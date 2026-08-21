# Design — commit scopes name the module

Source: `docs/superpowers/specs/2026-08-21-kan-202-commit-split-and-module-scopes-design.md`, which
carries the full problem statement and the evidence behind it. This file states the design as the
implementation needs it, plus the decisions and open questions the brainstorming stage produced.

## The shape of the change

Five edits, in three layers.

### Layer 1 — what the pipeline tells an agent to write

| File | Edit |
|------|------|
| `skills/myflow-start/SKILL.md` | The `**Commit:**` entry in the mechanically-checkable field list gains the scope rule: the scope names the module or area, derived from the paths in that task's own `**Files:**` field; never the change name, never a task id, never a list. |
| `skills/myflow-do/SKILL.md` | The COMMIT-PER-TASK block drops "where that convention has a scope, `<n>` is the scope", together with the `fix(<n>)` / `feat(<n>)` / `task(<n>)` examples built on it. The `Task-Id: <n>` trailer keeps its job; the subject follows the plan's declared `**Commit:**` field. |

### Layer 2 — the messages finish run 1 writes

Four sites state one or both of these messages. All four change together, since a message stated
two ways is a message that will be transcribed the wrong way.

| Site | Today | After |
|------|-------|-------|
| `skills/myflow-contracts/pipeline.md` — Git boundaries chain | `"<type>(<name>): <what the implementation does>"` / `"chore(<name>): plan and session records"` | `"<type>(<module>): <what the implementation does>"` / `"chore(openspec): plan and session records"` |
| `skills/myflow-contracts/finish-contract.md` — two-commit section | states the chain in prose, names no script | same messages, **and names `commit-split.sh`** |
| `skills/myflow-finish/SKILL.md:220` | `commit-split.sh <worktree> <name> "<type>(<name>): …" "chore(<name>): plan and session records"` | `… "<type>(<module>): …" "chore(openspec): plan and session records"` |
| `skills/myflow-do/SKILL.md:1041` | `commit-split.sh <worktree> <name> "<impl-msg>" "chore(<name>): plan and session records"` | `… "<impl-msg>" "chore(openspec): plan and session records"`, and `<impl-msg>`'s derivation drops `fix(<name>)` for a module scope |

`commit-split.sh` itself needs **no** change: it takes both messages as arguments and its `<name>`
parameter is documented as accepted-and-unused. Only the callers change.

### Layer 3 — enforcement and reach

**The guard.** `scripts/check-task-commit-fields.py` gains `check_commit_scope(task, change_name)`,
called from `check_task_commit` beside the existing `check_commit_subject`. It parses the scope out
of the **declared** `**Commit:**` field — not the real commit's subject, which the existing check
already compares against it — and fails when that scope is:

* the change name;
* the change name's leading Jira key (`kan-202` from `kan-202-commit-split-and-module-scopes`); or
* a dotted or numeric task id (`3`, `3.2`, `12.4.1`).

A `**Commit:**` field with no scope passes. So does any other scope: there is no vocabulary of legal
module names to maintain, and introducing one would be a list to keep in sync with the tree.

The change name comes from `tasks_md_path`, which the wrapper has already resolved to
`<worktree>/openspec/changes/<name>/tasks.md` — the basename of that path's parent directory. No new
argument.

**The rule.** `rules/commit-scope-is-the-module.mdc`, `alwaysApply: true`, with `<!-- core -->` /
`<!-- /core -->` markers around the part that belongs in every session's prompt. `setup.sh`'s
`always_on_rules()` discovers it from the frontmatter and `render_managed_block` extracts the core,
so the installer needs no edit. `rules/agent-baseline.md` gains a row in its rule table.

## Decisions

### How to handle KAN-202, whose stated symptom turned out to be stale

**ID:** `kan-202-close-and-fold`
**Status:** active
**Chosen:** Leave KAN-202 linked to this change and fold its two-line fix in — name `commit-split.sh`
in `finish-contract.md`'s two-commit section — letting `/myflow-finish` transition it to Done when
the edit lands. The ticket claims `/myflow-fast` hand-writes the guarded `&&` chain because it cites
`finish-contract.md`; it does not, and did not on the day it was filed — `skills/myflow-fast/SKILL.md`
routes run 1 through `skills/myflow-finish/SKILL.md` §1.2/§1.3, which calls the script. Verified
against the file at `05bac7a`.
**Considered:**
- *Close it and skip the edit* — accurate about the bug, but leaves the canonical statement of the
  two-commit step describing a chain and naming no implementation of it, in a file this change is
  editing anyway.
