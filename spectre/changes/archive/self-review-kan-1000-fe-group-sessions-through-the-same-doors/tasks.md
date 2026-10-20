# self-review-kan-1000-fe-group-sessions-through-the-same-doors

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** fix the five kan-1000 self-review findings and repair the store rows and records they
already wrote wrong.

**Architecture:** Tasks 1 and 2 change flowd's harvester (`stats/internal/harvest`) and add
store methods it calls (`stats/internal/store/records.go`); Task 3 extends a Go guard; Task 4 is
verifier prose; Task 5 is a CLI refusal. Tasks 1–3 end on a live-store repair step.

**Tech Stack:** Go (`stats/`), Postgres via `pgx`, Markdown skills.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-21-self-review-kan-1000-fe-group-sessions-through-the-same-doors-design.md`
in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer.
  Never write `spectre/changes/` — this change's `verbatim-moves.txt` already lists every prose
  sentence Tasks 1, 3 and 4 add or reword, exactly as the steps below spell them. Never merge or
  push.
- Every task runs in this worktree,
  `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1000-fe-group-sessions-through-the-same-doors`
  (`<wt>` below), on branch `self-review-kan-1000-fe-group-sessions-through-the-same-doors`;
  never touch a main checkout.
- Run `cd stats && gofmt -w .` before committing any Go change; `gofmt -l .` and `make -s web-dist
  && go vet ./...` stay clean.
- The repair steps write the live dev flow store (`flow-postgres`, `localhost:5433`, project
  `gymie-7c1f238a`) — local, not production, and allowed for this change. They run after their
  task's commit, once each; a command that fails is reported with its output, never retried
  blind. A repair step whose tool call is denied is reported with the exact commands left
  unrun, for the parent to hand to the operator.
- `<transcript>` is
  `/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/45e57ddc-ab4c-4f0e-bffc-e105ed098d03.jsonl`
  and `<agents>` is
  `/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/45e57ddc-ab4c-4f0e-bffc-e105ed098d03/subagents`.

## Review focus

- A launch whose description's first word names no row (`Task 3/22 (port guard)`) must pair by
  time exactly as before; a keyed stamp must never overwrite a row that already carries an agent
  (Task 1).
- A begin whose key a keyed stamp already used must never claim another launch, in the same batch
  or a later one (Task 1).
- A back-filled inline row's figures must be replaced, never added to, and must count only records
  up to the committed offset — the next batch credits the rest (Task 2).
- `flowd restamp-dispatch` given a parent transcript must refuse a row that is not `inline`
  (Task 2).
- A report carrying no count lines, or the parent's `no verbatim report captured` placeholder,
  must never fail the guard (Task 3).

- [x] 1. Stamp a launch onto the row its description names before pairing by time, and restamp kan-1000's 22 rows
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Write
    `TestWatcherStampsALaunchByItsDescriptionKey` (Step 2) and run it against the untouched
    harvester: it fails, the three launches stamped by time onto the rotated rows (kan-1000 round
    9's shape). Keep the test.
  - [x] **Step 2: Write the failing tests.**
    - `stats/internal/harvest/watcher_test.go`, `TestWatcherStampsALaunchByItsDescriptionKey`:
      one transcript, session `"s"`, holding one Bash line at `00:00:00Z` whose command begins
      three rows in this order — keys `panel-9-primary+principles`,
      `panel-9-test-sanity+call-efficiency`, `visual-verify-fixrun1-final-gymie`, token
      `mf-k1000f1` — then three launch results at `00:00:20Z`, `00:00:23Z`, `00:00:26Z` with
      `"description"` `visual-verify-fixrun1-final-gymie`, `panel-9-primary+principles`,
      `panel-9-test-sanity+call-efficiency` and agents `av`, `ap`, `at`. A fake deps embedding
      `harvest.NoDeps` records `StampSessionDispatchAgent` calls and answers `true` for those
      three keys, `false` for any other. Assert the
      keyed calls are exactly `{s, visual-verify-fixrun1-final-gymie, av}`,
      `{s, panel-9-primary+principles, ap}`, `{s, panel-9-test-sanity+call-efficiency, at}`, and
      that `StampDispatchAgent` (time pairing) was never called. Then append only a launch at
      `00:00:41Z` with description `Task 1/1 (x)`, agent `ai` and no begin of its own, run
      `RunOnce` again, and assert `StampDispatchAgent` is still never called — none of the three
      begins the keyed stamps used is left to claim it.
      Build the lines the way `stampBeginLine` / `stampLaunchLine` do.
    - `stats/internal/store/records_test.go`,
      `TestStampSessionDispatchAgentFillsTheSessionsLatestEmptyRow`: two rows under one key, one
      `session_id` `s1` and one `s2`, plus an older `s1` row under the same key already carrying
      an agent. `StampSessionDispatchAgent(ctx, "s1", key, "a1")` returns `true` and fills only
      the empty `s1` row; a second call returns `false`; a call for session `s3` returns `false`.
      Seed rows the way `TestStampLoopDispatchAgentFillsLatestRowByKey` does, with a session id.
  - [x] **Step 3: Run them to see them fail.** `cd stats && go test ./internal/harvest/
    ./internal/store/ -run
    'TestWatcherStampsALaunchByItsDescriptionKey|TestStampSessionDispatchAgentFillsTheSessionsLatestEmptyRow'
    -count=1` — the store test fails to compile (no method), the watcher test fails on time
    stamps.
  - [x] **Step 4: Implement.**
    - `stats/internal/harvest/transcript.go`: `rawToolUseResult` gains
      ``Description string `json:"description"` ``; `AgentLaunch` gains `Key string` — the first
      `strings.Fields` word of the description, `""` when it has none — set in
      `ParseAgentLaunches`. Its doc comment names `**Every dispatch's Agent-tool description**`
      (`skills/flow/implement.md`) as the contract that puts the key there.
    - `stats/internal/harvest/deps.go`: `DispatchAgentStamper` gains
      `StampSessionDispatchAgent(ctx context.Context, sessionID, key, agentID string) (stamped
      bool, err error)`; `NoDeps` answers `false, nil`.
    - `stats/internal/store/records.go`: `StampSessionDispatchAgent` — one `UPDATE dispatches
      SET agent_id = $3 WHERE id = (SELECT d.id FROM dispatches d WHERE d.session_id = $1 AND
      d.dispatch_key = $2 AND COALESCE(d.agent_id, '') = '' ORDER BY d.started_at DESC, d.id DESC
      LIMIT 1)`, returning rows-affected `> 0`, errors wrapped as `StampLoopDispatchAgent` wraps
      them.
    - `stats/internal/harvest/watcher.go`: `stampKeyedLaunches(ctx, path, launches, begins)`
      runs in `stampDispatchAgents` after the pending lists merge and before the candidates are
      built. For each launch with `Key != ""` and `SessionID != ""` it calls
      `StampSessionDispatchAgent`; on `true` it credits the agent file
      (`creditStampedAgentFile`), removes the launch, and adds the key to a per-path set
      `w.keyStampedKeys[path]` (cleared wholesale past `maxOpenAgentCalls`, as
      `rememberOpenAgentCalls` bounds its set). Every begin whose `Key` is in that set is removed
      before pairing, so a begin arriving in a later batch is dropped too. An error is warned
      with the posture `stampLoopLaunches` uses and the launch stays for time pairing. Rewrite
      the `PAIRING IS BY TIME` paragraph of `stampDispatchAgents`' doc comment to say a launch is
      first stamped by its description's key and only a launch naming no row pairs by time.
    - `skills/flow/review-panel.md`, the `-agent-id` paragraph: replace its line
      `pairing each launch with the begin nearest it in time;` with this one line:

