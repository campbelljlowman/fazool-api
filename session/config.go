package session

import (
	"fmt"
	"context"
	"strconv"

	"golang.org/x/exp/slog"

)

func (sc *Session) SetSessionConfig(sessionID, maximumVoters int, adminAccountID string) {
	sc.redisClient.HSet(context.Background(),  getSessionConfigKey(sessionID), "sessionID", sessionID, "maximumVoters", maximumVoters, "adminAccountID", adminAccountID)
}

func (sc *Session) GetSessionAdmin(sessionID int) string {
	sessionMaximumVoters, err := sc.redisClient.HGet(context.Background(), getSessionConfigKey(sessionID), "adminAccountID").Result()
	if err != nil {
		slog.Warn("Error getting session admin", "error", err)
	}

	return sessionMaximumVoters
}

func (sc *Session) GetSessionMaximumVoters(sessionID int) int {
	sessionMaximumVoters, err := sc.redisClient.HGet(context.Background(), getSessionConfigKey(sessionID), "maximumVoters").Result()
	if err != nil {
		slog.Warn("Error getting session maximum voters", "error", err)
	}
	sessionMaximumVotersInt, err := strconv.Atoi(sessionMaximumVoters)
	if err != nil {
		slog.Warn("Error converting session maximum voters to int", "error", err)
	}

	return sessionMaximumVotersInt
}

func (sc *Session) DoesSessionExist(sessionID int) bool {
	_, err := sc.redisClient.HGetAll(context.Background(), getSessionConfigKey(sessionID)).Result()
	if err != nil {
		slog.Warn("Error getting session config", "error", err)
		return false
	}
	return true
}

func getSessionConfigKey(sessionID int) string {
	return fmt.Sprintf("session-config-%d", sessionID)
}