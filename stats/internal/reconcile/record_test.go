package reconcile_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/reconcile"
	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// nopRecordStore satisfies api.RecordWriter with nothing but errors -- for
// the tests that exercise the state or stage journal only and never touch a
// record write, now that reconcile.New requires a record writer alongside
// the other two stores. Mirrors nopStageStore, which exists for the
// identical reason on the stage side.
type nopRecordStore struct{}

var errRecordStoreNotExercised = errors.New("nopRecordStore: record writes are not exercised by this test")

func (nopRecordStore) RecordDispatch(context.Context, string, string, records.Dispatch) (records.Dispatch, error) {
	return records.Dispatch{}, errRecordStoreNotExercised
}

func (nopRecordStore) EndDispatch(context.Context, string, string, records.DispatchEnd) (records.Dispatch, error) {
	return records.Dispatch{}, errRecordStoreNotExercised
}

func (nopRecordStore) UpsertFinding(context.Context, string, string, records.Finding) (records.Finding, bool, error) {
	return records.Finding{}, false, errRecordStoreNotExercised
}

func (nopRecordStore) SetFindingStatus(context.Context, string, string, string, string) error {
	return errRecordStoreNotExercised
}

var _ api.RecordWriter = nopRecordStore{}

// recordJournalPath mirrors cmd/myflow/record.go's own recordJournalPath
// (the state journal path with ".record" appended) -- reproduced here
// rather than imported, for the same reason stageJournalPath is: record.go
// lives in a main package, and nothing may import one. See reconcile.go's
// recordJournalBody doc comment for the boundary this crosses and why it
// is a literal convention rather than a shared constant.
func recordJournalPath(root, project, name string) string {
	return journalPath(root, project, name) + ".record"
}

// recordEnvelope is cmd/myflow/record.go's own recordJournalBody shape,
// reproduced here for the same reason recordJournalPath is: no import
// across the main-package boundary.
type recordEnvelope struct {
	Kind    string `json:"kind"`
	Request any    `json:"request"`
}

// recordStatusRequest is cmd/myflow/record.go's own recordStatusRequest:
// the ref the wire PATCH carries in its URL, alongside the status its body
// carries, so a journalled status write is replayable from what it holds
// rather than from a route this test would have to encode a second time.
type recordStatusRequest struct {
	Ref    string `json:"ref"`
	Status string `json:"status"`
}

// appendRecordWrite journals kind/req exactly as cmd/myflow/record.go's
// journalRecordWrite does -- the same envelope, the same
// fallback.AppendJournalEntry call, at the same suffixed path -- so a test
// using this exercises reconcile's real decoder against the real shape the
// CLI writes, not a hand-simplified stand-in for it. req is a
// records.Dispatch, a records.Finding or a recordStatusRequest, which is
// the actual shared contract between the two sides; only the envelope and
// the path suffix are duplicated conventions.
func appendRecordWrite(t *testing.T, root, project, name, kind string, req any) {
	t.Helper()
	body, err := json.Marshal(recordEnvelope{Kind: kind, Request: req})
	if err != nil {
		t.Fatalf("marshal record journal body: %v", err)
	}
	if err := fallback.AppendJournalEntry(recordJournalPath(root, project, name), project, name, body, time.Now()); err != nil {
		t.Fatalf("append record journal entry: %v", err)
	}
}

func pendingRecordCount(t *testing.T, root, project, name string) int {
	t.Helper()
	entries, err := fallback.ReadJournalEntries(recordJournalPath(root, project, name))
	if err != nil {
		t.Fatalf("read record journal entries: %v", err)
	}
	return len(entries)
}

// fakeRecordStore is an in-memory api.RecordWriter that appends one
// short description per call, in the order the calls arrive. File order is
// the whole point of a replay -- a dispatch's seq is allocated in the order
// the store sees it, and a status write that overtook the finding it
// updates would be refused -- so the assertion these tests make is on the
// sequence, not on a set.
//
// It is in-memory rather than a real store because these tests are about
// replay's own mechanics: which files are walked, which entries decode,
// which outcomes retire. internal/store's own tests already cover what each
// of these three methods does against a real PostgreSQL.
type fakeRecordStore struct {
	mu      sync.Mutex
	applied []string
}

