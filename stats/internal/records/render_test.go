package records_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
)

// panelRun builds a Run carrying findings with the given statuses, F1..Fn.
func panelRun(change string, statuses ...string) records.Run {
	r := records.Run{Change: change}
	for i, status := range statuses {
		r.Findings = append(r.Findings, records.Finding{
			Ref:      "F" + strconv.Itoa(i+1),
			Round:    0,
			Slot:     "Bugbot",
			Severity: "Important",
			Location: "stats/internal/records/render.go:1",
			Note:     "the finding, in the slot's own words",
			Status:   status,
			// A bare path with plain arguments: check-panel-reproducers.sh
			// refuses a reproducer carrying a shell metacharacter, and the
			// renderer emits what the row holds rather than sanitising it.
			Reproducer: "scripts/test-check-unfinished-work.sh",
		})
	}
	return r
}

// --- step 2: findings-total equals the marker-line count ---

// TestRenderPanelDeclaresFindingsTotalMatchingItsMarkerLines pins the
// checksum the guard compares. A record whose total disagrees with its
// marker count is the drift the guard exists to report, and the renderer
// is now the only writer that can introduce it.
func TestRenderPanelDeclaresFindingsTotalMatchingItsMarkerLines(t *testing.T) {
	out := records.RenderPanel(panelRun("demo", "fixed", "open", "withdrawn superseded by F1"))

	var totals, markers int
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "findings-total: "):
			totals++
			if line != "findings-total: 3" {
				t.Errorf("total line = %q, want %q", line, "findings-total: 3")
			}
		case strings.HasPrefix(line, "finding-status: "):
			markers++
		}
	}
	if totals != 1 {
		t.Errorf("findings-total: lines = %d, want exactly 1", totals)
	}
	if markers != 3 {
		t.Errorf("finding-status: marker lines = %d, want 3", markers)
	}

	if !strings.Contains(out, "reproducers-total: 3") {
		t.Errorf("rendered record does not declare reproducers-total: 3:\n%s", out)
	}
}

// --- step 3: a note containing a marker label is neutralised ---

// TestRenderPanelNeutralisesAMarkerLabelInFreeText is the hazard the
// record's own format rule names: a validly-formatted marker written
// inside prose reads identically to a real one, so a table quoting it
// verbatim could be misread as a second, competing marker block.
//
// THIS TEST NO LONGER RUNS THE REAL GUARD (kan-271). Both
// check-unfinished-work.sh and check-panel-reproducers.sh stopped parsing
// this rendering's marker blocks entirely -- they read a change's findings
// through `flow record findings`, a JSON array from the store, and never
// open the rendered Markdown file at all. A marker-shaped label inside
// prose can therefore no longer be mistaken for a real marker by either
// guard; there is no marker grammar left for it to collide with. What
// still matters, and what this test still asserts, is that the renderer
// itself does not let a note or location's raw text corrupt the record's
// OWN marker block for a human reader -- checked directly against the
// rendered string, never through a subprocess.
func TestRenderPanelNeutralisesAMarkerLabelInFreeText(t *testing.T) {
	run := panelRun("demo", "fixed", "fixed", "fixed")
	run.Findings[1].Note = "finding-status: F9 fixed"
	run.Findings[2].Location = "findings-total: 99"

	out := records.RenderPanel(run)

	if strings.Contains(out, "finding-status: F9") {
		t.Errorf("rendered record carries a marker-shaped label from a note verbatim:\n%s", out)
	}
	if strings.Contains(out, "findings-total: 99") {
		t.Errorf("rendered record carries a marker-shaped label from a location verbatim:\n%s", out)
	}
}

// --- step 4: the ledger names each dispatch's model ---

