package session

import (
	"fmt"
	"time"

	"golang.org/x/exp/slog"
	"github.com/go-redsync/redsync/v4"

)

func (sc *Session) ExpireSession(sessionID int) {
	expiryMutex := sc.redsync.NewMutex(getExpiryMutexKey(sessionID))

	expiryMutex.Lock()
	expiresAt := time.Now()
	err := sc.setStructToRedis(getExpiryKey(sessionID), expiresAt)
	expiryMutex.Unlock()

	if err != nil {
		slog.Warn("Error setting session Expiration", "error", err)
	}
}

func (sc *Session) IsSessionExpired(sessionID int) bool {
	expiresAt, expiryMutex := sc.lockAndGetSessionExpiry(sessionID)

	expiryMutex.Unlock()

	isExpired := expiresAt.Before(time.Now())

	return isExpired
}

func (sc *Session) RefreshSession(sessionID int) {
	expiryMutex := sc.redsync.NewMutex(getExpiryMutexKey(sessionID))
	expiryMutex.Lock()

	expiresAt := time.Now().Add(sessionTimeout * time.Minute)
	err := sc.setStructToRedis(getExpiryKey(sessionID), expiresAt)

	expiryMutex.Unlock()

	if err != nil {
		slog.Warn("Error refreshing session", "error", err)
	}
}

func (sc *Session) lockAndGetSessionExpiry(sessionID int) (time.Time, *redsync.Mutex) {
	expiryMutex := sc.redsync.NewMutex(getExpiryMutexKey(sessionID))
	
	expiryMutex.Lock()
	var expiresAt time.Time
	err := sc.getStructFromRedis(getExpiryKey(sessionID), &expiresAt)

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