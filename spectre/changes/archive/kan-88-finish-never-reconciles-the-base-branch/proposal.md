# kan-88-finish-never-reconciles-the-base-branch

**Jira:** KAN-88 (absorbs KAN-150, KAN-195, KAN-238)

## Why

Nothing in the integrate/archive phase reconciles the base branch against its remote before acting
on it. Observed four times, on KAN-87, KAN-129, KAN-73 and KAN-201.

- A stale local base ref flips the preflight verdict: a merged branch returns `RUN1`, sending the
  operator back through integration — another landing question, another push to a branch the forge
  may have deleted, a duplicate-PR attempt. It fails safe but silently; the wrong verdict line reads
  exactly like a legitimate not-yet-merged answer.
- Run 1 asks how the branch should land without checking whether the base moved since the recorded
  merge base. On KAN-129 the conflict surfaced at the merge step, after two of three repositories had
  already merged and pushed, briefly leaving develop inconsistent.

The caller-side half of the first problem already landed — `finish-contract.md` composes
`<base-ref>` as `origin/$BASE` — leaving the guard correct by convention rather than by construction,
and leaving the ticket's own Verification item unmet.

## What changes

- `scripts/check-finish-preflight.sh` resolves a base ref it is handed to its remote-tracking
  counterpart when one exists, and names the ref it actually used in every verdict line.
- `scripts/lib/resolve-remote-base.sh` — that resolution, shared by both guards below.
- `scripts/check-base-moved.sh` — a new shipped guard reporting, per worktree, how far the base has
  moved since the recorded merge base and which of the moved paths this change also touched.
- `skills/flow/integrate.md` step 2 runs it before the landing prompt and, on overlap, asks once
  whether to stop and rebase or land anyway.
- `skills/flow-contracts/finish-contract.md` becomes canonical for the new guard's verdicts, exit
  contract, per-worktree rule and aggregated ask.
- Regression coverage: the ticket's own stale-local-base case in
  `scripts/test-check-finish-preflight.sh`, and `scripts/test-check-base-moved.sh`.

Design: `docs/superpowers/specs/2026-08-28-kan-88-finish-never-reconciles-the-base-branch-design.md`
