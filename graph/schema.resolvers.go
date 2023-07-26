package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/campbelljlowman/fazool-api/auth"
	"github.com/campbelljlowman/fazool-api/graph/generated"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/streaming"
	"github.com/campbelljlowman/fazool-api/utils"
	"github.com/campbelljlowman/fazool-api/voter"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

// CreateSession is the resolver for the createSession field.
func (r *mutationResolver) CreateSession(ctx context.Context) (*model.Account, error) {
	accountID, _ := ctx.Value("accountID").(int)
	if accountID == 0 {
		return nil, utils.LogAndReturnError("Account ID is required to create session", nil)
	}

	sessionID := r.accountService.GetAccountActiveSession(accountID)
	if sessionID != 0 {
		return nil, utils.LogAndReturnError("You already have an active session", nil)
	}

	refreshToken := r.accountService.GetSpotifyRefreshToken(accountID)
	if refreshToken == "" {
		return nil, utils.LogAndReturnError("No spotify refresh token found", nil)
	}

	client := streaming.NewSpotifyClient(refreshToken)

	accountType := r.accountService.GetAccountType(accountID)

	sessionID, err := r.sessionService.CreateSession(accountID, accountType, client, r.accountService)
	if err != nil {
		return nil, utils.LogAndReturnError("Error creating new session", err)
	}

	r.accountService.SetAccountActiveSession(accountID, sessionID)

	account := &model.Account{
		ID:            accountID,
		ActiveSession: &sessionID,
	}

	return account, nil
}

// CreateAccount is the resolver for the createAccount field.
func (r *mutationResolver) CreateAccount(ctx context.Context, newAccount model.NewAccount) (string, error) {
	slog.Debug("Creating new account")
	passwordHash, err := utils.HashPassword(newAccount.Password)
	newAccount.Password = "BLANK"
	if err != nil {
		utils.LogAndReturnError("Error processing credentials", err)
	}
	// Get this from new account request!
	accountType := model.AccountTypeFree
	voterType := model.VoterTypeFree
	streamingService := model.StreamingServiceNone

	isVaildEmail := utils.ValidateEmail(newAccount.Email)
	if !isVaildEmail {
		return "", utils.LogAndReturnError("Invalid email format", nil)
	}

	accountExists := r.accountService.CheckIfEmailHasAccount(newAccount.Email)

	if accountExists {
		return "", utils.LogAndReturnError("Account with this email already exists!", nil)
	}

	accountID := r.accountService.CreateAccount(newAccount.FirstName, newAccount.LastName, newAccount.Email, passwordHash, accountType, voterType, 0, streamingService)

	jwtToken, err := auth.GenerateJWTForAccount(accountID)
	if err != nil {
		return "", utils.LogAndReturnError("Error creating account token", err)
	}

	return jwtToken, nil
}

// UpdateQueue is the resolver for the updateQueue field.
func (r *mutationResolver) UpdateQueue(ctx context.Context, sessionID int, song model.SongUpdate) (*model.SessionState, error) {
	slog.Debug("Updating queue", "sessionID", sessionID, "song", song.Title)
	voterID, _ := ctx.Value("voterID").(string)
	if voterID == "" {
		return nil, utils.LogAndReturnError("Voter ID is required to update queue", nil)
	}

	exists := r.sessionService.DoesSessionExist(sessionID)
	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	existingVoter, voterExists := r.sessionService.GetVoterInSession(sessionID, voterID)
	if !voterExists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Voter not in active voters! Voter: %v", voterID), nil)
	}

	numberOfVotes, isBonusVote, err := existingVoter.CalculateAndProcessVote(song.ID, &song.Vote, &song.Action)
	if err != nil {
		return nil, utils.LogAndReturnError("Error processing vote", err)
	}

	if isBonusVote {
		r.sessionService.AddBonusVote(song.ID, existingVoter.AccountID, numberOfVotes, sessionID)
	}

	r.sessionService.RefreshVoterExpiration(sessionID, voterID)
	r.sessionService.UpsertVoterInSession(sessionID, existingVoter)
	r.sessionService.UpsertQueue(sessionID, numberOfVotes, song)

	return r.sessionService.GetSessionState(sessionID), nil
}

