-- +migrate Up notransaction

CREATE TYPE "role" AS ENUM ('ADMIN','USER');

CREATE TABLE IF NOT EXISTS "users" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT UNIQUE NOT NULL, 
    password TEXT NOT NULL,
    username VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    "role" role NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_id ON users USING btree(id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users USING hash(email);

-- +migrate Down

DROP INDEX IF EXISTS idx_users_id;
DROP INDEX IF EXISTS idx_users_email;

DROP TABLE IF EXISTS "users";