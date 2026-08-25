# myflow on spectre — design

**Date:** 2026-08-25
**Status:** approved, pending implementation plan

## Purpose

Move the myflow pipeline off OpenSpec and onto `spectre`, the minimal spec tool at
`github.com/tweety53/spectre`. One artifact layer, chosen by this repository, replaces a third-party
CLI whose surface myflow used a fraction of.

The switch is a full transition. Nothing in the pipeline speaks OpenSpec afterwards, and no project
is expected to run both.

## Scope

**The existing `openspec/` tree is frozen in place and never migrated.** Its 35 capability specs,
roughly 60 archived changes and 2 unfinished changes — `kan-295-cut-pipeline-load-cost-split-by-consumer`
and `kan-327-review-per-round-delta-after-round-1` — stay exactly where they are, readable, as the
record of how this repository got here. Both unfinished changes are abandoned; neither is recreated.

A fresh `spectre/specs/` and `spectre/changes/` start empty. No `spectre/config.md` is written: the
defaults are what myflow wants, and a configuration file added at cutover would be configuring for
its own sake.

## A rider: plans are arguments, not scripts

Not part of the cutover, added to this branch at the author's request because the cutover is what
demonstrated it. Seven times during this work an implementer measured something that contradicted
the plan — the task shape spectre could not read, the budget guard already red on `main`, the
archive path with no date prefix, sub-changes that are flat siblings, a blocked-state signal that
would have stopped `/myflow-do` running at all. Every one improved the result, and every one arrived
because the implementer measured rather than complied.

`plan-provenance.md` already carries the vocabulary for this — `verified:`, `unverified:`,
`measured:` — because a plan is a mix of checked facts and guesses. What it did not say is what to do
when a guess is disproved mid-run. It says so now, under **When a measurement contradicts the plan**
(`skills/myflow-contracts/plan-provenance.md`), which is canonical for it.

## What the pipeline stops doing

Three of the integration points have no spectre equivalent, by spectre's design. Each is a deletion
rather than a port, and each removes work the pipeline was doing.

### Delta specs

OpenSpec has a change carry `changes/<name>/specs/<capability>/spec.md` delta files, which
`/myflow-finish` run 2 merges back into the main specs tree at archive time.

spectre has no deltas. A change edits `spectre/specs/<capability>.md` **directly on its branch**, and
git computes the diff. So:

- `/myflow-start` writes spec changes into the specs tree itself, not into delta files.
- `/myflow-finish` run 2 loses step 3's "sync delta specs" half entirely. What remains is: archive the
  change folder, commit on `chore/archive-<name>`.

This removes the step in the whole pipeline most able to lose an edit silently, and it is the
strongest single reason to make this switch.

### `openspec instructions`

`/myflow-start` and `/myflow-do` call `openspec instructions <artifact-id> --change <name> --json` for
enriched artifact instructions. spectre keeps conventions out of the binary deliberately — its own
design records this as a non-goal — so the artifact shape moves into the skills that create the
artifacts. Most of that prose already lives there; what the CLI supplied is the residue.

### `openspec status --change`

Used for the `changeRoot` field. Under spectre a change's root is `<tree>/changes/<id>` by
construction, so this becomes derivation rather than a process call.

## The plan's task shape

Discovered during implementation, and the design was wrong about it. This document previously
claimed the artifact shapes themselves were unchanged. They are not.

A myflow plan marks each task with a `### <n> <title>` heading and each step beneath it with a
`- [x] **Step N: ...**` checkbox. spectre reads a task as a flat `- [ ] <n>. <text>` line and
nothing else. Measured against the archived `kan-295` plan, spectre sees no tasks at all:
`spectre list` reports `0/0`, `spectre validate` emits 31 findings — one false "malformed task line"
per step checkbox — and `spectre archive` refuses with "tasks.md has no tasks".

