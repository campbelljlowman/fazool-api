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
	"github.com/campbelljlowman/fazool-api/utils"
)


// Sessions get watched at this frequency in seconds
const sessionWatchFrequency time.Duration = 10

// Session gets removed after being inactive for this long in minutes
const sessionTimeout time.Duration = 30

// Spotify gets watched by default at this frequency in milliseconds
const spotifyWatchFrequency time.Duration = 250

// Voters get watched at this frequency in seconds
const voterWatchFrequency time.Duration = 1

type Session struct {
	redsync 		*redsync.Redsync
	redisClient 		*redis.Client
}

func NewSessionClient(redsync *redsync.Redsync, redisClient *redis.Client) *Session {
	sessionCache := &Session{
		redsync: redsync,
		redisClient: redisClient,
	}
	// Clean up any data from previous run, this should be removed when running in production
	redisClient.FlushAllAsync(context.Background()).Result()
	return sessionCache
}

func (s *Session) CreateSession(adminAccountID, accountLevel string) (int, error) {
	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		return 0, utils.LogAndReturnError("Error generating session ID", err)
	}

	maximumVoters := 0
	if accountLevel == constants.RegularAccountLevel {
		maximumVoters = 50
	}

	s.refreshSession(sessionID)
	s.initVoterMap(sessionID)
	s.setSessionConfig(sessionID, maximumVoters, adminAccountID)
	s.initQueue(sessionID)
	s.initCurrentlyPlaying(sessionID)

	return sessionID, nil
}

// TODO: This code hasn't been tested
func (s *Session) EndSession(sessionID int) {
// Delete all keys from Redis
// Remove shedulers
// Remove session from database
}

func (s *Session) CheckSessionExpiry(sessionID int) {
	isSessionExpired := s.isSessionExpired(sessionID)
	if isSessionExpired == true {
		s.EndSession(sessionID)
	}
}

