package repository

import (
	"context"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/domain"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/jmdict"
	"github.com/jackc/pgx/v5"
)

type dictionaryCategoryRepository struct {
	db DBSaver
}

func (dr *dictionaryCategoryRepository) BulkSave(dictionaries *[]*jmdict.DictionaryCategory) error {
	connections := make([][]any, len(*dictionaries))
	for i, d := range *dictionaries {
		connections[i] = []any{d.Id, d.Name, d.Description}
	}

	_, err := dr.db.CopyFrom(
		context.Background(),
		pgx.Identifier{"dictionary_categories"},
		[]string{"id", "name", "description"},
		pgx.CopyFromRows(connections),
	)

	return err
}

func NewDictionaryCategoryRepository(db DBSaver) domain.DictionaryCategoryRepository {
	return &dictionaryCategoryRepository{db: db}
}