var _ api.RecordWriter = (*fakeRecordStore)(nil)

func (f *fakeRecordStore) record(desc string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.applied = append(f.applied, desc)
}

func (f *fakeRecordStore) RecordDispatch(_ context.Context, projectKey, change string, in records.Dispatch) (records.Dispatch, error) {
	f.record(fmt.Sprintf("dispatch %s/%s task=%s role=%s model=%s", projectKey, change, in.TaskID, in.Role, in.Model))
	return in, nil
}

func (f *fakeRecordStore) EndDispatch(_ context.Context, projectKey, change string, in records.DispatchEnd) (records.Dispatch, error) {
	f.record(fmt.Sprintf("dispatch-end %s/%s key=%s commit=%s outcome=%s", projectKey, change, in.Key, in.CommitSHA, in.Outcome))
	return records.Dispatch{Key: in.Key, CommitSHA: in.CommitSHA, Outcome: in.Outcome, EndedAt: &in.EndedAt}, nil
}

func (f *fakeRecordStore) UpsertFinding(_ context.Context, projectKey, change string, in records.Finding) (records.Finding, bool, error) {
	f.record(fmt.Sprintf("finding %s/%s ref=%s status=%s", projectKey, change, in.Ref, in.Status))
	return in, true, nil
}

func (f *fakeRecordStore) SetFindingStatus(_ context.Context, projectKey, change, ref, status string) error {
	f.record(fmt.Sprintf("status %s/%s ref=%s status=%s", projectKey, change, ref, status))
	return nil
}

func (f *fakeRecordStore) appliedCalls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.applied...)
}

func assertAppliedCalls(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("store saw %d record write(s) %q, want %d %q", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("record write %d = %q, want %q (replay must apply entries in file order)", i, got[i], want[i])
		}
	}
}

func testDispatch() records.Dispatch {
	return records.Dispatch{
		TaskID:       "7",
		Role:         "implementer",
		Model:        "opus",
		CommitSHA:    "3aa9a4a",
		Outcome:      "completed",
		SessionToken: "kan-258-record-replay",
		StartedAt:    time.Date(2026, 8, 22, 9, 0, 0, 0, time.UTC),
	}
}

func testFinding(ref, status string) records.Finding {
	return records.Finding{
		Ref:      ref,
		Round:    0,
		Slot:     "principles",
		Severity: "major",
		Location: "internal/reconcile/reconcile.go:1",
		Note:     "the record journal is never replayed",
		Status:   status,
	}
}

// testDispatchEnd is the closing half of testDispatch: the two fields that
// name the row and the three that close it.
func testDispatchEnd() records.DispatchEnd {
	return records.DispatchEnd{
		SessionToken: "kan-258-record-replay",
		Key:          "task-7-implementer",
		CommitSHA:    "3aa9a4a",
		Outcome:      "completed",
		EndedAt:      time.Date(2026, 8, 22, 9, 40, 0, 0, time.UTC),
	}
}

// --- a record write journalled while the store is unreachable is
// observable in the store after a replay ---

