package graph

import (
	"context"
	"net/http"
	"time"

	"github.com/campbelljlowman/fazool-api/graph/generated"
	"github.com/campbelljlowman/fazool-api/utils"
	"golang.org/x/exp/slog"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"
)

func NewGraphQLServer(resolver *Resolver) *handler.Server {
	config := generated.Config{Resolvers: resolver}
	config.Directives.HasVoterID = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		voterID, _ := ctx.Value("voterID").(string)
		if voterID == "" {
			return nil, utils.LogAndReturnError("no voter ID passed to resolver that requires voter ID", nil)
		}

		return next(ctx)
	}

	config.Directives.HasAccountID = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		accountID, _ := ctx.Value("accountID").(int)
		if accountID == 0 {
			return nil, utils.LogAndReturnError("no account ID passed to resolver that requires account ID", nil)
		}

		return next(ctx)
	}

	srv := handler.New(generated.NewExecutableSchema(config))
	
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
	slog.Debug("init payload: ", "p", initalPayload)
	authToken, ok := initalPayload["SubscriptionAuthentication"].(string)
	if !ok || authToken == "" {
		return ctx, nil
	}

	if authToken != "Subscription-Allowed" {
		return ctx, nil
	}

	// TODO: Once voter token is passed dynamically, check the value
	ctxNew := context.WithValue(ctx, "voterID", "ID1")
	return ctxNew, nil
}