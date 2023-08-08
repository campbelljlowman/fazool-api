package payments

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/constants"
	"github.com/campbelljlowman/fazool-api/graph/model"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"github.com/stripe/stripe-go/v74/webhook"
	"golang.org/x/exp/slog"
)

type StripeService struct {
    accountService account.AccountService
    endpointSecret string
    frontendDomain string
    fazoolTokensOptionsMapping map[model.FazoolTokenAmount]string
}

func NewStripeService(accountService account.AccountService) *StripeService {
    stripe.Key = os.Getenv("STRIPE_KEY")
    endpointSecret := os.Getenv("STRIPE_WEBHOOK_ENDPOINT_SECRET")
    frontendDomain := os.Getenv("FRONTEND_DOMAIN")
    fazoolTokensFivePrice := os.Getenv("FAZOOL_TOKENS_FIVE_PRICE")
    
    if (endpointSecret == "") || (frontendDomain == "") || (fazoolTokensFivePrice == "") {
		slog.Warn("At least one environment variable needed for stripe service is empty", 
        "endpointSecret", endpointSecret, "frontendDomain", frontendDomain, "fazoolTokensFivePrice", fazoolTokensFivePrice)
		os.Exit(1)
    }

    stripeService := &StripeService{
        accountService: accountService,
        endpointSecret: endpointSecret,
        frontendDomain: frontendDomain,
        fazoolTokensOptionsMapping: map[model.FazoolTokenAmount]string{
            model.FazoolTokenAmountFive: fazoolTokensFivePrice,
        },
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
        var checkoutSessionResponse stripe.CheckoutSession
        json.Unmarshal(event.Data.Raw, &checkoutSessionResponse)
        params := &stripe.CheckoutSessionParams{};
        params.AddExpand("line_items")
        checkoutSessionData, err := session.Get(checkoutSessionResponse.ID, params);
        if err != nil {
            slog.Error("Error getting checkout session data", err)
            break
        }

        accountID, err := strconv.Atoi(checkoutSessionData.ClientReferenceID)
        if err != nil {
            slog.Error("Error getting account ID from checkout session data", err)
            break
        }

        for _, lineItem := range(checkoutSessionData.LineItems.Data) {
            fazoolTokensToAdd := constants.FazoolTokenProductMapping[lineItem.Price.Product.ID]
            slog.Info("fazool tokens to add:", "data", fazoolTokensToAdd)
            s.accountService.AddFazoolTokens(accountID, fazoolTokensToAdd)
        }

        // LOG info about payment if not from one of the links or no client reference id
    default:
        slog.Warn("Unhandled stripe event type", "event-type", event.Type)
    }

    w.WriteHeader(http.StatusOK)
}

func (s *StripeService) CreateCheckoutSession(sessionID, accountID int, fazoolTokensAmount model.FazoolTokenAmount) (string, error) {
    accountIDString := fmt.Sprintf("%v", accountID)
    redirectURL := fmt.Sprintf("%v/session/%v", s.frontendDomain, sessionID)
    priceID := s.fazoolTokensOptionsMapping[fazoolTokensAmount]

    params := &stripe.CheckoutSessionParams{
      LineItems: []*stripe.CheckoutSessionLineItemParams{
        &stripe.CheckoutSessionLineItemParams{
          // Provide the exact Price ID (for example, pr_1234) of the product you want to sell
          Price: stripe.String(priceID),
          Quantity: stripe.Int64(1),
        },
      },
      Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
    //   SuccessURL: stripe.String(s.frontendDomain + "/session/" + sessionID),
      SuccessURL: &redirectURL,
      CancelURL: &redirectURL,
      AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{Enabled: stripe.Bool(true)},
      ClientReferenceID: &accountIDString,
    }
  
    checkoutSession, err := session.New(params)
  
    if err != nil {
        return "", err
    }
  
    return checkoutSession.URL, nil
}