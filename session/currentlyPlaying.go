package session

import (
	"fmt"

	"golang.org/x/exp/slog"
	"github.com/go-redsync/redsync/v4"

	"github.com/campbelljlowman/fazool-api/graph/model"

)

func (s *Session) initCurrentlyPlaying(sessionID int) {
	currentlyPlayingMutex := s.redsync.NewMutex(getCurrentlyPlayingMutexKey(sessionID))
	currentlyPlayingMutex.Lock()

	currentlyPlaying := &model.CurrentlyPlayingSong{
		SimpleSong: &model.SimpleSong{},
		Playing:    false,
	}
	s.setStructToRedis(getCurrentlyPlayingKey(sessionID), currentlyPlaying)
	currentlyPlayingMutex.Unlock()
}

func (s *Session) getCurrentlyPlaying(sessionID int) *model.CurrentlyPlayingSong {
	currentlyPlaying, currentlyPlayingMutex := s.lockAndGetCurrentlyPlaying(sessionID)
	currentlyPlayingMutex.Unlock()
	return currentlyPlaying
}

func (s *Session) lockAndGetCurrentlyPlaying(sessionID int) (*model.CurrentlyPlayingSong, *redsync.Mutex) {
	currentlyPlayingMutex := s.redsync.NewMutex(getCurrentlyPlayingMutexKey(sessionID))
	currentlyPlayingMutex.Lock()

	var currentlyPlaying *model.CurrentlyPlayingSong
	err := s.getStructFromRedis(getCurrentlyPlayingKey(sessionID), &currentlyPlaying)

	if err != nil {
		slog.Warn("Error getting session currentlyPlaying", "error", err)
	}

	return currentlyPlaying, currentlyPlayingMutex
}

func (s *Session) setAndUnlockCurrentlyPlaying(sessionID int, newCurrentlyPlaying *model.CurrentlyPlayingSong, currentlyPlayingMutex *redsync.Mutex) {
	err := s.setStructToRedis(getCurrentlyPlayingKey(sessionID), newCurrentlyPlaying)
	if err != nil {
		slog.Warn("Error setting currently playing", "error", err)
	}
	currentlyPlayingMutex.Unlock()
}

func getCurrentlyPlayingMutexKey(sessionID int) string {
	return fmt.Sprintf("currently-playing-mutex-%d", sessionID)
}

func getCurrentlyPlayingKey(sessionID int) string {
	return fmt.Sprintf("currently-playing-%d", sessionID)
}