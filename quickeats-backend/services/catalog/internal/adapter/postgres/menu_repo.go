package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quickeats/services/catalog/internal/domain"
)

type MenuItemRepo struct{ db *pgxpool.Pool }

func NewMenuItemRepo(db *pgxpool.Pool) *MenuItemRepo { return &MenuItemRepo{db: db} }

func (r *MenuItemRepo) Create(ctx context.Context, item *domain.MenuItem) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO menu_items (id,shop_id,name,description,price,category,image_url,stock,available,created_at,updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		item.ID, item.ShopID, item.Name, item.Description, item.Price, item.Category, item.ImageURL, item.Stock, item.Available, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *MenuItemRepo) GetByID(ctx context.Context, id string) (*domain.MenuItem, error) {
	var m domain.MenuItem
	err := r.db.QueryRow(ctx,
		`SELECT id,shop_id,name,description,price,category,image_url,stock,available,created_at,updated_at FROM menu_items WHERE id=$1`, id,
	).Scan(&m.ID, &m.ShopID, &m.Name, &m.Description, &m.Price, &m.Category, &m.ImageURL, &m.Stock, &m.Available, &m.CreatedAt, &m.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
	if err != nil { return nil, fmt.Errorf("get menu item: %w", err) }
	return &m, nil
}

func (r *MenuItemRepo) ListByShop(ctx context.Context, shopID string) ([]*domain.MenuItem, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id,shop_id,name,description,price,category,image_url,stock,available,created_at,updated_at
		 FROM menu_items WHERE shop_id=$1 ORDER BY category, name`, shopID)
	if err != nil { return nil, err }
	defer rows.Close()
	var items []*domain.MenuItem
	for rows.Next() {
		var m domain.MenuItem
		if err := rows.Scan(&m.ID, &m.ShopID, &m.Name, &m.Description, &m.Price, &m.Category, &m.ImageURL, &m.Stock, &m.Available, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, &m)
	}
	return items, nil
}

func (r *MenuItemRepo) Update(ctx context.Context, item *domain.MenuItem) error {
	_, err := r.db.Exec(ctx,
		`UPDATE menu_items SET name=$1,description=$2,price=$3,category=$4,image_url=$5,stock=$6,available=$7,updated_at=now() WHERE id=$8`,
		item.Name, item.Description, item.Price, item.Category, item.ImageURL, item.Stock, item.Available, item.ID)
	return err
}

func (r *MenuItemRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM menu_items WHERE id=$1`, id)
	return err
}

func (r *MenuItemRepo) UpdateStock(ctx context.Context, id string, stock int) error {
	_, err := r.db.Exec(ctx, `UPDATE menu_items SET stock=$1, available=$2, updated_at=now() WHERE id=$3`, stock, stock > 0, id)
	return err
}
