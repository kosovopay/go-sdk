package kosovopay

import (
	"context"
	"fmt"
)

// CreateRefundParams are the parameters for creating a refund.
type CreateRefundParams struct {
	// Payment is the ID of the payment to refund. Required.
	Payment string `json:"payment"`
	// Amount is the refund amount in minor units. Omit for a full refund.
	Amount int `json:"amount,omitempty"`
	// Reason is the reason for the refund.
	Reason RefundReason `json:"reason,omitempty"`
}

// Validate checks CreateRefundParams for obvious client-side errors.
func (p *CreateRefundParams) Validate() error {
	if p.Payment == "" {
		return fmt.Errorf("kosovopay: payment id is required")
	}
	if p.Amount < 0 {
		return fmt.Errorf("kosovopay: amount, when given, must be a positive integer in minor units")
	}
	return nil
}

// ListRefundsParams are the query parameters for listing refunds.
type ListRefundsParams struct {
	Payment       string
	Limit         int
	StartingAfter string
	EndingBefore  string
}

func (p *ListRefundsParams) toQuery() map[string]string {
	q := map[string]string{}
	if p == nil {
		return q
	}
	if p.Payment != "" {
		q["payment"] = p.Payment
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
	return q
}

// RefundsResource provides access to refund operations.
type RefundsResource struct {
	client *Client
}

// Create creates a new refund. An Idempotency-Key is auto-generated; pass a
// non-empty idempotencyKey to override it.
func (r *RefundsResource) Create(ctx context.Context, params CreateRefundParams, idempotencyKey string) (*Refund, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	if err := r.client.post(ctx, "/refunds", params, &raw, idempotencyKey); err != nil {
		return nil, err
	}
	return refundFromMap(raw), nil
}

// Retrieve fetches a refund by ID.
func (r *RefundsResource) Retrieve(ctx context.Context, id string) (*Refund, error) {
	var raw map[string]interface{}
	if err := r.client.get(ctx, "/refunds/"+id, nil, &raw); err != nil {
		return nil, err
	}
	return refundFromMap(raw), nil
}

// Iter returns a cursor-paginated iterator over refunds matching the params.
func (r *RefundsResource) Iter(ctx context.Context, params *ListRefundsParams) *PageIter[*Refund] {
	return newPageIter(ctx, r.client, "/refunds", params.toQuery(),
		func(row map[string]interface{}) *Refund { return refundFromMap(row) })
}

// All collects every matching refund across all pages. Use Iter for large result sets.
func (r *RefundsResource) All(ctx context.Context, params *ListRefundsParams) ([]*Refund, error) {
	iter := r.Iter(ctx, params)
	var out []*Refund
	for iter.Next() {
		out = append(out, iter.Value())
	}
	return out, iter.Err()
}
