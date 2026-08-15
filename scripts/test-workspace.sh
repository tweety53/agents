#!/usr/bin/env bash
# test-workspace.sh — round-trip harness for scripts/workspace.sh's create,
# remove and survivors commands, exercised against the real myflow-postgres
# container (superuser myflow, database myflow, host port 5433).
#
# Skips the whole file, printing why, when that container is not reachable —
# the same "SKIP: ... not reachable" style stats/health_test.go and
# stats/internal/store/testsupport_test.go use for the Go suite, so this
# harness behaves consistently with it in an environment with no docker
# stack running.
#
# WHY A THROWAWAY ID RATHER THAN A CHANGE'S OWN. Reusing this change's own
# workspace id would race remove/create against a genuine worktree using it.
# A private id namespaced under "wstest-" plus this process's pid can never
# collide with a real workspace id, which is always `<prefix>-<4-hex-digest>`
# per skills/myflow-contracts/workspace-isolation.md.
#
# WHAT THIS ASSERTS. The round-trip the plan requires, in order: create makes
# myflow_<id>; create again is a no-op (not an error); survivors then prints
# that database and exits 0; remove drops it; survivors then prints nothing
# and STILL exits 0 (exit 0 + empty output is the only result that verifies
# a removal, per project-configuration.md's "What survivors prints"). Also:
# create/remove refuse an empty or a missing id rather than operating on the
# shared "myflow" database; `remove` with no flag still behaves exactly as
# before (no --force); `remove --force` drops a database that has a live
# connection, which plain `remove` cannot; and `remove --force` refuses an id
# whose database does not end in "_uitest", so a command-line/environment id
# that does not name the UI-test database can never reach the forced drop
# (stats/Makefile's own UITEST_ID `override` is the first layer against that;
# this is the second, independent one, inside workspace.sh itself).

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT="$SCRIPT_DIR/workspace.sh"
CONTAINER="myflow-postgres"
SUPERUSER="myflow"

if ! docker exec "$CONTAINER" true >/dev/null 2>&1; then
  echo "SKIP: $CONTAINER compose stack not reachable — scripts/test-workspace.sh needs it" >&2
  exit 0
fi

[ -x "$SCRIPT" ] || { echo "ERROR: $SCRIPT not found or not executable" >&2; exit 2; }

ID="wstest-$$"
DB="myflow_${ID//-/_}"
ID2="wstest-force-$$-uitest"
DB2="myflow_${ID2//-/_}"

PASSED=0
FAILED=0
pass() { PASSED=$((PASSED + 1)); printf '  ok: %s\n' "$1"; }
fail() { FAILED=$((FAILED + 1)); printf '  FAIL: %s\n      %s\n' "$1" "$2" >&2; }

db_exists() {
  local out
  out="$(docker exec "$CONTAINER" psql -U "$SUPERUSER" -d "$SUPERUSER" -tAc \
    "select 1 from pg_database where datname = '$1'" 2>/dev/null)"
  [ "$out" = "1" ]
}

