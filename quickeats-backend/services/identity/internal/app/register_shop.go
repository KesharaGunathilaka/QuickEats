package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/quickeats/services/identity/internal/domain"
	"github.com/quickeats/services/identity/internal/port"
)

type RegisterShopRequest struct {
	RegisterRequest
	ShopName        string `json:"shopName"`
	ShopDescription string `json:"shopDescription"`
}

// RegisterShopUseCase signs up a shop owner and creates their shop in the catalog service.
type RegisterShopUseCase struct {
	userRepo    port.UserRepository
	orgRepo     port.OrganisationRepository
	shopCreator port.ShopCreator
}

func NewRegisterShopUseCase(userRepo port.UserRepository, orgRepo port.OrganisationRepository, shopCreator port.ShopCreator) *RegisterShopUseCase {
	return &RegisterShopUseCase{userRepo: userRepo, orgRepo: orgRepo, shopCreator: shopCreator}
}

func (uc *RegisterShopUseCase) Execute(ctx context.Context, req RegisterShopRequest) (*domain.User, error) {
	// 1. Same eligibility checks as a customer sign-up, but as shop owner
	user, err := prepareUser(ctx, uc.userRepo, uc.orgRepo, req.RegisterRequest, domain.RoleShopOwner)
	if err != nil {
		return nil, err
	}

	// 2. Create the shop in catalog (owned by catalog's database), owned by the new user
	shopID, err := uc.shopCreator.CreateShop(ctx, user.OrganisationID, user.ID, req.ShopName, req.ShopDescription)
	if err != nil {
		return nil, fmt.Errorf("create shop: %w", err)
	}

	// 3. Persist the user linked to the shop, so their JWT carries the shop claim
	user.ShopID = &shopID
	if err := uc.userRepo.Create(ctx, user); err != nil {
		// Uniqueness was checked above, so this is rare; the shop is left without a usable owner
		slog.Error("register shop: user insert failed after shop was created", "shop_id", shopID, "error", err)
		return nil, err
	}

	return user, nil
}
