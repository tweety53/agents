# Slim the myflow skills — design

**Change:** `kan-106-slim-the-myflow-skills-cut-meta-prose-extract`
**Jira:** KAN-106
**Date:** 2026-08-09

## Problem

KAN-95/KAN-82/KAN-87 already built the core/appendix partition machinery — the
`myflow-contract-economy` capability's verbatim-partition rule, the bounded rule-extraction
carve-out, the per-move ledger, and `scripts/check-contract-budget.sh` as the regrowth ratchet — and
used it to slim `pipeline.md`, `project-configuration.md`, `workspace-isolation.md` and
`myflow-do/SKILL.md`. It is not applied consistently: `myflow-finish/SKILL.md`, `myflow-start
/SKILL.md`, `pipeline.md` and several contracts still carry restatement boilerplate and justification
prose that the machinery already knows how to evict. This change applies the existing mechanism
further, adds one new shared contract for a repeated interaction shape, and removes duplicated
procedures — no new partition mechanism, no new capability for the split itself.

Two behavior changes to `/myflow-fast` are folded in on top (not slimming, raised during this
session): its setup question round goes silent-default, and its design-approval gate is dropped.

## Scope

This uses **lever 1 (evict inline rationale)** and **lever 3 (rule extraction under the carve-out)**
from `myflow-contract-economy`, applied to files KAN-95 left untouched, plus new procedural
deduplication that is not a core/appendix split at all.