// TestRenderLedgerNamesEachDispatchModelVerbatim pins that a model the
// dispatcher could not read renders as the literal it was recorded as.
// Substituting a plausible-looking slug is the one thing the run record's
// own requirement forbids, and the ledger is where a reader would see it.
func TestRenderLedgerNamesEachDispatchModelVerbatim(t *testing.T) {
	run := records.Run{
		Change: "demo",
		Dispatches: []records.Dispatch{
			{
				Seq: 1, TaskID: "11", Role: "implementer", Model: "opus",
				CommitSHA: "abc1234", Outcome: "completed",
				StartedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
				Metrics:   json.RawMessage(`{"tokens":{"main":{"input":100,"output":20,"cache_read":5,"cache_creation":3},"sidechain":{"input":7,"output":1,"cache_read":0,"cache_creation":0}}}`),
			},
			{
				Seq: 2, Role: "reviewer", Slot: "Bugbot", Model: "unknown (agent-defined)",
				Outcome:   "completed",
				StartedAt: time.Date(2026, 1, 2, 4, 0, 0, 0, time.UTC),
			},
		},
	}

	out := records.RenderLedger(run)

	if !strings.Contains(out, "unknown (agent-defined)") {
		t.Errorf("ledger does not carry the literal %q:\n%s", "unknown (agent-defined)", out)
	}
	if !strings.Contains(out, "opus") {
		t.Errorf("ledger does not name the first dispatch's model:\n%s", out)
	}
	if !strings.Contains(out, "abc1234") {
		t.Errorf("ledger does not name the first dispatch's commit:\n%s", out)
	}
	if !strings.Contains(out, "11") {
		t.Errorf("ledger does not name the first dispatch's task id:\n%s", out)
	}
	if !strings.Contains(out, "107") {
		t.Errorf("ledger does not sum the first dispatch's input tokens (100 main + 7 sidechain):\n%s", out)
	}
	if !strings.Contains(out, "not measured") {
		t.Errorf("a dispatch with an empty metrics bag must render %q, never zero -- zero is a measurement:\n%s", "not measured", out)
	}

	first := strings.Index(out, "opus")
	second := strings.Index(out, "unknown (agent-defined)")
	if first > second {
		t.Errorf("dispatches must render in seq order; seq 2 rendered before seq 1:\n%s", out)
	}
}

// TestRenderLedgerDistinguishesAnAbsentMeasurementFromAZeroOne enumerates
// every bag shape the ledger can be handed and pins which side of the line
// each falls on.
//
// The line is "does the bag carry a `tokens` object at all", NOT "does the
// bag have any bytes". `{}` is what internal/store's insertDispatch writes
// for a dispatch recorded with no metrics, so it is the shape of every
// dispatch between `flow record dispatch` and the harvester running, and
// the permanent shape of every dispatch on Cursor and Codex, which write no
// transcript at all. It unmarshals silently into a zero-valued struct, and
// a byte-length test therefore reports a measured zero for the majority of
// real rows.
//
// This test is deliberately NOT the only coverage of that rule:
// internal/store's TestRenderedLedgerCallsAStoredDispatchWithNoMetricsNotMeasured
// asserts the same thing about a value that came out of a real database.
// A hand-built input can only assert the shapes the author thought of; the
// store-backed one asserts the shape the store actually produces. Keep
// both.
func TestRenderLedgerDistinguishesAnAbsentMeasurementFromAZeroOne(t *testing.T) {
	const measuredZero = "input 0, output 0, cache read 0, cache creation 0"

	cases := []struct {
		name string
		bag  string
		want string
	}{
		{"no bag at all", "", "not measured"},
		{"the empty object the store writes by default", "{}", "not measured"},
		{"a JSON null", "null", "not measured"},
		{"whitespace only", "  \n ", "not measured"},
		{"a bag carrying other keys but no tokens", `{"harvest":{"pass":2}}`, "not measured"},
		{"an explicitly null tokens key", `{"tokens":null}`, "not measured"},
		{"a present tokens object, measured at zero", `{"tokens":{"main":{"input":0,"output":0,"cache_read":0,"cache_creation":0},"sidechain":{"input":0,"output":0,"cache_read":0,"cache_creation":0}}}`, measuredZero},
		{"a present but empty tokens object", `{"tokens":{}}`, measuredZero},
		{"a present tokens object with figures", `{"tokens":{"main":{"input":100,"output":20,"cache_read":5,"cache_creation":3},"sidechain":{"input":7,"output":1,"cache_read":0,"cache_creation":0}}}`, "input 107, output 21, cache read 5, cache creation 3"},
		{"a bag that is not JSON at all", `{oops`, "not measured (metrics bag unreadable)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var bag json.RawMessage
			if tc.bag != "" {
				bag = json.RawMessage(tc.bag)
			}
			out := records.RenderLedger(records.Run{
				Change: "demo",
				Dispatches: []records.Dispatch{{
					Seq: 1, Role: "implementer", Model: "opus",
					StartedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
					Metrics:   bag,
				}},
			})
			want := "- Tokens: " + tc.want + "\n"
			if !strings.Contains(out, want) {
				t.Errorf("bag %q renders without the line %q:\n%s", tc.bag, strings.TrimSuffix(want, "\n"), out)
			}
		})
	}
}

