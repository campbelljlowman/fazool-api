package main

import(
	"net/http"

	"github.com/gin-gonic/gin"
)

func initializeRoutes(router *gin.Engine){
	router.GET("/hc", healthCheck)

	// Sessions
	router.POST("/session", createNewSession)
	router.GET("/session", getAllSessions)
	router.GET("/session/:sessionID", getSessionFromID)
}


func healthCheck(c *gin.Context){
	c.String(http.StatusOK, "API is healthy!")
}