#!/usr/bin/env bash
# Assertion harness for resolve-base-branch.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR and asserts the guard's exit status,
# stdout AND stderr together for every case. Never touches the real
# repository tree.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated
# contract in openspec/changes/kan-236-base-branch-resolution-retyped-by-hand/design.md,
# never against observed output.
#
# Every case checks stdout, stderr AND exit status together. A guard that
# exits correctly while printing a stale name on stdout — or that refuses
# correctly but leaves stdout non-empty — is the exact failure this whole
# change exists to remove, so a case that only checks the exit code proves
# nothing.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/resolve-base-branch.sh"
REAL_GIT="$(command -v git)"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# skip <label> <reason> — a case this environment cannot decide, reported as
# neither a pass nor a failure. Matches test-check-cleanup-complete.sh's own
# `skip` helper and its two environmental reasons: a file mode of 000 still
# readable by root, and a git that stores refs packed (or in the reftable
# backend) rather than as loose files, so there is nothing to `chmod` at all.
# Printing "ok" for either would be a green result for an assertion that
# never ran, which is the vacuous pass every guard in this repository is
# written against.
skip() { printf 'skip: %s (%s)\n' "$1" "$2"; }

# An indexed array, not a space-separated string: sandbox paths come from
# mktemp under TMPDIR, which may contain spaces, and word-splitting a string
# would then rm -rf the fragments. Bash 3.2 has indexed arrays; only
# associative arrays are unavailable.
REPOS=()
cleanup() {
  # ${REPOS[@]} is unset-expansion-unsafe under `set -u` on bash 3.2 when empty.
  [ "${#REPOS[@]}" -eq 0 ] && return 0
  for repo in "${REPOS[@]}"; do
    rm -rf "$repo"
  done
}
trap cleanup EXIT

# run_guard <dir> -> sets OUT, ERR, RC. stdout and stderr are captured
# separately (never merged) so a case can assert both independently.
#
# Runs the guard through its own `#!/usr/bin/env bash` shebang, unless the
# caller sets GUARD_INTERP as a temporary assignment prefixing the call
# (same style as the `PATH=...` and `LC_ALL=...` prefixes already used at
# call sites below), in which case the guard is invoked as
# "$GUARD_INTERP" "$GUARD" instead. GUARD_INTERP exists for exactly one
# caller: case 12 below (the non-ASCII locale case) pins it to /bin/bash
# because on a machine where a newer bash sits ahead of the system one on
# PATH, `env bash` resolves to it, and that newer bash does not reproduce
# the collating-range defect the LC_ALL=C fix and its mutation-proof both
# depend on — pinning that one case to the system /bin/bash is what "bash
# 3.2 is the floor" actually means here, matching
# test-check-plan-provenance.sh's own precedent for forcing an interpreter
# (case 182, `/bin/bash "$GUARD"`).
run_guard() {
  ERRFILE="$(mktemp "${TMPDIR:-/tmp}/resolve-base-branch-err.XXXXXX")"
  set +e
  if [ -n "${GUARD_INTERP:-}" ]; then
    OUT="$("$GUARD_INTERP" "$GUARD" "$@" 2>"$ERRFILE")"
  else
    OUT="$("$GUARD" "$@" 2>"$ERRFILE")"
  fi
  RC=$?
  set -e
  ERR="$(cat "$ERRFILE")"
  rm -f "$ERRFILE"
}

# expect_success <case-name> <expected-branch> — exit 0, stdout is exactly
# the branch name, stderr is empty.
expect_success() {
  [ "$RC" -eq 0 ] && pass "$1: exit 0" \
    || fail "$1: expected exit 0, got rc=$RC out=[$OUT] err=[$ERR]"
  [ "$OUT" = "$2" ] && pass "$1: stdout is '$2'" \
    || fail "$1: expected stdout '$2', got '$OUT'"
  [ -z "$ERR" ] && pass "$1: stderr empty" \
    || fail "$1: expected empty stderr, got '$ERR'"
}

# expect_refusal <case-name> <expected-exit> <stderr-substring> — the given
# exit code, empty stdout, and stderr containing the substring.
expect_refusal() {
  [ "$RC" -eq "$2" ] && pass "$1: exit $2" \
    || fail "$1: expected exit $2, got rc=$RC out=[$OUT] err=[$ERR]"
  [ -z "$OUT" ] && pass "$1: stdout empty" \
    || fail "$1: expected empty stdout, got '$OUT'"
  case "$ERR" in
    *"$3"*) pass "$1: stderr names the failure" ;;
    *) fail "$1: expected stderr to mention '$3', got '$ERR'" ;;
  esac
}

