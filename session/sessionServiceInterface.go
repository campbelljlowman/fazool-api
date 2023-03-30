package session

import (
	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/voter"
	"github.com/campbelljlowman/fazool-api/musicplayer"
	"github.com/campbelljlowman/fazool-api/graph/model"

)

type SessionService interface {
	CreateSession(adminAccountID int, accountLevel string, musicPlayer musicplayer.MusicPlayer) (int, error)

	GetSessionInfo(sessionID int) *model.SessionInfo
	GetSessionAdminAccountID(sessionID int) int
	GetVoterInSession(sessionID int, voterID string) (*voter.Voter, bool)
	// TODO: Probably shouldn't do this
	GetMusicPlayer(sessionID int) musicplayer.MusicPlayer
	IsSessionFull(sessionID int) bool
	DoesSessionExist(sessionID int) bool

	UpsertQueue(sessionID, vote int, song model.SongUpdate)
	UpsertVoterInSession(sessionID int, newVoter *voter.Voter)
	AdvanceQueue(sessionID int, force bool, accountService account.AccountService) error
	AddBonusVote(songID string, accountID, numberOfVotes, sessionID int)
	SetQueue(sessionID int, newQueue [] *model.QueuedSong)

	WatchSpotifyCurrentlyPlaying(sessionID int, accountService account.AccountService)
	WatchVotersExpirations(sessionID int)

	EndSession(sessionID int, accountService account.AccountService)

}