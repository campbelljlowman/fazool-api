package main

import (
	"os"

	"golang.org/x/exp/slog"

	"github.com/campbelljlowman/fazool-api/api"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout))
	slog.SetDefault(logger)

	router := api.InitializeRoutes()
	router.Run()
}