- *Keep it open and correct its description* — ticket surgery for a documentation-only residual with
  no live route to the failure; the comment recording the staleness carries the same information at
  a fraction of the cost.

### What scope finish run 1's two commits carry

**ID:** `finish-commit-scopes`
**Status:** active
**Chosen:** The implementation commit derives its scope from the reshaped diff; the planning commit
becomes the fixed literal `chore(openspec): plan and session records`. These are the two messages
that actually reach `main` — run 1's `git reset --soft <merge-base>` collapses every per-task commit
before they are written.
**Considered:**
- *Keep `chore(<name>)` for the planning commit* — would preserve the change name once per change in
  the history, but the scope still names no module, which is the whole defect. The branch and the
  ticket already carry the change name.
- *Fixed literals for both* — simpler, but the implementation commit is the one carrying the work,
  and a fixed scope there would stop the subject naming where the change landed.

### How the module-scope rule is enforced

**ID:** `scope-enforcement`
**Status:** active
**Chosen:** Extend `check-task-commit-fields.py`, which already parses `**Commit:**` and already runs
once per task commit during `/myflow-do`. The regression is caught mechanically at a call site that
exists, and the repository's convention that every guard carries a mutation test applies to the new
check.
**Considered:**
- *Documentation only* — the sites all say the right thing and nothing stops the next plan declaring
  a change-name scope; this is precisely how the current state arose.
- *A check in the plan-provenance guard, at `/myflow-start`* — earlier feedback, but a second guard
  reading `tasks.md` for a field the first guard already parses, and provenance is a different
  concern.

### Whether a scope is mandatory

**ID:** `scope-optional`
**Status:** active
**Chosen:** Optional. A `**Commit:**` field declaring `fix: <subject>` passes; only a present-and-wrong
scope fails. This matches what the repository already does — `e61603b fix: drop an invented spec
citation` carries no scope.
**Considered:**
- *Mandatory for every task commit* — more uniform, but forces a scope onto a genuinely cross-cutting
  task, where the honest answer is that no single module carries it. A required field answered
  dishonestly is worse than an absent one.

### Whether the convention reaches beyond the pipeline

**ID:** `global-rule`
**Status:** active
**Chosen:** Add `rules/commit-scope-is-the-module.mdc` as an always-on rule, plus a row in
`rules/agent-baseline.md`. The convention is about commits, not about myflow; a commit made outside a
`/myflow-*` run is subject to it in exactly the same way.
**Considered:**
- *State it in the myflow files only* — smaller blast radius, but leaves the convention unstated for
  every commit the pipeline does not produce, which is most of them in most repositories.

### Whether prose alone should guard the principles-path resolution

**ID:** `top-level-skill-symlink-rule`
**Status:** active
**Chosen:** Add a fifth rule to `check-guard-symlinks.sh` rejecting any symlink placed directly under
`skills/<skill>/`. `skills/myflow-do/SKILL.md` resolves `[PRINCIPLES_PATH]` **beside itself**, always
— including when `/myflow-fast` runs that section — and commit `d149107` added a sentence naming
"symlinking one in" as the wrong fix. A session created three such symlinks in `skills/myflow-fast/`
at 23:58 on 2026-08-20, roughly five hours later. Rule 1 validates entries *under*
`skills/*/scripts/` and reaches nothing at the top level, so nothing mechanical existed. The three
symlinks were untracked, existed on one machine only, and have been removed.
**Considered:**
- *Leave prose as the countermeasure* — it is already written, and it already failed once within
  hours of being written. A rule that a reader must reach before the file-not-found error is a rule
  most readers meet second.
- *File it as its own ticket* — accurate scoping, but the guard and its harness are a self-contained
  pair, this change is already editing the repository's own guard layer, and the operator chose to
  fold it in.

## Open questions

*(none — the brainstorming stage ended with nothing held)*

## What is deliberately out of scope

- **The finish commits stay unguarded.** Their messages are derived at integration time from a diff,
  not declared in a file a guard reads. Guarding them would mean re-deriving the module inside a
  guard — the same judgment the agent just made, duplicated.
- **History is not rewritten**, and archived plans under `openspec/changes/archive/` are not repaired.
  The change governs plans written from here on.
- **No vocabulary of legal module names** is introduced. The guard rejects three wrong shapes and
  accepts everything else.
- **Rule 5 forbids the placement, not a particular target.** Any symlink directly under
  `skills/<skill>/` is reported, rather than only one pointing at a sibling skill. The repository has
  no such symlink today, so the broad form has no false positives, and a narrow one would be a
  target-matching rule to keep in sync.
