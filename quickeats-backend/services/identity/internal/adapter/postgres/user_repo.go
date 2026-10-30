package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quickeats/services/identity/internal/domain"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

// Admins have no organisation: "" is stored as NULL and read back as "".
const userColumns = `id, COALESCE(organisation_id::text, ''), email, password_hash, display_name, role, shop_id, verified, created_at, COALESCE(tokens_invalid_before, 'epoch')`

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO users (id, organisation_id, email, password_hash, display_name, role, shop_id, verified, created_at)
		 VALUES ($1, NULLIF($2, '')::uuid, $3, $4, $5, $6, $7, $8, $9)`,
		user.ID, user.OrganisationID, user.Email, user.PasswordHash,
		user.DisplayName, user.Role, user.ShopID, user.Verified, user.CreatedAt,
	)
	if isUniqueViolation(err) {
		return domain.ErrUserAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE lower(email) = lower($1)`, email)
	return scanUser(row)
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id)
	return scanUser(row)
}

func (r *UserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE lower(email) = lower($1))`, email).Scan(&exists)
	return exists, err
}

func (r *UserRepo) UpdateDisplayName(ctx context.Context, id, displayName string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET display_name = $1 WHERE id = $2`, displayName, id)
	return err
}

func (r *UserRepo) UpdatePasswordHash(ctx context.Context, id, hash string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, hash, id)
	return err
}

func (r *UserRepo) UpdateOrganisation(ctx context.Context, id, orgID string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET organisation_id = NULLIF($1, '')::uuid WHERE id = $2`, orgID, id)
	return err
}

func (r *UserRepo) InvalidateTokens(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET tokens_invalid_before = now() WHERE id = $1`, id)
	return err
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(&u.ID, &u.OrganisationID, &u.Email, &u.PasswordHash,
		&u.DisplayName, &u.Role, &u.ShopID, &u.Verified, &u.CreatedAt, &u.TokensInvalidBefore)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &u, nil
}