```markdown verified:applied at plan time on 19fd4a8c, check-verbatim-moves.sh FAIL line captured into verbatim-moves.txt, then reverted
stamping each launch onto the row its Agent-tool `description`'s first word names in the launching session, and pairing a launch whose description names no row with the begin nearest it in time;
```

    - `skills/flow/SKILL-rationale.md`: in the paragraph opening `` `begin` is recorded
      immediately before the dispatch ``, replace
      `the harvester pairs that result with the begin nearest it in time,` with
      ``the harvester stamps that result onto the row its `description`'s first word names, else pairs it with the begin nearest it in time,``.
  - [x] **Step 5: Verify.** `cd stats && gofmt -w . && gofmt -l . && make -s web-dist && go vet
    ./...` is clean; `go test ./internal/harvest/ ./internal/store/ -run
    'TestWatcherStampsALaunchByItsDescriptionKey|TestStampSessionDispatchAgentFillsTheSessionsLatestEmptyRow|TestWatcherStamps|TestWatcherCarries|TestLateStamp|TestStampLoopDispatchAgent'
    -count=1` passes; from `<wt>`, `scripts/check-verbatim-moves.sh` exits 0 and
    `scripts/check-markdown-integrity.py` exits 0. Commit.
  - [x] **Step 6: Repair the store.** Build the binary once and restamp each row onto its true
    agent, each pair `<dispatch id>:<agent id>`:

