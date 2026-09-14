#!/usr/bin/env bash
# Assertion harness for break-and-prove.sh (KAN-329). Builds throwaway git
# repositories under a sandboxed TMPDIR, each carrying one committed
# config.yaml and README.md, and asserts break-and-prove.sh's exit codes,
# report blocks and restore behavior against them. Never touches the real
# repository tree. Same shape as test-mutate-and-verify.sh: an indexed REPOS
# array (bash 3.2 has no associative arrays), removed by an EXIT trap, real
# git repositories rather than fixture trees.
#
# The proof loop the harness pins down, per KAN-329's own statement of it:
# apply the mutation, force a clean re-run, assert non-zero, restore from a
# pre-mutation snapshot (NEVER `git checkout --`, which was the observed
# failure that destroyed an agent's uncommitted edits), run again, assert
# zero, and print both runs' observed output.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BREAK="$SCRIPT_DIR/break-and-prove.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

assert_rc() { # assert_rc <expected> <what>
  if [ "$PROOF_RC" -eq "$1" ]; then
    pass "$2"
  else
    fail "$2 — expected exit $1, got $PROOF_RC"
  fi
}

assert_contains() { # assert_contains <needle> <what>
  if grep -qF -- "$1" "$OUT_FILE"; then
    pass "$2"
  else
    fail "$2 — output has no line matching: $1"
  fi
}

REPOS=()
cleanup() {
  [ "${#REPOS[@]}" -eq 0 ] && return 0
  for p in "${REPOS[@]}"; do
    rm -rf "$p" "${p}.out" 2>/dev/null || true
  done
}
trap cleanup EXIT

OUT_FILE=""

# new_fixture_repo — commits config.yaml (2 lines) and README.md (1 line).
new_fixture_repo() {
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/break-and-prove-test.XXXXXX")"
  REPOS+=("$REPO")
  OUT_FILE="${REPO}.out"
  git -C "$REPO" init -q
  git -C "$REPO" config user.email test@example.com
  git -C "$REPO" config user.name test
  printf 'mode: strict\nretries: 3\n' > "$REPO/config.yaml"
  printf 'readme v1\n' > "$REPO/README.md"
  git -C "$REPO" add config.yaml README.md
  git -C "$REPO" commit -qm fixture
}

# run_proof <args...> — runs break-and-prove.sh from inside the fixture repo,
# capturing combined stdout+stderr outside the repo tree (a capture inside it
# would itself be untracked drift).
run_proof() {
  set +e
  ( cd "$REPO" && "$BREAK" "$@" ) > "$OUT_FILE" 2>&1
  PROOF_RC=$?
  set -e
}

# config_matches_head — true when config.yaml in the working tree is
# byte-identical to the committed version.
config_matches_head() {
  git -C "$REPO" show HEAD:config.yaml | cmp -s - "$REPO/config.yaml"
}

# A hand-written unified diff flipping mode: strict to mode: lax.
write_flip_patch() { # write_flip_patch <patch-path>
  cat > "$1" <<'EOF'
--- a/config.yaml
+++ b/config.yaml
@@ -1,2 +1,2 @@
-mode: strict
+mode: lax
 retries: 3
EOF
}

# 1. proof held — sed mutation.
new_fixture_repo
cp "$REPO/config.yaml" "$REPO/.config-before"
run_proof --sed 's/mode: strict/mode: lax/' config.yaml -- grep -q 'mode: strict' config.yaml
assert_rc 0 "proof held (sed mutation) — exit code"
assert_contains "run 1 exited" "proof held (sed mutation) — run 1 reported"
if config_matches_head; then
  pass "proof held (sed mutation) — config restored byte-exact"
else
  fail "proof held (sed mutation) — config not restored byte-exact"
fi
rm -f "$REPO/.config-before"

# 2. proof held — patch mutation.
new_fixture_repo
write_flip_patch "$REPO.parent.patch"
run_proof --patch "$REPO.parent.patch" config.yaml -- grep -q 'mode: strict' config.yaml
assert_rc 0 "proof held (patch mutation) — exit code"
if config_matches_head; then
  pass "proof held (patch mutation) — config restored byte-exact"
