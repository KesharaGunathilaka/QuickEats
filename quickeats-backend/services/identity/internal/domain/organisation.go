package domain

import (
	"time"

	"github.com/google/uuid"
)

type VerifyMethod string

const (
	VerifyEmailDomain VerifyMethod = "email_domain"
	VerifyInviteCode  VerifyMethod = "invite_code"
)

type Organisation struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	VerifyMethod VerifyMethod           `json:"verifyMethod"`
	VerifyConfig map[string]interface{} `json:"verifyConfig"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
}

func NewOrganisation(name, slug string, method VerifyMethod, config map[string]interface{}) *Organisation {
	now := time.Now().UTC()
	return &Organisation{
		ID:           uuid.New().String(),
		Name:         name,
		Slug:         slug,
		VerifyMethod: method,
		VerifyConfig: config,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
