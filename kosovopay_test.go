package kosovopay_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	kosovopay "github.com/kosovopay/go-sdk"
)

// ---- test helpers ----

// newTestClient creates a Client pointed at a test server.
func newTestClient(t *testing.T, handler http.Handler) (*kosovopay.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := kosovopay.New("sk_test_abc",
		kosovopay.WithBaseURL(srv.URL),
		kosovopay.WithMaxRetries(0),
	)
	return c, srv
}

func jsonBody(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("jsonBody: %v", err)
	}
	return b
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func signPayload(payload []byte, secret string, ts int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(ts, 10)))
	mac.Write([]byte("."))
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("t=%d,v1=%s", ts, sig)
}

// ---- Me ----

func TestMe(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sdk/me" || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"object":           "me",
			"team":             map[string]interface{}{"id": "team_1", "name": "Acme", "logo_url": ""},
			"mode":             "test",
			"key_prefix":       "sk_test",
			"enabled_banks":    []string{"onefor"},
			"default_currency": "EUR",
		})
	})
	c, _ := newTestClient(t, handler)
	me, err := c.Me(context.Background())
	if err != nil {
		t.Fatalf("Me() error: %v", err)
	}
	if me.Team.ID != "team_1" {
		t.Errorf("team.id = %q, want team_1", me.Team.ID)
	}
	if me.Mode != kosovopay.BankModeTest {
		t.Errorf("mode = %q, want test", me.Mode)
	}
	if len(me.EnabledBanks) != 1 || me.EnabledBanks[0] != kosovopay.BankCodeOnefor {
		t.Errorf("enabled_banks = %v, want [onefor]", me.EnabledBanks)
	}
}

// ---- Banks ----

func bankFixture() map[string]interface{} {
	return map[string]interface{}{
		"code":         "onefor",
		"display_name": "Onefor",
		"logo_url":     "https://example.com/logo.png",
		"enabled":      true,
		"modes":        []string{"test", "live"},
		"capabilities": map[string]interface{}{
			"currencies":  []string{"EUR"},
			"min_amount":  150,
			"amount_step": 1,
			"refunds":     map[string]interface{}{"supported": true, "partial": true},
		},
	}
}

func TestBanksAll(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sdk/banks" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"object":   "list",
			"data":     []interface{}{bankFixture()},
			"has_more": false,
			"url":      "/banks",
		})
	})
	c, _ := newTestClient(t, handler)
	banks, err := c.Banks.All(context.Background())
	if err != nil {
		t.Fatalf("Banks.All() error: %v", err)
	}
	if len(banks) != 1 {
		t.Fatalf("len(banks) = %d, want 1", len(banks))
	}
	if banks[0].Code != kosovopay.BankCodeOnefor {
		t.Errorf("banks[0].Code = %q, want onefor", banks[0].Code)
	}
	if banks[0].Capabilities.MinAmount != 150 {
		t.Errorf("capabilities.min_amount = %d, want 150", banks[0].Capabilities.MinAmount)
	}
}

func TestBanksRetrieve(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sdk/banks/onefor" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, bankFixture())
	})
	c, _ := newTestClient(t, handler)
	bank, err := c.Banks.Retrieve(context.Background(), kosovopay.BankCodeOnefor)
	if err != nil {
		t.Fatalf("Banks.Retrieve() error: %v", err)
	}
	if bank.Code != kosovopay.BankCodeOnefor {
		t.Errorf("code = %q, want onefor", bank.Code)
	}
	if !bank.Enabled {
		t.Error("enabled = false, want true")
	}
}

// ---- Currencies ----

func TestCurrenciesAll(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"object": "list",
			"data": []interface{}{
				map[string]interface{}{"code": "EUR", "name": "Euro", "symbol": "€", "decimals": 2, "is_default": true},
			},
			"has_more": false,
		})
	})
	c, _ := newTestClient(t, handler)
	currencies, err := c.Currencies.All(context.Background())
	if err != nil {
		t.Fatalf("Currencies.All() error: %v", err)
	}
	if len(currencies) != 1 || currencies[0].Code != kosovopay.CurrencyEUR {
		t.Errorf("currencies = %v", currencies)
	}
}

