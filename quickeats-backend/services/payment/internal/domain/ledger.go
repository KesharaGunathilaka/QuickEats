package domain

import (
	"time"
	"github.com/google/uuid"
)

type EntryType string
const (
	Debit  EntryType = "DEBIT"
	Credit EntryType = "CREDIT"
)

type LedgerEntry struct {
	ID          string    `json:"id"`
	PaymentID   string    `json:"paymentId"`
	Account     string    `json:"account"`
	EntryType   EntryType `json:"entryType"`
	Amount      int64     `json:"amount"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

func NewLedgerEntry(paymentID, account string, entryType EntryType, amount int64, description string) *LedgerEntry {
	return &LedgerEntry{ID: uuid.New().String(), PaymentID: paymentID, Account: account, EntryType: entryType, Amount: amount, Description: description, CreatedAt: time.Now().UTC()}
}
