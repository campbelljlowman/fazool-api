package main

import (
	// "net/http"

	"github.com/gin-gonic/gin"
)

var sessions = make(map[int]Session)
var songs = make(map[int]Song)

func main()  {
	populateSongs()

	router := gin.Default()
	initializeRoutes(router)
	router.Run()
}

