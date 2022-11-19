package api

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/campbelljlowman/fazool-api/graph"
	"github.com/campbelljlowman/fazool-api/database"
	"github.com/campbelljlowman/fazool-api/cache"
)

func InitializeRoutes() *gin.Engine {
	router := gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "authorization")
	router.Use(cors.New(corsConfig))
	router.Use(jwtAuthMiddleware())

	router.GET("/hc", healthCheck)

	// Playground - not needed for prod
	router.GET("/playground", func(c *gin.Context) {
		playground.Handler("GraphQL", "/query").ServeHTTP(c.Writer, c.Request)
	})

	pgClient := database.NewPostgresClient()
	redisClient := cache.GetRedisClient()
	r := graph.NewResolver(pgClient, redisClient)
	srv := graph.NewGraphQLServer(r)
	router.Any("/query", func(c *gin.Context) {
		srv.ServeHTTP(c.Writer, c.Request)
	})

	return router
}

func healthCheck(c *gin.Context) {
	c.String(http.StatusOK, "API is healthy!")
}
