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
	"github.com/campbelljlowman/fazool-api/session"
	"github.com/campbelljlowman/fazool-api/utils"
	"github.com/campbelljlowman/fazool-api/voter"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

// CreateSession is the resolver for the createSession field.
func (r *mutationResolver) CreateSession(ctx context.Context) (*model.User, error) {
	userID, _ := ctx.Value("user").(string)
	if userID == "" {
		return nil, utils.LogAndReturnError("No userID found on token for creating session", nil)
	}

	accountLevel, err := r.database.GetAccountLevel(userID)
	if userID == "" {
		return nil, utils.LogAndReturnError("Error getting account level from database", nil)
	}

	sessionSize := 0
	if accountLevel == constants.RegularAccountLevel {
		sessionSize = 50
	}

	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		return nil, utils.LogAndReturnError("Error generating session ID", err)
	}

	session := session.NewSession()

	r.sessionsMutex.Lock()
	r.sessions[sessionID] = &session
	r.sessionsMutex.Unlock()

	// Create session info
	sessionInfo := &model.SessionInfo{
		ID: sessionID,
		CurrentlyPlaying: &model.CurrentlyPlayingSong{
			SimpleSong: &model.SimpleSong{},
			Playing:    false,
		},
		Queue: nil,
		Admin: userID,
		Size:  sessionSize,
	}
	session.SessionInfo = sessionInfo

	// TODO: Maybe combine these two sets to a single db function and query
	err = r.database.SetUserSession(userID, sessionID)

	refreshToken, err := r.database.GetSpotifyRefreshToken(userID)
	if err != nil {
		return nil, utils.LogAndReturnError("Error getting Spotify refresh token", err)
	}

	spotifyToken, err := musicplayer.RefreshSpotifyToken(refreshToken)
	if err != nil {
		return nil, utils.LogAndReturnError("Error refreshing Spotify Token", err)
	}

	err = r.database.SetSpotifyAccessToken(userID, spotifyToken)
	if err != nil {
		return nil, utils.LogAndReturnError("Error setting new Spotify token", err)
	}

	client := musicplayer.NewSpotifyClient(spotifyToken)
	session.MusicPlayer = client

	go session.WatchSpotifyCurrentlyPlaying()
	go session.WatchVoters()

	user := &model.User{
		ID:        userID,
		SessionID: &sessionID,
	}

	return user, nil
}

// EndSession is the resolver for the endSession field.
func (r *mutationResolver) EndSession(ctx context.Context, sessionID int) (string, error) {
	userID := ctx.Value("user").(string)

	session, exists := utils.GetValueFromMutexedMap(r.sessions, sessionID, r.sessionsMutex)
	if !exists {
		return "", utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	if userID != session.SessionInfo.Admin {
		return "", utils.LogAndReturnError(fmt.Sprintf("User %v is not the admin for this session!", userID), nil)
	}

	err := r.endSession(session)
	if err != nil {
		return "", utils.LogAndReturnError("Error removing session for the user", err)
	}

	return fmt.Sprintf("Session %v successfully deleted", sessionID), nil
}

// UpdateQueue is the resolver for the updateQueue field.
func (r *mutationResolver) UpdateQueue(ctx context.Context, sessionID int, song model.SongUpdate) (*model.SessionInfo, error) {
	// slog.Info("Updating queue", "sessionID", sessionID, "song", song.Title)
	session, exists := utils.GetValueFromMutexedMap(r.sessions, sessionID, r.sessionsMutex)
	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	userID := ctx.Value("user").(string)

	existingVoter, voterExists := utils.GetValueFromMutexedMap(session.Voters, userID, session.VotersMutex)
	if !voterExists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("User not in active voters! User: %v", userID), nil)
	}

	vote, isBonusVote, err := existingVoter.GetVoteAmountAndType(song.ID, &song.Vote, &song.Action)
	if err != nil {
		return nil, utils.LogAndReturnError("Error processing vote", err)
	}
	if isBonusVote {
		session.BonusVoteMutex.Lock()
		if _, exists := session.BonusVotes[song.ID][existingVoter.Id]; !exists {
			session.BonusVotes[song.ID] = make(map[string]int)
		}
		session.BonusVotes[song.ID][existingVoter.Id] += vote
		session.BonusVoteMutex.Unlock()
	}

	existingVoter.RefreshVoterExpiration()

	slog.Info("Currently playing", "artist", session.SessionInfo.CurrentlyPlaying.SimpleSong.Artist)

	session.UpsertQueue(song, vote)

	// Update subscription
	session.SendUpdate()

	return session.SessionInfo, nil
}