// TestReplayAppliesPendingRecordEntries is the delta spec's "a journalled
// record write is replayed" scenario: every entry cmd/myflow/record.go's
// fallback path can write -- a dispatch begin, a dispatch end, a finding
// and a status -- must reach the store, in the order the file holds them, and be retired.
// Before replayRecordFile existed, nothing walked "*.journal.record" at
// all, so a record write that fell back to the journal stayed on disk
// forever, which turns the never-block guarantee into a data-loss
// guarantee.
//
// Every kind is exercised in one file rather than one each, because order
// across kinds is exactly what matters: a status write replayed before the
// finding it updates would be refused by a real store, and a dispatch end
// replayed before the begin it closes names a row that does not exist yet.
func TestReplayAppliesPendingRecordEntries(t *testing.T) {
	root := t.TempDir()
	const project, change = "proj-record", "chg-record"

	appendRecordWrite(t, root, project, change, "dispatch", testDispatch())
	appendRecordWrite(t, root, project, change, "dispatch-end", testDispatchEnd())
	appendRecordWrite(t, root, project, change, "finding", testFinding("F1", "open"))
	appendRecordWrite(t, root, project, change, "status", recordStatusRequest{Ref: "F1", Status: "fixed"})

	rs := &fakeRecordStore{}
	rec := reconcile.New(&fakeStore{}, nopStageStore{}, rs, root, nil)

	result, err := rec.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Journals != 1 || result.Applied != 4 || result.Refused != 0 {
		t.Fatalf("Run result = %+v, want {Journals:1 Applied:4 Refused:0}", result)
	}

	assertAppliedCalls(t, rs.appliedCalls(), []string{
		"dispatch proj-record/chg-record task=7 role=implementer model=opus",
		"dispatch-end proj-record/chg-record key=task-7-implementer commit=3aa9a4a outcome=completed",
		"finding proj-record/chg-record ref=F1 status=open",
		"status proj-record/chg-record ref=F1 status=fixed",
	})

	if n := pendingRecordCount(t, root, project, change); n != 0 {
		t.Fatalf("pending record entries after replay = %d, want 0 (every applied entry retired)", n)
	}
}

// TestPartialTrailingRecordJournalLineIsIgnored pins that
// splitCompleteLines' guarantee holds for this third journal kind too: a
// file whose final bytes have no trailing newline -- as
// fallback.AppendJournalEntry's single Write would leave it if the process
// died partway through -- has its complete entry applied and retired, and
// its partial tail left exactly where it was.
func TestPartialTrailingRecordJournalLineIsIgnored(t *testing.T) {
	root := t.TempDir()
	const project, change = "proj-record-partial", "chg-record-partial"
	path := recordJournalPath(root, project, change)

	body, err := json.Marshal(recordEnvelope{Kind: "dispatch", Request: testDispatch()})
	if err != nil {
		t.Fatalf("marshal record journal body: %v", err)
	}
	complete := fmt.Sprintf(
		`{"recordedAt":"2026-08-22T09:00:00Z","project":%q,"name":%q,"body":%s}`+"\n",
		project, change, body,
	)
	partialTail := `{"recordedAt":"2026-08-22T10:00:00Z","project":"proj-record-partial","na`

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(complete+partialTail), 0o644); err != nil {
		t.Fatalf("write record journal fixture: %v", err)
	}

	rs := &fakeRecordStore{}
	rec := reconcile.New(&fakeStore{}, nopStageStore{}, rs, root, nil)

	result, err := rec.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Applied != 1 || result.Refused != 0 {
		t.Fatalf("Run result = %+v, want {Applied:1 Refused:0} (the one complete line)", result)
	}
	assertAppliedCalls(t, rs.appliedCalls(), []string{
		"dispatch proj-record-partial/chg-record-partial task=7 role=implementer model=opus",
	})

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read record journal after replay: %v", err)
	}
	if !bytes.Equal(got, []byte(partialTail)) {
		t.Fatalf("record journal after replay = %q, want exactly the untouched partial tail %q", got, partialTail)
	}
}

// TestUndecodableRecordEntryIsRefusedAndAppliesEntryBehindIt is the same
// shape errChangeEntryDecodeFailed already establishes for a change entry:
// a body that will never decode is retired as refused rather than left at
// the head of the file, where it would block every valid entry queued
// behind it on every future replay. One bad line must not cost the file.
func TestUndecodableRecordEntryIsRefusedAndAppliesEntryBehindIt(t *testing.T) {
	root := t.TempDir()
	const project, change = "proj-record-bad", "chg-record-bad"

	// A complete, well-formed journal line whose *body* cannot decode:
	// records.Dispatch.Role is a string, so a number there fails on every
	// replay, identically, forever.
	undecodable := []byte(`{"kind":"dispatch","request":{"role":42}}`)
	if err := fallback.AppendJournalEntry(recordJournalPath(root, project, change), project, change, undecodable, time.Now()); err != nil {
		t.Fatalf("append undecodable record journal entry: %v", err)
	}
	appendRecordWrite(t, root, project, change, "finding", testFinding("F2", "open"))

	rs := &fakeRecordStore{}
	rec := reconcile.New(&fakeStore{}, nopStageStore{}, rs, root, nil)

	result, err := rec.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Applied != 1 || result.Refused != 1 {
		t.Fatalf("Run result = %+v, want {Applied:1 Refused:1} -- an undecodable entry must retire, not block the valid entry behind it", result)
	}
	assertAppliedCalls(t, rs.appliedCalls(), []string{
		"finding proj-record-bad/chg-record-bad ref=F2 status=open",
	})

	if n := pendingRecordCount(t, root, project, change); n != 0 {
		t.Fatalf("pending record entries after replay = %d, want 0 (both entries retired)", n)
	}
}

