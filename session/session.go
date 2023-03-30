package session

import (
	"sort"
	"sync"
	"time"

	"github.com/campbelljlowman/fazool-api/constants"
	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/musicplayer"
	"github.com/campbelljlowman/fazool-api/voter"
	"github.com/campbelljlowman/fazool-api/utils"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"

)

type Session struct {
	sessionInfo    *model.SessionInfo
	voters         map[string]*voter.Voter
	musicPlayer    musicplayer.MusicPlayer
	expiresAt      time.Time
	// Map of [song][account][votes]
	bonusVotes     map[string]map[int]int
	queueMutex     *sync.Mutex
	votersMutex    *sync.Mutex
	expiryMutex    *sync.Mutex
	bonusVoteMutex *sync.Mutex
}

type SessionServiceInMemory struct {
	sessions	map[int]*Session
}

// Session gets removed after being inactive for this long in minutes
const sessionTimeout time.Duration = 30
// Spotify gets watched by default at this frequency in milliseconds
const spotifyWatchFrequency time.Duration = 250
// Voters get watched at this frequency in seconds
const voterWatchFrequency time.Duration = 1

func NewSessionServiceInMemoryImpl() *SessionServiceInMemory{
	sessionInMemory := &SessionServiceInMemory{
		sessions: make(map[int]*Session),
	}
	return sessionInMemory
}

func (s *SessionServiceInMemory) CreateSession(adminAccountID int, accountLevel string, musicPlayer musicplayer.MusicPlayer) (int, error) {
	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		return 0, utils.LogAndReturnError("Error generating session ID", err)
	}

	maximumVoters := 0
	if accountLevel == constants.RegularAccountLevel {
		maximumVoters = 50
	}

	sessionInfo := &model.SessionInfo{
		ID: sessionID,
		CurrentlyPlaying: &model.CurrentlyPlayingSong{
			SimpleSong: &model.SimpleSong{},
			Playing:    false,
		},
		Queue: nil,
		AdminAccountID: adminAccountID,
		MaximumVoters: maximumVoters,
	}

	session := Session{
		sessionInfo:    sessionInfo,
		voters:         make(map[string]*voter.Voter),
		musicPlayer:    musicPlayer,
		expiresAt:      time.Now().Add(sessionTimeout * time.Minute),
		bonusVotes:     make(map[string]map[int]int),
		queueMutex:     &sync.Mutex{},
		votersMutex:    &sync.Mutex{},
		expiryMutex:    &sync.Mutex{},
		bonusVoteMutex: &sync.Mutex{},
	}

	s.sessions[sessionID] = &session

	return sessionID, nil
}

func (s *SessionServiceInMemory) WatchSpotifyCurrentlyPlaying(sessionID int, accountService account.AccountService) {
	session := s.sessions[sessionID]
	sendUpdateFlag := false
	addNextSongFlag := false

	for {
		if time.Now().After(session.expiresAt) {
			slog.Info("Session has expired, ending session spotify watcher", "session_id", session.sessionInfo.ID)
			return
		}

		sendUpdateFlag = false
		spotifyCurrentlyPlayingSong, spotifyCurrentlyPlaying, err := session.musicPlayer.CurrentSong()
		if err != nil {
			slog.Warn("Error getting music player state", "error", err)
			continue
		}

		if spotifyCurrentlyPlaying == true {
			if session.sessionInfo.CurrentlyPlaying.SimpleSong.ID != spotifyCurrentlyPlayingSong.SimpleSong.ID {
				// If song has changed, update currently playing, send update, and set flag to pop next song from queue
				session.sessionInfo.CurrentlyPlaying = spotifyCurrentlyPlayingSong
				sendUpdateFlag = true
				addNextSongFlag = true
			} else if session.sessionInfo.CurrentlyPlaying.Playing != spotifyCurrentlyPlaying {
				// If same song is paused and then played, set the new state
				session.sessionInfo.CurrentlyPlaying.Playing = spotifyCurrentlyPlaying
				sendUpdateFlag = true
			}

			// If the currently playing song is about to end, pop the top of the session and add to spotify queue
			// If go spotify client adds API for checking current queue, checking this is a better way to tell if it's
			// Safe to add song
			timeLeft, err := session.musicPlayer.TimeRemaining()
			if err != nil {
				slog.Warn("Error getting song time remaining", "error", err)
				continue
			}

			if timeLeft < 5000 && addNextSongFlag {
				s.AdvanceQueue(sessionID, false, accountService)

				sendUpdateFlag = true
				addNextSongFlag = false
			}
		} else {
			// Change currently playing to false if music gets paused
			if session.sessionInfo.CurrentlyPlaying.Playing != spotifyCurrentlyPlaying {
				session.sessionInfo.CurrentlyPlaying.Playing = spotifyCurrentlyPlaying
				sendUpdateFlag = true
			}
		}

		if sendUpdateFlag {
			s.refreshSession(sessionID)
			s.sendUpdate(sessionID)
		}

		// TODO: Maybe make this refresh value dynamic to adjust refresh frequency at the end of a song
		time.Sleep(spotifyWatchFrequency * time.Millisecond)
	}
}

func (s *SessionServiceInMemory) sendUpdate(sessionID int) {
	//fill this in
}

func (s *SessionServiceInMemory) refreshSession(sessionID int) {
	session := s.sessions[sessionID]
	session.expiresAt = time.Now().Add(sessionTimeout * time.Minute)
}

