-- Record the shop being paid, so ledger entries credit (and refunds debit) the right
-- shop. Payments made before this have no shop, their ledger entries were posted to
-- the order ID by mistake, and they were all wallet payments (the wallet is gone), so
-- they are removed rather than carried over. All statements run in one transaction.
DELETE FROM ledger_entries;
DELETE FROM payments;

ALTER TABLE payments ADD COLUMN shop_id VARCHAR(36) NOT NULL;
