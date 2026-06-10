package kosovopay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

const (
	// Version is the SDK version.
	Version = "1.1.2"
	// DefaultBaseURL is the KosovoPay API base URL.
	DefaultBaseURL = "https://api.kosovo.sh"
	// DefaultAPIVersion is the Kosovopay-Version header value.
	DefaultAPIVersion = "2026-06-01"
	// apiPrefix is the path prefix for all SDK routes.
	apiPrefix = "/api/sdk"
)

// Config holds immutable client configuration.
type Config struct {
	// APIKey is the secret key (sk_test_… or sk_live_…).
	APIKey string
	// BaseURL overrides the default API base URL.
	BaseURL string
	// APIVersion is the Kosovopay-Version header value.
	APIVersion string
	// ConnectTimeout is the TCP connection timeout (default 10s).
	ConnectTimeout time.Duration
	// RequestTimeout is the total request timeout (default 30s).
	RequestTimeout time.Duration
	// MaxRetries is the maximum number of retry attempts (default 3).
	MaxRetries int
	// RetryWaitTime is the initial retry back-off (default 500ms).
	RetryWaitTime time.Duration
}

func (c *Config) applyDefaults() {
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
	if c.APIVersion == "" {
		c.APIVersion = DefaultAPIVersion
	}
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = 10 * time.Second
	}
	if c.RequestTimeout == 0 {
		c.RequestTimeout = 30 * time.Second
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}
	if c.RetryWaitTime == 0 {
		c.RetryWaitTime = 500 * time.Millisecond
	}
}

// Client is the KosovoPay API client. Construct with New and access resources
// via the typed fields.
//
//	kp := kosovopay.New("sk_test_...")
//	payment, err := kp.Payments.Create(ctx, params)
type Client struct {
	cfg   Config
	resty *resty.Client

	// Payments provides access to payment operations.
	Payments *PaymentsResource
	// Refunds provides access to refund operations.
	Refunds *RefundsResource
	// Banks provides access to bank operations.
	Banks *BanksResource
	// Currencies provides access to currency operations.
	Currencies *CurrenciesResource
	// Rates provides access to exchange rate operations.
	Rates *RatesResource
	// WebhookEndpoints provides access to webhook endpoint operations.
	WebhookEndpoints *WebhookEndpointsResource
}

// New creates a new KosovoPay client with the given API key and optional
// configuration overrides. Panics if apiKey is empty.
func New(apiKey string, opts ...func(*Config)) *Client {
	if apiKey == "" {
		panic("kosovopay: API key must not be empty")
	}
	cfg := Config{APIKey: apiKey}
	for _, opt := range opts {
		opt(&cfg)
	}
	cfg.applyDefaults()

	r := resty.New().
		SetBaseURL(cfg.BaseURL).
		SetHeader("Authorization", "Bearer "+cfg.APIKey).
		SetHeader("Kosovopay-Version", cfg.APIVersion).
		SetHeader("User-Agent", "kosovopay-go/"+Version).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetTimeout(cfg.RequestTimeout).
		SetRetryCount(cfg.MaxRetries).
		SetRetryWaitTime(cfg.RetryWaitTime).
		SetRetryMaxWaitTime(30 * time.Second).
		AddRetryCondition(func(resp *resty.Response, err error) bool {
			if err != nil {
				return true // network error
			}
			sc := resp.StatusCode()
			if sc == http.StatusTooManyRequests {
				return true
			}
			// Retry 5xx only when the request was safe or had an Idempotency-Key.
			if sc >= 500 {
				method := resp.Request.Method
				ik := resp.Request.Header.Get("Idempotency-Key")
				if method == http.MethodGet || method == http.MethodDelete || ik != "" {
					return true
				}
			}
			return false
		})

	c := &Client{cfg: cfg, resty: r}
	c.Payments = &PaymentsResource{client: c}
	c.Refunds = &RefundsResource{client: c}
	c.Banks = &BanksResource{client: c}
	c.Currencies = &CurrenciesResource{client: c}
	c.Rates = &RatesResource{client: c}
	c.WebhookEndpoints = &WebhookEndpointsResource{client: c}
	return c
}

// WithBaseURL returns an option that overrides the API base URL.
func WithBaseURL(url string) func(*Config) {
	return func(c *Config) { c.BaseURL = url }
}

// WithAPIVersion returns an option that overrides the Kosovopay-Version header.
func WithAPIVersion(v string) func(*Config) {
	return func(c *Config) { c.APIVersion = v }
}

// WithTimeout returns an option that sets both connect and request timeouts.
func WithTimeout(d time.Duration) func(*Config) {
	return func(c *Config) {
		c.ConnectTimeout = d
		c.RequestTimeout = d
	}
}

// WithMaxRetries returns an option that sets the maximum retry count.
func WithMaxRetries(n int) func(*Config) {
	return func(c *Config) { c.MaxRetries = n }
}

// Me retrieves the identity associated with the API key.
func (c *Client) Me(ctx context.Context) (*Me, error) {
	var me Me
	if err := c.get(ctx, "/me", nil, &me); err != nil {
		return nil, err
	}
	// The /me endpoint returns typed enum strings; re-parse them safely.
	return &me, nil
}

// ValidateAmount performs a client-side pre-check of amount against a bank's
// live capabilities. Catches amount_below_minimum and amount_step_invalid
// before a round-trip to the server.
func (c *Client) ValidateAmount(ctx context.Context, amount int, currency CurrencyCode, bankCode BankCode) (*AmountValidation, error) {
	bank, err := c.Banks.Retrieve(ctx, bankCode)
	if err != nil {
		return nil, err
	}
	return validateAmount(bank, amount, currency), nil
}

// ---- internal HTTP helpers ----

// get executes a GET request and JSON-decodes the response into out.
func (c *Client) get(ctx context.Context, path string, query map[string]string, out interface{}) error {
	req := c.resty.R().SetContext(ctx)
	if query != nil {
		req = req.SetQueryParams(query)
	}
	resp, err := req.Get(apiPrefix + path)
	return c.handleResponse(resp, err, out)
}

// post executes a POST request with a JSON body and decodes the response into out.
func (c *Client) post(ctx context.Context, path string, body interface{}, out interface{}, idempotencyKey string) error {
	if idempotencyKey == "" {
		idempotencyKey = uuid.New().String()
	}
	resp, err := c.resty.R().
		SetContext(ctx).
		SetHeader("Idempotency-Key", idempotencyKey).
		SetBody(body).
		Post(apiPrefix + path)
	return c.handleResponse(resp, err, out)
}

// del executes a DELETE request and decodes the response into out.
func (c *Client) del(ctx context.Context, path string, out interface{}) error {
	resp, err := c.resty.R().SetContext(ctx).Delete(apiPrefix + path)
	return c.handleResponse(resp, err, out)
}

// handleResponse checks the HTTP status and either decodes a success body or
// returns a typed error.
func (c *Client) handleResponse(resp *resty.Response, err error, out interface{}) error {
	if err != nil {
		return fmt.Errorf("kosovopay: request failed: %w", err)
	}
	statusCode := resp.StatusCode()
	body := resp.Body()

	if statusCode >= 400 {
		var envelope map[string]interface{}
		_ = json.Unmarshal(body, &envelope)
		retryAfter := 0
		if ra := resp.Header().Get("Retry-After"); ra != "" {
			retryAfter, _ = strconv.Atoi(ra)
		}
		return mapError(envelope, statusCode, retryAfter)
	}

	if out != nil && len(body) > 0 {
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("kosovopay: failed to decode response: %w", err)
		}
	}
	return nil
}
