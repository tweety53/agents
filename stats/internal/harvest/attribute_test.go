package harvest_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/harvest"
)

// mustParse parses an RFC 3339 timestamp or fails the test -- used
// throughout to build Window boundaries against the fixtures' own real
// timestamps.
func mustParse(t *testing.T, s string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return ts
}

// fakeWindowSource answers WindowsForSession from a fixed map, so
// attribute_test.go never needs a database -- exactly what
// TestHarvestNeedsNoDatabase asserts about this package as a whole.
type fakeWindowSource struct {
	bySession map[string][]harvest.Window
	// calls records every session id WindowsForSession was asked about,
	// so a test can assert it was consulted exactly once per session
	// present in a batch, not once per record.
	calls []string
	err   error
}

func (f *fakeWindowSource) WindowsForSession(_ context.Context, sessionID string) ([]harvest.Window, error) {
	f.calls = append(f.calls, sessionID)
	if f.err != nil {
		return nil, f.err
	}
	return f.bySession[sessionID], nil
}

// mainThreadRecords and sidechainRecords parse the two fixtures once per
// test, via the package's own already-tested parser -- attribute_test.go
// is about attribution, not re-proving parsing works.
func mainThreadRecords(t *testing.T) []harvest.Record {
	t.Helper()
	complete, _ := harvest.SplitCompleteLines(readFixture(t, mainThreadFixture))
	return harvest.ParseAssistantRecords(complete)
}

func sidechainRecords(t *testing.T) []harvest.Record {
	t.Helper()
	complete, _ := harvest.SplitCompleteLines(readFixture(t, sidechainFixture))
	return harvest.ParseAssistantRecords(complete)
}

const mainSessionID = "session-main-a1b2c3"

// TestAttributeToOpenWindow attributes the six main-thread fixture
// records belonging to mainSessionID to one open window and checks the
// resulting delta against sums computed independently from the fixture
// (recorded here as literals) -- so this test fails if attribution
// double-counts, drops a record, or mixes up which bucket a token figure
// lands in.
func TestAttributeToOpenWindow(t *testing.T) {
	records := mainThreadRecords(t)

	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
		mainSessionID: {{
			StageRunID: 42,
			Attempt:    1,
			SessionID:  mainSessionID,
			StartedAt:  mustParse(t, "2025-12-01T00:00:00Z"),
			EndedAt:    nil, // still open
		}},
	}}
	a := harvest.NewAttributor(windows)

	// The 7th record belongs to a different session with no window
	// registered for it, so only stage run 42 should ever appear in the
	// result -- see TestMessagesOutsideAnyWindowAreNotAttributed for that
	// assertion in isolation.
	deltas, err := a.Attribute(context.Background(), records)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	if len(deltas) != 1 {
		t.Fatalf("deltas = %v, want exactly one entry (stage run 42)", deltas)
	}

	d, ok := deltas[42]
	if !ok {
		t.Fatalf("no delta for stage run 42: %v", deltas)
	}
	if d.Total.Main.Input != 109 {
		t.Errorf("tokens.main.input = %v, want 109", d.Total.Main.Input)
	}
	if d.Total.Main.CacheCreation != 33063 {
		t.Errorf("tokens.main.cache_creation = %v, want 33063", d.Total.Main.CacheCreation)
	}
	if d.Total.Main.CacheRead != 109042 {
		t.Errorf("tokens.main.cache_read = %v, want 109042", d.Total.Main.CacheRead)
	}
	if d.Total.Main.Output != 1650 {
		t.Errorf("tokens.main.output = %v, want 1650", d.Total.Main.Output)
	}
}

// TestMessagesOutsideAnyWindowAreNotAttributed asserts the negative case
// design.md draws no window for: a message whose session has no
// registered window at all produces no delta at all, never one
// attributed to the nearest window by guesswork.
func TestMessagesOutsideAnyWindowAreNotAttributed(t *testing.T) {
	records := mainThreadRecords(t) // includes one record from a second, unregistered session
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
		mainSessionID: {{
			StageRunID: 1,
			SessionID:  mainSessionID,
			StartedAt:  mustParse(t, "2025-12-01T00:00:00Z"),
		}},
	}}
	a := harvest.NewAttributor(windows)

	deltas, err := a.Attribute(context.Background(), records)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}

	// Stage run 1 got the six mainSessionID records; nothing else should
	// ever appear -- in particular nothing invented for the 7th record's
	// own, unregistered session.
	if len(deltas) != 1 {
		t.Fatalf("deltas = %v, want exactly one entry (stage run 1)", deltas)
	}
	if _, ok := deltas[1]; !ok {
		t.Fatalf("no delta for stage run 1: %v", deltas)
	}
}

