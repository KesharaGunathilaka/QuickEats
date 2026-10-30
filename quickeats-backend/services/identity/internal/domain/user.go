package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleCustomer  Role = "customer"
	RoleShopStaff Role = "shop_staff"
	RoleShopOwner Role = "shop_owner"
	RoleAdmin     Role = "admin"
)

type User struct {
	ID             string    `json:"id"`
	OrganisationID string    `json:"organisationId"`
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"`
	DisplayName    string    `json:"displayName"`
	Role           Role      `json:"role"`
	ShopID         *string   `json:"shopId,omitempty"`
	Verified       bool      `json:"verified"`
	CreatedAt      time.Time `json:"createdAt"`
	// TokensInvalidBefore: refresh tokens issued before this time are rejected.
	// Zero means no sessions have been revoked. Set on password change and logout.
	TokensInvalidBefore time.Time `json:"-"`
}

func NewUser(orgID, email, passwordHash, displayName string, role Role) *User {
	return &User{
		ID:             uuid.New().String(),
		OrganisationID: orgID,
		Email:          email,
		PasswordHash:   passwordHash,
		DisplayName:    displayName,
		Role:           role,
		Verified:       false,
		CreatedAt:      time.Now().UTC(),
	}
}
