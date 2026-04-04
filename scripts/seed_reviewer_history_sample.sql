-- Reviewer history UI sample: 4 baris dengan judul / section / tanggal assign (MM-DD) / review / keputusan editor.
-- Prasyarat: migrasi `20260404143000_manuscripts_section.sql` sudah jalan (kolom `manuscripts.section`).
-- Hapus baris lama (ID tetap) lalu insert ulang — aman dijalankan berkali-kali.

-- Referensi seed: journal 990e8400-e29b-41d4-a716-446655440001, author ...4040, editor ...4020, reviewer ...4050

BEGIN;

DELETE FROM review_assignments WHERE id IN (
    'ffee2000-0000-4000-8000-000000000001'::uuid,
    'ffee2000-0000-4000-8000-000000000002'::uuid,
    'ffee2000-0000-4000-8000-000000000003'::uuid,
    'ffee2000-0000-4000-8000-000000000004'::uuid
);
DELETE FROM review_rounds WHERE id IN (
    'ffee1000-0000-4000-8000-000000000001'::uuid,
    'ffee1000-0000-4000-8000-000000000002'::uuid,
    'ffee1000-0000-4000-8000-000000000003'::uuid,
    'ffee1000-0000-4000-8000-000000000004'::uuid
);
DELETE FROM manuscripts WHERE id IN (
    'ffeedd01-0000-4000-8000-000000000001'::uuid,
    'ffeedd01-0000-4000-8000-000000000002'::uuid,
    'ffeedd01-0000-4000-8000-000000000003'::uuid,
    'ffeedd01-0000-4000-8000-000000000004'::uuid
);

INSERT INTO manuscripts (
    id,
    journal_id,
    volume_number_id,
    title,
    section,
    abstract,
    status,
    main_author_id,
    assigned_editor_id,
    is_tnc_accepted,
    published_at,
    created_at,
    updated_at
) VALUES
(
    'ffeedd01-0000-4000-8000-000000000001'::uuid,
    '990e8400-e29b-41d4-a716-446655440001'::uuid,
    NULL,
    'Neural Pathways in Cognitive Rehabilitation After Stroke',
    'Original Research',
    'Sample abstract for reviewer history mock.',
    'UNDER_REVIEW',
    '550e8400-e29b-41d4-a716-446655440040'::uuid,
    '550e8400-e29b-41d4-a716-446655440020'::uuid,
    TRUE,
    '2026-01-01 00:00:00+00',
    '2026-01-15 00:00:00+00',
    NOW()
),
(
    'ffeedd01-0000-4000-8000-000000000002'::uuid,
    '990e8400-e29b-41d4-a716-446655440001'::uuid,
    NULL,
    'Long-Term Outcomes of Minimally Invasive Lumbar Spine Surgery',
    'Case Study',
    'Sample abstract for reviewer history mock.',
    'UNDER_REVIEW',
    '550e8400-e29b-41d4-a716-446655440040'::uuid,
    '550e8400-e29b-41d4-a716-446655440020'::uuid,
    TRUE,
    '2026-01-01 00:00:00+00',
    '2026-01-15 00:00:00+00',
    NOW()
),
(
    'ffeedd01-0000-4000-8000-000000000003'::uuid,
    '990e8400-e29b-41d4-a716-446655440001'::uuid,
    NULL,
    'Platelet-Rich Plasma in Tendon Healing: A Systematic Review and Meta-Analysis',
    'Original Research',
    'Sample abstract for reviewer history mock.',
    'UNDER_REVIEW',
    '550e8400-e29b-41d4-a716-446655440040'::uuid,
    '550e8400-e29b-41d4-a716-446655440020'::uuid,
    TRUE,
    '2026-01-01 00:00:00+00',
    '2026-01-15 00:00:00+00',
    NOW()
),
(
    'ffeedd01-0000-4000-8000-000000000004'::uuid,
    '990e8400-e29b-41d4-a716-446655440001'::uuid,
    NULL,
    'Biomarkers for Early Detection of Alzheimer''s Disease',
    'Case Study',
    'Sample abstract for reviewer history mock.',
    'UNDER_REVIEW',
    '550e8400-e29b-41d4-a716-446655440040'::uuid,
    '550e8400-e29b-41d4-a716-446655440020'::uuid,
    TRUE,
    '2026-01-01 00:00:00+00',
    '2026-01-15 00:00:00+00',
    NOW()
);

