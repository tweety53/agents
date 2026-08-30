# kan-300-myflow-reviewing-a-relocation-through-a-diff-is

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no — this plan adds new prose and a new script; it moves nothing between files.

**Goal:** when a plan declares itself a relocation, `/flow`'s review-panel stage generates a
before/after passage comparison across the plan's own scoped files, before the panel runs, and
hands every slot its path alongside `final-review.diff`.

**Spec:** `design.md` beside this file, and `openspec/specs/myflow-review-panel-economics/spec.md`.

## Global constraints

- **Generation never blocks the panel.** A script failure prints one line and the panel dispatches
  without the comparison file — the same fallback shape `dispatch-context.md`'s own missing-bundle
  case already uses.
- **Passage granularity, not sentence.** A passage is a paragraph, a bullet-list item, or a table
  row — the same unit `myflow-contract-economy`'s per-move ledger uses.
- **No new guard mechanically enforces the `**Relocation:**` header's presence.** It is stated as a
  writing-plans authoring requirement, the same "stated and deliberately not enforced" treatment
  this corpus already gives comparable judgment rules (e.g. `myflow-contract-economy`'s rationale
  appendix rule). A plan with no `**Relocation:**` line is read by the generation step as `no`.
- **`scripts/check-contract-budget.sh` stays green.** Both edited skill files
  (`skills/flow/brainstorm.md`, `skills/flow/review-panel.md`) have headroom in the declared budget
  table (26964 and 35985 bytes respectively, current sizes 22023 and 29920 bytes
  <!-- measured: wc -c skills/flow/brainstorm.md skills/flow/review-panel.md @ branch spectre/kan-300-myflow-reviewing-a-relocation-through-a-diff-is -->);
  no budget row needs raising unless a task's diff exceeds it, in which case that task raises the
  row for the file it touched.

---

- [x] 1. Declare and read the plan-level `**Relocation:**` header

**Build:** green

**Files:**
- Edit: `skills/flow/brainstorm.md` (§D, beside the existing `**Execution:**` header text)
- Edit: `openspec/specs/myflow-review-panel-economics/spec.md`

**Interfaces:**
- Consumes: nothing.
- Produces: the header line task 2's generation script reads.

  - [ ] **Step 1: Add the header instruction to writing-plans' output**

In `skills/flow/brainstorm.md`'s **D. Basic Workflow #3 — Writing plans** section, immediately
after the existing "Add this header to `tasks.md`" instruction and its `**Execution:**` example
block, add a second required line to the same header:

```markdown verified:this is the literal header text task 3 reads back
> **Relocation:** yes — <one-line reason>
```

