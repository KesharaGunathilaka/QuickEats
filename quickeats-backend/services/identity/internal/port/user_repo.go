package port

import (
	"context"

	"github.com/quickeats/services/identity/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateDisplayName(ctx context.Context, id, displayName string) error
	UpdatePasswordHash(ctx context.Context, id, hash string) error
	UpdateOrganisation(ctx context.Context, id, orgID string) error
}
