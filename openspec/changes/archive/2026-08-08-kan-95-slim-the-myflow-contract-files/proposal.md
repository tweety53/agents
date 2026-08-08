## Why

Every `/myflow-*` command reads its contract files in full before it acts. A `/myflow-do` run loads
**210,481 bytes** of contract and skill prose before dispatching its first implementer, and that
text is resent on every tool call for the whole run.

<!-- verified: wc -c on each file listed below @ f763481 -->

| File | Bytes |
|------|-------|
| `skills/myflow-contracts/pipeline.md` | 64,701 |
| `skills/myflow-contracts/project-configuration.md` | 45,274 |
| `skills/myflow-do/SKILL.md` | 37,885 |
| `skills/myflow-contracts/workspace-isolation.md` | 31,079 |
| `skills/myflow-contracts/state-file.md` | 15,951 |
| `skills/myflow-contracts/jira-integration.md` | 15,591 |

KAN-95's issue text quotes ~244 KB for the first five. That is the **budget** column of
`scripts/check-contract-budget.sh` — each budget is its file's size when its row landed plus 25% —
not the files' actual size. This change is built against measured bytes, and the correction is
stated here rather than left to be discovered by a reader comparing the two.

KAN-82 and KAN-87 already worked this problem and both declared `project-configuration.md` and
`workspace-isolation.md` out of scope. This change takes them, and takes four things beyond byte
count that were added during planning.

## What Changes

**Five parts.** KAN-95 as filed is part 1.

### Part 1 — slim the four loaded contracts

- **Evict inline rationale** from `pipeline.md`, `skills/myflow-do/SKILL.md`,
  `project-configuration.md` and `workspace-isolation.md`. Two new appendices are created for the
  latter two, which have none today; a `/myflow-*` run loads neither.
- **Move `The block each state renders` and the three state block templates** (~14,000 B) to
  `skills/myflow-contracts/handoff-blocks.md`, whose sole consumer is `/myflow-status`. The general
  handoff rules stay in the core, because every command prints a handoff.
- **Every eviction leaves a pointer** — one sentence naming the rule, then the appendix heading
  holding the reasoning.
- **Re-anchor `scripts/check-contract-budget.sh`** to what lands, plus 25%, per its own header. **No
  numeric target is set**: that header warns a target-first budget creates pressure to push
  normative text into an appendix to make a number.

`model-policy.md` and `artifacts-registry.md` were considered and rejected — the first relieves only
the cheap commands, the second re-creates KAN-82's panel finding F1 by moving a section
`/myflow-do`'s own contracts cite.

### Part 2 — remove `/myflow-info`

`skills/myflow-info/` and both command files are deleted. The state diagram, the Level 1 stage table
and every Level 2 expansion — 13,769 B of `pipeline.md` — are rewritten for humans into the root
`README.md`. **Nothing loads them**, which is why no `pipeline-stages.md` is created.

### Part 3 — slim `/myflow-status`

Six columns instead of seven: the `Worktree / branch` column is dropped, and the absolute path stays
in the detail view. The command reads the state file plus local git and **never the network** — the
`gh pr list` probe is removed, so the `PR` column shows the recorded pull request's number but no
longer its state. Merge status is kept: it is local git, and it is what splits the `IN_PROGRESS`
next-command answer between integrating and archiving.

### Part 4 — delete self-heal

`/myflow-status` is the only command that performs self-heal, so part 3 leaves the mechanism with no
performer. `skills/myflow-contracts/state-self-heal.md` is deleted and every citation stripped.

### Part 5 — compress the handoff blocks

Related fields fold onto one line, in `handoff-blocks.md` and in each producing command's copy at
once, so the two renderings never diverge. **No field is dropped.** The fold groups on-disk values
onto a `**Recorded:**` line and leaves the run-only fields as their own lines.

## Capabilities

### New Capabilities

*(none — this change modifies existing capabilities and removes one)*

### Modified Capabilities

- `myflow-contract-economy`: the verbatim-partition requirement gains a bounded carve-out for a
  mixed passage — the original moves to the appendix byte-for-byte and the core gains the rule plus
  whatever operative detail the rule cannot be followed without — and a new requirement carries the
  per-move ledger that verifies it. **A section reachable from only one command lives in its own
  file** is also amended: such a file is still not split merely for being large, but **is** split
  where the resulting appendix would not be loaded at all, since the "extra file" that rule declines
  to pay for is then not paid. `jira-followups.md` stays whole; `handoff-blocks.md` splits.
