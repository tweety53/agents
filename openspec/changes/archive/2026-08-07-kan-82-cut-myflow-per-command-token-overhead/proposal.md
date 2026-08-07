## Why

Every `/myflow-*` command loads its contract files before doing any work, and two of those files
have grown large enough that the load dominates the run: `skills/myflow-contracts/pipeline.md` is
103326 bytes and is read by **every** command, and `skills/myflow-contracts/jira-integration.md` is
52783 bytes and is read by three of them. A `/myflow-start` run pays both before it asks its first
question.

The content is not the problem — the audience mix is. Each file interleaves rules an agent must
obey mid-run with the reasoning behind those rules, which serves whoever *edits* the contract. The
agent pays for both on every run; the editor needs both once, rarely. Splitting the two apart costs
the editor one extra file to open and gives every command back the larger half.

## What Changes

- **Split `pipeline.md`** into a normative core (same path) and
  `skills/myflow-contracts/pipeline-rationale.md`, by strict partition: every existing sentence
  lands in exactly one of the two, verbatim, with one permitted edit — repointing a citation whose
  target moved.
- **Split `jira-integration.md` on two axes.** Its `### Follow-up issues` section is reachable only
  from `/myflow-finish` run 1 and is 457 lines / 35467 bytes.
  <!-- measured: awk '/^### Follow-up issues/{f=1} f{n++; b+=length($0)+1}' skills/myflow-contracts/jira-integration.md @ adedf66 -->
  It moves whole to `skills/myflow-contracts/jira-followups.md`, keeping its rules *and* its
  reasoning inline. What remains is partitioned into a core (same path) and
  `skills/myflow-contracts/jira-integration-rationale.md`.
- **Appendices mirror their core's heading tree** — same `##`/`###` headings, same order — so a
  citation resolves against whichever file its path names, and an editor reads the two side by side.
- **Add `scripts/check-contract-budget.sh`**, a new lint guard carrying a path→max-bytes table
  covering every `skills/myflow-contracts/*.md`. Each budget is the size the split actually achieves
  plus 25%: a ratchet against regrowth, never a target the split can fail against. Companion
  `scripts/test-check-contract-budget.sh`, and an entry in `.myflow/project.md`'s `## lint`.
- **State the runtime rule**: a `/myflow-*` run never loads an appendix. Appendices are for whoever
  edits a contract. Stated, deliberately unenforced.
- **Correct the already-stale file inventory** in `myflow-contract-distribution`, which asserts the
  contracts directory holds exactly six files. It holds nine on this tree today, before this change
  adds three more.

**Not a behaviour change.** No rule is added, removed or reworded. Every `/myflow-*` command
behaves identically before and after.

**Out of scope:** the linked issue's third cut, pruning the harness system prompt. That is an
operator action against an installed environment, not a change to this repository, and belongs on
its own issue. Also out of scope: splitting `project-configuration.md` (45236 bytes),
`workspace-isolation.md` and `plan-provenance.md` — they are ratcheted where they stand by the new
guard, and a later split of any of them is a budget edit plus the same partition work.

## Capabilities

### New Capabilities

- `myflow-contract-economy`: how a contract file is partitioned into a normative core and a
  rationale appendix, what distinguishes the two, the verbatim-move and citation-repointing rules,
  the heading-mirroring rule, the rule that a run never loads an appendix, and the byte budget that
  keeps the cores small.

### Modified Capabilities

- `myflow-contract-distribution`: its file inventory names exactly six files and is already stale;
  it must name the directory's real contents including the three new files, and its
  installs-without-new-installer-code requirement must cover them.

## Impact

**Files split:** `skills/myflow-contracts/pipeline.md`, `skills/myflow-contracts/jira-integration.md`.

**Files created:** `skills/myflow-contracts/pipeline-rationale.md`,
`skills/myflow-contracts/jira-integration-rationale.md`,
`skills/myflow-contracts/jira-followups.md`, `scripts/check-contract-budget.sh`,
`scripts/test-check-contract-budget.sh`.

**Files edited:** `skills/myflow-contracts/SKILL.md` (index rows, runtime rule),
`rules/myflow-manual-review.mdc` (contract table), `CLAUDE.md` and `AGENTS.md` (skill index rows),
`.myflow/project.md` (`## lint`), and every file carrying a citation whose target moved — there are
132 backticked references to `pipeline.md` and 61 to `jira-integration.md` across the corpus.

**No installer change.** `install_skills` (`setup.sh:188`) symlinks whole skill *directories*, so
the new files ship to `~/.claude/skills/`, `~/.cursor/skills/` and `~/.codex/skills/` unchanged.
Verified against the source, and asserted as a scenario rather than implemented.

**Guards that must stay green:** `scripts/check-references.sh` proves no citation went stale across
all 258 bold-token references; the full `## lint` and `## test` lists in `.myflow/project.md` prove
nothing else regressed.
