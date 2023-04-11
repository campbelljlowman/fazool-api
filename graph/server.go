package graph

import (
	"context"
	"net/http"
	"time"
	"errors"

	"github.com/campbelljlowman/fazool-api/graph/generated"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"
)

func NewGraphQLServer(resolver *Resolver) *handler.Server {
	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))
	
	srv.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		KeepAlivePingInterval: 10 * time.Second,
		InitFunc: func(ctx context.Context, initalPayload transport.InitPayload) (context.Context, error) {
			return authenticateSubscription(ctx, initalPayload)
		},
		
	})
	srv.Use(extension.Introspection{})


	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	return srv
}

func authenticateSubscription(ctx context.Context, initalPayload transport.InitPayload) (context.Context, error) {
	authToken, ok := initalPayload["SubscriptionAuthentication"].(string)
	if !ok || authToken == "" {
		return nil, errors.New("authToken not found in transport payload")
	}

	if authToken != "Subscription-Allowed" {
		return nil, errors.New("authToken not found in transport payload")
	}

	// TODO: Once voter token is passed dynamically, check the value
	ctxNew := context.WithValue(ctx, "voterID", "ID1")
	return ctxNew, nil
}