package domain

import (
	"errors"
	"testing"
	"time"
)

var now = time.Date(2026, time.October, 28, 12, 0, 0, 0, time.UTC)

func TestNewCardAcceptsValidCard(t *testing.T) {
	card, err := NewCard("user-1", "4242 4242 4242 4242", "  Demo Customer ", 12, 30, "123", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.Brand != "Visa" || card.Last4 != "4242" {
		t.Errorf("brand/last4 = %s/%s, want Visa/4242", card.Brand, card.Last4)
	}
	if card.HolderName != "Demo Customer" {
		t.Errorf("holder = %q, want trimmed name", card.HolderName)
	}
	if card.ExpYear != 2030 {
		t.Errorf("expYear = %d, want 2030 (YY expanded)", card.ExpYear)
	}
	if card.Balance != DemoCardBalance {
		t.Errorf("balance = %d, want the demo balance %d", card.Balance, DemoCardBalance)
	}
}

func TestNewCardRejectsBadInput(t *testing.T) {
	tests := []struct {
		name, number, holder, cvv string
		month, year               int
		field                     string
	}{
		{"fails checksum", "4242 4242 4242 4241", "Demo", "123", 12, 2030, "number"},
		{"too short", "4242 4242", "Demo", "123", 12, 2030, "number"},
		{"letters", "4242 4242 abcd 4242", "Demo", "123", 12, 2030, "number"},
		{"no holder", "4242424242424242", " ", "123", 12, 2030, "holderName"},
		{"month 13", "4242424242424242", "Demo", "123", 13, 2030, "expiry"},
		{"expired last year", "4242424242424242", "Demo", "123", 12, 2025, "expiry"},
		{"expired last month", "4242424242424242", "Demo", "123", 9, 2026, "expiry"},
		{"cvv too short", "4242424242424242", "Demo", "12", 12, 2030, "cvv"},
		{"cvv letters", "4242424242424242", "Demo", "12a", 12, 2030, "cvv"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCard("user-1", tt.number, tt.holder, tt.month, tt.year, tt.cvv, now)
			var verr *CardValidationError
			if !errors.As(err, &verr) {
				t.Fatalf("err = %v, want a CardValidationError", err)
			}
			if verr.Field != tt.field {
				t.Errorf("field = %s, want %s", verr.Field, tt.field)
			}
		})
	}
}

func TestNewCardAcceptsCurrentMonth(t *testing.T) {
	if _, err := NewCard("user-1", "4242424242424242", "Demo", 10, 2026, "123", now); err != nil {
		t.Errorf("a card expiring this month is still valid: %v", err)
	}
}

func TestCardBrand(t *testing.T) {
	tests := map[string]string{
		"4242424242424242": "Visa",
		"5555555555554444": "Mastercard",
		"2223003122003222": "Mastercard",
		"378282246310005":  "Amex",
		"6011111111111117": "Card",
	}
	for number, want := range tests {
		if got := CardBrand(number); got != want {
			t.Errorf("CardBrand(%s) = %s, want %s", number, got, want)
		}
	}
}

func TestFailureReason(t *testing.T) {
	if got := FailureReason(ErrNoCard); got != FailureNoCard {
		t.Errorf("no card -> %s", got)
	}
	if got := FailureReason(ErrInsufficientFunds); got != FailureInsufficientFunds {
		t.Errorf("insufficient -> %s", got)
	}
	if got := FailureReason(errors.New("boom")); got != FailureOther {
		t.Errorf("other -> %s", got)
	}
}
