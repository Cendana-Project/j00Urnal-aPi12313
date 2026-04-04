-- Required for GET /v1/reviewer/history (SELECT m.section). Safe to apply on all envs (IF NOT EXISTS).
-- +goose Up
-- +goose StatementBegin
ALTER TABLE manuscripts ADD COLUMN IF NOT EXISTS section VARCHAR(150);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE manuscripts DROP COLUMN IF EXISTS section;
-- +goose StatementEnd
