-- +goose Up
-- +goose StatementBegin

-- 1. Add assigned_editor_id to manuscripts
ALTER TABLE manuscripts
    ADD COLUMN assigned_editor_id UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX idx_manuscripts_assigned_editor ON manuscripts(assigned_editor_id);

-- 2. Review Rounds
CREATE TABLE review_rounds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    manuscript_id UUID NOT NULL REFERENCES manuscripts(id) ON DELETE CASCADE,
    round_number INT NOT NULL DEFAULT 1,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    editor_decision VARCHAR(30),
    decision_comments TEXT,
    decision_at TIMESTAMP,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    UNIQUE(manuscript_id, round_number)
);

CREATE INDEX idx_review_rounds_manuscript ON review_rounds(manuscript_id);

-- 3. Review Assignments
CREATE TABLE review_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_round_id UUID NOT NULL REFERENCES review_rounds(id) ON DELETE CASCADE,
    reviewer_id UUID NOT NULL REFERENCES users(id),
    assigned_by UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'INVITED',
    invitation_token VARCHAR(128) UNIQUE,
    invitation_expires_at TIMESTAMP NOT NULL,
    invitation_accepted_at TIMESTAMP,
    due_date TIMESTAMP NOT NULL,
    recommendation VARCHAR(20),
    comments TEXT,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);

CREATE INDEX idx_review_assignments_round ON review_assignments(review_round_id);
CREATE INDEX idx_review_assignments_reviewer ON review_assignments(reviewer_id);
CREATE INDEX idx_review_assignments_token ON review_assignments(invitation_token);

-- 4. Review Files
CREATE TABLE review_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_assignment_id UUID REFERENCES review_assignments(id) ON DELETE CASCADE,
    review_round_id UUID NOT NULL REFERENCES review_rounds(id) ON DELETE CASCADE,
    uploaded_by UUID NOT NULL REFERENCES users(id),
    file_type VARCHAR(20) NOT NULL,
    file_path TEXT NOT NULL,
    filename TEXT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_review_files_assignment ON review_files(review_assignment_id);
CREATE INDEX idx_review_files_round ON review_files(review_round_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS review_files;
DROP TABLE IF EXISTS review_assignments;
DROP TABLE IF EXISTS review_rounds;
ALTER TABLE manuscripts DROP COLUMN IF EXISTS assigned_editor_id;
-- +goose StatementEnd
