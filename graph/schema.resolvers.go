package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/campbelljlowman/fazool-api/auth"
	"github.com/campbelljlowman/fazool-api/constants"
	"github.com/campbelljlowman/fazool-api/graph/generated"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/musicplayer"
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

	refreshToken := r.accountService.GetSpotifyRefreshToken(accountID)
	slog.Info("spotify refersh token: ", "t", refreshToken)
	if refreshToken == "" {
		return nil, utils.LogAndReturnError("No spotify refresh token found", nil)
	}

	spotifyToken, err := musicplayer.RefreshSpotifyToken(refreshToken)
	if err != nil {
		return nil, utils.LogAndReturnError("Error refreshing Spotify Token", err)
	}

	client := musicplayer.NewSpotifyClient(spotifyToken)

	accountLevel := r.accountService.GetAccountLevel(accountID)

	sessionID, err := r.sessionService.CreateSession(accountID, accountLevel)

	r.musicPlayers[sessionID] = client

	r.accountService.SetAccountActiveSession(accountID, sessionID)

	go r.sessionService.CheckSpotifyCurrentlyPlaying(sessionID, client, r.accountService)
	// TODO: Add this to scheduler, and make this only run once
	go r.sessionService.CheckVotersExpirations(sessionID)

	account := &model.Account{
		ID:        		accountID,
		ActiveSession: &sessionID,
	}

	return account, nil
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

	r.sessionService.EndSession(sessionID)

	return fmt.Sprintf("Session %v successfully deleted", sessionID), nil
}

// UpdateQueue is the resolver for the updateQueue field.
func (r *mutationResolver) UpdateQueue(ctx context.Context, sessionID int, song model.SongUpdate) (*model.SessionInfo, error) {
	// slog.Info("Updating queue", "sessionID", sessionID, "song", song.Title)
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

	// TODO: Remove bonus votes if applicable?
	if isBonusVote {
		r.sessionService.AddBonusVote(song.ID, existingVoter.AccountID, numberOfVotes, sessionID)
	}

	existingVoter.RefreshVoterExpiration()
	r.sessionService.UpsertVoterInSession(sessionID, existingVoter)

	r.sessionService.UpsertQueue(sessionID, numberOfVotes, song)

	return r.sessionService.GetSessionInfo(sessionID), nil
}

// UpdateCurrentlyPlaying is the resolver for the updateCurrentlyPlaying field.
func (r *mutationResolver) UpdateCurrentlyPlaying(ctx context.Context, sessionID int, action model.QueueAction) (*model.SessionInfo, error) {
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

	musicPlayer := r.musicPlayers[sessionID]

	switch action {
	case "PLAY":
		err := musicPlayer.Play()
		if err != nil {
			return nil, utils.LogAndReturnError("Error playing song", err)
		}
	case "PAUSE":
		err := musicPlayer.Pause()
		if err != nil {
			return nil, utils.LogAndReturnError("Error pausing song", err)
		}
	case "ADVANCE":
		err := r.sessionService.AdvanceQueue(sessionID, true, musicPlayer, r.accountService)
		if err != nil {
			return nil, utils.LogAndReturnError("Error advancing queue", err)
		}
	}

	return r.sessionService.GetSessionInfo(sessionID), nil
}

// CreateAccount is the resolver for the createAccount field.
func (r *mutationResolver) CreateAccount(ctx context.Context, newAccount model.NewAccount) (string, error) {
	// Get this from new account request!
	accountLevel := constants.RegularAccountLevel
	voterLevel := constants.RegularVoterType

	isVaildEmail := utils.ValidateEmail(newAccount.Email)
	if !isVaildEmail {
		return "", utils.LogAndReturnError("Invalid email format", nil)
	}

	accountExists := r.accountService.CheckIfEmailHasAccount(newAccount.Email)

	if accountExists {
		return "", utils.LogAndReturnError("Account with this email already exists!", nil)
	}

	passwordHash := utils.HashHelper(newAccount.Password)

	accountID := r.accountService.AddAccount(newAccount, passwordHash, accountLevel, voterLevel, 0)

	jwtToken, err := auth.GenerateJWTForAccount(accountID)
	if err != nil {
		return "", utils.LogAndReturnError("Error creating account token", err)
	}

	return jwtToken, nil
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, accountLogin model.AccountLogin) (string, error) {
	accountID, accountPassword := r.accountService.GetAccountIDAndPassHash(accountLogin.Email)

	if utils.HashHelper(accountLogin.Password) != accountPassword {
		return "", utils.LogAndReturnError("Invalid Login Credentials!", nil)
	}

	jwtToken, err := auth.GenerateJWTForAccount(accountID)
	if err != nil {
		return "", utils.LogAndReturnError("Error creating account token", err)
	}

	return jwtToken, nil
}

