package api

import (
	"fmt"

	"github.com/campbelljlowman/fazool-api/auth"
	"github.com/gin-gonic/gin"
)

func jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bearerToken := c.Request.Header.Get("Authorization")

		userId, authLevel, err := auth.VerifyJWT(bearerToken)
		if err != nil {
			println(err.Error())
			return
		}
		
		fmt.Printf("Token variables: user - %v, Auth - %v", userId, authLevel)

		c.Set("user", userId)
		// c.Set("auth", authLevel)
		fmt.Printf("\nMiddleware context: %v\n", c)
        c.Next()
	}
}