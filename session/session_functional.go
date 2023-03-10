package session

import (
	"context"
	"fmt"
	"time"
	"encoding/json"

	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/slog"
)

type SessionCache struct {
	rs 		*redsync.Redsync
	rc 		*redis.Client
}

func NewSessionCache(rs *redsync.Redsync, rc *redis.Client) *SessionCache {
	sessionCache := &SessionCache{
		rs: rs,
		rc: rc,
	}
	return sessionCache
}

func (sc *SessionCache) CreateSession(sessionID int) {
	sc.RefreshSession(sessionID)
	// expiryMutex := sc.rs.NewMutex(fmt.Sprintf("expiry-%d", sessionID))
	// expiryMutex.Lock()
	// expiresAt := time.Now().Add(sessionTimeout * time.Minute)
	// sc.setSessionExpiration(sessionID, expiresAt)
	// expiryMutex.Unlock()

}

func (sc *SessionCache) ExpireSession(sessionID int) {
	expiryMutex := sc.rs.NewMutex(fmt.Sprintf("expiry-mutex-%d", sessionID))

	expiryMutex.Lock()
	expiresAt := time.Now()
	err := sc.setSessionExpiration(sessionID, expiresAt)
	expiryMutex.Unlock()

	if err != nil {
		slog.Warn("Error setting session Expiration", "error", err)
	}
}

func (sc *SessionCache) IsExpired(sessionID int) bool {
	expiryMutex := sc.rs.NewMutex(fmt.Sprintf("expiry-mutex-%d", sessionID))
	var expiresAt time.Time

	expiryMutex.Lock()
	err := sc.getSessionExpiration(sessionID, &expiresAt)
	expiryMutex.Unlock()

	if err != nil {
		slog.Warn("Error getting session's expiration", "error", err)
		return false
	}
	isExpired := expiresAt.Before(time.Now())

	return isExpired
}

func (sc *SessionCache) RefreshSession(sessionID int) {
	expiryMutex := sc.rs.NewMutex(fmt.Sprintf("expiry-mutex-%d", sessionID))
	expiryMutex.Lock()
	expiresAt := time.Now().Add(sessionTimeout * time.Minute)
	err := sc.setSessionExpiration(sessionID, expiresAt)
	if err != nil {
		slog.Warn("Error refreshing session", "error", err)
	}
	expiryMutex.Unlock()
}

func (sc *SessionCache) getSessionExpiration(sessionID int, dest interface{}) error {
    result, err := sc.rc.Get(context.Background(), fmt.Sprintf("expiry-%d", sessionID)).Result()
    if err != nil {
       return err
    }
	json.Unmarshal([]byte(result), dest)
    return nil
}

func (sc *SessionCache) setSessionExpiration(sessionID int, value interface{}) error {
	valueString, err := json.Marshal(value)
    if err != nil {
       return err
    }
    return sc.rc.Set(context.Background(), fmt.Sprintf("expiry-%d", sessionID), string(valueString), 0).Err()
}