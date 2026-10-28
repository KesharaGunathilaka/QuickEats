-- Dev seed data for local testing of the mobile apps.
-- Safe to re-run: every insert is ON CONFLICT DO NOTHING.
--
-- Logins (password for all: password123)
--   customer@quickeats.dev  -> customer (wallet: LKR 5,000.00)
--   staff@quickeats.dev     -> shop_staff for "Campus Canteen"
-- New customers and shop owners can self-register with organisation code "demo"
-- and an @quickeats.dev or @gmail.com email address.

\connect identity_db

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

\connect catalog_db

INSERT INTO shops (id, organisation_id, name, description, is_open, owner_id)
VALUES
  ('33333333-3333-3333-3333-000000000001', '11111111-1111-1111-1111-111111111111',
   'Campus Canteen', 'Rice & curry, short eats and drinks', true, '22222222-2222-2222-2222-000000000002'),
  ('33333333-3333-3333-3333-000000000002', '11111111-1111-1111-1111-111111111111',
   'Library Cafe', 'Coffee, tea and snacks', true, '22222222-2222-2222-2222-000000000002')
ON CONFLICT DO NOTHING;

-- Customers only see shops linked to their organisation
INSERT INTO shop_organisations (shop_id, organisation_id)
VALUES
  ('33333333-3333-3333-3333-000000000001', '11111111-1111-1111-1111-111111111111'),
  ('33333333-3333-3333-3333-000000000002', '11111111-1111-1111-1111-111111111111')
ON CONFLICT DO NOTHING;

-- Prices are in cents (LKR)
INSERT INTO menu_items (id, shop_id, name, description, price, category, stock, available)
VALUES
  ('44444444-4444-4444-4444-000000000001', '33333333-3333-3333-3333-000000000001',
   'Chicken Rice & Curry', 'Rice, chicken curry, dhal and two vegetables', 45000, 'Mains', 30, true),
  ('44444444-4444-4444-4444-000000000002', '33333333-3333-3333-3333-000000000001',
   'Vegetable Fried Rice', 'With egg and chilli paste', 38000, 'Mains', 25, true),
  ('44444444-4444-4444-4444-000000000003', '33333333-3333-3333-3333-000000000001',
   'Fish Bun', 'Freshly baked', 12000, 'Short Eats', 40, true),
  ('44444444-4444-4444-4444-000000000004', '33333333-3333-3333-3333-000000000001',
   'Iced Milo', '', 20000, 'Drinks', 20, true),
  ('44444444-4444-4444-4444-000000000005', '33333333-3333-3333-3333-000000000002',
   'Cappuccino', '', 35000, 'Coffee', 50, true),
  ('44444444-4444-4444-4444-000000000006', '33333333-3333-3333-3333-000000000002',
   'Chocolate Muffin', '', 18000, 'Snacks', 15, true)
ON CONFLICT DO NOTHING;

\connect payment_db

INSERT INTO wallets (id, user_id, balance, currency)
VALUES ('55555555-5555-5555-5555-000000000001', '22222222-2222-2222-2222-000000000001', 500000, 'LKR')
ON CONFLICT DO NOTHING;
