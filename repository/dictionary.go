package repository

import (
	"context"

	"github.com/g13n4/LuteSentencePicker/domain"
	"github.com/g13n4/LuteSentencePicker/jmdict"
	"github.com/jackc/pgx/v5"
)

type dictionaryRepository struct {
	db DBSaver
}

func (dr *dictionaryRepository) BulkSave(objs *[]*jmdict.Dictionary) error {
	connections := make([][]any, len(*objs))
	for i, d := range *objs {
		connections[i] = []any{d.Id, d.Name}
	}

	_, err := dr.db.CopyFrom(
		context.Background(),
		pgx.Identifier{"dictionaries"},
		[]string{"id", "name"},
		pgx.CopyFromRows(connections),
	)

	return err
}

func NewDictionaryRepository(db DBSaver) domain.DictionaryRepository {
	return &dictionaryRepository{db: db}
}
