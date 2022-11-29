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

		userId, accountLevel, voterLevel, err := auth.VerifyJWT(bearerToken)
		if err != nil {
			slog.Warn("Couldn't verify JWT token", "error", err.Error())
			return
		}
		
		ctx1 := context.WithValue(c.Request.Context(), "user", userId)
		ctx2 := context.WithValue(ctx1, "account-level", accountLevel)
		ctx3 := context.WithValue(ctx2, "voter-level", voterLevel)
        c.Request = c.Request.WithContext(ctx3)
        c.Next()
	}
}