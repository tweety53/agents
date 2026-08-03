#!/usr/bin/env bash
# Assertion harness for checkpoint and uncommitted-review-package (both in
# skills/myflow-do/scripts/). Builds a throwaway git repo per test case under a
# sandboxed TMPDIR and asserts on their output; never touches the real
# repository tree.
#
# WHY THIS SUITE EXISTS. In the NO-COMMITS pipeline, nothing the implementer
# does ever reaches HEAD, so a review round has no base..head commit range to
# diff. `checkpoint` prints a commit-ish that snapshots the working tree (a
# stash-create object, or HEAD on a clean tree) — it stages the tree first
# (excluding the NO-COMMITS planning paths) so untracked files are captured
# too, which is a documented index-mutating side effect, not a no-op — and
# `uncommitted-review-package` diffs the *current* working tree against a
# `checkpoint` sha, so a reviewer sees only the work done since that
# checkpoint was taken. Case 4.4 (isolation) proves the one property this
# whole change exists to provide: a second round's package must show round
# two's diff and NOT round one's, even though round one is still sitting
# uncommitted in the same working tree.
#
# Conventions follow test-check-unfinished-work.sh: `set -euo pipefail`, an
# indexed-array SANDBOXES trap-cleaned on EXIT, `mktemp -d
# "${TMPDIR:-/tmp}/<name>-test.XXXXXX"`, fail/pass helpers, a FAILURES
# counter, exit 1 if FAILURES is non-zero at the end.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHECKPOINT="$SCRIPT_DIR/../skills/myflow-do/scripts/checkpoint"
REVIEW_PKG="$SCRIPT_DIR/../skills/myflow-do/scripts/uncommitted-review-package"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# An indexed array, not a space-separated string: sandbox paths come from
# mktemp under TMPDIR, which may contain spaces, and word-splitting a string
# would then rm -rf the fragments.
SANDBOXES=()
cleanup() {
  # ${SANDBOXES[@]} is unset-expansion-unsafe under `set -u` on bash 3.2 when empty.
  [ "${#SANDBOXES[@]}" -eq 0 ] && return 0
  for s in "${SANDBOXES[@]}"; do rm -rf "$s"; done
}
trap cleanup EXIT

WORK="$(mktemp -d "${TMPDIR:-/tmp}/uncommitted-review-package-test.XXXXXX")"
SANDBOXES+=("$WORK")
ERRFILE="$WORK/stderr"

# 4.1 setup: a throwaway git repo per test case, never the real repository
# tree. An initial empty commit gives HEAD something to resolve to before any
# case makes its own changes.
#
# new_repo does NOT register its own result into SANDBOXES: every call site
# assigns `repo="$(new_repo)"`, and command substitution always runs the
# function in a subshell, so an append to SANDBOXES done inside new_repo is
# lost the moment the subshell exits (`ARR=(); f(){ local x=one; ARR+=("$x");
# }; y="$(f)"; echo "${#ARR[@]}"` prints `0`). Each call site registers the
# returned path itself, in the parent shell, right after the assignment.
new_repo() {
  local repo
  repo="$(mktemp -d "${TMPDIR:-/tmp}/uncommitted-review-package-test.XXXXXX")"
  git -C "$repo" init -q
  git -C "$repo" config user.email "test@example.com"
  git -C "$repo" config user.name "Test"
  git -C "$repo" commit -q --allow-empty -m "initial"
  printf '%s' "$repo"
}

# run_checkpoint <repo> -> sets OUT (stdout), RC. checkpoint is run with the
# repo as cwd since it operates on the current git repository via bare `git`
# commands, with no repo-path argument of its own.
run_checkpoint() {
  local repo="$1"
  set +e
  OUT="$(cd "$repo" && "$CHECKPOINT" 2>"$ERRFILE")"
  RC=$?
  set -e
}

# run_review_pkg <repo> <plan> <base> [outfile] -> sets OUT, ERR, RC. Same
# cwd convention as run_checkpoint: uncommitted-review-package diffs "the
# current git repository", which is whatever repo the caller's cwd is in.
run_review_pkg() {
  local repo="$1"
  shift
  set +e
  OUT="$(cd "$repo" && "$REVIEW_PKG" "$@" 2>"$ERRFILE")"
  RC=$?
  set -e
  ERR="$(cat "$ERRFILE")"
}

