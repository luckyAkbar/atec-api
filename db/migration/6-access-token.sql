-- +migrate Up notransaction

CREATE TABLE IF NOT EXISTS "access_tokens" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    token TEXT NOT NULL,
    user_id UUID NOT NULL,
    valid_until TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

ALTER TABLE "access_tokens" ADD FOREIGN KEY (user_id) REFERENCES "users" ("id");
ALTER TABLE "access_tokens" ADD CONSTRAINT unique_access_tokens_token_user_id UNIQUE (token,user_id);

CREATE INDEX IF NOT EXISTS idx_access_tokens_id ON "access_tokens" USING HASH(id);
CREATE INDEX IF NOT EXISTS idx_access_tokens_token ON "access_tokens" USING HASH(token);

-- +migrate Down

DROP INDEX IF EXISTS idx_access_tokens_id;
DROP INDEX IF EXISTS idx_access_tokens_token;
DROP TABLE IF EXISTS "access_tokens";