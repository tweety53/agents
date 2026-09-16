# flow-plan-inits-spectre

## Why

`/flow-plan` captures a session as a staging note under `docs/research/`, with a plan and a
decision beside it, and `/flow <KEY>` later adopts the three by rewriting the note into
`proposal.md` and `design.md`, copying the plan and re-recording the decision, then deleting the
note. The round trip exists only because `/flow-plan` is forbidden to touch pipeline state. It
costs an LLM re-rendering of content the operator already approved, and it carries a seed lookup,
a partial-seed and fully-seeded exception, a delete-on-adopt step, a research worktree, a direct
push onto the default branch and a `docs/research/` pathspec in every planning-path exclusion.

## What changes

- `/flow-plan` ends at `STARTED`. A captured session runs `/flow`'s own cited sections — name
  and `STARTED` write, kickoff, artifact creation, writing-plans, Decide, the plan review gate —
  writes the change's own `proposal.md`, `design.md` and `tasks.md` in the change worktree,
  records the decision against the change, and commits and pushes the planning artifacts on
  `spectre/<name>`. It still marks only `plan.session`, closed `started`.
- `/flow <KEY>` resumes at `STARTED` through the existing resume logic and skips to implement.
  Section A looks the change up by key prefix before deriving a slug, and kickoff checks out an
  existing `origin/spectre/<name>` instead of branching fresh.
- The staging-note path is deleted everywhere: the seed section and its exceptions, the seeded
  startup block, the fully-seeded bypass, the delete-on-adopt step, the research worktree and its
  push, and every `docs/research/` mention and pathspec in skills, contracts, commands, README,
  `CLAUDE.md`, `AGENTS.md`, `commit-split.sh` and `recover-guard-incident.sh`.
- `docs/research/` and its five notes are deleted.
