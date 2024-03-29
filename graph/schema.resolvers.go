package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/campbelljlowman/fazool-api/constants"
	"github.com/campbelljlowman/fazool-api/graph/generated"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/streaming"
	"github.com/campbelljlowman/fazool-api/utils"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

// CreateSession is the resolver for the createSession field.
func (r *mutationResolver) CreateSession(ctx context.Context) (*model.Account, error) {
	accountID, _ := ctx.Value("accountID").(int)

	sessionID := r.accountService.GetAccountActiveSession(accountID)
	if sessionID != 0 {
		return nil, utils.LogAndReturnError("You already have an active session", nil)
	}

	// TODO: Add logic here to make this not dependent on spotify alone
	spotifyRefreshToken := r.accountService.GetSpotifyRefreshToken(accountID)
	if spotifyRefreshToken == "" {
		return nil, utils.LogAndReturnError("No spotify refresh token found", nil)
	}

	streamingService := streaming.NewSpotifyClient(spotifyRefreshToken)

	accountType := r.accountService.GetAccountType(accountID)

	sessionID, err := r.sessionService.CreateSession(accountID, accountType, streamingService)
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
	passwordHash, err := r.authService.GenerateBcryptHashForString(newAccount.Password)
	newAccount.Password = "BLANK"
	if err != nil {
		utils.LogAndReturnError("Error processing credentials", err)
	}
	// Get this from new account request!
	accountType := model.AccountTypeFree
	streamingService := model.StreamingServiceNone

	isVaildEmail := utils.ValidateEmail(newAccount.Email)
	if !isVaildEmail {
		return "", utils.LogAndReturnError("Invalid email format", nil)
	}

	accountExists := r.accountService.CheckIfEmailHasAccount(newAccount.Email)

	if accountExists {
		return "", utils.LogAndReturnError("Account with this email already exists!", nil)
	}

	accountID := r.accountService.CreateAccount(newAccount.FirstName, newAccount.LastName, newAccount.Email, newAccount.PhoneNumber, passwordHash, accountType, 0, streamingService)

	jwtAccessToken, err := r.authService.GenerateJWTAccessTokenForAccount(accountID)
	if err != nil {
		return "", utils.LogAndReturnError("Error creating account token", err)
	}

	return jwtAccessToken, nil
}

// UpdateQueue is the resolver for the updateQueue field.
func (r *mutationResolver) UpdateQueue(ctx context.Context, sessionID int, song model.SongUpdate) (*model.SessionState, error) {
	slog.Debug("Updating queue", "sessionID", sessionID, "song", song.Title)
	voterID, _ := ctx.Value("voterID").(string)

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
		r.sessionService.AddUnusedBonusVote(song.ID, existingVoter.AccountID, numberOfVotes, sessionID)
		r.accountService.SubtractBonusVotes(existingVoter.AccountID, numberOfVotes)
	}

	r.sessionService.RefreshVoterExpiration(sessionID, voterID)
	r.sessionService.UpsertVoterInSession(sessionID, existingVoter)
	r.sessionService.UpsertSongInQueue(sessionID, numberOfVotes, song)

	return r.sessionService.GetSessionState(sessionID), nil
}

// UpdateCurrentlyPlaying is the resolver for the updateCurrentlyPlaying field.
func (r *mutationResolver) UpdateCurrentlyPlaying(ctx context.Context, sessionID int, action model.QueueAction) (*model.SessionState, error) {
	accountID, _ := ctx.Value("accountID").(int)

	exists := r.sessionService.DoesSessionExist(sessionID)
	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	if accountID != r.sessionService.GetSessionAdminAccountID(sessionID) {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Account %v is not the admin for this session!", accountID), nil)
	}

	err := r.sessionService.UpdateCurrentlyPlaying(sessionID, action)
	if err != nil {
		return nil, utils.LogAndReturnError("Error updating currently playing", err)
	}

	return r.sessionService.GetSessionState(sessionID), nil
}

// SetSpotifyStreamingService is the resolver for the setSpotifyStreamingService field.
func (r *mutationResolver) SetSpotifyStreamingService(ctx context.Context, spotifyRefreshToken string) (*model.Account, error) {
	accountID, _ := ctx.Value("accountID").(int)

	r.accountService.SetSpotifyStreamingService(accountID, spotifyRefreshToken)

	return &model.Account{ID: accountID}, nil
}

