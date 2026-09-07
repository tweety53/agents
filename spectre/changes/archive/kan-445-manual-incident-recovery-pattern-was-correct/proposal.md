# kan-445-manual-incident-recovery-pattern-was-correct

## Why

KAN-423's guard incident left a stuck `git revert` and a wiped planning directory. The parent
session recovered correctly, but under pressure and with one failed attempt folded in: it spotted
the reflog's alternating reset pattern, ran `git revert --abort`, then restored the untracked
planning files from the stash's third parent via `git show "stash@{0}^3:<path>"` redirects — after
first losing a restore to `git checkout stash@{0}^3 -- <path>`, which stages the restored files,
so the run's own later `revert --abort` discarded them again. The sequence is correct but was
re-derived live; KAN-445 asks that it be reproduced deliberately next time rather than reasoned
out again mid-incident.

## What changes

A generic recovery tool and its test harness, both new files:

- `scripts/recover-guard-incident.sh` — verifies the incident's preconditions (revert in progress,
  `stash@{0}^3` present, every restore target missing or untracked), prints the reflog diagnosis,
  then aborts the revert and restores the untracked planning files with `git show` redirects,
  abort strictly before all restores. Dry-run by default; `--apply` executes.
- `scripts/test-recover-guard-incident.sh` — a sandbox-repo harness covering refusal without a
  revert, refusal without a `^3` parent, dry-run changing nothing, apply restoring files untracked
  and unstaged, and tracked-target refusal.

Observable difference: an operator facing the same incident runs one script instead of re-deriving
a five-step recovery under pressure, and the staging pitfall is encoded in the tool's ordering
rather than in someone's memory.
