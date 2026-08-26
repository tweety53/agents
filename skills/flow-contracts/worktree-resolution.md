# Resolving a change's worktrees

How any step resolves the set of worktrees belonging to a change.

**Loaded by `/myflow-do`, `/myflow-finish`, `/myflow-status` and `/myflow-fast`** — before a
preflight verdict, a gate that runs once per worktree, a status report, or a removal.

This file is **canonical** for everything in it.

The reasoning behind this file lives in
`skills/flow-contracts/worktree-resolution-rationale.md`; **a `/myflow-*` run never loads it.**

## Resolving a change's worktrees

Any step in any command that needs "the worktrees" for a change — a preflight verdict, a gate that
runs once per worktree, a status report, or a removal — resolves the set first; it never loops over
the state file's `worktrees` map directly. See **Resolving a change's worktrees**
(`skills/flow-contracts/worktree-resolution-rationale.md`) for why.

**A resolved set that comes back empty is never a vacuous pass.** Report it explicitly rather than
let a zero-iteration loop read as "every worktree passed," "every worktree is merged," or whatever
verdict the calling step would otherwise default to on no evidence at all. A gate stops and asks the
operator, exactly as it would on any other refusal; a read-only report says so in its own output
instead of silently omitting the change.

This binds every command that iterates a change's worktrees: `/myflow-do`'s workspace-isolation
gate, `/myflow-status`'s merge-status report, and `/myflow-finish`'s preflight verdict,
unfinished-work gate and run 2 removal alike — each resolves its own set through this rule rather
than restating it. How a command resolves the set beyond reading the state file's map — whether it
falls back to a filesystem scan, and what an inconclusive answer does next — is that command's own;
see **Resolving a change's worktrees** (`skills/flow-contracts/finish-contract.md`).
