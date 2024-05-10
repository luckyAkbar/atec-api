-- +migrate Up notransaction

ALTER TABLE "emails" ADD COLUMN IF NOT EXISTS deadline BIGINT DEFAULT NULL;

-- +migrate Down

ALTER TABLE "emails" DROP COLUMN IF EXISTS deadline;