package mapper

import (
	"testing"
	"time"

	reviewrepo "github.com/api-monolith-template/internal/repository/review"
)

func TestFormatReviewerHistoryMMDD(t *testing.T) {
	got := formatReviewerHistoryMMDD(time.Date(2026, 5, 24, 15, 0, 0, 0, time.FixedZone("JKT", 7*3600)))
	if got != "05-24" {
		t.Fatalf("expected UTC month-day 05-24, got %q", got)
	}
	if formatReviewerHistoryMMDD(time.Time{}) != "" {
		t.Fatal("zero time should yield empty string")
	}
}

func TestToReviewerHistoryItemResponse(t *testing.T) {
	rec := "ACCEPT"
	ed := "REVISION_REQUIRED"
	completed := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	row := reviewrepo.ReviewerHistoryRow{
		AssignmentID:   "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		ManuscriptID:   "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
		AssignedAt:     time.Date(2026, 6, 2, 8, 30, 0, 0, time.UTC),
		Section:        "Case Study",
		Title:          "Example Title",
		Recommendation: &rec,
		EditorDecision: &ed,
		CompletedAt:    &completed,
	}
	out := ToReviewerHistoryItemResponse(row, 7)
	if out.ID != 7 {
		t.Fatalf("id: got %d want 7", out.ID)
	}
	if out.AssignmentID != row.AssignmentID || out.ManuscriptID != row.ManuscriptID {
		t.Fatal("ids mismatch")
	}
	if out.MMDDAssigned != "06-02" {
		t.Fatalf("mm_dd_assigned: got %q", out.MMDDAssigned)
	}
	if out.Sec != "Case Study" || out.Title != "Example Title" {
		t.Fatal("sec/title mismatch")
	}
	if out.Review != "Accepted" || out.EditorDecision != "Revision required" {
		t.Fatalf("labels: review=%q editor=%q", out.Review, out.EditorDecision)
	}
	if out.RecommendationCode == nil || *out.RecommendationCode != "ACCEPT" {
		t.Fatal("recommendation_code")
	}
	if out.EditorDecisionCode == nil || *out.EditorDecisionCode != "REVISION_REQUIRED" {
		t.Fatal("editor_decision_code")
	}
	if out.CompletedAt == nil || !out.CompletedAt.Equal(completed) {
		t.Fatal("completed_at")
	}
}

func TestToReviewerHistoryItemResponse_EmptySectionUsesDefault(t *testing.T) {
	rec := "REJECT"
	row := reviewrepo.ReviewerHistoryRow{
		AssignmentID:   "a",
		ManuscriptID:   "b",
		AssignedAt:     time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
		Section:        "   ",
		Title:          "T",
		Recommendation: &rec,
		EditorDecision: nil,
	}
	out := ToReviewerHistoryItemResponse(row, 1)
	if out.Sec != defaultManuscriptSection {
		t.Fatalf("sec fallback: got %q", out.Sec)
	}
	if out.EditorDecision != "" || out.EditorDecisionCode != nil {
		t.Fatal("expected empty editor decision for nil code")
	}
}
