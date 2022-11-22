package graph

//go:generate go get github.com/99designs/gqlgen@v0.17.15 && go run github.com/99designs/gqlgen generate
import (
	"github.com/campbelljlowman/fazool-api/session"
	
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	sessions 		map[int]*session.Session
	postgresClient *pgxpool.Pool
	redisClient    *redis.Client
}

func NewResolver(pgClient *pgxpool.Pool, redisClient *redis.Client) *Resolver {
	return &Resolver{
		sessions:      	make(map[int]*session.Session),
		postgresClient: pgClient,
		redisClient:    redisClient,
	}
}