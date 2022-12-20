package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"sort"

	"github.com/campbelljlowman/fazool-api/auth"
	"github.com/campbelljlowman/fazool-api/constants"
	"github.com/campbelljlowman/fazool-api/graph/generated"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/musicplayer"
	"github.com/campbelljlowman/fazool-api/session"
	"github.com/campbelljlowman/fazool-api/utils"
	"github.com/campbelljlowman/fazool-api/voter"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

// CreateSession is the resolver for the createSession field.
func (r *mutationResolver) CreateSession(ctx context.Context) (*model.User, error) {
	userID, _ := ctx.Value("user").(string)
	if userID == "" {
		return nil, utils.LogErrorMessage("No userID found on token for creating session")
	}

	accountLevel, err := r.database.GetAccountLevel(userID)
	if userID == "" {
		return nil, utils.LogErrorMessage("Error getting account level from database")
	}

	sessionSize := 0
	if accountLevel == constants.RegularAccountLevel {
		sessionSize = 50
	}

	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		return nil, utils.LogErrorObject("Error generating session ID", err)
	}

	session := session.NewSession()

	r.sessionsMutex.Lock()
	r.sessions[sessionID] = &session
	r.sessionsMutex.Unlock()

	// Create session info
	sessionInfo := &model.SessionInfo{
		ID:               sessionID,
		CurrentlyPlaying: nil,
		Queue:            nil,
		Admin:            userID,
		Size:             sessionSize,
	}
	session.SessionInfo = sessionInfo

	// TODO: Maybe combine these two sets to a single db function and query
	err = r.database.SetUserSession(userID, sessionID)

	refreshToken, err := r.database.GetSpotifyRefreshToken(userID)
	if err != nil {
		return nil, utils.LogErrorObject("Error getting Spotify refresh token", err)
	}

	spotifyToken, err := musicplayer.RefreshSpotifyToken(refreshToken)
	if err != nil {
		return nil, utils.LogErrorObject("Error refreshing Spotify Token", err)
	}

	err = r.database.SetSpotifyAccessToken(userID, spotifyToken)
	if err != nil {
		return nil, utils.LogErrorObject("Error setting new Spotify token", err)
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

	session, exists := utils.GetFromMutexedMap(r.sessions, sessionID, r.sessionsMutex)
	if !exists {
		return "", utils.LogErrorMessage(fmt.Sprintf("Session %v not found!", sessionID))
	}

	if userID != session.SessionInfo.Admin {
		return "", utils.LogErrorMessage(fmt.Sprintf("User %v is not the admin for this session!", userID))
	}

	err := r.endSession(session, userID)
	if err != nil {
		return "", utils.LogErrorObject("Error removing session for the user", err)
	}

	return fmt.Sprintf("Session %v successfully deleted", sessionID), nil
}

// UpdateQueue is the resolver for the updateQueue field.
func (r *mutationResolver) UpdateQueue(ctx context.Context, sessionID int, song model.SongUpdate) (*model.SessionInfo, error) {
	// slog.Info("Updating queue", "sessionID", sessionID, "song", song.Title)
	session, exists := utils.GetFromMutexedMap(r.sessions, sessionID, r.sessionsMutex)
	if !exists {
		return nil, utils.LogErrorMessage(fmt.Sprintf("Session %v not found!", sessionID))
	}

	userID := ctx.Value("user").(string)

	existingVoter, voterExists := utils.GetFromMutexedMap(session.Voters, userID, session.VotersMutex)
	if !voterExists {
		return nil, utils.LogErrorMessage(fmt.Sprintf("User not in active voters! User: %v", userID))
	}

	vote, isBonusVote, err := existingVoter.ProcessVote(song.ID, &song.Vote, &song.Action)
	if err != nil {
		return nil, utils.LogErrorObject("Error processing vote", err)
	}
	if isBonusVote {
		session.BonusVoteMutex.Lock()
		if _, exists := session.BonusVotes[song.ID][existingVoter.Id]; !exists {
			session.BonusVotes[song.ID] = make(map[string]int)
		}
		session.BonusVotes[song.ID][existingVoter.Id] += vote
		session.BonusVoteMutex.Unlock()
	}

	existingVoter.Refresh()

	slog.Info("Currently playing", "artist", session.SessionInfo.CurrentlyPlaying.Artist)
	idx := slices.IndexFunc(session.SessionInfo.Queue, func(s *model.Song) bool { return s.ID == song.ID })
	session.QueueMutex.Lock()
	if idx == -1 {
		// add new song to queue
		newSong := &model.Song{
			ID:     song.ID,
			Title:  *song.Title,
			Artist: *song.Artist,
			Image:  *song.Image,
			Votes:  vote,
		}
		session.SessionInfo.Queue = append(session.SessionInfo.Queue, newSong)
	} else {
		queuedSong := session.SessionInfo.Queue[idx]
		queuedSong.Votes += vote
	}

	// Sort queue
	sort.Slice(session.SessionInfo.Queue, func(i, j int) bool { return session.SessionInfo.Queue[i].Votes > session.SessionInfo.Queue[j].Votes })
	session.QueueMutex.Unlock()

	// Update subscription
	session.SendUpdate()

	return session.SessionInfo, nil
}

