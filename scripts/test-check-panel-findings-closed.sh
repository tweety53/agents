#!/usr/bin/env bash
# Assertion harness for check-panel-findings-closed.sh.
#
# Every case builds a worktree-shaped sandbox and a stub `flow` binary placed
# ahead of the real one on PATH inside that sandbox: the stub answers
# `record findings -change <name> [-C <dir>]` with a canned JSON array (the
# shape `flow record findings` itself prints -- one object per finding with
# at least `ref` and `status`), or exits non-zero to simulate a store the
# guard could not reach. Follows test-check-panel-reproducers.sh's own shape
# for these helpers.
#
# `-e` as well as `-u`/`pipefail`, matching this repository's sibling
# harnesses: without `-e`, a failed `mktemp` in make_worktree_json would
# leave a stale sandbox path in play and the harness would keep going anyway.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-panel-findings-closed.sh"
FAILED=0

WORKTREES=()
cleanup() {
  [ "${#WORKTREES[@]}" -eq 0 ] && return 0
  for wt in "${WORKTREES[@]}"; do
    rm -rf "$wt"
  done
}
trap cleanup EXIT

# findings_json <ref> <status> [<ref> <status> ...] -- builds a compact JSON
# array of finding objects from ref/status pairs, via `jq -n --args` rather
# than hand-quoted string interpolation.
findings_json() {
  jq -nc '
    [$ARGS.positional as $a
     | range(0; ($a | length) / 2)
     | {ref: $a[. * 2], status: $a[. * 2 + 1]}]
  ' --args -- "$@"
}

# make_worktree_json <json-array> [<stderr-line>] -- a worktree-shaped
# sandbox carrying a stub `flow` on its own bin/, which prints <json-array>
# on stdout for `record findings` (optionally also printing <stderr-line> on
# stderr first, simulating flow's own `flow: using FLOW_ADDR=...` diagnostic)
# and exits 0 regardless of the flags it was called with. Prints the
# worktree path.
make_worktree_json() {
  local wt json="$1" stderr_line="${2:-}"
  wt="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-findings-closed-test.XXXXXX")" || {
    printf 'make_worktree_json: mktemp failed -- aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  WORKTREES+=("$wt")
  mkdir -p "$wt/bin"
  printf '%s' "$json" > "$wt/bin/findings.json"
  printf '%s' "$stderr_line" > "$wt/bin/stderr.txt"
  cat > "$wt/bin/flow" <<'STUB'
#!/usr/bin/env bash
dir="$(dirname -- "$0")"
line="$(cat "$dir/stderr.txt")"
if [ -n "$line" ]; then
  echo "$line" >&2
fi
cat "$dir/findings.json"
exit 0
STUB
  chmod +x "$wt/bin/flow"
  printf '%s' "$wt"
}

# make_worktree_marker -- a worktree-shaped sandbox whose stub `flow` writes
# a marker file every time it is invoked, so a case can assert the store was
# NEVER called (a bad change name must be rejected before any store read).
make_worktree_marker() {
  local wt
  wt="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-findings-closed-test.XXXXXX")" || {
    printf 'make_worktree_marker: mktemp failed -- aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  WORKTREES+=("$wt")
  mkdir -p "$wt/bin"
  cat > "$wt/bin/flow" <<STUB
#!/usr/bin/env bash
touch "$wt/bin/called.marker"
echo '[]'
exit 0
STUB
  chmod +x "$wt/bin/flow"
  printf '%s' "$wt"
}

# make_worktree_store_unreachable -- a worktree-shaped sandbox whose stub
# `flow` exits non-zero and prints nothing useful to stdout, simulating a
# store `record findings` could not reach.
make_worktree_store_unreachable() {
  local wt
  wt="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-findings-closed-test.XXXXXX")" || {
    printf 'make_worktree_store_unreachable: mktemp failed -- aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  WORKTREES+=("$wt")
  mkdir -p "$wt/bin"
  cat > "$wt/bin/flow" <<'STUB'
#!/usr/bin/env bash
echo "flow: connect: connection refused" >&2
exit 1
STUB
  chmod +x "$wt/bin/flow"
  printf '%s' "$wt"
}

run_with_stub() {
  local guard="$1" wt="$2" name="$3"
  PATH="$wt/bin:$PATH" "$guard" "$wt" "$name"
}

run_guard() {
  run_with_stub "$GUARD" "$1" "${2:-demo}"
}

expect_exit() {
  local label="$1" want="$2"; shift 2
  local out got
  set +e
  out="$("$@" 2>&1)"; got=$?
  set -e
  if [[ "$got" != "$want" ]]; then
    printf 'FAIL %s: expected exit %s, got %s\n%s\n' "$label" "$want" "$got" "$out"
    FAILED=1
  else
    printf 'ok: %s\n' "$label"
  fi
}

expect_exit_and_names() {
  local label="$1" want="$2" needle="$3"; shift 3
  local out got
  set +e
  out="$("$@" 2>&1)"; got=$?
  set -e
  if [[ "$got" != "$want" ]]; then
    printf 'FAIL %s: expected exit %s, got %s\n%s\n' "$label" "$want" "$got" "$out"
    FAILED=1
    return
  fi
  case "$out" in
    *"$needle"*) printf 'ok: %s\n' "$label" ;;
    *)
      printf 'FAIL %s: expected output to name %s, got:\n%s\n' "$label" "$needle" "$out"
      FAILED=1
      ;;
  esac
}

