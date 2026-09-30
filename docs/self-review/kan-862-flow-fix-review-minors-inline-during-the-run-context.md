# Self-review context bundle for kan-862-flow-fix-review-minors-inline-during-the-run

found: 6 of 7 sources; skipped: 1 of 7 sources
skipped: change summary (absent)

## .superpowers/sdd/ledgers/kan-862-flow-fix-review-minors-inline-during-the-run.md

# SDD ledger — kan-862-flow-fix-review-minors-inline-during-the-run

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — implementer

- Task: 1
- Role: implementer
- Key: task-1-implementer
- Model: opus effort=high
- Commit: 43d7d887
- Outcome: completed
- Started: 2026-09-30T10:59:36Z
- Tokens: not measured

## Dispatch 2 — implementer

- Task: 2
- Role: implementer
- Key: task-2-implementer
- Model: opus effort=high
- Commit: fef6f859
- Outcome: completed
- Started: 2026-09-30T11:00:58Z
- Tokens: not measured

## Dispatch 3 — implementer

- Task: 3
- Role: implementer
- Key: task-3-implementer
- Model: opus effort=high
- Commit: da88acf4
- Outcome: completed
- Started: 2026-09-30T11:02:14Z
- Tokens: not measured

## Dispatch 4 — implementer

- Task: 4
- Role: implementer
- Key: task-4-implementer
- Model: opus effort=high
- Commit: a36da4eb
- Outcome: completed
- Started: 2026-09-30T11:04:34Z
- Tokens: not measured

## Dispatch 5 — implementer

- Task: 5
- Role: implementer
- Key: task-5-implementer
- Model: opus effort=high
- Commit: ce80d098
- Outcome: completed
- Started: 2026-09-30T11:06:22Z
- Tokens: not measured

## Dispatch 6 — implementer

- Task: 6
- Role: implementer
- Key: task-6-implementer
- Model: opus effort=high
- Commit: 0090d18a
- Outcome: completed
- Started: 2026-09-30T11:06:43Z
- Tokens: not measured

## Dispatch 7 — implementer

- Task: 7
- Role: implementer
- Key: task-7-implementer
- Model: opus effort=high
- Commit: d0b70969
- Outcome: completed
- Started: 2026-09-30T11:07:44Z
- Tokens: not measured

## Dispatch 8 — implementer

- Task: 8
- Role: implementer
- Key: task-8-implementer
- Model: opus effort=high
- Commit: 247a5f24
- Outcome: completed
- Started: 2026-09-30T11:08:05Z
- Tokens: not measured

## Dispatch 9 — reviewer

- Task: no task
- Role: reviewer
- Key: task-1+2+3+4+7-reviewer
- Model: opus effort=default
- Commit: no commit
- Outcome: clean
- Started: 2026-09-30T11:13:04Z
- Tokens: input 80, output 710, cache read 3237039, cache creation 117719

## Dispatch 10 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: opus effort=medium
- Commit: no commit
- Outcome: completed
- Started: 2026-09-30T11:21:00Z
- Tokens: not measured

## Dispatch 11 — verifier

- Task: no task
- Role: verifier
- Key: verify
- Model: opus effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-30T11:27:50Z
- Tokens: not measured

## Dispatch 12 — verifier

- Task: no task
- Role: verifier
- Key: visual-verify
- Model: opus effort=low
- Commit: no commit
- Outcome: blocked
- Started: 2026-09-30T11:30:55Z
- Tokens: input 20, output 198, cache read 310555, cache creation 45086

## Dispatch 13 — verifier

- Task: no task
- Role: verifier
- Key: visual-verify-2
- Model: opus effort=low
- Commit: no commit
- Outcome: blocked
- Started: 2026-09-30T11:32:46Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-862-flow-fix-review-minors-inline-during-the-run-panel.md

# Review panel — kan-862-flow-fix-review-minors-inline-during-the-run

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Minor | stats/cmd/flow/record.go:157 | three comments still justify keeping the deferred vocabulary by a dashboard feature this change removed (the deferred-Minor breakdown, numerator and rate) |   |
| F2 | primary | Minor | spectre/changes/kan-862-flow-fix-review-minors-inline-during-the-run/design.md:41 | design.md and Task 4 Step 1 still say 'the one source change', contradicting the delivered review-panel.md; Task 4 has no Correction note |   |
| F3 | principles | Minor | stats/internal/guard/unfinishedwork.go:347 | the two open-finding predicates kept identical on purpose now disagree: the panel guard reads deferred as open, the integrate gate still reads it as closed |   |

findings-total: 3
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed

reproducers-total: 3
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-principles-1.sh

## Pass log

### Round 0

- diff size: 1765 lines, under cap; docs-only: exit 1 (first non-doc path scripts/check-panel-findings-closed.sh) — resolved roster unchanged: primary+principles; roster: compact — 6; standards: CLAUDE.md, AGENTS.md; no addition this round — the resolved list ran alone; citation check: ran scripts/check-references.sh (exit 0)
- decided under: KAN-862's own Panel re-runs rule (worktree review-panel.md) — a Minor-only round's Minors fixed inline by the parent in 977b3f4f, no slot re-run; the installed main-checkout guard, still on the Minor-deferral rule, exits 1 demanding re-runs for F1-F3, the worktree guard exits 0 FINDINGS-CLOSED
- visual-verify capture run by the parent on the operator's explicit approval after verifier visual-verify-2's capture was denied by the permission classifier
## spectre/changes/archive/kan-862-flow-fix-review-minors-inline-during-the-run/tasks.md

# kan-862-flow-fix-review-minors-inline-during-the-run

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** review Minors are fixed inline by the parent at the moment they are raised. Nothing is
deferred to `KNOWN-BUGS.md` again, and the deferral surface is removed.

**Architecture:** two guard/CLI changes (Tasks 1–2) and one dashboard removal (Task 3) land first.
The prose that relies on them follows: the inline Minor path and the KNOWN-BUGS contract in one
citation-closed commit (Task 4), the DRIFT CHECK paragraph (Task 5), then the handoff and stray
mentions (Task 6). The backlog deletion (Task 7) and a live check (Task 8) come last.

**Spec:** `design.md` beside this file is canonical for every scope decision. Tasks cite its
`## Decisions` by ID and never restate them.

## Global Constraints

- Go: new logic is in the existing files named below; no new packages, no new dependencies.
- No new inline suppressions, and no weakening of any `scripts/check-*` guard.
- A commit's scope names the module it moved (**Commit scopes name the module**,
  `rules/commit-scope-is-the-module.mdc`).
- Every run-loaded sentence a task deletes or rewords on purpose is listed, exactly as
  `scripts/check-verbatim-moves.sh` prints it after `::`, in
  `spectre/changes/kan-862-flow-fix-review-minors-inline-during-the-run/verbatim-moves.txt`. The
  file is edited in the worktree and **left uncommitted**: it is a planning path, carried by the
  next `chore(spectre): plan` commit (**Planning commits**,
  `skills/flow-contracts/git-boundaries.md`), never by a task commit.
- Every task's verify step runs the lint lines from `.flow/project.md` `## lint` that its
  `**Files:**` need. For Go: `cd stats && gofmt -l . && go vet ./...`. For Markdown:
  `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`,
  `scripts/check-installed-citations.sh`, `scripts/check-verbatim-moves.sh`,
  `scripts/check-dispatch-paragraphs.sh`, `scripts/check-stage-mark-calls.sh`. For the SPA:
  `cd stats/web && npx tsc -b`. Each also runs the task's targeted tests.
