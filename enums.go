// Package kosovopay provides a Go client for the KosovoPay payment API.
//
// Create a client with your secret key and use the resource accessors:
//
//	kp := kosovopay.New("sk_test_...")
//	payment, err := kp.Payments.Create(ctx, kosovopay.CreatePaymentParams{...})
package kosovopay

import "strings"

// CheckoutMode is how the customer reaches the bank.
type CheckoutMode string

const (
	// CheckoutModeHosted renders a KosovoPay-branded checkout page where the
	// customer picks the bank.
	CheckoutModeHosted CheckoutMode = "hosted"
	// CheckoutModeDirect skips the hosted page and redirects the customer
	// straight to a specific bank. Requires BankCode to be set.
	CheckoutModeDirect CheckoutMode = "direct"
)

// BankMode indicates whether the key / bank is in test or live mode.
type BankMode string

const (
	BankModeTest BankMode = "test"
	BankModeLive BankMode = "live"
)

// BankModeFromWire parses a wire value, defaulting to BankModeTest for
// unrecognised values (forward-compatible).
func BankModeFromWire(v string) BankMode {
	switch v {
	case "live":
		return BankModeLive
	default:
		return BankModeTest
	}
}

// BankCode identifies a bank integration.
type BankCode string

const (
	BankCodeProcredit BankCode = "procredit"
	BankCodeProcard   BankCode = "procard"
	BankCodeOnefor    BankCode = "onefor"
	// BankCodeUnknown is the forward-compatible fallback for an unrecognised wire value.
	BankCodeUnknown BankCode = "unknown"
)

// BankCodeFromWire parses a wire value, returning BankCodeUnknown for
// unrecognised codes (forward-compatible).
func BankCodeFromWire(v string) BankCode {
	switch v {
	case "procredit":
		return BankCodeProcredit
	case "procard":
		return BankCodeProcard
	case "onefor":
		return BankCodeOnefor
	default:
		return BankCodeUnknown
	}
}

// CurrencyCode is a full ISO 4217 currency code.
type CurrencyCode string

