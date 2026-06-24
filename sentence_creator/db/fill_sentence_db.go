package db

import (
	"context"
	"errors"
	"io"

	conns "github.com/g13n4/LuteSentencePicker/sentence_creator/connections"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/parser"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/repository"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/state"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/tatoeba"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/utils"
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

	uniqueSentenceFilter := utils.NewSentenceFilter()
	uniqueReadingsList := make([]*int, 0, 1_000)

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

		counter := 0
		for rId := range uniqueReadings {
			uniqueReadingsList[counter] = &rId
			counter++
		}

		if uniqueSentenceFilter.Fits(&uniqueReadingsList) {
			for rId := range uniqueReadings {
				srSaver.Add(
					&conns.SentenceReading{
						SentenceId: s.Id,
						ReadingId:  rId,
					},
				)
			}
		}

		clear(uniqueReadings)
		clear(uniqueReadingsList)
	}

	err = srSaver.SaveInBatches()
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
