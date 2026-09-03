#!/usr/bin/env bash
# Assertion harness for check-panel-docs-only.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR and asserts the guard's exit code
# (and, where named, its stdout). Never touches the real repository tree.
# Same shape as test-check-panel-citation-trigger.sh: an indexed REPOS
# array (bash 3.2 has no associative arrays), removed by an EXIT trap,
# real git repositories rather than fixture trees.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the contract
# stated in tasks.md's task 1, never against observed output.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-panel-docs-only.sh"
# LIB — the GIT_BIN resolution, worktree/merge-base validation and
# COMMITTED/STAGED/UNSTAGED collection this guard sources (KAN-312 review
# round 0, F1): the guard's own file carries only its classification logic
# now, so the structural call-site check below scans $LIB too.
LIB="$SCRIPT_DIR/lib/panel-touched-paths.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

REPOS=()
cleanup() {
  [ "${#REPOS[@]}" -eq 0 ] && return 0
  for repo in "${REPOS[@]}"; do
    rm -rf "$repo"
  done
}
trap cleanup EXIT

# run_guard <worktree> <merge-base> -> sets RC, OUT
run_guard() {
  set +e
  OUT="$("$GUARD" "$1" "$2" 2>&1)"
  RC=$?
  set -e
}

# new_repo -> sets REPO, MERGEBASE
# A repository on `main` with one commit carrying base.go and tracked.md —
# the merge base every case starts from. tracked.md is tracked from the
# start so an unstaged-only edit is a modification `git diff` detects
# without an intervening `git add` — a brand-new untracked file would not
# appear in either the staged or unstaged diff at all.
new_repo() {
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/docs-only-test.XXXXXX")"
  REPOS+=("$REPO")
  git -C "$REPO" init -q -b main
  git -C "$REPO" config user.email test@example.invalid
  git -C "$REPO" config user.name "Test"
  echo base > "$REPO/base.go"
  echo "# tracked" > "$REPO/tracked.md"
  git -C "$REPO" add base.go tracked.md
  git -C "$REPO" commit -qm "base"
  MERGEBASE="$(git -C "$REPO" rev-parse HEAD)"
}

# 1. committed .md change only -> exit 0, empty stdout.
new_repo
echo "# notes" > "$REPO/notes.md"
git -C "$REPO" add notes.md
git -C "$REPO" commit -qm "add notes.md"
run_guard "$REPO" "$MERGEBASE"
[ "$RC" -eq 0 ] && pass "committed .md change only -> exit 0" \
  || fail "committed .md change only: expected exit 0, got rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "committed .md change only -> empty stdout" \
  || fail "committed .md change only: expected empty stdout, got: $OUT"

# 2. committed .mdc change only -> exit 0, empty stdout.
new_repo
echo "rule" > "$REPO/rule.mdc"
git -C "$REPO" add rule.mdc
git -C "$REPO" commit -qm "add rule.mdc"
run_guard "$REPO" "$MERGEBASE"
[ "$RC" -eq 0 ] && pass "committed .mdc change only -> exit 0" \
  || fail "committed .mdc change only: expected exit 0, got rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "committed .mdc change only -> empty stdout" \
  || fail "committed .mdc change only: expected empty stdout, got: $OUT"

# 3. committed .md plus committed main.go -> exit 1, stdout is "main.go".
new_repo
echo "# notes" > "$REPO/notes.md"
echo "package main" > "$REPO/main.go"
git -C "$REPO" add notes.md main.go
git -C "$REPO" commit -qm "add notes.md and main.go"
run_guard "$REPO" "$MERGEBASE"
[ "$RC" -eq 1 ] && pass "committed .md plus committed main.go -> exit 1" \
  || fail "committed .md plus committed main.go: expected exit 1, got rc=$RC out=$OUT"
[ "$OUT" = "main.go" ] && pass "committed .md plus committed main.go -> stdout is main.go" \
  || fail "committed .md plus committed main.go: expected stdout 'main.go', got: $OUT"

