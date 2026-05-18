package state

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

const DefaultBatchSize = 512

type Singleton struct {
	Pool      *pgxpool.Pool
	DBUrl     string
	BatchSize int

	KanjiPool map[int]struct{}
	EntryPool map[string][]int
}

func GetStateSingleton() *Singleton {
	url := fmt.Sprintf(
		"%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DATABASE_USERNAME"),
		os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_ADDRESS"),
		os.Getenv("DATABASE_PORT"),
		os.Getenv("DATABASE_NAME"),
	)

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		panic(err)
	}

	batchSize, err := strconv.Atoi(os.Getenv("BATCH_SIZE"))
	if err != nil {
		batchSize = DefaultBatchSize
	}

	return &Singleton{
		Pool:      pool,
		DBUrl:     url,
		BatchSize: batchSize,

		KanjiPool: map[int]struct{}{},
		EntryPool: map[string][]int{},
	}

}
