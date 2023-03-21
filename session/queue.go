package session

import (
	"fmt"
	
	"golang.org/x/exp/slog"
	"github.com/go-redsync/redsync/v4"

	"github.com/campbelljlowman/fazool-api/graph/model"

)

func (sc *SessionCache) InitQueue(sessionID int) {
	var queue []*model.QueuedSong
	sc.SetQueue(sessionID, queue)
}

func (sc *SessionCache) GetQueue(sessionID int) []*model.QueuedSong {
	queue, queueMutex := sc.lockAndGetQueue(sessionID)
	queueMutex.Unlock()

	return queue
}

func (sc *SessionCache) SetQueue(sessionID int, newQueue [] *model.QueuedSong) {
	queueMutex := sc.redsync.NewMutex(getQueueMutexKey(sessionID))
	queueMutex.Lock()

	sc.setStructToRedis(getQueueKey(sessionID), newQueue)
	queueMutex.Unlock()
}

func (sc *SessionCache) lockAndGetQueue(sessionID int) ([]*model.QueuedSong, *redsync.Mutex) {
	queueMutex := sc.redsync.NewMutex(getQueueMutexKey(sessionID))
	queueMutex.Lock()

	var queue [] *model.QueuedSong
	err := sc.getStructFromRedis(getQueueKey(sessionID), &queue)

	if err != nil {
		slog.Warn("Error getting session queue", "error", err)
	}

	return queue, queueMutex
}

func (sc *SessionCache) setAndUnlockQueue(sessionID int, newQueue [] *model.QueuedSong, queueMutex *redsync.Mutex) {
	sc.setStructToRedis(getQueueKey(sessionID), newQueue)
	queueMutex.Unlock()
}

func getQueueMutexKey(sessionID int) string {
	return fmt.Sprintf("queue-mutex-%d", sessionID)
}

func getQueueKey(sessionID int) string {
	return fmt.Sprintf("queue-%d", sessionID)
}