// SetPlaylist is the resolver for the setPlaylist field.
func (r *mutationResolver) SetPlaylist(ctx context.Context, sessionID int, playlistID string) (*model.SessionState, error) {
	accountID, _ := ctx.Value("accountID").(int)

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

// SetAccountType is the resolver for the setAccountType field.
func (r *mutationResolver) SetAccountType(ctx context.Context, targetAccountID int, accountType model.AccountType) (*model.Account, error) {
	accountID, _ := ctx.Value("accountID").(int)
	if accountID != targetAccountID {
		return nil, utils.LogAndReturnError("You can only set your own account level!", nil)
	}

	return r.accountService.SetAccountType(targetAccountID, accountType), nil
}

// AddSuperVoter is the resolver for the addSuperVoter field.
func (r *mutationResolver) SetSuperVoterSession(ctx context.Context, sessionID int, targetAccountID int) (*model.Account, error) {
	voterID, _ := ctx.Value("voterID").(string)
	accountID, _ := ctx.Value("accountID").(int)
	if accountID != targetAccountID {
		return nil, utils.LogAndReturnError("You can only set your own voter super voter type", nil)
	}

	accountFazoolTokens := r.accountService.GetAccountFazoolTokens(accountID)
	if accountFazoolTokens < constants.SuperVoterCost {
		return nil, utils.LogAndReturnError("You don't have enough Fazool tokens", nil)
	}

	voter, _ := r.sessionService.GetVoterInSession(sessionID, voterID)
	if voter.VoterType == model.VoterTypeSuper {
		return nil, utils.LogAndReturnError("You're already a super voter", nil)
	}
	voter.VoterType = model.VoterTypeSuper
	r.sessionService.UpsertVoterInSession(sessionID, voter)

	return r.accountService.SetSuperVoterSession(targetAccountID, sessionID, constants.SuperVoterCost), nil
}

// AddBonusVotes is the resolver for the addBonusVotes field.
func (r *mutationResolver) AddBonusVotes(ctx context.Context, sessionID int, targetAccountID int, bonusVoteAmount model.BonusVoteAmount) (*model.Account, error) {
	voterID, _ := ctx.Value("voterID").(string)
	accountID, _ := ctx.Value("accountID").(int)
	if accountID != targetAccountID {
		return nil, utils.LogAndReturnError("You can only set your own bonus votes!", nil)
	}

	bonusVoteCostMapping := constants.BonusVoteCostMapping[bonusVoteAmount]

	accountFazoolTokens := r.accountService.GetAccountFazoolTokens(accountID)
	if accountFazoolTokens < bonusVoteCostMapping.CostInFazoolTokens {
		return nil, utils.LogAndReturnError("You don't have enough Fazool tokens", nil)
	}

	voter, _ := r.sessionService.GetVoterInSession(sessionID, voterID)
	voter.BonusVotes += bonusVoteCostMapping.NumberOfBonusVotes
	r.sessionService.UpsertVoterInSession(sessionID, voter)

	return r.accountService.AddBonusVotes(targetAccountID, bonusVoteCostMapping.NumberOfBonusVotes, bonusVoteCostMapping.CostInFazoolTokens), nil
}

// AddFazoolTokens is the resolver for the addFazoolTokens field.
func (r *mutationResolver) AddFazoolTokens(ctx context.Context, sessionID int, targetAccountID int, fazoolTokenAmount model.FazoolTokenAmount) (string, error) {
	accountID, _ := ctx.Value("accountID").(int)
	if accountID != targetAccountID {
		return "", utils.LogAndReturnError("You can only set your own bonus votes!", nil)
	}

	return r.stripeService.CreateCheckoutSession(sessionID, targetAccountID, fazoolTokenAmount)
}

// CreatePasswordChangeRequest is the resolver for the createPasswordChangeRequest field.
func (r *mutationResolver) CreatePasswordChangeRequest(ctx context.Context, email string) (string, error) {
	accountExists := r.accountService.CheckIfEmailHasAccount(email)
	if !accountExists {
		return "", utils.LogAndReturnError("this email address doesn't have an account", nil)
	}
	
	account := r.accountService.GetAccountFromEmail(email)

	err := r.authService.CreateAndSendPasswordChangeRequest(*account.Email, account.ID)
	if err != nil {
		return "", utils.LogAndReturnError("error generating password change request", err)
	}
	return "Password change request send", nil
}

// ChangePassword is the resolver for the changePassword field.
func (r *mutationResolver) ChangePassword(ctx context.Context, passwordChangeRequestID string, newPassword string) (*model.Account, error) {
	accountID, passwordChangeRequestIsValid := r.authService.ValidatePasswordChangeRequest(passwordChangeRequestID)
	if !passwordChangeRequestIsValid {
		return nil, utils.LogAndReturnError("password change request isn't valid", nil)
	}

	passwordHash, err := r.authService.GenerateBcryptHashForString(newPassword)
	if err != nil {
		return nil, utils.LogAndReturnError("error generating password hash", err)
	}

	return r.accountService.SetAccountPasswordHash(accountID, passwordHash), nil
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, accountLogin model.AccountLogin) (string, error) {
	accountID, accountPasswordHash := r.accountService.GetAccountIDAndPassHash(accountLogin.Email)

	if !r.authService.CompareBcryptHashAndString(accountPasswordHash, accountLogin.Password) {
		return "", utils.LogAndReturnError("Invalid Login Credentials!", nil)
	}

	jwtAccessToken, err := r.authService.GenerateJWTAccessTokenForAccount(accountID)
	if err != nil {
		return "", utils.LogAndReturnError("Error creating account access token", err)
	}

	return jwtAccessToken, nil
}

// EndSession is the resolver for the endSession field.
func (r *mutationResolver) EndSession(ctx context.Context, sessionID int) (string, error) {
	accountID, _ := ctx.Value("accountID").(int)

	exists := r.sessionService.DoesSessionExist(sessionID)
	if !exists {
		return "", utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	if accountID != r.sessionService.GetSessionAdminAccountID(sessionID) {
		return "", utils.LogAndReturnError(fmt.Sprintf("Account %v is not the admin for this session!", accountID), nil)
	}

	r.sessionService.EndSession(sessionID)

	return fmt.Sprintf("Session %v successfully deleted", sessionID), nil
}

// RemoveSongFromQueue is the resolver for the removeSongFromQueue field.
func (r *mutationResolver) RemoveSongFromQueue(ctx context.Context, sessionID int, songID string) (*model.SessionState, error) {
	accountID, _ := ctx.Value("accountID").(int)

	exists := r.sessionService.DoesSessionExist(sessionID)
	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	if accountID != r.sessionService.GetSessionAdminAccountID(sessionID) {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Account %v is not the admin for this session!", accountID), nil)
	}

	r.sessionService.RemoveSongFromQueue(sessionID, songID)

	return r.sessionService.GetSessionState(sessionID), nil
}

// RemoveSpotifyStreamingService is the resolver for the removeSpotifyStreamingService field.
func (r *mutationResolver) RemoveSpotifyStreamingService(ctx context.Context, targetAccountID int) (*model.Account, error) {
	accountID, _ := ctx.Value("accountID").(int)
	if accountID != targetAccountID {
		return nil, utils.LogAndReturnError("you can only remove spotify on your own account", nil)
	}

	return r.accountService.RemoveSpotifyStreamingService(targetAccountID), nil
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
	exists := r.sessionService.DoesSessionExist(sessionID)

	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	return r.sessionService.GetSessionConfig(sessionID), nil
}

// SessionState is the resolver for the sessionState field.
func (r *queryResolver) SessionState(ctx context.Context, sessionID int) (*model.SessionState, error) {
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

	exists := r.sessionService.DoesSessionExist(sessionID)
	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	existingVoter, exists := r.sessionService.GetVoterInSession(sessionID, voterID)

	if exists && existingVoter.AccountID == accountID {
		slog.Debug("Return existing voter", "voter", existingVoter.VoterID)
		return existingVoter.ConvertVoterType(), nil
	}

	if exists && existingVoter.AccountID != accountID {
		slog.Debug("Adding account to existing voter", "voter", existingVoter.VoterID)
		existingVoter = r.sessionService.UpdateVoterAccount(sessionID, accountID, existingVoter)
		return existingVoter.ConvertVoterType(), nil
	}

	if r.sessionService.IsSessionFull(sessionID) {
		return nil, utils.LogAndReturnError("Session is full of voters!", nil)
	}

	return r.sessionService.CreateVoterInSession(sessionID, accountID, voterID)
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

	account := r.accountService.GetAccountFromID(accountID)

	return account, nil
}

// Playlists is the resolver for the playlists field.
func (r *queryResolver) Playlists(ctx context.Context, sessionID int) ([]*model.Playlist, error) {
	accountID, _ := ctx.Value("accountID").(int)

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
