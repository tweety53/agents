#!/usr/bin/env bash
# Assertion harness for mutate-and-verify.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR, each carrying one fixture "guard"
# script and one fixture test-*.sh harness for it (2-3+ cases, cheap and
# fast — fixtures for mutate-and-verify.sh, not real guards), and asserts
# mutate-and-verify.sh's exit code, report and restore behavior against
# them. Never touches the real repository tree. Same shape as
# test-check-base-moved.sh: an indexed REPOS array (bash 3.2 has no
# associative arrays), removed by an EXIT trap, real git repositories
# rather than fixture trees.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated
# contract in design.md and tasks.md's task 2, never against observed
# output.
#
# PROVED BY MUTATION, NOT MERELY ASSERTED. The "caught" and "suspicious
# blast radius" cases were checked by hand against a deliberately-broken
# scratch copy of mutate-and-verify.sh with the new-failures set-difference
# step (step 12) replaced by "always empty" — every fixture's mutation then
# read back as a surviving mutant, including the blast-radius fixture's
# broad breakage, and both cases below caught that (they no longer got
# `caught`/`suspicious blast radius`, they got `surviving mutant`).
# Restoring the real set-difference made both pass again. This is the same
# manual-mutation-and-revert method test-check-base-moved.sh's own header
# records; it is not reproduced automatically on every run.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MUTATE="$SCRIPT_DIR/mutate-and-verify.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

REPOS=()
cleanup() {
  [ "${#REPOS[@]}" -eq 0 ] && return 0
  for p in "${REPOS[@]}"; do
    rm -rf "$p"
  done
}
trap cleanup EXIT

# build_guard <file> <n-cases> [<mutation>] — writes a fixture "guard"
# script: `guard.sh <n>` exits 0 for even n, 1 for odd n, for n in
# 1..<n-cases>, and exits 1 for any n outside that range (the catch-all).
# <mutation> alters exactly one thing so a patch can be built as the diff
# between two calls with different mutations:
#   ""            baseline — correct behavior described above.
#   "flip:<i>"    case <i>'s own arm exits the WRONG code (breaks exactly
#                 the one fixture-harness case that tests n=<i>).
#   "catchall"    the catch-all arm exits 0 instead of 1 — never observed
#                 by the fixture harness, since it only calls n in
#                 1..<n-cases> (a surviving mutant).
#   "always-fail" an unconditional `exit 1` right after argument capture,
#                 before the case statement — breaks every even-n case
#                 (roughly half of <n-cases>), a broad blast radius when
#                 <n-cases> is large enough.
build_guard() {
  local file="$1" n_cases="$2" mutation="${3:-}"
  {
    echo '#!/usr/bin/env bash'
    echo 'n="$1"'
    if [ "$mutation" = "always-fail" ]; then
      echo 'exit 1'
    fi
    echo 'case "$n" in'
    local i want
    for i in $(seq 1 "$n_cases"); do
      if [ $((i % 2)) -eq 0 ]; then want=0; else want=1; fi
      if [ "$mutation" = "flip:$i" ]; then
        want=$((1 - want))
      fi
      printf '  %s) exit %s ;;\n' "$i" "$want"
    done
    if [ "$mutation" = "catchall" ]; then
      echo '  *) exit 0 ;;'
    else
      echo '  *) exit 1 ;;'
    fi
    echo 'esac'
  } >"$file"
  chmod +x "$file"
}