- `myflow-contract-distribution`: **Requirement: The pipeline diagram and its stage table live in
  `pipeline.md`, and nowhere else** is contradicted directly; they move to `README.md`. The
  index-and-directory agreement rule gains `handoff-blocks.md` and two appendices, and loses
  `state-self-heal.md`.
- `myflow-handoff-output`: the block definition moves file, and the identical-label-set rule becomes
  an identical-folded-line-set rule. It also gains a requirement governing `/myflow-status`'s table —
  its columns and its data sources — which no capability specifies today.
- `myflow-command-surface`: three pipeline commands plus **one** read-only one.
- `myflow-progress-visibility`: its read-only-commands-register-nothing rule names `/myflow-info`.

- `myflow-state-machine`: removes **Requirement: Self-heal validates state against artifacts**, replaces
  **State writes are monotonic with one exception** with an exception-free version, and drops the
  self-heal rebuild rule from **The state file carries only fields with a live consumer**. This
  capability was initially recorded as needing no change — that was wrong, and the error is left
  visible here rather than quietly corrected: it was checked for `/myflow-info` references and never
  for self-heal requirements, so three live `SHALL`s would have survived the archive asserting
  behaviour the corpus no longer has.

### Removed Capabilities

- `myflow-state-integrity`: both of its requirements are self-heal requirements, and part 4 deletes
  the mechanism. Monotonic state writes are unaffected — they live in `state-file.md`, not here.

## Impact

**Projected** — from per-section byte tallies and a judgment of which passages are evictable, not a
commitment:

| File | Now | Projected |
|------|-----|-----------|
| `skills/myflow-contracts/pipeline.md` | 64,701 | ~23,000 |
| `skills/myflow-do/SKILL.md` | 37,885 | ~26,000 |
| `skills/myflow-contracts/project-configuration.md` | 45,274 | ~28,000 |
| `skills/myflow-contracts/workspace-isolation.md` | 31,079 | ~20,000 |
| **`/myflow-do` per run** | **210,481** | **~128,500** |

<!-- unverified: projections from per-section byte tallies and a judgment of which passages are evictable -->

**Files deleted:** `skills/myflow-info/SKILL.md`, `commands/myflow-info.md`,
`commands-claude/myflow-info.md`, `skills/myflow-contracts/state-self-heal.md`.

**Files created:** `skills/myflow-contracts/handoff-blocks.md`,
`skills/myflow-contracts/project-configuration-rationale.md`,
`skills/myflow-contracts/workspace-isolation-rationale.md`.

**Files edited:** the four contracts, `skills/myflow-status/SKILL.md`, the three pipeline
`SKILL.md`s and their rationale siblings, `skills/myflow-contracts/SKILL.md`,
`scripts/check-contract-budget.sh` and its harness, `CLAUDE.md`, `AGENTS.md`,
`rules/myflow-manual-review.mdc`, `README.md`, `skills/README.md`. Twelve files under `docs/` mention
`/myflow-info`; they are records of past runs and are left untouched.

**Repointing load**, measured — `Model policy` (29 citations), `Handoff output` (12) and
`Temporary artifacts registry` (18) stay in the core and need no repointing, which is a large part of
why they stay:

<!-- verified: grep -rl over the working tree excluding .git and archive @ f763481 -->

| Moving section | Files citing it |
|---|---|
| `The block each state renders` | 19 |
| `Pipeline flow` | 14 |
| `Level 1 — the stages of each command` | 7 |

**Parts 1 and 5 change no rule. Parts 2, 3 and 4 are behaviour changes** and are named as such
rather than folded into the slimming claim: a command is removed, a report loses a column and a
network probe, and a correctness mechanism is deleted.

**The guard cannot see the failure part 1 can cause.** `scripts/check-references.sh` resolves a
citation whenever a heading of that name exists, so a stub carrying a heading and a pointer resolves
identically to a stub carrying a heading and nothing. That is KAN-84, which stays open. The per-move
ledger and the mandatory pointer are the mitigation, and this proposal states the gap rather than
implying the guard covers it.

**A stale state file is now never corrected, and never flagged.** That was raised as a concern
during planning and the decision was to delete the mechanism regardless.

**Out of scope:** `state-file.md` and `jira-integration.md`, beyond the re-homing part 4 forces on
the first; KAN-84 itself; and the harness system prompt, which is an operator action against an
installed environment.
