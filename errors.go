package kosovopay

import "fmt"

// KosovoPayError is the base error type returned by all SDK operations. It carries
// the full server error envelope so callers can branch on a stable machine code,
// surface the doc URL, and quote the request ID in support tickets.
type KosovoPayError struct {
	// Message is the human-readable error message from the server.
	Message string
	// Code is the machine-stable error code (e.g. "amount_below_minimum").
	Code string
	// Type is the error family (e.g. "payment_error").
	Type string
	// Param is the request parameter that caused the error, if applicable.
	Param string
	// RequestID is the server-assigned request identifier for support.
	RequestID string
	// DocURL is a link to the documentation for this specific error code.
	DocURL string
	// StatusCode is the HTTP status code of the response.
	StatusCode int
}

func (e *KosovoPayError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("kosovopay: %s (code=%s, status=%d, request_id=%s)", e.Message, e.Code, e.StatusCode, e.RequestID)
	}
	return fmt.Sprintf("kosovopay: %s (status=%d, request_id=%s)", e.Message, e.StatusCode, e.RequestID)
}

// AuthenticationError is returned when the API key is missing or invalid.
// HTTP 401, type "authentication_error".
type AuthenticationError struct{ *KosovoPayError }

// PermissionError is returned when the key lacks permission for this operation.
// type "permission_error".
type PermissionError struct{ *KosovoPayError }

// ValidationError is returned for malformed or invalid requests.
// HTTP 422/404, type "validation_error".
type ValidationError struct{ *KosovoPayError }

// IdempotencyError is returned when an idempotency key conflict is detected.
// HTTP 409, type "idempotency_error".
type IdempotencyError struct{ *KosovoPayError }

// RateLimitError is returned when the request is rate-limited (HTTP 429).
// RetryAfter holds the number of seconds to wait before retrying, if the
// server provided a Retry-After header.
type RateLimitError struct {
	*KosovoPayError
	// RetryAfter is seconds to wait before retrying (0 if not provided).
	RetryAfter int
}

// PaymentError is the base error for payment-specific failures.
// type "payment_error".
type PaymentError struct{ *KosovoPayError }

// AmountBelowMinimumError is returned when the amount is below the bank minimum.
type AmountBelowMinimumError struct{ *PaymentError }

// AmountStepInvalidError is returned when the amount is not a valid step.
type AmountStepInvalidError struct{ *PaymentError }

// BankNotEnabledError is returned when no enabled bank can process the request.
type BankNotEnabledError struct{ *PaymentError }

// BankUnreachableError is returned when the bank could not be reached (HTTP 502).
type BankUnreachableError struct{ *PaymentError }

// PaymentNotCancelableError is returned when the payment cannot be cancelled.
type PaymentNotCancelableError struct{ *PaymentError }

// PaymentNotRefundableError is returned when the payment cannot be refunded.
type PaymentNotRefundableError struct{ *PaymentError }

// RefundExceedsRemainingError is returned when the refund exceeds the remaining amount.
type RefundExceedsRemainingError struct{ *PaymentError }

// PartialRefundUnsupportedError is returned when the bank only supports full refunds.
type PartialRefundUnsupportedError struct{ *PaymentError }

// APIError is the generic fallback for server-side errors (HTTP 5xx, type "api_error").
type APIError struct{ *KosovoPayError }

// mapError converts a decoded server error envelope into the most specific
// typed error. Resolution order: exact code → rate-limit (429) → type family →
// APIError. An unrecognised code never panics.
func mapError(body map[string]interface{}, statusCode int, retryAfter int) error {
	errMap, _ := body["error"].(map[string]interface{})

	msg, _ := errMap["message"].(string)
	if msg == "" {
		msg = "the request failed"
	}
	code, _ := errMap["code"].(string)
	errType, _ := errMap["type"].(string)
	param, _ := errMap["param"].(string)
	requestID, _ := errMap["request_id"].(string)
	docURL, _ := errMap["doc_url"].(string)

	base := &KosovoPayError{
		Message:    msg,
		Code:       code,
		Type:       errType,
		Param:      param,
		RequestID:  requestID,
		DocURL:     docURL,
		StatusCode: statusCode,
	}

	// Rate-limit is identified by type or status before the code dispatch.
	if errType == "rate_limit_error" || statusCode == 429 {
		return &RateLimitError{KosovoPayError: base, RetryAfter: retryAfter}
	}

	pe := &PaymentError{base}

	// Dispatch on exact error code first.
	switch code {
	case "amount_below_minimum":
		return &AmountBelowMinimumError{pe}
	case "amount_step_invalid":
		return &AmountStepInvalidError{pe}
	case "bank_not_enabled":
		return &BankNotEnabledError{pe}
	case "bank_unreachable":
		return &BankUnreachableError{pe}
	case "payment_not_cancelable":
		return &PaymentNotCancelableError{pe}
	case "payment_not_refundable":
		return &PaymentNotRefundableError{pe}
	case "refund_exceeds_remaining":
		return &RefundExceedsRemainingError{pe}
	case "partial_refund_unsupported":
		return &PartialRefundUnsupportedError{pe}
	}

	// Fall through to type family.
	switch errType {
	case "authentication_error":
		return &AuthenticationError{base}
	case "permission_error":
		return &PermissionError{base}
	case "validation_error":
		return &ValidationError{base}
	case "idempotency_error":
		return &IdempotencyError{base}
	case "payment_error":
		return &PaymentError{base}
	case "api_error":
		return &APIError{base}
	}

	// Generic fallback — unknown code never crashes.
	return &APIError{base}
}
