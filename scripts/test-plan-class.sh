#!/usr/bin/env bash
# test-plan-class.sh — assertion harness for scripts/plan-class.sh.
#
# Builds throwaway tasks.md fixtures under a sandboxed TMPDIR and asserts
# plan-class.sh's three-line stdout and exit status. Never touches this
# repository's own tasks.md files except the one read-only reproduction the
# task's own worked check names.
#
# Same shape as scripts/test-check-base-moved.sh: FAILURES counter,
# fail()/pass() helpers, an EXIT trap removing every fixture dir.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/plan-class.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

DIRS=()
cleanup() {
  [ "${#DIRS[@]}" -eq 0 ] && return 0
  for d in "${DIRS[@]}"; do
    rm -rf "$d"
  done
}
trap cleanup EXIT

# new_fixture <change-name> -> sets TASKS_FILE to
# <tmp>/changes/<change-name>/tasks.md (so basename "$(dirname "$1")" reads
# back <change-name>, exactly as plan-class.sh derives it).
new_fixture() {
  local name="$1"
  local dir
  dir="$(mktemp -d "${TMPDIR:-/tmp}/plan-class-test.XXXXXX")"
  DIRS+=("$dir")
  mkdir -p "$dir/changes/$name"
  TASKS_FILE="$dir/changes/$name/tasks.md"
}

# task_line <n> <files...> -> one "- [ ] <n>. Task <n>" line plus its
# "**Files:** \`f1\`, \`f2\`" line, appended to TASKS_FILE.
task_line() {
  local n="$1"
  shift
  {
    printf -- '- [ ] %s. Task %s\n' "$n" "$n"
    if [ "$#" -gt 0 ]; then
      local files=""
      local f
      for f in "$@"; do
        if [ -n "$files" ]; then files="$files, "; fi
        files="$files\`$f\`"
      done
      printf '**Files:** %s\n' "$files"
    fi
  } >>"$TASKS_FILE"
}

# make_tasks <change-name> <task-count> [migration] -> a fixture with
# <task-count> tasks, each carrying one distinct file under src/, plus one
# migration-path file on task 1 when the third argument is "migration".
make_tasks() {
  local name="$1" count="$2" want_migration="${3:-}"
  new_fixture "$name"
  : >"$TASKS_FILE"
  local i
  for i in $(seq 1 "$count"); do
    if [ "$i" = 1 ] && [ "$want_migration" = migration ]; then
      task_line "$i" "stats/internal/store/migrations/0099_x.sql"
    else
      task_line "$i" "src/file$i.go"
    fi
  done
}

run_guard() {
  set +e
  OUT="$("$GUARD" "$@" 2>&1)"
  RC=$?
  set -e
}

# --- case: 5 tasks, 12 files, repos=1 -> small ---
new_fixture small-5-12
: >"$TASKS_FILE"
for i in 1 2 3 4 5; do
  case "$i" in
    1) task_line "$i" a1.go a2.go a3.go ;;
    2) task_line "$i" b1.go b2.go b3.go ;;
    3) task_line "$i" c1.go c2.go c3.go ;;
    4) task_line "$i" d1.go d2.go d3.go ;;
    5) task_line "$i" ;;
  esac
done
run_guard "$TASKS_FILE" 1
if [ "$RC" -eq 0 ] && printf '%s\n' "$OUT" | grep -q '^inputs: tasks=5 files=12 repos=1 migration=no spec=no red=no unverified=no$' \
  && printf '%s\n' "$OUT" | grep -q '^class: small$'; then
  pass "5 tasks / 12 files / repos=1 -> small"
else
  fail "5 tasks / 12 files / repos=1 -> small (rc=$RC out=$OUT)"
fi

# --- case: 6 tasks -> regular ---
make_tasks regular-6 6
run_guard "$TASKS_FILE" 1
if [ "$RC" -eq 0 ] && printf '%s\n' "$OUT" | grep -q '^class: regular$'; then
  pass "6 tasks -> regular"
else
  fail "6 tasks -> regular (rc=$RC out=$OUT)"
fi

# --- case: 15 tasks -> big ---
make_tasks big-15 15
run_guard "$TASKS_FILE" 1
if [ "$RC" -eq 0 ] && printf '%s\n' "$OUT" | grep -q '^class: big$'; then
  pass "15 tasks -> big"
else
  fail "15 tasks -> big (rc=$RC out=$OUT)"
fi

# --- case: 1 migration + 8 tasks -> big ---
make_tasks migration-8-big 8 migration
run_guard "$TASKS_FILE" 1
if [ "$RC" -eq 0 ] && printf '%s\n' "$OUT" | grep -q '^class: big$' \
  && printf '%s\n' "$OUT" | grep -q 'migration=yes'; then
  pass "1 migration + 8 tasks -> big"
else
  fail "1 migration + 8 tasks -> big (rc=$RC out=$OUT)"
fi

