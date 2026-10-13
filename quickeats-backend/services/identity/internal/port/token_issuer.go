package port

import "github.com/quickeats/services/identity/internal/domain"

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type TokenIssuer interface {
	Issue(user *domain.User) (*TokenPair, error)
	ValidateRefreshToken(tokenStr string) (userID string, err error)
}
