-- Dev seed data for local testing of the mobile apps: identity database.
-- Each service has its own Postgres server, so each database is seeded separately
-- (see catalog.sql and payment.sql).
-- Safe to re-run: every insert is ON CONFLICT DO NOTHING / DO UPDATE.
--
-- Logins (password for all: password123)
--   customer@quickeats.dev  -> customer (no card yet: add one in the app's Card tab)
--   staff@quickeats.dev     -> shop_staff for "Campus Canteen"
-- New customers and shop owners can self-register with organisation code "demo"
-- and an @quickeats.dev or @gmail.com email address.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO organisations (id, name, slug, verify_method, verify_config)
VALUES ('11111111-1111-1111-1111-111111111111', 'QuickEats Demo University', 'demo',
        'email_domain', '{"domains": ["quickeats.dev", "gmail.com"]}')
ON CONFLICT (id) DO UPDATE SET verify_config = EXCLUDED.verify_config, updated_at = now();

INSERT INTO users (id, organisation_id, email, password_hash, display_name, role, shop_id, verified)
VALUES
  ('22222222-2222-2222-2222-000000000001', '11111111-1111-1111-1111-111111111111',
   'customer@quickeats.dev', crypt('password123', gen_salt('bf', 10)), 'Demo Customer', 'customer', NULL, true),
  ('22222222-2222-2222-2222-000000000002', '11111111-1111-1111-1111-111111111111',
   'staff@quickeats.dev', crypt('password123', gen_salt('bf', 10)), 'Canteen Staff', 'shop_staff',
   '33333333-3333-3333-3333-000000000001', true)
ON CONFLICT DO NOTHING;
