package kosovopay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// SignatureHeader is the HTTP header carrying the webhook signature.
	SignatureHeader = "Kosovopay-Signature"
	// defaultTolerance is the default replay window in seconds (5 minutes).
	defaultTolerance = 300
)

// WebhookSignatureError is returned when signature verification fails.
type WebhookSignatureError struct {
	Message string
}

func (e *WebhookSignatureError) Error() string {
	return "kosovopay: webhook signature verification failed: " + e.Message
}

// ConstructEvent verifies the webhook signature and parses the raw payload
// into a typed Event. payload must be the raw (un-decoded) request body.
// tolerance is the replay window in seconds; pass 0 to use the default (300s).
func ConstructEvent(payload []byte, signatureHeader, secret string, tolerance int) (*Event, error) {
	if tolerance <= 0 {
		tolerance = defaultTolerance
	}
	if err := verifySignature(payload, signatureHeader, secret, time.Now().Unix(), tolerance); err != nil {
		return nil, err
	}
	var d map[string]interface{}
	if err := json.Unmarshal(payload, &d); err != nil {
		return nil, &WebhookSignatureError{Message: "payload is not valid JSON: " + err.Error()}
	}
	return eventFromMap(d), nil
}

// Verify checks only the signature of a webhook delivery without parsing the
// body. Returns nil on success, *WebhookSignatureError on failure.
// now is the current Unix timestamp; pass 0 to use time.Now().
func Verify(payload []byte, signatureHeader, secret string, now int64, tolerance int) error {
	if tolerance <= 0 {
		tolerance = defaultTolerance
	}
	if now <= 0 {
		now = time.Now().Unix()
	}
	return verifySignature(payload, signatureHeader, secret, now, tolerance)
}

// verifySignature is the internal constant-time HMAC-SHA256 check.
func verifySignature(payload []byte, signatureHeader, secret string, now int64, tolerance int) error {
	parts := parseSignatureHeader(signatureHeader)
	tsStr, ok := parts["t"]
	if !ok || tsStr == "" {
		return &WebhookSignatureError{Message: "missing timestamp in signature header"}
	}
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil || ts <= 0 {
		return &WebhookSignatureError{Message: "invalid timestamp in signature header"}
	}
	given, ok := parts["v1"]
	if !ok || given == "" {
		return &WebhookSignatureError{Message: "missing v1 signature in header"}
	}
	if abs64(now-ts) > int64(tolerance) {
		return &WebhookSignatureError{Message: fmt.Sprintf("timestamp is outside the tolerance window (skew=%ds)", abs64(now-ts))}
	}
	// Signed payload: "{t}.{raw_body}"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(ts, 10)))
	mac.Write([]byte("."))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	givenBytes, err := hex.DecodeString(given)
	if err != nil {
		return &WebhookSignatureError{Message: "v1 signature is not valid hex"}
	}
	expectedBytes, _ := hex.DecodeString(expected)
	if !hmac.Equal(expectedBytes, givenBytes) {
		return &WebhookSignatureError{Message: "signature mismatch"}
	}
	return nil
}

// parseSignatureHeader parses "t=123,v1=abc" into {"t":"123","v1":"abc"}.
func parseSignatureHeader(header string) map[string]string {
	parts := map[string]string{}
	for _, piece := range strings.Split(header, ",") {
		kv := strings.SplitN(strings.TrimSpace(piece), "=", 2)
		if len(kv) == 2 {
			parts[kv[0]] = kv[1]
		}
	}
	return parts
}

func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// eventFromMap deserialises a raw JSON map into an Event.
func eventFromMap(d map[string]interface{}) *Event {
	e := &Event{}
	e.ID, _ = d["id"].(string)
	if s, ok := d["type"].(string); ok {
		e.Type = WebhookEventTypeFromWire(s)
	}
	e.Created = toInt64(d["created"])
	e.Livemode, _ = d["livemode"].(bool)
	e.APIVersion, _ = d["api_version"].(string)
	if data, ok := d["data"].(map[string]interface{}); ok {
		e.Data = data
		if obj, ok := data["object"].(map[string]interface{}); ok {
			e.Object = obj
		}
		if prev, ok := data["previous_attributes"].(map[string]interface{}); ok {
			e.PreviousAttributes = prev
		}
	}
	return e
}