func (s *SessionServiceInMemory) AdvanceQueue(sessionID int, force bool, accountService account.AccountService) error { 
	session := s.sessions[sessionID]
	var song *model.SimpleSong

	session.queueMutex.Lock()
	if len(session.sessionInfo.Queue) == 0 {
		session.queueMutex.Unlock()
		return nil
	}

	song, session.sessionInfo.Queue = session.sessionInfo.Queue[0].SimpleSong, session.sessionInfo.Queue[1:]
	session.queueMutex.Unlock()

	err := session.musicPlayer.QueueSong(song.ID)
	if err != nil {
		return err
	}

	err = s.processBonusVotes(sessionID, song.ID, accountService)
	if err != nil {
		return err
	}

	if !force {
		return nil
	}

	err = session.musicPlayer.Next()
	if err != nil {
		return err
	}

	return nil
}

// TODO: Write function that watches voters and removes any inactive ones
func (s *SessionServiceInMemory) WatchVotersExpirations(sessionID int) {
	session := s.sessions[sessionID]

	for {
		if time.Now().After(session.expiresAt) {
			slog.Info("Session has expired, ending session voter watcher", "session_id", session.sessionInfo.ID)
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
			}

		}
		session.votersMutex.Unlock()

		time.Sleep(voterWatchFrequency * time.Second)
	}
}

// TODO: This code hasn't been tested
func (s *SessionServiceInMemory) processBonusVotes(sessionID int, songID string, accountService account.AccountService) error {
	session := s.sessions[sessionID]

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

func (s *SessionServiceInMemory) SetQueue(sessionID int, newQueue [] *model.QueuedSong) {
	session := s.sessions[sessionID]
	session.queueMutex.Lock()
	session.sessionInfo.Queue = newQueue
	session.queueMutex.Unlock()
}

func (s *SessionServiceInMemory) UpsertQueue(sessionID, vote int, song model.SongUpdate) {
	session := s.sessions[sessionID]
	session.queueMutex.Lock()
	idx := slices.IndexFunc(session.sessionInfo.Queue, func(s *model.QueuedSong) bool { return s.SimpleSong.ID == song.ID })
	if idx == -1 {
		// add new song to queue
		newSong := &model.QueuedSong{
			SimpleSong: &model.SimpleSong{
				ID:     song.ID,
				Title:  *song.Title,
				Artist: *song.Artist,
				Image:  *song.Image,
			},
			Votes: vote,
		}
		session.sessionInfo.Queue = append(session.sessionInfo.Queue, newSong)
	} else {
		queuedSong := session.sessionInfo.Queue[idx]
		queuedSong.Votes += vote
	}

	// Sort queue
	sort.Slice(session.sessionInfo.Queue, func(i, j int) bool { return session.sessionInfo.Queue[i].Votes > session.sessionInfo.Queue[j].Votes })
	session.queueMutex.Unlock()
	s.refreshSession(sessionID)
}

func (s *SessionServiceInMemory) UpsertVoterInSession(sessionID int, newVoter *voter.Voter){
	session := s.sessions[sessionID]
	session.votersMutex.Lock()
	session.voters[newVoter.VoterID] = newVoter
	session.votersMutex.Unlock()
	// TODO: update number of active voters in session
}

func (s *SessionServiceInMemory) GetVoterInSession(sessionID int, voterID string) (*voter.Voter, bool){
	session := s.sessions[sessionID]
	session.votersMutex.Lock()
	voter, exists := session.voters[voterID]
	session.votersMutex.Unlock()
	return voter, exists
} 

func (s *SessionServiceInMemory) IsSessionFull(sessionID int) bool {
	session := s.sessions[sessionID]
	isFull := false
	session.votersMutex.Lock()
	// TODO: Check the number of active voters
	if len(session.voters) >= session.sessionInfo.MaximumVoters  {
		isFull = true
	}
	session.votersMutex.Unlock()
	return isFull
}

func (s *SessionServiceInMemory) expireSession(sessionID int) {
	session := s.sessions[sessionID]
	session.expiryMutex.Lock()
	session.expiresAt = time.Now()
	session.expiryMutex.Unlock()
}

// func (s *SessionServiceInMemory) IsExpired() bool {
// 	s.expiryMutex.Lock()
// 	isExpired := s.expiresAt.Before(time.Now())
// 	s.expiryMutex.Unlock()
// 	return isExpired
// }

func (s *SessionServiceInMemory) AddBonusVote(songID string, accountID, numberOfVotes, sessionID int) {
	session := s.sessions[sessionID]
	session.bonusVoteMutex.Lock()
	if _, exists := session.bonusVotes[songID][accountID]; !exists {
		session.bonusVotes[songID] = make(map[int]int)
	}
	session.bonusVotes[songID][accountID] += numberOfVotes
	session.bonusVoteMutex.Unlock()
}

func (s *SessionServiceInMemory) DoesSessionExist(sessionID int) bool {
	_, exists := s.sessions[sessionID]
	return exists
}

func (s *SessionServiceInMemory) EndSession(sessionID int, accountService account.AccountService) {
	session := s.sessions[sessionID]
	slog.Info("Ending session!", "sessionID", session.sessionInfo.ID)
	accountService.SetAccountActiveSession(session.sessionInfo.AdminAccountID, 0)

	delete(s.sessions, session.sessionInfo.ID)

	s.expireSession(sessionID)	
}

func (s *SessionServiceInMemory) GetSessionAdminAccountID(sessionID int) int {
	session := s.sessions[sessionID]
	return session.sessionInfo.AdminAccountID
}

func (s *SessionServiceInMemory) GetSessionInfo(sessionID int) *model.SessionInfo {
	session := s.sessions[sessionID]
	return session.sessionInfo
}

func (s *SessionServiceInMemory) GetMusicPlayer(sessionID int) musicplayer.MusicPlayer {
	session := s.sessions[sessionID]
	return session.musicPlayer
}