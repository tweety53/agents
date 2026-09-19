#!/usr/bin/env bash
# test-run-guard-tests.sh — assertion harness for scripts/run-guard-tests.sh,
# the KAN-362 task 2 runner that discovers scripts/test-*.sh and runs them
# through scripts/lib/parallel.sh (task 1). Every case below points the
# runner at a throwaway fixture directory via RUN_GUARD_TESTS_ROOT — see
# design.md's runner-root-override decision — and never invokes it against
# this repository's own scripts/ directory, which is exactly what
# RUN_GUARD_TESTS_ROOT exists to make possible: scripts/test-run-guard-tests.sh
# is itself matched by the runner's own scripts/test-*.sh glob, so without
# the override this harness would re-enter the real 40+-harness suite.
#
# The four bullets design.md's Testing section names for this file:
# discovery finds every test-*.sh in the fixture and nothing else; one
# ok/FAIL line per harness plus the summary counts; non-zero exit on any
# failure; RUN_GUARD_TESTS_ROOT set but empty exits 2. A "shape" section
# below those four exercises the output-shape decision (replay-failures-only):
# a clean run stays quiet, a failing run replays the failing harness's
# captured output in full and names every failure.
#
# Bash 3.2 is the floor, as test-check-finish-preflight.sh's header records:
# indexed arrays only, no associative arrays, no `wait -n`.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUNNER="$SCRIPT_DIR/run-guard-tests.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

if [ ! -x "$RUNNER" ] && [ ! -r "$RUNNER" ]; then
  echo "test-run-guard-tests: cannot read $RUNNER" >&2
  exit 2
fi

# run_runner <fixture_dir> -> sets RC and OUT. Always via
# RUN_GUARD_TESTS_ROOT — never a bare invocation, which would glob this
# repository's real scripts/ directory.
run_runner() {
  set +e
  OUT="$(RUN_GUARD_TESTS_ROOT="$1" bash "$RUNNER" 2>&1)"
  RC=$?
  set -e
}

# new_fixture -> a fresh mktemp -d with a canonical mixed set of harnesses:
#   test-alpha.sh   passes,  prints ALPHA-OUT-MARKER on stdout
#   test-beta.sh    FAILS,   prints BETA-OUT-MARKER (stdout) and
#                            BETA-ERR-MARKER (stderr), exits 3
#   test-gamma.sh   passes,  prints GAMMA-OUT-MARKER on stdout
#   helper.sh              a decoy that does NOT match test-*.sh
#   subdir/test-nested.sh  a decoy in a subdirectory — proves the glob is
#                          flat, not recursive
new_fixture() {
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/run-guard-tests-fixture.XXXXXX")"
  cat > "$FIXTURE/test-alpha.sh" <<'EOF'
#!/usr/bin/env bash
printf 'ALPHA-OUT-MARKER\n'
exit 0
EOF
  cat > "$FIXTURE/test-beta.sh" <<'EOF'
#!/usr/bin/env bash
printf 'BETA-OUT-MARKER\n'
printf 'BETA-ERR-MARKER\n' >&2
exit 3
EOF
  cat > "$FIXTURE/test-gamma.sh" <<'EOF'
#!/usr/bin/env bash
printf 'GAMMA-OUT-MARKER\n'
exit 0
EOF
  cat > "$FIXTURE/helper.sh" <<'EOF'
#!/usr/bin/env bash
printf 'HELPER-SHOULD-NEVER-RUN-MARKER\n'
exit 1
EOF
  mkdir -p "$FIXTURE/subdir"
  cat > "$FIXTURE/subdir/test-nested.sh" <<'EOF'
#!/usr/bin/env bash
printf 'NESTED-SHOULD-NEVER-RUN-MARKER\n'
exit 1
EOF
  chmod +x "$FIXTURE"/test-*.sh "$FIXTURE/helper.sh" "$FIXTURE/subdir/test-nested.sh"
}

