package email

import (
	"strings"
	"testing"
)

func TestNormalizeSubject(t *testing.T) {
	t.Parallel()
	got := NormalizeSubject("  Manuskrip uji curl \u2014 editor001  ")
	want := "Manuskrip uji curl - editor001"
	if got != want {
		t.Fatalf("NormalizeSubject() = %q, want %q", got, want)
	}
}

func TestReviewerInviteSubject(t *testing.T) {
	t.Parallel()
	s := ReviewerInviteSubject("Manuskrip uji curl — editor001")
	if !strings.Contains(s, "Manuskrip uji curl - editor001") {
		t.Fatalf("expected normalized dash in subject, got %q", s)
	}
	if !strings.Contains(s, "MedikaOne Journal") {
		t.Fatalf("expected brand suffix, got %q", s)
	}
	long := strings.Repeat("x", 80)
	s2 := ReviewerInviteSubject(long)
	if !strings.Contains(s2, "…") {
		t.Fatalf("expected truncation ellipsis in long subject, got %q", s2)
	}
	if len([]rune(s2)) > 120 {
		t.Fatalf("subject unexpectedly long: %d runes", len([]rune(s2)))
	}
}
