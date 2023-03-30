package account

import (
	"os"

	"golang.org/x/exp/slog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/campbelljlowman/fazool-api/graph/model"
)

type AccountService interface {
	CreateAccount(newAccount model.NewAccount, passwordHash, account_level, voter_level string, bonusVotes int) int 

	GetAccountFromEmail(accountEmail string) *model.Account
	GetAccountFromID(accountID int) *model.Account
	GetAccountIDAndPassHash(accountEmail string) (int, string)
	GetSpotifyRefreshToken(accountID int) string
	GetAccountLevel(accountID int) string
	GetVoterLevelAndBonusVotes(accountID int) (string, int)
	CheckIfEmailHasAccount(email string) bool

	SetAccountActiveSession(accountID int, sessionID int)
	SetSpotifyRefreshToken(accountID int, refreshToken string)
	SubtractBonusVotes(accountID, bonusVotes int)

	// TODO: Delete account
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

type AccountServiceGorm struct {
	gorm 	*gorm.DB
}

func NewAccountServiceGormImpl() *AccountServiceGorm {
	postgresURL := os.Getenv("POSTRGRES_URL")

    gormDB, err := gorm.Open(postgres.Open(postgresURL), &gorm.Config{})
	if err != nil {
		slog.Error("Unable to connect to database", err)
		os.Exit(1)
	}

	gormDB.AutoMigrate(&account{})

	gormDB.Exec("UPDATE public.accounts SET active_session = 0;")

	accountGorm := AccountServiceGorm{gorm: gormDB}
	return &accountGorm
}

func (a *AccountServiceGorm) GetAccountFromEmail(accountEmail string) *model.Account {
	var fullAccount account
	a.gorm.Where("email = ?", accountEmail).First(&fullAccount)

	return transformAccountType(fullAccount)
}

func (a *AccountServiceGorm) GetAccountFromID(accountID int) *model.Account {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	return transformAccountType(fullAccount)
}

func transformAccountType(fullAccount account) *model.Account {
	accountToReturn := &model.Account{
		ID: int(fullAccount.ID),
		FirstName: &fullAccount.FirstName,
		LastName: &fullAccount.LastName,
		Email: &fullAccount.Email,
		ActiveSession: &fullAccount.ActiveSession,
	}
	return accountToReturn
}

func (a *AccountServiceGorm) GetAccountIDAndPassHash(accountEmail string) (int, string) {
	var fullAccount account
	a.gorm.Where("email = ?", accountEmail).First(&fullAccount)

	return int(fullAccount.ID), fullAccount.PasswordHash
}

func (a *AccountServiceGorm) GetSpotifyRefreshToken(accountID int) string {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	return fullAccount.SpotifyRefreshToken
}

func (a *AccountServiceGorm) GetAccountLevel(accountID int) string {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	return fullAccount.AccountLevel
}

func (a *AccountServiceGorm) GetVoterLevelAndBonusVotes(accountID int) (string, int) {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	return fullAccount.VoterLevel, fullAccount.BonusVotes
}

func (a *AccountServiceGorm) SetAccountActiveSession(accountID int, sessionID int) {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	fullAccount.ActiveSession = sessionID
	a.gorm.Save(&fullAccount)
}

func (a *AccountServiceGorm) SetSpotifyRefreshToken(accountID int, refreshToken string) {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	fullAccount.SpotifyRefreshToken = refreshToken
	a.gorm.Save(&fullAccount)
}

func (a *AccountServiceGorm) SubtractBonusVotes(accountID, bonusVotes int) {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	fullAccount.BonusVotes -= bonusVotes
	a.gorm.Save(&fullAccount)
}

func (a *AccountServiceGorm) CheckIfEmailHasAccount(email string) bool {
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
func (a *AccountServiceGorm) CreateAccount(newAccount model.NewAccount, passwordHash, accountLevel, voterLevel string, bonusVotes int) int {
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

	return int(accountToAdd.ID)
}