-- Dev seed data: payment database (the demo customer's wallet).
-- user_id matches customer@quickeats.dev in identity.sql. Safe to re-run.

INSERT INTO wallets (id, user_id, balance, currency)
VALUES ('55555555-5555-5555-5555-000000000001', '22222222-2222-2222-2222-000000000001', 500000, 'LKR')
ON CONFLICT DO NOTHING;
