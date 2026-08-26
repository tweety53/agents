package reconcile_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/reconcile"
	"github.com/tweety53/agents/stats/internal/store"
)

// journalPath mirrors fallback.JournalFilePath's own "<root>/<project>/<name>.journal"
// shape, built directly against a test's own root rather than through
// fallback.StateRoot()/FLOW_STATE_DIR -- reconcile.New already takes an
// explicit root, so these tests never need to touch that env var.
func journalPath(root, project, name string) string {
	return filepath.Join(root, project, name+".journal")
}

// changeBody builds a wire-shaped PUT/journal-entry body for state/name at
// updatedAt, carrying mainCheckoutPath so a real store.PutChange can
// bootstrap the project row on the first write for that project.
func changeBody(t *testing.T, state, updatedAt, updatedBy string) []byte {
	t.Helper()
	body := fmt.Sprintf(
		`{"state":%q,"mainCheckoutPath":"/tmp/reconcile-test","updatedAt":%q,"updatedBy":%q}`,
		state, updatedAt, updatedBy,
	)
	return []byte(body)
}

func appendEntry(t *testing.T, root, project, name, state, updatedAt, updatedBy string) {
	t.Helper()
	path := journalPath(root, project, name)
	body := changeBody(t, state, updatedAt, updatedBy)
	if err := fallback.AppendJournalEntry(path, project, name, body, time.Now()); err != nil {
		t.Fatalf("append journal entry: %v", err)
	}
}

func pendingCount(t *testing.T, root, project, name string) int {
	t.Helper()
	entries, err := fallback.ReadJournalEntries(journalPath(root, project, name))
	if err != nil {
		t.Fatalf("read journal entries: %v", err)
	}
	return len(entries)
}

