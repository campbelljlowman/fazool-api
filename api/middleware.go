package api

import (
	"strings"
	"context"

	"golang.org/x/exp/slog"
	
	"github.com/campbelljlowman/fazool-api/auth"

	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
)

func jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authString := c.Request.Header.Get("Authentication")

		if authString == "" {
			slog.Debug("No Authentication header passed on request!")
			return
		}

		var tokenString string
		if len(strings.Split(authString, " ")) == 2 {
			tokenString = strings.Split(authString, " ")[1]
		} else {
			slog.Debug("No value passed after Bearer")
			return
		}

		user := ""
		// Try to parse token as UUID
		_, err := uuid.Parse(tokenString)
		if err == nil {
			user = tokenString
		}

		// Try to parse userID from token
		userID, err := auth.VerifyJWT(tokenString)
		if err == nil {
			user = userID
		}

		if userID == "" {
			slog.Debug("Request made with no user or voter token")
		}

		ctx1 := context.WithValue(c.Request.Context(), "user", user)
		c.Request = c.Request.WithContext(ctx1)
		c.Next()
	}
}