const (
	CurrencyAED CurrencyCode = "AED"
	CurrencyAFN CurrencyCode = "AFN"
	CurrencyALL CurrencyCode = "ALL"
	CurrencyAMD CurrencyCode = "AMD"
	CurrencyANG CurrencyCode = "ANG"
	CurrencyAOA CurrencyCode = "AOA"
	CurrencyARS CurrencyCode = "ARS"
	CurrencyAUD CurrencyCode = "AUD"
	CurrencyAWG CurrencyCode = "AWG"
	CurrencyAZN CurrencyCode = "AZN"
	CurrencyBAM CurrencyCode = "BAM"
	CurrencyBBD CurrencyCode = "BBD"
	CurrencyBDT CurrencyCode = "BDT"
	CurrencyBGN CurrencyCode = "BGN"
	CurrencyBHD CurrencyCode = "BHD"
	CurrencyBIF CurrencyCode = "BIF"
	CurrencyBMD CurrencyCode = "BMD"
	CurrencyBND CurrencyCode = "BND"
	CurrencyBOB CurrencyCode = "BOB"
	CurrencyBRL CurrencyCode = "BRL"
	CurrencyBSD CurrencyCode = "BSD"
	CurrencyBTN CurrencyCode = "BTN"
	CurrencyBWP CurrencyCode = "BWP"
	CurrencyBYN CurrencyCode = "BYN"
	CurrencyBZD CurrencyCode = "BZD"
	CurrencyCAD CurrencyCode = "CAD"
	CurrencyCDF CurrencyCode = "CDF"
	CurrencyCHF CurrencyCode = "CHF"
	CurrencyCLP CurrencyCode = "CLP"
	CurrencyCNY CurrencyCode = "CNY"
	CurrencyCOP CurrencyCode = "COP"
	CurrencyCRC CurrencyCode = "CRC"
	CurrencyCUP CurrencyCode = "CUP"
	CurrencyCVE CurrencyCode = "CVE"
	CurrencyCZK CurrencyCode = "CZK"
	CurrencyDJF CurrencyCode = "DJF"
	CurrencyDKK CurrencyCode = "DKK"
	CurrencyDOP CurrencyCode = "DOP"
	CurrencyDZD CurrencyCode = "DZD"
	CurrencyEGP CurrencyCode = "EGP"
	CurrencyERN CurrencyCode = "ERN"
	CurrencyETB CurrencyCode = "ETB"
	CurrencyEUR CurrencyCode = "EUR"
	CurrencyFJD CurrencyCode = "FJD"
	CurrencyFKP CurrencyCode = "FKP"
	CurrencyGBP CurrencyCode = "GBP"
	CurrencyGEL CurrencyCode = "GEL"
	CurrencyGHS CurrencyCode = "GHS"
	CurrencyGIP CurrencyCode = "GIP"
	CurrencyGMD CurrencyCode = "GMD"
	CurrencyGNF CurrencyCode = "GNF"
	CurrencyGTQ CurrencyCode = "GTQ"
	CurrencyGYD CurrencyCode = "GYD"
	CurrencyHKD CurrencyCode = "HKD"
	CurrencyHNL CurrencyCode = "HNL"
	CurrencyHTG CurrencyCode = "HTG"
	CurrencyHUF CurrencyCode = "HUF"
	CurrencyIDR CurrencyCode = "IDR"
	CurrencyILS CurrencyCode = "ILS"
	CurrencyINR CurrencyCode = "INR"
	CurrencyIQD CurrencyCode = "IQD"
	CurrencyIRR CurrencyCode = "IRR"
	CurrencyISK CurrencyCode = "ISK"
	CurrencyJMD CurrencyCode = "JMD"
	CurrencyJOD CurrencyCode = "JOD"
	CurrencyJPY CurrencyCode = "JPY"
	CurrencyKES CurrencyCode = "KES"
	CurrencyKGS CurrencyCode = "KGS"
	CurrencyKHR CurrencyCode = "KHR"
	CurrencyKMF CurrencyCode = "KMF"
	CurrencyKPW CurrencyCode = "KPW"
	CurrencyKRW CurrencyCode = "KRW"
	CurrencyKWD CurrencyCode = "KWD"
	CurrencyKYD CurrencyCode = "KYD"
	CurrencyKZT CurrencyCode = "KZT"
	CurrencyLAK CurrencyCode = "LAK"
	CurrencyLBP CurrencyCode = "LBP"
	CurrencyLKR CurrencyCode = "LKR"
	CurrencyLRD CurrencyCode = "LRD"
	CurrencyLSL CurrencyCode = "LSL"
	CurrencyLYD CurrencyCode = "LYD"
	CurrencyMAD CurrencyCode = "MAD"
	CurrencyMDL CurrencyCode = "MDL"
	CurrencyMGA CurrencyCode = "MGA"
	CurrencyMKD CurrencyCode = "MKD"
	CurrencyMMK CurrencyCode = "MMK"
	CurrencyMNT CurrencyCode = "MNT"
	CurrencyMOP CurrencyCode = "MOP"
	CurrencyMRU CurrencyCode = "MRU"
	CurrencyMUR CurrencyCode = "MUR"
	CurrencyMVR CurrencyCode = "MVR"
	CurrencyMWK CurrencyCode = "MWK"
	CurrencyMXN CurrencyCode = "MXN"
	CurrencyMYR CurrencyCode = "MYR"
	CurrencyMZN CurrencyCode = "MZN"
	CurrencyNAD CurrencyCode = "NAD"
	CurrencyNGN CurrencyCode = "NGN"
	CurrencyNIO CurrencyCode = "NIO"
	CurrencyNOK CurrencyCode = "NOK"
	CurrencyNPR CurrencyCode = "NPR"
	CurrencyNZD CurrencyCode = "NZD"
	CurrencyOMR CurrencyCode = "OMR"
	CurrencyPAB CurrencyCode = "PAB"
	CurrencyPEN CurrencyCode = "PEN"
	CurrencyPGK CurrencyCode = "PGK"
	CurrencyPHP CurrencyCode = "PHP"
	CurrencyPKR CurrencyCode = "PKR"
	CurrencyPLN CurrencyCode = "PLN"
	CurrencyPYG CurrencyCode = "PYG"
	CurrencyQAR CurrencyCode = "QAR"
	CurrencyRON CurrencyCode = "RON"
	CurrencyRSD CurrencyCode = "RSD"
	CurrencyRUB CurrencyCode = "RUB"
	CurrencyRWF CurrencyCode = "RWF"
	CurrencySAR CurrencyCode = "SAR"
	CurrencySBD CurrencyCode = "SBD"
	CurrencySCR CurrencyCode = "SCR"
	CurrencySDG CurrencyCode = "SDG"
	CurrencySEK CurrencyCode = "SEK"
	CurrencySGD CurrencyCode = "SGD"
	CurrencySHP CurrencyCode = "SHP"
	CurrencySLE CurrencyCode = "SLE"
	CurrencySOS CurrencyCode = "SOS"
	CurrencySRD CurrencyCode = "SRD"
	CurrencySSP CurrencyCode = "SSP"
	CurrencySTN CurrencyCode = "STN"
	CurrencySVC CurrencyCode = "SVC"
	CurrencySYP CurrencyCode = "SYP"
	CurrencySZL CurrencyCode = "SZL"
	CurrencyTHB CurrencyCode = "THB"
	CurrencyTJS CurrencyCode = "TJS"
	CurrencyTMT CurrencyCode = "TMT"
	CurrencyTND CurrencyCode = "TND"
	CurrencyTOP CurrencyCode = "TOP"
	CurrencyTRY CurrencyCode = "TRY"
	CurrencyTTD CurrencyCode = "TTD"
	CurrencyTWD CurrencyCode = "TWD"
	CurrencyTZS CurrencyCode = "TZS"
	CurrencyUAH CurrencyCode = "UAH"
	CurrencyUGX CurrencyCode = "UGX"
	CurrencyUSD CurrencyCode = "USD"
	CurrencyUYU CurrencyCode = "UYU"
	CurrencyUZS CurrencyCode = "UZS"
	CurrencyVES CurrencyCode = "VES"
	CurrencyVND CurrencyCode = "VND"
	CurrencyVUV CurrencyCode = "VUV"
	CurrencyWST CurrencyCode = "WST"
	CurrencyXAF CurrencyCode = "XAF"
	CurrencyXCD CurrencyCode = "XCD"
	CurrencyXOF CurrencyCode = "XOF"
	CurrencyXPF CurrencyCode = "XPF"
	CurrencyYER CurrencyCode = "YER"
	CurrencyZAR CurrencyCode = "ZAR"
	CurrencyZMW CurrencyCode = "ZMW"
	CurrencyZWG CurrencyCode = "ZWG"
	// CurrencyUnknown is the forward-compatible fallback for an unrecognised wire value.
	CurrencyUnknown CurrencyCode = "unknown"
)

