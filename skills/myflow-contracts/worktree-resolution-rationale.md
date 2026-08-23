# Resolving a change's worktrees — rationale

This file is the reasoning behind `skills/myflow-contracts/worktree-resolution.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

## Resolving a change's worktrees

A map with zero keys and a map that was never populated
look identical to a raw read, so a direct read cannot tell "nothing to do" from "unpopulated," and
a zero-iteration loop over either reads as a pass it never earned.
