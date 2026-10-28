DROP INDEX IF EXISTS idx_stock_res_status_expires;
ALTER TABLE stock_reservations DROP COLUMN IF EXISTS status;
