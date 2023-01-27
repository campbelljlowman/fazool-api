package voter

import (
	"testing"
	"time"

	"github.com/campbelljlowman/fazool-api/constants"
	"github.com/campbelljlowman/fazool-api/graph/model"
)

var two = 2
var GetVoterInfoTests = []struct {
	inputVoter Voter
	expectedOutputVoter model.VoterInfo
}{
	{Voter{
		ID: "asdf",
		VoterType: constants.RegularVoterType,
		ExpiresAt: time.Now(),
		songsUpVoted: map[string]struct{}{"song1": emptyStructValue},
		songsDownVoted:  make(map[string]struct{}),
		bonusVotes: 2,
	}, model.VoterInfo{
		Type: constants.RegularVoterType,
		SongsUpVoted: []string{"song1"},
		SongsDownVoted: []string{},
		BonusVotes: &two,
	}},
}
func TestGetVoterInfo(t *testing.T){
	for _, testCase := range(GetVoterInfoTests) {
		resultVoter := testCase.inputVoter.GetVoterInfo()

		if resultVoter.Type != testCase.expectedOutputVoter.Type {
			t.Errorf("GetVoterInfo() Type failed! Wanted: %v, got: %v", testCase.expectedOutputVoter.Type, resultVoter.Type)

		}

		if resultVoter.SongsUpVoted[0] != testCase.expectedOutputVoter.SongsUpVoted[0] {
			t.Errorf("GetVoterInfo() SongsUpVoted failed! Wanted: %v, got: %v", testCase.expectedOutputVoter.SongsUpVoted, resultVoter.SongsUpVoted)

		}

		if len(resultVoter.SongsDownVoted) != len(testCase.expectedOutputVoter.SongsDownVoted) {
			t.Errorf("GetVoterInfo() SongsDownVoted failed! Wanted: %v, got: %v", testCase.expectedOutputVoter.SongsDownVoted, resultVoter.SongsDownVoted)

		}

		if *resultVoter.BonusVotes != *testCase.expectedOutputVoter.BonusVotes {
			t.Errorf("GetVoterInfo() BonusVotes failed! Wanted: %v, got: %v", testCase.expectedOutputVoter.BonusVotes, resultVoter.BonusVotes)

		}
	}
}

// var CalculateAndProcessVoteTests = []struct {
// 	inputVoteDirection model.SongVoteDirection
// 	inputVoteAction model.SongVoteAction
// }{
// 	{model.SongVoteDirectionUp, model.SongVoteActionAdd},
// 	{model.SongVoteDirectionDown, model.SongVoteActionAdd},
// 	{model.SongVoteDirectionUp, model.SongVoteActionAdd},
// 	{model.SongVoteDirectionDown, model.SongVoteActionRemove},
// }
// func TestCalculateAndProcessVote(t *testing.T){
// 	for _, testCase := range(GetVoteAmountAndTypeTests) {
// 		// Assert correct function was called
// 		print(testCase.inputVoteAction)
// 		resulteVote, resultBonusVote, err := testCase.inputVoter.GetVoteAmountAndType(testCase.inputSong, &testCase.inputVoteDirection, &testCase.inputVoteAction)

// 		if resultVoter != &testCase.expectedOutputVoter {
// 			t.Errorf("GetVoterInfo() failed! Wanted: %v, got: %v", testCase.expectedOutputVoter, resultVoter)

// 		}

// 	}
// }


var calculateAndAddUpVoteTests = []struct {
	voterType string
	songsUpVoted map[string]struct{}
	songsDownVoted map[string]struct{}
	bonusVotes int
	songVotingFor string
	expectedVoteAmount int
	expectedBonusVoteValue bool
	expectedError bool
}{
	// Test voter types for regular vote
	{constants.RegularVoterType, map[string]struct{}{}, map[string]struct{}{},  0, "song1", 1, false, false},
	{constants.PrivilegedVoterType, map[string]struct{}{}, map[string]struct{}{}, 0, "song1", 2, false, false},
	{constants.AdminVoterType, map[string]struct{}{}, map[string]struct{}{}, 0, "song1", 1, false, false},
	// Test if song already exists 
	{constants.RegularVoterType, map[string]struct{}{"song1": emptyStructValue}, map[string]struct{}{}, 0, "song1", 0, false, true},
	{constants.RegularVoterType, map[string]struct{}{"song1": emptyStructValue}, map[string]struct{}{}, 1, "song1", 1, true, false},
	{constants.AdminVoterType, map[string]struct{}{"song1": emptyStructValue}, map[string]struct{}{}, 0, "song1", 1, false, false},
	// Test vote adjustment
	{constants.RegularVoterType, map[string]struct{}{}, map[string]struct{}{"song1": emptyStructValue}, 0, "song1", 2, false, false},
	{constants.PrivilegedVoterType, map[string]struct{}{}, map[string]struct{}{"song1": emptyStructValue}, 0, "song1", 3, false, false},
	{constants.AdminVoterType, map[string]struct{}{}, map[string]struct{}{"song1": emptyStructValue}, 0, "song1", 1, false, false},
}
func TestCalculateAndAddUpVote(t *testing.T){
	for _, testCase := range(calculateAndAddUpVoteTests) {
		voter, _ := NewVoter("ID1", testCase.voterType, testCase.bonusVotes)
		voter.songsUpVoted = testCase.songsUpVoted
		voter.songsDownVoted = testCase.songsDownVoted

		resulteVoteAmount, isResultBonusVote, err := voter.calculateAndAddUpVote(testCase.songVotingFor)

		if err != nil && !testCase.expectedError{
			t.Errorf("calculateAndAddUpVote() failed! Got an error: %v", err)
		}

		if resulteVoteAmount != testCase.expectedVoteAmount {
			t.Errorf("calculateAndAddUpVote() failed! Wanted vote amount: %v, got: %v", testCase.expectedVoteAmount, resulteVoteAmount)

		}

		if isResultBonusVote != testCase.expectedBonusVoteValue {
			t.Errorf("calculateAndAddUpVote() failed! Wanted bonus vote: %v, got: %v", testCase.expectedBonusVoteValue, isResultBonusVote)

		}

		_, upVoteExists := voter.songsUpVoted[testCase.songVotingFor]
		if !upVoteExists {
			t.Errorf("calculateAndAddUpVote() failed! Song not in map of upvoted songs")
		}

		_, downVoteExists := voter.songsDownVoted[testCase.songVotingFor]
		if downVoteExists {
			t.Errorf("calculateAndAddUpVote() failed! Song in map of downvoted songs")
		}
	}
}

