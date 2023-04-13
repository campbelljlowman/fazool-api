package main

import (
	"os"

	"golang.org/x/exp/slog"

	"github.com/campbelljlowman/fazool-api/api"
)

func main() {
	logLevelString := os.Getenv("LOG_LEVEL")
	var logLevel slog.Leveler
	switch logLevelString {
	case "DEBUG":
		logLevel = slog.DebugLevel
	case "INFO":
		logLevel = slog.InfoLevel
	case "ERROR":
		logLevel = slog.ErrorLevel
	}
	h := slog.HandlerOptions{Level: logLevel}.NewJSONHandler(os.Stderr)
	logger := slog.New(h)
	slog.SetDefault(logger)

	router := api.InitializeRoutes()
	router.Run()
}