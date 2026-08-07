# myflow-finish — rationale

This file is the reasoning behind `skills/myflow-finish/SKILL.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

## State gate

## Deciding which run this is

A PR a human merged on the forge, a colleague's merge, and run 1's own merge are indistinguishable
here, and that is correct — all three mean the same thing.

# Run 1 — integrate

## 1.0 Check for unfinished work

The signals that make a change outstanding are the script's own and are not restated here — a second
list of them would drift from the one that is actually run.

Asking the landing question first and only then reporting unfinished work would make the operator
choose a route for a branch they have not yet been told is incomplete.

## 1.1 Ask how the branch should land, before any git action

## 1.2 Commit the staged work

**The `reset` is what makes the split hold; the exclusion alone only assumes it** — the reason is
stated under **Git boundaries** (`skills/myflow-contracts/pipeline.md`), which `/myflow-do` follows
too. What is specific to this gate is *whose* staging it retracts: the operator may have run their
own `git add -A` while reviewing, and the excluding `add` cannot take those paths back out.

`scripts/preserve-session-records.sh` still runs **before** the first `add`, unchanged:
`docs/superpowers/` is one of the excluded paths, so its files are picked up by the second staging
pass. The second `add` carries no pathspec, which is what makes it pick them up.

## 1.3 Take the chosen route

## 1.4 No verification gate

## 1.5 State and handoff

The block below is **not** a second definition of the handoff. It is this run's rendering of the
`IN_PROGRESS`-after-run-1 template, which is defined once under
**The block each state renders** (`skills/myflow-contracts/pipeline.md`) and is canonical for the
labels, the field set and their order. What this block adds is the enumeration of the literal
alternatives run 1 writes — the three `PR:` cases below are that enumeration of the template's
placeholder, one per landing route. **Change the template first and bring this block with it** — a
field added here and not there is drift the moment `/myflow-status <name>` regenerates the same
state.

**Outstanding is the same list the planning commit carries**, and it is stated in both places
because either alone is a record the next reader may never reach: a commit message nobody reads
back, or a handoff that scrolls out of a session nobody kept.

The last line is this same command, because that is what the operator runs once the branch is
merged.

# Run 2 — archive and clean up

   The inner `{ commit && push; }` grouping matters: without it, `&&`/`||` at equal precedence
   parse left-to-right as `(diff --quiet || commit) && push`, which runs `push` unconditionally
   once the outer group is entered — even when nothing was staged or committed. Grouping `commit`
   and `push` together means `push` only runs on the branch where `commit` actually ran, matching
   the skip-if-empty shape **Git boundaries** (`skills/myflow-contracts/pipeline.md`) already
   documents for the worktree commits.

## Guardrails
