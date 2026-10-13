package domain

import (
	"time"

	"github.com/google/uuid"
)

type MenuItem struct {
	ID          string    `json:"id"`
	ShopID      string    `json:"shopId"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Price       int64     `json:"price"` // cents
	Category    string    `json:"category"`
	ImageURL    string    `json:"imageUrl,omitempty"`
	Stock       int       `json:"stock"`
	Available   bool      `json:"available"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func NewMenuItem(shopID, name, description string, price int64, category string, stock int) *MenuItem {
	now := time.Now().UTC()
	return &MenuItem{
		ID:          uuid.New().String(),
		ShopID:      shopID,
		Name:        name,
		Description: description,
		Price:       price,
		Category:    category,
		Stock:       stock,
		Available:   stock > 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
