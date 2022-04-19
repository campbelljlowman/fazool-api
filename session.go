package main

import(
	"math/rand"
	"github.com/gin-gonic/gin"
)

type Session struct {
	id int
	currently_playing Song
	queue map[int] Song
}

func createNewSession(c *gin.Context){
	sessionId := getNewSessionId()
}

func getNewSessionId(){
	newSessionId := rand.Intn(1000)
	_, key_exists := sessions[newSessionId]
	if key_exists {
		return getNewSessionId()
	}
	return newSessionId
}