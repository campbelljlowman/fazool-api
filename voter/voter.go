package voter

import (
	"time"

	"github.com/campbelljlowman/fazool-api/graph/model"
)

type Voter struct {
	Id string
	VoterType string
	Expires time.Time
	SongsVotedFor []string
	BonusVotes int
}

// TODO: Making this long for testing, should be 5
const regularVoterDuration time.Duration = 15

func NewVoter(id string, bonusVotes int) Voter {
	v := Voter{
		Id: id,
		VoterType: "regular-voter",
		Expires: time.Now().Add(regularVoterDuration * time.Minute),
		SongsVotedFor: nil,
		BonusVotes: bonusVotes,
	}
	return v
}

func (v *Voter) GetVoterInfo() *model.VoterInfo {
	voter := model.VoterInfo{
		Type: v.VoterType,
		SongsVotedFor: v.SongsVotedFor,
		BonusVotes: &v.BonusVotes,
	}

	return &voter
}

func (v *Voter) Refresh() {
	v.Expires = time.Now().Add(5 * time.Minute)
}