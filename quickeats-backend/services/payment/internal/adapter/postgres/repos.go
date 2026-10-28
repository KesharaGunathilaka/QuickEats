package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quickeats/services/payment/internal/domain"
)

type PaymentRepo struct{ db *pgxpool.Pool }
func NewPaymentRepo(db *pgxpool.Pool) *PaymentRepo { return &PaymentRepo{db: db} }

func (r *PaymentRepo) Create(ctx context.Context, p *domain.Payment) error {
	_, err := r.db.Exec(ctx, `INSERT INTO payments (id,order_id,shop_id,method,status,amount,currency,gateway_ref,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		p.ID, p.OrderID, p.ShopID, p.Method, p.Status, p.Amount, p.Currency, p.GatewayRef, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *PaymentRepo) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	var p domain.Payment
	err := r.db.QueryRow(ctx, `SELECT id,order_id,shop_id,method,status,amount,currency,gateway_ref,created_at,updated_at FROM payments WHERE id=$1`, id).Scan(
		&p.ID, &p.OrderID, &p.ShopID, &p.Method, &p.Status, &p.Amount, &p.Currency, &p.GatewayRef, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
	if err != nil { return nil, fmt.Errorf("get payment: %w", err) }
	return &p, nil
}

func (r *PaymentRepo) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	var p domain.Payment
	err := r.db.QueryRow(ctx, `SELECT id,order_id,shop_id,method,status,amount,currency,gateway_ref,created_at,updated_at FROM payments WHERE order_id=$1`, orderID).Scan(
		&p.ID, &p.OrderID, &p.ShopID, &p.Method, &p.Status, &p.Amount, &p.Currency, &p.GatewayRef, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
	if err != nil { return nil, err }
	return &p, nil
}

func (r *PaymentRepo) UpdateStatus(ctx context.Context, id string, status domain.PaymentStatus, gatewayRef string) error {
	_, err := r.db.Exec(ctx, `UPDATE payments SET status=$1, gateway_ref=$2, updated_at=now() WHERE id=$3`, status, gatewayRef, id)
	return err
}

func (r *PaymentRepo) RefundToCard(ctx context.Context, paymentID, userID string) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil { return false, fmt.Errorf("begin tx: %w", err) }
	defer tx.Rollback(ctx)

	var amount int64
	err = tx.QueryRow(ctx,
		`UPDATE payments SET status=$1, updated_at=now() WHERE id=$2 AND status=$3 RETURNING amount`,
		domain.PaymentRefunded, paymentID, domain.PaymentCompleted).Scan(&amount)
	if errors.Is(err, pgx.ErrNoRows) { return false, nil }
	if err != nil { return false, fmt.Errorf("mark refunded: %w", err) }

	// If the customer removed the card since, the refund is still recorded but has
	// nowhere to go in this demo.
	if _, err := tx.Exec(ctx, `UPDATE cards SET balance = balance + $1, updated_at = now() WHERE user_id = $2`, amount, userID); err != nil {
		return false, fmt.Errorf("credit card: %w", err)
	}
	if err := tx.Commit(ctx); err != nil { return false, fmt.Errorf("commit: %w", err) }
	return true, nil
}

type LedgerRepo struct{ db *pgxpool.Pool }
func NewLedgerRepo(db *pgxpool.Pool) *LedgerRepo { return &LedgerRepo{db: db} }

func (r *LedgerRepo) CreateEntries(ctx context.Context, entries []*domain.LedgerEntry) error {
	for _, e := range entries {
		_, err := r.db.Exec(ctx, `INSERT INTO ledger_entries (id,payment_id,account,entry_type,amount,description,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			e.ID, e.PaymentID, e.Account, e.EntryType, e.Amount, e.Description, e.CreatedAt)
		if err != nil { return err }
	}
	return nil
}

func (r *LedgerRepo) GetByPaymentID(ctx context.Context, paymentID string) ([]*domain.LedgerEntry, error) {
	rows, err := r.db.Query(ctx, `SELECT id,payment_id,account,entry_type,amount,description,created_at FROM ledger_entries WHERE payment_id=$1`, paymentID)
	if err != nil { return nil, err }
	defer rows.Close()
	var entries []*domain.LedgerEntry
	for rows.Next() {
		var e domain.LedgerEntry
		rows.Scan(&e.ID, &e.PaymentID, &e.Account, &e.EntryType, &e.Amount, &e.Description, &e.CreatedAt)
		entries = append(entries, &e)
	}
	return entries, nil
}

type CardRepo struct{ db *pgxpool.Pool }
func NewCardRepo(db *pgxpool.Pool) *CardRepo { return &CardRepo{db: db} }

// Save inserts the card, or replaces the customer's existing card. A replacement is
// a new card, so it starts with the new card's balance.
func (r *CardRepo) Save(ctx context.Context, c *domain.Card) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO cards (id,user_id,brand,last4,holder_name,exp_month,exp_year,balance,created_at,updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		 ON CONFLICT (user_id) DO UPDATE SET
		   id=EXCLUDED.id, brand=EXCLUDED.brand, last4=EXCLUDED.last4, holder_name=EXCLUDED.holder_name,
		   exp_month=EXCLUDED.exp_month, exp_year=EXCLUDED.exp_year, balance=EXCLUDED.balance,
		   created_at=EXCLUDED.created_at, updated_at=EXCLUDED.updated_at`,
		c.ID, c.UserID, c.Brand, c.Last4, c.HolderName, c.ExpMonth, c.ExpYear, c.Balance, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *CardRepo) GetByUserID(ctx context.Context, userID string) (*domain.Card, error) {
	var c domain.Card
	err := r.db.QueryRow(ctx,
		`SELECT id,user_id,brand,last4,holder_name,exp_month,exp_year,balance,created_at,updated_at FROM cards WHERE user_id=$1`, userID,
	).Scan(&c.ID, &c.UserID, &c.Brand, &c.Last4, &c.HolderName, &c.ExpMonth, &c.ExpYear, &c.Balance, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
	if err != nil { return nil, fmt.Errorf("get card: %w", err) }
	return &c, nil
}

func (r *CardRepo) Delete(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM cards WHERE user_id=$1`, userID)
	return err
}

// Debit is a single conditional UPDATE, so two payments at once can't overdraw the card.
func (r *CardRepo) Debit(ctx context.Context, userID string, amount int64) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE cards SET balance = balance - $1, updated_at = now() WHERE user_id = $2 AND balance >= $1`, amount, userID)
	if err != nil { return fmt.Errorf("debit card: %w", err) }
	if tag.RowsAffected() == 1 { return nil }

	// Nothing changed: either there is no card or it can't cover the amount.
	var exists bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM cards WHERE user_id = $1)`, userID).Scan(&exists); err != nil {
		return fmt.Errorf("debit card: %w", err)
	}
	if !exists { return domain.ErrNoCard }
	return domain.ErrInsufficientFunds
}
