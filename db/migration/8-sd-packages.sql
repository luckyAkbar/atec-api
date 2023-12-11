-- +migrate Up notransaction

CREATE TABLE IF NOT EXISTS "test_packages" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_by UUID NOT NULL,
    package JSONB NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    is_locked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

ALTER TABLE "test_packages" ADD FOREIGN KEY (created_by) REFERENCES "users" (id);
ALTER TABLE "test_packages" ADD FOREIGN KEY (template_id) REFERENCES "test_templates" (id);
CREATE INDEX IF NOT EXISTS idx_test_packages_id ON "test_packages" (id);

-- +migrate Down

DROP INDEX IF EXISTS idx_test_packages_id;
DROP TABLE IF EXISTS "test_packages";