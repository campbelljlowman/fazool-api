package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

func NewPostgresClient() *pgxpool.Pool {
	// databaseUrl = os.Getenv("DATABASE_URL")
	databaseUrl := "postgres://clowman:asdf@localhost:5433/fazool"

	dbPool, err := pgxpool.Connect(context.Background(), databaseUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	// TODO: close db connection?
	return dbPool
}