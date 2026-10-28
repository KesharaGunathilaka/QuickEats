-- Dev seed data: catalog database (demo shops and menus).
-- IDs match identity.sql (organisation, shop staff). Safe to re-run.

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
