# kan-445-manual-incident-recovery-pattern-was-correct — design

Approved 2026-09-07 in the `/flow` brainstorming round for KAN-445. This is the design the change's
`spectre/changes/kan-445-manual-incident-recovery-pattern-was-correct/design.md` was sourced from.

## Problem

KAN-423's guard incident left a stuck `git revert` and a wiped planning directory. The recovery
that worked: read the reflog's alternating reset pattern, `git revert --abort`, restore untracked
planning files from the stash's third parent with `git show "stash@{0}^3:<file>" > <file>`
redirects. The first attempt lost its restore to `git checkout stash@{...} --`, which stages the
files so the run's own later `revert --abort` discarded them. KAN-445 asks that the sequence be
reproduced deliberately, not re-derived under pressure.

## Approach

A generic Bash tool, `scripts/recover-guard-incident.sh`, plus `scripts/test-recover-guard-incident.sh`
(discovered automatically by `scripts/run-guard-tests.sh`'s glob). Dry-run by default, `--apply`
to execute. One abort before all restores. The script's header comment is the runbook; the harness
is the proof. No doc file, no edits to existing files.

Ruled out in brainstorming: this-repo-only hard-coded paths (the kan-451 incident table is
per-project); runbook-only (a doc reminds, a script enforces); always-execute (irreversible work,
no pause); interactive confirm (needs a `--yes` escape to be testable — always-execute with extra
steps).

## CLI

```
scripts/recover-guard-incident.sh [--apply] [repo-dir] [path...]
```

`repo-dir` defaults to cwd and must resolve through `git rev-parse --show-toplevel`. `path...`
defaults to `docs/superpowers` and `spectre/changes`; paths given replace the defaults and resolve
relative to the repo root. `--help` prints usage. Exit codes: 0 success, 1 precondition failure
(named cause), 2 usage error.

## Preconditions (in order; a non-repository exits 2 as a usage error; the remaining precondition failures exit 1)

1. Repo dir is a git repository.
2. Revert in progress: `git rev-parse -q --verify REVERT_HEAD` succeeds.
3. `git rev-parse -q --verify "stash@{0}^3"` succeeds; no-stash and stash-without-`-u` are
   distinguished in the message.
4. Every restore target is missing or untracked. A tracked target refuses the whole run, reported on the
   first offending file; an untracked existing target is overwritten and the overwrite named in the plan.

Diagnosis (both modes): the last 15 `git reflog -g HEAD` entries.

## Sequence (order load-bearing)

Enumerate the restore set with `git ls-tree -r --name-only "stash@{0}^3" -- <path>` filtered by
the missing/untracked check. Dry-run prints every command it would run and exits 0 unchanged.
`--apply` runs `git revert --abort` first, then per file `mkdir -p <dir>` and
`git show "stash@{0}^3:<file>" > <file>`, each step printed, first failure exiting nonzero. All
aborts precede all restores; the redirects leave files unstaged, so nothing after the abort point
discards them.

## Testing

Five sandbox-repo cases (`mktemp -d`, cleanup trap, bash 3.2): refusal without a revert; refusal
without a `^3` parent; dry-run changes nothing; `--apply` aborts and restores untracked-unstaged
with stash content; tracked target refuses with revert state and tracked content untouched.

## Files

- `scripts/recover-guard-incident.sh` (new)
- `scripts/test-recover-guard-incident.sh` (new)
