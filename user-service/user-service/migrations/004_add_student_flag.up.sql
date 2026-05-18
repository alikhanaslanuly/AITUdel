ALTER TABLE users ADD COLUMN IF NOT EXISTS promo_assigned VARCHAR(50);

CREATE INDEX IF NOT EXISTS idx_users_is_student ON users(is_student) WHERE is_student = TRUE;
