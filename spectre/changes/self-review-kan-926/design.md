## Context

Found by the self-review pass of kan-916 (filed as KAN-926, angle `flow-fix`); planned standalone on
`self-review-kan-926`, based at `origin/main` `b76c2cf1`. Brainstorm design:
`.superpowers/sdd/2026-10-21-self-review-kan-926-design.md` in this worktree.

**Finding — landed drift blocks the main checkout's fast-forward (blast radius 3).**
<!-- measured: Task 1's **Files:** names 3 @ b76c2cf1 -->
Run 2's step 6 (`skills/flow-contracts/finish-contract-run2.md`) brings each main checkout forward
with `refresh-main-checkout.sh`, whose Go guard (`stats/internal/guard/refreshmaincheckout.go`)
refuses any tracked change outright. When run 1's `DRIFT-DIRTY` drift is landed on `origin/<base>`
from another worktree, the main checkout still holds the same bytes uncommitted, and git's
fast-forward refuses to overwrite them; a scratch spike — an unstaged `f` and a staged-added `g`,
both identical to `origin/main` — printed `error: Your local changes to the following files would
be overwritten by merge: f`, and `git restore --source=HEAD --staged --worktree -- d f g` followed
by `git merge --ff-only` left the checkout clean at `origin/main`.
<!-- measured: scratch repo spike, git merge --ff-only rc=1 then restore + merge rc=0 with empty status @ b76c2cf1 -->
Reproduced at `b76c2cf1`: Task 1's test `TestRefreshMainCheckoutDiscardsLandedCopies` fails on the
unchanged guard (`REFRESH-REFUSED ... has tracked changes`) and passes with the fix; the whole
`./internal/guard/` package stays green.
<!-- measured: go test ./internal/guard/ -run TestRefreshMainCheckout -count=1 with and without the refreshmaincheckout.go hunk, and go test ./internal/guard/ -count=1 @ b76c2cf1 -->

Nothing to repair: the main checkout kan-916 left dirty was reset by the operator and
`git status --porcelain --untracked-files=no` there is empty today.

**The main-checkout read-only rule holds.** Only byte-identical copies are discarded, only by the
guard `hooks/protect-main-checkout.py` already admits in a main checkout, and the fast-forward
writes the same bytes back.

The operator asked that this pass plan no new Go guard; this is a change to the existing refresh
guard, the step that lands a main checkout.

## Decisions

### Discard landed copies inside refresh-main-checkout

**ID:** discard-in-refresh
**Status:** active
**Chosen:** extend the existing `refresh-main-checkout` guard — after its existing refusals, a dirty checkout whose every changed path is byte-identical to `origin/<base>` has those paths restored to `HEAD`, then fast-forwards — the one step that already brings the main checkout forward, so every landing route gets it.
**Considered:** a new guard or a step in run 1's drift question — the operator ruled out new Go guards, and run 1 runs before the drift lands; handing back the exact `git restore` command — the same identity check run by hand, one more operator step for no added safety.

### Byte-identity is worktree copy plus staged index copy

**ID:** identity-check
**Status:** active
**Chosen:** compare the worktree copy of every path changed against `HEAD` (worktree or index) with `origin/<base>`, and the index copy of every staged path; refuse on any unmerged entry. An unstaged path's index copy is `HEAD`'s, already committed, so it is not compared.
**Considered:** comparing the index of every path — refuses the kan-916 case itself, whose copies were unstaged; `git status` path parsing — quoting and rename lines, where `diff --name-only -z --no-renames` is exact.

### Restore to HEAD, not stage-then-fast-forward

**ID:** restore-to-head
**Status:** active
**Chosen:** `git --literal-pathspecs restore --source=HEAD --staged --worktree -- <paths>`, then the existing `merge --ff-only` — what the ticket names, one git verb, and the spike showed it handles an unstaged edit, a staged addition and an unstaged deletion alike.
**Considered:** `git add` the copies so the fast-forward keeps them — writes the main checkout's index, against the read-only rule's spirit, for the same end state.

### A new output line, not a changed verdict

**ID:** discarded-line
**Status:** active
**Chosen:** print `REFRESH-DISCARDED: <path> <n> tracked paths identical to origin/<base>` before `REFRESH-DONE`; exit codes unchanged — the handoff and `scripts/integrate-change.sh` keep reading `REFRESH-DONE`/`REFRESH-REFUSED` as before.
**Considered:** folding it into `REFRESH-DONE`'s text — changes a line `scripts/test-integrate-change.sh` matches; staying silent — a discard in a main checkout would go unreported.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — the change deletes uncommitted content in the operator's main checkout; the identity check's edge cases (staged vs unstaged, deletions, additions) deserve a full first look.
**Considered:** `flow-low` — one small, test-pinned task, but a wrong comparison loses operator work.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-kan-926-fix`, Task 1.
**Considered:** none — the plan has one task.

## Open questions
