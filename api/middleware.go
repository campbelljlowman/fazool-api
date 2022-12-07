package api

import (
	"context"

	"golang.org/x/exp/slog"
	
	"github.com/campbelljlowman/fazool-api/auth"

	"github.com/gin-gonic/gin"
)

func jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authString := c.Request.Header.Get("Authentication")

		if authString == "" {
			slog.Debug("No Authentication header passed on request!")
			return
		}

		userID, err := auth.VerifyJWT(authString)
		if err != nil {
			slog.Debug("Couldn't verify JWT token", "error", err.Error())
			userID = authString
		}
		
		ctx1 := context.WithValue(c.Request.Context(), "user", userID)
        c.Request = c.Request.WithContext(ctx1)
        c.Next()
	}
}