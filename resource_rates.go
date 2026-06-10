package kosovopay

import (
	"context"
	"encoding/json"
	"fmt"
)

// RatesResource provides access to exchange rate operations.
type RatesResource struct {
	client *Client
}

// Retrieve fetches the exchange rate between two currencies.
func (r *RatesResource) Retrieve(ctx context.Context, from, to CurrencyCode) (*Rate, error) {
	resp, err := r.client.resty.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"from": string(from),
			"to":   string(to),
		}).
		Get(apiPrefix + "/rates")
	if err != nil {
		return nil, fmt.Errorf("kosovopay: request failed: %w", err)
	}
	if resp.StatusCode() >= 400 {
		var envelope map[string]interface{}
		_ = json.Unmarshal(resp.Body(), &envelope)
		return nil, mapError(envelope, resp.StatusCode(), 0)
	}
	var wire struct {
		From     string `json:"from"`
		To       string `json:"to"`
		Rate     string `json:"rate"`
		SyncedAt string `json:"synced_at"`
		Stale    bool   `json:"stale"`
	}
	if err := json.Unmarshal(resp.Body(), &wire); err != nil {
		return nil, fmt.Errorf("kosovopay: failed to decode rate response: %w", err)
	}
	return &Rate{
		From:     CurrencyCodeFromWire(wire.From),
		To:       CurrencyCodeFromWire(wire.To),
		Rate:     wire.Rate,
		SyncedAt: wire.SyncedAt,
		Stale:    wire.Stale,
	}, nil
}
