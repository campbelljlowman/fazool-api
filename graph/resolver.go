package graph

//go:generate go get github.com/99designs/gqlgen@v0.17.15 && go run github.com/99designs/gqlgen generate
import (
	"sync"

	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/zmb3/spotify/v2"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	sessions map[int]*model.SessionInfo
	channels map[int][]chan *model.SessionInfo
	// TODO: Make this map of interfaces
	// musicPlayers		map[int] *musicplayer.MusicPlayer
	spotifyPlayers map[int]*spotify.Client
	postgresClient *pgxpool.Pool
	redisClient    *redis.Client
	// These need to be per session!
	channelMutex sync.Mutex
	queueMutex   sync.Mutex
}

func NewResolver(pgClient *pgxpool.Pool, redisClient *redis.Client) *Resolver {
	return &Resolver{
		sessions:       make(map[int]*model.SessionInfo),
		channels:       make(map[int][]chan *model.SessionInfo),
		spotifyPlayers: make(map[int]*spotify.Client),
		postgresClient: pgClient,
		redisClient:    redisClient,
		channelMutex:   sync.Mutex{},
		queueMutex:     sync.Mutex{},
	}
}
