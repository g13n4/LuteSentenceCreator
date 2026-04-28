package postgres

import (
	"context"

	"github.com/g13n4/LuteSentencePicker/domain"
	"github.com/g13n4/LuteSentencePicker/jmdict"
	"github.com/g13n4/LuteSentencePicker/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dictionaryRepository struct {
	pool *pgxpool.Pool
}

func (dr *dictionaryRepository) GetIdByName(ctx context.Context, name string) (utils.PostgresID, error) {
	var objId utils.PostgresID
	err := dr.pool.QueryRow(ctx, "SELECT id FROM entry_dictionaries WHERE name=$1", name).Scan(&objId)
	return objId, err
}

func (dr *dictionaryRepository) Save(ctx context.Context, obj jmdict.DictionaryEntry) error {
	query := "INSERT INTO entry_dictionaries (name, description) VALUES ($1, $2)"
	_, err := dr.pool.Exec(ctx, query, obj.Category, obj.Description)
	return err
}

func NewDictionaryRepository(pool *pgxpool.Pool) domain.DictionaryRepository {
	return &dictionaryRepository{pool: pool}
}
