package graph

import (
	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/payments"
	"github.com/campbelljlowman/fazool-api/session"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	sessionService		    session.SessionService
	accountService  		account.AccountService
	stripeService			*payments.StripeService
}

func NewResolver(sessionService session.SessionService, accountService account.AccountService, stripeService *payments.StripeService) *Resolver {
	return &Resolver{
		sessionService: sessionService,
		accountService: accountService,
		stripeService: stripeService,
	}
}
