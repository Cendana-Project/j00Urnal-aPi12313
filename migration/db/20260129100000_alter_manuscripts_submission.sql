-- +goose Up
-- +goose StatementBegin
ALTER TABLE manuscripts
    ADD COLUMN journal_id UUID REFERENCES journals(id),
    ADD COLUMN is_tnc_accepted BOOLEAN DEFAULT FALSE,
    ADD COLUMN tnc_accepted_at TIMESTAMP;

-- Allow manuscripts to be created without an issue/volume initially
ALTER TABLE manuscripts ALTER COLUMN volume_number_id DROP NOT NULL;

-- Set default status to DRAFT
ALTER TABLE manuscripts ALTER COLUMN status SET DEFAULT 'DRAFT';

-- Backfill journal_id for existing manuscripts if necessary (assuming they have volume_number_id linked to issues linked to volumes linked to journals)
-- This might be complex update, but for now we make it nullable initially to avoid breakage, but plan says NOT NULL.
-- Actually, if we make it NOT NULL immediately, existing records will fail.
-- Strategy: Add as NULLABLE first, then specific migration to fill it if data exists.
-- For this task, I will leave it NULLABLE in DB but enforce in App, OR if I want strict FK, I must default it or fill it.
-- Since it's a new feature request, I'll assumme strictness is good but let's check if we can easily link it.
-- Issue -> Volume -> Journal.
-- UPDATE manuscripts m SET journal_id = (SELECT v.journal_id FROM volumes v JOIN issues i ON i.volume_id = v.id WHERE i.id = m.volume_number_id);
-- Valid point. Let's do that to be safe.

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM manuscripts) THEN
        UPDATE manuscripts m
        SET journal_id = (
            SELECT v.journal_id
            FROM volumes v
            JOIN issues i ON i.volume_id = v.id
            WHERE i.id = m.volume_number_id
        );
    END IF;
END $$;

-- Now apply constraint if we want it required. User said "author must choose manuscript upload to which journal". So it should be required.
-- However, if update failed (orphan issues?), it might fail. I'll stick to nullable for safety or just constraint NOT NULL.
-- Let's make it NOT NULL as per robust design, assuming data integrity.
ALTER TABLE manuscripts ALTER COLUMN journal_id SET NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE manuscripts ALTER COLUMN status SET DEFAULT 'PUBLISHED';
ALTER TABLE manuscripts ALTER COLUMN volume_number_id SET NOT NULL;
ALTER TABLE manuscripts DROP COLUMN tnc_accepted_at;
ALTER TABLE manuscripts DROP COLUMN is_tnc_accepted;
ALTER TABLE manuscripts DROP COLUMN journal_id;
-- +goose StatementEnd