// TestMessageOutsideItsWindowIntervalIsNotAttributed is the timestamp-side
// twin of TestMessagesOutsideAnyWindowAreNotAttributed: the session
// matches, but the window's interval already excludes the message.
func TestMessageOutsideItsWindowIntervalIsNotAttributed(t *testing.T) {
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
		mainSessionID: {{
			StageRunID: 7,
			SessionID:  mainSessionID,
			StartedAt:  mustParse(t, "2020-01-01T00:00:00Z"),
			EndedAt:    ptrTime(mustParse(t, "2020-01-01T00:00:01Z")),
		}},
	}}
	a := harvest.NewAttributor(windows)

	records := mainThreadRecords(t) // every real record's timestamp is on 2026-01-01, none in the window above
	deltas, err := a.Attribute(context.Background(), records)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	if len(deltas) != 0 {
		t.Fatalf("deltas = %v, want empty: every record's timestamp falls outside the one registered window", deltas)
	}
}

func ptrTime(t time.Time) *time.Time { return &t }

// TestSidechainAccumulatesSeparately attributes main-thread and sidechain
// fixture records that share one session id into one window, and checks
// that main and sidechain totals land in separate buckets rather than
// one combined figure -- the entire reason design.md rejects
// self-reported marks as the metric source: a dispatching command's own
// cost must stay distinguishable from what it dispatched.
func TestSidechainAccumulatesSeparately(t *testing.T) {
	var records []harvest.Record
	records = append(records, mainThreadRecords(t)...)
	records = append(records, sidechainRecords(t)...)

	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
		mainSessionID: {{
			StageRunID: 9,
			SessionID:  mainSessionID,
			StartedAt:  mustParse(t, "2025-12-01T00:00:00Z"),
			EndedAt:    ptrTime(mustParse(t, "2027-01-01T00:00:00Z")),
		}},
	}}
	a := harvest.NewAttributor(windows)

	deltas, err := a.Attribute(context.Background(), records)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}

	d, ok := deltas[9]
	if !ok {
		t.Fatalf("no delta for stage run 9: %v", deltas)
	}
	if d.Total.Main.Input != 109 {
		t.Errorf("tokens.main.input = %v, want 109 (main-thread total, unaffected by sidechain records)", d.Total.Main.Input)
	}
	if d.Total.Sidechain.Input != 8 {
		t.Errorf("tokens.sidechain.input = %v, want 8", d.Total.Sidechain.Input)
	}
	if d.Total.Sidechain.CacheCreation != 35348 {
		t.Errorf("tokens.sidechain.cache_creation = %v, want 35348", d.Total.Sidechain.CacheCreation)
	}
	if d.Total.Sidechain.CacheRead != 77676 {
		t.Errorf("tokens.sidechain.cache_read = %v, want 77676", d.Total.Sidechain.CacheRead)
	}
	if d.Total.Sidechain.Output != 125 {
		t.Errorf("tokens.sidechain.output = %v, want 125", d.Total.Sidechain.Output)
	}
}

