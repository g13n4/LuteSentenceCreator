package repository

import (
	"context"

	conns "github.com/g13n4/LuteSentencePicker/connections"
	"github.com/g13n4/LuteSentencePicker/domain"
	"github.com/jackc/pgx/v5"
)

type dictionaryEntryConnectionsRepository struct {
	db DBSaver
}

func (dec *dictionaryEntryConnectionsRepository) BulkSave(objs []*conns.DictionaryEntry) error {
	connections := make([][]any, len(objs))
	for i, o := range objs {

		connections[i] = []any{o.Entry, o.DictionaryId}
	}

	_, err := dec.db.CopyFrom(
		context.Background(),
		pgx.Identifier{"dictionaries__mtm__entries"},
		[]string{"entry", "ed_id"},
		pgx.CopyFromRows(connections),
	)

	return err
}

func NewDictionaryEntryConnectionsRepository(db DBSaver) domain.ConnectionsRepository[*conns.DictionaryEntry] {
	return &dictionaryEntryConnectionsRepository{db: db}
}
