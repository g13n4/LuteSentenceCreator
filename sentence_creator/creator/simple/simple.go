package simple

import (
	"bufio"
	"context"
	"os"

	mw "github.com/g13n4/LuteSentencePicker/sentence_creator/middleware"
	"github.com/g13n4/LuteSentencePicker/sentence_creator/repository"
)

type Executor struct {
	db              repository.DBSaver
	maxPermutations int
}

type SentenceReadingRow struct {
	SentenceId int64
	ReadingId  int64
}

func NewExecutor(db repository.DBSaver) *Executor {
	return &Executor{
		db:              db,
		maxPermutations: 1,
	}
}

func (exc *Executor) GetSentences(ctx context.Context, outputFile *os.File, qh mw.QueryHelper, limit int) error {
	rows, err := exc.db.Query(ctx, qh.CreateSimpleQuery(limit))
	if err != nil {
		return err
	}

	var sentence string
	writer := bufio.NewWriter(outputFile)
	for rows.Next() {
		err := rows.Scan(&sentence)
		if err != nil {
			return err
		}
		_, err = writer.WriteString(sentence + "\n")
		if err != nil {
			return err
		}
	}

	err = writer.Flush()

	return err
}
