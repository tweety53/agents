## Why

myflow's twelve stages encode two different facts in one field — how far the work got, and who
is waiting on it. That conflation is what forces seven commands whose only job is to say "a
human looked at this", an `originStage` field to remember where a fix was raised, and a
monotonic-gates contract to stop those recorded confirmations from being demoted.

The cost is paid three times over: eighteen commands duplicated across two harness trees,
nineteen skills whose names do not match the commands that load them, and a vocabulary guard
that has to be taught every retired spelling because a rename touches so many layers. The
operator pays too — they have to know which of eighteen commands is legal at the current stage.

Collapse the model so the state says how far the work got and the human gate is a property of
that state. Three states, three pipeline commands, one human gate, and no command whose only
output is a state write.

Reviewing a staged diff and running the apps against a checklist are the same sitting at the
keyboard, so they become one gate rather than two stages. And integration is not a stage at all
— it is the first half of finishing, which is why `/myflow-finish` absorbs it.

## What Changes

- **BREAKING** Replace the twelve stages with three states: `STARTED`, `IN_PROGRESS`,
  `FINISHED`. Each pipeline command ends in the state named after it.
- **BREAKING** Reduce the command surface to five: `/myflow-start`, `/myflow-do`,
  `/myflow-finish`, `/myflow-status`, `/myflow-info`. Removed entirely: `/myflow-full`,
  `/myflow-fast-path`, `/myflow-manual-test`, `/myflow-review`, and the seven pure-state-write
  commands (`/myflow-start-fix` and `/myflow-start-done`, `/myflow-do-manual-review`,
  `/myflow-do-done`, `/myflow-do-fix` and its two, `/myflow-manual-test-done`,
  `/myflow-review-done`) together with the `myflow-state-advance` skill (which carries
  `state-advance.sh`) and `scripts/test-state-advance.sh`.
- **BREAKING** `/myflow-do` produces **both** the staged diff **and** the manual test guide, so
  reviewing the diff and running the apps become one human gate at `IN_PROGRESS` instead of two.
- **BREAKING** Make every command re-entrant: `/myflow-start` re-run revises the proposal,
  `/myflow-do` re-run applies a fix. A fix never moves the state, which removes `originStage`.
- **BREAKING** `/myflow-finish` becomes a two-run command. **Run 1** integrates: it asks up front
  how to land the branch — open a PR (default), merge to the base branch and push, or leave it to
  be handled manually — then stops, leaving the change at `IN_PROGRESS`. **Run 2**, once the
  branch is actually merged, does what finish does today plus the new work: sync delta specs,
  archive, **commit and push the archive**, and **remove the worktrees and branches**.
- **BREAKING** Remove the pre-integration verification gate entirely — no tests, linters, or
  spec-coverage check runs before a PR or merge. Correctness is established during `/myflow-do`
  (TDD per task, per-task review, and the final review panel) and by the human gate.
- **BREAKING** Remove **every** command flag — `automerge`, `skip-review`, `skip-manual-test`,
  `skip-propose`, `propose-only`, `checkpoint`, `commit-during-apply` and `full-panel`. No
  `/myflow-*` command accepts any flag; the only argument is the optional change name.
- **BREAKING** Shrink the state file to `state`, `branch`, `worktrees`, `artifactUrl`,
  `jiraIssue`, `prUrl`, `updatedAt`, `updatedBy`. Dropped: the whole `gates` object,
  `originStage`, `REVIEWED_TREE`, `fastPath`, and the separate `worktree` + `MERGE_BASE` pair
  (merged into one `worktrees` map keyed by absolute path).
- **Every review-panel reviewer runs on Sonnet.** The panel keeps its full roster and trigger
  table, but the per-provider economy-tier mapping is deleted — there is no longer a tier
  distinction to resolve.
- Add worktree cleanup to `/myflow-finish` run 2: four gating preflight checks and a disclosure of the ignored files `--force` will destroy, then
  `git worktree remove --force`, `git branch -d`, `git worktree prune`. Any failed check leaves
  everything alone and reports why.
- Add an optional `## stop` key to `.myflow/project.md` naming the command that stops a
  project's local stack. Absent means the check is skipped, not failed.
- Fix handoff output everywhere: absolute paths only, no bodies pasted into chat, and the next
  command as the bare last line with nothing after it.
- Exclude `openspec/` from the diff presented for review at `IN_PROGRESS`.
- **BREAKING** Restructure `skills/` to one skill per command, named after it — nineteen skills
  to seven. Delete the five `/opsx:*` commands that duplicate pipeline steps, keeping
  `/opsx:explore`.
- Rewrite `README.md` (with a mermaid transition graph), `myflow-info`, `myflow-status`,
  `CLAUDE.md`, `AGENTS.md` and `skills/README.md` against the three states; teach
  `scripts/check-vocabulary.sh` the retired state values and command names.

## Capabilities

### New Capabilities

- `myflow-state-machine`: The three states, the command→state transition table, re-entrancy and
  fix semantics, the state file shape, monotonicity and self-heal.
- `myflow-command-surface`: Which commands exist, which states each accepts, the git actions each
  may take, the wrong-state handoff, and what the removals guarantee.
- `myflow-finish-cleanup`: The two-run finish — the integration choice, merge verification,
  archive commit and push, and the worktree cleanup contract including the `## stop` project key.
- `myflow-handoff-output`: The fixed handoff shape — absolute paths, no pasted bodies, the bare
  next-command last line, and the `openspec/` exclusion from the review diff.

### Modified Capabilities

- `myflow-state-advance`: Removed in full — the script, its harness, the skill, and the commands
  that invoked it no longer exist.
- `agents-repo-verification`: The guards are no longer run by a review stage, because there is no
  review stage; they are the project's `## lint`/`## test` commands, run during `/myflow-do`.
  `scripts/test-state-advance.sh` is deleted with the script it asserted against.
- `myflow-contract-distribution`: `pipeline.md` is rewritten for three states, so the section
  names the distribution requirements enumerate change. The always-on stub mechanism itself is
  unchanged.
- `myflow-review-panel-economics`: Every panel slot runs on Sonnet; the per-provider economy-tier
  mapping and the parent-model inheritance rule are both removed.

## Impact

- `rules/myflow-manual-review.mdc` — frontmatter description and state diagram; the stub
  mechanism is untouched.
- `skills/myflow-contracts/pipeline.md` — rewritten. `state-file.md` and `state-self-heal.md` —
  rewritten for the new shape. `project-configuration.md` — gains `## stop`.
- `skills/` — three pipeline skills created by renaming and merging their predecessors; twelve
  skills deleted, leaving seven in total. `openspec-archive-change`, `openspec-sync-specs` and
  `openspec-propose` are delegated into by surviving skills, so their used content must be
  inlined **before** removal or `/myflow-finish` and `/myflow-start` break.
- `commands/` and `commands-claude/` — eighteen myflow command files each become five;
  `commands/opsx-*.md` drops from six files to one.
- `scripts/check-vocabulary.sh` — new retired literals. `scripts/test-state-advance.sh` —
  deleted; the `state-advance.sh` it asserted against goes with its skill directory.
- `README.md`, `CLAUDE.md`, `AGENTS.md`, `skills/README.md` — rewritten sections.
- State files already on disk under `/Users/tweety53/Agents/myflow/state/` use the old shape.
  No migration is written: every existing change is finished, and self-heal rewrites any file it
  cannot parse.