# new_clean_fixture -> a fixture where every harness passes, for the
# quiet-on-pass bullet.
new_clean_fixture() {
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/run-guard-tests-clean.XXXXXX")"
  cat > "$FIXTURE/test-alpha.sh" <<'EOF'
#!/usr/bin/env bash
printf 'ALPHA-OUT-MARKER\n'
exit 0
EOF
  cat > "$FIXTURE/test-gamma.sh" <<'EOF'
#!/usr/bin/env bash
printf 'GAMMA-OUT-MARKER\n'
exit 0
EOF
  chmod +x "$FIXTURE"/test-*.sh
}

# ---------------------------------------------------------------------------
# 1. Discovery finds every test-*.sh in the fixture and nothing else.
# ---------------------------------------------------------------------------
new_fixture
run_runner "$FIXTURE"
case "$OUT" in
  *test-alpha.sh*) pass "case 1: discovers test-alpha.sh" ;;
  *) fail "case 1: did not discover test-alpha.sh — out=$OUT" ;;
esac
case "$OUT" in
  *test-beta.sh*) pass "case 1: discovers test-beta.sh" ;;
  *) fail "case 1: did not discover test-beta.sh — out=$OUT" ;;
esac
case "$OUT" in
  *test-gamma.sh*) pass "case 1: discovers test-gamma.sh" ;;
  *) fail "case 1: did not discover test-gamma.sh — out=$OUT" ;;
esac
case "$OUT" in
  *HELPER-SHOULD-NEVER-RUN-MARKER*) fail "case 1: ran helper.sh, which does not match test-*.sh — out=$OUT" ;;
  *) pass "case 1: does not run helper.sh (glob excludes non test-*.sh names)" ;;
esac
case "$OUT" in
  *NESTED-SHOULD-NEVER-RUN-MARKER*) fail "case 1: ran subdir/test-nested.sh — the glob must be flat, not recursive — out=$OUT" ;;
  *) pass "case 1: does not descend into subdir/ (glob is flat, not recursive)" ;;
esac
rm -rf "$FIXTURE"

# ---------------------------------------------------------------------------
# 2. One ok/FAIL line per harness, plus the summary counts.
# ---------------------------------------------------------------------------
new_fixture
run_runner "$FIXTURE"
if printf '%s\n' "$OUT" | grep -Eq '^ok:.*test-alpha\.sh'; then
  pass "case 2: an ok: line names the passing test-alpha.sh"
else
  fail "case 2: no ok: line for test-alpha.sh — out=$OUT"
fi
if printf '%s\n' "$OUT" | grep -Eq '^ok:.*test-gamma\.sh'; then
  pass "case 2: an ok: line names the passing test-gamma.sh"
else
  fail "case 2: no ok: line for test-gamma.sh — out=$OUT"
fi
if printf '%s\n' "$OUT" | grep -Eq '^FAIL:.*test-beta\.sh'; then
  pass "case 2: a FAIL: line names the failing test-beta.sh"
else
  fail "case 2: no FAIL: line for test-beta.sh — out=$OUT"
fi
if printf '%s\n' "$OUT" | grep -Eq '^ok:.*test-beta\.sh'; then
  fail "case 2: test-beta.sh (which fails) was reported with ok:, not FAIL: — out=$OUT"
else
  pass "case 2: the failing harness is never reported as ok:"
fi
if printf '%s\n' "$OUT" | grep -Eq '3 harnesses'; then
  pass "case 2: summary names the total (3 harnesses)"
else
  fail "case 2: summary does not name the total 3 harnesses — out=$OUT"
fi
if printf '%s\n' "$OUT" | grep -Eq '2 passed'; then
  pass "case 2: summary names the passed count (2)"
else
  fail "case 2: summary does not name 2 passed — out=$OUT"
fi
if printf '%s\n' "$OUT" | grep -Eq '1 failed'; then
  pass "case 2: summary names the failed count (1)"
else
  fail "case 2: summary does not name 1 failed — out=$OUT"
