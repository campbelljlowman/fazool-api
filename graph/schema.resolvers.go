package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/campbelljlowman/fazool-api/auth"
	"github.com/campbelljlowman/fazool-api/database"
	"github.com/campbelljlowman/fazool-api/graph/generated"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/spotifyUtil"
	"github.com/campbelljlowman/fazool-api/utils"
	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/exp/slices"
	"golang.org/x/oauth2"
)

// CreateSession is the resolver for the createSession field.
func (r *mutationResolver) CreateSession(ctx context.Context, userID int) (*model.User, error) {
	// TODO: Get userID from context
	// 	userID, _ := ctx.Value("user").(int)
	// TODO: Make session ID random - use UUID
	sessionID := 81

	// Create session
	session := &model.Session{
		ID:               sessionID,
		CurrentlyPlaying: nil,
		Queue:            nil,
	}
	r.sessions[session.ID] = session

	queryString := fmt.Sprintf(`
		UPDATE public.user
		SET session_id = %v
		WHERE user_id = %v;`, sessionID, userID)

	commandTag, err := r.PostgresClient.Exec(context.Background(), queryString)

	if err != nil {
		return nil, errors.New("Error adding new session to database")
	}
	if commandTag.RowsAffected() != 1 {
		return nil, errors.New("No user found to update")
	}

	spotifyToken, err := spotifyUtil.RefreshToken(r.PostgresClient, userID)
	if err != nil {
		return nil, errors.New("Error adding new session to database")
	}

	// TODO: Use refresh token as well? https://pkg.go.dev/golang.org/x/oauth2#Token
	token := &oauth2.Token{
		AccessToken: spotifyToken,
	}
	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)
	r.spotifyPlayers[sessionID] = client

	go watchCurrentlyPlaying(r, sessionID)

	user := &model.User{
		ID:        userID,
		SessionID: &sessionID,
	}

	return user, nil
}

// UpdateQueue is the resolver for the updateQueue field.
func (r *mutationResolver) UpdateQueue(ctx context.Context, sessionID int, song model.SongUpdate) (*model.Session, error) {
	session := r.sessions[sessionID]
	println("currently playing: ", session.CurrentlyPlaying.Artist)
	idx := slices.IndexFunc(session.Queue, func(s *model.Song) bool { return s.ID == song.ID })
	r.queueMutex.Lock()
	if idx == -1 {
		// add new song to queue
		newSong := &model.Song{
			ID:     song.ID,
			Title:  *song.Title,
			Artist: *song.Artist,
			Image:  *song.Image,
			Votes:  song.Vote,
		}
		session.Queue = append(session.Queue, newSong)
	} else {
		queuedSong := session.Queue[idx]
		queuedSong.Votes += song.Vote
	}

	// Sort queue
	sort.Slice(session.Queue, func(i, j int) bool { return session.Queue[i].Votes > session.Queue[j].Votes })
	r.queueMutex.Unlock()

	// Update subscription
	go func() {
		r.channelMutex.Lock()
		channels := r.channels[sessionID]
		r.channelMutex.Unlock()
		for _, ch := range channels {
			select {
			case ch <- session: // This is the actual send.
				// Our message went through, do nothing
			default: // This is run when our send does not work.
				fmt.Println("Channel closed in update.")
				// You can handle any deregistration of the channel here.
			}
		}
	}()

	return session, nil
}

// UpdateCurrentlyPlaying is the resolver for the updateCurrentlyPlaying field.
func (r *mutationResolver) UpdateCurrentlyPlaying(ctx context.Context, sessionID int, action model.QueueAction) (*model.Session, error) {
	spotifyClient := r.spotifyPlayers[sessionID]
	switch action {
	case "PLAY":
		spotifyClient.Play(ctx)
	case "PAUSE":
		spotifyClient.Pause(ctx)
	case "ADVANCE":
		print("Advance")
		// TODO: This should use the same advance code as in the queue watcher
		// spotifyClient.Advance("Next Song")
	}

	return r.sessions[sessionID], nil
}