// UpdateCurrentlyPlaying is the resolver for the updateCurrentlyPlaying field.
func (r *mutationResolver) UpdateCurrentlyPlaying(ctx context.Context, sessionID int, action model.QueueAction) (*model.SessionState, error) {
	accountID, _ := ctx.Value("accountID").(int)
	if accountID == 0 {
		return nil, utils.LogAndReturnError("Account ID is required to update currently playing song", nil)
	}

	exists := r.sessionService.DoesSessionExist(sessionID)
	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	if accountID != r.sessionService.GetSessionAdminAccountID(sessionID) {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Account %v is not the admin for this session!", accountID), nil)
	}

	err := r.sessionService.UpdateCurrentlyPlaying(sessionID, action, r.accountService)
	if err != nil {
		return nil, utils.LogAndReturnError("Error updating currently playing", err)
	}

	return r.sessionService.GetSessionState(sessionID), nil
}

// UpsertSpotifyToken is the resolver for the upsertSpotifyToken field.
func (r *mutationResolver) UpsertSpotifyToken(ctx context.Context, spotifyCreds model.SpotifyCreds) (*model.Account, error) {
	accountID, _ := ctx.Value("accountID").(int)
	if accountID == 0 {
		return nil, utils.LogAndReturnError("Account ID required for adding Spotify token", nil)
	}

	r.accountService.SetSpotifyRefreshToken(accountID, spotifyCreds.RefreshToken)

	return &model.Account{ID: accountID}, nil
}

// SetPlaylist is the resolver for the setPlaylist field.
func (r *mutationResolver) SetPlaylist(ctx context.Context, sessionID int, playlistID string) (*model.SessionState, error) {
	accountID, _ := ctx.Value("accountID").(int)
	if accountID == 0 {
		return nil, utils.LogAndReturnError("Account ID required for setting playlist", nil)
	}

	exists := r.sessionService.DoesSessionExist(sessionID)
	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	if accountID != r.sessionService.GetSessionAdminAccountID(sessionID) {
		return nil, utils.LogAndReturnError("Only session Admin is permitted to set playlists", nil)
	}

	err := r.sessionService.SetPlaylist(sessionID, playlistID)
	if err != nil {
		return nil, utils.LogAndReturnError("Error setting sesison playlist", err)
	}

	return r.sessionService.GetSessionState(sessionID), nil
}

// SetVoterType is the resolver for the setVoterType field.
func (r *mutationResolver) SetVoterType(ctx context.Context, targetAccountID int, voterType model.VoterType) (*model.Account, error) {
	accountID, _ := ctx.Value("accountID").(int)
	if accountID != targetAccountID {
		return nil, utils.LogAndReturnError("You can only set your own voter level!", nil)
	}

	return r.accountService.SetVoterType(targetAccountID, voterType), nil
}

// SetAccountType is the resolver for the setAccountType field.
func (r *mutationResolver) SetAccountType(ctx context.Context, targetAccountID int, accountType model.AccountType) (*model.Account, error) {
	accountID, _ := ctx.Value("accountID").(int)
	if accountID != targetAccountID {
		return nil, utils.LogAndReturnError("You can only set your own account level!", nil)
	}

	return r.accountService.SetAccountType(targetAccountID, accountType), nil
}

