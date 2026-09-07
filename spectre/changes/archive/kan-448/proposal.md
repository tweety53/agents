# kan-448

## Why

Self-review finding from KAN-423 (flow-automation angle), filed as KAN-448 with companion flow-fix
issue KAN-439. Two gaps remain after KAN-442:

1. **Guards that mutate the working tree can still leave residue behind, and nothing inside them
   notices.** KAN-442 removed the incident's own mechanism (`check-task-commit-fields.py` no
   longer stashes, reverts or resets; its git use is read-only), and `skills/flow/implement.md`
   gained a conductor-side inspect-before-retry rule. But `scripts/mutate-and-verify.sh` (git
   apply → harnesses → `git checkout --` restore) verifies only the files it touched — residual
   dirt exits 3, and there is no stash-delta check at all — and `scripts/prepare-archive-branch.sh`
   (`git checkout -q $BASE` + `git merge --ff-only` in the landing worktree) is guarded only by a
   pre-checkout dirty refusal. The ticket's failure — a conductor retrying blind over residue —
   stays reachable through both.
2. **The cross-repo `spectre link` cannot succeed where the pipeline runs it, and its failure is
   non-gating.** `skills/flow/implement.md` §2 runs the link inside each new worktree after
   `git worktree add`, where the peers file's cwd-relative entries resolve into `.worktrees/` and
   the link is always refused ("peer gymie is not present" — KAN-439's logged cause), and states
   "the link is not a gate." Every cross-repo change therefore lands at integrate with a false
   OUTSTANDING verdict from `check-unfinished-work.sh` and a hand-verification step instead of a
   trusted guard.

## What changes

- New `scripts/lib/post-mutation-check.sh` (bash 3.2, sourced): `snapshot_tree_state` captures
  `git status --porcelain=v2` and `git stash list` before a guard's first mutation;
  `check_tree_restored` diffs against that snapshot afterwards, names every new stash entry and
  unexpected status line, and exits nonzero on drift. Read-only; mutates nothing.
- `scripts/mutate-and-verify.sh` snapshots immediately before `git apply` and runs the check in
  its EXIT trap after the existing touched-file residual test: touched-file dirt keeps its
  documented exit 3 (precedence), broader drift exits 2 naming the stash names. The header
  documents both meanings of exit 2.
- `scripts/prepare-archive-branch.sh` snapshots after its dirty-tree refusal and checks after its
  checkout and fast-forward merge; drift exits 2 naming the stash names.
- `skills/flow/implement.md` §2 runs `spectre link --root <worktree>/spectre
  <canonical-peer>:<name>` with cwd at the satellite repo's primary checkout, after each
  additional worktree's `git worktree add`; a refusal hard-fails the stage and the "the link is
  not a gate" sentence is deleted. This resolves KAN-439, which closes when this change lands.
- `scripts/check-unfinished-work.sh` is not touched; `spectre` itself is not modified.

<!-- measured: KAN-423's incident (55 minutes of hand recovery, one wasted conductor dispatch) is
     KAN-442's proposal quoting KAN-423's self-review; a live incident, not re-runnable. The
     false-OUTSTANDING mechanism is live on every cross-repo change, per KAN-439. -->
