//go:build ignore

package main

import (
	"os"

	"github.com/jackc/pgx/v5"
)

func InitiateDB(conn *pgx.Conn) (bool, err) {
	sqlFile, err := os.ReadFile("init.sql")

	if err != nil {
		panic(err)
	}

	conn.Exec()
}
