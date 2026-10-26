# self-review-kan-1036-late-round-finding-prompt-tuning

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close kan-1036's four non-big self-review findings as one self-review fix loop.

**Architecture:** two prose edits applied as verified diffs; a lexical reproducer-name check in
`check-panel-reproducers` plus a repair of kan-1036's F3 row; a Round column in the rendered
panel findings table; a zero delta from the dispatch attributor for an inline window the harvest
has read past without finding a parent turn.

**Tech Stack:** Go (`stats/internal/guard`, `stats/internal/records`, `stats/internal/harvest`,
`stats/internal/store`), Markdown skills.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-26-self-review-kan-1036-late-round-finding-prompt-tuning-design.md` in
this worktree.

## Global Constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/kan-1036-late-round-finding-prompt-tuning`, branch `spectre/kan-1036-late-round-finding-prompt-tuning`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- A prose diff given in a step is applied with `git apply` from the worktree's root after copying the lines between its fence markers into a scratch file outside the worktree; each was applied and checked at `c580312c`. A diff that does not apply is reported, never hand-merged by guesswork. Each prose sentence lands character for character: `scripts/check-verbatim-moves.sh` matches the reworded ones against `verbatim-moves.txt` exactly.
- Run `cd stats && gofmt -w .` on Go changes before committing; `cd stats && gofmt -l .` and `cd stats && make -s web-dist && go vet ./...` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container; the store tests create and drop their own databases in it, as they already do.

## Review Focus

- Task 2: a finding joined across slots (`primary+principles`) whose script is `0-primary-1.sh` still passes; a reproducer outside `.superpowers/sdd/reproducers/` and the `none — <reason>` exemption are never name-checked.
- Task 2: kan-1036's F3 row keeps every field but its reproducer.
- Task 4: an inline window already measured, still open, or ending after the batch's last main-thread record gets no zero delta; a launched window never does.

- [x] 1. Count round 1 as a late round in the late-round-miss definition
  - [x] **Step 1: Apply the prose diff** (decision `late-round-from-1`):

````diff verified:applied with git apply at c580312c; check-verbatim-moves.sh then reports exactly the sentences verbatim-moves.txt lists
diff --git a/skills/flow-self-review/SKILL.md b/skills/flow-self-review/SKILL.md
index 3bbef742..98a2be49 100644
--- a/skills/flow-self-review/SKILL.md
+++ b/skills/flow-self-review/SKILL.md
@@ -103,13 +103,13 @@ gap in instructions:
   `skills/flow/sdd-dispatch.md`, `skills/flow/engineering-principles.md`) would have prevented the
   bad code.
 - **`reviewer`** — either a **miss**, a defect raised only in a later round or by a later slot that