// ---- Rates ----

func TestRatesRetrieve(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("from") != "EUR" || r.URL.Query().Get("to") != "USD" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"from": "EUR", "to": "USD", "rate": "1.08", "synced_at": "2026-06-01T00:00:00Z", "stale": false,
		})
	})
	c, _ := newTestClient(t, handler)
	rate, err := c.Rates.Retrieve(context.Background(), kosovopay.CurrencyEUR, kosovopay.CurrencyUSD)
	if err != nil {
		t.Fatalf("Rates.Retrieve() error: %v", err)
	}
	if rate.Rate != "1.08" {
		t.Errorf("rate = %q, want 1.08", rate.Rate)
	}
}

// ---- Payments ----

func paymentFixture(id string) map[string]interface{} {
	return map[string]interface{}{
		"id":              id,
		"object":          "payment",
		"status":          "pending",
		"mode":            "test",
		"amount":          500,
		"amount_captured": 0,
		"amount_refunded": 0,
		"currency":        "EUR",
		"created":         float64(1717200000),
		"refunds":         []interface{}{},
		"metadata":        map[string]interface{}{},
	}
}

func TestPaymentsCreate(t *testing.T) {
	var gotIdempotencyKey string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sdk/payments" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		gotIdempotencyKey = r.Header.Get("Idempotency-Key")
		writeJSON(w, http.StatusOK, paymentFixture("pi_1"))
	})
	c, _ := newTestClient(t, handler)
	p, err := c.Payments.Create(context.Background(), kosovopay.CreatePaymentParams{
		Amount:     500,
		Currency:   kosovopay.CurrencyEUR,
		SuccessURL: "https://example.com/ok",
		Mode:       kosovopay.CheckoutModeHosted,
	}, "")
	if err != nil {
		t.Fatalf("Payments.Create() error: %v", err)
	}
	if p.ID != "pi_1" {
		t.Errorf("id = %q, want pi_1", p.ID)
	}
	if gotIdempotencyKey == "" {
		t.Error("Idempotency-Key header was not sent")
	}
}

func TestPaymentsCreateIdempotencyKeyOverride(t *testing.T) {
	var gotKey string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("Idempotency-Key")
		writeJSON(w, http.StatusOK, paymentFixture("pi_2"))
	})
	c, _ := newTestClient(t, handler)
	_, err := c.Payments.Create(context.Background(), kosovopay.CreatePaymentParams{
		Amount:     500,
		Currency:   kosovopay.CurrencyEUR,
		SuccessURL: "https://example.com/ok",
	}, "my-custom-key")
	if err != nil {
		t.Fatalf("Payments.Create() error: %v", err)
	}
	if gotKey != "my-custom-key" {
		t.Errorf("Idempotency-Key = %q, want my-custom-key", gotKey)
	}
}

func TestPaymentsRetrieve(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sdk/payments/pi_99" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, paymentFixture("pi_99"))
	})
	c, _ := newTestClient(t, handler)
	p, err := c.Payments.Retrieve(context.Background(), "pi_99")
	if err != nil {
		t.Fatalf("Payments.Retrieve() error: %v", err)
	}
	if p.ID != "pi_99" {
		t.Errorf("id = %q, want pi_99", p.ID)
	}
}

func TestPaymentsTimeline(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sdk/payments/pi_1/timeline" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"object":   "list",
			"data":     []interface{}{map[string]interface{}{"type": "created", "at": float64(1717200000)}},
			"has_more": false,
		})
	})
	c, _ := newTestClient(t, handler)
	events, err := c.Payments.Timeline(context.Background(), "pi_1")
	if err != nil {
		t.Fatalf("Payments.Timeline() error: %v", err)
	}
	if len(events) != 1 || events[0].Type != "created" {
		t.Errorf("events = %v", events)
	}
}

func TestPaymentsCancel(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sdk/payments/pi_1/cancel" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		p := paymentFixture("pi_1")
		p["status"] = "canceled"
		writeJSON(w, http.StatusOK, p)
	})
	c, _ := newTestClient(t, handler)
	p, err := c.Payments.Cancel(context.Background(), "pi_1", "")
	if err != nil {
		t.Fatalf("Payments.Cancel() error: %v", err)
	}
	if p.Status != kosovopay.PaymentStatusCanceled {
		t.Errorf("status = %q, want canceled", p.Status)
	}
}

