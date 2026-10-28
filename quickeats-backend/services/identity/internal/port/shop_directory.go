package port

import "context"

// ShopDirectory manages shops in the catalog service, which owns shop data.
type ShopDirectory interface {
	// CreateShop creates a shop visible to the given organisations and returns its ID.
	CreateShop(ctx context.Context, orgIDs []string, ownerID, name, description string) (string, error)
	// ShopOrganisations returns the IDs of the organisations a shop serves.
	ShopOrganisations(ctx context.Context, shopID string) ([]string, error)
	// SetShopOrganisations replaces the organisations a shop serves.
	SetShopOrganisations(ctx context.Context, shopID string, orgIDs []string) error
}
