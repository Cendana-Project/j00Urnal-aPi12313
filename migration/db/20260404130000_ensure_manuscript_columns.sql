-- +goose Up
-- Align manuscripts with entity.Manuscript (fixes pq: bind message has 15 result formats but query has 14 columns
-- when older DBs missed a migration or were partially applied).
ALTER TABLE manuscripts ADD COLUMN IF NOT EXISTS journal_id UUID REFERENCES journals(id);
ALTER TABLE manuscripts ADD COLUMN IF NOT EXISTS assigned_editor_id UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE manuscripts ADD COLUMN IF NOT EXISTS term_id UUID REFERENCES publication_terms(id);
ALTER TABLE manuscripts ADD COLUMN IF NOT EXISTS is_tnc_accepted BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE manuscripts ADD COLUMN IF NOT EXISTS tnc_accepted_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_manuscripts_assigned_editor ON manuscripts(assigned_editor_id);

-- +goose Down
-- Intentional no-op: dropping columns risks data loss if this migration added them.
