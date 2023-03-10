package session

import (
	"context"
	"fmt"
	"time"
	"encoding/json"
	"os"

	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/slog"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/campbelljlowman/fazool-api/database"

)

type SessionCache struct {
	redsync 		*redsync.Redsync
	redisClient 		*redis.Client
}

func NewSessionCache(redsync *redsync.Redsync, redisClient *redis.Client) *SessionCache {
	sessionCache := &SessionCache{
		redsync: redsync,
		redisClient: redisClient,
	}
	return sessionCache
}

func (sc *SessionCache) CreateSession(sessionID int) {
	sc.RefreshSession(sessionID)
}

func (sc *SessionCache) ExpireSession(sessionID int) {
	expiryMutex := sc.redsync.NewMutex(fmt.Sprintf("expiry-mutex-%d", sessionID))

	expiryMutex.Lock()
	expiresAt := time.Now()
	err := sc.set(fmt.Sprintf("expiry-%d", sessionID), expiresAt)
	expiryMutex.Unlock()

	if err != nil {
		slog.Warn("Error setting session Expiration", "error", err)
	}
}

func (sc *SessionCache) IsExpired(sessionID int) bool {
	expiryMutex := sc.redsync.NewMutex(fmt.Sprintf("expiry-mutex-%d", sessionID))
	var expiresAt time.Time

	expiryMutex.Lock()
	err := sc.get(fmt.Sprintf("expiry-%d", sessionID), &expiresAt)
	expiryMutex.Unlock()

	if err != nil {
		slog.Warn("Error getting session's expiration", "error", err)
		return false
	}
	isExpired := expiresAt.Before(time.Now())

	return isExpired
}

func (sc *SessionCache) RefreshSession(sessionID int) {
	expiryMutex := sc.redsync.NewMutex(fmt.Sprintf("expiry-mutex-%d", sessionID))
	expiryMutex.Lock()
	expiresAt := time.Now().Add(sessionTimeout * time.Minute)
	err := sc.set(fmt.Sprintf("expiry-%d", sessionID), expiresAt)
	if err != nil {
		slog.Warn("Error refreshing session", "error", err)
	}
	expiryMutex.Unlock()
}

func (sc *SessionCache) AddBonusVote(songID, accountID string, numberOfVotes, sessionID int) {
	bonusVoteMutex := sc.redsync.NewMutex(getBonusVoteMutexKey(sessionID))
	// Map of [song][account][votes]
	var bonusVotes map[string]map[string]int

	bonusVoteMutex.Lock()
	sc.get(getBonusVoteKey(sessionID), bonusVotes)
	if _, exists := bonusVotes[songID][accountID]; !exists {
		bonusVotes[songID] = make(map[string]int)
	}
	bonusVotes[songID][accountID] += numberOfVotes
	err := sc.set(getBonusVoteKey(sessionID), bonusVotes)
	bonusVoteMutex.Unlock()

	if err != nil {
		slog.Warn("Error updating bonus votes", "error", err)
	}
}


// TODO: This code hasn't been tested
func (sc *SessionCache) processBonusVotes(sessionID int, songID string) error {
	bonusVoteMutex := sc.redsync.NewMutex(getBonusVoteMutexKey(sessionID))
	// Map of [song][account][votes]
	var sessionBonusVotes map[string]map[string]int


	bonusVoteMutex.Lock()
	sc.get(getBonusVoteKey(sessionID), sessionBonusVotes)

	songBonusVotes, exists := sessionBonusVotes[songID]
	delete(sessionBonusVotes, songID)

	sc.set(getBonusVoteKey(sessionID), sessionBonusVotes)
	bonusVoteMutex.Unlock()

	if !exists {
		return nil
	}

	databaseURL := os.Getenv("POSTRGRES_URL")

	dbPool, err := pgxpool.Connect(context.Background(), databaseURL)
	if err != nil {
		return err
	}

	pg := database.PostgresWrapper{PostgresClient: dbPool}

	for accountID, votes := range songBonusVotes {
		err = pg.SubtractBonusVotes(accountID, votes)
		if err != nil {
			slog.Warn("Error updating account's bonus votes", "account", accountID)
		}
	}

	pg.CloseConnection()
	return nil
}

func (sc *SessionCache) get(key string, dest interface{}) error {
    result, err := sc.redisClient.Get(context.Background(), key).Result()
    if err != nil {
       return err
    }
	json.Unmarshal([]byte(result), dest)
    return nil
}

func (sc *SessionCache) set(key string, value interface{}) error {
	valueString, err := json.Marshal(value)
    if err != nil {
       return err
    }
    return sc.redisClient.Set(context.Background(), key, string(valueString), 0).Err()
}

func getBonusVoteMutexKey(sessionID int) string {
	return fmt.Sprintf("bonus-vote-mutex-%d", sessionID)
}

func getBonusVoteKey(sessionID int) string {
	return fmt.Sprintf("bonus-vote-%d", sessionID)
}