# Remove the manual test guide and its gate — design

**Change:** `kan-107-remove-manual-test-guide-and-gate`
**Jira:** KAN-107
**Date:** 2026-08-08

## Purpose

`/myflow-do` writes `docs/manual-test/<name>.md` — a tickable behaviour checklist with a
`## Known incomplete` section — and `/myflow-finish`'s unfinished-work guard reads it back as two of
its four signals. Remove that artifact and everything that reads it. In its place, the `/myflow-do`
handoff prints the instructions for running whatever is in scope.

The human gate at `IN_PROGRESS` is unchanged in substance: the operator still reviews the staged
diff and still runs the apps. What goes away is the generated file, its checkboxes, and the guard
signals derived from them.

## Scope

In scope: the `myflow-manual-test-guide` capability, the guide-related requirements inside four
other capabilities, `/myflow-do` section 6, the `IN_PROGRESS` handoff template, the planning-path
staging exclusions, `check-unfinished-work.sh`, `gather-self-review-context.sh`, the affected test
harnesses, and the digest blocks in the repository's documentation.

Out of scope: the archived changes under `openspec/changes/archive/`, which record history and are
never rewritten. `check-vocabulary.sh`'s banned-vocabulary allowlist is also out of scope — its
`awaiting-manual-test`, `myflow-manual-test` and `skip-manual-test` entries ban retired *command*
names, not this artifact.

## Decisions

### The human gate keeps both halves

**ID:** gate-keeps-both-halves
**Status:** active
**Chosen:** `IN_PROGRESS` continues to mean "review the staged diff and run the apps" — only the
generated artifact and its checkboxes disappear.
**Considered:**
- *Diff review only* — ruled out: running the apps is the half of the gate that catches what a diff
  cannot, and dropping it would silently narrow the pipeline's only behavioural check.
- *No gate at all* — ruled out: that is a state-machine change, not an artifact removal, and it
  would leave nothing between implementation and the irreversible finish step.

### Run instructions are printed in the handoff, not written to a file

**ID:** run-instructions-in-handoff
**Status:** active
**Chosen:** the `/myflow-do` handoff block carries the commands to run whatever is in scope, with
every path absolute and every URL resolved from this worktree's workspace id. Nothing is written to
disk and nothing is committed.
**Considered:**
- *A file with no checkboxes* — ruled out: keeps the file, the path and the staging exclusion, so
  the removal would be cosmetic.
- *Cite `.myflow/project.md` and stop* — ruled out: loses the absolute-path and per-worktree URL
  resolution, which is the part of today's preamble that carries actual value. A worktree bound to
  a moved port would be sent to the project's declared `localhost` URL.

### The run-instructions template lives in `handoff-blocks.md`

**ID:** template-in-handoff-blocks
**Status:** active
**Chosen:** the `IN_PROGRESS` template in `skills/myflow-contracts/handoff-blocks.md` gains the
run-instructions section in place of its `Test guide:` line. `/myflow-do` carries the block it
prints; `/myflow-status` loads that file and regenerates from the same template.
**Considered:**
- *Each command composes its own* — ruled out: cheaper diff, but the two renderings drift, which is
  the exact failure `handoff-blocks.md` was split out to prevent.

### Both guide signals are deleted from the unfinished-work guard

**ID:** delete-guide-signals
**Status:** active
**Chosen:** `check-unfinished-work.sh` drops from four signals to two — plan checkboxes (including
`<name>-fix-N` sub-changes) and open review-panel findings. The `GUIDE` variable, the missing-guide
branch, the guide's unticked-box count and the fence-aware `## Known incomplete` scan are removed.
**Considered:**
- *Relocate `## Known incomplete` onto `tasks.md`* — ruled out: more rework for a section the plan's
  own unticked boxes already express. An unfinished task is an unticked box.

### `docs/manual-test/` is deleted outright

