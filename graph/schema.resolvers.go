package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
	
	"github.com/campbelljlowman/fazool-api/auth"
	"github.com/campbelljlowman/fazool-api/graph/generated"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/session"
	"github.com/campbelljlowman/fazool-api/spotifyUtil"
	"github.com/campbelljlowman/fazool-api/utils"

	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

// CreateSession is the resolver for the createSession field.
func (r *mutationResolver) CreateSession(ctx context.Context) (*model.User, error) {
	userID, _ := ctx.Value("user").(int)
	// TODO: Make session ID random - use UUID
	sessionID := 81

	session := session.NewSession()
	r.sessions[sessionID] = &session

	// Create session info
	sessionInfo := &model.SessionInfo{
		ID:               sessionID,
		CurrentlyPlaying: nil,
		Queue:            nil,
	}
	session.SessionInfo = sessionInfo

	err := r.database.SetUserSession(userID, sessionID)

	refreshToken, err := r.database.GetSpotifyRefreshToken(userID)
	if err != nil {
		return nil, errors.New("Error get Spotify refresh token")
	}

	spotifyToken, err := spotifyUtil.RefreshToken(userID, refreshToken)
	if err != nil {
		return nil, errors.New("Error refreshing Spotify Token")
	}

	err = r.database.SetSpotifyAccessToken(userID, spotifyToken)
	if err != nil {
		return nil, errors.New("Error setting new Spotify token")
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
	session := r.sessions[sessionID]

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
			Votes:  song.Vote,
		}
		session.SessionInfo.Queue = append(session.SessionInfo.Queue, newSong)
	} else {
		queuedSong := session.SessionInfo.Queue[idx]
		queuedSong.Votes += song.Vote
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
func (r *mutationResolver) CreateUser(ctx context.Context, newUser model.NewUser) (*model.Token, error) {
	// Get this from new user request!
	authLevel := 1
	vaildEmail := utils.ValidateEmail(newUser.Email)
	if !vaildEmail {
		return nil, errors.New("Invalid email format")
	}

	emailExists := r.database.CheckIfEmailExists(newUser.Email)

	if emailExists {
		return nil, errors.New("User with this email already exists!")
	}
	// TODO: Salt password and use Argon2id
	passwordHash := utils.HashHelper(newUser.Password)

	userID, err := r.database.AddUserToDatabase(newUser, passwordHash, authLevel)

	if err != nil {
		println("Error adding user to database")
		println(err.Error())
		return nil, errors.New("Error adding user to database")
	}

	token, err := auth.GenerateJWT(userID, authLevel)
	if err != nil {
		println("Error creating user token")
		println(err.Error())
		return nil, errors.New("Error creating user token")
	}

	returnValue := &model.Token{
		Jwt: token,
	}

	return returnValue, nil
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, userLogin model.UserLogin) (*model.Token, error) {
	userID, authLevel, password, err := r.database.GetUserLoginValues(userLogin.Email)
	if err != nil {
		return nil, err
	}

	if utils.HashHelper(userLogin.Password) != password {
		return nil, errors.New("Invalid Login Credentials!")
	}

	token, err := auth.GenerateJWT(userID, authLevel)
	if err != nil {
		return nil, errors.New("Error creating user token")
	}

	returnValue := &model.Token{
		Jwt: token,
	}

	return returnValue, nil
}

// UpsertSpotifyToken is the resolver for the upsertSpotifyToken field.
func (r *mutationResolver) UpsertSpotifyToken(ctx context.Context, spotifyCreds model.SpotifyCreds) (*model.User, error) {
	userID, _ := ctx.Value("user").(int)

	err := r.database.SetSpotifyAccessToken(userID, spotifyCreds.AccessToken)

	if err != nil {
		return nil, err
	}

	err = r.database.SetSpotifyRefreshToken(userID, spotifyCreds.RefreshToken)

	if err != nil {
		return nil, err
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
	if exists {
		return session.SessionInfo, nil
	} else {
		return nil, errors.New("Session not found!")
	}
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context) (*model.User, error) {
	userID, _ := ctx.Value("user").(int)

	user, err := r.database.GetUserByID(userID)
	if err != nil {
		println(err.Error())
		return nil, err
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

// SessionUpdated is the resolver for the sessionUpdated field.
func (r *subscriptionResolver) SessionUpdated(ctx context.Context, sessionID int) (<-chan *model.SessionInfo, error) {
	session := r.sessions[sessionID]
	channel := make(chan *model.SessionInfo)

	session.ChannelMutex.Lock()
	session.Channels = append(session.Channels, channel)
	session.ChannelMutex.Unlock()

	return channel, nil

	// TODO: Cleanup channel?
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