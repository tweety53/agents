# kan-513-flow-bias-free-implementer-grouping-toward

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Five tasks, dependency order. `spectre/changes/kan-513-flow-bias-free-implementer-grouping-toward/design.md`
is canonical for every decision a task implements; a task names the decision it implements and what
the edit must say, never a second copy of it.

**Baseline, measured before any edit, on `main` at `912c9fd`:**

- `scripts/run-guard-tests.sh` runs every `scripts/test-*.sh`; `scripts/test-plan-dispatch-bundles.sh`
  passes 45 cases. `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
  `scripts/check-guard-symlinks.sh`, `scripts/check-contract-budget.sh` and
  `scripts/check-markdown-integrity.py` exit 0.
  <!-- measured: scripts/test-plan-dispatch-bundles.sh | grep -c '^ok:' and each guard from the repository root @ main 912c9fd, 2026-09-11 -->
- Byte sizes against `scripts/check-contract-budget.sh`'s `budgets()` rows:
  `skills/flow/brainstorm-planner.md` 31,594 of 31,664; `skills/flow/implement.md` 41,021 of
  50,307; `skills/flow-contracts/pipeline.md` 25,296 of 36,155; `skills/flow/brainstorm.md`
  11,832 of 12,712. Task 2 grows `brainstorm-planner.md` past its row, so that task raises the row;
  task 3 may grow `brainstorm-planner.md` further (raise the row again if needed), `pipeline.md`
  and `brainstorm.md` have headroom but raise their rows too if the edit exceeds it; task 4's one
  clause fits inside `implement.md`'s headroom.
  <!-- measured: wc -c on each file and the budgets() rows at scripts/check-contract-budget.sh @ main 912c9fd -->

**Every prose task's verify step names the guards that scan owned Markdown**
(`check-vocabulary.sh`, `check-references.sh`, `check-markdown-integrity.py`,
`check-contract-budget.sh`) **plus `check-guard-symlinks.sh` where the edit invokes a script** —
never the project's whole `## lint` list.

---

- [x] 1. `scripts/plan-dispatch-groups.py` + wrapper + harness — the mechanical grouping

Per design.md section 2. A stdlib-only Python script that imports `check_file` from
`plan-dispatch-bundles.py` (same `sys.path` insertion of `lib/` for `plan_grammar`), computes
`deps(k) = A_k \ M_k` per bundle, runs the chain merge (a bundle joins the first group g with
`deps ∩ members(g) ≠ ∅` and `deps ⊆ members(g) ∪ ready(g)`, else opens a new group) then the fold
(groups with an identical `ready` set are folded down to two, alternating in plan order), and
prints `group <g>: <bundle ids>` one per line, numbered from 1 by lowest bundle id. Exit codes
mirror `plan-dispatch-bundles.py`: 0 printed (zero unchecked tasks prints nothing), 1 with the
bundle script's own violation lines passed through to stderr, 2 on a bad argument count or an
unreadable file. The `.sh` wrapper is `plan-dispatch-bundles.sh`'s one-argument form only
(python3 probe, then `exec`), with the module docstring and header comment stating the rule once
and citing design.md's worked examples rather than restating them. The harness is modelled on
`test-plan-dispatch-bundles.sh` (fixture per case under TMPDIR, `run_guard`, explicit pass/fail)
and covers at least: serial-default five bundles → one group; two independent chains → two groups;
four `after: none` singletons plus a join → `[1,3]`, `[2,4]`, `[5]`; the ceiling case `1, 2→1, 3→1,
4→2 3` → one group; a join point straddling two groups opens its own; a missing `**Files:**` field
passes the bundle script's exit 1 and message through; zero unchecked tasks prints nothing, exit 0;
wrong argument count exits 2.

  - [x] **Step 1: Write the failing harness** `scripts/test-plan-dispatch-groups.sh` with the cases
    above; run it and confirm every case fails against the absent script.
  - [x] **Step 2: Write** `scripts/plan-dispatch-groups.py` and `scripts/plan-dispatch-groups.sh`
    until the harness passes.
  - [x] **Step 3: Verify** — `scripts/test-plan-dispatch-groups.sh` passes every case;
    `scripts/test-plan-dispatch-bundles.sh` still passes 45; `scripts/check-guard-symlinks.sh`
    exits 0 (no skill invokes the script yet).