// CreateUser is the resolver for the createUser field.
func (r *mutationResolver) CreateUser(ctx context.Context, newUser model.NewUser) (*model.Token, error) {
	// Get this from new user request!
	authLevel := 1
	vaildEmail := utils.ValidateEmail(newUser.Email)
	if !vaildEmail {
		return nil, errors.New("Invalid email format")
	}

	checkEmailQueryString := fmt.Sprintf("SELECT exists (SELECT 1 FROM public.user WHERE email = '%v' LIMIT 1);", newUser.Email)
	var emailExists bool
	r.PostgresClient.QueryRow(context.Background(), checkEmailQueryString).Scan(&emailExists)
	if emailExists {
		return nil, errors.New("Email already exists!")
	}
	// TODO: Salt password and use Argon2id
	passwordHash := utils.HashHelper(newUser.Password)

	// TODO: Add this to database helpers
	newUserQueryString := fmt.Sprintf(`
		INSERT INTO public.user(first_name, last_name, email, pass_hash, auth_level)
		VALUES ('%v', '%v', '%v', '%v', '%v')
		RETURNING user_id;`,
		newUser.FirstName, newUser.LastName, newUser.Email, passwordHash, authLevel)

	var userID int
	err := r.PostgresClient.QueryRow(context.Background(), newUserQueryString).Scan(&userID)
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
	userID, authLevel, password, err := database.GetUserLoginValues(r.PostgresClient, userLogin.Email)
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

// UpdateSpotifyToken is the resolver for the updateSpotifyToken field.
func (r *mutationResolver) UpdateSpotifyToken(ctx context.Context, spotifyCreds model.SpotifyCreds) (*model.User, error) {
	userID, _ := ctx.Value("user").(int)

	err := database.SetSpotifyAccessToken(r.PostgresClient, userID, spotifyCreds.AccessToken)

	if err != nil {
		return nil, err
	}

	err = database.SetSpotifyRefreshToken(r.PostgresClient, userID, spotifyCreds.RefreshToken)

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
func (r *queryResolver) Session(ctx context.Context, sessionID *int) (*model.Session, error) {
	session, exists := r.sessions[*sessionID]
	if exists {
		return session, nil
	} else {
		return nil, errors.New("Session not found!")
	}
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context) (*model.User, error) {
	userID, _ := ctx.Value("user").(int)

	user, err := database.GetUserByID(r.PostgresClient, userID)
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
func (r *subscriptionResolver) SessionUpdated(ctx context.Context, sessionID int) (<-chan *model.Session, error) {
	channel := make(chan *model.Session)

	r.channelMutex.Lock()
	r.channels[sessionID] = append(r.channels[sessionID], channel)
	r.channelMutex.Unlock()

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


func watchCurrentlyPlaying(r *mutationResolver, sessionID int) {
	client := r.spotifyPlayers[sessionID]
	session := r.sessions[sessionID]
	session.CurrentlyPlaying = &model.CurrentlyPlayingSong{}
	sendUpdate := false
	addNextSong := false
	var song *model.Song

	for {
		sendUpdate = false
		playerState, err := client.PlayerState(context.Background())
		if err != nil {
			fmt.Println(err)
			continue
		}

		if playerState.CurrentlyPlaying.Playing == true {
			if session.CurrentlyPlaying.ID != playerState.CurrentlyPlaying.Item.ID.String() {
				// If song has changed, update currently playing, send update, and set flag to pop next song from queue
				session.CurrentlyPlaying.ID = playerState.CurrentlyPlaying.Item.ID.String()
				session.CurrentlyPlaying.Title = playerState.CurrentlyPlaying.Item.Name
				session.CurrentlyPlaying.Artist = playerState.CurrentlyPlaying.Item.Artists[0].Name
				session.CurrentlyPlaying.Image = playerState.CurrentlyPlaying.Item.Album.Images[0].URL
				session.CurrentlyPlaying.Playing = playerState.CurrentlyPlaying.Playing
				sendUpdate = true
				addNextSong = true
			}

			// If the currently playing song is about to end, pop the top of the session and add to spotify queue
			// If go spotify client adds API for checking current queue, checking this is a better way to tell if it's 
			// Safe to add song. This is a pretty hard requirement for running multiple API instances, although
			// If spotify API is too slow, might have to make addNextSong flag part of session information stored in Redis
			// And cache is refreshed before performing this operaiton
			timeLeft := playerState.CurrentlyPlaying.Item.SimpleTrack.Duration - playerState.CurrentlyPlaying.Progress
			if timeLeft < 5000 && addNextSong{
				r.queueMutex.Lock()
				if len(session.Queue) != 0{
					song, session.Queue = session.Queue[0], session.Queue[1:]
					r.queueMutex.Unlock()

					client.QueueSong(context.Background(), spotify.ID(song.ID))

					sendUpdate = true
					addNextSong = false
				} else {
					// This else block is so we can unlock right after we update the queue in the true condition
					r.queueMutex.Unlock()
				}
			}
		} else {
			// Change currently playing to false if music gets paused
			if session.CurrentlyPlaying.Playing != playerState.CurrentlyPlaying.Playing {
				session.CurrentlyPlaying.Playing = playerState.CurrentlyPlaying.Playing
				sendUpdate = true
			}
		}

		if sendUpdate {
			r.channelMutex.Lock()
			channels := r.channels[sessionID]
			r.channelMutex.Unlock()
	
			for _, ch := range channels {
				select {
				case ch <- session: // This is the actual send.
					// Our message went through, do nothing
				default: // This is run when our send does not work.
					fmt.Println("Channel closed in update.")
					// You can handle any deregistration of the channel here.
				}
			}
		}

		time.Sleep(time.Second)
	}
}
