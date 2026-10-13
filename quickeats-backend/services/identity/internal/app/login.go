package app

import (
	"context"

	"github.com/quickeats/services/identity/internal/domain"
	"github.com/quickeats/services/identity/internal/port"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUseCase struct {
	userRepo    port.UserRepository
	tokenIssuer port.TokenIssuer
}

func NewLoginUseCase(userRepo port.UserRepository, tokenIssuer port.TokenIssuer) *LoginUseCase {
	return &LoginUseCase{userRepo: userRepo, tokenIssuer: tokenIssuer}
}

func (uc *LoginUseCase) Execute(ctx context.Context, req LoginRequest) (*port.TokenPair, error) {
	user, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	tokens, err := uc.tokenIssuer.Issue(user)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}
