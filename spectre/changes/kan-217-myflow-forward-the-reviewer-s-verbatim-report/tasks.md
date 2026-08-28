# kan-217-myflow-forward-the-reviewer-s-verbatim-report

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

`design.md` is canonical for every decision below, for the filename rule, and for the guard's
resulting paragraph table. Nothing here restates them.

**Baseline, measured before any edit:**

- `scripts/` carries 30 guards and 38 test harnesses.
<!-- measured: ls scripts/ | grep -c '^check-'; ls scripts/ | grep -c '^test-' @ 8875344 (before this change) -->
- `scripts/test-check-reproduce-not-read.sh` runs 7 cases and makes 14 assertions, all passing.
<!-- measured: grep -c 'pass "' scripts/test-check-reproduce-not-read.sh; scripts/test-check-reproduce-not-read.sh @ 8875344 (before this change) -->
- `scripts/check-references.sh` exits 0.
<!-- measured: scripts/check-references.sh; echo $? @ 8875344 (before this change) -->
- `skills/flow/review-panel.md` is 24572 bytes against a `check-contract-budget.sh` budget of
  27420; `skills/flow-contracts/artifacts-registry.md` is 6393 against 7620.
<!-- measured: wc -c skills/flow/review-panel.md skills/flow-contracts/artifacts-registry.md; grep -n 'review-panel.md\|artifacts-registry.md' scripts/check-contract-budget.sh @ 8875344 (before this change) -->

**Task order is load-bearing.** Task 3 generalizes the guard to require the VERBATIM REPORT
paragraph at `skills/flow/review-panel.md`, and the guard self-scopes to this repository's own
root. Landing task 3 before task 1 would leave `## lint` red on a paragraph that does not exist
yet. Tasks 1 and 2 are independent of each other.

**Every task that grows an owned `.md` file runs `scripts/check-contract-budget.sh` and reconciles
its `budgets()` row in the same commit** — raise a row only where the guard actually fails on that
file, and only to the size its own rule gives. Both files this change grows have headroom measured
above, so no row is expected to need raising.

**The stale citation in `spectre/changes/kan-121-run-the-guards-own-parsers-over-tasks-md-at-plan/tasks.md`
is left alone**, and no task names it. See `design.md`'s `leave-finished-plans-frozen`.

---

- [x] 1. Capture, forward, and stop paraphrasing

  Three edits to `skills/flow/review-panel.md`, all prose.

  **Capture**, in **Recording findings, and the record's format**, beside the existing dispatch
  recording: as each slot's report comes back and its `flow record dispatch end` is recorded, write
  that report byte for byte to `<abs-worktree>/.superpowers/sdd/panel-report-<round>-<id>.md`.
  State that `<round>` is the same value that round's findings carry on `-round` (`0` initial,
  `1..n` fix rounds) and `<id>` is the resolved reviewer id, never the slot display name — so
  `Code review (low)` needs no slugging rule. State that every dispatched slot writes one,
  including a slot that raised nothing and a slot substituted per **An unspawnable id is
  substituted, not skipped**, and that a report which cannot be captured verbatim still writes its
  file, carrying the single line `no verbatim report captured — <reason>`.

  **Forward**, at **Carry each surviving finding to the fix subagent as a structured block**
  (line 395): add its slot's report path to the per-finding field list, and add the paragraph below
  to the fix subagent's dispatch prompt. Leave **Inline no source excerpt** exactly as it stands —
  the report is handed over as a path, so nothing is inlined by it.

  ```markdown verified:authored in-tree for this change; the three phrases below are the literals task 3's guard table holds
  > **VERBATIM REPORT — THE FACT:** each finding below names the file its slot's report was written
  > to. That file is the reviewer's own report, unedited — read it before you act on the finding.
  > The structured block is the dispatcher's summary of it: direction on what to work on, and
  > never a source of fact. Where the two disagree the report wins. Where the block asserts
  > something the report does not, treat it as unchecked and establish it yourself before building
  > on it.
  ```

  **Stop paraphrasing**, at the `flow record finding` call (line 205): `-note` carries the
  reviewer's own sentence naming the defect, not a dispatcher restatement. Where the reviewer's
  wording runs long, quote the sentence that names the defect and leave the rest to the report
  file.

  Run `scripts/check-references.sh`, `scripts/check-vocabulary.sh`,
  `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh` and
  `scripts/check-reproduce-not-read.sh` and leave all five clean.

**Files:** `skills/flow/review-panel.md`
**Tests:** none — this task edits skill prose; what verifies it is `scripts/check-references.sh`,
  `scripts/check-vocabulary.sh`, `scripts/check-markdown-integrity.py` and
  `scripts/check-contract-budget.sh`, with task 3's renamed guard covering the new paragraph's
  continued presence from then on
