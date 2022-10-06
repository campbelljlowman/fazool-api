package api

import (
	"fmt"
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
		}
		
		fmt.Printf("Token variables: user - %v, Auth - %v", userId, authLevel)

		// TODO: Don't use string for key: https://www.howtographql.com/graphql-go/6-authentication/
		ctx1 := context.WithValue(c.Request.Context(), "user", userId)
		ctx2 := context.WithValue(ctx1, "auth", authLevel)
        c.Request = c.Request.WithContext(ctx2)
        c.Next()
	}
}