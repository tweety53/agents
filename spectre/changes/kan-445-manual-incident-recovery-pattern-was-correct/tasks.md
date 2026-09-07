# kan-445-manual-incident-recovery-pattern-was-correct

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** one script, `scripts/recover-guard-incident.sh`, that reproduces KAN-423's incident
recovery on demand — diagnose from the reflog, abort a stuck `git revert`, then restore the
untracked planning files from the stash's third parent with `git show` redirects, abort strictly
before all restores — plus the sandbox harness that proves it.

**Architecture:** a red-green task pair. Task 1 writes the failing harness; task 2 implements the
tool against it. The tool is a single standalone Bash file with no lib/ dependency: its logic is
precondition checks and command emission, and every one of its behaviors is observable from
outside the process, which is what the harness asserts.

**Tech Stack:** Bash (3.2 floor — indexed arrays only, no associative arrays), real `git` in
`mktemp -d` sandbox repos.

**Spec:** `spectre/changes/kan-445-manual-incident-recovery-pattern-was-correct/design.md`

## Global Constraints

- Bash 3.2 is the floor: indexed arrays only, no associative arrays, no `${var,,}`.
- Script header-comment style follows `scripts/check-worktree-location.sh`: usage, exit-code
  contract, and the load-bearing rationale in capitalised topic paragraphs. `set -euo pipefail`.
- Harness shape is copied from `scripts/test-check-worktree-location.sh`: sandboxed `TMPDIR` root,
  a `SANDBOXES` array with an `EXIT` cleanup trap, `FAILURES` counter, `fail`/`pass` printers, and
  a `run_tool` capturing stdout and stderr separately (a refusal puts its message on stderr and
  must leave stdout empty; a merged capture cannot tell).
- A new `scripts/test-*.sh` is discovered by `scripts/run-guard-tests.sh`'s glob with no
  registration edit anywhere.
- Ordering is normative (design decision `redirect-restore-after-single-abort`): every
  `git revert --abort` precedes every restore; restores use `git show … >` redirects and never
  stage. Exit codes: 0 success, 1 precondition failure (cause named on stderr), 2 usage error.
- No tracked file is ever clobbered: a restore target tracked in the index or HEAD refuses the
  whole run before any action (design decision `refuse-tracked-targets`).
- Every guard in `.flow/project.md`'s `## lint` exits clean before a task is reported done; the
  new harness must also pass through `scripts/run-guard-tests.sh` without disturbing its siblings.

---

- [x] 1. Write the failing harness

**Build:** red
**Squash-with:** Task 2
**Files:** `scripts/test-recover-guard-incident.sh`
**Tests:** `no-revert-refusal`, `stash-without-third-parent-refusal`, `usage-error-exit-2`, `unknown-flag-usage-error`, `dry-run-changes-nothing`, `apply-restores-unstaged-after-abort`, `tracked-target-refusal`, `subdirectory-invocation-anchors-at-root`, `overwrite-naming`, `HEAD-tracked-target-refusal`, `custom-paths-replace-defaults`, `empty-restore-set-refusal`, `reflog-block-bounded-at-15`
**Regression:** reverting the fold commit removes the only executable statement of the design's
behavior contract — all thirteen cases die on a missing sibling script — and the guard-test suite loses
its one harness over `scripts/recover-guard-incident.sh`.
**Baseline:** before=53 after=54 harnesses matched by `scripts/test-*.sh`
<!-- measured: ls scripts/test-*.sh | wc -l @ branch spectre/kan-445-manual-incident-recovery-pattern-was-correct -->
<!-- predicted: ls scripts/test-*.sh | wc -l after this task — the new file matches the glob -->
**Commit:** `feat(scripts): add the guard-incident recovery tool and its harness`

  - [x] **Step 1: RED — create the harness skeleton and the thirteen cases**

    Create `scripts/test-recover-guard-incident.sh` with the skeleton below (header comment naming
    the exit-code contract it is the executable statement of, then the copied shape), followed by
    thirteen cases (six from this step, seven added by the round-1 fix). `GUARD` points at the sibling `scripts/recover-guard-incident.sh`, overridable only
    for a mutation case the way `CHECK_WORKTREE_LOCATION_GUARD` is.