# new_worktree -> sets WT, REMOTE. A bare "remote" repository with a main
# branch, cloned into WT with refs/remotes/origin/HEAD explicitly set, then
# checked out onto a branch named openspec/fixture.
new_worktree() {
  REMOTE="$(mktemp -d "${TMPDIR:-/tmp}/resolve-base-branch-test.XXXXXX")"
  REPOS+=("$REMOTE")
  rm -rf "$REMOTE"
  git init -q -b main --bare "$REMOTE"

  SEED="$(mktemp -d "${TMPDIR:-/tmp}/resolve-base-branch-test.XXXXXX")"
  REPOS+=("$SEED")
  git init -q -b main "$SEED"
  git -C "$SEED" config user.email test@example.invalid
  git -C "$SEED" config user.name "Test"
  echo base > "$SEED/file.txt"
  git -C "$SEED" add file.txt
  git -C "$SEED" commit -qm base
  git -C "$SEED" remote add origin "$REMOTE"
  git -C "$SEED" push -q origin main

  WT="$(mktemp -d "${TMPDIR:-/tmp}/resolve-base-branch-test.XXXXXX")"
  REPOS+=("$WT")
  rm -rf "$WT"
  git clone -q "$REMOTE" "$WT"
  git -C "$WT" remote set-head origin -a >/dev/null
  git -C "$WT" checkout -q -b openspec/fixture
}

# new_worktree_no_head -> like new_worktree, but refs/remotes/origin/HEAD is
# deliberately deleted, so the `git remote show origin` fallback is the only
# resolution path left.
new_worktree_no_head() {
  new_worktree
  git -C "$WT" remote set-head origin -d
}

# new_worktree_no_origin -> sets WT. An ordinary repository with a commit and
# no remote configured at all.
new_worktree_no_origin() {
  WT="$(mktemp -d "${TMPDIR:-/tmp}/resolve-base-branch-test.XXXXXX")"
  REPOS+=("$WT")
  rm -rf "$WT"
  git init -q -b openspec/fixture "$WT"
  git -C "$WT" config user.email test@example.invalid
  git -C "$WT" config user.name "Test"
  echo base > "$WT/file.txt"
  git -C "$WT" add file.txt
  git -C "$WT" commit -qm base
}

# shim_git [<text-file>] -> sets SHIM_DIR, a PATH entry whose `git`:
#   - turns `fetch --quiet origin` into a silent no-op, so it cannot restore
#     a deleted refs/remotes/origin/HEAD the way this machine's git (which
#     writes a remote's HEAD ref on fetch when one is missing and the remote
#     advertises it) otherwise would — the fixture's deletion must survive
#     the guard's own fetch for the fallback path to be exercised at all;
#   - when a text file is given, answers `remote show origin` with that
#     file's contents verbatim instead of the real answer, to build a
#     `HEAD branch:` value the real command will never produce;
#   - forwards every other invocation, including a real `remote show origin`
#     when no file is given, to the real git unmodified.
shim_git() {
  SHOW_FILE="${1:-}"
  SHIM_DIR="$(mktemp -d "${TMPDIR:-/tmp}/resolve-base-branch-shim.XXXXXX")"
  REPOS+=("$SHIM_DIR")
  {
    printf '#!/usr/bin/env bash\n'
    printf 'case "$*" in\n'
    printf '  *"fetch --quiet origin"*)\n'
    printf '    exit 0 ;;\n'
    if [ -n "$SHOW_FILE" ]; then
      printf '  *"remote show origin"*)\n'
      printf '    cat "%s"\n' "$SHOW_FILE"
      printf '    exit 0 ;;\n'
    fi
    printf 'esac\n'
    printf 'exec "%s" "$@"\n' "$REAL_GIT"
  } > "$SHIM_DIR/git"
  chmod +x "$SHIM_DIR/git"
}

# 1. Clone with origin/HEAD set: the symbolic-ref path answers directly.
new_worktree
run_guard "$WT"
expect_success "origin/HEAD set" "main"

# 2. origin/HEAD deleted: the `git remote show origin` fallback answers.
#    The guard's own fetch is shimmed to a no-op so the deletion survives
#    into the resolution step; `remote show origin` itself is the real
#    command, run against the real remote.
new_worktree_no_head
shim_git
PATH="$SHIM_DIR:$PATH" run_guard "$WT"
expect_success "origin/HEAD deleted, remote show answers" "main"

# 3. Detached HEAD: no current branch to compare against.
new_worktree
git -C "$WT" checkout -q --detach main
run_guard "$WT"
expect_refusal "detached HEAD" 1 "detached"

# 4. Base resolves to the checked-out branch: refusing to compare a branch
#    with itself. `main` is both the resolved base and the checked-out
#    branch.
new_worktree
git -C "$WT" checkout -q main
run_guard "$WT"
expect_refusal "base equals current branch" 1 "same as the current branch"

# 5. Neither resolution path answers: origin/HEAD is gone, and the remote
#    show fallback is shimmed to a response with no `HEAD branch:` line at
#    all.
new_worktree_no_head
NO_HEAD_LINE="$(mktemp "${TMPDIR:-/tmp}/resolve-base-branch-remote.XXXXXX")"
REPOS+=("$NO_HEAD_LINE")
printf '* remote origin\n  Fetch URL: %s\n  Remote branch:\n    main tracked\n' "$REMOTE" > "$NO_HEAD_LINE"
shim_git "$NO_HEAD_LINE"
PATH="$SHIM_DIR:$PATH" run_guard "$WT"
expect_refusal "neither resolution path answers" 1 "could not resolve"

