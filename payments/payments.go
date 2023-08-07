package payments

import (
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
    slog.Info("Stripe webhook called")
    stripe.Key = os.Getenv("STRIPE_KEY")

    const MaxBodyBytes = int64(65536)
    req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
    payload, err := io.ReadAll(req.Body)
    if err != nil {
        slog.Warn("Error reading request body from stripe webhook", "error", err)
        w.WriteHeader(http.StatusServiceUnavailable)
    return
    }

    // This is your Stripe CLI webhook secret for testing your endpoint locally.
    endpointSecret := os.Getenv("STRIPE_WEBHOOK_ENDPOINT_SECRET")
    // test
    // endpointSecret := "whsec_2WIqSly5FZWIBsz45Dxyr9ukDctoMKUj"
    // Pass the request body and Stripe-Signature header to ConstructEvent, along
    // with the webhook signing key.
    event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"),
    endpointSecret)

    if err != nil {
        slog.Warn("Error verifying stripe webhook signature", "error", err)
        w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
        return
    }

    // Unmarshal the event data into an appropriate struct depending on its Type
    switch event.Type {
    case "payment_intent.succeeded":
        // Then define and call a function to handle the event payment_intent.succeeded
        // ... handle other event types
        slog.Info("Received stripe webhook call", "data", event.Data.Object)
    default:
        slog.Warn("Unhandled stripe event type", "event-type", event.Type)
    }

    w.WriteHeader(http.StatusOK)
}