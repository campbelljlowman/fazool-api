package voter

import (
	"testing"

	"github.com/campbelljlowman/fazool-api/graph/model"
)

var GetVoterInfoTests = []struct {
	inputVoter Voter
	expectedOutputVoter model.VoterInfo
}{
	{Voter{}},
}
func TestGetVoterInfo(t *testing.T){
	for _, testCase := range(GetVoterInfoTests) {
		resultVoter := testCase.inputVoter.GetVoterInfo()

		if resultVoter != &testCase.expectedOutputVoter {
			t.Errorf("GetVoterInfo() failed! Wanted: %v, got: %v", testCase.expectedOutputVoter, resultVoter)

		}

	}
}

var GetVoteAmountAndTypeTests = []struct {
	inputVoter Voter
	inputSong string
	inputVoteDirection model.SongVoteDirection
	inputVoteAction model.SongVoteAction
	expectedOutputVote int
	expectedOutputBonusVote bool

}{
	{Voter{}},
}
func TestGetVoteAmountAndType(t *testing.T){
	for _, testCase := range(GetVoteAmountAndTypeTests) {
		resulteVote, resultBonusVote, err := testCase.inputVoter.GetVoteAmountAndType(testCase.inputSong, &testCase.inputVoteDirection, &testCase.inputVoteAction)

		if resultVoter != &testCase.expectedOutputVoter {
			t.Errorf("GetVoterInfo() failed! Wanted: %v, got: %v", testCase.expectedOutputVoter, resultVoter)

		}

	}
}