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
	CreateSession(adminAccountID int, accountType model.AccountType, streaming streaming.StreamingService) (int, error)

	GetSessionConfig(sessionID int) *model.SessionConfig
	GetSessionState(sessionID int) *model.SessionState
	GetSessionAdminAccountID(sessionID int) int
	GetVoterInSession(sessionID int, voterID string) (*voter.Voter, bool)
	GetPlaylists(sessionID int) ([]*model.Playlist, error)
	IsSessionFull(sessionID int) bool
	DoesSessionExist(sessionID int) bool
	SearchForSongs(sessionID int, query string) ([]*model.SimpleSong, error)

	UpsertQueue(sessionID, numberOfVotes int, song model.SongUpdate)
	UpsertVoterInSession(sessionID int, voter *voter.Voter)
	UpdateCurrentlyPlaying(sessionID int, action model.QueueAction) error
	PopQueue(sessionID int) error
	AddVoterToSession(sessionID, accountID int, voterID string) (*model.Voter, error)
	AddBonusVote(songID string, accountID, numberOfVotes, sessionID int)
	AddChannel(sessionID int, channel chan *model.SessionState)
	SetPlaylist(sessionID int, playlist string) error
	RefreshVoterExpiration(sessionID int, voterID string)

	EndSession(sessionID int)
}

type session struct {
	sessionConfig    		*model.SessionConfig
	sessionState    		*model.SessionState
	channels 				[]chan *model.SessionState
	voters         			map[string]*voter.Voter
	streaming    			streaming.StreamingService
	expiresAt      			time.Time
	// Map of [song][account][votes]
	bonusVotes     			map[string]map[int]int
	streamingServiceUpdater	chan string
	sessionStateMutex		*sync.Mutex
	channelMutex 			*sync.Mutex
	votersMutex    			*sync.Mutex
	expiryMutex    			*sync.Mutex
	bonusVoteMutex 			*sync.Mutex
}

type SessionServiceInMemory struct {
	sessions			map[int]*session
	allSessionsMutex 	*sync.Mutex
	accountService 		account.AccountService
}

//lint:file-ignore ST1011 Ignore rule for time.Duration unit in variable name
const sessionWatchFrequencySeconds time.Duration = 10 
const sessionTimeoutMinutes time.Duration = 30 
const streamingServiceWatchFrequencySlowMilliseconds time.Duration = 4000
const streamingServiceWatchFrequencyFastMilliseconds time.Duration = 250
const voterWatchFrequencySeconds time.Duration = 5

func NewSessionServiceInMemoryImpl(accountService account.AccountService) *SessionServiceInMemory{
	sessionInMemory := &SessionServiceInMemory{
		sessions: 			make(map[int]*session),
		allSessionsMutex: 	&sync.Mutex{},
		accountService: 	accountService,
	}

	go sessionInMemory.watchSessions()
	return sessionInMemory
}

func (s *SessionServiceInMemory) CreateSession(adminAccountID int, accountType model.AccountType, streaming streaming.StreamingService) (int, error) {
	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		return 0, err
	}

	maximumVoters := 0
	if accountType == model.AccountTypeFree {
		maximumVoters = 50
	}

	sessionConfig := &model.SessionConfig{
		SessionID: sessionID,
		AdminAccountID: adminAccountID,
		MaximumVoters: maximumVoters,
	}

	sessionState := &model.SessionState{
		CurrentlyPlaying: &model.CurrentlyPlayingSong{
			SimpleSong: &model.SimpleSong{},
			IsPlaying:    false,
			SongProgressSeconds: 0,
			SongDurationSeconds: 0,
		},
		Queue: nil,
		NumberOfVoters: 0,
	}

	session := session{
		sessionConfig:    			sessionConfig,
		sessionState: 				sessionState,
		channels: 					nil,
		voters:         			make(map[string]*voter.Voter),
		streaming:   				streaming,
		expiresAt:      			time.Now().Add(sessionTimeoutMinutes * time.Minute),
		bonusVotes:     			make(map[string]map[int]int),
		streamingServiceUpdater:	make(chan string),
		sessionStateMutex:			&sync.Mutex{},
		channelMutex: 				&sync.Mutex{},
		votersMutex:    			&sync.Mutex{},
		expiryMutex:    			&sync.Mutex{},
		bonusVoteMutex: 			&sync.Mutex{},
	}

	s.allSessionsMutex.Lock()
	s.sessions[sessionID] = &session
	s.allSessionsMutex.Unlock()

	go s.watchStreamingServiceCurrentlyPlaying(sessionID)
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

	session.sessionStateMutex.Lock()
	if session.sessionState.NumberOfVoters >= session.sessionConfig.MaximumVoters  {
		isFull = true
	}
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