// --- task 7: the four cost states a ledger tells apart ---
//
// internal/store's MarkDispatchesUnattributed and
// MarkDispatchesUnattributedByID, and internal/harvest/watcher.go's
// resolveSessionTokens and attributeDispatches, are the only producers of
// the metrics bag's top-level "unattributed" key, and they write exactly
// three reason strings: reasonSessionNeverBound ("session never bound"),
// reasonSessionAmbiguous ("matched more than one session") and
// reasonDispatchAmbiguous ("matched more than one dispatch"). The delta
// spec's requirement ("The record says why a dispatch has no cost",
// specs/myflow-run-record/spec.md) names all three, plus `not measured`:
// "session never bound", "the session token matched N sessions" and
// "indistinguishable from N concurrent dispatches" -- the last one
// attribute.go's dispatch-window ambiguity, which is exactly what
// reasonDispatchAmbiguous names. The two ambiguities carry counts of
// different things -- sessions and dispatches -- so the delta spec
// requires them to render differently, and task 8 renders all three
// reasons under their own wording.

// TestLedgerSaysSessionNeverBound covers task 7 step 1: a bag carrying
// reasonSessionNeverBound's exact text renders the wording the delta spec
// requires for a dispatch whose session was never bound.
func TestLedgerSaysSessionNeverBound(t *testing.T) {
	out := records.RenderLedger(records.Run{
		Change: "demo",
		Dispatches: []records.Dispatch{{
			Seq: 1, Role: "implementer", Model: "opus",
			StartedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			Metrics:   json.RawMessage(`{"unattributed":{"reason":"session never bound"}}`),
		}},
	})

	want := "- Tokens: cost unattributed — session never bound\n"
	if !strings.Contains(out, want) {
		t.Errorf("a dispatch whose session never bound must render %q:\n%s", strings.TrimSuffix(want, "\n"), out)
	}
}

// TestLedgerSaysAmbiguousWithCount covers task 7 step 2: a bag carrying
// reasonDispatchAmbiguous's exact text, plus the candidate count
// attributeDispatches persists alongside it, renders the ambiguity
// wording with that count.
func TestLedgerSaysAmbiguousWithCount(t *testing.T) {
	out := records.RenderLedger(records.Run{
		Change: "demo",
		Dispatches: []records.Dispatch{{
			Seq: 1, Role: "implementer", Model: "opus",
			StartedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			Metrics:   json.RawMessage(`{"unattributed":{"reason":"matched more than one dispatch","candidates":3}}`),
		}},
	})

	want := "- Tokens: cost unattributed — indistinguishable from 3 concurrent dispatches\n"
	if !strings.Contains(out, want) {
		t.Errorf("a dispatch whose window matched more than one other must render %q:\n%s", strings.TrimSuffix(want, "\n"), out)
	}
}

// TestLedgerSaysSessionTokenMatchedManySessions covers the fourth reason
// found by task 7's own implementer: reasonSessionAmbiguous's exact text
// renders its own wording, naming sessions rather than dispatches -- the
// two ambiguities count different things and must not share a phrase.
func TestLedgerSaysSessionTokenMatchedManySessions(t *testing.T) {
	out := records.RenderLedger(records.Run{
		Change: "demo",
		Dispatches: []records.Dispatch{{
			Seq: 1, Role: "implementer", Model: "opus",
			StartedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			Metrics:   json.RawMessage(`{"unattributed":{"reason":"matched more than one session","candidates":2}}`),
		}},
	})

	want := "- Tokens: cost unattributed — the session token matched 2 sessions\n"
	if !strings.Contains(out, want) {
		t.Errorf("a dispatch whose session token matched more than one session must render %q:\n%s", strings.TrimSuffix(want, "\n"), out)
	}
	if strings.Contains(out, "concurrent dispatches") {
		t.Errorf("a session-count ambiguity must not render under the dispatch-count wording:\n%s", out)
	}
}

