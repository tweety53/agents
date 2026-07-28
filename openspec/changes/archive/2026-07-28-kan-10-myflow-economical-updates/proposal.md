## Why

myflow's token cost is dominated by fixed overhead paid regardless of change size: a 58137-byte (~57 KB)
always-on rule file inlined into every session in every project on the machine (~150k input tokens
per change across the twelve-stage pipeline, plus its cost on every unrelated session), seven
commands that load a 6 KB skill to set two JSON fields, and a review panel slot running on the
parent model where an economy agent would do.

Cutting fixed overhead is worth more than cutting any single stage, because it is charged on every
session whether or not myflow is in use.

## What Changes

- Split `rules/myflow-manual-review.mdc` (58137 bytes, `alwaysApply: true`): **State file**, **State
  self-heal**, **Project configuration**, and **Jira integration** move to a new on-demand
  `myflow-contracts` skill, one file per section. Each vacated section leaves a stub naming what
  it governs and which file to load. Always-on layer drops from 58137 bytes (~57 KB) to ~32 KB.
- Add `skills/myflow-state-advance/state-advance.sh`: the seven pure-state-write commands run it
  first and load the skill only when it escalates via a distinct exit code.
- Move review panel slot 4 (Adversarial) from the parent model to the economy tier; slots 0 and 2
  stay on the parent.
- Initialize OpenSpec in this repo and add `.myflow/project.md` so myflow reads this project's
  real test, lint, and standards commands instead of another project's.
- Add `scripts/check-references.sh`, a third guard that fails when a cross-referenced section no
  longer exists in the file it is referenced from. `/myflow-review` runs it alongside the other two.
- **BREAKING (behavioral, internal):** on the happy path the seven `*-done` / `*-manual-review`
  commands no longer run the artifact-contradiction checks from State self-heal — only "state file
  parses" and "worktree still exists". Full self-heal still runs whenever the script escalates.

## Capabilities

### New Capabilities

- `myflow-contract-distribution`: where myflow's pipeline contracts live, which of them are
  always-on versus loaded on demand, how each harness installs and resolves them, and the stub
  requirement that keeps an extracted contract discoverable.
- `myflow-state-advance`: the pure state-write path — the mechanical fast path, the escalation
  boundary between script and skill, and the invariants (monotonic gates, carry-forward, no Jira)
  that hold on both sides of it.
- `myflow-review-panel-economics`: which model tier each review panel slot runs on and why.
- `agents-repo-verification`: this repository's own myflow project configuration and its guard
  scripts, including the reference guard.

### Modified Capabilities

<!-- None. OpenSpec is being initialized in this repository by this change, so there are no
     existing specs to modify. -->

## Impact

- `rules/myflow-manual-review.mdc` — four sections removed, four stubs added.
- `skills/myflow-contracts/` — new (SKILL.md + four contract files).
- `skills/myflow-state-advance/` — new `state-advance.sh`; SKILL.md gains the escalation contract.
- `commands/` and `commands-claude/` — the seven `*-done` / `*-manual-review` command files in both
  trees invoke the script first.
- `skills/openspec-apply-superpowers/SKILL.md` — slot 4 model tier and the mapping's scope.
- `skills/openspec-apply-fix-superpowers/SKILL.md`, `skills/openspec-fast-path-superpowers/SKILL.md`
  — prose naming the slot range.
- Every skill, command, and root doc carrying a cross-reference to one of the four moved sections.
- `scripts/check-references.sh` — new; `scripts/test-setup.sh` — additive assertions for the new
  skill directory.
- `.myflow/project.md`, `openspec/` — new.
- No runtime code, no public API, no dependencies. Existing installs pick the change up on their
  next `setup.sh global`, which rewrites the managed block in place.
