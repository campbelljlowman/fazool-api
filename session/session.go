package session

import (
	"sort"
	"sync"
	"time"

	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/constants"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/streamingService"
	"github.com/campbelljlowman/fazool-api/utils"
	"github.com/campbelljlowman/fazool-api/voter"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

type SessionService interface {
	CreateSession(adminAccountID int, accountLevel string, streamingService streamingService.StreamingService) (int, error)

	GetSessionInfo(sessionID int) *model.SessionInfo
	GetSessionAdminAccountID(sessionID int) int
	GetVoterInSession(sessionID int, voterID string) (*voter.Voter, bool)
	GetPlaylists(sessionID int) ([]*model.Playlist, error)
	IsSessionFull(sessionID int) bool
	DoesSessionExist(sessionID int) bool
	SearchForSongs(sessionID int, query string) ([]*model.SimpleSong, error)

	UpsertQueue(sessionID, vote int, song model.SongUpdate)
	UpsertVoterInSession(sessionID int, newVoter *voter.Voter)
	UpdateCurrentlyPlaying(sessionID int, action model.QueueAction, accountService account.AccountService) error
	AdvanceQueue(sessionID int, force bool, accountService account.AccountService) error
	AddBonusVote(songID string, accountID, numberOfVotes, sessionID int)
	SetPlaylist(sessionID int, playlist string) error

	WatchSpotifyCurrentlyPlaying(sessionID int, accountService account.AccountService)
	WatchSessions(accountService account.AccountService)
	WatchVotersExpirations(sessionID int)

	EndSession(sessionID int, accountService account.AccountService)
}

type Session struct {
	sessionInfo    		*model.SessionInfo
	voters         		map[string]*voter.Voter
	streamingService    streamingService.StreamingService
	expiresAt      		time.Time
	// Map of [song][account][votes]
	bonusVotes     		map[string]map[int]int
	queueMutex     		*sync.Mutex
	votersMutex    		*sync.Mutex
	expiryMutex    		*sync.Mutex
	bonusVoteMutex 		*sync.Mutex
}

type SessionServiceInMemory struct {
	sessions			map[int]*Session
	allSessionsMutex 	*sync.Mutex
}

const sessionWatchFrequency time.Duration = 10 // Seconds
const sessionTimeout time.Duration = 30 // Minutes
const spotifyWatchFrequency time.Duration = 250 // Milliseconds
const voterWatchFrequency time.Duration = 1 // Seconds

func NewSessionServiceInMemoryImpl() *SessionServiceInMemory{
	sessionInMemory := &SessionServiceInMemory{
		sessions: 			make(map[int]*Session),
		allSessionsMutex: 	&sync.Mutex{},
	}
	return sessionInMemory
}

func (s *SessionServiceInMemory) CreateSession(adminAccountID int, accountLevel string, streamingService streamingService.StreamingService) (int, error) {
	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		return 0, err
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
		sessionInfo:    	sessionInfo,
		voters:         	make(map[string]*voter.Voter),
		streamingService:   streamingService,
		expiresAt:      	time.Now().Add(sessionTimeout * time.Minute),
		bonusVotes:     	make(map[string]map[int]int),
		queueMutex:     	&sync.Mutex{},
		votersMutex:    	&sync.Mutex{},
		expiryMutex:    	&sync.Mutex{},
		bonusVoteMutex: 	&sync.Mutex{},
	}

	s.allSessionsMutex.Lock()
	s.sessions[sessionID] = &session
	s.allSessionsMutex.Unlock()
	
	return sessionID, nil
}

func (s *SessionServiceInMemory) GetSessionInfo(sessionID int) *model.SessionInfo {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	return session.sessionInfo
}

func (s *SessionServiceInMemory) GetSessionAdminAccountID(sessionID int) int {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	return session.sessionInfo.AdminAccountID
}

func (s *SessionServiceInMemory) GetVoterInSession(sessionID int, voterID string) (*voter.Voter, bool){
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	session.votersMutex.Lock()
	voter, exists := session.voters[voterID]
	session.votersMutex.Unlock()
	return voter, exists
}

func (s *SessionServiceInMemory) GetPlaylists(sessionID int) ([]*model.Playlist, error){
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	return session.streamingService.GetPlaylists()
}

