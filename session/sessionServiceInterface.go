package session

import (
	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/musicplayer"
	"github.com/campbelljlowman/fazool-api/voter"
	"github.com/campbelljlowman/fazool-api/graph/model"

)

type SessionService interface {
	CreateSession(adminAccountID int, accountLevel string) (int, error)
	EndSession(sessionID int)
	CheckVotersExpirations(sessionID int)
	IsSessionFull(sessionID int) bool
	AdvanceQueue(sessionID int, force bool, musicPlayer musicplayer.MusicPlayer, accountService account.AccountService) error
	UpsertQueue(sessionID int, vote int, song model.SongUpdate)
	CheckSpotifyCurrentlyPlaying(sessionID int, musicPlayer musicplayer.MusicPlayer, accountService account.AccountService)
	GetSessionInfo(sessionID int) *model.SessionInfo
	AddBonusVote(songID string, accountID, numberOfVotes, sessionID int)
	GetSessionAdminAccountID(sessionID int) int
	SetQueue(sessionID int, newQueue [] *model.QueuedSong)
	UpsertVoterInSession(sessionID int, newVoter *voter.Voter)
	GetVoterInSession(sessionID int, voterID string) (*voter.Voter, bool)
	DoesSessionExist(sessionID int) bool
}