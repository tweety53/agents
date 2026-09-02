# kan-377-flow-cut-run-cost-and-time-conductor-subagent

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

> **Relocation:** no

Eight tasks. Four are Go under `stats/` (TDD, RED before GREEN): the `conductor` role, the
harvester's message-id dedupe, the fable pricing row, and interval narrowing of an id-ambiguous
dispatch window set. Four are skill prose under `skills/flow/` and `skills/flow-contracts/`: the
panel re-run rules, the report/plan-field/trim changes, the per-task review overlap, and the
conductor dispatch itself. `design.md` is canonical for every rule the prose tasks write and for
the decisions; `proposal.md` for why. Every `go` command below runs from `stats/`; every guard
from the repository root. No prose task names a test file, so each carries `**Tests:** none`.

**Baseline, measured before any edit:**

- `cmd/flow/record_test.go`: 40 tests.
  <!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ main (d4d6cbe) -->
- `internal/harvest/transcript_test.go`: 15 tests.
  <!-- measured: grep -c '^func Test' stats/internal/harvest/transcript_test.go @ main (d4d6cbe) -->
- `internal/store/pricing_seed_test.go`: 3 tests.
  <!-- measured: grep -c '^func Test' stats/internal/store/pricing_seed_test.go @ main (d4d6cbe) -->
- `internal/harvest/attribute_test.go`: 34 tests.
  <!-- measured: grep -c '^func Test' stats/internal/harvest/attribute_test.go @ main (d4d6cbe) -->
