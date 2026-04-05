-- +goose Up
-- +goose StatementBegin

-- Display reference for manuscript list / reviewer UI (nullable until backfilled)
ALTER TABLE manuscripts ADD COLUMN IF NOT EXISTS reference_number BIGINT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_manuscripts_reference_number
    ON manuscripts (reference_number) WHERE reference_number IS NOT NULL;

-- Backfill display numbers for existing manuscripts (stable ordering by creation time)
UPDATE manuscripts m
SET reference_number = s.n
FROM (
    SELECT id, ROW_NUMBER() OVER (ORDER BY created_at ASC, id ASC)::bigint AS n
    FROM manuscripts
    WHERE reference_number IS NULL
) s
WHERE m.id = s.id;

-- Reviewer structured report (draft vs submitted); answers JSON matches form schema field ids
CREATE TABLE IF NOT EXISTS review_assignment_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_assignment_id UUID NOT NULL REFERENCES review_assignments(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    schema_version INT NOT NULL DEFAULT 1,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    UNIQUE(review_assignment_id)
);

CREATE INDEX IF NOT EXISTS idx_review_assignment_reports_assignment
    ON review_assignment_reports(review_assignment_id);

-- Extension requests (editor approval in follow-up)
CREATE TABLE IF NOT EXISTS review_extension_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_assignment_id UUID NOT NULL REFERENCES review_assignments(id) ON DELETE CASCADE,
    requested_due TIMESTAMP NOT NULL,
    reason TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    decided_by UUID REFERENCES users(id) ON DELETE SET NULL,
    decided_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_review_extension_assignment
    ON review_extension_requests(review_assignment_id);

CREATE INDEX IF NOT EXISTS idx_review_assignments_reviewer_status
    ON review_assignments(reviewer_id, status)
    WHERE reviewer_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_review_assignments_reviewer_status;
DROP TABLE IF EXISTS review_extension_requests;
DROP TABLE IF EXISTS review_assignment_reports;
DROP INDEX IF EXISTS idx_manuscripts_reference_number;
ALTER TABLE manuscripts DROP COLUMN IF EXISTS reference_number;
-- +goose StatementEnd
