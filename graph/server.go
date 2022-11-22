
package graph

import (
	"net/http"
	"time"

	"github.com/campbelljlowman/fazool-api/graph/generated"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"
)

func NewGraphQLServer(resolver *Resolver) *handler.Server {
	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))
	
	// Configure WebSocket with CORS
	srv.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		KeepAlivePingInterval: 10 * time.Second,
	})

	// srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	// srv.AddTransport(transport.MultipartForm{})

	// srv.SetQueryCache(lru.New(1000))

	// srv.Use(extension.Introspection{})
	// srv.Use(extension.AutomaticPersistedQuery{
	// 	Cache: lru.New(100),
	// })

	return srv
}