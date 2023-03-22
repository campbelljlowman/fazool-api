package session

import (
	"fmt"

	"golang.org/x/exp/slog"
	"github.com/go-redsync/redsync/v4"

	"github.com/campbelljlowman/fazool-api/voter"

)

func (s *Session) initVoterMap(sessionID int) {
	voters := make(map[string]*voter.Voter)
	votersMutex := s.redsync.NewMutex(getVotersMutexKey(sessionID))
	votersMutex.Lock()
	err := s.setStructToRedis(getVotersKey(sessionID), voters)
	votersMutex.Unlock()

	if err != nil {
		slog.Warn("Error setting session voters", "error", err)
	}
}

func (s *Session) UpsertVoterInSession(sessionID int, newVoter *voter.Voter){
	voters, votersMutex := s.lockAndGetAllVotersInSession(sessionID)
	voters[newVoter.VoterID] = newVoter
	err := s.setStructToRedis(getVotersKey(sessionID), voters)

	votersMutex.Unlock()

	if err != nil {
		slog.Warn("Error setting session voters", "error", err)
	}

	s.setNumberOfVoters(sessionID, len(voters))
}

func (s *Session) GetVoterInSession(sessionID int, voterID string) (*voter.Voter, bool){
	voters, votersMutex := s.lockAndGetAllVotersInSession(sessionID)
	voter, exists := voters[voterID]

	votersMutex.Unlock()

	return voter, exists
}

func (s *Session) lockAndGetAllVotersInSession(sessionID int) (map[string]*voter.Voter, *redsync.Mutex) {
	votersMutex := s.redsync.NewMutex(getVotersMutexKey(sessionID))
	votersMutex.Lock()

	var voters map[string] *voter.Voter
	err := s.getStructFromRedis(getVotersKey(sessionID), &voters)

	if err != nil {
		slog.Warn("Error getting session voters", "error", err)
	}

	return voters, votersMutex
}

func (s *Session) setAndUnlockAllVotersInSession(sessionID int, voters map[string]*voter.Voter, votersMutex *redsync.Mutex) {
	err := s.setStructToRedis(getVotersMutexKey(sessionID), voters)
	if err != nil {
		slog.Warn("Error setting voters", "error", err)
	}
	votersMutex.Unlock()
}

func (s *Session) getNumberOfVoters(sessionID int) int {
	voterCountMutex := s.redsync.NewMutex(getVoterCountMutexKey(sessionID))
	voterCountMutex.Lock()

	var voterCount int
	err := s.getStructFromRedis(getVoterCountKey(sessionID), &voterCount)

	if err != nil {
		slog.Warn("Error getting session voterCount", "error", err)
	}

	voterCountMutex.Unlock()

	return voterCount	
}

func (s *Session) setNumberOfVoters(sessionID, voterCount int) {
	voterCountMutex := s.redsync.NewMutex(getVoterCountMutexKey(sessionID))
	voterCountMutex.Lock()

	err := s.setStructToRedis(getVoterCountKey(sessionID), voterCount)

	if err != nil {
		slog.Warn("Error getting session voterCount", "error", err)
	}

	voterCountMutex.Unlock()

}

func getVotersMutexKey(sessionID int) string {
	return fmt.Sprintf("voters-mutex-%d", sessionID)
}

func getVotersKey(sessionID int) string {
	return fmt.Sprintf("voters-%d", sessionID)
}

func getVoterCountMutexKey(sessionID int) string {
	return fmt.Sprintf("voter-count-mutex-%d", sessionID)
}

func getVoterCountKey(sessionID int) string {
	return fmt.Sprintf("voter-count-%d", sessionID)
}
