package session

import (
	"testing"
	"time"

	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/streaming"
	"github.com/campbelljlowman/fazool-api/voter"
	"golang.org/x/exp/slices"
	// "github.com/golang/mock/gomock"
)

// TODO: see if there's enough common code creating songs to pull into helper function

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
	accountService, streamingService, sessionService := newTestingServices()
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, streamingService, accountService)
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
	accountService, streamingService, sessionService := newTestingServices()
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, streamingService, accountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	newVoter, err := voter.NewVoter("asdf", model.VoterTypeFree, 123, 0)
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

// func TestUpdateCurrentlyPlaying (t *testing.T) {
// 	// TODO: figure out how to mock the streaming calls
// }

func TestPopQueue(t *testing.T) {
	// TODO: assert function calls to streaming service
	accountService, streamingService, sessionService := newTestingServices()
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, streamingService, accountService)
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

	sessionService.PopQueue(sessionID, accountService)

	sessionState := sessionService.GetSessionState(sessionID)
	if len(sessionState.Queue) != 0 {
		t.Errorf("PopQueue() failed! Wanted length of queue %v, got: %v", 0, len(sessionState.Queue))	
	}

	sessionService.PopQueue(sessionID, accountService)

	sessionState = sessionService.GetSessionState(sessionID)
	if len(sessionState.Queue) != 0 {
		t.Errorf("PopQueue() failed! Wanted length of queue %v, got: %v", 0, len(sessionState.Queue))	
	}
}

func TestAddBonusVote(t *testing.T) {
	accountService, streamingService, sessionService := newTestingServices()
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, streamingService, accountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	sessionService.AddBonusVote("asdf", 123, 2, sessionID)
}

func TestAddChannel(t *testing.T) {
	accountService, streamingService, sessionService := newTestingServices()
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, streamingService, accountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}
	
	newChannel := make(chan *model.SessionState)
	sessionService.AddChannel(sessionID, newChannel)
}

// func TestSePlaylist(t *testing.T) {
// 	// TODO: Need to figure out how to mock return value for playlist session function
// }

func TestRefreshVoterExpiration(t *testing.T) {
	accountService, streamingService, sessionService := newTestingServices()
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, streamingService, accountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	newVoter, err := voter.NewVoter("voter-id", model.VoterTypeFree, 123, 0)
	sessionService.UpsertVoterInSession(sessionID, newVoter)

	sessionService.RefreshVoterExpiration(sessionID, "voter-id")

	if time.Now().After(newVoter.ExpiresAt) { 
		t.Errorf("RefreshVoterExpiration() failed! voter has expired")	
	}
}

// func TestEndSession(t *testing.T) {
// 	// TODO: assert cleanup functions are called and mock db call
// 	accountService, streamingService, sessionService := newTestingServices()
// 	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, streamingService, accountService)
// 	if err != nil {
// 		t.Errorf("CreateSession() failed! Got an error: %v", err)
// 	}

// 	sessionService.EndSession(sessionID, accountService)

// 	sessionExists := sessionService.DoesSessionExist(sessionID)
// 	if sessionExists {
// 		t.Errorf("EndSession() failed! session still exists")	
// 	}
// }

func newTestingServices() (account.AccountService, streaming.StreamingService, SessionServiceInMemory) {
	accountService := account.NewAccountServiceMockImpl()
	streamingService := streaming.NewMockStreamingServiceClient("asdf")
	sessionService := NewSessionServiceInMemoryImpl(accountService)
	return accountService, streamingService, *sessionService 
}