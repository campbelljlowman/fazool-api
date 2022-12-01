package api

import (
	"context"

	"golang.org/x/exp/slog"
	
	"github.com/campbelljlowman/fazool-api/auth"

	"github.com/gin-gonic/gin"
)

func jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bearerToken := c.Request.Header.Get("Authorization")

		if bearerToken == "" {
			return
		}

		userId, err := auth.VerifyJWT(bearerToken)
		if err != nil {
			slog.Warn("Couldn't verify JWT token", "error", err.Error())
			return
		}
		
		ctx1 := context.WithValue(c.Request.Context(), "user", userId)
        c.Request = c.Request.WithContext(ctx1)
        c.Next()
	}
}