**Files:** `scripts/plan-dispatch-groups.py`, `scripts/plan-dispatch-groups.sh`,
`scripts/test-plan-dispatch-groups.sh`
**Tests:** `scripts/test-plan-dispatch-groups.sh`
**Regression:** reverting this commit removes the script and its harness; task 2's step 4 then
invokes a script that does not exist and `check-guard-symlinks.sh` rule 1 reports a dangling
symlink.
**Baseline:** before=0 after=8 — one new harness with eight cases
<!-- predicted: scripts/test-plan-dispatch-groups.sh after task 1 -->
**Commit:** `feat(scripts): add plan-dispatch-groups, the mechanical implementer grouping`
**Build:** green

- [x] 2. `skills/flow/brainstorm-planner.md` — Decide step 4 runs the script, split-only override

Per design.md section 3. Step 4 becomes: run `plan-dispatch-groups.sh <changeRoot>/tasks.md`
after `plan-dispatch-bundles.sh`; record its output as `groups_mechanical`; `groups` is that
result unless the planner **splits** a mechanical group with a one-line `groups_override` reason,
and never merges across it; `groups_override` is `null` when the mechanical grouping is taken
verbatim; `groups_reason` defaults to the literal `mechanical`. The decision JSON's field list
gains `groups_mechanical` and `groups_override` beside `groups` (both `null` when `execution` is
inline). The printed `## Decision` table's implementer-groups row shows the groups followed by
`— <groups_reason>` and, on an override, `(mechanical: <groups_mechanical>; override: <reason>)`,
mirroring the `class:` line. The two-in-flight cap sentence stays. Add the `.sh` and `.py` symlinks
under `skills/flow/scripts/` beside the `plan-dispatch-bundles` pair, and raise the file's
`budgets()` row.

  - [x] **Step 1: Rewrite step 4** and the decision JSON field list and printed row per the content
    above.
  - [x] **Step 2: Add the symlinks** `skills/flow/scripts/plan-dispatch-groups.sh` and
    `skills/flow/scripts/plan-dispatch-groups.py`, relative targets, as the bundles pair.
  - [x] **Step 3: Raise the budget row** for `skills/flow/brainstorm-planner.md` in
    `scripts/check-contract-budget.sh` to `wc -c` plus 25% headroom.
  - [x] **Step 4: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
    `scripts/check-guard-symlinks.sh`, `scripts/check-installed-citations.sh` all exit 0.

**Files:** `skills/flow/brainstorm-planner.md`, `skills/flow/scripts/plan-dispatch-groups.sh`,
`skills/flow/scripts/plan-dispatch-groups.py`, `scripts/check-contract-budget.sh`
**Tests:** none — Markdown and symlinks; the verify step's guard scripts are the check
**Regression:** reverting this commit puts step 4 back on "group freely", so the next sdd run's
grouping is the planner's instinct again and no `groups_mechanical` is recorded.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `docs(flow): decide implementer groups mechanically, split-only override`
**Build:** green

- [x] 3. Skip brainstorming's interactive checklist when the seed is fully complete

Per design.md's "A fully-seeded research note skips brainstorming's interactive checklist
entirely" decision. Three edits:

1. `skills/flow/brainstorm-planner.md`'s "Seed from a staged research note, if one exists"
   section: immediately after the existing "Seeding never skips the interactive round" paragraph,
   add the exception — when the found note carries all three of itself, a sibling
   `<stem>/tasks.md` and a sibling `<stem>/decision.json`, the interactive checklist and the merged
   convergence-and-approval confirm are both skipped: present the note's parsed structure, then
   continue directly into **C**. A note missing either sibling file still runs the full checklist.
