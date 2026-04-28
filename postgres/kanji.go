package postgres

import (
	"context"

	"github.com/g13n4/LuteSentencePicker/domain"
	"github.com/g13n4/LuteSentencePicker/kanji"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type kanjiRepository struct {
	pool *pgxpool.Pool
}

func (k *kanjiRepository) Save(ctx context.Context, obj kanji.Kanji) error {
	query := "INSERT INTO kanjis (literal, jlpt, freq, grade, stroke_count) VALUES ($1, $2, $3, $4, $5)"
	_, err := k.pool.Exec(ctx, query, obj.Literal, obj.JLPT, obj.Frequency, obj.Grade, obj.StrokeCount)
	return err
}

func (k *kanjiRepository) BulkSave(ctx context.Context, objs []kanji.Kanji) error {
	kanjis := make([][]any, len(objs))
	for _, obj := range objs {
		kanjis = append(kanjis, []any{obj.Literal, obj.JLPT, obj.Frequency, obj.Grade, obj.StrokeCount})
	}

	_, err := k.pool.CopyFrom(
		context.Background(),
		pgx.Identifier{"readings"},
		[]string{"literal", "jlpt", "freq", "grade", "stroke_count"},
		pgx.CopyFromRows(kanjis),
	)

	return err
}

func NewKanjiRepository(pool *pgxpool.Pool) domain.KanjiRepository {
	return &kanjiRepository{pool: pool}
}
