#!/usr/bin/env bash
# Assertion harness for check-task-reviewer-single-dispatch.sh.
#
# Follows test-check-panel-fix-single-dispatch.sh's stub-`flow`-on-PATH
# pattern for dispatch/decision rows, and test-plan-dispatch-groups.sh's
# fixture-tasks.md `task` helper for group membership — this guard's own
# group-membership check runs the REAL plan-dispatch-bundles.sh and
# plan-dispatch-groups.sh against a fixture tasks.md, never a stub, since
# both are pure and standard-library.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-task-reviewer-single-dispatch.sh"
FAILED=0

WORKTREES=()
cleanup() {
  [ "${#WORKTREES[@]}" -eq 0 ] && return 0
  for wt in "${WORKTREES[@]}"; do
    rm -rf "$wt"
  done
}
trap cleanup EXIT

# task <id> <files-path> [after-value] — one unchecked task, one **Files:**
# entry, an optional **After:** field. Verbatim from test-plan-dispatch-
# groups.sh's own helper.
task() {
  local id="$1" files="$2" after="${3:-}"
  printf -- '- [ ] %s. Task %s\n\n' "$id" "$id"
  printf '**Files:**\n- Create: `%s`\n\n' "$files"
  if [ -n "$after" ]; then
    printf '**After:** %s\n\n' "$after"
  fi
  printf -- '  - [ ] **Step 1: do it**\n\n'
}

# fixture_tasks_md <worktree> <change-name> -- writes the shared fixture
# plan (test-plan-dispatch-groups.sh's own case 2: two independent chains
# 1->3, 2->4, plus a join point 5 straddling both) at
# <worktree>/spectre/changes/<name>/tasks.md, so group-dispatch-groups.sh
# answers `group 1: 1 3`, `group 2: 2 4`, `group 3: 5` for every case below.
fixture_tasks_md() {
  local wt="$1" name="$2"
  mkdir -p "$wt/spectre/changes/$name"
  {
    task 1 a.txt none
    task 2 b.txt none
    task 3 c.txt "Task 1"
    task 4 d.txt "Task 2"
    task 5 e.txt "Task 3, 4"
  } > "$wt/spectre/changes/$name/tasks.md"
}

# dispatch_json <key> <role> <token> [...] -- verbatim from
# test-check-panel-fix-single-dispatch.sh's own helper.
dispatch_json() {
  jq -nc '
    [$ARGS.positional as $a
     | range(0; ($a | length) / 3)
     | {key: $a[. * 3], role: $a[. * 3 + 1], sessionToken: $a[. * 3 + 2]}]
  ' --args -- "$@"
}

# make_worktree <change-name> <dispatches-json> [<decisions-json>] -- a
# worktree-shaped sandbox: the fixture plan plus a stub `flow` on its own
# bin/ answering `record dispatches` and `record decisions`; any other verb
# exits 2. Prints the worktree path.
make_worktree() {
  local name="$1" dispatches="$2" decisions="${3:-[]}"
  local wt
  wt="$(mktemp -d "${TMPDIR:-/tmp}/check-task-reviewer-single-dispatch-test.XXXXXX")" || {
    printf 'make_worktree: mktemp failed -- aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  WORKTREES+=("$wt")
  fixture_tasks_md "$wt" "$name"
  mkdir -p "$wt/bin"
  printf '%s' "$dispatches" > "$wt/bin/dispatches.json"
  printf '%s' "$decisions" > "$wt/bin/decisions.json"
  cat > "$wt/bin/flow" <<'STUB'
#!/usr/bin/env bash
dir="$(dirname -- "$0")"
case "${2:-}" in
  dispatches) cat "$dir/dispatches.json" ;;
  decisions)  cat "$dir/decisions.json" ;;
  *) printf 'stub flow: unexpected invocation: %s\n' "$*" >&2; exit 2 ;;
esac
exit 0
STUB
  chmod +x "$wt/bin/flow"
  printf '%s' "$wt"
}

