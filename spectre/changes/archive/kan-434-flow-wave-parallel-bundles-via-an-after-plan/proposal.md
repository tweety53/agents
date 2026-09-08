# kan-434-flow-wave-parallel-bundles-via-an-after-plan

## Why

`implement.md` §4's "at most one implementer subagent in flight against a given worktree" rule
serializes every dispatch bundle, and file-disjoint bundling does not — the bundling rule only
excludes conflicting bundles, it does not require running the rest one after another. A quarter of
this repository's recent changes are now 6–9 file-disjoint bundles (kan-372, kan-374, kan-377,
kan-378, kan-379) paying that serialization for no reason the bundling rule imposes. This is
wall-clock lever 1 of `docs/superpowers/research/flow-speedup.md` (section 5, reopened and decided
in section 8): a gain of one to five implementer-lengths (3–15 min) on the 6–9-bundle plans, and
nothing on a 1–2-task change until the planner opts a task in.

## What changes

A task may declare its predecessors with an `**After:** Task <ids>` field (or `**After:** none`),
the same line-scoped shape as `**Squash-with:**`. `plan-dispatch-bundles.py` resolves each bundle's
`after` set (absent field = every earlier task; the default is opt-out, so every existing plan and
every unannotated task stays fully serial) and prints it as an `after <k>:` line beside the
unchanged `bundle <k>:` line. `implement.md` §4 then dispatches concurrently-ready bundles
together in waves: a bundle alone in its wave runs in the canonical worktree exactly as today; a
shared wave dispatches one implementer per bundle into its own throwaway worktree, cherry-picks
each onto the change branch in plan order as it returns, runs `check-task-commit-fields.sh` on each
picked commit, and hands a conflicted or guard-failing bundle back to its own implementer on the
rebased tree. When the plan's last task belongs to a shared wave, the conductor runs the resolved
`## test` list once on the canonical worktree after that wave's final pick passes.

Touched: `scripts/lib/plan_grammar.py` (new field regexes beside `SQUASH_WITH_VALUE_RE`),
`scripts/plan-dispatch-bundles.py`, `scripts/check-plan-shape.py` (new findings F7–F10) with
`scripts/check-task-commit-fields.py`'s field vocabulary, `skills/flow/implement.md` §4,
`skills/flow/brainstorm-planner.md` (plan template), and the guards' harnesses.
