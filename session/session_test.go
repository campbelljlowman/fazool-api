package session

import (
	"testing"

	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/streaming"
	"github.com/golang/mock/gomock"
)

func TestNewSessionServiceInMemoryImpl(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockAccountService := NewMockAccountService(ctrl)

	mockAccountService.EXPECT().Bar(gomock.Eq(99)).Return(101)

  	// SUT(m)

	// accountService := account.NewAccountServiceMockImpl()
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

func newTestingServices() (account.AccountService, streaming.StreamingService, SessionServiceInMemory) {
	accountService := account.NewAccountServiceMockImpl()
	streamingService := streaming.NewMockStreamingServiceClient("asdf")
	sessionService := NewSessionServiceInMemoryImpl(accountService)
	return accountService, streamingService, *sessionService 
}