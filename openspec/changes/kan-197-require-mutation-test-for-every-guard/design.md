# KAN-197 — A corpus-scanning guard reports what it checked

**Date:** 2026-08-18
**Change:** `kan-197-require-mutation-test-for-every-guard`
**Issue:** [KAN-197](https://tweety53.atlassian.net/browse/KAN-197)

## The problem

A guard exits 0 when it finds no violation. That result is produced identically by a guard that
examined the whole corpus and found it clean, and by a guard that examined **nothing at all**. The
two are indistinguishable from the outside, and the second is silent forever.

KAN-73 hit exactly this. `scripts/check-guard-symlinks.sh` rule 2 required every guard a skill
invokes to be symlinked into that skill. For `skills/myflow-fast/` — the skill carrying the **most**
symlinks, 17 of them — the required set computed to empty, because that skill delegates by citation
and names no guard in a form the classifier could see. Deleting one of its symlinks still produced
`GUARD-SYMLINKS-OK`, exit 0.

Three reviewers read that guard. Its own harness passed. The defect was found only because someone
deliberately broke the tree and noticed the guard did not care.

## What this change is not

**Not a rule that harnesses must contain a failure fixture.** That was KAN-197's own cheaper
proposal, and it is vacuous: **all 16** `scripts/test-check-*.sh` harnesses already assert a failure,
and `test-check-guard-symlinks.sh` specifically carried **two** rule-2 violating fixtures
(`:265`, `:286`) while missing `myflow-fast` completely.
<!-- measured: grep over scripts/test-check-*.sh on 2026-08-18 — 16 of 16 assert a non-zero exit or an INVALID verdict; the two rule-2 fixtures read at the cited lines -->

The fixtures covered the guard's **rules**. They did not cover its **input classes** — and the person
choosing the fixtures is the person with the blind spot, so a class nobody imagined stays untested.

**Not a `Mutation:` field in `tasks.md`.** KAN-197's other proposal. It puts per-task friction on
every future guard change to catch a defect that lives in the guard, not in the task, and it would
have to be written by the same author with the same blind spot.

Both are recorded as considered and rejected in **Decisions**, and KAN-197's description is corrected
so the ticket does not contradict the change made against it.

## Decisions

### The mechanism is coverage reporting, not more failure fixtures

**ID:** `coverage-not-fixtures`
**Status:** active
**Chosen:** A corpus-scanning guard reports what it actually checked, per member of the corpus, so
that "nothing was checked here" is visible on a **healthy** tree — no mutation required, and no
reviewer required to imagine the missing case.
**Considered:**
- *A required failure fixture per harness* — already universally satisfied, and demonstrably
  compatible with the defect it is supposed to catch.
- *A `Mutation:` field per task* — friction in the wrong layer, authored by the same blind spot.
- *Mutating the real tree during the run* — this is what actually caught KAN-73, and it works, but
  it is a procedure a human must remember rather than a property the guard has. Coverage reporting
  makes the guard state the fact itself, every run.

### An undeclared zero is a violation; a declared one is not

**ID:** `declared-expected-zeros`
**Status:** active
**Chosen:** Each in-scope guard carries an explicit list of corpus members that legitimately check
nothing. A member reporting zero coverage while absent from that list is a **violation**, named and
non-zero-exit. `skills/myflow-start/` is a legitimate zero — it invokes no guard; `skills/myflow-fast/`
was not, and would have failed.
**Considered:**
- *Report zeros without failing* — cheapest, but it relies on a human reading the line, and three
  reviewers already read this guard without noticing the same fact.
- *Fail only when coverage falls from non-zero to zero* — needs a committed baseline, and would
  **not** have caught KAN-73, where the coverage was zero from the guard's very first commit.

The declaration is the load-bearing part: it converts "I assume this zero is fine" into a statement
someone had to write down, and a reviewer can question.

### Scope is the corpus-scanning guards, and only those

**ID:** `scope-corpus-scanners`
**Status:** active
**Chosen:** `check-guard-symlinks.sh`, `check-references.sh`, `check-vocabulary.sh` and
`check-stage-mark-calls.sh` — the guards that scan a corpus discovered from the tree, where a member
can be silently skipped.
**Considered:** *All 15 `check-*.sh` guards* — ten of them take an explicit single target
(`<worktree> <change-name>`, one project config, one `tasks.md`). They have no member to skip, so
their count would be a constant, and a line that never varies teaches readers to skip it.

## Open questions

*(none)*

## How it works

Each in-scope guard gains two things:

1. **A per-member coverage count in its verdict line**, listing any member whose count is zero.
2. **A declared expected-zero set**, and a violation for any undeclared zero.

`check-guard-symlinks.sh` already prints `GUARD-SYMLINKS-OK: … 53 guard(s) across 6 skill(s)
validated`, which is a corpus-wide total. The change is that the total is broken out per member and
the zeros are named:

```text verified:the guard's actual output against the real tree at df9d5dd, wrapped here for width
GUARD-SYMLINKS-OK: <repo> — 55 guard(s) across 6 skill(s) validated
  myflow-do 13 · myflow-fast 16 · myflow-finish 6 · myflow-start 0 (declared: …) ·
  myflow-status 0 (declared: …) · openspec-explore 0 (declared: …)
```

**Two things in that line differ from what this design first guessed, and both were corrected by
building it.** The counts are the size of rule 2's post-delegation **required set** (13/16/6), not
the raw symlink count (14/17/6) — the required set is what the spec's own wording names, and what a
zero would mean. And the corpus members are not the ones assumed: `openspec-explore` is a member and
needed declaring, while `myflow-contracts` is **not** a member at all, being filtered out of this
guard's corpus by name before coverage is computed. A declaration for a non-member would have been
inert.

and, on the defect KAN-73 shipped:

```text unverified:reconstructs what the guard would have printed at KAN-73's task 5, before the fix round
GUARD-SYMLINKS-INVALID: <repo> — 1 violation(s)
  myflow-fast: 0 guards required, but the skill is not declared as expected-zero (coverage)
```

**The point is the second block appears on a clean tree.** No symlink is missing there; the tree is
healthy. What is wrong is that the guard is not looking, and that is now the thing it says.

## Testing

Each in-scope guard's harness gains two cases: a fixture where a member legitimately reports zero and
is declared (passes), and one where a member reports zero and is **not** declared (fails, naming the
member). The second is the case that would have caught KAN-73.

**And the regression test that closes the loop:** a fixture reproducing KAN-73's own shape — a skill
that delegates by citation and names no guard directly — asserted to fail as an undeclared zero.

## A limitation this design accepts, found by mutation

**Only a collapse all the way to zero is caught.** Stripping most guard citations from
`skills/myflow-do/SKILL.md` took rule 2's required set for that skill from 13 to 5 — a large silent
regression — and the guard reported `GUARD-SYMLINKS-OK`, exit 0, because 5 is not 0.
<!-- measured: mutation run against the worktree at df9d5dd on 2026-08-18; the same file restored immediately after and the guard returned to exit 0 -->

That is correct behaviour for the rule as specified, and the total collapse it *does* catch is the
shape KAN-73 actually shipped. But the partial collapse is the same class of defect and stays
invisible. Catching it needs a committed per-member baseline to compare against — the option
considered and rejected under `declared-expected-zeros`, on the grounds that a baseline would not
have caught KAN-73 at all, since that coverage was zero from the guard's first commit.

Both mechanisms have a blind spot the other covers. This change takes the one that catches the
observed defect; the baseline is worth its own ticket rather than being smuggled in here.

## Out of scope

- Changing what any guard checks. This adds reporting and one new violation class; no existing rule
  is altered.
- A per-member coverage baseline, per the limitation above.
- The ten single-target guards, per `scope-corpus-scanners`.
- KAN-198's per-slot findings telemetry and KAN-199's cross-derivation note, which are separate
  tickets from the same self-review.
