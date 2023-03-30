package graph

import (
	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/session"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	sessionService		    session.SessionService
	accountService  		account.AccountService
}

func NewResolver(sessionService session.SessionService, accountService account.AccountService) *Resolver {
	return &Resolver{
		sessionService: sessionService,
		accountService: accountService,
	}
}
