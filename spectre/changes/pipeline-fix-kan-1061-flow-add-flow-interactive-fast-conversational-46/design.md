## Context

Found in run kan-1061: the `Correction:` transcription step of `skills/flow/implement.md` invokes
`check-plan-provenance.sh`, which no skill ships. Reproduced in this worktree at `f3cf9bd9`:

- With the reworded sentence and no symlink, `scripts/check-guard-symlinks.sh` exits 1 with
  `implement.md:900: invokes check-plan-provenance.sh here, but skill "flow" carries no symlink …
  (rule 2)`; with the symlink it exits 0. With the original wording it exits 0 either way, and
  with the symlink but the original wording it exits 1 on rule 6 (symlink required by nothing) —
  the rewording and the symlink land together.
- Through the symlink, the shim at `f3cf9bd9` exits 2 with `spectre/changes not found under
  …/skills/flow — refusing to report a clean run`; with `root="$(flow_guard_root)"` it prints
  `check-plan-provenance: 7 file(s) scanned, all provenance stated`.
- The guard takes no argument (it scans the checkout), so the sentence names none.
- `skills/flow/brainstorm-planner.md:373` names `check-plan-provenance` only as prose about the
  project-configured guard's exit codes; it invokes nothing and stays.

Blast radius: the three files of Task 1.
Nothing to repair: the failed call wrote nothing.

## Decisions

### Fix shape

**ID:** ship-provenance-symlink
**Status:** active
**Chosen:** add the symlink, reword the sentence into a rule-2 invocation, and make the shim resolve its root through `flow_guard_root` so the symlinked call works. The regression check is `check-guard-symlinks.sh` on the tree itself (`TestCheckGuardSymlinks`' own-tree case), which fails if either the symlink or the invoking wording is dropped.
**Considered:** extending rule 2 to read "`x.sh` … run" — widens a classifier other prose relies on, for one sentence; rewording to "the project's configured plan-provenance guard" with no symlink — leaves the parent without a runnable path in this repository, where the call is the point.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — one symlink, one sentence and a one-line root derivation reusing an existing helper.
**Considered:** `flow-medium` — the shim feeds the lint gate, but the change mirrors `check-plan-shape.sh` and the full guard package passes with it.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — the plan holds one task.
**Considered:** none — a single task has no split.

## Open questions

None.
