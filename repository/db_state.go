package repository

import (
	"context"

	"github.com/g13n4/LuteSentencePicker/domain"
)

type dbStateRepository struct {
	db DBSaver
}

func NewDBStateRepository(db DBSaver) domain.DBStateRepository {
	return &dbStateRepository{db: db}
}

func (dbsr *dbStateRepository) SetStatus(ctx context.Context, val int) error {
	query := "UPDATE db_state SET status = $1 WHERE id = 999"
	_, err := dbsr.db.Exec(ctx, query, val)
	return err
}

func (dbsr *dbStateRepository) GetStatus(ctx context.Context) (int, error) {
	query := "SELECT status FROM db_state WHERE id = 999"
	var status int
	err := dbsr.db.QueryRow(ctx, query).Scan(&status)

	return status, err
}
