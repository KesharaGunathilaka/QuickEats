package port

import (
	"context"
	"github.com/quickeats/services/catalog/internal/domain"
)

type ShopRepository interface {
	Create(ctx context.Context, shop *domain.Shop) error
	GetByID(ctx context.Context, id string) (*domain.Shop, error)
	ListByOrganisation(ctx context.Context, orgID string) ([]*domain.Shop, error)
	Update(ctx context.Context, shop *domain.Shop) error
}
