package session

import (
	"fmt"

	"golang.org/x/exp/slog"
	"github.com/go-redsync/redsync/v4"

)


func (sc *Session) AddBonusVote(songID, accountID string, numberOfVotes, sessionID int) {
	bonusVotes, bonusVoteMutex := sc.lockAndGetBonusVotes(sessionID)

	if _, exists := bonusVotes[songID][accountID]; !exists {
		bonusVotes[songID] = make(map[string]int)
	}
	bonusVotes[songID][accountID] += numberOfVotes

	err := sc.setStructToRedis(getBonusVoteKey(sessionID), bonusVotes)
	bonusVoteMutex.Unlock()

	if err != nil {
		slog.Warn("Error updating bonus votes", "error", err)
	}
}


func (sc *Session) lockAndGetBonusVotes(sessionID int) (map[string]map[string]int, *redsync.Mutex) {
	bonusVoteMutex := sc.redsync.NewMutex(fmt.Sprintf("bonus-vote-mutex-%d", sessionID))
	// Map of [song][account][votes]
	var bonusVotes map[string]map[string]int

	bonusVoteMutex.Lock()
	err := sc.getStructFromRedis(getBonusVoteKey(sessionID), &bonusVotes)

	if err != nil {
		slog.Warn("Error getting session bonus votes", "error", err)
	}

	return bonusVotes, bonusVoteMutex
}

func getBonusVoteKey(sessionID int) string {
	return fmt.Sprintf("bonus-vote-%d", sessionID)
}