// TODO: This code hasn't been tested
func (s *Session) processBonusVotes(sessionID int, songID string) error {
	sessionBonusVotes, bonusVoteMutex := s.lockAndGetBonusVotes(sessionID)

	songBonusVotes, exists := sessionBonusVotes[songID]
	delete(sessionBonusVotes, songID)

	s.setAndUnlockBonusVotes(sessionID, sessionBonusVotes, bonusVoteMutex)

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

func (s *Session) CheckVotersExpirations(sessionID int) {

	for {
		if s.isSessionExpired(sessionID) {
			slog.Info("Session has expired, ending session voter watcher", "session_id", sessionID)
			// TODO: Remove this when the sheduler calls this function
			return
		}

		voters, votersMutex := s.lockAndGetAllVotersInSession(sessionID)

		for _, voter := range voters {
			if voter.VoterType == constants.AdminVoterType {
				continue
			}

			if time.Now().After(voter.ExpiresAt) {
				slog.Info("Voter exipred! Removing", "voter", voter.VoterID)
				delete(voters, voter.VoterID)
			}

		}
		s.setAndUnlockAllVotersInSession(sessionID, voters, votersMutex)

		time.Sleep(voterWatchFrequency * time.Second)
	}
}

func (s *Session) IsSessionFull(sessionID int) bool {
	isFull := false

	voterCount := s.getNumberOfVoters(sessionID)

	sessionMaximumVoters := s.getSessionMaximumVoters(sessionID)

	if voterCount >= sessionMaximumVoters {
		isFull = true
	}
	return isFull
}

func (s *Session) AdvanceQueue(sessionID int, force bool, musicPlayer musicplayer.MusicPlayer) error { 
	var song *model.SimpleSong

	queue, queueMutex := s.lockAndGetQueue(sessionID)
	if len(queue) == 0 {
		queueMutex.Unlock()
		return nil
	}

	song, queue = queue[0].SimpleSong, queue[1:]
	s.setAndUnlockQueue(sessionID, queue, queueMutex)

	err := musicPlayer.QueueSong(song.ID)
	if err != nil {
		return err
	}

	err = s.processBonusVotes(sessionID, song.ID)
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

	s.refreshSession(sessionID)
	return nil
}

func (s *Session) UpsertQueue(sessionID int, vote int, song model.SongUpdate) {
	queue, queueMutex := s.lockAndGetQueue(sessionID)
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
	s.setAndUnlockQueue(sessionID, queue, queueMutex)
	s.refreshSession(sessionID)
}


func (s *Session) CheckSpotifyCurrentlyPlaying(sessionID int, musicPlayer musicplayer.MusicPlayer) {
	// s.SessionInfo.CurrentlyPlaying = &model.CurrentlyPlayingSong{}
	updateSessionFlag := false
	addNextSongFlag := false

	for {
		if s.isSessionExpired(sessionID) {
			slog.Info("Session has expired, ending session spotify watcher", "session_id", sessionID)
			return
		}

		updateSessionFlag = false
		spotifyCurrentlyPlayingSong, isSpotifyCurrentlyPlaying, err := musicPlayer.CurrentSong()
		if err != nil {
			slog.Warn("Error getting music player state", "error", err)
			continue
		}

		currentlyPlaying, currentlyPlayingMutex := s.lockAndGetCurrentlyPlaying(sessionID)
		if isSpotifyCurrentlyPlaying == true {
			if currentlyPlaying.SimpleSong.ID != spotifyCurrentlyPlayingSong.SimpleSong.ID {
				// If song has changed, update currently playing, send update, and set flag to pop next song from queue
				currentlyPlaying = spotifyCurrentlyPlayingSong
				updateSessionFlag = true
				addNextSongFlag = true
			} else if currentlyPlaying.Playing != isSpotifyCurrentlyPlaying {
				// If same song is paused and then played, set the new state
				currentlyPlaying.Playing = isSpotifyCurrentlyPlaying
				updateSessionFlag = true
			}

			// If the currently playing song is about to end, pop the top of the session and add to spotify queue
			// If go spotify client adds API for checking current queue, checking this is a better way to tell if it's
			// Safe to add song
			timeLeft, err := musicPlayer.TimeRemaining()
			if err != nil {
				slog.Warn("Error getting song time remaining", "error", err)
				continue
			}

			if timeLeft < 5000 && addNextSongFlag {
				s.AdvanceQueue(sessionID, false, musicPlayer)

				updateSessionFlag = true
				addNextSongFlag = false
			}
		} else {
			// Change currently playing to false if music gets paused
			if currentlyPlaying.Playing != isSpotifyCurrentlyPlaying {
				currentlyPlaying.Playing = isSpotifyCurrentlyPlaying
				updateSessionFlag = true
			}
		}

		if updateSessionFlag {
			s.setAndUnlockCurrentlyPlaying(sessionID, currentlyPlaying, currentlyPlayingMutex)
			s.refreshSession(sessionID)
		} else {
			currentlyPlayingMutex.Unlock()
		}

		// TODO: Maybe make this refresh value dynamic to adjust refresh frequency at the end of a song
		time.Sleep(spotifyWatchFrequency * time.Millisecond)
	}
}


func (s *Session) GetSessionInfo(sessionID int) *model.SessionInfo {
	sessionInfo := &model.SessionInfo{
		ID: sessionID,
		CurrentlyPlaying: s.getCurrentlyPlaying(sessionID),
		Queue: s.getQueue(sessionID),
		Admin: s.GetSessionAdmin(sessionID),
		NumberOfVoters: s.getNumberOfVoters(sessionID),
		MaximumVoters: s.getSessionMaximumVoters(sessionID),
	}

	return sessionInfo
}

func (s *Session) getStructFromRedis(key string, dest interface{}) error {
    result, err := s.redisClient.Get(context.Background(), key).Result()
    if err != nil {
    	return err
    }
	err = json.Unmarshal([]byte(result), dest)
	if err != nil {
		slog.Info("Error unmarshaling json", "error", err)
	}
    return nil
}

func (s *Session) setStructToRedis(key string, value interface{}) error {
	valueString, err := json.Marshal(value)
    if err != nil {
       return err
    }
    return s.redisClient.Set(context.Background(), key, string(valueString), 0).Err()
}
