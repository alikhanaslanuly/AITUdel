CREATE TABLE IF NOT EXISTS promo_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    discount_percent INT NOT NULL,
    student_only BOOLEAN NOT NULL DEFAULT FALSE,
    max_uses INT NOT NULL DEFAULT 1,
    used_count INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS promo_usages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    promo_id UUID NOT NULL REFERENCES promo_codes(id),
    user_id UUID NOT NULL,
    order_id UUID,
    used_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(promo_id, user_id)
);

INSERT INTO promo_codes (code, discount_percent, student_only, max_uses)
VALUES ('AITU2025', 15, TRUE, 10000);