package api

import(
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/campbelljlowman/fazool-api/graph"
	"github.com/campbelljlowman/fazool-api/graph/generated"
)

var sessions = make(map[int]*Session)
var songs = make(map[int]Song)


func InitializeRoutes(router *gin.Engine){
	router.GET("/hc", healthCheck)

	// Graphql
	router.GET("/playground", func(c *gin.Context) {
		playground.Handler("GraphQL", "/query").ServeHTTP(c.Writer, c.Request)
	})
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))
	router.POST("/query", func(c *gin.Context) {
		srv.ServeHTTP(c.Writer, c.Request)
	})

	// Sessions
	router.POST("/session", createNewSession)
	router.GET("/session", getAllSessions)
	router.GET("/session/:sessionID", getSessionFromID)
	// TODO: Change this function to updateQueue, have it add a song to the queue if it's not in it, otherwise increment the vote
	// Takes song id and vote increment as required, optionally name, artist and album to create a new song object

	router.POST("/session/:sessionID", updateQueue)
	router.PATCH("/session/:sessionID", updateCurrentlyPlaying)
}


func healthCheck(c *gin.Context){
	c.String(http.StatusOK, "API is healthy!")
}