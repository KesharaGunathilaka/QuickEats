package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quickeats/services/identity/internal/domain"
)

type OrgRepo struct {
	db *pgxpool.Pool
}

func NewOrgRepo(db *pgxpool.Pool) *OrgRepo {
	return &OrgRepo{db: db}
}

func (r *OrgRepo) Create(ctx context.Context, org *domain.Organisation) error {
	configJSON, _ := json.Marshal(org.VerifyConfig)
	_, err := r.db.Exec(ctx,
		`INSERT INTO organisations (id, name, slug, verify_method, verify_config, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		org.ID, org.Name, org.Slug, org.VerifyMethod, configJSON, org.CreatedAt, org.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert organisation: %w", err)
	}
	return nil
}

func (r *OrgRepo) GetByID(ctx context.Context, id string) (*domain.Organisation, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, slug, verify_method, verify_config, created_at, updated_at
		 FROM organisations WHERE id = $1`, id)
	return scanOrg(row)
}

func (r *OrgRepo) GetBySlug(ctx context.Context, slug string) (*domain.Organisation, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, slug, verify_method, verify_config, created_at, updated_at
		 FROM organisations WHERE slug = $1`, slug)
	return scanOrg(row)
}

func (r *OrgRepo) List(ctx context.Context) ([]*domain.Organisation, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, slug, verify_method, verify_config, created_at, updated_at FROM organisations ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*domain.Organisation
	for rows.Next() {
		org, err := scanOrgRows(rows)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, nil
}

func scanOrg(row pgx.Row) (*domain.Organisation, error) {
	var o domain.Organisation
	var configJSON []byte
	err := row.Scan(&o.ID, &o.Name, &o.Slug, &o.VerifyMethod, &configJSON, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan organisation: %w", err)
	}
	json.Unmarshal(configJSON, &o.VerifyConfig)
	return &o, nil
}

func scanOrgRows(rows pgx.Rows) (*domain.Organisation, error) {
	var o domain.Organisation
	var configJSON []byte
	err := rows.Scan(&o.ID, &o.Name, &o.Slug, &o.VerifyMethod, &configJSON, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(configJSON, &o.VerifyConfig)
	return &o, nil
}