// CurrencyCodeFromWire parses a wire value (case-insensitive), returning
// CurrencyUnknown for unrecognised codes (forward-compatible).
func CurrencyCodeFromWire(v string) CurrencyCode {
	// Upper-case the input for comparison.
	upper := strings.ToUpper(v)
	switch CurrencyCode(upper) {
	case CurrencyAED, CurrencyAFN, CurrencyALL, CurrencyAMD, CurrencyANG,
		CurrencyAOA, CurrencyARS, CurrencyAUD, CurrencyAWG, CurrencyAZN,
		CurrencyBAM, CurrencyBBD, CurrencyBDT, CurrencyBGN, CurrencyBHD,
		CurrencyBIF, CurrencyBMD, CurrencyBND, CurrencyBOB, CurrencyBRL,
		CurrencyBSD, CurrencyBTN, CurrencyBWP, CurrencyBYN, CurrencyBZD,
		CurrencyCAD, CurrencyCDF, CurrencyCHF, CurrencyCLP, CurrencyCNY,
		CurrencyCOP, CurrencyCRC, CurrencyCUP, CurrencyCVE, CurrencyCZK,
		CurrencyDJF, CurrencyDKK, CurrencyDOP, CurrencyDZD, CurrencyEGP,
		CurrencyERN, CurrencyETB, CurrencyEUR, CurrencyFJD, CurrencyFKP,
		CurrencyGBP, CurrencyGEL, CurrencyGHS, CurrencyGIP, CurrencyGMD,
		CurrencyGNF, CurrencyGTQ, CurrencyGYD, CurrencyHKD, CurrencyHNL,
		CurrencyHTG, CurrencyHUF, CurrencyIDR, CurrencyILS, CurrencyINR,
		CurrencyIQD, CurrencyIRR, CurrencyISK, CurrencyJMD, CurrencyJOD,
		CurrencyJPY, CurrencyKES, CurrencyKGS, CurrencyKHR, CurrencyKMF,
		CurrencyKPW, CurrencyKRW, CurrencyKWD, CurrencyKYD, CurrencyKZT,
		CurrencyLAK, CurrencyLBP, CurrencyLKR, CurrencyLRD, CurrencyLSL,
		CurrencyLYD, CurrencyMAD, CurrencyMDL, CurrencyMGA, CurrencyMKD,
		CurrencyMMK, CurrencyMNT, CurrencyMOP, CurrencyMRU, CurrencyMUR,
		CurrencyMVR, CurrencyMWK, CurrencyMXN, CurrencyMYR, CurrencyMZN,
		CurrencyNAD, CurrencyNGN, CurrencyNIO, CurrencyNOK, CurrencyNPR,
		CurrencyNZD, CurrencyOMR, CurrencyPAB, CurrencyPEN, CurrencyPGK,
		CurrencyPHP, CurrencyPKR, CurrencyPLN, CurrencyPYG, CurrencyQAR,
		CurrencyRON, CurrencyRSD, CurrencyRUB, CurrencyRWF, CurrencySAR,
		CurrencySBD, CurrencySCR, CurrencySDG, CurrencySEK, CurrencySGD,
		CurrencySHP, CurrencySLE, CurrencySOS, CurrencySRD, CurrencySSP,
		CurrencySTN, CurrencySVC, CurrencySYP, CurrencySZL, CurrencyTHB,
		CurrencyTJS, CurrencyTMT, CurrencyTND, CurrencyTOP, CurrencyTRY,
		CurrencyTTD, CurrencyTWD, CurrencyTZS, CurrencyUAH, CurrencyUGX,
		CurrencyUSD, CurrencyUYU, CurrencyUZS, CurrencyVES, CurrencyVND,
		CurrencyVUV, CurrencyWST, CurrencyXAF, CurrencyXCD, CurrencyXOF,
		CurrencyXPF, CurrencyYER, CurrencyZAR, CurrencyZMW, CurrencyZWG:
		return CurrencyCode(upper)
	}
	return CurrencyUnknown
}