// AddBonusVotes is the resolver for the addBonusVotes field.
func (r *mutationResolver) AddBonusVotes(ctx context.Context, targetAccountID int, bonusVotes int) (*model.Account, error) {
	accountID, _ := ctx.Value("accountID").(int)
	if accountID != targetAccountID {
		return nil, utils.LogAndReturnError("You can only set your own bonus votes!", nil)
	}

	return r.accountService.AddBonusVotes(targetAccountID, bonusVotes), nil
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, accountLogin model.AccountLogin) (string, error) {
	accountID, accountPasswordHash := r.accountService.GetAccountIDAndPassHash(accountLogin.Email)

	if !utils.CompareHashAndPassword(accountPasswordHash, accountLogin.Password) {
		return "", utils.LogAndReturnError("Invalid Login Credentials!", nil)
	}

	jwtToken, err := auth.GenerateJWTForAccount(accountID)
	if err != nil {
		return "", utils.LogAndReturnError("Error creating account token", err)
	}

	return jwtToken, nil
}

// EndSession is the resolver for the endSession field.
func (r *mutationResolver) EndSession(ctx context.Context, sessionID int) (string, error) {
	accountID, _ := ctx.Value("accountID").(int)
	if accountID == 0 {
		return "", utils.LogAndReturnError("Account ID is required to end session", nil)
	}

	exists := r.sessionService.DoesSessionExist(sessionID)
	if !exists {
		return "", utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	if accountID != r.sessionService.GetSessionAdminAccountID(sessionID) {
		return "", utils.LogAndReturnError(fmt.Sprintf("Account %v is not the admin for this session!", accountID), nil)
	}

	r.sessionService.EndSession(sessionID, r.accountService)

	return fmt.Sprintf("Session %v successfully deleted", sessionID), nil
}

// DeleteAccount is the resolver for the deleteAccount field.
func (r *mutationResolver) DeleteAccount(ctx context.Context, targetAccountID int) (string, error) {
	accountID, _ := ctx.Value("accountID").(int)
	if accountID != targetAccountID {
		return "Failed to delete account", utils.LogAndReturnError("You can only delete your own account!", nil)
	}
	r.accountService.DeleteAccount(targetAccountID)
	return fmt.Sprintf("Account %v successfully deleted", targetAccountID), nil
}

// SessionConfig is the resolver for the sessionConfig field.
func (r *queryResolver) SessionConfig(ctx context.Context, sessionID int) (*model.SessionConfig, error) {
	voterID, _ := ctx.Value("voterID").(string)
	if voterID == "" {
		return nil, utils.LogAndReturnError("Voter ID is required to get session config", nil)
	}

	exists := r.sessionService.DoesSessionExist(sessionID)

	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	return r.sessionService.GetSessionConfig(sessionID), nil
}

// SessionState is the resolver for the sessionState field.
func (r *queryResolver) SessionState(ctx context.Context, sessionID int) (*model.SessionState, error) {
	voterID, _ := ctx.Value("voterID").(string)
	if voterID == "" {
		return nil, utils.LogAndReturnError("Voter ID is required to get session state", nil)
	}

	exists := r.sessionService.DoesSessionExist(sessionID)

	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	return r.sessionService.GetSessionState(sessionID), nil
}

// Voter is the resolver for the voter field.
func (r *queryResolver) Voter(ctx context.Context, sessionID int) (*model.Voter, error) {
	voterID, _ := ctx.Value("voterID").(string)
	accountID, _ := ctx.Value("accountID").(int)
	if voterID == "" {
		return nil, utils.LogAndReturnError("Voter ID required for getting voter", nil)
	}

	exists := r.sessionService.DoesSessionExist(sessionID)
	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	existingVoter, exists := r.sessionService.GetVoterInSession(sessionID, voterID)

	if exists {
		slog.Debug("Return existing voter", "voter", existingVoter.VoterID)
		return existingVoter.ConvertVoterType(), nil
	}

	if r.sessionService.IsSessionFull(sessionID) {
		return nil, utils.LogAndReturnError("Session is full of voters!", nil)
	}

	voterType := model.VoterTypeFree
	bonusVotes := 0

	if accountID != 0 {
		voterTypeLocal, bonusVotesValue := r.accountService.GetVoterTypeAndBonusVotes(accountID)
		bonusVotes = bonusVotesValue

		if voterTypeLocal == model.VoterTypePrivileged {
			voterType = model.VoterTypePrivileged
		}
		if r.sessionService.GetSessionAdminAccountID(sessionID) == accountID {
			voterType = model.VoterTypeAdmin
		}
	}

	newVoter, err := voter.NewVoter(voterID, voterType, accountID, bonusVotes)
	if err != nil {
		return nil, utils.LogAndReturnError("Error generating new voter", nil)
	}

	slog.Debug("New voter created:", "voter", newVoter)
	r.sessionService.UpsertVoterInSession(sessionID, newVoter)

	return newVoter.ConvertVoterType(), nil
}

// VoterToken is the resolver for the voterToken field.
func (r *queryResolver) VoterToken(ctx context.Context, sessionID int) (string, error) {
	sessionExists := r.sessionService.DoesSessionExist(sessionID)
	if !sessionExists {
		return "", utils.LogAndReturnError("Voter token requested for session that doesn't exist", nil)
	}
	
	slog.Debug("Giving new voter token")
	voterToken := uuid.New()
	return voterToken.String(), nil
}

// Account is the resolver for the account field.
func (r *queryResolver) Account(ctx context.Context) (*model.Account, error) {
	accountID, _ := ctx.Value("accountID").(int)
	if accountID == 0 {
		return nil, utils.LogAndReturnError("No account found on token for getting account", nil)
	}

	account := r.accountService.GetAccountFromID(accountID)

	return account, nil
}

// Playlists is the resolver for the playlists field.
func (r *queryResolver) Playlists(ctx context.Context, sessionID int) ([]*model.Playlist, error) {
	accountID, _ := ctx.Value("accountID").(int)
	if accountID == 0 {
		return nil, utils.LogAndReturnError("accountID required for getting playlists", nil)
	}

	exists := r.sessionService.DoesSessionExist(sessionID)
	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	if accountID != r.sessionService.GetSessionAdminAccountID(sessionID) {
		return nil, utils.LogAndReturnError("Only session Admin is permitted to get playlists", nil)
	}

	playlists, err := r.sessionService.GetPlaylists(sessionID)
	if err != nil {
		return nil, utils.LogAndReturnError("Error getting playlists for the session!", err)
	}

	return playlists, nil
}

// MusicSearch is the resolver for the musicSearch field.
func (r *queryResolver) MusicSearch(ctx context.Context, sessionID int, query string) ([]*model.SimpleSong, error) {
	voterID, _ := ctx.Value("voterID").(string)
	if voterID == "" {
		return nil, utils.LogAndReturnError("Voter ID required for searching for music", nil)
	}

	exists := r.sessionService.DoesSessionExist(sessionID)
	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	_, voterExists := r.sessionService.GetVoterInSession(sessionID, voterID)
	if !voterExists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("You're not in session %v!", sessionID), nil)
	}

	searchResult, err := r.sessionService.SearchForSongs(sessionID, query)
	if err != nil {
		return nil, utils.LogAndReturnError("Error executing music search!", err)
	}

	return searchResult, nil
}

// SubscribeSessionState is the resolver for the subscribeSessionState field.
func (r *subscriptionResolver) SubscribeSessionState(ctx context.Context, sessionID int) (<-chan *model.SessionState, error) {
	slog.Debug("Creating new subscription")
	// voterID, _ := ctx.Value("voterID").(string)
	// if voterID == "" {
	// 	return nil, utils.LogAndReturnError("Voter ID is required to subscribe to session state", nil)
	// }

	exists := r.sessionService.DoesSessionExist(sessionID)
	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	newChannel := make(chan *model.SessionState)

	r.sessionService.AddChannel(sessionID, newChannel)

	return newChannel, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
