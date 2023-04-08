package session

import (
	"time"
	
	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/constants"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"golang.org/x/exp/slog"

)

func (s *SessionServiceInMemory) sendUpdatedState(sessionID int) {
	go func() {
		var activeChannels []chan *model.SessionState

		s.allSessionsMutex.Lock()
		session := s.sessions[sessionID]
		s.allSessionsMutex.Unlock()

		session.channelMutex.Lock()
		channels := session.channels

		for _, ch := range channels {
			select {
			case ch <- session.sessionState: // This is the actual send.
				// Our message went through, do nothing
				activeChannels = append(activeChannels, ch)
			default: // This is run when our send does not work.
				slog.Info("Channel closed in update.")
				// You can handle any deregistration of the channel here.
			}
		}

		session.channels = activeChannels
		session.channelMutex.Unlock()

		session.expiryMutex.Lock()
		session.expiresAt = time.Now().Add(sessionTimeout * time.Minute)
		session.expiryMutex.Unlock()
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

func expireSession(session *Session) {
	session.expiryMutex.Lock()
	session.expiresAt = time.Now()
	session.expiryMutex.Unlock()
}

func closeChannels(session *Session) {
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

func (s *SessionServiceInMemory) watchSpotifyCurrentlyPlaying(sessionID int, accountService account.AccountService) {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	sendUpdateFlag := false
	addNextSongFlag := false
	advanceQueueFlag := false

	for {
		// TODO: Maybe make this refresh value dynamic to adjust refresh frequency at the end of a song
		time.Sleep(spotifyWatchFrequency * time.Millisecond)

		session.expiryMutex.Lock()
		sessionExpired := time.Now().After(session.expiresAt)
		session.expiryMutex.Unlock()

		if sessionExpired {
			slog.Info("Session has expired, ending session spotify watcher", "session_id", sessionID)
			return
		}

		sendUpdateFlag = false
		spotifyCurrentlyPlayingSong, spotifyCurrentlyPlaying, err := session.streaming.CurrentSong()
		if err != nil {
			slog.Warn("Error getting music player state", "error", err)
			continue
		}

		session.sessionStateMutex.Lock()
		if spotifyCurrentlyPlaying == true {
			if session.sessionState.CurrentlyPlaying.SimpleSong.ID != spotifyCurrentlyPlayingSong.SimpleSong.ID {
				// If song has changed, update currently playing, send update, and set flag to pop next song from queue
				session.sessionState.CurrentlyPlaying = spotifyCurrentlyPlayingSong
				sendUpdateFlag = true
				addNextSongFlag = true
			} else if session.sessionState.CurrentlyPlaying.Playing != spotifyCurrentlyPlaying {
				// If same song is paused and then played, set the new state
				session.sessionState.CurrentlyPlaying.Playing = spotifyCurrentlyPlaying
				sendUpdateFlag = true
			}

			// If the currently playing song is about to end, pop the top of the session and add to spotify queue
			// If go spotify client adds API for checking current queue, checking this is a better way to tell if it's
			// Safe to add song
			timeLeft, err := session.streaming.TimeRemaining()
			if err != nil {
				slog.Warn("Error getting song time remaining", "error", err)
				continue
			}

			if timeLeft < 5000 && addNextSongFlag {
				advanceQueueFlag = true
				sendUpdateFlag = true
				addNextSongFlag = false
			}
		} else {
			// Change currently playing to false if music gets paused
			if session.sessionState.CurrentlyPlaying.Playing != spotifyCurrentlyPlaying {
				session.sessionState.CurrentlyPlaying.Playing = spotifyCurrentlyPlaying
				sendUpdateFlag = true
			}
		}
		session.sessionStateMutex.Unlock()

		if advanceQueueFlag {
			s.AdvanceQueue(sessionID, false, accountService)
			advanceQueueFlag = false
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
		session.expiryMutex.Lock()
		sessionExpired := time.Now().After(session.expiresAt)
		session.expiryMutex.Unlock()

		if sessionExpired {
			slog.Info("Session has expired, ending session voter watcher", "session_id", sessionID)
			return
		}

		session.votersMutex.Lock()
		for _, voter := range session.voters {
			if voter.VoterType == constants.AdminVoterType {
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

		time.Sleep(voterWatchFrequency * time.Second)
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


		for sessionID := range sessionsToEnd {
			s.EndSession(sessionID, accountService)
		}
		sessionsToEnd = nil

		time.Sleep(sessionWatchFrequency * time.Second)
	}
}