-- +migrate Up notransaction

ALTER TABLE "pins" RENAME COLUMN "failed_count" TO "remaining_attempts";
ALTER TABLE "pins" ALTER COLUMN "remaining_attempts" SET DEFAULT 3;

-- +migrate Down

ALTER TABLE "pins" RENAME COLUMN "remaining_attempts" TO "failed_count";