package payments

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/constants"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
	"golang.org/x/exp/slog"
)

type StripeService struct {
    accountService account.AccountService
    endpointSecret string
}

func NewStripeService(accountService account.AccountService) *StripeService {
    stripe.Key = os.Getenv("STRIPE_KEY")
    endpointSecret := os.Getenv("STRIPE_WEBHOOK_ENDPOINT_SECRET")
    
    stripeService := &StripeService{
        accountService: accountService,
        endpointSecret: endpointSecret,
    }
    return stripeService
}

func (s *StripeService) HandleStripeWebhook(w http.ResponseWriter, req *http.Request) {

    const MaxBodyBytes = int64(65536)
    req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
    payload, err := io.ReadAll(req.Body)
    if err != nil {
        slog.Warn("Error reading request body from stripe webhook", "error", err)
        w.WriteHeader(http.StatusServiceUnavailable)
    return
    }

    event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"),
    s.endpointSecret)

    if err != nil {
        slog.Warn("Error verifying stripe webhook signature", "error", err)
        w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
        return
    }

    switch event.Type {
    case "checkout.session.completed":
        var checkoutSessionData stripe.CheckoutSession
        json.Unmarshal(event.Data.Raw, &checkoutSessionData)
        if checkoutSessionData.PaymentLink != nil {
            slog.Info("Fazool tokens to add ", "data", constants.FazoolTokenPLinkMapping[checkoutSessionData.PaymentLink.ID])
            // s.accountService.AddFazoolTokens()
        }
        slog.Info("client reference ID: ", "data", checkoutSessionData.ClientReferenceID)
    default:
        slog.Warn("Unhandled stripe event type", "event-type", event.Type)
    }

    w.WriteHeader(http.StatusOK)
}