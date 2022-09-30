package api

import (
	"fmt"

	"github.com/campbelljlowman/fazool-api/auth"
	"github.com/gin-gonic/gin"
)

func jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bearerToken := c.Request.Header.Get("Authorization")
		fmt.Printf("Token: %v", bearerToken)

		id, authLevel, err := auth.VerifyJWT(bearerToken)
		if err != nil {
			println(err.Error())
		}
		
		_ = fmt.Sprintf("Token variables: ID - %v, Auth - %v", id, authLevel)

	}
}