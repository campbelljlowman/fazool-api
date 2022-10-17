package database

import (
	"fmt"
	"context"
	"errors"

	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/jackc/pgx/v4/pgxpool"
)

func GetUserByEmail (db *pgxpool.Pool, userEmail string) (*model.User, error) {
	getUserQueryString := fmt.Sprintf(`
	SELECT user_id, first_name, last_name, email, coalesce(session_id,0) as session_id FROM public.user WHERE email = '%v'`,
	userEmail)
	var userID, sessionID int
	var firstName, lastName, email string
	err := db.QueryRow(context.Background(), getUserQueryString).Scan(&userID, &firstName, &lastName, &email, &sessionID)
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

func GetUserByID (db *pgxpool.Pool, ID int) (*model.User, error) {
	getUserQueryString := fmt.Sprintf(`
	SELECT user_id, first_name, last_name, email, coalesce(session_id,0) as session_id FROM public.user WHERE user_id = '%v'`,
	ID)
	var userID, sessionID int
	var firstName, lastName, email string
	err := db.QueryRow(context.Background(), getUserQueryString).Scan(&userID, &firstName, &lastName, &email, &sessionID)
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

func GetUserLoginValues (db *pgxpool.Pool, userEmail string) (int, int, string, error) {
	getUserQueryString := fmt.Sprintf(`
	SELECT user_id, auth_level, pass_hash FROM public.user WHERE email = '%v'`,
	userEmail)
	var password string
	var userID, authLevel int
	err := db.QueryRow(context.Background(), getUserQueryString).Scan(&userID, &authLevel, &password)
	if err != nil {
		println("Error getting user from database")
		println(err.Error())
		return 0, 0, "", errors.New("Invalid Login Credentials!")
	}

	return userID, authLevel, password, nil
}

func SetSpotifyAccessToken(db *pgxpool.Pool, userID int, AccessToken string) error {
	queryString := fmt.Sprintf(`
	UPDATE public.user
	SET spotify_access_token = '%v'
	WHERE user_id = %v;`, AccessToken, userID)

	commandTag, err := db.Exec(context.Background(), queryString)

	if err != nil {
		fmt.Printf("Error: %v", err)
		return errors.New("Error adding spotify access token to database")
	}
	if commandTag.RowsAffected() != 1 {
		return errors.New("No user found to update")
	}
	return nil
}

func SetSpotifyRefreshToken(db *pgxpool.Pool, userID int, RefreshToken string) error {
	queryString := fmt.Sprintf(`
	UPDATE public.user
	SET spotify_refresh_token = '%v'
	WHERE user_id = %v;`, RefreshToken, userID)

	commandTag, err := db.Exec(context.Background(), queryString)

	if err != nil {
		fmt.Printf("Error: %v", err)
		return errors.New("Error adding spotify credentials to database")
	}
	if commandTag.RowsAffected() != 1 {
		return errors.New("No user found to update")
	}
	return nil
}

func GetSpotifyRefreshToken(db *pgxpool.Pool, userID int) (string, error) {
	getUserQueryString := fmt.Sprintf(`
	SELECT spotify_refresh_token FROM public.user WHERE user_id = '%v'`,
	userID)
	var spotifyRefreshToken string
	err := db.QueryRow(context.Background(), getUserQueryString).Scan(&spotifyRefreshToken)
	if err != nil {
		println("Error getting spotify refresh token from database")
		println(err.Error())
		return "", errors.New("Spotify refresh token not found!")
	}

	return spotifyRefreshToken, nil
}