| Item | What | Mechanism |
|------|------|-----------|
| A | Collapse anti-restatement boilerplate corpus-wide | Rule extraction (carve-out) — see below |
| B | Move justification prose to rationale appendices | Lever 1, existing appendices (all three `SKILL-rationale.md` siblings already exist) |
| C | New `skills/myflow-contracts/operator-prompts.md` | New shared contract, cited from 5 call sites |
| D | Stop mirroring guard grammar into skills | Lever 1 (the mirrored grammar is pure rationale — it changes no agent behavior once the guard's own error text is what informs a caller) |
| E | Deduplicate 3 repeated procedures | Cite-instead-of-restate + one new script |
| F | `scripts/prepare-workspace.sh` | New script, existing pattern |
| G | Table-drive guard test suites | **Deferred** — filed as its own follow-up at finish time |
| H | Compress report-only subagent prompts | Caveman-style compression, runs after A/B |
| I | `/myflow-fast`: silent defaults + drop design-approval gate | Behavior change, not slimming |

**Order:** C and A first (ticket's own recommendation — lowest risk, highest value). Then B, which H
depends on. D, E, F in any order. H last. I lands alongside A in the same file
(`myflow-fast/SKILL.md`).

## Item A — anti-restatement boilerplate

The pattern repeated at nearly every citation site: a full paragraph re-explaining that the cited
file is canonical and must not be restated or acted on from memory. This is rule extraction under
`myflow-contract-economy`'s bounded carve-out:

1. **State the doctrine once**, near the top of each file that cites other contracts:
   > *"Every citation below is canonical at its target. Never restate its content here and never act
   > on a remembered version of it — read it fresh each time it is needed."*
2. **The original per-citation paragraph moves to that file's `-rationale.md`, byte-for-byte**, per
   the carve-out's condition 1.
3. **Each citation site keeps its bare form** — `see Model policy (pipeline.md)` — which is what the
   core needs to act; the one-time doctrine statement is what licenses dropping the individual
   restatement, so the extracted "core gain" (condition 2) is the doctrine sentence itself, not a
   per-site sentence.
4. **Per-move ledger**, one row per collapsed paragraph, `Destination` reading
   `— (rule extracted)` and quoting the doctrine sentence, exactly as KAN-95's ledger rows do.

Files touched: `skills/myflow-do/SKILL.md`, `skills/myflow-finish/SKILL.md`,
`skills/myflow-start/SKILL.md`, `skills/myflow-fast/SKILL.md`, and every `skills/myflow-contracts/
*.md` core that currently repeats the boilerplate (confirmed present in at least `pipeline.md`,
`jira-integration.md`; the plan enumerates the rest by grep during execution, not guessed here).

**Hoist-before-move check**: the doctrine paragraph is not itself a cited target (nothing cites
*into* it), so `myflow-contract-economy`'s "passage another command depends on" rule does not apply
to it — it applies to citation targets, not to the doctrine text that licenses brevity around them.

## Item B — justification prose to existing appendices

Lever 1, unchanged from KAN-95's method, applied to files it didn't touch:

- **`myflow-start/SKILL.md` `## Convergence`** (~90 lines): rule extraction. Core keeps the
  ~12-line rule (loop until an explicit "move on"; rounds 1–2 open silently, round 3+ is offered; an
  unanswerable question is recorded rather than guessed). The defense of those choices — why the
  confirm and the offer recommend opposite courses, the tuned threshold's history — moves to
  `myflow-start/SKILL-rationale.md` under the same heading (mirrors already exist, sibling already
  present).
- Any other section in `myflow-finish/SKILL.md`, `myflow-do/SKILL.md`'s remaining prose, or
  `pipeline.md` that fails the core/rationale behavioral test (would removing it change what an
  agent does?) moves the same way. Enumerated during planning by reading each file section by
  section — not guessed here, per KAN-95's own precedent of not pre-committing to a passage list in
  the design doc.
- **`skills/myflow-fast/SKILL-rationale.md` is created** — it doesn't exist today. It receives item
  B's moves for `myflow-fast/SKILL.md` plus the rationale for item I's two behavior changes (why
  silent defaults, why the design-approval gate is dropped specifically here and nowhere else).

**Per-move ledger** applies to every paragraph moved this way, same as item A.

## Item C — `skills/myflow-contracts/operator-prompts.md`

New file, new shared contract — not a split of an existing file, so the verbatim-partition rule
doesn't govern its *creation* (there's no original to partition), only the citation sites that lose
their inlined copy of the pattern.

**Shape**, extracted from the five call sites listed in the ticket:

```text
A prompt in this shape SHALL state:
- the question, with named options
- exactly one option marked (recommended)
- what happens if the operator is silent (the safe default — always the recommended option)
- a ⚠ marker in the handoff when that silent default actually fired
```

Call sites, each shrunk to question text + options + a citation to this contract for the mechanics:

1. Brainstorming's convergence confirm (`myflow-start/SKILL.md`)
2. Review panel handback (`myflow-do/SKILL.md` §5)
3. Run-1 unfinished-work gate (`myflow-finish/SKILL.md`)
4. Fix-documenting ask (`myflow-do/SKILL.md` §3)
5. Self-review skip prompt (wherever it's currently inlined — located during planning)

Each site's own prior prose about *why* its particular default is what it is stays at the site (it's
call-site-specific rationale, not part of the shared mechanic) or moves to that file's own
`-rationale.md` if it's pure justification — same lever 1 test as item B.

**Budget guard**: `operator-prompts.md` is a new `.md` under `skills/myflow-contracts/`, so it gets a
new row in `scripts/check-contract-budget.sh`'s table, sized to its landed bytes + 25%, added at the
same re-anchoring pass as every other touched file (see Verification below).

## Item D — stop mirroring guard grammar

Confirmed by reading `scripts/check-unfinished-work.sh`: it already emits its own reject reasons
(`add "the review panel record's ... marker line(s) are spread over ... lines — they must be one
unbroken block..."` etc., and exits 2 on a genuine parse failure vs. 0/1 on a reached verdict). The
skill's copy of these rules is therefore pure rationale by the behavioral test — a caller that
violates the format learns why from the guard's own message, not from the skill re-explaining it in
advance.

- `myflow-do/SKILL.md` §5: keep only the **emit format** (what a marker line looks like when writing
  one); the parse/rejection rules move to `SKILL-rationale.md` (lever 1) or are dropped as pure
  duplication of the guard's own error text where nothing distinguishes them from that text.
- §7's exit-code table: replaced with a citation to "run the guard and read its own output," since
  the guard's exit codes and their meaning are already stated once in its header/comments.

## Item E — deduplicate three repeated procedures

1. **Empty-worktree-set stop** (3 copies: `myflow-do` §2, §7; `myflow-finish` twice) → each site
   cites "Resolving a change's worktrees" (`pipeline.md`) instead of restating the reasoning. The
   rule already lives in `pipeline.md` per KAN-95's part 4 finding; this item is purely removing the
   3 remaining restatements, no rule relocation.
2. **Two-commit chain** (`myflow-do` §7's PR-exception path, `myflow-finish` §1.2) → new
   `scripts/commit-split.sh <worktree> <name> <impl-msg> <plan-msg>`, wrapping exactly the guarded
   `reset -q` / `add -A ... exclude` / conditional-commit / `add -A` / conditional-commit chain
   already specified in `pipeline.md`'s Git boundaries section (that section's own bash block becomes
   this script's body, verbatim in behavior). Both call sites invoke the script; `pipeline.md`'s Git
   boundaries section keeps the chain as the canonical spec of what the script must do (a spec is not
   a restatement the carve-out needs to touch — it's the contract the script is measured against).
3. **`planningEffort` retired-key fallback** (3 call sites: `myflow-start`, `myflow-do`,
   `myflow-finish`) → one description in `state-file.md` (already the canonical file for the state
   shape), the other two sites cite it instead of re-explaining the fallback.

## Item F — `scripts/prepare-workspace.sh`

`myflow-do` §7 currently does five jobs in prose: the workspace-isolation guard, the variable
export, lint, test, staging, state write. The first two are mechanical and script-shaped exactly like
`commit-split.sh` above: `prepare-workspace.sh <worktree>` validates the workspace-isolation
preconditions, exports the derived variables, and prints what it exported (so the skill's remaining
prose can say "run the script, then lint/test with what it exported" instead of carrying the
derivation rules inline). `workspace-isolation.md` stays the canonical spec the script implements.

## Item G — deferred

`test-check-plan-provenance.sh` (4711 lines) and `test-check-cleanup-complete.sh` (1608 lines
against a 993-line guard) need their own fixture-format design — what a fixture directory contains,
how an expected verdict is declared, how the harness diffs actual vs. expected. That's a second
design effort with its own risk surface (rewriting ~6000 lines of test assertions), independent of
every other item here. **Filed as a follow-up Jira issue at `/myflow-finish` run 1**, not attempted
in this change.

## Item H — compress report-only subagent prompts

Runs after A and B (ticket's own stated sequencing: "once the meta prose and the justification text
are gone, what remains is almost all load-bearing rule, so compressing beforehand would only make the
duplication harder to see").

Caveman-style compression, applied only to:

- `skills/myflow-do/principles-reviewer-prompt.md`
- `skills/myflow-do/adversarial-reviewer-prompt.md`
- any locate-and-report dispatch prompt in `myflow-do` §4

**Not compressed** (ticket's own exclusion list, carried forward unchanged): implementer dispatch
clauses, negation-built prohibition text, guard grammar compared byte-for-byte, handoff blocks,
commit messages, spec text, Jira descriptions — anything persisted, or anything whose meaning a
compression pass could silently invert.

## Item I — `/myflow-fast` behavior changes

Two changes to `skills/myflow-fast/SKILL.md`'s creating-run branch, both approved this session:

1. **Silent defaults.** The setup question round (planning effort, 3 model roles, review panel
   roster — currently 5 `AskUserQuestion` calls per `myflow-start`'s "Ask the planning effort, the
   models, and the review panel roster" section) is skipped entirely on `/myflow-fast`. The recorded
   values are written directly: `planningEffort: default`, `models.implementation: sonnet`,
   `models.reviewPanel: sonnet`, `models.panelFix: sonnet`, `reviewPanelRoster: light`. An explicit
   session instruction (e.g., operator says "use opus for implementation" before or during the run)
   still overrides the named field — recorded with the dispatch exactly as `pipeline.md`'s Model
   policy section already permits for any operator override, just without an interactive prompt as
   the trigger.
2. **Drop the design-approval gate.** Brainstorming's clarifying questions and design-section
   presentation stay fully interactive — that's where requirements get gathered and where the
   operator can redirect. What's removed is the separate explicit "does this look right, yes/no"
   confirm after the design is presented: `/myflow-fast` proceeds straight into artifact creation
   once the design is presented, unless the operator objects during that presentation. **This is an
   explicit, scoped override of `superpowers:brainstorming`'s hard-gate, and applies to `/myflow-fast`
   only** — `/myflow-start` keeps the gate exactly as it is today. The human checkpoint this removes
   is replaced by the one that already exists downstream and is unaffected by this change: the
   `IN_PROGRESS` staged-diff-and-run-instructions review.

Both changes are recorded with their reasoning in the new `myflow-fast/SKILL-rationale.md` (item B).

`skills/myflow-fast/SKILL.md`'s own guardrails list gains two lines and loses the two that these
changes now contradict ("Never auto-answer a brainstorming question" is narrowed to cover only the
clarifying/design questions, not the dropped confirm; a new line states the question round never
asks).

## Not a behavior change, except where it is

Items A, B, C, D, E, F, H are pure slimming — no rule an agent follows changes, verified by the
per-move ledger and the carve-out's "no restatement of the argument" bound. Item G is deferred
rather than attempted. **Item I is a behavior change**, named as such rather than folded into the
slimming claim, exactly as KAN-95 named parts 2–4.

## Verification

- **Per-move ledger**: every task in items A, B, D that evicts or extracts a passage emits the
  four-column ledger table before its diff reaches the review panel, per
  `myflow-contract-economy`'s "A move or eviction is recorded in a per-move ledger" requirement —
  unchanged mechanism, applied to more files.
- **Budget guard re-anchor**: `scripts/check-contract-budget.sh`'s table gains rows for
  `operator-prompts.md` and `myflow-fast/SKILL-rationale.md` (new files), and every touched file's
  existing row is re-anchored to its landed size + 25%, done once as the last action, per the
  guard's own stated rule.
- **New scripts get test coverage**: `commit-split.sh` and `prepare-workspace.sh` are new executable
  surface, not prose — they get their own `test-*.sh` harnesses (existing convention: every guard
  script in `scripts/` has a `test-` sibling) and `## test` entries in `.myflow/project.md` if that
  file lists scripts individually (checked during planning).
- **`check-references.sh`** must stay green throughout — it catches a citation whose target heading
  no longer exists, though not a hollowed one (KAN-84, still open, unaffected by this change).

## Risks

1. **Same gap KAN-95 named**: `check-references.sh` cannot detect a stub that lost its pointer
   without losing its heading. The ledger is the mitigation, not a fix — KAN-84 stays open.
2. **Item C's five call sites span three different skill files** (`myflow-start`, `myflow-do`,
   `myflow-finish`) plus one not-yet-located self-review prompt — more files touched by one item than
   any other, raising the chance one site's call-site-specific rationale gets misclassified as shared
   mechanic (or vice versa) during extraction.
3. **Item I removes a stated hard-gate** from one command. The mitigation is that it's scoped
   explicitly to `/myflow-fast`, documented in that skill's own rationale file, and the downstream
   `IN_PROGRESS` review remains the operator's actual checkpoint — but it's a real reduction in
   pre-implementation stops for this one command, accepted knowingly.
4. **Scope**: 9 items (A–I, G deferred so effectively 8 landing), multiple files, one worktree, one
   review panel. Smaller than KAN-95's 5-part change in mechanism (this reuses KAN-95's machinery
   rather than building new), comparable in file count touched.

## Out of scope

- Item G (test-suite table-driving) — filed as a follow-up at finish time.
- Any new capability or requirement in `myflow-contract-economy` — this change is an *application*
  of that capability's existing rules to more files, not a change to the rules themselves.
- `state-file.md`, `jira-integration.md`, `project-configuration.md`, `workspace-isolation.md` beyond
  what item E's `planningEffort` consolidation forces on `state-file.md`.
