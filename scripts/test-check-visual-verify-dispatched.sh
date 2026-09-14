#!/usr/bin/env bash
# test-check-visual-verify-dispatched.sh — assertion harness for
# check-visual-verify-dispatched.sh, /flow's run-1 finish-gate guard that
# answers whether a change touching a declared `## visual verification` UI
# path actually dispatched flow.visual-verify's verifier (kan-30's fix).
#
# Every case builds a worktree-shaped sandbox: a real git repository
# carrying .flow/project.md (with or without the declared section), a base
# commit and a change commit, and a stub `flow` binary placed ahead of the
# real one on PATH that answers `record dispatches -change <name> -C <dir>`
# with a canned JSON array, or exits non-zero / prints non-JSON to simulate
# a store the guard cannot establish. The stub follows
# test-check-panel-fix-single-dispatch.sh's own helper shape — the guard
# names that harness's posture as the one its store read duplicates.
#
# The cases pin the guard's three verdicts and its exit contract: not
# configured and no-UI-paths are VISUAL-VERIFY-OK (exit 0) without
# consulting the store at all; UI paths touched with a completed verifier
# dispatch whose key starts with `visual-verify` is VISUAL-VERIFY-OK
# (exit 0); the same paths with no such row — empty store, wrong role,
# wrong key, non-completed outcome — is VISUAL-VERIFY-MISSING (exit 1);
# and every inability to answer (arguments, store unreachable, non-JSON)
# is exit 2, never read as a verdict.
#
# Bash 3.2 is the floor: indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-visual-verify-dispatched.sh"
FAILURES=0

ROOTS=()
cleanup() {
  [ "${#ROOTS[@]}" -eq 0 ] && return 0
  for r in "${ROOTS[@]}"; do
    rm -rf "$r"
  done
}
trap cleanup EXIT

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# dispatch_json <key> <role> <outcome> [<key> <role> <outcome> ...] --
# builds a compact JSON array of dispatch-row objects from triples, via
# `jq -n --args` rather than hand-quoted string interpolation.
dispatch_json() {
  jq -nc '
    [$ARGS.positional as $a
     | range(0; ($a | length) / 3)
     | {key: $a[. * 3], role: $a[. * 3 + 1], outcome: $a[. * 3 + 2]}]
  ' --args -- "$@"
}

# make_wt <dispatches-json|UNREACHABLE|NOT-JSON> <declare-section|plain> --
# a worktree-shaped sandbox: a git repo with .flow/project.md (declaring a
# `## visual verification` section unless "plain"), and a stub `flow` on
# its own bin/ answering `record dispatches` from the canned array. Sets
# WT and BASE (the root commit the guard diffs against). Prints nothing.
make_wt() {
  local store="$1" section="$2" rows="[]"
  WT="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/check-vv-dispatched-test.XXXXXX")" && pwd -P)"
  ROOTS+=("$WT")
  git -C "$WT" init -q
  git -C "$WT" config user.email guard-test@example.com
  git -C "$WT" config user.name guard-test
  mkdir -p "$WT/.flow" "$WT/bin"
  if [ "$section" = "declare-section" ]; then
    cat > "$WT/.flow/project.md" <<'EOF'
# project

## visual verification

| Setting | Value |
|---------|-------|
| `ui paths` | `app/src/**` |
EOF
  else
    printf '# project\n' > "$WT/.flow/project.md"
  fi
  git -C "$WT" add .flow/project.md
  git -C "$WT" commit -q -m base
  BASE="$(git -C "$WT" rev-parse HEAD)"
  case "$store" in
    UNREACHABLE)
      cat > "$WT/bin/flow" <<'STUB'
#!/usr/bin/env bash
echo "flow: connect: connection refused" >&2
exit 1
STUB
      ;;
    NOT-JSON)
      cat > "$WT/bin/flow" <<'STUB'
