## Context

myflow is installed globally from this repository. `setup.sh`'s `always_on_rules()` (setup.sh:116)
selects every `rules/*.mdc` declaring `alwaysApply: true` and `render_managed_block` inlines each
one into `~/.claude/CLAUDE.md` and `~/.codex/AGENTS.md`; `install_rules_cursor` links the same
files into `~/.cursor/rules/`. `rules/myflow-manual-review.mdc` is 58137 bytes (~57 KB) and always-on, so all
three harnesses load it on every session, in every project, whether or not myflow is in use.

Roughly half that file is contract detail consulted only by skills that already load a `SKILL.md`
of their own: State file, State self-heal, Project configuration, Jira integration. The other half
— stage tables, boundary rules, per-stage sections — is what an agent must know without being
asked to fetch anything.

Two smaller costs sit alongside it. The seven pure-state-write commands load a 6 KB skill and
reason through it to set two JSON fields. And the review panel dispatches slot 4 (Adversarial) on
the parent model, though its work is the narrowed breadth pass the economic mapping already
justifies delegating.

The full brainstorming record is
`docs/superpowers/specs/2026-07-28-kan-10-myflow-economical-updates-design.md`.

## Goals / Non-Goals

**Goals:**

- Cut the always-on rule layer from 58137 bytes (~57 KB) to ~32 KB without losing any rule or its rationale.
- Make the pure-state-write path mechanical in the common case, preserving today's behavior on
  every uncommon one.
- Move one review panel slot to the economy tier.
- Give this repository the project configuration and the guard it needs to verify its own changes.

**Non-Goals:**

- Reducing the number of pipeline stages or gates, or changing what any gate verifies.
- Compressing rationale prose out of the always-on layer. It exists to stop agents rationalizing
  around the rules; removing it from context risks the failures it was written to prevent.
- Any change to runtime code, public APIs, or dependencies — this repository has none.

## Decisions

### Packaging of the extracted contract sections

**ID:** `contracts-packaging`
**Status:** active
**Chosen:** A new `myflow-contracts` skill with one file per section — `install_skills` already
symlinks every skill directory into all three harness skill directories, skills already resolve by
name in all three, and resolution has direct precedent in `[PRINCIPLES_PATH]` ("the absolute path
of `engineering-principles.md` in the skill directory you are reading this file from"). Per-file
granularity means the average consumer loads ~85 lines instead of ~385.
**Considered:** A `.mdc` under `rules/` with `alwaysApply: false` — by `always_on_rules()`'s
contract that marks it an opt-in *project standard*, never installed by any path and reaching a
project only when its `.myflow/project.md` names it, which renders it back into an always-loaded
project block. A plain `.md` under `rules/` — ignored by both installer globs, so nothing installs
it and consumers would need the agents-repo path, which varies by machine. Co-locating each section
beside its main consumer — State file has ~10 consumers, so it still needs one canonical home plus
cross-skill absolute paths, multiplying the problem rather than solving it.

### Depth of the rule-file split

**ID:** `split-depth`
**Status:** active
**Chosen:** Contract sections only, ~57 KB → ~32 KB. The four sections are read solely by skills
that already load a `SKILL.md`, which makes the extraction mechanical rather than a judgment call
about what an agent needs in context.
**Considered:** Also moving rationale prose, reaching ~15-18 KB — rejected because that prose is
the guard against misreading the rules it accompanies. A middle ground leaving one-line "why" stubs
(~20-22 KB) — rejected as carrying most of that risk for part of the saving.

### Division of labor between script and skill

**ID:** `state-advance-split`
**Status:** active
**Chosen:** The script handles the happy path and escalates with distinct exit codes (3 unreadable
state or stale worktree, 4 stage mismatch, 5 null `originStage`, 6 corrupt `originStage`). Every
behavior that needs judgment — name resolution across multiple candidates, the override prompt,
artifact-based self-heal — stays in the skill and runs unchanged when the script escalates.
**Considered:** Script-only with a hard fail — cheapest, but drops self-heal, automatic name
resolution, and the override prompt. Script also performing self-heal — keeps the rarest paths
cheap at the cost of substantially more shell logic to test, for paths that are rare by definition.

### Host repository for this change

**ID:** `host-repo`
**Status:** active
**Chosen:** Initialize OpenSpec here and add `.myflow/project.md`. The change touches zero files in
any other repository, and it makes `/myflow-review` run the verification that actually applies.
**Considered:** Following the KAN-7 precedent of hosting myflow changes in the gymie repository —
no new setup, but the change's lint, test, apps, and manual-test guide would all be drawn from a
project it does not touch; `/myflow-review` would run `./gradlew ktlintCheck detekt` against a
change containing no Kotlin.

### Guard for moved-section references

**ID:** `reference-guard`
**Status:** active
**Chosen:** A new `scripts/check-references.sh`. A stale cross-reference is this change's principal
failure mode, `check-vocabulary.sh` cannot catch it (it greps a fixed list of retired literals),
and commit `af76a3a` records the same drift class biting this repository before.
**Considered:** Relying on the review panel and a manual sweep — nothing would re-check it on later
changes. Extending `check-vocabulary.sh` — mixes two concerns in a script whose stated contract is
"grep for retired literals".

## Risks / Trade-offs

- **Stale cross-references after the move** → `scripts/check-references.sh`, run by
  `/myflow-review` as part of the verification gate.
- **Agents never load a contract they need** → each vacated section leaves a stub carrying its
  original heading, what the contract governs, and the file to load; consuming skills carry an
  explicit guardrail naming the file.
- **Narrowed self-heal hides real drift** → the script still escalates on an unreadable state file
  and on a `worktree` missing from `git worktree list`; `/myflow-review` and `/myflow-finish`
  continue to verify gate and merge state independently. The narrowing is stated in both the skill
  and `state-self-heal.md` rather than left implicit.
- **Authority split from the text it governs** → canonical-authority declarations move with their
  sections and are reworded to name the contracts file; stubs explicitly do not claim authority.
- **A guard that is too strict becomes noise** → `check-references.sh` requires only that *at least
  one* bold token on a referencing line resolves to a heading, which tolerates the phrasing
  variants in use while still failing when a section moves.

## Migration Plan

No migration step is required. Existing installs continue to work from their current managed block
until the next `setup.sh global`, which rewrites that block in place from the new rule file and
installs the new skill directory alongside the existing ones. Rollback is `git revert` plus a
re-run of `setup.sh global`.

## Open Questions

None. All five design decisions were settled during brainstorming and are recorded above.