```bash verified:setup, cleanup trap, counters and stream-splitting runner copied from scripts/test-check-worktree-location.sh @ branch spectre/kan-445-manual-incident-recovery-pattern-was-correct
#!/usr/bin/env bash
# Assertion harness for recover-guard-incident.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR and asserts the tool's refusals, its
# dry-run plan and its --apply effects. Never touches the real repository.
#
# THE CONTRACT THIS FILE IS THE EXECUTABLE STATEMENT OF:
#
#   exit 0   dry-run plan printed, nothing changed; or --apply completed
#   exit 1   precondition failure, cause on stderr, nothing on stdout
#   exit 2   usage error
#   --apply   REVERT_HEAD gone first, then untracked restores, never staged
#
# Shape copied from test-check-worktree-location.sh. Bash 3.2 is the floor.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="${RECOVER_GUARD_INCIDENT_TOOL:-$SCRIPT_DIR/recover-guard-incident.sh}"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

SANDBOXES=()
cleanup() {
  if [ "${#SANDBOXES[@]}" -ne 0 ]; then
    for s in "${SANDBOXES[@]}"; do
      chmod -R u+rwX "$s" 2>/dev/null || true
      rm -rf "$s"
    done
  fi
}
trap cleanup EXIT

WORK="$(mktemp -d "${TMPDIR:-/tmp}/recover-guard-incident-test.XXXXXX")"
SANDBOXES+=("$WORK")
ERRFILE="$WORK/stderr"

# run_tool <arg ...> -> OUT (stdout only), ERR (stderr only), RC. The streams
# are captured separately, never merged with 2>&1.
run_tool() {
  set +e
  OUT="$(bash "$GUARD" "$@" 2>"$ERRFILE")"
  RC=$?
  set -e
  ERR="$(cat "$ERRFILE")"
}

# new_repo -> REPO (physical form), a git repo with one commit, a tracked
# file two commits touch (so a conflicting revert is available), and planning
# files that only ever exist untracked (so a `git stash -u` captures them
# into the stash's third parent).
new_repo() {
  local r
  r="$(mktemp -d "${TMPDIR:-/tmp}/recover-guard-incident-repo.XXXXXX")"
  SANDBOXES+=("$r")
  git init -q "$r"
  git -C "$r" config user.email t@example.com
  git -C "$r" config user.name t
  printf 'one\n' >"$r/shared.txt"
  git -C "$r" add shared.txt
  git -C "$r" commit -q -m base
  printf 'two\n' >"$r/shared.txt"
  git -C "$r" commit -qam two
  printf 'one\n' >"$r/shared.txt"   # reverts `two` WITH conflict on commit, abort-able
  git -C "$r" commit -qam three
  REPO="$(cd "$r" && pwd -P)"
}

# start_conflicting_revert -> leaves a revert in progress (REVERT_HEAD set),
# the state the incident was recovered from.
start_conflicting_revert() {
  set +e
  git -C "$REPO" revert --no-commit HEAD~1 >/dev/null 2>&1
  set -e
}
```

    The six cases from this step, each in the file's `case`-block assertion style:

    - `usage-error-exit-2` — run with no arguments against a plain directory
      that is not a git repo: RC=2, stdout empty, stderr non-empty.
    - `no-revert-refusal` — a clean repo (no revert in progress), planning
      files present in `stash@{0}^3`: RC=1, stdout empty, stderr names the
      missing `REVERT_HEAD`, and `git -C "$REPO" rev-parse -q --verify
      REVERT_HEAD` still fails.
    - `stash-without-third-parent-refusal` — conflicting revert started, stash
      taken WITHOUT `-u` (no `^3` parent): RC=1, stderr distinguishes the
      no-third-parent cause from a missing stash, and the revert is still in
      progress afterwards.
    - `dry-run-changes-nothing` — conflicting revert started, planning files
      stashed with `-u` and then deleted from the worktree: run with no flag.
      RC=0; stdout lists `revert --abort` and one `git show "stash@{0}^3:<f>"`
      line per planning file, one under `docs/superpowers/` and one under
      `spectre/changes/`; stdout carries the 15-entry reflog block; `REVERT_HEAD` still verifies;
      no planning file exists on disk; `git -C "$REPO" diff --cached` is empty.
    - `apply-restores-unstaged-after-abort` — same setup, run with `--apply`.
      RC=0; `REVERT_HEAD` gone; each planning file back with the stashed
      content; `git -C "$REPO" status --porcelain` marks every restored file
      `??` (untracked) and no restored file is staged (`git diff --cached`
      empty) — the property the incident turned on.
    - `tracked-target-refusal` — conflicting revert started, planning files
      stashed with `-u`, then one planning path re-created AND staged: RC=1,
      stdout empty, stderr names the tracked file, the revert is still in
      progress, and the staged file's content is untouched.

  - [x] **Step 2: RED — run the harness and see every case fail**

    `scripts/test-recover-guard-incident.sh` — thirteen cases fail (the sibling
    tool does not exist yet, so `run_tool` cannot produce a passing exit) and
    exit 1.

  - [x] **Step 3: leave the commit to the fold**

    This task is `Build: red` with `**Squash-with:** Task 2` — its work lands
    in task 2's commit; no separate commit is made here.

