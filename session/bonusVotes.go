package session

import (
	"fmt"

	"golang.org/x/exp/slog"
	"github.com/go-redsync/redsync/v4"

)


func (s *Session) AddBonusVote(songID string, accountID, numberOfVotes, sessionID int) {
	bonusVotes, bonusVoteMutex := s.lockAndGetBonusVotes(sessionID)

	if _, exists := bonusVotes[songID][accountID]; !exists {
		bonusVotes[songID] = make(map[int]int)
	}
	bonusVotes[songID][accountID] += numberOfVotes

	err := s.setStructToRedis(getBonusVoteKey(sessionID), bonusVotes)
	bonusVoteMutex.Unlock()

	if err != nil {
		slog.Warn("Error updating bonus votes", "error", err)
	}
}


func (s *Session) lockAndGetBonusVotes(sessionID int) (map[string]map[int]int, *redsync.Mutex) {
	bonusVoteMutex := s.redsync.NewMutex(fmt.Sprintf("bonus-vote-mutex-%d", sessionID))
	// Map of [song][account][votes]
	var bonusVotes map[string]map[int]int

	bonusVoteMutex.Lock()
	err := s.getStructFromRedis(getBonusVoteKey(sessionID), &bonusVotes)

	if err != nil {
		slog.Warn("Error getting session bonus votes", "error", err)
	}

	return bonusVotes, bonusVoteMutex
}

func (s *Session) setAndUnlockBonusVotes(sessionID int, bonusVotes map[string]map[int]int, bonusVoteMutex *redsync.Mutex) {
	err := s.setStructToRedis(getBonusVoteKey(sessionID), bonusVotes)
	if err != nil {
		slog.Warn("Error setting bonus votes:", "error", err)
	}
	bonusVoteMutex.Unlock()
}

func getBonusVoteKey(sessionID int) string {
	return fmt.Sprintf("bonus-vote-%d", sessionID)
}