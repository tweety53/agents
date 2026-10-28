#!/usr/bin/env bash
# test-live-store-copy.sh — live-store-copy.sh copies only into
# flow_<name>_live, starts and stops only its own flowd, and refuses the
# protected ports and a malformed name before any side effect (kan-1041
# self-review). docker, make, go and curl are stubs on PATH; no database or
# daemon is touched. Exit codes: 0 pass, 1 fail.
set -u
cd "$(dirname "$0")/.." || exit 1
rc=0
fail() { printf 'FAIL: %s\n' "$1"; rc=1; }
stub=$(mktemp -d)
name="lsctest$$"
port=$((20000 + $$ % 20000))
log="$stub/log"
trap 'rm -rf "$stub" "/tmp/flow-live-$name" "/tmp/flowd-$port.pid"' EXIT
export STUB="$stub" LOG="$log"
cat >"$stub/docker" <<'S'
#!/usr/bin/env bash
echo "docker $*" >>"$LOG"
S
cat >"$stub/make" <<'S'
#!/usr/bin/env bash
echo "make $*" >>"$LOG"
S
cat >"$stub/curl" <<'S'
#!/usr/bin/env bash
[ -e "$STUB/up" ]
S
cat >"$stub/go" <<'S'
#!/usr/bin/env bash
echo "go $*" >>"$LOG"
out=$3
printf '#!/usr/bin/env bash\nprintf "%%s\\n%%s\\n" $$ "$0" >"$TMPDIR/flowd-$FLOWD_PORT.pid"\necho "$FLOWD_DSN" >>"$LOG"\ntouch "$STUB/up"\nexec sleep 300\n' >"$out"
chmod +x "$out"
S
chmod +x "$stub"/*
export PATH="$stub:$PATH"
lsc() { scripts/live-store-copy.sh "$@" >"$stub/out" 2>&1; }

for bad in "" Flow ../a "a-b;rm"; do
  : >"$log"
  lsc up "$bad" "$port"; [ $? -eq 2 ] || fail "up '$bad': not exit 2"
  [ -s "$log" ] && fail "up '$bad': ran $(cat "$log")"
done
for p in 4173 4174 80 x; do
  : >"$log"
  lsc up "$name" "$p"; [ $? -eq 2 ] || fail "up --port $p: not exit 2"
  [ -s "$log" ] && fail "up port $p: ran $(cat "$log")"
done

: >"$log"
lsc up "$name" "$port" || fail "up: exit $? ($(cat "$stub/out"))"
grep -qx "docker exec -i flow-postgres createdb -U flow flow_${name}_live" "$log" || fail "up: no createdb of flow_${name}_live"
grep -qx "docker exec -i flow-postgres pg_dump -U flow --no-owner --no-acl flow" "$log" || fail "up: no read-only pg_dump of flow"
grep -q "psql -q -v ON_ERROR_STOP=1 -U flow -d flow_${name}_live$" "$log" || fail "up: no restore into flow_${name}_live"
grep -qx "postgres://flow:flow@localhost:5433/flow_${name}_live?sslmode=disable" "$log" || fail "up: flowd not pointed at the copy"
grep -qF "http://127.0.0.1:$port" "$stub/out" || fail "up: address not printed"
pid=$(head -n1 "/tmp/flowd-$port.pid" 2>/dev/null)
kill -0 "$pid" 2>/dev/null || fail "up: scratch flowd not running"
lsc up "$name" "$port"; [ $? -eq 2 ] || fail "second up: not exit 2"

sleep 300 &
sentinel=$!
: >"$log"
lsc down "$name" || fail "down: exit $? ($(cat "$stub/out"))"
kill -0 "$pid" 2>/dev/null && fail "down: scratch flowd still running"
kill -0 "$sentinel" 2>/dev/null || fail "down: killed a process it did not start"
kill "$sentinel"
grep -qx "docker exec -i flow-postgres dropdb -U flow --if-exists --force flow_${name}_live" "$log" || fail "down: no dropdb of flow_${name}_live"
grep -E '(dropdb|createdb).* flow$|-d flow$' "$log" && fail "down: touched the flow database"
[ -e "/tmp/flow-live-$name" ] && fail "down: state dir left"
lsc down "$name" || fail "second down: exit $?"

[ "$rc" -eq 0 ] && printf 'all cases passed\n'
exit "$rc"
