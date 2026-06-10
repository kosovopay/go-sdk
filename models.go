package kosovopay

import "time"

// Team is the KosovoPay team associated with an API key.
type Team struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	LogoURL string `json:"logo_url"`
}

// Me is the identity returned by GET /me.
type Me struct {
	Team            Team         `json:"team"`
	Mode            BankMode     `json:"mode"`
	KeyPrefix       string       `json:"key_prefix"`
	EnabledBanks    []BankCode   `json:"enabled_banks"`
	DefaultCurrency CurrencyCode `json:"default_currency"`
}

// RefundCapability describes the refund support for a bank.
type RefundCapability struct {
	Supported bool `json:"supported"`
	Partial   bool `json:"partial"`
}

// BankCapabilities describes what a bank supports.
type BankCapabilities struct {
	Currencies []CurrencyCode   `json:"currencies"`
	MinAmount  int              `json:"min_amount"`
	AmountStep int              `json:"amount_step"`
	Refunds    RefundCapability `json:"refunds"`
}

// Bank represents one bank integration.
type Bank struct {
	Code         BankCode         `json:"code"`
	DisplayName  string           `json:"display_name"`
	LogoURL      string           `json:"logo_url"`
	Enabled      bool             `json:"enabled"`
	Modes        []BankMode       `json:"modes"`
	Capabilities BankCapabilities `json:"capabilities"`
}

// Currency is a supported currency.
type Currency struct {
	Code      CurrencyCode `json:"code"`
	Name      string       `json:"name"`
	Symbol    string       `json:"symbol"`
	Decimals  int          `json:"decimals"`
	IsDefault bool         `json:"is_default"`
}

// Rate is an exchange rate between two currencies.
type Rate struct {
	From     CurrencyCode `json:"from"`
	To       CurrencyCode `json:"to"`
	Rate     string       `json:"rate"`
	SyncedAt string       `json:"synced_at"`
	Stale    bool         `json:"stale"`
}

// Payer holds optional payer identity attached to a payment.
type Payer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Fx holds foreign-exchange details for a payment.
type Fx struct {
	From CurrencyCode `json:"from"`
	To   CurrencyCode `json:"to"`
	Rate string       `json:"rate"`
}

// LineItem is one line in a payment's order.
type LineItem struct {
	Name            string `json:"name"`
	Quantity        int    `json:"quantity"`
	UnitAmountCents int    `json:"unit_amount_cents"`
	SKU             string `json:"sku,omitempty"`
	ImageURL        string `json:"image_url,omitempty"`
	Variant         string `json:"variant,omitempty"`
}

// Refund is a refund issued against a captured Payment.
type Refund struct {
	ID            string       `json:"id"`
	Payment       string       `json:"payment"`
	Amount        int          `json:"amount"`
	Status        RefundStatus `json:"status"`
	Reason        RefundReason `json:"reason"`
	FailureReason string       `json:"failure_reason"`
	Created       int64        `json:"created"`
	SucceededAt   int64        `json:"succeeded_at"`
}

// CreatedAt returns the refund creation time.
func (r *Refund) CreatedAt() time.Time {
	if r.Created == 0 {
		return time.Time{}
	}
	return time.Unix(r.Created, 0)
}

// Payment is a payment object returned by the API.
type Payment struct {
	ID                string                 `json:"id"`
	Status            PaymentStatus          `json:"status"`
	Mode              BankMode               `json:"mode"`
	Amount            int                    `json:"amount"`
	AmountCaptured    int                    `json:"amount_captured"`
	AmountRefunded    int                    `json:"amount_refunded"`
	Currency          CurrencyCode           `json:"currency"`
	BankCode          BankCode               `json:"bank_code"`
	MerchantReference string                 `json:"merchant_reference"`
	Description       string                 `json:"description"`
	Payer             *Payer                 `json:"payer"`
	LineItems         []LineItem             `json:"line_items"`
	Metadata          map[string]interface{} `json:"metadata"`
	Fx                *Fx                    `json:"fx"`
	LastError         string                 `json:"last_error"`
	ExpiresAt         int64                  `json:"expires_at"`
	CapturedAt        int64                  `json:"captured_at"`
	Created           int64                  `json:"created"`
	Refunds           []Refund               `json:"refunds"`
	CheckoutMode      CheckoutMode           `json:"checkout_mode"`
	HostedURL         string                 `json:"hosted_url"`
	RedirectURL       string                 `json:"redirect_url"`
}

// CreatedAt returns the payment creation time.
func (p *Payment) CreatedAt() time.Time {
	if p.Created == 0 {
		return time.Time{}
	}
	return time.Unix(p.Created, 0)
}

// TimelineEvent is one entry in a payment's audit timeline.
type TimelineEvent struct {
	Type string `json:"type"`
	At   int64  `json:"at"`
}

// Time returns the timeline event time.
func (t *TimelineEvent) Time() time.Time {
	return time.Unix(t.At, 0)
}

// WebhookEndpoint is a registered webhook destination.
type WebhookEndpoint struct {
	ID            string             `json:"id"`
	URL           string             `json:"url"`
	Description   string             `json:"description"`
	EnabledEvents []WebhookEventType `json:"enabled_events"`
	Status        string             `json:"status"`
	Mode          BankMode           `json:"mode"`
	Created       int64              `json:"created"`
	Secret        string             `json:"secret"`
}

