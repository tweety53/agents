> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

## 1. `checkpoint` script

- [x] 1.1 Create `skills/myflow-do/scripts/checkpoint` (executable, `chmod +x`), modeled on the
  header style of `skills/subagent-driven-development/scripts/sdd-workspace`
  (`verified:read` — file at
  `/Users/tweety53/.claude/plugins/cache/claude-plugins-official/superpowers/6.2.0/skills/subagent-driven-development/scripts/sdd-workspace`).
  No arguments. Body (drafted here, not yet run; task 1.2 exercises it against a real repo before it
  is considered done):
  ```bash unverified:
  #!/usr/bin/env bash
  set -euo pipefail
  ref=$(git stash create) || true
  if [ -z "$ref" ]; then
    ref=$(git rev-parse HEAD)
  fi
  echo "$ref"
  ```
  `unverified:` — `git stash create`'s exit status on a clean tree specifically (it is documented to
  print nothing, but the `|| true` guards against a non-zero exit accompanying the empty output;
  confirm behavior empirically in task 1.2 before relying on it in the real script).
- [x] 1.2 Manually verify `checkpoint`'s two paths in a scratch repo before writing the test suite:
  `cd $(mktemp -d) && git init -q && git commit -q --allow-empty -m init`, run `checkpoint` on the
  clean tree and confirm it prints the `HEAD` sha (`git rev-parse HEAD`); then `echo x > f.txt`
  (untracked — `git stash create` does not snapshot untracked files by default, so also run
  `git add f.txt` to confirm the staged case) and confirm `checkpoint` prints a *different*
  commit-ish, and that `git status --short`, `git stash list`, and `git rev-parse HEAD` are
  unchanged before and after the call.

## 2. `uncommitted-review-package` script

- [x] 2.1 Create `skills/myflow-do/scripts/uncommitted-review-package` (executable), adapted from
  `skills/subagent-driven-development/scripts/review-package`
  (`verified:read` — same plugin path as above, `scripts/review-package`). Usage:
  `uncommitted-review-package PLAN_FILE BASE [OUTFILE]`. Differences from `review-package` to
  implement:
  - validate `PLAN_FILE` exists and `BASE` resolves via `git rev-parse --verify --quiet`, exiting 2
    with `no such plan file: $plan` / `bad BASE: $base` on failure — same two checks
    `review-package` makes on its own first two arguments (`verified:read`, same file, lines 20-23)
  - default `OUTFILE` via the same `sdd-workspace` helper `review-package` uses, but named
    `review-<base7>.diff` (single sha, no `..head7`, since there is no HEAD-side commit to name),
    resolved with
    `dir=$("$(cd "$(dirname "$0")" && pwd)/../../subagent-driven-development/scripts/sdd-workspace" "$plan")`.
    `unverified:` — confirm this relative path from `skills/myflow-do/scripts/` to the installed
    `subagent-driven-development` skill's `sdd-workspace` actually resolves once both skills are
    installed side by side under the same `skills/` root (true for this repo's own layout and for
    every harness's plugin-cache layout observed in `skills/myflow-do/SKILL.md`'s own references to
    sibling skill files, but not independently confirmed here for `subagent-driven-development`
    specifically since it ships from a different install root, the `superpowers` plugin cache, not
    this repo). If it does not resolve in the harness's actual layout, fall back to invoking
    `sdd-workspace` by an absolute path resolved the same way `myflow-do/SKILL.md` already resolves
    `engineering-principles.md` (see `[PRINCIPLES_PATH]` in `SKILL.md`) and record that as a design
    note in this task's completion.
  - write `## Files changed` (`git diff --stat "$base"`) and `## Diff` (`git diff -U10 "$base"`) —
    single-ref diff against the live working tree, not a `base..head` range
  - write no `## Commits` section
  - final stdout line mirrors `review-package`'s own: `wrote <out>: <bytes> bytes` (drop the
    `<commits> commit(s)` clause, since there are none)
- [x] 2.2 Manually verify `uncommitted-review-package` end-to-end in the same scratch repo from 1.2:
  record `BASE=$(checkpoint)`, make a further change, run
  `uncommitted-review-package some-plan.md "$BASE"`, and inspect the written file for a
  `## Files changed` section, a `## Diff` section, and the absence of any `## Commits` heading.

## 3. Wire into `skills/myflow-do/SKILL.md`