// ---- Cursor pagination across 2 pages ----

func TestPaymentsPaginationTwoPages(t *testing.T) {
	page := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		switch page {
		case 1:
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"object":   "list",
				"data":     []interface{}{paymentFixture("pi_A"), paymentFixture("pi_B")},
				"has_more": true,
				"url":      "/payments",
			})
		case 2:
			// Verify starting_after was passed
			if r.URL.Query().Get("starting_after") != "pi_B" {
				t.Errorf("starting_after = %q, want pi_B", r.URL.Query().Get("starting_after"))
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"object":   "list",
				"data":     []interface{}{paymentFixture("pi_C")},
				"has_more": false,
				"url":      "/payments",
			})
		default:
			t.Error("unexpected extra page request")
			http.Error(w, "unexpected", http.StatusInternalServerError)
		}
	})
	c, _ := newTestClient(t, handler)
	all, err := c.Payments.All(context.Background(), nil)
	if err != nil {
		t.Fatalf("Payments.All() error: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("len(all) = %d, want 3", len(all))
	}
	ids := []string{all[0].ID, all[1].ID, all[2].ID}
	want := []string{"pi_A", "pi_B", "pi_C"}
	for i, id := range ids {
		if id != want[i] {
			t.Errorf("all[%d].ID = %q, want %q", i, id, want[i])
		}
	}
}

// ---- Refunds ----

func refundFixture(id string) map[string]interface{} {
	return map[string]interface{}{
		"id":             id,
		"payment":        "pi_1",
		"amount":         200,
		"status":         "succeeded",
		"reason":         "requested_by_customer",
		"failure_reason": nil,
		"created":        float64(1717200000),
		"succeeded_at":   float64(1717200010),
	}
}

func TestRefundsCreate(t *testing.T) {
	var gotKey string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("Idempotency-Key")
		writeJSON(w, http.StatusOK, refundFixture("re_1"))
	})
	c, _ := newTestClient(t, handler)
	ref, err := c.Refunds.Create(context.Background(), kosovopay.CreateRefundParams{
		Payment: "pi_1",
		Amount:  200,
		Reason:  kosovopay.RefundReasonRequestedByCustomer,
	}, "")
	if err != nil {
		t.Fatalf("Refunds.Create() error: %v", err)
	}
	if ref.ID != "re_1" {
		t.Errorf("id = %q, want re_1", ref.ID)
	}
	if gotKey == "" {
		t.Error("Idempotency-Key header was not sent")
	}
}

func TestRefundsRetrieve(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, refundFixture("re_1"))
	})
	c, _ := newTestClient(t, handler)
	ref, err := c.Refunds.Retrieve(context.Background(), "re_1")
	if err != nil {
		t.Fatalf("Refunds.Retrieve() error: %v", err)
	}
	if ref.Status != kosovopay.RefundStatusSucceeded {
		t.Errorf("status = %q, want succeeded", ref.Status)
	}
}

func TestRefundsPaginationTwoPages(t *testing.T) {
	page := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		switch page {
		case 1:
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"object":   "list",
				"data":     []interface{}{refundFixture("re_A"), refundFixture("re_B")},
				"has_more": true,
			})
		case 2:
			if r.URL.Query().Get("starting_after") != "re_B" {
				t.Errorf("starting_after = %q, want re_B", r.URL.Query().Get("starting_after"))
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"object":   "list",
				"data":     []interface{}{refundFixture("re_C")},
				"has_more": false,
			})
		}
	})
	c, _ := newTestClient(t, handler)
	all, err := c.Refunds.All(context.Background(), nil)
	if err != nil {
		t.Fatalf("Refunds.All() error: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("len(all) = %d, want 3", len(all))
	}
}

// ---- WebhookEndpoints ----

func webhookEndpointFixture(id string) map[string]interface{} {
	return map[string]interface{}{
		"id":             id,
		"url":            "https://example.com/hooks",
		"description":    "test",
		"enabled_events": []interface{}{"payment.captured"},
		"status":         "enabled",
		"mode":           "test",
		"created":        float64(1717200000),
		"secret":         "whsec_abc",
	}
}

