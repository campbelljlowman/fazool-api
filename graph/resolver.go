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