// PaymentStatus is the lifecycle state of a Payment.
type PaymentStatus string

const (
	PaymentStatusPending           PaymentStatus = "pending"
	PaymentStatusAuthorized        PaymentStatus = "authorized"
	PaymentStatusCaptured          PaymentStatus = "captured"
	PaymentStatusPartiallyRefunded PaymentStatus = "partially_refunded"
	PaymentStatusRefunded          PaymentStatus = "refunded"
	PaymentStatusFailed            PaymentStatus = "failed"
	PaymentStatusCanceled          PaymentStatus = "canceled"
	// PaymentStatusUnknown is the forward-compatible fallback.
	PaymentStatusUnknown PaymentStatus = "unknown"
)

// PaymentStatusFromWire parses a wire value, returning PaymentStatusUnknown for
// unrecognised values (forward-compatible).
func PaymentStatusFromWire(v string) PaymentStatus {
	switch v {
	case "pending":
		return PaymentStatusPending
	case "authorized":
		return PaymentStatusAuthorized
	case "captured":
		return PaymentStatusCaptured
	case "partially_refunded":
		return PaymentStatusPartiallyRefunded
	case "refunded":
		return PaymentStatusRefunded
	case "failed":
		return PaymentStatusFailed
	case "canceled":
		return PaymentStatusCanceled
	default:
		return PaymentStatusUnknown
	}
}

