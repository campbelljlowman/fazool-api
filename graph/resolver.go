package graph

import (
	"sync"

	"github.com/campbelljlowman/fazool-api/database"
	"github.com/campbelljlowman/fazool-api/musicplayer"
	"github.com/campbelljlowman/fazool-api/session"
	// "golang.org/x/exp/slog"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Sessions get watched at this frequency in seconds
// const sessionWatchFrequency time.Duration = 10

type Resolver struct {
	musicPlayers	map[int]musicplayer.MusicPlayer
	sessionsMutex 	*sync.Mutex
	database 		database.Database
	// TODO: Change this to cache and wrap Redis in an interface
	session		    *session.Session
}

func NewResolver(database database.Database, session *session.Session) *Resolver {
	return &Resolver{
		musicPlayers: 	make(map[int]musicplayer.MusicPlayer),
		sessionsMutex:  &sync.Mutex{},
		database: 		database,
		session: session,	
	}
}

// func (r *Resolver) WatchSessions() {
// 	for {
// 		r.sessionsMutex.Lock()

// 		for _, session := range r.sessions{
// 			if r.session.IsSessionExpired(session.SessionInfo.ID) {
// 				r.endSession(session)
// 			}
// 		}
// 		r.sessionsMutex.Unlock()
// 		time.Sleep(sessionWatchFrequency * time.Second)
// 	}
// }

// func (r *Resolver) endSession(session *session.Session) error {
// 	slog.Info("Ending session!", "sessionID", session.SessionInfo.ID)
// 	err := r.database.SetAccountSession(session.SessionInfo.Admin, 0)
// 	if err != nil {
// 		return err
// 	}

// 	delete(r.sessions, session.SessionInfo.ID)

// 	r.session.ExpireSession(session.SessionInfo.ID)
	
// 	return nil
// }

// func (r *Resolver) getSession(sessionID int) (*session.Session, bool) {
// 	r.sessionsMutex.Lock()
// 	session, exists := r.sessions[sessionID]
// 	r.sessionsMutex.Unlock()
// 	return session, exists
// }