- Task 3 touches `stats/web/src/**`, so `flow.visual-verify` runs. The run-detail baseline in
  `stats/web/tests/visual/` loses its `Deferred minor` panel; that stage owns the regeneration,
  not Task 3.

## Review Focus

- **A leftover `deferred` row.** A finding recorded `deferred` by an older skill copy mid-flight
  must read as open, never closed — Task 1, cases 4b and 13.
- **A fixed Minor with no slot re-run.** A slot raised only Minors in a round where another slot's
  Important went to the fix. Its fixed Minor must not be demanded a re-run, while the Important's
  slot still is — Task 1, `Case 30`.
- **handoff-lines with the store down.** `flow record handoff-lines` must print both of its lines
  and make no findings request — Task 2, `TestRecordHandoffLinesReadsNoFindings`.
- **`/flow-fast` reach.** The inline Minor rule must sit inside `## Panel re-runs`, not after it,
  or `/flow-fast`'s "through **Panel re-runs**" range misses it — Task 4, Step 6's grep.
- **A slot marked stale by the Minor commit.** The inline Minor commit must not make any slot's
  result stale — Task 4's carve-out sentence. It is prose only; the panel reads it.

**Live verification:** Task 8 exercises the guard, `flow record handoff-lines` and the reviewers
API against the worktree's own isolated stack, recording before (merge-base build) and after.

---

- [x] 1. check-panel-findings-closed reads `deferred` as open and needs no re-run for a fixed Minor