var calculateAndAddDownVoteTests = []struct {
	voterType string
	songsUpVoted map[string]struct{}
	songsDownVoted map[string]struct{}
	songVotingFor string
	expectedVoteAmount int
	expectedError bool
}{
	// Test voter types for regular vote
	{constants.RegularVoterType, map[string]struct{}{}, map[string]struct{}{}, "song1", -1, false},
	{constants.PrivilegedVoterType, map[string]struct{}{}, map[string]struct{}{}, "song1", -1, false},
	{constants.AdminVoterType, map[string]struct{}{}, map[string]struct{}{}, "song1", -1, false},
	// Test if song already exists 
	{constants.RegularVoterType, map[string]struct{}{}, map[string]struct{}{"song1": emptyStructValue}, "song1", 0, true},
	{constants.PrivilegedVoterType, map[string]struct{}{}, map[string]struct{}{"song1": emptyStructValue}, "song1", 0, true},
	{constants.AdminVoterType, map[string]struct{}{}, map[string]struct{}{"song1": emptyStructValue}, "song1", -1, false},
	// Test vote adjustment
	{constants.RegularVoterType, map[string]struct{}{"song1": emptyStructValue}, map[string]struct{}{}, "song1", -2, true},
	{constants.PrivilegedVoterType, map[string]struct{}{"song1": emptyStructValue}, map[string]struct{}{}, "song1", -3, true},
	{constants.AdminVoterType, map[string]struct{}{"song1": emptyStructValue}, map[string]struct{}{}, "song1", -1, false},
}

func TestCalculateAndAddDownVote(t *testing.T){
	for _, testCase := range(calculateAndAddDownVoteTests) {
		voter, _ := NewVoter("ID1", testCase.voterType, 0)
		voter.songsUpVoted = testCase.songsUpVoted
		voter.songsDownVoted = testCase.songsDownVoted

		resulteVoteAmount, isResultBonusVote, err := voter.calculateAndAddDownVote(testCase.songVotingFor)

		if err != nil && !testCase.expectedError{
			t.Errorf("calculateAndAddDownVote() failed! Got an error: %v", err)
		}

		if resulteVoteAmount != testCase.expectedVoteAmount {
			t.Errorf("calculateAndAddDownVote() failed! Wanted vote amount: %v, got: %v", testCase.expectedVoteAmount, resulteVoteAmount)

		}

		if isResultBonusVote != false {
			t.Errorf("calculateAndAddDownVote() failed! Wanted bonus vote: %v, got: %v", false, isResultBonusVote)

		}

		_, downVoteExists := voter.songsDownVoted[testCase.songVotingFor]
		if !downVoteExists {
			t.Errorf("calculateAndAddUpVote() failed! Song not in map of downvoted songs")
		}

		_, upVoteExists := voter.songsUpVoted[testCase.songVotingFor]
		if upVoteExists {
			t.Errorf("calculateAndAddUpVote() failed! Song in map of upvoted songs")
		}
	}
}


