package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quickeats/services/catalog/internal/domain"
)

type ShopRepo struct{ db *pgxpool.Pool }

func NewShopRepo(db *pgxpool.Pool) *ShopRepo { return &ShopRepo{db: db} }

func (r *ShopRepo) Create(ctx context.Context, shop *domain.Shop) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO shops (id, organisation_id, name, description, image_url, is_open, owner_id, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		shop.ID, shop.OrganisationID, shop.Name, shop.Description, shop.ImageURL, shop.IsOpen, shop.OwnerID, shop.CreatedAt, shop.UpdatedAt)
	return err
}

func (r *ShopRepo) GetByID(ctx context.Context, id string) (*domain.Shop, error) {
	var s domain.Shop
	err := r.db.QueryRow(ctx,
		`SELECT id,organisation_id,name,description,image_url,is_open,owner_id,created_at,updated_at FROM shops WHERE id=$1`, id,
	).Scan(&s.ID, &s.OrganisationID, &s.Name, &s.Description, &s.ImageURL, &s.IsOpen, &s.OwnerID, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
	if err != nil { return nil, fmt.Errorf("get shop: %w", err) }
	return &s, nil
}

func (r *ShopRepo) ListByOrganisation(ctx context.Context, orgID string) ([]*domain.Shop, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id,organisation_id,name,description,image_url,is_open,owner_id,created_at,updated_at FROM shops WHERE organisation_id=$1 ORDER BY name`, orgID)
	if err != nil { return nil, err }
	defer rows.Close()
	var shops []*domain.Shop
	for rows.Next() {
		var s domain.Shop
		if err := rows.Scan(&s.ID, &s.OrganisationID, &s.Name, &s.Description, &s.ImageURL, &s.IsOpen, &s.OwnerID, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		shops = append(shops, &s)
	}
	return shops, nil
}

func (r *ShopRepo) Update(ctx context.Context, shop *domain.Shop) error {
	_, err := r.db.Exec(ctx,
		`UPDATE shops SET name=$1, description=$2, image_url=$3, is_open=$4, updated_at=now() WHERE id=$5`,
		shop.Name, shop.Description, shop.ImageURL, shop.IsOpen, shop.ID)
	return err
}
