package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/quickeats/services/identity/internal/domain"
	"github.com/quickeats/services/identity/internal/port"
)

type RegisterShopRequest struct {
	Email             string
	Password          string
	DisplayName       string
	OrganisationSlugs []string
	ShopName          string
	ShopDescription   string
}

// RegisterShopUseCase signs up a merchant and creates their shop in the catalog service.
type RegisterShopUseCase struct {
	userRepo port.UserRepository
	orgRepo  port.OrganisationRepository
	shops    port.ShopDirectory
}

func NewRegisterShopUseCase(userRepo port.UserRepository, orgRepo port.OrganisationRepository, shops port.ShopDirectory) *RegisterShopUseCase {
	return &RegisterShopUseCase{userRepo: userRepo, orgRepo: orgRepo, shops: shops}
}

// Execute signs up a merchant. Unlike customers, merchants may use any valid email
// (e.g. Gmail); they choose one or more organisation codes their shop will serve.
func (uc *RegisterShopUseCase) Execute(ctx context.Context, req RegisterShopRequest) (*domain.User, error) {
	orgs, err := resolveOrgs(ctx, uc.orgRepo, req.OrganisationSlugs)
	if err != nil {
		return nil, err
	}

	// The user's own organisation is the first one; shop visibility uses the full list.
	user, err := newAccount(ctx, uc.userRepo, req.Email, req.Password, req.DisplayName, orgs[0].ID, domain.RoleShopOwner)
	if err != nil {
		return nil, err
	}

	shopID, err := uc.shops.CreateShop(ctx, orgIDs(orgs), user.ID, req.ShopName, req.ShopDescription)
	if err != nil {
		return nil, fmt.Errorf("create shop: %w", err)
	}

	// Persist the user linked to the shop, so their JWT carries the shop claim
	user.ShopID = &shopID
	if err := uc.userRepo.Create(ctx, user); err != nil {
		// Uniqueness was checked above, so this is rare; the shop is left without a usable owner
		slog.Error("register shop: user insert failed after shop was created", "shop_id", shopID, "error", err)
		return nil, err
	}
	return user, nil
}

// resolveOrgs looks up every organisation code, dropping duplicates. At least one is required.
func resolveOrgs(ctx context.Context, orgRepo port.OrganisationRepository, slugs []string) ([]*domain.Organisation, error) {
	var orgs []*domain.Organisation
	seen := map[string]bool{}
	for _, s := range slugs {
		org, err := findOrg(ctx, orgRepo, s)
		if err != nil {
			return nil, err
		}
		if !seen[org.ID] {
			seen[org.ID] = true
			orgs = append(orgs, org)
		}
	}
	if len(orgs) == 0 {
		return nil, domain.ErrNoOrganisations
	}
	return orgs, nil
}

func orgIDs(orgs []*domain.Organisation) []string {
	ids := make([]string, len(orgs))
	for i, o := range orgs {
		ids[i] = o.ID
	}
	return ids
}
