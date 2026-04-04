-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS publication_terms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content TEXT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_publication_terms_version ON publication_terms(version DESC);

ALTER TABLE manuscripts ADD COLUMN IF NOT EXISTS term_id UUID REFERENCES publication_terms(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE manuscripts DROP COLUMN IF EXISTS term_id;
DROP TABLE IF EXISTS publication_terms;
-- +goose StatementEnd