# 4. committed .md plus staged main.go (not committed) -> exit 1, stdout is "main.go".
new_repo
echo "# notes" > "$REPO/notes.md"
git -C "$REPO" add notes.md
git -C "$REPO" commit -qm "add notes.md"
echo "package main" > "$REPO/main.go"
git -C "$REPO" add main.go
run_guard "$REPO" "$MERGEBASE"
[ "$RC" -eq 1 ] && pass "committed .md plus staged main.go -> exit 1" \
  || fail "committed .md plus staged main.go: expected exit 1, got rc=$RC out=$OUT"
[ "$OUT" = "main.go" ] && pass "committed .md plus staged main.go -> stdout is main.go" \
  || fail "committed .md plus staged main.go: expected stdout 'main.go', got: $OUT"

# 5. committed .md plus unstaged edit to base.go -> exit 1, stdout is "base.go".
new_repo
echo "# notes" > "$REPO/notes.md"
git -C "$REPO" add notes.md
git -C "$REPO" commit -qm "add notes.md"
echo "unstaged edit" >> "$REPO/base.go"
run_guard "$REPO" "$MERGEBASE"
[ "$RC" -eq 1 ] && pass "committed .md plus unstaged edit to base.go -> exit 1" \
  || fail "committed .md plus unstaged edit to base.go: expected exit 1, got rc=$RC out=$OUT"
[ "$OUT" = "base.go" ] && pass "committed .md plus unstaged edit to base.go -> stdout is base.go" \
  || fail "committed .md plus unstaged edit to base.go: expected stdout 'base.go', got: $OUT"

# 6. staged-only .md change, nothing committed -> exit 0.
new_repo
echo "# staged" > "$REPO/staged.md"
git -C "$REPO" add staged.md
run_guard "$REPO" "$MERGEBASE"
[ "$RC" -eq 0 ] && pass "staged-only .md change -> exit 0" \
  || fail "staged-only .md change: expected exit 0, got rc=$RC out=$OUT"

# 7. unstaged-only edit to tracked.md -> exit 0.
new_repo
echo "unstaged edit" >> "$REPO/tracked.md"
run_guard "$REPO" "$MERGEBASE"
[ "$RC" -eq 0 ] && pass "unstaged-only edit to tracked.md -> exit 0" \
  || fail "unstaged-only edit to tracked.md: expected exit 0, got rc=$RC out=$OUT"

# 8. no changes at all -> exit 1, empty stdout.
new_repo
run_guard "$REPO" "$MERGEBASE"
[ "$RC" -eq 1 ] && pass "no changes at all -> exit 1" \
  || fail "no changes at all: expected exit 1, got rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "no changes at all -> empty stdout" \
  || fail "no changes at all: expected empty stdout, got: $OUT"

# 9. missing arguments -> exit 2.
run_guard "" ""
[ "$RC" -eq 2 ] && pass "missing arguments -> exit 2" \
  || fail "missing arguments: expected exit 2, got rc=$RC out=$OUT"

# 10. <worktree> not a directory -> exit 2.
NOT_A_DIR="$(mktemp -u "${TMPDIR:-/tmp}/docs-only-test-nodir.XXXXXX")"
run_guard "$NOT_A_DIR" deadbeef
[ "$RC" -eq 2 ] && pass "worktree not a directory -> exit 2" \
  || fail "worktree not a directory: expected exit 2, got rc=$RC out=$OUT"

# 11. <worktree> not a git repository -> exit 2.
NOT_A_REPO="$(mktemp -d "${TMPDIR:-/tmp}/docs-only-test-notrepo.XXXXXX")"
REPOS+=("$NOT_A_REPO")
run_guard "$NOT_A_REPO" deadbeef
[ "$RC" -eq 2 ] && pass "worktree not a git repository -> exit 2" \
  || fail "worktree not a git repository: expected exit 2, got rc=$RC out=$OUT"

# 12. <merge-base> not resolving -> exit 2.
new_repo
run_guard "$REPO" 0000000000000000000000000000000000000000
[ "$RC" -eq 2 ] && pass "merge base not resolving -> exit 2" \
  || fail "merge base not resolving: expected exit 2, got rc=$RC out=$OUT"

# 13. flag-shaped merge base "--evil" -> exit 2, OUT contains "does not resolve".
new_repo
run_guard "$REPO" "--evil"
[ "$RC" -eq 2 ] && pass "flag-shaped merge base ('--evil') -> exit 2" \
  || fail "flag-shaped merge base: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  *"does not resolve"*) pass "flag-shaped merge base: reported as not resolving" ;;
  *) fail "flag-shaped merge base: expected a 'does not resolve' message, got: $OUT" ;;