# build_harness <file> <guard-basename> <n-cases> — writes a fixture
# test-*.sh-shaped harness: calls guard.sh once per n in 1..<n-cases>,
# printing `ok: case-<n>` / `FAIL: case-<n>` per this repository's own
# convention.
build_harness() {
  local file="$1" guard_name="$2" n_cases="$3"
  {
    echo '#!/usr/bin/env bash'
    echo 'set -euo pipefail'
    echo 'DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"'
    printf 'GUARD="$DIR/%s"\n' "$guard_name"
    local i want
    for i in $(seq 1 "$n_cases"); do
      if [ $((i % 2)) -eq 0 ]; then want=0; else want=1; fi
      printf 'if "$GUARD" %s >/dev/null 2>&1; then got=0; else got=1; fi\n' "$i"
      printf 'if [ "$got" -eq %s ]; then echo "ok: case-%s"; else echo "FAIL: case-%s"; fi\n' \
        "$want" "$i" "$i"
    done
    echo 'exit 0'
  } >"$file"
  chmod +x "$file"
}

# new_fixture_repo <n-cases> -> sets REPO, GUARD, HARNESS. A repo on `main`
# with one commit carrying guard.sh and test-fixture.sh built by the
# helpers above.
new_fixture_repo() {
  local n_cases="$1"
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/mutate-verify-test.XXXXXX")"
  REPOS+=("$REPO")
  git -C "$REPO" init -q -b main
  git -C "$REPO" config user.email test@example.invalid
  git -C "$REPO" config user.name "Test"
  build_guard "$REPO/guard.sh" "$n_cases"
  build_harness "$REPO/test-fixture.sh" "guard.sh" "$n_cases"
  git -C "$REPO" add guard.sh test-fixture.sh
  git -C "$REPO" commit -qm "fixture"
  GUARD="$REPO/guard.sh"
  HARNESS="$REPO/test-fixture.sh"
}

# make_patch <repo> <n-cases> <mutation> -> sets PATCH (a tracked tempfile)
# to the diff between guard.sh's committed baseline and the same guard
# rebuilt with <mutation>. Leaves guard.sh back at its committed content
# (the patch is not applied — mutate-and-verify.sh applies it itself).
make_patch() {
  local repo="$1" n_cases="$2" mutation="$3"
  build_guard "$repo/guard.sh" "$n_cases" "$mutation"
  PATCH="$(mktemp "${TMPDIR:-/tmp}/mutate-verify-test-patch.XXXXXX")"
  REPOS+=("$PATCH")
  git -C "$repo" diff -- guard.sh >"$PATCH"
  git -C "$repo" checkout -q -- guard.sh
}

# run_mutate <repo> <patch> <harness>... -> sets OUT, RC. Invoked with cwd
# inside the repo so mutate-and-verify.sh's own `git rev-parse
# --show-toplevel` resolves it.
run_mutate() {
  local repo="$1" patch="$2"
  shift 2
  set +e
  OUT="$(cd "$repo" && "$MUTATE" "$patch" "$@" 2>&1)"
  RC=$?
  set -e
}

# repo_clean <repo> -> 0 if `git status --porcelain` is empty.
repo_clean() {
  [ -z "$(git -C "$1" status --porcelain)" ]
}

# 1. caught — a patch that flips exactly one fixture-harness case's
#    expected outcome.
new_fixture_repo 4
make_patch "$REPO" 4 "flip:2"
run_mutate "$REPO" "$PATCH" "$HARNESS"
[ "$RC" -eq 0 ] && pass "caught: exit 0" || fail "caught: expected exit 0, got rc=$RC out=$OUT"
case "$OUT" in
  *"caught"*) pass "caught: verdict names caught" ;;
  *) fail "caught: verdict missing 'caught': $OUT" ;;
esac
case "$OUT" in
  *"suspicious"*) fail "caught: wrongly also flagged suspicious blast radius: $OUT" ;;
  *) pass "caught: does not flag suspicious blast radius" ;;
esac
case "$OUT" in
  *"case-2"*) pass "caught: names case-2 as the new failure" ;;
  *) fail "caught: case-2 not named as the new failure: $OUT" ;;
esac
case "$OUT" in
  *"case-1"*"new failure"*|*"case-3"*"new failure"*|*"case-4"*"new failure"*)
    fail "caught: names an unrelated case as a new failure: $OUT" ;;
  *) pass "caught: names exactly one new failure" ;;