else
  fail "proof held (patch mutation) — config not restored byte-exact"
fi
rm -f "$REPO.parent.patch"

# 3. uncommitted edit on the target file survives — the observed KAN-329
# failure: `git checkout --` destroyed such edits; the snapshot must not.
new_fixture_repo
printf '# tuned by hand\n' >> "$REPO/config.yaml"
run_proof --sed 's/mode: strict/mode: lax/' config.yaml -- grep -q 'mode: strict' config.yaml
assert_rc 0 "uncommitted edit on the target file survives — exit code"
if grep -q '^# tuned by hand$' "$REPO/config.yaml" && grep -q '^mode: strict$' "$REPO/config.yaml"; then
  pass "uncommitted edit on the target file survives — edit and original content both present"
else
  fail "uncommitted edit on the target file survives — edit lost or content wrong"
fi

# 4. uncommitted edit on a second file survives.
new_fixture_repo
printf 'uncommitted note\n' >> "$REPO/README.md"
run_proof --sed 's/mode: strict/mode: lax/' config.yaml -- grep -q 'mode: strict' config.yaml
assert_rc 0 "uncommitted edit on a second file survives — exit code"
if grep -q '^uncommitted note$' "$REPO/README.md"; then
  pass "uncommitted edit on a second file survives — edit still present"
else
  fail "uncommitted edit on a second file survives — edit lost"
fi

# 5. no `git checkout` call in the script — the anti-pattern is banned
# structurally, not just in the happy path. Comment lines are stripped
# first: the header documents the ban and may name the command in prose;
# the assertion is about executable code.
if grep -v '^[[:space:]]*#' "$BREAK" | grep -q 'git checkout'; then
  fail "no git checkout in the script — found a git checkout call"
else
  pass "no git checkout in the script"
fi

# 6. mutated run passed — the proof did not hold (surviving mutant).
new_fixture_repo
run_proof --sed 's/mode: strict/mode: lax/' config.yaml -- grep -q 'mode:' config.yaml
assert_rc 1 "mutated run passed — proof did not hold"
assert_contains "run 1 exited 0" "mutated run passed — leg named"
if config_matches_head; then
  pass "mutated run passed — config still restored"
else
  fail "mutated run passed — config not restored"
fi

# 7. restored run failed — the proof did not hold.
new_fixture_repo
run_proof --sed 's/mode: strict/mode: lax/' config.yaml -- grep -q 'impossible-token' config.yaml
assert_rc 1 "restored run failed — proof did not hold"
assert_contains "run 2" "restored run failed — leg named"

# 8. patch not applying — refused before mutating anything.
new_fixture_repo
cat > "$REPO.parent.patch" <<'EOF'
--- a/config.yaml
+++ b/config.yaml
@@ -1,2 +1,2 @@
-mode: strict
+mode: lax
 retries: 9
EOF
run_proof --patch "$REPO.parent.patch" config.yaml -- grep -q 'mode: strict' config.yaml
assert_rc 2 "patch not applying refused"
if config_matches_head; then
  pass "patch not applying refused — config untouched"
else
  fail "patch not applying refused — config modified"
fi
rm -f "$REPO.parent.patch"

# 9. sed matching nothing — refused.
new_fixture_repo
run_proof --sed 's/zzz-no-match/yyy/' config.yaml -- grep -q 'mode: strict' config.yaml
assert_rc 2 "sed matching nothing refused"
if config_matches_head; then
  pass "sed matching nothing refused — config untouched"
else
  fail "sed matching nothing refused — config modified"
fi

# 10. patch touching other files — refused.
new_fixture_repo
cat > "$REPO.parent.patch" <<'EOF'
--- a/config.yaml
+++ b/config.yaml
@@ -1,2 +1,2 @@
-mode: strict
+mode: lax
 retries: 3
--- a/README.md
+++ b/README.md
@@ -1 +1 @@
-readme v1
+readme v2
EOF
run_proof --patch "$REPO.parent.patch" config.yaml -- grep -q 'mode: strict' config.yaml
assert_rc 2 "patch touching other files refused"
rm -f "$REPO.parent.patch"

