package graph

//go:generate go get github.com/99designs/gqlgen@v0.17.15 && go run github.com/99designs/gqlgen generate
import (
	"sync"

	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/zmb3/spotify/v2"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	sessions 			map[int]*model.Session
	channels 			map[int] []chan *model.Session
	// TODO: Make this map of interfaces
	// musicPlayers		map[int] *musicplayer.MusicPlayer 		
	spotifyPlayers		map[int] *spotify.Client	
	// TODO: Make this lower case
	PostgresClient 		*pgxpool.Pool
	channelMutex 		sync.Mutex
	queueMutex			sync.Mutex
}

func NewResolver (client *pgxpool.Pool) *Resolver {
	return &Resolver{
		sessions:			make(map[int]*model.Session),
		channels:			make(map[int][]chan *model.Session),
		spotifyPlayers:		make(map[int] *spotify.Client),	
		PostgresClient: 	client,
		channelMutex: 		sync.Mutex{},
		queueMutex: 		sync.Mutex{},	
	}
}