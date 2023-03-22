package session

import (
	"fmt"
	"time"

	"golang.org/x/exp/slog"
	"github.com/go-redsync/redsync/v4"

)

func (s *Session) expireSession(sessionID int) {
	expiryMutex := s.redsync.NewMutex(getExpiryMutexKey(sessionID))

	expiryMutex.Lock()
	expiresAt := time.Now()
	err := s.setStructToRedis(getExpiryKey(sessionID), expiresAt)
	expiryMutex.Unlock()

	if err != nil {
		slog.Warn("Error setting session Expiration", "error", err)
	}
}

func (s *Session) isSessionExpired(sessionID int) bool {
	expiresAt, expiryMutex := s.lockAndGetSessionExpiry(sessionID)

	expiryMutex.Unlock()

	isExpired := expiresAt.Before(time.Now())

	return isExpired
}

func (s *Session) refreshSession(sessionID int) {
	expiryMutex := s.redsync.NewMutex(getExpiryMutexKey(sessionID))
	expiryMutex.Lock()

	expiresAt := time.Now().Add(sessionTimeout * time.Minute)
	err := s.setStructToRedis(getExpiryKey(sessionID), expiresAt)

	expiryMutex.Unlock()

	if err != nil {
		slog.Warn("Error refreshing session", "error", err)
	}
}

func (s *Session) lockAndGetSessionExpiry(sessionID int) (time.Time, *redsync.Mutex) {
	expiryMutex := s.redsync.NewMutex(getExpiryMutexKey(sessionID))
	
	expiryMutex.Lock()
	var expiresAt time.Time
	err := s.getStructFromRedis(getExpiryKey(sessionID), &expiresAt)

	if err != nil {
		slog.Warn("Error getting session expiration", "error", err)
	}

	return expiresAt, expiryMutex
}

func getExpiryMutexKey(sessionID int) string {
	return fmt.Sprintf("expiry-mutex-%d", sessionID)
}

func getExpiryKey(sessionID int) string {
	return fmt.Sprintf("expiry-%d", sessionID)
}