// TestAttributeBucketsTokensPerModel attributes the mixed main-thread
// (claude-opus-5) and sidechain (claude-sonnet-5) fixtures into one
// window and checks that Attribute's per-model buckets carry each
// model's own totals, and that they sum to Total -- the same whole-run
// figure TestSidechainAccumulatesSeparately already checks against Main
// and Sidechain directly. This is task 22's "buckets sum to the
// top-level total" guarantee: nothing is redistributed or lost on the
// way from Total to Models, because both are derived from the same
// records in the same pass.
func TestAttributeBucketsTokensPerModel(t *testing.T) {
	var records []harvest.Record
	records = append(records, mainThreadRecords(t)...)
	records = append(records, sidechainRecords(t)...)

	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
		mainSessionID: {{
			StageRunID: 9,
			SessionID:  mainSessionID,
			StartedAt:  mustParse(t, "2025-12-01T00:00:00Z"),
			EndedAt:    ptrTime(mustParse(t, "2027-01-01T00:00:00Z")),
		}},
	}}
	a := harvest.NewAttributor(windows)

	deltas, err := a.Attribute(context.Background(), records)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	d, ok := deltas[9]
	if !ok {
		t.Fatalf("no delta for stage run 9: %v", deltas)
	}

	opus, ok := d.Models["claude-opus-5"]
	if !ok {
		t.Fatalf("no models bucket for claude-opus-5: %v", d.Models)
	}
	if opus.Main.Input != 109 {
		t.Errorf("models[claude-opus-5].main.input = %v, want 109 (the whole main-thread fixture, all opus-5)", opus.Main.Input)
	}
	if opus.Sidechain.Input != 0 {
		t.Errorf("models[claude-opus-5].sidechain.input = %v, want 0 (every sidechain record is sonnet-5)", opus.Sidechain.Input)
	}

	sonnet, ok := d.Models["claude-sonnet-5"]
	if !ok {
		t.Fatalf("no models bucket for claude-sonnet-5: %v", d.Models)
	}
	if sonnet.Sidechain.Input != 8 {
		t.Errorf("models[claude-sonnet-5].sidechain.input = %v, want 8", sonnet.Sidechain.Input)
	}
	if sonnet.Main.Input != 0 {
		t.Errorf("models[claude-sonnet-5].main.input = %v, want 0 (no sonnet-5 record is on the main thread)", sonnet.Main.Input)
	}

	// The models buckets must sum to exactly Total -- the whole-run
	// figure the top-level "tokens" key stores and every existing
	// aggregation already reads, unchanged by this task.
	if got, want := opus.Main.Input+sonnet.Main.Input, d.Total.Main.Input; got != want {
		t.Errorf("models main.input sum = %v, want Total.Main.Input = %v", got, want)
	}
	if got, want := opus.Sidechain.Input+sonnet.Sidechain.Input, d.Total.Sidechain.Input; got != want {
		t.Errorf("models sidechain.input sum = %v, want Total.Sidechain.Input = %v", got, want)
	}
}

// TestRecordWithNoModelContributesToTotalOnlyNotToAnyBucket asserts the
// rule task 22 exists to protect: a record whose message carried no
// model still counts toward the whole-run Total (nothing is dropped),
// but is filed under no models key at all -- not "unknown", not "" --
// because a fabricated model name would appear in a model dropdown as
// though someone had actually chosen it, and would break
// absence-is-never-zero at its source.
func TestRecordWithNoModelContributesToTotalOnlyNotToAnyBucket(t *testing.T) {
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
		mainSessionID: {{
			StageRunID: 5,
			SessionID:  mainSessionID,
			StartedAt:  mustParse(t, "2025-12-01T00:00:00Z"),
		}},
	}}
	a := harvest.NewAttributor(windows)

	rec := harvest.Record{
		Timestamp: mustParse(t, "2026-01-01T00:00:00Z"),
		SessionID: mainSessionID,
		Model:     "", // no model recorded on this message
		Usage:     harvest.Usage{InputTokens: 7},
	}

	deltas, err := a.Attribute(context.Background(), []harvest.Record{rec})
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	d, ok := deltas[5]
	if !ok {
		t.Fatalf("no delta for stage run 5: %v", deltas)
	}
	if d.Total.Main.Input != 7 {
		t.Errorf("Total.Main.Input = %v, want 7 -- a record with no model must still count toward the whole-run total", d.Total.Main.Input)
	}
	if len(d.Models) != 0 {
		t.Errorf("Models = %v, want empty -- a record with no model must not be filed under any model key, fabricated or otherwise", d.Models)
	}
}

