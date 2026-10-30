package port

import (
	"time"

	"github.com/quickeats/services/identity/internal/domain"
)

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type TokenIssuer interface {
	Issue(user *domain.User) (*TokenPair, error)
	// ValidateRefreshToken also returns the token's issued-at time, so the caller
	// can reject tokens issued before the user's sessions were last revoked.
	ValidateRefreshToken(tokenStr string) (userID string, issuedAt time.Time, err error)
}
