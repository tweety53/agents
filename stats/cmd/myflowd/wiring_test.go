package main

import (
	"testing"

	"github.com/tweety53/agents/stats/internal/harvest"
	"github.com/tweety53/agents/stats/internal/store"
)

// TestNewTranscriptWatcherWiresBinderAndPricer is KAN-172 task 7's own
// deliverable (tasks.md, "the test is the deliverable, not step 1"): it
// asserts the *constructed* Watcher carries both a session-token binder
// and a pricer, never main.go's source text -- a test that merely grepped
// for "WithSessionTokenBinder" would still pass for a refactor that kept
// the call site's text but dropped its effect.
//
// Run against the unfixed newTranscriptWatcher (harvest.WithPricer(st)
// only, no harvest.WithSessionTokenBinder(st)), this test failed:
//
//	watcher has no session-token binder: cmd/myflowd built it without
//	harvest.WithSessionTokenBinder -- no stage run can ever be bound
//
// which is exactly the live defect this task fixes: a real mark recorded
// session_token mf-k172-live-7f3a91c, the token was present in its own
// session transcript, the daemon harvested that transcript to EOF, and
// the stage run stayed unbound because pendingSessionTokens
// (internal/harvest/watcher.go) always returned nil with no binder
// configured.
func TestNewTranscriptWatcherWiresBinderAndPricer(t *testing.T) {
	var st *store.Store // never dereferenced: newTranscriptWatcher only stores it behind the HarvestSink/Pricer/SessionTokenBinder interfaces, none of whose methods this test calls.
	attributor := harvest.NewAttributor(nil)

	w := newTranscriptWatcher(t.TempDir(), st, attributor, nil)

	if !w.HasPricer() {
		t.Error("watcher has no pricer: cmd/myflowd built it without harvest.WithPricer -- no stage run would ever be priced")
	}
	if !w.HasSessionTokenBinder() {
		t.Error("watcher has no session-token binder: cmd/myflowd built it without harvest.WithSessionTokenBinder -- no stage run can ever be bound")
	}
	// KAN-258 extends this test to the second attribution pass rather than
	// writing a second one: the defect it guards against is identical --
	// a Watcher constructed without harvest.WithDispatchAttribution
	// harvests every transcript exactly as before and leaves every
	// dispatch's metrics bag empty forever, with nothing failing.
	if !w.HasDispatchAttribution() {
		t.Error("watcher has no dispatch attribution: cmd/myflowd built it without harvest.WithDispatchAttribution -- no dispatch would ever be charged for the usage it caused")
	}
}
