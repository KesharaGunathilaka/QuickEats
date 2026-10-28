package domain

import "errors"

var (
	ErrPaymentNotFound   = errors.New("payment not found")
	ErrNoCard            = errors.New("no card on file")
	ErrInsufficientFunds = errors.New("insufficient card balance")
	ErrGatewayTimeout    = errors.New("payment gateway timeout")
	ErrDuplicatePayment  = errors.New("payment for this order already exists")
)

// Failure reasons sent in payment.failed events and to the order service, which
// stores them on the cancelled order so the app can say why payment didn't go through.
const (
	FailureNoCard            = "no_card"
	FailureInsufficientFunds = "insufficient_funds"
	FailureOther             = "payment_failed"
)

// FailureReason maps a strategy error to the reason code clients understand.
func FailureReason(err error) string {
	switch {
	case errors.Is(err, ErrNoCard):
		return FailureNoCard
	case errors.Is(err, ErrInsufficientFunds):
		return FailureInsufficientFunds
	default:
		return FailureOther
	}
}
