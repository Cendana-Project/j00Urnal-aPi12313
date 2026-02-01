-- +goose Up
-- +goose StatementBegin
ALTER TABLE manuscripts ADD COLUMN current_step INTEGER NOT NULL DEFAULT 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE manuscripts DROP COLUMN current_step;
-- +goose StatementEnd
