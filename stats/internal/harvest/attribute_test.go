package harvest_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
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

// TestAttributeSplitsByAgentID asserts KAN-201's own dispatch breakout:
// two records carrying different AgentIDs against one stage run produce
// two entries in Delta.Dispatches, each carrying only its own record's
// usage -- Delta.Models' own per-model split (TestAttributeBucketsTokensPerModel
// above) applied to a second key, agentId instead of model.
func TestAttributeSplitsByAgentID(t *testing.T) {
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
		mainSessionID: {{
			StageRunID: 12,
			SessionID:  mainSessionID,
			StartedAt:  mustParse(t, "2025-12-01T00:00:00Z"),
		}},
	}}
	a := harvest.NewAttributor(windows)

	records := []harvest.Record{
		{
			Timestamp:   mustParse(t, "2026-01-01T00:00:00Z"),
			SessionID:   mainSessionID,
			IsSidechain: true,
			AgentID:     "agent-one",
			Usage:       harvest.Usage{InputTokens: 5},
		},
		{
			Timestamp:   mustParse(t, "2026-01-01T00:00:01Z"),
			SessionID:   mainSessionID,
			IsSidechain: true,
			AgentID:     "agent-two",
			Usage:       harvest.Usage{InputTokens: 9},
		},
	}

	deltas, err := a.Attribute(context.Background(), records)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	d, ok := deltas[12]
	if !ok {
		t.Fatalf("no delta for stage run 12: %v", deltas)
	}
	if len(d.Dispatches) != 2 {
		t.Fatalf("Dispatches = %v, want exactly two entries", d.Dispatches)
	}
	one, ok := d.Dispatches["agent-one"]
	if !ok || one.Sidechain.Input != 5 {
		t.Errorf("Dispatches[agent-one] = %+v, ok=%v, want sidechain.input=5", one, ok)
	}
	two, ok := d.Dispatches["agent-two"]
	if !ok || two.Sidechain.Input != 9 {
		t.Errorf("Dispatches[agent-two] = %+v, ok=%v, want sidechain.input=9", two, ok)
	}
}

// TestAttributeNoAgentIDCreatesNoDispatch mirrors
// TestRecordWithNoModelContributesToTotalOnlyNotToAnyBucket at the
// dispatch granularity: a record carrying no AgentID -- a parent-session
// message -- contributes to Total and to Models, and creates no entry in
// Dispatches at all, fabricated placeholder included.
func TestAttributeNoAgentIDCreatesNoDispatch(t *testing.T) {
	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
		mainSessionID: {{
			StageRunID: 13,
			SessionID:  mainSessionID,
			StartedAt:  mustParse(t, "2025-12-01T00:00:00Z"),
		}},
	}}
	a := harvest.NewAttributor(windows)

	rec := harvest.Record{
		Timestamp: mustParse(t, "2026-01-01T00:00:00Z"),
		SessionID: mainSessionID,
		Model:     "claude-opus-5",
		AgentID:   "", // parent-session message
		Usage:     harvest.Usage{InputTokens: 11},
	}

	deltas, err := a.Attribute(context.Background(), []harvest.Record{rec})
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	d, ok := deltas[13]
	if !ok {
		t.Fatalf("no delta for stage run 13: %v", deltas)
	}
	if d.Total.Main.Input != 11 {
		t.Errorf("Total.Main.Input = %v, want 11", d.Total.Main.Input)
	}
	if _, ok := d.Models["claude-opus-5"]; !ok {
		t.Errorf("Models[claude-opus-5] missing, want present: %v", d.Models)
	}
	if len(d.Dispatches) != 0 {
		t.Errorf("Dispatches = %v, want empty -- no agentId must create no dispatch entry, fabricated or otherwise", d.Dispatches)
	}
}