make_worktree_store_unreachable() {
  local name="$1" wt
  wt="$(mktemp -d "${TMPDIR:-/tmp}/check-task-reviewer-single-dispatch-test.XXXXXX")" || {
    printf 'make_worktree_store_unreachable: mktemp failed -- aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  WORKTREES+=("$wt")
  fixture_tasks_md "$wt" "$name"
  mkdir -p "$wt/bin"
  cat > "$wt/bin/flow" <<'STUB'
#!/usr/bin/env bash
echo "flow: connect: connection refused" >&2
exit 1
STUB
  chmod +x "$wt/bin/flow"
  printf '%s' "$wt"
}

run_guard() {
  local wt="$1" name="${2:-demo}" token="${3-mf-tok}"
  PATH="$wt/bin:$PATH" "$GUARD" "$wt" "$name" "$token"
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
  local errfile
  errfile="$(mktemp "${TMPDIR:-/tmp}/check-task-reviewer-single-dispatch-test-stderr.XXXXXX")"
  set +e
  out="$("$@" 2>"$errfile")"; got=$?
  set -e
  err="$(cat "$errfile")"
  rm -f "$errfile"
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
# 1. One bundle per group (1+3, 2+4, 5), class big -- exit 0, no stderr.
# ===========================================================================
wt="$(make_worktree demo "$(dispatch_json task-1+3-reviewer reviewer mf-tok task-2+4-reviewer reviewer mf-tok task-5-reviewer reviewer mf-tok)")"
expect_exit_and_no_stderr 'case 1: one bundle per group exits 0 with no stderr' 0 run_guard "$wt"

# ===========================================================================
# 2. Original plus -retry on one bundle -- the handshake's one allowed
#    second dispatch -- exit 0.
# ===========================================================================
wt="$(make_worktree demo "$(dispatch_json task-1+3-reviewer reviewer mf-tok task-1+3-reviewer-retry reviewer mf-tok)")"
expect_exit 'case 2: original plus -retry exits 0' 0 run_guard "$wt"

# ===========================================================================
# 3. Two non-retry rows under the SAME key -- one dispatch per task run
#    twice -- exit 1, naming the key.
# ===========================================================================
wt="$(make_worktree demo "$(dispatch_json task-1+3-reviewer reviewer mf-tok task-1+3-reviewer reviewer mf-tok)")"
expect_exit_and_names 'case 3: two originals under one key exit 1, naming it' 1 'task-1+3-reviewer' run_guard "$wt"

# ===========================================================================
# 4. Tasks 1 and 3 (same group) dispatched as two SEPARATE reviewer bundles
#    -- exactly the KAN-527 shape -- exit 1, naming both keys.
# ===========================================================================
wt="$(make_worktree demo "$(dispatch_json task-1-reviewer reviewer mf-tok task-3-reviewer reviewer mf-tok)")"
expect_exit_and_names 'case 4: same-group tasks split across two bundles exit 1' 1 'group' run_guard "$wt"

# ===========================================================================
# 5. Invented key task-5-review (not -reviewer) -- out of canonical shape --
#    exit 1, naming it verbatim. The jq prefix filter only requires
#    `task-<n>-reviewer`-prefixed keys to reach the shape check, so a key
#    with a totally different role never reaches this guard at all — this
#    case instead uses a malformed suffix on an otherwise-matching prefix.
# ===========================================================================
wt="$(make_worktree demo "$(dispatch_json task-1-reviewer-oops reviewer mf-tok)")"
expect_exit_and_names 'case 5: malformed key exits 1, naming it verbatim' 1 'task-1-reviewer-oops' run_guard "$wt"

# ===========================================================================
# 6. Two original bundles for DIFFERENT groups (1+3, then 2+4) on class
#    `small` -- implement.md requires ALL gate-fired tasks of the run in
#    one bundle at the last boundary on small/regular -- exit 1.
# ===========================================================================
DECISIONS_SMALL='[{"decision":{"class":"small"}}]'
wt="$(make_worktree demo "$(dispatch_json task-1+3-reviewer reviewer mf-tok task-2+4-reviewer reviewer mf-tok)" "$DECISIONS_SMALL")"
expect_exit_and_names 'case 6: two original bundles on class small exit 1' 1 "class 'small'" run_guard "$wt"

# ===========================================================================
# 7. Same two bundles as case 6, but class `big` (the default when no
#    decision row exists) -- two different groups in two different bundles
#    is exactly the intended shape on `big` -- exit 0.
# ===========================================================================
wt="$(make_worktree demo "$(dispatch_json task-1+3-reviewer reviewer mf-tok task-2+4-reviewer reviewer mf-tok)")"
expect_exit 'case 7: two original bundles on class big (default) exits 0' 0 run_guard "$wt"

# ===========================================================================
# 8. A fix-round bundle for group 1 (task-1+3-reviewer-fix-1) alongside the
#    original task-1+3-reviewer -- a fix key is exempt from the
#    same-group-split check by construction -- exit 0.
# ===========================================================================
wt="$(make_worktree demo "$(dispatch_json task-1+3-reviewer reviewer mf-tok task-1+3-reviewer-fix-1 reviewer mf-tok)")"
expect_exit 'case 8: original plus its own fix-round bundle exits 0' 0 run_guard "$wt"

# ===========================================================================
# 9. Store unreachable -- cannot answer -- exit 2.
# ===========================================================================
wt="$(make_worktree_store_unreachable demo)"
expect_exit 'case 9: unreachable store exits 2' 2 run_guard "$wt"

# ===========================================================================
# 10. Missing tasks.md -- cannot establish group membership -- exit 2.
# ===========================================================================
wt="$(mktemp -d "${TMPDIR:-/tmp}/check-task-reviewer-single-dispatch-test.XXXXXX")"
WORKTREES+=("$wt")
mkdir -p "$wt/bin"
printf '%s' '[]' > "$wt/bin/dispatches.json"
cat > "$wt/bin/flow" <<'STUB'
#!/usr/bin/env bash
dir="$(dirname -- "$0")"
case "${2:-}" in
  dispatches) cat "$dir/dispatches.json" ;;
  *) exit 2 ;;
esac
exit 0
STUB
chmod +x "$wt/bin/flow"
expect_exit 'case 10: missing tasks.md exits 2' 2 run_guard "$wt"

# ===========================================================================
# 11. Empty session token is a usage error, matching check-panel-fix-single-
#     dispatch.sh's own `[[ -n "$TOKEN" ]]` requirement -- exit 2.
# ===========================================================================
wt="$(make_worktree demo "$(dispatch_json task-1+3-reviewer reviewer mf-tok)")"
expect_exit 'case 11: empty token is a usage error, exits 2' 2 run_guard "$wt" demo ""

if [ "$FAILED" -ne 0 ]; then
  exit 1
fi
exit 0