fi
rm -rf "$FIXTURE"

# MUTATION PROOF for case 2: an all-passing fixture must report 0 failed and
# every harness's line as ok: — proves the ok:/FAIL: and count assertions
# above actually discriminate the two outcomes rather than always matching.
new_clean_fixture
run_runner "$FIXTURE"
if printf '%s\n' "$OUT" | grep -Eq '^FAIL:'; then
  fail "case 2 mutation: an all-passing fixture reported a FAIL: line — out=$OUT"
else
  pass "case 2 mutation: an all-passing fixture reports no FAIL: line"
fi
if printf '%s\n' "$OUT" | grep -Eq '0 failed'; then
  pass "case 2 mutation: an all-passing fixture's summary names 0 failed"
else
  fail "case 2 mutation: an all-passing fixture's summary does not name 0 failed — out=$OUT"
fi
rm -rf "$FIXTURE"

# ---------------------------------------------------------------------------
# 3. Non-zero exit on any failure; zero exit when every harness passes.
# ---------------------------------------------------------------------------
new_fixture
run_runner "$FIXTURE"
if [ "$RC" -ne 0 ]; then
  pass "case 3: a fixture with a failing harness exits non-zero"
else
  fail "case 3: expected a non-zero exit, got 0"
fi
rm -rf "$FIXTURE"

new_clean_fixture
run_runner "$FIXTURE"
if [ "$RC" -eq 0 ]; then
  pass "case 3: an all-passing fixture exits 0"
else
  fail "case 3: expected exit 0, got $RC — out=$OUT"
fi
rm -rf "$FIXTURE"

# ---------------------------------------------------------------------------
# 4. RUN_GUARD_TESTS_ROOT set but empty exits 2 — mirrors
#    CHECK_REFERENCES_ROOT / CHECK_GUARD_SYMLINKS_ROOT verbatim.
# ---------------------------------------------------------------------------
set +e
EMPTY_OUT="$(RUN_GUARD_TESTS_ROOT="" bash "$RUNNER" 2>&1)"
EMPTY_RC=$?
set -e
if [ "$EMPTY_RC" -eq 2 ]; then
  pass "case 4: RUN_GUARD_TESTS_ROOT set but empty exits 2"
else
  fail "case 4: RUN_GUARD_TESTS_ROOT set but empty exited $EMPTY_RC, want 2 — out=$EMPTY_OUT"
fi

# MUTATION PROOF for case 4: RUN_GUARD_TESTS_ROOT genuinely UNSET must not
# hit that same exit-2 path (it falls back to the runner's own directory) —
# proves the check is "set but empty", not "empty or unset" collapsed
# together. Point the fallback at this repository's own scripts/ directory
# indirectly is unsafe (forbidden by this task), so instead prove the
# distinction at the parsing level: a fixture path that legitimately exists
# and is merely unrelated to RUN_GUARD_TESTS_ROOT being unset must not exit 2.
new_clean_fixture
set +e
unset RUN_GUARD_TESTS_ROOT 2>/dev/null || true
UNSET_OUT="$(cd "$FIXTURE" && RUN_GUARD_TESTS_ROOT="$FIXTURE" bash "$RUNNER" 2>&1)"
UNSET_RC=$?
set -e
if [ "$UNSET_RC" -eq 2 ] && printf '%s\n' "$UNSET_OUT" | grep -q 'set but empty'; then
  fail "case 4 mutation: a non-empty RUN_GUARD_TESTS_ROOT was rejected as if empty — out=$UNSET_OUT"
else
  pass "case 4 mutation: a non-empty RUN_GUARD_TESTS_ROOT is honoured, not rejected as empty"
fi
rm -rf "$FIXTURE"

