package session

import (
	"time"
	
	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/voter"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"golang.org/x/exp/slog"

)

func (s *SessionServiceInMemory) sendUpdatedState(sessionID int) {
	go func(){
		var activeChannels []chan *model.SessionState

		s.allSessionsMutex.Lock()
		session := s.sessions[sessionID]
		s.allSessionsMutex.Unlock()

		session.expiryMutex.Lock()
		session.expiresAt = time.Now().Add(sessionTimeoutMinutes * time.Minute)
		session.expiryMutex.Unlock()

		session.channelMutex.Lock()
		channels := session.channels

		for _, ch := range channels {
			select {
			case ch <- session.sessionState: 
				slog.Debug("Sent update")
				activeChannels = append(activeChannels, ch)
			case  <-time.After(100 * time.Millisecond):
				slog.Debug("Waiting for channel to become unblocked timed out")
			}
		}

		session.channels = activeChannels
		session.channelMutex.Unlock()
	}()
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

func cleanupSuperVoters(sessionID int, voter *voter.Voter) {
	
}

func expireSession(session *session) {
	session.expiryMutex.Lock()
	session.expiresAt = time.Now()
	session.expiryMutex.Unlock()
}

func closeChannels(session *session) {
	session.channelMutex.Lock()
	for _, ch := range session.channels {
		close(ch)
	}
	session.channelMutex.Unlock()
}

func (s *SessionServiceInMemory) setQueue(sessionID int, newQueue [] *model.QueuedSong) {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	session.sessionStateMutex.Lock()
	session.sessionState.Queue = newQueue
	session.sessionStateMutex.Unlock()

	s.sendUpdatedState(sessionID)
}

func (s *SessionServiceInMemory) watchStreamingServiceCurrentlyPlaying(sessionID int, accountService account.AccountService) {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	sendUpdateFlag := false
	addNextSongFlag := true
	popQueueFlag := false
	//lint:file-ignore ST1011 Ignore rule for time.Duration unit in variable name
	streamingServiceWatchFrequencyMilliseconds := streamingServiceWatchFrequencySlowMilliseconds

	for {
		// Refresh value is dynamic to increase the sensitivity when the song is about to change
		if !addNextSongFlag {
			streamingServiceWatchFrequencyMilliseconds = streamingServiceWatchFrequencyFastMilliseconds
		} else {
			streamingServiceWatchFrequencyMilliseconds = streamingServiceWatchFrequencySlowMilliseconds
		}

		select{
		case <-time.After(streamingServiceWatchFrequencyMilliseconds * time.Millisecond):
		case <-session.streamingServiceUpdater:
		}

		session.expiryMutex.Lock()
		sessionExpired := time.Now().After(session.expiresAt)
		session.expiryMutex.Unlock()

		if sessionExpired {
			slog.Info("Session has expired, ending session streaming service watcher", "session_id", sessionID)
			return
		}

		sendUpdateFlag = false
		currentlyPlayingSong, isCurrentlyPlaying, err := session.streaming.CurrentSong()
		if err != nil {
			slog.Warn("Error getting music player state", "error", err)
			continue
		}
	
		session.sessionStateMutex.Lock()
		if isCurrentlyPlaying {
			if session.sessionState.CurrentlyPlaying.SimpleSong.ID != currentlyPlayingSong.SimpleSong.ID {
				addNextSongFlag = true
			}

			session.sessionState.CurrentlyPlaying = currentlyPlayingSong
			sendUpdateFlag = true

			// If the currently playing song is about to end, pop the top of the session and add to streaming service queue
			// If go streaming service clients adds API for checking current queue, checking this is a better way to tell if it's
			// Safe to add song
			timeLeft, err := session.streaming.TimeRemaining()
			if err != nil {
				slog.Warn("Error getting song time remaining", "error", err)
				continue
			}
	
			if timeLeft < 5000 && addNextSongFlag {
				popQueueFlag = true
				sendUpdateFlag = true
				addNextSongFlag = false
			}
		} else {
			// Change currently playing to false if music gets paused
			if session.sessionState.CurrentlyPlaying.IsPlaying != isCurrentlyPlaying {
				session.sessionState.CurrentlyPlaying.IsPlaying = isCurrentlyPlaying
				sendUpdateFlag = true
			}
		}
		session.sessionStateMutex.Unlock()
	
		if popQueueFlag {
			s.PopQueue(sessionID, accountService)
			popQueueFlag = false
		}
		if sendUpdateFlag {
			s.sendUpdatedState(sessionID)
		}
	}
}

func (s *SessionServiceInMemory) watchVotersExpirations(sessionID int) {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()


	for {
		time.Sleep(voterWatchFrequencySeconds * time.Second)

		session.expiryMutex.Lock()
		sessionExpired := time.Now().After(session.expiresAt)
		session.expiryMutex.Unlock()

		if sessionExpired {
			slog.Info("Session has expired, ending session voter watcher", "session_id", sessionID)
			return
		}

		sessionIsFull := s.IsSessionFull(sessionID)
		if !sessionIsFull {
			slog.Debug("Session isn't full, skipping voter check", "sessionID", sessionID)
			continue
		}

		session.votersMutex.Lock()
		for _, voter := range session.voters {
			if voter.VoterType == model.VoterTypeAdmin {
				continue
			}
			
			if time.Now().After(voter.ExpiresAt) {
				slog.Info("Voter exipred! Removing", "voter", voter.VoterID)
				delete(session.voters, voter.VoterID)

				session.sessionStateMutex.Lock()
				session.sessionState.NumberOfVoters--
				session.sessionStateMutex.Unlock()
				s.sendUpdatedState(sessionID)
			}

		}
		session.votersMutex.Unlock()
	}
}

func (s *SessionServiceInMemory) watchSessions(accountService account.AccountService) {
	var sessionsToEnd []int

	for {
		s.allSessionsMutex.Lock()

		for sessionID, session := range s.sessions {
			session.expiryMutex.Lock()
			sessionExpired := time.Now().After(session.expiresAt)
			session.expiryMutex.Unlock()

			if sessionExpired {
				sessionsToEnd = append(sessionsToEnd, sessionID)
			}
		}

		s.allSessionsMutex.Unlock()


		for _, sessionID := range sessionsToEnd {
			s.EndSession(sessionID, accountService)
		}
		sessionsToEnd = nil

		time.Sleep(sessionWatchFrequencySeconds * time.Second)
	}
}