// TestExistingTokenKeysUnchanged pins the strictly-additive constraint
// task 4's own instructions state as a hard requirement: for a batch
// mixing main-thread and sidechain fixture records, Total.Main,
// Total.Sidechain and every Models entry hold exactly what
// TestSidechainAccumulatesSeparately and TestAttributeBucketsTokensPerModel
// already assert they held before Delta.Dispatches existed -- a stage
// run harvested before this capability must not become retroactively
// wrong.
func TestExistingTokenKeysUnchanged(t *testing.T) {
	var records []harvest.Record
	records = append(records, mainThreadRecords(t)...)
	records = append(records, sidechainRecords(t)...)

	windows := &fakeWindowSource{bySession: map[string][]harvest.Window{
		mainSessionID: {{
			StageRunID: 14,
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
	d, ok := deltas[14]
	if !ok {
		t.Fatalf("no delta for stage run 14: %v", deltas)
	}

	if d.Total.Main.Input != 109 {
		t.Errorf("Total.Main.Input = %v, want 109 (unchanged from TestSidechainAccumulatesSeparately)", d.Total.Main.Input)
	}
	if d.Total.Sidechain.Input != 8 {
		t.Errorf("Total.Sidechain.Input = %v, want 8 (unchanged from TestSidechainAccumulatesSeparately)", d.Total.Sidechain.Input)
	}
	opus, ok := d.Models["claude-opus-5"]
	if !ok || opus.Main.Input != 109 {
		t.Errorf("Models[claude-opus-5] = %+v, ok=%v, want main.input=109 (unchanged from TestAttributeBucketsTokensPerModel)", opus, ok)
	}
	sonnet, ok := d.Models["claude-sonnet-5"]
	if !ok || sonnet.Sidechain.Input != 8 {
		t.Errorf("Models[claude-sonnet-5] = %+v, ok=%v, want sidechain.input=8 (unchanged from TestAttributeBucketsTokensPerModel)", sonnet, ok)
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
// windows for the same session, both containing the message's timestamp,
// must resolve to the one with the higher Attempt, never the lower one
// and never whichever WindowsForSession happens to list first. Despite
// its name and its variables' names below, this is *not* a same-instant
// replay case: orphan and replay start ten minutes apart (15:00 and
// 15:10), so latest-start-wins already picks stage run 302 on its own
// StartedAt, before Attempt is ever consulted -- the same resolution
// TestOverlappingOpenWindowsPreferTheLatestStarted checks directly (fix
// round 4, findings F7 and F11: task 2's supersede closes a genuine
// replay's earlier attempt at the same instant it started, so the
// ordinary replay path never reaches this same-instant tiebreak at all;
// see bestWindow's own doc comment). What this test still pins is that a
// *higher* attempt does not lose to a *later* start when both point the
// same way, and that neither list order changes the answer -- the same
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

// TestOverlappingOpenWindowsPreferTheLatestStarted is the incident from
// design.md, reduced: an orphan window left open by a dropped end mark
// (started 15:00, attempt 1) and the stage actually running (started
// 15:10, attempt 1) both contain a message at 15:25. Both windows sit at
// the same attempt, so the old highest-attempt tie-break has nothing to
// say and used to fall back to iteration order -- exactly what let stage
// run 146 (kan-175) absorb two hours of kan-184's tokens. The message
// must go to the later-started window, and the orphan must receive
// nothing, in both list orders.
func TestOverlappingOpenWindowsPreferTheLatestStarted(t *testing.T) {
	orphan := harvest.Window{
		StageRunID: 401, SessionID: mainSessionID, Attempt: 1,
		StartedAt: mustParse(t, "2026-07-26T15:00:00Z"), EndedAt: nil, // still open, dropped end mark
	}
	live := harvest.Window{
		StageRunID: 402, SessionID: mainSessionID, Attempt: 1,
		StartedAt: mustParse(t, "2026-07-26T15:10:00Z"), EndedAt: nil, // still open, later start
	}
	rec := harvest.Record{
		Timestamp: mustParse(t, "2026-07-26T15:25:00Z"),
		SessionID: mainSessionID,
		Usage:     harvest.Usage{InputTokens: 1},
	}

	for _, order := range [][]harvest.Window{{orphan, live}, {live, orphan}} {
		windows := &fakeWindowSource{bySession: map[string][]harvest.Window{mainSessionID: order}}
		a := harvest.NewAttributor(windows)

		deltas, err := a.Attribute(context.Background(), []harvest.Record{rec})
		if err != nil {
			t.Fatalf("Attribute: %v", err)
		}
		if _, ok := deltas[402]; !ok {
			t.Fatalf("order %v: deltas = %v, want an entry for 402 (the window that started last)", order, deltas)
		}
		if _, ok := deltas[401]; ok {
			t.Fatalf("order %v: stage run 401 (the orphan, started first) has a delta, want untouched", order)
		}
	}
}

// TestSameInstantWindowsFallBackToHighestAttempt asserts the tie-break
// survives with something to do: two windows sharing one StartedAt,
// attempts 1 and 2 -- latest-start alone cannot distinguish them, so the
// higher attempt still wins.
func TestSameInstantWindowsFallBackToHighestAttempt(t *testing.T) {
	sameStart := mustParse(t, "2026-07-26T15:00:00Z")
	low := harvest.Window{
		StageRunID: 501, SessionID: mainSessionID, Attempt: 1,
		StartedAt: sameStart, EndedAt: nil,
	}
	high := harvest.Window{
		StageRunID: 502, SessionID: mainSessionID, Attempt: 2,
		StartedAt: sameStart, EndedAt: nil,
	}
	rec := harvest.Record{
		Timestamp: sameStart.Add(time.Minute),
		SessionID: mainSessionID,
		Usage:     harvest.Usage{InputTokens: 1},
	}

	for _, order := range [][]harvest.Window{{low, high}, {high, low}} {
		windows := &fakeWindowSource{bySession: map[string][]harvest.Window{mainSessionID: order}}
		a := harvest.NewAttributor(windows)

		deltas, err := a.Attribute(context.Background(), []harvest.Record{rec})
		if err != nil {
			t.Fatalf("Attribute: %v", err)
		}
		if _, ok := deltas[502]; !ok {
			t.Fatalf("order %v: deltas = %v, want an entry for 502 (Attempt=2, same start instant)", order, deltas)
		}
		if _, ok := deltas[501]; ok {
			t.Fatalf("order %v: stage run 501 (Attempt=1, same start instant) has a delta, want untouched", order)
		}
	}
}

// TestLatestStartOutranksAHigherAttempt pins bestWindow's tie-break
// *order* directly -- something none of the three tests above actually
// discriminates (fix round 5, finding F15). Swap bestWindow's two clauses
// to attempt-first in a scratch copy and every one of them still passes:
// TestOverlappingOpenWindowsPreferTheLatestStarted holds both windows at
// attempt 1, so it has no attempt difference to get wrong;
// TestSameInstantWindowsFallBackToHighestAttempt holds both starts equal,
// so it never exercises the StartedAt comparison at all; and
// TestReplayedBeginPrefersHighestAttempt's winner carries *both* the later
// start and the higher attempt, so either clause alone would pick it. That
// is exactly how the KAN-185 incident could return -- an orphan at a
// higher attempt number reabsorbing a live stage's tokens -- with nothing
// in this file failing.
//
// This test is the one that does fail on that swap: window A starts
// earlier but carries the *higher* attempt, window B starts later but
// carries the *lower* attempt, and only latest-start-first picks B.
func TestLatestStartOutranksAHigherAttempt(t *testing.T) {
	earlierHigherAttempt := harvest.Window{
		StageRunID: 601, SessionID: mainSessionID, Attempt: 5,
		StartedAt: mustParse(t, "2026-07-26T15:00:00Z"), EndedAt: nil,
	}
	laterLowerAttempt := harvest.Window{
		StageRunID: 602, SessionID: mainSessionID, Attempt: 1,
		StartedAt: mustParse(t, "2026-07-26T15:10:00Z"), EndedAt: nil,
	}
	rec := harvest.Record{
		Timestamp: mustParse(t, "2026-07-26T15:25:00Z"),
		SessionID: mainSessionID,
		Usage:     harvest.Usage{InputTokens: 1},
	}

	for _, order := range [][]harvest.Window{
		{earlierHigherAttempt, laterLowerAttempt},
		{laterLowerAttempt, earlierHigherAttempt},
	} {
		windows := &fakeWindowSource{bySession: map[string][]harvest.Window{mainSessionID: order}}
		att := harvest.NewAttributor(windows)

		deltas, err := att.Attribute(context.Background(), []harvest.Record{rec})
		if err != nil {
			t.Fatalf("Attribute: %v", err)
		}
		if _, ok := deltas[602]; !ok {
			t.Fatalf("order %v: deltas = %v, want an entry for 602 (later start, lower attempt)", order, deltas)
		}
		if _, ok := deltas[601]; ok {
			t.Fatalf("order %v: stage run 601 (earlier start, higher attempt) has a delta, want untouched", order)
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

// fakeDispatchWindowSource answers DispatchWindowsForSession from a fixed
// map, the dispatch-grain twin of fakeWindowSource above -- so the second
// attribution pass is testable with no database at all, exactly like the
// first (TestHarvestNeedsNoDatabase asserts that about this package as a
// whole).
type fakeDispatchWindowSource struct {
	bySession map[string][]harvest.DispatchWindow
	err       error
}

func (f *fakeDispatchWindowSource) DispatchWindowsForSession(_ context.Context, sessionID string) ([]harvest.DispatchWindow, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.bySession[sessionID], nil
}

// dispatchWindow is the sidechain fixture's own span, widened by a minute
// on each side: every one of its four assistant records (00:05:00.500 to
// 00:05:03.500 on 2026-01-01) falls inside it, and every main-thread
// record (00:00:00.500 to 00:00:17.000) falls outside it. Both facts are
// deliberate -- the cases below use this window to separate "a sidechain
// record inside the window" from "the parent's own main-thread record",
// which is the whole distinction dispatch attribution exists to draw.
func dispatchWindow(t *testing.T, dispatchID int64) harvest.DispatchWindow {
	t.Helper()
	return harvest.DispatchWindow{
		DispatchID: dispatchID,
		StartedAt:  mustParse(t, "2026-01-01T00:04:00Z"),
		EndedAt:    ptrTime(mustParse(t, "2026-01-01T00:06:00Z")),
	}
}

// TestDispatchWindowAttributesSidechainUsage is the second attribution
// pass's central case: sidechain usage recorded inside a dispatch's own
// window lands on that dispatch, with the fixture's own figures rather
// than anything the dispatched agent reported about itself.
//
// The expected totals are the same sidechain figures
// TestSidechainAccumulatesSeparately already checks at the stage grain --
// deliberately, since the two passes are two grains over the same usage
// and neither redistributes it.
func TestDispatchWindowAttributesSidechainUsage(t *testing.T) {
	var records []harvest.Record
	records = append(records, mainThreadRecords(t)...)
	records = append(records, sidechainRecords(t)...)

	windows := &fakeDispatchWindowSource{bySession: map[string][]harvest.DispatchWindow{
		mainSessionID: {dispatchWindow(t, 77)},
	}}
	a := harvest.NewDispatchAttributor(windows)

	deltas, _, err := a.Attribute(context.Background(), records)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	if len(deltas) != 1 {
		t.Fatalf("deltas = %v, want exactly one entry (dispatch 77)", deltas)
	}

	d, ok := deltas[77]
	if !ok {
		t.Fatalf("no delta for dispatch 77: %v", deltas)
	}
	if d.Sidechain.Input != 8 {
		t.Errorf("tokens.sidechain.input = %v, want 8", d.Sidechain.Input)
	}
	if d.Sidechain.CacheCreation != 35348 {
		t.Errorf("tokens.sidechain.cache_creation = %v, want 35348", d.Sidechain.CacheCreation)
	}
	if d.Sidechain.CacheCreation5m != 35348 {
		t.Errorf("tokens.sidechain.cache_creation_5m = %v, want 35348 (the fixture's split is known and entirely 5m)", d.Sidechain.CacheCreation5m)
	}
	if d.Sidechain.CacheRead != 77676 {
		t.Errorf("tokens.sidechain.cache_read = %v, want 77676", d.Sidechain.CacheRead)
	}
	if d.Sidechain.Output != 125 {
		t.Errorf("tokens.sidechain.output = %v, want 125", d.Sidechain.Output)
	}
}

// TestDispatchWindowIgnoresMainThreadUsage pins the boundary between the
// two grains: the parent's own main-thread tokens belong to the stage run
// it is running under, never to a subagent it dispatched, so a dispatch's
// Main bucket is zero even for a window wide enough to contain every
// main-thread record in the batch.
//
// Without this, a dispatch window that happened to span its parent's own
// turns would charge the parent's thinking to the agent it dispatched --
// and the split between a dispatching command's own cost and what it
// dispatched is exactly what TokenDelta's two buckets exist to keep.
func TestDispatchWindowIgnoresMainThreadUsage(t *testing.T) {
	var records []harvest.Record
	records = append(records, mainThreadRecords(t)...)
	records = append(records, sidechainRecords(t)...)

	// Wide enough to contain every record in both fixtures, main-thread
	// records included -- the point being that containment alone is not
	// enough to make a main-thread record a dispatch's cost.
	windows := &fakeDispatchWindowSource{bySession: map[string][]harvest.DispatchWindow{
		mainSessionID: {{
			DispatchID: 5,
			StartedAt:  mustParse(t, "2025-12-01T00:00:00Z"),
			EndedAt:    ptrTime(mustParse(t, "2027-01-01T00:00:00Z")),
		}},
	}}
	a := harvest.NewDispatchAttributor(windows)

	deltas, _, err := a.Attribute(context.Background(), records)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}

	d, ok := deltas[5]
	if !ok {
		t.Fatalf("no delta for dispatch 5: %v", deltas)
	}
	if d.Main != (harvest.Bucket{}) {
		t.Errorf("tokens.main = %+v, want the zero bucket: a dispatch is charged for sidechain usage only", d.Main)
	}
	if d.Sidechain.Input != 8 {
		t.Errorf("tokens.sidechain.input = %v, want 8 -- the sidechain records must still be attributed", d.Sidechain.Input)
	}
}

// TestDispatchAttributionLeavesStageAttributionUnchanged is the case that
// proves the second pass is additive rather than a change to the first:
// the same records slice, attributed at the stage grain before and after
// the dispatch pass has run over it, produces exactly the same deltas --
// and those deltas are the figures TestSidechainAccumulatesSeparately
// already pins.
//
// Both halves matter. Comparing before against after catches a dispatch
// pass that mutates the batch it shares with the first pass; comparing
// against the literal figures catches a change to what stage attribution
// records that both calls would report identically.
func TestDispatchAttributionLeavesStageAttributionUnchanged(t *testing.T) {
	var records []harvest.Record
	records = append(records, mainThreadRecords(t)...)
	records = append(records, sidechainRecords(t)...)

	// The same stage window TestSidechainAccumulatesSeparately registers,
	// so the expected figures below are that test's own.
	stage := harvest.NewAttributor(&fakeWindowSource{bySession: map[string][]harvest.Window{
		mainSessionID: {{
			StageRunID: 9,
			SessionID:  mainSessionID,
			StartedAt:  mustParse(t, "2025-12-01T00:00:00Z"),
			EndedAt:    ptrTime(mustParse(t, "2027-01-01T00:00:00Z")),
		}},
	}})

	before, err := stage.Attribute(context.Background(), records)
	if err != nil {
		t.Fatalf("stage Attribute (before): %v", err)
	}

	dispatches := harvest.NewDispatchAttributor(&fakeDispatchWindowSource{
		bySession: map[string][]harvest.DispatchWindow{mainSessionID: {dispatchWindow(t, 77)}},
	})
	if _, _, err := dispatches.Attribute(context.Background(), records); err != nil {
		t.Fatalf("dispatch Attribute: %v", err)
	}

	after, err := stage.Attribute(context.Background(), records)
	if err != nil {
		t.Fatalf("stage Attribute (after): %v", err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Errorf("stage deltas changed once the dispatch pass ran over the same records:\nbefore = %+v\nafter  = %+v", before, after)
	}

	want := harvest.TokenDelta{
		Main: harvest.Bucket{
			Input: 109, Output: 1650,
			CacheCreation: 33063, CacheCreation1h: 33063,
			CacheRead: 109042,
		},
		Sidechain: harvest.Bucket{
			Input: 8, Output: 125,
			CacheCreation: 35348, CacheCreation5m: 35348,
			CacheRead: 77676,
		},
	}
	if got := after[9].Total; got != want {
		t.Errorf("stage run 9's total = %+v, want %+v (exactly what TestSidechainAccumulatesSeparately pins)", got, want)
	}
}

// TestRecordOutsideEveryDispatchWindowIsNotAttributed is the negative
// case at the dispatch grain: sidechain usage recorded while no dispatch
// was open belongs to no dispatch at all, and is never rounded to the
// nearest one. Attribution that guesses here would be the self-reporting
// this requirement exists to avoid, one level down.
func TestRecordOutsideEveryDispatchWindowIsNotAttributed(t *testing.T) {
	windows := &fakeDispatchWindowSource{bySession: map[string][]harvest.DispatchWindow{
		mainSessionID: {{
			DispatchID: 12,
			StartedAt:  mustParse(t, "2020-01-01T00:00:00Z"),
			EndedAt:    ptrTime(mustParse(t, "2020-01-01T00:00:01Z")),
		}},
	}}
	a := harvest.NewDispatchAttributor(windows)

	deltas, _, err := a.Attribute(context.Background(), sidechainRecords(t))
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	if len(deltas) != 0 {
		t.Fatalf("deltas = %v, want empty: every sidechain record's timestamp falls outside the one registered dispatch window", deltas)
	}
}

// TestDispatchWindowIntervalIsHalfOpen pins that a dispatch window carries
// the same [StartedAt, EndedAt) convention Window already documents: a
// record timestamped exactly at one window's EndedAt, which is also the
// next window's StartedAt, resolves to the window that is *starting*.
//
// Two back-to-back dispatches with no gap between them is the ordinary
// case for a review panel's slots, so this seam is reached routinely
// rather than exceptionally, and both list orders are checked so that an
// incidental query order can never stand in for the stated rule.
func TestDispatchWindowIntervalIsHalfOpen(t *testing.T) {
	seam := mustParse(t, "2026-01-01T00:05:01.5Z")
	first := harvest.DispatchWindow{
		DispatchID: 1,
		StartedAt:  mustParse(t, "2026-01-01T00:05:00Z"), EndedAt: ptrTime(seam),
	}
	second := harvest.DispatchWindow{
		DispatchID: 2,
		StartedAt:  seam, EndedAt: ptrTime(mustParse(t, "2026-01-01T00:06:00Z")),
	}

	for _, order := range [][]harvest.DispatchWindow{{first, second}, {second, first}} {
		a := harvest.NewDispatchAttributor(&fakeDispatchWindowSource{
			bySession: map[string][]harvest.DispatchWindow{mainSessionID: order},
		})

		deltas, _, err := a.Attribute(context.Background(), sidechainRecords(t))
		if err != nil {
			t.Fatalf("order %v: Attribute: %v", order, err)
		}

		// Only the fixture's first record (00:05:00.500) precedes the
		// seam; the record *on* the seam and the two after it belong to
		// the window that is starting.
		if got := deltas[1].Sidechain.Output; got != 30 {
			t.Errorf("order %v: dispatch 1 output = %v, want 30 (its one record before the seam)", order, got)
		}
		if got := deltas[2].Sidechain.Output; got != 95 {
			t.Errorf("order %v: dispatch 2 output = %v, want 95 (the record on the seam plus the two after it)", order, got)
		}
		if got := deltas[1].Sidechain.CacheCreation; got != 20000 {
			t.Errorf("order %v: dispatch 1 cache_creation = %v, want 20000", order, got)
		}
		if got := deltas[2].Sidechain.CacheCreation; got != 15348 {
			t.Errorf("order %v: dispatch 2 cache_creation = %v, want 15348", order, got)
		}
	}
}

// panelWindow builds one review-panel slot's dispatch window: an interval
// wide enough to contain every record the agent-id cases below place
// inside it, so that *every* case in this group is one where the interval
// rule alone cannot separate the slots and only the agent id can. That is
// deliberate -- a case whose windows do not overlap would pass with or
// without agent-id matching and would prove nothing about it.
func panelWindow(t *testing.T, dispatchID int64, agentID string, startedAt string) harvest.DispatchWindow {
	t.Helper()
	return harvest.DispatchWindow{
		DispatchID: dispatchID,
		AgentID:    agentID,
		StartedAt:  mustParse(t, startedAt),
		EndedAt:    ptrTime(mustParse(t, "2026-01-01T00:11:00Z")),
	}
}

// sidechainRecord builds one sidechain record inside every window
// panelWindow builds, carrying agentID and a distinguishable input figure.
func sidechainRecord(t *testing.T, agentID string, at string, input int64) harvest.Record {
	t.Helper()
	return harvest.Record{
		Timestamp:   mustParse(t, at),
		SessionID:   mainSessionID,
		IsSidechain: true,
		AgentID:     agentID,
		Usage:       harvest.Usage{InputTokens: input},
	}
}

// attributeInEveryOrder runs the dispatch attribution pass once per
// permutation of windows and hands each result to check, so no case in
// this group can pass by depending on the order
// DispatchWindowsForSession happened to return its rows in. The order is
// the store's ORDER BY, and a rule that survives only one of its
// permutations is not a rule.
func attributeInEveryOrder(t *testing.T, windows []harvest.DispatchWindow, records []harvest.Record, check func(t *testing.T, deltas map[int64]harvest.TokenDelta)) {
	t.Helper()
	for _, order := range permuteWindows(windows) {
		a := harvest.NewDispatchAttributor(&fakeDispatchWindowSource{
			bySession: map[string][]harvest.DispatchWindow{mainSessionID: order},
		})
		deltas, _, err := a.Attribute(context.Background(), records)
		if err != nil {
			t.Fatalf("order %v: Attribute: %v", dispatchIDs(order), err)
		}
		t.Run(fmt.Sprintf("order %v", dispatchIDs(order)), func(t *testing.T) {
			check(t, deltas)
		})
	}
}

// permuteWindows returns every ordering of in, by Heap's algorithm. Every
// case in this group is checked against all of them rather than against a
// sample: the orderings of two or three windows are six at most, so
// exhausting them costs nothing and removes "the test happened to pick a
// passing order" as an explanation of a green run entirely.
func permuteWindows(in []harvest.DispatchWindow) [][]harvest.DispatchWindow {
	var out [][]harvest.DispatchWindow
	cur := append([]harvest.DispatchWindow(nil), in...)
	var generate func(k int)
	generate = func(k int) {
		if k == 1 {
			out = append(out, append([]harvest.DispatchWindow(nil), cur...))
			return
		}
		for i := range k {
			generate(k - 1)
			if k%2 == 0 {
				cur[i], cur[k-1] = cur[k-1], cur[i]
			} else {
				cur[0], cur[k-1] = cur[k-1], cur[0]
			}
		}
	}
	generate(len(cur))
	return out
}

func dispatchIDs(windows []harvest.DispatchWindow) []int64 {
	out := make([]int64, len(windows))
	for i, w := range windows {
		out[i] = w.DispatchID
	}
	return out
}

// TestConcurrentSlotsAreSeparatedByAgentID is the case this pair exists
// for. A review panel dispatches its slots at once against one parent
// session, so their windows overlap and the interval alone cannot say
// which slot a record inside the overlap belongs to. Every record here
// falls inside both windows, and the slot that started *first* still
// receives its own usage -- which is exactly what the interval rule alone
// cannot produce, since latest-start-wins would hand both records to
// slot B.
func TestConcurrentSlotsAreSeparatedByAgentID(t *testing.T) {
	windows := []harvest.DispatchWindow{
		panelWindow(t, 1, "agent-slot-a", "2026-01-01T00:10:00Z"),
		panelWindow(t, 2, "agent-slot-b", "2026-01-01T00:10:00.5Z"),
	}
	records := []harvest.Record{
		sidechainRecord(t, "agent-slot-a", "2026-01-01T00:10:30Z", 5),
		sidechainRecord(t, "agent-slot-b", "2026-01-01T00:10:31Z", 9),
	}

	attributeInEveryOrder(t, windows, records, func(t *testing.T, deltas map[int64]harvest.TokenDelta) {
		if got := deltas[1].Sidechain.Input; got != 5 {
			t.Errorf("dispatch 1 (agent-slot-a) input = %v, want 5 -- its own agent's record, not whichever slot started last", got)
		}
		if got := deltas[2].Sidechain.Input; got != 9 {
			t.Errorf("dispatch 2 (agent-slot-b) input = %v, want 9", got)
		}
	})
}

// TestRecordAgentIDMatchingNoDispatchFallsBackToTheWindowRule pins the
// fallback as the ordinary path it is. A record can carry an agent id no
// dispatch row records -- a subagent dispatched by something other than
// the pipeline, or a record harvested before this column existed -- and
// the interval rule must still place it, unchanged and without complaint.
// The windows are non-overlapping (kan-212's identity-beats-interval pass 2
// is ambiguous rather than tie-broken wherever two windows both contain the
// record, so a fallback case must give the interval pass exactly one
// candidate to prove the fallback itself, not the ambiguity rule).
func TestRecordAgentIDMatchingNoDispatchFallsBackToTheWindowRule(t *testing.T) {
	windows := []harvest.DispatchWindow{
		{DispatchID: 1, AgentID: "agent-slot-a", StartedAt: mustParse(t, "2026-01-01T00:10:00Z"), EndedAt: ptrTime(mustParse(t, "2026-01-01T00:10:15Z"))},
		{DispatchID: 2, AgentID: "agent-slot-b", StartedAt: mustParse(t, "2026-01-01T00:10:15Z"), EndedAt: ptrTime(mustParse(t, "2026-01-01T00:11:00Z"))},
	}
	records := []harvest.Record{
		sidechainRecord(t, "agent-nobody-recorded", "2026-01-01T00:10:30Z", 7),
	}

	attributeInEveryOrder(t, windows, records, func(t *testing.T, deltas map[int64]harvest.TokenDelta) {
		if got := deltas[2].Sidechain.Input; got != 7 {
			t.Errorf("dispatch 2 input = %v, want 7 -- no agent id matches, so the interval pass places it on the one window that contains it", got)
		}
		if got := deltas[1].Sidechain.Input; got != 0 {
			t.Errorf("dispatch 1 input = %v, want 0", got)
		}
	})
}

// TestDispatchWithNoAgentIDReceivesRecordsByTheWindowRule is the mirror,
// and the defect a naive `record.AgentID == window.AgentID` comparison
// produces. Cursor and Codex expose no subagent identifier at all, so a
// dispatch recorded without one is ordinary rather than degraded, and an
// absent id is "" meaning "not reported" -- never a value that matches
// another absent one.
//
// Both subtests use non-overlapping windows, for the same reason
// TestRecordAgentIDMatchingNoDispatchFallsBackToTheWindowRule does: two
// windows both containing the record is the ambiguous case
// (TestDispatchAmbiguousOverlapAttributesToNone), not this one. Each is
// still built so that treating "" as a match would change the answer: the
// window carrying no agent id is the one that would be reached by an
// identifier comparison first, so an implementation that lets ""=="" win
// would pull the record onto it and away from the window the interval rule
// alone would place it on.
func TestDispatchWithNoAgentIDReceivesRecordsByTheWindowRule(t *testing.T) {
	t.Run("a record with no agent id does not match a dispatch with none", func(t *testing.T) {
		windows := []harvest.DispatchWindow{
			{DispatchID: 1, AgentID: "", StartedAt: mustParse(t, "2026-01-01T00:10:00Z"), EndedAt: ptrTime(mustParse(t, "2026-01-01T00:10:15Z"))},
			{DispatchID: 2, AgentID: "agent-slot-b", StartedAt: mustParse(t, "2026-01-01T00:10:15Z"), EndedAt: ptrTime(mustParse(t, "2026-01-01T00:11:00Z"))},
		}
		records := []harvest.Record{sidechainRecord(t, "", "2026-01-01T00:10:30Z", 4)}

		attributeInEveryOrder(t, windows, records, func(t *testing.T, deltas map[int64]harvest.TokenDelta) {
			if got := deltas[2].Sidechain.Input; got != 4 {
				t.Errorf("dispatch 2 input = %v, want 4 -- two absent agent ids are not a match, so the interval pass decides", got)
			}
			if got := deltas[1].Sidechain.Input; got != 0 {
				t.Errorf("dispatch 1 input = %v, want 0 -- an absent agent id means \"not reported\", never \"matches the empty agent id\"", got)
			}
		})
	})

	t.Run("a dispatch with no agent id is still reachable by the window rule", func(t *testing.T) {
		windows := []harvest.DispatchWindow{
			{DispatchID: 1, AgentID: "agent-slot-a", StartedAt: mustParse(t, "2026-01-01T00:10:00Z"), EndedAt: ptrTime(mustParse(t, "2026-01-01T00:10:15Z"))},
			{DispatchID: 2, AgentID: "", StartedAt: mustParse(t, "2026-01-01T00:10:15Z"), EndedAt: ptrTime(mustParse(t, "2026-01-01T00:11:00Z"))},
		}
		records := []harvest.Record{sidechainRecord(t, "agent-unrelated", "2026-01-01T00:10:30Z", 6)}

		attributeInEveryOrder(t, windows, records, func(t *testing.T, deltas map[int64]harvest.TokenDelta) {
			if got := deltas[2].Sidechain.Input; got != 6 {
				t.Errorf("dispatch 2 input = %v, want 6 -- a dispatch recorded with no agent id still receives records", got)
			}
		})
	})
}

// TestDispatchIdentityBeatsInterval is the kan-286 case: -ended-at is typed
// by the dispatching agent, so a record legitimately falls outside a window
// that is approximate by construction. A record whose agent id matches a
// dispatch must attribute to it even when the record's own timestamp falls
// outside that dispatch's interval entirely.
func TestDispatchIdentityBeatsInterval(t *testing.T) {
	windows := []harvest.DispatchWindow{
		{
			DispatchID: 1,
			AgentID:    "agent-one",
			StartedAt:  mustParse(t, "2026-01-01T10:00:00Z"),
			EndedAt:    ptrTime(mustParse(t, "2026-01-01T10:05:00Z")),
		},
	}
	records := []harvest.Record{
		sidechainRecord(t, "agent-one", "2026-01-01T10:07:00Z", 42),
	}

	a := harvest.NewDispatchAttributor(&fakeDispatchWindowSource{
		bySession: map[string][]harvest.DispatchWindow{mainSessionID: windows},
	})
	deltas, _, err := a.Attribute(context.Background(), records)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	if got := deltas[1].Sidechain.Input; got != 42 {
		t.Errorf("dispatch 1 input = %v, want 42 -- the record's agent id matches this dispatch even though its timestamp falls two minutes outside the dispatch's own interval", got)
	}
}

// TestDispatchDuplicateAgentIDAttributesToNeither is the identity pass's own
// ambiguous case: two dispatches recording the same agent id, with a record
// carrying that id landing inside both windows. A tie broken by interval or
// by row order would reintroduce the same silent misattribution this change
// removes, one layer down -- so the returned map must be empty, not resolved
// by either fallback.
func TestDispatchDuplicateAgentIDAttributesToNeither(t *testing.T) {
	windows := []harvest.DispatchWindow{
		panelWindow(t, 1, "agent-one", "2026-01-01T00:10:00Z"),
		panelWindow(t, 2, "agent-one", "2026-01-01T00:10:00.5Z"),
	}
	records := []harvest.Record{
		sidechainRecord(t, "agent-one", "2026-01-01T00:10:30Z", 5),
	}

	attributeInEveryOrder(t, windows, records, func(t *testing.T, deltas map[int64]harvest.TokenDelta) {
		if len(deltas) != 0 {
			t.Fatalf("deltas = %v, want empty: two dispatches record the same agent id, so the identity pass is ambiguous and must not fall back to the interval or to row order", deltas)
		}
	})
}

// TestDispatchAmbiguousOverlapAttributesToNone is the kan-295 case in
// miniature: three windows with byte-identical intervals and no agent id on
// either side, so the identity pass finds no candidate and the interval pass
// finds three. The returned map must be empty -- not that the
// latest-started window wins, which is exactly the silent misattribution
// this change removes.
func TestDispatchAmbiguousOverlapAttributesToNone(t *testing.T) {
	windows := []harvest.DispatchWindow{
		{DispatchID: 21, StartedAt: mustParse(t, "2026-01-01T00:10:00Z"), EndedAt: ptrTime(mustParse(t, "2026-01-01T00:11:00Z"))},
		{DispatchID: 22, StartedAt: mustParse(t, "2026-01-01T00:10:00Z"), EndedAt: ptrTime(mustParse(t, "2026-01-01T00:11:00Z"))},
		{DispatchID: 23, StartedAt: mustParse(t, "2026-01-01T00:10:00Z"), EndedAt: ptrTime(mustParse(t, "2026-01-01T00:11:00Z"))},
	}
	records := []harvest.Record{sidechainRecord(t, "", "2026-01-01T00:10:30Z", 3)}

	attributeInEveryOrder(t, windows, records, func(t *testing.T, deltas map[int64]harvest.TokenDelta) {
		if len(deltas) != 0 {
			t.Fatalf("deltas = %v, want empty: three windows contain the record and none carries an agent id, so the interval pass is ambiguous too -- not that the latest-started window wins", deltas)
		}
	})
}

// TestDispatchNonOverlappingIntervalStillAttributes is the guard against
// over-correcting: the interval rule is the fallback, not a legacy path. A
// single containing window with no agent id on either side still attributes
// -- it is the only rule available on Cursor and Codex, and it is exactly
// correct wherever windows do not overlap.
func TestDispatchNonOverlappingIntervalStillAttributes(t *testing.T) {
	windows := []harvest.DispatchWindow{
		{DispatchID: 31, StartedAt: mustParse(t, "2026-01-01T00:10:00Z"), EndedAt: ptrTime(mustParse(t, "2026-01-01T00:11:00Z"))},
	}
	records := []harvest.Record{sidechainRecord(t, "", "2026-01-01T00:10:30Z", 9)}

	a := harvest.NewDispatchAttributor(&fakeDispatchWindowSource{
		bySession: map[string][]harvest.DispatchWindow{mainSessionID: windows},
	})
	deltas, _, err := a.Attribute(context.Background(), records)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	if got := deltas[31].Sidechain.Input; got != 9 {
		t.Errorf("dispatch 31 input = %v, want 9 -- one containing window with no agent id on either side must still attribute", got)
	}
}
