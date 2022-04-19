package main

import (
	// "net/http"

	"github.com/gin-gonic/gin"
)

var sessions  map[int]Session

func main()  {
	router := gin.Default()
	initializeRoutes(router)
	router.Run()
}

