package session

import (
	"testing"
	// "time"

	"github.com/campbelljlowman/fazool-api/graph/model"
	// "github.com/golang/mock/gomock"
)

func TestSendUpdatedState(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	newChannel := make(chan *model.SessionState)
	sessionService.AddChannel(sessionID, newChannel)
	sessionService.sendUpdatedState(sessionID)

	// Wait for goroutine to run before asserting
	// time.Sleep(1 * time.Second)

	// TODO: Assert session expiry was refreshed
}

func TestProcessBonusVotes(t *testing.T) {
	accountID := 123
	numberOfBonusVotes := 1

	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(accountID, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}
	
	sessionService.AddBonusVote("song1", accountID, numberOfBonusVotes, sessionID)
	mockAccountService.EXPECT().SubtractBonusVotes(accountID, numberOfBonusVotes)

	sessionService.processBonusVotes(sessionID, "song1", mockAccountService)
	sessionService.processBonusVotes(sessionID, "song2", mockAccountService)

	session := sessionService.sessions[sessionID]
	_, bonusVoteExists := session.bonusVotes["song1"]

	if bonusVoteExists {
		t.Errorf("processBonusVotes() failed! Song is still present in bonus votes map")	
	}
}

// func TestExpireSession(t *testing.T) {
// 	// TODO: Try to mock account service function call for SetAccountActiveSession.
// 	// Session watcher can try to end session after the expire session function is called
// 	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
// 	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
// 	if err != nil {
// 		t.Errorf("CreateSession() failed! Got an error: %v", err)
// 	}
// 	// mockAccountService.EXPECT().SetAccountActiveSession(gomock.Any(), gomock.Any())
// 	session := sessionService.sessions[sessionID]

// 	expireSession(session)
// 	sessionExpired := time.Now().After(session.expiresAt)

// 	if !sessionExpired {
// 		t.Errorf("expireSession() failed! Session isn't expired")	
// 	}
// }

func TestCloseChannels(t *testing.T) {
	mockAccountService, mockStreamingService, sessionService := newTestingServices(t)
	sessionID, err := sessionService.CreateSession(123, model.AccountTypeFree, mockStreamingService, mockAccountService)
	if err != nil {
		t.Errorf("CreateSession() failed! Got an error: %v", err)
	}

	newChannel := make(chan *model.SessionState)
	sessionService.AddChannel(sessionID, newChannel)

	session := sessionService.sessions[sessionID]
	closeChannels(session)

	_, isChannelOpen := <- newChannel
	if isChannelOpen {
		t.Errorf("closeChannels() failed! channel isn't closed")	
	}
}

func TestSetQueue(t *testing.T){
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
	session := sessionService.sessions[sessionID]

	if len(session.sessionState.Queue) != 1 {
		t.Errorf("setQueue() failed! wanted queue length %v, got: %v", 1, len(session.sessionState.Queue))	
	}
}