-- +goose Up
-- +goose StatementBegin
CREATE TABLE publication_terms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content TEXT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for ordering by version
CREATE INDEX idx_publication_terms_version ON publication_terms(version DESC);

-- Alter Manuscripts to link to accepted term
ALTER TABLE manuscripts
    ADD COLUMN term_id UUID REFERENCES publication_terms(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE manuscripts DROP COLUMN term_id;
DROP TABLE publication_terms;
-- +goose StatementEnd
