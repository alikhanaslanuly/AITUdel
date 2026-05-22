
INSERT INTO users (id, email, password_hash, name, phone, role, is_student, promo_assigned) VALUES
  (
    '00000000-0000-0004-0000-000000000001',
    'demo@aitu.kz',
    crypt('demo1234', gen_salt('bf', 10)),
    'Aizat Bekova',
    '+77011234567',
    'user',
    FALSE,
    NULL
  ),
  (
    '00000000-0000-0004-0000-000000000002',
    'student@aitu.kz',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LExZCNSiu2G',
    'Damir Seitkali',
    '+77027654321',
    'user',
    TRUE,
    'AITU2025'
  ),
  (
    '00000000-0000-0004-0000-000000000003',
    'courier@aitu.kz',
    crypt('demo1234', gen_salt('bf', 10)),
    'Yerlan Dzhaksybekov',
    '+77030001122',
    'courier',
    FALSE,
    NULL
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO couriers (id, user_id, status, latitude, longitude) VALUES
  (
    '00000000-0000-0003-0000-000000000001',
    '00000000-0000-0004-0000-000000000003',
    'active',
    51.0906,
    71.4100
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO notifications_log (user_id, type, subject, success) VALUES
  ('00000000-0000-0004-0000-000000000001', 'welcome',         'Welcome to AITUdel!',                TRUE),
  ('00000000-0000-0004-0000-000000000002', 'welcome',         'Welcome to AITUdel, student!',       TRUE),
  ('00000000-0000-0004-0000-000000000002', 'promo',           'Your student promo code: AITU2025',  TRUE),
  ('00000000-0000-0004-0000-000000000001', 'order_delivered', 'Your order has been delivered!',     TRUE)
ON CONFLICT DO NOTHING;
