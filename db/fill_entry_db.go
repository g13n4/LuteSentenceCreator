package db

import (
	"context"
	"errors"
	"io"

	conns "github.com/g13n4/LuteSentencePicker/connections"
	"github.com/g13n4/LuteSentencePicker/jmdict"
	"github.com/g13n4/LuteSentencePicker/parser"
	"github.com/g13n4/LuteSentencePicker/repository"
	"github.com/g13n4/LuteSentencePicker/state"
	"github.com/g13n4/LuteSentencePicker/utils"
	"github.com/jackc/pgx/v5"
)

func FillEntry(ss *state.Singleton, entryData io.Reader) error {
	dictPool := jmdict.NewDictionaryPool()

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

	entryRepo := repository.NewEntryRepository(tx)
	dictCatRepo := repository.NewDictionaryCategoryRepository(tx)
	dictRepo := repository.NewDictionaryRepository(tx)
	rkRepo := repository.NewReadingKanjiConnectionsRepository(tx)
	deRepo := repository.NewDictionaryEntryConnectionsRepository(tx)

	err = dictCatRepo.BulkSave(dictPool.GetAllCategories())
	if err != nil {
		return err
	}

	eChan := parser.CreateXMLParsingChan[*jmdict.Entry](entryData, jmdict.EntryNodeName, ss.BatchSize)

	eSaver := utils.NewBatchSaveHelper[*jmdict.Entry](entryRepo, ss.BatchSize)

	rkSaver := utils.NewBulkSaveHelper[*conns.ReadingKanji](rkRepo, ss.BatchSize)
	deSaver := utils.NewBulkSaveHelper[*conns.DictionaryEntry](deRepo, ss.BatchSize)

	uniqueKanjiTOReadingConnection := make(map[int]struct{})
	uniqueReadingDictionary := make(map[int]struct{})
	for e := range eChan {
		if !e.IsPopular() {
			continue
		}

		for _, r := range e.Readings {
			_, ok := ss.EntryPool[r.Word]
			if !ok {
				ss.EntryPool[r.Word] = make([]int, 0)
			}
			ss.EntryPool[r.Word] = append(ss.EntryPool[r.Word], r.CombinedId)

			if r.IsKanji {
				for _, k := range r.Word {
					kInt := int(k)
					_, ok := uniqueKanjiTOReadingConnection[kInt]
					if ok {
						continue
					}

					_, ok = ss.KanjiPool[kInt]
					if ok {
						rk := conns.ReadingKanji{ReadingId: r.CombinedId, KanjiId: kInt}
						rkSaver.Add(&rk)

						uniqueKanjiTOReadingConnection[kInt] = struct{}{}
					}
				}
				clear(uniqueKanjiTOReadingConnection)

			}
		}
		for _, dName := range *e.GetAllDictionaries() {
			dictObj, dictCatObj := dictPool.GetDictionaryData(dName)

			_, ok := uniqueReadingDictionary[dictObj.Id]
			if ok {
				continue
			}

			deObj := conns.DictionaryEntry{Entry: e.EntryId, DictionaryId: dictObj.Id, DictionaryCategoryId: dictCatObj.Id}
			deSaver.Add(&deObj)

			uniqueReadingDictionary[dictObj.Id] = struct{}{}
		}

		err = eSaver.Add(e)
		if err != nil {
			return err
		}
	}

	err = dictRepo.BulkSave(dictPool.GetAllDictionaries())
	if err != nil {
		return err
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
