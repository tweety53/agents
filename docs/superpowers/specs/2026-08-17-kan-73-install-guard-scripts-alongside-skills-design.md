# KAN-73 — Install the myflow guard scripts alongside the skills

**Date:** 2026-08-17
**Change:** `kan-73-install-guard-scripts-alongside-skills`
**Issue:** [KAN-73](https://tweety53.atlassian.net/browse/KAN-73)

## The problem

`setup.sh`'s `install_skills()` symlinks each `skills/<name>/` directory into
`~/.claude/skills/`, `~/.cursor/skills/` and `~/.codex/skills/`. Every guard script,
however, lives at repo-root `scripts/` — outside that tree — and every skill and contract
cites it as a bare, repo-relative `scripts/<name>.sh`.

That path resolves only when the project a `/myflow-*` command is running against happens
to be this repository. In every other project it names nothing. Confirmed on this machine:

```
$ ls ~/.claude/skills/myflow-finish/
SKILL-rationale.md
SKILL.md
```

No `scripts/` directory under any installed `myflow-*` skill.

The contracts do carry a hand-run fallback — "when the script is absent, perform the same
signals by hand" — so a hand-run is legitimate. The failure is that absence is discovered
*silently, at each call site*, and several checks specify behaviour prose cannot faithfully
reproduce: distinguishing a missing verdict line from a verdict, and checking the exit code
**as well as** the line.

### Why this matters more than a missing convenience

**KAN-70.** At `/myflow-finish` run 2, signal 1 of the preflight — `HEAD` against the
recorded merge base — established that the branch carried no commits of its own and that
everything was still staged. The ancestor test alone reports *merged* for such a branch,
because a branch with no commits is an ancestor of every branch. Had the run trusted the
ancestor test, run 2 would have archived the change and `--force`-removed both worktrees,
holding roughly 12,000 lines of uncommitted work at that moment.

**KAN-37.** Both finish runs performed every check by hand. The hand-run cleanup
verification caught a real leftover — an IDE-recreated `.idea/workspace.xml` surviving
`git worktree remove --force`, which exited 0.

## Scope

The ticket names three fixes. Two are in scope; the third is already delivered.

| Ticket item | Disposition |
|---|---|
| Install the scripts so they ship with the skills | **In scope** — the core of this change |
| Have each skill detect absence once at start and say so loudly | **In scope** |
| Add a panel-record linter for the marker-block rules | **Already delivered — dropped** |

### Why the panel-record linter is dropped

The ticket states the marker-block rules "were verified with a throwaway script written
during the session, not by a checked-in guard." That stopped being true after the ticket
was last updated (2026-08-08). `scripts/check-unfinished-work.sh` now enforces every rule
the ticket names:

| Rule | Enforced at |
|---|---|
| `findings-total: <n>` equals the number of marker lines | `check-unfinished-work.sh:385` |
| statuses compared byte for byte (`open` / `fixed` / `withdrawn`) | `:315-317`, `:442` |
| a `withdrawn` marker carries its reason on the same line | `:322`, `:341`, `:445` |
| marker lines occupy one unbroken consecutive span | `:399` |
| a reused `F<n>` cannot balance a missing one (multiset comparison) | `:413-425` |

It shares `scripts/lib/panel-record.sh` with `scripts/check-panel-reproducers.sh`, and
`scripts/test-check-unfinished-work.sh` passes clean.

Building a second guard over the same rules would create precisely the drift
`lib/panel-record.sh` exists to prevent — its own header records that
`check-unfinished-work.sh` and `check-panel-reproducers.sh` had already drifted apart on
the declared-total pattern before the library was extracted. One enforcement point, not two.

## Decisions

### Where the guards live

**ID:** `guards-ship-inside-skill-dirs`
**Status:** active
**Chosen:** Per-skill `scripts/` directories, populated with git-tracked **relative
symlinks** into repo-root `scripts/` — the one real copy of each guard stays where it is.
Ships through the existing `install_skills()` symlink with no installer change, and
repo-root `scripts/`, every `scripts/test-*.sh` harness, and `.myflow/project.md`'s `##
lint` and `## test` lists keep resolving unchanged.
**Considered:**
- *One shared dir inside `myflow-contracts`* — a single copy, but `myflow-contracts` is
  never the running command, so nothing would resolve against it under the resolution rule
  below.
- *Real files per skill, repo root symlinking back* — inverts the ownership of 20-odd files
  and forces a per-guard ownership call for the seven guards two or three skills share.
- *`setup.sh` materializes them at install time* — costs a manifest, a materialize step and
  a prune step, and leaves the repo checkout looking different from an installed user's
  tree, which is the mismatch that hid this bug in the first place.

### How a cited guard path resolves

**ID:** `one-resolution-rule-in-pipeline`
**Status:** active
**Chosen:** Skills and contracts name a guard by **basename**. `pipeline.md` states once
that a named guard resolves to `<the running command's own skill directory>/scripts/<name>`.
No call site carries a repo-relative `scripts/…` path any more.
**Considered:** *Literal skill-relative paths at every site* — explicit per site, but a
contract loaded by three different commands would have to name one skill's copy, and
`myflow-contracts` has no command of its own to be that answer.

### Which skills get a `scripts/` directory

**ID:** `scripts-dirs-on-command-skills-only`
**Status:** active
**Chosen:** The five command skills — `myflow-do`, `myflow-finish`, `myflow-start`,
`myflow-status`, `myflow-fast`. Each carries symlinks for exactly the guards it can invoke,
plus `lib` where one of those guards sources it.
**Considered:** *Command skills plus `myflow-contracts`* — a fifth symlink farm nothing
resolves against, and a directory nothing reads is a directory that silently goes stale.

### What a missing guard does

**ID:** `loud-once-then-hand-run`
**Status:** active
**Chosen:** Each command checks presence **once at start**, prints one prominent block
naming every missing guard and the command that installs them, then proceeds under the
hand-run fallback the contracts already define.
**Considered:**
- *Refuse to run* — the strongest guarantee, but it contradicts the contracts' existing
  "the check is never skipped for want of the script", which would have to be rewritten in
  every contract carrying it.
- *Refuse only for the finish guards* — targets the KAN-70 blast radius specifically, but
  splits one rule into two behaviours a reader has to keep straight, for a case the loud
  block already makes impossible to miss.

## Open questions

*(none)*

## How it works

### The symlink farm

One real file per guard stays at `scripts/`. Each command skill gets a `scripts/`
directory of relative symlinks:

```
scripts/check-unfinished-work.sh              <- the one real file
scripts/lib/panel-record.sh

skills/myflow-finish/scripts/
    check-unfinished-work.sh -> ../../../scripts/check-unfinished-work.sh
    lib                      -> ../../../scripts/lib
skills/myflow-do/scripts/
    check-unfinished-work.sh -> ../../../scripts/check-unfinished-work.sh
    lib                      -> ../../../scripts/lib
```

`../../../` is counted from the symlink's own directory
(`skills/myflow-finish/scripts/`), so it lands on the repository root.

### Why relative symlinks survive the install

`install_skills()` symlinks the whole skill directory:

```
~/.claude/skills/myflow-finish -> <repo>/skills/myflow-finish
```

Resolving `~/.claude/skills/myflow-finish/scripts/check-unfinished-work.sh` follows the
directory symlink to its real location first, then reads `scripts/check-unfinished-work.sh`
inside it, then follows that symlink's relative target **from the real directory** — landing
on `<repo>/scripts/check-unfinished-work.sh`. The guards' own
`SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"` idiom then resolves
`lib/` beside the invoked path, which is itself a symlink onto `<repo>/scripts/lib`.

Verified empirically against a scratch tree reproducing the installed layout, using the
guards' exact idiom rather than a simplification:

```
SCRIPT_DIR=<scratch>/repo/skills/myflow-finish/scripts
LIB-SOURCED
OK: lib reachable via the real guard idiom
```

The consequence worth stating: an installed skill's guards resolve **into the agents repo
checkout**. That is already true of the skill's own `SKILL.md`, which is the same symlink,
so this adds no new coupling — moving or deleting the checkout already breaks the install.

### Two families of guard, and only one of them ships

Not every `scripts/…` citation in a skill is an invocation. The citations divide cleanly,
and conflating them would ship files nothing resolves against:

- **Pipeline guards** — invoked by the command itself, in a fenced block or an imperative
  "run this". These are the ones that must travel with the skill, because the project being
  worked on has no reason to carry them.
- **Project-configured guards** — named in the project's own `.myflow/project.md` `##
  lint` / `## test` lists and resolved in that project. `check-references.sh`,
  `check-stage-mark-calls.sh`, `check-vocabulary.sh`, `check-contract-budget.sh`,
  `check-plan-provenance.sh` and `check-task-build-green.sh` appear in skill text only as
  **prose about this repository's own guards** — never as something a command runs.

The resolution rule therefore governs **invoked** guards only. Prose describing this
repository's own lint keeps its repo-relative `scripts/…` form, because there it genuinely
names a file in this repository.

`/myflow-start` invokes no guard of its own: its plan-provenance and build-green step reads
"run **the project's configured** plan-provenance guard … if the project declares them," so
those resolve through `.myflow/project.md`. It gets **no** `scripts/` directory.

### The guard-to-skill map

Derived from the invoking call sites in each skill's own text, with each guard's sibling
dependencies resolved from its `$SCRIPT_DIR/…` references.

| Skill | Guards symlinked in |
|---|---|
| `myflow-do` | `check-panel-diff-size.sh`, `check-panel-reproducers.sh`, `check-task-commit-fields.sh` + `.py`, `check-unfinished-work.sh`, `check-workspace-isolation.sh`, `commit-split.sh`, `plan-dispatch-bundles.sh` + `.py`, `prepare-workspace.sh`, `preserve-session-records.sh`, `reproducer-metachars.sh`, `run-reproducer.sh`, `lib` |
| `myflow-finish` | `check-cleanup-complete.sh`, `check-finish-preflight.sh`, `check-unfinished-work.sh`, `commit-split.sh`, `gather-self-review-context.sh`, `preserve-session-records.sh`, `lib` |
| `myflow-status` | `check-finish-preflight.sh` |
| `myflow-fast` | the union of `myflow-do`, `myflow-finish` and `myflow-status`'s sets |
| `myflow-start` | *(none — it invokes no guard of its own)* |

**Sibling dependencies are part of the map, not an afterthought.** Each of these resolves a
neighbour from its own `$SCRIPT_DIR`, so the neighbour must be symlinked beside it or the
guard fails at run time:

| Guard | Needs beside it |
|---|---|
| `check-unfinished-work.sh` | `lib/panel-record.sh` |
| `check-panel-reproducers.sh` | `lib/panel-record.sh`, `reproducer-metachars.sh` |
| `run-reproducer.sh` | `reproducer-metachars.sh` |
| `check-task-commit-fields.sh` | `check-task-commit-fields.py` |
| `plan-dispatch-bundles.sh` | `plan-dispatch-bundles.py` |
| `prepare-workspace.sh` | `check-workspace-isolation.sh` |

### The `$SCRIPT_DIR/..` hazard

**ID:** `repo-root-must-not-be-one-level-up`
**Status:** active

Two shipped guards derive a repository root by going exactly one level up from their own
directory:

```
scripts/plan-dispatch-bundles.sh:59
    REPO_ROOT="${PLAN_DISPATCH_BUNDLES_ROOT:-$(cd "$SCRIPT_DIR/.." && pwd)}"
scripts/check-workspace-isolation.sh:104
    REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
```

That holds only while the guard lives at `<repo>/scripts/`. Invoked through
`skills/myflow-do/scripts/`, `$SCRIPT_DIR/..` is `<repo>/skills/myflow-do` — a skill
directory silently standing in for a repository root.

The blast radius is bounded: both take an explicit target (`plan-dispatch-bundles.sh` a
`tasks.md` path, `check-workspace-isolation.sh` a project root, which is what
`prepare-workspace.sh` passes), and the `..` form is only the no-argument default. But a
default that resolves to a skill directory is worse than no default at all — it produces a
confident wrong answer rather than an error.

**Fix:** both derive the root from the script's **resolved physical** location, so the
answer is the same whichever symlink the guard was invoked through. The new
`check-guard-symlinks.sh` additionally rejects any shipped guard that derives a root as a
fixed number of levels above `$SCRIPT_DIR`, so the pattern cannot come back.

### The resolution rule

`pipeline.md` gains one short section, stated once and cited from everywhere else:

> **A named guard resolves to `<the running command's own skill directory>/scripts/<name>`.**
> Skills and contracts name a guard by basename and never by a repository-relative path — a
> path relative to the repository resolves only when the project being worked on *is* the
> agents repository, which is the one case that is never the interesting one.

Every existing `scripts/<name>.sh` citation in `skills/` is rewritten to the basename form.

### The presence check

Each command skill's first stage gains a presence check over exactly the guards its own
`scripts/` directory should carry. On a complete set it prints nothing. On any absence it
prints one block, once, and continues:

```
⚠ GUARDS MISSING — 3 of 6 not found at
  <skill-dir>/scripts/
    check-finish-preflight.sh
    check-unfinished-work.sh
    preserve-session-records.sh
These checks will be performed BY HAND.
Re-run ./setup.sh global to install them.
```

It is a report, never a gate: the run proceeds under the hand-run fallback each contract
already defines, and the handoff says the checks were run manually — which is what those
contracts already require of a hand-run.

## Testing

- **`scripts/test-setup.sh`** gains the regression case this change exists to prevent:
  install into a sandbox `HOME`, then assert that every guard in the map above is
  **executable through the installed skill path**, and that `check-unfinished-work.sh`
  invoked that way finds its `lib/`. A test that only checks the symlink exists in the repo
  would not have caught the original bug.
- **A new guard, `scripts/check-guard-symlinks.sh`**, with its own
  `scripts/test-check-guard-symlinks.sh` harness, added to `.myflow/project.md`'s `## lint`
  and `## test` lists. It asserts four things:
  1. every `skills/*/scripts/` entry is a symlink that resolves;
  2. every guard **invoked** in a skill's text has a symlink in that skill's own `scripts/`
     directory, sibling dependencies included — the map above, checked rather than trusted;
  3. no skill text carries a repo-relative `scripts/…` path **in an invoking position**
     (prose about this repository's own guards is exempt, per the two families above);
  4. no shipped guard derives a root as a fixed number of levels above `$SCRIPT_DIR`.

  Rules 3 and 4 are what stop the two defects this change fixes from creeping back in.
- **`scripts/check-references.sh`** and **`scripts/check-contract-budget.sh`** must stay
  green. `pipeline.md` grows by the resolution rule, so its budget row is raised
  deliberately, per that guard's own ratchet discipline.

## Out of scope

- Any change to what a guard checks. This change moves guards and fixes how they are
  addressed; it does not alter a single check's logic.
- The panel-record linter, per **Scope** above.
- KAN-76's unchanged-worktree check, which depends on this landing but is its own ticket.
