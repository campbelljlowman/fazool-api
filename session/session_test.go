package session

import (
	"testing"

	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/streaming"
	"github.com/campbelljlowman/fazool-api/voter"
	// "github.com/golang/mock/gomock"
)

// func TestNewSessionServiceInMemoryImpl(t *testing.T) {
// 	ctrl := gomock.NewController(t)

// 	mockAccountService := NewMockAccountService(ctrl)

// 	mockAccountService.EXPECT().Bar(gomock.Eq(99)).Return(101)

//   	// SUT(m)

// 	// accountService := account.NewAccountServiceMockImpl()
// 	_ = NewSessionServiceInMemoryImpl(mockAccountService)
// }

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
	accountService, streamingService, sessionService := newTestingServices()

	for _, testCase := range(CreateSessionTests) {
		sessionID, err := sessionService.CreateSession(testCase.adminAccountID, testCase.accountType, streamingService, accountService)
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
	accountService, streamingService, sessionService := newTestingServices()

	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, streamingService, accountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	sessionConfig := sessionService.GetSessionConfig(sessionID)
	if sessionConfig.AdminAccountID != 123 {
		t.Errorf("GetSessionConfig() failed! Wanted %v, got: %v", 123, sessionConfig.AdminAccountID)
	}
}

func TestGetSessionState(t *testing.T) {
	accountService, streamingService, sessionService := newTestingServices()

	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, streamingService, accountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	sessionState := sessionService.GetSessionState(sessionID)
	if sessionState.NumberOfVoters != 0 {
		t.Errorf("GetSessionState() failed! Wanted %v, got: %v", 0, sessionState.NumberOfVoters)
	}
}

func TestGetSessionAdminAccountID(t *testing.T) {
	accountService, streamingService, sessionService := newTestingServices()

	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, streamingService, accountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	adminAccountID := sessionService.GetSessionAdminAccountID(sessionID)
	if adminAccountID != 123 {
		t.Errorf("GetSessionAdminAccountID() failed! Wanted %v, got: %v", 123, adminAccountID)
	}
}

func TestGetVoterInSession(t *testing.T) {
	accountService, streamingService, sessionService := newTestingServices()

	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, streamingService, accountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	newVoter, err := voter.NewVoter("asdf", model.VoterTypeFree, 123, 0)
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

// func TestGetPlaylists(t testing.T) {
// 	// TODO: Figure out how to mock calls to streaming service
// }

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
	accountService, streamingService, sessionService := newTestingServices()

	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, streamingService, accountService)
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
	accountService, streamingService, sessionService := newTestingServices()

	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, streamingService, accountService)
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

// func TestSearchForSongs(t testing.T) {
// 	// TODO: Figre out how to mock calls to streaming service
// }

// func TestUpsertQueue(t testing.T) {
// 	accountService, streamingService, sessionService := newTestingServices()

// 	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, streamingService, accountService)
// 	if err != nil {
// 		t.Errorf("CreateSession() failed! Got an error: %v", err)
// 	}

// 	title := "the boogie"
// 	artist := "charles"
// 	image := "Image url"
// 	newSong := &model.SongUpdate{
// 		ID: "id",
// 		Title: &title,
// 		Artist: &artist,
// 		Image: &image,
// 		// Action: model.SongVoteActionAdd,
// 	}
// 	sessionService.UpsertQueue(sessionID, 1, *newSong)
// 	sessionState := sessionService.GetSessionState(sessionID)
// 	if sessionState.Queue
// }

func newTestingServices() (account.AccountService, streaming.StreamingService, SessionServiceInMemory) {
	accountService := account.NewAccountServiceMockImpl()
	streamingService := streaming.NewMockStreamingServiceClient("asdf")
	sessionService := NewSessionServiceInMemoryImpl(accountService)
	return accountService, streamingService, *sessionService 
}