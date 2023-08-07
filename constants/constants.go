package constants

import (
	"github.com/campbelljlowman/fazool-api/graph/model"
)

// THESE SHOULD BE EXACTLY THE SMAE VALUES AS IN THE FRONTEND!!!
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

var FazoolTokenPLinkMapping = map[string]int{
	"plink_1Nbn5WFrScZw72TaAINiKhPG": 5,
}