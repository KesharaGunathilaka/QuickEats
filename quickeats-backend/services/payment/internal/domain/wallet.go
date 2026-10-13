package domain

import (
	"time"
	"github.com/google/uuid"
)

type Wallet struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Balance   int64     `json:"balance"`
	Currency  string    `json:"currency"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewWallet(userID string) *Wallet {
	return &Wallet{ID: uuid.New().String(), UserID: userID, Balance: 0, Currency: "LKR", UpdatedAt: time.Now().UTC()}
}
