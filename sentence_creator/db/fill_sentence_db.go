package db

import (
	"context"
	"errors"
	"io"

	conns "github.com/g13n4/LuteSentencePicker/sentence_creator/connections"
	parser2 "github.com/g13n4/LuteSentencePicker/sentence_creator/parser"
	repository2 "github.com/g13n4/LuteSentencePicker/sentence_creator/repository"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/state"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/tatoeba"
	utils2 "github.com/g13n4/LuteSentencePicker/sentence_creator/utils"
	"github.com/jackc/pgx/v5"
)

func FillSentence(ss *state.Singleton, sentencesR, parsedSentencedR io.Reader) error {
	ctx := context.Background()
	tx, err := ss.Pool.Begin(ctx)
	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			panic(err)
		}
	}()

	if err != nil {
		return err
	}

	sentenceRepo := repository2.NewSentenceRepository(tx)

	sChan := parser2.CreateTSVParsingChan(sentencesR, ss.BatchSize)
	kSaver := utils2.NewBatchSaveHelper[*tatoeba.Sentence](sentenceRepo, ss.BatchSize)

	for s := range sChan {
		err = kSaver.Add(s)
		if err != nil {
			return err
		}
	}

	err = kSaver.BulkSave(true)
	if err != nil {
		return err
	}

	sentenceReadingRepo := repository2.NewSentenceReadingConnectionsRepository(tx)

	tsChan := parser2.CreateSudachiTSVParsingChan(parsedSentencedR, ss.BatchSize)
	srSaver := utils2.NewBulkSaveHelper[*conns.SentenceReading](sentenceReadingRepo, ss.BatchSize)

	uniqueReadings := make(map[int]struct{})
	for s := range tsChan {
		for _, t := range *s.Tokens {
			readingIds, ok := ss.EntryPool[t]
			if ok {
				for _, rId := range readingIds {
					uniqueReadings[rId] = struct{}{}
				}
			}

		}

		for rId := range uniqueReadings {
			srSaver.Add(
				&conns.SentenceReading{
					SentenceId: s.Id,
					ReadingId:  rId,
				},
			)
		}
		clear(uniqueReadings)
	}

	err = srSaver.SaveInBatches()
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
