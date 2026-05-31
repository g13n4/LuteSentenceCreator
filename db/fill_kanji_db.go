package db

import (
	"context"
	"errors"
	"io"

	"github.com/g13n4/LuteSentencePicker/kanji"
	"github.com/g13n4/LuteSentencePicker/parser"
	"github.com/g13n4/LuteSentencePicker/repository"
	"github.com/g13n4/LuteSentencePicker/state"
	"github.com/g13n4/LuteSentencePicker/utils"
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

	kChan := parser.CreateXMLParsingChan[*kanji.Kanji](r, kanji.KanjiNodeName, ss.BatchSize)
	kSaver := utils.NewBatchSaveHelper[*kanji.Kanji](kanjiRepo, ss.BatchSize)

	for k := range kChan {
		literal := utils.GetUTFValue(k.Literal)
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
