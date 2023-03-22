package session

import (
	"fmt"
	
	"golang.org/x/exp/slog"
	"github.com/go-redsync/redsync/v4"

	"github.com/campbelljlowman/fazool-api/graph/model"

)

func (s *Session) initQueue(sessionID int) {
	var queue []*model.QueuedSong
	s.SetQueue(sessionID, queue)
}

func (s *Session) getQueue(sessionID int) []*model.QueuedSong {
	queue, queueMutex := s.lockAndGetQueue(sessionID)
	queueMutex.Unlock()

	return queue
}

func (s *Session) SetQueue(sessionID int, newQueue [] *model.QueuedSong) {
	queueMutex := s.redsync.NewMutex(getQueueMutexKey(sessionID))
	queueMutex.Lock()

	s.setAndUnlockQueue(sessionID, newQueue, queueMutex)
}

func (s *Session) lockAndGetQueue(sessionID int) ([]*model.QueuedSong, *redsync.Mutex) {
	queueMutex := s.redsync.NewMutex(getQueueMutexKey(sessionID))
	queueMutex.Lock()

	var queue [] *model.QueuedSong
	err := s.getStructFromRedis(getQueueKey(sessionID), &queue)

	if err != nil {
		slog.Warn("Error getting session queue", "error", err)
	}

	return queue, queueMutex
}

func (s *Session) setAndUnlockQueue(sessionID int, newQueue [] *model.QueuedSong, queueMutex *redsync.Mutex) {
	err := s.setStructToRedis(getQueueKey(sessionID), newQueue)
	if err != nil {
		slog.Warn("Error setting queue", "error", err)
	}
	queueMutex.Unlock()
}

func getQueueMutexKey(sessionID int) string {
	return fmt.Sprintf("queue-mutex-%d", sessionID)
}

func getQueueKey(sessionID int) string {
	return fmt.Sprintf("queue-%d", sessionID)
}