func TestWebhookEndpointsCreate(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, webhookEndpointFixture("we_1"))
	})
	c, _ := newTestClient(t, handler)
	ep, err := c.WebhookEndpoints.Create(context.Background(), kosovopay.CreateWebhookEndpointParams{
		URL:           "https://example.com/hooks",
		EnabledEvents: []kosovopay.WebhookEventType{kosovopay.WebhookEventTypePaymentCaptured},
	})
	if err != nil {
		t.Fatalf("WebhookEndpoints.Create() error: %v", err)
	}
	if ep.ID != "we_1" {
		t.Errorf("id = %q, want we_1", ep.ID)
	}
	if ep.Secret != "whsec_abc" {
		t.Errorf("secret = %q, want whsec_abc", ep.Secret)
	}
}

func TestWebhookEndpointsAll(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"object":   "list",
			"data":     []interface{}{webhookEndpointFixture("we_1")},
			"has_more": false,
		})
	})
	c, _ := newTestClient(t, handler)
	eps, err := c.WebhookEndpoints.All(context.Background())
	if err != nil {
		t.Fatalf("WebhookEndpoints.All() error: %v", err)
	}
	if len(eps) != 1 || eps[0].ID != "we_1" {
		t.Errorf("endpoints = %v", eps)
	}
}

func TestWebhookEndpointsDelete(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"id": "we_1", "deleted": true})
	})
	c, _ := newTestClient(t, handler)
	del, err := c.WebhookEndpoints.Delete(context.Background(), "we_1")
	if err != nil {
		t.Fatalf("WebhookEndpoints.Delete() error: %v", err)
	}
	if !del.Deleted {
		t.Error("deleted = false, want true")
	}
}

func TestWebhookEndpointsRotateSecret(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ep := webhookEndpointFixture("we_1")
		ep["secret"] = "whsec_new"
		writeJSON(w, http.StatusOK, ep)
	})
	c, _ := newTestClient(t, handler)
	ep, err := c.WebhookEndpoints.RotateSecret(context.Background(), "we_1")
	if err != nil {
		t.Fatalf("WebhookEndpoints.RotateSecret() error: %v", err)
	}
	if ep.Secret != "whsec_new" {
		t.Errorf("secret = %q, want whsec_new", ep.Secret)
	}
}

// ---- Error mapping ----

func TestErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		body       map[string]interface{}
		statusCode int
		wantType   interface{}
	}{
		{
			name: "validation_error",
			body: map[string]interface{}{"error": map[string]interface{}{
				"type": "validation_error", "code": "invalid_request", "message": "bad request",
			}},
			statusCode: 422,
			wantType:   &kosovopay.ValidationError{},
		},
		{
			name: "authentication_error",
			body: map[string]interface{}{"error": map[string]interface{}{
				"type": "authentication_error", "code": "invalid_key", "message": "bad key",
			}},
			statusCode: 401,
			wantType:   &kosovopay.AuthenticationError{},
		},
		{
			name: "permission_error",
			body: map[string]interface{}{"error": map[string]interface{}{
				"type": "permission_error", "code": "forbidden", "message": "no permission",
			}},
			statusCode: 403,
			wantType:   &kosovopay.PermissionError{},
		},
		{
			name: "rate_limit",
			body: map[string]interface{}{"error": map[string]interface{}{
				"type": "rate_limit_error", "code": "rate_limited", "message": "too many requests",
			}},
			statusCode: 429,
			wantType:   &kosovopay.RateLimitError{},
		},
		{
			name: "amount_below_minimum",
			body: map[string]interface{}{"error": map[string]interface{}{
				"type": "payment_error", "code": "amount_below_minimum", "message": "too small",
			}},
			statusCode: 422,
			wantType:   &kosovopay.AmountBelowMinimumError{},
		},
		{
			name: "amount_step_invalid",
			body: map[string]interface{}{"error": map[string]interface{}{
				"type": "payment_error", "code": "amount_step_invalid", "message": "bad step",
			}},
			statusCode: 422,
			wantType:   &kosovopay.AmountStepInvalidError{},
		},
		{
			name: "bank_not_enabled",
			body: map[string]interface{}{"error": map[string]interface{}{
				"type": "payment_error", "code": "bank_not_enabled", "message": "not enabled",
			}},
			statusCode: 422,
			wantType:   &kosovopay.BankNotEnabledError{},
		},
		{
			name: "bank_unreachable",
			body: map[string]interface{}{"error": map[string]interface{}{
				"type": "payment_error", "code": "bank_unreachable", "message": "unreachable",
			}},
			statusCode: 502,
			wantType:   &kosovopay.BankUnreachableError{},
		},
		{
			name: "payment_not_cancelable",
			body: map[string]interface{}{"error": map[string]interface{}{
				"type": "payment_error", "code": "payment_not_cancelable", "message": "not cancelable",
			}},
			statusCode: 422,
			wantType:   &kosovopay.PaymentNotCancelableError{},
		},
		{
			name: "payment_not_refundable",
			body: map[string]interface{}{"error": map[string]interface{}{
				"type": "payment_error", "code": "payment_not_refundable", "message": "not refundable",
			}},
			statusCode: 422,
			wantType:   &kosovopay.PaymentNotRefundableError{},
		},
		{
			name: "refund_exceeds_remaining",
			body: map[string]interface{}{"error": map[string]interface{}{
				"type": "payment_error", "code": "refund_exceeds_remaining", "message": "exceeds",
			}},
			statusCode: 422,
			wantType:   &kosovopay.RefundExceedsRemainingError{},
		},
		{
			name: "partial_refund_unsupported",
			body: map[string]interface{}{"error": map[string]interface{}{
				"type": "payment_error", "code": "partial_refund_unsupported", "message": "no partial",
			}},
			statusCode: 422,
			wantType:   &kosovopay.PartialRefundUnsupportedError{},
		},
		{
			name: "idempotency_error",
			body: map[string]interface{}{"error": map[string]interface{}{
				"type": "idempotency_error", "code": "idempotency_payload_mismatch", "message": "mismatch",
			}},
			statusCode: 409,
			wantType:   &kosovopay.IdempotencyError{},
		},
		{
			name: "api_error",
			body: map[string]interface{}{"error": map[string]interface{}{
				"type": "api_error", "code": "internal_error", "message": "server error",
			}},
			statusCode: 500,
			wantType:   &kosovopay.APIError{},
		},
		{
			name: "unknown_code_falls_back_to_type",
			body: map[string]interface{}{"error": map[string]interface{}{
				"type": "payment_error", "code": "some_future_code_xyz", "message": "future error",
			}},
			statusCode: 422,
			wantType:   &kosovopay.PaymentError{},
		},
		{
			name: "unknown_code_and_type_falls_back_to_api_error",
			body: map[string]interface{}{"error": map[string]interface{}{
				"type": "future_error_type", "code": "some_future_code", "message": "future",
			}},
			statusCode: 500,
			wantType:   &kosovopay.APIError{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, tc.statusCode, tc.body)
			})
			c, _ := newTestClient(t, handler)
			_, err := c.Payments.Retrieve(context.Background(), "pi_test")
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			// Use a type switch to verify the concrete error type.
			got := false
			switch tc.wantType.(type) {
			case *kosovopay.ValidationError:
				var e *kosovopay.ValidationError
				got = isType[*kosovopay.ValidationError](err, &e)
			case *kosovopay.AuthenticationError:
				var e *kosovopay.AuthenticationError
				got = isType[*kosovopay.AuthenticationError](err, &e)
			case *kosovopay.PermissionError:
				var e *kosovopay.PermissionError
				got = isType[*kosovopay.PermissionError](err, &e)
			case *kosovopay.RateLimitError:
				var e *kosovopay.RateLimitError
				got = isType[*kosovopay.RateLimitError](err, &e)
			case *kosovopay.AmountBelowMinimumError:
				var e *kosovopay.AmountBelowMinimumError
				got = isType[*kosovopay.AmountBelowMinimumError](err, &e)
			case *kosovopay.AmountStepInvalidError:
				var e *kosovopay.AmountStepInvalidError
				got = isType[*kosovopay.AmountStepInvalidError](err, &e)
			case *kosovopay.BankNotEnabledError:
				var e *kosovopay.BankNotEnabledError
				got = isType[*kosovopay.BankNotEnabledError](err, &e)
			case *kosovopay.BankUnreachableError:
				var e *kosovopay.BankUnreachableError
				got = isType[*kosovopay.BankUnreachableError](err, &e)
			case *kosovopay.PaymentNotCancelableError:
				var e *kosovopay.PaymentNotCancelableError
				got = isType[*kosovopay.PaymentNotCancelableError](err, &e)
			case *kosovopay.PaymentNotRefundableError:
				var e *kosovopay.PaymentNotRefundableError
				got = isType[*kosovopay.PaymentNotRefundableError](err, &e)
			case *kosovopay.RefundExceedsRemainingError:
				var e *kosovopay.RefundExceedsRemainingError
				got = isType[*kosovopay.RefundExceedsRemainingError](err, &e)
			case *kosovopay.PartialRefundUnsupportedError:
				var e *kosovopay.PartialRefundUnsupportedError
				got = isType[*kosovopay.PartialRefundUnsupportedError](err, &e)
			case *kosovopay.IdempotencyError:
				var e *kosovopay.IdempotencyError
				got = isType[*kosovopay.IdempotencyError](err, &e)
			case *kosovopay.PaymentError:
				var e *kosovopay.PaymentError
				got = isType[*kosovopay.PaymentError](err, &e)
			case *kosovopay.APIError:
				var e *kosovopay.APIError
				got = isType[*kosovopay.APIError](err, &e)
			}
			if !got {
				t.Errorf("error type = %T, want %T", err, tc.wantType)
			}
		})
	}
}

