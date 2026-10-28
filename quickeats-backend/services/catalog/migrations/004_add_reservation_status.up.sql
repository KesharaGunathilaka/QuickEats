-- A reservation is HELD while the order awaits payment (stock is taken from the Redis
-- counter only) and COMMITTED once the order is confirmed (stock is also taken from
-- menu_items.stock, the record of what was sold).
-- Rows from before this migration never reduced menu_items.stock, so they start as
-- HELD; the expiry sweeper releases them and Redis falls back in line with Postgres.
ALTER TABLE stock_reservations ADD COLUMN status VARCHAR(16) NOT NULL DEFAULT 'HELD';
CREATE INDEX idx_stock_res_status_expires ON stock_reservations(status, expires_at);