INSERT INTO review_rounds (
    id,
    manuscript_id,
    round_number,
    status,
    editor_decision,
    created_by,
    created_at,
    updated_at
) VALUES
(
    'ffee1000-0000-4000-8000-000000000001'::uuid,
    'ffeedd01-0000-4000-8000-000000000001'::uuid,
    1,
    'IN_REVIEW',
    'ACCEPT',
    '550e8400-e29b-41d4-a716-446655440020'::uuid,
    '2026-05-01 00:00:00+00',
    NOW()
),
(
    'ffee1000-0000-4000-8000-000000000002'::uuid,
    'ffeedd01-0000-4000-8000-000000000002'::uuid,
    1,
    'IN_REVIEW',
    'REVISION_REQUIRED',
    '550e8400-e29b-41d4-a716-446655440020'::uuid,
    '2026-05-01 00:00:00+00',
    NOW()
),
(
    'ffee1000-0000-4000-8000-000000000003'::uuid,
    'ffeedd01-0000-4000-8000-000000000003'::uuid,
    1,
    'IN_REVIEW',
    'REJECT',
    '550e8400-e29b-41d4-a716-446655440020'::uuid,
    '2026-05-01 00:00:00+00',
    NOW()
),
(
    'ffee1000-0000-4000-8000-000000000004'::uuid,
    'ffeedd01-0000-4000-8000-000000000004'::uuid,
    1,
    'IN_REVIEW',
    NULL,
    '550e8400-e29b-41d4-a716-446655440020'::uuid,
    '2026-05-01 00:00:00+00',
    NOW()
);

-- Urutan tampilan API: completed_at DESC → baris 1 paling baru
INSERT INTO review_assignments (
    id,
    review_round_id,
    reviewer_id,
    invited_email,
    assigned_by,
    status,
    invitation_token,
    invitation_expires_at,
    due_date,
    recommendation,
    comments,
    completed_at,
    created_at,
    updated_at
) VALUES
(
    'ffee2000-0000-4000-8000-000000000001'::uuid,
    'ffee1000-0000-4000-8000-000000000001'::uuid,
    '550e8400-e29b-41d4-a716-446655440050'::uuid,
    'reviewer001@journalapi.id',
    '550e8400-e29b-41d4-a716-446655440020'::uuid,
    'COMPLETED',
    NULL,
    NOW(),
    NOW(),
    'ACCEPT',
    'History mock row 1',
    '2026-04-22 12:00:00+00',
    '2026-05-24 09:00:00+00',
    NOW()
),
(
    'ffee2000-0000-4000-8000-000000000002'::uuid,
    'ffee1000-0000-4000-8000-000000000002'::uuid,
    '550e8400-e29b-41d4-a716-446655440050'::uuid,
    'reviewer001@journalapi.id',
    '550e8400-e29b-41d4-a716-446655440020'::uuid,
    'COMPLETED',
    NULL,
    NOW(),
    NOW(),
    'MINOR_REVISION',
    'History mock row 2',
    '2026-04-10 12:00:00+00',
    '2026-06-02 11:30:00+00',
    NOW()
),
(
    'ffee2000-0000-4000-8000-000000000003'::uuid,
    'ffee1000-0000-4000-8000-000000000003'::uuid,
    '550e8400-e29b-41d4-a716-446655440050'::uuid,
    'reviewer001@journalapi.id',
    '550e8400-e29b-41d4-a716-446655440020'::uuid,
    'COMPLETED',
    NULL,
    NOW(),
    NOW(),
    'REJECT',
    'History mock row 3',
    '2026-03-25 12:00:00+00',
    '2026-04-15 08:15:00+00',
    NOW()
),
(
    'ffee2000-0000-4000-8000-000000000004'::uuid,
    'ffee1000-0000-4000-8000-000000000004'::uuid,
    '550e8400-e29b-41d4-a716-446655440050'::uuid,
    'reviewer001@journalapi.id',
    '550e8400-e29b-41d4-a716-446655440020'::uuid,
    'COMPLETED',
    NULL,
    NOW(),
    NOW(),
    'MAJOR_REVISION',
    'History mock row 4 — editor decision still pending',
    '2026-02-01 12:00:00+00',
    '2026-03-10 16:45:00+00',
    NOW()
);

COMMIT;

-- GET /v1/reviewer/history (JWT reviewer001) → expect items[].id 1..4, mm_dd_assigned 05-24, 06-02, 04-15, 03-10,
-- sec / title sesuai INSERT; review: Accepted, Minor revision, Declined, Major revision;
-- editor_decision: Accept, Revision required, Reject, (kosong → FE: Pending). UUID di assignment_id.
