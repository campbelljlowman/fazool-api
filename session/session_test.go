package session

import (
	"testing"
	"time"

	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/voter"
	"github.com/campbelljlowman/fazool-api/mocks"
	"golang.org/x/exp/slices"
	"github.com/golang/mock/gomock"
)

// TODO: see if there's enough common code creating songs to pull into helper function

func TestNewSessionServiceInMemoryImpl(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockAccountService := mocks.NewMockAccountService(ctrl)

	_ = NewSessionServiceInMemoryImpl(mockAccountService)
}

var CreateSessionTests = []struct {
	adminAccountID 			int
	accountType 			model.AccountType
	expectedMaximumVoters 	int
}{
	{123, model.AccountTypeFree, 50},
	{456, model.AccountTypeSmallVenue, 0},
	{789, model.AccountTypeLargeVenue, 0},
}
func TestCreateSession(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)

	for _, testCase := range(CreateSessionTests) {
		sessionID, err := sessionService.CreateSession(testCase.adminAccountID, testCase.accountType, mockStreamingService, mockAccountService)
		if err != nil {
			t.Errorf("CreateSession() failed! Got an error: %v", err)
		}
		
		sessionConfig := sessionService.GetSessionConfig(sessionID)
		if sessionConfig.MaximumVoters != testCase.expectedMaximumVoters {
			t.Errorf("CreateSession() failed! Wanted maximum voters %v, got: %v", testCase.expectedMaximumVoters, sessionConfig.MaximumVoters)
		}
		// TODO: assert watchers were called
	}
}

func TestGetSessionConfig(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	sessionConfig := sessionService.GetSessionConfig(sessionID)
	if sessionConfig.AdminAccountID != 123 {
		t.Errorf("GetSessionConfig() failed! Wanted %v, got: %v", 123, sessionConfig.AdminAccountID)
	}
}

func TestGetSessionState(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	sessionState := sessionService.GetSessionState(sessionID)
	if sessionState.NumberOfVoters != 0 {
		t.Errorf("GetSessionState() failed! Wanted %v, got: %v", 0, sessionState.NumberOfVoters)
	}
}

func TestGetSessionAdminAccountID(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	adminAccountID := sessionService.GetSessionAdminAccountID(sessionID)
	if adminAccountID != 123 {
		t.Errorf("GetSessionAdminAccountID() failed! Wanted %v, got: %v", 123, adminAccountID)
	}
}

func TestGetVoterInSession(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	newVoter, _ := voter.NewVoter("asdf", model.VoterTypeFree, 123, 0)
	sessionService.UpsertVoterInSession(sessionID, newVoter)

	voterInSession, exists := sessionService.GetVoterInSession(sessionID, "asdf")
	if !exists {
		t.Errorf("GetVoterInSession() failed! Wanted voter to exist")
	}
	if voterInSession != newVoter {
		t.Errorf("GetVoterInSession() failed! Wanted %v, got: %v", newVoter, voterInSession)
	}

	_, exists = sessionService.GetVoterInSession(sessionID, "nonexistent key")
	if exists {
		t.Errorf("GetVoterInSession() failed! Wanted voter to not exist")
	}
}

func TestGetPlaylists(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}
	mockStreamingService.EXPECT().GetPlaylists()

	sessionService.GetPlaylists(sessionID)
}

var IsSessionFullTests = []struct {
	numberOfVoters	int
	maximumVoters 	int
	expectedIsFull 	bool
}{
	{10, 20, false},
	{30, 20, true},
	{20, 20, true},
}
func TestIsSessionFull(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	for _, testCase := range(IsSessionFullTests) {
		sessionService.sessions[sessionID].sessionConfig.MaximumVoters = testCase.maximumVoters
		sessionService.sessions[sessionID].sessionStateMutex.Lock()
		sessionService.sessions[sessionID].sessionState.NumberOfVoters = testCase.numberOfVoters
		sessionService.sessions[sessionID].sessionStateMutex.Unlock()
	
		isFull := sessionService.IsSessionFull(sessionID)
		if isFull != testCase.expectedIsFull {
			t.Errorf("IsSessionFull() failed! Wanted %v, got: %v", testCase.expectedIsFull, isFull)
		}
	}
}

func TestDoesSessionExist(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	exists := sessionService.DoesSessionExist(sessionID)
	if exists != true {
		t.Errorf("DoesSessionExist() failed! Wanted %v, got: %v", true, exists)
	}

	exists = sessionService.DoesSessionExist(123)
	if exists != false {
		t.Errorf("DoesSessionExist() failed! Wanted %v, got: %v", false, exists)
	}
}

func TestSearchForSongs(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}
	mockStreamingService.EXPECT().Search("query")

	sessionService.SearchForSongs(sessionID, "query")
}

