# Integrate — stops

Loaded by `skills/flow/integrate.md` on a `STOP:` line it does not handle in place. Every section
name below without a path is a section of `skills/flow/integrate.md`; every stop closes the stage
mark it was in `stopped`, and the parent re-runs the call that stopped once it is handled.

## `prepare`

- **`STOP: main-checkout`** → stop at `IN_PROGRESS` and report every repository's lines together,
  per **Surface main-checkout drift before the preflight**
  (`skills/flow-contracts/finish-contract-run1.md`); the operator cleans the main checkout and
  re-runs. A `DRIFT-SWITCHED:` line is relayed as printed and asks nothing: `prepare` already switched that clean checkout back to its default branch and carried on, per **Surface main-checkout drift before the preflight** (`skills/flow-contracts/finish-contract-run1.md`).
- **`STOP: preflight` or `STOP: base`** → stop, report what the script reported, relay the guard's
  hand-verification procedure per **Hand-verifying a guard verdict** (`skills/flow-contracts/pipeline.md`), and ask
  the operator.
- **`STOP: base-mismatch`** → report the line and ask the operator. A worktree on its default branch whose origin lacks the base another worktree records is not a stop: `prepare` creates that base on its origin from the default branch, records it as the worktree's base with the first and third commands below (the merge base unchanged, since the new base sits at the default branch's commit), leaves moving the branch to the base-moved sync, and prints one `BASE-CREATED:` line.
  The answer that moves the worktree is `git -C <wt> config branch.spectre/<name>.flowBase <b>`,
  `git -C <wt> rebase --onto origin/<b> <its recorded merge base>`,
  `flow state add-worktree -C <the change's canonical worktree> <name> <wt> $(git -C <wt> merge-base HEAD origin/<b>)`, then re-run
  `prepare`.
- **`STOP: no-remote`** → per **The routes** (`skills/flow-contracts/finish-contract-run1.md`).
- **`STOP: no-verdict`** → a guard of the gate printed no verdict line, or exited non-zero without
  one: stop and ask, reporting its output.

## `land`

**Every git step here can fail, and none of them may fail silently.** `STOP: push`, `STOP: sync`,
`STOP: sync-conflict`, `STOP: rebase-lint` and `STOP: state` are
reported with the command's own output, leaving the change at `IN_PROGRESS`; `STOP: base` and
`STOP: no-remote` are handled as in `prepare` above; a conflict on the merge-and-push route's one re-sync is resolved per
**Sync the branch onto the base** (`skills/flow/sync-onto-base.md`) and `land` re-run as the one
retry.

**`STOP: pr`** — no usable `gh`: print the forge's create-PR URL and ask whether it was opened.
**Human confirmation is a legitimate substitute for an API probe** on a forge with no usable CLI. If
the answer is No, leave `prUrl` null and say what to do next. Either answer re-runs `land` adding
`--pr-url <worktree>=<url>` or `--pr-url <worktree>=none` for the worktree the stop names, every
earlier `--pr-url` kept.
