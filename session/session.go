package session

import (
	"sort"
	"sync"
	"time"

	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/streaming"
	"github.com/campbelljlowman/fazool-api/utils"
	"github.com/campbelljlowman/fazool-api/voter"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

type SessionService interface {
	CreateSession(adminAccountID int, accountType model.AccountType, streaming streaming.StreamingService, accountService account.AccountService) (int, error)

	GetSessionConfig(sessionID int) *model.SessionConfig
	GetSessionState(sessionID int) *model.SessionState
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
	AddChannel(sessionID int, channel chan *model.SessionState)
	SetPlaylist(sessionID int, playlist string) error

	EndSession(sessionID int, accountService account.AccountService)
}

type Session struct {
	sessionConfig    	*model.SessionConfig
	sessionState    	*model.SessionState
	channels 			[]chan *model.SessionState
	voters         		map[string]*voter.Voter
	streaming    streaming.StreamingService
	expiresAt      		time.Time
	// Map of [song][account][votes]
	bonusVotes     		map[string]map[int]int
	sessionStateMutex	*sync.Mutex
	channelMutex 		*sync.Mutex
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
const spotifyWatchFrequency time.Duration = 1000 // Milliseconds
const voterWatchFrequency time.Duration = 1 // Seconds

func NewSessionServiceInMemoryImpl(accountService account.AccountService) *SessionServiceInMemory{
	sessionInMemory := &SessionServiceInMemory{
		sessions: 			make(map[int]*Session),
		allSessionsMutex: 	&sync.Mutex{},
	}

	go sessionInMemory.watchSessions(accountService)
	return sessionInMemory
}

func (s *SessionServiceInMemory) CreateSession(adminAccountID int, accountType model.AccountType, streaming streaming.StreamingService, accountService account.AccountService) (int, error) {
	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		return 0, err
	}

	maximumVoters := 0
	if accountType == model.AccountTypeFree {
		maximumVoters = 50
	}

	sessionConfig := &model.SessionConfig{
		ID: sessionID,
		AdminAccountID: adminAccountID,
		MaximumVoters: maximumVoters,
	}

	sessionState := &model.SessionState{
		CurrentlyPlaying: &model.CurrentlyPlayingSong{
			SimpleSong: &model.SimpleSong{},
			Playing:    false,
		},
		Queue: nil,
		NumberOfVoters: 0,
	}

	session := Session{
		sessionConfig:    	sessionConfig,
		sessionState: 		sessionState,
		channels: 			nil,
		voters:         	make(map[string]*voter.Voter),
		streaming:   streaming,
		expiresAt:      	time.Now().Add(sessionTimeout * time.Minute),
		bonusVotes:     	make(map[string]map[int]int),
		sessionStateMutex:	&sync.Mutex{},
		channelMutex: 		&sync.Mutex{},
		votersMutex:    	&sync.Mutex{},
		expiryMutex:    	&sync.Mutex{},
		bonusVoteMutex: 	&sync.Mutex{},
	}

	s.allSessionsMutex.Lock()
	s.sessions[sessionID] = &session
	s.allSessionsMutex.Unlock()

	go s.watchSpotifyCurrentlyPlaying(sessionID, accountService)
	go s.watchVotersExpirations(sessionID)
	
	
	return sessionID, nil
}

func (s *SessionServiceInMemory) GetSessionConfig(sessionID int) *model.SessionConfig {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	return session.sessionConfig
}

func (s *SessionServiceInMemory) GetSessionState(sessionID int) *model.SessionState {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	session.sessionStateMutex.Lock()
	sessionState := session.sessionState
	session.sessionStateMutex.Unlock()

	return sessionState
}

func (s *SessionServiceInMemory) GetSessionAdminAccountID(sessionID int) int {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	return session.sessionConfig.AdminAccountID
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

	return session.streaming.GetPlaylists()
}

func (s *SessionServiceInMemory) IsSessionFull(sessionID int) bool {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	isFull := false

	session.votersMutex.Lock()
	session.sessionStateMutex.Lock()
	if session.sessionState.NumberOfVoters >= session.sessionConfig.MaximumVoters  {
		isFull = true
	}
	session.votersMutex.Unlock()
	session.sessionStateMutex.Unlock()

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

	return session.streaming.Search(query)
}

func (s *SessionServiceInMemory) UpsertQueue(sessionID, vote int, song model.SongUpdate) {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	session.sessionStateMutex.Lock()
	idx := slices.IndexFunc(session.sessionState.Queue, func(s *model.QueuedSong) bool { return s.SimpleSong.ID == song.ID })
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
		session.sessionState.Queue = append(session.sessionState.Queue, newSong)
	} else {
		queuedSong := session.sessionState.Queue[idx]
		queuedSong.Votes += vote
	}

	sort.Slice(session.sessionState.Queue, func(i, j int) bool { return session.sessionState.Queue[i].Votes > session.sessionState.Queue[j].Votes })
	session.sessionStateMutex.Unlock()
	s.sendUpdatedState(sessionID)
}

func (s *SessionServiceInMemory) UpsertVoterInSession(sessionID int, newVoter *voter.Voter){
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	session.votersMutex.Lock()
	session.voters[newVoter.VoterID] = newVoter
	numberOfVoters := len(session.voters)
	session.votersMutex.Unlock()

	session.sessionStateMutex.Lock()
	session.sessionState.NumberOfVoters = numberOfVoters
	session.sessionStateMutex.Unlock()
	s.sendUpdatedState(sessionID)
}

func (s *SessionServiceInMemory) UpdateCurrentlyPlaying(sessionID int, action model.QueueAction, accountService account.AccountService) error {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	switch action {
	case "PLAY":
		err := session.streaming.Play()
		if err != nil {
			return err
		}
	case "PAUSE":
		err := session.streaming.Pause()
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

	session.sessionStateMutex.Lock()
	if len(session.sessionState.Queue) == 0 {
		session.sessionStateMutex.Unlock()
		return nil
	}

	song, session.sessionState.Queue = session.sessionState.Queue[0].SimpleSong, session.sessionState.Queue[1:]
	session.sessionStateMutex.Unlock()

	s.sendUpdatedState(sessionID)

	err := session.streaming.QueueSong(song.ID)
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

	err = session.streaming.Next()
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

func (s *SessionServiceInMemory) AddChannel(sessionID int, channel chan *model.SessionState) {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()
	
	session.channelMutex.Lock()
	session.channels = append(session.channels, channel)
	session.channelMutex.Unlock()
}


func (s *SessionServiceInMemory) SetPlaylist(sessionID int, playlist string) error {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	songs, err := session.streaming.GetSongsInPlaylist(playlist)
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

func (s *SessionServiceInMemory) EndSession(sessionID int, accountService account.AccountService) {
	slog.Info("Ending session!", "sessionID", sessionID)
	
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	delete(s.sessions, sessionID)

	accountService.SetAccountActiveSession(session.sessionConfig.AdminAccountID, 0)

	expireSession(session)
	closeChannels(session)

	s.allSessionsMutex.Unlock()
}
