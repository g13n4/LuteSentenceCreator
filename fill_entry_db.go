//go:build ignore

package main

import (
	"github.com/jackc/pgx/v5"
)

func InitiateDB(conn *pgx.Conn) (bool, err) {

	if err != nil {
		panic(err)
	}

	conn.Exec()
}
