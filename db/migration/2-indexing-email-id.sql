-- +migrate Up notransaction

CREATE INDEX IF NOT EXISTS idx_emails_id ON emails USING btree(id);

-- +migrate Down

DROP INDEX IF EXISTS idx_emails_id;