- [x] 3.1 Rewrite the NO-COMMITS block in step 4 (currently lines 107-111 —
  `verified:read`, `skills/myflow-do/SKILL.md` in this repo) so the sentence "The parent records
  `TASK_BASE=$(git rev-parse HEAD)` before dispatch; your diff for review is `git diff TASK_BASE`."
  instead reads: "The parent records `TASK_BASE=$(skills/myflow-do/scripts/checkpoint)` before
  dispatch; your diff for review is `skills/myflow-do/scripts/uncommitted-review-package
  <plan-file> "$TASK_BASE"`." Keep the rest of the NO-COMMITS block unchanged.
- [x] 3.2 Update the paragraph in step 4 (currently line 133 —
  `verified:read`, same file: "Per-task review without commits: write `git diff TASK_BASE >
  .superpowers/sdd/task-N.diff` and give the reviewer that path, never a commit range.") to instead
  say the parent runs `skills/myflow-do/scripts/uncommitted-review-package <plan-file> "$TASK_BASE"`
  and gives the reviewer the printed path.
- [x] 3.3 Update the fix-round paragraph in step 5's "Panel re-runs" section (currently line 270 —
  `verified:read`, same file: "Record `FIX_BASE` before each fix, then `git diff FIX_BASE >
  .superpowers/sdd/fix-round-N.diff`.") to: "Record `FIX_BASE=$(skills/myflow-do/scripts/checkpoint)`
  before each fix, then `skills/myflow-do/scripts/uncommitted-review-package <plan-file> "$FIX_BASE"
  .superpowers/sdd/fix-round-N.diff`."
- [x] 3.4 Grep the rest of `SKILL.md` for any other `rev-parse HEAD` / `TASK_BASE` / `FIX_BASE`
  reference this change did not already account for (`grep -n "rev-parse HEAD\|TASK_BASE\|FIX_BASE"
  skills/myflow-do/SKILL.md`) and update any remaining occurrence the same way.

## 4. Tests

- [x] 4.1 Create `scripts/test-uncommitted-review-package.sh`, following the structure of
  `scripts/test-check-unfinished-work.sh` (`verified:read` — this repo, same conventions: `set -euo
  pipefail`, an indexed-array `SANDBOXES` trap-cleaned on `EXIT`, `mktemp -d
  "${TMPDIR:-/tmp}/<name>-test.XXXXXX"`, `fail`/`pass` helpers, a `FAILURES` counter, `exit 1` if
  `FAILURES` is non-zero at the end). Build a throwaway git repo per test case under the sandbox
  (`git init -q`, an initial empty commit so `HEAD` exists) — never touch the real repository tree.
- [x] 4.2 Case: clean-tree `checkpoint` prints the current `HEAD` sha.
- [x] 4.3 Case: dirty-tree `checkpoint` prints a commit-ish that is not equal to `HEAD`.
- [x] 4.4 Case: two-round isolation — make change A, capture `BASE=$(checkpoint)`, make change B,
  run `uncommitted-review-package`, and assert the written diff contains change B's content and does
  **not** contain change A's content (the property this whole change exists to provide).
- [x] 4.5 Case: the written package file contains a `## Files changed` line and a `## Diff` line, and
  does not contain a `## Commits` line.
- [x] 4.6 Case: `uncommitted-review-package` given a nonexistent plan file exits non-zero and prints
  `no such plan file:` on stderr, writing no output file.
- [x] 4.7 Case: `uncommitted-review-package` given a `BASE` that does not resolve exits non-zero and
  prints `bad BASE:` on stderr, writing no output file.
- [x] 4.8 Run `scripts/test-uncommitted-review-package.sh` and confirm every case reports `ok:` with
  `FAILURES=0`.
- [x] 4.9 Case (added during implementation — task-group-4 review flagged this spec scenario as
  untested): a second `uncommitted-review-package` call against the same repo with a distinct explicit
  `OUTFILE` writes a new file and leaves the prior package's file byte-for-byte intact.

## 5. Repo-wide verification

- [x] 5.1 Run `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
  `scripts/check-plan-provenance.sh` and confirm each exits clean against this change's artifacts
  (`verified:read` — these three are this repo's `.myflow/project.md` `## lint` commands).
- [x] 5.2 Run every `scripts/test-*.sh` in the repository (`for t in scripts/test-*.sh; do "$t" ||
  echo "FAILED: $t"; done`) and confirm none regressed.