// IsTerminal reports whether the payment has reached a final state.
func (s PaymentStatus) IsTerminal() bool {
	return s == PaymentStatusCaptured || s == PaymentStatusRefunded ||
		s == PaymentStatusFailed || s == PaymentStatusCanceled
}

// RefundStatus is the lifecycle state of a Refund.
type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusSucceeded RefundStatus = "succeeded"
	RefundStatusFailed    RefundStatus = "failed"
	// RefundStatusUnknown is the forward-compatible fallback.
	RefundStatusUnknown RefundStatus = "unknown"
)

// RefundStatusFromWire parses a wire value, returning RefundStatusUnknown for
// unrecognised values (forward-compatible).
func RefundStatusFromWire(v string) RefundStatus {
	switch v {
	case "pending":
		return RefundStatusPending
	case "succeeded":
		return RefundStatusSucceeded
	case "failed":
		return RefundStatusFailed
	default:
		return RefundStatusUnknown
	}
}

// RefundReason is the reason a refund was issued.
type RefundReason string

const (
	RefundReasonRequestedByCustomer RefundReason = "requested_by_customer"
	RefundReasonDuplicate           RefundReason = "duplicate"
	RefundReasonFraudulent          RefundReason = "fraudulent"
	RefundReasonOther               RefundReason = "other"
)

// WebhookEventType identifies the category of a webhook event.
type WebhookEventType string

const (
	WebhookEventTypePaymentCreated  WebhookEventType = "payment.created"
	WebhookEventTypePaymentCaptured WebhookEventType = "payment.captured"
	WebhookEventTypePaymentFailed   WebhookEventType = "payment.failed"
	WebhookEventTypePaymentCanceled WebhookEventType = "payment.canceled"
	WebhookEventTypePaymentExpired  WebhookEventType = "payment.expired"
	WebhookEventTypeRefundSucceeded WebhookEventType = "refund.succeeded"
	WebhookEventTypeRefundFailed    WebhookEventType = "refund.failed"
	// WebhookEventTypeUnknown is the forward-compatible fallback.
	WebhookEventTypeUnknown WebhookEventType = "unknown"
)

// WebhookEventTypeFromWire parses a wire value, returning
// WebhookEventTypeUnknown for unrecognised values (forward-compatible).
func WebhookEventTypeFromWire(v string) WebhookEventType {
	switch v {
	case "payment.created":
		return WebhookEventTypePaymentCreated
	case "payment.captured":
		return WebhookEventTypePaymentCaptured
	case "payment.failed":
		return WebhookEventTypePaymentFailed
	case "payment.canceled":
		return WebhookEventTypePaymentCanceled
	case "payment.expired":
		return WebhookEventTypePaymentExpired
	case "refund.succeeded":
		return WebhookEventTypeRefundSucceeded
	case "refund.failed":
		return WebhookEventTypeRefundFailed
	default:
		return WebhookEventTypeUnknown
	}
}

// IsRefund reports whether the event relates to a refund.
func (t WebhookEventType) IsRefund() bool {
	return t == WebhookEventTypeRefundSucceeded || t == WebhookEventTypeRefundFailed
}