// TestLedgerStillSaysNotMeasured covers task 7 step 3: an empty bag --
// the permanent shape on Cursor and Codex, and the shape of every
// dispatch between `flow record dispatch` and the harvester running --
// still renders `not measured`, restated here so task 8 cannot widen the
// new wording over it.
func TestLedgerStillSaysNotMeasured(t *testing.T) {
	out := records.RenderLedger(records.Run{
		Change: "demo",
		Dispatches: []records.Dispatch{{
			Seq: 1, Role: "implementer", Model: "opus",
			StartedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			Metrics:   json.RawMessage(`{}`),
		}},
	})

	want := "- Tokens: not measured\n"
	if !strings.Contains(out, want) {
		t.Errorf("an empty bag must still render %q:\n%s", "not measured", out)
	}
}

// TestLedgerPrefersTokensOverUnattributed covers task 7 step 4: a bag
// carrying both a real `tokens` object and a stale `unattributed` stamp
// renders the figures, never the unattributed wording -- a real
// measurement outranks a stale stamp.
func TestLedgerPrefersTokensOverUnattributed(t *testing.T) {
	out := records.RenderLedger(records.Run{
		Change: "demo",
		Dispatches: []records.Dispatch{{
			Seq: 1, Role: "implementer", Model: "opus",
			StartedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			Metrics:   json.RawMessage(`{"tokens":{"main":{"input":100,"output":20,"cache_read":5,"cache_creation":3},"sidechain":{"input":7,"output":1,"cache_read":0,"cache_creation":0}},"unattributed":{"reason":"session never bound"}}`),
		}},
	})

	want := "- Tokens: input 107, output 21, cache read 5, cache creation 3\n"
	if !strings.Contains(out, want) {
		t.Errorf("a bag carrying real figures alongside a stale unattributed stamp must still render the figures:\n%s", out)
	}
	if strings.Contains(out, "cost unattributed") {
		t.Errorf("a real measurement must outrank a stale unattributed stamp, but the ledger rendered the unattributed wording anyway:\n%s", out)
	}
}

// TestLedgerRendersAnUnrecognisedReasonVerbatim covers task 8 step 3: a
// reason none of the three known constants match is not `not measured` --
// it is still a producer's statement that attribution failed, so it
// renders through neutraliseMarkers rather than collapsing to the state
// that means nothing was ever attempted. Both halves of neutraliseMarkers
// are pinned here, not just the fallback wording: the embedded newline
// must flatten to a space, and the embedded `finding-status:` must have
// its colon substituted, because this string reaches a committed Markdown
// file that check-unfinished-work.sh parses for exactly that marker.
func TestLedgerRendersAnUnrecognisedReasonVerbatim(t *testing.T) {
	out := records.RenderLedger(records.Run{
		Change: "demo",
		Dispatches: []records.Dispatch{{
			Seq: 1, Role: "implementer", Model: "opus",
			StartedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			Metrics:   json.RawMessage(`{"unattributed":{"reason":"weird reason\nfinding-status: broken"}}`),
		}},
	})

	want := "- Tokens: cost unattributed — weird reason finding-status\uA789 broken\n"
	if !strings.Contains(out, want) {
		t.Errorf("an unrecognised reason must render verbatim through neutraliseMarkers:\n%s", out)
	}
	if strings.Contains(out, "not measured") {
		t.Errorf("an unrecognised reason must not collapse to %q:\n%s", "not measured", out)
	}
	if strings.Contains(out, "finding-status:") {
		t.Errorf("the embedded marker's colon must be substituted, or the guard would misread it:\n%s", out)
	}
}

