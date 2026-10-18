package app

import (
	"context"
	"fmt"

	"github.com/quickeats/services/identity/internal/domain"
	"github.com/quickeats/services/identity/internal/port"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	OrganisationSlug string `json:"organisationSlug"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	DisplayName      string `json:"displayName"`
	InviteCode       string `json:"inviteCode,omitempty"`
}

type RegisterUseCase struct {
	userRepo port.UserRepository
	orgRepo  port.OrganisationRepository
}

func NewRegisterUseCase(userRepo port.UserRepository, orgRepo port.OrganisationRepository) *RegisterUseCase {
	return &RegisterUseCase{userRepo: userRepo, orgRepo: orgRepo}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, req RegisterRequest) (*domain.User, error) {
	user, err := prepareUser(ctx, uc.userRepo, uc.orgRepo, req, domain.RoleCustomer)
	if err != nil {
		return nil, err
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// prepareUser runs the checks shared by every sign-up flow (organisation exists,
// email is eligible for it, email is unused) and builds the user with a hashed password.
func prepareUser(ctx context.Context, userRepo port.UserRepository, orgRepo port.OrganisationRepository, req RegisterRequest, role domain.Role) (*domain.User, error) {
	// 1. Find organisation
	org, err := orgRepo.GetBySlug(ctx, req.OrganisationSlug)
	if err != nil || org == nil {
		return nil, domain.ErrOrgNotFound
	}

	// 2. Verify eligibility
	verifier, err := domain.ResolveVerifier(org.VerifyMethod)
	if err != nil {
		return nil, fmt.Errorf("resolve verifier: %w", err)
	}

	// For invite code, inject the code into config for verification
	config := org.VerifyConfig
	if org.VerifyMethod == domain.VerifyInviteCode {
		codes, _ := config["codes"].([]interface{})
		found := false
		for _, c := range codes {
			if c == req.InviteCode {
				found = true
				break
			}
		}
		if !found {
			return nil, domain.ErrVerificationFailed
		}
	}

	if err := verifier.Verify(req.Email, config); err != nil {
		return nil, domain.ErrVerificationFailed
	}

	// 3. Check uniqueness
	exists, err := userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrUserAlreadyExists
	}

	// 4. Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := domain.NewUser(org.ID, req.Email, string(hash), req.DisplayName, role)
	user.Verified = true // auto-verified after passing org verification
	return user, nil
}
