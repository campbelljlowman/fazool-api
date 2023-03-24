package database

import (
	"context"
	"fmt"
	"os"

	"github.com/campbelljlowman/fazool-api/graph/model"
	"golang.org/x/exp/slog"

	"github.com/jackc/pgx/v4/pgxpool"
)


type PostgresWrapper struct {
	PostgresClient *pgxpool.Pool
}

func NewPostgresClient() *PostgresWrapper {
	databaseURL := os.Getenv("POSTRGRES_URL")

	dbPool, err := pgxpool.Connect(context.Background(), databaseURL)
	if err != nil {
		slog.Error("Unable to connect to database", err)
		os.Exit(1)
	}

	queryString := fmt.Sprintf(
	`CREATE TABLE IF NOT EXISTS public.old_accounts
	(
		account_id       		int GENERATED ALWAYS AS IDENTITY primary key,
		first_name    			varchar(100) not null,
		last_name     			varchar(100) not null,
		email 		 			varchar(100) not null,
		pass_hash 	  			varchar(100) not null,
		account_level 	  		varchar(20) not null,
		voter_level				varchar(20) not null,
		bonus_votes				int not null,
		session_id 	  			int,
		spotify_access_token 	varchar(200),
		spotify_refresh_token 	varchar(150)
	);

	UPDATE public.old_accounts
	SET session_id = 0;
	`)

	_, err = dbPool.Exec(context.Background(), queryString)
	if err != nil {
		slog.Error("Error initializing database", err)
		os.Exit(1)
	}

	pg := PostgresWrapper{dbPool}
	return &pg
}

func (p *PostgresWrapper) GetAccountByEmail(accountEmail string) (*model.Account, error) {
	queryString := fmt.Sprintf(
	`SELECT account_id::VARCHAR, first_name, last_name, email, coalesce(session_id,0) as session_id FROM public.old_accounts WHERE email = '%v'`,
	accountEmail)
	var sessionID int
	var accountID, firstName, lastName, email string
	err := p.PostgresClient.QueryRow(context.Background(), queryString).Scan(&accountID, &firstName, &lastName, &email, &sessionID)
	if err != nil {
		return nil, err
	}

	account := &model.Account{
		ID:        accountID,
		FirstName: &firstName,
		LastName:  &lastName,
		Email:     &email,
		SessionID: &sessionID,
	}

	return account, nil
}

func (p *PostgresWrapper) GetAccountByID(accountID string) (*model.Account, error) {
	queryString := fmt.Sprintf(
	`SELECT first_name, last_name, email, coalesce(session_id,0) as session_id FROM public.old_accounts WHERE account_id = '%v'`,
	accountID)
	slog.Debug("Query string:", "query-string", queryString)

	var sessionID int
	var firstName, lastName, email string
	err := p.PostgresClient.QueryRow(context.Background(), queryString).Scan(&firstName, &lastName, &email, &sessionID)
	if err != nil {
		return nil, err
	}

	account := &model.Account{
		ID:        accountID,
		FirstName: &firstName,
		LastName:  &lastName,
		Email:     &email,
		SessionID: &sessionID,
	}

	return account, nil
}

func (p *PostgresWrapper) GetAccountLoginValues(accountEmail string) (string, string, error) {
	queryString := fmt.Sprintf(
	`SELECT account_id::VARCHAR, pass_hash FROM public.old_accounts WHERE email = '%v'`,
	accountEmail)
	slog.Debug("Query string:", "query-string", queryString)

	var accountID, password string
	err := p.PostgresClient.QueryRow(context.Background(), queryString).Scan(&accountID, &password)
	if err != nil {
		return "", "", err
	}

	return accountID, password, nil
}

func (p *PostgresWrapper) GetSpotifyRefreshToken(accountID string) (string, error) {
	queryString := fmt.Sprintf(
	`SELECT spotify_refresh_token FROM public.old_accounts WHERE account_id = '%v'`,
	accountID)
	var spotifyRefreshToken string
	err := p.PostgresClient.QueryRow(context.Background(), queryString).Scan(&spotifyRefreshToken)
	if err != nil {
		return "", err
	}

	return spotifyRefreshToken, nil
}