# ---------------------------------------------------------------------------
# 5. Output shape (replay-failures-only): a clean run stays quiet — it must
#    not dump the harnesses' own stdout/stderr; a failing run replays the
#    failing harness's captured output in full, and does not replay a
#    passing neighbour's output.
# ---------------------------------------------------------------------------
new_clean_fixture
run_runner "$FIXTURE"
case "$OUT" in
  *ALPHA-OUT-MARKER*) fail "case 5: a clean run dumped test-alpha.sh's own output — out=$OUT" ;;
  *) pass "case 5: a clean run does not dump a passing harness's output" ;;
esac
case "$OUT" in
  *GAMMA-OUT-MARKER*) fail "case 5: a clean run dumped test-gamma.sh's own output — out=$OUT" ;;
  *) pass "case 5: a clean run does not dump another passing harness's output" ;;
esac
rm -rf "$FIXTURE"

new_fixture
run_runner "$FIXTURE"
case "$OUT" in
  *BETA-OUT-MARKER*) pass "case 5: a failing run replays the failing harness's stdout" ;;
  *) fail "case 5: failing run did not replay test-beta.sh's stdout — out=$OUT" ;;
esac
case "$OUT" in
  *BETA-ERR-MARKER*) pass "case 5: a failing run replays the failing harness's stderr" ;;
  *) fail "case 5: failing run did not replay test-beta.sh's stderr — out=$OUT" ;;
esac
case "$OUT" in
  *ALPHA-OUT-MARKER*) fail "case 5: a failing run also replayed a passing neighbour's output — out=$OUT" ;;
  *) pass "case 5: a failing run does not replay a passing neighbour's output" ;;
esac
case "$OUT" in
  *test-beta.sh*) pass "case 5: the failing harness is named in the output" ;;
  *) fail "case 5: the failing harness's name never appears in the output — out=$OUT" ;;
esac
rm -rf "$FIXTURE"

# ---------------------------------------------------------------------------
# 6. Companion presence (kan-387): every check-*.sh guard in the fixture
#    root must carry a test-check-*.sh companion, or the runner refuses
#    the suite with exit 1, naming the guard and the missing harness, and
#    runs nothing — the gap is a suite-contract violation, not a harness
#    result. A balanced fixture (guard plus companion) passes through to
#    the ordinary run.
# ---------------------------------------------------------------------------
new_clean_fixture
cat > "$FIXTURE/check-lonely.sh" <<'EOF'
#!/usr/bin/env bash
printf 'LONELY-SHOULD-NEVER-RUN-MARKER\n'
EOF
chmod +x "$FIXTURE/check-lonely.sh"
run_runner "$FIXTURE"
if [ "$RC" -eq 1 ]; then
  pass "case 6a: a fixture with an uncompanioned guard exits 1"
else
  fail "case 6a: expected exit 1, got $RC — out=$OUT"
fi
case "$OUT" in
  *check-lonely.sh*) pass "case 6a: the guard is named" ;;
  *) fail "case 6a: the uncompanioned guard is not named — out=$OUT" ;;
esac
case "$OUT" in
  *test-check-lonely.sh*) pass "case 6a: the missing companion is named" ;;
  *) fail "case 6a: the missing companion is not named — out=$OUT" ;;
esac
case "$OUT" in
  *LONELY-SHOULD-NEVER-RUN-MARKER*|*ALPHA-OUT-MARKER*)
    fail "case 6a: the runner ran harnesses despite the companion gap — out=$OUT" ;;
  *) pass "case 6a: nothing ran — the refusal fires before any harness" ;;
esac

new_clean_fixture
cat > "$FIXTURE/check-paired.sh" <<'EOF'
#!/usr/bin/env bash
printf 'PAIRED-GUARD\n'
EOF
cat > "$FIXTURE/test-check-paired.sh" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
chmod +x "$FIXTURE/check-paired.sh" "$FIXTURE/test-check-paired.sh"
run_runner "$FIXTURE"
if [ "$RC" -eq 0 ]; then
  pass "case 6b: a balanced fixture (guard plus companion) exits 0"
else
  fail "case 6b: expected exit 0, got $RC — out=$OUT"