// --- a journalled entry missing a required field is refused, not written ---

// TestRecordEntryMissingARequiredFieldIsRefusedWithoutWritingARow pins the
// one guarantee a replay that bypassed the API layer's required-field
// checks could not make: an entry whose required fields are empty must be
// refused by exactly the same rule the live HTTP handler applies, and no
// row may be written for it.
//
// The failure this closes is silent, not loud. Migration 0010's columns are
// NOT NULL, and an empty string satisfies NOT NULL -- so a hand-edited or
// corrupted ".journal.record" line carrying "role":"" reached the store,
// inserted a row with an empty role, returned nil, and was retired as a
// success. That is silent corruption of the very record this change exists
// to make trustworthy, and the exact opposite of the delta spec's "the
// daemon refuses the request" scenario, which the live path has always
// honoured.
//
// Each case asserts the *effect* rather than the call: the store is never
// reached, Run counts the entry as refused rather than applied, and the
// entry is retired. Retired -- not left for the next run -- because the
// delta spec is explicit that a refusal the store would repeat identically
// forever must not stay queued: leaving it would block every valid entry
// behind it on every future replay, which is the same reasoning
// errRecordEntryDecodeFailed already retires an undecodable body under.
func TestRecordEntryMissingARequiredFieldIsRefusedWithoutWritingARow(t *testing.T) {
	dispatchWithout := func(mutate func(*records.Dispatch)) records.Dispatch {
		d := testDispatch()
		mutate(&d)
		return d
	}
	endWithout := func(mutate func(*records.DispatchEnd)) records.DispatchEnd {
		e := testDispatchEnd()
		mutate(&e)
		return e
	}
	findingWithout := func(mutate func(*records.Finding)) records.Finding {
		f := testFinding("F1", "open")
		mutate(&f)
		return f
	}

	for _, tc := range []struct {
		name string
		kind string
		req  any
	}{
		{"dispatch, no role", "dispatch", dispatchWithout(func(d *records.Dispatch) { d.Role = "" })},
		{"dispatch, no model", "dispatch", dispatchWithout(func(d *records.Dispatch) { d.Model = "" })},
		{"dispatch, no startedAt", "dispatch", dispatchWithout(func(d *records.Dispatch) { d.StartedAt = time.Time{} })},
		{"dispatch end, no key", "dispatch-end", endWithout(func(e *records.DispatchEnd) { e.Key = "" })},
		{"dispatch end, no session token", "dispatch-end", endWithout(func(e *records.DispatchEnd) { e.SessionToken = "" })},
		{"dispatch end, no endedAt", "dispatch-end", endWithout(func(e *records.DispatchEnd) { e.EndedAt = time.Time{} })},
		{"finding, no ref", "finding", findingWithout(func(f *records.Finding) { f.Ref = "" })},
		{"finding, no severity", "finding", findingWithout(func(f *records.Finding) { f.Severity = "" })},
		{"finding, no note", "finding", findingWithout(func(f *records.Finding) { f.Note = "" })},
		{"finding, no slot", "finding", findingWithout(func(f *records.Finding) { f.Slot = "" })},
		{"finding, no status", "finding", findingWithout(func(f *records.Finding) { f.Status = "" })},
		{"status, no status", "status", recordStatusRequest{Ref: "F1", Status: ""}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			const project, change = "proj-record-required", "chg-record-required"

			appendRecordWrite(t, root, project, change, tc.kind, tc.req)

			rs := &fakeRecordStore{}
			rec := reconcile.New(&fakeStore{}, nopStageStore{}, rs, root, nil)

			result, err := rec.Run(context.Background())
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if calls := rs.appliedCalls(); len(calls) != 0 {
				t.Errorf("the store saw %q, want no call at all -- an entry missing a required field must never reach it", calls)
			}
			if result.Applied != 0 || result.Refused != 1 {
				t.Errorf("Run result = %+v, want {Applied:0 Refused:1} -- a refusal must never be counted (or reported) as a success", result)
			}
			if n := pendingRecordCount(t, root, project, change); n != 0 {
				t.Errorf("pending record entries after replay = %d, want 0 -- a refusal the store would repeat forever must retire, not block the file", n)
			}
		})
	}
}

