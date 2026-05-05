-- +goose Up
ALTER TABLE manuscript_authors
  ADD COLUMN IF NOT EXISTS is_primary_author BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_manuscript_authors_primary
  ON manuscript_authors (manuscript_id)
  WHERE is_primary_author = TRUE;

-- +goose Down
DROP INDEX IF EXISTS idx_manuscript_authors_primary;
ALTER TABLE manuscript_authors DROP COLUMN IF EXISTS is_primary_author;
