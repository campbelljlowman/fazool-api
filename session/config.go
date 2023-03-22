package session

import (
	"fmt"
	"context"
	"strconv"

	"golang.org/x/exp/slog"

)

func (s *Session) setSessionConfig(sessionID, maximumVoters int, adminAccountID string) {
	s.redisClient.HSet(context.Background(),  getSessionConfigKey(sessionID), "sessionID", sessionID, "maximumVoters", maximumVoters, "adminAccountID", adminAccountID)
}

func (s *Session) GetSessionAdmin(sessionID int) string {
	sessionMaximumVoters, err := s.redisClient.HGet(context.Background(), getSessionConfigKey(sessionID), "adminAccountID").Result()
	if err != nil {
		slog.Warn("Error getting session admin", "error", err)
	}

	return sessionMaximumVoters
}

func (s *Session) getSessionMaximumVoters(sessionID int) int {
	sessionMaximumVoters, err := s.redisClient.HGet(context.Background(), getSessionConfigKey(sessionID), "maximumVoters").Result()
	if err != nil {
		slog.Warn("Error getting session maximum voters", "error", err)
	}
	sessionMaximumVotersInt, err := strconv.Atoi(sessionMaximumVoters)
	if err != nil {
		slog.Warn("Error converting session maximum voters to int", "error", err)
	}

	return sessionMaximumVotersInt
}

func (s *Session) DoesSessionExist(sessionID int) bool {
	_, err := s.redisClient.HGetAll(context.Background(), getSessionConfigKey(sessionID)).Result()
	if err != nil {
		slog.Warn("Error getting session config", "error", err)
		return false
	}
	return true
}

func getSessionConfigKey(sessionID int) string {
	return fmt.Sprintf("session-config-%d", sessionID)
}