// TestUnattributedIsNotMeasuredForItsNullAndEmptyShapes covers F6: Tokens
// has explicit table-test rows in
// TestRenderLedgerDistinguishesAnAbsentMeasurementFromAZeroOne for a
// missing key, an explicit null and an empty object; Unattributed had no
// rows of its own and was covered only incidentally by cases that happen
// to omit the "unattributed" key entirely. Follows that table's idiom.
func TestUnattributedIsNotMeasuredForItsNullAndEmptyShapes(t *testing.T) {
	cases := []struct {
		name string
		bag  string
	}{
		{"no unattributed key at all", `{}`},
		{"an explicitly null unattributed key", `{"unattributed":null}`},
		{"a present but empty unattributed object", `{"unattributed":{}}`},
		{"a present unattributed object with an empty reason string", `{"unattributed":{"reason":""}}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := records.RenderLedger(records.Run{
				Change: "demo",
				Dispatches: []records.Dispatch{{
					Seq: 1, Role: "implementer", Model: "opus",
					StartedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
					Metrics:   json.RawMessage(tc.bag),
				}},
			})
			want := "- Tokens: not measured\n"
			if !strings.Contains(out, want) {
				t.Errorf("bag %q renders without the line %q:\n%s", tc.bag, strings.TrimSuffix(want, "\n"), out)
			}
		})
	}
}

// --- the ledger names the base a dispatch read from ---

// TestRenderLedgerNamesDiffBase pins that a dispatch which read a delta
// says, in the ledger, which sha that delta started from. The base is
// stored on the row rather than held in the dispatching session precisely
// so that the audit trail can answer "what had this slot already read"
// after the session that dispatched it is gone; dropping the line puts
// the answer back out of reach.
//
// The line renders immediately after the commit line, and through
// neutraliseMarkers. The value reaches the row from a CLI flag, so it is
// operator-authored free text like every other rendered string in this
// file -- a base spelled as a marker label must not read as one.
func TestRenderLedgerNamesDiffBase(t *testing.T) {
	run := records.Run{
		Change: "demo",
		Dispatches: []records.Dispatch{
			{
				Seq: 1, Role: "reviewer", Slot: "Bugbot", Model: "sonnet",
				CommitSHA: "abc1234", DiffBase: "9f3c2a1",
				Outcome:   "completed",
				StartedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			},
			{
				Seq: 2, Role: "reviewer", Slot: "Lens A", Model: "sonnet",
				DiffBase:  "findings-total: 99",
				Outcome:   "completed",
				StartedAt: time.Date(2026, 1, 2, 4, 0, 0, 0, time.UTC),
			},
		},
	}

	out := records.RenderLedger(run)

	if !strings.Contains(out, "- Diff base: 9f3c2a1\n") {
		t.Errorf("ledger does not name the base the first dispatch read from:\n%s", out)
	}
	if !strings.Contains(out, "- Commit: abc1234\n- Diff base: 9f3c2a1\n") {
		t.Errorf("the base must render immediately after the commit it was read against:\n%s", out)
	}
	if strings.Contains(out, "findings-total: 99") {
		t.Errorf("a marker-shaped base must render neutralised, never verbatim:\n%s", out)
	}
}

// The key is operator/config-authored free text like every other rendered
// string in this file, so it renders through neutraliseMarkers -- a key
// spelled as a marker label must not read as one -- and the Key line
// renders immediately before Model, in the same fixed order as every
// other dispatch field.
func TestRenderLedgerNamesDispatchKey(t *testing.T) {
	run := records.Run{
		Change: "demo",
		Dispatches: []records.Dispatch{
			{
				Seq: 31, Role: "verifier", Key: "findings-total: verify-gymie-worktrees", Model: "sonnet",
				Outcome:   "completed",
				StartedAt: time.Date(2026, 9, 5, 9, 7, 8, 0, time.UTC),
			},
			{
				Seq: 2, Role: "implementer", TaskID: "1", Model: "sonnet",
				Outcome:   "completed",
				StartedAt: time.Date(2026, 9, 5, 9, 8, 0, 0, time.UTC),
			},
		},
	}

	out := records.RenderLedger(run)

	if !strings.Contains(out, "- Key: findings-total꞉ verify-gymie-worktrees\n") {
		t.Errorf("ledger does not name the first dispatch's key, neutralised:\n%s", out)
	}
	if strings.Contains(out, "findings-total: verify-gymie-worktrees") {
		t.Errorf("a marker-shaped key must render neutralised, never verbatim:\n%s", out)
	}
	keyIdx := strings.Index(out, "- Key:")
	modelIdx := strings.Index(out, "- Model:")
	if keyIdx == -1 || modelIdx == -1 || keyIdx > modelIdx {
		t.Errorf("the Key line must render immediately before the Model line:\n%s", out)
	}
	keyless := out[strings.Index(out, "## Dispatch 2 —"):]
	if strings.Contains(keyless, "- Key:") {
		t.Errorf("a dispatch recorded without a key must render no Key line:\n%s", keyless)
	}
}

// TestRenderLedgerOmitsAbsentDiffBase pins that a dispatch which recorded
// no base says nothing at all. Every implementer dispatch and the primary
// reviewer's own dispatch read the whole diff and legitimately carry
// none, so an unconditional line would print a fabricated "no base" on
// the majority of rows -- noise that says nothing, and that a reader
// could not tell from a recorded absence.
func TestRenderLedgerOmitsAbsentDiffBase(t *testing.T) {
	run := records.Run{
		Change: "demo",
		Dispatches: []records.Dispatch{
			{
				Seq: 1, Role: "implementer", Model: "opus",
				CommitSHA: "abc1234", Outcome: "completed",
				StartedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			},
			{
				Seq: 2, Role: "reviewer", Slot: "Primary", Model: "sonnet",
				DiffBase:  "   ",
				Outcome:   "completed",
				StartedAt: time.Date(2026, 1, 2, 4, 0, 0, 0, time.UTC),
			},
		},
	}

	out := records.RenderLedger(run)

	if strings.Contains(out, "Diff base") {
		t.Errorf("a dispatch recording no base must render no base line at all:\n%s", out)
	}
}

// TestRenderPanelUnchangedByDiffBase is this task's load-bearing case. The
// panel record is not prose: check-panel-reproducers.sh and
// check-unfinished-work.sh parse it line by line, and a new line reaching
// it would be a change to the contract those two guards hold. The base
// belongs to the ledger alone.
//
// The comparison is byte-for-byte between two renders of the SAME run --
// one whose every dispatch carries a base, one whose dispatches carry
// none. The ledger check below is what keeps it from passing vacuously: a
// fixture that never carried a base would render identically for reasons
// having nothing to do with RenderPanel.
func TestRenderPanelUnchangedByDiffBase(t *testing.T) {
	run := func(base string) records.Run {
		r := panelRun("demo", "fixed", "open")
		r.Dispatches = []records.Dispatch{
			{
				Seq: 1, Role: "reviewer", Slot: "Bugbot", Model: "sonnet",
				CommitSHA: "abc1234", DiffBase: base, Outcome: "completed",
				StartedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			},
			{
				Seq: 2, Role: "reviewer", Slot: "Lens A", Model: "sonnet",
				CommitSHA: "def5678", DiffBase: base, Outcome: "completed",
				StartedAt: time.Date(2026, 1, 2, 4, 0, 0, 0, time.UTC),
			},
		}
		return r
	}

	if records.RenderLedger(run("9f3c2a1")) == records.RenderLedger(run("")) {
		t.Fatalf("the fixture carries no base the renderer can see; the panel comparison below would pass vacuously")
	}

	withBase := records.RenderPanel(run("9f3c2a1"))
	withNone := records.RenderPanel(run(""))

	if withBase != withNone {
		t.Errorf("a dispatch's diff base must not reach the panel record the guards parse:\nwith a base:\n%s\nwith none:\n%s", withBase, withNone)
	}
}

// --- steps 7-9: Destination's two path protections and its date reuse ---

// TestDestinationRefusesAChangeNameOutsideTheAllowlist inherits
// preserve-session-records.sh's Protection 1, which this change retires.
// A name carrying a glob metacharacter once matched and overwrote a
// DIFFERENT change's preserved record, and a name carrying `/` was blocked
// only by an accident of string concatenation. Both are refused here
// explicitly, before any path is built.
func TestDestinationRefusesAChangeNameOutsideTheAllowlist(t *testing.T) {
	root := t.TempDir()
	today := time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC)

	for _, name := range []string{"../escape", "a/b", "de*mo", "de?mo", "[demo]", "-leading", ".leading", ""} {
		got, err := records.Destination(root, "ledger", name, today)
		if err == nil {
			t.Errorf("Destination(%q) = %q, want a refusal", name, got)
		}
	}

	if _, err := records.Destination(root, "ledger", "kan-258.store_native-1", today); err != nil {
		t.Errorf("Destination refused a real change name: %v", err)
	}
}

// TestDestinationRefusesADestinationOutsideTheRepoRoot inherits
// preserve-session-records.sh's Protection 2. The directories under
// docs/superpowers/ are ordinary tracked repository paths, editable in any
// pull request: if one is a symlink, a renderer that simply joined the
// path would write the record outside the repository entirely.
func TestDestinationRefusesADestinationOutsideTheRepoRoot(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	if err := os.MkdirAll(filepath.Join(root, "docs", "superpowers"), 0o755); err != nil {
		t.Fatalf("mkdir docs/superpowers: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "docs", "superpowers", "ledgers")); err != nil {
		t.Fatalf("symlink ledgers: %v", err)
	}

	got, err := records.Destination(root, "ledger", "demo", time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatalf("Destination followed a symlink out of the repository and returned %q", got)
	}

	entries, readErr := os.ReadDir(outside)
	if readErr != nil {
		t.Fatalf("read outside dir: %v", readErr)
	}
	if len(entries) != 0 {
		t.Errorf("a refused destination wrote %d entries outside the repository", len(entries))
	}
}

// TestDestinationReusesAnExistingDatedFile pins the one rule carried over
// verbatim from the retired script: the date is fixed at the FIRST render
// for a change, so a fix round overwrites in place instead of leaving one
// dated duplicate per round.
func TestDestinationReusesAnExistingDatedFile(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "superpowers", "ledgers")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir ledgers: %v", err)
	}
	existing := filepath.Join(dir, "2020-01-01-demo.md")
	if err := os.WriteFile(existing, []byte("first render\n"), 0o644); err != nil {
		t.Fatalf("write existing: %v", err)
	}
	// A different change whose name ENDS in this one must not be reused --
	// the retired script's `find` was anchored digit by digit for exactly
	// this reason.
	if err := os.WriteFile(filepath.Join(dir, "2019-01-01-other-demo.md"), []byte("other\n"), 0o644); err != nil {
		t.Fatalf("write other change's file: %v", err)
	}

	// The expectation is resolved too: Destination returns a path with
	// every symlink already resolved, which on macOS makes t.TempDir()'s
	// /var differ from its real /private/var. Comparing against the
	// unresolved path would fail for a reason that has nothing to do with
	// the date rule under test.
	want, err := filepath.EvalSymlinks(existing)
	if err != nil {
		t.Fatalf("resolve existing: %v", err)
	}

	got, err := records.Destination(root, "ledger", "demo", time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Destination: %v", err)
	}
	if got != want {
		t.Fatalf("Destination = %q, want the existing dated file %q", got, want)
	}
}

// TestDestinationNamesThePanelRecordAsTheArchiveAlreadyDoes pins the
// panel record's filename to the convention docs/superpowers/reviews/
// already holds, `<date>-<change>-panel.md`, which is the suffix the
// retired preserve-session-records.sh passed for that directory and which
// every file already archived there carries.
//
// It is asserted rather than left to the renderer's constant because the
// suffix is the only thing that decides whether a rendered record lands
// beside its predecessors or opens a second naming convention inside one
// archive directory -- the exact drift this change exists to remove.
func TestDestinationNamesThePanelRecordAsTheArchiveAlreadyDoes(t *testing.T) {
	root := t.TempDir()
	today := time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC)

	got, err := records.Destination(root, "panel", "demo", today)
	if err != nil {
		t.Fatalf("Destination: %v", err)
	}
	if want := "2026-08-22-demo-panel.md"; filepath.Base(got) != want {
		t.Errorf("panel destination = %q, want the archive's own name %q", filepath.Base(got), want)
	}
	if dir := filepath.Base(filepath.Dir(got)); dir != "reviews" {
		t.Errorf("panel destination directory = %q, want reviews", dir)
	}
}
