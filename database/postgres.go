package database

import (
	"os"
	"fmt"
	"errors"
	"context"


	"github.com/campbelljlowman/fazool-api/graph/model"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresWrapper struct {
	postgresClient *pgxpool.Pool
}

func NewPostgresClient() *PostgresWrapper {
	databaseURL := os.Getenv("POSTRGRES_URL")

	dbPool, err := pgxpool.Connect(context.Background(), databaseURL)
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

	pg := PostgresWrapper{dbPool}
	return &pg
}

func (p *PostgresWrapper) GetUserByEmail(userEmail string) (*model.User, error) {
	getUserQueryString := fmt.Sprintf(`
	SELECT user_id, first_name, last_name, email, coalesce(session_id,0) as session_id FROM public.user WHERE email = '%v'`,
	userEmail)
	var userID, sessionID int
	var firstName, lastName, email string
	err := p.postgresClient.QueryRow(context.Background(), getUserQueryString).Scan(&userID, &firstName, &lastName, &email, &sessionID)
	if err != nil {
		println("Error getting user from database")
		println(err.Error())
		return nil, errors.New("Invalid Login Credentials!")
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

func (p *PostgresWrapper) GetUserByID(ID int) (*model.User, error) {
	getUserQueryString := fmt.Sprintf(`
	SELECT user_id, first_name, last_name, email, coalesce(session_id,0) as session_id FROM public.user WHERE user_id = '%v'`,
	ID)
	var userID, sessionID int
	var firstName, lastName, email string
	err := p.postgresClient.QueryRow(context.Background(), getUserQueryString).Scan(&userID, &firstName, &lastName, &email, &sessionID)
	if err != nil {
		println("Error getting user from database")
		println(err.Error())
		return nil, errors.New("Invalid Login Credentials!")
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

func (p *PostgresWrapper) GetUserLoginValues(userEmail string) (int, int, string, error) {
	getUserQueryString := fmt.Sprintf(`
	SELECT user_id, auth_level, pass_hash FROM public.user WHERE email = '%v'`,
	userEmail)
	var password string
	var userID, authLevel int
	err := p.postgresClient.QueryRow(context.Background(), getUserQueryString).Scan(&userID, &authLevel, &password)
	if err != nil {
		println("Error getting user from database")
		println(err.Error())
		return 0, 0, "", errors.New("Invalid Login Credentials!")
	}

	return userID, authLevel, password, nil
}

func (p *PostgresWrapper) GetSpotifyRefreshToken(userID int) (string, error) {
	getUserQueryString := fmt.Sprintf(`
	SELECT spotify_refresh_token FROM public.user WHERE user_id = '%v'`,
	userID)
	var spotifyRefreshToken string
	err := p.postgresClient.QueryRow(context.Background(), getUserQueryString).Scan(&spotifyRefreshToken)
	if err != nil {
		println("Error getting spotify refresh token from database")
		println(err.Error())
		return "", errors.New("Spotify refresh token not found!")
	}

	return spotifyRefreshToken, nil
}

func (p *PostgresWrapper) SetUserSession(userID int, sessionID int) error {
	queryString := fmt.Sprintf(`
	UPDATE public.user
	SET session_id = %v
	WHERE user_id = %v;`, sessionID, userID)

	commandTag, err := p.postgresClient.Exec(context.Background(), queryString)

	if err != nil {
		return errors.New("Error adding new session to database")
	}
	if commandTag.RowsAffected() != 1 {
		return errors.New("No user found to update")
	}
	return nil
}

func (p *PostgresWrapper) SetSpotifyAccessToken(userID int, AccessToken string) error {
	queryString := fmt.Sprintf(`
	UPDATE public.user
	SET spotify_access_token = '%v'
	WHERE user_id = %v;`, AccessToken, userID)

	commandTag, err := p.postgresClient.Exec(context.Background(), queryString)

	if err != nil {
		fmt.Printf("Error: %v", err)
		return errors.New("Error adding spotify access token to database")
	}
	if commandTag.RowsAffected() != 1 {
		return errors.New("No user found to update")
	}
	return nil
}

func (p *PostgresWrapper) SetSpotifyRefreshToken(userID int, RefreshToken string) error {
	queryString := fmt.Sprintf(`
	UPDATE public.user
	SET spotify_refresh_token = '%v'
	WHERE user_id = %v;`, RefreshToken, userID)

	commandTag, err := p.postgresClient.Exec(context.Background(), queryString)

	if err != nil {
		fmt.Printf("Error: %v", err)
		return errors.New("Error adding spotify credentials to database")
	}
	if commandTag.RowsAffected() != 1 {
		return errors.New("No user found to update")
	}
	return nil
}

func (p *PostgresWrapper) CheckIfEmailExists(email string) bool {
	checkEmailQueryString := fmt.Sprintf("SELECT exists (SELECT 1 FROM public.user WHERE email = '%v' LIMIT 1);", email)
	var emailExists bool
	p.postgresClient.QueryRow(context.Background(), checkEmailQueryString).Scan(&emailExists)

	return emailExists
}

func (p *PostgresWrapper) AddUserToDatabase(newUser model.NewUser, passwordHash string, authLevel int) (int, error) {
	newUserQueryString := fmt.Sprintf(`
	INSERT INTO public.user(first_name, last_name, email, pass_hash, auth_level)
	VALUES ('%v', '%v', '%v', '%v', '%v')
	RETURNING user_id;`,
	newUser.FirstName, newUser.LastName, newUser.Email, passwordHash, authLevel)

	var userID int
	err := p.postgresClient.QueryRow(context.Background(), newUserQueryString).Scan(&userID)

	return userID, err
}