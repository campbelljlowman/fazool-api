package database

import (
	"context"
	"fmt"
	"os"

	"github.com/campbelljlowman/fazool-api/graph/model"
	"golang.org/x/exp/slog"

	"github.com/jackc/pgx/v4/pgxpool"
)


type PostgresWrapper struct {
	postgresClient *pgxpool.Pool
}

func NewPostgresClient() *PostgresWrapper {
	databaseURL := os.Getenv("POSTRGRES_URL")

	dbPool, err := pgxpool.Connect(context.Background(), databaseURL)
	if err != nil {
		slog.Error("Unable to connect to database", err)
		os.Exit(1)
	}

	queryString := fmt.Sprintf(
	`CREATE TABLE IF NOT EXISTS public.user
	(
		user_id       			int GENERATED ALWAYS AS IDENTITY primary key,
		first_name    			varchar(100) not null,
		last_name     			varchar(100) not null,
		email 		 			varchar(100) not null,
		pass_hash 	  			varchar(100) not null,
		account_level 	  		varchar(20) not null,
		voter_level				varchar(20) not null,
		bonus_votes				int not null,
		session_id 	  			int,
		spotify_access_token 	varchar(200),
		spotify_refresh_token 	varchar(150)
	);

	UPDATE public.user
	SET session_id = 0;
	`)

	_, err = dbPool.Exec(context.Background(), queryString)
	if err != nil {
		slog.Error("Error initializing database", err)
		os.Exit(1)
	}

	pg := PostgresWrapper{dbPool}
	return &pg
}

func (p *PostgresWrapper) GetUserByEmail(userEmail string) (*model.User, error) {
	queryString := fmt.Sprintf(
	`SELECT user_id::VARCHAR, first_name, last_name, email, coalesce(session_id,0) as session_id FROM public.user WHERE email = '%v'`,
	userEmail)
	var sessionID int
	var userID, firstName, lastName, email string
	err := p.postgresClient.QueryRow(context.Background(), queryString).Scan(&userID, &firstName, &lastName, &email, &sessionID)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:        userID,
		FirstName: &firstName,
		LastName:  &lastName,
		Email:     &email,
		SessionID: &sessionID,
	}

	return user, nil
}

func (p *PostgresWrapper) GetUserByID(userID string) (*model.User, error) {
	queryString := fmt.Sprintf(
	`SELECT first_name, last_name, email, coalesce(session_id,0) as session_id FROM public.user WHERE user_id = '%v'`,
	userID)
	slog.Debug("Query string:", "query-string", queryString)

	var sessionID int
	var firstName, lastName, email string
	err := p.postgresClient.QueryRow(context.Background(), queryString).Scan(&firstName, &lastName, &email, &sessionID)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:        userID,
		FirstName: &firstName,
		LastName:  &lastName,
		Email:     &email,
		SessionID: &sessionID,
	}

	return user, nil
}

func (p *PostgresWrapper) GetUserLoginValues(userEmail string) (string, string, error) {
	queryString := fmt.Sprintf(
	`SELECT user_id::VARCHAR, pass_hash FROM public.user WHERE email = '%v'`,
	userEmail)
	slog.Debug("Query string:", "query-string", queryString)

	var userID, password string
	err := p.postgresClient.QueryRow(context.Background(), queryString).Scan(&userID, &password)
	if err != nil {
		return "", "", err
	}

	return userID, password, nil
}

func (p *PostgresWrapper) GetSpotifyRefreshToken(userID string) (string, error) {
	queryString := fmt.Sprintf(
	`SELECT spotify_refresh_token FROM public.user WHERE user_id = '%v'`,
	userID)
	var spotifyRefreshToken string
	err := p.postgresClient.QueryRow(context.Background(), queryString).Scan(&spotifyRefreshToken)
	if err != nil {
		return "", err
	}

	return spotifyRefreshToken, nil
}

func (p *PostgresWrapper) SetUserSession(userID string, sessionID int) error {
	queryString := fmt.Sprintf(
	`UPDATE public.user
	SET session_id = %v
	WHERE user_id = %v;`, sessionID, userID)

	commandTag, err := p.postgresClient.Exec(context.Background(), queryString)

	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("User %v not found to update", userID)
	}
	return nil
}

func (p *PostgresWrapper) SetSpotifyAccessToken(userID string, AccessToken string) error {
	queryString := fmt.Sprintf(
	`UPDATE public.user
	SET spotify_access_token = '%v'
	WHERE user_id = %v;`, AccessToken, userID)

	commandTag, err := p.postgresClient.Exec(context.Background(), queryString)

	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("User %v not found to update", userID)
	}
	return nil
}

func (p *PostgresWrapper) SetSpotifyRefreshToken(userID string, RefreshToken string) error {
	queryString := fmt.Sprintf(
	`UPDATE public.user
	SET spotify_refresh_token = '%v'
	WHERE user_id = %v;`, RefreshToken, userID)

	commandTag, err := p.postgresClient.Exec(context.Background(), queryString)

	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("User %v not found to update", userID)
	}
	return nil
}

func (p *PostgresWrapper) CheckIfEmailExists(email string) (bool, error) {
	queryString := fmt.Sprintf("SELECT exists (SELECT 1 FROM public.user WHERE email = '%v' LIMIT 1);", email)
	var emailExists bool
	err := p.postgresClient.QueryRow(context.Background(), queryString).Scan(&emailExists)
	if err != nil {
		return true, err
	}

	return emailExists, nil
}

func (p *PostgresWrapper) AddUserToDatabase(newUser model.NewUser, passwordHash, account_level, voter_level string, bonusVotes int) (string, error) {
	queryString := fmt.Sprintf(
	`INSERT INTO public.user(first_name, last_name, email, pass_hash, account_level, voter_level, bonus_votes)
	VALUES ('%v', '%v', '%v', '%v', '%v', '%v', %v')
	RETURNING user_id::VARCHAR;`,
	newUser.FirstName, newUser.LastName, newUser.Email, passwordHash, account_level, voter_level, bonusVotes)

	var userID string
	err := p.postgresClient.QueryRow(context.Background(), queryString).Scan(&userID)
	if err != nil {
		return "", err
	}

	return userID, nil
}