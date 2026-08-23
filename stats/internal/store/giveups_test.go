package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/harvest"
)

// TestRecordSessionTokenGiveUp states the contract RecordSessionTokenGiveUp
// and PersistedGiveUps satisfy together: recording a give-up for a token
// makes it show up in the persisted list, carrying the reason the watcher
// gave and starting at a retry count of zero -- this is the first time the
// watcher has abandoned this token, so nothing has retried it yet.
func TestRecordSessionTokenGiveUp(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	const (
		token  = "mf-giveup-record"
		reason = "session never bound"
	)
	gaveUpAt := time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC)

	if err := st.RecordSessionTokenGiveUp(ctx, token, reason, gaveUpAt); err != nil {
		t.Fatalf("RecordSessionTokenGiveUp: %v", err)
	}

	giveUps, err := st.PersistedGiveUps(ctx)
	if err != nil {
		t.Fatalf("PersistedGiveUps: %v", err)
	}
	if len(giveUps) != 1 {
		t.Fatalf("PersistedGiveUps returned %d rows, want 1", len(giveUps))
	}

	got := giveUps[0]
	if got.Token != token {
		t.Errorf("Token = %q, want %q", got.Token, token)
	}
	if got.Reason != reason {
		t.Errorf("Reason = %q, want %q", got.Reason, reason)
	}
	if got.Retries != 0 {
		t.Errorf("Retries = %d, want 0 on the first recording", got.Retries)
	}
}

// TestPersistedGiveUpsAreListed asserts PersistedGiveUps returns one entry
// per distinct token, each carrying the reason it was recorded with -- this
// is the read the watcher makes at start, over every token it has ever
// abandoned, not just the most recent one.
func TestPersistedGiveUpsAreListed(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	first := time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC)
	second := first.Add(time.Hour)

	if err := st.RecordSessionTokenGiveUp(ctx, "mf-giveup-a", "session never bound", first); err != nil {
		t.Fatalf("RecordSessionTokenGiveUp a: %v", err)
	}
	if err := st.RecordSessionTokenGiveUp(ctx, "mf-giveup-b", "matched more than one session", second); err != nil {
		t.Fatalf("RecordSessionTokenGiveUp b: %v", err)
	}

	giveUps, err := st.PersistedGiveUps(ctx)
	if err != nil {
		t.Fatalf("PersistedGiveUps: %v", err)
	}
	if len(giveUps) != 2 {
		t.Fatalf("PersistedGiveUps returned %d rows, want 2", len(giveUps))
	}

	byToken := make(map[string]harvest.GiveUp, len(giveUps))
	for _, g := range giveUps {
		byToken[g.Token] = g
	}
	if g, ok := byToken["mf-giveup-a"]; !ok || g.Reason != "session never bound" {
		t.Errorf("mf-giveup-a = %+v, ok=%v, want reason %q", g, ok, "session never bound")
	}
	if g, ok := byToken["mf-giveup-b"]; !ok || g.Reason != "matched more than one session" {
		t.Errorf("mf-giveup-b = %+v, ok=%v, want reason %q", g, ok, "matched more than one session")
	}
}

// TestGiveUpIsIdempotent covers the case the watcher hits on every restart
// that still cannot resolve a token: recording the same token a second time
// must update the existing row rather than insert a second one, and the
// retry count must climb, so a persisted give-up records how many times the
// watcher has come back to it rather than just the latest attempt.
func TestGiveUpIsIdempotent(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	const token = "mf-giveup-retry"
	first := time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC)
	second := first.Add(5 * time.Minute)

	if err := st.RecordSessionTokenGiveUp(ctx, token, "session never bound", first); err != nil {
		t.Fatalf("first RecordSessionTokenGiveUp: %v", err)
	}
	if err := st.RecordSessionTokenGiveUp(ctx, token, "session never bound", second); err != nil {
		t.Fatalf("second RecordSessionTokenGiveUp: %v", err)
	}

	giveUps, err := st.PersistedGiveUps(ctx)
	if err != nil {
		t.Fatalf("PersistedGiveUps: %v", err)
	}
	if len(giveUps) != 1 {
		t.Fatalf("PersistedGiveUps returned %d rows, want 1 -- recording the same token twice must update, not insert a second row", len(giveUps))
	}
	if giveUps[0].Retries != 1 {
		t.Errorf("Retries = %d, want 1 after the second recording", giveUps[0].Retries)
	}
}
