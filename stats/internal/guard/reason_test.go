package guard

import (
	"slices"
	"sync"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// TestReasonKeepsDistinctIDsInFirstSeenOrder pins Reason's contract: distinct
// ids in first-seen order, safe under concurrent calls, a copy out of IDs,
// and a no-op on an Env with no collector.
func TestReasonKeepsDistinctIDsInFirstSeenOrder(t *testing.T) {
	t.Parallel()
	r := &core.Reasons{}
	env := core.Env{Reasons: r}
	core.Reason(env, "b")
	core.Reason(env, "a")
	core.Reason(env, "b")
	if got := r.IDs(); !slices.Equal(got, []string{"b", "a"}) {
		t.Fatalf("IDs() = %q, want [b a]", got)
	}
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			core.Reason(env, "c")
		}()
	}
	wg.Wait()
	want := []string{"b", "a", "c"}
	got := r.IDs()
	if !slices.Equal(got, want) {
		t.Fatalf("IDs() after concurrent calls = %q, want %q", got, want)
	}
	_ = append(got[:1], "z")
	if again := r.IDs(); !slices.Equal(again, want) {
		t.Errorf("IDs() after mutating the returned slice = %q, want %q", again, want)
	}
	core.Reason(core.Env{}, "x")
}

// TestProbeRecordsNoReasonIntoTheCaller: run-verify-list's unchanged probe
// answers "do not skip" with a non-zero exit, which is not run-verify-list's
// refusal, so the parent run's collector stays empty.
func TestProbeRecordsNoReasonIntoTheCaller(t *testing.T) {
	t.Parallel()
	r := &core.Reasons{}
	env := core.Env{Getenv: func(k string) string {
		if k == "FLOW_VERIFY_FULL" {
			return "1"
		}
		return ""
	}, Reasons: r}
	rvlGuardSkip(env, t.TempDir(), "flow-guard unchanged a/ || go test -count=1 ./...")
	if ids := r.IDs(); len(ids) != 0 {
		t.Fatalf("probe recorded %v into the parent run's reasons", ids)
	}
}