`design.md` § 3.

  - [x] **Step 1: Write the failing cases** in `stats/internal/guard/check_panel_findings_closed_test.go`
    (table in `TestCheckPanelFindingsClosed`):
    - case 4b becomes `"case 4b: a deferred finding is open, exits 1, naming it"` — `want: 1`,
      needle `"check-panel-findings-closed: finding(s) still open: F1\n"`.
    - Cases 11, 12, 13, 15 and 16 keep their fixtures and now expect `want: 1`, each deferred
      Minor named on the `finding(s) still open: …` line in ref order. Case 13's label becomes
      `"case 13: a Minor-only round with its Minors deferred exits 1, naming both as open"`, and
      each other label's "exits 0" or "never deferred" wording becomes the open verdict. No
      needle may carry `never deferred` any more.
    - Case 17 becomes `"case 17: an open finding and a deferred Minor are both reported open"`,
      needle `"finding(s) still open: F1 F2\n"`.
    - New `"case 29: a fixed Minor with no later re-run of its slot exits 0"`:
      `cfcFinding{Ref: "F1", Severity: "Minor", Status: "fixed", Slot: "primary"}`, no dispatch
      rows, `want: 0`.
    - New `"case 30: a fixed Minor needs no re-run while a fixed Important of another slot still does"`:
      `F1 Important fixed slot "failure-modes"`, `F2 Minor fixed slot "primary"`, both round 0,
      no dispatch rows; `want: 1`, needle
      `"with no clean re-run dispatch of their slot in any later round: F1 (failure-modes)\n"`,
      and assert stderr does not contain `F2`.
  - [x] **Step 2: Run them; they fail** —
    `cd stats && go test ./internal/guard -run 'TestCheckPanelFindingsClosed' -count=1`.
  - [x] **Step 3: Implement** in `stats/internal/guard/panelfindingsclosed.go`:
    - an open finding is one whose status is neither `fixed` nor a `withdrawn` value; drop the
      `deferred` prefix from that condition and from its comment;
    - delete the "A MINOR IS NEVER DEFERRED BESIDE A CRITICAL OR IMPORTANT" block (`fixRound`,
      `misdeferred`, its loop and message);
    - in the fixed-without-re-run loop, skip a finding whose severity is Minor
      (`strings.EqualFold(f.severity, "minor")`), and add one comment line: a fixed Minor closes
      on its fix alone (review-panel.md's **Panel re-runs**).
    In `scripts/check-panel-findings-closed.sh`'s header, the exit-0 line becomes "no finding is
    open or deferred, and every Critical or Important recorded `fixed` has a clean re-run
    dispatch of its slot in a later round". Exit 1 becomes "a finding is open or `deferred`, or
    a Critical or Important is recorded `fixed` with no clean re-run dispatch of its slot in any
    later round …". Drop the "wrongly deferred ref with its round" clause.
  - [x] **Step 4: Verify** — the targeted run passes; `cd stats && gofmt -l . && go vet ./...`
    is clean; `scripts/check-references.sh` is clean.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/guard/panelfindingsclosed.go`,
`stats/internal/guard/check_panel_findings_closed_test.go`, `scripts/check-panel-findings-closed.sh`
**Tests:** `Case 29`, `Case 30`
**Regression:** reverting brings back `deferred` read as closed and fixed Minors held to the
re-run rule. Cases 4b, 13 and 29 fail, and case 30's stderr names F2.
**Baseline:** before=25 after=27
<!-- measured: grep -c 'label: "case ' stats/internal/guard/check_panel_findings_closed_test.go @ 0b7e1c58 -->
**Commit:** `fix(guard): read a deferred finding as open and need no re-run for a fixed Minor`
**After:** none
**Build:** green

**Decision:** no-deferral-withdraw-only
**Decision:** store-keeps-deferred-vocabulary

- [x] 2. flow record handoff-lines prints Records and Costs only

`design.md` § 4, "Handoff".

  - [x] **Step 1: Rewrite the tests** in `stats/cmd/flow/record_test.go`:
    - `handoffDaemon(t)` takes no findings arguments. It answers `/cost-status` as now, and
      fails the test (`t.Errorf("handoff-lines requested %s; it reads cost-status only", r.URL.Path)`)
      on any other path.
    - In `TestRecordHandoffLinesFormat`, drop the `mixed` fixture and the `findings` column. The
      three wants become `"**Records:** all writes reached the store\n**Costs:** 0 unattributed\n"`,
      `"**Records:** 2 write(s) journalled — the store was unreachable\n**Costs:** 0 unattributed\n"`
      and `"**Records:** unknown — the journal could not be counted\n**Costs:** 0 unattributed\n"`.
      Its doc comment names the two lines only.
    - `TestRecordHandoffLinesDedupByProject` drops the `reads` counter and expects
      `"**Records:** 3 write(s) journalled — the store was unreachable\n**Costs:** 0 unattributed\n"`.
    - Replace `TestRecordHandoffLinesUnknown` (the findings read it pinned is gone) with
      `TestRecordHandoffLinesReadsNoFindings`: a daemon whose records endpoint answers 500 and
      whose `/cost-status` answers as `handoffDaemon` does; the verb exits 0, prints exactly
      `"**Records:** all writes reached the store\n**Costs:** 0 unattributed\n"`, and the records
      endpoint's hit counter reads 0.
  - [x] **Step 2: Run them; they fail** —
    `cd stats && go test ./cmd/flow -run 'TestRecordHandoffLines' -count=1`.
  - [x] **Step 3: Implement** in `stats/cmd/flow/record.go`'s `runRecordHandoffLines`:
    - delete the `deferred`/`findingsOK` block and the `**Deferred:**` and `### Deferred minors`
      prints, and the now-unused `handoffFindingsUnknown`, `handoffDeferredPrefix` and
      `handoffDeferredMinorsNone` constants;
    - the function's doc comment and the verb's usage paragraph (the "handoff-lines prints the
      verify handoff's record lines …" text) name `**Records:**` and `**Costs:**` only, with no
      findings read;
    - `stats/cmd/flow/main.go`'s usage line becomes
      `record handoff-lines  print the handoff's Records and Costs lines`.
  - [x] **Step 4: Verify** — the targeted run and `cd stats && go test ./cmd/flow -count=1` pass;
    `cd stats && gofmt -l . && go vet ./...` is clean.
  - [x] **Step 5: Commit.**

**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/cmd/flow/main.go`
**Tests:** `TestRecordHandoffLinesReadsNoFindings`
**Regression:** reverting brings back the Deferred line, the list and the findings request.
`TestRecordHandoffLinesFormat`'s wants fail, and `TestRecordHandoffLinesReadsNoFindings` sees
the records endpoint hit.
**Baseline:** before=94 after=94
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ 0b7e1c58 -->
**Commit:** `refactor(flow): drop the Deferred line and list from record handoff-lines`
**After:** none
**Build:** green

**Decision:** handoff-drops-deferred-line

- [x] 3. The dashboard drops its deferred metrics

`design.md` § 6.

  - [x] **Step 1: Update the tests first**:
    - `stats/internal/store/aggregate_test.go` (`TestReviewersCountsBySeverityAndMarksExperimental`):
      delete the two `DeferredShare` assertions. Keep the deferred-finding fixture row, since it
      still counts toward `total` and `withdrawnShare`.
    - `stats/internal/api/stats_test.go` (`TestEveryViewCarriesItsRealNumbersThrough`): drop
      `DeferredShare` from the fixture, from the decoded struct and from the comparison. Add
      `TestReviewersWireCarriesNoDeferredShare` beside it: the same reviewers fixture served
      through the handler, asserting the raw response body does not contain `deferredShare`
      while it still contains `withdrawnShare`.
    - `stats/web/src/views/views.test.tsx`: drop `deferredShare` from both fixture rows and
      delete the `"33%"` deferred-share assertion.
    - `stats/web/src/views/RunDetail.test.tsx`: delete the `describe("the deferred-Minor ratio
      (kan-508)", …)` block (5 `it`s) and the comment above it, plus `fetchRunRecordMock` (its
      hoist, its `vi.mock` entry, its `beforeEach` reset and default) and `runRecordEnvelope`.
    - `stats/web/src/api.test.ts`: delete `it("fetchRunRecord requests the change's run record
      endpoint", …)` and its import.
  - [x] **Step 2: Run them; they fail** —
    `cd stats && go test ./internal/api -run 'TestReviewersWireCarriesNoDeferredShare' -count=1`
    (fails while the field exists), and `cd stats/web && npx tsc -b`
    (`deferredShare` is missing from the fixtures).
  - [x] **Step 3: Implement**:
    - `stats/internal/store/aggregate.go`: delete the `deferred` count in `finding_agg`, the
      `DeferredShare` `CASE` column and its `Scan` target, and `ReviewerRow.DeferredShare`. Its
      doc comment reads "…what it found by severity, and how much of it was withdrawn rather than
      fixed".
    - `stats/internal/api/stats.go`: delete `reviewerRowDTO.DeferredShare` and its mapping in
      `toReviewerDTOs`.
    - `stats/web/src/api.ts`: delete `deferredShare` from the reviewer row type, and
      `RunRecordFindingDTO`, `RunRecordDTO` and `fetchRunRecord` with their doc comments.
    - `stats/web/src/views/Reviewers.tsx`: delete the `deferredShare` column; the header comment
      reads "…what it found by severity, and how much of it was withdrawn rather than fixed".
    - `stats/web/src/views/RunDetail.tsx`: delete the `Deferred minor` `RunPanel` and its KAN-508
      comment; the line-43 comment drops "deferred and".
    - `stats/web/src/hooks/useRunDetail.ts`: delete `DeferredMinorRatio`, `deferredMinorOf`, the
      `fetchRunRecord(...)` entry of the `Promise.all`, and the `deferredMinor` state field.
  - [x] **Step 4: Verify** — `cd stats && go test ./internal/store ./internal/api -count=1`
    (the store tests need the `flow-postgres` compose stack and skip without it: say so if they
    skipped); `cd stats/web && npx vitest run src/api.test.ts src/views && npx tsc -b`;
    `cd stats && gofmt -l . && go vet ./...`.
  - [x] **Step 5: Commit.**

**Files:** `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`,
`stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/web/src/api.ts`,
`stats/web/src/api.test.ts`, `stats/web/src/views/Reviewers.tsx`,
`stats/web/src/views/views.test.tsx`, `stats/web/src/views/RunDetail.tsx`,
`stats/web/src/views/RunDetail.test.tsx`, `stats/web/src/hooks/useRunDetail.ts`
**Tests:** `TestReviewersWireCarriesNoDeferredShare`
**Regression:** reverting brings `deferredShare` back onto the wire, and
`TestReviewersWireCarriesNoDeferredShare` fails.
**Baseline:** before=124 after=119
<!-- measured: cat stats/internal/store/aggregate_test.go stats/internal/api/stats_test.go stats/web/src/api.test.ts stats/web/src/views/views.test.tsx stats/web/src/views/RunDetail.test.tsx | grep -c -E '^func Test|^[[:space:]]*it\(' @ 0b7e1c58 -->
**Commit:** `refactor(stats): drop the deferred share and the deferred-Minor ratio`
**After:** none
**Build:** green

**Decision:** drop-deferred-metrics

Correction (2026-09-30): Step 2's `npx tsc -b` did not fail — the view fixtures are not typed
against `ReviewerRow`, so dropping `deferredShare` from them compiles either way; the Go
`TestReviewersWireCarriesNoDeferredShare` run was the RED. Beyond Step 3, the Reviewers view's
user-visible `ViewFrame` description drops "deferred or", and `RunDetail.tsx`'s local
`formatShare` and its comment are deleted, their only caller having been the removed panel.

- [x] 4. The parent fixes Minors inline; the KNOWN-BUGS deferral goes

`design.md` § 2 and § 4. One commit, because these files cite each other's deferral sections.

  - [x] **Step 1: `skills/flow/review-panel.md`, `## Panel re-runs`** — replace the paragraph
    opening "**Every Critical and Important goes to the fix; whether a Minor does follows…**"
    and the paragraph "**A deferral's reason is one clause naming the mechanism…**" with:

    > **Every Critical and Important goes to the fix; every Minor is fixed too, and a Minor never
    > causes a fix round of its own.** Every Critical and Important the round raised goes to the
    > fix subagent below, closed by the verification that follows it — the reproducer re-run
    > exits 0 *and* the fix diff touches a path the finding named, or one of the path condition's
    > accepted alternatives below closes the finding. **A round that raised a Critical or
    > Important sends every Minor it raised to that same fix**, closed by the same verification.
    > **A round that raised no Critical and no Important has its Minors fixed by the parent,
    > inline, at the round's close** — no fix subagent, no reproducer run, no mutation proof, no
    > slot re-run and no dispatch record. The parent edits each Minor's named location, stages
    > the changed paths and makes one commit per worktree at the branch tip —
    > `git -C <worktree> commit -m "fix(<module>): review Minors" -- <the changed paths>`, the
    > scope naming the module the edits moved — pushed per **Branch backup**
    > (`skills/flow-contracts/git-boundaries.md`), then records each finding
    > `flow record status -change <name> -ref F<n> -status fixed`. `flow.verify`'s lint and
    > tests, which run after this stage, are that fix's verification. **A Minor is never
    > deferred**: one no change to the tree can resolve, or that names no defect, is withdrawn
    > with `-status 'withdrawn <reason>'`, the reason one clause naming the mechanism — out of
    > scope, pre-existing and cosmetic are not reasons. Then proceed to
    > `check-panel-findings-closed.sh` and the stage close. A fixed finding that fails
    > verification takes the handback below, and that loop re-runs no slot either.

    In the stale-result paragraph ("Handoff still requires **zero open findings at any
    severity**…"), after "…covers what the fix changed.", insert: "**The parent's inline Minor
    commit leaves every slot's result current** — a source change after a slot's last read
    that does not make that result stale." In the rule-change paragraph, "— the Minor-deferral
    default above included —" becomes "— the Minor rule above included —". In the ROUND SCOPE
    paragraph, delete the sentence "Deferral rationale is written nowhere — a Minor is deferred
    with its one-clause reason in the store."
  - [x] **Step 2: `skills/flow/review-panel.md`, the close** — delete the heading
    `### Deferred findings go to KNOWN-BUGS.md at round close` and the paragraph under it,
    "**A round close that leaves findings newly recorded `deferred` records them in…**", so that
    "**Before closing the stage**…" continues `## Panel re-runs`. Replace the guard's exit-1
    sentence ("Exit 1 means a finding still reads `open` in the store, or a Minor reads
    `deferred`…" through "…in any later round.") with: "Exit 1 means a finding still reads `open`
    or `deferred` in the store — nothing is deferred (**Panel re-runs** above); the line names
    the refs. It also fires on the class **Panel re-runs**' ordering above added: a Critical or
    Important recorded `fixed` whose slot has no clean re-run dispatch of it in any later round —
    a fixed Minor needs none."
  - [x] **Step 3: `skills/flow/review-panel-fix-round.md`**:
    - "A Minor, fixed or deferred, re-runs no slot: a fixed Minor closes on the verification
      below alone." becomes "A Minor re-runs no slot: one fixed beside a Critical or Important
      closes on the verification below alone, and one the parent fixed inline (**Panel
      re-runs**, `skills/flow/review-panel.md`) on its commit."
    - "A Minor either fixed or deferred blocks nothing; a Minor left `open` blocks exactly as a
      Critical does." becomes "A Minor fixed or withdrawn blocks nothing; a Minor left `open` or
      `deferred` blocks exactly as a Critical does."
    - "…with none, it is deferred under the round's Minor-deferral default above." becomes
      "…with none, the parent fixes it inline, as **Panel re-runs**
      (`skills/flow/review-panel.md`) fixes a Minor-only round's Minors."
  - [x] **Step 4: `skills/flow/review-panel-late-fix.md`** — "and Minors defer under the standing
    rule" → "and a Minor-only result is fixed inline under the standing rule"; "or raises only
    Minors that defer under the standing rule," → "or raises only Minors, fixed inline under the
    standing rule,"; "or with every finding it raised a deferred Minor under the standing rule:"
    → "or with every finding it raised a Minor the parent fixed inline under the standing rule:".
  - [x] **Step 5: the per-task pass, the contract, the follow-up file**:
    - `skills/flow/implement.md`, **The gated per-task reviewer**: replace from "**A pass whose
      findings are all Minor is `clean`**" through "…`task-<n>` in place of `F<n>`." with:
      "**A pass whose findings are all Minor is `clean`** — no fix round and no re-review; before
      ticking the task the parent fixes each such Minor itself, inline, exactly as **Panel
      re-runs** (`skills/flow/review-panel.md`) fixes a Minor-only round's Minors — one commit at
      the branch tip, pushed, no dispatch and no dispatch record — and only while no implementer
      dispatch is writing that worktree. A Minor no change to the tree can resolve is named, one
      clause, in the parent's output; nothing is written to `KNOWN-BUGS.md`." (The following "A
      `fix` pass sends its Minors to the same fix." stays.)
    - `skills/flow-contracts/known-bugs.md`: the title becomes
      `# Known bugs — the KNOWN-BUGS.md sweep`; the opening sentence ends "…the sweep that
      records pre-existing test failures.**"; delete `## Deferred review findings` with its body,
      and the **Where it runs** bullet "**Review panel round close** — deferred findings only…".
    - `skills/flow-contracts/jira-followups.md`: delete the sentence "The review panel's deferred
      findings go to `<project>/KNOWN-BUGS.md` instead (**Deferred review findings**,
      `skills/flow-contracts/known-bugs.md`)."
  - [x] **Step 6: Verify** —
    `grep -rn -i "defer" skills/flow/review-panel.md skills/flow/review-panel-fix-round.md skills/flow/review-panel-late-fix.md skills/flow-contracts/known-bugs.md skills/flow-contracts/jira-followups.md`
    prints only the kept "`deferred`"-as-open wording from Steps 2 and 3. Next,
    `awk '/^## /{h=$0} /inline, at the round.s close/{print h}' skills/flow/review-panel.md`
    prints `## Panel re-runs`, which is the `/flow-fast` reach. Then run the Markdown lint lines,
    and record `scripts/check-verbatim-moves.sh`'s FAIL sentences in `verbatim-moves.txt` until it
    exits 0.
  - [x] **Step 7: Commit.**

**Files:** `skills/flow/review-panel.md`, `skills/flow/review-panel-fix-round.md`,
`skills/flow/review-panel-late-fix.md`, `skills/flow/implement.md`,
`skills/flow-contracts/known-bugs.md`, `skills/flow-contracts/jira-followups.md`
**Tests:** none — prose; Task 1's cases pin the guard these sentences describe
**Regression:** none — prose; reverting restores the deferral text, which Task 1's guard then
contradicts.
**Baseline:** before=0 after=0
<!-- measured: no test files in this task @ 0b7e1c58 -->
**Commit:** `feat(flow): the parent fixes review Minors inline instead of deferring them`
**After:** Task 1
**Build:** green

**Decision:** parent-fixes-minors-inline
**Decision:** minor-fix-no-own-verification
**Decision:** no-deferral-withdraw-only
**Decision:** minor-rule-inside-panel-reruns
**Decision:** minor-commit-at-tip

Correction (2026-09-30): the gated per-task review found two plan-verbatim defects, fixed in the
`fix(flow): review Minors` commit at the tip. Step 1's carve-out said "the one source change",
false beside the no-finding and late-fix carve-outs; it now reads "a source change", here and in
`design.md` § 2. Step 5's per-task path said "exactly as **Panel re-runs**", whose procedure ends
in a `flow record status` write for a ref the store never holds on a per-task pass; the dash
clause now adds "no `flow record status`".

- [x] 5. The implementer's DRIFT CHECK paragraph

`design.md` § 5.

  - [x] **Step 1: Add the paragraph** to `skills/flow/implement.md`'s "Every implementer dispatch
    **must** carry" list, directly after the **TARGETED TESTS** paragraph, verbatim:

    > **DRIFT CHECK:** Before you commit, read your own diff once for what it leaves behind. For
    > every behaviour, name, path, flag or exit code the diff changes, grep for the comments, file
    > headers, usage lines, docs and cross-file citations that describe it, and bring each one in
    > line in this same commit. Every refusal or exit path the diff adds gets a test case of its
    > own.

    Add `DRIFT CHECK` to **Inline — the parent implements**' list of implementer paragraphs that
    bind the parent ("FLOW — COMMIT-PER-TASK, the TDD sub-skill, TARGETED TESTS, …").
  - [x] **Step 2: Verify** — `grep -c '^> \*\*DRIFT CHECK:\*\*' skills/flow/implement.md` prints
    `1`. Run the Markdown lint lines, and record the new paragraph's sentences in
    `verbatim-moves.txt` as `check-verbatim-moves.sh` prints them.
  - [x] **Step 3: Commit.**

**Files:** `skills/flow/implement.md`
**Tests:** none — a dispatch paragraph
**Regression:** none — prose.
**Baseline:** before=0 after=0
<!-- measured: no test files in this task @ 0b7e1c58 -->
**Commit:** `feat(flow): the implementer runs a drift check before committing`
**After:** Task 4
**Build:** green

**Decision:** drift-check-paragraph

- [x] 6. The handoff and every stray mention stop describing a deferral

`design.md` § 4.

  - [x] **Step 1: The handoff**:
    - `skills/flow/verify-and-handoff.md`: "**Produce the handoff's `Records:`, `Deferred:` and
      `Costs:` lines and `### Deferred minors` list with one call**" becomes "**Produce the
      handoff's `Records:` and `Costs:` lines with one call**". In the `## Implementation staged`
      template, delete the `**Deferred:** …` line and the `### Deferred minors` heading with its
      row line.
    - `skills/flow-contracts/handoff-blocks.md`: drop ` · **Deferred:** <count of deferred
      Minors>` from the `**Records:**` line; delete the `### Deferred minors` heading and its row
      line; delete the paragraph "**`Deferred` and `### Deferred minors` are on-disk, not
      `(run-only)`.**…".
  - [x] **Step 2: Stray mentions**:
    - `skills/flow-contracts/pipeline.md`: "- findings fixed, and findings deferred or
      withdrawn;" → "- findings fixed, and findings withdrawn;".
    - `skills/flow-contracts/operator-prompts-auto-resolution.md`: delete "a Minor's disposition, ".
    - `skills/flow-self-review/SKILL.md`: "…is left out of the prompt, and so is every finding
      the review panel deferred, since the panel already decided its disposition." → "…is left
      out of the prompt."
    - `README.md`: the `check-panel-findings-closed.sh` bullet becomes "no handoff while any
      finding is open or deferred, whatever its severity."
  - [x] **Step 3: The rationale** — append to `skills/flow/SKILL-rationale.md`'s
    `### review-panel.md — Panel re-runs` section, verbatim:

    > *Minors are fixed inline, never deferred (KAN-862).* The Minor-deferral default — a
    > Minor-only round deferred its Minors to `KNOWN-BUGS.md`, because a fix round cost more
    > than they were worth — had left 96 entries by `0b7e1c58`, 30 of them from kan-860 alone.
    > Fixing them later took whole `/flow` runs: KAN-861 planned 17 tasks at opus/high over 67 of
    > them. What made a fix round expensive was its subagent, reproducers, mutation proof and
    > slot re-runs, never the edits. The parent's inline fix drops all four, so a Minor now costs
    > about what writing its `KNOWN-BUGS.md` entry did.

    History stays as it is: `docs/**`, archived changes and the other `-rationale.md` records.
  - [x] **Step 4: Verify** — `grep -rn -i "deferred minor\|Deferred:\*\*" skills README.md`
    prints nothing outside `skills/flow-self-review/SKILL.md`'s own self-review `**Deferred:**`
    line, which is about deferred self-review. Run the Markdown lint lines, and record
    `check-verbatim-moves.sh`'s FAIL sentences in `verbatim-moves.txt` until it exits 0.
  - [x] **Step 5: Commit.**

**Files:** `skills/flow/verify-and-handoff.md`, `skills/flow-contracts/handoff-blocks.md`,
`skills/flow-contracts/pipeline.md`, `skills/flow-contracts/operator-prompts-auto-resolution.md`,
`skills/flow-self-review/SKILL.md`, `README.md`, `skills/flow/SKILL-rationale.md`
**Tests:** none — prose
**Regression:** none — prose.
**Baseline:** before=0 after=0
<!-- measured: no test files in this task @ 0b7e1c58 -->
**Commit:** `refactor(flow): retire Minor deferral from the handoff and the contracts`
**After:** Task 2, 4
**Build:** green

**Decision:** handoff-drops-deferred-line

Correction (2026-09-30): beyond Step 2, `operator-prompts-auto-resolution.md`'s "the
deferred-findings follow-up filing" became "the follow-up filing" — the integrate follow-up files
outstanding work, never deferred findings. Step 3's rationale paragraph cites
`<project>/KNOWN-BUGS.md` rather than a bare `KNOWN-BUGS.md`, which `check-installed-citations.sh`
refuses as rootless.

- [x] 7. KNOWN-BUGS.md loses its deferred Minors

`design.md` § 4 and § 7.

  - [x] **Step 1: Delete** the `## Deferred review findings` heading and every entry under it
    from `KNOWN-BUGS.md`. Keep the `# Known bugs` title and the sweep entry for
    `scripts/test-check-cleanup-complete.sh`.
  - [x] **Step 2: Verify** — `grep -c '^- ' KNOWN-BUGS.md` prints `1`;
    `grep -c 'Deferred review findings' KNOWN-BUGS.md` prints `0`; `scripts/check-references.sh`
    is clean.
  - [x] **Step 3: Commit.**

**Files:** `KNOWN-BUGS.md`
**Tests:** none — a record file
**Regression:** none — the deleted entries are the record.
**Baseline:** before=0 after=0
<!-- measured: no test files in this task @ 0b7e1c58 -->
**Commit:** `docs(known-bugs): delete the deferred review Minors`
**After:** Task 4
**Build:** green

**Decision:** delete-known-bugs-minors
**Decision:** kan-861-withdrawn

- [x] 8. Live verification against the worktree's own isolated stack

  - [x] **Step 1: Stack.** Bring up the worktree's isolated stack per `.flow/project.md`
    `## workspace isolation` (`scripts/workspace.sh create <id>`, its derived `flow_<id>` database
    and port), never the dev workspace's `flowd` on 4173 or its `flow` database. Export
    **both** `FLOW_ADDR` and `FLOW_RECORDS_ADDR` to the isolated port. `FLOW_RECORDS_ADDR` is
    deliberately not isolated by default, so leaving it unset writes the scratch findings into
    the persistent store.
  - [x] **Step 2: Before**, from binaries built in a throwaway detached worktree at the merge
    base (`git worktree add --detach <tmp> 0b7e1c58`; `go build` `flowd`, `flow` and
    `flow-guard`):
    1. Create a scratch change `kan-862-live-check`.
    2. Record one `panel-0-primary` reviewer dispatch, then, against it, `F1` Minor
       `deferred cosmetic` and `F2` Minor `fixed` (slot `primary`, round 0). No later-round
       dispatch exists.
    3. Run `check-panel-findings-closed.sh` against it (the base shim and its guard). Expect exit
       1 naming F2 as unverified, and F1 not open.
    4. Run `flow record handoff-lines -change kan-862-live-check -C <worktree>`, which prints a
       `**Deferred:** 1` line.
    5. `curl` the reviewers stats endpoint, whose rows carry `deferredShare`.
  - [x] **Step 3: After**, with the worktree HEAD's binaries against the same records: the guard
    exits 1 naming F1 as still open and nothing as unverified; `handoff-lines` prints exactly the
    `**Records:**` and `**Costs:**` lines; the reviewers response carries no `deferredShare`.
  - [x] **Step 4: Record** the before and after figures in the commit message body.
    **Not working looks like:**
    - the guard still passing F1, or still naming F2;
    - a `**Deferred:**` line or `### Deferred minors` in the after output;
    - `deferredShare` still in the after response;
    - any scratch row landing in the dev workspace's `flow` database — checked with
      `flow record findings -change kan-862-live-check` against 4173, which must print `[]`.
  - [x] **Step 5: Clean up** — remove the throwaway worktree and run
    `scripts/workspace.sh remove <id>`; the scratch change lived only in `flow_<id>`.
  - [x] **Step 6: Commit** (an empty commit carrying the record).

**Tests:** none — live run; the figures are the record
**Files:** none
**Regression:** none — verification only.
**Baseline:** before=0 after=0
<!-- measured: no test files in this task @ 0b7e1c58 -->
**Commit:** `test(stats): live-verify the Minor-deferral retirement`
**After:** Task 1, 2, 3
**Build:** green
## spectre/changes/archive/kan-862-flow-fix-review-minors-inline-during-the-run/design.md

## Context

One change, because the guard, the review prose, the handoff and the dashboard all retire the same
`deferred` status — §§ 1–7 below.

## 1. Where Minors go today, and what that costs

- **Per-task gated reviewer** (`skills/flow/implement.md`, **The gated per-task reviewer**): a pass
  whose findings are all Minor closes `clean`, and the parent appends each Minor to
  `KNOWN-BUGS.md` before ticking the task. A `fix` pass sends its Minors to the gated fix.
- **Panel round** (`skills/flow/review-panel.md`, **Panel re-runs**): a round with no Critical and
  no Important records every Minor `deferred <reason>` with a category, then appends it to
  `KNOWN-BUGS.md` (**Deferred findings go to KNOWN-BUGS.md at round close**). A round with a
  Critical or Important sends its Minors to that round's fix.
- **The pile:** at `0b7e1c58`, `KNOWN-BUGS.md` held 97 entries — 96 deferred Minors and one
  sweep entry for a pre-existing load-sensitive test failure. Kan-860 alone added 30.
  <!-- measured: grep -c '^- ' KNOWN-BUGS.md → 97; grep -c Minor KNOWN-BUGS.md → 96 @ 0b7e1c58 -->
- **The later fix is the expensive part.** KAN-861 (`/flow-fast`, class `big`, 17 tasks, 67
  findings, opus/high implementers, four waves) is the Minor-fixing route this change replaces.
- **What the Minors are:** about 46 are drift (comments, file headers, usage lines or citations
  that no longer match what the task changed), about 16 are missing test cases (an untested
  refusal or exit path), 18 are marked cosmetic, and the rest are edge cases.
  <!-- measured: keyword counts over the `## Deferred review findings` entries @ 0b7e1c58 -->

## 2. The inline Minor fix

- **Who fixes:** the parent, inline, at the point it wrote the `KNOWN-BUGS.md` entry before —
  reading the Minor costs about what writing the entry did. No dispatch, no dispatch record.
- **Panel:** a round with no Critical and no Important has every Minor fixed by the parent at the
  round's close: one commit per worktree at the branch tip (the on-top route, pathspec-scoped,
  pushed per **Branch backup**), then `flow record status -status fixed` per finding. No fix
  subagent, no reproducer run, no mutation proof, no slot re-run.
- **Per-task:** a Minor-only pass is still `clean`. Before ticking the task, the parent fixes its
  Minors the same way, one commit at the tip, applied only while no implementer dispatch is
  writing that worktree.
- **Unchanged:** a Minor beside a Critical or Important still goes to that round's fix or the
  gated fix, closed by that fix's verification.
- **Verification:** none of the fix's own. Per-task fixes are read by the panel, which reads the
  whole branch diff later. Panel fixes are covered by `flow.verify`'s lint and tests, which run
  after the panel.
- **Staleness carve-out:** the inline Minor commit leaves every slot's result current — a
  source change after a slot's last read that does not make that result stale. Without it, a
  Minor fix would force the re-runs this path exists to avoid.
- **The only exit is `withdrawn`:** a Minor no tree change can resolve, or one that is not a
  defect, is withdrawn with a one-clause reason, as the handback loop withdraws. Out of scope,
  pre-existing and cosmetic are not reasons: they are how 96 entries piled up.
- **`/flow-fast` parity:** `/flow-fast` runs `review-panel.md` "**Check base movement first**
  through **Panel re-runs**" and never runs `close-task.sh`, so no gated reviewer fires there. The
  new panel rule therefore has to live inside `## Panel re-runs` to reach `/flow-fast` with no
  edit of its own. `/flow-fast` runs no guards, so there the prose rule is the whole enforcement.

## 3. The findings-closed guard

`check-panel-findings-closed.sh` (`stats/internal/guard/panelfindingsclosed.go`) today:

- counts `deferred <reason>` as closed;
- fails a Minor recorded `deferred` in a round with a live Critical or Important;
- fails every `fixed` finding whose slot has no clean re-run in a later round, Minors included.
  This contradicts **Panel re-runs**' own "a fixed Minor closes on the verification below alone"
  wherever a slot raised only Minors in a round whose Important went to a fix.

After: `deferred` is open (so the misdeferred class collapses into "still open" and is deleted),
and a Minor recorded `fixed` needs no re-run. The store and CLI keep accepting and reading
`deferred`, because historical rows and migration `0025` carry it.

## 4. Retiring the deferral surface

- `skills/flow-contracts/known-bugs.md` loses `## Deferred review findings` and its **Where it
  runs** bullet. The sweep for pre-existing test failures stays.
- `skills/flow/review-panel.md` loses **Deferred findings go to KNOWN-BUGS.md at round close**,
  and the ROUND SCOPE paragraph's sentence on deferral. `review-panel-fix-round.md`'s
  non-convergence loop and `review-panel-late-fix.md`'s "Minors defer under the standing rule"
  switch to the inline fix.
- **Handoff:** the `**Deferred:**` line and `### Deferred minors` list are always `0`/`none` now,
  so they are removed from `verify-and-handoff.md`, `handoff-blocks.md` and
  `flow record handoff-lines`. That verb's findings read existed only for them, so it goes too,
  along with its unknown spelling.
- **Stray mentions:** `jira-followups.md`, `pipeline.md`'s "findings deferred or withdrawn",
  `operator-prompts-auto-resolution.md`'s "a Minor's disposition", `flow-self-review/SKILL.md`'s
  "every finding the review panel deferred", `README.md`'s guard line.
- `KNOWN-BUGS.md`: the `## Deferred review findings` heading and all 96 entries under it are
  deleted, unfixed. The title and the sweep entry stay.
- **Rationale:** the reversal of the Minor-deferral default is recorded in
  `skills/flow/SKILL-rationale.md`'s **review-panel.md — Panel re-runs** section. The old rule's
  recorded reasons stay where they are, as a rejected alternative's reasons.

## 5. Preventing Minors at the source

A new implementer dispatch paragraph in `skills/flow/implement.md`'s "Every implementer
dispatch **must** carry" list:

> **DRIFT CHECK:** Before you commit, read your own diff once for what it leaves behind. For every
> behaviour, name, path, flag or exit code the diff changes, grep for the comments, file headers,
> usage lines, docs and cross-file citations that describe it, and bring each one in line in this
> same commit. Every refusal or exit path the diff adds gets a test case of its own.

It targets the two biggest Minor classes above (about 62 of 96). Under `inline` execution it binds
the parent in the same words, as every implementer paragraph does (**Inline — the parent
implements**). It is not pinned in `check-dispatch-paragraphs`.

## 6. Stats: drop the deferred metrics

- Reviewers view: the `Deferred` column (`ReviewerRow.DeferredShare`, the aggregate's
  `status ILIKE 'deferred%'` count, the API's `deferredShare`).
- Run detail: the `Deferred minor` panel (KAN-508's `DeferredMinorRatio`). The run-record fetch
  in `useRunDetail` fed that panel alone, so the fetch, `fetchRunRecord` and the SPA's run-record
  DTO types go with it. The server's records endpoint stays: the CLI reads it.
- The store's `deferred` status vocabulary, categories and migration stay (**3** above).

## 7. KAN-861

KAN-861 (`kan-861-fix-all-minor-findings-in-known-bugs-md`, in flight, 6 commits) is withdrawn by
the operator in favour of deleting the backlog unfixed. This change never touches its worktrees
or its Jira issue; withdrawing it is the operator's act. This change deletes whatever Minor
entries its own base holds.

## Step-by-step breakdown

### Findings-closed guard

**What:** `check-panel-findings-closed` treats `deferred` as open and exempts a `fixed` Minor
from the clean-re-run requirement.
**Why:** it enforces "nothing is deferred" where the pipeline runs guards, and stops failing the
re-run-free Minor fix.
**Uses:** `stats/internal/guard/panelfindingsclosed.go`,
`stats/internal/guard/check_panel_findings_closed_test.go`, `scripts/check-panel-findings-closed.sh`.

### Handoff lines without Deferred

**What:** `flow record handoff-lines` prints `**Records:**` and `**Costs:**` only.
**Why:** with nothing deferred, the Deferred line and list are always `0`/`none`.
**Uses:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/cmd/flow/main.go`.

### Dashboard without deferred metrics

**What:** remove the Reviewers `Deferred` column and the Run detail `Deferred minor` panel, end
to end (store aggregate, API field, SPA).
**Why:** they would read 0% forever. The operator chose dropping them over keeping them as an
alarm.
**Uses:** `stats/internal/store/aggregate.go`, `stats/internal/api/stats.go`, `stats/web/src/`.

### Inline Minor fix in the review panel and per-task review

**What:** the panel's and the gated reviewer's Minor-only paths fix Minors inline; the KNOWN-BUGS
close step and the deferral contract section go.
**Why:** a Minor fixed where it is raised costs a few edits; fixed later, it costs a `/flow` run.
**Uses:** `skills/flow/review-panel.md`, `skills/flow/review-panel-fix-round.md`,
`skills/flow/review-panel-late-fix.md`, `skills/flow/implement.md`,
`skills/flow-contracts/known-bugs.md`, `skills/flow-contracts/jira-followups.md`.

### DRIFT CHECK paragraph

**What:** a new implementer dispatch paragraph requiring the drift grep and a test per new exit
path, before the commit.
**Why:** it prevents the two largest Minor classes, instead of fixing them after review.
**Uses:** `skills/flow/implement.md`.

### Handoff and stray mentions

**What:** remove the Deferred handoff line and list from the skill templates; fix every
remaining deferral mention; record the reversal's rationale.
**Why:** no run-loaded text may describe a deferral that no longer exists.
**Uses:** `skills/flow/verify-and-handoff.md`, `skills/flow-contracts/handoff-blocks.md`,
`skills/flow-contracts/pipeline.md`, `skills/flow-contracts/operator-prompts-auto-resolution.md`,
`skills/flow-self-review/SKILL.md`, `README.md`, `skills/flow/SKILL-rationale.md`.

### Delete the KNOWN-BUGS Minors

**What:** delete `## Deferred review findings` and its 96 entries from `KNOWN-BUGS.md`.
**Why:** the operator chose deleting over fixing; the entries' home no longer exists.
**Uses:** `KNOWN-BUGS.md`.

### Live verification

**What:** exercise the guard, `flow record handoff-lines` and the reviewers API against the
worktree's isolated store, with before (merge-base build) and after figures.
**Why:** the guard and the CLI read the running store, and flowd's aggregate query changes.
**Uses:** the worktree's isolated `flowd` (`## workspace isolation`), `scripts/workspace.sh`,
`FLOW_ADDR`/`FLOW_RECORDS_ADDR`.

## Decisions

### The parent fixes Minors inline

**ID:** parent-fixes-minors-inline
**Status:** active
**Chosen:** the parent, inline, at the point the Minor is read — no dispatch, one commit at the tip.
**Considered:** the raising reviewer fixing its own Minors in the same dispatch — it would edit
the worktree while the next group's implementer is also editing it, and reviewers would stop
being read-only. One sonnet/low fixer per boundary — the operator picked it first and then
reversed in favour of the parent: it adds a dispatch at every boundary.

### A Minor fix has no verification of its own

**ID:** minor-fix-no-own-verification
**Status:** active
**Chosen:** no re-review, reproducer or mutation proof; the panel (per-task fixes) and
`flow.verify` (panel fixes) cover it.
**Considered:** lint plus the task's targeted tests right after each Minor commit — catches a
break earlier, but costs a test run at every boundary.

### Nothing is deferred; withdraw is the only exit

**ID:** no-deferral-withdraw-only
**Status:** active
**Chosen:** every Minor is fixed; a Minor no tree change can resolve, or that is not a defect,
is withdrawn with one clause.
**Considered:** also allowing an out-of-scope withdrawal for code the change never touched — it
reopens the escape that let 96 entries accumulate.

### Delete the KNOWN-BUGS Minors unfixed

**ID:** delete-known-bugs-minors
**Status:** active
**Chosen:** delete every Minor entry and the heading.
**Considered:** fixing them all in this change — the expensive route this change exists to end.
Triage (fix what still stands, drop the stale) — still a sizeable pass for Minors.

### DRIFT CHECK implementer paragraph, unpinned

**ID:** drift-check-paragraph
**Status:** active
**Chosen:** one new implementer paragraph, prompt text only.
**Considered:** the same paragraph pinned in `check-dispatch-paragraphs` — a Go row and a test
nobody asked for. No prevention — leaves the two largest Minor classes to review.

### Drop the dashboard's deferred metrics

**ID:** drop-deferred-metrics
**Status:** active
**Chosen:** remove the Reviewers `Deferred` column and the Run detail `Deferred minor` panel end
to end.
**Considered:** leave as-is (no code, and any non-zero value on a new run would flag a broken
rule) — the operator chose dropping. Replace with a Minor-fix count — per-task Minors never reach
the store, so it would count panel Minors only.

### KAN-861 is withdrawn, not waited on

**ID:** kan-861-withdrawn
**Status:** active
**Chosen:** the operator withdraws KAN-861; this change deletes every Minor entry on its own base.
**Considered:** treating KAN-861 as independent — the operator chose to withdraw it. Landing
KAN-861's commits first — ties this change to a big run it replaces.

### The store keeps the `deferred` vocabulary

**ID:** store-keeps-deferred-vocabulary
**Status:** active
**Chosen:** the store and CLI keep accepting and reading `deferred <reason>` and its categories;
the guard reads it as open.
**Considered:** removing it from the store, the CLI, reconcile and the API errors — historical
rows and migration `0025` carry it, and nothing writes it once the skills stop.

### The Minor rule lives inside Panel re-runs

**ID:** minor-rule-inside-panel-reruns
**Status:** active
**Chosen:** state the inline Minor fix inside `## Panel re-runs`.
**Considered:** a section of its own after it — `/flow-fast`'s range ("through **Panel
re-runs**") would miss it.

### Drop the handoff's Deferred line and list

**ID:** handoff-drops-deferred-line
**Status:** active
**Chosen:** remove `**Deferred:**` and `### Deferred minors` from the templates and
`flow record handoff-lines`, along with the findings read that fed them.
**Considered:** keeping them — they would print `0`/`none` forever.

### One Minor commit at the tip

**ID:** minor-commit-at-tip
**Status:** active
**Chosen:** one pathspec commit per worktree at the branch tip, pushed like any other.
**Considered:** folding into the task commit — `fold-fixup.sh`'s set-aside, rebase and guard
re-runs are the cost this change removes; the branch is pushed at every boundary, so the on-top
route always applies.

## Open questions
## spectre/changes/archive/kan-862-flow-fix-review-minors-inline-during-the-run/narrative.md

# kan-862-flow-fix-review-minors-inline-during-the-run — session narrative

## 2026-09-30 — creating run

- Inline execution, 8 tasks. `flow record dispatch -effort` refuses `xhigh`; the parent's rows
  record `high`.
- The installed skill (main checkout) still carries the Minor-deferral rule this change retires.
  Both review rounds (the gated per-task bundle, 4 Minors; panel round 0, 3 Minors) were
  Minor-only and were fixed inline under this change's own rule. The installed findings-closed
  guard therefore exits 1 asking for slot re-runs; the worktree's guard exits 0.
- Panel F3 extended the change: the integrate gate (`unfinishedwork.go`) now also reads
  `deferred` as open, so the two duplicated predicates agree again.
- The citation pre-check's `project-get.sh` output is a fenced block; eval'ing it raw exit-127'd.
  It was re-run as the bare command.
- Task 8's first scratch write went to the journal (the scratch change did not exist yet); zsh
  did not word-split the cleanup `rm`, so the stray journal was removed by explicit path.
- Visual verify: `runs.spec.ts` fails on a baseline that has been stale since dd995076 (recorded
  in KNOWN-BUGS.md). The second verifier's capture was denied by the permission classifier; the
  operator approved it, and the parent ran capture, created the full app suite, and refreshed the
  stale Reviewers baseline (old copy passing within tolerance).

## 2026-09-30 — integrate run

- Preflight: guards present, STAGED-CLEAN, DRIFT-CLEAN, RUN1. Base not moved; no rebase.
- Unfinished-work gate: `check-unfinished-work.sh` CLEAR; `check-visual-verify-dispatched.sh`
  VISUAL-VERIFY-MISSING — no closed `visual-verify*` verifier row, because the parent ran capture
  after the verifier's capture was denied (see visual-verification.md). Operator chose
  **Continue — integrate anyway**.
- Route: merge and push, from the project's default.
## git log --stat

commit cde856d36a368bb659df2b06ead4ff14e458fcbc
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Wed Sep 30 16:44:10 2026 +0300

    feat(flow): fix review Minors inline and retire Minor deferral

 KNOWN-BUGS.md                                      | 159 +--------------------
 README.md                                          |   4 +-
 scripts/check-panel-findings-closed.sh             |  22 ++-
 skills/flow-contracts/handoff-blocks.md            |  12 +-
 skills/flow-contracts/jira-followups.md            |   4 +-
 skills/flow-contracts/known-bugs.md                |  22 +--
 .../operator-prompts-auto-resolution.md            |   4 +-
 skills/flow-contracts/pipeline.md                  |   2 +-
 skills/flow-self-review/SKILL.md                   |   3 +-
 skills/flow/SKILL-rationale.md                     |   8 ++
 skills/flow/implement.md                           |  23 +--
 skills/flow/review-panel-fix-round.md              |  14 +-
 skills/flow/review-panel-late-fix.md               |   8 +-
 skills/flow/review-panel.md                        |  66 ++++-----
 skills/flow/verify-and-handoff.md                  |   7 +-
 stats/cmd/flow/main.go                             |   2 +-
 stats/cmd/flow/record.go                           |  73 +++-------
 stats/cmd/flow/record_test.go                      |  92 +++++-------
 stats/internal/api/stats.go                        |   2 -
 stats/internal/api/stats_test.go                   |  26 +++-
 .../guard/check_panel_findings_closed_test.go      |  38 +++--
 stats/internal/guard/check_unfinished_work_test.go |   7 +-
 stats/internal/guard/panelfindingsclosed.go        |  42 ++----
 stats/internal/guard/unfinishedwork.go             |   7 +-
 stats/internal/records/types.go                    |   4 +-
 stats/internal/store/aggregate.go                  |  18 +--
 stats/internal/store/aggregate_test.go             |   6 -
 stats/internal/store/records.go                    |   4 +-
 stats/web/src/api.test.ts                          |  16 ---
 stats/web/src/api.ts                               |  30 ----
 stats/web/src/hooks/useRunDetail.ts                |  45 +-----
 stats/web/src/views/Reviewers.tsx                  |  14 +-
 stats/web/src/views/RunDetail.test.tsx             | 115 +--------------
 stats/web/src/views/RunDetail.tsx                  |  16 ---
 stats/web/src/views/views.test.tsx                 |   3 -
 stats/web/tests/visual/full-app-suite.spec.ts      |  37 +++++
 .../full-cache-efficiency-darwin.png               | Bin 0 -> 80178 bytes
 .../full-decisions-darwin.png                      | Bin 0 -> 74831 bytes
 .../full-flow-health-darwin.png                    | Bin 0 -> 139282 bytes
 .../full-reviewers-darwin.png                      | Bin 0 -> 66692 bytes
 .../full-run-detail-darwin.png                     | Bin 0 -> 93766 bytes
 .../full-runs-darwin.png                           | Bin 0 -> 108174 bytes
 .../full-stage-leaderboard-darwin.png              | Bin 0 -> 74823 bytes
 .../full-state-board-darwin.png                    | Bin 0 -> 99137 bytes
 .../full-trend-darwin.png                          | Bin 0 -> 73108 bytes
 stats/web/tests/visual/full-app-suite.zip          | Bin 0 -> 747326 bytes
 .../reviewers-darwin.png                           | Bin 66512 -> 66690 bytes
 47 files changed, 263 insertions(+), 692 deletions(-)

commit 16ecca342931ead11a40810062f7cc3be26bbcd0
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Wed Sep 30 16:44:10 2026 +0300

    chore(spectre): plan
    
    Integrated over: VISUAL-VERIFY-MISSING — no closed visual-verify verifier dispatch row; the
    parent ran the capture after the verifier's was denied (visual-verification.md).

 .../narrative.md                                                 | 9 +++++++++
 1 file changed, 9 insertions(+)

commit bba7741416aa6feef8c465beba23c50dd1f7f1ea
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Wed Sep 30 16:45:19 2026 +0300

    chore(spectre): archive kan-862-flow-fix-review-minors-inline-during-the-run

 .../design.md                                      |   0
 .../ledger.md                                      | 148 +++++++++++++++++++++
 .../narrative.md                                   |   0
 .../panel.md                                       |  26 ++++
 .../proposal.md                                    |   0
 .../tasks.md                                       |   0
 .../verbatim-moves.txt                             |   0
 .../visual-verification.md                         |   0
 8 files changed, 174 insertions(+)

## Session narrative

Run 2 followed run 1 in the same invocation via the merge-and-push route (project default). Run 1's
unfinished-work gate raised VISUAL-VERIFY-MISSING — no closed `visual-verify*` verifier row,
because the parent ran the capture after the verifier's was denied — and the operator integrated
over it. Cleanup's check 6 first returned HELD on the worktree's own flowd (pid 60477, :5903, left
from the visual-verify stack); the operator approved killing it, after which the worktrees, local
and remote branches and the `kan-862-flow-569c` workspace were removed and the cleanup verified
COMPLETE. The first FINISHED write also emptied `repos`; a restore attempt read back null.
