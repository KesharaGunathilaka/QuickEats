-- Why an order was cancelled, so the app can explain it: e.g. insufficient_funds or
-- no_card (payment declined), payment_timeout, cancelled_by_customer.
ALTER TABLE orders ADD COLUMN cancel_reason TEXT NOT NULL DEFAULT '';