# 4.2 Clean-tree checkpoint prints the current HEAD sha — no dirty state to
# snapshot, so `git stash create` prints nothing and checkpoint falls back to
# `git rev-parse HEAD`.
repo="$(new_repo)"
SANDBOXES+=("$repo")
head_sha="$(git -C "$repo" rev-parse HEAD)"
run_checkpoint "$repo"
if [ "$RC" -eq 0 ] && [ "$OUT" = "$head_sha" ]; then
  pass "a clean tree's checkpoint equals HEAD"
else
  fail "a clean tree's checkpoint: expected HEAD ($head_sha), got rc=$RC out=$OUT"
fi

# 4.3 Dirty-tree checkpoint prints a commit-ish that is not HEAD — an
# unstaged, tracked modification gives `git stash create` something to
# snapshot.
repo="$(new_repo)"
SANDBOXES+=("$repo")
printf 'a\n' > "$repo/file.txt"
git -C "$repo" add file.txt
git -C "$repo" commit -q -m "add file.txt"
head_sha="$(git -C "$repo" rev-parse HEAD)"
printf 'a\nb\n' > "$repo/file.txt"
run_checkpoint "$repo"
if [ "$RC" -eq 0 ] && [ -n "$OUT" ] && [ "$OUT" != "$head_sha" ]; then
  pass "a dirty tree's checkpoint differs from HEAD"
else
  fail "a dirty tree's checkpoint: expected a sha != HEAD ($head_sha), got rc=$RC out=$OUT"
fi

# 4.4 Two-round isolation — THE PROPERTY THIS CHANGE EXISTS TO PROVIDE. Round
# one's change (A) is captured into BASE by checkpoint without ever leaving
# the working tree; round two's change (B) is made on top of it; the written
# package must show B and must NOT show A, even though A is still sitting
# uncommitted in the same working tree as B when the package is generated.
#
# A and B are placed more than 10 lines apart (the diff's -U10 context width)
# so that B's context lines cannot incidentally include A's marker line —
# that would look like a leak without being one.
repo="$(new_repo)"
SANDBOXES+=("$repo")
seq 1 30 | sed 's/^/line /' > "$repo/file.txt"
git -C "$repo" add file.txt
git -C "$repo" commit -q -m "add file.txt"
sed -i.bak '1s/.*/CHANGE_A_MARKER/' "$repo/file.txt"
rm -f "$repo/file.txt.bak"
base="$(cd "$repo" && "$CHECKPOINT")"
printf 'CHANGE_B_MARKER\n' >> "$repo/file.txt"
plan="$repo/plan.md"
printf '# plan\n' > "$plan"
pkg_out="$repo/review.diff"
run_review_pkg "$repo" "$plan" "$base" "$pkg_out"
if [ "$RC" -ne 0 ]; then
  fail "two-round isolation: uncommitted-review-package exited rc=$RC err=$ERR"
else
  if grep -q "CHANGE_B_MARKER" "$pkg_out"; then
    pass "the written package contains round two's change"
  else
    fail "two-round isolation: package does not contain CHANGE_B_MARKER"
  fi
  if grep -q "CHANGE_A_MARKER" "$pkg_out"; then
    fail "two-round isolation: package leaks round one's change (CHANGE_A_MARKER) which BASE already covers"
  else
    pass "the written package does not contain round one's change"
  fi
fi

# 4.5 The written package's shape: a Files changed section, a Diff section,
# and no Commits section — there is no commit range to log under NO-COMMITS.
repo="$(new_repo)"
SANDBOXES+=("$repo")
printf 'original\n' > "$repo/file.txt"
git -C "$repo" add file.txt
git -C "$repo" commit -q -m "add file.txt"
base="$(cd "$repo" && "$CHECKPOINT")"
printf 'more\n' >> "$repo/file.txt"
plan="$repo/plan.md"
printf '# plan\n' > "$plan"
pkg_out="$repo/review.diff"
run_review_pkg "$repo" "$plan" "$base" "$pkg_out"
if [ "$RC" -ne 0 ]; then
  fail "package shape: uncommitted-review-package exited rc=$RC err=$ERR"
else
  if grep -q "^## Files changed$" "$pkg_out"; then
    pass "the written package has a Files changed section"
  else
    fail "package shape: no '## Files changed' line in $pkg_out"
  fi
  if grep -q "^## Diff$" "$pkg_out"; then
    pass "the written package has a Diff section"
  else
    fail "package shape: no '## Diff' line in $pkg_out"
  fi
  if grep -q "^## Commits$" "$pkg_out"; then
    fail "package shape: a '## Commits' line is present, but NO-COMMITS has no commit range to log"
  else
    pass "the written package has no Commits section"
  fi
fi

