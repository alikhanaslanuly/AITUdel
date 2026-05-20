CREATE TABLE IF NOT EXISTS notifications_log (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       VARCHAR(50) NOT NULL,    
    subject    TEXT        NOT NULL,
    success    BOOLEAN     NOT NULL DEFAULT FALSE,
    sent_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notif_user_id ON notifications_log(user_id);
CREATE INDEX IF NOT EXISTS idx_notif_type    ON notifications_log(type);
