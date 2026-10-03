# kan-829-flow-improvement-fix-pipeline-defects-at-the

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** a `/flow` or `/flow-fast` run fixes a non-`big` pipeline defect it hits within the run,
through a one-shot fix → review → fix → re-review loop on its own `<agents repo>` branch, landed by
the default route; self-review stops re-filing what was fixed.

**Architecture:** Task 1 adds the contract section and the two citations that let `/flow` and
`/flow-fast` dispatch under it, in one citation-closed commit. Task 2 teaches self-review's angle 1
to read the narrative line Task 1 defines.

**Spec:** `design.md` beside this file is canonical for every scope decision. Tasks cite its
`## Decisions` by ID and never restate them.

## Global Constraints

- Markdown only; no guard, CLI or SPA change.
- A commit's scope names the module it moved (**Commit scopes name the module**,
  `rules/commit-scope-is-the-module.mdc`).
- Every run-loaded sentence a task adds, deletes or rewords is listed, exactly as
  `scripts/check-verbatim-moves.sh` prints it after `::`, in
  `spectre/changes/kan-829-flow-improvement-fix-pipeline-defects-at-the/verbatim-moves.txt`, under
  a first line `# Operator, 2026-10-03: KAN-829 — in-run pipeline fixes.` The file is edited in the
  worktree and **left uncommitted**: it is a planning path, carried by the next
  `chore(spectre): plan` commit (**Planning commits**, `skills/flow-contracts/git-boundaries.md`).
- Every task's verify step runs the Markdown lint lines: `scripts/check-references.sh`,
  `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh`,
  `scripts/check-verbatim-moves.sh`, `scripts/check-dispatch-paragraphs.sh`,
  `scripts/check-stage-mark-calls.sh`, `scripts/check-normative-inventory.sh` (output unchanged
  against the merge base).

## Review Focus

- **Reach.** `/flow-fast` loads no `pipeline.md`; its guardrail must cite the section, or a
  `/flow-fast` run is still forbidden the new dispatches.
- **The closed list.** `implement.md`'s "every Agent-tool dispatch" table must carry the new row,
  or the parent's self-check refuses the dispatch.
- **Main checkout.** The fix worktree must never be the main checkout or the change's worktree.

**Live verification:** none — prose only; no service, store or runtime state changes.

---

- [ ] 1. The in-run pipeline-fix contract section, reachable from `/flow` and `/flow-fast`

`design.md` §§ 1–4.

  - [ ] **Step 1: Add the section** to `skills/flow-contracts/pipeline.md`, directly after
    `## Fewest operator actions` and before the first `## Wrong state for this command`, verbatim the
    block under **Task 1 — the section text** below.

  - [ ] **Step 2: Add the closed-list row** to `skills/flow/implement.md`'s
    **Dispatch sites — the parent's closed list**: change `These six rows are **every**` to
    `These seven rows are **every**`, and append the row under **Task 1 — the closed-list row**
    below.

    If the row count sentence reads differently once the board-swap change has landed on
    `origin/main`, keep its wording and change only the count.
  - [ ] **Step 3: Cite the section from `/flow-fast`.** In `skills/flow-fast/SKILL.md`'s
    **Guardrails, the whole list.** paragraph, change
    `Never dispatch a subagent the recorded decision does not name — an implementer per group on
    `sdd`, the decision's panel dispatches, the panel-fix subagent; never a planner or a verifier.`
    to end `…; never a planner or a verifier, save the in-run fix loop of **Pipeline defects found
    mid-run** (`skills/flow-contracts/pipeline.md`), which a `/flow-fast` run follows too.`
  - [ ] **Step 4: Verify** — `grep -c '^## Pipeline defects found mid-run$' skills/flow-contracts/pipeline.md`
    prints `1`; `grep -c 'Pipeline defects found mid-run' skills/flow/implement.md skills/flow-fast/SKILL.md`
    prints `1` for each. Run the Markdown lint lines; record every sentence
    `check-verbatim-moves.sh` flags in `verbatim-moves.txt`.
  - [ ] **Step 5: Commit.**

**Files:** `skills/flow-contracts/pipeline.md`, `skills/flow/implement.md`, `skills/flow-fast/SKILL.md`
**Tests:** none — contract prose
**Regression:** none — prose.
**Baseline:** before=0 after=0
<!-- measured: no test files in this task @ 5e86aef6 -->
**Commit:** `feat(flow): fix a pipeline defect within the run that finds it`
**After:** none
**Build:** green

**Decision:** rule-home-pipeline-contract
**Decision:** land-by-default-route
**Decision:** blast-radius-class
**Decision:** one-shot-fix-loop
**Decision:** narrative-line-record

### Task 1 — the section text

````markdown verified:authored in-tree for this change
## Pipeline defects found mid-run

