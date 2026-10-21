#!/usr/bin/env bash
# Assertion harness for rebuild-full-app-suite-zip.sh: builds a throwaway
# project whose `screenshots` root holds the full app suite's PNGs beside
# another spec's, and runs the REAL script against it.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REBUILD="$SCRIPT_DIR/rebuild-full-app-suite-zip.sh"
FAILURES=0
fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

P="$(mktemp -d "${TMPDIR:-/tmp}/rebuild-full-app-suite-zip.XXXXXX")"
trap 'rm -rf "$P"' EXIT
mkdir -p "$P/.flow" "$P/shots/full-app-suite.spec.ts-snapshots" "$P/shots/other.spec.ts-snapshots"
printf '# p\n\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `screenshots` | `shots` |\n' >"$P/.flow/project.md"
printf a >"$P/shots/full-app-suite.spec.ts-snapshots/a.png"
printf b >"$P/shots/full-app-suite.spec.ts-snapshots/b.png"
printf c >"$P/shots/other.spec.ts-snapshots/c.png"
Z="$P/shots/full-app-suite.zip"

# Case 1: the zip holds exactly the suite's PNGs, by base name.
"$REBUILD" "$P" "$P/shots" && rc=0 || rc=$?
names=$(unzip -Z1 "$Z" | sort | tr '\n' ' ')
[ "$rc" -eq 0 ] && [ "$names" = "a.png b.png " ] && pass "case 1 the suite's PNGs only" \
  || fail "case 1: exit $rc, entries '$names'"

# Case 2: a rebuild over unchanged images is byte-identical, whatever their mtimes.
cp "$Z" "$P/first.zip"
touch "$P/shots/full-app-suite.spec.ts-snapshots/a.png"
"$REBUILD" "$P" "$P/shots"
cmp -s "$Z" "$P/first.zip" && pass "case 2 byte-identical rebuild" || fail "case 2: the rebuild differs"

# Case 3: a PNG the suite no longer produces leaves the zip.
rm "$P/shots/full-app-suite.spec.ts-snapshots/b.png"
"$REBUILD" "$P" "$P/shots"
names=$(unzip -Z1 "$Z" | sort | tr '\n' ' ')
[ "$names" = "a.png " ] && pass "case 3 a gone PNG leaves the zip" || fail "case 3: entries '$names'"

# Case 4: no suite PNG is exit 1 with no zip left behind.
rm "$P/shots/full-app-suite.spec.ts-snapshots/a.png"
"$REBUILD" "$P" "$P/shots" && rc=0 || rc=$?
[ "$rc" -eq 1 ] && [ ! -e "$Z" ] && pass "case 4 nothing to zip" || fail "case 4: exit $rc"

# Case 5: a usage error is exit 2.
"$REBUILD" "$P" 2>/dev/null && rc=0 || rc=$?
[ "$rc" -eq 2 ] && pass "case 5 usage" || fail "case 5: exit $rc"

if [ "$FAILURES" -gt 0 ]; then
  echo "$FAILURES failure(s)" >&2
  exit 1
fi
echo "all rebuild-full-app-suite-zip cases pass"
