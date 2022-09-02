package api

import(
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/campbelljlowman/fazool-api/graph"
)

var sessions = make(map[int]*Session)
var songs = make(map[int]Song)


func InitializeRoutes(router *gin.Engine){
	router.GET("/hc", healthCheck)

	// Graphql
	router.GET("/playground", func(c *gin.Context) {
		playground.Handler("GraphQL", "/query").ServeHTTP(c.Writer, c.Request)
	})
	srv := graph.NewGraphQLServer(&graph.Resolver{})
	router.Any("/query", func(c *gin.Context) {
		srv.ServeHTTP(c.Writer, c.Request)
	})

	// Sessions
	router.POST("/session", createNewSession)
	router.GET("/session", getAllSessions)
	router.GET("/session/:sessionID", getSessionFromID)
	router.POST("/session/:sessionID", updateQueue)
	router.PATCH("/session/:sessionID", updateCurrentlyPlaying)
}


func healthCheck(c *gin.Context){
	c.String(http.StatusOK, "API is healthy!")
}