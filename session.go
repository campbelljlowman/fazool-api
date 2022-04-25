package main

import (
	// "encoding/json"
	// "fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Session struct {
	SessionID int `json:"sessionID"`
	CurrentlyPlaying Song `json:"currentlyPlaying"`
	Queue map[int] Song `json:"queue"`
}

type Vote struct{
	SongId int `json:"songID"`
	Vote int `json:"vote"`
}

// Handlers
func createNewSession(c *gin.Context){
	sessionID := getNewSessionId()
	new_session := Session{
		SessionID: sessionID,
		Queue: make(map[int]Song),
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
	if err != nil{
		// TODO: Error handling
	} 
	c.JSON(http.StatusOK, session)
}

func voteForSong(c *gin.Context){
	// Get sesion ID from URL
	sessionIDString := c.Param("sessionID")
	sessionID, err := strconv.Atoi(sessionIDString)
	if err != nil{
		// TODO: return error for incorrect param
	}

	// Get vote from JSON body
	newVote := Vote{}
	c.BindJSON(&newVote)

	// Get session from memory using ID
	session, key_exists := sessions[sessionID]
	if !key_exists{
		// error if queue doesn't exist
	}

	// Add vote to queue
	// Get copy of song voted on
	song := songs[newVote.SongId]
	var currentVotes int
	// Get current number of votes song has
	_, songExistsInQueue := session.Queue[newVote.SongId]
	if songExistsInQueue{
		currentVotes = session.Queue[newVote.SongId].Votes
	} else {
		currentVotes = 0
	}
	song.Votes = currentVotes + newVote.Vote
	// Store new song in queue
	session.Queue[newVote.SongId] = song

	c.Status(http.StatusOK)
}

func currentlyPlayingAction(c *gin.Context){
	
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