// UpdateCurrentlyPlaying is the resolver for the updateCurrentlyPlaying field.
func (r *mutationResolver) UpdateCurrentlyPlaying(ctx context.Context, sessionID int, action model.QueueAction) (*model.SessionInfo, error) {
	session, exists := utils.GetValueFromMutexedMap(r.sessions, sessionID, r.sessionsMutex)

	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	switch action {
	case "PLAY":
		err := session.MusicPlayer.Play()
		if err != nil {
			return nil, utils.LogAndReturnError("Error playing song", err)
		}
	case "PAUSE":
		err := session.MusicPlayer.Pause()
		if err != nil {
			return nil, utils.LogAndReturnError("Error pausing song", err)
		}
	case "ADVANCE":
		err := session.AdvanceQueue(true)
		if err != nil {
			return nil, utils.LogAndReturnError("Error advancing queue", err)
		}
		session.SendUpdate()
	}

	return session.SessionInfo, nil
}

// CreateUser is the resolver for the createUser field.
func (r *mutationResolver) CreateUser(ctx context.Context, newUser model.NewUser) (string, error) {
	// Get this from new user request!
	accountLevel := constants.RegularAccountLevel
	voterLevel := constants.RegularVoterType
	vaildEmail := utils.ValidateEmail(newUser.Email)
	if !vaildEmail {
		return "", utils.LogAndReturnError("Invalid email format", nil)
	}

	emailExists, err := r.database.CheckIfEmailExists(newUser.Email)

	if err != nil {
		return "", utils.LogAndReturnError("Error searching for email in database", err)
	}

	if emailExists {
		return "", utils.LogAndReturnError("User with this email already exists!", nil)
	}

	passwordHash := utils.HashHelper(newUser.Password)

	userID, err := r.database.AddUserToDatabase(newUser, passwordHash, accountLevel, voterLevel, 0)
	if err != nil {
		return "", utils.LogAndReturnError("Error adding user to database", err)
	}

	token, err := auth.GenerateJWT(userID)
	if err != nil {
		return "", utils.LogAndReturnError("Error creating user token", err)
	}

	slog.Info("Returning JWT:", "token", token)
	return token, nil
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, userLogin model.UserLogin) (string, error) {
	userID, password, err := r.database.GetUserLoginValues(userLogin.Email)
	if err != nil {
		return "", utils.LogAndReturnError("Error getting user login info from database", err)
	}

	if utils.HashHelper(userLogin.Password) != password {
		return "", utils.LogAndReturnError("Invalid Login Credentials!", nil)
	}

	token, err := auth.GenerateJWT(userID)
	if err != nil {
		return "", utils.LogAndReturnError("Error creating user token", err)
	}

	return token, nil
}

// UpsertSpotifyToken is the resolver for the upsertSpotifyToken field.
func (r *mutationResolver) UpsertSpotifyToken(ctx context.Context, spotifyCreds model.SpotifyCreds) (*model.User, error) {
	userID, _ := ctx.Value("user").(string)
	if userID == "" {
		return nil, utils.LogAndReturnError("No userID found on token for adding Spotify token", nil)
	}

	// TODO: Probably combine these db sets to a single query
	err := r.database.SetSpotifyAccessToken(userID, spotifyCreds.AccessToken)

	if err != nil {
		return nil, utils.LogAndReturnError("Error setting Spotify access token in database", err)
	}

	err = r.database.SetSpotifyRefreshToken(userID, spotifyCreds.RefreshToken)

	if err != nil {
		return nil, utils.LogAndReturnError("Error setting Spotify refresh token in database", err)
	}

	return &model.User{ID: userID}, nil
}

// SetPlaylist is the resolver for the setPlaylist field.
func (r *mutationResolver) SetPlaylist(ctx context.Context, sessionID int, playlist string) (*model.SessionInfo, error) {
	userID, _ := ctx.Value("user").(string)
	if userID == "" {
		return nil, utils.LogAndReturnError("No userID found on token for setting playlist", nil)
	}

	session, exists := utils.GetValueFromMutexedMap(r.sessions, sessionID, r.sessionsMutex)
	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	if userID != session.SessionInfo.Admin {
		return nil, utils.LogAndReturnError("Only session Admin is permitted to get playlists", nil)
	}

	songs, err := session.MusicPlayer.GetSongsInPlaylist(playlist)
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

	session.SetQueue(songsToQueue)

	return session.SessionInfo, nil
}

// Session is the resolver for the session field.
func (r *queryResolver) Session(ctx context.Context, sessionID *int) (*model.SessionInfo, error) {
	session, exists := utils.GetValueFromMutexedMap(r.sessions, *sessionID, r.sessionsMutex)

	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", *sessionID), nil)
	}

	return session.SessionInfo, nil
}