**A plan's tasks therefore lead with a spectre checkbox.** `- [ ] 1. Capture the verification
baselines` replaces `### 1 Capture the verification baselines`, and the task's steps are indented
two columns beneath it.

The indent is not cosmetic, and the first attempt at this proved it. spectre reports any column-0
`- [` line that is not a well-formed task as a malformed task line, so a plan whose steps stayed at
column 0 still drew one finding per step — 31 on the measured plan, with the tasks now read
correctly. Indenting the steps two columns takes that to `no findings`, exit 0, and full progress,
and both myflow guards are indifferent to the indent. It also reads truer: a step belongs to its
task, and now looks like it.

**A task's id is a flat integer.** spectre reads `- [ ] <n>.` and nothing else, so `- [ ] 1.1.` is a
task to myflow's grammar and a malformed line to spectre. The task line takes a flat integer; the
dotted form stays available to whatever else in the plan grammar uses it, but never as a task id. Both tools then agree what a task is: progress in `spectre list` becomes real,
`archive`'s unchecked-tasks refusal becomes meaningful, and `validate` is clean on a change the
pipeline itself produced.

The alternatives were rejected. Turning the task rules off in a `config.md` leaves `list` reporting
`0/0` and `archive` needing `--force`, which is the flag that also disarms the guards that matter.
Splitting the plan into a rich `plan.md` beside a thin `tasks.md` puts the task list in two places
that can drift. Teaching spectre to read headings adds to the tool the configurability it was built
to avoid.

The cost is real and lands on the guards: `scripts/lib/plan_grammar.py`,
`scripts/check-task-build-green.py`, `scripts/check-task-commit-fields.py` and
`scripts/plan-dispatch-bundles.py` all key off the `###` heading, and the contracts and
`/myflow-start` describe it. All of them move together.

## Command mapping

| OpenSpec call | Replacement | Notes |
|---------------|-------------|-------|
| `openspec new <name>` | `spectre new <id>` | refuses when the tree does not exist; there is no `init` |
| `openspec list --json` | `spectre list --json` | shape becomes `{"changes":[{"id","done","total"}]}` |
| `openspec validate [name]` | `spectre validate [id]` | exit 1 on findings, 2 on usage or IO error |
| `openspec archive <name>` | `spectre archive <id>` | `git mv` into `changes/archive/`; does not commit |
| `openspec status --change` | derived path | `<tree>/changes/<id>` |
| `openspec instructions` | skill prose | no replacement command |

`/myflow-status` reports only non-archived changes, which is exactly what `spectre list --json`
returns — it needs nothing spectre does not have.

`spectre refs <capability>#<id>` has no OpenSpec counterpart and no myflow caller. It is not wired
into the pipeline by this change.

## Guards and scripts

24 guard scripts and 21 guard tests reference `openspec/changes/...`, including 208 fixture paths.
All move to `spectre/changes/...`. The change is mechanical but wide, and the guard tests are the
safety net that proves it.

Two guards need judgement rather than substitution:

- `scripts/check-contract-budget.sh` carries a size baseline naming files under `openspec/specs/`.
  Five of those rows are ALREADY failing on `main`, before this cutover — three over budget, two
  undeclared. A frozen tree has no budget to enforce, because nothing in it can grow, so the guard
  drops its `openspec/specs/` rows entirely. That fixes a pre-existing failure as a side effect. If
  any of the frozen tree is ever unfrozen, its rows come back with it.
- Any guard asserting that delta specs were synced loses that assertion, because there are no delta
  specs to sync.

## Research mode

`skills/openspec-explore/` and `commands/opsx-explore.md` are third-party (MIT, authored by
openspec), 290 and 174 lines, and depend on `openspec` CLI calls through an `allowed-tools` grant.

They are **replaced, not ported**, by a new and much smaller skill written for this repository:

- Skill-only. Exploring a spectre tree means reading markdown; it needs no binary and no
  `allowed-tools` grant.
