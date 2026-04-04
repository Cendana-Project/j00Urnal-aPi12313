-- +goose Up
-- +goose StatementBegin

-- Invite-by-email: reviewer may not exist until onboarding completes.
ALTER TABLE review_assignments
    ADD COLUMN IF NOT EXISTS invited_email VARCHAR(255);

UPDATE review_assignments ra
SET invited_email = u.email
FROM users u
WHERE ra.reviewer_id IS NOT NULL
  AND u.id = ra.reviewer_id
  AND (ra.invited_email IS NULL OR ra.invited_email = '');

ALTER TABLE review_assignments
    ALTER COLUMN reviewer_id DROP NOT NULL;

CREATE INDEX IF NOT EXISTS idx_review_assignments_invited_email
    ON review_assignments (LOWER(invited_email));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_review_assignments_invited_email;
ALTER TABLE review_assignments DROP COLUMN IF EXISTS invited_email;
-- reviewer_id remains nullable; restoring NOT NULL requires a follow-up migration if needed

-- +goose StatementEnd
