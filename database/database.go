package database

import (
	"github.com/campbelljlowman/fazool-api/graph/model"
)

type Database interface {
	GetUserByEmail(userEmail string) (*model.User, error)
	GetUserByID(userID string) (*model.User, error)
	GetUserLoginValues(userEmail string) (string, string, error)
	GetSpotifyRefreshToken(userID string) (string, error)
	GetAccountLevel(userID string) (string, error)
	GetVoterValues(userID string) (string, int, error)

	SetUserSession(userID string, sessionID int) error
	SetSpotifyAccessToken(userID, AccessToken string) error
	SetSpotifyRefreshToken(userID, RefreshToken string) error
	SubtractBonusVotes(userID string, bonusVotes int) error

	CheckIfEmailExists(email string) (bool, error)

	AddUserToDatabase(newUser model.NewUser, passwordHash, account_level, voter_level string, bonusVotes int) (string, error) 

	CloseConnection()
}