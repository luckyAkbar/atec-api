-- +migrate Up notransaction

CREATE TABLE IF NOT EXISTS "test_templates" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_by UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    template JSONB NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    is_locked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

ALTER TABLE "test_templates" ADD FOREIGN KEY (created_by) REFERENCES "users" (id);
CREATE INDEX IF NOT EXISTS idx_test_templates_id ON "test_templates" USING hash(id);

-- +migrate Down

DROP INDEX IF EXISTS idx_test_templates_id;
DROP TABLE IF EXISTS "test_templates";