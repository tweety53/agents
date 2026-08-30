## Context

`skills/flow/integrate.md`'s step 2 already detects when the base branch has moved and overlaps
this change's own touched paths (`check-base-moved.sh`), and stops to ask the operator. Today the
only choices are "stop, rebase by hand" or "continue anyway" — there is no in-pipeline rebase.
Observed live integrating KAN-304: `main` moved 6, then 9, commits ahead of the recorded merge
base, overlapping `skills/flow/review-panel.md` (touched by both changes), and the operator had to
stop and rebase by hand on every re-invocation.

## How

### A third option on the existing overlap prompt

`skills/flow/integrate.md`'s step 2, on an overlap:

```markdown unverified:this exact shape confirmed against the existing two-option prompt when task 1 writes it
**The base branch has moved and touches paths this change also touched — how should
integration proceed?**
- **Stop — I'll rebase or reorder first** *(recommended, unchanged)*
- **Rebase onto `<base>` now, then continue** *(new)*
- **Continue — land anyway** *(unchanged)*
```

### On the rebase choice

```bash unverified:this is a standard git rebase invocation, not yet run against a live overlap by this change
git -C <worktree> rebase origin/$BASE
```

- **Clean** — the merge base recorded for the rest of this run becomes `origin/$BASE`'s resolved
  tip at rebase time. Re-run `check-base-moved.sh` once more against the new merge base (a
  concurrent push during the rebase is rare but possible); a fresh `MOVED` overlap re-offers this
  same prompt rather than looping silently. Otherwise, proceed to re-verification below, then the
  landing question.
- **Conflict** — never auto-abort. Leave the worktree mid-rebase, report the conflicting file(s)
  from `git status`, and hand off `git -C <worktree> rebase --continue` (after resolving) or
  `git -C <worktree> rebase --abort` as the operator's next manual step. State stays `IN_PROGRESS`.

### Scoped re-verification, not the full lint/test list

`integrate.md`'s "no verification gate" rule holds because correctness was already established by
the review panel and the human gate over the diff as staged — a rebase onto unrelated upstream
commits does not touch anything that gate covered. That assumption breaks exactly when the rebase's
incoming commits touch a file this change's own diff also touches — the same condition
`check-base-moved.sh` already computes as its `overlaps:` list.

Re-verification therefore runs only over that overlap list, never the project's whole `## lint`/
`## test` list: for each overlapping path, re-run the discoverable guard test that covers it —
`scripts/test-<basename-without-ext>.sh` beside `scripts/<name>.sh`, the same naming convention
`scripts/run-guard-tests.sh` already discovers by glob. A path with no such discoverable test (a
prose-only skill file, say) has no verification to run for it; the handoff states that plainly
rather than skipping silently. A non-zero exit blocks the same way any other verify-stage failure
blocks: report it, leave the change `IN_PROGRESS`, and stop before the landing question.

## Decisions

### A confirm before rebasing, not an automatic rebase

**ID:** rebase-is-a-confirmed-choice
**Status:** active
**Chosen:** The rebase runs only after the operator picks it from the three-option prompt.
**Considered:** Rebase automatically whenever an overlap is detected, no prompt — rejected: a
rebase is a history-rewriting operation, and a conflict leaves the worktree in a state the operator
must resolve by hand regardless, so skipping the confirm buys nothing but removes the operator's
chance to choose Stop or Continue instead.

### Scoped re-verification over the overlap set, not the full test suite

**ID:** scoped-reverify-not-full-suite
**Status:** active
**Chosen:** Re-run only the discoverable test(s) for the file(s) `check-base-moved.sh` reported as
overlapping.
**Considered:** Re-run the project's whole `## lint`/`## test` list after every rebase — rejected:
most rebases in this repository's own history carry no real content overlap (two changes touching
different sections of the same long contract file, say), and paying the full suite's cost (60+
seconds of guard tests, plus the Go/TS suites) on every rebase punishes the common case for a risk
the scoped re-verification already covers. Skip re-verification entirely after a clean rebase —
rejected: it would silently let a rebase that changed the very file this change also edited land
without ever having been checked against the new content.

### Never auto-abort a conflicting rebase

**ID:** never-auto-abort
**Status:** active
**Chosen:** A conflict stops the run, leaves the worktree mid-rebase, and reports the conflicting
files for the operator to resolve.
**Considered:** Auto-abort on conflict and fall back to "Continue — land anyway" — rejected outright:
silently landing a change that could not even cleanly rebase past the overlap is exactly the failure
mode this change exists to prevent.

## Open questions

None — the design converged in one round with no deferred question.
