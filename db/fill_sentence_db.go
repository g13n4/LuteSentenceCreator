package db

import (
	"context"
	"errors"
	"io"

	conns "github.com/g13n4/LuteSentencePicker/connections"
	"github.com/g13n4/LuteSentencePicker/parser"
	"github.com/g13n4/LuteSentencePicker/repository"
	"github.com/g13n4/LuteSentencePicker/state"
	"github.com/g13n4/LuteSentencePicker/tatoeba"
	"github.com/g13n4/LuteSentencePicker/utils"
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

	sentenceRepo := repository.NewSentenceRepository(tx)

	sChan := parser.CreateTSVParsingChan(sentencesR, ss.BatchSize)
	kSaver := utils.NewBatchSaveHelper[*tatoeba.Sentence](sentenceRepo, ss.BatchSize)

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

	sentenceReadingRepo := repository.NewSentenceReadingConnectionsRepository(tx)

	tsChan := parser.CreateSudachiTSVParsingChan(parsedSentencedR, ss.BatchSize)
	srSaver := utils.NewBulkSaveHelper[*conns.SentenceReading](sentenceReadingRepo, ss.BatchSize)

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

		for k, _ := range uniqueReadings {
			srSaver.Add(
				&conns.SentenceReading{
					SentenceId: s.Id,
					ReadingId:  k,
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