**ID:** delete-guide-directory
**Status:** active
**Chosen:** remove the directory and the fourteen guides committed in it. Git history retains them.
**Considered:**
- *Keep them as history* — ruled out: leaves a directory nothing produces and nothing reads, which
  is the state "completely" was meant to avoid.

### The planning-path list drops from three paths to two

**ID:** two-planning-paths
**Status:** active
**Chosen:** the paths `/myflow-do` never stages become `openspec/` and `docs/superpowers/`.
**Considered:**
- *Keep the three-path list* — ruled out: an exclusion for a path that can no longer exist is a
  contract stating something untrue, and a later reader cannot tell it from an oversight.

### The capability is removed; one requirement moves to `myflow-handoff-output`

**ID:** remove-capability-move-requirement
**Status:** active
**Chosen:** delete `openspec/specs/myflow-manual-test-guide/` in full, and add one requirement to
`myflow-handoff-output`, which already owns handoff shape and the absolute-path rule.
**Considered:**
- *Repurpose the capability in place* — ruled out: keeps a capability name that no longer describes
  what it governs.
- *Remove it and add nothing* — ruled out: the run instructions would then be governed only by a
  skill file, with no requirement stating what the handoff must show.

### The plan commit is renamed, and the grep matches both wordings

**ID:** rename-plan-commit
**Status:** active
**Chosen:** the second commit becomes `chore(<name>): plan and session records`.
`gather-self-review-context.sh` matches both wordings through one ERE alternation,
`plan(, test guide and| and) session records`, applied at both of its sites.
**Considered:**
- *Keep the old subject verbatim* — ruled out: the wording names an artifact that no longer exists.
- *Match the new wording only* — ruled out: every change already in history would stop resolving its
  plan commit, and `IMPL_SHA`'s exclusion would then pick that commit up as the implementation
  commit. A wrong answer, not a missing one.

## Open questions

*(none — every question raised during brainstorming was answered)*

## What is deleted

- `openspec/specs/myflow-manual-test-guide/` — the whole capability, all three requirements, via a
  `REMOVED` delta.
- `docs/manual-test/` — the directory and its fourteen committed guides.
- `check-unfinished-work.sh` — signals one and four: the `GUIDE` variable, the missing-guide check,
  the guide's unticked-box count and the fence-aware `## Known incomplete` awk pass. `count_unticked`
  stays; the plan signal uses it. `unreadable()` stays for the same reason.
- `test-check-unfinished-work.sh` — every guide-fixture case.
- `/myflow-status` section 4 — the guide's path-and-box-count bullet.
- `docs/manual-test/` from the staging exclusions in `pipeline.md`, `skills/myflow-do/SKILL.md`,
  `skills/myflow-finish/SKILL.md`, `finish-contract.md` and `test-uncommitted-review-package.sh`.
- The guide requirements inside `myflow-state-machine`, `myflow-handoff-output`,
  `myflow-finish-cleanup` and `myflow-command-surface`.

## What replaces it

`skills/myflow-do/SKILL.md` section 6 is renamed **Resolve the run instructions**. It writes no
file. It resolves, for the handoff:

- every app root, absolute, from `git worktree list` or the state file's `worktrees` keys;
- every start command from `.myflow/project.md`'s `## run`, with paths absolute;
- every URL resolved from this worktree's workspace id, the way section 2 computes it — never the
  project's declared base;
- for a project declaring no isolation, the declared URLs unchanged;
- for a project with no runnable application — this repository is that case — the guard and test
  commands from `## lint` and `## test`, absolute.

Those resolution rules carry over verbatim from the old preamble. What is dropped is the checklist,
the `## Known incomplete` section, the capability grouping and the tickable lines.

The `IN_PROGRESS` template's `Test guide: <path>` line is replaced by:

```text verified:authored in-tree for this change
Worktree:   <absolute worktree path>

Run it:
  <command>          # <app or check name>
  <command>
```

and its instruction line becomes `Review the diff, then run it:`.

