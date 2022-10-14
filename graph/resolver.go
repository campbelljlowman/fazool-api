package graph

//go:generate go get github.com/99designs/gqlgen@v0.17.15 && go run github.com/99designs/gqlgen generate
import (
	"context"
	"fmt"
	"sync"

	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/musicPlayer"
	"github.com/campbelljlowman/fazool-api/spotify"
	"github.com/jackc/pgx/v4/pgxpool"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	sessions 			map[int]*model.Session
	channels 			map[int] []chan *model.Session
	// TODO: Make this map of interfaces
	// musicPlayers		map[int] *musicplayer.MusicPlayer 		
	spotifyPlayers		map[int] *spotify.SpotifyClient	
	// TODO: Make this lower case
	PostgresClient 		*pgxpool.Pool
	mutex 				sync.Mutex
}

func NewResolver (client *pgxpool.Pool) *Resolver {
	// TODO: Migrations? Make tables if they don't exist?
	// Not sure if this is the right place for this but it needs to happen somewhere
	queryString := fmt.Sprintf(`
	UPDATE public.user
	SET session_id = 0
	`)
	commandTag, err := client.Exec(context.Background(), queryString)
	if err != nil {
		print("Error adding new session to database")
	}
	if commandTag.RowsAffected() != 1 {
		print("No user found to update")
	}
	
	return &Resolver{
		sessions:			make(map[int]*model.Session),
		channels:			make(map[int][]chan *model.Session),
		spotifyPlayers:		make(map[int] *spotify.SpotifyClient),	
		PostgresClient: 	client,
		mutex: 				sync.Mutex{},
	}
}