// Voter is the resolver for the voter field.
func (r *queryResolver) Voter(ctx context.Context, sessionID int) (*model.VoterInfo, error) {
	userID := ctx.Value("user").(string)
	// slog.Info("Getting voter", "sessionID", sessionID, "user", userID)

	if userID == "" {
		return nil, utils.LogAndReturnError("No userID found on token for adding Spotify token", nil)
	}

	session, exists := utils.GetValueFromMutexedMap(r.sessions, sessionID, r.sessionsMutex)

	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	existingVoter, exists := utils.GetValueFromMutexedMap(session.Voters, userID, session.VotersMutex)

	if exists {
		slog.Info("return existing voter", "voter", existingVoter.Id)
		return existingVoter.GetVoterInfo(), nil
	}

	if len(session.Voters) >= session.SessionInfo.Size {
		return nil, utils.LogAndReturnError("Session is full of voters!", nil)
	}

	voterType := constants.RegularVoterType
	bonusVotes := 0
	// If userID parses as a UUID, it's a guest voter, so no need to check it in the database
	_, err := uuid.Parse(userID)
	if err != nil {
		voterLevel, bonusVotesValue, err := r.database.GetVoterValues(userID)
		bonusVotes = bonusVotesValue
		if err != nil {
			return nil, utils.LogAndReturnError("Error getting voter from database", err)
		}

		if voterLevel == constants.PrivilegedVoterType {
			voterType = constants.PrivilegedVoterType
		}
		if session.SessionInfo.Admin == userID {
			voterType = constants.AdminVoterType
		}
	}

	slog.Info("Generating new voter", "voter", userID)
	newVoter, err := voter.NewVoter(userID, voterType, bonusVotes)
	if err != nil {
		return nil, utils.LogAndReturnError("Error generating new voter", nil)
	}

	session.VotersMutex.Lock()
	session.Voters[userID] = newVoter
	session.VotersMutex.Unlock()

	return newVoter.GetVoterInfo(), nil
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context) (*model.User, error) {
	userID, _ := ctx.Value("user").(string)
	if userID == "" {
		return nil, utils.LogAndReturnError("No userID found on token for getting user", nil)
	}

	user, err := r.database.GetUserByID(userID)
	if err != nil {
		return nil, utils.LogAndReturnError("Error getting user from database", err)
	}

	return user, nil
}

// Playlists is the resolver for the playlists field.
func (r *queryResolver) Playlists(ctx context.Context, sessionID int) ([]*model.Playlist, error) {
	userID, _ := ctx.Value("user").(string)
	if userID == "" {
		return nil, utils.LogAndReturnError("No userID found on token for getting playlists", nil)
	}

	session, exists := utils.GetValueFromMutexedMap(r.sessions, sessionID, r.sessionsMutex)
	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	if userID != session.SessionInfo.Admin {
		return nil, utils.LogAndReturnError("Only session Admin is permitted to get playlists", nil)
	}

	playlists, err := session.MusicPlayer.GetPlaylists()
	if err != nil {
		return nil, utils.LogAndReturnError("Error getting playlists for the session!", err)
	}

	return playlists, nil
}

// VoterToken is the resolver for the voterToken field.
func (r *queryResolver) VoterToken(ctx context.Context) (string, error) {
	id := uuid.New()
	return id.String(), nil
}

// MusicSearch is the resolver for the musicSearch field.
func (r *queryResolver) MusicSearch(ctx context.Context, sessionID int, query string) ([]*model.SimpleSong, error) {
	userID, _ := ctx.Value("user").(string)

	session, sessionExists := utils.GetValueFromMutexedMap(r.sessions, sessionID, r.sessionsMutex)

	if !sessionExists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	_, voterExists := utils.GetValueFromMutexedMap(session.Voters, userID, session.VotersMutex)
	if !voterExists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("You're not in session %v!", sessionID), nil)
	}

	searchResult, err := session.MusicPlayer.Search(query)
	if err != nil {
		return nil, utils.LogAndReturnError("Error executing music search!", err)
	}
	
	return searchResult, nil
}

// SessionUpdated is the resolver for the sessionUpdated field.
func (r *subscriptionResolver) SessionUpdated(ctx context.Context, sessionID int) (<-chan *model.SessionInfo, error) {
	// slog.Info("Subscribing to session")
	session, exists := utils.GetValueFromMutexedMap(r.sessions, sessionID, r.sessionsMutex)

	if !exists {
		return nil, utils.LogAndReturnError(fmt.Sprintf("Session %v not found!", sessionID), nil)
	}

	channel := make(chan *model.SessionInfo)

	session.AddChannel(channel)

	return channel, nil
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