or `**Relocation:** no`. State plainly that this line is required and explicit on every plan
(never omitted — this repository's "missing rather than dropped" convention), and that `yes`
scopes the comparison to the union of every task's own `**Files:**` field.

  - [ ] **Step 2: Add the spec requirement**

Add `### Requirement: A relocation-declaring plan gets a mechanical passage comparison` to
`openspec/specs/myflow-review-panel-economics/spec.md`, covering: the header field and its
required-explicit shape; that the comparison scopes from every task's `**Files:**` field; the
four-way passage classification (moved / repointed / added / removed) with repointed defined
identically to `myflow-contract-economy`'s two permitted moved-passage edits; that generation never
blocks the panel; and that every panel slot's dispatch prompt names the comparison file's path when
it exists. Include at least one scenario per clause, following this file's existing scenario style.

**Tests:** none — this task is prose only.

**Regression:** n/a — reverting removes the stated contract, which task 2's script would then have
no requirement to conform to.

**Baseline:** `scripts/check-contract-budget.sh` before=exit 0 after=exit 0
<!-- measured: scripts/check-contract-budget.sh @ branch spectre/kan-300-myflow-reviewing-a-relocation-through-a-diff-is -->;
`scripts/check-references.sh` before=exit 0 after=exit 0.

**Commit:** `docs(flow): declare the relocation header and its spec requirement`

---

- [x] 2. Add the relocation-comparison generation script

**Build:** green

**Files:**
- Create: `scripts/generate-relocation-comparison.py`
- Create: `scripts/generate-relocation-comparison.sh`
- Create: `scripts/test-generate-relocation-comparison.sh`
- Create: `skills/flow/scripts/generate-relocation-comparison.py` (symlink, `check-guard-symlinks.sh`
  rule 1)
- Create: `skills/flow/scripts/generate-relocation-comparison.sh` (symlink)
- Create: `skills/flow/scripts/test-generate-relocation-comparison.sh` (symlink)

**Interfaces:**
- Consumes: a worktree path, a merge-base ref, and a `tasks.md` path (to read `**Relocation:**`
  and every task's `**Files:**` field).
- Produces: `<worktree>/.superpowers/sdd/relocation-comparison.md` on a `yes` declaration with a
  non-empty scope; nothing otherwise.

  - [ ] **Step 1: Write `generate-relocation-comparison.py`**

Python 3, standard library only, following the `check-plan-shape.py`/`plan-dispatch-bundles.py`
precedent: import `parse_tasks` and the `**Files:**` field parsing from
`plan-dispatch-bundles.py` (via the same `importlib.util.spec_from_file_location` +
`sys.modules` registration pattern `check-plan-shape.py` already uses to import
`check-task-commit-fields.py`) rather than re-implementing the task/field grammar.

Algorithm:
1. Read `tasks.md`'s header for a `**Relocation:**` line (`^\s*>\s*\*\*Relocation:\*\*\s+(yes|no)\b`,
   scanned line by line). Anything other than an explicit `yes` (missing line, `no`, or a
   malformed value) exits 0 and writes nothing — this mirrors the accepted-limit style
   `plan-provenance.md`'s asymmetry rule already uses for absent tags.
2. On `yes`, union every task's `**Files:**` bullets across the whole plan into the file scope.
3. For each scoped file, extract passages — a run of non-blank lines between blank lines or `##`/`###`
   headings, or a single `- `/`* ` bullet, or a single Markdown table row — from the blob at
   `<merge-base>:<path>` (via `git show`) and from the file's current content on disk in the
   worktree.
4. Build the before-set and after-set of `(text, file, heading)` tuples. Classify every passage
   whose exact text appears in both sets but at a different `(file, heading)` as **moved**. Classify
   a before-only passage whose text matches an after-only passage's text modulo exactly the two
   edits `myflow-contract-economy` permits (a changed backticked path + bold section token, or a
   deleted trailing `above`/`below`) as **repointed**. Classify a remaining after-only passage as
   **added**, and a remaining before-only passage as **removed**.
5. Write the table to `<worktree>/.superpowers/sdd/relocation-comparison.md`: columns Passage
   (first eight words), Source, Destination, Class — one row per moved/repointed/added/removed
   passage. A passage unchanged in both text and location is not a row.

Exit codes: `0` written (including "declared no / empty scope, nothing written" — not an error);
`2` cannot answer — merge-base unreadable, a scoped file missing from the worktree, or `tasks.md`
unreadable.

  - [ ] **Step 2: Write the `.sh` wrapper**

`generate-relocation-comparison.sh <worktree> <changeRoot> <merge-base>` — thin wrapper, same
`python3` presence/liveness probe `plan-dispatch-bundles.sh` already runs, then
`exec python3 "$PYTHON_GUARD" "$@" "$changeRoot/tasks.md"`.

  - [ ] **Step 3: Write the test harness**

`test-generate-relocation-comparison.sh`: build a throwaway git repo fixture with a merge-base
commit and a worktree commit that (a) moves a paragraph verbatim between two files, (b) moves a
paragraph with a repointed citation, (c) adds a new paragraph, (d) removes a paragraph without
relocating it. Assert the four classes appear correctly in the output, and assert a plan declaring
`**Relocation:** no` produces no output file.

**Tests:** `skills/flow/scripts/test-generate-relocation-comparison.sh` (new) — the four
classification cases above, plus the `no`/absent-header no-op case.

**Regression:** reverting this task removes the only place these four classes are computed;
`test-generate-relocation-comparison.sh` fails with `No such file or directory`.

**Baseline:** new test file, no prior count to compare. `scripts/check-guard-symlinks.sh`
before=exit 0 after=exit 0.

**Commit:** `feat(flow): add the relocation-comparison generation script`

---

- [x] 3. Wire generation and dispatch into `skills/flow/review-panel.md`

**Build:** green

**Files:**
- Edit: `skills/flow/review-panel.md`

**Interfaces:**
- Consumes: task 2's script; task 1's header contract.
- Produces: the dispatch-prompt pointer every panel slot receives.

  - [ ] **Step 1: Call the generator before `final-review.diff`**

In `skills/flow/review-panel.md`, immediately before the existing `check-panel-diff-size.sh`
call, add:

```bash unverified:the wrapper's argument order matches task 2's own Step 2
generate-relocation-comparison.sh <worktree> <changeRoot> <merge-base>
```

State that a non-zero exit prints one line and the run continues without the comparison file —
generation is never a gate.

  - [ ] **Step 2: Add the dispatch-prompt pointer**

Beside the existing "Every slot's dispatch prompt also carries the CONTEXT BUNDLE paragraph"
sentence, add: when `<worktree>/.superpowers/sdd/relocation-comparison.md` exists, every slot's
dispatch prompt also names its absolute path, framed as a review input to audit against the diff —
never a substitute for reading `final-review.diff` itself.

**Tests:** none — this task is prose only; task 2's harness already covers the script's own
behaviour.

**Regression:** reverting this task leaves the script written but never invoked, so a relocation
plan's panel runs exactly as it did before this change.

**Baseline:** `scripts/check-contract-budget.sh` before=exit 0 after=exit 0
<!-- measured: scripts/check-contract-budget.sh @ branch spectre/kan-300-myflow-reviewing-a-relocation-through-a-diff-is -->;
`scripts/check-references.sh` before=exit 0 after=exit 0.

**Commit:** `docs(flow): wire the relocation comparison into review-panel dispatch`
