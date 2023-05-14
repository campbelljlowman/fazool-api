package session

import (
	"testing"

	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/streaming"
)

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

// TODO: Need to find out how to mock the session, this will make things easier

// func TestIsSessionFull(t *testing.T) {
// 	accountService, streamingService, sessionService := newTestingServices()

// 	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, streamingService, accountService)
// 	if err != nil {
// 		t.Errorf("CreateSession() failed! Got an error: %v", err)
// 	}

// 	sessionState := sessionService.GetSessionState(sessionID)
// }

func newTestingServices() (account.AccountService, streaming.StreamingService, SessionService) {
	accountService := account.NewAccountServiceMockImpl()
	streamingService := streaming.NewMockStreamingServiceClient("asdf")
	sessionService := NewSessionServiceInMemoryImpl(accountService)
	return accountService, streamingService, sessionService 
}