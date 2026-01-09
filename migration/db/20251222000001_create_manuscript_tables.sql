-- +goose Up
-- +goose StatementBegin
CREATE TABLE manuscripts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    volume_number_id UUID NOT NULL REFERENCES issues(id),
    title TEXT NOT NULL,
    abstract TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PUBLISHED',
    main_author_id UUID NOT NULL REFERENCES users(id),
    published_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_manuscripts_deleted_at ON manuscripts(deleted_at);
CREATE INDEX idx_manuscripts_issue_id ON manuscripts(volume_number_id);

CREATE TABLE manuscript_authors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    manuscript_id UUID NOT NULL REFERENCES manuscripts(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    author_name VARCHAR(255) NOT NULL,
    author_email VARCHAR(255) NOT NULL,
    affiliation TEXT,
    is_corresponding BOOLEAN DEFAULT FALSE,
    order_position INT NOT NULL
);

CREATE INDEX idx_manuscript_authors_manuscript_id ON manuscript_authors(manuscript_id);

CREATE TABLE manuscript_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    manuscript_id UUID NOT NULL REFERENCES manuscripts(id) ON DELETE CASCADE,
    file_type VARCHAR(20) NOT NULL,
    file_path TEXT NOT NULL,
    filename TEXT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,
    checksum_sha256 VARCHAR(64),
    version INT NOT NULL DEFAULT 1,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_manuscript_files_manuscript_id ON manuscript_files(manuscript_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS manuscript_files;
DROP TABLE IF EXISTS manuscript_authors;
DROP TABLE IF EXISTS manuscripts;
-- +goose StatementEnd
