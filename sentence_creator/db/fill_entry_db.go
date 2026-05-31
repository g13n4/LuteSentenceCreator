package db

import (
	"context"
	"errors"
	"io"

	conns "github.com/g13n4/LuteSentencePicker/sentence_creator/connections"
	jmdict2 "github.com/g13n4/LuteSentencePicker/sentence_creator/jmdict"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/parser"
	repository2 "github.com/g13n4/LuteSentencePicker/sentence_creator/repository"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/state"
	utils2 "github.com/g13n4/LuteSentencePicker/sentence_creator/utils"
	"github.com/jackc/pgx/v5"
)

func FillEntry(ss *state.Singleton, entryData io.Reader) error {
	dictPool := jmdict2.NewDictionaryPool()

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

	entryRepo := repository2.NewEntryRepository(tx)
	dictCatRepo := repository2.NewDictionaryCategoryRepository(tx)
	dictRepo := repository2.NewDictionaryRepository(tx)
	rkRepo := repository2.NewReadingKanjiConnectionsRepository(tx)
	deRepo := repository2.NewDictionaryEntryConnectionsRepository(tx)

	err = dictCatRepo.BulkSave(dictPool.GetAllCategories())
	if err != nil {
		return err
	}

	eChan := parser.CreateXMLParsingChan[*jmdict2.Entry](entryData, jmdict2.EntryNodeName, ss.BatchSize)

	eSaver := utils2.NewBatchSaveHelper[*jmdict2.Entry](entryRepo, ss.BatchSize)

	rkSaver := utils2.NewBulkSaveHelper[*conns.ReadingKanji](rkRepo, ss.BatchSize)
	deSaver := utils2.NewBulkSaveHelper[*conns.DictionaryEntry](deRepo, ss.BatchSize)

	uniqueKanjiTOReadingConnection := make(map[int]struct{})

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

			deObj := conns.DictionaryEntry{Entry: e.EntryId, DictionaryId: dictObj.Id, DictionaryCategoryId: dictCatObj.Id}
			deSaver.Add(&deObj)

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
