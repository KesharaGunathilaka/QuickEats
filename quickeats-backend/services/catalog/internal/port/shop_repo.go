package port

import (
	"context"

	"github.com/quickeats/services/catalog/internal/domain"
)

type ShopRepository interface {
	// Create saves the shop and links it to shop.OrganisationIDs.
	Create(ctx context.Context, shop *domain.Shop) error
	// GetByID returns the shop with its OrganisationIDs filled in.
	GetByID(ctx context.Context, id string) (*domain.Shop, error)
	// ListByOrganisation returns the shops serving an organisation.
	ListByOrganisation(ctx context.Context, orgID string) ([]*domain.Shop, error)
	Update(ctx context.Context, shop *domain.Shop) error
	// SetOrganisations replaces the organisations a shop serves (the first becomes primary).
	SetOrganisations(ctx context.Context, shopID string, orgIDs []string) error
}
