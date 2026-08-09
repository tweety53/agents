## Why

The myflow contract/skill files are read in full on every `/myflow-*` run. KAN-95/82/87 already
built the partition machinery (`myflow-contract-economy`) and used it to slim four of the largest
files; it hasn't been applied to the rest, so `myflow-do`, `myflow-finish`, `myflow-start`,
`pipeline.md` and several contracts still carry restatement boilerplate and justification prose the
machinery already knows how to evict, plus several procedures are duplicated instead of cited. This
change applies the existing mechanism further and removes the duplication — no new partition
mechanism.

Two `/myflow-fast`-specific behavior changes are folded in, approved during this session: its setup
question round goes silent-default, and its design-approval confirm is dropped.

## What Changes

- Collapse the per-citation "this is canonical, never restate" boilerplate corpus-wide into one
  doctrine statement per file, moving the original paragraphs to each file's `-rationale.md` under
  the existing `myflow-contract-economy` rule-extraction carve-out.
- Move remaining justification prose (e.g. `myflow-start`'s `## Convergence`) into the already-
  existing `SKILL-rationale.md` siblings for `myflow-do`, `myflow-finish`, `myflow-start`; create
  `myflow-fast/SKILL-rationale.md`, which doesn't exist yet.
- New `skills/myflow-contracts/operator-prompts.md`: a shared shape for the five call sites that
  already independently implement "named options, one marked recommended, a safe silent default, a
  warning marker when it fires" — no behavior change, this consolidates prose already required by
  the existing `myflow-planning-gate` capability.
- Trim `myflow-do`'s mirrored copy of `check-unfinished-work.sh`'s marker-parse and exit-code rules
  down to the emit format only; the guard already reports its own rejections.
- Deduplicate three repeated procedures: the empty-worktree-set stop (cite `pipeline.md` instead of
  restating 3x), the two-commit chain (new `scripts/commit-split.sh`), and the `planningEffort`
  retired-key fallback (one description in `state-file.md`, cited elsewhere).
- New `scripts/prepare-workspace.sh` folds `myflow-do` §7's workspace-isolation guard + variable
  export into a script.
- Caveman-compress the report-only reviewer prompts (`principles-reviewer-prompt.md`,
  `adversarial-reviewer-prompt.md`, locate-and-report dispatches) — nothing else.
- **BREAKING** (to operator experience, not to any external interface): `/myflow-fast`'s creating-run
  setup question round (planning effort, 3 model roles, roster) no longer asks — it records the
  recommended defaults directly, still overridable by an explicit session instruction.
- **BREAKING**: `/myflow-fast` no longer stops for an explicit design-approval confirm after
  presenting the brainstormed design — it proceeds directly into artifact creation. Brainstorming's
  clarifying questions and design presentation remain fully interactive. `/myflow-start` is
  unaffected; this is scoped to `/myflow-fast` only.
- Deferred, not in this change: table-driving the oversized guard test suites
  (`test-check-plan-provenance.sh`, `test-check-cleanup-complete.sh`) — filed as a follow-up Jira
  issue at finish time.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `myflow-fast-command`: **Requirement: The first invocation chains brainstorming into
  implementation** drops its claim that brainstorming ends "with the operator approving the design
  ... exactly as `/myflow-start` runs it, with no auto-answering" — the confirm step itself is no
  longer part of what `/myflow-fast` runs, while the clarifying-question and design-presentation
  stages stay interactive and unchanged. **Requirement: Recorded defaults favor speed, still
  overridable** changes from "SHALL ask ... recommending sonnet/light" to "SHALL NOT ask; SHALL
  record the defaults directly," keeping the still-overridable-by-explicit-instruction guarantee.

## Impact

- Files: `skills/myflow-do/SKILL.md`, `skills/myflow-finish/SKILL.md`, `skills/myflow-start
  /SKILL.md`, `skills/myflow-fast/SKILL.md` (+ new `SKILL-rationale.md`), the three existing
  `SKILL-rationale.md` siblings, `skills/myflow-contracts/pipeline.md` and other contracts carrying
  the boilerplate, new `skills/myflow-contracts/operator-prompts.md`, new `scripts/commit-split.sh`
  and `scripts/prepare-workspace.sh` with their `test-*.sh` harnesses, `scripts/check-contract-
  budget.sh`'s table (re-anchored + two new rows), `.myflow/project.md` if it lists scripts
  individually.
- No change to the state file shape, the three-state machine, the review panel roster, or any
  `/myflow-start`/`/myflow-do`/`/myflow-finish`/`/myflow-status` behavior beyond `/myflow-fast`'s two
  named changes.
- Design: `docs/superpowers/specs/2026-08-09-kan-106-slim-myflow-skills-design.md`.