// TestAttributePerModelBucketsAreAdditiveAcrossBatches simulates what
// jsonb_deep_add does in production (0005_jsonb_deep_add.sql): two
// separate Attribute calls over disjoint slices of the same fixture,
// each standing in for its own harvest batch, must produce per-model
// deltas that simply add together -- the harvester never recomputes a
// cumulative total from scratch, so Attribute must never assume it is
// seeing a session's whole history in one call.
func TestAttributePerModelBucketsAreAdditiveAcrossBatches(t *testing.T) {
	all := mainThreadRecords(t) // 6 mainSessionID records (claude-opus-5) + 1 other-session record
	boundary := mustParse(t, "2026-01-01T00:00:05Z")
	var batch1, batch2 []harvest.Record
	for _, r := range all {
		if r.SessionID != mainSessionID {
			continue
		}
		if r.Timestamp.Before(boundary) {
			batch1 = append(batch1, r)
		} else {
			batch2 = append(batch2, r)
		}
	}
	if len(batch1) == 0 || len(batch2) == 0 {
		t.Fatalf("fixture split produced an empty batch: batch1=%d batch2=%d", len(batch1), len(batch2))
	}

	window := map[string][]harvest.Window{
		mainSessionID: {{
			StageRunID: 11,
			SessionID:  mainSessionID,
			StartedAt:  mustParse(t, "2025-12-01T00:00:00Z"),
			EndedAt:    ptrTime(mustParse(t, "2027-01-01T00:00:00Z")),
		}},
	}

	d1, err := harvest.NewAttributor(&fakeWindowSource{bySession: window}).Attribute(context.Background(), batch1)
	if err != nil {
		t.Fatalf("Attribute (batch 1): %v", err)
	}
	d2, err := harvest.NewAttributor(&fakeWindowSource{bySession: window}).Attribute(context.Background(), batch2)
	if err != nil {
		t.Fatalf("Attribute (batch 2): %v", err)
	}

	// 109 is the same whole-run total TestAttributeToOpenWindow computes
	// from one unsplit batch -- jsonb_deep_add's job in production is
	// exactly to reproduce this sum from two separate commits.
	summed := d1[11].Models["claude-opus-5"].Main.Input + d2[11].Models["claude-opus-5"].Main.Input
	if summed != 109 {
		t.Errorf("models[claude-opus-5].main.input summed across two batches = %v, want 109", summed)
	}
}

// TestBoundaryStraddlingMessagesSplitByTimestamp uses two adjacent
// windows over the same session -- split at the seven-second gap the
// fixture's third and fourth assistant messages leave between them --
// and checks that each message lands in the window whose interval
// actually contains its timestamp, not in whichever window is listed
// first or happens to be "closest".
func TestBoundaryStraddlingMessagesSplitByTimestamp(t *testing.T) {
	boundary := mustParse(t, "2026-01-01T00:00:05Z") // strictly between msg 3 (00:00:02.5) and msg 4 (00:00:10)
	windowStart := mustParse(t, "2025-12-01T00:00:00Z")
	windowEnd := mustParse(t, "2026-06-01T00:00:00Z")

	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
		mainSessionID: {
			{StageRunID: 101, SessionID: mainSessionID, StartedAt: windowStart, EndedAt: ptrTime(boundary)},
			{StageRunID: 102, SessionID: mainSessionID, StartedAt: boundary, EndedAt: ptrTime(windowEnd)},
		},
	}}
	a := harvest.NewAttributor(windows)

	deltas, err := a.Attribute(context.Background(), mainThreadRecords(t))
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}

	before, ok := deltas[101]
	if !ok {
		t.Fatalf("no delta for stage run 101: %v", deltas)
	}
	// Messages 1-3: input 2+2+2=6, cache_creation 10597*3=31791,
	// cache_read 12673*3=38019, output 302*3=906.
	if before.Total.Main.Input != 6 {
		t.Errorf("stage run 101 (before boundary) tokens.main.input = %v, want 6", before.Total.Main.Input)
	}
	if before.Total.Main.Output != 906 {
		t.Errorf("stage run 101 (before boundary) tokens.main.output = %v, want 906", before.Total.Main.Output)
	}

	after, ok := deltas[102]
	if !ok {
		t.Fatalf("no delta for stage run 102: %v", deltas)
	}
	// Messages 4-6: input 99+2+2=103, output 357+316+71=744.
	if after.Total.Main.Input != 103 {
		t.Errorf("stage run 102 (after boundary) tokens.main.input = %v, want 103", after.Total.Main.Input)
	}
	if after.Total.Main.Output != 744 {
		t.Errorf("stage run 102 (after boundary) tokens.main.output = %v, want 744", after.Total.Main.Output)
	}

	// 6 + 103 = 109, the same total TestAttributeToOpenWindow computes
	// against one unsplit window -- the split changes attribution, never
	// the grand total.
	if before.Total.Main.Input+after.Total.Main.Input != 109 {
		t.Errorf("combined tokens.main.input across both windows = %v, want 109", before.Total.Main.Input+after.Total.Main.Input)
	}
}

