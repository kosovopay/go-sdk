package kosovopay

import (
	"fmt"
	"math"
	"strconv"
)

// Money provides helpers for working with minor-unit integer amounts.
type Money struct{}

// Format formats a minor-unit integer amount as a decimal string using the
// given number of decimal places.
//
//	Money{}.Format(1999, 2) // "19.99"
//	Money{}.Format(1999, 0) // "1999"
func (Money) Format(amount int, decimals int) string {
	if decimals <= 0 {
		return strconv.Itoa(amount)
	}
	divisor := intPow10(decimals)
	whole := amount / divisor
	frac := amount % divisor
	if frac < 0 {
		frac = -frac
	}
	return fmt.Sprintf("%d.%0*d", whole, decimals, frac)
}

// Convert converts a minor-unit amount using a decimal rate string, rounding
// to the nearest integer.
//
//	Money{}.Convert(1000, "1.08") // 1080
func (Money) Convert(amount int, rate string) int {
	r, err := strconv.ParseFloat(rate, 64)
	if err != nil || r == 0 {
		return 0
	}
	return int(math.Round(float64(amount) * r))
}

// intPow10 returns 10^n as an integer.
func intPow10(n int) int {
	result := 1
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}

// ValidateAmountLocal performs a client-side pre-check of amount against a bank's
// live capabilities without a server round-trip. It is the exported equivalent
// of validateAmount for use in tests and by callers who already have a Bank.
func ValidateAmountLocal(bank *Bank, amount int, currency CurrencyCode) *AmountValidation {
	return validateAmount(bank, amount, currency)
}

// validateAmount performs a client-side pre-check of amount against a bank's
// live capabilities. Catches amount_below_minimum and amount_step_invalid
// before a round-trip to the server. Always reads the bank's live capabilities —
// never hardcodes a minimum or step.
func validateAmount(bank *Bank, amount int, currency CurrencyCode) *AmountValidation {
	caps := bank.Capabilities

	if len(caps.Currencies) > 0 {
		found := false
		for _, c := range caps.Currencies {
			if c == currency {
				found = true
				break
			}
		}
		if !found {
			return &AmountValidation{
				Valid:   false,
				Code:    "currency_not_supported",
				Message: fmt.Sprintf("%s does not support %s", bank.DisplayName, string(currency)),
			}
		}
	}

	if amount < caps.MinAmount {
		return &AmountValidation{
			Valid:   false,
			Code:    "amount_below_minimum",
			Message: fmt.Sprintf("amount is below the %s minimum of %d", bank.DisplayName, caps.MinAmount),
		}
	}

	step := caps.AmountStep
	if step < 1 {
		step = 1
	}
	if amount%step != 0 {
		lower := (amount / step) * step
		upper := lower + step
		return &AmountValidation{
			Valid:        false,
			Code:         "amount_step_invalid",
			Message:      fmt.Sprintf("%s requires amounts in steps of %d", bank.DisplayName, step),
			NearestValid: [2]int{lower, upper},
		}
	}

	return &AmountValidation{Valid: true}
}
