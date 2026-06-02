package db

import (
	"context"
	"errors"
	"io"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/kanji"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/parser"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/repository"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/state"
	utils2 "github.com/g13n4/LuteSentencePicker/sentence_creator/utils"
	"github.com/jackc/pgx/v5"
)

func FillKanji(ss *state.Singleton, r io.Reader) error {
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

	kanjiRepo := repository.NewKanjiRepository(tx)

	kChan := parser.CreateXMLParsingChan[*kanji.Kanji](r, kanji.NodeName, ss.BatchSize)
	kSaver := utils2.NewBatchSaveHelper[*kanji.Kanji](kanjiRepo, ss.BatchSize)

	for k := range kChan {
		literal := utils2.GetUTFValue(k.Literal)
		ss.KanjiPool[literal] = struct{}{}
		err = kSaver.Add(k)
		if err != nil {
			return err
		}
	}

	err = kSaver.BulkSave(true)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
