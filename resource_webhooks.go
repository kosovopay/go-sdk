package kosovopay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// CreateWebhookEndpointParams are the parameters for creating a webhook endpoint.
type CreateWebhookEndpointParams struct {
	// URL is the HTTPS destination for event deliveries.
	URL string `json:"url"`
	// EnabledEvents is the list of event types to subscribe to. Must be non-empty.
	EnabledEvents []WebhookEventType `json:"enabled_events"`
	// Description is an optional human-readable label.
	Description string `json:"description,omitempty"`
}

// Validate checks CreateWebhookEndpointParams for obvious client-side errors.
func (p *CreateWebhookEndpointParams) Validate() error {
	if len(p.EnabledEvents) == 0 {
		return fmt.Errorf("kosovopay: enabled_events must contain at least one event type")
	}
	u, err := url.Parse(p.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("kosovopay: url must be an http or https URL")
	}
	return nil
}

// WebhookEndpointsResource provides access to webhook endpoint operations.
type WebhookEndpointsResource struct {
	client *Client
}

// Create registers a new webhook endpoint.
func (r *WebhookEndpointsResource) Create(ctx context.Context, params CreateWebhookEndpointParams) (*WebhookEndpoint, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	if err := r.client.post(ctx, "/webhook-endpoints", params, &raw, ""); err != nil {
		return nil, err
	}
	return webhookEndpointFromMap(raw), nil
}

// All returns all registered webhook endpoints.
func (r *WebhookEndpointsResource) All(ctx context.Context) ([]WebhookEndpoint, error) {
	resp, err := r.client.resty.R().SetContext(ctx).Get(apiPrefix + "/webhook-endpoints")
	if err != nil {
		return nil, fmt.Errorf("kosovopay: request failed: %w", err)
	}
	if resp.StatusCode() >= 400 {
		var envelope map[string]interface{}
		_ = json.Unmarshal(resp.Body(), &envelope)
		return nil, mapError(envelope, resp.StatusCode(), 0)
	}
	var env struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &env); err != nil {
		return nil, fmt.Errorf("kosovopay: failed to decode webhook endpoints response: %w", err)
	}
	out := make([]WebhookEndpoint, len(env.Data))
	for i, d := range env.Data {
		out[i] = *webhookEndpointFromMap(d)
	}
	return out, nil
}

// Delete deletes a webhook endpoint by ID.
func (r *WebhookEndpointsResource) Delete(ctx context.Context, id string) (*DeletedResource, error) {
	var result DeletedResource
	if err := r.client.del(ctx, "/webhook-endpoints/"+id, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RotateSecret rotates the signing secret for a webhook endpoint. The response
// contains the new secret (shown once).
func (r *WebhookEndpointsResource) RotateSecret(ctx context.Context, id string) (*WebhookEndpoint, error) {
	var raw map[string]interface{}
	if err := r.client.post(ctx, "/webhook-endpoints/"+id+"/rotate-secret", map[string]interface{}{}, &raw, ""); err != nil {
		return nil, err
	}
	return webhookEndpointFromMap(raw), nil
}

// webhookEndpointFromMap deserialises a raw JSON map into a WebhookEndpoint.
func webhookEndpointFromMap(d map[string]interface{}) *WebhookEndpoint {
	ep := &WebhookEndpoint{}
	ep.ID, _ = d["id"].(string)
	ep.URL, _ = d["url"].(string)
	ep.Description, _ = d["description"].(string)
	ep.Status, _ = d["status"].(string)
	if ep.Status == "" {
		ep.Status = "enabled"
	}
	if s, ok := d["mode"].(string); ok {
		ep.Mode = BankModeFromWire(s)
	}
	ep.Created = toInt64(d["created"])
	ep.Secret, _ = d["secret"].(string)
	if events, ok := d["enabled_events"].([]interface{}); ok {
		for _, e := range events {
			if s, ok := e.(string); ok {
				ep.EnabledEvents = append(ep.EnabledEvents, WebhookEventTypeFromWire(s))
			}
		}
	}
	return ep
}
