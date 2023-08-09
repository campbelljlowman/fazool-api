package constants

import (
	"github.com/campbelljlowman/fazool-api/graph/model"
)

// THESE SHOULD BE EXACTLY THE SAME VALUES AS IN THE FRONTEND!!!
// https://github.com/campbelljlowman/fazool-ui/blob/master/src/constants.ts
const (
	SuperVoterCost = 6
)

type BonusVoteCostPair struct {
	NumberOfBonusVotes 	int
	CostInFazoolTokens 	int
}

var BonusVoteCostMapping = map[model.BonusVoteAmount]BonusVoteCostPair{
	model.BonusVoteAmountTen: {
		NumberOfBonusVotes: 10,
		CostInFazoolTokens: 3,
	},
	model.BonusVoteAmountTwentyFive: {
		NumberOfBonusVotes: 25,
		CostInFazoolTokens: 6,
	},
	model.BonusVoteAmountFifty: {
		NumberOfBonusVotes: 50,
		CostInFazoolTokens: 10,
	},
}