esac

# 14. structural: every `-C "$WORKTREE"` call site in the guard carries the
#     literal `"$GIT_BIN" -C "$WORKTREE"` (the positive check case 11b of
#     the citation harness uses).
CALL_SITE_LINES="$(grep -n -- '-C[[:space:]]*"\$WORKTREE"' "$GUARD" "$LIB" || true)"
NON_GITBIN_CALLS=0
while IFS= read -r line; do
  [ -z "$line" ] && continue
  case "$line" in
    *'"$GIT_BIN" -C "$WORKTREE"'*) ;;
    *) NON_GITBIN_CALLS=$((NON_GITBIN_CALLS + 1)) ;;
  esac
done <<EOF
$CALL_SITE_LINES
EOF
[ "$NON_GITBIN_CALLS" -eq 0 ] \
  && pass "every real call site is literally \"\$GIT_BIN\" -C \"\$WORKTREE\", never bare git" \
  || fail "a call site does not literally invoke \"\$GIT_BIN\" -C \"\$WORKTREE\""

# 15. mutation: replace the guard's `\.mdc?$` pattern with `\.never$` via
#     sed into a copy, run case 1 against the copy, assert it now exits 1;
#     then run case 3 against the copy, assert it still exits 1 — proving
#     the pattern is what decides case 1. Mutates a COPY, never the
#     checked-in file, so the harness leaves the tree untouched even when
#     interrupted.
if [ -f "$GUARD" ]; then
  MUTANT_REPO_DIR="$(mktemp -d "${TMPDIR:-/tmp}/docs-only-mutant.XXXXXX")"
  REPOS+=("$MUTANT_REPO_DIR")
  # The guard now sources "$SCRIPT_DIR/lib/panel-touched-paths.sh"
  # (KAN-312 review round 0, F1) — a mutant copied out with no sibling
  # lib/ would fail at `source` before ever reaching the `.mdc?$` pattern
  # this proof mutates, passing both assertions below vacuously (exit 1
  # from a missing file, not from the broken pattern). Copying lib/
  # alongside keeps this an honest proof of the pattern's effect.
  cp -R "$SCRIPT_DIR/lib" "$MUTANT_REPO_DIR/lib"

  new_repo
  MUTANT_GUARD_REPO="$REPO"
  sed 's/\\\.mdc?\$/\\.never$/' "$GUARD" > "$MUTANT_REPO_DIR/guard-mutant.sh"
  chmod +x "$MUTANT_REPO_DIR/guard-mutant.sh"

  echo "# notes" > "$MUTANT_GUARD_REPO/notes.md"
  git -C "$MUTANT_GUARD_REPO" add notes.md
  git -C "$MUTANT_GUARD_REPO" commit -qm "add notes.md"
  set +e
  OUT="$("$MUTANT_REPO_DIR/guard-mutant.sh" "$MUTANT_GUARD_REPO" "$MERGEBASE" 2>&1)"
  RC=$?
  set -e
  [ "$RC" -eq 1 ] && pass "mutation: case 1 flips to exit 1 with .mdc?\$ broken" \
    || fail "mutation: case 1 still exits $RC with .mdc?\$ broken (out=$OUT)"

  new_repo
  MUTANT_GUARD_REPO="$REPO"
  echo "# notes" > "$MUTANT_GUARD_REPO/notes.md"
  echo "package main" > "$MUTANT_GUARD_REPO/main.go"
  git -C "$MUTANT_GUARD_REPO" add notes.md main.go
  git -C "$MUTANT_GUARD_REPO" commit -qm "add notes.md and main.go"
  set +e
  OUT="$("$MUTANT_REPO_DIR/guard-mutant.sh" "$MUTANT_GUARD_REPO" "$MERGEBASE" 2>&1)"
  RC=$?
  set -e
  [ "$RC" -eq 1 ] && pass "mutation: case 3 still exits 1 with .mdc?\$ broken" \
    || fail "mutation: case 3 exits $RC with .mdc?\$ broken (out=$OUT)"
else
  fail "mutation proof skipped: $GUARD does not exist yet"
fi

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-panel-docs-only: all cases pass\n'