var calculateAndRemoveUpVoteTests = []struct {
	voterType string
	songsUpVoted map[string]struct{}
	songVotingFor string
	expectedVoteAmount int
}{
	// Test voter types for regular vote
	{constants.RegularVoterType, map[string]struct{}{"song1": emptyStructValue}, "song1", -1},
	{constants.PrivilegedVoterType, map[string]struct{}{"song1": emptyStructValue}, "song1", -2},
	{constants.AdminVoterType, map[string]struct{}{"song1": emptyStructValue},"song1", -1},
}
func TestCalculateAndRemoveUpVote(t *testing.T){
	for _, testCase := range(calculateAndRemoveUpVoteTests) {
		voter, _ := NewVoter("ID1", testCase.voterType, 0)
		voter.songsUpVoted = testCase.songsUpVoted

		resulteVoteAmount, isResultBonusVote, err := voter.calculateAndRemoveUpVote(testCase.songVotingFor)

		if err != nil{
			t.Errorf("calculateAndRemoveUpVote() failed! Got an error: %v", err)
		}

		if resulteVoteAmount != testCase.expectedVoteAmount {
			t.Errorf("calculateAndRemoveUpVote() failed! Wanted vote amount: %v, got: %v", testCase.expectedVoteAmount, resulteVoteAmount)

		}

		if isResultBonusVote != false {
			t.Errorf("calculateAndRemoveUpVote() failed! Wanted bonus vote: %v, got: %v", false, isResultBonusVote)

		}

		_, upVoteExists := voter.songsUpVoted[testCase.songVotingFor]
		if upVoteExists {
			t.Errorf("calculateAndRemoveUpVote() failed! Song in map of upvoted songs")
		}
	}
}

var calculateAndRemoveDownVoteTests = []struct {
	voterType string
	songsDownVoted map[string]struct{}
	songVotingFor string
	expectedVoteAmount int
}{
	// Test voter types for regular vote
	{constants.RegularVoterType, map[string]struct{}{"song1": emptyStructValue}, "song1", 1},
	{constants.PrivilegedVoterType, map[string]struct{}{"song1": emptyStructValue}, "song1", 1},
	{constants.AdminVoterType, map[string]struct{}{"song1": emptyStructValue},"song1", 1},
}
func TestCalculateAndRemoveDownVote(t *testing.T){
	for _, testCase := range(calculateAndRemoveDownVoteTests) {
		voter, _ := NewVoter("ID1", testCase.voterType, 0)
		voter.songsDownVoted = testCase.songsDownVoted

		resulteVoteAmount, isResultBonusVote, err := voter.calculateAndRemoveDownVote(testCase.songVotingFor)

		if err != nil{
			t.Errorf("calculateAndRemoveDownVote() failed! Got an error: %v", err)
		}

		if resulteVoteAmount != testCase.expectedVoteAmount {
			t.Errorf("calculateAndRemoveDownVote() failed! Wanted vote amount: %v, got: %v", testCase.expectedVoteAmount, resulteVoteAmount)

		}

		if isResultBonusVote != false {
			t.Errorf("calculateAndRemoveDownVote() failed! Wanted bonus vote: %v, got: %v", false, isResultBonusVote)

		}

		_, upVoteExists := voter.songsDownVoted[testCase.songVotingFor]
		if upVoteExists {
			t.Errorf("calculateAndRemoveDownVote() failed! Song in map of downvoted songs")
		}
	}
}

var GetVoteAmountFromTypeTests = []struct {
	voterType string
	expectedVoteAmount int
}{
	{constants.RegularVoterType, 1},
	{constants.PrivilegedVoterType, 2},
	{constants.AdminVoterType, 1},
	{"not a voter type", 1},
}
func TestGetVoteAmountFromType(t *testing.T){
	for _, testCase := range(GetVoteAmountFromTypeTests) {
		voteAmount := getVoteAmountFromType(testCase.voterType)

		if voteAmount != testCase.expectedVoteAmount{
			t.Errorf("getVoteAmountFromType() failed! Wanted vote amount: %v, got %v", testCase.expectedVoteAmount, voteAmount)
		}
	}
}

var GetVoterDurationTests = []struct {
	voterType string
	expectedVoteDuration time.Duration
}{
	{constants.RegularVoterType, regularVoterDurationInMinutes},
	{constants.PrivilegedVoterType, priviledgedVoterDurationInMinutes},
	{constants.AdminVoterType, regularVoterDurationInMinutes},
	{"not a voter type", regularVoterDurationInMinutes},
}
func TestGetVoterDuration(t *testing.T){
	for _, testCase := range(GetVoterDurationTests) {
		voteDuration := getVoterDuration(testCase.voterType)

		if voteDuration != testCase.expectedVoteDuration{
			t.Errorf("getVoterDuration() failed! Wanted vote duration: %v, got %v", testCase.expectedVoteDuration, voteDuration)
		}
	}
}

var ContainsTests = []struct {
	testSlice []string
	searchValue string
	expectedExists bool
}{
	{[]string{"test string"}, "test string", true},
	{[]string{"test string"}, "string doesn't exist", false},
}
func TestContains(t *testing.T){
	for _, testCase := range(ContainsTests) {
		exists := contains(testCase.testSlice, testCase.searchValue)

		if exists != testCase.expectedExists{
			t.Errorf("contains() failed! Wanted contains value: %v, got %v", testCase.expectedExists, exists)
		}
	}
}