-  a round-1 reviewer instruction (`skills/flow/*-reviewer-prompt.md`,
+  a first-round reviewer instruction (`skills/flow/*-reviewer-prompt.md`,
   `skills/flow/reviewer-calibration.md`) would have caught, or a **false positive**, a finding later
   disproved or withdrawn that cost a fix round for nothing. A visual-verify finding's reviewer
   is the verifier: its miss or false positive targets `skills/flow/visual-verify-verifier.md` or
   `skills/flow/visual-verify.md`.
 
-A **late-round miss** — a finding a slot first raises in round 2 or later on a hunk that same slot
+A **late-round miss** — a finding a slot first raises in round 1 or later on a hunk that same slot
 already reviewed, unchanged, in an earlier round — is traced to why the earlier rounds missed it,
 whatever its severity: the pass compares that slot's earlier dispatches with the round that raised
 it — the slot's reviewer prompt file at the change's merge base, the diff each was given, the
````

  - [x] **Step 2: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` and
    `scripts/check-markdown-integrity.sh` exit 0, and
    `git grep -n -F -e 'round 2 or later' -e 'round-1 reviewer' -- skills/flow-self-review/SKILL.md`
    prints nothing.
  - [x] **Step 3: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow-self-review/SKILL.md`
**Tests:** **none** — a two-phrase prose correction; `check-verbatim-moves.sh` pins the exact sentences.
**Regression:** none — no test is declared.
**Commit:** docs(flow-self-review): count round 1 as a late round
**After:** none

**Decision:** late-round-from-1

- [x] 2. Name the reproducer shape in the slot prompt and refuse an off-shape path
  - [x] **Step 1: Write the failing test** `TestCheckPanelReproducersRefusesAnOffShapeName` in
    `stats/internal/guard/check_panel_reproducers_test.go`: a table run through `cprRun` and
    `cprExpect` on raw JSON findings carrying `ref`, `round`, `slot`, `status`, `severity` and
    `reproducer`. Cases — exit 1 naming the ref and `<round>-<id>-<n>.sh`: round 0 `mutation` with
    `.superpowers/sdd/reproducers/r1-M1-1.sh`; round 0 `primary` with
    `.superpowers/sdd/reproducers/r0-P1-1.sh`; round 1 `primary` with
    `.superpowers/sdd/reproducers/0-primary-1.sh` (wrong round); round 0 `primary` with
    `.superpowers/sdd/reproducers/0-Primary-1.sh` (case); round 0 `primary` with
    `.superpowers/sdd/reproducers/0-primary-01.sh`. Exit 0: round 0 `primary` with
    `.superpowers/sdd/reproducers/0-primary-1.sh`; round 2 `mutation` with
    `.superpowers/sdd/reproducers/2-mutation-12.sh --strict`; round 0 `primary+principles` with
    `.superpowers/sdd/reproducers/0-principles-2.sh` and with
    `.superpowers/sdd/reproducers/0-primary+principles-1.sh`; round 0 `primary` with
    `scripts/check-panel-reproducers.sh`; a Minor round 0 `primary` with
    `none — prose-only`. Run `cd stats && go test ./internal/guard/ -run TestCheckPanelReproducersRefusesAnOffShapeName`
    and see the exit-1 cases fail.
  - [x] **Step 2: Implement** (decision `reproducer-name-shape-guarded`) in
    `stats/internal/guard/panelreproducers.go`: read each finding's `round` (a JSON number;
    absent reads as 0) and `slot` index-aligned with `pcParseFindings`, the way `cprSeverities`
    reads severity; for a runnable reproducer whose path token (text before the first space)
    starts `.superpowers/sdd/reproducers/`, add the violation
    `<ref>'s reproducer '<path token>' is off the <round>-<id>-<n>.sh shape — expected <round>-<one of: slots>-<n>.sh`
    unless the remainder equals `<round>-<id>-<n>.sh` for the finding's round, an `<id>` that is
    the whole slot or one of its `+`-separated parts, compared exactly, and `<n>` matching
    `[1-9][0-9]*`. Doc comment beside it: the kan-1036 collision (`r1-m1-1.sh` / `r1-M1-1.sh` on
    a case-insensitive filesystem) is why. Leave the exemption branch and `cprCommandViolations`
    as they are. Run Step 1's selector and the whole
    `cd stats && go test ./internal/guard/ -run 'TestCheckPanelReproducer'` green.
  - [x] **Step 3: Apply the prose diff:**

````diff verified:applied with git apply at c580312c; check-verbatim-moves.sh then reports exactly the sentences verbatim-moves.txt lists
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index 77c71cd2..a359b88f 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -544,7 +544,7 @@ gives it a shebang and `chmod +x`, and records that same path, the path relative
 not the `<abs-worktree>/`-prefixed form above; `run-reproducer.sh` refuses an absolute token.
 Mutation writes its own into the canonical worktree, never its
 `<worktree>-<slot>-<round>` copy, which is removed the moment its dispatch closes. The parent
-records the path the slot supplied verbatim — there is no rename step. **The cwd contract**: the
+records the path the slot supplied verbatim — there is no rename step. `check-panel-reproducers.sh` refuses a recorded path under `.superpowers/sdd/reproducers/` whose name is off `<round>-<id>-<n>.sh`, `<id>` one of the finding's slots, as a malformed field. **The cwd contract**: the
 reproducer always executes with its working directory set to the worktree the parent passes
 `run-reproducer.sh` — per finding, the worktree whose copy of the recorded relative path resolves:
 the canonical worktree first, else the one other entry of the state record's `worktrees` map
@@ -638,7 +638,7 @@ the **AskUserQuestion** it states.
 in the shape a slot authors to, since a slot reads its prompt and not this file:
 
 > **REPRODUCER:** every finding carries a reproducer: a script under the canonical worktree's
