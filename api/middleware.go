package api

import (
	"strings"
	"context"
	"errors"

	"golang.org/x/exp/slog"
	
	"github.com/campbelljlowman/fazool-api/auth"
	"github.com/golang-jwt/jwt/v4"

	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
)

func getAccountIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountAuthenticationValue, err := parseAuthenticationHeader("AccountAuthentication", c)
		if err != nil {
			slog.Debug("Account authentication header wasn't parsed", "error", err)
			c.Next()
			return
		}

		accountID, err := auth.GetAccountIDFromJWT(accountAuthenticationValue)
		if errors.Is(err, jwt.ErrTokenExpired) {
			ctx1 := context.WithValue(c.Request.Context(), "accountTokenIsExpired", true)
			c.Request = c.Request.WithContext(ctx1)
			return
		}
		if err != nil {
			slog.Debug("Account authentication passed isn't valid")
			c.Next()
			return
		}

		ctx1 := context.WithValue(c.Request.Context(), "accountID", accountID)
		c.Request = c.Request.WithContext(ctx1)

		c.Next()
	}
}

func getVoterIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
	voterAuthenticationValue, err := parseAuthenticationHeader("VoterAuthentication", c)
	if err != nil {
		slog.Debug("Voter authentication header wasn't parsed", "error", err)
		c.Next()
		return
	}

	_, err = uuid.Parse(voterAuthenticationValue)
	if err != nil {
		slog.Debug("Voter authentication passed isn't a valid")
		c.Next()
		return
	}

	voterID := voterAuthenticationValue

	ctx1 := context.WithValue(c.Request.Context(), "voterID", voterID)
	c.Request = c.Request.WithContext(ctx1)
	c.Next()
	}
}

func parseAuthenticationHeader(header string, c *gin.Context) (string, error) {
	authenticationRaw := c.Request.Header.Get(header)

	if authenticationRaw == "" {
		return "", errors.New("no authentication header passed on request")
	}

	if len(strings.Split(authenticationRaw, " ")) != 2 {
		return "", errors.New("incorrect number of spaces in authentication raw string")
	} 

	authenticationValue := strings.Split(authenticationRaw, " ")[1]	

	return authenticationValue, nil
}