func isType[T any](err error, out *T) bool {
	v, ok := err.(T)
	if ok {
		*out = v
	}
	return ok
}

func TestRateLimitRetryAfter(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		writeJSON(w, 429, map[string]interface{}{"error": map[string]interface{}{
			"type": "rate_limit_error", "code": "rate_limited", "message": "slow down",
		}})
	})
	c, _ := newTestClient(t, handler)
	_, err := c.Payments.Retrieve(context.Background(), "pi_1")
	var rle *kosovopay.RateLimitError
	if !isType(err, &rle) {
		t.Fatalf("expected RateLimitError, got %T", err)
	}
	if rle.RetryAfter != 30 {
		t.Errorf("RetryAfter = %d, want 30", rle.RetryAfter)
	}
}

// ---- Webhook verification ----

func TestWebhookConstructEvent_Valid(t *testing.T) {
	secret := "whsec_testsecret"
	payload := []byte(`{"id":"evt_1","type":"payment.captured","created":1717200000,"livemode":false,"api_version":"2026-06-01","data":{"object":{}}}`)
	ts := time.Now().Unix()
	header := signPayload(payload, secret, ts)

	event, err := kosovopay.ConstructEvent(payload, header, secret, 300)
	if err != nil {
		t.Fatalf("ConstructEvent() error: %v", err)
	}
	if event.Type != kosovopay.WebhookEventTypePaymentCaptured {
		t.Errorf("type = %q, want payment.captured", event.Type)
	}
}

func TestWebhookConstructEvent_Tampered(t *testing.T) {
	secret := "whsec_testsecret"
	payload := []byte(`{"id":"evt_1","type":"payment.captured","created":1717200000,"livemode":false,"api_version":"2026-06-01","data":{"object":{}}}`)
	ts := time.Now().Unix()
	header := signPayload(payload, secret, ts)

	// Tamper the payload after signing.
	tampered := []byte(`{"id":"evt_1","type":"payment.failed","created":1717200000,"livemode":false,"api_version":"2026-06-01","data":{"object":{}}}`)
	_, err := kosovopay.ConstructEvent(tampered, header, secret, 300)
	if err == nil {
		t.Fatal("expected error for tampered payload, got nil")
	}
	var sigErr *kosovopay.WebhookSignatureError
	if !isType(err, &sigErr) {
		t.Errorf("error type = %T, want *WebhookSignatureError", err)
	}
}

