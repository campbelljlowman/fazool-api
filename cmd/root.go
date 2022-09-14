package cmd

import (
	"github.com/campbelljlowman/fazool-api/api"
)


func Execute()  {
	router := api.InitializeRoutes()
	router.Run()
}
