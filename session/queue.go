package session

import (
	"fmt"
	
	"golang.org/x/exp/slog"
	"github.com/go-redsync/redsync/v4"

	"github.com/campbelljlowman/fazool-api/graph/model"

)

func (sc *Session) initQueue(sessionID int) {
	var queue []*model.QueuedSong
	sc.SetQueue(sessionID, queue)
}

func (sc *Session) getQueue(sessionID int) []*model.QueuedSong {
	queue, queueMutex := sc.lockAndGetQueue(sessionID)
	queueMutex.Unlock()

	return queue
}

func (sc *Session) SetQueue(sessionID int, newQueue [] *model.QueuedSong) {
	queueMutex := sc.redsync.NewMutex(getQueueMutexKey(sessionID))
	queueMutex.Lock()

	sc.setAndUnlockQueue(sessionID, newQueue, queueMutex)
}

func (sc *Session) lockAndGetQueue(sessionID int) ([]*model.QueuedSong, *redsync.Mutex) {
	queueMutex := sc.redsync.NewMutex(getQueueMutexKey(sessionID))
	queueMutex.Lock()

	var queue [] *model.QueuedSong
	err := sc.getStructFromRedis(getQueueKey(sessionID), &queue)

	if err != nil {
		slog.Warn("Error getting session queue", "error", err)
	}

	return queue, queueMutex
}

func (sc *Session) setAndUnlockQueue(sessionID int, newQueue [] *model.QueuedSong, queueMutex *redsync.Mutex) {
	sc.setStructToRedis(getQueueKey(sessionID), newQueue)
	queueMutex.Unlock()
}

func getQueueMutexKey(sessionID int) string {
	return fmt.Sprintf("queue-mutex-%d", sessionID)
}

func getQueueKey(sessionID int) string {
	return fmt.Sprintf("queue-%d", sessionID)
}