DROP INDEX IF EXISTS idx_users_is_student;
ALTER TABLE users DROP COLUMN IF EXISTS promo_assigned;
