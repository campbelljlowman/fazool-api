package voter

import (
	"time"
)

type Voter struct {
	id string
	voterType string
	expires time.Time
	songsVotedFor []string
	bonusVotes int
}

func NewVoter(id string, bonusVotes int) Voter {
	v := Voter{
		id: id,
		voterType: "regular-voter",
		expires: time.Now().Add(5 * time.Minute),
		songsVotedFor: nil,
		bonusVotes: bonusVotes,
	}
	return v
}

func (v *Voter) refresh() {
	v.expires = time.Now().Add(5 * time.Minute)
}