package domain

import (
	"time"

	"github.com/google/uuid"
)

type Shop struct {
	ID             string    `json:"id"`
	OrganisationID string    `json:"organisationId"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	ImageURL       string    `json:"imageUrl,omitempty"`
	IsOpen         bool      `json:"isOpen"`
	OwnerID        string    `json:"ownerId"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func NewShop(orgID, name, description, ownerID string) *Shop {
	now := time.Now().UTC()
	return &Shop{
		ID:             uuid.New().String(),
		OrganisationID: orgID,
		Name:           name,
		Description:    description,
		IsOpen:         false,
		OwnerID:        ownerID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}