- The three harvest fixtures carry one distinct `message.id` per assistant line (`main-thread.jsonl`
  7 of 7, `sidechain.jsonl` 4 of 4), so the dedupe changes no existing count.
  <!-- measured: grep -c '"type":"assistant"' and grep -o '"id":"msg_[^"]*"' | sort -u | wc -l on stats/testdata/transcripts/*.jsonl @ main (d4d6cbe) -->
- `scripts/check-contract-budget.sh` rows against current sizes: `skills/flow/SKILL.md` 16278
  (file 15742 bytes), `skills/flow/implement.md` 20877 (18520 bytes), `skills/flow/review-panel.md`
  35985 (35853 bytes), `skills/flow/verify-and-handoff.md` 24862 (21408 bytes),
  `skills/flow-contracts/pipeline.md` 36155 (28623 bytes), `skills/flow-contracts/model-policy.md`
  8010 (6515 bytes), `commands/flow.md` 3874 (2896 bytes), `commands-claude/flow.md` 3188 (2684
  bytes). Tasks 7 and 8 grow `implement.md` past its row and task 8 grows `SKILL.md` past its; each
  raises the row in its own commit — a budget raise is a visible diff by that guard's design.
  <!-- measured: grep in scripts/check-contract-budget.sh and wc -c on each file @ main (d4d6cbe) -->
- The Targeted/Full ladder region of `review-panel.md` (its lines from the mode table to the end of
  the cap-check section) carries no `SHALL`/`MUST` sentence, so deleting it changes
  `check-normative-inventory.sh`'s set only by what a task deliberately adds.
  <!-- measured: awk 'NR>=385 && NR<=460 && /SHALL|MUST/' skills/flow/review-panel.md @ main (d4d6cbe) -->

- [x] 1. `conductor` joins the dispatch role list

`recordRoles` in `stats/cmd/flow/record.go` is the closed list the CLI checks before contacting the
store; the store validates no role. Add `conductor` beside `planner`, and update every place the
list is spelled out: the comment above `recordRoles`, the usage text line `-role is one of: …`.

**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordAcceptsConductorRole`; `TestRecordRejectsUnknownRoleWithoutContactingStore`
extended so its accepted-role loop also expects `conductor`.
**Regression:** `TestRecordAcceptsConductorRole` — reverting this commit makes `record dispatch
begin -role conductor` exit 2 without contacting the store, so the test's exit-code check fails;
the extended loop fails on the missing role name in stderr.
**Baseline:** before=40 after=41
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ main (d4d6cbe); after predicted: one new test function -->
**Commit:** `feat(flow-cli): accept the conductor dispatch role`
**Build:** green

  - [x] **Step 1: Write the failing test**

Beside `TestRecordAcceptsPlannerRole` (`record_test.go`), add a copy for the conductor and extend
the rejection test's loop:

```go verified:copied from TestRecordAcceptsPlannerRole in stats/cmd/flow/record_test.go with the role, model and token changed
// TestRecordAcceptsConductorRole pins that "conductor" is in the
// accepted-role allowlist -- the role `/flow` records for the subagent
// that runs implement.md, review-panel.md and verify-and-handoff.md as
// the conductor (design.md's conductor-runs-implementation-half).
func TestRecordAcceptsConductorRole(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"seq":1,"role":"conductor","model":"sonnet","startedAt":"2026-01-02T03:04:05Z"}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"record", "dispatch", "begin", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-change", "kan-258", "-role", "conductor", "-model", "sonnet", "-key", "conductor",
			"-session-token", "mf-record-conductor-role", "-started-at", "2026-01-02T03:04:05Z"},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if !contacted {
		t.Error("the store was not contacted for the conductor role")
	}
}
```

In `TestRecordRejectsUnknownRoleWithoutContactingStore`, change the loop's list to
`[]string{"implementer", "reviewer", "panel-fix", "red-partner", "planner", "conductor"}`.

  - [x] **Step 2: Run the tests and confirm they fail**

Run: `go test ./cmd/flow -run 'TestRecordAcceptsConductorRole|TestRecordRejectsUnknownRoleWithoutContactingStore' -v`
Expected: FAIL — `exit code = 2, want 0` and `stderr does not name the accepted role "conductor"`.

  - [x] **Step 3: Add the role**

```go verified:stats/cmd/flow/record.go line 41 at main (d4d6cbe), with the new element appended
var recordRoles = []string{"implementer", "reviewer", "panel-fix", "red-partner", "planner", "conductor"}
```

Extend the comment above it: `conductor` is the subagent `/flow` dispatches to run
`skills/flow/implement.md`, `review-panel.md` and `verify-and-handoff.md` (design.md's
`conductor-runs-implementation-half`). Update the usage line `-role is one of: implementer,
reviewer, panel-fix, red-partner, planner.` to end `…, planner, conductor.`

  - [x] **Step 4: Run the tests and confirm they pass, then the package checks**

Run: `go test ./cmd/flow -run 'TestRecordAcceptsConductorRole|TestRecordRejectsUnknownRoleWithoutContactingStore' -v`
Expected: PASS.
Run: `gofmt -l . && go vet ./...`
Expected: no output from `gofmt -l`, exit 0 from `vet`.

  - [x] **Step 5: Commit**

```bash verified:the commit subject is this task's own Commit: field
git add cmd/flow/record.go cmd/flow/record_test.go
git commit -m "feat(flow-cli): accept the conductor dispatch role" -m "Task-Id: 1"
```

- [x] 2. Harvester deduplicates usage by `message.id`

Claude Code writes one JSONL line per content block of one API response, each line carrying the
same `message.id` and identical `usage`. `ParseAssistantRecords` emits one Record per line, so
every token figure in the store is inflated by the block count. Keep the first Record per
non-empty `message.id` within one call.

**Files:** `stats/internal/harvest/transcript.go`, `stats/internal/harvest/transcript_test.go`,
`stats/testdata/transcripts/multi-block.jsonl`
**Tests:** `TestParseAssistantRecordsDeduplicatesByMessageID`
**Regression:** reverting this commit makes the fixture parse to five records instead of three, so
the length assertion fails.
**Baseline:** before=15 after=16
<!-- measured: grep -c '^func Test' stats/internal/harvest/transcript_test.go @ main (d4d6cbe); after predicted: one new test function -->
**Commit:** `fix(harvest): count one usage record per message id, not per transcript line`
**Build:** green

  - [x] **Step 1: Write the fixture**

Create `stats/testdata/transcripts/multi-block.jsonl` with exactly these five lines — three lines
sharing `msg_dup` with identical usage and different block types (the real shape: thinking, text,
tool_use), one line with no `id` at all, one line with a second id:

```json verified:field names read from rawLine/rawMessage/rawUsage in stats/internal/harvest/transcript.go and from a live transcript line read during planning on 2026-09-02
{"type":"assistant","timestamp":"2026-01-01T00:00:01Z","sessionId":"s-multi","isSidechain":false,"message":{"id":"msg_dup","model":"claude-sonnet-5","content":[{"type":"thinking","thinking":""}],"usage":{"input_tokens":5,"cache_creation_input_tokens":0,"cache_read_input_tokens":100,"output_tokens":7}}}
{"type":"assistant","timestamp":"2026-01-01T00:00:01Z","sessionId":"s-multi","isSidechain":false,"message":{"id":"msg_dup","model":"claude-sonnet-5","content":[{"type":"text","text":"hi"}],"usage":{"input_tokens":5,"cache_creation_input_tokens":0,"cache_read_input_tokens":100,"output_tokens":7}}}
{"type":"assistant","timestamp":"2026-01-01T00:00:01Z","sessionId":"s-multi","isSidechain":false,"message":{"id":"msg_dup","model":"claude-sonnet-5","content":[{"type":"tool_use","name":"Bash","input":{"command":"true"}}],"usage":{"input_tokens":5,"cache_creation_input_tokens":0,"cache_read_input_tokens":100,"output_tokens":7}}}
{"type":"assistant","timestamp":"2026-01-01T00:00:02Z","sessionId":"s-multi","isSidechain":false,"message":{"model":"claude-sonnet-5","content":[{"type":"text","text":"no id"}],"usage":{"input_tokens":11,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"output_tokens":1}}}
{"type":"assistant","timestamp":"2026-01-01T00:00:03Z","sessionId":"s-multi","isSidechain":false,"message":{"id":"msg_two","model":"claude-sonnet-5","content":[{"type":"text","text":"second"}],"usage":{"input_tokens":13,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"output_tokens":2}}}
```

  - [x] **Step 2: Write the failing test**

In `transcript_test.go`, beside `TestParseAssistantRecordsCarriesAgentID`:

```go verified:readFixture and mainThreadFixture are the file's own helpers (transcript_test.go lines 14-19 at main d4d6cbe); the fixture path follows their relative form
const multiBlockFixture = "../../testdata/transcripts/multi-block.jsonl"

// TestParseAssistantRecordsDeduplicatesByMessageID pins that a
// multi-block API response -- one JSONL line per content block, every
// line carrying the same message.id and identical usage -- is counted
// once, and that a line carrying no id at all is never collapsed into a
// neighbour (design.md's harvest-dedupe-by-message-id).
func TestParseAssistantRecordsDeduplicatesByMessageID(t *testing.T) {
	complete, _ := harvest.SplitCompleteLines(readFixture(t, multiBlockFixture))
	records := harvest.ParseAssistantRecords(complete)
	if len(records) != 3 {
		t.Fatalf("got %d records, want 3 (three lines share msg_dup, one has no id, one is msg_two)", len(records))
	}
	if got := records[0].Usage.InputTokens; got != 5 {
		t.Errorf("first record input = %d, want 5 (msg_dup, counted once)", got)
	}
	if got := records[1].Usage.InputTokens; got != 11 {
		t.Errorf("second record input = %d, want 11 (the id-less line is kept)", got)
	}
	if got := records[2].Usage.InputTokens; got != 13 {
		t.Errorf("third record input = %d, want 13 (msg_two)", got)
	}
}
```

  - [x] **Step 3: Run the test and confirm it fails**

Run: `go test ./internal/harvest -run TestParseAssistantRecordsDeduplicatesByMessageID -v`
Expected: FAIL — `got 5 records, want 3`.

  - [x] **Step 4: Implement the dedupe**

In `transcript.go`, add `ID string `json:"id"`` to `rawMessage`, and in `ParseAssistantRecords`
declare `seen := map[string]bool{}` before the loop and, immediately after the timestamp parse
succeeds:

```go verified:written against ParseAssistantRecords in stats/internal/harvest/transcript.go at main (d4d6cbe)
		// One API response is written as one line per content block, each
		// line repeating the same message.id and the same usage (measured:
		// 56 assistant lines for 15 ids on one real transcript). The first
		// line per id is the whole message's usage; the rest are repeats.
		// A line with no id is an older or foreign shape and is kept as is.
		// ponytail: dedupe is per call, so a message whose lines straddle
		// two reads counts twice -- lines of one message land in one write
		// burst, so this is bounded by one message per read boundary; carry
		// the last id beside harvest_offsets if it ever shows up in the data.
		if id := raw.Message.ID; id != "" {
			if seen[id] {
				continue
			}
			seen[id] = true
		}
```

Update the doc comment on `ParseAssistantRecords` (one sentence: one Record per `message.id`, not
per line) and the package doc's "one usage record per line" wording where it appears.

  - [x] **Step 5: Run the harvest tests, then the package checks**

Run: `go test ./internal/harvest`
Expected: PASS — the existing fixtures carry one id per line, so no count changes.
Run: `gofmt -l . && go vet ./...`
Expected: clean.

  - [x] **Step 6: Commit**

```bash verified:the commit subject is this task's own Commit: field
git add internal/harvest/transcript.go internal/harvest/transcript_test.go testdata/transcripts/multi-block.jsonl
git commit -m "fix(harvest): count one usage record per message id, not per transcript line" -m "Task-Id: 2"
```

- [x] 3. Pricing row for `claude-fable-5-1`

The operator's session model has no rate row, so a stage run whose `models` bucket includes fable
prices that bucket as unavailable and the top-level `cost_usd` is omitted. Add the published row.

**Files:** `stats/internal/store/pricing_seed.go`, `stats/internal/store/pricing_seed_test.go`
**Tests:** `TestSeedPricingRatesCarryFable`; `TestSeedPricingRatesOmitFastForModelsWithNone`
extended so its no-fast-rate case also names `claude-fable-5-1`.
**Regression:** `TestSeedPricingRatesCarryFable` — reverting this commit removes the row, so the
lookup fails; `TestSeedPricingRatesRoundTrip` iterates every seeded row and so covers the new one
without change.
**Baseline:** before=3 after=4
<!-- measured: grep -c '^func Test' stats/internal/store/pricing_seed_test.go @ main (d4d6cbe); after predicted: one new test function -->
**Commit:** `feat(stats): price claude-fable-5-1`
**Build:** green

  - [x] **Step 1: Write the failing test**

In `pricing_seed_test.go`:

```go verified:PricingRate field names read from the struct literals in stats/internal/store/pricing_seed.go at main (d4d6cbe)
// TestSeedPricingRatesCarryFable pins the row for the operator's own
// session model -- absent, every fable-orchestrated run priced its parent
// at zero (design.md's fable-pricing-row). Rates are the published table
// read on 2026-09-02; the cache-read rate is the one column that is not
// 0.1x input on this model.
func TestSeedPricingRatesCarryFable(t *testing.T) {
	var fable *store.PricingRate
	for _, rate := range store.SeedPricingRates() {
		if rate.Model == "claude-fable-5-1" {
			rate := rate
			fable = &rate
		}
	}
	if fable == nil {
		t.Fatal("no seeded rate for claude-fable-5-1")
	}
	if fable.InputPerMTok != 10 || fable.OutputPerMTok != 50 {
		t.Errorf("input/output = %v/%v, want 10/50", fable.InputPerMTok, fable.OutputPerMTok)
	}
	if fable.CacheWrite5mPerMTok != 12.5 || fable.CacheWrite1hPerMTok == nil || *fable.CacheWrite1hPerMTok != 20 {
		t.Errorf("cache writes = %v/%v, want 12.5/20", fable.CacheWrite5mPerMTok, fable.CacheWrite1hPerMTok)
	}
	if fable.CacheReadPerMTok != 0.25 {
		t.Errorf("cache read = %v, want 0.25", fable.CacheReadPerMTok)
	}
	if fable.FastInputPerMTok != nil || fable.FastOutputPerMTok != nil {
		t.Error("fast rate present, want nil (fast mode is Opus 5 / 4.8 only)")
	}
}
```

In `TestSeedPricingRatesOmitFastForModelsWithNone`, change the first case to
`case "claude-sonnet-5", "claude-haiku-4-5", "claude-fable-5-1":`.

  - [x] **Step 2: Run the test and confirm it fails**

Run: `go test ./internal/store -run TestSeedPricingRatesCarryFable -v`
Expected: FAIL — `no seeded rate for claude-fable-5-1`. (The store tests need the dev Postgres on
port 5433 for the round-trip test; this one does not touch the store.)

  - [x] **Step 3: Add the row**

Insert before the `claude-opus-5` literal in `SeedPricingRates`:

```go verified:rates read from https://platform.claude.com/docs/en/about-claude/pricing on 2026-09-02; struct shape copied from the claude-opus-5 literal in stats/internal/store/pricing_seed.go
		{
			Model:               "claude-fable-5-1",
			EffectiveFrom:       pricingSeedEffectiveFrom,
			InputPerMTok:        10,
			OutputPerMTok:       50,
			CacheWritePerMTok:   12.50, // legacy column; superseded by the 5m/1h split below.
			CacheWrite5mPerMTok: 12.50,
			CacheWrite1hPerMTok: fastRate(20),
			CacheReadPerMTok:    0.25, // 0.025x input on this model, not the 0.1x every other row uses.
			// No fast-mode rate published for this model (Opus 5 / 4.8 only).
		},
```

Extend the doc comment on `SeedPricingRates`: the fable row was read from the same page on
2026-09-02.

  - [x] **Step 4: Run the store tests, then the package checks**

Run: `go test ./internal/store -run 'TestSeedPricing' -v`
Expected: PASS, the round-trip subtest `claude-fable-5-1` included (dev Postgres on 5433 must be
up; if it is not, the round-trip test skips or fails for that reason alone — say so).
Run: `gofmt -l . && go vet ./...`
Expected: clean.

  - [x] **Step 5: Commit**

```bash verified:the commit subject is this task's own Commit: field
git add internal/store/pricing_seed.go internal/store/pricing_seed_test.go
git commit -m "feat(stats): price claude-fable-5-1" -m "Task-Id: 3"
```

- [x] 4. `bestDispatchWindow` narrows an id-ambiguous window set by interval

A resumed fix dispatch shares its agent id with the original implementer dispatch, so the id pass
finds two windows and today declares ambiguity without looking at the timestamp. Filter the
id-matched set by `contains(ts)` first.

**Files:** `stats/internal/harvest/attribute.go`, `stats/internal/harvest/attribute_test.go`
**Tests:** `TestDispatchSameAgentIDNarrowedByInterval`, `TestDispatchSameAgentIDOutsideBothStaysAmbiguous`
**Regression:** `TestDispatchSameAgentIDNarrowedByInterval` — reverting this commit leaves both
same-id windows as candidates, so nothing is attributed and the delta assertion fails;
`TestDispatchSameAgentIDOutsideBothStaysAmbiguous` guards the other direction and passes before and
after.
**Baseline:** before=34 after=36
<!-- measured: grep -c '^func Test' stats/internal/harvest/attribute_test.go @ main (d4d6cbe); after predicted: two new test functions -->
**Commit:** `fix(harvest): narrow an id-ambiguous dispatch window set by interval`
**Build:** green

  - [x] **Step 1: Write the failing tests**

Beside `TestDispatchAmbiguousOverlapAttributesToNone` in `attribute_test.go`, using that file's
own `sidechainRecord(t, agentID, at, input)`, `mustParse`, `ptrTime`, `attributeInEveryOrder`
and `fakeDispatchWindowSource` helpers:

```go verified:helper signatures read from stats/internal/harvest/attribute_test.go at main (d4d6cbe): sidechainRecord line 1283, attributeInEveryOrder line 1300, ptrTime line 168
// TestDispatchSameAgentIDNarrowedByInterval is kan-374's own case: a
// task-N-implementer-fix-1 dispatch resumed the implementer and so
// carries its agent id, giving two windows for one id. The record's
// timestamp falls inside exactly one of them, and that one wins
// (design.md's agent-id-on-begin).
func TestDispatchSameAgentIDNarrowedByInterval(t *testing.T) {
	windows := []harvest.DispatchWindow{
		{DispatchID: 41, AgentID: "a1", StartedAt: mustParse(t, "2026-01-01T00:10:00Z"), EndedAt: ptrTime(mustParse(t, "2026-01-01T00:11:00Z"))},
		{DispatchID: 42, AgentID: "a1", StartedAt: mustParse(t, "2026-01-01T00:20:00Z"), EndedAt: ptrTime(mustParse(t, "2026-01-01T00:21:00Z"))},
	}
	records := []harvest.Record{sidechainRecord(t, "a1", "2026-01-01T00:20:30Z", 5)}

	attributeInEveryOrder(t, windows, records, func(t *testing.T, deltas map[int64]harvest.TokenDelta) {
		if got := deltas[42].Sidechain.Input; got != 5 {
			t.Fatalf("dispatch 42 input = %v, want 5 -- two windows share the id and only 42 contains the timestamp", got)
		}
		if _, ok := deltas[41]; ok {
			t.Fatalf("dispatch 41 received usage; the interval narrowed the id match to 42 alone")
		}
	})
}

// TestDispatchSameAgentIDOutsideBothStaysAmbiguous guards the other
// direction: an id-matched record inside neither window is still refused,
// with both windows reported as the candidates.
func TestDispatchSameAgentIDOutsideBothStaysAmbiguous(t *testing.T) {
	windows := []harvest.DispatchWindow{
		{DispatchID: 41, AgentID: "a1", StartedAt: mustParse(t, "2026-01-01T00:10:00Z"), EndedAt: ptrTime(mustParse(t, "2026-01-01T00:11:00Z"))},
		{DispatchID: 42, AgentID: "a1", StartedAt: mustParse(t, "2026-01-01T00:20:00Z"), EndedAt: ptrTime(mustParse(t, "2026-01-01T00:21:00Z"))},
	}
	records := []harvest.Record{sidechainRecord(t, "a1", "2026-01-01T00:15:00Z", 5)}

	a := harvest.NewDispatchAttributor(&fakeDispatchWindowSource{
		bySession: map[string][]harvest.DispatchWindow{mainSessionID: windows},
	})
	deltas, ambiguous, err := a.Attribute(context.Background(), records)
	if err != nil {
		t.Fatalf("Attribute: %v", err)
	}
	if len(deltas) != 0 {
		t.Fatalf("deltas = %v, want empty", deltas)
	}
	if len(ambiguous) != 1 || len(ambiguous[0].DispatchIDs) != 2 {
		t.Fatalf("ambiguous = %+v, want one entry naming dispatches 41 and 42", ambiguous)
	}
}
```

  - [x] **Step 2: Run the tests and confirm the first fails**

Run: `go test ./internal/harvest -run 'TestDispatchSameAgentID' -v`
Expected: `TestDispatchSameAgentIDNarrowedByInterval` FAIL (`dispatch 42 input = 0, want 5`);
`TestDispatchSameAgentIDOutsideBothStaysAmbiguous` PASS.

  - [x] **Step 3: Narrow the id-matched set**

Replace the `default:` arm of the `len(byID)` switch in `bestDispatchWindow` (`attribute.go`):

```go verified:written against bestDispatchWindow in stats/internal/harvest/attribute.go lines 835-867 at main (d4d6cbe)
		default:
			// More than one window carries this id -- a dispatch resumed via
			// SendMessage shares the original's agent id. The timestamp
			// decides where the id cannot: exactly one containing window
			// wins; otherwise the narrowed set (or, if none contains ts, the
			// whole id-matched set) is the ambiguity reported.
			var narrowed []DispatchWindow
			for _, w := range byID {
				if w.contains(ts) {
					narrowed = append(narrowed, w)
				}
			}
			if len(narrowed) == 1 {
				return narrowed[0], true, nil
			}
			if len(narrowed) > 1 {
				return DispatchWindow{}, false, narrowed
			}
			return DispatchWindow{}, false, byID
```

Update `bestDispatchWindow`'s doc comment to state the narrowing.

  - [x] **Step 4: Run the harvest tests, then the package checks**

Run: `go test ./internal/harvest`
Expected: PASS — the existing ambiguity test uses windows with no agent id and is untouched by the
id pass.
Run: `gofmt -l . && go vet ./...`
Expected: clean.

  - [x] **Step 5: Commit**

```bash verified:the commit subject is this task's own Commit: field
git add internal/harvest/attribute.go internal/harvest/attribute_test.go
git commit -m "fix(harvest): narrow an id-ambiguous dispatch window set by interval" -m "Task-Id: 4"
```

- [x] 5. Panel re-runs: the Minor rule and delta re-runs replace the Targeted/Full ladder

`design.md` §2 is canonical for every rule below; this task writes it into
`skills/flow/review-panel.md`'s **Panel re-runs** section and rewords the three "stale" lines.

**Files:** `skills/flow/review-panel.md`, `skills/flow/SKILL.md`, `skills/flow/verify-and-handoff.md`
**Tests:** none — prose; verification is the guard list in step 5 and the acceptance grep.
**Regression:** reverting this commit brings the mode table and the five triggers back, so
`grep -n "Targeted\|Full re-run" skills/flow/review-panel.md` prints lines again.
**Baseline:** before=0 after=0 (no test-count-bearing files touched)
**Commit:** `feat(flow): replace the panel's Targeted/Full ladder with delta re-runs and the Minor rule`
**Build:** green

  - [x] **Step 1: Capture the normative inventory**

Run: `scripts/check-normative-inventory.sh > /tmp/kan-377-normative-before.txt`
Expected: exit 0. The region this task deletes carries no `SHALL`/`MUST` sentence (baseline above),
so the after-capture in step 5 may differ only by sentences this task adds.

  - [x] **Step 2: Rewrite Panel re-runs**

In `review-panel.md`'s `## Panel re-runs`, keep its first four paragraphs unchanged (pass 1 runs
the roster; `FIX_BASE` and `fix-round-N.diff`; fixup + autosquash; the clean-rebase warning). Delete
everything from the `| Mode | Who re-runs | Diff they get |` table through the end of
`### The diff-size cap check on a re-run` (its last paragraph begins "**On a Targeted re-run**"),
and put this in its place:

```markdown verified:authored in-tree for this change from design.md §2
**Which slots re-run, and on what, follows from the severities the round raised — never from a
mode table, a trigger list, or a round count.**

**A Minor finding blocks and is fixed, and triggers no re-run.** Every finding the round raised goes
to the fix subagent below; each is closed by the verification that follows it — the reproducer
re-run exits 0 *and* the fix diff touches a path the finding named. When every finding the round
raised was Minor, no slot re-runs: proceed to `check-panel-findings-closed.sh` and the stage close.
A finding that fails verification takes the handback below, and that loop re-runs no slot either.

**When the round raised anything above Minor, re-run on deltas.** A delta is `git diff <the HEAD
sha that slot last reviewed> HEAD`, written to `<abs-worktree>/.superpowers/sdd/slot-delta-<round>-<slot>.diff`.
Each dispatch sets that slot's sha to the HEAD it was dispatched against, and a slot not dispatched
in a round keeps the sha it had; a slot for which no last-reviewed sha is held reads the whole
`final-review.diff`. Every slot's dispatch prompt names the path it was given and, for a delta, the
sha that delta starts from. Then:

- **every diff-reading slot** in the resolved roster — Primary included, reading a delta like the
  rest — plus every operator-added slot already dispatched in an earlier pass of this run, re-runs
  on its delta. **A slot whose delta is empty is not dispatched**, and the record states `not re-run
  — nothing new since its last read`;
- **Bugbot and Security**, which read no diff file, re-run only when that slot raised a finding in
  the previous round or the previous round raised a new Critical;
- **a slot the operator has not named for this run is never added here** — that addition happens
  only through the explicit-request check **The roster** states, at the start of any round.

**The cap check on a re-run** is `check-panel-diff-size.sh <worktree> <sha> <cap>` once per
**distinct** last-reviewed sha among the diff-reading slots dispatched this round (two slots
sharing a sha need one call, not two); a slot with no held sha counts as the merge base. **The
gating count is the largest of those counts** — the largest single read any one slot this round
faces — and an exit-1 result from the call that produced it puts the over-cap choice to the
operator (**The roster**, above), naming the gating count and, when it differs, the full-branch
count too. Record both in `<abs-worktree>/.superpowers/sdd/final-review-panel.md` for this round,
alongside the agents-ran/why/diff-path fields the fix pass records.

Handoff still requires **zero open findings at any severity** from every agent that has run, and
no stale result — where **a slot's clean result is stale when the rule above required that slot to
re-run and it has not**. A non-Minor fix Bugbot did not raise leaves Bugbot's result current: the
round's own mutation-proof (below) covers what the fix changed.
```

Then, further down the same section: change "A minor finding blocks the handoff exactly as a
critical one does." to "A minor finding blocks the handoff exactly as a critical one does, and
triggers no re-run — see above." Change the stage's closing line "Once this stage ends clean —
zero open findings at any severity, no stale result — continue into" to read "…, no stale result
as **Panel re-runs** defines it — continue into".

  - [x] **Step 3: Reword the two guardrails**

`skills/flow/SKILL.md`, the guardrail beginning "**Never** hand off with an open finding of any
severity, or a stale clean result": append " — stale as **Panel re-runs**
(`skills/flow/review-panel.md`) defines it: a slot the re-run rule required that did not run".
`skills/flow/verify-and-handoff.md`, the guardrail of the same wording: the same append.

  - [x] **Step 4: The acceptance grep**

Run: `grep -n "Targeted\|Full re-run\|Full mode\|Escalate automatically\|coverage waiver" skills/flow/review-panel.md`
Expected: no output.

  - [x] **Step 5: Run the guards**

Run, from the repository root: `scripts/check-normative-inventory.sh | diff /tmp/kan-377-normative-before.txt -`
Expected: only added lines (`>`), each a sentence step 2 wrote; a removed line (`<`) is a sentence
to restore. Then `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
`scripts/check-contract-budget.sh`, `scripts/check-markdown-integrity.py`,
`scripts/check-stage-mark-calls.sh`, `scripts/check-dispatch-paragraphs.sh`,
`scripts/check-installed-citations.sh` — each exit 0. `review-panel.md` shrinks under this task,
so its budget row is untouched.

  - [x] **Step 6: Commit**

```bash verified:the commit subject is this task's own Commit: field
git add skills/flow/review-panel.md skills/flow/SKILL.md skills/flow/verify-and-handoff.md
git commit -m "feat(flow): replace the panel's Targeted/Full ladder with delta re-runs and the Minor rule" -m "Task-Id: 5"
```

- [x] 6. Slots write their own reports; the fix subagent maintains plan fields; orchestrator trims

`design.md` §2's **Reports**, **Plan fields** and **Trims** paragraphs, written into
`review-panel.md` and `implement.md`.

**Files:** `skills/flow/review-panel.md`, `skills/flow/implement.md`, `scripts/check-contract-budget.sh`
**Tests:** none — prose; verification is the guard list in step 5.
**Regression:** reverting this commit puts the dispatcher's byte-for-byte report re-emission back
and drops the plan-field instruction from the fix subagent's prompt, so the "write that report byte
for byte" sentence reappears in `review-panel.md`.
**Baseline:** before=0 after=0 (no test-count-bearing files touched)
**Commit:** `feat(flow): slots write their own panel reports and the fix subagent maintains plan fields`
**Build:** green

  - [x] **Step 1: Capture the normative inventory**

Run: `scripts/check-normative-inventory.sh > /tmp/kan-377-normative-before.txt`

  - [x] **Step 2: Reports — the slot writes, the dispatcher checks**

In `review-panel.md`, under the slot-prompt paragraphs (beside the REPRODUCE, DON'T READ paragraph
every slot carries), add a paragraph every slot's dispatch prompt must carry:

```markdown verified:authored in-tree for this change from design.md §2
**Every slot's dispatch prompt also carries the REPORT FILE paragraph**, with the round and the
slot's resolved id substituted:

> **REPORT FILE:** write your full report to
> `<abs-worktree>/.superpowers/sdd/panel-report-<round>-<id>.md` before you end your turn — every
> finding with its `file:line`, severity, the sentence naming the defect, and its reproducer. Your
> return message is the findings summary; the file is the record.
```

Replace the first paragraph of `## Recording findings, and the record's format` ("**Beside the
existing dispatch recording**: as each slot's report comes back … `no verbatim report captured —
<reason>`.") with:

```markdown verified:authored in-tree for this change from design.md §2
**The slot writes its own report.** Each slot writes
`<abs-worktree>/.superpowers/sdd/panel-report-<round>-<id>.md` itself, per the REPORT FILE paragraph
its prompt carries — `<round>` the same value that round's findings carry on `-round` (`0` initial,
`1..n` fix rounds), `<id>` the resolved reviewer id, never the slot display name. As each slot's
`flow record dispatch end` is recorded, confirm the file exists and is non-empty (`test -s`); when
it is not, write it yourself carrying the single line `no verbatim report captured — <reason>`.
Every dispatched slot ends up with one, a slot that raised nothing and a substituted slot
included. **Never re-emit a slot's report from this context** — record its `F<n>` rows and cite the
file.
```

  - [x] **Step 3: The fix subagent maintains plan fields**

In `review-panel.md`, beside the VERBATIM REPORT — THE FACT paragraph the fix subagent's prompt
carries, add:

```markdown verified:authored in-tree for this change from design.md §2
**Every fix subagent's dispatch prompt also carries the PLAN FIELDS paragraph**:

> **PLAN FIELDS:** when your fix changes what a task's `**Baseline:**` counts or `**Files:**` paths
> declare — a test case added, a file created — update that field in the worktree's
> `<project>/spectre/changes/<name>/tasks.md` in this same pass. Edit it; do not stage or commit
> it — it is a planning path and is committed later by the pipeline, never in a fixup.
```

  - [x] **Step 4: Trims — pass paths, never read subagent-facing files into this context**

In `review-panel.md`'s **The roster** section, after the roster table, add one paragraph:

```markdown verified:authored in-tree for this change from design.md §2
**A subagent-facing file is passed by absolute path, never read into this context.** Superpowers'
`code-reviewer.md` (Primary), `principles-reviewer-prompt.md` and `engineering-principles.md`
(Principles), and `<project>/.flow/project.md`'s standards files are inputs to the slot that reads
them; the dispatcher resolves their paths, confirms each exists, and names them in the prompt.
```

In `implement.md`'s bundle paragraph (section 4, after "Confirm the bundle was actually written"),
add: "**Never read the bundle back into this context** — `test -f` is the whole check; its
content is the implementer's input, not the dispatcher's."

  - [x] **Step 5: Run the guards**

Run: `scripts/check-normative-inventory.sh | diff /tmp/kan-377-normative-before.txt -`
Expected: only lines this task added. Then `scripts/check-dispatch-paragraphs.sh` (the existing
sites and minimums are untouched), `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
`scripts/check-contract-budget.sh`, `scripts/check-markdown-integrity.py`,
`scripts/check-stage-mark-calls.sh` — each exit 0. `review-panel.md` after tasks 5 and 6 is
predicted under its 35985 row; if `check-contract-budget.sh` reports otherwise, raise the row to
the new size plus a quarter in this commit.
<!-- predicted: wc -c skills/flow/review-panel.md after this task's edits -->

  - [x] **Step 6: Commit**

```bash verified:the commit subject is this task's own Commit: field
git add skills/flow/review-panel.md skills/flow/implement.md
git commit -m "feat(flow): slots write their own panel reports and the fix subagent maintains plan fields" -m "Task-Id: 6"
```

- [x] 7. Per-task review overlapped with the next implementer; `-agent-id` on `begin`; reviewer outcome word

`design.md` §3 and §4's first bullet are canonical. This task rewrites `implement.md` section 4's
loop and moves `-agent-id` onto `begin` in every record block of `implement.md` and in
`review-panel.md`'s panel-fix block.

**Files:** `skills/flow/implement.md`, `skills/flow/review-panel.md`
**Tests:** none — prose; verification is the guard list in step 5.
**Regression:** reverting this commit restores the strictly serial "waits for the previous
implementer's commit sha … before dispatching the next" loop and `-agent-id goes on end`.
**Baseline:** before=0 after=0 (no test-count-bearing files touched)
**Commit:** `feat(flow): overlap per-task review with the next implementer`
**Build:** green

  - [x] **Step 1: Capture the normative inventory**

Run: `scripts/check-normative-inventory.sh > /tmp/kan-377-normative-before.txt`

  - [x] **Step 2: Record blocks — `-agent-id` on `begin`**

In `implement.md` section 4, change the implementer pair so `begin` carries `-agent-id <id>` and
replace the paragraph "**Both calls are required, and `begin` must go out BEFORE the dispatch does
— never delayed to obtain an identifier.**" through "`-agent-id` goes on `end` here — a serialized
implementer dispatch reports its own identifier only once it comes back." with:

```markdown verified:authored in-tree for this change from design.md §4; the flag set is record.go's own usage (stats/cmd/flow/record.go lines 84-88 at main d4d6cbe)
**Both calls are required. Every launch is asynchronous and returns the agent's identifier at
launch, so `begin` carries `-agent-id <id>` and is recorded immediately after the launch returns,
before any other action; `end` may repeat the id.** `-key` is this dispatch's own literal label,
unique within the run's session token — `task-<n>-implementer`, reused identically in both calls.
`-role` is one of `implementer`, `reviewer`, `panel-fix` or `red-partner`; `-task` is the task's
flat integer id, omitted for a dispatch against no single task; `-started-at`/`-ended-at` are
RFC 3339 — `-started-at` the launch time. `-session-token` takes a literal, never a shell
substitution. Two dispatches starting at one instant are told apart only by id, and a resumed
dispatch shares its id with the original — which is why the id is recorded at launch.
```

Apply the same move — `-agent-id <id>` on `begin`, kept on `end` — to the per-task reviewer pair
(section 4, "Record the reviewer's dispatch too") and, in `review-panel.md`, to the panel-fix
pair, deleting its trailing "`-agent-id` goes on `end`." sentence.

  - [x] **Step 3: The overlapped loop**

Replace `implement.md`'s paragraph "**At most one implementer subagent may be in flight against a
given worktree at any moment.** The parent waits for the previous implementer's commit sha …" with:

```markdown verified:authored in-tree for this change from design.md §3
**At most one implementer subagent may be in flight against a given worktree at any moment** —
a reviewer is not one: it reads an immutable commit range, and any number of them may run beside
the one implementer. Dispatches into different worktrees remain free to run concurrently. This
explicitly overrides `superpowers:subagent-driven-development`'s parallel dispatch guidance and
`superpowers:dispatching-parallel-agents` for same-worktree tasks.
```

Replace the paragraph "**Guard the commit before dispatching review.**" through "**Per-task
review:** … a step's checkbox tracks the step and gates nothing." (keeping the two required
FOREGROUND BUILDS and REPRODUCE, DON'T READ reviewer blockquotes that follow it exactly where they
are) with:

````markdown verified:authored in-tree for this change from design.md §3
**Review overlaps the next implementer.** The unit is the bundle `plan-dispatch-bundles.sh` emits;
"bundle N's reviewers" is one reviewer per task in it. At each boundary, in this order:

1. **Bundle N+1's implementer commits** and reports its shas.
2. **A pending fix for bundle N folds in first.** If bundle N's review raised a fix, resume that
   bundle's implementer (`SendMessage`; record the resumption as its own pair under
   `task-<n>-implementer-fix-<k>`, `-agent-id` the implementer's own id) with the reviewer's
   report path; it commits `git commit --fixup=<task-sha>` and runs `git rebase --autosquash`. A
   conflict there is between two of the branch's own commits and the implementer resolves it —
   `skills/flow/integrate.md`'s never-auto-resolve rule concerns the operator's base branch.
3. **Guard every commit whose sha is new** — bundle N+1's tasks and every task the rebase rewrote —
   before dispatching review for it, passing the canonical worktree's absolute path (the worktree
   created or resumed in step **2** above) as the guard's fifth argument and this run's resolved
   `<name>` as its sixth:

   ```bash
   check-task-commit-fields.sh <worktree> <task-id> <task-sha> <task-base> <canonical-worktree> <name>
   ```

   The guard reads git objects and `tasks.md` only, so it is safe while the tree changes. A
   nonzero exit is a guard failure, not a review finding — it does **not** consume a fix-round
   slot; send the task back to the **same implementer**, then re-run the guard, before anything
   below.
4. **Dispatch together:** bundle N's re-reviewer if step 2 ran (the task's full range,
   `git diff <task-base>..<new-task-sha>`), bundle N+1's reviewers, and bundle N+2's implementer.

**When the script cannot be located**, apply `flow-task-commit-fields`'s rules by hand: check the
commit's `Files:` against `git diff --name-only <task-base>..<task-sha>`, its `Tests:` against the
commit's diff, and its `Commit:` against the commit's actual subject line.

**Per-task review:** the reviewer gets the commit-range diff `git diff <task-base>..<task-sha>` —
a real commit diff, never a snapshot of the working tree, which the next implementer is editing.
**Record the reviewer's dispatch too** — `-role reviewer`, the same `-task <n>`, `-model
DEFAULT_MODEL`, `-agent-id` on `begin` — and close it with **`-outcome clean` or `-outcome fix`**,
so per-task review yield is measurable. **The per-task review is a single combined reviewer**,
covering spec compliance and code quality together, dispatched on `DEFAULT_MODEL` — `/flow`'s
panel carries no roster, so there is no `full`-preset split into two per-task reviewers. Mark a
**task's** checkbox `[x]` (`flow tasks tick`) only after that task passes spec **and** quality
review — the guard has already passed by then; a step's checkbox tracks the step and gates nothing.

**The last bundle's reviewers run alone.** Overlap them only with the review panel's pre-work —
its citation check, bundle rebuild, relocation comparison and diff-size check; `final-review.diff`
is written and the slots dispatched once the last review is clean and any fix folded.

**Never end a turn with a child in flight** — wait for every implementer and reviewer launched
before reporting a stage boundary or asking the operator anything.
````

  - [x] **Step 4: Raise the budget row**

`implement.md` grows past its 20877 row under this task. Run `wc -c skills/flow/implement.md`,
and set the `skills/flow/implement.md` row in `scripts/check-contract-budget.sh` to that size plus a
quarter, rounded up to the next hundred.
<!-- predicted: wc -c skills/flow/implement.md after this task's edits, about 21000 bytes -->

  - [x] **Step 5: Run the guards**

Run: `scripts/check-normative-inventory.sh | diff /tmp/kan-377-normative-before.txt -`
Expected: only lines this task added or deliberately removed (the retired "begin must go out
BEFORE" sentence carries no keyword). Then `scripts/check-dispatch-paragraphs.sh` — `implement.md`
still carries its two FOREGROUND BUILDS blocks and both REPRODUCE, DON'T READ variants;
`scripts/check-stage-mark-calls.sh` — every record block keeps its literal session token;
`scripts/check-vocabulary.sh`, `scripts/check-references.sh`, `scripts/check-contract-budget.sh`,
`scripts/check-markdown-integrity.py`, `scripts/test-check-contract-budget.sh` — each exit 0.

  - [x] **Step 6: Commit**

```bash verified:the commit subject is this task's own Commit: field
git add skills/flow/implement.md skills/flow/review-panel.md scripts/check-contract-budget.sh
git commit -m "feat(flow): overlap per-task review with the next implementer" -m "Task-Id: 7"
```

- [x] 8. The conductor dispatch: implement.md, SKILL.md routing, the `## Handoff` return, progress visibility

`design.md` §1 is canonical. The parent's side lives in a new **Dispatch the conductor** section at
the head of `implement.md`; the conductor's side is the existing sections, addressed as "you".
Section 3 stays where it is and gains one leading line saying it runs in the parent first.

**Files:** `skills/flow/implement.md`, `skills/flow/SKILL.md`, `skills/flow/verify-and-handoff.md`,
`skills/flow-contracts/pipeline.md`, `skills/flow-contracts/model-policy.md`, `commands/flow.md`,
`commands-claude/flow.md`, `scripts/check-contract-budget.sh`
**Tests:** none — prose; verification is the guard list in step 8.
**Regression:** reverting this commit removes the **Dispatch the conductor** section, so
`grep -n "role conductor" skills/flow/implement.md` prints nothing and the router again runs the
implementation half in the parent.
**Baseline:** before=0 after=0 (no test-count-bearing files touched)
**Commit:** `feat(flow): run the implementation half in a conductor subagent`
**Build:** green

  - [x] **Step 1: Capture the normative inventory**

Run: `scripts/check-normative-inventory.sh > /tmp/kan-377-normative-before.txt`

  - [x] **Step 2: Write Dispatch the conductor**

In `implement.md`, after the opening skill table and its "**Never** invoke
`finishing-a-development-branch`" line and before `## 1. Load context and validate the plan`,
insert:

````markdown verified:authored in-tree for this change from design.md §1; the flag set is record.go's own usage (stats/cmd/flow/record.go lines 84-88 at main d4d6cbe) plus the conductor role task 1 adds
## Dispatch the conductor

Sections **1**, **2** and **4** below, `skills/flow/review-panel.md` and
`skills/flow/verify-and-handoff.md` are the **conductor's** work, not the parent's — one conductor
subagent runs `flow.load-context` through `flow.write-in-progress`, resumed between its returns so
its context carries from the plan to the handoff. Every "you" in those files addresses it. The
parent's own work on this branch is what this section states, plus — on a fix run — section **3**
below, which runs **before** the dispatch: the parent resolves the worktree from the state file's
`worktrees` map, runs section 3's planner dispatch and Jira sync, and only then dispatches the
conductor. A fix run's stage order is therefore document-fix → load-context → isolate (resume) →
sdd-tdd → …, so the appended plan is validated after the fix's edit.

**Resolve `DEFAULT_MODEL` and `REVIEWERS`** per **Model resolution** (`skills/flow/SKILL.md`),
and run the guard-presence check, before dispatching. Dispatch one subagent with the Agent tool's
`model` parameter set to `DEFAULT_MODEL` — or the run's plain-language session override, recorded
with the dispatch — and `subagent_type: general-purpose`. Its prompt carries, verbatim:

> Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
> Include this instruction verbatim in any prompt you write for another agent.

and states: the change name `<name>`; the project root; `<changeRoot>`; this run's literal session
token; the harness; `DEFAULT_MODEL` and the resolved `REVIEWERS` list; the guard-presence result;
the run kind (creating or fix) and, on a fix run, the operator's fix instructions; and the
instruction to read this file's sections **1**, **2** and **4**, `skills/flow/review-panel.md` and
`skills/flow/verify-and-handoff.md` and follow them **as the conductor**, running every stage mark
in that range itself with the token it was given.

**The relay contract**, stated in the same prompt. The conductor has no channel to the operator and
no task-list tool. It ends a turn only with one of three blocks, and never with a child subagent
still in flight — it waits for every implementer, reviewer, slot and fix subagent it launched
first:

- `## Question` — the question plus named options; the parent asks it verbatim through
  **AskUserQuestion** and resumes the conductor via **SendMessage** with the answer. Every operator
  prompt inside the covered stages goes this way: the over-cap choice, a second wall-clock breach,
  the non-converging-finding handback, BLOCKED, an empty resolved-worktree set, a plan-quality
  repair that needs the operator.
- `## Stage flow.<key>` plus one line of outcome, at every `flow stage end` it runs; the parent
  updates the harness task list — the stage is the granularity on this branch, per **Progress
  visibility** (`skills/flow-contracts/pipeline.md`) — and resumes it with `continue`.
- `## Handoff` carrying the `IN_PROGRESS` handoff block verbatim; the parent prints it unchanged.

The first line of its first reply is `Model: <the model named in its own system prompt>`.

**Record the dispatch immediately after the launch returns its identifier**, before anything else:

```bash
flow record dispatch begin -change <name> -role conductor -model <DEFAULT_MODEL> \
  -key conductor -agent-id <id> -session-token mf-<literal-token> -started-at <ts>
```

**The handshake.** Compare the `Model:` line against `DEFAULT_MODEL` (or the override). A match
proceeds. A mismatch records the model that answered and **continues on the running agent — no
re-dispatch**: the planner's opus re-dispatch (**Dispatch the planner**,
`skills/flow/brainstorm.md`) exists because fable may be unavailable, which `DEFAULT_MODEL` does
not share, and a re-dispatch onto a costlier model would raise what this dispatch exists to cut:

```bash
flow record dispatch end -change <name> -key conductor -session-token mf-<literal-token> \
  -outcome fallback -ended-at <ts>
flow record dispatch begin -change <name> -role conductor -model <the model the handshake named> \
  -key conductor-<that model, lowercased> -agent-id <id> -session-token mf-<literal-token> -started-at <ts>
```

**A mark or a record never blocks** — proceed on the handshake's outcome regardless of whether any
`flow` call reached the store.

**The return.** Once `## Handoff` arrives, print the block unchanged and close the record under
whichever key is open:

```bash
flow record dispatch end -change <name> -key <the key currently open> -session-token mf-<literal-token> \
  -outcome completed -ended-at <ts>
```

The run is at `IN_PROGRESS`; nothing further runs in this invocation.

**A conductor that ends without one of the three blocks, or whose agent dies, is closed with
`-outcome aborted`, reported, and not retried**: print `/flow <name>` for the operator — a re-run
resumes from whatever the conductor left (checkbox state, the state file's worktrees, findings in
the store) through this file's own re-entry rules, and the operator should see the death rather
than have it hidden by a second dispatch.

**At depth, the Agent tool offers no `bugbot` or `security-review` type.** **An unspawnable id is
substituted, not skipped** (`skills/flow/review-panel.md`) applies unchanged; under the conductor
it is the norm, not the exception.
````

Add to the head of `## 3. Documenting a fix, before implementing it`, after its heading: "**Parent
work, run before the conductor is dispatched** — see **Dispatch the conductor** above. Everything
below is the parent's own; the conductor never sees this section."

  - [x] **Step 3: Route both run kinds through it**

`skills/flow/SKILL.md`, **Reading the state**: on the creating-run bullet, after "See **A. Resolve
the change and write `STARTED`** (`skills/flow/brainstorm.md`)", add "; once `## Plan` returns,
**Dispatch the conductor** (`skills/flow/implement.md`)". On the fix-run bullet, replace "See **3.
Documenting a fix, before implementing it** (`skills/flow/implement.md`)" with "See **3.
Documenting a fix, before implementing it** and then **Dispatch the conductor**
(`skills/flow/implement.md`)". Add a guardrail bullet: "**Never** run `implement.md` sections 1, 2
or 4, `review-panel.md` or `verify-and-handoff.md` in the parent session — the conductor runs
them; the parent dispatches it, relays its questions and prints its handoff."

`skills/flow/verify-and-handoff.md`, at the end of `## Write IN_PROGRESS` after the handoff block's
explanatory paragraphs, add: "**This block is returned, not printed, by the conductor** — its
turn ends with `## Handoff` and the block verbatim beneath it, and the parent prints it unchanged
(**Dispatch the conductor**, `skills/flow/implement.md`)."

  - [x] **Step 4: Progress visibility**

`skills/flow-contracts/pipeline.md`, `## Progress visibility`, after its first paragraph, add:
"**On the implementation branch the granularity is the stage.** The stages from
`flow.load-context` to `flow.write-in-progress` run inside a conductor subagent (**Dispatch the
conductor**, `skills/flow/implement.md`) that has no task-list tool; the parent updates the list at
each `## Stage` return it relays — one entry per stage, not per `tasks.md` item. The per-task
granularity named above is what a harness with the conductor in the parent session would show; this
is the accepted visibility trade of design.md's `conductor-stage-returns` in the change that
introduced it." Adjust the first paragraph's "`tasks.md` items on the implementation branch" to
"stages on the implementation branch".

  - [x] **Step 5: The role list and the two command files**

`skills/flow-contracts/model-policy.md`: the sentence listing "the implementer, panel and planner"
dispatches gains "and conductor". `commands/flow.md` and `commands-claude/flow.md`: in the sentence
describing what a creating run does after brainstorming, add "then implementation, review and
verification in a conductor subagent on the default model, the parent relaying its questions".

  - [x] **Step 6: Raise the budget rows**

Run `wc -c skills/flow/implement.md skills/flow/SKILL.md` and set each file's row in
`scripts/check-contract-budget.sh` to its size plus a quarter, rounded up to the next hundred,
where the size exceeds the row. `pipeline.md`, `verify-and-handoff.md`, `model-policy.md` and the
two command files are predicted under their rows.
<!-- predicted: wc -c on each file after this task's edits; implement.md about 25000 bytes, SKILL.md about 16800 bytes -->

  - [x] **Step 7: The acceptance greps**

Run: `grep -n "role conductor" skills/flow/implement.md`
Expected: the two `begin` lines from step 2.
Run: `grep -n "Dispatch the conductor" skills/flow/SKILL.md skills/flow/verify-and-handoff.md skills/flow-contracts/pipeline.md`
Expected: one hit per file at least.

  - [x] **Step 8: Run the guards**

Run: `scripts/check-normative-inventory.sh | diff /tmp/kan-377-normative-before.txt -`
Expected: only lines this task added. Then every guard `<project>/.flow/project.md`'s `## lint`
names, in order — `scripts/check-vocabulary.sh`, `scripts/check-references.sh` (every bold
citation above names a real heading), `scripts/check-plan-provenance.sh`,
`scripts/check-task-build-green.sh`, `scripts/check-plan-shape.sh`,
`scripts/check-contract-budget.sh`, `scripts/check-markdown-integrity.py`,
`scripts/check-stage-mark-calls.sh` (the conductor's `begin` lines carry the literal token),
`scripts/check-guard-symlinks.sh`, `scripts/check-dispatch-paragraphs.sh`,
`scripts/check-installed-citations.sh`, `scripts/check-installed-rules.sh` — each exit 0; then
`scripts/test-check-contract-budget.sh` and `scripts/test-check-references.sh`.

  - [x] **Step 9: Commit**

```bash verified:the commit subject is this task's own Commit: field
git add skills/flow/implement.md skills/flow/SKILL.md skills/flow/verify-and-handoff.md \
  skills/flow-contracts/pipeline.md skills/flow-contracts/model-policy.md \
  commands/flow.md commands-claude/flow.md scripts/check-contract-budget.sh
git commit -m "feat(flow): run the implementation half in a conductor subagent" -m "Task-Id: 8"
```
