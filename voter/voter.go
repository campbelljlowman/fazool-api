package voter

import (
	"time"
)

type Voter struct {
	Id string
	VoterType string
	Expires time.Time
	SongsVotedFor []string
	BonusVotes int
}

func NewVoter(id string, bonusVotes int) Voter {
	v := Voter{
		Id: id,
		VoterType: "regular-voter",
		Expires: time.Now().Add(5 * time.Minute),
		SongsVotedFor: nil,
		BonusVotes: bonusVotes,
	}
	return v
}

func (v *Voter) Refresh() {
	v.Expires = time.Now().Add(5 * time.Minute)
}