# --- case: 1 migration + 7 tasks -> regular ---
make_tasks migration-7-regular 7 migration
run_guard "$TASKS_FILE" 1
if [ "$RC" -eq 0 ] && printf '%s\n' "$OUT" | grep -q '^class: regular$' \
  && printf '%s\n' "$OUT" | grep -q 'migration=yes'; then
  pass "1 migration + 7 tasks -> regular"
else
  fail "1 migration + 7 tasks -> regular (rc=$RC out=$OUT)"
fi

# --- case: the override never appears (the script has none) ---
make_tasks no-override 15
run_guard "$TASKS_FILE" 1
if [ "$RC" -eq 0 ] && ! printf '%s\n' "$OUT" | grep -qi 'override'; then
  pass "override never appears in plan-class.sh output"
else
  fail "override never appears in plan-class.sh output (rc=$RC out=$OUT)"
fi

# --- case: a fixed name's two rolls are equal across two runs ---
make_tasks roll-stability-name 4
run_guard "$TASKS_FILE" 1
FIRST="$OUT"
run_guard "$TASKS_FILE" 1
SECOND="$OUT"
FIRST_ROLLS="$(printf '%s\n' "$FIRST" | grep '^rolls:' || true)"
SECOND_ROLLS="$(printf '%s\n' "$SECOND" | grep '^rolls:' || true)"
if [ -n "$FIRST_ROLLS" ] && [ "$FIRST_ROLLS" = "$SECOND_ROLLS" ]; then
  pass "rolls are reproducible for a fixed change name"
else
  fail "rolls are reproducible for a fixed change name (first=$FIRST_ROLLS second=$SECOND_ROLLS)"
fi

# --- case: the rolls line carries all three rolls in shape ---
make_tasks bundle-shape 4
run_guard "$TASKS_FILE" 1
if [ "$RC" -eq 0 ] && printf '%s\n' "$OUT" \
  | grep -qE '^rolls: compact [0-9]+ · experimental [0-9]+ · bundle [0-9]+$'; then
  pass "rolls line carries compact, experimental and bundle"
else
  fail "rolls line carries compact, experimental and bundle (rc=$RC out=$OUT)"
fi

# --- case: a fixed name's three rolls are equal across two runs ---
make_tasks roll-stability-name-three 4
run_guard "$TASKS_FILE" 1
FIRST3="$OUT"
run_guard "$TASKS_FILE" 1
SECOND3="$OUT"
FIRST3_ROLLS="$(printf '%s\n' "$FIRST3" | grep '^rolls:' || true)"
SECOND3_ROLLS="$(printf '%s\n' "$SECOND3" | grep '^rolls:' || true)"
if [ -n "$FIRST3_ROLLS" ] && [ "$FIRST3_ROLLS" = "$SECOND3_ROLLS" ]; then
  pass "all three rolls are reproducible for a fixed change name"
else
  fail "all three rolls are reproducible for a fixed change name (first=$FIRST3_ROLLS second=$SECOND3_ROLLS)"
fi

# --- case: a change name whose bundle_roll < 30 (static grouping) ---
# name found by trying candidates: sha256("bundle-static-2bundle") mod 100 = 26
make_tasks bundle-static-2 4
run_guard "$TASKS_FILE" 1
if [ "$RC" -eq 0 ] && printf '%s\n' "$OUT" | grep -q '^rolls:.*bundle 26$'; then
  pass "bundle-static-2 rolls bundle 26 (<30 -> static)"
else
  fail "bundle-static-2 rolls bundle 26 (<30 -> static) (rc=$RC out=$OUT)"
fi

# --- case: a change name whose bundle_roll >= 30 (free grouping) ---
# name found by trying candidates: sha256("bundle-free-1bundle") mod 100 = 46
make_tasks bundle-free-1 4
run_guard "$TASKS_FILE" 1
if [ "$RC" -eq 0 ] && printf '%s\n' "$OUT" | grep -q '^rolls:.*bundle 46$'; then
  pass "bundle-free-1 rolls bundle 46 (>=30 -> free)"
else
  fail "bundle-free-1 rolls bundle 46 (>=30 -> free) (rc=$RC out=$OUT)"
fi

# --- case: missing file -> exit 2 ---
run_guard "/nonexistent/tasks.md" 1
if [ "$RC" -eq 2 ]; then
  pass "missing file -> exit 2"
else
  fail "missing file -> exit 2 (rc=$RC out=$OUT)"
fi

# --- case: non-integer repos -> exit 2 ---
make_tasks bad-repos 3
run_guard "$TASKS_FILE" notanumber
if [ "$RC" -eq 2 ]; then
  pass "non-integer repos -> exit 2"
else
  fail "non-integer repos -> exit 2 (rc=$RC out=$OUT)"
fi

if [ "$FAILURES" -gt 0 ]; then
  printf 'test-plan-class.sh: %d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'test-plan-class.sh: all cases passed\n'