fi
case "$OUT" in
  *companion*) fail "case 6b: a balanced fixture reported a companion gap — out=$OUT" ;;
  *) pass "case 6b: a balanced fixture never reports a companion gap" ;;
esac

# ---------------------------------------------------------------------------
# 7. Tree-integrity gate (KAN-584): the runner snapshots the watched
#    repository's `git status --porcelain` before launching the suite and
#    re-reads it after; a net change fails the runner naming the changed
#    paths — a harness case mutating a real shared file is a suite-contract
#    violation, the same code a failing harness gets. GUARD_TESTS_REPO_ROOT
#    overrides the watched root (the runner's own repository when unset),
#    the same opt-in-override idiom as RUN_GUARD_TESTS_ROOT and for the
#    same reason: this harness must never point the runner at this
#    repository's own tree. A watched root that is not a git work tree
#    skips the gate with one stderr line rather than failing it.
# ---------------------------------------------------------------------------
# 7a. A harness that writes into the watched repository fails the runner,
#     which names the changed path; the writing harness itself still gets
#     its own ok: line — the gate is a separate verdict from the results.
new_dirty_fixtures() {
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/run-guard-tests-writer.XXXXXX")"
  WATCHED="$(mktemp -d "${TMPDIR:-/tmp}/run-guard-tests-watched.XXXXXX")"
  git -C "$WATCHED" init -q
  git -C "$WATCHED" config user.email test@example.com
  git -C "$WATCHED" config user.name test
  printf 'base\n' > "$WATCHED/base.txt"
  git -C "$WATCHED" add base.txt
  git -C "$WATCHED" commit -qm base
  cat > "$FIXTURE/test-dirty.sh" <<'EOF'
#!/usr/bin/env bash
printf 'mutated by a harness case\n' > "$GUARD_TESTS_REPO_ROOT/dirt.txt"
exit 0
EOF
  chmod +x "$FIXTURE/test-dirty.sh"
}
new_dirty_fixtures
set +e
OUT="$(RUN_GUARD_TESTS_ROOT="$FIXTURE" GUARD_TESTS_REPO_ROOT="$WATCHED" bash "$RUNNER" 2>&1)"
RC=$?
set -e
if [ "$RC" -eq 1 ]; then
  pass "case 7a: a suite that dirties the watched repository exits 1"
else
  fail "case 7a: expected exit 1, got $RC — out=$OUT"
fi
case "$OUT" in
  *dirt.txt*) pass "case 7a: the changed path is named" ;;
  *) fail "case 7a: the changed path dirt.txt is not named — out=$OUT" ;;
esac
case "$OUT" in
  *tree\ changed\ during\ the\ suite*) pass "case 7a: the gate names the violation" ;;
  *) fail "case 7a: no tree-changed report — out=$OUT" ;;
esac
case "$OUT" in
  *ok:*test-dirty.sh*) pass "case 7a: the writing harness still gets its own ok: line" ;;
  *) fail "case 7a: the writing harness has no ok: line — out=$OUT" ;;
esac
rm -rf "$FIXTURE" "$WATCHED"

# 7b. A suite that touches only its own sandbox leaves the watched
#     repository clean and passes the gate.
new_clean_fixture
WATCHED="$(mktemp -d "${TMPDIR:-/tmp}/run-guard-tests-clean-watched.XXXXXX")"
git -C "$WATCHED" init -q
printf 'base\n' > "$WATCHED/base.txt"
set +e
OUT="$(RUN_GUARD_TESTS_ROOT="$FIXTURE" GUARD_TESTS_REPO_ROOT="$WATCHED" bash "$RUNNER" 2>&1)"
RC=$?
set -e
if [ "$RC" -eq 0 ]; then
  pass "case 7b: a sandbox-only suite passes the tree gate (exit 0)"
else
  fail "case 7b: expected exit 0, got $RC — out=$OUT"
