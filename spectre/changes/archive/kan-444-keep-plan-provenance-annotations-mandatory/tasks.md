# kan-444-keep-plan-provenance-annotations-mandatory

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Two wiring tasks in dependency order: ship the guard, then point the writing-plans stage at it
unconditionally. `design.md` is canonical for every decision.

**Baseline, measured before any edit:**

- The guard is not shipped: `skills/flow/scripts/` carries neither
  `check-plan-provenance.sh` nor `check-plan-provenance.py`.
  <!-- measured: ls skills/flow/scripts/check-plan-provenance.sh skills/flow/scripts/check-plan-provenance.py @ branch spectre/kan-444-keep-plan-provenance-annotations-mandatory -->
- `check-guard-symlinks.sh` rule 3 exempts the project-configured guards by name,
  `check-plan-provenance.sh` among them.
  <!-- measured: grep -n 'EXEMPT\[' scripts/check-guard-symlinks.sh @ branch spectre/kan-444-keep-plan-provenance-annotations-mandatory -->
- `scripts/test-check-guard-symlinks.sh` passes with 82 ok cases.
  <!-- measured: bash scripts/test-check-guard-symlinks.sh | grep -c '^ok:' @ branch spectre/kan-444-keep-plan-provenance-annotations-mandatory -->
- `scripts/check-contract-budget.sh`, `scripts/check-references.sh`,
  `scripts/check-dispatch-paragraphs.sh`, `scripts/check-installed-citations.sh` and
  `scripts/check-plan-provenance.sh` all exit 0 on this worktree.
  <!-- measured: the five guards run from this worktree before any edit @ branch spectre/kan-444-keep-plan-provenance-annotations-mandatory -->

**Every task that grows an owned `.md` file runs `scripts/check-contract-budget.sh` and
reconciles its `budgets()` row in the same commit** — raise a row only when the guard actually
fails on that file, and only to the size its own rule gives.

---

- [x] 1. Ship the plan-provenance guard with the flow skill

**Build:** green

  - [x] **Step 1: Symlink both guard files into the skill's scripts directory**

```bash verified:the relative-target shape read from skills/flow/scripts/check-plan-shape.sh's existing symlinks
ln -s ../../../scripts/check-plan-provenance.sh skills/flow/scripts/check-plan-provenance.sh
ln -s ../../../scripts/check-plan-provenance.py skills/flow/scripts/check-plan-provenance.py
```

  - [x] **Step 2: Drop the rule 3 exemption**

In `scripts/check-guard-symlinks.sh`, delete the `EXEMPT["check-plan-provenance.sh"] = 1` line
and update the comment above the map: the project-configured family it names shrinks by the
provenance guard, which ships with this change.

  - [x] **Step 3: Run the symlink guard and its harness**

```bash unverified:expect exit 0 only once Step 1's links exist and Step 2's exemption is gone
scripts/check-guard-symlinks.sh && scripts/test-check-guard-symlinks.sh
```

Expected: the guard exits 0 — rule 2 sees the two new links as legitimate shipped guards and
rule 3 sees no repository-relative invocation of the name in skill prose — and the harness still
prints its pass verdict.

  - [x] **Step 4: Commit**

```bash verified:subject matches the task's declared Commit field below
git add skills/flow/scripts/check-plan-provenance.sh skills/flow/scripts/check-plan-provenance.py scripts/check-guard-symlinks.sh
git commit -m "feat(scripts): ship the plan-provenance guard with the flow skill"
```

**Files:** `skills/flow/scripts/check-plan-provenance.sh`,
`skills/flow/scripts/check-plan-provenance.py`, `scripts/check-guard-symlinks.sh`,
`scripts/test-check-guard-symlinks.sh`
**Tests:** `scripts/test-check-guard-symlinks.sh` — the kan-444 rule 3 case: a
repository-relative invocation of the shipped provenance guard is reported
**Regression:** reverting the commit un-ships the guard — the symlinks vanish with it and rule 3
re-exempts the name, so a project that never declares the guard never rejects an untagged plan,
and the kan-444 harness case fails.
**Baseline:** before=82 after=84 ok cases in `scripts/test-check-guard-symlinks.sh`; the kan-444
case adds the two assertions closing panel finding F1.
<!-- measured: bash scripts/test-check-guard-symlinks.sh | grep -c '^ok:' @ branch spectre/kan-444-keep-plan-provenance-annotations-mandatory -->
**Commit:** `feat(scripts): ship the plan-provenance guard with the flow skill`

- [x] 2. Enforce the guard unconditionally at the writing-plans stage

**Build:** green

  - [x] **Step 1: Rewire the stage sentence**

In `skills/flow/brainstorm-planner.md` section D, replace the guard sentence so the stage runs
`check-plan-shape.sh` and `check-plan-provenance.sh` — shipped guards, run unconditionally — and
the project's configured build-green guard, if the project declares one; keep "and fix any hit".
Do not restate what the guard enforces — `skills/flow-contracts/plan-provenance-guard.md` is
canonical for that.

  - [x] **Step 2: Add the guard to the presence list**

In `skills/flow/SKILL.md`'s guard-presence paragraph, insert `check-plan-provenance.sh` into the
union list after `check-plan-shape.sh`. The `.py` needs no entry — a guard's sibling dependencies
are derived by grepping its source for `$SCRIPT_DIR/<name>`.

  - [x] **Step 3: Run the owned-prose guards**

```bash verified:all four exited 0 on this worktree before any edit, measured in this plan's baseline
scripts/check-contract-budget.sh && scripts/check-references.sh \
  && scripts/check-dispatch-paragraphs.sh && scripts/check-installed-citations.sh
```

On a budget failure, raise only the failing file's `budgets()` row, only to the size its rule
gives.

  - [x] **Step 4: Run the provenance guard over this change's own plan**

```bash verified:exit 0 on this worktree at plan time — the shipped basename form takes over once the change merges
scripts/check-plan-provenance.sh
```

During this run the shipped basename resolves against the installed skill directory, which still
tracks the main checkout; the repository-root invocation is the same script and is the by-hand
fallback the guard-presence rule names.

  - [x] **Step 5: Commit**

```bash verified:subject matches the task's declared Commit field below
git add skills/flow/brainstorm-planner.md skills/flow/SKILL.md
git commit -m "feat(flow): reject untagged plans at the writing-plans stage in every project"
```

**After:** Task 1
**Files:** `skills/flow/brainstorm-planner.md`, `skills/flow/SKILL.md`, `.flow/project.md`
**Tests:** none — no test cases are added; the stage wiring is guarded by the owned-prose checks
in Step 3 and the harness run in the first task
**Regression:** reverting leaves section D naming only a project-configured guard, so a project
that never declares it accepts untagged plans again, and the presence list stops naming what the
stage invokes.
**Baseline:** before=56 after=56 guard harnesses under `scripts/test-*.sh`; none added.
<!-- measured: ls scripts/test-*.sh | wc -l @ branch spectre/kan-444-keep-plan-provenance-annotations-mandatory -->
**Commit:** `feat(flow): reject untagged plans at the writing-plans stage in every project`
