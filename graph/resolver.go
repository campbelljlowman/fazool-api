package graph

//go:generate go get github.com/99designs/gqlgen@v0.17.15 && go run github.com/99designs/gqlgen generate
import (
	"sync"

	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/jackc/pgx/v4/pgxpool"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	sessions 			map[int]*model.Session
	channels 			map[int] []chan *model.Session
	PostgresClient 		*pgxpool.Pool
	mutex 				sync.Mutex
}

func NewResolver (client *pgxpool.Pool) *Resolver {
	return &Resolver{
		sessions:			make(map[int]*model.Session),
		channels:			make(map[int][]chan *model.Session),
		PostgresClient: 	client,
		mutex: 				sync.Mutex{},
	}
}