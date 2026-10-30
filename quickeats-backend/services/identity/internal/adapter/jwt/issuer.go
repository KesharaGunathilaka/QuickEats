package jwt

import (
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/quickeats/services/identity/internal/domain"
	"github.com/quickeats/services/identity/internal/port"
)

type Issuer struct {
	secret       []byte
	expiry       time.Duration
	refreshExpiry time.Duration
}

func NewIssuer(secret string, expiry, refreshExpiry time.Duration) *Issuer {
	return &Issuer{
		secret:        []byte(secret),
		expiry:        expiry,
		refreshExpiry: refreshExpiry,
	}
}

func (i *Issuer) Issue(user *domain.User) (*port.TokenPair, error) {
	now := time.Now()

	// Access token
	accessClaims := jwtlib.MapClaims{
		"sub":  user.ID,
		"org":  user.OrganisationID,
		"role": string(user.Role),
		"iat":  now.Unix(),
		"exp":  now.Add(i.expiry).Unix(),
	}
	if user.ShopID != nil {
		accessClaims["shop"] = *user.ShopID
	}

	accessToken := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(i.secret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	// Refresh token
	refreshClaims := jwtlib.MapClaims{
		"sub":  user.ID,
		"type": "refresh",
		"iat":  now.Unix(),
		"exp":  now.Add(i.refreshExpiry).Unix(),
	}
	refreshToken := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString(i.secret)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &port.TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	}, nil
}

func (i *Issuer) ValidateRefreshToken(tokenStr string) (string, time.Time, error) {
	token, err := jwtlib.Parse(tokenStr, func(t *jwtlib.Token) (interface{}, error) {
		return i.secret, nil
	})
	if err != nil || !token.Valid {
		return "", time.Time{}, fmt.Errorf("invalid refresh token")
	}
	claims, ok := token.Claims.(jwtlib.MapClaims)
	if !ok {
		return "", time.Time{}, fmt.Errorf("invalid claims")
	}
	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return "", time.Time{}, fmt.Errorf("not a refresh token")
	}
	userID, _ := claims["sub"].(string)
	iat, _ := claims["iat"].(float64)
	return userID, time.Unix(int64(iat), 0), nil
}
