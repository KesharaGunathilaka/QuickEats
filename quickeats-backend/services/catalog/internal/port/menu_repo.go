package port

import (
	"context"
	"github.com/quickeats/services/catalog/internal/domain"
)

type MenuItemRepository interface {
	Create(ctx context.Context, item *domain.MenuItem) error
	GetByID(ctx context.Context, id string) (*domain.MenuItem, error)
	ListByShop(ctx context.Context, shopID string) ([]*domain.MenuItem, error)
	Update(ctx context.Context, item *domain.MenuItem) error
	Delete(ctx context.Context, id string) error
	UpdateStock(ctx context.Context, id string, stock int) error
}
