package api

import (
	"context"

	"github.com/campbelljlowman/fazool-api/auth"
	
	"github.com/gin-gonic/gin"
)

func jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bearerToken := c.Request.Header.Get("Authorization")

		if bearerToken == "" {
			return
		}

		userId, authLevel, err := auth.VerifyJWT(bearerToken)
		if err != nil {
			println(err.Error())
			return
		}
		
		ctx1 := context.WithValue(c.Request.Context(), "user", userId)
		ctx2 := context.WithValue(ctx1, "auth", authLevel)
        c.Request = c.Request.WithContext(ctx2)
        c.Next()
	}
}