// TestRecordStatusEntryWithAnEmptyRefStillReachesTheStore pins the one
// record path that must *not* grow a required-field check. An empty ref on
// a status write is already safe end to end: store.SetFindingStatus's
// UPDATE matches zero rows and reports store.ErrFindingNotFound, which
// isDefinitiveRecordOutcome retires. Refusing it here instead would move
// the answer without improving it, and would diverge from the live PATCH
// route, whose ref comes from the URL and is likewise handed to the store
// unexamined.
func TestRecordStatusEntryWithAnEmptyRefStillReachesTheStore(t *testing.T) {
	root := t.TempDir()
	const project, change = "proj-record-emptyref", "chg-record-emptyref"

	appendRecordWrite(t, root, project, change, "status", recordStatusRequest{Ref: "", Status: "fixed"})

	rs := &fakeRecordStore{}
	rec := reconcile.New(&fakeStore{}, nopStageStore{}, rs, root, nil)

	if _, err := rec.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	assertAppliedCalls(t, rs.appliedCalls(), []string{
		"status proj-record-emptyref/chg-record-emptyref ref= status=fixed",
	})
}

// notFoundRecordStore answers every dispatch end with
// store.ErrDispatchNotFound and nothing else -- the store having been
// reached and having found no row under the key.
type notFoundRecordStore struct{ nopRecordStore }

func (notFoundRecordStore) EndDispatch(_ context.Context, projectKey, change string, in records.DispatchEnd) (records.Dispatch, error) {
	return records.Dispatch{}, fmt.Errorf("%w: %s/%s key %q", store.ErrDispatchNotFound, projectKey, change, in.Key)
}

// TestDispatchEndNamingNoRowStaysQueued pins the one 404 this package
// treats as retryable rather than definitive, and it is the mirror image of
// the finding cases above.
//
// A finding ref naming nothing is a typo: the store answers the same way
// forever, so retiring it is what stops one bad entry blocking every valid
// entry behind it. "No dispatch under this key" has a second, entirely
// ordinary cause -- the begin that would have created the row was itself
// journalled, and has not been replayed yet -- so retiring the end would
// discard it and leave the window that begin opened open forever, which is
// exactly the defect the end call exists to prevent. It stays queued, and
// resolves itself the moment the begin lands.
func TestDispatchEndNamingNoRowStaysQueued(t *testing.T) {
	root := t.TempDir()
	const project, change = "proj-record-end-queued", "chg-record-end-queued"

	appendRecordWrite(t, root, project, change, "dispatch-end", testDispatchEnd())

	rec := reconcile.New(&fakeStore{}, nopStageStore{}, notFoundRecordStore{}, root, nil)
	result, err := rec.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Applied != 0 || result.Refused != 0 {
		t.Errorf("Run result = %+v, want {Applied:0 Refused:0} -- neither applied nor definitively refused", result)
	}
	if n := pendingRecordCount(t, root, project, change); n != 1 {
		t.Errorf("pending record entries after replay = %d, want 1 -- an end whose begin has not landed must wait for it, not be discarded", n)
	}
}
