# kan-357-record-each-dispatch-own-token-usage

> **Execution:** `/flow-fast` implements this plan. Mark a task's own
> checkbox when its commit lands and its verify step is green.
> **Relocation:** no

## Global constraints

- Every file lives under `stats/`; the targeted lint for touched Go is
  `cd stats && gofmt -l .` plus `cd stats && go vet ./...`.
- `internal/harvest` imports nothing from `internal/store`
  (`TestHarvestNeedsNoDatabase` rests on it); the new dependency follows
  `DispatchWindowSource`'s consumer-side interface pattern.
- Per-dispatch figures merge as batch deltas through `MergeDispatchMetrics`
  (additive, `jsonb_deep_add`), never as cumulative replacements.
- Stage-grain attribution (`Attributor`, `Watcher.RunOnce`'s first pass)
  is not modified by this plan.

---

- [x] 1. Store and Deps resolve dispatch windows by agent id

**Files:**
- `stats/internal/store/records.go`
- `stats/internal/store/records_test.go`
- `stats/internal/harvest/deps.go`

**Build:** green
**Tests:** `TestDispatchWindowsForAgentOrdersByStartedAt`
**Regression:** reverting this commit removes the only agent-keyed window
source; the tree still builds until task 2 lands (the method has no caller
yet), after which the watcher's routing no longer compiles.
**Baseline:** before=190 after=191
<!-- measured: go test ./internal/store/ -count=1 -v @ branch spectre/kan-357-record-each-dispatch-own-token-usage (worktree, from main e2d6b02) -->
<!-- predicted: go test ./internal/store/ -count=1 -v after this task -->
**Commit:** feat(stats): resolve dispatch windows by agent id

  - [ ] **Step 1: Write the failing store test**

Add to `stats/internal/store/records_test.go`, mirroring
`TestDispatchWindowsForSessionResolvesTheBoundToken`'s fixtures:

```go verified:authored in-tree for this change
// TestDispatchWindowsForAgentOrdersByStartedAt pins the query the
// agent-file attribution pass rests on: one resumed agent shares its
// agentId across several dispatch rows, and the split that separates
// their usage reads the rows back ordered by (started_at, id).
func TestDispatchWindowsForAgentOrdersByStartedAt(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-agent-windows-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	first := baseDispatch("implementer", "opus")
	first.AgentID = "a2ad390ad5a6c1507"
	firstRec, err := st.RecordDispatch(ctx, projectKey, "kan-1", first)
	if err != nil {
		t.Fatalf("RecordDispatch (first resume): %v", err)
	}
	second := baseDispatch("implementer", "opus")
	second.AgentID = "a2ad390ad5a6c1507"
	second.StartedAt = first.StartedAt.Add(10 * time.Minute)
	secondRec, err := st.RecordDispatch(ctx, projectKey, "kan-1", second)
	if err != nil {
		t.Fatalf("RecordDispatch (second resume): %v", err)
	}

	// A different agent's row, and one carrying no agent id at all:
	// neither is a window for this agent.
	other := baseDispatch("reviewer", "sonnet")
	other.AgentID = "a52d5a1bd2e13ac30"
	if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", other); err != nil {
		t.Fatalf("RecordDispatch (other agent): %v", err)
	}
	if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", baseDispatch("reviewer", "sonnet")); err != nil {
		t.Fatalf("RecordDispatch (no agent id): %v", err)
	}

	windows, err := st.DispatchWindowsForAgent(ctx, "a2ad390ad5a6c1507")
	if err != nil {
		t.Fatalf("DispatchWindowsForAgent: %v", err)
	}
	if len(windows) != 2 {
		t.Fatalf("windows = %+v, want the two rows carrying this agentId", windows)
	}
	if windows[0].DispatchID != firstRec.ID || windows[1].DispatchID != secondRec.ID {
		t.Errorf("windows ordered %d, %d — want started_at order %d, %d",
			windows[0].DispatchID, windows[1].DispatchID, firstRec.ID, secondRec.ID)
	}

	none, err := st.DispatchWindowsForAgent(ctx, "a-nobody-recorded")
	if err != nil {
		t.Fatalf("DispatchWindowsForAgent for an unknown agent: %v", err)
	}
	if len(none) != 0 {
		t.Errorf("windows for an unknown agent = %+v, want none", none)
	}
}
```

  - [ ] **Step 2: Run it to verify it fails**

Run: `cd stats && go test ./internal/store/ -run TestDispatchWindowsForAgentOrdersByStartedAt -count=1`
Expected: FAIL — `st.DispatchWindowsForAgent undefined`.

  - [ ] **Step 3: Implement the store query and the Deps composition**

In `stats/internal/store/records.go`, beside `DispatchWindowsForSession`
(records.go:738), whose SELECT shape this mirrors with the session-token
join replaced by an agent filter:

```sql unverified:confirm the dispatches column list against dispatchColumns in the same file
		SELECT d.id, d.started_at, d.ended_at, d.agent_id
		FROM dispatches d
		WHERE d.agent_id = $1 AND d.agent_id <> ''
		ORDER BY d.started_at, d.id
```

Scanned into `[]harvest.DispatchWindow` exactly as the sibling query does
(`DispatchID`, `StartedAt`, `EndedAt`, `AgentID`), returning the slice
directly so `*store.Store` satisfies the new interface with no adapter.

In `stats/internal/harvest/deps.go`: a new one-method consumer interface
beside `DispatchWindowSource` (watcher.go), composed into `Deps`, with a
`NoDeps` stub returning `(nil, nil)`:

```go verified:authored in-tree for this change
// AgentWindowSource answers which dispatch rows carry an agentId — the
// agent-file attribution pass's source. Returned directly as
// harvest.DispatchWindow for the same reason DispatchWindowSource's
// answer is: *store.Store then satisfies it with no adapter.
type AgentWindowSource interface {
	DispatchWindowsForAgent(ctx context.Context, agentID string) ([]DispatchWindow, error)
}
```

`Deps` gains `AgentWindowSource` as a fifth constituent; `cmd/flowd` needs
no edit — its existing `var _ harvest.Deps = (*store.Store)(nil)` compile
check now covers the new method automatically.

  - [ ] **Step 4: Run the test to verify it passes, then targeted lint**

Run: `cd stats && go test ./internal/store/ -run TestDispatchWindowsForAgent -count=1 && go test ./internal/harvest/ -count=1 && go test ./cmd/flowd/ -count=1`
Expected: PASS (the flowd run proves the store still satisfies `Deps` whole).
Run: `cd stats && gofmt -l . && go vet ./...`
Expected: no output, exit 0.

  - [ ] **Step 5: Commit**

```bash verified:authored in-tree for this change
git add stats/internal/store/records.go stats/internal/store/records_test.go stats/internal/harvest/deps.go
git commit -m "feat(stats): resolve dispatch windows by agent id"
```

- [x] 2. Agent-transcript batches credit their dispatch rows directly

**Files:**
- `stats/internal/harvest/transcript.go`
- `stats/internal/harvest/attribute.go`
- `stats/internal/harvest/watcher.go`
- `stats/internal/harvest/attribute_test.go`

**Build:** green
**Tests:** `TestAgentFileRecordsCreditOneRowWhole`, `TestAgentFileRecordsSplitAcrossResumedRows`, `TestAgentFileRecordsBeforeTheFirstRowFloorToIt`, `TestAgentFileRecordsWithNoRowsCreditNothing`, `TestAttributeDispatchesRoutesAgentFileBatchesDirectly`
**Regression:** reverting this commit returns resumed-agent batches to
`DispatchAttributor`'s identity-then-interval pass, which stamps every row
sharing the agentId unattributed — the measured kan-377 failure
(6 of 28 rows) returns.
**Baseline:** before=104 after=109
<!-- measured: go test ./internal/harvest/ -count=1 -v @ branch spectre/kan-357-record-each-dispatch-own-token-usage (worktree, from main e2d6b02) -->
<!-- predicted: go test ./internal/harvest/ -count=1 -v after this task -->
**Commit:** feat(harvest): credit agent-transcript batches by agent id

  - [ ] **Step 1: Write the failing tests**

In `stats/internal/harvest/attribute_test.go`. The pure function under
test takes the windows for one agentId plus that file batch's records and
returns per-row deltas — no database, no watcher:

```go verified:authored in-tree for this change
func TestAgentFileRecordsCreditOneRowWhole(t *testing.T) {
	row := DispatchWindow{DispatchID: 7, AgentID: "a1", StartedAt: time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)}
	records := []Record{
		{SessionID: "sess-parent", AgentID: "a1", IsSidechain: true, Timestamp: row.StartedAt.Add(time.Minute), Usage: Usage{InputTokens: 10, OutputTokens: 5}},
		{SessionID: "sess-parent", AgentID: "a1", IsSidechain: true, Timestamp: row.StartedAt.Add(2 * time.Minute), Usage: Usage{InputTokens: 20}},
	}
	got := attributeAgentFileRecords([]DispatchWindow{row}, records)
	if len(got) != 1 {
		t.Fatalf("deltas cover %d rows, want 1", len(got))
	}
	if got[7].Sidechain.InputTokens != 30 || got[7].Sidechain.OutputTokens != 5 {
		t.Errorf("row 7 = %+v, want the batch's whole sidechain usage", got[7])
	}
}

func TestAgentFileRecordsSplitAcrossResumedRows(t *testing.T) {
	first := DispatchWindow{DispatchID: 10, AgentID: "a1", StartedAt: time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)}
	second := DispatchWindow{DispatchID: 12, AgentID: "a1", StartedAt: first.StartedAt.Add(10 * time.Minute)}
	records := []Record{
		{AgentID: "a1", IsSidechain: true, Timestamp: first.StartedAt.Add(time.Minute), Usage: Usage{InputTokens: 10}},
		{AgentID: "a1", IsSidechain: true, Timestamp: second.StartedAt.Add(time.Minute), Usage: Usage{InputTokens: 40}},
	}
	got := attributeAgentFileRecords([]DispatchWindow{first, second}, records)
	if got[10].Sidechain.InputTokens != 10 || got[12].Sidechain.InputTokens != 40 {
		t.Errorf("split = %+v, want each resume's own records only", got)
	}
}

func TestAgentFileRecordsBeforeTheFirstRowFloorToIt(t *testing.T) {
	row := DispatchWindow{DispatchID: 10, AgentID: "a1", StartedAt: time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)}
	records := []Record{
		{AgentID: "a1", IsSidechain: true, Timestamp: row.StartedAt.Add(-time.Minute), Usage: Usage{InputTokens: 3}},
	}
	got := attributeAgentFileRecords([]DispatchWindow{row}, records)
	if got[10].Sidechain.InputTokens != 3 {
		t.Errorf("row 10 = %+v, want the pre-start record floored to the first row", got[10])
	}
}

func TestAgentFileRecordsWithNoRowsCreditNothing(t *testing.T) {
	records := []Record{{AgentID: "a1", IsSidechain: true, Usage: Usage{InputTokens: 1}}}
	if got := attributeAgentFileRecords(nil, records); len(got) != 0 {
		t.Errorf("deltas = %+v, want none — an agent with no rows has nothing to credit", got)
	}
}
```

And the routing test in the same file, against a fake
`AgentWindowSource`-satisfying watcher: a batch whose path sits in a
`subagents/` directory credits rows by the file's agentId and never calls
`DispatchAttributor`'s ambiguous path, while a top-level path keeps the
inference pass:

```go verified:authored in-tree for this change
func TestAttributeDispatchesRoutesAgentFileBatchesDirectly(t *testing.T) {
	w := newWatcherForDispatchTests(t, fakeAgentWindows{
		"a1": {{DispatchID: 7, AgentID: "a1", StartedAt: time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)}},
	})
	w.attributeDispatches(context.Background(), []Record{
		{AgentID: "a1", IsSidechain: true, Usage: Usage{InputTokens: 9}},
	}, filepath.Join(t.TempDir(), "sess-x", "subagents", "agent-a1.jsonl"))
	// mergeSink records MergeDispatchMetrics calls; assert patch landed on row 7.
}
```

(The sink and watcher fakes follow `watcher_test.go`'s existing
embed-`NoDeps`-and-override pattern; extend the existing fake sink with a
`DispatchWindowsForAgent` method only where the test's own `Deps` needs
it.)

  - [ ] **Step 2: Run them to verify they fail**

Run: `cd stats && go test ./internal/harvest/ -run 'TestAgentFileRecords|TestAttributeDispatchesRoutes' -count=1`
Expected: FAIL — `attributeAgentFileRecords undefined`.

  - [ ] **Step 3: Implement**

`stats/internal/harvest/transcript.go` — export the path test
`ReadDispatchMeta` already applies privately:

```go verified:authored in-tree for this change
// IsAgentTranscriptPath reports whether path is a per-agent subagent
// transcript (.../subagents/agent-<id>.jsonl) — the same directory rule
// ReadDispatchMeta applies, exported for the watcher's batch routing.
func IsAgentTranscriptPath(path string) bool {
	return filepath.Base(filepath.Dir(path)) == subagentsDirName
}
```

`stats/internal/harvest/attribute.go` — the pure split, beside
`bestDispatchWindow`:

```go verified:authored in-tree for this change
// attributeAgentFileRecords sums one agent-file batch's sidechain records
// onto the dispatch windows its agentId resolves to. Windows must be
// ordered by (started_at, id) — AgentWindowSource's own contract. A
// record joins the latest window whose StartedAt is at or before the
// record's timestamp; before the first window's start it floors to the
// first. This chooses only among resumes of one agent — it can never move
// spend across agents (design.md, resumed-split-by-started-at). Non-
// sidechain records contribute nothing, matching DispatchAttributor's own
// filter. With no windows there is nothing to credit.
func attributeAgentFileRecords(windows []DispatchWindow, records []Record) map[int64]TokenDelta
```

`stats/internal/harvest/watcher.go` — in `attributeDispatches` (watcher.go:839),
branch on the batch's path before anything else:

```go verified:authored in-tree for this change
	if IsAgentTranscriptPath(path) {
		w.attributeAgentFile(ctx, records, path)
		return
	}
```

`attributeAgentFile` takes the agentId from the path's own filename
(`agent-<id>.jsonl`), asks `w.deps.DispatchWindowsForAgent`, and merges
each returned row's delta through `w.deps.MergeDispatchMetrics` exactly as
the inference path already does — same warn-and-continue failure posture.
Zero windows returns silently (design.md, unrecorded-agents-skipped).

  - [ ] **Step 4: Run the tests to verify they pass, then targeted lint**

Run: `cd stats && go test ./internal/harvest/ -count=1`
Expected: PASS — all 109.
Run: `cd stats && gofmt -l . && go vet ./...`
Expected: no output, exit 0.

  - [ ] **Step 5: Commit**

```bash verified:authored in-tree for this change
git add stats/internal/harvest/transcript.go stats/internal/harvest/attribute.go stats/internal/harvest/watcher.go stats/internal/harvest/attribute_test.go
git commit -m "feat(harvest): credit agent-transcript batches by agent id"
```
