package domain

import (
	"time"

	"github.com/google/uuid"
)

type Shop struct {
	ID string `json:"id"`
	// OrganisationID is the shop's primary organisation (the first one its owner chose).
	OrganisationID string `json:"organisationId"`
	// OrganisationIDs lists every organisation whose customers can see the shop.
	OrganisationIDs []string  `json:"organisationIds,omitempty"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	ImageURL        string    `json:"imageUrl,omitempty"`
	IsOpen          bool      `json:"isOpen"`
	OwnerID         string    `json:"ownerId"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func NewShop(orgIDs []string, name, description, ownerID string) *Shop {
	now := time.Now().UTC()
	primary := ""
	if len(orgIDs) > 0 {
		primary = orgIDs[0]
	}
	return &Shop{
		ID:              uuid.New().String(),
		OrganisationID:  primary,
		OrganisationIDs: orgIDs,
		Name:            name,
		Description:     description,
		IsOpen:          false,
		OwnerID:         ownerID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// ServesOrganisation reports whether customers of the organisation may see the shop.
func (s *Shop) ServesOrganisation(orgID string) bool {
	for _, id := range s.OrganisationIDs {
		if id == orgID {
			return true
		}
	}
	return false
}
