package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

func NewPostgresClient() *pgxpool.Pool {
	// databaseUrl = os.Getenv("DATABASE_URL")
	databaseUrl := "postgres://clowman:asdf@localhost:5433/postgres"

	dbPool, err := pgxpool.Connect(context.Background(), databaseUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	  }
	defer dbPool.Close()

	return dbPool
}