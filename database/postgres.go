package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

func NewPostgresClient() *pgxpool.Pool {
	databaseUrl := os.Getenv("DATABASE_URL")

	dbPool, err := pgxpool.Connect(context.Background(), databaseUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	// TODO: close db connection?

	queryString := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS public.user
	(
		user_id       int GENERATED ALWAYS AS IDENTITY primary key,
		first_name    varchar(100) not null,
		last_name     varchar(100) not null,
		email 		  varchar(100) not null,
		pass_hash 	  varchar(100) not null,
		auth_level 	  int not null,
		session_id 	  int,
		spotify_access_token varchar(200),
		spotify_refresh_token varchar (150)
	);

	UPDATE public.user
	SET session_id = 0;
	`)

	_, err = dbPool.Exec(context.Background(), queryString)
	if err != nil {
		print("Error initializing database")
	}

	return dbPool
}