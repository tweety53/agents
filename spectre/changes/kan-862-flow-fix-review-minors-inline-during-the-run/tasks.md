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
    commit leaves every slot's result current** — the one source change after a slot's last read
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
