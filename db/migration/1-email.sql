-- +migrate Up notransaction
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS "emails" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    subject VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    "to" VARCHAR(255) ARRAY CHECK (CARDINALITY("to") > 0),
    cc VARCHAR(255) ARRAY DEFAULT null,
    bcc VARCHAR(255) ARRAY DEFAULT null,
    sent_at TIMESTAMPTZ DEFAULT null,
    client_signature VARCHAR(50) DEFAULT null,
    metadata VARCHAR(255) DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL DEFAULT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS "emails";