package repository

import (
	"context"

	conns "github.com/g13n4/LuteSentencePicker/connections"
	"github.com/g13n4/LuteSentencePicker/domain"
	"github.com/jackc/pgx/v5"
)

type sentenceReadingConnectionsRepository struct {
	db DBSaver
}

func (src *sentenceReadingConnectionsRepository) BulkSave(objs []*conns.SentenceReading) error {
	connections := make([][]any, len(objs))
	for i, o := range objs {
		connections[i] = []any{o.ReadingId, o.SentenceId}
	}

	_, err := src.db.CopyFrom(
		context.Background(),
		pgx.Identifier{"sentences__mtm__readings"},
		[]string{"r_id", "s_id"},
		pgx.CopyFromRows(connections),
	)

	return err
}

func NewSentenceReadingConnectionsRepository(db DBSaver) domain.ConnectionsRepository[*conns.SentenceReading] {
	return &sentenceReadingConnectionsRepository{db: db}
}
