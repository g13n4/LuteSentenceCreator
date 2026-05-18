package db

import (
	"context"
	"io"

	conns "github.com/g13n4/LuteSentencePicker/connections"
	"github.com/g13n4/LuteSentencePicker/jmdict"
	"github.com/g13n4/LuteSentencePicker/parser"
	"github.com/g13n4/LuteSentencePicker/repository"
	"github.com/g13n4/LuteSentencePicker/state"
	"github.com/g13n4/LuteSentencePicker/utils"
)

func FillEntry(ss *state.Singleton, entryData io.Reader) error {
	dictPool := jmdict.NewDictionaryPool()

	ctx := context.Background()
	tx, err := ss.Pool.Begin(ctx)
	defer func() {
		err := tx.Rollback(ctx)
		panic(err)
	}()

	if err != nil {
		return err
	}

	entryRepo := repository.NewEntryRepository(tx)
	dictRepo := repository.NewDictionaryRepository(tx)
	rkRepo := repository.NewReadingKanjiConnectionsRepository(tx)
	deRepo := repository.NewDictionaryEntryConnectionsRepository(tx)

	err = dictRepo.BulkSave(dictPool.GetAllDictionaries())
	if err != nil {
		return err
	}

	eChan := parser.CreateXMLParsingChan[*jmdict.Entry](entryData, jmdict.EntryNodeName, ss.BatchSize)

	eSaver := utils.NewBatchSaveHelper[*jmdict.Entry](entryRepo, ss.BatchSize)

	rkSaver := utils.NewBulkSaveHelper[*conns.ReadingKanji](rkRepo, ss.BatchSize)
	deSaver := utils.NewBulkSaveHelper[*conns.DictionaryEntry](deRepo, ss.BatchSize)

	for e := range eChan {
		for _, r := range e.Readings {
			_, ok := ss.EntryPool[r.Word]
			if !ok {
				ss.EntryPool[r.Word] = make([]int, 1)
			}
			ss.EntryPool[r.Word] = append(ss.EntryPool[r.Word], r.OrderId)

			if r.IsKanji {
				for _, k := range r.Word {
					kInt := int(k)
					_, ok := ss.KanjiPool[kInt]
					if ok {
						rk := conns.ReadingKanji{ReadingId: r.OrderId, KanjiId: kInt}
						rkSaver.Add(&rk)
					}
				}

			}
		}
		for _, dName := range *e.GetAllDictionaries() {
			dObj := dictPool.GetDictionary(dName)
			deObj := conns.DictionaryEntry{Entry: e.EntryId, DictionaryId: dObj.Id}
			deSaver.Add(&deObj)
		}

		err = eSaver.Add(e)
		if err != nil {
			return err
		}
	}

	err = eSaver.BulkSave(true)
	if err != nil {
		return err
	}

	err = deSaver.SaveInBatches()
	if err != nil {
		return err
	}

	err = rkSaver.SaveInBatches()
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