```bash verified:every agent file below exists under <agents>, and each pair's agent meta description opens with that row's dispatch_key (store rows read 2026-10-21)
cd <wt>/stats && go build -o /private/tmp/flowd-restamp ./cmd/flowd
A=/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/45e57ddc-ab4c-4f0e-bffc-e105ed098d03/subagents
for pair in 6689:a408d251ac30407de 6690:a3716f5c82cb55bb4 6694:a5fb016ee5bceb3ee \
  6697:a33744221e2ff446d 6698:af356a3951430255d 6705:a370947ac4d7956d6 \
  6706:ac67e6ac4295ca89e 6708:a31b330e54acd0b98 6710:a7b53a4363e1874c1 \
  6711:a4f0f560bc5ba7a63 6712:a6ff0d4128631e27b 6716:ab076db08895cb3cc \
  6717:acca67fa039c01eda 6718:ae3f1716cd2bb62df 6727:ae0a8219b2046172a \
  6728:ad0f9d76eb3600ab8 6732:a4dea4a68353bbbd1 6733:a4556ca7c690b8ee4 \
  6734:a769b929d38dce23e 6739:ad8ec196be47c4b95 6740:ac2e06e388f489522 \
  6741:ae7f31e03b6f094bf; do
  /private/tmp/flowd-restamp restamp-dispatch "${pair%%:*}" "$A/agent-${pair#*:}.jsonl" || echo "FAILED ${pair%%:*}"
done
```

    Then `flow record dispatches -C /Users/tweety53/Projects/gymie -change
    kan-1000-fe-group-sessions-through-the-same-doors` shows each of those 22 ids carrying its
    pair's agent and a `tokens` bag. Report any `FAILED` line with its output.