- [x] 2. Implement `scripts/recover-guard-incident.sh`

**Build:** green
**Files:** `scripts/recover-guard-incident.sh`
**Allowed-collateral:** `docs/superpowers/specs/2026-09-07-kan-445-manual-incident-recovery-pattern-was-correct-design.md` (only if the implementation surfaces a design correction worth recording)
**Tests:** **none**
**Regression:** reverting this commit removes the tool; task 1's harness reports thirteen failing
cases and exits 1, and no script named `recover-guard-incident.sh` exists to recover an incident with.
**Baseline:** before=13 after=0 failing harness cases
<!-- measured: scripts/test-recover-guard-incident.sh exits 1 with thirteen failing cases @ branch spectre/kan-445-manual-incident-recovery-pattern-was-correct, after task 1 -->
<!-- predicted: the same harness exits 0 with zero failing cases after this task's commit -->
**Commit:** `feat(scripts): add the guard-incident recovery tool and its harness`

  - [x] **Step 1: GREEN — write the tool**

    Create `scripts/recover-guard-incident.sh`:

```bash unverified:confirm every construct against the bash 3.2 floor and real git in task 1's sandbox harness before trusting it
#!/usr/bin/env bash
# recover-guard-incident.sh — reproduce KAN-423's incident recovery on
# demand: abort a stuck `git revert`, then restore the untracked planning
# files the incident stash holds, unstaged, via `git show "stash@{0}^3:<f>" >
# <f>` redirects.
#
# THE ORDER IS THE POINT. `git checkout stash@{...} -- <path>` stages the
# restore, and any later `revert --abort` discards it — the incident's first
# restore attempt was lost exactly that way. Here every abort precedes every
# restore, and the redirects leave the files unstaged, so nothing after the
# abort point can discard them.
#
# Usage: recover-guard-incident.sh [--apply] [repo-dir] [path...]
#
#   --apply    execute; without it, print what would run and change nothing
#   repo-dir   defaults to the caller's cwd; must be a git repository
#   path...    planning paths to restore, relative to the repo root;
#              defaults to `docs/superpowers` and `spectre/changes`
#
# Exit 0 on a printed dry-run plan or a completed apply; 1 on a failed
# precondition (cause on stderr, stdout empty); 2 on a usage error.
#
# Bash 3.2 is the floor: indexed arrays only, no associative arrays.
set -euo pipefail

SELF="recover-guard-incident"

die() { printf '%s: %s\n' "$SELF" "$*" >&2; exit "${2:-1}"; }

usage() {
  printf 'usage: %s [--apply] [repo-dir] [path...]\n' "$SELF"
  printf '  --apply    execute; default is a dry-run plan\n'
  printf '  repo-dir   git repository, default cwd\n'
  printf '  path...    planning paths, repo-root-relative; default docs/superpowers spectre/changes\n'
}

APPLY=0
if [ "${1:-}" = "--help" ]; then usage; exit 0; fi
if [ "${1:-}" = "--apply" ]; then APPLY=1; shift; fi

if [ "$#" -gt 0 ]; then
  REPO="$(cd "$1" 2>/dev/null && pwd -P)" || die "not a directory: $1" 2
  shift
else
  REPO="$(pwd -P)"
fi
git -C "$REPO" rev-parse --show-toplevel >/dev/null 2>&1 \
  || die "not a git repository: $REPO" 2