expect_exit_and_no_stderr() {
  local label="$1" want="$2"; shift 2
  local out err got
  set +e
  out="$("$@" 2>/tmp/check-panel-findings-closed-test-stderr.$$)"; got=$?
  set -e
  err="$(cat "/tmp/check-panel-findings-closed-test-stderr.$$")"
  rm -f "/tmp/check-panel-findings-closed-test-stderr.$$"
  if [[ "$got" != "$want" ]]; then
    printf 'FAIL %s: expected exit %s, got %s\nstdout: %s\nstderr: %s\n' "$label" "$want" "$got" "$out" "$err"
    FAILED=1
    return
  fi
  if [ -n "$err" ]; then
    printf 'FAIL %s: expected no stderr, got:\n%s\n' "$label" "$err"
    FAILED=1
    return
  fi
  printf 'ok: %s\n' "$label"
}

# ===========================================================================
# 1. Every finding closed (all fixed) -- exit 0, no stderr.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 fixed F2 fixed)")"
expect_exit_and_no_stderr 'case 1: every finding closed exits 0 with no stderr' 0 run_guard "$wt"

# ===========================================================================
# 2. One still open -- exit 1, F1 named on stderr.
#
#    PROVED BY MUTATION: removing the open-count branch that adds the
#    violation (deleting the `if [ -n "$OPEN_REFS" ]; ... exit 1` block, so
#    the guard always falls through to `FINDINGS-CLOSED` at exit 0) turns
#    this case's exit 1 into exit 0 -- confirmed by hand against a scratch
#    copy of the guard before this comment was written.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 open F2 fixed)")"
expect_exit_and_names 'case 2: one still-open finding exits 1, naming it' 1 'F1' run_guard "$wt"

# ===========================================================================
# 3. Several still open -- exit 1, both named.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 open F2 fixed F3 open)")"
out="$(run_with_stub "$GUARD" "$wt" demo 2>&1)" && got=0 || got=$?
if [[ "$got" == 1 ]] && [[ "$out" == *"F1"* ]] && [[ "$out" == *"F3"* ]]; then
  printf 'ok: %s\n' 'case 3: several still-open findings exit 1, naming both'
else
  printf 'FAIL %s: expected exit 1 naming F1 and F3, got exit %s:\n%s\n' 'case 3' "$got" "$out"
  FAILED=1
fi

# ===========================================================================
# 4. Withdrawn counts as closed -- exit 0.
#
#    PROVED BY MUTATION: removing the `startswith("withdrawn")` clause from
#    the guard's jq predicate (leaving only `.status != "fixed"`) turns this
#    case's exit 0 into exit 1 -- confirmed by hand against a scratch copy
#    of the guard before this comment was written.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 'withdrawn — not a real defect')")"
expect_exit 'case 4: a withdrawn finding counts as closed, exits 0' 0 run_guard "$wt"

# ===========================================================================
# 5. No findings at all -- an empty JSON array -- exit 0.
# ===========================================================================
wt="$(make_worktree_json '[]')"
expect_exit 'case 5: zero findings exits 0' 0 run_guard "$wt"

# ===========================================================================
# 6. Store unreachable -- exit 2.
# ===========================================================================
wt="$(make_worktree_store_unreachable)"
expect_exit_and_names 'case 6: an unreachable store is cannot-answer' 2 'cannot determine anything' run_guard "$wt"

# ===========================================================================
# 7. Change name outside the allowlist -- exit 2, and the store is never
#    called (the stub `flow` writes a marker file on every invocation).
# ===========================================================================
wt="$(make_worktree_marker)"
for bad_name in "../../../planted/clear" "demo*" "demo/../demo" ".hidden" "demo?x"; do
  expect_exit_and_names "case 7: change name '$bad_name' is rejected" 2 'is not a plain change name' run_with_stub "$GUARD" "$wt" "$bad_name"
  if [ -e "$wt/bin/called.marker" ]; then
    printf 'FAIL case 7: change name %s reached the store -- the marker file exists\n' "$bad_name"
    FAILED=1
    rm -f "$wt/bin/called.marker"
  fi
done

# ===========================================================================
# 8. Missing arguments -- exit 2. With zero arguments the worktree check
#    (the guard's first check, matching check-panel-reproducers.sh's own
#    order) fires first and names the missing worktree, not "usage:"; the
#    missing-change-name message is reached only once a real worktree is
#    supplied.
# ===========================================================================
expect_exit_and_names 'case 8a: no arguments at all exits 2' 2 'not a directory' "$GUARD"
wt="$(make_worktree_json '[]')"
expect_exit_and_names 'case 8b: a missing change name exits 2' 2 'usage:' "$GUARD" "$wt"

# ===========================================================================
# 9. Worktree is not a directory -- exit 2.
# ===========================================================================
not_a_dir="$(mktemp "${TMPDIR:-/tmp}/check-panel-findings-closed-test.XXXXXX")"
expect_exit_and_names 'case 9: non-directory argument exits 2 and names it' 2 'not a directory' "$GUARD" "$not_a_dir" demo
rm -f "$not_a_dir"

# ===========================================================================
# 10. A diagnostic on stderr from the store read (`flow`'s own
#     `flow: using FLOW_ADDR=...` line) does NOT break a successful read --
#     exit 0.
#
#     PROVED BY MUTATION: replacing the guard's separated capture
#     (`2>"$FINDINGS_ERR"`) with `2>&1` puts this diagnostic line at the head
#     of what jq parses, and jq fails on it -- turning this case's exit 0
#     into exit 2. Confirmed by hand against a scratch copy of the guard
#     before this comment was written.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 fixed)" 'flow: using FLOW_ADDR=http://127.0.0.1:4174')"
expect_exit 'case 10: a stderr diagnostic on the store read does not break a clean answer' 0 run_guard "$wt"

if [ "$FAILED" -ne 0 ]; then
  printf 'check-panel-findings-closed-test: one or more cases failed\n' >&2
  exit 1
fi
printf 'check-panel-findings-closed-test: all cases pass\n'