**Build:** green
**Files:** `stats/internal/harvest/transcript.go`, `stats/internal/harvest/watcher.go`, `stats/internal/harvest/deps.go`, `stats/internal/harvest/watcher_test.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `skills/flow/review-panel.md`, `skills/flow/SKILL-rationale.md`
**Tests:** `TestWatcherStampsALaunchByItsDescriptionKey`, `TestStampSessionDispatchAgentFillsTheSessionsLatestEmptyRow`
**Regression:** reverting `stampKeyedLaunches` stamps round 9's launches onto the rotated rows by time, failing the watcher test; reverting the store method fails its test to compile.
**Baseline:** before=148 after=150
<!-- measured: cat stats/internal/harvest/watcher_test.go stats/internal/store/records_test.go | grep -c '^func Test' -->
**Commit:** fix(harvest): stamp a launch onto the row its description names before pairing by time
**After:** none

**Decision:** launch-stamp-by-description-key

- [x] 2. Measure a back-filled inline row from the parent transcript, and repair kan-1000's four
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Write
    `TestWatcherMeasuresABackfilledInlineRow` (Step 2) and run it against the untouched watcher:
    no `RestampInlineDispatch` call is made — the row stays unmeasured. Keep the test.
  - [x] **Step 2: Write the failing tests.**
    - `stats/internal/harvest/watcher_test.go`, `TestWatcherMeasuresABackfilledInlineRow`: a
      transcript, session `"s"`, whose first `RunOnce` reads one main-thread assistant record at
      `00:01:00Z` (output 7) while the fake window source returns no windows; then a Bash line at
      `00:05:00Z` running `flow record dispatch backfill -change c -task 1 -role implementer
      -model opus -agent-id inline -key task-1-implementer-fix-1 -session-token mf-x -started-at
      "$(date -u +%FT%TZ)" -ended-at "$(date -u +%FT%TZ)" -outcome completed` is appended, the
      fake now returning an inline window id `7` over `[00:00:30Z, 00:02:00Z)` and
      `InlineDispatchID("mf-x", "task-1-implementer-fix-1")` → `7, true`. After the second
      `RunOnce`, exactly one `RestampInlineDispatch` call is recorded, for id `7`, whose patch's
      `tokens.main.output` is `7`.
    - same file, `TestInlineDispatchPatchCreditsOnlyTheRowsWindow`: `harvest.InlineDispatchPatch`
      over three main-thread records at `00:00:10Z`, `00:01:00Z`, `00:03:00Z` (outputs 1, 2, 4),
      inline window `7` over `[00:00:30Z, 00:02:00Z)`, returns a patch whose `tokens.main.output`
      is `2`; for an id no window carries it returns `harvest.ErrNoInlineShare`.
    - `stats/internal/store/records_test.go`,
      `TestInlineDispatchIDAndRestampInlineDispatch`: `InlineDispatchID(token, key)` returns the
      latest `inline` row under that token and key, `false` for a key with only an agent row;
      `RestampInlineDispatch` replaces an inline row's `tokens` and `signals`, keeps its other
      metrics keys, and returns `store.ErrDispatchNotFound` for a non-inline row, leaving it
      untouched.
  - [x] **Step 3: Run them to see them fail.** `cd stats && go test ./internal/harvest/
    ./internal/store/ -run
    'TestWatcherMeasuresABackfilledInlineRow|TestInlineDispatchPatchCreditsOnlyTheRowsWindow|TestInlineDispatchIDAndRestampInlineDispatch'
    -count=1` fails.
  - [x] **Step 4: Implement.**
    - `stats/internal/harvest/deps.go`: interface `InlineDispatchRestamper` —
      `InlineDispatchID(ctx context.Context, sessionToken, key string) (id int64, found bool, err
      error)` and `RestampInlineDispatch(ctx context.Context, dispatchID int64, patch
      json.RawMessage) error` — composed into `Deps`; `NoDeps` answers `0, false, nil` and `nil`.
    - `stats/internal/store/records.go`: `InlineDispatchID` — `SELECT id FROM dispatches WHERE
      session_token = $1 AND dispatch_key = $2 AND agent_id = 'inline' ORDER BY started_at DESC,
      id DESC LIMIT 1`, no row → `0, false, nil`; `RestampInlineDispatch` — `RestampDispatch`'s
      `UPDATE` with `WHERE id = $1 AND agent_id = 'inline'` and no `agent_id` change, zero rows →
      `ErrDispatchNotFound`, a nil patch → `ErrNilMetricsPatch`.
    - `stats/internal/harvest/watcher.go`:
      - `var ErrNoInlineShare = errors.New(...)` and `func InlineDispatchPatch(ctx
        context.Context, windows DispatchWindowSource, complete []byte, dispatchID int64)
        (json.RawMessage, error)` beside `AgentFileDispatchPatch`: records are
        `ParseAssistantRecords(complete)` plus `ParseSignalRecords(complete)`,
        `NewDispatchAttributor(windows).Attribute`, and the patch is built from
        `deltas[dispatchID]` exactly as `attributeDispatches` builds one (tokens, sidechain
        signals, `Apportioned` when non-zero); no entry → `ErrNoInlineShare`.
      - The begin parser's split loop becomes one helper both judges share:
        `dispatchBeginsFromCommand` keeps its behaviour, and `inlineBackfillsFromCommand(command
        string) []dispatchBeginEvent` judges each invocation by `inlineBackfillFromInvocation`:
        it must contain `flow record dispatch backfill`, `-agent-id` must read `inline`, and
        `-key` / `-session-token` must resolve to literals holding no `$` or backquote
        (`resolvedFlagValue`, as `dispatchBeginFromInvocation` reads them).
      - `measureInlineBackfills(ctx, path string, commands []CommandRecord, committed int64)`,
        called in `RunOnce` right after `stampDispatchAgents`: for each backfill event,
        `InlineDispatchID`; found → read `path`, cut to `committed` bytes, `InlineDispatchPatch`
        with `w.deps`, then `RestampInlineDispatch`. Every failure, `ErrNoInlineShare` included,
        is warned and stepped over — the row stays as it was.
    - `stats/cmd/flowd/main.go`, `restampDispatch`: a path that is not an agent transcript is a
      parent transcript — read it whole, `harvest.InlineDispatchPatch(ctx, st, raw, id)`, then
      `st.RestampInlineDispatch(ctx, id, patch)`. The usage line becomes
      `usage: flowd restamp-dispatch <dispatch-id> <agent-transcript.jsonl | parent-transcript.jsonl>`
      and the doc comment names both modes.
  - [x] **Step 5: Verify.** `cd stats && gofmt -w . && gofmt -l . && make -s web-dist && go vet
    ./...` is clean; `go test ./internal/harvest/ ./internal/store/ ./cmd/flowd/ -run
    'TestWatcherMeasuresABackfilledInlineRow|TestInlineDispatchPatchCreditsOnlyTheRowsWindow|TestInlineDispatchIDAndRestampInlineDispatch|TestWatcherStamps|TestWatcherCarries|Wiring'
    -count=1` passes. Commit.
  - [x] **Step 6: Repair the store.** Rebuild the binary from this commit and measure rows 24, 42,
    60 and 73 from the parent transcript:

```bash unverified:confirm each command prints no error; the four ids are the inline rows with no tokens in the 2026-10-21 store read
cd <wt>/stats && go build -o /private/tmp/flowd-restamp ./cmd/flowd
for id in 6688 6707 6730 6743; do
  /private/tmp/flowd-restamp restamp-dispatch "$id" /Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/45e57ddc-ab4c-4f0e-bffc-e105ed098d03.jsonl || echo "FAILED $id"
