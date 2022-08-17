package api

import(
	"net/http"

	"github.com/gin-gonic/gin"
)

var sessions = make(map[int]Session)
var songs = make(map[int]Song)


func InitializeRoutes(router *gin.Engine){
	router.GET("/hc", healthCheck)

	// Sessions
	router.POST("/session", createNewSession)
	router.GET("/session", getAllSessions)
	router.GET("/session/:sessionID", getSessionFromID)
	// TODO: Change this function to updateQueue, have it add a song to the queue if it's not in it, otherwise increment the vote
	// Takes song id and vote increment as required, optionally name, artist and album to create a new song object

	// First change data structures to use pointers instead
	router.PUT("/session/:sessionID", voteForSong)
	router.PATCH("/session/:sessionID", updateCurrentlyPlaying)
}


func healthCheck(c *gin.Context){
	c.String(http.StatusOK, "API is healthy!")
}