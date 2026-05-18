package repository

import (
	"context"

	"github.com/g13n4/LuteSentencePicker/domain"
	"github.com/g13n4/LuteSentencePicker/jmdict"
	"github.com/g13n4/LuteSentencePicker/utils"
	"github.com/jackc/pgx/v5"
)

type entryRepository struct {
	db DBSaver
	utils.Cache[string, utils.PostgresID]
}

func (e *entryRepository) GetIdByReading(ctx context.Context, reading string) (utils.PostgresID, error) {
	var objId utils.PostgresID
	err := e.db.QueryRow(ctx, "SELECT id FROM readings WHERE reading=$1 AND no_kanji=TRUE ORDER BY in_news ASC", reading).Scan(&objId)
	return objId, err
}

func (e *entryRepository) GetIdByKanjiReading(ctx context.Context, kanjiReading string) (utils.PostgresID, error) {
	var objId utils.PostgresID
	err := e.db.QueryRow(ctx, "SELECT id FROM readings WHERE reading=$1 AND no_kanji=False ORDER BY in_news ASC", kanjiReading).Scan(&objId)
	return objId, err
}

func (e *entryRepository) getReadings(obj *jmdict.Entry) *[][]any {
	readings := make([][]any, len(obj.Readings))
	for i, reading := range obj.Readings {
		readings[i] = []any{reading.OrderId, obj.EntryId, reading.Word, reading.IsKanji, obj.IsInNews()}
	}

	return &readings
}

func (e *entryRepository) BulkSave(objs []*jmdict.Entry) error {
	allReadings := make([][]any, len(objs))
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
		[]string{"entry_id", "reading", "kanji", "in_news"},
		pgx.CopyFromRows(allReadings),
	)
	return err
}

func NewEntryRepository(db DBSaver) domain.EntryRepository {
	return &entryRepository{db: db, Cache: *utils.NewCache[string, utils.PostgresID](3)}
}