esac
repo_clean "$REPO" && pass "caught: touched file clean after restore" \
  || fail "caught: touched file left dirty: $(git -C "$REPO" status --porcelain)"

# 2. surviving mutant — a patch that changes the fixture guard's catch-all
#    arm, which no fixture-harness case observes.
new_fixture_repo 4
make_patch "$REPO" 4 "catchall"
run_mutate "$REPO" "$PATCH" "$HARNESS"
[ "$RC" -eq 0 ] && pass "surviving mutant: exit 0" \
  || fail "surviving mutant: expected exit 0, got rc=$RC out=$OUT"
case "$OUT" in
  *"surviving mutant"*) pass "surviving mutant: verdict names surviving mutant" ;;
  *) fail "surviving mutant: verdict missing 'surviving mutant': $OUT" ;;
esac
case "$OUT" in
  *"0 new failures"*) pass "surviving mutant: reports 0 new failures" ;;
  *) fail "surviving mutant: does not report 0 new failures: $OUT" ;;
esac
repo_clean "$REPO" && pass "surviving mutant: touched file clean after restore" \
  || fail "surviving mutant: touched file left dirty: $(git -C "$REPO" status --porcelain)"

# 3. suspicious blast radius — 12 fixture-harness cases (more than the
#    default bound of 5), and a patch that breaks the guard unconditionally
#    before its case statement, failing every even-numbered case (6 total).
new_fixture_repo 12
make_patch "$REPO" 12 "always-fail"
run_mutate "$REPO" "$PATCH" "$HARNESS"
[ "$RC" -eq 0 ] && pass "blast radius: exit 0" \
  || fail "blast radius: expected exit 0, got rc=$RC out=$OUT"
case "$OUT" in
  *"suspicious blast radius"*) pass "blast radius: verdict names suspicious blast radius" ;;
  *) fail "blast radius: verdict missing 'suspicious blast radius': $OUT" ;;
esac
case "$OUT" in
  *"6 new failures"*) pass "blast radius: reports the actual new-failure count (6)" ;;
  *) fail "blast radius: does not report 6 new failures: $OUT" ;;
esac
case "$OUT" in
  *"bound 5"*) pass "blast radius: names the default bound (5)" ;;
  *) fail "blast radius: does not name the default bound: $OUT" ;;
esac
repo_clean "$REPO" && pass "blast radius: touched file clean after restore" \
  || fail "blast radius: touched file left dirty: $(git -C "$REPO" status --porcelain)"

# 4. MUTATE_AND_VERIFY_MAX_NEW_FAILURES override — the same broad-breakage
#    patch as case 3, with the bound raised above the actual new-failure
#    count: verdict flips from suspicious blast radius to caught, proving
#    the bound is read from the environment rather than hardcoded.
new_fixture_repo 12
make_patch "$REPO" 12 "always-fail"
set +e
OUT="$(cd "$REPO" && MUTATE_AND_VERIFY_MAX_NEW_FAILURES=10 "$MUTATE" "$PATCH" "$HARNESS" 2>&1)"
RC=$?
set -e
[ "$RC" -eq 0 ] && pass "bound override: exit 0" \
  || fail "bound override: expected exit 0, got rc=$RC out=$OUT"
case "$OUT" in
  *"suspicious"*) fail "bound override: still flagged suspicious blast radius: $OUT" ;;
  *) pass "bound override: no longer flags suspicious blast radius" ;;
esac
case "$OUT" in
  *"caught"*) pass "bound override: verdict names caught" ;;
  *) fail "bound override: verdict missing 'caught': $OUT" ;;
esac
case "$OUT" in
  *"bound: 10"*|*"bound 10"*) pass "bound override: names the overridden bound (10)" ;;
  *) fail "bound override: does not name the overridden bound: $OUT" ;;
esac
repo_clean "$REPO" && pass "bound override: touched file clean after restore" \
  || fail "bound override: touched file left dirty: $(git -C "$REPO" status --porcelain)"

