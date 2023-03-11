package session

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/slog"

	"github.com/campbelljlowman/fazool-api/constants"
	"github.com/campbelljlowman/fazool-api/database"
	"github.com/campbelljlowman/fazool-api/voter"
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

func (sc *SessionCache) CreateSession(sessionID, maximumVoters int, adminAccountID string) {
	sc.RefreshSession(sessionID)
	sc.InitVoterMap(sessionID)
	sc.SetSessionConfig(sessionID, maximumVoters, adminAccountID)
}

func (sc *SessionCache) SetSessionConfig(sessionID, maximumVoters int, adminAccountID string) {
	sc.redisClient.HSet(context.Background(),  getSessionConfigKey(sessionID), "sessionID", sessionID, "maximumVoters", maximumVoters, "adminAccountID", adminAccountID)
}

func (sc *SessionCache) InitVoterMap(sessionID int) {
	votersMutex := sc.redsync.NewMutex(getVotersMutexKey(sessionID))
	voters := make(map[string]*voter.Voter)
	votersMutex.Lock()
	sc.set(getVotersKey(sessionID), voters)
	votersMutex.Unlock()
}

func (sc *SessionCache) ExpireSession(sessionID int) {
	expiryMutex := sc.redsync.NewMutex(getExpiryMutexKey(sessionID))

	expiryMutex.Lock()
	expiresAt := time.Now()
	err := sc.set(getExpiryKey(sessionID), expiresAt)
	expiryMutex.Unlock()

	if err != nil {
		slog.Warn("Error setting session Expiration", "error", err)
	}
}

func (sc *SessionCache) IsSessionExpired(sessionID int) bool {
	expiryMutex := sc.redsync.NewMutex(getExpiryMutexKey(sessionID))
	
	expiryMutex.Lock()
	var expiresAt time.Time
	err := sc.get(getExpiryKey(sessionID), &expiresAt)
	expiryMutex.Unlock()

	if err != nil {
		slog.Warn("Error getting session's expiration", "error", err)
		return false
	}
	isExpired := expiresAt.Before(time.Now())

	return isExpired
}

func (sc *SessionCache) RefreshSession(sessionID int) {
	expiryMutex := sc.redsync.NewMutex(getExpiryMutexKey(sessionID))
	expiryMutex.Lock()
	expiresAt := time.Now().Add(sessionTimeout * time.Minute)
	err := sc.set(getExpiryKey(sessionID), expiresAt)
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
	sc.get(getBonusVoteKey(sessionID), &bonusVotes)
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
	sc.get(getBonusVoteKey(sessionID), &sessionBonusVotes)

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

func (sc *SessionCache) CheckVotersExpirations(sessionID int) {

	for {
		if sc.IsSessionExpired(sessionID) {
			slog.Info("Session has expired, ending session voter watcher", "session_id", sessionID)
			// TODO: Deregister this check from the scheduler
			return
		}

		votersMutex := sc.redsync.NewMutex(getVotersMutexKey(sessionID))
		votersMutex.Lock()

		var voters map[string] *voter.Voter
		sc.get(getVotersKey(sessionID), &voters)
		for _, voter := range voters {
			if voter.VoterType == constants.AdminVoterType {
				continue
			}

			if time.Now().After(voter.ExpiresAt) {
				slog.Info("Voter exipred! Removing", "voter", voter.VoterID)
				delete(voters, voter.VoterID)
			}

		}
		sc.set(getVotersKey(sessionID), voters)
		votersMutex.Unlock()

		time.Sleep(voterWatchFrequency * time.Second)
	}
}

func (sc *SessionCache) UpsertVoterToSession(sessionID int, newVoter *voter.Voter){
	slog.Info("Adding Voter:", "newVoter", newVoter)
	votersMutex := sc.redsync.NewMutex(getVotersMutexKey(sessionID))
	votersMutex.Lock()

	var voters map[string] *voter.Voter
	sc.get(getVotersKey(sessionID), &voters)
	voters[newVoter.VoterID] = newVoter
	sc.set(getVotersKey(sessionID), voters)

	votersMutex.Unlock()
}

func (sc *SessionCache) GetVoterInSession(sessionID int, voterID string) (*voter.Voter, bool){
	votersMutex := sc.redsync.NewMutex(getVotersMutexKey(sessionID))
	votersMutex.Lock()

	var voters map[string] *voter.Voter
	sc.get(getVotersKey(sessionID), &voters)
	voter, exists := voters[voterID]
	slog.Info("existing voter", "voter", voter)
	votersMutex.Unlock()

	return voter, exists
}

func (sc *SessionCache) IsSessionFull(sessionID int) bool {
	isFull := false

	votersMutex := sc.redsync.NewMutex(getVotersMutexKey(sessionID))

	votersMutex.Lock()
	var voters map[string] *voter.Voter
	sc.get(getVotersKey(sessionID), &voters)
	votersMutex.Unlock()

	sessionMaximumVoters, err := sc.redisClient.HGet(context.Background(), getSessionConfigKey(sessionID), "maximumVoters").Result()
	if err != nil {
		slog.Warn("Error getting sessions maximum voters", "error", err)
		return true
	}
	sessionMaximumVotersInt, err := strconv.Atoi(sessionMaximumVoters)
	if err != nil {
		slog.Warn("Error converting maximum voters from string to int", "error", err)
		return true
	}

	if len(voters) >= sessionMaximumVotersInt {
		isFull = true
	}
	return isFull
}


func (sc *SessionCache) get(key string, dest interface{}) error {
    result, err := sc.redisClient.Get(context.Background(), key).Result()
    if err != nil {
    	return err
    }
	err = json.Unmarshal([]byte(result), dest)
	if err != nil {
		slog.Info("Error unmarshaling json", "error", err)
	}
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

func getExpiryMutexKey(sessionID int) string {
	return fmt.Sprintf("expiry-mutex-%d", sessionID)
}

func getExpiryKey(sessionID int) string {
	return fmt.Sprintf("expiry-%d", sessionID)
}

func getVotersMutexKey(sessionID int) string {
	return fmt.Sprintf("voters-mutex-%d", sessionID)
}

func getVotersKey(sessionID int) string {
	return fmt.Sprintf("voters-%d", sessionID)
}

func getSessionConfigKey(sessionID int) string {
	return fmt.Sprintf("session-config-%d", sessionID)
}