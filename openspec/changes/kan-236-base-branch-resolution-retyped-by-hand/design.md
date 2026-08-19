# Design — base-branch resolution becomes a guard

Source: `docs/superpowers/specs/2026-08-19-kan-236-base-branch-resolution-retyped-by-hand-design.md`,
approved during this change's brainstorming.

## Context

`check-finish-preflight.sh <worktree> <base-ref> <recorded-merge-base|->` takes the base ref as an
argument. The block that **produces** that ref lives only in
`skills/myflow-contracts/finish-contract.md`, as a fenced `bash` block a run is expected to retype:
a wrapped `git fetch`, `git symbolic-ref refs/remotes/origin/HEAD`, a `git remote show origin`
fallback, and three assertions — non-detached `HEAD`, non-empty base, base ≠ current branch.

The cost of a wrong answer is recorded in the contract itself. Resolving via `HEAD@{upstream}` from
inside the apply worktree compares `openspec/<name>` against `origin/openspec/<name>`, which is true
the moment the branch is pushed. An unmerged change is then reported merged, and run 2 archives it
and `--force`-removes the worktree holding all of it.

**A second call site the ticket does not name.** `skills/myflow-status/SKILL.md`'s merge-status step
3 runs `git merge-base --is-ancestor <branch> origin/<base>` and states no source for `<base>`. There
the assertion is not retyped — it is absent, in the command an operator runs most often.

## Goals

- One implementation of the resolution, invoked rather than retyped.
- All three assertions survive, and the third one — base ≠ current branch — is not skippable.
- The finish contract keeps the *reasoning* that stops a reader reintroducing the block.
- `/myflow-status`'s hand-derived base is closed by the same guard, without giving a read-only
  command a way to block.

## Non-Goals

- **KAN-224** (`myflow workspace-id`) — the same shape of finding for the workspace id, and its own
  change. Nothing here presumes its route.
- **Changing `check-finish-preflight.sh`'s signature.** It keeps taking `<base-ref>`.
- **A timing test for the unreachable-remote hang.** A wall-clock assertion on a TCP timeout is
  flaky by construction; the wrapping is carried forward verbatim and its reason recorded in the
  guard's header.

## The resolver

`scripts/resolve-base-branch.sh <dir>` — one required argument, the directory to resolve in.

stdout carries the bare branch name and nothing else, so a caller composes it directly:
`BASE="$(resolve-base-branch.sh "$WT")"`. Every refusal writes its reason to stderr and leaves
stdout empty, so no caller can read empty output as a resolved base.

| Exit | Meaning |
|------|---------|
| `0` | Resolved; the name is on stdout |
| `1` | A named refusal — detached `HEAD`, no base resolved, base equal to the current branch, or a base name that fails validation |
| `2` | Cannot answer — the argument is missing, `<dir>` is absent, unreadable, or not a git worktree, or `HEAD`'s own ref cannot be read |
| `3` | The repository has no `origin` remote at all |

Order of operations:

1. `<dir>` is a readable directory and `git -C "$dir" rev-parse --git-dir` succeeds → else exit 2.
2. `git -C "$dir" remote get-url origin` succeeds → else exit 3.
3. `git -C "$dir" -c core.askpass=true fetch --quiet origin 2>/dev/null || true`.
4. `BASE` from `git symbolic-ref --quiet --short refs/remotes/origin/HEAD`, `origin/` stripped; when
   empty, from `GIT_TERMINAL_PROMPT=0 git remote show origin | sed -n 's/^ *HEAD branch: //p'`.
5. `CUR="$(git -C "$dir" branch --show-current)"`; the command failing → exit 2, `HEAD`'s own ref
   cannot be read, which is a different fact from a detached `HEAD` and is not reported as one;
   succeeding with empty output → exit 1, detached `HEAD`.
6. `BASE` empty → exit 1, no base resolved.
7. `BASE` = `CUR` → exit 1, refusing to compare a branch with itself.
8. `BASE` matches `^[A-Za-z0-9._][A-Za-z0-9._/-]*$` → else exit 1.

Steps 5-7 keep the fenced block's own order, so each refusal reports the same failure in the same
sequence a reader of the old block would expect.

## Call sites

| Call site | Today | After |
|-----------|-------|-------|
| Preflight's `<base-ref>` | retyped block | `BASE="$(resolve-base-branch.sh "$WT")"` |
| Run 1 merge-and-push | retyped block | same `BASE` |
| Run 2 archive commit | retyped block | resolved from the apply worktree, before cleanup |
| Run 2 cleanup check | retyped block | resolved per worktree, inside the cleanup loop |
| `/myflow-status` step 3 | no stated source | same guard, per worktree |

## Decisions

### Ship the resolution as a standalone guard

**ID:** standalone-guard
**Status:** active
**Chosen:** A new `scripts/resolve-base-branch.sh` every call site invokes — the base *name*, not
merely a preflight verdict, is needed at four points, and one script serves all four with the
resolution in exactly one file.
**Considered:**
- *`check-finish-preflight.sh` resolves the base itself when passed `-`*, mirroring the `-` it
  already accepts for the recorded merge base — ruled out because it fixes only the preflight call
  site: run 1 still needs `BASE` in a shell variable to merge into and push, so the fenced block
  survives anyway, and the script's own header (lines 28-30) argues against a second copy of the
  resolution inside itself.