# 11. missing target file — refused.
new_fixture_repo
run_proof --sed 's/a/b/' no-such-file.yaml -- grep -q x no-such-file.yaml
assert_rc 2 "missing target file refused"

# 12. the clean command runs before each leg — the Gradle cleanTest trap,
# handled once by the script instead of remembered every time.
new_fixture_repo
LOG="$REPO.parent.log"
run_proof --clean "echo clean >> $LOG" --sed 's/mode: strict/mode: lax/' config.yaml -- \
  sh -c "echo test >> $LOG; grep -q 'mode: strict' config.yaml"
assert_rc 0 "clean command runs before each leg — exit code"
if [ "$(cat "$LOG" 2>/dev/null)" = "clean
test
clean
test" ]; then
  pass "clean command runs before each leg — clean,test,clean,test"
else
  fail "clean command runs before each leg — log was: $(cat "$LOG" 2>/dev/null | tr '\n' ',')"
fi
rm -f "$LOG"

# 13. restore failure — exit 3, mutated content may remain.
new_fixture_repo
run_proof --sed 's/mode: strict/mode: lax/' config.yaml -- \
  sh -c 'chmod 444 config.yaml; exit 7'
assert_rc 3 "restore failure exits 3"
assert_contains "could not restore" "restore failure exits 3 — named in the report"
chmod 644 "$REPO/config.yaml" 2>/dev/null || true

# 14. tree drift — an untracked file the test command leaves behind is
# reported and forces exit 2.
new_fixture_repo
run_proof --sed 's/mode: strict/mode: lax/' config.yaml -- \
  sh -c 'touch stray.txt; exit 5'
assert_rc 2 "tree drift exits 2"
assert_contains "stray.txt" "tree drift exits 2 — stray file named"

# 15. usage errors — exit 4.
new_fixture_repo
run_proof config.yaml -- grep -q 'mode: strict' config.yaml
assert_rc 4 "usage errors exit 4 — no mutation flag"
run_proof --sed 's/a/b/' config.yaml
assert_rc 4 "usage errors exit 4 — no test command"
run_proof --patch "$REPO/no-such.patch" config.yaml -- grep -q x config.yaml
assert_rc 4 "usage errors exit 4 — unreadable patch file"

# 16. not inside a git worktree — exit 4.
BARE="$(mktemp -d "${TMPDIR:-/tmp}/break-and-prove-test.XXXXXX")"
REPOS+=("$BARE")
OUT_FILE="${BARE}.out"
set +e
( cd "$BARE" && "$BREAK" --sed 's/a/b/' whatever.yaml -- true ) > "$OUT_FILE" 2>&1
PROOF_RC=$?
set -e
assert_rc 4 "not a git worktree exits 4"

# 17. test command that cannot exec (127) — exit 4, never a false verdict.
new_fixture_repo
cp "$REPO/config.yaml" "$REPO/.config-before"
run_proof --sed 's/mode: strict/mode: lax/' config.yaml -- ./no-such-command-anywhere
assert_rc 4 "test command 127 exits 4"
assert_contains "not usable as evidence" "test command 127 — the 126/127 message, not a proof verdict"
config_matches_head && pass "test command 127 — config restored" || fail "test command 127 — config not restored"
rm -f "$REPO/.config-before"

# 18. test command that exists but is not executable (126) — exit 4 too.
new_fixture_repo
printf '#!/usr/bin/env bash\necho should-not-run\n' > "$REPO/nonexec-cmd"
run_proof --sed 's/mode: strict/mode: lax/' config.yaml -- ./nonexec-cmd
assert_rc 4 "test command 126 exits 4"
assert_contains "not usable as evidence" "test command 126 — the 126/127 message, not a proof verdict"
config_matches_head && pass "test command 126 — config restored" || fail "test command 126 — config not restored"

if [ "$FAILURES" -gt 0 ]; then
  printf 'FAIL: test-break-and-prove — %d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'ok: test-break-and-prove — all cases passed\n'