// JoinVoters is the resolver for the joinVoters field.
func (r *mutationResolver) JoinVoters(ctx context.Context) (string, error) {
	voterToken := uuid.New()
	return voterToken.String(), nil
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
func (r *mutationResolver) SetPlaylist(ctx context.Context, sessionID int, playlist string) (*model.SessionInfo, error) {
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

	musicPlayer := r.musicPlayers[sessionID]
	songs, err := musicPlayer.GetSongsInPlaylist(playlist)
	if err != nil {
		return nil, utils.LogAndReturnError("Error getting songs in playlist", err)
	}

	var songsToQueue []*model.QueuedSong
	for _, song := range songs {
		songToQueue := &model.QueuedSong{
			SimpleSong: song,
			Votes:      0,
		}
		songsToQueue = append(songsToQueue, songToQueue)
	}

	r.sessionService.SetQueue(sessionID, songsToQueue)

	return r.sessionService.GetSessionInfo(sessionID), nil
}

// Session is the resolver for the session field.
func (r *queryResolver) Session(ctx context.Context, sessionID int) (*model.SessionInfo, error) {
	voterID, _ := ctx.Value("voterID").(string)
	if voterID == "" {
		return nil, utils.LogAndReturnError("Voter ID is required to get session", nil)
	}

	exists := r.sessionService.DoesSessionExist(sessionID)

	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	returnSession := r.sessionService.GetSessionInfo(sessionID)

	return returnSession, nil
}

// Voter is the resolver for the voter field.
func (r *queryResolver) Voter(ctx context.Context, sessionID int) (*model.VoterInfo, error) {
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
		slog.Info("Return existing voter", "voter", existingVoter.VoterID)
		return existingVoter.GetVoterInfo(), nil
	}

	if r.sessionService.IsSessionFull(sessionID) {
		return nil, utils.LogAndReturnError("Session is full of voters!", nil)
	}

	voterType := constants.RegularVoterType
	bonusVotes := 0

	if accountID != 0 {
		voterLevel, bonusVotesValue := r.accountService.GetVoterLevelAndBonusVotes(accountID)
		bonusVotes = bonusVotesValue

		if voterLevel == constants.PrivilegedVoterType {
			voterType = constants.PrivilegedVoterType
		}
		if r.sessionService.GetSessionAdminAccountID(sessionID) == accountID {
			voterType = constants.AdminVoterType
		}
	}

	slog.Info("Generating new voter", "voter", voterID, "bonus-votes", bonusVotes)
	newVoter, err := voter.NewVoter(voterID, voterType, accountID, bonusVotes)
	if err != nil {
		return nil, utils.LogAndReturnError("Error generating new voter", nil)
	}

	slog.Info("New voter created:", "voter", newVoter)
	r.sessionService.UpsertVoterInSession(sessionID, newVoter)

	return newVoter.GetVoterInfo(), nil
}

// Account is the resolver for the account field.
func (r *queryResolver) Account(ctx context.Context) (*model.Account, error) {
	slog.Info("account ID")
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

	musicPlayer := r.musicPlayers[sessionID]
	playlists, err := musicPlayer.GetPlaylists()
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

	slog.Info("Session ID", "sessionID", sessionID)
	exists := r.sessionService.DoesSessionExist(sessionID)
	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	_, voterExists := r.sessionService.GetVoterInSession(sessionID, voterID)
	if !voterExists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("You're not in session %v!", sessionID), nil)
	}
	musicPlayer := r.musicPlayers[sessionID]
	searchResult, err := musicPlayer.Search(query)
	if err != nil {
		return nil, utils.LogAndReturnError("Error executing music search!", err)
	}

	return searchResult, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
