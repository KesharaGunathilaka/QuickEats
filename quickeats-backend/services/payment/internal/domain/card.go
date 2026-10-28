package domain

import (
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

// DemoCardBalance is the credit every newly added card starts with (LKR 5,000.00 in
// cents). This is a demonstration system: no real card network is charged, so each
// card carries its own spendable balance instead. It is never sent to clients.
const DemoCardBalance int64 = 500000

// Card is a customer's saved payment card. Only display details are kept; the full
// number and the CVV are checked when the card is added and then discarded.
type Card struct {
	ID         string    `json:"id"`
	UserID     string    `json:"-"`
	Brand      string    `json:"brand"`
	Last4      string    `json:"last4"`
	HolderName string    `json:"holderName"`
	ExpMonth   int       `json:"expMonth"`
	ExpYear    int       `json:"expYear"`
	Balance    int64     `json:"-"` // hidden: the API only ever says whether it covers an amount
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// NewCard validates the details a customer entered and returns the card to save,
// loaded with DemoCardBalance. now decides whether the card has expired.
func NewCard(userID, number, holderName string, expMonth, expYear int, cvv string, now time.Time) (*Card, error) {
	digits := stripSpaces(number)
	if len(digits) < 13 || len(digits) > 19 || !allDigits(digits) {
		return nil, &CardValidationError{Field: "number", Message: "enter a valid card number"}
	}
	if !luhnValid(digits) {
		return nil, &CardValidationError{Field: "number", Message: "this card number is not valid"}
	}

	holderName = strings.TrimSpace(holderName)
	if len(holderName) < 2 || len(holderName) > 100 {
		return nil, &CardValidationError{Field: "holderName", Message: "enter the name on the card"}
	}

	if expYear < 100 {
		expYear += 2000 // accept YY as well as YYYY
	}
	if expMonth < 1 || expMonth > 12 {
		return nil, &CardValidationError{Field: "expiry", Message: "enter a valid expiry month"}
	}
	// A card is valid through the last day of its expiry month.
	if expYear < now.Year() || (expYear == now.Year() && expMonth < int(now.Month())) {
		return nil, &CardValidationError{Field: "expiry", Message: "this card has expired"}
	}
	if expYear > now.Year()+20 {
		return nil, &CardValidationError{Field: "expiry", Message: "enter a valid expiry year"}
	}

	cvv = strings.TrimSpace(cvv)
	if (len(cvv) != 3 && len(cvv) != 4) || !allDigits(cvv) {
		return nil, &CardValidationError{Field: "cvv", Message: "enter the 3 or 4 digit security code"}
	}

	t := now.UTC()
	return &Card{
		ID:         uuid.New().String(),
		UserID:     userID,
		Brand:      CardBrand(digits),
		Last4:      digits[len(digits)-4:],
		HolderName: holderName,
		ExpMonth:   expMonth,
		ExpYear:    expYear,
		Balance:    DemoCardBalance,
		CreatedAt:  t,
		UpdatedAt:  t,
	}, nil
}

// CardBrand names the card network from the number's leading digits.
func CardBrand(digits string) string {
	switch {
	case strings.HasPrefix(digits, "4"):
		return "Visa"
	case strings.HasPrefix(digits, "34"), strings.HasPrefix(digits, "37"):
		return "Amex"
	case len(digits) >= 2 && digits[:2] >= "51" && digits[:2] <= "55":
		return "Mastercard"
	case len(digits) >= 4 && digits[:4] >= "2221" && digits[:4] <= "2720":
		return "Mastercard"
	default:
		return "Card"
	}
}

// CardValidationError says which field of a new card is wrong.
type CardValidationError struct {
	Field   string
	Message string
}

func (e *CardValidationError) Error() string { return e.Message }

// luhnValid runs the standard card-number checksum.
func luhnValid(digits string) bool {
	sum := 0
	double := false
	for i := len(digits) - 1; i >= 0; i-- {
		d := int(digits[i] - '0')
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}

func stripSpaces(s string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '-' {
			return -1
		}
		return r
	}, s)
}

func allDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) || r > unicode.MaxASCII {
			return false
		}
	}
	return s != ""
}
