package voter

import (
	"errors"
	"fmt"
	"time"

	"github.com/campbelljlowman/fazool-api/graph/model"
	"github.com/campbelljlowman/fazool-api/constants"
)

type Voter struct {
	Id string
	VoterType string
	Expires time.Time
	SongsUpVoted map[string]struct{}
	SongsDownVoted map[string]struct{}
	BonusVotes int
}

var empty struct{}
const regularVoterDuration time.Duration = 15
const priviledgedVoterDuration time.Duration = 15
var validVoterTypes = []string{constants.AdminVoterType, constants.PrivilegedVoterType, constants.RegularVoterType}


func NewVoter(id, voterType string, bonusVotes int) (*Voter, error) {
	if !contains(validVoterTypes, voterType) {
		return nil, fmt.Errorf("Invalid voter type passed!")
	}

	v := Voter{
		Id: id,
		VoterType: voterType,
		Expires: time.Now().Add(getVoterDuration(voterType) * time.Minute),
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

func (v *Voter) ProcessVote(song string, direction *model.SongVoteDirection, action *model.SongVoteAction) (int, bool, error) {
	if direction.String() == "UP" {
		if action.String() == "ADD" {
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

		} else if action.String() == "REMOVE" {
			delete(v.SongsUpVoted, song)
			return -getVoteValue(v.VoterType), false, nil
		}
	} else if direction.String() == "DOWN"{
		if action.String() == "ADD" {
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

		} else if action.String() == "REMOVE" {
			delete(v.SongsDownVoted, song)
			return 1, false, nil
		}
	}

	return 0, false, fmt.Errorf("Song vote inputs aren't valid!")
}

func (v *Voter) Refresh() {
	v.Expires = time.Now().Add(getVoterDuration(v.VoterType) * time.Minute)
}

func getVoteValue (voterType string) int {
	if voterType == constants.PrivilegedVoterType {
		return 2
	}
	return 1
}

func getVoterDuration (voterType string) time.Duration {
	if voterType == constants.PrivilegedVoterType {
		return priviledgedVoterDuration
	}
	return regularVoterDuration
}

func contains(elems []string, v string) bool {
    for _, s := range elems {
        if v == s {
            return true
        }
    }
    return false
}