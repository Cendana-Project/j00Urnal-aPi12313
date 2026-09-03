package review

import (
	"strings"
	"testing"
)

func TestClaimInvitedAssignmentQueryIsAtomicAndLiteral(t *testing.T) {
	const assignmentID = "11111111-1111-4111-8111-111111111111"
	const reviewerID = "22222222-2222-4222-8222-222222222222"

	query := claimInvitedAssignmentQuery(assignmentID, reviewerID)
	assertLiteralAcceptanceQuery(t, query)

	wants := []string{
		"SET reviewer_id = '" + reviewerID + "'::uuid",
		"WHERE id = '" + assignmentID + "'::uuid",
		"AND status = 'INVITED'",
		"AND reviewer_id IS NULL",
		"AND invitation_token IS NOT NULL",
		"invitation_token = NULL",
	}
	for _, want := range wants {
		if !strings.Contains(query, want) {
			t.Errorf("claim query missing %q:\n%s", want, query)
		}
	}
}

func TestAcceptInvitedAssignmentQueryChecksReviewerAndStatus(t *testing.T) {
	const assignmentID = "11111111-1111-4111-8111-111111111111"
	const reviewerID = "22222222-2222-4222-8222-222222222222"

	query := acceptInvitedAssignmentQuery(assignmentID, reviewerID)
	assertLiteralAcceptanceQuery(t, query)

	wants := []string{
		"WHERE id = '" + assignmentID + "'::uuid",
		"AND reviewer_id = '" + reviewerID + "'::uuid",
		"AND status = 'INVITED'",
	}
	for _, want := range wants {
		if !strings.Contains(query, want) {
			t.Errorf("logged-in acceptance query missing %q:\n%s", want, query)
		}
	}
}

func TestAcceptanceQueriesRejectUnsafeUUIDText(t *testing.T) {
	const unsafe = "not-a-uuid' OR TRUE --"

	queries := []string{
		claimInvitedAssignmentQuery(unsafe, unsafe),
		acceptInvitedAssignmentQuery(unsafe, unsafe),
	}
	for _, query := range queries {
		if strings.Contains(query, unsafe) {
			t.Fatalf("unsafe UUID text was interpolated into query:\n%s", query)
		}
		assertLiteralAcceptanceQuery(t, query)
	}
}

func assertLiteralAcceptanceQuery(t *testing.T, query string) {
	t.Helper()
	if strings.Contains(query, "?") {
		t.Fatalf("acceptance transition query uses a bound placeholder:\n%s", query)
	}
	if !strings.Contains(query, "status = 'ACCEPTED'") {
		t.Fatalf("acceptance transition query does not set ACCEPTED:\n%s", query)
	}
}
