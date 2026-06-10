package kosovopay

import (
	"context"
	"encoding/json"
	"fmt"
)

// BanksResource provides access to bank operations.
type BanksResource struct {
	client *Client
}

// All returns every bank enabled for this team/key in the key's mode.
func (r *BanksResource) All(ctx context.Context) ([]Bank, error) {
	resp, err := r.client.resty.R().SetContext(ctx).Get(apiPrefix + "/banks")
	if err != nil {
		return nil, fmt.Errorf("kosovopay: request failed: %w", err)
	}
	if resp.StatusCode() >= 400 {
		var envelope map[string]interface{}
		_ = json.Unmarshal(resp.Body(), &envelope)
		return nil, mapError(envelope, resp.StatusCode(), 0)
	}
	var env struct {
		Data []bankWire `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &env); err != nil {
		return nil, fmt.Errorf("kosovopay: failed to decode banks response: %w", err)
	}
	banks := make([]Bank, len(env.Data))
	for i, b := range env.Data {
		banks[i] = b.toBank()
	}
	return banks, nil
}

// Retrieve fetches one bank by its code. Returns resource_missing if the bank
// is not enabled for this team.
func (r *BanksResource) Retrieve(ctx context.Context, code BankCode) (*Bank, error) {
	resp, err := r.client.resty.R().SetContext(ctx).Get(apiPrefix + "/banks/" + string(code))
	if err != nil {
		return nil, fmt.Errorf("kosovopay: request failed: %w", err)
	}
	if resp.StatusCode() >= 400 {
		var envelope map[string]interface{}
		_ = json.Unmarshal(resp.Body(), &envelope)
		return nil, mapError(envelope, resp.StatusCode(), 0)
	}
	var b bankWire
	if err := json.Unmarshal(resp.Body(), &b); err != nil {
		return nil, fmt.Errorf("kosovopay: failed to decode bank response: %w", err)
	}
	bank := b.toBank()
	return &bank, nil
}

// bankWire is the raw JSON shape of a Bank returned by the API.
type bankWire struct {
	Code         string   `json:"code"`
	DisplayName  string   `json:"display_name"`
	LogoURL      string   `json:"logo_url"`
	Enabled      bool     `json:"enabled"`
	Modes        []string `json:"modes"`
	Capabilities struct {
		Currencies []string `json:"currencies"`
		MinAmount  int      `json:"min_amount"`
		AmountStep int      `json:"amount_step"`
		Refunds    struct {
			Supported bool `json:"supported"`
			Partial   bool `json:"partial"`
		} `json:"refunds"`
	} `json:"capabilities"`
}

func (b *bankWire) toBank() Bank {
	modes := make([]BankMode, len(b.Modes))
	for i, m := range b.Modes {
		modes[i] = BankModeFromWire(m)
	}
	currencies := make([]CurrencyCode, len(b.Capabilities.Currencies))
	for i, c := range b.Capabilities.Currencies {
		currencies[i] = CurrencyCodeFromWire(c)
	}
	return Bank{
		Code:        BankCodeFromWire(b.Code),
		DisplayName: b.DisplayName,
		LogoURL:     b.LogoURL,
		Enabled:     b.Enabled,
		Modes:       modes,
		Capabilities: BankCapabilities{
			Currencies: currencies,
			MinAmount:  b.Capabilities.MinAmount,
			AmountStep: b.Capabilities.AmountStep,
			Refunds: RefundCapability{
				Supported: b.Capabilities.Refunds.Supported,
				Partial:   b.Capabilities.Refunds.Partial,
			},
		},
	}
}
