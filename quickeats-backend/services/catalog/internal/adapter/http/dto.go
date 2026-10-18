package http

type CreateShopRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// InternalCreateShopRequest is sent by identity when a shop owner signs up.
type InternalCreateShopRequest struct {
	OrganisationID string `json:"organisationId"`
	OwnerID        string `json:"ownerId"`
	Name           string `json:"name"`
	Description    string `json:"description"`
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

type ReleaseStockRequest struct {
	OrderID        string   `json:"orderId"`
	ReservationIDs []string `json:"reservationIds"`
}