cleanup() {
  [ -n "${HOLD_PID:-}" ] && kill "$HOLD_PID" >/dev/null 2>&1 || true
  docker exec "$CONTAINER" dropdb -U "$SUPERUSER" --if-exists "$DB" >/dev/null 2>&1 || true
  docker exec "$CONTAINER" dropdb -U "$SUPERUSER" --if-exists --force "$DB2" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "== round trip for $ID =="

# 1. create makes the database.
"$SCRIPT" create "$ID" >/dev/null
RC=$?
[ "$RC" -eq 0 ] && pass "create exits 0" || fail "create exits 0" "rc=$RC"
db_exists "$DB" && pass "create makes $DB" || fail "create makes $DB" "no row in pg_database"

# 2. create is a no-op when the database already exists (a re-run of the
#    pipeline in an existing worktree is not an error).
"$SCRIPT" create "$ID" >/dev/null
RC=$?
[ "$RC" -eq 0 ] && pass "create is a no-op on an existing database" \
  || fail "create is a no-op on an existing database" "rc=$RC"

# 3. survivors then prints that database and exits 0.
OUT="$("$SCRIPT" survivors "$ID")"
RC=$?
[ "$RC" -eq 0 ] && pass "survivors exits 0 while the database survives" \
  || fail "survivors exits 0 while the database survives" "rc=$RC"
[ "$OUT" = "$DB" ] && pass "survivors prints $DB" || fail "survivors prints $DB" "got [$OUT]"

# 4. remove drops it.
"$SCRIPT" remove "$ID" >/dev/null
RC=$?
[ "$RC" -eq 0 ] && pass "remove exits 0" || fail "remove exits 0" "rc=$RC"
db_exists "$DB" && fail "remove drops $DB" "still present in pg_database" \
  || pass "remove drops $DB"

# 5. survivors then prints nothing and still exits 0.
OUT="$("$SCRIPT" survivors "$ID")"
RC=$?
[ "$RC" -eq 0 ] && pass "survivors exits 0 with nothing surviving" \
  || fail "survivors exits 0 with nothing surviving" "rc=$RC"
[ -z "$OUT" ] && pass "survivors prints nothing after remove" \
  || fail "survivors prints nothing after remove" "got [$OUT]"

# 6. create/remove refuse an empty or missing id rather than operating on
#    the shared "myflow" database.
"$SCRIPT" create "" >/dev/null 2>&1
[ $? -ne 0 ] && pass "create refuses an empty id" || fail "create refuses an empty id" "exited 0"
"$SCRIPT" create >/dev/null 2>&1
[ $? -ne 0 ] && pass "create refuses a missing id" || fail "create refuses a missing id" "exited 0"
"$SCRIPT" remove "" >/dev/null 2>&1
[ $? -ne 0 ] && pass "remove refuses an empty id" || fail "remove refuses an empty id" "exited 0"
"$SCRIPT" remove >/dev/null 2>&1
[ $? -ne 0 ] && pass "remove refuses a missing id" || fail "remove refuses a missing id" "exited 0"
db_exists "$SUPERUSER" && pass "the shared $SUPERUSER database is untouched" \
  || fail "the shared $SUPERUSER database is untouched" "gone?!"

# 7. remove --force drops a database with a live connection; plain remove
#    (no flag) cannot, and default behaviour is otherwise unchanged.
echo "== --force round trip for $ID2 =="
"$SCRIPT" create "$ID2" >/dev/null
docker exec "$CONTAINER" psql -U "$SUPERUSER" -d "$DB2" -c "select pg_sleep(30)" \
  >/dev/null 2>&1 &
HOLD_PID=$!
for i in $(seq 1 20); do
  HELD="$(docker exec "$CONTAINER" psql -U "$SUPERUSER" -d "$SUPERUSER" -tAc \
    "select count(*) from pg_stat_activity where datname = '$DB2' and pid <> pg_backend_pid()" \
    2>/dev/null)"
  [ "${HELD:-0}" -gt 0 ] 2>/dev/null && break
  sleep 0.5
done
[ "${HELD:-0}" -gt 0 ] 2>/dev/null && pass "a live connection is held open on $DB2" \
  || fail "a live connection is held open on $DB2" "no connection observed in pg_stat_activity"

"$SCRIPT" remove "$ID2" >/dev/null 2>&1
RC=$?
[ "$RC" -ne 0 ] && pass "remove (no flag) fails against a database with a live connection" \
  || fail "remove (no flag) fails against a database with a live connection" "exited 0"
db_exists "$DB2" && pass "$DB2 still exists after the unforced remove failed" \
  || fail "$DB2 still exists after the unforced remove failed" "already gone"

"$SCRIPT" remove "$ID2" --force >/dev/null
RC=$?
[ "$RC" -eq 0 ] && pass "remove --force exits 0 against a database with a live connection" \
  || fail "remove --force exits 0 against a database with a live connection" "rc=$RC"
db_exists "$DB2" && fail "remove --force drops $DB2" "still present in pg_database" \
  || pass "remove --force drops $DB2"

kill "$HOLD_PID" >/dev/null 2>&1 || true
unset HOLD_PID

# 8. remove --force refuses an id whose database does not end in "_uitest",
#    without touching the target database at all -- the second, independent
#    layer against the defect stats/Makefile's UITEST_ID `override` guards
#    on the caller's side (KAN-180 review, F1/F4).
echo "== --force suffix guard for $ID =="
"$SCRIPT" create "$ID" >/dev/null
"$SCRIPT" remove "$ID" --force >/dev/null 2>&1
RC=$?
[ "$RC" -ne 0 ] && pass "remove --force refuses a database not ending in _uitest" \
  || fail "remove --force refuses a database not ending in _uitest" "exited 0"
db_exists "$DB" && pass "$DB survives the refused --force" \
  || fail "$DB survives the refused --force" "dropped despite the refusal"
"$SCRIPT" remove "$ID" >/dev/null 2>&1

echo
echo "$PASSED passed, $FAILED failed"
[ "$FAILED" -eq 0 ]
