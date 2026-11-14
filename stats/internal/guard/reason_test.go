package guard

import (
	"slices"
	"sync"
	"testing"
)

// TestReasonKeepsDistinctIDsInFirstSeenOrder pins Reason's contract: distinct
// ids in first-seen order, safe under concurrent calls, a copy out of IDs,
// and a no-op on an Env with no collector.
func TestReasonKeepsDistinctIDsInFirstSeenOrder(t *testing.T) {
	t.Parallel()
	r := &Reasons{}
	env := Env{Reasons: r}
	Reason(env, "b")
	Reason(env, "a")
	Reason(env, "b")
	if got := r.IDs(); !slices.Equal(got, []string{"b", "a"}) {
		t.Fatalf("IDs() = %q, want [b a]", got)
	}
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Reason(env, "c")
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
	Reason(Env{}, "x")
}