The new requirement in `myflow-handoff-output` states that `/myflow-do`'s `IN_PROGRESS` handoff
SHALL carry run instructions for everything in scope, with every path absolute and every URL the one
this worktree resolved. It absorbs the removed capability's third requirement — the no-runnable-
application case — as a scenario.

## Guards, commit subject and tests

`check-unfinished-work.sh` keeps its exit contract unchanged: exit `0` whenever a verdict is
reached — the verdict line itself, `CLEAR:` or `OUTSTANDING:`, carries the answer, never the exit
status — and exit `2` only when it cannot determine at all (an unreadable worktree, a change name
outside the allowlist, or a file that exists but cannot be read). `myflow-finish-cleanup`'s signal
list goes from four to two, and the scenarios naming the guide's boxes and its `## Known incomplete`
section are removed.

The commit subject changes at three chain sites — `pipeline.md`, `skills/myflow-do/SKILL.md` and
`skills/myflow-finish/SKILL.md`:

```text verified:authored in-tree for this change
chore(<name>): plan and session records
```

`gather-self-review-context.sh` changes at both of its sites:

```bash verified:authored in-tree for this change
--grep="^chore\(${NAME_RE}\): plan(, test guide and| and) session records"
```

```bash verified:authored in-tree for this change
| grep -Ev "plan(, test guide and| and) session records|^[0-9a-f]+ chore\(${NAME_RE}\): .*archive" \
```

Both sites gain a comment naming why the alternation exists.

Test changes:

- `test-check-unfinished-work.sh` — guide cases deleted; the remaining cases keep full coverage of
  the two surviving signals.
- `test-gather-self-review-context.sh` — its two existing fixtures use the old subject and stay, as
  back-compat proof; a fixture using the new subject is added.
- `test-uncommitted-review-package.sh` — the `docs/manual-test/` fixtures and assertions drop to the
  two remaining paths.
- `test-check-references.sh` — its fixture uses `docs/manual-test/<name>.md` as a reference-shape
  example; the path is swapped and the test kept.
- `check-references.sh` and `check-plan-provenance.py` — stale comments mentioning the guide are
  updated.
- `plan-provenance.md` — the guide's row in the quoting-exception table is removed.
- `workspace-isolation.md` — "written into the change's manual test guide" becomes "printed in the
  `/myflow-do` handoff".

## Documentation and digest blocks

The three-line digest is copied deliberately into several files; the facts are fixed, the wording is
not. Updated in: `CLAUDE.md`, `AGENTS.md`, `rules/myflow-manual-review.mdc`, `README.md`,
`skills/README.md`, and both `myflow-do` command files (frontmatter `description:` and body).

The gate wording "review the staged diff and run the apps" stays everywhere. Only the mentions of
the manual test guide are removed.

`myflow-state-machine`'s requirement that `/myflow-do` produce both the diff and the guide becomes a
requirement that it produce the staged diff and the run instructions in its handoff. The
`IN_PROGRESS` state's meaning is unchanged. `myflow-command-surface` loses its "Do carries the test
guide generation" scenario, and its staging scenario goes from three paths to two.

## Verification

Every command in this repository's `## test` and `## lint` sections must pass:

```bash verified:copied from .myflow/project.md's ## test and ## lint sections
scripts/test-check-unfinished-work.sh
scripts/test-gather-self-review-context.sh
scripts/test-uncommitted-review-package.sh
scripts/test-check-references.sh
scripts/test-setup.sh
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/check-task-build-green.sh
scripts/check-workspace-isolation.sh
scripts/check-contract-budget.sh
```

Plus a sandboxed installer pass:

```bash verified:copied from .myflow/project.md's ## run section
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
```

Shell and guard changes are made red-first: adjust or delete the test, watch it fail, then change
the script. Markdown contract edits are verified by `check-references.sh`, which resolves every
citation, and by `check-contract-budget.sh`.
