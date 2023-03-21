package session

import (
	"fmt"

	"golang.org/x/exp/slog"
	"github.com/go-redsync/redsync/v4"

	"github.com/campbelljlowman/fazool-api/graph/model"

)

func (sc *Session) initCurrentlyPlaying(sessionID int) {
	currentlyPlayingMutex := sc.redsync.NewMutex(getCurrentlyPlayingMutexKey(sessionID))
	currentlyPlayingMutex.Lock()

	currentlyPlaying := &model.CurrentlyPlayingSong{
		SimpleSong: &model.SimpleSong{},
		Playing:    false,
	}
	sc.setStructToRedis(getCurrentlyPlayingKey(sessionID), currentlyPlaying)
	currentlyPlayingMutex.Unlock()
}

func (sc *Session) getCurrentlyPlaying(sessionID int) *model.CurrentlyPlayingSong {
	currentlyPlaying, currentlyPlayingMutex := sc.lockAndGetCurrentlyPlaying(sessionID)
	currentlyPlayingMutex.Unlock()
	return currentlyPlaying
}

func (sc *Session) lockAndGetCurrentlyPlaying(sessionID int) (*model.CurrentlyPlayingSong, *redsync.Mutex) {
	currentlyPlayingMutex := sc.redsync.NewMutex(getCurrentlyPlayingMutexKey(sessionID))
	currentlyPlayingMutex.Lock()

	var currentlyPlaying *model.CurrentlyPlayingSong
	err := sc.getStructFromRedis(getCurrentlyPlayingKey(sessionID), &currentlyPlaying)

	if err != nil {
		slog.Warn("Error getting session currentlyPlaying", "error", err)
	}

	return currentlyPlaying, currentlyPlayingMutex
}

func (sc *Session) setAndUnlockCurrentlyPlaying(sessionID int, newCurrentlyPlaying *model.CurrentlyPlayingSong, currentlyPlayingMutex *redsync.Mutex) {
	sc.setStructToRedis(getCurrentlyPlayingKey(sessionID), newCurrentlyPlaying)
	currentlyPlayingMutex.Unlock()
}

func getCurrentlyPlayingMutexKey(sessionID int) string {
	return fmt.Sprintf("currently-playing-mutex-%d", sessionID)
}

func getCurrentlyPlayingKey(sessionID int) string {
	return fmt.Sprintf("currently-playing-%d", sessionID)
}