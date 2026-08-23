# Cut `pipeline.md`'s load cost — split by consumer, move rationale to appendices

## Why

`skills/myflow-contracts/pipeline.md` is **50,290 bytes**, and `rules/myflow-manual-review.mdc`
requires every `/myflow-*` command to load it in full before any other step.
<!-- measured: wc -c skills/myflow-contracts/pipeline.md @ 15f11cc -->
It is the largest fixed context cost in the pipeline, paid identically by `/myflow-start`,
`/myflow-do`, `/myflow-finish`, `/myflow-fast` and `/myflow-status` — including commands needing a
fraction of it.

Surfaced while running `/myflow-start` for KAN-242: ~140k tokens were spent before the first design
question, and the mandated load was the largest single contributor.

Five of its sections are unreachable from at least two of the five commands, and together weigh
**19,793 bytes** — 39% of the file.
<!-- measured: awk section-size pass over skills/myflow-contracts/pipeline.md @ 15f11cc -->

The consumer split this needs is not a new mechanism. It is the same one that already produced
`state-file.md`, `jira-integration.md`, `finish-contract.md`, `jira-followups.md` and
`handoff-blocks.md` — established, and unfinished.

## What changes

- **Five sections move out of `pipeline.md`**, each into its own file loaded only by the commands
  that need it: `git-boundaries.md`, `model-policy.md`, `artifacts-registry.md`,
  `session-records.md`, `worktree-resolution.md`.
- **Five rationale appendices**, one per new core, with the matching sections moved out of
  `pipeline-rationale.md` so every heading tree stays mirrored.
- **A rationale sweep** across `pipeline.md` and all five new cores: inline argument prose moves to
  the appendix, leaving the normative sentence and a pointer.
- **Load declarations** in each command's own `SKILL.md`, at the step needing the contract. The
  always-on rule layer is untouched.
- **57 citations repointed** across `skills/`, `rules/`, `commands/`, `commands-claude/`,
  `CLAUDE.md`, `AGENTS.md`, `README.md` and `openspec/specs/`.
- **A spec requirement** recording why a section not reachable from every command lives in its own
  file, generalising the single-command rule `myflow-contract-economy` already carries.
- **Budget rows** for ten new files, with `pipeline.md`'s and `pipeline-rationale.md`'s **lowered**
  to their new sizes.

## Impact

- Affected specs: `myflow-contract-economy`, `myflow-contract-distribution` (MODIFIED — panel finding F20: `pipeline.md`'s canonical-authority requirement still claimed git boundaries after this change moved it out)
- Affected code: `skills/myflow-contracts/` (2 files split into 12), the five `/myflow-*`
  `SKILL.md` files, `scripts/check-contract-budget.sh`, and every file citing a moved section
  <!-- predicted: five cores plus five appendices added to pipeline.md and pipeline-rationale.md -->
- `setup.sh` needs **no** edit — `install_skills` symlinks whole skill directories, so the new files
  reach `~/.claude/skills/`, `~/.cursor/skills/` and `~/.codex/skills/` on the next run
- `setup.sh global` must be re-run so installed harnesses resolve the new paths

## Non-goals

- Changing what any contract says. This is a relocation and a split; a behavioural change riding
  along would be invisible in a diff this size.
- Evicting `Stage marks` or `Change name resolution` from the core — both are reachable from every
  producing command.
- A permanent no-loss guard. The sentence-set check this change runs is a throwaway: a future
  relocation would have no baseline for it to diff against.
- Repairing the pre-existing `Rendering` / `Preserving the session records` heading mismatch, which
  `pipeline.md` deliberately states and which travels with the section unchanged.