- Keeps the stance that earns the mode its place: a thinking partner that investigates, asks and
  captures, and never writes application code.
- Drops the OpenSpec-specific machinery — the status probing, the change-exists branching, the
  artifact-creation choreography — in favour of reading the tree directly.
- Installed as `myflow-research`, invoked as `/myflow-research`. It sits in the myflow family, not
  the spectre one, because that is where its coupling is: every guardrail it carries — no state
  file, no worktree, no branch, no commit, and `spectre new` and `spectre archive` reserved to
  `/myflow-start` and `/myflow-finish` — is a myflow concept. The spectre-specific content is one
  short section on reading a markdown tree. Naming it `spectre-research` would also have left it the
  only skill in `skills/` not matching the `/myflow-*` convention this repository uses for the
  commands it owns. Leaving a command named after the removed
  tool is the drift this repository's own rules exist to prevent.

## Documentation and distribution

`CLAUDE.md`, `AGENTS.md`, `README.md`, `.myflow/project.md`, `skills/README.md`, `setup.sh`, the five
myflow skills, the contract files under `skills/myflow-contracts/` that name OpenSpec (10 of its 27), and
the ten myflow slash commands
across `commands/` and `commands-claude/` all describe the pipeline in terms of OpenSpec and are
updated together.

Historical artifacts are NOT rewritten: `docs/self-review/`, `docs/superpowers/ledgers/`,
`docs/superpowers/reviews/`, `docs/superpowers/reports/`, existing files under
`docs/superpowers/specs/`, and the whole frozen `openspec/` tree keep saying OpenSpec, because that
is what happened at the time. Rewriting them would falsify the record.

## Execution

One cutover on a single branch, by hand, with the guard test suite as the net. The repository is
self-hosting: a half-ported repository fails its own guards, so there is no coherent intermediate
state at which to stop, and phasing would only defer the same landing. The cutover is not itself run
through `/myflow-*`, because neither pipeline is usable while it is in progress.

## Verification

1. The repository's 21 guard tests pass.
2. `spectre validate` runs clean against the new empty tree.
3. A real pipeline run end to end — `/myflow-start`, `/myflow-do`, `/myflow-finish` twice — against a
   scratch project holding a spectre tree, proving the pipeline works on the new layer rather than
   merely compiling.
4. `git status openspec/` reports nothing. The frozen tree must prove it stayed frozen.

## Non-goals

No migration of OpenSpec content, now or later. No dual-path support: nothing in the pipeline
detects which tool a project uses. No `spectre/config.md`. No new spectre features — if the pipeline
needs something spectre lacks, that is a change to spectre, proposed separately.

## Rejected alternatives

**Teaching myflow to speak both, cutting over later.** Rejected because every contract and guard
would grow a second path, which is the configurability spectre exists to remove, and because the
agents repository would then carry two artifact layers indefinitely.

**Phasing the cutover across three changes** — contracts and skills, then guards, then docs.
Rejected because the repository runs its own guards against itself: after phase 1 the guards fail,
so the phases have to land together regardless. The phasing would be presentational.

**Running the cutover as the last OpenSpec change, through `/myflow-*` itself.** Rejected despite the
appeal of exercising the old path one final time: the run's own guards would inspect a repository
mid-cutover, and it would add an entry to a tree being frozen in the same commit.

**Migrating the 35 capability specs with `spectre migrate`.** Rejected by the author. OpenSpec states
the modal verb in a requirement's body while spectre states it in the bullet, so every migrated
requirement needs a hand pass — 249 of 278 warnings on this very corpus are that one artifact. A
frozen, readable `openspec/` tree plus an empty `spectre/` tree is honest; a machine-converted tree
carrying 249 findings would not be.

**Porting `openspec-explore` as it stands.** Rejected: it is 290 lines of third-party prose built
around CLI calls that no longer exist, and a smaller skill written for this repository serves the
same purpose with less to maintain. Naming it research rather than explore is the author's call.
