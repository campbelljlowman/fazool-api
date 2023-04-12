package voter

import (
	"testing"
	"time"

	"github.com/campbelljlowman/fazool-api/graph/model"
)

var two = 2
var GetVoterInfoTests = []struct {
	inputVoter Voter
	expectedOutputVoter model.VoterInfo
}{
	{Voter{
		VoterID: "asdf",
		AccountID: 1234,
		VoterType: model.VoterTypeFree,
		ExpiresAt: time.Now(),
		SongsUpVoted: map[string]struct{}{"song1": emptyStructValue},
		SongsDownVoted:  make(map[string]struct{}),
		BonusVotes: 2,
	}, model.VoterInfo{
		Type: model.VoterTypeFree,
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
	voterType model.VoterType
	SongsUpVoted map[string]struct{}
	SongsDownVoted map[string]struct{}
	bonusVotes int
	songVotingFor string
	expectedVoteAmount int
	expectedBonusVoteValue bool
	expectedError bool
}{
	// Test voter types for regular vote
	{model.VoterTypeFree, map[string]struct{}{}, map[string]struct{}{},  0, "song1", 1, false, false},
	{model.VoterTypePrivileged, map[string]struct{}{}, map[string]struct{}{}, 0, "song1", 2, false, false},
	{model.VoterTypeAdmin, map[string]struct{}{}, map[string]struct{}{}, 0, "song1", 1, false, false},
	// Test if song already exists 
	{model.VoterTypeFree, map[string]struct{}{"song1": emptyStructValue}, map[string]struct{}{}, 0, "song1", 0, false, true},
	{model.VoterTypeFree, map[string]struct{}{"song1": emptyStructValue}, map[string]struct{}{}, 1, "song1", 1, true, false},
	{model.VoterTypeAdmin, map[string]struct{}{"song1": emptyStructValue}, map[string]struct{}{}, 0, "song1", 1, false, false},
	// Test vote adjustment
	{model.VoterTypeFree, map[string]struct{}{}, map[string]struct{}{"song1": emptyStructValue}, 0, "song1", 2, false, false},
	{model.VoterTypePrivileged, map[string]struct{}{}, map[string]struct{}{"song1": emptyStructValue}, 0, "song1", 3, false, false},
	{model.VoterTypeAdmin, map[string]struct{}{}, map[string]struct{}{"song1": emptyStructValue}, 0, "song1", 1, false, false},
}
func TestCalculateAndAddUpVote(t *testing.T){
	for _, testCase := range(calculateAndAddUpVoteTests) {
		voter, _ := NewVoter("voterID", testCase.voterType, 1234, testCase.bonusVotes)
		voter.SongsUpVoted = testCase.SongsUpVoted
		voter.SongsDownVoted = testCase.SongsDownVoted

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

		_, upVoteExists := voter.SongsUpVoted[testCase.songVotingFor]
		if !upVoteExists {
			t.Errorf("calculateAndAddUpVote() failed! Song not in map of upvoted songs")
		}

		_, downVoteExists := voter.SongsDownVoted[testCase.songVotingFor]
		if downVoteExists {
			t.Errorf("calculateAndAddUpVote() failed! Song in map of downvoted songs")
		}
	}
}

var calculateAndAddDownVoteTests = []struct {
	voterType model.VoterType
	SongsUpVoted map[string]struct{}
	SongsDownVoted map[string]struct{}
	songVotingFor string
	expectedVoteAmount int
	expectedError bool
}{
	// Test voter types for regular vote
	{model.VoterTypeFree, map[string]struct{}{}, map[string]struct{}{}, "song1", 0, false},
	{model.VoterTypePrivileged, map[string]struct{}{}, map[string]struct{}{}, "song1", -1, false},
	{model.VoterTypeAdmin, map[string]struct{}{}, map[string]struct{}{}, "song1", -1, false},
	// Test if song already exists 
	{model.VoterTypeFree, map[string]struct{}{}, map[string]struct{}{"song1": emptyStructValue}, "song1", 0, true},
	{model.VoterTypePrivileged, map[string]struct{}{}, map[string]struct{}{"song1": emptyStructValue}, "song1", 0, true},
	{model.VoterTypeAdmin, map[string]struct{}{}, map[string]struct{}{"song1": emptyStructValue}, "song1", -1, false},
	// Test vote adjustment
	{model.VoterTypeFree, map[string]struct{}{"song1": emptyStructValue}, map[string]struct{}{}, "song1", 0, true},
	{model.VoterTypePrivileged, map[string]struct{}{"song1": emptyStructValue}, map[string]struct{}{}, "song1", -3, true},
	{model.VoterTypeAdmin, map[string]struct{}{"song1": emptyStructValue}, map[string]struct{}{}, "song1", -1, false},
}

func TestCalculateAndAddDownVote(t *testing.T){
	for _, testCase := range(calculateAndAddDownVoteTests) {
		voter, _ := NewVoter("voterID", testCase.voterType, 1234, 0)
		voter.SongsUpVoted = testCase.SongsUpVoted
		voter.SongsDownVoted = testCase.SongsDownVoted

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

		if voter.VoterType != model.VoterTypeFree {
			_, downVoteExists := voter.SongsDownVoted[testCase.songVotingFor]
			if !downVoteExists {
				t.Errorf("calculateAndAddUpVote() failed! Song not in map of downvoted songs")
			}
		}

		if voter.VoterType != model.VoterTypeFree {
			_, upVoteExists := voter.SongsUpVoted[testCase.songVotingFor]
			if upVoteExists {
				t.Errorf("calculateAndAddUpVote() failed! Song in map of upvoted songs")
			}
		}

	}
}


var calculateAndRemoveUpVoteTests = []struct {
	voterType model.VoterType
	SongsUpVoted map[string]struct{}
	songVotingFor string
	expectedVoteAmount int
}{
	// Test voter types for regular vote
	{model.VoterTypeFree, map[string]struct{}{"song1": emptyStructValue}, "song1", -1},
	{model.VoterTypePrivileged, map[string]struct{}{"song1": emptyStructValue}, "song1", -2},
	{model.VoterTypeAdmin, map[string]struct{}{"song1": emptyStructValue},"song1", -1},
}
func TestCalculateAndRemoveUpVote(t *testing.T){
	for _, testCase := range(calculateAndRemoveUpVoteTests) {
		voter, _ := NewVoter("voterID", testCase.voterType, 1234, 0)
		voter.SongsUpVoted = testCase.SongsUpVoted

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

		_, upVoteExists := voter.SongsUpVoted[testCase.songVotingFor]
		if upVoteExists {
			t.Errorf("calculateAndRemoveUpVote() failed! Song in map of upvoted songs")
		}
	}
}

var calculateAndRemoveDownVoteTests = []struct {
	voterType model.VoterType
	SongsDownVoted map[string]struct{}
	songVotingFor string
	expectedVoteAmount int
}{
	// Test voter types for regular vote
	{model.VoterTypeFree, map[string]struct{}{"song1": emptyStructValue}, "song1", 1},
	{model.VoterTypePrivileged, map[string]struct{}{"song1": emptyStructValue}, "song1", 1},
	{model.VoterTypeAdmin, map[string]struct{}{"song1": emptyStructValue},"song1", 1},
}
func TestCalculateAndRemoveDownVote(t *testing.T){
	for _, testCase := range(calculateAndRemoveDownVoteTests) {
		voter, _ := NewVoter("voterID", testCase.voterType, 1234, 0)
		voter.SongsDownVoted = testCase.SongsDownVoted

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

		_, upVoteExists := voter.SongsDownVoted[testCase.songVotingFor]
		if upVoteExists {
			t.Errorf("calculateAndRemoveDownVote() failed! Song in map of downvoted songs")
		}
	}
}

var GetVoteAmountFromTypeTests = []struct {
	voterType model.VoterType
	expectedVoteAmount int
}{
	{model.VoterTypeFree, 1},
	{model.VoterTypePrivileged, 2},
	{model.VoterTypeAdmin, 1},
	{"not a voter type", 1},
}
func TestGetVoteAmountFromType(t *testing.T){
	for _, testCase := range(GetVoteAmountFromTypeTests) {
		voteAmount := getNumberOfVotesFromType(testCase.voterType)

		if voteAmount != testCase.expectedVoteAmount{
			t.Errorf("getNumberOfVotesFromType() failed! Wanted vote amount: %v, got %v", testCase.expectedVoteAmount, voteAmount)
		}
	}
}

var GetVoterDurationTests = []struct {
	voterType model.VoterType
	expectedVoteDuration time.Duration
}{
	{model.VoterTypeFree, regularVoterDurationInMinutes},
	{model.VoterTypePrivileged, priviledgedVoterDurationInMinutes},
	{model.VoterTypeAdmin, regularVoterDurationInMinutes},
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
	testSlice []model.VoterType
	searchValue model.VoterType
	expectedExists bool
}{
	{[]model.VoterType{model.VoterTypeAdmin}, model.VoterTypeAdmin, true},
	{[]model.VoterType{model.VoterTypeAdmin}, model.VoterTypeFree, false},
}
func TestContains(t *testing.T){
	for _, testCase := range(ContainsTests) {
		exists := contains(testCase.testSlice, testCase.searchValue)

		if exists != testCase.expectedExists{
			t.Errorf("contains() failed! Wanted contains value: %v, got %v", testCase.expectedExists, exists)
		}
	}
}
