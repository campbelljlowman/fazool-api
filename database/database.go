package database

import (
	"github.com/campbelljlowman/fazool-api/graph/model"
)

type Database interface {
	GetUserByEmail(userEmail string) (*model.User, error)
	GetUserByID(userID string) (*model.User, error)
	GetUserLoginValues(userEmail string) (string, string, error)
	GetSpotifyRefreshToken(userID string) (string, error)

	SetUserSession(userID string, sessionID int) error
	SetSpotifyAccessToken(userID string, AccessToken string) error
	SetSpotifyRefreshToken(userID string, RefreshToken string) error

	CheckIfEmailExists(email string) (bool, error)

	AddUserToDatabase(newUser model.NewUser, passwordHash, account_level string, bonusVotes int) (string, error) 
}