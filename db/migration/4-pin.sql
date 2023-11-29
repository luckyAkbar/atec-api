-- +migrate Up notransaction

CREATE TABLE IF NOT EXISTS "pins" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pin TEXT NOT NULL,
    user_id UUID NOT NULL,
    expired_at TIMESTAMPTZ NOT NULL,
    failed_count INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

ALTER TABLE "pins" ADD FOREIGN KEY (user_id) REFERENCES "users" (id);

CREATE INDEX IF NOT EXISTS idx_pins_id ON "pins" USING hash(id);

-- +migrate Down

DROP INDEX IF EXISTS idx_pins_id;
DROP TABLE IF EXISTS "pins";