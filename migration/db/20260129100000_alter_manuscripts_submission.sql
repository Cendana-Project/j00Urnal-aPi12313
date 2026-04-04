-- +goose Up
-- +goose StatementBegin
-- Idempotent: schema may already match (GORM/seed/manual) while goose replays migrations.
ALTER TABLE manuscripts ADD COLUMN IF NOT EXISTS journal_id UUID REFERENCES journals(id);
ALTER TABLE manuscripts ADD COLUMN IF NOT EXISTS is_tnc_accepted BOOLEAN DEFAULT FALSE;
ALTER TABLE manuscripts ADD COLUMN IF NOT EXISTS tnc_accepted_at TIMESTAMP;

ALTER TABLE manuscripts ALTER COLUMN volume_number_id DROP NOT NULL;

ALTER TABLE manuscripts ALTER COLUMN status SET DEFAULT 'DRAFT';

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM manuscripts) THEN
        UPDATE manuscripts m
        SET journal_id = (
            SELECT v.journal_id
            FROM volumes v
            JOIN issues i ON i.volume_id = v.id
            WHERE i.id = m.volume_number_id
        )
        WHERE m.journal_id IS NULL
          AND m.volume_number_id IS NOT NULL;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM manuscripts WHERE journal_id IS NULL) THEN
        ALTER TABLE manuscripts ALTER COLUMN journal_id SET NOT NULL;
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE manuscripts ALTER COLUMN status SET DEFAULT 'PUBLISHED';
ALTER TABLE manuscripts ALTER COLUMN volume_number_id SET NOT NULL;
ALTER TABLE manuscripts DROP COLUMN tnc_accepted_at;
ALTER TABLE manuscripts DROP COLUMN is_tnc_accepted;
ALTER TABLE manuscripts DROP COLUMN journal_id;
-- +goose StatementEnd
