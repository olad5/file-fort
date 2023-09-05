package data

import (
	"log"

	"github.com/jmoiron/sqlx"
)

func StartPostgres(DatabaseUrl string) *sqlx.DB {
	connection, err := sqlx.Connect("postgres", DatabaseUrl)
	if err != nil {
		log.Fatal("Failed to create PostgresConnection:  %w", err)
	}
	return connection
}