#!/usr/bin/env bash
if [ "${1:-}" = record ] && [ "${2:-}" = dispatches ]; then
  printf 'sorry, not json\n'
  exit 0
fi
exit 2
STUB
      ;;
    *)
      rows="$store"
      cat > "$WT/bin/flow" <<STUB
#!/usr/bin/env bash
if [ "\${1:-}" = record ] && [ "\${2:-}" = dispatches ]; then
  cat "$WT/bin/dispatches.json"
  exit 0
fi
echo "stub flow: unexpected invocation: \$*" >&2
exit 2
STUB
      printf '%s' "$rows" > "$WT/bin/dispatches.json"
      ;;
  esac
  chmod +x "$WT/bin/flow"
}

# touch_paths <path>... -- writes each path and commits it, so the diff
# BASE..HEAD names exactly those paths.
touch_paths() {
  for p in "$@"; do
    mkdir -p "$WT/$(dirname "$p")"
    printf 'x\n' > "$WT/$p"
    git -C "$WT" add "$p"
  done
  git -C "$WT" commit -q -m change
}

run_guard() {
  set +e
  OUT="$(PATH="$WT/bin:$PATH" "$GUARD" "$WT" demo "$BASE" 2>&1)"
  RC=$?
  set -e
}

# ---------------------------------------------------------------------------
# 1. No declared section — VISUAL-VERIFY-OK: not configured, exit 0. The
#    verdict is reached on the trigger's exit-2 path, before the store
#    read in the guard's own code order; the rc pins that behaviour.
# ---------------------------------------------------------------------------
make_wt "[]" plain
touch_paths "docs/note.md"
run_guard
[ "$RC" -eq 0 ] && pass "case 1: no declared section exits 0" \
  || fail "case 1: rc=$RC out=$OUT"
case "$OUT" in
  *"VISUAL-VERIFY-OK: not configured"*) pass "case 1: the verdict is not-configured" ;;
  *) fail "case 1: no not-configured verdict: out=$OUT" ;;
esac

# ---------------------------------------------------------------------------
# 2. Section declared, the diff touches no ui path — VISUAL-VERIFY-OK: no
#    UI paths touched, exit 0, reached on the trigger's exit-1 path before
#    the store read in the guard's own code order.
# ---------------------------------------------------------------------------
make_wt "[]" declare-section
touch_paths "docs/note.md"
run_guard
[ "$RC" -eq 0 ] && pass "case 2: no ui path touched exits 0" \
  || fail "case 2: rc=$RC out=$OUT"
case "$OUT" in
  *"VISUAL-VERIFY-OK: no UI paths touched"*) pass "case 2: the verdict is no-UI-paths" ;;
  *) fail "case 2: no no-UI-paths verdict: out=$OUT" ;;
esac

# ---------------------------------------------------------------------------
# 3. A ui path touched and one completed verifier dispatch whose key starts
#    with `visual-verify` — VISUAL-VERIFY-OK, exit 0. The key carries a
#    worktree suffix, one of the shapes the guard's header accepts.
# ---------------------------------------------------------------------------
make_wt "$(dispatch_json visual-verify-wt2 verifier completed)" declare-section
touch_paths "app/src/Widget.tsx"
run_guard
[ "$RC" -eq 0 ] && pass "case 3: completed verifier dispatch exits 0" \
  || fail "case 3: rc=$RC out=$OUT"
case "$OUT" in
  VISUAL-VERIFY-OK:*) pass "case 3: the verdict is VISUAL-VERIFY-OK" ;;
  *) fail "case 3: no OK verdict: out=$OUT" ;;
esac

# ---------------------------------------------------------------------------
# 4. A ui path touched, a verifier row closed aborted — a non-completed
#    outcome is not evidence the verifier's report was read: exit 1.
# ---------------------------------------------------------------------------
make_wt "$(dispatch_json visual-verify verifier aborted)" declare-section
touch_paths "app/src/Widget.tsx"
run_guard
[ "$RC" -eq 1 ] && pass "case 4: an aborted dispatch is not evidence, exit 1" \
  || fail "case 4: rc=$RC out=$OUT"
