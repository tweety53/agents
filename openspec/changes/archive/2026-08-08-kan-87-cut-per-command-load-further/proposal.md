## Why

KAN-82 cut what a `/myflow-start` run loads from 156,109 to 103,844 bytes by splitting the two
largest contract files into a normative core and a rationale appendix. Two levers remain, and both
are **pure file moves** using machinery that change built and its review panel tested across two
full passes.

`skills/myflow-contracts/pipeline.md` is still 88,253 bytes — 1,348 lines, 30,351 tokens.
<!-- measured: Read tool on skills/myflow-contracts/pipeline.md @ 63fa6ee -->
It **still cannot be read in one page**: the read is split in two because it exceeds the tool's
per-read cap. Every `/myflow-*` command loads it, and 37% of it serves `/myflow-finish` alone.

And the files that *load* the contracts were never measured at all. `skills/myflow-do/SKILL.md` is
40,529 bytes — larger than every contract except `pipeline.md` and `project-configuration.md` — and
is read in full on every `/myflow-do` run. KAN-82 optimised what the skills load and never looked at
the skills themselves.

<!-- measured: wc -c skills/*/SKILL.md @ 63fa6ee -->

## What Changes

- **Hoist two passages into `pipeline.md`'s core first.** The `preserve-session-records.sh` outcome
  table becomes its own core section, and `Temporary artifacts registry` stays in the core. Both are
  cited by files `/myflow-do` loads, so moving them would put a citation in front of a file that
  command never reads — KAN-82's panel finding F1, reintroduced by design.
- **Move what is then genuinely finish-only** — `Finish contract` and `Worktree cleanup`, 25,760
  bytes — to `skills/myflow-contracts/finish-contract.md`, carrying rules **and** reasoning inline
  because a single command loads it. `pipeline.md` lands at **62,493 bytes**.
- **Split the three pipeline skills.** `skills/myflow-do/SKILL.md` (40,529),
  `skills/myflow-start/SKILL.md` (28,944) and `skills/myflow-finish/SKILL.md` (28,086) each gain a
  sibling `SKILL-rationale.md`. `myflow-status` and `myflow-info` are left alone.
- **Widen `scripts/check-contract-budget.sh`** from `skills/myflow-contracts/*.md` to also cover
  `skills/*/SKILL.md` and the new rationale siblings, with harness cases for the new glob.
- **`/myflow-info` keeps the stage table.** The state diagram, the level-1 stage table and every
  level-2 expansion stay in the core, so it still describes both finish runs and every stage; it
  gains one line naming where the finish contract's detail now lives.

**Not a behaviour change.** No rule is added, removed or reworded. Every `/myflow-*` command behaves
identically before and after.

**Out of scope:** sentence-level granularity, which needs a wrapping-normalised verification and its
own design gate; `project-configuration.md` (45,258) and `workspace-isolation.md` (31,079), both
ratcheted where they stand; and pruning the harness system prompt, which is an operator action
against an installed environment.

## Capabilities

### New Capabilities

*(none — this change widens an existing capability rather than introducing one)*

### Modified Capabilities

- `myflow-contract-economy`: its requirements govern "a contract file under
  `skills/myflow-contracts/`". The same partition rule, the same two permitted edits, the same
  heading mirroring and the same byte budget now apply to any file the pipeline loads per run,
  including a `SKILL.md`. One capability, one rule set — the operation is identical, and two
  near-identical rule sets are the duplication this corpus repeatedly punishes.
- `myflow-contract-distribution`: its index-and-directory agreement rule covers
  `skills/myflow-contracts/`; `finish-contract.md` joins that directory and must be indexed, and the
  rule that no `/myflow-*` run loads an appendix needs to name the skill-level appendices too.

## Impact

**Files split:** `skills/myflow-contracts/pipeline.md`, `skills/myflow-do/SKILL.md`,
`skills/myflow-start/SKILL.md`, `skills/myflow-finish/SKILL.md`.

**Files created:** `skills/myflow-contracts/finish-contract.md`,
`skills/myflow-do/SKILL-rationale.md`, `skills/myflow-start/SKILL-rationale.md`,
`skills/myflow-finish/SKILL-rationale.md`.

**Files edited:** `scripts/check-contract-budget.sh` and its harness,
`skills/myflow-contracts/SKILL.md`, `rules/myflow-manual-review.mdc`, `CLAUDE.md`, `AGENTS.md`, and
every file citing a section that moved.

**The verification that matters is not mechanical.** `scripts/check-references.sh` resolves a
citation whenever a heading of that name exists, and the mirroring rule guarantees one does — so it
stays green whether or not a citation still reaches its substance. KAN-82 had three findings in that
gap. `Finish contract` is cited from four skill files; each is checked by reading, and
`myflow-do`'s is the dependency the hoist exists to resolve.
