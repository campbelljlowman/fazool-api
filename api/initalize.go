package api

import (
	"net/http"


	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/graph"
	"github.com/campbelljlowman/fazool-api/session"
	"github.com/campbelljlowman/fazool-api/payments"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitializeRoutes() *gin.Engine {
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
	stripeService := payments.NewStripeService(accountService)

	r := graph.NewResolver(sessionService, accountService, stripeService)
	srv := graph.NewGraphQLServer(r)

	router.Any("/query", func(c *gin.Context) {
		srv.ServeHTTP(c.Writer, c.Request)
	})

	router.POST("/stripe-webhook", func(c *gin.Context) {
		stripeService.HandleStripeWebhook(c.Writer, c.Request)
	})

	router.GET("/active-session-metrics", sessionService.GetActiveSessionMetrics)
	router.GET("/asm", sessionService.GetActiveSessionMetrics)
	router.GET("/completed-session-metrics", sessionService.GetCompletedSessionMetrics)
	router.GET("/csm", sessionService.GetCompletedSessionMetrics)

	return router
}

func healthCheck(c *gin.Context) {
	c.String(http.StatusOK, "API is healthy!")
}
