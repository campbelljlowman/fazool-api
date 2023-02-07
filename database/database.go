package database

import (
	"github.com/campbelljlowman/fazool-api/graph/model"
)

type Database interface {
	GetAccountByEmail(accountEmail string) (*model.Account, error)
	GetAccountByID(accountID string) (*model.Account, error)
	GetAccountLoginValues(accountEmail string) (string, string, error)
	GetSpotifyRefreshToken(accountID string) (string, error)
	GetAccountLevel(accountID string) (string, error)
	GetVoterValues(accountID string) (string, int, error)

	SetAccountSession(accountID string, sessionID int) error
	SetSpotifyAccessToken(accountID, AccessToken string) error
	SetSpotifyRefreshToken(accountID, RefreshToken string) error
	SubtractBonusVotes(accountID string, bonusVotes int) error

	CheckIfEmailExists(email string) (bool, error)

	AddAccountToDatabase(newAccount model.NewAccount, passwordHash, account_level, voter_level string, bonusVotes int) (string, error) 

	CloseConnection()
}