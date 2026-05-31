package repository

import (
	"context"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/domain"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/jmdict"
	utils2 "github.com/g13n4/LuteSentencePicker/sentence_creator/utils"
)

type entryRepository struct {
	db DBSaver
	utils2.Cache[string, utils2.PostgresID]
}

func (e *entryRepository) GetIdByReading(ctx context.Context, reading string) (utils2.PostgresID, error) {
	var objId utils2.PostgresID
	err := e.db.QueryRow(ctx, "SELECT id FROM readings WHERE reading=$1 AND no_kanji=TRUE ORDER BY in_news ASC", reading).Scan(&objId)
	return objId, err
}

func (e *entryRepository) GetIdByKanjiReading(ctx context.Context, kanjiReading string) (utils2.PostgresID, error) {
	var objId utils2.PostgresID
	err := e.db.QueryRow(ctx, "SELECT id FROM readings WHERE reading=$1 AND no_kanji=False ORDER BY in_news ASC", kanjiReading).Scan(&objId)
	return objId, err
}

func (e *entryRepository) getReadings(obj *jmdict.Entry) *[][]any {
	readings := make([][]any, len(obj.Readings))
	for i, reading := range obj.Readings {
		readings[i] = []any{reading.CombinedId, obj.EntryId, reading.Word, reading.IsKanji, obj.IsInNews()}
	}

	return &readings
}

func (e *entryRepository) BulkSave(objs []*jmdict.Entry) error {
	allReadings := make([][]any, 0)
	for _, obj := range objs {
		if obj == nil {
			continue
		}

		readings := e.getReadings(obj)
		allReadings = append(allReadings, *readings...)
	}

	_, err := e.db.CopyFrom(
		context.Background(),
		pgx.Identifier{"readings"},
		[]string{"id", "entry", "reading", "kanji", "in_news"},
		pgx.CopyFromRows(allReadings),
	)
	return err
}

func NewEntryRepository(db DBSaver) domain.EntryRepository {
	return &entryRepository{db: db, Cache: *utils2.NewCache[string, utils2.PostgresID](3)}
}