# 6. No origin remote configured at all.
new_worktree_no_origin
run_guard "$WT"
expect_refusal "no origin remote" 3 "no 'origin' remote"

# 7. <dir> is a plain directory, not a git worktree.
PLAIN_DIR="$(mktemp -d "${TMPDIR:-/tmp}/resolve-base-branch-test.XXXXXX")"
REPOS+=("$PLAIN_DIR")
run_guard "$PLAIN_DIR"
expect_refusal "plain directory, not a worktree" 2 "not a git worktree"

# 8. <dir> does not exist.
MISSING_DIR="${TMPDIR:-/tmp}/resolve-base-branch-test-missing.$$"
run_guard "$MISSING_DIR"
expect_refusal "directory does not exist" 2 "not a directory"

# 9. No argument at all: exit 2, usage line on stderr.
run_guard
expect_refusal "no argument" 2 "usage: resolve-base-branch.sh"

# 10. Remote reports a HEAD branch value of `-x` — a leading `-` a downstream
#     git call would read as an option.
new_worktree_no_head
DASH_LINE="$(mktemp "${TMPDIR:-/tmp}/resolve-base-branch-remote.XXXXXX")"
REPOS+=("$DASH_LINE")
printf '* remote origin\n  HEAD branch: -x\n' > "$DASH_LINE"
shim_git "$DASH_LINE"
PATH="$SHIM_DIR:$PATH" run_guard "$WT"
expect_refusal "hostile HEAD branch value: leading dash" 1 "invalid"

# 11. Remote reports a HEAD branch value carrying a control character.
new_worktree_no_head
CTRL_LINE="$(mktemp "${TMPDIR:-/tmp}/resolve-base-branch-remote.XXXXXX")"
REPOS+=("$CTRL_LINE")
printf '* remote origin\n  HEAD branch: bad\x01name\n' > "$CTRL_LINE"
shim_git "$CTRL_LINE"
PATH="$SHIM_DIR:$PATH" run_guard "$WT"
expect_refusal "hostile HEAD branch value: control character" 1 "invalid"

# 12. Remote reports a HEAD branch value carrying a non-ASCII byte, run under
#     a UTF-8 locale. `A-Za-z0-9` inside a `case` bracket expression is a
#     COLLATING range, not a byte range: under en_US.UTF-8 on bash 3.2, a
#     value like this — "mainéx" as UTF-8 bytes — collates inside that range
#     and both validation blocks would admit it if the guard did not pin its
#     own locale. This case proves the guard refuses it regardless of the
#     caller's ambient locale, not merely under LC_ALL=C.
new_worktree_no_head
NONASCII_LINE="$(mktemp "${TMPDIR:-/tmp}/resolve-base-branch-remote.XXXXXX")"
REPOS+=("$NONASCII_LINE")
printf '* remote origin\n  HEAD branch: main\xc3\xa9x\n' > "$NONASCII_LINE"
shim_git "$NONASCII_LINE"
LC_ALL=en_US.UTF-8 PATH="$SHIM_DIR:$PATH" GUARD_INTERP=/bin/bash run_guard "$WT"
expect_refusal "hostile HEAD branch value: non-ASCII byte under a UTF-8 locale" 1 "invalid"

# 13. HEAD's own ref is unreadable (e.g. corrupt, or permission-denied): the
#     `branch --show-current` call itself fails (non-zero exit), which is a
#     different fact from detached HEAD, where the same command succeeds and
#     simply prints nothing. Refused as its own case, exit 2 ("cannot
#     answer"), never folded into the exit-1 detached-HEAD refusal above.
#     `chmod 000` on the checked-out branch's loose ref file reproduces it,
#     which this environment cannot decide for two reasons: it requires
#     running as a non-root user, since root ignores file permission bits,
#     and it requires a git that stores the branch as a loose ref file at
#     all — a git using the packed-refs or reftable backend leaves nothing
#     at REF_FILE to chmod. Skipped rather than asserted when either holds,
#     matching test-check-cleanup-complete.sh's own unreadable-ref case.
new_worktree
REF_FILE="$WT/.git/refs/heads/openspec/fixture"
if [ ! -f "$REF_FILE" ]; then
  skip "current branch ref unreadable" \
    "this git stores the ref packed rather than loose"
else
  chmod 000 "$REF_FILE"
  if [ -r "$REF_FILE" ]; then
    chmod 644 "$REF_FILE"
    skip "current branch ref unreadable" \
      "this process can read a mode-000 file"
  else
    run_guard "$WT"
    chmod 644 "$REF_FILE"
    expect_refusal "current branch ref unreadable" 2 "could not read the current branch"
  fi
fi

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'resolve-base-branch: all cases pass\n'
