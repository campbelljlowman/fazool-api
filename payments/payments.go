package payments

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
	"golang.org/x/exp/slog"
)

// The library needs to be configured with your account's secret key.
// Ensure the key is kept out of any version control system you might be using.

// func main() {
//     http.HandleFunc("/webhook", handleWebhook)
//     addr := "localhost:4242"
//     log.Printf("Listening on %s", addr)
//     log.Fatal(http.ListenAndServe(addr, nil))
// }

func HandleStripeWebhook(w http.ResponseWriter, req *http.Request) {
    stripe.Key = os.Getenv("STRIPE_KEY")

    const MaxBodyBytes = int64(65536)
    req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
    payload, err := io.ReadAll(req.Body)
    if err != nil {
        slog.Warn("Error reading request body from stripe webhook", "error", err)
        w.WriteHeader(http.StatusServiceUnavailable)
    return
    }

    endpointSecret := os.Getenv("STRIPE_WEBHOOK_ENDPOINT_SECRET")
    event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"),
    endpointSecret)

    if err != nil {
        slog.Warn("Error verifying stripe webhook signature", "error", err)
        w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
        return
    }

    switch event.Type {
    case "checkout.session.completed":
        var checkoutSessionData stripe.CheckoutSession
        json.Unmarshal(event.Data.Raw, &checkoutSessionData)
        slog.Info("Payment link: ", "data", checkoutSessionData.PaymentLink)
    default:
        slog.Warn("Unhandled stripe event type", "event-type", event.Type)
    }

    w.WriteHeader(http.StatusOK)
}