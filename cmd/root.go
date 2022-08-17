package cmd

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	
	"github.com/campbelljlowman/fazool-api/api"

)


func Execute()  {
	api.PopulateSongs()

	router := gin.Default()
	//TODO: Restrict this to only required methods and domains
	router.Use(cors.Default())

	api.InitializeRoutes(router)
	router.Run()
}