// TestExactBoundaryMessageGoesToTheStartingWindow is F4's own test: a
// message timestamped exactly at one window's EndedAt, which also equals
// the next window's StartedAt, must resolve to exactly the second window
// -- never both, and never depending on WindowsForSession's return
// order. This is the case TestBoundaryStraddlingMessagesSplitByTimestamp
// deliberately avoids (its boundary sits strictly between two message
// timestamps); this test constructs a synthetic record landing exactly
// on the seam instead, both with windows listed start-then-end and
// end-then-start, to rule out the "incidental ORDER BY" failure mode the
// review named directly.
func TestExactBoundaryMessageGoesToTheStartingWindow(t *testing.T) {
	seam := mustParse(t, "2026-07-26T15:25:40Z")
	before := harvest.Window{
		StageRunID: 201, SessionID: mainSessionID,
		StartedAt: mustParse(t, "2026-07-26T15:25:00Z"), EndedAt: ptrTime(seam),
	}
	after := harvest.Window{
		StageRunID: 202, SessionID: mainSessionID,
		StartedAt: seam, EndedAt: ptrTime(mustParse(t, "2026-07-26T15:26:00Z")),
	}
	onSeam := harvest.Record{
		Timestamp: seam,
		SessionID: mainSessionID,
		Usage:     harvest.Usage{InputTokens: 1},
	}

	for _, order := range [][]harvest.Window{{before, after}, {after, before}} {
		windows := &fakeWindowSource{bySession: map[string][]harvest.Window{mainSessionID: order}}
		a := harvest.NewAttributor(windows)

		deltas, err := a.Attribute(context.Background(), []harvest.Record{onSeam})
		if err != nil {
			t.Fatalf("Attribute: %v", err)
		}
		if _, ok := deltas[202]; !ok {
			t.Fatalf("order %v: deltas = %v, want an entry for 202 (the window starting exactly at the message's timestamp)", order, deltas)
		}
		if _, ok := deltas[201]; ok {
			t.Fatalf("order %v: stage run 201 (the window ending exactly at the message's timestamp) has a delta, want untouched", order)
		}
	}
}

// TestReplayedBeginPrefersHighestAttempt is F3's own test: two open
// windows for the same session, both containing the message's timestamp
// (Window's own doc comment: "A replayed begin mark can open a second
// attempt" -- design.md), must resolve to the one with the higher
// Attempt, never the lower one and never whichever WindowsForSession
// happens to list first. Both orderings are exercised, the same
// discipline TestExactBoundaryMessageGoesToTheStartingWindow uses for the
// boundary case, to rule out an incidental ORDER BY standing in for the
// stated tie-break rule.
func TestReplayedBeginPrefersHighestAttempt(t *testing.T) {
	ts := mustParse(t, "2026-07-26T15:25:40Z")
	orphan := harvest.Window{
		StageRunID: 301, SessionID: mainSessionID, Attempt: 1,
		StartedAt: mustParse(t, "2026-07-26T15:00:00Z"), EndedAt: nil, // still open
	}
	replay := harvest.Window{
		StageRunID: 302, SessionID: mainSessionID, Attempt: 2,
		StartedAt: mustParse(t, "2026-07-26T15:10:00Z"), EndedAt: nil, // still open, overlaps orphan
	}
	rec := harvest.Record{Timestamp: ts, SessionID: mainSessionID, Usage: harvest.Usage{InputTokens: 1}}

	for _, order := range [][]harvest.Window{{orphan, replay}, {replay, orphan}} {
		windows := &fakeWindowSource{bySession: map[string][]harvest.Window{mainSessionID: order}}
		a := harvest.NewAttributor(windows)

		deltas, err := a.Attribute(context.Background(), []harvest.Record{rec})
		if err != nil {
			t.Fatalf("Attribute: %v", err)
		}
		if _, ok := deltas[302]; !ok {
			t.Fatalf("order %v: deltas = %v, want an entry for 302 (Attempt=2, the live replay)", order, deltas)
		}
		if _, ok := deltas[301]; ok {
			t.Fatalf("order %v: stage run 301 (Attempt=1, the orphaned earlier attempt) has a delta, want untouched", order)
		}
	}
}

