package graph

//go:generate go get github.com/99designs/gqlgen@v0.17.15 && go run github.com/99designs/gqlgen generate
import (
	"sync"
	"time"

	"github.com/campbelljlowman/fazool-api/database"
	"github.com/campbelljlowman/fazool-api/session"
	"golang.org/x/exp/slog"

	"github.com/go-redis/redis/v8"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	sessions 		map[int]*session.Session
	sessionsMutex 	*sync.Mutex
	database 		database.Database
	// TODO: Change this to cache and wrap Redis in an interface
	redisClient    *redis.Client
}

func NewResolver(database database.Database, redisClient *redis.Client) *Resolver {
	return &Resolver{
		sessions:      	make(map[int]*session.Session),
		sessionsMutex:  &sync.Mutex{},
		database: 		database,
		redisClient:    redisClient,
	}
}

func (r *Resolver) WatchSessions() {
	for {
		r.sessionsMutex.Lock()

		for _, s := range r.sessions{
			if s.ExpiresAt.Before(time.Now()) {
				r.endSession(s, s.SessionInfo.Admin)
			}
		}
		r.sessionsMutex.Unlock()
		time.Sleep(10 * time.Second)
	}
}

func (r *Resolver) endSession(session *session.Session, userID string) error {
	slog.Info("Ending session!", "sessionID", session.SessionInfo.ID)
	err := r.database.SetUserSession(userID, 0)
	if err != nil {
		return err
	}

	session.ChannelMutex.Lock()
	for _, ch := range session.Channels {
		close(ch)
	}

	delete(r.sessions, session.SessionInfo.ID)

	session.ExpiryMutex.Lock()
	session.ExpiresAt = time.Now()
	session.ExpiryMutex.Unlock()
	
	return nil
}