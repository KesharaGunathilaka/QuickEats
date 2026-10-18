package port

import "context"

// ShopCreator creates a shop in the catalog service and returns its ID.
type ShopCreator interface {
	CreateShop(ctx context.Context, orgID, ownerID, name, description string) (string, error)
}
