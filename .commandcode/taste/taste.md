# Taste (Continuously Learned by [CommandCode][cmd])

[cmd]: https://commandcode.ai/

# api
- For `POST /v1/editor/invite-reviewer`: use `manuscript_id` and `email` in the request body (not `round_id`/`reviewer_id`/`due_date`); the reviewer is auto-assigned to the first available round with a 7-day default expiry. Confidence: 0.75
- For manuscript authors: authors array is purely informational (name/email/affiliation as strings, no user ID), not linked to user accounts. The submitter/primary contact is always the authenticated user and is separate from the authors list. Confidence: 0.80

# architecture
- Store reviewer form copywriting and questions (labels, options, sections) in a single external JSON file as the single source of truth, not hardcoded in Go logic; embed at build time with `go:embed`. Confidence: 0.85

