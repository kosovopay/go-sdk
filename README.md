# KosovoPay Go SDK

Official Go SDK for the [KosovoPay](https://kosovo.sh) payment API.

## Requirements

- Go 1.23 or later

## Installation

```bash
go get github.com/kosovopay/go-sdk
```

## Quick start

```go
import kosovopay "github.com/kosovopay/go-sdk"

kp := kosovopay.New("sk_test_...")
```

## Usage

### Hosted checkout (customer picks the bank)

```go
payment, err := kp.Payments.Create(ctx, kosovopay.CreatePaymentParams{
    Amount:     4990, // €49.90
    Currency:   kosovopay.CurrencyEUR,
    Mode:       kosovopay.CheckoutModeHosted,
    SuccessURL: "https://shop.example.com/thanks",
    CancelURL:  "https://shop.example.com/cart",
    Metadata:   map[string]interface{}{"order_id": "O-1234"},
}, "")
if err != nil {
    log.Fatal(err)
}
// Redirect customer to payment.HostedURL
fmt.Println(payment.HostedURL)
```

### Direct checkout (you pick the bank)

```go
payment, err := kp.Payments.Create(ctx, kosovopay.CreatePaymentParams{
    Amount:     2500,
    Currency:   kosovopay.CurrencyEUR,
    Mode:       kosovopay.CheckoutModeDirect,
    BankCode:   kosovopay.BankCodeOnefor,
    SuccessURL: "https://shop.example.com/ok",
    CancelURL:  "https://shop.example.com/cart",
    FailURL:    "https://shop.example.com/failed",
}, "")
// Redirect customer to payment.RedirectURL
```

### Retrieve a payment

```go
payment, err := kp.Payments.Retrieve(ctx, "pi_...")
if payment.Status == kosovopay.PaymentStatusCaptured {
    // fulfil the order
}
```

### Refund a payment

```go
// Full refund
refund, err := kp.Refunds.Create(ctx, kosovopay.CreateRefundParams{
    Payment: "pi_...",
    Reason:  kosovopay.RefundReasonRequestedByCustomer,
}, "")

// Partial refund
refund, err = kp.Refunds.Create(ctx, kosovopay.CreateRefundParams{
    Payment: "pi_...",
    Amount:  1000, // €10.00
}, "")
```

### Cursor pagination

The `Iter` method streams results across pages automatically:

```go
iter := kp.Payments.Iter(ctx, &kosovopay.ListPaymentsParams{
    Status: kosovopay.PaymentStatusCaptured,
})
for iter.Next() {
    p := iter.Value()
    fmt.Println(p.ID, p.Amount)
}
if err := iter.Err(); err != nil {
    log.Fatal(err)
}
```

To collect all results at once:

```go
payments, err := kp.Payments.All(ctx, nil)
```

### Webhooks

Register an endpoint:

```go
ep, err := kp.WebhookEndpoints.Create(ctx, kosovopay.CreateWebhookEndpointParams{
    URL: "https://shop.example.com/webhooks/kosovopay",
    EnabledEvents: []kosovopay.WebhookEventType{
        kosovopay.WebhookEventTypePaymentCaptured,
        kosovopay.WebhookEventTypeRefundSucceeded,
    },
})
// Store ep.Secret securely — it is shown only once.
```

Verify and parse inbound events in your HTTP handler:

```go
func webhookHandler(w http.ResponseWriter, r *http.Request) {
    payload, _ := io.ReadAll(r.Body)
    sigHeader := r.Header.Get(kosovopay.SignatureHeader)

    event, err := kosovopay.ConstructEvent(payload, sigHeader, os.Getenv("KP_WHSEC"), 0)
    if err != nil {
        http.Error(w, "signature verification failed", http.StatusBadRequest)
        return
    }

    switch event.Type {
    case kosovopay.WebhookEventTypePaymentCaptured:
        p, _ := event.AsPayment()
        fulfil(p.Metadata["order_id"])
    case kosovopay.WebhookEventTypeRefundSucceeded:
        ref, _ := event.AsRefund()
        reverse(ref.ID)
    }
    w.WriteHeader(http.StatusOK)
}
```

### Error handling

Every error returned by the SDK is a typed struct. Switch on the concrete type
or use `errors.As` to access the error fields:

```go
payment, err := kp.Payments.Create(ctx, params, "")
if err != nil {
    var rle *kosovopay.RateLimitError
    var vle *kosovopay.ValidationError
    var pne *kosovopay.PaymentNotCancelableError

    switch {
    case errors.As(err, &rle):
        time.Sleep(time.Duration(rle.RetryAfter) * time.Second)
    case errors.As(err, &vle):
        log.Printf("validation error on param %s: %s", vle.Param, vle.Message)
    case errors.As(err, &pne):
        log.Println("payment cannot be cancelled in current state")
    default:
        log.Printf("API error: %v", err)
    }
}
```

Available error types:

| Type | HTTP | When |
|---|---|---|
| `*AuthenticationError` | 401 | Missing or invalid API key |
| `*PermissionError` | 403 | Key lacks permission |
| `*ValidationError` | 422/404 | Malformed request |
| `*IdempotencyError` | 409 | Idempotency key conflict |
| `*RateLimitError` | 429 | Rate limited; check `RetryAfter` |
| `*PaymentError` | 422 | Generic payment error |
| `*AmountBelowMinimumError` | 422 | Amount below bank minimum |
| `*AmountStepInvalidError` | 422 | Amount not a valid step |
| `*BankNotEnabledError` | 422 | Bank not enabled for team |
| `*BankUnreachableError` | 502 | Bank unreachable |
| `*PaymentNotCancelableError` | 422 | Payment not cancelable |
| `*PaymentNotRefundableError` | 422 | Payment not refundable |
| `*RefundExceedsRemainingError` | 422 | Refund exceeds remaining amount |
| `*PartialRefundUnsupportedError` | 422 | Bank only supports full refunds |
| `*APIError` | 5xx | Generic server error |

### Client-side amount validation

Pre-check an amount against a bank's live capabilities before sending a request:

```go
result, err := kp.ValidateAmount(ctx, 100, kosovopay.CurrencyEUR, kosovopay.BankCodeOnefor)
if !result.Valid {
    fmt.Printf("invalid: %s (code: %s)\n", result.Message, result.Code)
    if result.Code == "amount_step_invalid" {
        fmt.Printf("nearest valid: %d or %d\n", result.NearestValid[0], result.NearestValid[1])
    }
}
```

### Money helpers

```go
m := kosovopay.Money{}

// Format minor units as a decimal string
m.Format(1999, 2) // "19.99"
m.Format(500, 2)  // "5.00"

// Convert between currencies using a rate string
m.Convert(1000, "1.08") // 1080
```

### Configuration options

```go
kp := kosovopay.New("sk_test_...",
    kosovopay.WithBaseURL("https://api.kosovo.sh"),
    kosovopay.WithAPIVersion("2026-06-01"),
    kosovopay.WithTimeout(15*time.Second),
    kosovopay.WithMaxRetries(3),
)
```

## Idempotency

Every mutating request (`POST`) automatically generates a UUID `Idempotency-Key`
header. To supply your own key (e.g. tied to an order ID), pass it as the last
argument:

```go
payment, err := kp.Payments.Create(ctx, params, "order-9912")
refund, err  := kp.Refunds.Create(ctx, refundParams, "refund-order-9912")
```

Replaying the same key returns the original response — a retry storm can never
double-charge.

## Retry behaviour

The client retries automatically on:

- Network errors
- HTTP 429 (rate limited)
- HTTP 5xx on safe (`GET`, `DELETE`) requests or any request with an `Idempotency-Key`

Retries use exponential back-off. `4xx` errors (except 429) are never retried.

## License

MIT — see [LICENSE](LICENSE).