- *A `myflow base-branch` Go subcommand*, matching the shape KAN-224 proposes for the workspace id —
  ruled out because this is a pure-git operation with no store involvement, every other guard is
  reached by shell, and a subcommand installs on a different path than the guard symlinks
  `setup.sh` already manages.

### stdout is the bare name, not a verdict line

**ID:** bare-name-stdout
**Status:** active
**Chosen:** stdout carries the branch name alone under a four-value exit contract, so callers use it
in a command substitution with no parsing.
**Considered:** *A `BASE: main` / `REFUSE: <reason>` verdict line*, matching
`check-finish-preflight.sh`'s shape — ruled out because every caller would then strip a prefix
before using the value, which is precisely the hand-work this change removes. The two guards answer
different kinds of question: one returns a verdict, one returns a value.

### Exit 3 is separate from exit 2

**ID:** no-remote-own-exit
**Status:** active
**Chosen:** A repository with no `origin` gets its own exit code, so a caller can print the finish
contract's distinct no-remote message.
**Considered:** *Folding it into exit 2 and matching the stderr text* — ruled out as fragile: the
contract requires that message specifically because a base-branch failure "sends the operator
debugging the wrong thing", and a caller that can only tell by string-matching will stop being able
to the first time the wording is edited.

### The `base ≠ current` assertion stays unconditional, and callers pass the worktree

**ID:** assert-always-pass-worktree
**Status:** active
**Chosen:** No flag skips the third assertion; every consumer passes the apply worktree, where
`HEAD` is `openspec/<name>` by construction. `/myflow-finish` run 2 resolves the base before its
step-4 cleanup removes that worktree, and `/myflow-status` already iterates the resolved worktree
set.
**Considered:**
- *An opt-out flag for callers that only need the name* — ruled out because the two callers that
  would reach for it are the two standing on the base branch, and the mode that disables a tripwire
  is the one a careless caller picks.
- *Dropping the assertion, since the implementation never reads `HEAD@{upstream}`* — ruled out
  because the ticket requires all three survive, and a tripwire's value is catching the resolution
  path nobody predicted.

### The fetch stays inside the guard, unconditional

**ID:** fetch-inside
**Status:** active
**Chosen:** The wrapped `git fetch … || true` is part of the block being consolidated, and a stale
`origin/HEAD` is exactly what makes resolution wrong. `/myflow-status` inherits one network round
trip per worktree, bounded by the same wrapper that keeps an unreachable host from hanging ~75s.
**Considered:**
- *Moving the fetch to the callers* — ruled out: it splits the block back into two pieces, one of
  which is again prose the finish steps retype.
- *An environment variable to skip it* — ruled out as a mode nobody asked for, and a second code
  path the mutation test would have to cover.

### `/myflow-status` gains its first guard, and the invariants that deny it are corrected

**ID:** status-acquires-guard
**Status:** active
**Chosen:** Close the defect where it is worse. `/myflow-status` gets a `scripts/` directory with
the one symlink, a real **Check guard presence.** paragraph, and `check-guard-symlinks.sh`'s
`declare_if_present "myflow-status"` removed — that declaration exists to keep a legitimate zero
from reading as an uncomputed empty set, and it becomes false the moment the required set is
non-empty.
**Considered:** *Scoping the change to `/myflow-finish` and leaving status's prose alone* — ruled
out because it would ship a guard for the value status hand-derives while leaving status
hand-deriving it, with no assertion at all that base ≠ the branch under test.

### A resolver refusal makes status inconclusive, never a stop

**ID:** status-refusal-inconclusive
**Status:** active
**Chosen:** A non-zero exit maps to **inconclusive**, the same disposition status already gives a
git failure or an unresolvable base ref, and the refusal's message is what the detail view reports.
**Considered:** *Treating a refusal as a stop, the way `/myflow-finish` treats `REFUSE`* — ruled out
because `/myflow-status` is read-only and never blocks; a report that refuses to render is worse
than one that says it could not tell.

## Risks / Trade-offs

- **`/myflow-status` gains network I/O.** One wrapped fetch per worktree. Accepted: the wrapper
  bounds it, and the alternative is a status report reading stale refs — which is a wrong answer
  rather than a slow one.
- **A three-way exit contract is more surface than the two guards beside it carry.** Accepted: each
  value is separately mutation-tested, and the alternative is callers string-matching stderr.
- **Reversing `/myflow-status`'s "invokes no guard" invariant touches three places** that currently
  assert it. All three are enumerated in the plan, and `check-guard-symlinks.sh`'s own coverage
  report is what proves the reversal took.

## Migration Plan

None. The guard is additive; no operator-visible behaviour changes on a correct run — the same base
branch resolves and the same refusals are reported in the same order. Existing changes at any
pipeline state are unaffected.

## Open Questions

*(none)*
