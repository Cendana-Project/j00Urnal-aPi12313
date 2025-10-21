-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS permissions (
                                           id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(96) NOT NULL,
    slug        VARCHAR(96) NOT NULL UNIQUE,
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE,
    deleted_at  TIMESTAMP WITH TIME ZONE
                              );
CREATE UNIQUE INDEX IF NOT EXISTS uq_permissions_name ON permissions(name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS permissions;
-- +goose StatementEnd