**Regression:** reverting this commit returns the fix subagent to receiving the dispatcher's
  distillation alone — the verdict without the reviewer's evidence — which is the path that
  produced KAN-189's F20 and F23.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `docs(flow): forward the reviewer's verbatim report to the fix agent`
**Build:** green

- [x] 2. Register the report files

  One row in `skills/flow-contracts/artifacts-registry.md`'s table, in the shape the rows around it
  use:

  | Artifact | Created by | Location | Removed by |
  |----------|-----------|----------|-----------|
  | Panel slot verbatim reports | `/flow`'s review panel | `<abs-worktree>/.superpowers/sdd/` in the worktree | with the worktree, at run 2 |

  This is what **An artifact no row accounts for is a defect in the registry** requires. The files
  are matched by `.gitignore`'s existing `.superpowers/` entry, so they are never committed and
  never archived — state that in the row's own vicinity only if the table's existing prose does not
  already cover it; do not restate the `Per-task and review diffs` row's reasoning.

  Run `scripts/check-references.sh`, `scripts/check-markdown-integrity.py` and
  `scripts/check-contract-budget.sh` and leave all three clean.

**Files:** `skills/flow-contracts/artifacts-registry.md`
**Tests:** none — this task adds one table row to a contract; what verifies it is
  `scripts/check-references.sh`, `scripts/check-markdown-integrity.py` and
  `scripts/check-contract-budget.sh`
**Regression:** reverting this commit leaves an artifact the pipeline creates with no row
  accounting for it, which that file defines as a registry defect.
**Baseline:** before=0 after=0 tests added
<!-- predicted: no test file is added by this task -->
**Commit:** `docs(flow-contracts): register the panel report files`
**Build:** green

- [x] 3. Generalize the dispatch-paragraph guard and rename it

  RED before GREEN: extend the harness first, watch it fail, then generalize the guard.

  **Rename both files with `git mv`**, so the history follows:
  `scripts/check-reproduce-not-read.sh` → `scripts/check-dispatch-paragraphs.sh`, and
  `scripts/test-check-reproduce-not-read.sh` → `scripts/test-check-dispatch-paragraphs.sh`.
  Rename the environment override with them: `CHECK_REPRODUCE_NOT_READ_ROOT` →
  `CHECK_DISPATCH_PARAGRAPHS_ROOT`, keeping its set-but-empty refusal exactly as it is.

  **Collapse the single-paragraph constants into one table.** `LABEL`, `SHARED_PHRASES`,
  `REVIEWER_PHRASE`, `IMPLEMENTER_PHRASE`, `SITE_PATHS`, `SITE_MIN_BLOCKS` and `SITE_VARIANTS`
  become one paragraph table whose entries each carry a label, its shared phrases, its variants'
  own phrases, and its required sites with each site's minimum block count and required variants.
  The REPRODUCE, DON'T READ entry reproduces today's values unchanged. The second entry is
  `**VERBATIM REPORT — THE FACT:**`, required once at `skills/flow/review-panel.md`, with no
  variants and these three phrases held as short literals:

  ```text verified:authored in-tree for this change; each substring appears in task 1's paragraph
  the reviewer's own report
  never a source of fact
  the report wins
  ```

  `extract_block_text`, the `file:line` reporting shape via `report_line`, the `die2` refusal, and
  the exit contract — `0` clean, `1` a required site missing its block or one of its phrases, `2`
  it cannot answer at all — are unchanged. Rewrite the module docstring for what the guard now
  covers, keeping its **WHAT A GREEN RUN DOES NOT PROVE** paragraph and widening it: a green run
  never proves a dispatcher wrote a report file, and never proves a fix agent read one.

  **The harness** keeps its existing 7 cases, retargeted at the renamed guard, and adds cases for
  the new paragraph: present and clean; the site missing the block entirely; and one case per
  required phrase missing, each asserting exit 1 and that the reported line names the missing
  phrase. Per KAN-197, **mutate the guard to prove each new case actually fails** — a case that
  passes against a guard with its own check removed proves nothing. Build every fixture under
  `TMPDIR` via the existing `new_root`/`write_site` helpers and point the guard at it with
  `CHECK_DISPATCH_PARAGRAPHS_ROOT`; never let a case touch this repository.

  **Follow the rename through its live citations.** `.flow/project.md`: the `## test` entry
  (line 106), the `## lint` entry (line 158), and the explanatory paragraph (line 247), which is
  rewritten for a guard covering a table of paragraphs rather than one. `scripts/test-check-plan-shape.sh`
  (line 20), whose comment cites the harness by its old name.

  Run `scripts/check-dispatch-paragraphs.sh`, `scripts/test-check-dispatch-paragraphs.sh`,
  `scripts/check-guard-symlinks.sh` and `scripts/check-references.sh` and leave all four clean.

**Files:** `scripts/check-dispatch-paragraphs.sh`, `scripts/test-check-dispatch-paragraphs.sh`, `scripts/check-reproduce-not-read.sh`, `scripts/test-check-reproduce-not-read.sh`, `.flow/project.md`, `scripts/test-check-plan-shape.sh`
**Tests:** `scripts/test-check-dispatch-paragraphs.sh`
**Regression:** reverting this commit removes the only check that either dispatch paragraph is
  still present at its required sites, returning both to depending on no later prose edit trimming
  them away a line at a time — the failure KAN-289 established the original guard to stop, now
  covering the paragraph this change adds as well.
**Baseline:** before=14 after=22 assertions in the harness; before=38 after=38 harnesses in `scripts/`
<!-- predicted: grep -c 'pass "' scripts/test-check-dispatch-paragraphs.sh, and ls scripts/ | grep -c '^test-', after task 3 -->
**Commit:** `refactor(scripts): guard dispatch paragraphs from a table`
**Build:** green
