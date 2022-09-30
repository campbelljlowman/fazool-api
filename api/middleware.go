package api

import (
	"fmt"

	"github.com/campbelljlowman/fazool-api/auth"
	"github.com/gin-gonic/gin"
)

func jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bearerToken := c.Request.Header.Get("Authorization")

		id, authLevel, err := auth.VerifyJWT(bearerToken)
		if err != nil {
			println(err.Error())
		}
		
		fmt.Printf("Token variables: ID - %v, Auth - %v", id, authLevel)

	}
}