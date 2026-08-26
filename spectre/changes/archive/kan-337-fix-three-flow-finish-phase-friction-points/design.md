# Design: three /flow finish-phase fixes

## Decisions

### Docker restart is documentation guidance, not a procedural reorder

**ID:** docker-restart-is-guidance
**Status:** active
**Chosen:** add a note to `finish-contract.md` clarifying that a `SKIPPED:` clause on
`check-cleanup-complete.sh`'s `COMPLETE:` line, a reported-and-continued removal failure at step 5,
and a journaled state write are all correct, expected outcomes once the project's stack has been
stopped by worktree cleanup's check 5 — none of them is a cue to restart the stack.
**Considered:** reordering the `## stop` command to run after every store-dependent step, or having
those later steps auto-restart the stack — rejected on inspection: every step after check 5 already
degrades gracefully by the contract's own existing design (verified directly against KAN-271's own
run), so reordering or auto-restarting would add real procedural complexity to fix a symptom that
was actually the operator's own over-caution, not a gap in the contract.

### The `cd`-out precondition is stated where check 6 is invoked

**ID:** cd-out-before-process-check
**Status:** active
**Chosen:** `finish-contract.md`'s Worktree cleanup section states, immediately at check 6, that the
orchestrating shell's own cwd must be outside every worktree in the resolved set before the check
runs.
**Considered:** changing `check-worktree-processes.sh` itself to exclude the caller's own shell —
rejected: the guard's existing behavior (counting the caller's shell) is deliberate and documented
in its own header as a real safety property, not a bug; the gap is only that the calling procedure
never stated its own precondition.

### The main checkout's stray copy is cleaned up where it is created

**ID:** cleanup-main-checkout-copy
**Status:** active
**Chosen:** `skills/flow/implement.md`'s isolate-workspace step now documents copying
`<project>/spectre/changes/<name>/` into the new worktree and then removing the main checkout's own
copy of it.
**Considered:** leaving removal to archive time (`skills/flow/archive.md`) instead — rejected: by
archive time there is no way to tell a genuinely-stale main-checkout copy from operator work in
progress there, whereas at isolate-workspace time the copy being removed was made moments earlier by
this same run and is unambiguously safe to remove.

## Open questions

None.
