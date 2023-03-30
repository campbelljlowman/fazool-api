package graph

import (
	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/musicplayer"
	"github.com/campbelljlowman/fazool-api/session"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	musicPlayers			map[int]musicplayer.MusicPlayer
	sessionService		    session.SessionService
	accountService  		account.AccountService
}

func NewResolver(sessionService session.SessionService, accountService account.AccountService) *Resolver {
	return &Resolver{
		musicPlayers: 	make(map[int]musicplayer.MusicPlayer),
		sessionService: sessionService,
		accountService: accountService,
	}
}
