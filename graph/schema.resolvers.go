package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/campbelljlowman/fazool-api/graph/generated"
	"github.com/campbelljlowman/fazool-api/graph/model"
)

// CreateSession is the resolver for the createSession field.
func (r *mutationResolver) CreateSession(ctx context.Context) (*model.Session, error) {
	if r.sessions == nil {
		r.sessions = make(map[int]*model.Session)
	}

	session := &model.Session{
		ID:               81,
		CurrentlyPlaying: nil,
		Queue:            nil,
	}

	r.sessions[session.ID] = session
	return session, nil
}

// UpdateQueue is the resolver for the updateQueue field.
func (r *mutationResolver) UpdateQueue(ctx context.Context, sessionID *int, song *model.SongUpdate) (*model.Session, error) {
	session := r.sessions[*sessionID]


	idx := slices.IndexFunc(session.Queue, func(s *model.Song) bool { return s.ID == song.ID })
	if (idx == -1){
		// add new song to queue
		newSong := &model.Song{
			ID: song.ID,
			Title: *song.Title,
			Artist: *song.Artist,
			Image: *song.Image,
			Votes: song.Vote,
		}
		session.Queue = append(session.Queue, newSong)
	} else{
		queuedSong := session.Queue[idx]
		queuedSong.Votes += song.Vote
	}

	print(session.Queue[idx])
	return session, nil
}

// UpdateCurrentlyPlaying is the resolver for the updateCurrentlyPlaying field.
func (r *mutationResolver) UpdateCurrentlyPlaying(ctx context.Context, sessionID *int, action *model.QueueAction) (*model.Session, error) {
	panic(fmt.Errorf("not implemented: UpdateCurrentlyPlaying - updateCurrentlyPlaying"))
}

// Session is the resolver for the session field.
func (r *queryResolver) Session(ctx context.Context, session *int) ([]*model.Session, error) {
	panic(fmt.Errorf("not implemented: Session - session"))
	//Code here
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
