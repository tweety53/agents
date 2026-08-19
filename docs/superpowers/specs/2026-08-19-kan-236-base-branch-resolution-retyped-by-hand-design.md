# KAN-236 — base-branch resolution becomes a guard

**Date:** 2026-08-19
**Change:** `kan-236-base-branch-resolution-retyped-by-hand`
**Jira:** [KAN-236](https://tweety53.atlassian.net/browse/KAN-236) — self-review finding, angle 4
(`myflow-automation`), from KAN-228.

## Problem

`check-finish-preflight.sh <worktree> <base-ref> <recorded-merge-base|->` takes the base ref as an
argument. The block that **produces** that ref exists only as a fenced `bash` block inside
`skills/myflow-contracts/finish-contract.md` — fetch, `refs/remotes/origin/HEAD`, the
`git remote show origin` fallback, and three assertions: non-detached `HEAD`, non-empty base, and
base ≠ current branch. Every `/myflow-finish` run retypes it.

The contract's own comments say what a wrong answer costs. Resolving via `HEAD@{upstream}` compares
`openspec/<name>` against `origin/openspec/<name>`, which is true the moment the branch is pushed,
so an unmerged change is reported merged — after which run 2 archives it and `--force`-removes its
worktree. That is the most destructive step in the pipeline, gated by a block that is copied by
hand.

**A second call site the ticket does not name.** `skills/myflow-status/SKILL.md`'s merge-status step
3 runs `git merge-base --is-ancestor <branch> origin/<base>` and states no source for `<base>` at
all. The same value is hand-derived there, with no assertion that the base differs from the branch
under test, in a command the operator runs far more often than `/myflow-finish`.

## Approaches considered

1. **A standalone guard script** every call site invokes. **Chosen.** The base *name* — not merely a
   preflight verdict — is needed at four points: the preflight's `<base-ref>` argument, run 1's
   merge-and-push route, run 2's archive commit, and `/myflow-status`'s ancestor test. One script
   serves all four, and the resolution lives in exactly one file.
2. **`check-finish-preflight.sh` resolves the base itself when passed `-`**, mirroring the `-` it
   already accepts for the recorded merge base. Rejected: it fixes only the preflight call site.
   Run 1 still needs `BASE` in a shell variable to merge into and push, so the fenced block survives
   anyway — and `check-finish-preflight.sh`'s own header (lines 28-30) argues against carrying a
   second copy of the resolution inside itself, for the drift reason this change exists to remove.
3. **A `myflow base-branch` subcommand** in Go, matching the shape KAN-224 proposes for the
   workspace id. Rejected: this is a pure-git operation with no store involvement, every other guard
   in the pipeline is reached by shell, and a Go subcommand installs on a different path than the
   guard symlinks `setup.sh` already manages.

## The resolver

`scripts/resolve-base-branch.sh <dir>` — one required argument, the directory to resolve in.

**Output.** On success, stdout carries exactly the base branch name and nothing else (`main`), so a
caller composes it directly: `BASE="$(resolve-base-branch.sh "$WT")"`. Every refusal writes its
reason to stderr and **nothing to stdout** — a caller must never read empty output as a resolved
base.

| Exit | Meaning |
|------|---------|
| `0` | Resolved; the name is on stdout |
| `1` | A named refusal — detached `HEAD`, no base resolved, or base equal to the current branch |
| `2` | Cannot answer — `<dir>` is missing, unreadable, or not a git worktree |
| `3` | The repository has no `origin` remote at all |

**Exit 3 is separate from exit 2 on purpose.** The finish contract requires a distinct operator
message for a repository with no remote — *"this repository has no remote, so there is nothing to
push to or merge into"* — rather than a base-branch failure, which sends the operator debugging the
wrong thing. A caller can only print that message if it can tell the two apart, and matching on
stderr text would be fragile.

**Order of operations.**

1. `<dir>` is a readable directory and `git -C "$dir" rev-parse --git-dir` succeeds → else exit 2.
2. `git -C "$dir" remote get-url origin` succeeds → else exit 3.
3. `git -C "$dir" -c core.askpass=true fetch --quiet origin 2>/dev/null || true` — a stale ref only
   fails safe, and the wrapping is what keeps an unreachable host from turning a correct refusal
   into a ~75s hang on the default TCP timeout.
4. `BASE` from `git symbolic-ref --quiet --short refs/remotes/origin/HEAD`, stripped of its
   `origin/` prefix; when empty, from `GIT_TERMINAL_PROMPT=0 git remote show origin` via
   `sed -n 's/^ *HEAD branch: //p'`.
5. `CUR="$(git -C "$dir" branch --show-current)"`; empty → exit 1, detached `HEAD`.
6. `BASE` empty → exit 1, no base resolved.
7. `BASE` = `CUR` → exit 1, refusing to compare a branch with itself.

**Steps 5-7 keep the block's own order**, so each refusal reports the same failure the fenced block
reported, in the same sequence.

**`BASE` is validated before it is printed.** It arrives from the remote, and it flows into refs
callers build (`origin/$BASE`) and into git commands. It must match `^[A-Za-z0-9._][A-Za-z0-9._/-]*$`
— rejecting an empty value, a leading `-` that a downstream git call would read as an option, and
any control character — or the resolver refuses with exit 1 rather than printing it.

**The `base ≠ current` assertion is unconditional, and there is no flag to skip it.** It is what
makes the `HEAD@{upstream}` class of misresolution impossible rather than merely unlikely, and the
one caller that would want it off is the one that most needs it on.

**Every caller therefore passes the apply worktree**, where `HEAD` is `openspec/<name>` by
construction, never the main checkout. This matters at two points that stand on the base branch:

- `/myflow-finish` run 2 commits the archive on the base branch in the main checkout. It resolves
  the base **from the apply worktree** — which is still present, because worktree cleanup is run 2's
  step 4 and the archive commit is step 3.
- `/myflow-status` already iterates the resolved worktree set and answers merge status once per
  worktree; it resolves the base in each worktree it is already visiting.

## Call sites

### `skills/myflow-contracts/finish-contract.md`

The fenced block under **Run 1 — the branch is not merged** is replaced by an invocation:

```bash
BASE="$(resolve-base-branch.sh "<abs-worktree>")"
```

with the exit table above stated beside it. **The prose stays.** The `HEAD@{upstream}` paragraph and
the no-remote paragraph are the *reasoning*, not the copied code, and they are what a reader needs
in order to not reintroduce the block. Both are rewritten to name the guard as the place the rule is
enforced.

The guard is named by **basename** in every invoking position, per **Guard resolution**
(`skills/myflow-contracts/pipeline.md`) — never a repository-relative `scripts/…` path, which
`check-guard-symlinks.sh`'s rule 3 rejects.

### `scripts/check-finish-preflight.sh`

Its signature does not change; it keeps taking `<base-ref>` as an argument, and it still resolves
nothing itself. Only its header comment changes: lines 28-30 currently argue that base-branch
resolution "deliberately stays in pipeline.md's Finish contract and is passed in." That is now
wrong about *where*, and right about *why*. Rewritten, it points at `resolve-base-branch.sh` as the
one owner and keeps the reason — a second copy could drift from it.

### `skills/myflow-finish/SKILL.md`

Its **Check guard presence.** paragraph gains `resolve-base-branch.sh`. The paragraph at line 78 —
*"The base branch is resolved, never assumed, and never derived from the current branch"* — keeps
its wording and points at the guard.

### `skills/myflow-status/SKILL.md`

Merge-status step 3 resolves `<base>` through the guard, in the worktree it is already visiting.

**A non-zero exit maps to inconclusive, never to a stop.** `/myflow-status` is read-only and never
blocks, and its step 3 already treats a git failure or an unresolvable base ref as inconclusive. A
refusal from the resolver is the same class of answer, and its stderr message is what the detail
view reports.

**Three invariant reversals, all in one direction: `/myflow-status` acquires its first guard.**

1. Its "No guard-presence check here — this command invokes no guard" block becomes a real presence
   check naming `resolve-base-branch.sh`.
2. Its "This command therefore carries no `scripts/` directory" claim is replaced; the directory is
   created with the one symlink.
3. `scripts/check-guard-symlinks.sh`'s `declare_if_present "myflow-status" "invokes no guard — …"`
   call is removed. That declaration exists to keep a legitimate zero from reading as an
   uncomputed empty set; once the skill ships a guard, rule 2 computes a non-empty required set for
   it and the declaration would be false.

### `setup.sh` and the symlink set

The guard is symlinked, relative, from `scripts/resolve-base-branch.sh` into
`skills/myflow-finish/scripts/`, `skills/myflow-status/scripts/` (new directory) and
`skills/myflow-fast/scripts/` — the last because `/myflow-fast` carries the union of the three
commands' guards, which its own delegating presence paragraph already declares.

## Testing

`scripts/test-resolve-base-branch.sh`, in the house style of
`scripts/test-check-finish-preflight.sh`: throwaway git repositories under `TMPDIR`, one assertion
per case, a `resolve-base-branch: all cases pass` line at the end. Added to `.myflow/project.md`'s
`## test` list.

**Every case mutation-proves the behaviour it covers**, per KAN-197 — a case whose assertion still
passes with the line it covers removed proves nothing. Cases:

| Case | Expects |
|------|---------|
| Clone with `origin/HEAD` set, branch checked out | stdout `main`, exit 0 |
| `origin/HEAD` absent, `git remote show origin` answers | stdout `main`, exit 0 (fallback path) |
| Detached `HEAD` | exit 1, stderr names detached `HEAD`, stdout empty |
| Base resolves to the checked-out branch | exit 1, stderr names the self-comparison, stdout empty |
| Neither resolution path answers | exit 1, stderr names the unresolved base, stdout empty |
| No `origin` remote configured | exit 3, stderr names the missing remote |
| `<dir>` is not a git worktree | exit 2 |
| `<dir>` does not exist | exit 2 |
| Missing argument | exit 2, usage line on stderr |
| Remote `HEAD branch:` naming `-x` or a control character | exit 1, nothing on stdout |

The unreachable-remote hang is **not** covered by a timing assertion — a wall-clock test of a TCP
timeout is flaky by construction. The wrapping is carried forward verbatim from the block, and its
reason is recorded in the guard's header.

## Files touched

| Path | Change |
|------|--------|
| `scripts/resolve-base-branch.sh` | new guard |
| `scripts/test-resolve-base-branch.sh` | new test harness |
| `skills/myflow-finish/scripts/resolve-base-branch.sh` | new relative symlink |
| `skills/myflow-status/scripts/resolve-base-branch.sh` | new directory, new relative symlink |
| `skills/myflow-fast/scripts/resolve-base-branch.sh` | new relative symlink |
| `skills/myflow-contracts/finish-contract.md` | block → invocation; prose kept and re-pointed |
| `skills/myflow-finish/SKILL.md` | presence paragraph; base-resolution paragraph re-pointed |
| `skills/myflow-status/SKILL.md` | step 3 invokes the guard; three no-guard claims reversed |
| `scripts/check-finish-preflight.sh` | header comment lines 28-30 re-pointed |
| `scripts/check-guard-symlinks.sh` | `declare_if_present "myflow-status"` removed |
| `.myflow/project.md` | `## test` gains `scripts/test-resolve-base-branch.sh` |

## Out of scope

- **KAN-224** (`myflow workspace-id`) is the same shape of finding for the workspace id and stays
  its own change. Nothing here presumes its route.
- **`check-finish-preflight.sh`'s signature.** It keeps taking `<base-ref>`; adding a `-`
  self-resolution mode was considered and rejected above, and adding it *alongside* the guard would
  ship two mechanisms for one job.
