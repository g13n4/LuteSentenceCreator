package postgres

import (
	"context"

	"github.com/g13n4/LuteSentencePicker/domain"
	"github.com/g13n4/LuteSentencePicker/jmdict"
	"github.com/g13n4/LuteSentencePicker/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type entryRepository struct {
	pool *pgxpool.Pool
}

func (e *entryRepository) GetIdByReading(ctx context.Context, reading string) (utils.PostgresID, error) {
	var objId utils.PostgresID
	err := e.pool.QueryRow(ctx, "SELECT id FROM readings WHERE reading=$1 AND no_kanji=TRUE", reading).Scan(&objId)
	return objId, err
}

func (e *entryRepository) GetIdByKanjiReading(ctx context.Context, kanjiReading string) (utils.PostgresID, error) {
	var objId utils.PostgresID
	err := e.pool.QueryRow(ctx, "SELECT id FROM readings WHERE reading=$1 AND no_kanji=False", kanjiReading).Scan(&objId)
	return objId, err
}

func (e *entryRepository) getReadings(obj jmdict.Entry) *[][]any {
	readingLen := len(obj.Reading)
	readings := make([][]any, readingLen+len(obj.KanjiReading))
	for i, reading := range obj.Reading {
		readings[i] = []any{obj.EntryId, reading, false}
	}
	for i, reading := range obj.KanjiReading {
		readings[i+readingLen] = []any{obj.EntryId, reading, true}
	}
	return &readings
}

func (e *entryRepository) Save(ctx context.Context, obj jmdict.Entry) error {
	readings := e.getReadings(obj)

	_, err := e.pool.CopyFrom(
		context.Background(),
		pgx.Identifier{"readings"},
		[]string{"entry_id", "reading", "no_kanji"},
		pgx.CopyFromRows(*readings),
	)
	return err
}

func (e *entryRepository) BulkSave(ctx context.Context, objs []jmdict.Entry) error {
	allReadings := make([][]any, len(objs))
	for _, obj := range objs {
		readings := e.getReadings(obj)
		allReadings = append(allReadings, *readings...)
	}

	_, err := e.pool.CopyFrom(
		context.Background(),
		pgx.Identifier{"readings"},
		[]string{"entry_id", "reading", "no_kanji"},
		pgx.CopyFromRows(allReadings),
	)
	return err
}

func NewEntryRepository(pool *pgxpool.Pool) domain.EntryRepository {
	return &entryRepository{pool: pool}
}