done
```

    Then `flow record cost-status -C /Users/tweety53/Projects/gymie -change
    kan-1000-fe-group-sessions-through-the-same-doors` prints `2 unattributed — not measured: 2`
    (rows 52 and 53). Report any other figure, or any `FAILED` line, with its output.
    <!-- predicted: flow record cost-status after Tasks 1 and 2's repair steps -->

**Build:** green
**Files:** `stats/internal/harvest/watcher.go`, `stats/internal/harvest/deps.go`, `stats/internal/harvest/watcher_test.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `stats/cmd/flowd/main.go`
**Tests:** `TestWatcherMeasuresABackfilledInlineRow`, `TestInlineDispatchPatchCreditsOnlyTheRowsWindow`, `TestInlineDispatchIDAndRestampInlineDispatch`
**Regression:** reverting `measureInlineBackfills` leaves the back-filled row unmeasured, failing the watcher test; reverting `InlineDispatchPatch` or the store methods fails their tests to compile.
**Baseline:** before=150 after=153
<!-- measured: cat stats/internal/harvest/watcher_test.go stats/internal/store/records_test.go | grep -c '^func Test' -->
**Commit:** fix(harvest): measure a back-filled inline row from the parent transcript
**After:** Task 1

**Decision:** inline-backfill-recompute

- [x] 3. Fail the panel close when a slot's report counts more findings than the store holds, and record kan-1000's F10
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `scripts/check-panel-findings-closed.sh
    /Users/tweety53/Projects/gymie-worktrees/kan-1000-fe-group-sessions-through-the-same-doors
    kan-1000-fe-group-sessions-through-the-same-doors` from `<wt>`: it prints `FINDINGS-CLOSED`,
    exit 0, while `panel-report-11-primary.md` there raises an Important the store lacks.
  - [x] **Step 2: Write the failing test.**
    `stats/internal/guard/check_panel_findings_closed_test.go`,
    `TestCheckPanelFindingsClosedReportCounts`, table-driven over a worktree whose
    `.superpowers/sdd/` holds report files, findings from `cfcJSON`, dispatches from
    `cfcBinDispatches`:
    - `panel-report-11-primary.md` opening `Critical: 0` / `Important: 1`, no findings → exit 1,
      stderr naming `panel-report-11-primary.md` with `report 1, recorded 0`.
    - same report, one round-11 Important `fixed` under slot `primary+principles` with a clean
      round-12 `primary+principles` re-run → exit 0.
    - same report, the Important recorded at round 10 → exit 1.
    - a report with no count lines and a report holding only `no verbatim report captured —
      timeout` → exit 0.
    - `panel-report-3-test-sanity.md` with `Critical: 1` / `Important: 0` and one round-3
      Critical `withdrawn duplicate of F2` under `test-sanity` → exit 0.
  - [x] **Step 3: Run it to see it fail.** `cd stats && go test ./internal/guard/ -run
    'TestCheckPanelFindingsClosedReportCounts' -count=1` fails on the first case.
  - [x] **Step 4: Implement.**
    - `stats/internal/guard/panelfindingsclosed.go`: after the fixed-without-rerun class, glob
      `<worktree>/.superpowers/sdd/panel-report-*.md`; a name matching
      `^panel-report-([0-9]+)-(.+)\.md$` gives round and id. Read the file's first
      `(?m)^Critical: ([0-9]+)\s*$` and `(?m)^Important: ([0-9]+)\s*$`; a file missing either is
      skipped. Recorded = findings with that `round`, severity `critical` or `important`
      (case-insensitive), any status, whose `slot` split on `+` contains the id. When the report's
      sum exceeds it, add `<file> (report <n>, recorded <m>)` to one stderr line
      `finding(s) a slot's report counts that the store does not hold: …` and set `violated`. An
      unreadable report is exit 2, the guard's cannot-answer posture.
    - `scripts/check-panel-findings-closed.sh`: the exit `1` line of the header gains the class —
      `or a slot's panel-report-<round>-<id>.md counts on its Critical: <n> and Important: <n>
      lines more findings than the store holds for that round and slot` — and the exit `2` line
      gains `an unreadable panel report`.
    - `skills/flow/review-panel.md`, the REPORT FILE paragraph (the one `render-slot-prompt.sh`
      reads): replace its last line `> too; the file is the record.` with:

```markdown verified:applied at plan time on 19fd4a8c, check-verbatim-moves.sh FAIL line captured into verbatim-moves.txt, then reverted
> too; the file is the record. Open the file with one `Critical: <n>` and one `Important: <n>` line
> counting the findings it raises at each severity.
```

    - same file, the paragraph under `check-panel-findings-closed.sh <worktree> <change>`:
      replace its line `none. Either way, return to the` with:

```markdown verified:applied at plan time on 19fd4a8c, check-verbatim-moves.sh FAIL line captured into verbatim-moves.txt, then reverted
none. It fires as well when a slot's report counts more Critical and Important findings on its
`Critical: <n>` and `Important: <n>` lines than the store holds for that round and slot. Either way, return to the
```
  - [x] **Step 5: Verify.** `cd stats && gofmt -w . && gofmt -l . && make -s web-dist && go vet
    ./...` is clean; `go test ./internal/guard/ -run
    'TestCheckPanelFindingsClosed|RenderSlot|DispatchParagraph' -count=1` passes; from `<wt>`,
    `scripts/check-verbatim-moves.sh`, `scripts/check-dispatch-paragraphs.sh` and
    `scripts/check-markdown-integrity.py` exit 0. Commit.
  - [x] **Step 6: Repair the record.** Record round 11's Important as F10 and close it:

```bash unverified:confirm F10 is still the next free ref (flow record findings lists F1-F9 on 2026-10-21)
N=kan-1000-fe-group-sessions-through-the-same-doors
flow record finding -C /Users/tweety53/Projects/gymie -change $N -ref F10 -round 11 -slot primary \
  -severity Important -status open -task 19 -dispatch-seq 74 \
  -location 'gymie-playwright:scripts/specs-for-diff.mjs:56' \
  -note 'A caller branch that dies untagged (no own tag reaching a spec, no further caller) is silently dropped while another branch'"'"'s hits are returned, so the narrowed list omits the screen reached only through the untagged branch.' \
  -reproducer 'none — fixed inline as gymie-playwright b2212bb before it was recorded; panel-report-11-primary.md states the defect'
flow record status -C /Users/tweety53/Projects/gymie -change $N -ref F10 -status fixed
```

    Then `flow record findings -C /Users/tweety53/Projects/gymie -change $N` lists F10 `fixed`.

**Build:** green
**Files:** `stats/internal/guard/panelfindingsclosed.go`, `stats/internal/guard/check_panel_findings_closed_test.go`, `scripts/check-panel-findings-closed.sh`, `skills/flow/review-panel.md`
**Tests:** `TestCheckPanelFindingsClosedReportCounts`
**Regression:** reverting the report read lets a report counting an Important the store lacks pass, failing the test's first case.
**Baseline:** before=1 after=2
<!-- measured: cat stats/internal/guard/check_panel_findings_closed_test.go | grep -c '^func Test' -->
**Commit:** fix(guard): fail the panel close when a slot's report counts findings the store lacks
**After:** Task 1