2. `skills/flow-contracts/pipeline.md`'s "Stage exit — never the command's own judgment" section:
   add this as a second, narrowly-scoped bounded exception beside the existing "no channel to ask
   through" one, cross-referencing `brainstorm-planner.md`'s section by name rather than restating
   its mechanics.
3. `skills/flow/brainstorm.md`'s mark text under "Run brainstorming and planning directly": note
   that on the fully-seeded bypass, the `flow.brainstorm` end and `flow.design-approval`
   begin/end marks still fire, back-to-back with no interactive gap between them, to keep stage
   bookkeeping consistent with every other run.

  - [x] **Step 1: Edit** `skills/flow/brainstorm-planner.md` per point 1 above.
  - [x] **Step 2: Edit** `skills/flow-contracts/pipeline.md` per point 2 above.
  - [x] **Step 3: Edit** `skills/flow/brainstorm.md` per point 3 above.
  - [x] **Step 4: Raise budget rows** for any of the three files in
    `scripts/check-contract-budget.sh` whose edit exceeded its row (`wc -c` plus 25% headroom).
  - [x] **Step 5: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh` all exit 0.

**Files:** `skills/flow/brainstorm-planner.md`, `skills/flow-contracts/pipeline.md`,
`skills/flow/brainstorm.md`, `scripts/check-contract-budget.sh`
**Tests:** none — Markdown; the verify step's guard scripts are the check
**Regression:** reverting this commit restores "seeding never skips the interactive round" as the
only rule, so a fully-seeded note still runs the checklist and confirm every time.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `docs(flow): skip brainstorming's checklist when the seed is fully complete`
**Build:** green

- [x] 4. `skills/flow/implement.md` — a group's readiness excludes its own members

Per design.md section 4. In the **Waves** paragraph, the readiness sentence reads: a group is
ready when every id in the union of its bundles' `after <k>:` lines **that is not itself a task of
one of the group's bundles** has landed. One clause; nothing else in the paragraph moves.

  - [x] **Step 1: Edit** the readiness sentence in `skills/flow/implement.md`.
  - [x] **Step 2: Verify** — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-contract-budget.sh`,
    `scripts/check-normative-inventory.sh` all exit 0.

**Files:** `skills/flow/implement.md`
**Tests:** none — Markdown; the verify step's guard scripts are the check
**Regression:** reverting this commit restores the literal reading under which a chain-merged
group waits on its own members and is never ready.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `docs(flow): exclude a group's own members from its readiness set`
**Build:** green

- [x] 5. Delete the adopted staging note and its plan directory

`docs/superpowers/research/kan-513.md` seeded this change's brainstorm; its plan and decision
(`docs/superpowers/research/kan-513/`) are now this change's own `tasks.md` and decision record —
delete all of it.

  - [x] **Step 1: Delete** `docs/superpowers/research/kan-513.md` and the directory
    `docs/superpowers/research/kan-513/`.
  - [x] **Step 2: Verify** — `scripts/check-vocabulary.sh` and `scripts/check-references.sh` from
    the worktree root; both exit 0.

**Files:** `docs/superpowers/research/kan-513.md`,
`docs/superpowers/research/kan-513/tasks.md`,
`docs/superpowers/research/kan-513/decision.json`
**Tests:** none — deletion of an adopted staging note; the verify step's guard scripts are the check
**Regression:** reverting this commit leaves a stale note and plan in the research tree after their
content has landed in this change's own artifacts, which the "delete once adopted" rule
(`skills/flow/brainstorm-planner.md` section B) exists to prevent.
**Baseline:** before=0 after=0 — no test file changes
**Commit:** `docs(flow-research): delete the adopted kan-513 staging note`
**Build:** green
