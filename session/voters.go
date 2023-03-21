package session

import (
	"fmt"

	"golang.org/x/exp/slog"
	"github.com/go-redsync/redsync/v4"

	"github.com/campbelljlowman/fazool-api/voter"

)

func (sc *Session) initVoterMap(sessionID int) {
	voters := make(map[string]*voter.Voter)
	votersMutex := sc.redsync.NewMutex(getVotersMutexKey(sessionID))
	votersMutex.Lock()
	err := sc.setStructToRedis(getVotersKey(sessionID), voters)
	votersMutex.Unlock()

	if err != nil {
		slog.Warn("Error setting session voters", "error", err)
	}
}

func (sc *Session) UpsertVoterInSession(sessionID int, newVoter *voter.Voter){
	voters, votersMutex := sc.lockAndGetAllVotersInSession(sessionID)
	voters[newVoter.VoterID] = newVoter
	err := sc.setStructToRedis(getVotersKey(sessionID), voters)

	votersMutex.Unlock()

	if err != nil {
		slog.Warn("Error setting session voters", "error", err)
	}

	sc.setNumberOfVoters(sessionID, len(voters))
}

func (sc *Session) GetVoterInSession(sessionID int, voterID string) (*voter.Voter, bool){
	voters, votersMutex := sc.lockAndGetAllVotersInSession(sessionID)
	voter, exists := voters[voterID]

	votersMutex.Unlock()

	return voter, exists
}

func (sc *Session) lockAndGetAllVotersInSession(sessionID int) (map[string]*voter.Voter, *redsync.Mutex) {
	votersMutex := sc.redsync.NewMutex(getVotersMutexKey(sessionID))
	votersMutex.Lock()

	var voters map[string] *voter.Voter
	err := sc.getStructFromRedis(getVotersKey(sessionID), &voters)

	if err != nil {
		slog.Warn("Error getting session voters", "error", err)
	}

	return voters, votersMutex
}

func (sc *Session) GetNumberOfVoters(sessionID int) int {
	voterCountMutex := sc.redsync.NewMutex(getVoterCountMutexKey(sessionID))
	voterCountMutex.Lock()

	var voterCount int
	err := sc.getStructFromRedis(getVoterCountKey(sessionID), &voterCount)

	if err != nil {
		slog.Warn("Error getting session voterCount", "error", err)
	}

	voterCountMutex.Unlock()

	return voterCount	
}

func (sc *Session) setNumberOfVoters(sessionID, voterCount int) {
	voterCountMutex := sc.redsync.NewMutex(getVoterCountMutexKey(sessionID))
	voterCountMutex.Lock()

	err := sc.setStructToRedis(getVoterCountKey(sessionID), voterCount)

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
