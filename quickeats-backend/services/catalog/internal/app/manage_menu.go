package app

import (
	"context"

	"github.com/quickeats/services/catalog/internal/domain"
	"github.com/quickeats/services/catalog/internal/port"
)

type ManageMenuUseCase struct {
	menuRepo   port.MenuItemRepository
	stockCache port.StockCache
}

func NewManageMenuUseCase(menuRepo port.MenuItemRepository, stockCache port.StockCache) *ManageMenuUseCase {
	return &ManageMenuUseCase{menuRepo: menuRepo, stockCache: stockCache}
}

func (uc *ManageMenuUseCase) ListItems(ctx context.Context, shopID string) ([]*domain.MenuItem, error) {
	return uc.menuRepo.ListByShop(ctx, shopID)
}

func (uc *ManageMenuUseCase) CreateItem(ctx context.Context, shopID, name, desc string, price int64, category string, stock int) (*domain.MenuItem, error) {
	item := domain.NewMenuItem(shopID, name, desc, price, category, stock)
	if err := uc.menuRepo.Create(ctx, item); err != nil {
		return nil, err
	}
	// Warm Redis cache
	uc.stockCache.SetStock(ctx, item.ID, stock)
	return item, nil
}

func (uc *ManageMenuUseCase) UpdateItem(ctx context.Context, item *domain.MenuItem) error {
	return uc.menuRepo.Update(ctx, item)
}

func (uc *ManageMenuUseCase) DeleteItem(ctx context.Context, id string) error {
	return uc.menuRepo.Delete(ctx, id)
}

func (uc *ManageMenuUseCase) UpdateStock(ctx context.Context, id string, stock int) error {
	if err := uc.menuRepo.UpdateStock(ctx, id, stock); err != nil {
		return err
	}
	return uc.stockCache.SetStock(ctx, id, stock)
}
