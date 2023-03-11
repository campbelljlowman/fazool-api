package voter

import (
	"errors"
	"fmt"
	"time"

	"github.com/campbelljlowman/fazool-api/constants"
	"github.com/campbelljlowman/fazool-api/graph/model"
	"golang.org/x/exp/slog"
)

type Voter struct {
	VoterID string
	AccountID string
	VoterType string
	ExpiresAt time.Time
	SongsUpVoted map[string]struct{}
	SongsDownVoted map[string]struct{}
	BonusVotes int
}

var emptyStructValue struct{}
const regularVoterDurationInMinutes time.Duration = 15
const priviledgedVoterDurationInMinutes time.Duration = 15
var validVoterTypes = []string{constants.AdminVoterType, constants.PrivilegedVoterType, constants.RegularVoterType}


func NewVoter(voterID, accountID, voterType string, BonusVotes int) (*Voter, error) {
	if !contains(validVoterTypes, voterType) {
		return nil, fmt.Errorf("Invalid voter type passed!")
	}

	v := Voter{
		VoterID: voterID,
		AccountID: accountID,
		VoterType: voterType,
		ExpiresAt: time.Now().Add(getVoterDuration(voterType) * time.Minute),
		SongsUpVoted: make(map[string]struct{}),
		SongsDownVoted: make(map[string]struct{}),
		BonusVotes: BonusVotes,
	}
	slog.Info("Created voter in constructor", "voter", v, "bonus-votes", v.BonusVotes)
	return &v, nil
}

func (v *Voter) GetVoterInfo() *model.VoterInfo {
	SongsUpVotedList := make([]string, len(v.SongsUpVoted))
	SongsDownVotedList := make([]string, len(v.SongsDownVoted))

	i := 0
	for k := range v.SongsUpVoted {
		SongsUpVotedList[i] = k
		i++
	}

	i = 0
	for k := range v.SongsDownVoted {
		SongsDownVotedList[i] = k
		i++
	}

	voter := model.VoterInfo{
		Type: v.VoterType,
		SongsUpVoted: SongsUpVotedList,
		SongsDownVoted: SongsDownVotedList,
		BonusVotes: &v.BonusVotes,
	}

	return &voter
}

func (v *Voter) CalculateAndProcessVote(song string, direction *model.SongVoteDirection, action *model.SongVoteAction) (int, bool, error) {
	switch {
	case *action == model.SongVoteActionAdd && *direction == model.SongVoteDirectionUp:
		return v.calculateAndAddUpVote(song)
	case *action == model.SongVoteActionAdd && *direction == model.SongVoteDirectionDown:
		return v.calculateAndAddDownVote(song)
	case *action == model.SongVoteActionRemove && *direction == model.SongVoteDirectionUp:
		return v.calculateAndRemoveUpVote(song)
	case *action == model.SongVoteActionRemove && *direction == model.SongVoteDirectionDown:
		return v.calculateAndRemoveDownVote(song)
	}

	return 0, false, fmt.Errorf("Song vote inputs aren't valid!")
}

func (v *Voter) calculateAndAddUpVote(song string) (int, bool, error){
	voteAdjustment := 0
	if v.VoterType != constants.AdminVoterType {
		if _, exists := v.SongsUpVoted[song]; exists {
			if v.BonusVotes <= 0 {
				return 0, false, errors.New("You've already voted for this song!")
			} else {
				// Handle bonus votes
				v.BonusVotes -= 1
				return 1, true, nil
			}
		}

		if _, exists := v.SongsDownVoted[song]; exists {
			// If song was downvoted and is being upvoted, vote needs to be double
			voteAdjustment = 1
		}
	}
	
	delete(v.SongsDownVoted, song)
	v.SongsUpVoted[song] = emptyStructValue
	return voteAdjustment + getNumberOfVotesFromType(v.VoterType), false, nil
}

func (v *Voter) calculateAndAddDownVote(song string) (int, bool, error){
	voteAdjustment := 0
	if v.VoterType != constants.AdminVoterType {
		if _, exists := v.SongsDownVoted[song]; exists {
			return 0, false, errors.New("You've already voted for this song!")
		}

		if _, exists := v.SongsUpVoted[song]; exists {
			// If song was downvoted and is being upvoted, vote needs to be double
			voteAdjustment = getNumberOfVotesFromType(v.VoterType)
		}
	}
	
	delete(v.SongsUpVoted, song)
	v.SongsDownVoted[song] = emptyStructValue
	return -(voteAdjustment + 1), false, nil
}

func (v *Voter) calculateAndRemoveUpVote(song string) (int, bool, error) {
	delete(v.SongsUpVoted, song)
	return -getNumberOfVotesFromType(v.VoterType), false, nil
}

func (v *Voter) calculateAndRemoveDownVote(song string) (int, bool, error) {
	delete(v.SongsDownVoted, song)
	return 1, false, nil
}

func (v *Voter) RefreshVoterExpiration() {
	v.ExpiresAt = time.Now().Add(getVoterDuration(v.VoterType) * time.Minute)
}

func getNumberOfVotesFromType (voterType string) int {
	if voterType == constants.PrivilegedVoterType {
		return 2
	}
	return 1
}

func getVoterDuration (voterType string) time.Duration {
	if voterType == constants.PrivilegedVoterType {
		return priviledgedVoterDurationInMinutes
	}
	return regularVoterDurationInMinutes
}

func contains(elems []string, v string) bool {
    for _, s := range elems {
        if v == s {
            return true
        }
    }
    return false
}