case "$OUT" in
  VISUAL-VERIFY-MISSING:*) pass "case 4: the verdict is VISUAL-VERIFY-MISSING" ;;
  *) fail "case 4: no MISSING verdict: out=$OUT" ;;
esac

# ---------------------------------------------------------------------------
# 5. The row's role is not verifier; 6. the store is empty; 7. the key
#    starts elsewhere — all three are MISSING, exit 1.
# ---------------------------------------------------------------------------
make_wt "$(dispatch_json visual-verify reviewer completed)" declare-section
touch_paths "app/src/Widget.tsx"
run_guard
[ "$RC" -eq 1 ] && pass "case 5: a non-verifier row is not evidence, exit 1" \
  || fail "case 5: rc=$RC out=$OUT"

make_wt "[]" declare-section
touch_paths "app/src/Widget.tsx"
run_guard
[ "$RC" -eq 1 ] && pass "case 6: an empty store is MISSING, exit 1" \
  || fail "case 6: rc=$RC out=$OUT"

make_wt "$(dispatch_json some-other-key verifier completed)" declare-section
touch_paths "app/src/Widget.tsx"
run_guard
[ "$RC" -eq 1 ] && pass "case 7: a foreign key prefix is not evidence, exit 1" \
  || fail "case 7: rc=$RC out=$OUT"

# ---------------------------------------------------------------------------
# 8. Cannot answer — exit 2, never a verdict: store unreachable, store
#    output not JSON, non-directory worktree, empty change name, empty
#    merge-base, and a change name outside the allowlist.
# ---------------------------------------------------------------------------
make_wt UNREACHABLE declare-section
touch_paths "app/src/Widget.tsx"
run_guard
[ "$RC" -eq 2 ] && pass "case 8a: an unreachable store exits 2" \
  || fail "case 8a: rc=$RC out=$OUT"
case "$OUT" in
  VISUAL-VERIFY-OK:*|VISUAL-VERIFY-MISSING:*) fail "case 8a: a verdict was printed despite exit 2: out=$OUT" ;;
  *) pass "case 8a: no verdict on an inability to answer" ;;
esac

make_wt NOT-JSON declare-section
touch_paths "app/src/Widget.tsx"
run_guard
[ "$RC" -eq 2 ] && pass "case 8b: non-JSON store output exits 2" \
  || fail "case 8b: rc=$RC out=$OUT"

make_wt "[]" declare-section
set +e
OUT="$(PATH="$WT/bin:$PATH" "$GUARD" "$WT/no-such-dir" demo "$BASE" 2>&1)"; RC=$?
set -e
[ "$RC" -eq 2 ] && pass "case 8c: a non-directory worktree exits 2" \
  || fail "case 8c: rc=$RC out=$OUT"

set +e
OUT="$(PATH="$WT/bin:$PATH" "$GUARD" "$WT" "" "$BASE" 2>&1)"; RC=$?
set -e
[ "$RC" -eq 2 ] && pass "case 8d: an empty change name exits 2" \
  || fail "case 8d: rc=$RC out=$OUT"

set +e
OUT="$(PATH="$WT/bin:$PATH" "$GUARD" "$WT" demo "" 2>&1)"; RC=$?
set -e
[ "$RC" -eq 2 ] && pass "case 8e: an empty merge-base exits 2" \
  || fail "case 8e: rc=$RC out=$OUT"

set +e
OUT="$(PATH="$WT/bin:$PATH" "$GUARD" "$WT" ../evil "$BASE" 2>&1)"; RC=$?
set -e
[ "$RC" -eq 2 ] && pass "case 8f: a change name outside the allowlist exits 2" \
  || fail "case 8f: rc=$RC out=$OUT"

# ---------------------------------------------------------------------------
if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-visual-verify-dispatched: all cases pass\n'