**Decision:** panel-report-count-lines
**Decision:** f10-recorded-fixed

- [x] 4. Pin each PNG's mtime before zipping the full-app suite
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In a scratch directory
    outside the worktree, zip one file with `zip -X -j` twice, `touch` between: `cmp` reports the
    two zips differ.
  - [x] **Step 2: Edit.** In `skills/flow/visual-verify-verifier.md` step 8, replace the block's
    resolver-into-zip line with:

```bash verified:spiked at plan time — two zips over the same files with mtimes pinned compare identical, unpinned they differ
   list=$(resolve-visual-screenshots.sh <worktree> full-app-suite.spec.ts)
   printf '%s\n' "$list" | tr '\n' '\0' | xargs -0 touch -t 198001010000
   printf '%s\n' "$list" | zip -X -j -@ <screenshots root>/full-app-suite.zip
```

    keeping `rm -f <screenshots root>/full-app-suite.zip` above it, and replace the line after
    the block that opens `` `<screenshots root>` is `screenshots` resolved `` so the paragraph
    reads:

````markdown verified:applied at plan time on 19fd4a8c, check-verbatim-moves.sh FAIL line captured into verbatim-moves.txt, then reverted
   `<screenshots root>` is `screenshots` resolved the way `resolve-visual-screenshots.sh`
   resolves it. Every PNG's mtime is pinned before zipping because the zip stores it: a rebuild over
   unchanged images is then byte-identical and commits nothing. A resolver exit 1 or 2, or a non-zero `zip`, blocks.
````
  - [x] **Step 3: Verify.** From `<wt>`, `scripts/check-verbatim-moves.sh` and
    `scripts/check-markdown-integrity.py` exit 0; run the three lines over two scratch PNGs
    twice, `touch`ing them between runs: the zips compare identical. Commit.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`
**Tests:** **none** — prose only; Step 3's rebuild-twice run is the check.
**Regression:** reverting the commit stores fresh mtimes again, and a rebuild over unchanged PNGs commits a changed zip.
**Commit:** docs(flow): pin each screenshot's mtime before zipping the full-app suite
**After:** none

**Decision:** zip-pin-mtime

- [x] 5. Refuse a visual-verify verifier begin with no diff base
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Write the test (Step 2) and
    run it on the untouched CLI: the begin exits 0 and contacts the daemon.
  - [x] **Step 2: Write the failing test.** `stats/cmd/flow/record_test.go`,
    `TestRecordDispatchBeginRefusesAVisualVerifyVerifierWithoutDiffBase`, on the
    `TestRecordDispatchBeginRefusesAPanelReviewerWithoutSlot` pattern: `-role verifier -key
    visual-verify-gymie` with no `-diff-base` → exit 2, stderr names `-diff-base`, the daemon not
    contacted; the same with `-diff-base 0123456789abcdef0123456789abcdef01234567` → exit 0;
    `-role verifier -key verify-gymie` with no `-diff-base` → exit 0.
  - [x] **Step 3: Implement.** In `runRecordDispatchBegin` (`stats/cmd/flow/record.go`), beside
    the panel-slot refusal: `if *role == "verifier" && strings.HasPrefix(*key, "visual-verify")
    && *diffBase == ""` print
    `flow: -key %q is a visual-verify verifier row: pass -diff-base <the HEAD it verifies> (skills/flow/visual-verify.md)`
    and return `2`.
  - [x] **Step 4: Verify.** `cd stats && gofmt -w . && gofmt -l . && make -s web-dist && go vet
    ./...` is clean; `go test ./cmd/flow/ -run 'TestRecordDispatchBegin|TestRecordDispatchBackfill'
    -count=1` passes. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordDispatchBeginRefusesAVisualVerifyVerifierWithoutDiffBase`
**Regression:** reverting the refusal lets a visual-verify verifier begin with no `-diff-base` record, failing the test.
**Baseline:** before=119 after=120
<!-- measured: cat stats/cmd/flow/record_test.go | grep -c '^func Test' -->
**Commit:** fix(stats): refuse a visual-verify verifier begin with no diff base
**After:** none
