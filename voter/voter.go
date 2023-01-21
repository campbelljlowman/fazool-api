package voter

import (
	"errors"
	"fmt"
	"time"

	"github.com/campbelljlowman/fazool-api/constants"
	"github.com/campbelljlowman/fazool-api/graph/model"
)

type Voter struct {
	ID string
	VoterType string
	ExpiresAt time.Time
	SongsUpVoted map[string]struct{}
	SongsDownVoted map[string]struct{}
	BonusVotes int
}

var empty struct{}
const regularVoterDurationInMinutes time.Duration = 15
const priviledgedVoterDurationInMinutes time.Duration = 15
var validVoterTypes = []string{constants.AdminVoterType, constants.PrivilegedVoterType, constants.RegularVoterType}


func NewVoter(ID, voterType string, bonusVotes int) (*Voter, error) {
	if !contains(validVoterTypes, voterType) {
		return nil, fmt.Errorf("Invalid voter type passed!")
	}

	v := Voter{
		ID: ID,
		VoterType: voterType,
		ExpiresAt: time.Now().Add(getVoterDuration(voterType) * time.Minute),
		SongsUpVoted: make(map[string]struct{}),
		SongsDownVoted: make(map[string]struct{}),
		BonusVotes: bonusVotes,
	}
	return &v, nil
}

func (v *Voter) GetVoterInfo() *model.VoterInfo {
	songsUpVotedList := make([]string, len(v.SongsUpVoted))
	songsDownVotedList := make([]string, len(v.SongsDownVoted))

	i := 0
	for k := range v.SongsUpVoted {
		songsUpVotedList[i] = k
		i++
	}

	i = 0
	for k := range v.SongsDownVoted {
		songsDownVotedList[i] = k
		i++
	}

	voter := model.VoterInfo{
		Type: v.VoterType,
		SongsUpVoted: songsUpVotedList,
		SongsDownVoted: songsDownVotedList,
		BonusVotes: &v.BonusVotes,
	}

	return &voter
}

func (v *Voter) GetVoteAmountAndType(song string, direction *model.SongVoteDirection, action *model.SongVoteAction) (int, bool, error) {
	switch {
	case action.String() == "ADD" && direction.String() == "UP":
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
				delete(v.SongsDownVoted, song)
				// If song was downvoted and is being upvoted, vote needs to be double
				voteAdjustment = 1
			}
		}

		v.SongsUpVoted[song] = empty
		return voteAdjustment + getVoteValue(v.VoterType), false, nil
	case action.String() == "ADD" && direction.String() == "DOWN":
		voteAdjustment := 0
		if v.VoterType != constants.AdminVoterType {
			if _, exists := v.SongsDownVoted[song]; exists {
				return 0, false, errors.New("You've already voted for this song!")
			}

			if _, exists := v.SongsUpVoted[song]; exists {
				delete(v.SongsUpVoted, song)
				// If song was downvoted and is being upvoted, vote needs to be double
				voteAdjustment = getVoteValue(v.VoterType)
			}
		}

		v.SongsDownVoted[song] = empty
		return -(voteAdjustment + 1), false, nil
	case action.String() == "REMOVE" && direction.String() == "UP":
		delete(v.SongsUpVoted, song)
		return -getVoteValue(v.VoterType), false, nil
	case action.String() == "REMOVE" && direction.String() == "DOWN":
		delete(v.SongsDownVoted, song)
		return 1, false, nil
	}

	return 0, false, fmt.Errorf("Song vote inputs aren't valid!")
}

func (v *Voter) RefreshVoterExpiration() {
	v.ExpiresAt = time.Now().Add(getVoterDuration(v.VoterType) * time.Minute)
}

func getVoteValue (voterType string) int {
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