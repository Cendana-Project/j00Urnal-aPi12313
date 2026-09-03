package review

import (
	"errors"
	"testing"
	"time"
)

func TestNotifyReviewerAcceptedAfterSuccessfulTransaction(t *testing.T) {
	const wantReviewerID = "reviewer-created-in-transaction"

	notified := make(chan string, 1)

	err := notifyReviewerAcceptedAfterTransaction(nil, wantReviewerID, func(reviewerID string) {
		notified <- reviewerID
	})
	if err != nil {
		t.Fatalf("notifyReviewerAcceptedAfterTransaction() error = %v", err)
	}

	select {
	case reviewerID := <-notified:
		if reviewerID != wantReviewerID {
			t.Fatalf("notification reviewer ID = %q, want %q", reviewerID, wantReviewerID)
		}
	case <-time.After(time.Second):
		t.Fatal("notification was not dispatched after the transaction completed")
	}
}

func TestNotifyReviewerAcceptedSkipsNotificationOnTransactionError(t *testing.T) {
	wantErr := errors.New("commit failed")
	notified := make(chan struct{}, 1)

	err := notifyReviewerAcceptedAfterTransaction(wantErr, "reviewer-that-was-rolled-back", func(string) {
		notified <- struct{}{}
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("notifyReviewerAcceptedAfterTransaction() error = %v, want %v", err, wantErr)
	}

	select {
	case <-notified:
		t.Fatal("notification was dispatched for a failed transaction")
	case <-time.After(100 * time.Millisecond):
	}
}
