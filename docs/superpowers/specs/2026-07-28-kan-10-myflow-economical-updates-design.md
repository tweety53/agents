# KAN-10 — myflow economical updates

**Date:** 2026-07-28
**Jira:** [KAN-10](https://tweety53.atlassian.net/browse/KAN-10) — Myflow economical updates
**Repo:** `agents` (this repo)

## Problem

myflow's per-change token cost is dominated by fixed overhead that is paid regardless of how
large the change is.

1. **`rules/myflow-manual-review.mdc` is 58137 bytes (~57 KB) and `alwaysApply: true`.** `setup.sh`'s
   `always_on_rules()` inlines every always-on rule into `~/.claude/CLAUDE.md` and
   `~/.codex/AGENTS.md`, and `install_rules_cursor` links it into `~/.cursor/rules/`. All three
   harnesses therefore load the whole file on every session, in every project, whether or not
   myflow is being used. A change that walks the twelve-stage pipeline spans roughly ten
   sessions, so the rule file alone costs ~150k input tokens per change — plus its cost on every
   unrelated session on the machine.

2. **Seven commands are pure state writes that still load the full skill.** `/myflow-do-done`,
   `/myflow-do-manual-review`, `/myflow-do-fix-done`, `/myflow-do-fix-manual-review`,
   `/myflow-manual-test-done`, `/myflow-review-done`, and `/myflow-start-done` each load the 6 KB
   `myflow-state-advance` skill and reason through it in order to set two JSON fields.

3. **The review panel's Adversarial slot runs on the parent model.** Slots 0, 2, and 4 all
   inherit the parent; only slots 5+ use the economy tier.

## Non-goals

- Reducing the number of pipeline stages or gates.
- Changing what any gate verifies.
- Compressing the rule text that governs always-on judgment. Rationale prose stays where it is;
  it exists to stop agents rationalizing around the rules, and removing it from always-on context
  risks the failures it was written to prevent.

## Design

### 1. Split the rule file

`rules/myflow-manual-review.mdc` keeps every section that governs always-on judgment: Model
policy, Change name resolution, Pipeline stages, Stage transitions, Fix re-entry, IntelliJ
commands, Stage boundaries, the per-stage sections (Do through Finish), Full cycle gates, and
Opt-out.

Four sections move into a new `skills/myflow-contracts/`, one file each:

| Section moved | New file | Approx. lines |
|---------------|----------|---------------|
| State file | `state-file.md` | 85 |
| State self-heal | `state-self-heal.md` | 40 |
| Project configuration | `project-configuration.md` | 96 |
| Jira integration | `jira-integration.md` | 164 |

`SKILL.md` is a thin index naming the four files and when each is needed.

Every vacated section **leaves a stub in the rule file** — its original heading, one sentence
naming what it governs, and the exact contracts file to load. The stub is load-bearing: without
it, an agent reading always-on context has no way to know the contract exists, which is a worse
failure than the token cost being removed.

The canonical-authority sentences move with their sections and are reworded so the contracts file
is canonical ("This file is the canonical definition… if a skill and this file disagree, this
file wins"). Authority must not be split from the text it governs.

**Result:** the always-on layer drops from 58137 bytes (~57 KB) to ~32 KB; the average consuming skill loads
~85 lines rather than ~385.

#### Why a skill, not a rule or a companion file

- A new `.mdc` under `rules/` with `alwaysApply: false` is, by `always_on_rules()`'s contract,
  an **opt-in project standard** — never installed by any path, reaching a project only when its
  `.myflow/project.md` names it, which renders it back into an always-loaded project block. Wrong
  mechanism.
- A plain `.md` under `rules/` is ignored by both globs, so nothing installs it anywhere and
  consumers would need the agents-repo path, which varies by machine (`AGENTS_DATA`, and the repo
  is only ever symlinked from).
- `install_skills` already symlinks every skill directory into `~/.claude/skills`,
  `~/.cursor/skills`, and `~/.codex/skills` (setup.sh:595-604), and skills resolve by name in all
  three harnesses. Resolution has direct precedent: `[PRINCIPLES_PATH]` is already "the absolute
  path of `engineering-principles.md` in the skill directory you are reading this file from".
- Cost of the new skill: one line in the always-present skill index, ~30 tokens.

### 2. `state-advance.sh`

New `skills/myflow-state-advance/state-advance.sh`, symlinked with its skill directory:

```bash
state-advance.sh --name <name> --target <stage|originStage> \
                 --accepted <csv> --by /myflow-do-done
```

Happy path: resolve `STATE_FILE` per the State file contract → parse → confirm the current stage
is in `--accepted` → `jq` a whole-object write carrying every unowned field forward → print the
handoff block → exit 0.

Escalation exit codes, on which the invoking command loads the `myflow-state-advance` skill and
proceeds exactly as today:

| Code | Meaning |
|------|---------|
| 3 | State file missing or unparseable, **or** a non-null `worktree` absent from `git worktree list` |
| 4 | Current stage not in `--accepted` (mismatch handoff + override prompt are the skill's job) |
| 5 | Dynamic target and `originStage` is `null`/missing |
| 6 | Dynamic target and `originStage` holds a value outside the six legal origins (corruption) |

The script never prompts, never lowers a gate, never touches Jira, and never writes a stage
outside the accepted set.

**Deliberate narrowing.** On the happy path, the artifact-contradiction checks from State
self-heal no longer run for these seven commands; only "state file parses" and "worktree still
exists" do. This is a real behavior change, not a pure refactor. It is acceptable because these
stages are pure human-confirmation writes, and `/myflow-review` and `/myflow-finish` still verify
merge and gate state independently. Full self-heal remains the skill's job whenever the script
escalates.

### 3. Adversarial slot to the economy tier

In `skills/openspec-apply-superpowers/SKILL.md`:

- Slot 4's Model column: `parent (omit model)` → economy tier.
- "Economic model mapping (slots 5+ only)" → "slots 4 and 5+", including the body sentence that
  restricts it to the conditional lens reviewers.
- The guardrail at :338 that forbids omitting `model` on a slot 5+ reviewer extends to slot 4.

Slots 0 and 2 stay on the parent model: plan alignment and principle tradeoffs are judgment work
that degrades on a weaker agent. Slot 4's job — regression and test-theater hunting over a bounded
diff — is the narrowed-breadth shape the mapping already justifies for economy agents.

`openspec-apply-fix-superpowers` and `openspec-fast-path-superpowers` do not restate the roster,
but both name the slot range in prose (`apply-fix:43`, `apply-fix:194`, `fast-path:296`) and are
swept for wording.

### 4. The agents repo becomes a first-class myflow project

`openspec init`, plus a `.myflow/project.md`:

| Key | Value |
|-----|-------|
| `## apps` | No runnable application. Verification is the guard scripts; `setup.sh` is exercised against a sandboxed `HOME`. |
| `## test` | `scripts/test-setup.sh` |
| `## lint` | `scripts/check-vocabulary.sh`, `scripts/check-references.sh`. **No auto-fix command exists** — stated explicitly so the Lint Fix Priority rule's "auto-fix first" step is knowingly inapplicable rather than silently skipped. |
| `## standards` | `CLAUDE.md`, `AGENTS.md` |
| `## jira` | `KAN` |

This fixes a live defect rather than merely tidying: hosted in gymie (the KAN-7 precedent),
`/myflow-review` would run `./gradlew ktlintCheck detekt` against a change containing no Kotlin,
and the Gate C guide would describe gymie's three apps.

### 5. `scripts/check-references.sh`

Moving a section invalidates every cross-reference to it across ~18 skills, 36 command files in
two trees, and the repo-root docs. `check-vocabulary.sh` cannot catch this — it greps a fixed list
of known-retired literals.

**Rule:** for every line carrying both a `**bold token**` and a backticked `.md`/`.mdc` path, at
least one bold token on that line must match a `##`/`###` heading in that file. Tolerant of the
phrasings in use ("see **X** in", "per **X** in", "defined once under **X** in"), and it fails
exactly when a referenced section moves or is renamed.

Takes no arguments; owns its own `DEFAULT_TARGETS`, the same contract as the other two guards.
`/myflow-review` runs it alongside them.

## Verification

- `scripts/test-setup.sh` green (setup.sh gains one skill directory; assertions are additive).
- `scripts/check-vocabulary.sh` green.
- `scripts/check-references.sh` green against the swept tree.
- A sandboxed-`HOME` `setup.sh global` run showing: the managed block down from ~60 KB to ~31 KB
  (the block also carries `lint-fix-priority.mdc` at ~2 KB plus the delimiter wrapper, so it is
  always ~2 KB larger than the rule file itself), all four stubs present, and
  `~/.claude/skills/myflow-contracts/` installed in all three harness skill directories.
- Before/after byte counts recorded as evidence.

## Risks

| Risk | Mitigation |
|------|------------|
| Stale cross-references after the move | `check-references.sh`, run by `/myflow-review` |
| Agents never load the contracts they need | Stubs name the contract and its file at the original location; consuming skills carry an explicit guardrail |
| Self-heal narrowing hides real drift | Script still escalates on a missing worktree; `/myflow-review` and `/myflow-finish` verify independently |
| Existing installs carry the old block | Next `setup.sh global` rewrites the managed block in place; no migration step needed |

## Decisions

### Packaging of the extracted contract sections

**ID:** `contracts-packaging`
**Status:** active
**Chosen:** A new `myflow-contracts` skill with one file per section — reuses `install_skills`
and name-based resolution that already work in all three harnesses, and per-file granularity means
the average consumer loads ~85 lines instead of ~385.
**Considered:** A plain `.md` under `rules/` — ignored by both installer globs, so consumers would
need a machine-varying agents-repo path. Co-locating each section beside its main consumer —
**State file** has ~10 consumers, so it still needs one canonical home plus cross-skill absolute
paths, multiplying the same problem.

### Depth of the rule-file split

**ID:** `split-depth`
**Status:** active
**Chosen:** Contract sections only (~57 KB → ~32 KB) — the four sections are read solely by skills
that already load a `SKILL.md`, making the extraction mechanical.
**Considered:** Also moving rationale prose (~15-18 KB) — that prose exists specifically to stop
agents rationalizing around the rules, so removing it from always-on context risks the exact
failures it prevents. A middle ground leaving one-line "why" stubs (~20-22 KB) was rejected as
carrying most of that risk for part of the saving.

### Division of labor between script and skill

**ID:** `state-advance-split`
**Status:** active
**Chosen:** Script handles the happy path and escalates with distinct exit codes — preserves the
full existing behavior (name resolution, override prompt, self-heal) while making the common case
cheap.
**Considered:** Script-only with a hard fail on anything unusual — cheapest, but loses self-heal,
automatic name resolution, and the override prompt. Script also performing self-heal — keeps the
rarest paths cheap, at the cost of substantially more shell logic to test for little gain.

### Host repository for this change

**ID:** `host-repo`
**Status:** active
**Chosen:** Initialize openspec in the agents repo and give it its own `.myflow/project.md` — the
change touches zero gymie files, and it makes `/myflow-review` run the verification that actually
applies.
**Considered:** Following the KAN-7 precedent and hosting in gymie — no new setup, but the
change's lint, test, apps, and manual-test guide would all be drawn from the wrong project.

### Guard for moved-section references

**ID:** `reference-guard`
**Status:** active
**Chosen:** A new `scripts/check-references.sh` — this change's main failure mode is a stale
pointer, and commit `af76a3a` shows the same drift class has bitten this repo before.
**Considered:** Relying on the review panel and a manual sweep — nothing would re-check it on
later changes. Extending `check-vocabulary.sh` — mixes two concerns in a script whose contract is
"grep for retired literals".
