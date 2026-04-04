package util

import (
	"slices"
	"testing"
)

func TestInvitationTokenCandidates_padDroppedEquals(t *testing.T) {
	// DB stores padded base64url; client path often omits trailing '='
	dbToken := "X7UuXC4V5NSIC0ak-mrb_xmB6eyZ-sBDUbt4qMneRM0="
	fromPath := "X7UuXC4V5NSIC0ak-mrb_xmB6eyZ-sBDUbt4qMneRM0"
	got := InvitationTokenCandidates(fromPath)
	if !slices.Contains(got, dbToken) {
		t.Fatalf("expected padded form in candidates, got %v", got)
	}
}

func TestInvitationTokenCandidates_roundTripPadded(t *testing.T) {
	s := "ab=="
	got := InvitationTokenCandidates(s)
	if !slices.Contains(got, s) {
		t.Fatalf("expected original, got %v", got)
	}
}
