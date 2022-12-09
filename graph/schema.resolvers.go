package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/campbelljlowman/fazool-api/auth"
	"github.com/campbelljlowman/fazool-api/graph/generated"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/session"
	"github.com/campbelljlowman/fazool-api/spotifyUtil"
	"github.com/campbelljlowman/fazool-api/utils"
	"github.com/campbelljlowman/fazool-api/voter"
	"github.com/google/uuid"
	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
)

// CreateSession is the resolver for the createSession field.
func (r *mutationResolver) CreateSession(ctx context.Context) (*model.User, error) {
	// TODO: Check users account level from db and set session size accordingly
	sessionSize := 100
	userID, _ := ctx.Value("user").(string)
	if userID == "" {
		return nil, utils.LogErrorMessage("No userID found on token for creating session")
	}

	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		return nil, utils.LogErrorObject("Error generating session ID", err)
	}

	session := session.NewSession()
	r.sessions[sessionID] = &session

	// Create session info
	sessionInfo := &model.SessionInfo{
		ID:               sessionID,
		CurrentlyPlaying: nil,
		Queue:            nil,
		Admin:            userID,
		Size:             sessionSize,
	}
	session.SessionInfo = sessionInfo

	err = r.database.SetUserSession(userID, sessionID)

	refreshToken, err := r.database.GetSpotifyRefreshToken(userID)
	if err != nil {
		return nil, utils.LogErrorObject("Error getting Spotify refresh token", err)
	}

	spotifyToken, err := spotifyUtil.RefreshToken(refreshToken)
	if err != nil {
		return nil, utils.LogErrorObject("Error refreshing Spotify Token", err)
	}

	err = r.database.SetSpotifyAccessToken(userID, spotifyToken)
	if err != nil {
		return nil, utils.LogErrorObject("Error setting new Spotify token", err)
	}

	// TODO: Use refresh token as well? https://pkg.go.dev/golang.org/x/oauth2#Token
	token := &oauth2.Token{
		AccessToken: spotifyToken,
	}
	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)
	session.SpotifyPlayer = client

	go session.WatchSpotifyCurrentlyPlaying()

	user := &model.User{
		ID:        userID,
		SessionID: &sessionID,
	}

	return user, nil
}