func (p *PostgresWrapper) GetAccountLevel(accountID string) (string, error){
	queryString := fmt.Sprintf(
		`SELECT account_level FROM public.old_accounts WHERE account_id = '%v'`,
	accountID)
	var accountLevel string
	err := p.PostgresClient.QueryRow(context.Background(), queryString).Scan(&accountLevel)
	if err != nil {
		return "", err
	}

	return accountLevel, nil
}

func (p *PostgresWrapper) GetVoterValues(accountID string) (string, int, error){
	queryString := fmt.Sprintf(
		`SELECT voter_level, bonus_votes FROM public.old_accounts WHERE account_id = '%v'`,
	accountID)
	var voterLevel string
	var bonusVotes int
	err := p.PostgresClient.QueryRow(context.Background(), queryString).Scan(&voterLevel, &bonusVotes)
	if err != nil {
		return "", 0, err
	}

	return voterLevel, bonusVotes, nil	
}

func (p *PostgresWrapper) SetAccountSession(accountID string, sessionID int) error {
	queryString := fmt.Sprintf(
	`UPDATE public.old_accounts
	SET session_id = %v
	WHERE account_id = %v;`, sessionID, accountID)

	commandTag, err := p.PostgresClient.Exec(context.Background(), queryString)

	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("Account %v not found to update", accountID)
	}
	return nil
}

func (p *PostgresWrapper) SetSpotifyAccessToken(accountID, AccessToken string) error {
	queryString := fmt.Sprintf(
	`UPDATE public.old_accounts
	SET spotify_access_token = '%v'
	WHERE account_id = %v;`, AccessToken, accountID)

	commandTag, err := p.PostgresClient.Exec(context.Background(), queryString)

	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("Account %v not found to update", accountID)
	}
	return nil
}

func (p *PostgresWrapper) SetSpotifyRefreshToken(accountID, RefreshToken string) error {
	queryString := fmt.Sprintf(
	`UPDATE public.old_accounts
	SET spotify_refresh_token = '%v'
	WHERE account_id = %v;`, RefreshToken, accountID)

	commandTag, err := p.PostgresClient.Exec(context.Background(), queryString)

	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("Account %v not found to update", accountID)
	}
	return nil
}

func (p *PostgresWrapper) SubtractBonusVotes(accountID string, bonusVotes int) error {
	queryString := fmt.Sprintf(
		`UPDATE public.old_accounts
		SET bonus_votes = bonus_votes - '%v'
		WHERE account_id = %v;`, bonusVotes, accountID)
	
		commandTag, err := p.PostgresClient.Exec(context.Background(), queryString)
	
		if err != nil {
			return err
		}
		if commandTag.RowsAffected() != 1 {
			return fmt.Errorf("Account %v not found to update", accountID)
		}
		return nil
}


func (p *PostgresWrapper) CheckIfEmailExists(email string) (bool, error) {
	queryString := fmt.Sprintf("SELECT exists (SELECT 1 FROM public.old_accounts WHERE email = '%v' LIMIT 1);", email)
	var emailExists bool
	err := p.PostgresClient.QueryRow(context.Background(), queryString).Scan(&emailExists)
	if err != nil {
		return true, err
	}

	return emailExists, nil
}

func (p *PostgresWrapper) AddAccountToDatabase(newAccount model.NewAccount, passwordHash, account_level, voter_level string, bonusVotes int) (string, error) {
	queryString := fmt.Sprintf(
	`INSERT INTO public.old_accounts(first_name, last_name, email, pass_hash, account_level, voter_level, bonus_votes)
	VALUES ('%v', '%v', '%v', '%v', '%v', '%v', '%v')
	RETURNING account_id::VARCHAR;`,
	newAccount.FirstName, newAccount.LastName, newAccount.Email, passwordHash, account_level, voter_level, bonusVotes)

	var accountID string
	err := p.PostgresClient.QueryRow(context.Background(), queryString).Scan(&accountID)
	if err != nil {
		return "", err
	}

	return accountID, nil
}

func (p *PostgresWrapper) CloseConnection() {
	p.PostgresClient.Close()
}