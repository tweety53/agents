# kan-432-flow-delete-per-task-reviewer

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

**Spec:** `design.md` beside this file. One task, `Build: green`, prose and comments only — no
script logic, Go or store change. Every quoted anchor below was read from `main` at `2a5c000`; an
anchor that has moved is found by its quoted text, never by its line number.
`docs/superpowers/research/` and this change's own directory are never in the task commit.

- [x] 1. Delete the per-task reviewer from `skills/flow/implement.md` §4 and correct every sentence and comment that named it

**Build:** green
**Files:** `skills/flow/implement.md`, `skills/README.md`, `skills/flow/brainstorm-planner.md`, `skills/flow/review-panel.md`, `skills/flow-contracts/finish-contract-run1.md`, `scripts/check-dispatch-paragraphs.sh`, `scripts/test-check-dispatch-paragraphs.sh`, `scripts/check-panel-docs-only.sh`
**Tests:** Case 1 of `scripts/test-check-dispatch-paragraphs.sh` — its header comment is the one line this task edits there; no test is added or changed in behaviour
**Regression:** Case 1 — none: the edit is a comment, so reverting this commit changes no verdict; `scripts/check-dispatch-paragraphs.sh`, `scripts/test-check-dispatch-paragraphs.sh`, `scripts/check-stage-mark-calls.sh`, `scripts/check-contract-budget.sh` and `scripts/check-vocabulary.sh` pass before and after
**Baseline:** before=0 after=0
<!-- measured: no test suite covers skills/ prose or script comments; the count is the number of tests this task touches @ 2a5c000 -->
<!-- predicted: unchanged after this task, on branch spectre/kan-432-flow-delete-per-task-reviewer -->
**Commit:** `docs(flow): delete the per-task reviewer`

  - [x] **Step 1: The step table, the COMMIT-PER-TASK blockquote and the in-flight sentence in
    `skills/flow/implement.md`.** Replace the row opening `| **6** | **superpowers:requesting-code-review** + the review panel |` with:

````markdown verified:authored in-tree for this change
| **6** | the review panel | The final whole-branch panel |
````

    In the **FLOW — COMMIT-PER-TASK** blockquote, replace `before the parent dispatches review
    for it` with `before the guard runs on it`. Replace the paragraph `**Never end a turn with a
    child in flight** — wait for every implementer and reviewer launched` … `anything.` with:

````markdown verified:authored in-tree for this change
**Never end a turn with a child in flight** — wait for every implementer launched before
reporting a stage boundary or asking the operator anything.
````

  - [x] **Step 2: The task boundary in `skills/flow/implement.md` §4.** Replace everything from
    the paragraph opening `**Review overlaps the next implementer.**` through the paragraph ending
    `clean and any fix folded.` — the four-step list, **When the script cannot be located**,
    **Per-task review** and **The last bundle's reviewers run alone** — with:

````markdown verified:authored in-tree for this change; the guard's argument shape is unchanged from main at 2a5c000
**The next implementer overlaps the guard.** The unit is the bundle `plan-dispatch-bundles.sh`
emits. At each boundary, in this order:

1. **Bundle N+1's implementer commits** and writes its report; the wait above ends.
2. **One Bash call: the implementer's `record dispatch end`, the guard on every commit whose sha
   is new, `flow tasks tick` for every task the guard passed, and bundle N+2's gather.** The
   guard takes the canonical worktree's absolute path (the worktree created or resumed in step
   **2** above) as its fifth argument and this run's resolved `<name>` as its sixth:

   ```bash
   check-task-commit-fields.sh <worktree> <task-id> <task-sha> <task-base> <canonical-worktree> <name>
   ```

   The guard reads git objects and `tasks.md` only, so it is safe while the tree changes. Every
   verdict is printed and read before anything launches: a nonzero exit sends that task back to
   the **same implementer**, which re-commits and re-runs the guard before anything below; exit 0
   ticks the task in the same call.
3. **One message launches bundle N+2's implementer. The next Bash call records its `begin`** —
   the very next action after the launch returns, which is what "recorded immediately after the
   launch returns" above requires.

**When the script cannot be located**, apply `flow-task-commit-fields`'s rules by hand: check the
commit's `Files:` against `git diff --name-only <task-base>..<task-sha>`, its `Tests:` against the
commit's diff, and its `Commit:` against the commit's actual subject line.

**The guard's pass is the tick.** Mark a **task's** checkbox `[x]` (`flow tasks tick`) once
`check-task-commit-fields.sh` exits 0 on its commit — no reviewer runs per task; the whole-branch
panel (`skills/flow/review-panel.md`) is this branch's review. A step's checkbox tracks the step
and gates nothing. A red task's checkbox is ticked together with its partner's, on their one
commit's guard pass.

