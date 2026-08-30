# kan-263-myflow-forbid-backgrounded-builds-in-the-shared

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

- [x] 1. Add `FOREGROUND BUILDS` at all four dispatch sites, guarded by `check-dispatch-paragraphs.sh`

**Build:** green
**Files:** `skills/flow/implement.md`, `skills/flow/review-panel.md`,
`scripts/check-dispatch-paragraphs.sh`, `scripts/test-check-dispatch-paragraphs.sh`
**Tests:** `scripts/test-check-dispatch-paragraphs.sh`'s new cases — one asserting `FOREGROUND
BUILDS` missing entirely from a required site is caught, three asserting each of its shared
phrases missing is caught individually
**Regression:** if `FOREGROUND BUILDS` disappears from any of the four dispatch sites,
`scripts/check-dispatch-paragraphs.sh` — and its own test suite's new cases — fail loud instead of
the removal passing silently.
**Baseline:** before=44 after=44
<!-- measured: scripts/run-guard-tests.sh @ branch spectre/kan-263-myflow-forbid-backgrounded-builds-in-the-shared (44 harnesses, 44 passed) -->
<!-- predicted: scripts/run-guard-tests.sh after task 1 (44 harnesses, 44 passed — same harness count, test-check-dispatch-paragraphs.sh grows internally but is still one harness) -->
**Commit:** feat(flow): forbid backgrounded builds in the shared dispatch preamble

  - [ ] **Step 1: Add the full paragraph to both `implement.md` sites**

    In `skills/flow/implement.md`, insert this exact blockquote twice: once in the implementer
    dispatch block, immediately after `PLAN PROVENANCE` (line 233) and before
    `REPRODUCE, DON'T READ` (line 235); once at the per-task reviewer dispatch, immediately before
    its own `REPRODUCE, DON'T READ` block (line 267).
    <!-- measured: grep -n '' skills/flow/implement.md | sed -n '215,240p;265,272p' @ branch spectre/kan-263-myflow-forbid-backgrounded-builds-in-the-shared -->

```markdown verified:authored in-tree for this change
> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.
```

    State plainly, in the implementer block's own lead-in sentence, that every implementer
    dispatch **must** carry this paragraph too, the same way it already says so for the ones
    before it.

  - [ ] **Step 2: Add the full paragraph to both `review-panel.md` sites**

    In `skills/flow/review-panel.md`, insert the same exact blockquote twice: once at the panel
    slot dispatch, immediately after the `CONTEXT BUNDLE` citation line and before the
    `REPRODUCE, DON'T READ` lead-in (between lines 138 and 140); once at the panel-fix subagent
    dispatch, immediately after the `VERBATIM REPORT — THE FACT` blockquote (after line 472).
    <!-- measured: grep -n '' skills/flow/review-panel.md | sed -n '135,146p;463,473p' @ branch spectre/kan-263-myflow-forbid-backgrounded-builds-in-the-shared -->

```markdown verified:authored in-tree for this change
**Every slot's dispatch prompt also carries the FOREGROUND BUILDS paragraph**:

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.
```

    Word the panel-fix site's own lead-in the same way, adapted to "Every fix subagent's dispatch
    prompt also carries the FOREGROUND BUILDS paragraph".

  - [ ] **Step 3: Add the `foreground` entry to the guard's paragraph table**

    In `scripts/check-dispatch-paragraphs.sh`:
    - `ENTRY_LABEL`: add `[foreground]="**FOREGROUND BUILDS:**"`.
    - `ENTRY_SHARED_PHRASES`: add `[foreground]="still executing in the background${US}run it in
      the foreground${US}poll it to completion"`.
    - `ENTRY_VARIANTS`: add `[foreground]=""` — no variants; every block carrying the label must
      carry all three phrases, identically at every site.
    - Extend the four `SITE_*` parallel arrays with two new rows: `(foreground, implement.md, 2,
      "")` and `(foreground, review-panel.md, 2, "")` — two required blocks per file, matching the
      two dispatch sites added per file in Steps 1–2.
    - Update the script's own header comment (the paragraph table listing and the module
      docstring's enumeration) to describe the new entry and its sites, matching how the existing
      header documents `reproduce` and `verbatim`.

  - [ ] **Step 4: Extend the guard's test suite**

    In `scripts/test-check-dispatch-paragraphs.sh`:
    - Add a correct `FOREGROUND BUILDS` block to every fixture file the suite already builds that
      must exit `0` — most directly Case 1's `review-panel.md`/`implement.md` fixtures, plus any
      other case whose fixture is asserted clean — so the new required site does not turn an
      existing "everything correct" case into a false failure.
    - Add a case where the label is absent from `implement.md` entirely (mirroring Case 8's shape
      for `VERBATIM REPORT — THE FACT`), asserting the guard reports the missing block at that
      site.
    - Add one case per shared phrase (three cases, mirroring Cases 9–11) where the block is
      present but missing that one phrase, asserting the specific missing-phrase message.
    - Update the file's own header comment describing what each case covers, per its existing
      convention.

  - [ ] **Step 5: Verify**

```bash verified:authored in-tree for this change
scripts/check-dispatch-paragraphs.sh
scripts/test-check-dispatch-paragraphs.sh
```

    Both must exit `0`.
    <!-- predicted: scripts/check-dispatch-paragraphs.sh && scripts/test-check-dispatch-paragraphs.sh after task 1 -->
    Also run this project's other configured markdown/guard lint commands
    (`scripts/check-vocabulary.sh`, `scripts/check-references.sh`, `scripts/check-plan-shape.sh`,
    `scripts/check-guard-symlinks.sh`, `scripts/check-markdown-integrity.py`,
    `scripts/check-contract-budget.sh`) and fix any hit.
