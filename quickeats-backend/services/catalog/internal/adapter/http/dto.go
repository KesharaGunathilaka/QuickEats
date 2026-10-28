package http

type CreateShopRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateShopRequest edits a shop; omitted fields are left unchanged.
type UpdateShopRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsOpen      *bool   `json:"isOpen,omitempty"`
}

// InternalCreateShopRequest is sent by identity when a shop owner signs up.
type InternalCreateShopRequest struct {
	OrganisationIDs []string `json:"organisationIds,omitempty"`
	OrganisationID  string   `json:"organisationId,omitempty"` // older single-organisation form
	OwnerID         string   `json:"ownerId"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
}

// InternalShopOrganisationsRequest replaces the organisations a shop serves.
type InternalShopOrganisationsRequest struct {
	OrganisationIDs []string `json:"organisationIds"`
}

type CreateMenuItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Category    string `json:"category"`
	Stock       int    `json:"stock"`
}

type UpdateStockRequest struct {
	Stock int `json:"stock"`
}

type ReserveStockRequest struct {
	OrderID string              `json:"orderId"`
	Items   []ReserveStockItem  `json:"items"`
}

type ReserveStockItem struct {
	MenuItemID string `json:"menuItemId"`
	Quantity   int    `json:"quantity"`
}

type CommitStockRequest struct {
	OrderID string `json:"orderId"`
}

type ReleaseStockRequest struct {
	OrderID        string   `json:"orderId"`
	ReservationIDs []string `json:"reservationIds"`
}
