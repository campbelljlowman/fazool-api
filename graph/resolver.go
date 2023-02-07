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

// Sessions get watched at this frequency in seconds
const sessionWatchFrequency time.Duration = 10

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

		for _, session := range r.sessions{
			if session.IsExpired() {
				r.endSession(session)
			}
		}
		r.sessionsMutex.Unlock()
		time.Sleep(sessionWatchFrequency * time.Second)
	}
}

func (r *Resolver) endSession(session *session.Session) error {
	slog.Info("Ending session!", "sessionID", session.SessionInfo.ID)
	err := r.database.SetAccountSession(session.SessionInfo.Admin, 0)
	if err != nil {
		return err
	}

	session.CloseChannels()

	delete(r.sessions, session.SessionInfo.ID)

	session.ExpireSession()
	
	return nil
}

func (r *Resolver) getSession(sessionID int) (*session.Session, bool) {
	r.sessionsMutex.Lock()
	session, exists := r.sessions[sessionID]
	r.sessionsMutex.Unlock()
	return session, exists
}