-> `<project>/.superpowers/sdd/reproducers/`, named by round, finding id and sequence number and ending `.sh`, recorded by its worktree-relative path, or
+> `<project>/.superpowers/sdd/reproducers/`, named `<round>-<id>-<n>.sh` — `<round>` this round's number, `<id>` your own resolved reviewer id, `<n>` your own 1-based finding index — and recorded by its worktree-relative path, or
 > `none — <reason>` on a Minor finding only. It exits non-zero when the defect is present, and 0
 > only when it is not; a `# mutation-reproducer` follows the mutation brief instead. Within its
 > first 10 lines it declares one `# demonstrates: <path>:<line>:<content>` line per cited location
````

  - [x] **Step 4: Repair kan-1036's F3 row** (decision `repair-f3-only`) — a store write, no
    file change: read the row with
    `flow record findings -change kan-1036-late-round-finding-prompt-tuning -C /Users/tweety53/Projects/agents-worktrees/kan-1036-late-round-finding-prompt-tuning`, then
    re-record it with `flow record finding -C /Users/tweety53/Projects/agents-worktrees/kan-1036-late-round-finding-prompt-tuning -change kan-1036-late-round-finding-prompt-tuning -ref F3 -round 0 -slot mutation -severity Minor -dispatch-seq 4 -task 1`
    plus `-location`, `-status` and `-note` copied byte for byte from that read, and
    `-reproducer "none — its mutation script was overwritten by another slot's r1-m1-1.sh on a case-insensitive filesystem (kan-1036)"`.
    Read it back: every field but `reproducer` is unchanged.
  - [x] **Step 5: Verify.** `scripts/check-panel-reproducers.sh /Users/tweety53/Projects/agents-worktrees/kan-1036-late-round-finding-prompt-tuning kan-1036-late-round-finding-prompt-tuning`
    prints `REPRODUCERS-OK (3 finding(s) declared)`; `scripts/check-verbatim-moves.sh`,
    `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`,
    `scripts/check-dispatch-paragraphs.sh`, `cd stats && gofmt -l .` and
    `cd stats && make -s web-dist && go vet ./...` exit clean.
  - [x] **Step 6: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/panelreproducers.go`, `stats/internal/guard/check_panel_reproducers_test.go`, `skills/flow/review-panel.md`
**Tests:** `TestCheckPanelReproducersRefusesAnOffShapeName`
**Regression:** `TestCheckPanelReproducersRefusesAnOffShapeName` — reverting accepts `r1-M1-1.sh` on a round-0 `mutation` finding and every other off-shape case at exit 0.
**Baseline:** before=1 after=2
<!-- measured: grep -c '^func Test' stats/internal/guard/check_panel_reproducers_test.go @ c580312c -->
**Commit:** fix(guard): refuse a panel reproducer named off the round-id-index shape
**After:** none

**Decision:** reproducer-name-shape-guarded, repair-f3-only

- [x] 3. Render each panel finding's round
  - [x] **Step 1: Write the failing test** `TestRenderPanelCarriesEachFindingsRound` in
    `stats/internal/records/render_test.go`: a `Run` with findings `F1` round 0 and `F2` round 2;
    `RenderPanel` carries the header `| ID | Round | Slot | Severity | Location | Note | Lineage |`
    and rows opening `| F1 | 0 |` and `| F2 | 2 |`. Run
    `cd stats && go test ./internal/records/ -run TestRenderPanelCarriesEachFindingsRound` and see
    it fail.
  - [x] **Step 2: Implement** (decision `round-column`) in `RenderPanel`,
    `stats/internal/records/render.go`: the header and separator gain the `Round` column after
    `ID`, each row the cell `strconv.Itoa(f.Round)`. Update the existing Lineage test's header
    assertion in `render_test.go` to the new header. In `skills/flow/SKILL-rationale.md`'s
    "the rendered record's format (KAN-857 …)" example, set the header to
    `| ID | Round | Slot | Severity | Location | Note | Lineage |`, its separator to seven
    columns, and the example row to `| F1 | 0 | Mutation | Minor | `src/Foo.kt:42` | replaced the silent catch | |`.
  - [x] **Step 3: Verify.** `cd stats && go test ./internal/records/ -run 'TestRenderPanel'` green; `cd stats && gofmt -l .`,
    `cd stats && make -s web-dist && go vet ./...`, `scripts/check-references.sh` and
    `scripts/check-markdown-integrity.sh` clean.
  - [x] **Step 4: Commit** the three files with the subject below.

**Build:** green
**Files:** `stats/internal/records/render.go`, `stats/internal/records/render_test.go`, `skills/flow/SKILL-rationale.md`
**Tests:** `TestRenderPanelCarriesEachFindingsRound`
**Regression:** `TestRenderPanelCarriesEachFindingsRound` — reverting renders the six-column header with no round cell.
**Baseline:** before=37 after=38
<!-- measured: grep -c '^func Test' stats/internal/records/render_test.go @ c580312c -->
**Commit:** feat(records): render each panel finding's round
**After:** none

**Decision:** round-column

- [x] 4. Measure an inline window that held no parent turn as zero
  - [x] **Step 1: Write the failing test** `TestInlineWindowWithNoTurnInsideMeasuresZero` in
    `stats/internal/harvest/attribute_test.go`, after `TestInlineDispatchTakesMainThreadUsage`,
    built on kan-1036's real stamps: main-thread records of `mainSessionID` at
    `2026-10-26T13:31:25.743Z` and `2026-10-26T13:37:28.709Z` carrying non-zero usage; windows,
    all inline unless stated — dispatch 6 `13:31:27.823Z`–`13:37:25.519Z`, `Measured` false (wants
    an entry, every token figure zero); dispatch 5 `13:30:22.713Z`–`13:31:00.301Z`, `Measured`
    true (wants no entry); dispatch 7 the same span as 6 but `Measured` false and `EndedAt`
    `13:40:00Z`, after the batch's last main-thread record (wants no entry); dispatch 8 open,
    `EndedAt` nil (wants no entry); dispatch 9 launched, agent `agent-x`, the same span as 6
    (wants no entry). Run `cd stats && go test ./internal/harvest/ -run TestInlineWindowWithNoTurnInsideMeasuresZero`
    and see it fail.
  - [x] **Step 2: Implement** (decision `inline-window-zero`): `DispatchWindow` gains
    `Measured bool` with a doc line (the row's metrics already carry a `tokens` object).
    `DispatchWindowsForSession` (`stats/internal/store/records.go`) selects
    `(d.metrics -> 'tokens') IS NOT NULL` and scans it into `Measured`. In
    `DispatchAttributor.Attribute` (`stats/internal/harvest/attribute.go`), after a session's
    records are attributed, take that session's latest main-thread record timestamp in the batch;
    every inline window with a non-nil `EndedAt` not after it, `Measured` false and no entry in
    `deltas` gets `deltas[w.DispatchID] = DispatchDelta{}`. Extend `Attribute`'s doc comment: an
    inline row recorded around one blocking command holds no parent turn, and the harvest having
    read past its end is what makes zero a measurement rather than a guess. Extend
    `TestDispatchWindowsForSessionResolvesTheBoundToken` in `stats/internal/store/records_test.go`
    to assert `Measured` false on a row with `{}` metrics and true once
    `MergeDispatchMetrics` has written a `tokens` patch to it.
  - [x] **Step 3: Verify.** `cd stats && go test ./internal/harvest/ ./internal/store/ -count=1 -run 'Inline|DispatchWindow|Attribut'`
    green — a pre-existing test whose fixture now receives a zero entry for an empty closed inline
    window is updated to expect it, and named in the commit body; `cd stats && gofmt -l .` and
    `cd stats && make -s web-dist && go vet ./...` clean.
  - [x] **Step 4: Commit** the four files with the subject below.

**Build:** green
**Files:** `stats/internal/harvest/attribute.go`, `stats/internal/harvest/attribute_test.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `stats/internal/harvest/watcher.go`
**Allowed-collateral:** `stats/internal/harvest/*_test.go`
**Tests:** `TestInlineWindowWithNoTurnInsideMeasuresZero`
**Regression:** `TestInlineWindowWithNoTurnInsideMeasuresZero` — reverting leaves dispatch 6 with no delta, so its bag stays `{}` and the ledger reads `not measured`.
**Baseline:** before=47 after=48
<!-- measured: grep -c '^func Test' stats/internal/harvest/attribute_test.go @ c580312c -->
**Commit:** fix(harvest): measure an inline window that held no parent turn as zero
**After:** none

**Decision:** inline-window-zero, inline-zero-no-repair
