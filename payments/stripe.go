package payments

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/campbelljlowman/fazool-api/account"
	"github.com/campbelljlowman/fazool-api/graph/model"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"github.com/stripe/stripe-go/v74/webhook"
	"golang.org/x/exp/slog"
)

type StripeService interface {
    HandleStripeWebhook(w http.ResponseWriter, req *http.Request)
    CreateCheckoutSession(sessionID, accountID int, fazoolTokensAmount model.FazoolTokenAmount) (string, error)
}

type stripeWrapper struct {
    accountService      account.AccountService
    endpointSecret      string
    frontendDomain      string
    fazoolTokenProducts [] *fazoolTokenProduct
}

type fazoolTokenProduct struct {
    fazoolTokenAmount       model.FazoolTokenAmount
    stripeProductID         string
    stripePriceID           string
    numberOfFazoolTokens    int
    priceUSD                int
}

func NewStripeService(accountService account.AccountService) StripeService {
    stripe.Key = os.Getenv("STRIPE_KEY")
    endpointSecret := os.Getenv("STRIPE_WEBHOOK_ENDPOINT_SECRET")
    frontendDomain := os.Getenv("FRONTEND_DOMAIN")
    stripeTestProductMode := os.Getenv("STRIPE_TEST_PRODUCT_MODE")
    
    if (endpointSecret == "") || (frontendDomain == "") {
		slog.Warn("At least one environment variable needed for stripe service is empty", 
        "endpointSecret", endpointSecret, "frontendDomain", frontendDomain)
		os.Exit(1)
    }

    // THESE SHOULD BE EXACTLY THE SAME AS THE FRONTEND!!!
    // https://github.com/campbelljlowman/fazool-ui/blob/master/src/constants.ts
    var fazoolTokenProducts [] *fazoolTokenProduct

    if (stripeTestProductMode == "true") {
        fazoolTokenProducts = append(fazoolTokenProducts, &fazoolTokenProduct{
            fazoolTokenAmount:      model.FazoolTokenAmountFive,
            stripeProductID:        "prod_OPijTEhckyCHAm",
            stripePriceID:          "price_1NctI5FrScZw72Ta1VoRFfSE",
            numberOfFazoolTokens:   5,
            priceUSD:               5,
        })
        fazoolTokenProducts = append(fazoolTokenProducts, &fazoolTokenProduct{
            fazoolTokenAmount:      model.FazoolTokenAmountTen,
            stripeProductID:        "prod_OPlsrZSSbZ6uO2",
            stripePriceID:          "price_1NcwKwFrScZw72TaXQBa7YRu",
            numberOfFazoolTokens:   10,
            priceUSD:               10,
        })
        fazoolTokenProducts = append(fazoolTokenProducts, &fazoolTokenProduct{
            fazoolTokenAmount:      model.FazoolTokenAmountTwentyTwo,
            stripeProductID:        "prod_OPltQrkYshQQIg",
            stripePriceID:          "price_1NcwLJFrScZw72TaFQ4sOj7f",
            numberOfFazoolTokens:   22,
            priceUSD:               20,
        })
    } else {
        fazoolTokenProducts = append(fazoolTokenProducts, &fazoolTokenProduct{
            fazoolTokenAmount:      model.FazoolTokenAmountFive,
            stripeProductID:        "prod_OQ54LRbqQ276Kh",
            stripePriceID:          "price_1NdEuoFrScZw72TaLjujszXB",
            numberOfFazoolTokens:   5,
            priceUSD:               5,
        })
        fazoolTokenProducts = append(fazoolTokenProducts, &fazoolTokenProduct{
            fazoolTokenAmount:      model.FazoolTokenAmountTen,
            stripeProductID:        "prod_OQ55OI2dpEa9D0",
            stripePriceID:          "price_1NdEutFrScZw72TajoNE73Ys",
            numberOfFazoolTokens:   10,
            priceUSD:               10,
        })
        fazoolTokenProducts = append(fazoolTokenProducts, &fazoolTokenProduct{
            fazoolTokenAmount:      model.FazoolTokenAmountTwentyTwo,
            stripeProductID:        "prod_OQ55wh7ZC3EpWa",
            stripePriceID:          "price_1NdEuwFrScZw72TayueTtJnu",
            numberOfFazoolTokens:   22,
            priceUSD:               20,
        })
    }

    stripeService := &stripeWrapper{
        accountService:         accountService,
        endpointSecret:         endpointSecret,
        frontendDomain:         frontendDomain,
        fazoolTokenProducts:    fazoolTokenProducts,
    }
    return stripeService
}

func (s *stripeWrapper) HandleStripeWebhook(w http.ResponseWriter, req *http.Request) {

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
            fazoolTokenProduct := s.getFazoolTokenProductFromProductID(lineItem.Price.Product.ID)
            s.accountService.AddFazoolTokens(accountID, fazoolTokenProduct.numberOfFazoolTokens)
        }
    default:
        slog.Warn("Unhandled stripe event type", "event-type", event.Type)
    }

    w.WriteHeader(http.StatusOK)
}

func (s *stripeWrapper) CreateCheckoutSession(sessionID, accountID int, fazoolTokensAmount model.FazoolTokenAmount) (string, error) {
    accountIDString := fmt.Sprintf("%v", accountID)
    redirectURL := fmt.Sprintf("%v/session/%v", s.frontendDomain, sessionID)
    fazoolTokenProduct := s.getFazoolTokenProductFromAmount(fazoolTokensAmount)

    params := &stripe.CheckoutSessionParams{
      LineItems: []*stripe.CheckoutSessionLineItemParams{
        &stripe.CheckoutSessionLineItemParams{
          Price: stripe.String(fazoolTokenProduct.stripePriceID),
          Quantity: stripe.Int64(1),
        },
      },
      Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
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

func (s *stripeWrapper) getFazoolTokenProductFromAmount(fazoolTokenAmount model.FazoolTokenAmount) *fazoolTokenProduct {
    for _, fazoolTokenProduct := range(s.fazoolTokenProducts) {
        if fazoolTokenProduct.fazoolTokenAmount == fazoolTokenAmount {
            return fazoolTokenProduct
        }
    }
    return nil
}

func (s *stripeWrapper) getFazoolTokenProductFromProductID(stripeProductID string) *fazoolTokenProduct {
    for _, fazoolTokenProduct := range(s.fazoolTokenProducts) {
        if fazoolTokenProduct.stripeProductID == stripeProductID {
            return fazoolTokenProduct
        }
    }
    return nil
}