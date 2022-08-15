package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

var sessions = make(map[int]Session)
var songs = make(map[int]Song)

func main()  {
	populateSongs()

	router := gin.Default()
	//TODO: Restrict this to only required methods and domains
	router.Use(cors.Default())

	initializeRoutes(router)
	router.Run()
}