fi
case "$OUT" in
  *tree\ changed\ during\ the\ suite*) fail "case 7b: a clean suite reported a tree change — out=$OUT" ;;
  *) pass "case 7b: a clean suite never reports a tree change" ;;
esac
rm -rf "$FIXTURE" "$WATCHED"

# 7c. A watched root that is not a git work tree skips the gate with one
#     stderr line — an unanswerable question, not a violation.
new_dirty_fixtures
rm -rf "$WATCHED"
WATCHED="$(mktemp -d "${TMPDIR:-/tmp}/run-guard-tests-nongit.XXXXXX")"
set +e
OUT="$(RUN_GUARD_TESTS_ROOT="$FIXTURE" GUARD_TESTS_REPO_ROOT="$WATCHED" bash "$RUNNER" 2>&1)"
RC=$?
set -e
if [ "$RC" -eq 0 ]; then
  pass "case 7c: a non-git watched root skips the gate and exits 0"
else
  fail "case 7c: expected exit 0, got $RC — out=$OUT"
fi
case "$OUT" in
  *skipping\ the\ tree\ gate*) pass "case 7c: the skip is announced, never silent" ;;
  *) fail "case 7c: no skip line for the non-git watched root — out=$OUT" ;;
esac
rm -rf "$FIXTURE" "$WATCHED"

# 7d. GUARD_TESTS_REPO_ROOT set but empty exits 2, mirroring
#     RUN_GUARD_TESTS_ROOT's own set-but-empty refusal.
new_clean_fixture
set +e
EMPTY_GATED_OUT="$(RUN_GUARD_TESTS_ROOT="$FIXTURE" GUARD_TESTS_REPO_ROOT="" bash "$RUNNER" 2>&1)"
EMPTY_GATED_RC=$?
set -e
if [ "$EMPTY_GATED_RC" -eq 2 ]; then
  pass "case 7d: GUARD_TESTS_REPO_ROOT set but empty exits 2"
else
  fail "case 7d: expected exit 2, got $EMPTY_GATED_RC — out=$EMPTY_GATED_OUT"
fi
rm -rf "$FIXTURE"

# 7e. A BARE watched repository skips the gate with the same announced
#     line — `rev-parse --is-inside-work-tree` exits 0 printing `false`
#     there, so validation must read the verdict, never the exit code
#     alone (kan-584's panel, F1/P1: treating exit 0 as a pass crashed the
#     runner's snapshot under set -e).
new_dirty_fixtures
rm -rf "$WATCHED"
git init -q --bare "$WATCHED"
set +e
OUT="$(RUN_GUARD_TESTS_ROOT="$FIXTURE" GUARD_TESTS_REPO_ROOT="$WATCHED" bash "$RUNNER" 2>&1)"
RC=$?
set -e
if [ "$RC" -eq 0 ]; then
  pass "case 7e: a bare watched repository skips the gate and exits 0"
else
  fail "case 7e: expected exit 0, got $RC — out=$OUT"
fi
case "$OUT" in
  *skipping\ the\ tree\ gate*) pass "case 7e: the bare-repo skip is announced" ;;
  *) fail "case 7e: no skip line for the bare watched repository — out=$OUT" ;;
esac
rm -rf "$FIXTURE" "$WATCHED"

# MUTATION PROOF for case 7: both 7a's fail and 7b's pass were hand-checked
# against a scratch copy of run-guard-tests.sh with the post-suite snapshot
# deleted (the gate reading "clean" unconditionally) — 7a then passed where
# it must fail and 7b was indistinguishable from it. Restoring the snapshot
# made 7a fail and 7b pass again. This is the same manual-mutation-and-revert
# method test-mutate-and-verify.sh's header records; it is not reproduced
# automatically on every run.

# ---------------------------------------------------------------------------
if [ "$FAILURES" -eq 0 ]; then
  printf '\n✓ PASS\n'
  exit 0
fi
printf '\n✗ FAIL — %s failure(s)\n' "$FAILURES" >&2
exit 1
