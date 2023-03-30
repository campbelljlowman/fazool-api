package api

import (
	"net/http"

	"golang.org/x/exp/slog"

	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/graph"
	"github.com/campbelljlowman/fazool-api/session"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitializeRoutes() *gin.Engine {
	slog.Debug("Initializing routes")
	router := gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "AccountAuthentication")
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "VoterAuthentication")
	router.Use(cors.New(corsConfig))
	router.Use(getAccountIDMiddleware())
	router.Use(getVoterIDMiddleware())

	router.GET("/hc", healthCheck)

	// Playground - not needed for prod
	router.GET("/playground", func(c *gin.Context) {
		playground.Handler("GraphQL", "/query").ServeHTTP(c.Writer, c.Request)
	})

	accountService := account.NewAccountServiceGormImpl()
	sessionService := session.NewSessionServiceInMemoryImpl(accountService)
	r := graph.NewResolver(sessionService, accountService)
	srv := graph.NewGraphQLServer(r)

	router.Any("/query", func(c *gin.Context) {
		srv.ServeHTTP(c.Writer, c.Request)
	})

	return router
}

func healthCheck(c *gin.Context) {
	c.String(http.StatusOK, "API is healthy!")
}
