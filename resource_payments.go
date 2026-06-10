package kosovopay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// CreatePaymentParams are the parameters for creating a payment.
type CreatePaymentParams struct {
	// Amount is in minor units (e.g. 500 = €5.00). Required, must be > 0.
	Amount int `json:"amount"`
	// Currency is the ISO 4217 currency code.
	Currency CurrencyCode `json:"currency"`
	// SuccessURL is the URL to redirect to after a successful payment.
	SuccessURL string `json:"success_url"`
	// Mode is the checkout mode (hosted or direct).
	Mode CheckoutMode `json:"mode,omitempty"`
	// BankCode is required when Mode is CheckoutModeDirect.
	BankCode BankCode `json:"bank_code,omitempty"`
	// CancelURL is the URL to redirect to if the customer cancels.
	CancelURL string `json:"cancel_url,omitempty"`
	// FailURL is the URL to redirect to on failure.
	FailURL string `json:"fail_url,omitempty"`
	// Description is an optional description shown on the checkout page.
	Description string `json:"description,omitempty"`
	// LineItems are the individual order lines.
	LineItems []CreateLineItem `json:"line_items,omitempty"`
	// Metadata is arbitrary key-value data attached to the payment.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	// ExpiresAt is a Unix timestamp for payment expiry.
	ExpiresAt int64 `json:"expires_at,omitempty"`
	// MerchantReference is an optional external reference for idempotent lookup.
	MerchantReference string `json:"merchant_reference,omitempty"`
}

// CreateLineItem is one item in a CreatePaymentParams line item list.
type CreateLineItem struct {
	Name            string `json:"name"`
	Quantity        int    `json:"quantity"`
	UnitAmountCents int    `json:"unit_amount_cents"`
	SKU             string `json:"sku,omitempty"`
	ImageURL        string `json:"image_url,omitempty"`
	Variant         string `json:"variant,omitempty"`
}

// Validate checks CreatePaymentParams for obvious client-side errors.
func (p *CreatePaymentParams) Validate() error {
	if p.Amount <= 0 {
		return fmt.Errorf("kosovopay: amount must be a positive integer in minor units")
	}
	if p.Mode == CheckoutModeDirect && p.BankCode == "" {
		return fmt.Errorf("kosovopay: bank_code is required for direct checkout mode")
	}
	if p.Mode == CheckoutModeHosted && p.BankCode != "" {
		return fmt.Errorf("kosovopay: bank_code must be omitted for hosted checkout mode")
	}
	if err := assertURL(p.SuccessURL, "success_url"); err != nil {
		return err
	}
	if p.CancelURL != "" {
		if err := assertURL(p.CancelURL, "cancel_url"); err != nil {
			return err
		}
	}
	if p.FailURL != "" {
		if err := assertURL(p.FailURL, "fail_url"); err != nil {
			return err
		}
	}
	return nil
}

func assertURL(raw, field string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("kosovopay: %s must be an http or https URL", field)
	}
	return nil
}

// ListPaymentsParams are the query parameters for listing payments.
type ListPaymentsParams struct {
	Limit             int
	StartingAfter     string
	EndingBefore      string
	Status            PaymentStatus
	BankCode          BankCode
	Currency          CurrencyCode
	MerchantReference string
	CreatedGte        int64
	CreatedLte        int64
}

func (p *ListPaymentsParams) toQuery() map[string]string {
	q := map[string]string{}
	if p == nil {
		return q
	}
	if p.Limit > 0 {
		q["limit"] = fmt.Sprintf("%d", p.Limit)
	}
	if p.StartingAfter != "" {
		q["starting_after"] = p.StartingAfter
	}
	if p.EndingBefore != "" {
		q["ending_before"] = p.EndingBefore
	}
	if p.Status != "" {
		q["status"] = string(p.Status)
	}
	if p.BankCode != "" {
		q["bank_code"] = string(p.BankCode)
	}
	if p.Currency != "" {
		q["currency"] = string(p.Currency)
	}
	if p.MerchantReference != "" {
		q["merchant_reference"] = p.MerchantReference
	}
	if p.CreatedGte > 0 {
		q["created[gte]"] = fmt.Sprintf("%d", p.CreatedGte)
	}
	if p.CreatedLte > 0 {
		q["created[lte]"] = fmt.Sprintf("%d", p.CreatedLte)
	}
	return q
}

// PaymentsResource provides access to payment operations.
type PaymentsResource struct {
	client *Client
}

// Create creates a new payment. An Idempotency-Key is auto-generated; pass a
// non-empty idempotencyKey to override it.
func (r *PaymentsResource) Create(ctx context.Context, params CreatePaymentParams, idempotencyKey string) (*Payment, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	if err := r.client.post(ctx, "/payments", params, &raw, idempotencyKey); err != nil {
		return nil, err
	}
	return paymentFromMap(raw), nil
}

// Retrieve fetches a payment by ID.
func (r *PaymentsResource) Retrieve(ctx context.Context, id string) (*Payment, error) {
	var raw map[string]interface{}
	if err := r.client.get(ctx, "/payments/"+id, nil, &raw); err != nil {
		return nil, err
	}
	return paymentFromMap(raw), nil
}

// Iter returns a cursor-paginated iterator over payments matching the params.
// Use iter.Next() / iter.Value() / iter.Err() to walk results.
func (r *PaymentsResource) Iter(ctx context.Context, params *ListPaymentsParams) *PageIter[*Payment] {
	return newPageIter(ctx, r.client, "/payments", params.toQuery(),
		func(row map[string]interface{}) *Payment { return paymentFromMap(row) })
}

// All collects every matching payment across all pages. Use Iter for large result sets.
func (r *PaymentsResource) All(ctx context.Context, params *ListPaymentsParams) ([]*Payment, error) {
	iter := r.Iter(ctx, params)
	var out []*Payment
	for iter.Next() {
		out = append(out, iter.Value())
	}
	return out, iter.Err()
}

// Timeline returns the audit timeline for a payment.
func (r *PaymentsResource) Timeline(ctx context.Context, id string) ([]TimelineEvent, error) {
	resp, err := r.client.resty.R().SetContext(ctx).Get(apiPrefix + "/payments/" + id + "/timeline")
	if err != nil {
		return nil, fmt.Errorf("kosovopay: request failed: %w", err)
	}
	if resp.StatusCode() >= 400 {
		var envelope map[string]interface{}
		_ = json.Unmarshal(resp.Body(), &envelope)
		return nil, mapError(envelope, resp.StatusCode(), 0)
	}
	var env struct {
		Data []struct {
			Type string  `json:"type"`
			At   float64 `json:"at"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &env); err != nil {
		return nil, fmt.Errorf("kosovopay: failed to decode timeline response: %w", err)
	}
	events := make([]TimelineEvent, len(env.Data))
	for i, e := range env.Data {
		events[i] = TimelineEvent{Type: e.Type, At: int64(e.At)}
	}
	return events, nil
}

// Cancel cancels a pending payment.
func (r *PaymentsResource) Cancel(ctx context.Context, id string, reason string) (*Payment, error) {
	body := map[string]interface{}{}
	if reason != "" {
		body["reason"] = reason
	}
	var raw map[string]interface{}
	if err := r.client.post(ctx, "/payments/"+id+"/cancel", body, &raw, ""); err != nil {
		return nil, err
	}
	return paymentFromMap(raw), nil
}
