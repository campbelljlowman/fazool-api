package api

import (
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Session struct {
	SessionID int `json:"session_id"`
	CurrentlyPlaying Song `json:"currently_playing"`
	Queue map[string] *Song `json:"queue"`
}

// Handlers
func createNewSession(c *gin.Context){
	sessionID := getNewSessionId()
	new_session := Session{
		SessionID: sessionID,
		Queue: make(map[string]*Song),
	}
	sessions[sessionID] = &new_session
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

func updateQueue(c *gin.Context){
	//TODO: Fix - This errors if called on a queue that doesn't exist
	sessionIDString := c.Param("sessionID")
	sessionID, err := strconv.Atoi(sessionIDString)
	if err != nil{
		// TODO: return error for incorrect param
	}

	votedSong := Song{}
	c.BindJSON(&votedSong)	

	session, key_exists := sessions[sessionID]
	if !key_exists{
		// error if queue doesn't exist
	}

	queuedSong, songExistsInQueue := session.Queue[votedSong.SongID]
	if songExistsInQueue{
		queuedSong.Votes += votedSong.Votes
		print("asdf")
		print(queuedSong.Votes)
	} else {
		//TODO: Validate the values of song
		session.Queue[votedSong.SongID] = &votedSong
	}

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
	// session := sessions[sessionID]

	// queue := session.Queue
	// var highestVotes int
	// var highestVotesKey int
	// // TODO: make sure default value is not 0
	// for key, element := range queue{
	// 	if (highestVotes < element.Votes) {
	// 		highestVotes = element.Votes
	// 		// highestVotesKey = key
	// 	} 
	// }
	// // session.CurrentlyPlaying = session.Queue[highestVotesKey]
	// sessions[sessionID] = session
	// delete(session.Queue, highestVotesKey)
}
