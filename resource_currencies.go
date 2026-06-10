package kosovopay

import (
	"context"
	"encoding/json"
	"fmt"
)

// CurrenciesResource provides access to currency operations.
type CurrenciesResource struct {
	client *Client
}

// All returns every currency supported by the API.
func (r *CurrenciesResource) All(ctx context.Context) ([]Currency, error) {
	resp, err := r.client.resty.R().SetContext(ctx).Get(apiPrefix + "/currencies")
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
			Code      string `json:"code"`
			Name      string `json:"name"`
			Symbol    string `json:"symbol"`
			Decimals  int    `json:"decimals"`
			IsDefault bool   `json:"is_default"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &env); err != nil {
		return nil, fmt.Errorf("kosovopay: failed to decode currencies response: %w", err)
	}
	out := make([]Currency, len(env.Data))
	for i, c := range env.Data {
		out[i] = Currency{
			Code:      CurrencyCodeFromWire(c.Code),
			Name:      c.Name,
			Symbol:    c.Symbol,
			Decimals:  c.Decimals,
			IsDefault: c.IsDefault,
		}
	}
	return out, nil
}
