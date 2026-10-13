package domain

import "errors"

var (
	ErrShopNotFound      = errors.New("shop not found")
	ErrMenuItemNotFound  = errors.New("menu item not found")
	ErrStockInsufficient = errors.New("insufficient stock for one or more items")
)
