
INSERT INTO promo_codes (id, code, discount_percent, student_only, max_uses, expires_at) VALUES
  (
    '00000000-0000-0002-0000-000000000001',
    'WELCOME10',
    10,
    FALSE,
    1000,
    NOW() + INTERVAL '6 months'
  ),
  (
    '00000000-0000-0002-0000-000000000002',
    'SUMMER20',
    20,
    FALSE,
    500,
    NOW() + INTERVAL '3 months'
  ),
  (
    '00000000-0000-0002-0000-000000000003',
    'FIRSTORDER',
    25,
    FALSE,
    1,
    NOW() + INTERVAL '1 year'
  )
ON CONFLICT (code) DO NOTHING;

INSERT INTO orders (id, user_id, restaurant_id, status, total_price, promo_code, discount_amount, delivery_address) VALUES
  (
    '00000000-0000-0001-0000-000000000001',
    '00000000-0000-0004-0000-000000000001',
    '00000000-0000-0000-0000-000000000002',
    'delivered',
    2150,
    NULL,
    0,
    'Astana, AITU Campus, Block C'
  ),
  (
    '00000000-0000-0001-0000-000000000002',
    '00000000-0000-0004-0000-000000000002',
    '00000000-0000-0000-0000-000000000003',
    'delivered',
    2117,
    'AITU2025',
    373.65,
    'Astana, AITU Dorm, Room 412'
  ),
  (
    '00000000-0000-0001-0000-000000000003',
    '00000000-0000-0004-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'pending',
    1750,
    NULL,
    0,
    'Astana, AITU Campus, Library'
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO order_items (order_id, item_id, name, quantity, price) VALUES
  (
    '00000000-0000-0001-0000-000000000001',
    '00000000-0000-0000-0000-000000000006',
    'Classic Smash',
    1,
    1400
  ),
  (
    '00000000-0000-0001-0000-000000000001',
    '00000000-0000-0000-0000-000000000010',
    'Loaded Fries',
    1,
    750
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO order_items (order_id, item_id, name, quantity, price) VALUES
  (
    '00000000-0000-0001-0000-000000000002',
    '00000000-0000-0000-0000-000000000011',
    'Pad Thai',
    1,
    1350
  ),
  (
    '00000000-0000-0001-0000-000000000002',
    '00000000-0000-0000-0000-000000000013',
    'Dumplings (8pc)',
    1,
    1100
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO order_items (order_id, item_id, name, quantity, price) VALUES
  (
    '00000000-0000-0001-0000-000000000003',
    '00000000-0000-0000-0000-000000000001',
    'Caesar Salad',
    1,
    950
  ),
  (
    '00000000-0000-0001-0000-000000000003',
    '00000000-0000-0000-0000-000000000002',
    'Chicken Soup',
    1,
    700
  ),
  (
    '00000000-0000-0001-0000-000000000003',
    '00000000-0000-0000-0000-000000000005',
    'Green Smoothie',
    1,
    550
  )
ON CONFLICT (id) DO NOTHING;
