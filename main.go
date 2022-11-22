package main

import (
	"github.com/campbelljlowman/fazool-api/api"
)

func main() {
	
	router := api.InitializeRoutes()
	router.Run()
}