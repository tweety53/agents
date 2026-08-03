# Manual test guide — kan-53-myflow-no-uncommitted-diff-helper-for-subagent

This repository declares no runnable application (`.myflow/project.md`'s `## apps` section). Run
the checks below from the repository root in this worktree:

```bash
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/test-uncommitted-review-package.sh
```

## Checks

### `checkpoint` — non-destructive per-task snapshot

- [x] On a clean tree, prints the current `HEAD` sha.
- [x] On a tree with a tracked, uncommitted modification, prints a commit-ish distinct from `HEAD`.
- [x] On a tree with only a brand-new untracked file (never `git add`ed), prints a commit-ish
      distinct from the checkpoint taken immediately before that file existed — a second checkpoint
      no longer collapses to the same value as the first.
- [x] Leaves `git status`, `git stash list`, and `HEAD` unaffected by the call, other than the
      documented staging of the working tree (excluding the three NO-COMMITS planning paths below).
- [x] Invoked from a subdirectory of the repo, still stages an untracked file elsewhere in the repo
      (repo-root-anchored, not scoped to the invocation cwd).
- [x] Never stages `openspec/`, `docs/manual-test/`, or `docs/superpowers/`, even when those paths
      carry uncommitted changes at invocation time.

### `uncommitted-review-package` — per-task/per-fix-round review package

- [x] Given a valid plan file and `BASE`, writes a package containing a `## Files changed` section
      and a `## Diff` section, and no `## Commits` section.
- [x] A later round's changes are isolated from an earlier round's — a second package built from a
      second checkpoint shows only the second round's own change.
- [x] A brand-new untracked file created after `BASE` appears correctly in the diff as an addition.
- [x] Invoked from a subdirectory of the repo, still stages and diffs an untracked file elsewhere in
      the repo.
- [x] Never stages or includes `openspec/`, `docs/manual-test/`, or `docs/superpowers/` in its
      output, even when those paths carry uncommitted changes at invocation time.
- [x] A nonexistent plan file exits non-zero, reports `no such plan file:` on stderr, and writes no
      output file.
- [x] A `BASE` that does not resolve exits non-zero, reports `bad BASE:` on stderr, and writes no
      output file.
- [x] Called again with a distinct explicit `OUTFILE`, writes to that new path and leaves the prior
      package file byte-for-byte untouched.
- [x] Called with no explicit `OUTFILE` (the shape `skills/myflow-do/SKILL.md` documents for the
      ordinary per-task call), resolves a default output path under
      `.superpowers/sdd/<plan-basename>/` and prints `wrote <path>: <bytes> bytes` on stdout.

### `skills/myflow-do/SKILL.md` — wiring

- [x] The NO-COMMITS block, the per-task review paragraph, and the fix-round paragraph all
      reference `checkpoint` and `uncommitted-review-package` — no `git rev-parse HEAD` /
      `git diff TASK_BASE` / `git diff FIX_BASE` reference remains anywhere in the file.

## Known incomplete

None.
