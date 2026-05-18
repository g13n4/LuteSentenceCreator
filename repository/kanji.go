package repository

import (
	"context"

	"github.com/g13n4/LuteSentencePicker/domain"
	"github.com/g13n4/LuteSentencePicker/kanji"
	"github.com/jackc/pgx/v5"
)

type kanjiRepository struct {
	db DBSaver
}

func NewKanjiRepository(db DBSaver) domain.KanjiRepository {
	return &kanjiRepository{db: db}
}

func (k *kanjiRepository) Save(ctx context.Context, obj *kanji.Kanji) error {
	literalVal := rune(obj.Literal[0])
	query := "INSERT INTO kanjis (literal, jlpt, freq, grade, stroke_count) VALUES ($1, $2, $3, $4, $5, $6)"
	_, err := k.db.Exec(ctx, query, literalVal, obj.Literal, obj.JLPT, obj.Frequency, obj.Grade, obj.StrokeCount)
	return err
}

func (k *kanjiRepository) BulkSave(objs []*kanji.Kanji) error {
	kanjis := make([][]any, len(objs))
	for _, obj := range objs {
		if obj == nil {
			continue
		}

		literalVal := rune(obj.Literal[0])
		kanjis = append(kanjis, []any{literalVal, obj.Literal, obj.JLPT, obj.Frequency, obj.Grade, obj.StrokeCount})
	}

	_, err := k.db.CopyFrom(
		context.Background(),
		pgx.Identifier{"readings"},
		[]string{"id", "literal", "jlpt", "freq", "grade", "stroke_count"},
		pgx.CopyFromRows(kanjis),
	)

	return err
}