func (s *SessionServiceInMemory) IsSessionFull(sessionID int) bool {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	isFull := false
	session.votersMutex.Lock()
	// TODO: Check the number of active voters
	if len(session.voters) >= session.sessionInfo.MaximumVoters  {
		isFull = true
	}
	session.votersMutex.Unlock()
	return isFull
}

func (s *SessionServiceInMemory) DoesSessionExist(sessionID int) bool {
	s.allSessionsMutex.Lock()
	_, exists := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	return exists
}

func (s *SessionServiceInMemory) SearchForSongs(sessionID int, query string) ([]*model.SimpleSong, error){
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	return session.streamingService.Search(query)
}

func (s *SessionServiceInMemory) UpsertQueue(sessionID, vote int, song model.SongUpdate) {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

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
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	session.votersMutex.Lock()
	session.voters[newVoter.VoterID] = newVoter
	session.votersMutex.Unlock()
	// TODO: update number of active voters in session
}

func (s *SessionServiceInMemory) UpdateCurrentlyPlaying(sessionID int, action model.QueueAction, accountService account.AccountService) error {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	switch action {
	case "PLAY":
		err := session.streamingService.Play()
		if err != nil {
			return err
		}
	case "PAUSE":
		err := session.streamingService.Pause()
		if err != nil {
			return err
		}
	case "ADVANCE":
		err := s.AdvanceQueue(sessionID, true, accountService)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SessionServiceInMemory) AdvanceQueue(sessionID int, force bool, accountService account.AccountService) error { 
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	var song *model.SimpleSong

	session.queueMutex.Lock()
	if len(session.sessionInfo.Queue) == 0 {
		session.queueMutex.Unlock()
		return nil
	}

	song, session.sessionInfo.Queue = session.sessionInfo.Queue[0].SimpleSong, session.sessionInfo.Queue[1:]
	session.queueMutex.Unlock()

	err := session.streamingService.QueueSong(song.ID)
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

	err = session.streamingService.Next()
	if err != nil {
		return err
	}

	return nil
}

func (s *SessionServiceInMemory) AddBonusVote(songID string, accountID, numberOfVotes, sessionID int) {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	session.bonusVoteMutex.Lock()
	if _, exists := session.bonusVotes[songID][accountID]; !exists {
		session.bonusVotes[songID] = make(map[int]int)
	}
	session.bonusVotes[songID][accountID] += numberOfVotes
	session.bonusVoteMutex.Unlock()
}

func (s *SessionServiceInMemory) SetPlaylist(sessionID int, playlist string) error {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	songs, err := session.streamingService.GetSongsInPlaylist(playlist)
	if err != nil {
		return err
	}

	var songsToQueue []*model.QueuedSong
	for _, song := range songs {
		songToQueue := &model.QueuedSong{
			SimpleSong: song,
			Votes:      0,
		}
		songsToQueue = append(songsToQueue, songToQueue)
	}

	s.setQueue(sessionID, songsToQueue)
	return nil
}


func (s *SessionServiceInMemory) WatchSpotifyCurrentlyPlaying(sessionID int, accountService account.AccountService) {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	sendUpdateFlag := false
	addNextSongFlag := false

	for {
		if time.Now().After(session.expiresAt) {
			slog.Info("Session has expired, ending session spotify watcher", "session_id", session.sessionInfo.ID)
			return
		}

		sendUpdateFlag = false
		spotifyCurrentlyPlayingSong, spotifyCurrentlyPlaying, err := session.streamingService.CurrentSong()
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
			timeLeft, err := session.streamingService.TimeRemaining()
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

func (s *SessionServiceInMemory) WatchSessions(accountService account.AccountService) {
	var sessionsToEnd []int

	for {

		s.allSessionsMutex.Lock()
		for sessionID := range s.sessions {
			if s.isExpired(sessionID) {
				sessionsToEnd = append(sessionsToEnd, sessionID)
				s.EndSession(sessionID, accountService)
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

// TODO: Write function that watches voters and removes any inactive ones
func (s *SessionServiceInMemory) WatchVotersExpirations(sessionID int) {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()


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

func (s *SessionServiceInMemory) EndSession(sessionID int, accountService account.AccountService) {
	slog.Info("Ending session!", "sessionID", sessionID)
	
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	delete(s.sessions, sessionID)
	s.allSessionsMutex.Unlock()

	accountService.SetAccountActiveSession(session.sessionInfo.AdminAccountID, 0)

	s.expireSession(sessionID)	
}
