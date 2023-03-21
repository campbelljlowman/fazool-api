package session

import (
	"context"
	"encoding/json"
	"os"
	"time"
	"sort"


	"github.com/go-redsync/redsync/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/slog"
	"golang.org/x/exp/slices"


	"github.com/campbelljlowman/fazool-api/constants"
	"github.com/campbelljlowman/fazool-api/database"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/musicplayer"
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
	sc.InitQueue(sessionID)
}

// TODO: This code hasn't been tested
func (sc *SessionCache) processBonusVotes(sessionID int, songID string) error {
	sessionBonusVotes, bonusVoteMutex := sc.lockAndGetBonusVotes(sessionID)

	songBonusVotes, exists := sessionBonusVotes[songID]
	delete(sessionBonusVotes, songID)

	err :=	sc.setStructToRedis(getBonusVoteKey(sessionID), sessionBonusVotes)

	bonusVoteMutex.Unlock()

	if err != nil {
		slog.Warn("Error setting session Expiration", "error", err)
	}

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

		voters, votersMutex := sc.lockAndGetAllVotersInSession(sessionID)

		for _, voter := range voters {
			if voter.VoterType == constants.AdminVoterType {
				continue
			}

			if time.Now().After(voter.ExpiresAt) {
				slog.Info("Voter exipred! Removing", "voter", voter.VoterID)
				delete(voters, voter.VoterID)
			}

		}
		err := sc.setStructToRedis(getVotersKey(sessionID), voters)

		votersMutex.Unlock()

		if err != nil {
			slog.Warn("Error setting session voters", "error", err)
		}

		time.Sleep(voterWatchFrequency * time.Second)
	}
}

func (sc *SessionCache) IsSessionFull(sessionID int) bool {
	isFull := false

	voterCount := sc.GetNumberOfVoters(sessionID)

	sessionMaximumVoters := sc.GetSessionMaximumVoters(sessionID)

	if voterCount >= sessionMaximumVoters {
		isFull = true
	}
	return isFull
}

func (sc *SessionCache) AdvanceQueue(sessionID int, force bool, musicPlayer musicplayer.MusicPlayer) error { 
	var song *model.SimpleSong

	queue, queueMutex := sc.lockAndGetQueue(sessionID)
	if len(queue) == 0 {
		queueMutex.Unlock()
		return nil
	}

	song, queue = queue[0].SimpleSong, queue[1:]
	sc.setAndUnlockQueue(sessionID, queue, queueMutex)
	queueMutex.Unlock()

	err := musicPlayer.QueueSong(song.ID)
	if err != nil {
		return err
	}

	err = sc.processBonusVotes(sessionID, song.ID)
	if err != nil {
		return err
	}

	if !force {
		return nil
	}

	err = musicPlayer.Next()
	if err != nil {
		return err
	}

	return nil
}

func (sc *SessionCache) UpsertQueue(sessionID int, vote int, song model.SongUpdate) {
	queue, queueMutex := sc.lockAndGetQueue(sessionID)
	idx := slices.IndexFunc(queue, func(s *model.QueuedSong) bool { return s.SimpleSong.ID == song.ID })
	if idx == -1 {
		// add new song to queue
		newSong := &model.QueuedSong{
			SimpleSong: &model.SimpleSong{
				ID:     song.ID,
				Title:  *song.Title,
				Artist: *song.Artist,
				Image:  *song.Image,
			},
			Votes: vote,
		}
		queue = append(queue, newSong)
	} else {
		queuedSong := queue[idx]
		queuedSong.Votes += vote
	}

	// Sort queue
	sort.Slice(queue, func(i, j int) bool { return queue[i].Votes > queue[j].Votes })
	sc.setAndUnlockQueue(sessionID, queue, queueMutex)
}

func (sc *SessionCache) GetSessionInfo(sessionID int) *model.SessionInfo {
	sessionInfo := &model.SessionInfo{
		ID: sessionID,
		// CurrentlyPlaying: session.SessionInfo.CurrentlyPlaying,
		Queue: sc.GetQueue(sessionID),
		Admin: sc.GetSessionAdmin(sessionID),
		NumberOfVoters: sc.GetNumberOfVoters(sessionID),
		MaximumVoters: sc.GetSessionMaximumVoters(sessionID),
	}

	return sessionInfo
}
func (sc *SessionCache) getStructFromRedis(key string, dest interface{}) error {
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

func (sc *SessionCache) setStructToRedis(key string, value interface{}) error {
	valueString, err := json.Marshal(value)
    if err != nil {
       return err
    }
    return sc.redisClient.Set(context.Background(), key, string(valueString), 0).Err()
}
