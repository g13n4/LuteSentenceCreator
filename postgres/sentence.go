package postgres

import (
	"context"

	"github.com/g13n4/LuteSentencePicker/domain"
	"github.com/g13n4/LuteSentencePicker/tatoeba"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sentenceRepository struct {
	pool *pgxpool.Pool
}

func (k *sentenceRepository) Save(ctx context.Context, sentence tatoeba.Sentence) error {
	query := "INSERT INTO sentences (id, sentence) VALUES ($1, $2)"
	_, err := k.pool.Exec(ctx, query, sentence.Id, sentence.Text)
	return err
}

func (k *sentenceRepository) BulkSave(ctx context.Context, sentences []tatoeba.Sentence) error {
	bulkSentences := make([][]any, len(sentences))
	for _, sentence := range sentences {
		bulkSentences = append(bulkSentences, []any{sentence.Id, sentence.Text})
	}

	_, err := k.pool.CopyFrom(
		context.Background(),
		pgx.Identifier{"sentences"},
		[]string{"id", "sentence"},
		pgx.CopyFromRows(bulkSentences),
	)

	return err
}

func NewSentenceRepository(pool *pgxpool.Pool) domain.SentenceRepository {
	return &sentenceRepository{pool: pool}
}