**The last bundle's guard pass is the stage's last boundary.** `final-review.diff` is written and
the slots dispatched once it has passed; the review panel's pre-work may share the last
implementer's wait, in its one call.
````

    The fenced `bash` block inside that replacement is the same guard call the deleted list
    carried, moved with its paragraph; `check-dispatch-paragraphs.sh`'s three blockquotes below
    that point are untouched.

  - [x] **Step 3: `skills/README.md` and `skills/flow/brainstorm-planner.md`.** In
    `skills/README.md`, replace the row `| **6** | requesting-code-review + the panel | `/flow`
    (implementation) |` with:

````markdown verified:authored in-tree for this change
| **6** | the review panel | `/flow` (implementation) |
````

    In `skills/flow/brainstorm-planner.md` **D**, in the fenced header block, replace the two
    lines `> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task`
    / `> passes spec + quality review.` with:

````markdown verified:authored in-tree for this change
> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
````

    This plan's own header keeps the old wording: it was written under the rule in force at
    `2a5c000`, and a plan is never rewritten to match a rule its change introduces.

  - [x] **Step 4: `skills/flow/review-panel.md` and `skills/flow-contracts/finish-contract-run1.md`.**
    In `review-panel.md`, replace the two lines `On a docs-only branch the whole-branch read covers
    the text every per-task reviewer already read` / `against the same plan; there is no code seam
    between commits for a second slot to find (KAN-312).` with:

````markdown verified:authored in-tree for this change; the justification is design.md's docs-only-gets-plan-alignment-only in kan-436
On a docs-only branch the implementer's self-review and the vocabulary and reference guards cover
the prose; there is no code seam between commits for a second slot to find (KAN-312).
````

    Replace `**This binds the review panel's fix round and not the per-task review's fix** in` /
    `` `skills/flow/implement.md`. `` with:

````markdown verified:authored in-tree for this change
**This binds the review panel's fix round.**
````

    In `finish-contract-run1.md`, replace `TDD per task, per-task review, the final review` with
    `TDD per task, the final review` on the line that carries it.

  - [x] **Step 5: The three script comments.** In `scripts/check-dispatch-paragraphs.sh`, replace
    `the implementer dispatch, the per-task reviewer dispatch, the review` / `# panel's slot
    dispatch, and the panel-fix subagent dispatch.` with the same two comment lines reading `the
    implementer dispatch, the conductor's own §4 instruction, the review` / `# panel's slot
    dispatch, and the panel-fix subagent dispatch.`, and `of implement.md (implementer dispatch,
    per-task reviewer dispatch) and` with `of implement.md (implementer dispatch, the conductor's
    own §4 instruction) and`. In `scripts/test-check-dispatch-paragraphs.sh`, replace `# dispatch,
    per-task reviewer dispatch).` with `# dispatch, the conductor's own §4 instruction).`. In
    `scripts/check-panel-docs-only.sh`, replace `# WHY THIS GUARD EXISTS (KAN-312): the
    whole-branch panel re-reads a` / `# docs-only branch every per-task reviewer already read.`
    with `# WHY THIS GUARD EXISTS (KAN-312): a docs-only branch has no code seam for a` / `#
    second slot to find; the implementer's self-review and the prose guards cover it.` Keep every
    comment line at or under 80 columns, as the surrounding header is.

  - [x] **Step 6: Nothing else names it.** Run, from the repository root:

````bash verified:ran at 2a5c000 during planning; the hits are the sites steps 1-5 replace, plus scripts/check-installed-citations.py, scripts/test-check-installed-citations.sh and docs/superpowers/ which are historical and left alone
grep -rn -e "per-task review" -e "per-task reviewer" -e "-role reviewer" -e "reviewers run alone" -e "re-reviewer" -e "spec + quality review" -e "spec \*\*and\*\* quality" skills scripts CLAUDE.md
````

    After steps 1–5 the only hits are `skills/flow/review-panel.md`'s panel-slot `-role reviewer`
    record (the panel's own; stays), `skills/flow/implement.md`'s `-role` value list (`reviewer`
    is still the panel's role; stays), this plan's own header, and the historical comments in
    `scripts/check-installed-citations.py` and `scripts/test-check-installed-citations.sh`, which
    record past findings by the reviewer that found them and are not statements about the
    pipeline. Any other hit is a site this plan missed: fix it in the same commit.

  - [x] **Step 7: The guards.** Run, from the repository root, and require exit 0 from each:

````bash verified:each is in .flow/project.md's ## lint list or its ## test list at 2a5c000
scripts/check-dispatch-paragraphs.sh
scripts/test-check-dispatch-paragraphs.sh
scripts/check-stage-mark-calls.sh
scripts/check-contract-budget.sh
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-markdown-integrity.py
scripts/check-installed-citations.sh
````

    `check-contract-budget.sh` cannot trip: every touched owned file shrinks or holds. Then
    commit with the declared subject and a `Task-Id: 1` trailer.
