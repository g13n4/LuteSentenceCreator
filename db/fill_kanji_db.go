package db

import (
	"context"
	"io"

	"github.com/g13n4/LuteSentencePicker/kanji"
	"github.com/g13n4/LuteSentencePicker/parser"
	"github.com/g13n4/LuteSentencePicker/repository"
	"github.com/g13n4/LuteSentencePicker/state"
	"github.com/g13n4/LuteSentencePicker/utils"
)

func FillKanji(ss *state.Singleton, r io.Reader) error {
	ctx := context.Background()
	tx, err := ss.Pool.Begin(ctx)
	defer func() {
		err := tx.Rollback(ctx)
		panic(err)
	}()

	if err != nil {
		return err
	}

	kanjiRepo := repository.NewKanjiRepository(tx)

	kChan := parser.CreateXMLParsingChan[*kanji.Kanji](r, kanji.KanjiNodeName, ss.BatchSize)
	kSaver := utils.NewBatchSaveHelper[*kanji.Kanji](kanjiRepo, ss.BatchSize)

	for k := range kChan {
		literal := int(k.Literal[0])
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
