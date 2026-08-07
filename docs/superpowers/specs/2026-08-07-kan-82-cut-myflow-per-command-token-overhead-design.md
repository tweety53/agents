# Cut myflow per-command token overhead — design

**Change:** `kan-82-cut-myflow-per-command-token-overhead`
**Jira:** KAN-82
**Date:** 2026-08-07

## Problem

Every `/myflow-*` command loads its contract files before doing any work, and those files have
grown large enough that the load dominates the run. Measured on this tree:

| File | Lines | Bytes | Loaded by |
|------|-------|-------|-----------|
| `skills/myflow-contracts/pipeline.md` | 1540 | 103326 | every `/myflow-*` command |
| `skills/myflow-contracts/jira-integration.md` | 732 | 52783 | `/myflow-start`, `/myflow-do`, `/myflow-finish` |
| `skills/myflow-contracts/project-configuration.md` | 542 | 45236 | `/myflow-do`, `/myflow-finish` |
| `skills/myflow-contracts/state-file.md` | 236 | 15951 | any state read or write |

A `/myflow-start` run loads the first, second and fourth of those before its first question.

The content itself is not the problem. Each file mixes two kinds of writing with two different
audiences: rules an agent must obey mid-run, and the reasoning behind those rules, which serves
whoever edits the contract. The agent pays for both on every run; the editor needs both once,
rarely.

Structural content and prose measure as follows — where *structural* is table rows, fenced blocks,
headings, list items and block quotes, and *prose* is everything else:

| File | structural | prose | prose % |
|------|-----------|-------|---------|
| `pipeline.md` | 21702 | 81624 | 78% |
| `jira-integration.md` | 4985 | 47798 | 90% |

**Prose is an upper bound on what can move, not an estimate of what will.** Many prose paragraphs
are normative — `pipeline.md`'s "Never substitute a commit count" and
`jira-integration.md`'s "Resolve transitions by name, never by identifier" are both prose and both
core. The partition test below is behavioural, not typographic.

One further measurement corrects the issue: `jira-integration.md`'s `### Follow-up issues` section
is **457 lines / 35467 bytes**, not the ~250 lines the issue estimated. It is 67% of that file, and
it is reachable only from `/myflow-finish` run 1.

## Scope

**In scope** — splitting `skills/myflow-contracts/pipeline.md` and
`skills/myflow-contracts/jira-integration.md`, and adding a guard that keeps the results small.

**Out of scope** — the issue's third cut, pruning the harness system prompt. That is an operator
action against an installed environment (uninstalling unused plugins, trimming skill descriptions),
not a change to this repository. It is worth doing and belongs on its own issue.

Also out of scope: `project-configuration.md` (45236 bytes), `workspace-isolation.md` and
`plan-provenance.md`. They are ratcheted by the budget guard where they stand, and a later split of
any of them is a budget edit plus the same partition work.

**No rule changes and no behaviour changes.** Every `/myflow-*` command behaves identically before
and after. This change moves text and adds one guard.

## The partition rule

One test decides every passage:

