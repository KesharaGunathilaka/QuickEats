package app

import (
	"context"

	"github.com/quickeats/services/catalog/internal/domain"
	"github.com/quickeats/services/catalog/internal/port"
)

type BrowseShopsUseCase struct {
	shopRepo port.ShopRepository
}

func NewBrowseShopsUseCase(shopRepo port.ShopRepository) *BrowseShopsUseCase {
	return &BrowseShopsUseCase{shopRepo: shopRepo}
}

func (uc *BrowseShopsUseCase) ListByOrg(ctx context.Context, orgID string) ([]*domain.Shop, error) {
	return uc.shopRepo.ListByOrganisation(ctx, orgID)
}

func (uc *BrowseShopsUseCase) GetShop(ctx context.Context, id string) (*domain.Shop, error) {
	return uc.shopRepo.GetByID(ctx, id)
}