// UpdateQueue is the resolver for the updateQueue field.
func (r *mutationResolver) UpdateQueue(ctx context.Context, sessionID int, song model.SongUpdate) (*model.SessionInfo, error) {
	// TODO: Check for voter auth on context and make sure sessionIDs match
	// slog.Info("Updating queue", "sessionID", sessionID, "song", song.Title)
	session := r.sessions[sessionID]

	userID := ctx.Value("user").(string)
	session.VotersMutex.Lock()
	existingVoter, voterExists := session.Voters[userID]
	session.VotersMutex.Unlock()

	if !voterExists {
		return nil, fmt.Errorf("User not in active voters! User: %v", userID)
	}

	vote, err := existingVoter.ProcessVote(song.ID, &song.Vote, &song.Action)
	if err != nil {
		errorMsg := "Error processing Vote"
		slog.Warn(errorMsg, "error", err)
		return nil, errors.New(errorMsg)
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
	session := r.sessions[sessionID]

	switch action {
	case "PLAY":
		session.SpotifyPlayer.Play(ctx)
	case "PAUSE":
		session.SpotifyPlayer.Pause(ctx)
	case "ADVANCE":
		session.AdvanceQueue(true)
		session.SendUpdate()
	}

	return session.SessionInfo, nil
}

// CreateUser is the resolver for the createUser field.
func (r *mutationResolver) CreateUser(ctx context.Context, newUser model.NewUser) (string, error) {
	// Get this from new user request!
	accountLevel := "free"
	vaildEmail := utils.ValidateEmail(newUser.Email)
	if !vaildEmail {
		errorMsg := "Invalid email format"
		slog.Warn(errorMsg)
		return "", errors.New(errorMsg)
	}

	emailExists, err := r.database.CheckIfEmailExists(newUser.Email)

	if err != nil {
		errorMsg := "Error searching for email in database"
		slog.Warn(errorMsg, "error", err)
		return "", errors.New(errorMsg)
	}
	if emailExists {
		errorMsg := "User with this email already exists!"
		return "", errors.New(errorMsg)
	}

	passwordHash := utils.HashHelper(newUser.Password)

	userID, err := r.database.AddUserToDatabase(newUser, passwordHash, accountLevel, 0)
	if err != nil {
		errorMsg := "Error adding user to database"
		slog.Warn(errorMsg, "error", err)
		return "", errors.New(errorMsg)
	}

	token, err := auth.GenerateJWT(userID)
	if err != nil {
		errorMsg := "Error creating user token"
		slog.Warn(errorMsg, "error", err)
		return "", errors.New(errorMsg)
	}

	slog.Info("Returning JWT:", "token", token)
	return token, nil
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, userLogin model.UserLogin) (string, error) {
	userID, password, err := r.database.GetUserLoginValues(userLogin.Email)
	if err != nil {
		errorMsg := "Error getting user login info from database"
		slog.Warn(errorMsg, "error", err)
		return "", errors.New(errorMsg)
	}

	if utils.HashHelper(userLogin.Password) != password {
		errorMsg := "Invalid Login Credentials!"
		slog.Warn(errorMsg)
		return "", errors.New(errorMsg)
	}

	token, err := auth.GenerateJWT(userID)
	if err != nil {
		errorMsg := "Error creating user token"
		slog.Warn(errorMsg, "error", err)
		return "", errors.New(errorMsg)
	}

	return token, nil
}

// UpsertSpotifyToken is the resolver for the upsertSpotifyToken field.
func (r *mutationResolver) UpsertSpotifyToken(ctx context.Context, spotifyCreds model.SpotifyCreds) (*model.User, error) {
	userID, _ := ctx.Value("user").(string)
	if userID == "" {
		return nil, fmt.Errorf("No userID found on token for adding Spotify token")
	}

	err := r.database.SetSpotifyAccessToken(userID, spotifyCreds.AccessToken)

	if err != nil {
		errorMsg := "Error setting Spotify access token in database"
		slog.Warn(errorMsg, "error", err)
		return nil, errors.New(errorMsg)
	}

	err = r.database.SetSpotifyRefreshToken(userID, spotifyCreds.RefreshToken)

	if err != nil {
		errorMsg := "Error setting Spotify refresh token in database"
		slog.Warn(errorMsg, "error", err)
		return nil, errors.New(errorMsg)
	}

	return &model.User{ID: userID}, nil
}

// SetOutputDevice is the resolver for the setOutputDevice field.
func (r *mutationResolver) SetOutputDevice(ctx context.Context, outputDevice model.OutputDevice) (*model.Device, error) {
	panic(fmt.Errorf("not implemented: SetOutputDevice - setOutputDevice"))
}

// SetPlaylist is the resolver for the setPlaylist field.
func (r *mutationResolver) SetPlaylist(ctx context.Context, playlist model.PlaylistInput) (*model.Playlist, error) {
	panic(fmt.Errorf("not implemented: SetPlaylist - setPlaylist"))
}

// Session is the resolver for the session field.
func (r *queryResolver) Session(ctx context.Context, sessionID *int) (*model.SessionInfo, error) {
	session, exists := r.sessions[*sessionID]
	if !exists {
		errorMsg := "Session not found!"
		slog.Warn(errorMsg)
		return nil, errors.New(errorMsg)
	}

	return session.SessionInfo, nil
}

// Voter is the resolver for the voter field.
func (r *queryResolver) Voter(ctx context.Context, sessionID int) (*model.VoterInfo, error) {
	userID := ctx.Value("user").(string)
	slog.Info("Getting voter", "sessionID", sessionID, "user", userID)

	if userID == "" {
		return nil, utils.LogErrorMessage("No userID found on token for adding Spotify token")
	}

	session := r.sessions[sessionID]

	session.VotersMutex.Lock()
	existingVoter, exists := session.Voters[userID]
	session.VotersMutex.Unlock()

	if exists {
		slog.Info("Returning existing voter")
		return existingVoter.GetVoterInfo(), nil
	}

	if len(session.Voters) >= session.SessionInfo.Size {
		return nil, utils.LogErrorMessage("Session is full of voters!")
	}

	voterType := voter.RegularVoterType
	priviledged := true
	if priviledged {
		voterType = voter.PrivilegedVoterType
	}
	if session.SessionInfo.Admin == userID {
		voterType = voter.AdminVoterType
	}


	// TODO: Check db for bonus votes
	newVoter, err := voter.NewVoter(userID, voterType, 0)
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
	// TODO: Validate tis input
	userID, _ := ctx.Value("user").(string)
	if userID == "" {
		return nil, fmt.Errorf("No userID found on token for getting user")
	}

	user, err := r.database.GetUserByID(userID)
	if err != nil {
		errorMsg := "Error getting user from database"
		slog.Warn(errorMsg, "error", err)
		return nil, errors.New(errorMsg)
	}

	return user, nil
}

// Devices is the resolver for the devices field.
func (r *queryResolver) Devices(ctx context.Context) ([]*model.Device, error) {
	panic(fmt.Errorf("not implemented: Devices - devices"))
}

// Playlists is the resolver for the playlists field.
func (r *queryResolver) Playlists(ctx context.Context) ([]*model.Playlist, error) {
	panic(fmt.Errorf("not implemented: Playlists - playlists"))
}

// VoterToken is the resolver for the voterToken field.
func (r *queryResolver) VoterToken(ctx context.Context) (string, error) {
	id := uuid.New()
	return id.String(), nil
}

// SessionUpdated is the resolver for the sessionUpdated field.
func (r *subscriptionResolver) SessionUpdated(ctx context.Context, sessionID int) (<-chan *model.SessionInfo, error) {
	// slog.Info("Subscribing to session")
	session := r.sessions[sessionID]
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
