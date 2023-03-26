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
	GetAccountFromEmail(accountEmail string) *model.Account
	GetAccountFromID(accountID string) *model.Account
	GetAccountIDAndPassHash(accountEmail string) (string, string)
	GetSpotifyRefreshToken(accountID string) string
	GetAccountLevel(accountID string) string
	GetVoterLevelAndBonusVotes(accountID string) (string, int)

	SetAccountActiveSession(accountID string, sessionID int)
	SetSpotifyAccessToken(accountID, accessToken string)
	SetSpotifyRefreshToken(accountID, refreshToken string)
	SubtractBonusVotes(accountID string, bonusVotes int)

	CheckIfEmailHasAccount(email string) bool

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

func (a *AccountGorm) GetAccountFromEmail(accountEmail string) *model.Account {
	var fullAccount account
	a.gorm.Where("email = ?", accountEmail).First(&fullAccount)

	return transformAccountType(fullAccount)
}

func (a *AccountGorm) GetAccountFromID(accountID string) *model.Account {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	return transformAccountType(fullAccount)
}

func transformAccountType(fullAccount account) *model.Account {
	accountToReturn := &model.Account{
		ID: fmt.Sprintf("%d", fullAccount.ID),
		FirstName: &fullAccount.FirstName,
		LastName: &fullAccount.LastName,
		Email: &fullAccount.Email,
		SessionID: &fullAccount.ActiveSession,
	}
	return accountToReturn
}

func (a *AccountGorm) GetAccountIDAndPassHash(accountEmail string) (string, string) {
	var fullAccount account
	a.gorm.Where("email = ?", accountEmail).First(&fullAccount)

	return fmt.Sprintf("%d", fullAccount.ID), fullAccount.PasswordHash
}

func (a *AccountGorm) GetSpotifyRefreshToken(accountID string) string {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	return fullAccount.SpotifyRefreshToken
}

func (a *AccountGorm) GetAccountLevel(accountID string) string {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	return fullAccount.AccountLevel
}

func (a *AccountGorm) GetVoterLevelAndBonusVotes(accountID string) (string, int) {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	return fullAccount.VoterLevel, fullAccount.BonusVotes
}

func (a *AccountGorm) SetAccountActiveSession(accountID string, sessionID int) {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	fullAccount.ActiveSession = sessionID
	a.gorm.Save(&fullAccount)
}

func (a *AccountGorm) SetSpotifyAccessToken(accountID, accessToken string) {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	fullAccount.SpotifyAccessToken = accessToken
	a.gorm.Save(&fullAccount)
}

func (a *AccountGorm) SetSpotifyRefreshToken(accountID, refreshToken string) {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	fullAccount.SpotifyRefreshToken = refreshToken
	a.gorm.Save(&fullAccount)
}

func (a *AccountGorm) SubtractBonusVotes(accountID string, bonusVotes int) {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	fullAccount.BonusVotes -= bonusVotes
	a.gorm.Save(&fullAccount)
}

func (a *AccountGorm) CheckIfEmailHasAccount(email string) bool {
	var fullAccount account
	err := a.gorm.Where("email = ?", email).First(&fullAccount).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false
		}
		slog.Warn("Error checking if email has an account", "error", err)
	}
	return true
}

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