# 5. patch does not apply — a patch built against text guard.sh does not
#    contain: refuse before mutating anything, guard.sh byte-for-byte
#    unchanged, and no baseline/mutated report at all.
new_fixture_repo 4
BEFORE_SUM="$(shasum -a 256 "$GUARD" | cut -d' ' -f1)"
PATCH="$(mktemp "${TMPDIR:-/tmp}/mutate-verify-test-patch.XXXXXX")"
REPOS+=("$PATCH")
cat >"$PATCH" <<'EOF'
diff --git a/nonexistent.sh b/nonexistent.sh
index 0000000..1111111 100644
--- a/nonexistent.sh
+++ b/nonexistent.sh
@@ -1,1 +1,1 @@
-old line that is not present anywhere
+new line
EOF
run_mutate "$REPO" "$PATCH" "$HARNESS"
[ "$RC" -eq 2 ] && pass "patch does not apply: exit 2" \
  || fail "patch does not apply: expected exit 2, got rc=$RC out=$OUT"
AFTER_SUM="$(shasum -a 256 "$GUARD" | cut -d' ' -f1)"
[ "$BEFORE_SUM" = "$AFTER_SUM" ] && pass "patch does not apply: fixture file byte-for-byte unchanged" \
  || fail "patch does not apply: fixture file changed"
case "$OUT" in
  *"baseline pass"*|*"mutated pass"*)
    fail "patch does not apply: a baseline/mutated report was printed: $OUT" ;;
  *) pass "patch does not apply: no harness ran, no report printed" ;;
esac
repo_clean "$REPO" && pass "patch does not apply: repo stays clean" \
  || fail "patch does not apply: repo left dirty: $(git -C "$REPO" status --porcelain)"

# 6. dirty touched file refused — an uncommitted edit already sits on the
#    file the patch also touches: refuse, the uncommitted edit survives
#    unchanged, the patch is never applied.
new_fixture_repo 4
make_patch "$REPO" 4 "flip:2"
printf '# pre-existing uncommitted edit\n' >>"$GUARD"
DIRTY_CONTENT="$(cat "$GUARD")"
run_mutate "$REPO" "$PATCH" "$HARNESS"
[ "$RC" -eq 2 ] && pass "dirty touched file: exit 2" \
  || fail "dirty touched file: expected exit 2, got rc=$RC out=$OUT"
AFTER_CONTENT="$(cat "$GUARD")"
[ "$DIRTY_CONTENT" = "$AFTER_CONTENT" ] \
  && pass "dirty touched file: uncommitted edit survives unchanged" \
  || fail "dirty touched file: uncommitted edit was altered"
case "$OUT" in
  *"uncommitted changes"*) pass "dirty touched file: names the reason" ;;
  *) fail "dirty touched file: reason missing: $OUT" ;;
esac

# 7. missing harness refused — a harness path that does not exist.
new_fixture_repo 4
make_patch "$REPO" 4 "flip:2"
NO_SUCH_HARNESS="$REPO/no-such-harness.sh"
run_mutate "$REPO" "$PATCH" "$NO_SUCH_HARNESS"
[ "$RC" -eq 2 ] && pass "missing harness: exit 2" \
  || fail "missing harness: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  *"no-such-harness.sh"*) pass "missing harness: names the harness" ;;
  *) fail "missing harness: harness not named: $OUT" ;;
esac
repo_clean "$REPO" && pass "missing harness: repo stays clean" \
  || fail "missing harness: repo left dirty: $(git -C "$REPO" status --porcelain)"

# 8. non-executable harness refused — a harness path that exists but is not
#    +x.
new_fixture_repo 4
make_patch "$REPO" 4 "flip:2"
NOEXEC_HARNESS="$REPO/test-fixture-noexec.sh"
cp "$HARNESS" "$NOEXEC_HARNESS"
chmod -x "$NOEXEC_HARNESS"
run_mutate "$REPO" "$PATCH" "$NOEXEC_HARNESS"
[ "$RC" -eq 2 ] && pass "non-executable harness: exit 2" \
  || fail "non-executable harness: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  *"test-fixture-noexec.sh"*) pass "non-executable harness: names the harness" ;;
  *) fail "non-executable harness: harness not named: $OUT" ;;
