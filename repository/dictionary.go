package repository

import (
	"context"

	"github.com/g13n4/LuteSentencePicker/domain"
	"github.com/g13n4/LuteSentencePicker/jmdict"
	"github.com/g13n4/LuteSentencePicker/utils"
	"github.com/jackc/pgx/v5"
)

type dictionaryRepository struct {
	db DBSaver
}

func (dr *dictionaryRepository) GetMap(ctx context.Context) (*map[string]utils.PostgresID, error) {
	rows, err := dr.db.Query(ctx, "SELECT id, name  FROM dictionaries")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var output map[string]utils.PostgresID
	for rows.Next() {
		var rowId utils.PostgresID
		var rowName string
		if err := rows.Scan(&rowId, &rowName); err != nil {
			return nil, err
		}
		output[rowName] = rowId
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &output, err
}

func (dr *dictionaryRepository) GetIdByName(ctx context.Context, name string) (utils.PostgresID, error) {
	var objId utils.PostgresID
	err := dr.db.QueryRow(ctx, "SELECT id FROM dictionaries WHERE name=$1", name).Scan(&objId)
	return objId, err
}

func (dr *dictionaryRepository) BulkSave(dictionaries *[]jmdict.Dictionary) error {
	connections := make([][]any, len(*dictionaries))
	for i, d := range *dictionaries {
		connections[i] = []any{d.Id, d.Category, d.Description}
	}

	_, err := dr.db.CopyFrom(
		context.Background(),
		pgx.Identifier{"sentences__mtm__readings"},
		[]string{"id", "name", "description"},
		pgx.CopyFromRows(connections),
	)

	return err
}

func NewDictionaryRepository(db DBSaver) domain.DictionaryRepository {
	return &dictionaryRepository{db: db}
}