// UpdateCurrentlyPlaying is the resolver for the updateCurrentlyPlaying field.
func (r *mutationResolver) UpdateCurrentlyPlaying(ctx context.Context, sessionID int, action model.QueueAction) (*model.SessionInfo, error) {
	session, exists := utils.GetFromMutexedMap(r.sessions, sessionID, r.sessionsMutex)

	if !exists {
		return nil, utils.LogErrorMessage(fmt.Sprintf("Session %v not found!", sessionID))
	}

	switch action {
	case "PLAY":
		err := session.MusicPlayer.Play()
		if err != nil {
			return nil, utils.LogErrorObject("Error playing song", err)
		}
	case "PAUSE":
		err := session.MusicPlayer.Pause()
		if err != nil {
			return nil, utils.LogErrorObject("Error pausing song", err)
		}
	case "ADVANCE":
		err := session.AdvanceQueue(true)
		if err != nil {
			return nil, utils.LogErrorObject("Error advancing queue", err)
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
		return "", utils.LogErrorMessage("Invalid email format")
	}

	emailExists, err := r.database.CheckIfEmailExists(newUser.Email)

	if err != nil {
		return "", utils.LogErrorObject("Error searching for email in database", err)
	}

	if emailExists {
		return "", utils.LogErrorMessage("User with this email already exists!")
	}

	passwordHash := utils.HashHelper(newUser.Password)

	userID, err := r.database.AddUserToDatabase(newUser, passwordHash, accountLevel, voterLevel, 0)
	if err != nil {
		return "", utils.LogErrorObject("Error adding user to database", err)
	}

	token, err := auth.GenerateJWT(userID)
	if err != nil {
		return "", utils.LogErrorObject("Error creating user token", err)
	}

	slog.Info("Returning JWT:", "token", token)
	return token, nil
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, userLogin model.UserLogin) (string, error) {
	userID, password, err := r.database.GetUserLoginValues(userLogin.Email)
	if err != nil {
		return "", utils.LogErrorObject("Error getting user login info from database", err)
	}

	if utils.HashHelper(userLogin.Password) != password {
		return "", utils.LogErrorMessage("Invalid Login Credentials!")
	}

	token, err := auth.GenerateJWT(userID)
	if err != nil {
		return "", utils.LogErrorObject("Error creating user token", err)
	}

	return token, nil
}

// UpsertSpotifyToken is the resolver for the upsertSpotifyToken field.
func (r *mutationResolver) UpsertSpotifyToken(ctx context.Context, spotifyCreds model.SpotifyCreds) (*model.User, error) {
	userID, _ := ctx.Value("user").(string)
	if userID == "" {
		return nil, utils.LogErrorMessage("No userID found on token for adding Spotify token")
	}

	// TODO: Probably combine these db sets to a single query
	err := r.database.SetSpotifyAccessToken(userID, spotifyCreds.AccessToken)

	if err != nil {
		return nil, utils.LogErrorObject("Error setting Spotify access token in database", err)
	}

	err = r.database.SetSpotifyRefreshToken(userID, spotifyCreds.RefreshToken)

	if err != nil {
		return nil, utils.LogErrorObject("Error setting Spotify refresh token in database", err)
	}

	return &model.User{ID: userID}, nil
}

// SetPlaylist is the resolver for the setPlaylist field.
func (r *mutationResolver) SetPlaylist(ctx context.Context, sessionID int, playlist string) (*model.SessionInfo, error) {
	userID, _ := ctx.Value("user").(string)
	if userID == "" {
		return nil, utils.LogErrorMessage("No userID found on token for setting playlist")
	}

	session, exists := utils.GetFromMutexedMap(r.sessions, sessionID, r.sessionsMutex)
	if !exists {
		return nil, utils.LogErrorMessage(fmt.Sprintf("Session %v not found!", sessionID))
	}

	if userID != session.SessionInfo.Admin {
		return nil, utils.LogErrorMessage("Only session Admin is permitted to get playlists")
	}

	songs, err := session.MusicPlayer.GetSongsInPlaylist(playlist)
	if err != nil {
		return nil, utils.LogErrorObject("Error getting songs in playlist", err)
	}

	session.QueueMutex.Lock()
	session.SessionInfo.Queue = songs
	session.QueueMutex.Unlock()

	return session.SessionInfo, nil
}

// Session is the resolver for the session field.
func (r *queryResolver) Session(ctx context.Context, sessionID *int) (*model.SessionInfo, error) {
	session, exists := utils.GetFromMutexedMap(r.sessions, *sessionID, r.sessionsMutex)

	if !exists {
		return nil, utils.LogErrorMessage(fmt.Sprintf("Session %v not found!", *sessionID))
	}

	return session.SessionInfo, nil
}

// Voter is the resolver for the voter field.
func (r *queryResolver) Voter(ctx context.Context, sessionID int) (*model.VoterInfo, error) {
	userID := ctx.Value("user").(string)
	// slog.Info("Getting voter", "sessionID", sessionID, "user", userID)

	if userID == "" {
		return nil, utils.LogErrorMessage("No userID found on token for adding Spotify token")
	}

	session, exists := utils.GetFromMutexedMap(r.sessions, sessionID, r.sessionsMutex)

	if !exists {
		return nil, utils.LogErrorMessage(fmt.Sprintf("Session %v not found!", sessionID))
	}

	existingVoter, exists := utils.GetFromMutexedMap(session.Voters, userID, session.VotersMutex)

	if exists {
		slog.Info("return existing voter", "voter", existingVoter.Id)
		return existingVoter.GetVoterInfo(), nil
	}

	if len(session.Voters) >= session.SessionInfo.Size {
		return nil, utils.LogErrorMessage("Session is full of voters!")
	}

	voterType := constants.RegularVoterType
	bonusVotes := 0
	// If userID parses as a UUID, it's a guest voter, so no need to check it in the database
	_, err := uuid.Parse(userID)
	if err != nil {
		voterLevel, bonusVotesValue, err := r.database.GetVoterValues(userID)
		bonusVotes = bonusVotesValue
		if err != nil {
			return nil, utils.LogErrorObject("Error getting voter from database", err)
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
		return nil, utils.LogErrorMessage("Error generating new voter")
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
		return nil, utils.LogErrorMessage("No userID found on token for getting user")
	}

	user, err := r.database.GetUserByID(userID)
	if err != nil {
		return nil, utils.LogErrorObject("Error getting user from database", err)
	}

	return user, nil
}

// Playlists is the resolver for the playlists field.
func (r *queryResolver) Playlists(ctx context.Context, sessionID int) ([]*model.Playlist, error) {
	userID, _ := ctx.Value("user").(string)
	if userID == "" {
		return nil, utils.LogErrorMessage("No userID found on token for getting playlists")
	}

	session, exists := utils.GetFromMutexedMap(r.sessions, sessionID, r.sessionsMutex)
	if !exists {
		return nil, utils.LogErrorMessage(fmt.Sprintf("Session %v not found!", sessionID))
	}

	if userID != session.SessionInfo.Admin {
		return nil, utils.LogErrorMessage("Only session Admin is permitted to get playlists")
	}

	playlists, err := session.MusicPlayer.GetPlaylists()
	if err != nil {
		return nil, utils.LogErrorObject("Error getting playlists for the session!", err)
	}

	return playlists, nil
}

// VoterToken is the resolver for the voterToken field.
func (r *queryResolver) VoterToken(ctx context.Context) (string, error) {
	id := uuid.New()
	return id.String(), nil
}

// SessionUpdated is the resolver for the sessionUpdated field.
func (r *subscriptionResolver) SessionUpdated(ctx context.Context, sessionID int) (<-chan *model.SessionInfo, error) {
	// slog.Info("Subscribing to session")
	session, exists := utils.GetFromMutexedMap(r.sessions, sessionID, r.sessionsMutex)

	if !exists {
		return nil, utils.LogErrorMessage(fmt.Sprintf("Session %v not found!", sessionID))
	}

	channel := make(chan *model.SessionInfo)

	session.ChannelMutex.Lock()
	session.Channels = append(session.Channels, channel)
	session.ChannelMutex.Unlock()

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
