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
	_, err := r.db.Exec(ctx, `INSERT INTO payments (id,order_id,method,status,amount,currency,gateway_ref,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		p.ID, p.OrderID, p.Method, p.Status, p.Amount, p.Currency, p.GatewayRef, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *PaymentRepo) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	var p domain.Payment
	err := r.db.QueryRow(ctx, `SELECT id,order_id,method,status,amount,currency,gateway_ref,created_at,updated_at FROM payments WHERE id=$1`, id).Scan(
		&p.ID, &p.OrderID, &p.Method, &p.Status, &p.Amount, &p.Currency, &p.GatewayRef, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
	if err != nil { return nil, fmt.Errorf("get payment: %w", err) }
	return &p, nil
}

func (r *PaymentRepo) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	var p domain.Payment
	err := r.db.QueryRow(ctx, `SELECT id,order_id,method,status,amount,currency,gateway_ref,created_at,updated_at FROM payments WHERE order_id=$1`, orderID).Scan(
		&p.ID, &p.OrderID, &p.Method, &p.Status, &p.Amount, &p.Currency, &p.GatewayRef, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
	if err != nil { return nil, err }
	return &p, nil
}

func (r *PaymentRepo) UpdateStatus(ctx context.Context, id string, status domain.PaymentStatus, gatewayRef string) error {
	_, err := r.db.Exec(ctx, `UPDATE payments SET status=$1, gateway_ref=$2, updated_at=now() WHERE id=$3`, status, gatewayRef, id)
	return err
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

type WalletRepo struct{ db *pgxpool.Pool }
func NewWalletRepo(db *pgxpool.Pool) *WalletRepo { return &WalletRepo{db: db} }

func (r *WalletRepo) Create(ctx context.Context, w *domain.Wallet) error {
	_, err := r.db.Exec(ctx, `INSERT INTO wallets (id,user_id,balance,currency,updated_at) VALUES ($1,$2,$3,$4,$5)`, w.ID, w.UserID, w.Balance, w.Currency, w.UpdatedAt)
	return err
}

func (r *WalletRepo) GetByUserID(ctx context.Context, userID string) (*domain.Wallet, error) {
	var w domain.Wallet
	err := r.db.QueryRow(ctx, `SELECT id,user_id,balance,currency,updated_at FROM wallets WHERE user_id=$1`, userID).Scan(&w.ID, &w.UserID, &w.Balance, &w.Currency, &w.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
	return &w, err
}

func (r *WalletRepo) Debit(ctx context.Context, userID string, amount int64) error {
	_, err := r.db.Exec(ctx, `UPDATE wallets SET balance = balance - $1, updated_at=now() WHERE user_id=$2 AND balance >= $1`, amount, userID)
	return err
}

func (r *WalletRepo) Credit(ctx context.Context, userID string, amount int64) error {
	_, err := r.db.Exec(ctx, `UPDATE wallets SET balance = balance + $1, updated_at=now() WHERE user_id=$2`, amount, userID)
	return err
}
