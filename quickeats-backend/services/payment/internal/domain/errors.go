package domain

import "errors"

var (
	ErrPaymentNotFound         = errors.New("payment not found")
	ErrWalletNotFound          = errors.New("wallet not found")
	ErrWalletInsufficientFunds = errors.New("insufficient wallet balance")
	ErrGatewayTimeout          = errors.New("payment gateway timeout")
	ErrDuplicatePayment        = errors.New("payment for this order already exists")
)
