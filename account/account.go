//go:generate mockgen -destination=../mocks/mock_account.go -package=mocks . AccountService

package account

import (
	"os"
	"strings"

	"golang.org/x/exp/slog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/campbelljlowman/fazool-api/graph/model"
)

type AccountService interface {
	CreateAccount(firstName, lastName, email, phoneNumber, passwordHash string, accountType model.AccountType, bonusVotes int, streamingService model.StreamingService) int

	GetAccountFromEmail(accountEmail string) *model.Account
	GetAccountFromID(accountID int) *model.Account
	GetAccountIDAndPassHash(accountEmail string) (int, string)
	GetSpotifyRefreshToken(accountID int) string
	GetAccountType(accountID int) model.AccountType
	GetSuperVoterSessionsAndBonusVotes(accountID int) (int, int)
	GetAccountActiveSession(accountID int) int
	GetAccountFazoolTokens(accountID int) int
	CheckIfEmailHasAccount(email string) bool

	SetAccountActiveSession(accountID int, sessionID int)
	SetSpotifyRefreshToken(accountID int, refreshToken string)
	SetAccountType(accountID int, accountType model.AccountType) *model.Account
	SetSuperVoterSession(accountID, sessionID, fazoolTokens int) *model.Account
	AddBonusVotes(accountID, bonusVotes, fazoolTokens int) *model.Account
	AddFazoolTokens(accountID, fazoolTokens int) *model.Account

	RemoveSuperVoter(accountID int, sessionID int)
	SubtractBonusVotes(accountID, bonusVotes int)

	DeleteAccount(accountID int)
}

type account struct {
	gorm.Model
	FirstName 				string
	LastName 				string
	Email 					string
	PhoneNumber 			string
	PasswordHash 			string
	AccountType				model.AccountType
	SuperVoterSession		int
	BonusVotes 				int
	FazoolTokens			int
	ActiveSession			int
	SpotifyRefreshToken		string
	StreamingService		model.StreamingService
}

type AccountServiceGorm struct {
	gorm 	*gorm.DB
}

func NewAccountServiceGormImpl() *AccountServiceGorm {
	postgresURL := os.Getenv("POSTGRES_URL")
	slog.Debug("Databse URL", "url", postgresURL)

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

func (a *AccountServiceGorm) CreateAccount(firstName, lastName, email, phoneNumber, passwordHash string, accountType model.AccountType, bonusVotes int, streamingService model.StreamingService) int {
	accountToAdd := &account{
		FirstName: 			firstName,
		LastName: 			lastName,
		Email: 				strings.ToLower(email),
		PhoneNumber: 		phoneNumber,
		PasswordHash: 		passwordHash,
		AccountType: 		accountType,
		SuperVoterSession: 	0,
		BonusVotes: 		bonusVotes,
		StreamingService: 	streamingService,
	}
	
	a.gorm.Create(accountToAdd)

	return int(accountToAdd.ID)
}

func (a *AccountServiceGorm) GetAccountFromEmail(accountEmail string) *model.Account {
	var fullAccount account
	a.gorm.Where("email = ?", strings.ToLower(accountEmail)).First(&fullAccount)

	return transformAccountType(fullAccount)
}

func (a *AccountServiceGorm) GetAccountFromID(accountID int) *model.Account {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	return transformAccountType(fullAccount)
}

func (a *AccountServiceGorm) GetAccountIDAndPassHash(accountEmail string) (int, string) {
	var fullAccount account
	a.gorm.Where("email = ?", strings.ToLower(accountEmail)).First(&fullAccount)

	return int(fullAccount.ID), fullAccount.PasswordHash
}

func (a *AccountServiceGorm) GetSpotifyRefreshToken(accountID int) string {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	return fullAccount.SpotifyRefreshToken
}

func (a *AccountServiceGorm) GetAccountType(accountID int) model.AccountType {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	return fullAccount.AccountType
}

func (a *AccountServiceGorm) GetSuperVoterSessionsAndBonusVotes(accountID int) (int, int) {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	return fullAccount.SuperVoterSession, fullAccount.BonusVotes
}

func (a *AccountServiceGorm) GetAccountActiveSession(accountID int) int {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	return fullAccount.ActiveSession
}

func (a *AccountServiceGorm) GetAccountFazoolTokens(accountID int) int {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	return fullAccount.FazoolTokens
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
	fullAccount.StreamingService = model.StreamingServiceSpotify
	a.gorm.Save(&fullAccount)
}

func (a *AccountServiceGorm) SetAccountType(accountID int, accountType model.AccountType) *model.Account {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	fullAccount.AccountType = accountType
	a.gorm.Save(&fullAccount)

	return transformAccountType(fullAccount)
}

func (a *AccountServiceGorm) SetSuperVoterSession(accountID, sessionID, fazoolTokens int) *model.Account {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	fullAccount.SuperVoterSession = sessionID
	fullAccount.FazoolTokens -= fazoolTokens
	a.gorm.Save(&fullAccount)

	return transformAccountType(fullAccount)
}

func (a *AccountServiceGorm) AddBonusVotes(accountID, bonusVotes, fazoolTokens int) *model.Account {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	fullAccount.BonusVotes += bonusVotes
	fullAccount.FazoolTokens -= fazoolTokens
	a.gorm.Save(&fullAccount)

	return transformAccountType(fullAccount)
}

func (a *AccountServiceGorm) AddFazoolTokens(accountID, fazoolTokens int) *model.Account {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	fullAccount.FazoolTokens += fazoolTokens
	a.gorm.Save(&fullAccount)

	return transformAccountType(fullAccount)
}

func (a *AccountServiceGorm) SubtractBonusVotes(accountID, bonusVotes int) {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	fullAccount.BonusVotes -= bonusVotes
	a.gorm.Save(&fullAccount)
}

func (a *AccountServiceGorm) RemoveSuperVoter(accountID int, sessionID int) {
	var fullAccount account
	a.gorm.First(&fullAccount, accountID)

	fullAccount.SuperVoterSession = 0
	a.gorm.Save(&fullAccount)
}

func (a *AccountServiceGorm) CheckIfEmailHasAccount(email string) bool {
	var fullAccount account
	err := a.gorm.Where("email = ?", strings.ToLower(email)).First(&fullAccount).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			slog.Info("No account found with email", "email", strings.ToLower(email))
			return false
		}
		slog.Warn("Error checking if email has an account", "error", err)
	}
	return true
}

func (a *AccountServiceGorm) DeleteAccount(accountID int) {
	var fullAccount account
	a.gorm.Delete(&fullAccount, accountID)
}

func transformAccountType(fullAccount account) *model.Account {
	accountToReturn := &model.Account{
		ID: int(fullAccount.ID),
		FirstName: fullAccount.FirstName,
		LastName: fullAccount.LastName,
		Email: &fullAccount.Email,
		ActiveSession: &fullAccount.ActiveSession,
		StreamingService: &fullAccount.StreamingService,
		FazoolTokens: &fullAccount.FazoolTokens,
	}
	return accountToReturn
}