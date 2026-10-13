package port

import (
	"context"
	"github.com/quickeats/services/order/internal/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error
	SetPickupCode(ctx context.Context, id, code string) error
	SetRejectReason(ctx context.Context, id, reason string) error
	ListByShop(ctx context.Context, shopID string, limit int, cursor string) ([]*domain.Order, error)
	ListByCustomer(ctx context.Context, customerID string, limit int, cursor string) ([]*domain.Order, error)
	FindTimedOutOrders(ctx context.Context, status domain.OrderStatus, olderThan int) ([]*domain.Order, error)
}
