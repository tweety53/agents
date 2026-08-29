package harvest_test

import (
	"context"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/harvest"
)

// TestNoDepsReturnsZeroValuesAndNoError asserts, per method, that
// harvest.NoDeps{} is a real no-op -- nil error, zero-value result --
// rather than merely a value that compiles against harvest.Deps. Task
// 2's 40 rewritten call sites all rely on NoDeps being inert; a method
// that quietly returned a non-nil error would fail every one of them
// with a misleading message that pointed at the call site rather than
// here.
func TestNoDepsReturnsZeroValuesAndNoError(t *testing.T) {
	ctx := context.Background()
	var d harvest.NoDeps

	if err := d.Price(ctx, 1); err != nil {
		t.Errorf("Price: got error %v, want nil", err)
	}

	tokens, err := d.UnresolvedSessionTokens(ctx)
	if err != nil {
		t.Errorf("UnresolvedSessionTokens: got error %v, want nil", err)
	}
	if tokens != nil {
		t.Errorf("UnresolvedSessionTokens: got %v, want nil map", tokens)
	}

	bound, err := d.BindSession(ctx, "token", "session")
	if err != nil {
		t.Errorf("BindSession: got error %v, want nil", err)
	}
	if bound != 0 {
		t.Errorf("BindSession: got bound=%d, want 0", bound)
	}

	if err := d.RecordSessionTokenGiveUp(ctx, "token", "reason", time.Now()); err != nil {
		t.Errorf("RecordSessionTokenGiveUp: got error %v, want nil", err)
	}

	giveUps, err := d.PersistedGiveUps(ctx)
	if err != nil {
		t.Errorf("PersistedGiveUps: got error %v, want nil", err)
	}
	if giveUps != nil {
		t.Errorf("PersistedGiveUps: got %v, want nil slice", giveUps)
	}

	if err := d.MarkDispatchesUnattributedByID(ctx, []int64{1, 2}, "reason", 2); err != nil {
		t.Errorf("MarkDispatchesUnattributedByID: got error %v, want nil", err)
	}

	if err := d.MarkDispatchesUnattributed(ctx, "token", "reason", 2); err != nil {
		t.Errorf("MarkDispatchesUnattributed: got error %v, want nil", err)
	}

	if err := d.MergeDispatchMetrics(ctx, 1, nil); err != nil {
		t.Errorf("MergeDispatchMetrics: got error %v, want nil", err)
	}

	windows, err := d.DispatchWindowsForSession(ctx, "session")
	if err != nil {
		t.Errorf("DispatchWindowsForSession: got error %v, want nil", err)
	}
	if windows != nil {
		t.Errorf("DispatchWindowsForSession: got %v, want nil slice", windows)
	}
}

// TestNoDepsIsUsableAsEveryConstituentInterface asserts that
// harvest.NoDeps{} satisfies each of the four interfaces harvest.Deps
// composes, individually -- so a Deps that has drifted from what it
// composes is caught here rather than at task 2's call sites.
func TestNoDepsIsUsableAsEveryConstituentInterface(t *testing.T) {
	var (
		_ harvest.Pricer               = harvest.NoDeps{}
		_ harvest.SessionTokenBinder   = harvest.NoDeps{}
		_ harvest.DispatchMetricsSink  = harvest.NoDeps{}
		_ harvest.DispatchWindowSource = harvest.NoDeps{}
		_ harvest.Deps                 = harvest.NoDeps{}
	)
}