# 4.6 A nonexistent plan file is a programmer error: exit non-zero, name the
# problem on stderr, and write no output file.
repo="$(new_repo)"
SANDBOXES+=("$repo")
plan="$repo/no-such-plan.md"
pkg_out="$repo/review.diff"
run_review_pkg "$repo" "$plan" "HEAD" "$pkg_out"
if [ "$RC" -ne 0 ]; then
  pass "a nonexistent plan file exits non-zero"
else
  fail "a nonexistent plan file: expected non-zero exit, got rc=$RC"
fi
case "$ERR" in
  *"no such plan file:"*) pass "a nonexistent plan file names the failure on stderr" ;;
  *) fail "a nonexistent plan file: no 'no such plan file:' message on stderr: $ERR" ;;
esac
if [ -e "$pkg_out" ]; then
  fail "a nonexistent plan file: an output file was written anyway ($pkg_out)"
else
  pass "a nonexistent plan file writes no output file"
fi

# 4.7 A BASE that does not resolve is a programmer error: exit non-zero, name
# the problem on stderr, and write no output file.
repo="$(new_repo)"
SANDBOXES+=("$repo")
plan="$repo/plan.md"
printf '# plan\n' > "$plan"
pkg_out="$repo/review.diff"
run_review_pkg "$repo" "$plan" "not-a-real-commit-ish" "$pkg_out"
if [ "$RC" -ne 0 ]; then
  pass "a BASE that does not resolve exits non-zero"
else
  fail "a bad BASE: expected non-zero exit, got rc=$RC"
fi
case "$ERR" in
  *"bad BASE:"*) pass "a BASE that does not resolve names the failure on stderr" ;;
  *) fail "a bad BASE: no 'bad BASE:' message on stderr: $ERR" ;;
esac
if [ -e "$pkg_out" ]; then
  fail "a bad BASE: an output file was written anyway ($pkg_out)"
else
  pass "a bad BASE writes no output file"
fi

# 4.9 A re-review after a fix round gets a distinct file (task-group-4 review
# flagged this spec scenario as untested): round one's package is written to
# an explicit OUTFILE; round two makes a further change, checkpoints again (a
# distinct BASE), and writes to a distinct explicit OUTFILE. The new file must
# contain round two's content and the prior file must be byte-for-byte
# untouched.
repo="$(new_repo)"
SANDBOXES+=("$repo")
printf 'original\n' > "$repo/file.txt"
git -C "$repo" add file.txt
git -C "$repo" commit -q -m "add file.txt"
base1="$(cd "$repo" && "$CHECKPOINT")"
printf 'CHANGE_ROUND1_MARKER\n' >> "$repo/file.txt"
plan="$repo/plan.md"
printf '# plan\n' > "$plan"
pkg_out1="$repo/review-round1.diff"
run_review_pkg "$repo" "$plan" "$base1" "$pkg_out1"
if [ "$RC" -ne 0 ]; then
  fail "distinct-outfile re-review: round one uncommitted-review-package exited rc=$RC err=$ERR"
else
  pass "round one's package is written"
fi
round1_snapshot="$repo/review-round1.diff.snapshot"
cp "$pkg_out1" "$round1_snapshot"

base2="$(cd "$repo" && "$CHECKPOINT")"
printf 'CHANGE_ROUND2_MARKER\n' >> "$repo/file.txt"
pkg_out2="$repo/review-round2.diff"
run_review_pkg "$repo" "$plan" "$base2" "$pkg_out2"
if [ "$RC" -ne 0 ]; then
  fail "distinct-outfile re-review: round two uncommitted-review-package exited rc=$RC err=$ERR"
else
  if [ -e "$pkg_out2" ] && grep -q "CHANGE_ROUND2_MARKER" "$pkg_out2"; then
    pass "round two's package is written to its own distinct OUTFILE with round two's content"
  else
    fail "distinct-outfile re-review: $pkg_out2 missing or lacks CHANGE_ROUND2_MARKER"
  fi
fi
if [ -e "$pkg_out1" ]; then
  if cmp -s "$pkg_out1" "$round1_snapshot"; then
    pass "round one's package file is left byte-for-byte intact after round two"
  else
    fail "distinct-outfile re-review: round one's package file changed after round two ($pkg_out1 vs $round1_snapshot)"
  fi
else
  fail "distinct-outfile re-review: round one's package file $pkg_out1 no longer exists"
fi

