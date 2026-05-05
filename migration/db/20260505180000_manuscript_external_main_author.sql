-- +goose Up
-- Primary contact may not be a registered user: store textual main author + who submitted.

ALTER TABLE manuscripts
  ADD COLUMN IF NOT EXISTS submitted_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS external_main_author_name VARCHAR(255),
  ADD COLUMN IF NOT EXISTS external_main_author_email VARCHAR(255),
  ADD COLUMN IF NOT EXISTS external_main_author_affiliation TEXT;

ALTER TABLE manuscripts
  ALTER COLUMN main_author_id DROP NOT NULL;

CREATE INDEX IF NOT EXISTS idx_manuscripts_submitted_by ON manuscripts(submitted_by_user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_manuscripts_submitted_by;

ALTER TABLE manuscripts
  DROP COLUMN IF EXISTS external_main_author_affiliation,
  DROP COLUMN IF EXISTS external_main_author_email,
  DROP COLUMN IF EXISTS external_main_author_name,
  DROP COLUMN IF EXISTS submitted_by_user_id;

-- Cannot safely restore NOT NULL if any NULL main_author_id rows exist.
-- ALTER TABLE manuscripts ALTER COLUMN main_author_id SET NOT NULL;
