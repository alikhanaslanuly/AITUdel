CREATE TABLE IF NOT EXISTS couriers (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    status     VARCHAR(20) NOT NULL DEFAULT 'offline',   -- active | offline
    latitude   DOUBLE PRECISION NOT NULL DEFAULT 0,
    longitude  DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_couriers_status  ON couriers(status);
CREATE INDEX IF NOT EXISTS idx_couriers_user_id ON couriers(user_id);