// TestReplayAppliesPendingEntries pins the basic replay path against a
// real store: a single pending journal entry is applied and retired.
func TestReplayAppliesPendingEntries(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	root := t.TempDir()

	appendEntry(t, root, "proj-apply", "chg-apply", "STARTED", "2026-08-13T10:00:00Z", "flow-start")

	rec := reconcile.New(st, st, st, root, nil)
	result, err := rec.Run(ctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Applied != 1 || result.Refused != 0 || result.Journals != 1 {
		t.Fatalf("Run result = %+v, want {Journals:1 Applied:1 Refused:0}", result)
	}

	got, err := st.GetChange(ctx, "proj-apply", "chg-apply")
	if err != nil {
		t.Fatalf("GetChange: %v", err)
	}
	if got.State != store.StateStarted || got.UpdatedBy != "flow-start" {
		t.Fatalf("stored change = %+v, want state STARTED, updatedBy flow-start", got)
	}

	if n := pendingCount(t, root, "proj-apply", "chg-apply"); n != 0 {
		t.Fatalf("pending entries after replay = %d, want 0", n)
	}
}

// TestStaleEntryCannotRegressFinished pins the scenario named in
// specs/myflow-state-store/spec.md: a journal entry recording an earlier
// state than the one already stored as FINISHED must not move the record
// backwards, and must still be retired rather than retried forever.
//
// Mutation check performed by hand: with retirePrefix's call in replayFile
// commented out, this test still passes (the store row is correctly
// unchanged either way) but pendingCount's assertion below fails --
// confirming that assertion, not the GetChange one, is what actually pins
// "retired rather than retried indefinitely".
func TestStaleEntryCannotRegressFinished(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	root := t.TempDir()

	finished := store.Change{
		ProjectKey:       "proj-stale",
		MainCheckoutPath: "/tmp/reconcile-test",
		Name:             "chg-stale",
		State:            store.StateFinished,
		UpdatedAt:        time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC),
		UpdatedBy:        "flow-finish",
	}
	if err := st.PutChange(ctx, finished); err != nil {
		t.Fatalf("seed FINISHED change: %v", err)
	}

	// A stale journal entry: an earlier pipeline state, timestamped after
	// the FINISHED write -- the monotonic rule's state dimension is what
	// must refuse this, not the timestamp.
	appendEntry(t, root, "proj-stale", "chg-stale", "IN_PROGRESS", "2026-08-13T13:00:00Z", "flow-do")

	rec := reconcile.New(st, st, st, root, nil)
	result, err := rec.Run(ctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Applied != 0 || result.Refused != 1 {
		t.Fatalf("Run result = %+v, want {Applied:0 Refused:1}", result)
	}

	got, err := st.GetChange(ctx, "proj-stale", "chg-stale")
	if err != nil {
		t.Fatalf("GetChange: %v", err)
	}
	if got.State != store.StateFinished || got.UpdatedBy != "flow-finish" {
		t.Fatalf("stored change regressed: got %+v, want the FINISHED record unchanged", got)
	}

	if n := pendingCount(t, root, "proj-stale", "chg-stale"); n != 0 {
		t.Fatalf("pending entries after replay = %d, want 0 (stale entry must be retired, not retried forever)", n)
	}
}

// TestReplayRetiresRefusedEntries broadens the refusal scenario beyond a
// single entry: a journal with one entry that legitimately advances the
// record and a second that is superseded (a benign duplicate of the
// first) both retire, and the store ends up holding the correctly
// advanced record.
func TestReplayRetiresRefusedEntries(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	root := t.TempDir()

	appendEntry(t, root, "proj-refuse", "chg-refuse", "IN_PROGRESS", "2026-08-13T10:00:00Z", "flow-do")
	// A byte-identical shape of the same write (same state, same
	// updatedAt) -- exactly the "benign duplicate" shape put()'s own doc
	// comment describes, e.g. a retried request whose first response was
	// lost.
	appendEntry(t, root, "proj-refuse", "chg-refuse", "IN_PROGRESS", "2026-08-13T10:00:00Z", "flow-do")

	rec := reconcile.New(st, st, st, root, nil)
	result, err := rec.Run(ctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Applied != 1 || result.Refused != 1 {
		t.Fatalf("Run result = %+v, want {Applied:1 Refused:1}", result)
	}

	got, err := st.GetChange(ctx, "proj-refuse", "chg-refuse")
	if err != nil {
		t.Fatalf("GetChange: %v", err)
	}
	if got.State != store.StateInProgress {
		t.Fatalf("stored change = %+v, want state IN_PROGRESS", got)
	}

	if n := pendingCount(t, root, "proj-refuse", "chg-refuse"); n != 0 {
		t.Fatalf("pending entries after replay = %d, want 0", n)
	}
}

// TestReplayRetiresInvalidStateEntryAndAppliesEntryBehindIt pins F1's fix:
// before api.IsDefinitiveChangeOutcome existed, replayFile's switch checked
// only errors.Is(putErr, store.ErrMonotonicViolation), so a PutChange
// refusal for any other reason -- an invalid state value included -- fell
// into the "unknown outcome" branch and stopped the whole file there. That
// permanently blocked every entry queued behind it, on every future replay,
// because retrying the identical bad entry never makes it valid.
//
// Two entries land in the same journal file (same project/name): the first
// names a state PutChange rejects with store.ErrInvalidState, the second is
// an ordinary valid write. Both must retire, and the second's state must be
// the one actually stored.
//
// Mutation check performed by hand: reverting replayFile's
// api.IsDefinitiveChangeOutcome(putErr) case to the bare
// errors.Is(putErr, store.ErrMonotonicViolation) it replaced makes this
// test fail -- Applied drops to 0 and the second entry is left pending,
// which is exactly F1's bug.
func TestReplayRetiresInvalidStateEntryAndAppliesEntryBehindIt(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	root := t.TempDir()

	appendEntry(t, root, "proj-badstate", "chg-badstate", "BOGUS_STATE", "2026-08-13T10:00:00Z", "flow-do")
	appendEntry(t, root, "proj-badstate", "chg-badstate", "IN_PROGRESS", "2026-08-13T10:05:00Z", "flow-do")

	rec := reconcile.New(st, st, st, root, nil)
	result, err := rec.Run(ctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Applied != 1 || result.Refused != 1 {
		t.Fatalf("Run result = %+v, want {Applied:1 Refused:1} -- an invalid-state entry must retire, not block the valid entry behind it", result)
	}

	got, err := st.GetChange(ctx, "proj-badstate", "chg-badstate")
	if err != nil {
		t.Fatalf("GetChange: %v", err)
	}
	if got.State != store.StateInProgress {
		t.Fatalf("stored change = %+v, want state IN_PROGRESS (the valid entry queued behind the bad one)", got)
	}

	if n := pendingCount(t, root, "proj-badstate", "chg-badstate"); n != 0 {
		t.Fatalf("pending entries after replay = %d, want 0 (both entries retired)", n)
	}
}

// TestReplayRetiresUndecodableEntryAndAppliesEntryBehindIt is F1's other
// reproduction: a journal entry whose body carries a field
// api.DecodeChangeBody's DisallowUnknownFields rejects will fail to decode
// identically on every future replay, so it must retire rather than block
// the file -- exactly like an invalid-state PutChange refusal, just caught
// one step earlier.
//
// Mutation check performed by hand: reverting the decode-failure branch to
// `break entryLoop` (its pre-fix shape) makes this test fail the same way
// TestReplayRetiresInvalidStateEntryAndAppliesEntryBehindIt does -- Applied
// drops to 0 and the valid second entry never applies.
func TestReplayRetiresUndecodableEntryAndAppliesEntryBehindIt(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	root := t.TempDir()

	path := journalPath(root, "proj-undecodable", "chg-undecodable")
	undecodable := []byte(`{"state":"STARTED","bogusField":"x","mainCheckoutPath":"/tmp/reconcile-test","updatedAt":"2026-08-13T10:00:00Z","updatedBy":"flow-do"}`)
	if err := fallback.AppendJournalEntry(path, "proj-undecodable", "chg-undecodable", undecodable, time.Now()); err != nil {
		t.Fatalf("append undecodable journal entry: %v", err)
	}
	appendEntry(t, root, "proj-undecodable", "chg-undecodable", "IN_PROGRESS", "2026-08-13T10:05:00Z", "flow-do")

	rec := reconcile.New(st, st, st, root, nil)
	result, err := rec.Run(ctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Applied != 1 || result.Refused != 1 {
		t.Fatalf("Run result = %+v, want {Applied:1 Refused:1} -- an undecodable entry must retire, not block the valid entry behind it", result)
	}

	got, err := st.GetChange(ctx, "proj-undecodable", "chg-undecodable")
	if err != nil {
		t.Fatalf("GetChange: %v", err)
	}
	if got.State != store.StateInProgress {
		t.Fatalf("stored change = %+v, want state IN_PROGRESS (the valid entry queued behind the undecodable one)", got)
	}

	if n := pendingCount(t, root, "proj-undecodable", "chg-undecodable"); n != 0 {
		t.Fatalf("pending entries after replay = %d, want 0 (both entries retired)", n)
	}
}

// TestInterruptedReplayResumesWithoutDuplicating exercises the requirement
// spec.md names as "Replay is interrupted": a transport-shaped failure on
// the second of two pending entries must leave that entry (and only that
// entry) in the journal, and a second Run -- once the failure clears --
// must apply it without re-applying the first.
//
// injectingStore wraps the real store, forwarding every call except one
// chosen (project, name) pair, which fails with a plain, non-monotonic
// error exactly once to simulate a dropped connection mid-replay.
func TestInterruptedReplayResumesWithoutDuplicating(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	root := t.TempDir()

	appendEntry(t, root, "proj-interrupt", "chg-a", "STARTED", "2026-08-13T10:00:00Z", "flow-start")
	appendEntry(t, root, "proj-interrupt", "chg-b", "STARTED", "2026-08-13T10:00:00Z", "flow-start")

	wrapped := &injectingStore{inner: st, failName: "chg-b", failsRemaining: 1}
	rec := reconcile.New(wrapped, nopStageStore{}, nopRecordStore{}, root, nil)

	result, err := rec.Run(ctx)
	if err != nil {
		t.Fatalf("Run (interrupted): %v", err)
	}
	if result.Applied != 1 {
		t.Fatalf("Run (interrupted) result = %+v, want Applied:1 (only chg-a)", result)
	}
	if n := pendingCount(t, root, "proj-interrupt", "chg-a"); n != 0 {
		t.Fatalf("chg-a pending entries = %d, want 0 (accepted, must be retired)", n)
	}
	if n := pendingCount(t, root, "proj-interrupt", "chg-b"); n != 1 {
		t.Fatalf("chg-b pending entries = %d, want 1 (unresolved, must remain)", n)
	}
	if calls := wrapped.callsFor("chg-a"); calls != 1 {
		t.Fatalf("chg-a was sent to the store %d times, want exactly 1", calls)
	}

	// Second Run, failure cleared: chg-b applies, chg-a is never sent to
	// the store again (it is no longer in the journal to replay).
	result2, err := rec.Run(ctx)
	if err != nil {
		t.Fatalf("Run (resumed): %v", err)
	}
	if result2.Applied != 1 {
		t.Fatalf("Run (resumed) result = %+v, want Applied:1 (chg-b)", result2)
	}
	if calls := wrapped.callsFor("chg-a"); calls != 1 {
		t.Fatalf("chg-a was sent to the store %d times across both runs, want exactly 1 (no duplicate application)", calls)
	}
	// chg-b is legitimately sent to the wrapper twice: the first attempt
	// (injected failure) and the second, successful one on resume -- the
	// invariant this test protects is that it reaches the *underlying*
	// store's PutChange, and therefore actually applies, exactly once.
	if calls := wrapped.callsFor("chg-b"); calls != 2 {
		t.Fatalf("chg-b was sent to the wrapper %d times, want exactly 2 (1 failed attempt + 1 resumed attempt)", calls)
	}
	if forwarded := wrapped.forwardedCountFor("chg-b"); forwarded != 1 {
		t.Fatalf("chg-b reached the underlying store %d times, want exactly 1 (no duplicate application)", forwarded)
	}

	gotA, err := st.GetChange(ctx, "proj-interrupt", "chg-a")
	if err != nil {
		t.Fatalf("GetChange chg-a: %v", err)
	}
	if gotA.State != store.StateStarted {
		t.Fatalf("chg-a state = %v, want STARTED", gotA.State)
	}
	gotB, err := st.GetChange(ctx, "proj-interrupt", "chg-b")
	if err != nil {
		t.Fatalf("GetChange chg-b: %v", err)
	}
	if gotB.State != store.StateStarted {
		t.Fatalf("chg-b state = %v, want STARTED", gotB.State)
	}
}

// injectingStore forwards to inner, except that it fails the first
// failsRemaining calls naming failName with a plain, non-monotonic error.
type injectingStore struct {
	inner reconcile.ChangeStore

	mu             sync.Mutex
	failName       string
	failsRemaining int
	calls          []store.Change
	forwarded      []store.Change
}

var errInjectedTransportFailure = errors.New("injected transport failure")

func (s *injectingStore) PutChange(ctx context.Context, c store.Change) error {
	s.mu.Lock()
	s.calls = append(s.calls, c)
	shouldFail := c.Name == s.failName && s.failsRemaining > 0
	if shouldFail {
		s.failsRemaining--
	}
	s.mu.Unlock()

	if shouldFail {
		return errInjectedTransportFailure
	}

	err := s.inner.PutChange(ctx, c)
	if err == nil {
		s.mu.Lock()
		s.forwarded = append(s.forwarded, c)
		s.mu.Unlock()
	}
	return err
}

func (s *injectingStore) forwardedCountFor(name string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, c := range s.forwarded {
		if c.Name == name {
			n++
		}
	}
	return n
}

func (s *injectingStore) callsFor(name string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, c := range s.calls {
		if c.Name == name {
			n++
		}
	}
	return n
}

// fakeStore is an in-memory ChangeStore for the tests below that exercise
// replay's file mechanics rather than the store's own monotonic guarantee
// -- concurrent Run calls, an append landing mid-replay, and a partial
// trailing journal line. hook, when set, runs on every PutChange call and
// may mutate external state (e.g. append to the journal file being
// replayed) or return an error.
type fakeStore struct {
	mu    sync.Mutex
	calls []store.Change
	hook  func(store.Change) error

	active    int32
	maxActive int32
}

func (f *fakeStore) PutChange(_ context.Context, c store.Change) error {
	n := atomic.AddInt32(&f.active, 1)
	defer atomic.AddInt32(&f.active, -1)
	for {
		old := atomic.LoadInt32(&f.maxActive)
		if n <= old {
			break
		}
		if atomic.CompareAndSwapInt32(&f.maxActive, old, n) {
			break
		}
	}

	f.mu.Lock()
	f.calls = append(f.calls, c)
	hook := f.hook
	f.mu.Unlock()

	if hook != nil {
		return hook(c)
	}
	return nil
}

func (f *fakeStore) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

// TestReconcilerRunSerializesConcurrentCalls exercises "replay running
// twice concurrently" -- a startup replay and a reconnect-triggered replay
// firing at once against the same Reconciler. Both calls are made to Run
// concurrently; the store hook sleeps briefly to widen the race window a
// missing lock would need to be caught by. If Run did not serialize
// through its own mutex, two PutChange calls from the two concurrent Run
// invocations could execute at the same time (fakeStore.maxActive would
// observe more than 1); with the mutex, every PutChange call the fakeStore
// ever sees is strictly ordered.
//
// Mutation check performed by hand: commenting out Reconciler.Run's
// r.mu.Lock()/Unlock() makes this test fail intermittently (maxActive
// observed as 2) under `go test -race -count=20`, confirming the mutex is
// what the assertion actually depends on.
func TestReconcilerRunSerializesConcurrentCalls(t *testing.T) {
	root := t.TempDir()
	appendEntry(t, root, "proj-concurrent", "chg-1", "STARTED", "2026-08-13T10:00:00Z", "flow-start")
	appendEntry(t, root, "proj-concurrent", "chg-2", "STARTED", "2026-08-13T10:00:00Z", "flow-start")

	fs := &fakeStore{hook: func(store.Change) error {
		time.Sleep(20 * time.Millisecond)
		return nil
	}}
	rec := reconcile.New(fs, nopStageStore{}, nopRecordStore{}, root, nil)

	var wg sync.WaitGroup
	results := make([]reconcile.Result, 2)
	errs := make([]error, 2)
	for i := range 2 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = rec.Run(context.Background())
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("Run[%d]: %v", i, err)
		}
	}
	if atomic.LoadInt32(&fs.maxActive) > 1 {
		t.Fatalf("observed %d concurrent PutChange calls, want at most 1 -- Run's serialization did not hold", fs.maxActive)
	}
	totalApplied := results[0].Applied + results[1].Applied
	if totalApplied != 2 {
		t.Fatalf("total Applied across both Run calls = %d, want 2 (each entry applied exactly once)", totalApplied)
	}
	if fs.callCount() != 2 {
		t.Fatalf("store saw %d PutChange calls, want 2 (no duplicate application)", fs.callCount())
	}
}

// TestRetirePreservesEntryAppendedDuringReplay exercises "the journal
// being appended to while replay reads it": the store hook, invoked while
// Run is processing the journal's one pre-existing entry, appends a
// *second* entry for the *same* project/name -- i.e. to the very same
// journal file replayFile already has a stale snapshot (raw) of -- mid
// call, simulating a CLI in another worktree writing a follow-on state for
// this change at the exact moment the daemon is replaying its journal.
//
// The freshly appended entry must not be lost: it was never part of the
// snapshot this Run's loop is iterating (Run does not re-scan mid-file),
// so it is correctly left pending rather than applied in this same pass --
// but retirePrefix must preserve it rather than discard it when it removes
// the entry that was actually processed. A second Run then picks it up.
func TestRetirePreservesEntryAppendedDuringReplay(t *testing.T) {
	root := t.TempDir()
	appendEntry(t, root, "proj-append", "chg-x", "STARTED", "2026-08-13T10:00:00Z", "flow-start")

	var once sync.Once
	fs := &fakeStore{}
	fs.hook = func(c store.Change) error {
		once.Do(func() {
			// Same project/name as the entry currently being processed --
			// this append lands in the exact file replayFile's raw is a
			// now-stale snapshot of, not a different one.
			appendEntry(t, root, "proj-append", "chg-x", "IN_PROGRESS", "2026-08-13T11:00:00Z", "flow-do")
		})
		return nil
	}

	rec := reconcile.New(fs, nopStageStore{}, nopRecordStore{}, root, nil)
	result, err := rec.Run(context.Background())
	if err != nil {
		t.Fatalf("Run (first pass): %v", err)
	}
	if result.Applied != 1 {
		t.Fatalf("Run (first pass) Applied = %d, want 1 (only the entry present when raw was read)", result.Applied)
	}

	if n := pendingCount(t, root, "proj-append", "chg-x"); n != 1 {
		t.Fatalf("chg-x pending entries after first pass = %d, want 1 (the entry appended mid-replay must survive retirement of the one processed)", n)
	}

	result2, err := rec.Run(context.Background())
	if err != nil {
		t.Fatalf("Run (second pass): %v", err)
	}
	if result2.Applied != 1 {
		t.Fatalf("Run (second pass) Applied = %d, want 1 (the appended entry now picked up)", result2.Applied)
	}
	if fs.callCount() != 2 {
		t.Fatalf("store saw %d PutChange calls across both passes, want 2", fs.callCount())
	}
}

// TestPartialTrailingJournalLineIsIgnored exercises the reachable-but-rare
// crash-mid-write case: a journal file whose final bytes have no trailing
// newline, as fallback.AppendJournalEntry's single f.Write call would
// leave it if the process died partway through that write. Replay must
// apply and retire the complete entry before it, and must leave the
// partial tail exactly as it found it -- neither guessing at it nor
// erroring the whole file out.
func TestPartialTrailingJournalLineIsIgnored(t *testing.T) {
	root := t.TempDir()
	path := journalPath(root, "proj-partial", "chg-partial")

	complete := fmt.Sprintf(
		`{"recordedAt":"2026-08-13T10:00:00Z","project":"proj-partial","name":"chg-partial","body":%s}`+"\n",
		mustMarshalRawBody(t, changeBody(t, "STARTED", "2026-08-13T10:00:00Z", "flow-start")),
	)
	partialTail := `{"recordedAt":"2026-08-13T11:00:00Z","project":"proj-partial","na`

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(complete+partialTail), 0o644); err != nil {
		t.Fatalf("write journal fixture: %v", err)
	}

	fs := &fakeStore{}
	rec := reconcile.New(fs, nopStageStore{}, nopRecordStore{}, root, nil)
	result, err := rec.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Applied != 1 {
		t.Fatalf("Applied = %d, want 1 (the one complete line)", result.Applied)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read journal after replay: %v", err)
	}
	if !bytes.Equal(got, []byte(partialTail)) {
		t.Fatalf("journal after replay = %q, want exactly the untouched partial tail %q", got, partialTail)
	}
}

func mustMarshalRawBody(t *testing.T, body []byte) string {
	t.Helper()
	// The journal's "body" field is itself a JSON value (json.RawMessage in
	// fallback.Entry) -- since body is already well-formed JSON bytes, it
	// can be embedded verbatim; this only exists to make that explicit at
	// the call site above rather than hand-splicing raw bytes into a
	// %s-formatted string quietly.
	return string(body)
}

// TestConcurrentAppendVersusRetirePreservesEveryEntry is F1's reproducer,
// scaled for a test suite rather than a one-off stress run: a tight-loop
// appender writes appendCount entries to one journal file for the same
// (project, name) while a second goroutine hammers that same file with
// back-to-back Reconciler.Run calls -- no artificial delay on either side,
// to maximize how often a retire's rename lands inside an in-flight
// append's open-then-write window.
//
// Without fallback.LockJournal/retirePrefix's flock, this loses a large
// fraction of entries silently (the original, unmitigated F1 reproducer:
// 20,000 appends racing a tight retire loop lost ~71%, with no error
// anywhere -- a lost append here does not fail, it just never reaches the
// store). With the lock in place, every appended entry -- identified by
// its own unique UpdatedBy value -- must eventually reach the store
// exactly once, with zero lost and zero append or replay errors.
//
// Mutation check performed by hand (see this task's report): commenting
// out the fallback.LockJournal call in retirePrefix reproduces real loss
// under this exact test, confirming the assertion below actually depends
// on the lock rather than passing regardless of it.
func TestConcurrentAppendVersusRetirePreservesEveryEntry(t *testing.T) {
	const appendCount = 600

	root := t.TempDir()
	path := journalPath(root, "race-proj", "race-chg")

	fs := &fakeStore{}
	seen := make(map[string]bool, appendCount)
	var seenMu sync.Mutex
	fs.hook = func(c store.Change) error {
		seenMu.Lock()
		seen[c.UpdatedBy] = true
		seenMu.Unlock()
		return nil
	}

	rec := reconcile.New(fs, nopStageStore{}, nopRecordStore{}, root, nil)

	stopRetirer := make(chan struct{})
	var retirerErrs []error
	var retirerMu sync.Mutex
	var retirerWG sync.WaitGroup
	retirerWG.Add(1)
	go func() {
		defer retirerWG.Done()
		for {
			select {
			case <-stopRetirer:
				return
			default:
			}
			if _, err := rec.Run(context.Background()); err != nil {
				retirerMu.Lock()
				retirerErrs = append(retirerErrs, err)
				retirerMu.Unlock()
			}
			// A small gap between retire passes, not a realistic pacing
			// choice (production callers are far less aggressive: a
			// reconnect-triggered replay fires once per
			// reconnectPingInterval, not back to back) but a deliberate
			// one for this test: without it, the retirer holds the
			// sidecar lock for very close to 100% of wall-clock time
			// against 600 back-to-back appends, which pushes occasional
			// appends past even a generous bound purely on scheduler/GC
			// tail latency rather than on the race this test exists to
			// catch. This keeps duty cycle high enough to still exercise
			// the race hard (the mutation check below still reliably
			// loses hundreds of entries with the lock removed) while
			// giving appends a realistic chance to interleave.
			time.Sleep(200 * time.Microsecond)
		}
	}()

	var appendErrs []error
	for i := 0; i < appendCount; i++ {
		body := []byte(fmt.Sprintf(
			`{"state":"IN_PROGRESS","mainCheckoutPath":"/tmp/race","updatedAt":"2026-08-13T10:00:00Z","updatedBy":"w%06d"}`, i,
		))
		if err := fallback.AppendJournalEntry(path, "race-proj", "race-chg", body, time.Now()); err != nil {
			appendErrs = append(appendErrs, err)
		}
	}

	// Give the retirer goroutine a bounded window to drain whatever is
	// still pending once appending is done, then stop it and do a final,
	// single-threaded drain pass to mop up anything left (the tight
	// retirer loop above has no reason to leave entries behind once
	// appending stops, but this keeps the test deterministic rather than
	// racing its own teardown).
	drainDeadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(drainDeadline) {
		entries, err := fallback.ReadJournalEntries(path)
		if err != nil {
			t.Fatalf("read journal entries while draining: %v", err)
		}
		if len(entries) == 0 {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}
	close(stopRetirer)
	retirerWG.Wait()

	for range 20 {
		entries, err := fallback.ReadJournalEntries(path)
		if err != nil {
			t.Fatalf("read journal entries during final drain: %v", err)
		}
		if len(entries) == 0 {
			break
		}
		if _, err := rec.Run(context.Background()); err != nil {
			t.Fatalf("final drain Run: %v", err)
		}
	}

	if len(appendErrs) > 0 {
		t.Fatalf("%d/%d appends failed, first error: %v", len(appendErrs), appendCount, appendErrs[0])
	}
	if len(retirerErrs) > 0 {
		t.Fatalf("%d Run calls returned an error, first: %v", len(retirerErrs), retirerErrs[0])
	}

	seenMu.Lock()
	gotN := len(seen)
	seenMu.Unlock()
	if gotN != appendCount {
		t.Fatalf("store observed %d of %d appended entries -- %d lost to the append-vs-retire race",
			gotN, appendCount, appendCount-gotN)
	}
}

// TestReplayRetiresMalformedMergeBaseEntryAndAppliesEntryBehindIt is the
// journal half of "A recorded merge base is a sha or nothing". `flow
// state set` refuses a malformed merge base before it can ever be
// journalled, so the only way one reaches replay is a hand-edited or
// out-of-band-modified fallback file -- the case
// skills/myflow-contracts/state-file.md already names.
//
// Such an entry is unfixable by retrying: the bytes already on disk
// produce the identical store.ErrInvalidMergeBase refusal on every future
// pass. It must therefore retire, exactly as an invalid-state entry does,
// rather than block every entry queued behind it forever. That is what
// makes store.ErrInvalidMergeBase's case in api.IsDefinitiveChangeOutcome
// load-bearing rather than decorative: without it the refusal falls into
// replayFile's "unknown outcome" branch and stops the file there, which is
// F1's bug reproduced by a different malformed field.
func TestReplayRetiresMalformedMergeBaseEntryAndAppliesEntryBehindIt(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	root := t.TempDir()

	const project, name = "proj-badbase", "chg-badbase"

	// A worktree path written into the merge-base position -- the value
	// KAN-265 recorded and could not correct.
	handEdited := []byte(`{"state":"IN_PROGRESS","mainCheckoutPath":"/tmp/reconcile-test",` +
		`"worktrees":{"/w/kan-265":"/Users/x/Projects/agents-worktrees/kan-265"},` +
		`"updatedAt":"2026-08-13T10:00:00Z","updatedBy":"flow-do"}`)
	if err := fallback.AppendJournalEntry(journalPath(root, project, name), project, name, handEdited, time.Now()); err != nil {
		t.Fatalf("append hand-edited journal entry: %v", err)
	}
	appendEntry(t, root, project, name, "IN_PROGRESS", "2026-08-13T10:05:00Z", "flow-fix")

	rec := reconcile.New(st, st, st, root, nil)
	result, err := rec.Run(ctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Applied != 1 || result.Refused != 1 {
		t.Fatalf("Run result = %+v, want {Applied:1 Refused:1} -- a hand-edited merge base must retire, not block the valid entry behind it", result)
	}

	got, err := st.GetChange(ctx, project, name)
	if err != nil {
		t.Fatalf("GetChange: %v", err)
	}
	if got.UpdatedBy != "flow-fix" {
		t.Fatalf("stored change = %+v, want the valid entry queued behind the bad one", got)
	}

	if n := pendingCount(t, root, project, name); n != 0 {
		t.Fatalf("pending entries after replay = %d, want 0 (both entries retired)", n)
	}
}
