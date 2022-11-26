package database

import (
	"github.com/campbelljlowman/fazool-api/graph/model"
)

type Database interface {
	GetUserByEmail(userEmail string) (*model.User, error)
	GetUserByID(ID int) (*model.User, error)
	GetUserLoginValues(userEmail string) (int, int, string, error)
	GetSpotifyRefreshToken(userID int) (string, error)

	SetUserSession(userID int, sessionID int) error
	SetSpotifyAccessToken(userID int, AccessToken string) error
	SetSpotifyRefreshToken(userID int, RefreshToken string) error

	CheckIfEmailExists(email string) (bool, error)

	AddUserToDatabase(newUser model.NewUser, passwordHash string, authLevel int) (int, error) 
}