// TestWindowContainsIsHalfOpen is a narrower, direct assertion of
// Window.contains itself -- half-open on the end, closed on the start --
// independent of bestWindow's tie-breaking, so a future change to
// bestWindow cannot accidentally restore the closed-on-both-ends
// behaviour without this test noticing.
func TestWindowContainsIsHalfOpen(t *testing.T) {
	start := mustParse(t, "2026-01-01T00:00:00Z")
	end := mustParse(t, "2026-01-01T01:00:00Z")
	w := harvest.Window{StartedAt: start, EndedAt: ptrTime(end)}

	cases := []struct {
		name string
		ts   time.Time
		want bool
	}{
		{"at start", start, true},
		{"just after start", start.Add(time.Second), true},
		{"at end", end, false},
		{"just before end", end.Add(-time.Nanosecond), true},
		{"just after end", end.Add(time.Nanosecond), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// contains is unexported; exercise it through Attribute
			// against a single-window fixture instead of a direct call.
			windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
				mainSessionID: {{StageRunID: 1, SessionID: mainSessionID, StartedAt: w.StartedAt, EndedAt: w.EndedAt}},
			}}
			a := harvest.NewAttributor(windows)
			rec := harvest.Record{Timestamp: c.ts, SessionID: mainSessionID, Usage: harvest.Usage{InputTokens: 1}}

			deltas, err := a.Attribute(context.Background(), []harvest.Record{rec})
			if err != nil {
				t.Fatalf("Attribute: %v", err)
			}
			got := len(deltas) == 1
			if got != c.want {
				t.Errorf("contains(%s) = %v, want %v", c.name, got, c.want)
			}
		})
	}
}

// TestAttributePropagatesWindowSourceFailure asserts that a failing
// WindowSource's error surfaces from Attribute rather than being
// swallowed -- Watcher.RunOnce depends on this: it must not proceed to
// commit anything for a batch whose attribution could not actually be
// computed.
func TestAttributePropagatesWindowSourceFailure(t *testing.T) {
	windows := &fakeWindowSource{err: errors.New("window source unavailable")}
	a := harvest.NewAttributor(windows)

	_, err := a.Attribute(context.Background(), mainThreadRecords(t))
	if err == nil {
		t.Fatalf("expected an error from a failing WindowSource, got nil")
	}
}

// TestWindowsForSessionConsultedOncePerSession is a narrow guard against a
// performance regression that would also be an interface-contract
// violation: a batch spanning many messages for one session must ask
// WindowSource once for that session, not once per message.
func TestWindowsForSessionConsultedOncePerSession(t *testing.T) {
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
		mainSessionID: {{
			StageRunID: 3,
			SessionID:  mainSessionID,
			StartedAt:  mustParse(t, "2025-12-01T00:00:00Z"),
			EndedAt:    ptrTime(mustParse(t, "2027-01-01T00:00:00Z")),
		}},
	}}
	a := harvest.NewAttributor(windows)

	if _, err := a.Attribute(context.Background(), mainThreadRecords(t)); err != nil {
		t.Fatalf("Attribute: %v", err)
	}

	seen := map[string]int{}
	for _, s := range windows.calls {
		seen[s]++
	}
	for session, n := range seen {
		if n != 1 {
			t.Errorf("WindowsForSession(%q) called %d times, want exactly 1", session, n)
		}
	}
	if len(seen) != 2 { // mainSessionID plus the 7th record's own session
		t.Errorf("consulted %d distinct sessions, want 2", len(seen))
	}
}

