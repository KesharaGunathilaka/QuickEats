package port

import (
	"context"

	"github.com/quickeats/services/identity/internal/domain"
)

type OrganisationRepository interface {
	Create(ctx context.Context, org *domain.Organisation) error
	GetByID(ctx context.Context, id string) (*domain.Organisation, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Organisation, error)
	List(ctx context.Context) ([]*domain.Organisation, error)
}