func TestWebhookConstructEvent_Stale(t *testing.T) {
	secret := "whsec_testsecret"
	payload := []byte(`{"id":"evt_1","type":"payment.captured","created":1717200000,"livemode":false,"api_version":"2026-06-01","data":{"object":{}}}`)
	// Use a timestamp 10 minutes in the past — outside the 5-minute window.
	staleTS := time.Now().Unix() - 600
	header := signPayload(payload, secret, staleTS)

	_, err := kosovopay.ConstructEvent(payload, header, secret, 300)
	if err == nil {
		t.Fatal("expected error for stale timestamp, got nil")
	}
}

func TestWebhookVerify_MissingHeader(t *testing.T) {
	payload := []byte(`{}`)
	err := kosovopay.Verify(payload, "", "secret", 0, 300)
	if err == nil {
		t.Fatal("expected error for empty header, got nil")
	}
}

// ---- Money helpers ----

func TestMoneyFormat(t *testing.T) {
	m := kosovopay.Money{}
	tests := []struct {
		amount   int
		decimals int
		want     string
	}{
		{1999, 2, "19.99"},
		{500, 2, "5.00"},
		{100, 0, "100"},
		{0, 2, "0.00"},
		{1000, 3, "1.000"},
	}
	for _, tc := range tests {
		got := m.Format(tc.amount, tc.decimals)
		if got != tc.want {
			t.Errorf("Format(%d, %d) = %q, want %q", tc.amount, tc.decimals, got, tc.want)
		}
	}
}

func TestMoneyConvert(t *testing.T) {
	m := kosovopay.Money{}
	tests := []struct {
		amount int
		rate   string
		want   int
	}{
		{1000, "1.08", 1080},
		{100, "0.5", 50},
		{100, "0", 0},
		{100, "invalid", 0},
	}
	for _, tc := range tests {
		got := m.Convert(tc.amount, tc.rate)
		if got != tc.want {
			t.Errorf("Convert(%d, %q) = %d, want %d", tc.amount, tc.rate, got, tc.want)
		}
	}
}

// ---- Amount validation ----

func TestValidateAmount(t *testing.T) {
	bank := &kosovopay.Bank{
		DisplayName: "Onefor",
		Capabilities: kosovopay.BankCapabilities{
			Currencies: []kosovopay.CurrencyCode{kosovopay.CurrencyEUR},
			MinAmount:  150,
			AmountStep: 50,
		},
	}

	tests := []struct {
		name      string
		amount    int
		currency  kosovopay.CurrencyCode
		wantValid bool
		wantCode  string
	}{
		{"valid", 200, kosovopay.CurrencyEUR, true, ""},
		{"below_minimum", 100, kosovopay.CurrencyEUR, false, "amount_below_minimum"},
		{"step_invalid", 175, kosovopay.CurrencyEUR, false, "amount_step_invalid"},
		{"currency_not_supported", 200, kosovopay.CurrencyUSD, false, "currency_not_supported"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := kosovopay.ValidateAmountLocal(bank, tc.amount, tc.currency)
			if result.Valid != tc.wantValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tc.wantValid)
			}
			if result.Code != tc.wantCode {
				t.Errorf("Code = %q, want %q", result.Code, tc.wantCode)
			}
		})
	}
}

func TestValidateAmountNearestValid(t *testing.T) {
	bank := &kosovopay.Bank{
		DisplayName: "Onefor",
		Capabilities: kosovopay.BankCapabilities{
			Currencies: []kosovopay.CurrencyCode{kosovopay.CurrencyEUR},
			MinAmount:  150,
			AmountStep: 50,
		},
	}
	result := kosovopay.ValidateAmountLocal(bank, 175, kosovopay.CurrencyEUR)
	if result.Valid {
		t.Fatal("expected invalid")
	}
	if result.NearestValid[0] != 150 || result.NearestValid[1] != 200 {
		t.Errorf("NearestValid = %v, want [150 200]", result.NearestValid)
	}
}

// ---- Enum forward-compat ----

