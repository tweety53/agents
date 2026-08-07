# Manual test — kan-82-cut-myflow-per-command-token-overhead

This repository declares **no runnable application**: it is the source of the myflow skills, commands
and rules, installed elsewhere by `setup.sh`. "Running the apps" here means running its guard
scripts, its assertion harnesses, and a sandboxed installer pass.

Run everything from the apply worktree:

```bash
cd /Users/tweety53/Projects/agents-worktrees/kan-82-cut-myflow-per-command-token-overhead
```

## The guards

- [ ] check `scripts/check-contract-budget.sh` — reports `BUDGET-OK` and names how many files it checked
- [ ] check the budget guard fails a file over budget — append 200KB to `skills/myflow-contracts/build-green.md`, re-run, confirm it names that file and exits 1, then `git checkout --` it
- [ ] check the budget guard fails a contract with no budget row — `touch skills/myflow-contracts/scratch.md`, re-run, confirm it names that file, then delete it
- [ ] check the budget guard is self-scoped — run it by absolute path from `/tmp` and confirm it still reports 12 files, not a clean run over nothing
- [ ] check the budget guard refuses a symlink — `ln -s /etc/hosts skills/myflow-contracts/ghost.md`, re-run, confirm it names it as a symlink and never prints that file's size, then delete it
- [ ] check `scripts/check-references.sh` — every citation across the corpus still resolves
- [ ] check `scripts/check-vocabulary.sh` — both vocabulary guards clean
- [ ] check `scripts/check-plan-provenance.sh`, `scripts/check-task-build-green.sh`, `scripts/check-workspace-isolation.sh` — all exit 0

## The harnesses

- [ ] check `scripts/test-check-contract-budget.sh` — prints `all checks passed`, fourteen cases
- [ ] check every other harness in `.myflow/project.md`'s `## test` list still passes

## The split itself

- [ ] check nothing was lost from `pipeline.md` — the line-multiset diff against `adedf66` shows 14 removed lines, every one the pre-edit form of a repointed citation or of a line whose stale `above`/`below` was deleted
- [ ] check nothing was lost from `jira-integration.md` — the same diff against core + rationale + followups shows 13 removed lines, each accounted for the same way
- [ ] check `pipeline.md` and `pipeline-rationale.md` carry identical heading trees, fence-aware
- [ ] check `jira-integration.md` and `jira-integration-rationale.md` carry identical heading trees
- [ ] check no rule moved to an appendix — read `pipeline-rationale.md` and `jira-integration-rationale.md` and confirm nothing in them is something an agent must obey
- [ ] check `jira-followups.md` carries both its rules and its reasoning, and that no `jira-followups-rationale.md` exists

## What a command now loads

- [ ] check `/myflow-info` is still answerable from `pipeline.md` alone — the state diagram, the level-1 stage table and all eight level-2 expansions are in the core
- [ ] check `skills/myflow-contracts/SKILL.md` names every file in the directory, and the directory holds every file it names
- [ ] check no `/myflow-*` skill instructs an agent to load a `-rationale.md` file
- [ ] check `rules/myflow-manual-review.mdc` lists `jira-followups.md` among the narrower contracts and does **not** list either appendix

## Distribution

- [ ] check a sandboxed `HOME="$(mktemp -d)" ./setup.sh global` places all five contract files under `.claude/skills/`, `.cursor/skills/` and `.codex/skills/`
- [ ] check `git diff -- setup.sh` is empty — the new files ship with no installer change

## Run a real command against the split contracts

- [ ] check `/myflow-status` renders correctly reading the partitioned `pipeline.md`
- [ ] check `/myflow-info` prints the pipeline explanation from the partitioned `pipeline.md`

## Known incomplete

- **The linked issue's headline target is not met.** KAN-82 asked for `pipeline.md` to drop from
  roughly 34k tokens to roughly 8k. It went 103,326 → 88,253 bytes, a 15% cut, and the two files a
  `/myflow-start` run loads went 156,109 → 103,844 bytes, a 33% cut. The panel's fix rounds returned
  several passages to the core, so the final cut is smaller than the figure measured mid-run. The gap is a consequence of the
  `paragraph-granularity-partition` decision taken during implementation: sentence-level splitting
  turned out to be unexecutable against the line-multiset check that proves nothing was lost, so a
  paragraph mixing a rule with its justification stays in the core whole. Most of `pipeline.md`'s
  remaining prose is that shape.
- **About 35KB of bold-lead-in prose paragraphs remain in `pipeline.md`'s core.** Sampled by hand and
  found to be mostly rule-stating rather than pure rationale, so they correctly stay under the
  paragraph rule — but that sample was not exhaustive. A later pass could move a few more whole
  paragraphs — the preservation-ordering note among them. Not attempted here: the yield is a few
  thousand bytes and each one is a fresh judgment call. Both `**Which file to change first.**`
  paragraphs are deliberately **not** candidates: the review panel disputed moving one of them and the
  operator ruled both stay in the core.
- **`project-configuration.md` is not split.** At 45,258 bytes it is the next file to have this
  problem, and `/myflow-do` and `/myflow-finish` both load it. It is ratcheted where it stands by the
  new budget guard, so it cannot grow further unnoticed, but it was deliberately left out of scope.
- **The no-appendix-at-runtime rule is unenforced by design.** Nothing stops a future skill from
  naming a `-rationale.md` in a load instruction; the rule is stated in three places and guarded
  nowhere, which is the `runtime-rule-unenforced` decision.
