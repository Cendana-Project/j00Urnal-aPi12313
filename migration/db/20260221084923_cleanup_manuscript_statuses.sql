-- +goose Up
-- +goose StatementBegin
-- 1. Sync existing statuses to the new simplified flow
UPDATE manuscripts SET status = 'SUBMITTED' WHERE status IN ('DRAFT', 'UNDER_CHIEF_REVIEW');

-- 2. Update default value for future rows
ALTER TABLE manuscripts ALTER COLUMN status SET DEFAULT 'SUBMITTED';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE manuscripts ALTER COLUMN status SET DEFAULT 'DRAFT';
-- +goose StatementEnd