esac
# guard.sh itself (the only patch-touched file) stays clean; the untracked
# noexec harness copy this case added on purpose is not part of that check.
[ -z "$(git -C "$REPO" status --porcelain -- guard.sh)" ] \
  && pass "non-executable harness: touched file stays clean" \
  || fail "non-executable harness: touched file left dirty: $(git -C "$REPO" status --porcelain -- guard.sh)"

# 9. boundary — new-failure count exactly equals the default bound (5): 10
#    fixture-harness cases, broken unconditionally before the case
#    statement, fails exactly the 5 even-numbered cases. Pins the `-le`
#    comparison at its own boundary — an off-by-one (`-le` turned `-lt`)
#    would flip this verdict from caught to suspicious blast radius.
new_fixture_repo 10
make_patch "$REPO" 10 "always-fail"
run_mutate "$REPO" "$PATCH" "$HARNESS"
[ "$RC" -eq 0 ] && pass "boundary: exit 0" \
  || fail "boundary: expected exit 0, got rc=$RC out=$OUT"
case "$OUT" in
  *"caught"*) pass "boundary: verdict names caught at new_count == bound" ;;
  *) fail "boundary: verdict missing 'caught': $OUT" ;;
esac
case "$OUT" in
  *"suspicious"*) fail "boundary: wrongly flagged suspicious blast radius at new_count == bound: $OUT" ;;
  *) pass "boundary: does not flag suspicious blast radius" ;;
esac
case "$OUT" in
  *"5 new failure(s)"*) pass "boundary: reports exactly 5 new failures" ;;
  *) fail "boundary: does not report 5 new failures: $OUT" ;;
esac
repo_clean "$REPO" && pass "boundary: touched file clean after restore" \
  || fail "boundary: touched file left dirty: $(git -C "$REPO" status --porcelain)"

# 10. could not fully restore — exit 3 path: the patch's only touched file
#     is one `git apply` creates as a brand-new file. Such a file is
#     untracked afterward, so `git checkout --` cannot remove it (it has no
#     committed content to fall back to); the EXIT trap's residual
#     `git status --porcelain` check then finds it still present and forces
#     exit 3. This is the only way to make the restore step genuinely fail
#     without touching the real repository tree.
new_fixture_repo 4
NEWFILE="new-file.txt"
printf 'net new content\n' >"$REPO/$NEWFILE"
git -C "$REPO" add "$NEWFILE"
PATCH="$(mktemp "${TMPDIR:-/tmp}/mutate-verify-test-patch.XXXXXX")"
REPOS+=("$PATCH")
git -C "$REPO" diff --cached -- "$NEWFILE" >"$PATCH"
git -C "$REPO" reset -q -- "$NEWFILE"
rm -f "$REPO/$NEWFILE"
run_mutate "$REPO" "$PATCH" "$HARNESS"
[ "$RC" -eq 3 ] && pass "cannot restore: exit 3" \
  || fail "cannot restore: expected exit 3, got rc=$RC out=$OUT"
case "$OUT" in
  *"could not fully restore"*) pass "cannot restore: names the reason" ;;
  *) fail "cannot restore: reason missing: $OUT" ;;
esac
case "$OUT" in
  *"ran clean"*) fail "cannot restore: wrongly also claimed ran clean: $OUT" ;;
  *) pass "cannot restore: does not claim ran clean" ;;
esac
[ -e "$REPO/$NEWFILE" ] && pass "cannot restore: leftover untracked file still present" \
  || fail "cannot restore: leftover file unexpectedly absent"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'mutate-and-verify: all cases passed\n'
