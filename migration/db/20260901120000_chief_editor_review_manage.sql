-- +goose Up
-- Chief Editor is already special-cased to bypass the "must be the assigned editor" check in
-- validateEditorAssignment() (internal/service/review/review.service.go) for accept/decline/
-- request-revision/round-decision — the code has always intended Chief Editor to act on these,
-- and to see review round/assignment detail (who's reviewing, round status). But the review.manage
-- permission that gates the entire /v1/editor/* route group (including the read-only
-- GET /v1/editor/manuscripts/:id/reviews) was never granted to the CHIEF_EDITOR role, so none of
-- that was actually reachable. Symptom reported: the reviewer name never showed up in the Chief
-- Editor's participant sidebar, because the frontend has to fall back to a generic endpoint that
-- doesn't return review-round data at all when the user lacks this permission.
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, NOW()
FROM roles r, permissions p
WHERE r.slug = 'CHIEF_EDITOR' AND p.slug = 'review.manage'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- +goose Down
DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE slug = 'CHIEF_EDITOR')
  AND permission_id = (SELECT id FROM permissions WHERE slug = 'review.manage');