func (s *SessionServiceInMemory) UpsertQueue(sessionID, numberOfVotes int, song model.SongUpdate) {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	session.sessionStateMutex.Lock()
	index := slices.IndexFunc(session.sessionState.Queue, func(s *model.QueuedSong) bool { return s.SimpleSong.ID == song.ID })
	if index == -1 {
		// add new song to queue
		newSong := &model.QueuedSong{
			SimpleSong: &model.SimpleSong{
				ID:     song.ID,
				Title:  *song.Title,
				Artist: *song.Artist,
				Image:  *song.Image,
			},
			Votes: numberOfVotes,
		}
		session.sessionState.Queue = append(session.sessionState.Queue, newSong)
	} else {
		queuedSong := session.sessionState.Queue[index]
		queuedSong.Votes += numberOfVotes
	}

	sort.Slice(session.sessionState.Queue, func(i, j int) bool { return session.sessionState.Queue[i].Votes > session.sessionState.Queue[j].Votes })
	session.sessionStateMutex.Unlock()
	s.sendUpdatedState(sessionID)
}

func (s *SessionServiceInMemory) UpsertVoterInSession(sessionID int, voter *voter.Voter){
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()

	session.votersMutex.Lock()
	session.voters[voter.VoterID] = voter
	numberOfVoters := len(session.voters)
	session.votersMutex.Unlock()

	session.sessionStateMutex.Lock()
	session.sessionState.NumberOfVoters = numberOfVoters
	session.sessionStateMutex.Unlock()
	s.sendUpdatedState(sessionID)
}

func (s *SessionServiceInMemory) UpdateCurrentlyPlaying(sessionID int, action model.QueueAction) error {
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
		err := s.PopQueue(sessionID)
		if err != nil {
			return err
		}
		err = session.streaming.Next()
		if err != nil {
			return err
		}
	}

	// Do this twice bc ADVANCE updates streaming service too slow for the first one to get the new state
	session.streamingServiceUpdater <- "Update!"
	session.streamingServiceUpdater <- "Update!"

	return nil
}

func (s *SessionServiceInMemory) PopQueue(sessionID int) error { 
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

	err = s.processBonusVotes(sessionID, song.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *SessionServiceInMemory) AddVoterToSession(sessionID, accountID int, voterID string) (*model.Voter, error) {
	voterType := model.VoterTypeFree
	bonusVotes := 0

	if accountID != 0 {
		superVoterSession, bonusVotesValue := s.accountService.GetSuperVoterSessionsAndBonusVotes(accountID)
		bonusVotes = bonusVotesValue

		if superVoterSession == sessionID {
			voterType = model.VoterTypeSuper
		}
		if s.GetSessionAdminAccountID(sessionID) == accountID {
			voterType = model.VoterTypeAdmin
		}
	}

	newVoter, err := voter.NewVoter(voterID, voterType, accountID, bonusVotes)
	if err != nil {
		return nil, utils.LogAndReturnError("Error generating new voter", nil)
	}

	slog.Debug("New voter created:", "voter", newVoter)
	s.UpsertVoterInSession(sessionID, newVoter)

	return newVoter.ConvertVoterType(), nil
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

func (s *SessionServiceInMemory) RefreshVoterExpiration(sessionID int, voterID string) {
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	s.allSessionsMutex.Unlock()
	
	session.votersMutex.Lock()
	v := session.voters[voterID]
	v.ExpiresAt = time.Now().Add(voter.GetVoterDuration(v.VoterType) * time.Minute)
	session.votersMutex.Unlock()

}

func (s *SessionServiceInMemory) EndSession(sessionID int) {
	slog.Info("Ending session!", "sessionID", sessionID)
	
	s.allSessionsMutex.Lock()
	session := s.sessions[sessionID]
	delete(s.sessions, sessionID)

	s.accountService.SetAccountActiveSession(session.sessionConfig.AdminAccountID, 0)

	expireSession(session)
	closeChannels(session)

	s.allSessionsMutex.Unlock()
}
