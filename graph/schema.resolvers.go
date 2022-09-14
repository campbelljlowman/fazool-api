package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"sort"

	"github.com/campbelljlowman/fazool-api/graph/generated"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// CreateSession is the resolver for the createSession field.
func (r *mutationResolver) CreateSession(ctx context.Context) (*model.Session, error) {
	// TODO: Make session ID random
	sessionID := 81

	session := &model.Session{
		ID:               sessionID,
		CurrentlyPlaying: nil,
		Queue:            nil,
	}
	r.sessions[session.ID] = session

	return session, nil
}

// UpdateQueue is the resolver for the updateQueue field.
func (r *mutationResolver) UpdateQueue(ctx context.Context, sessionID int, song model.SongUpdate) (*model.Session, error) {
	session := r.sessions[sessionID]

	idx := slices.IndexFunc(session.Queue, func(s *model.Song) bool { return s.ID == song.ID })
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

	// Update subscription
	go func() {
		r.mutex.Lock()
		channels := r.channels[sessionID]
		r.mutex.Unlock()
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
	panic(fmt.Errorf("not implemented: UpdateCurrentlyPlaying - updateCurrentlyPlaying"))
}

// CreateUser is the resolver for the createUser field.
func (r *mutationResolver) CreateUser(ctx context.Context, newUser model.NewUser) (*model.User, error) {
	print(newUser.FirstName)
	user := &model.User{
		ID:        "1",
		FirstName: &newUser.FirstName,
		LastName:  &newUser.LastName,
		Email:     &newUser.Email,
	}
	return user, nil
}

// Session is the resolver for the session field.
func (r *queryResolver) Session(ctx context.Context, sessionID *int) ([]*model.Session, error) {
	if sessionID != nil {
		session := r.sessions[*sessionID]
		return []*model.Session{session}, nil
	} else {
		return maps.Values(r.sessions), nil
	}
}

// SessionUpdated is the resolver for the sessionUpdated field.
func (r *subscriptionResolver) SessionUpdated(ctx context.Context, sessionID int) (<-chan *model.Session, error) {
	channel := make(chan *model.Session)

	r.mutex.Lock()
	r.channels[sessionID] = append(r.channels[sessionID], channel)
	r.mutex.Unlock()

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
