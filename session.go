package main

import (
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Session struct {
	SessionID int `json:"session_id"`
	CurrentlyPlaying Song `json:"currently_playing"`
	Queue map[int] Song `json:"queue"`
}

type Vote struct{
	SongId int `json:"song_id"`
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
	// Describe expected JSON
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

func updateCurrentlyPlaying(c *gin.Context){
	// update currently playing song
	// Advance to next song (highest in queue), play or pause
	// TODO: Describe expected JSON
	// Options are strings: advance, play, pause

	// Get sesion ID from URL
	sessionIDString := c.Param("sessionID")
	sessionID, err := strconv.Atoi(sessionIDString)
	if err != nil{
		// TODO: return error for incorrect param
	}

	songAction := SongAction{}
	c.BindJSON(&songAction)
	action := songAction.Action

	switch action {
	case "advance":
		// Move song with most votes to next playing
		advanceSong(sessionID)
	case "play":
		// call play song on player plugin
	case "pause":
		// call pause song on player plugin
	default:
		// return error for invalid action being called
	}
	c.Status(http.StatusOK)

}

// Helpers
func getNewSessionId() int {
	// TODO: Make not recursive (Use UUID?)
	// TODO: Make session ID range more like kahoot (all same # of digits) 
	newSessionId := rand.Intn(1000)
	_, key_exists := sessions[newSessionId]
	if key_exists {
		return getNewSessionId()
	} else {
		return newSessionId
	}
}

func advanceSong(sessionID int) {
	session := sessions[sessionID]

	queue := session.Queue
	var highestVotes int
	var highestVotesKey int
	// TODO: make sure default value is not 0
	for key, element := range queue{
		if (highestVotes < element.Votes) {
			highestVotes = element.Votes
			highestVotesKey = key
		} 
	}
	session.CurrentlyPlaying = session.Queue[highestVotesKey]
	sessions[sessionID] = session
	delete(session.Queue, highestVotesKey)
}