# 4.10 checkpoint captures brand-new untracked files (Finding 1 regression
# guard): `git stash create` alone never sees untracked files unless they are
# staged, so a checkpoint taken right after a brand-new file appears must
# differ from a checkpoint taken before that file existed — otherwise a later
# review package would stage the file itself and misattribute round one's
# untracked file to whichever round happens to run second.
repo="$(new_repo)"
SANDBOXES+=("$repo")
head_sha="$(git -C "$repo" rev-parse HEAD)"
checkpoint1="$(cd "$repo" && "$CHECKPOINT")"
if [ "$checkpoint1" = "$head_sha" ]; then
  pass "checkpoint on a clean tree still equals HEAD before the untracked file appears"
else
  fail "checkpoint on a clean tree: expected HEAD ($head_sha), got $checkpoint1"
fi
printf 'UNTRACKED_ROUND_ONE_MARKER\n' > "$repo/new-file.txt"
checkpoint2="$(cd "$repo" && "$CHECKPOINT")"
if [ -n "$checkpoint2" ] && [ "$checkpoint2" != "$head_sha" ]; then
  pass "checkpoint captures a brand-new untracked file (differs from the checkpoint taken before it existed)"
else
  fail "checkpoint did not capture the untracked file: got checkpoint2=$checkpoint2 (same as HEAD $head_sha)"
fi
printf 'TRACKED_ROUND_TWO_MARKER\n' > "$repo/second-file.txt"
plan="$repo/plan.md"
printf '# plan\n' > "$plan"
pkg_out="$repo/review.diff"
run_review_pkg "$repo" "$plan" "$checkpoint2" "$pkg_out"
if [ "$RC" -ne 0 ]; then
  fail "untracked-file isolation: uncommitted-review-package exited rc=$RC err=$ERR"
else
  if grep -q "TRACKED_ROUND_TWO_MARKER" "$pkg_out"; then
    pass "the written package contains round two's new file"
  else
    fail "untracked-file isolation: package does not contain TRACKED_ROUND_TWO_MARKER"
  fi
  if grep -q "UNTRACKED_ROUND_ONE_MARKER" "$pkg_out"; then
    fail "untracked-file isolation: package leaks round one's untracked file, which checkpoint2 already covers"
  else
    pass "the written package does not leak round one's untracked file"
  fi
fi

# 4.11 checkpoint never stages the NO-COMMITS planning paths (Finding 2):
# openspec/, docs/manual-test/, docs/superpowers/ must never enter the index
# via checkpoint's staging step, even though everything else in the repo —
# including a brand-new untracked file in an unrelated subdirectory, and even
# when checkpoint is invoked from a cwd other than the repo root — must.
repo="$(new_repo)"
SANDBOXES+=("$repo")
mkdir -p "$repo/openspec/changes/x" "$repo/docs/manual-test" "$repo/docs/superpowers" "$repo/src/sub"
printf 'openspec change\n' > "$repo/openspec/changes/x/tasks.md"
printf 'manual test notes\n' > "$repo/docs/manual-test/notes.md"
printf 'superpowers spec\n' > "$repo/docs/superpowers/spec.md"
printf 'impl change\n' > "$repo/src/sub/impl.txt"
(cd "$repo/src/sub" && "$CHECKPOINT" >/dev/null)
staged="$(git -C "$repo" diff --cached --name-only)"
if printf '%s\n' "$staged" | grep -q '^openspec/'; then
  fail "checkpoint staged openspec/ despite NO-COMMITS: $staged"
else
  pass "checkpoint did not stage openspec/"
fi
if printf '%s\n' "$staged" | grep -q '^docs/manual-test/'; then
  fail "checkpoint staged docs/manual-test/ despite NO-COMMITS: $staged"
else
  pass "checkpoint did not stage docs/manual-test/"
fi
if printf '%s\n' "$staged" | grep -q '^docs/superpowers/'; then
  fail "checkpoint staged docs/superpowers/ despite NO-COMMITS: $staged"
else
  pass "checkpoint did not stage docs/superpowers/"
fi
if printf '%s\n' "$staged" | grep -q '^src/sub/impl.txt$'; then
  pass "checkpoint staged an untracked file elsewhere in the repo, invoked from a subdirectory cwd"
else
  fail "checkpoint did not stage src/sub/impl.txt from a subdirectory cwd (staged: $staged)"
fi