func TestUnknownEnumValues(t *testing.T) {
	if kosovopay.PaymentStatusFromWire("not_a_real_status") != kosovopay.PaymentStatusUnknown {
		t.Error("PaymentStatusFromWire unknown")
	}
	if kosovopay.RefundStatusFromWire("not_a_real_status") != kosovopay.RefundStatusUnknown {
		t.Error("RefundStatusFromWire unknown")
	}
	if kosovopay.BankCodeFromWire("not_a_real_bank") != kosovopay.BankCodeUnknown {
		t.Error("BankCodeFromWire unknown")
	}
	if kosovopay.CurrencyCodeFromWire("ZZZ") != kosovopay.CurrencyUnknown {
		t.Error("CurrencyCodeFromWire unknown")
	}
	if kosovopay.WebhookEventTypeFromWire("future.event") != kosovopay.WebhookEventTypeUnknown {
		t.Error("WebhookEventTypeFromWire unknown")
	}
}

// ---- CreatePaymentParams validation ----

func TestCreatePaymentParamsValidation(t *testing.T) {
	tests := []struct {
		name    string
		params  kosovopay.CreatePaymentParams
		wantErr bool
	}{
		{
			name: "valid hosted",
			params: kosovopay.CreatePaymentParams{
				Amount: 500, Currency: kosovopay.CurrencyEUR,
				SuccessURL: "https://example.com/ok", Mode: kosovopay.CheckoutModeHosted,
			},
			wantErr: false,
		},
		{
			name: "valid direct",
			params: kosovopay.CreatePaymentParams{
				Amount: 500, Currency: kosovopay.CurrencyEUR,
				SuccessURL: "https://example.com/ok",
				Mode:       kosovopay.CheckoutModeDirect, BankCode: kosovopay.BankCodeOnefor,
			},
			wantErr: false,
		},
		{
			name: "zero amount",
			params: kosovopay.CreatePaymentParams{
				Amount: 0, Currency: kosovopay.CurrencyEUR, SuccessURL: "https://example.com/ok",
			},
			wantErr: true,
		},
		{
			name: "direct without bank_code",
			params: kosovopay.CreatePaymentParams{
				Amount: 500, Currency: kosovopay.CurrencyEUR,
				SuccessURL: "https://example.com/ok", Mode: kosovopay.CheckoutModeDirect,
			},
			wantErr: true,
		},
		{
			name: "hosted with bank_code",
			params: kosovopay.CreatePaymentParams{
				Amount: 500, Currency: kosovopay.CurrencyEUR,
				SuccessURL: "https://example.com/ok",
				Mode:       kosovopay.CheckoutModeHosted, BankCode: kosovopay.BankCodeOnefor,
			},
			wantErr: true,
		},
		{
			name: "invalid success_url",
			params: kosovopay.CreatePaymentParams{
				Amount: 500, Currency: kosovopay.CurrencyEUR, SuccessURL: "not-a-url",
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// ---- Headers ----

func TestAuthorizationHeader(t *testing.T) {
	var gotAuth string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"object": "me", "team": map[string]interface{}{"id": "t1", "name": "T", "logo_url": ""},
			"mode": "test", "key_prefix": "sk_test", "enabled_banks": []string{}, "default_currency": "EUR",
		})
	})
	c, _ := newTestClient(t, handler)
	_, err := c.Me(context.Background())
	if err != nil {
		t.Fatalf("Me() error: %v", err)
	}
	if gotAuth != "Bearer sk_test_abc" {
		t.Errorf("Authorization = %q, want Bearer sk_test_abc", gotAuth)
	}
}

func TestVersionHeader(t *testing.T) {
	var gotVersion string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotVersion = r.Header.Get("Kosovopay-Version")
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"object": "me", "team": map[string]interface{}{"id": "t1", "name": "T", "logo_url": ""},
			"mode": "test", "key_prefix": "sk_test", "enabled_banks": []string{}, "default_currency": "EUR",
		})
	})
	c, _ := newTestClient(t, handler)
	_, err := c.Me(context.Background())
	if err != nil {
		t.Fatalf("Me() error: %v", err)
	}
	if gotVersion != kosovopay.DefaultAPIVersion {
		t.Errorf("Kosovopay-Version = %q, want %q", gotVersion, kosovopay.DefaultAPIVersion)
	}
}
