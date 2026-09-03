# kan-285-myflow-prompt-a-rebase-before-the-review-panel

## Why

`check-base-moved.sh` runs once, at integrate run 1 step 2 (`skills/flow/integrate.md`, added by
KAN-371). The review panel runs earlier and reviews a diff against a merge base that may already
be stale — nothing at panel entry compares that merge base against the base tip. So the conflict,
and every proof relative to that base, surfaces at the merge, after the panel has closed and the
change is verified. Observed on KAN-265: three PRs landed on `main` mid-change, one touching the
same guard and spec files; the stale baseline had to be recaptured and every proof re-derived, and
two genuine defects were only found then.

## What changes

- `skills/flow/review-panel.md` gains a first step under the `flow.review-panel` begin mark, before
  the citation pre-check and the dispatch-context rebuild, on every panel run: per worktree,
  `resolve-base-branch.sh` then `check-base-moved.sh` against the run's working-notes merge base,
  with integrate step 2's verdict handling and, on overlap, its Stop / Rebase now / Continue prompt.
- A clean rebase at panel entry updates that worktree's working-notes merge base to `origin/$BASE`'s
  tip, re-checks once, clears every slot's held last-reviewed sha, and runs no scoped
  re-verification — the panel and `flow.verify` follow. A conflict stops the run mid-rebase, never
  auto-aborted, never resolved by the pipeline.
- After a clean rebase at either site — panel entry, and integrate step 2 — the report ends with
  one fixed sentence saying a recorded baseline must be recaptured now.
- The integrate-time check stays as the later net; `skills/flow-contracts/finish-contract-run1.md`
  and every script are unchanged.
