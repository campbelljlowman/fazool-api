package api

import "github.com/gin-gonic/gin"

func jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bearerToken := c.Request.Header.Get("Authorization")
		println(bearerToken)
	}
}