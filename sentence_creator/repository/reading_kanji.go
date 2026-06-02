package repository

import (
	"context"

	conns "github.com/g13n4/LuteSentencePicker/sentence_creator/connections"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/domain"
	"github.com/jackc/pgx/v5"
)

type readingKanjiConnectionsRepository struct {
	db DBSaver
}

func (rkc *readingKanjiConnectionsRepository) BulkSave(objs []*conns.ReadingKanji) error {
	connections := make([][]any, len(objs))
	for i, o := range objs {
		connections[i] = []any{o.ReadingId, o.KanjiId}
	}

	_, err := rkc.db.CopyFrom(
		context.Background(),
		pgx.Identifier{"readings__mtm__kanjis"},
		[]string{"r_id", "k_id"},
		pgx.CopyFromRows(connections),
	)

	return err
}

func NewReadingKanjiConnectionsRepository(db DBSaver) domain.ConnectionsRepository[*conns.ReadingKanji] {
	return &readingKanjiConnectionsRepository{db: db}
}
