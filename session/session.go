package session

import (
	"sync"
	"time"

	"github.com/campbelljlowman/fazool-api/constants"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/musicplayer"
	"github.com/campbelljlowman/fazool-api/utils"
	"golang.org/x/exp/slog"

)

type Session struct {
	SessionInfo    *model.SessionInfo
	queueMutex     *sync.Mutex
	sc 				*SessionCache
}

// Session gets removed after being inactive for this long in minutes
const sessionTimeout time.Duration = 30

// Spotify gets watched by default at this frequency in milliseconds
const spotifyWatchFrequency time.Duration = 250

// Voters get watched at this frequency in seconds
const voterWatchFrequency time.Duration = 1

func NewSession(accountID, accountLevel string, sc *SessionCache) (*Session, int, error) {
	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		return nil, 0, utils.LogAndReturnError("Error generating session ID", err)
	}

	sessionSize := 0
	if accountLevel == constants.RegularAccountLevel {
		sessionSize = 50
	}

	// Create session info
	sessionInfo := &model.SessionInfo{
		CurrentlyPlaying: &model.CurrentlyPlayingSong{
			SimpleSong: &model.SimpleSong{},
			Playing:    false,
		},
	}

	sc.CreateSession(sessionID, sessionSize, accountID)

	session := Session{
		SessionInfo:    sessionInfo,
		sc: 			sc,
	}

	return &session, sessionID, nil
}

func (s *Session) WatchSpotifyCurrentlyPlaying(sessionID int, musicPlayer musicplayer.MusicPlayer) {
	// s.SessionInfo.CurrentlyPlaying = &model.CurrentlyPlayingSong{}
	updateCacheFlag := false
	addNextSongFlag := false

	for {
		if s.sc.IsSessionExpired(sessionID) {
			slog.Info("Session has expired, ending session spotify watcher", "session_id", sessionID)
			return
		}

		updateCacheFlag = false
		spotifyCurrentlyPlayingSong, spotifyCurrentlyPlaying, err := musicPlayer.CurrentSong()
		if err != nil {
			slog.Warn("Error getting music player state", "error", err)
			continue
		}

		if spotifyCurrentlyPlaying == true {
			if s.SessionInfo.CurrentlyPlaying.SimpleSong.ID != spotifyCurrentlyPlayingSong.SimpleSong.ID {
				// If song has changed, update currently playing, send update, and set flag to pop next song from queue
				s.SessionInfo.CurrentlyPlaying = spotifyCurrentlyPlayingSong
				updateCacheFlag = true
				addNextSongFlag = true
			} else if s.SessionInfo.CurrentlyPlaying.Playing != spotifyCurrentlyPlaying {
				// If same song is paused and then played, set the new state
				s.SessionInfo.CurrentlyPlaying.Playing = spotifyCurrentlyPlaying
				updateCacheFlag = true
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
				s.sc.AdvanceQueue(sessionID, false, musicPlayer)

				updateCacheFlag = true
				addNextSongFlag = false
			}
		} else {
			// Change currently playing to false if music gets paused
			if s.SessionInfo.CurrentlyPlaying.Playing != spotifyCurrentlyPlaying {
				s.SessionInfo.CurrentlyPlaying.Playing = spotifyCurrentlyPlaying
				updateCacheFlag = true
			}
		}

		if updateCacheFlag {
			s.sc.RefreshSession(sessionID)
		}

		// TODO: Maybe make this refresh value dynamic to adjust refresh frequency at the end of a song
		time.Sleep(spotifyWatchFrequency * time.Millisecond)
	}
}