var title = "the boogie"
var artist = "charles"
var image = "Image url"
var UpsertQueueTests = []struct {
	songToQueue				*model.SongUpdate
	votesToAdd 				int
	expectedNumberOfVotes 	int
}{
	{
		&model.SongUpdate{
			ID: "id",
			Title: &title,
			Artist: &artist,
			Image: &image,
		}, 1, 1,
	},
	{	&model.SongUpdate{
			ID: "id",
		}, 2, 3,
	},
}
func TestUpsertQueue(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	for _, testCase := range(UpsertQueueTests) {
		sessionService.UpsertQueue(sessionID, testCase.votesToAdd, *testCase.songToQueue)

		sessionState := sessionService.GetSessionState(sessionID)
		index := slices.IndexFunc(sessionState.Queue, func(s *model.QueuedSong) bool { return s.SimpleSong.ID == testCase.songToQueue.ID })
		songInQueue := sessionState.Queue[index]
		if songInQueue.Votes != testCase.expectedNumberOfVotes {
			t.Errorf("UpsertQueue() failed! Wanted votes %v, got: %v", testCase.expectedNumberOfVotes, songInQueue.Votes)	
		}
	}
}

func TestUpsertVoterInSession(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	newVoter, _ := voter.NewVoter("asdf", model.VoterTypeFree, 123, 0)
	sessionService.UpsertVoterInSession(sessionID, newVoter)

	voterInSession, exists := sessionService.GetVoterInSession(sessionID, "asdf")
	if !exists {
		t.Errorf("UpsertVoterInSesion() failed! Wanted voter to exist")
	}
	if voterInSession != newVoter {
		t.Errorf("UpsertVoterInSesion() failed! Wanted %v, got: %v", newVoter, voterInSession)
	}

	sessionState := sessionService.GetSessionState(sessionID)
	if sessionState.NumberOfVoters != 1 {
		t.Errorf("UpsertVoterInSesion() failed! Wanted number of voters %v, got: %v", 1, sessionState.NumberOfVoters)
	}
}

func TestUpdateCurrentlyPlaying(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}
	mockStreamingService.EXPECT().CurrentSong().AnyTimes()

	mockStreamingService.EXPECT().Play()
	sessionService.UpdateCurrentlyPlaying(sessionID, model.QueueActionPlay, mockAccountService)

	mockStreamingService.EXPECT().Pause()
	sessionService.UpdateCurrentlyPlaying(sessionID, model.QueueActionPause, mockAccountService)

	mockStreamingService.EXPECT().Next()
	sessionService.UpdateCurrentlyPlaying(sessionID, model.QueueActionAdvance, mockAccountService)
}

func TestPopQueue(t *testing.T) {
	// TODO: assert function calls to streaming service
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	var queue []*model.QueuedSong
	songToQueue := &model.QueuedSong{
		SimpleSong: &model.SimpleSong{
			ID: "id",
			Title: "asdf",
			Artist: "qwer",
			Image: "qewr",
		},
		Votes: 1,
	}
	queue = append(queue, songToQueue)
	sessionService.setQueue(sessionID, queue)
	mockStreamingService.EXPECT().QueueSong("id")

	sessionService.PopQueue(sessionID, mockAccountService)

	sessionState := sessionService.GetSessionState(sessionID)
	if len(sessionState.Queue) != 0 {
		t.Errorf("PopQueue() failed! Wanted length of queue %v, got: %v", 0, len(sessionState.Queue))	
	}

	sessionService.PopQueue(sessionID, mockAccountService)

	sessionState = sessionService.GetSessionState(sessionID)
	if len(sessionState.Queue) != 0 {
		t.Errorf("PopQueue() failed! Wanted length of queue %v, got: %v", 0, len(sessionState.Queue))	
	}
}

func TestAddBonusVote(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	sessionService.AddBonusVote("asdf", 123, 2, sessionID)
}

func TestAddChannel(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}
	
	newChannel := make(chan *model.SessionState)
	sessionService.AddChannel(sessionID, newChannel)
}

func TestSetPlaylist(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	mockStreamingService.EXPECT().GetSongsInPlaylist("playlist")
	sessionService.SetPlaylist(sessionID, "playlist")
}

func TestRefreshVoterExpiration(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	newVoter, _ := voter.NewVoter("voter-id", model.VoterTypeFree, 123, 0)
	sessionService.UpsertVoterInSession(sessionID, newVoter)

	sessionService.RefreshVoterExpiration(sessionID, "voter-id")

	if time.Now().After(newVoter.ExpiresAt) { 
		t.Errorf("RefreshVoterExpiration() failed! voter has expired")	
	}
}

func TestEndSession(t *testing.T) {
	accountID := 123
	// TODO: assert cleanup functions are called
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(accountID, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}
	mockAccountService.EXPECT().SetAccountActiveSession(accountID, 0)

	sessionService.EndSession(sessionID, mockAccountService)

	sessionExists := sessionService.DoesSessionExist(sessionID)
	if sessionExists {
		t.Errorf("EndSession() failed! session still exists")	
	}
}

func newTestingServices(t *testing.T) (*mocks.MockAccountService, *mocks.MockStreamingService, SessionServiceInMemory) {
	ctrl := gomock.NewController(t)

	mockAccountService := mocks.NewMockAccountService(ctrl)
	mockStreamingService := mocks.NewMockStreamingService(ctrl)
	sessionService := NewSessionServiceInMemoryImpl(mockAccountService)
	return mockAccountService, mockStreamingService, *sessionService 
}