**A `/flow` or `/flow-fast` run that hits a defect in the pipeline itself fixes it within the
run, unless the fix is `big`.** The pipeline is everything `<agents repo>` ships — a skill, a
contract, a guard, the `flow` CLI or flowd, a rule — and any saved memory that encodes its
behaviour. For this section alone `<agents repo>` is a repository the run was given
(**Reporting back**, `rules/agent-baseline.md`), whichever project the run is in.

**Predict the class before any work, from the blast radius — never from a plan.** Count the
files the fix must touch: the defective file, its test, and every file `grep -rl` finds in
`<agents repo>` citing the name or section the fix changes. The fix is `big` — `plan-class.sh`'s
own `files>=60` threshold — when that count is 60 or more, when it reaches outside
`<agents repo>`, or when it needs a design choice only the operator can make. A `big` fix is
deferred to self-review, and so is one found `big` once under way, which stops there.

**Every dispatch in the loop is one-shot.** No `SendMessage` to a finished child, and no child
parks a question for a resume. Each runs on `opus`, `subagent_type: flow-high`, its key per the
steps below and `<k>` counting this run's in-run fixes:

1. **Fix** — `pipeline-fix-<k>`, its prompt carrying the defect, its evidence and the counted
   blast radius. It works in its own worktree,
   `git -C <agents repo> worktree add -b fix/<slug> <agents repo>/.worktrees/<slug> origin/<default-branch>`
   — never the main checkout, never the change's own worktree — adds one test or guard that
   fails without the fix, updates a saved memory that encoded the defect, runs the
   `<agents repo>/.flow/project.md` `## lint` lines its files need, and commits with a module
   scope. It never merges or pushes.
2. **Review** — a fresh `pipeline-fix-<k>-review-<r>` over
   `git diff origin/<default-branch>...fix/<slug>`.
3. **Fix the findings** — the parent fixes them inline in that worktree, or dispatches a fresh
   `pipeline-fix-<k>-fix-<r>`; then step 2 again, `<r>` plus one. The loop runs to a clean
   review under **Fewest operator actions** above.
4. **Land** — by `<agents repo>`'s `## default landing route`
   (`project-get.sh <agents repo> 'default landing route'`), without asking. Merge and push is:
   rebase `fix/<slug>` onto `origin/<default-branch>`,
   `git push origin fix/<slug>:<default-branch>`, `git -C <agents repo> pull --ff-only`, then
   remove the worktree and the branch.

The loop runs in the background while the run's own work continues; the run still never ends a
turn with one of its children in flight.

**Record each one.** Every in-run fix, landed or deferred, is one line in the run's narrative —
`narrative.md` on `/flow` (**Write `IN_PROGRESS`**, `skills/flow/verify-and-handoff.md`),
`## Session narrative` on `/flow-fast` — and one decision bullet in the run's summary
(**Summary and live-stack line, before every handoff**, above):

```text
In-run pipeline fix: <agents sha | deferred> — <the defect, one line> (blast radius <N> files)
```
````

### Task 1 — the closed-list row

```markdown verified:authored in-tree for this change; table columns read from implement.md @ 5e86aef6
| in-run pipeline fix, review and fixer, one each per loop step | `pipeline-fix` | `pipeline-fix-<k>`, `pipeline-fix-<k>-review-<r>`, `pipeline-fix-<k>-fix-<r>` | `skills/flow-contracts/pipeline.md`, **Pipeline defects found mid-run** |
```

- [ ] 2. Self-review's angle 1 skips a defect fixed in-run

`design.md` § 4.

  - [ ] **Step 1: Add the sentence** to `skills/flow-self-review/SKILL.md`, as its own paragraph
    directly after the paragraph opening `Angle 5's remit covers`, verbatim:

    > An `In-run pipeline fix:` line in the bundle (**Pipeline defects found mid-run**,
    > `skills/flow-contracts/pipeline.md`) that names a sha is reported under angle 1 as fixed and
    > is never offered for filing; one that reads `deferred` is an angle-1 finding like any other.

  - [ ] **Step 2: Verify** — `grep -c 'In-run pipeline fix:' skills/flow-self-review/SKILL.md`
    prints `1`. Run the Markdown lint lines plus `scripts/check-self-review-report.sh`; record the
    new sentences in `verbatim-moves.txt`.
  - [ ] **Step 3: Commit.**

**Files:** `skills/flow-self-review/SKILL.md`
**Tests:** none — skill prose
**Regression:** none — prose.
**Baseline:** before=0 after=0
<!-- measured: no test files in this task @ 5e86aef6 -->
**Commit:** `feat(flow-self-review): report an in-run pipeline fix as fixed, never re-file it`
**After:** Task 1
**Build:** green

**Decision:** narrative-line-record
