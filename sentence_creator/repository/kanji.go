package repository

import (
	"context"
	"fmt"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/domain"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/kanji"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/utils"
)

type kanjiRepository struct {
	db DBSaver
}

func NewKanjiRepository(db DBSaver) domain.KanjiRepository {
	return &kanjiRepository{db: db}
}

func (k *kanjiRepository) Save(ctx context.Context, obj *kanji.Kanji) error {
	literalVal := utils.GetUTFValue(obj.Literal)
	query := "INSERT INTO kanjis (id, literal, jlpt, freq, grade, stroke_count) VALUES ($1, $2, $3, $4, $5, $6)"
	_, err := k.db.Exec(ctx, query, literalVal, obj.Literal, obj.JLPT, obj.Frequency, obj.Grade, obj.StrokeCount)
	return err
}

func (k *kanjiRepository) BulkSave(objs []*kanji.Kanji) error {
	kanjis := make([][]any, 0)
	for _, obj := range objs {
		if obj == nil {
			continue
		}

		literalVal := utils.GetUTFValue(obj.Literal)
		kanjis = append(kanjis, []any{literalVal, obj.Literal, obj.JLPT, obj.Frequency, obj.Grade, obj.StrokeCount})
	}

	_, err := k.db.CopyFrom(
		context.Background(),
		pgx.Identifier{"kanjis"},
		[]string{"id", "literal", "jlpt", "freq", "grade", "stroke_count"},
		pgx.CopyFromRows(kanjis),
	)

	return err
}

func (k *kanjiRepository) GetUniqueFields(ctx context.Context, field string) ([]int, error) {
	query := fmt.Sprintf("SELECT DISTINCT %s FROM kanjis WHERE %s IS NOT NULL ORDER BY %s", field, field, field)
	rows, err := k.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	out := make([]int, 0)
	for rows.Next() {
		var s int

		err := rows.Scan(&s)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}

	return out, nil
}
