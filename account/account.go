package account

import (
	"fmt"
	"os"

	"golang.org/x/exp/slog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/campbelljlowman/fazool-api/graph/model"
)

type AccountService interface {
	// GetAccountByEmail(accountEmail string) (*model.Account, error)
	// GetAccountByID(accountID string) (*model.Account, error)
	// GetAccountLoginValues(accountEmail string) (string, string, error)
	// GetSpotifyRefreshToken(accountID string) (string, error)
	// GetAccountLevel(accountID string) (string, error)
	// GetVoterValues(accountID string) (string, int, error)

	// SetAccountSession(accountID string, sessionID int) error
	// SetSpotifyAccessToken(accountID, AccessToken string) error
	// SetSpotifyRefreshToken(accountID, RefreshToken string) error
	// SubtractBonusVotes(accountID string, bonusVotes int) error

	// CheckIfEmailExists(email string) (bool, error)

	AddAccountToDatabase(newAccount model.NewAccount, passwordHash, account_level, voter_level string, bonusVotes int) string 
}

type account struct {
	gorm.Model
	FirstName 				string
	LastName 				string
	Email 					string
	PasswordHash 			string
	AccountLevel			string
	VoterLevel 				string
	BonusVotes 				int
	ActiveSession			int
	SpotifyAccessToken 		string
	SpotifyRefreshToken		string
}

// TODO: Use type alias?
type AccountGorm struct {
	gorm 	*gorm.DB
}

func NewAccountGorm() *AccountGorm {
	postgresURL := os.Getenv("POSTRGRES_URL")

    gormDB, err := gorm.Open(postgres.Open(postgresURL), &gorm.Config{})
	if err != nil {
		slog.Error("Unable to connect to database", err)
		os.Exit(1)
	}

	gormDB.AutoMigrate(&account{})

	accountGorm := AccountGorm{gorm: gormDB}
	return &accountGorm
}

// func (a *AccountGorm) GetAccountByEmail(accountEmail string) (*model.Account, error)

// func (a *AccountGorm) GetAccountByID(accountID string) (*model.Account, error)
// func (a *AccountGorm) GetAccountLoginValues(accountEmail string) (string, string, error)
// func (a *AccountGorm) GetSpotifyRefreshToken(accountID string) (string, error)
// func (a *AccountGorm) GetAccountLevel(accountID string) (string, error)
// func (a *AccountGorm) GetVoterValues(accountID string) (string, int, error)

// func (a *AccountGorm) SetAccountSession(accountID string, sessionID int) error
// func (a *AccountGorm) SetSpotifyAccessToken(accountID, AccessToken string) error
// func (a *AccountGorm) SetSpotifyRefreshToken(accountID, RefreshToken string) error
// func (a *AccountGorm) SubtractBonusVotes(accountID string, bonusVotes int) error

// func (a *AccountGorm) CheckIfEmailExists(email string) (bool, error)

// TODO, password is passed to this function on newAccount struct, this is bad
func (a *AccountGorm) AddAccountToDatabase(newAccount model.NewAccount, passwordHash, accountLevel, voterLevel string, bonusVotes int) string {
	accountToAdd := &account{
		FirstName: 		newAccount.FirstName,
		LastName: 		newAccount.LastName,
		Email: 			newAccount.Email,
		PasswordHash: 	passwordHash,
		AccountLevel: 	accountLevel,
		VoterLevel: 	voterLevel,
		BonusVotes: 	bonusVotes,
	}
	
	a.gorm.Create(accountToAdd)

	return fmt.Sprintf("%d", accountToAdd.ID)
}