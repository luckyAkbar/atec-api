-- +migrate Up notransaction

CREATE TABLE IF NOT EXISTS "test_results" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    package_id UUID NOT NULL,
    user_id UUID DEFAULT NULL,
    answer JSONB DEFAULT NULL,
    result JSONB DEFAULT NULL,
    finished_at TIMESTAMPTZ DEFAULT NULL,
    open_until TIMESTAMPTZ NOT NULL,
    submit_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

ALTER TABLE "test_results" ADD FOREIGN KEY (package_id) REFERENCES "test_packages" (id);
ALTER TABLE "test_results" ADD FOREIGN KEY (user_id) REFERENCES "users" (id);
CREATE INDEX IF NOT EXISTS idx_test_results_id ON "test_results" USING HASH(id);
CREATE INDEX IF NOT EXISTS idx_test_results_user_id ON "test_results" USING BTREE(user_id);
CREATE INDEX IF NOT EXISTS idx_test_results_package_id ON "test_results" USING BTREE(package_id);

-- +migrate Down

DROP INDEX IF EXISTS idx_test_results_id;
DROP INDEX IF EXISTS idx_test_results_user_id;
DROP INDEX IF EXISTS idx_test_results_package_id;
DROP TABLE IF EXISTS "test_results";