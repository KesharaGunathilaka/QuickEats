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

const shopColumns = `s.id, s.organisation_id, s.name, s.description, s.image_url, s.is_open, s.owner_id, s.created_at, s.updated_at`

func scanShop(row pgx.Row) (*domain.Shop, error) {
	var s domain.Shop
	err := row.Scan(&s.ID, &s.OrganisationID, &s.Name, &s.Description, &s.ImageURL, &s.IsOpen, &s.OwnerID, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *ShopRepo) Create(ctx context.Context, shop *domain.Shop) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO shops (id, organisation_id, name, description, image_url, is_open, owner_id, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		shop.ID, shop.OrganisationID, shop.Name, shop.Description, shop.ImageURL, shop.IsOpen, shop.OwnerID, shop.CreatedAt, shop.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert shop: %w", err)
	}
	if err := linkOrganisations(ctx, tx, shop.ID, shop.OrganisationIDs); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *ShopRepo) GetByID(ctx context.Context, id string) (*domain.Shop, error) {
	s, err := scanShop(r.db.QueryRow(ctx, `SELECT `+shopColumns+` FROM shops s WHERE s.id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get shop: %w", err)
	}

	rows, err := r.db.Query(ctx, `SELECT organisation_id::text FROM shop_organisations WHERE shop_id=$1`, id)
	if err != nil {
		return nil, fmt.Errorf("get shop organisations: %w", err)
	}
	defer rows.Close()
	s.OrganisationIDs = []string{}
	for rows.Next() {
		var orgID string
		if err := rows.Scan(&orgID); err != nil {
			return nil, err
		}
		s.OrganisationIDs = append(s.OrganisationIDs, orgID)
	}
	return s, rows.Err()
}

func (r *ShopRepo) ListByOrganisation(ctx context.Context, orgID string) ([]*domain.Shop, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+shopColumns+` FROM shops s
		 JOIN shop_organisations so ON so.shop_id = s.id
		 WHERE so.organisation_id = $1 ORDER BY s.name`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	shops := []*domain.Shop{}
	for rows.Next() {
		s, err := scanShop(rows)
		if err != nil {
			return nil, err
		}
		shops = append(shops, s)
	}
	return shops, rows.Err()
}

func (r *ShopRepo) Update(ctx context.Context, shop *domain.Shop) error {
	_, err := r.db.Exec(ctx,
		`UPDATE shops SET name=$1, description=$2, image_url=$3, is_open=$4, updated_at=now() WHERE id=$5`,
		shop.Name, shop.Description, shop.ImageURL, shop.IsOpen, shop.ID)
	return err
}

func (r *ShopRepo) SetOrganisations(ctx context.Context, shopID string, orgIDs []string) error {
	if len(orgIDs) == 0 {
		return fmt.Errorf("a shop must serve at least one organisation")
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM shop_organisations WHERE shop_id=$1`, shopID); err != nil {
		return err
	}
	if err := linkOrganisations(ctx, tx, shopID, orgIDs); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE shops SET organisation_id=$1, updated_at=now() WHERE id=$2`, orgIDs[0], shopID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func linkOrganisations(ctx context.Context, tx pgx.Tx, shopID string, orgIDs []string) error {
	for _, orgID := range orgIDs {
		_, err := tx.Exec(ctx,
			`INSERT INTO shop_organisations (shop_id, organisation_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, shopID, orgID)
		if err != nil {
			return fmt.Errorf("link shop to organisation: %w", err)
		}
	}
	return nil
}