> **A passage is core if removing it would change what an agent does.** Tables, ordered step lists,
> templates, verdict tables, code blocks an agent runs, and every prohibition are core.
> **A passage is rationale if removing it changes no agent behaviour**: justification of an
> ordering, alternatives considered and rejected, history ("this reverses an earlier decision"),
> measurements, and meta-commentary about the document itself ("stated here rather than left to be
> re-derived", "this file is the only copy").

The test is applied to sentence groups — a paragraph, a bullet, a table — never to individual
clauses inside a sentence.

**Every sentence lands in exactly one file, verbatim.** Nothing is rewritten, summarised or
re-authored. The issue's constraint — a split must not produce two independently authored
statements of one rule — then holds by construction rather than by review diligence, and the diff
is reviewable as a move.

**One edit is permitted: repointing a citation whose target moved.** A backticked path, and the
bold section token beside it, may change to name the file that now holds the target. Every other
character of a moved sentence is untouched. `scripts/check-references.sh` verifies that each
repointed citation resolves to a real heading in the file it now names, so the repointing is
mechanically checked rather than trusted.

## The resulting files

Five files where there were two.

| File | Holds | Loaded by |
|------|-------|-----------|
| `skills/myflow-contracts/pipeline.md` | the normative core | every `/myflow-*` command |
| `skills/myflow-contracts/pipeline-rationale.md` | its reasoning | whoever edits the contract |
| `skills/myflow-contracts/jira-integration.md` | the normative core | `/myflow-start`, `/myflow-do`, `/myflow-finish` |
| `skills/myflow-contracts/jira-integration-rationale.md` | its reasoning | whoever edits the contract |
| `skills/myflow-contracts/jira-followups.md` | `Follow-up issues` — its rules **and** its reasoning | `/myflow-finish` run 1 only |

**Appendices mirror their core's heading tree.** The same `##` and `###` headings appear in the
same order in both files, each appendix section holding only that section's reasoning. A section
that is wholly normative leaves its appendix heading present with no body under it, rather than
absent — the same missing-rather-than-dropped rule this pipeline applies to handoff fields, and
what makes it visible that the section was examined rather than skipped. A citation naming a
section then resolves against whichever file the path names, and an editor reads the two side by
side.

**`jira-followups.md` keeps its reasoning inline and is not core/appendix split.** Only
`/myflow-finish` run 1 loads it, so its reasoning costs nothing on `/myflow-start` or `/myflow-do`,
and splitting it would buy a saving on the one command that already pays the least attention to
this file's size. Moving the section out at all is the largest single lever in the Jira half:
`/myflow-start` and `/myflow-do` stop loading 35467 bytes per run before any core/appendix work
happens.

## The budget guard

`scripts/check-contract-budget.sh` — new, argument-free, self-scoped from its own location like
`check-references.sh` and `check-vocabulary.sh`.

- **Covers every `skills/myflow-contracts/*.md`**, split or not, including the appendices and
  `SKILL.md`. A file in that directory with no budget entry is itself a failure, so a contract
  added later cannot silently escape the ratchet.
- **Measures bytes**, declared as a path→max-bytes table inside the script. One place to read, no
  dependency on a tokenizer, and raising a budget is a visible diff in the guard rather than an
  edit to the file being ratcheted.
- **Each budget is the size the split actually achieves, plus 25%.** The guard is a ratchet against
  regrowth, never a target the split itself can fail against — which is what stops normative text
  being pushed into an appendix to make a number.
- **Exit codes** follow this repository's guard convention: 0 every file within budget, 1 a file
  over budget or missing an entry, 2 the guard cannot answer at all.
- `scripts/test-check-contract-budget.sh` beside it, and an entry in `.myflow/project.md`'s
  `## lint`.

## The runtime rule

**A `/myflow-*` run never loads an appendix.** Appendices exist for whoever edits a contract; a
command that loads one has paid the cost this change exists to remove.

Stated in `skills/myflow-contracts/SKILL.md`, in each core file's header, and in
`rules/myflow-manual-review.mdc`. **Nothing enforces it** — the same treatment this corpus gives
its other judgment rules, and a naming-convention guard was considered and rejected as machinery
around a rule no observed failure has yet broken.

## Distribution

**No installer change is needed, and this was verified rather than assumed.** `install_skills`
(`setup.sh:188`) iterates `skills/*/` and symlinks each whole skill *directory* into
`~/.claude/skills/`, `~/.cursor/skills/` and `~/.codex/skills/`. New files inside
`skills/myflow-contracts/` therefore ship to all three harnesses with no edit to `setup.sh`.

This is asserted as a delta-spec scenario, not implemented as code.

## Verification

| What must hold | How it is proved |
|----------------|------------------|
| No sentence was lost or altered | The multiset of non-blank lines in core+appendix equals that of the original file taken from `git show HEAD:<path>`, after excluding lines whose only difference is a repointed citation path. Lines, not sentences: a line is what `sort`/`comm` can compare without a parser, and the partition moves whole lines by construction |
| No citation went stale | `scripts/check-references.sh` stays green across all 258 bold-token references in the corpus |
| The cores are and stay small | `scripts/check-contract-budget.sh` |
| The new files ship | An asserted scenario over `install_skills`, plus `scripts/test-setup.sh` |
| Nothing else regressed | The full `## lint` and `## test` lists in `.myflow/project.md` |

## Consequential edits

- `openspec/specs/myflow-contract-distribution/spec.md` — its scenario asserts the directory holds
  *exactly* six files. That is already stale on this tree (`build-green.md`,
  `plan-provenance.md` and `workspace-isolation.md` exist), so it needs a delta regardless of this
  change, and this change is where it gets one.
- `skills/myflow-contracts/SKILL.md` — index rows for the three new files, and the runtime rule.
- `rules/myflow-manual-review.mdc` — its contract table gains the new rows.
- `CLAUDE.md` and `AGENTS.md` — the `skills/myflow-contracts/` row in each skill index.

## Alternatives considered

**One shared `rationale.md` for all contracts.** Fewer files, but an editor of one contract loads
the reasoning of all of them — the same audience mismatch this change exists to fix, one level up.

**Same file, rationale below a runtime-stop marker.** No new files and no repointing, but a `Read`
of the file still pulls the whole thing into context, so it saves nothing that matters.

**Delete the rationale outright.** Smallest result. Rejected: this corpus's reasoning is load-
bearing for whoever edits it, and git history is not where an editor looks. Several sections exist
precisely because a previous editor re-derived a decision wrongly.

**Rewrite the core as a fresh summary.** Smallest core, and exactly the two-independently-authored-
statements failure the issue forbids.

**Target-first budgets** (declare `pipeline.md ≤ 24000` up front and make the split fit).
Guarantees the saving, at the cost of pressure to push normative text into an appendix to hit a
number.

**A guard enforcing the runtime rule** via an appendix naming convention. Rejected as machinery
around a rule nothing has yet broken.

## Risk

The achieved size is not known until the partition runs. The issue targets `pipeline.md` at roughly
8000 tokens from roughly 34000; the measured 78% prose fraction makes that plausible but is an
upper bound, since normative prose stays. If the partition lands the core at 40000 bytes rather
than 24000, that is the result and the budget records it. The budget-from-achieved-size decision
exists so that this outcome is a smaller win rather than a failed change.