# 4.12 uncommitted-review-package never stages (or diffs) the NO-COMMITS
# planning paths, invoked from a subdirectory (Finding 2, plus the
# cwd-scoping regression check: a `-- .` pathspec run without `-C` would only
# cover the invocation cwd's subtree, so this also proves the repo-root
# anchoring still works when called from src/sub rather than the repo root).
repo="$(new_repo)"
SANDBOXES+=("$repo")
mkdir -p "$repo/openspec/changes/x" "$repo/docs/manual-test" "$repo/docs/superpowers" "$repo/src/sub"
printf 'tracked\n' > "$repo/src/sub/base.txt"
git -C "$repo" add src/sub/base.txt
git -C "$repo" commit -q -m "add base"
base="$(cd "$repo" && "$CHECKPOINT")"
printf 'openspec change\n' > "$repo/openspec/changes/x/tasks.md"
printf 'manual test notes\n' > "$repo/docs/manual-test/notes.md"
printf 'superpowers spec\n' > "$repo/docs/superpowers/spec.md"
printf 'IMPL_MARKER\n' > "$repo/src/sub/new-impl.txt"
plan="$repo/plan.md"
printf '# plan\n' > "$plan"
pkg_out="$repo/review.diff"
run_review_pkg "$repo/src/sub" "$plan" "$base" "$pkg_out"
if [ "$RC" -ne 0 ]; then
  fail "planning-path exclusion: uncommitted-review-package exited rc=$RC err=$ERR"
else
  if grep -q "IMPL_MARKER" "$pkg_out"; then
    pass "the written package contains an untracked implementation file from a subdirectory invocation"
  else
    fail "planning-path exclusion: package does not contain IMPL_MARKER"
  fi
  if grep -q "openspec/changes" "$pkg_out"; then
    fail "planning-path exclusion: package includes openspec/ despite NO-COMMITS"
  else
    pass "the written package does not include openspec/"
  fi
  if grep -q "docs/manual-test" "$pkg_out"; then
    fail "planning-path exclusion: package includes docs/manual-test/ despite NO-COMMITS"
  else
    pass "the written package does not include docs/manual-test/"
  fi
  if grep -q "docs/superpowers" "$pkg_out"; then
    fail "planning-path exclusion: package includes docs/superpowers/ despite NO-COMMITS"
  else
    pass "the written package does not include docs/superpowers/"
  fi
fi
staged="$(git -C "$repo" diff --cached --name-only)"
if printf '%s\n' "$staged" | grep -qE '^(openspec/|docs/manual-test/|docs/superpowers/)'; then
  fail "planning-path exclusion: uncommitted-review-package staged an excluded path: $staged"
else
  pass "uncommitted-review-package's own staging step did not touch the excluded paths"
fi

# 4.13 the default-OUTFILE code path (Finding 3 regression guard): every case
# above passes an explicit OUTFILE, but SKILL.md's documented per-task call
# omits it (`uncommitted-review-package <plan-file> "$TASK_BASE"`, 2 args).
# subagent-driven-development is not vendored in this repo, so the sibling
# relative path never resolves and the harness-cache fallback loop genuinely
# fires against this machine's real Claude Code plugin cache.
repo="$(new_repo)"
SANDBOXES+=("$repo")
printf 'original\n' > "$repo/file.txt"
git -C "$repo" add file.txt
git -C "$repo" commit -q -m "add file.txt"
base="$(cd "$repo" && "$CHECKPOINT")"
printf 'more\n' >> "$repo/file.txt"
plan="$repo/my-plan.md"
printf '# plan\n' > "$plan"
run_review_pkg "$repo" "$plan" "$base"
if [ "$RC" -eq 0 ]; then
  pass "the default-OUTFILE call (2 args, no explicit OUTFILE) succeeds"
else
  fail "the default-OUTFILE call: exited rc=$RC err=$ERR"
fi
base7="$(git -C "$repo" rev-parse --short "$base")"
# sdd-workspace resolves the repo root via `git rev-parse --show-toplevel`,
# which canonicalizes symlinks (e.g. macOS's /tmp -> /private/tmp); $repo
# itself may not be canonicalized the same way, so canonicalize it the same
# way before building the expected path, or the comparison is a path-shape
# false failure rather than a real one.
canonical_repo="$(cd "$repo" && git rev-parse --show-toplevel)"
expected="$canonical_repo/.superpowers/sdd/my-plan/review-$base7.diff"
if [ -e "$expected" ]; then
  pass "the default OUTFILE is written at <repo-root>/.superpowers/sdd/<plan-basename>/review-<base7>.diff"
else
  fail "the default OUTFILE: expected file at $expected, not found (stdout: $OUT)"
fi
case "$OUT" in
  "wrote $expected: "*" bytes")
    pass "stdout matches the documented 'wrote <path>: <bytes> bytes' format" ;;
  *)
    fail "stdout does not match 'wrote <path>: <bytes> bytes' format: $OUT" ;;
esac

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'uncommitted-review-package: all cases pass\n'
