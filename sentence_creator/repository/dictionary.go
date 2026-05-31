package repository

import (
	"context"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/domain"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/jmdict"
	"github.com/jackc/pgx/v5"
)

type dictionaryRepository struct {
	db DBSaver
}

func (dr *dictionaryRepository) GetDictionaries(ctx context.Context) ([]*jmdict.Dictionary, error) {
	query := "SELECT id, name, category, number from dictionaries"
	rows, err := dr.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	out := make([]*jmdict.Dictionary, 0)
	for rows.Next() {
		var d jmdict.Dictionary

		err := rows.Scan(&d.Id, &d.Name, &d.Category, &d.Number)
		if err != nil {
			return nil, err
		}
		out = append(out, &d)
	}

	return out, nil
}

func (dr *dictionaryRepository) BulkSave(objs *[]*jmdict.Dictionary) error {
	connections := make([][]any, len(*objs))
	for i, d := range *objs {
		connections[i] = []any{d.Id, d.Name, d.Category, d.Number}
	}

	_, err := dr.db.CopyFrom(
		context.Background(),
		pgx.Identifier{"dictionaries"},
		[]string{"id", "name", "category", "number"},
		pgx.CopyFromRows(connections),
	)

	return err
}

func NewDictionaryRepository(db DBSaver) domain.DictionaryRepository {
	return &dictionaryRepository{db: db}
}