// TestAttributeSplitsCacheCreationByRate is task 23's own attribution-level
// guard: the main-thread fixture's cache-creation usage is entirely
// 1-hour writes (confirmed by reading the fixture directly, not assumed),
// so a correct split must land the whole collapsed total already proven
// by TestAttributeToOpenWindow (33063) in CacheCreation1h, with nothing
// left in CacheCreation5m or CacheCreationUnknown.
func TestAttributeSplitsCacheCreationByRate(t *testing.T) {
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
		mainSessionID: {{
			StageRunID: 42,
			SessionID:  mainSessionID,
			StartedAt:  mustParse(t, "2025-12-01T00:00:00Z"),
		}},
	}}
	a := harvest.NewAttributor(windows)

	deltas, err := a.Attribute(context.Background(), mainThreadRecords(t))
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	main := deltas[42].Total.Main
	if main.CacheCreation1h != 33063 {
		t.Errorf("tokens.main.cache_creation_1h = %v, want 33063", main.CacheCreation1h)
	}
	if main.CacheCreation5m != 0 {
		t.Errorf("tokens.main.cache_creation_5m = %v, want 0 (every main-thread record is a 1-hour write)", main.CacheCreation5m)
	}
	if main.CacheCreationUnknown != 0 {
		t.Errorf("tokens.main.cache_creation_unknown = %v, want 0 (every record carried a known split)", main.CacheCreationUnknown)
	}
	if main.CacheCreation != main.CacheCreation5m+main.CacheCreation1h {
		t.Errorf("cache_creation (%v) != cache_creation_5m + cache_creation_1h (%v)", main.CacheCreation, main.CacheCreation5m+main.CacheCreation1h)
	}
}

// TestAttributeRecordsUnknownCacheSplitSeparately covers the absent-split
// case: a record with no "cache_creation" object at all must add its
// whole collapsed total to CacheCreationUnknown, never guess a split by
// leaving both dedicated fields at zero (which would misread as "no
// cache-creation usage at that rate happened", a different, false claim).
func TestAttributeRecordsUnknownCacheSplitSeparately(t *testing.T) {
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
		"s-unknown-split": {{
			StageRunID: 99,
			SessionID:  "s-unknown-split",
			StartedAt:  mustParse(t, "2025-12-01T00:00:00Z"),
		}},
	}}
	a := harvest.NewAttributor(windows)

	records := []harvest.Record{{
		Timestamp: mustParse(t, "2025-12-01T00:00:01Z"),
		SessionID: "s-unknown-split",
		Model:     "claude-opus-5",
		Usage: harvest.Usage{
			InputTokens:              1,
			CacheCreationInputTokens: 500,
			CacheSplitKnown:          false,
		},
	}}
	deltas, err := a.Attribute(context.Background(), records)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	main := deltas[99].Total.Main
	if main.CacheCreationUnknown != 500 {
		t.Errorf("tokens.main.cache_creation_unknown = %v, want 500", main.CacheCreationUnknown)
	}
	if main.CacheCreation5m != 0 || main.CacheCreation1h != 0 {
		t.Errorf("split tokens = (5m=%v, 1h=%v), want both 0 for an unknown-split record", main.CacheCreation5m, main.CacheCreation1h)
	}
	if main.CacheCreation != 500 {
		t.Errorf("tokens.main.cache_creation = %v, want 500 (the collapsed total is still recorded)", main.CacheCreation)
	}
}

// TestAttributeCapturesSpeedLastWriteWins covers Delta.Speed: the last
// non-empty speed value among a stage run's records in this batch wins,
// mirroring the rule Effort's own doc comment already states for that
// field.
func TestAttributeCapturesSpeedLastWriteWins(t *testing.T) {
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
		"s-speed": {{
			StageRunID: 7,
			SessionID:  "s-speed",
			StartedAt:  mustParse(t, "2025-12-01T00:00:00Z"),
		}},
	}}
	a := harvest.NewAttributor(windows)

	records := []harvest.Record{
		{
			Timestamp: mustParse(t, "2025-12-01T00:00:01Z"),
			SessionID: "s-speed",
			Model:     "claude-opus-5",
			Usage:     harvest.Usage{InputTokens: 1, Speed: "standard"},
		},
		{
			Timestamp: mustParse(t, "2025-12-01T00:00:02Z"),
			SessionID: "s-speed",
			Model:     "claude-opus-5",
			Usage:     harvest.Usage{InputTokens: 1, Speed: "fast"},
		},
	}
	deltas, err := a.Attribute(context.Background(), records)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	if got := deltas[7].Speed; got != "fast" {
		t.Errorf("Delta.Speed = %q, want %q (the last record's speed)", got, "fast")
	}
}

// Compile-time sanity: fakeWindowSource must actually satisfy the
// interface it stands in for, so a signature drift in attribute.go fails
// the build rather than silently changing what these tests exercise.
var _ harvest.WindowSource = (*fakeWindowSource)(nil)
