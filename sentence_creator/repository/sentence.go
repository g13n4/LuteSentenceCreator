package repository

import (
	"context"

	"github.com/g13n4/LuteSentencePicker/sentence_creator/domain"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/tatoeba"
	"github.com/jackc/pgx/v5"
)

type sentenceRepository struct {
	db DBSaver
}

func (k *sentenceRepository) Save(ctx context.Context, sentence tatoeba.Sentence) error {
	query := "INSERT INTO sentences (id, sentence) VALUES ($1, $2)"
	_, err := k.db.Exec(ctx, query, sentence.Id, sentence.Text)
	return err
}

func (k *sentenceRepository) BulkSave(sentences []*tatoeba.Sentence) error {
	bulkSentences := make([][]any, 0)
	for _, sentence := range sentences {
		if sentence == nil {
			continue
		}
		bulkSentences = append(bulkSentences, []any{sentence.Id, sentence.Text})
	}

	_, err := k.db.CopyFrom(
		context.Background(),
		pgx.Identifier{"sentences"},
		[]string{"id", "sentence"},
		pgx.CopyFromRows(bulkSentences),
	)

	return err
}

func NewSentenceRepository(db DBSaver) domain.SentenceRepository {
	return &sentenceRepository{db: db}
}
