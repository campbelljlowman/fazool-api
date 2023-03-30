package session

import (
	"time"
	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/graph/model"
)

func (s *SessionServiceInMemory) sendUpdate(sessionID int) {
	//fill this in
}

func (s *SessionServiceInMemory) refreshSession(sessionID int) {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	session.expiryMutex.Lock()
	session.expiresAt = time.Now().Add(sessionTimeout * time.Minute)
	session.expiryMutex.Unlock()
}

// TODO: This code hasn't been tested
func (s *SessionServiceInMemory) processBonusVotes(sessionID int, songID string, accountService account.AccountService) error {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()


	session.bonusVoteMutex.Lock()
	songBonusVotes, exists := session.bonusVotes[songID]
	delete(session.bonusVotes, songID)
	session.bonusVoteMutex.Unlock()

	if !exists {
		return nil
	}

	for accountID, votes := range songBonusVotes {
		accountService.SubtractBonusVotes(accountID, votes)
	}

	return nil
}

func (s *SessionServiceInMemory) expireSession(sessionID int) {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	session.expiryMutex.Lock()
	session.expiresAt = time.Now()
	session.expiryMutex.Unlock()
}

func (s *SessionServiceInMemory) setQueue(sessionID int, newQueue [] *model.QueuedSong) {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	session.queueMutex.Lock()
	session.sessionInfo.Queue = newQueue
	session.queueMutex.Unlock()
}