// Event is a webhook event delivered to an endpoint.
type Event struct {
	ID                 string                 `json:"id"`
	Type               WebhookEventType       `json:"type"`
	Created            int64                  `json:"created"`
	Livemode           bool                   `json:"livemode"`
	APIVersion         string                 `json:"api_version"`
	Data               map[string]interface{} `json:"data"`
	Object             map[string]interface{} `json:"-"`
	PreviousAttributes map[string]interface{} `json:"-"`
}

// CreatedAt returns the event creation time.
func (e *Event) CreatedAt() time.Time {
	return time.Unix(e.Created, 0)
}

// AsPayment hydrates the event's data.object as a Payment.
// Only valid for payment.* events.
func (e *Event) AsPayment() (*Payment, error) {
	obj, ok := e.Data["object"].(map[string]interface{})
	if !ok {
		return nil, &KosovoPayError{Message: "event data.object is not a payment object"}
	}
	return paymentFromMap(obj), nil
}

// AsRefund hydrates the event's data.object as a Refund.
// Only valid for refund.* events.
func (e *Event) AsRefund() (*Refund, error) {
	obj, ok := e.Data["object"].(map[string]interface{})
	if !ok {
		return nil, &KosovoPayError{Message: "event data.object is not a refund object"}
	}
	return refundFromMap(obj), nil
}

// DeletedResource is the response from a DELETE endpoint.
type DeletedResource struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// Collection is a typed list response from a list endpoint (non-paginated).
type Collection[T any] struct {
	Object  string `json:"object"`
	Data    []T    `json:"data"`
	HasMore bool   `json:"has_more"`
	URL     string `json:"url"`
}

// AmountValidation is the result of a local amount pre-check.
type AmountValidation struct {
	Valid        bool
	Code         string
	Message      string
	NearestValid [2]int // [lower, upper] valid amounts; zero-value if Valid is true
}

// paymentFromMap deserialises a raw JSON map into a Payment, applying
// forward-compatible enum parsing.
func paymentFromMap(d map[string]interface{}) *Payment {
	p := &Payment{}
	p.ID, _ = d["id"].(string)
	if s, ok := d["status"].(string); ok {
		p.Status = PaymentStatusFromWire(s)
	}
	if s, ok := d["mode"].(string); ok {
		p.Mode = BankModeFromWire(s)
	}
	p.Amount = toInt(d["amount"])
	p.AmountCaptured = toInt(d["amount_captured"])
	p.AmountRefunded = toInt(d["amount_refunded"])
	if s, ok := d["currency"].(string); ok {
		p.Currency = CurrencyCodeFromWire(s)
	}
	if s, ok := d["bank_code"].(string); ok {
		p.BankCode = BankCodeFromWire(s)
	}
	p.MerchantReference, _ = d["merchant_reference"].(string)
	p.Description, _ = d["description"].(string)
	if pm, ok := d["payer"].(map[string]interface{}); ok {
		p.Payer = &Payer{}
		p.Payer.Name, _ = pm["name"].(string)
		p.Payer.Email, _ = pm["email"].(string)
	}
	if items, ok := d["line_items"].([]interface{}); ok {
		for _, item := range items {
			if m, ok := item.(map[string]interface{}); ok {
				li := LineItem{}
				li.Name, _ = m["name"].(string)
				li.Quantity = toInt(m["quantity"])
				li.UnitAmountCents = toInt(m["unit_amount_cents"])
				li.SKU, _ = m["sku"].(string)
				li.ImageURL, _ = m["image_url"].(string)
				li.Variant, _ = m["variant"].(string)
				p.LineItems = append(p.LineItems, li)
			}
		}
	}
	if meta, ok := d["metadata"].(map[string]interface{}); ok {
		p.Metadata = meta
	}
	if fx, ok := d["fx"].(map[string]interface{}); ok {
		p.Fx = &Fx{}
		if s, ok := fx["from"].(string); ok {
			p.Fx.From = CurrencyCodeFromWire(s)
		}
		if s, ok := fx["to"].(string); ok {
			p.Fx.To = CurrencyCodeFromWire(s)
		}
		p.Fx.Rate, _ = fx["rate"].(string)
	}
	p.LastError, _ = d["last_error"].(string)
	p.ExpiresAt = toInt64(d["expires_at"])
	p.CapturedAt = toInt64(d["captured_at"])
	p.Created = toInt64(d["created"])
	if refunds, ok := d["refunds"].([]interface{}); ok {
		for _, r := range refunds {
			if m, ok := r.(map[string]interface{}); ok {
				p.Refunds = append(p.Refunds, *refundFromMap(m))
			}
		}
	}
	if s, ok := d["checkout_mode"].(string); ok {
		p.CheckoutMode = CheckoutMode(s)
	}
	p.HostedURL, _ = d["hosted_url"].(string)
	p.RedirectURL, _ = d["redirect_url"].(string)
	return p
}

// refundFromMap deserialises a raw JSON map into a Refund.
func refundFromMap(d map[string]interface{}) *Refund {
	r := &Refund{}
	r.ID, _ = d["id"].(string)
	r.Payment, _ = d["payment"].(string)
	r.Amount = toInt(d["amount"])
	if s, ok := d["status"].(string); ok {
		r.Status = RefundStatusFromWire(s)
	}
	if s, ok := d["reason"].(string); ok {
		r.Reason = RefundReason(s)
	}
	r.FailureReason, _ = d["failure_reason"].(string)
	r.Created = toInt64(d["created"])
	r.SucceededAt = toInt64(d["succeeded_at"])
	return r
}

// toInt safely converts an interface{} JSON number to int.
func toInt(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	}
	return 0
}

// toInt64 safely converts an interface{} JSON number to int64.
func toInt64(v interface{}) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case int64:
		return n
	}
	return 0
}
