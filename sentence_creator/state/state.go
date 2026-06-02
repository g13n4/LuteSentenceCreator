package state

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
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
	err := godotenv.Load()

	_, dockerErr := os.Stat("/.dockerenv")
	if err != nil && dockerErr != nil {
		log.Fatal("Error loading .env file")
	}

	url := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_ADDRESS"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
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