[ "$#" -gt 0 ] || set -- docs/superpowers spectre/changes

git -C "$REPO" rev-parse -q --verify REVERT_HEAD >/dev/null 2>&1 \
  || die "no revert in progress in $REPO (REVERT_HEAD missing) — nothing to recover"

if ! git -C "$REPO" rev-parse -q --verify "stash@{0}" >/dev/null 2>&1; then
  die "no stash entry in $REPO — recovery restores stash@{0}^3, which needs a stash"
fi
if ! git -C "$REPO" rev-parse -q --verify "stash@{0}^3" >/dev/null 2>&1; then
  die "stash@{0} has no untracked third parent — it was not created with 'git stash -u'; re-stash with -u before any abort"
fi

# The restore set: every file the stash's third parent holds under the given
# paths, each checked against the working tree. Tracked anywhere (index or
# HEAD) -> refuse the whole run; already present untracked -> restore over
# it, named as an overwrite; absent -> a plain restore.
FILES=""
OVERWRITES=""
for p in "$@"; do
  while IFS= read -r f; do
    [ -n "$f" ] || continue
    if [ -n "$(git -C "$REPO" ls-files -- "$f")" ] \
       || git -C "$REPO" cat-file -e "HEAD:$f" 2>/dev/null; then
      die "refusing: $f is tracked in $REPO — recovery never clobbers tracked state" 1
    fi
    FILES="${FILES}${f}
"
    if [ -e "$REPO/$f" ]; then
      OVERWRITES="${OVERWRITES}${f}
"
    fi
  done < <(git -C "$REPO" ls-tree -r --name-only "stash@{0}^3" -- "$p")
done
[ -n "$FILES" ] || die "stash@{0}^3 holds no files under: $*"

printf 'reflog diagnosis — the alternating reset pattern the incident showed:\n'
git -C "$REPO" reflog -g HEAD -n 15 || true

printf 'plan:\n'
printf '  git -C %s revert --abort\n' "$REPO"
while IFS= read -r f; do
  [ -n "$f" ] || continue
  case "$OVERWRITES" in *"$f"*) printf '  (overwrite) ' ;; esac
  printf '  mkdir -p %s/%s\n' "$REPO" "$(dirname "$f")"
  printf '  git -C %s show "stash@{0}^3:%s" > %s/%s\n' "$REPO" "$f" "$REPO" "$f"
done <<<"$FILES"

if [ "$APPLY" -eq 1 ]; then
  printf 'running: git -C %s revert --abort\n' "$REPO"
  git -C "$REPO" revert --abort
  while IFS= read -r f; do
    [ -n "$f" ] || continue
    printf 'running: git show "stash@{0}^3:%s" > %s/%s\n' "$f" "$REPO" "$f"
    mkdir -p "$REPO/$(dirname "$f")"
    git -C "$REPO" show "stash@{0}^3:$f" > "$REPO/$f"
  done <<<"$FILES"
fi
```

    Adjust freely where the harness disagrees — the harness is the executable
    statement of the contract; this listing is the starting sketch. What may
    not change: the precondition order, abort-before-all-restores, unstaged
    redirects, the tracked-target refusal, and the exit-code contract.

  - [x] **Step 2: GREEN — run the harness**

    `scripts/test-recover-guard-incident.sh` — thirteen cases pass, exit 0,
    `FAILURES=0`.

  - [x] **Step 3: mutation check**

    Break one load-bearing line (e.g. move `git revert --abort` after the
    restore loop), re-run the harness, see `apply-restores-unstaged-after-abort`
    go red, restore the line, see it go green again.

  - [x] **Step 4: run the guard suite and the lint-relevant guards**

    `scripts/run-guard-tests.sh` — 54 harnesses, the new one included, all
    green. Then the plan-relevant subset of `.flow/project.md`'s `## lint`:
    `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.py` (this plan and the design spec are
    owned Markdown), `scripts/check-guard-symlinks.sh`,
    `scripts/check-contract-budget.sh` (no budget-bearing file is edited, so
    it must pass unchanged).

  - [x] **Step 5: Commit**

    `git add scripts/recover-guard-incident.sh scripts/test-recover-guard-incident.sh`
    then commit `feat(scripts): add the guard-incident recovery tool and its
    harness` — task 1's harness squashes into this one commit.
