package main

import (
	// "encoding/json"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Session struct {
	ID int
	CurrentlyPlaying Song
	Queue map[int] Song
}

// Handlers
func createNewSession(c *gin.Context){
	sessionID := getNewSessionId()
	new_session := Session{
		ID: sessionID,
	}
	sessions[sessionID] = new_session
	c.JSON(http.StatusOK, gin.H{"sessionID": sessionID})
}

func getAllSessions(c *gin.Context){
	c.JSON(http.StatusOK, sessions)
}

func getSessionFromID(c *gin.Context){
	sessionIDString := c.Param("sessionID")
	sessionID, err := strconv.Atoi(sessionIDString)
	if err != nil{
		// TODO: return error for incorrect param
	}
	session := sessions[sessionID]
	// sessionJSON, err := json.Marshal(session)
	if err != nil{
		// TODO: Error handling
	} 
	c.JSON(http.StatusOK, session)
	// c.JSON(http.StatusOK, sessionJSON)
}


// Helpers
func getNewSessionId() int {
	// TODO: Make not recursive 
	// TODO: Make session ID range more like kahoot (all same # of digits)
	newSessionId := rand.Intn(1000)
	_, key_exists := sessions[newSessionId]
	if key_exists {
		return getNewSessionId()
	} else {
		return newSessionId
	}
}