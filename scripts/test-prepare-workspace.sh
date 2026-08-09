#!/usr/bin/env bash
# Assertion harness for prepare-workspace.sh. Builds sandboxed git repositories
# under TMPDIR for every case that needs a declared `## workspace isolation`
# section; never touches the real repository tree for those. The no-op case is
# the one exception — it runs the script against this repository itself, since
# the thing under test there is exactly the same no-op check-workspace-
# isolation.sh already keeps: this repository declares no section.
#
# Three cases, matching the three the task record asks for: (1) no `##
# workspace isolation` section — no-op, exit 0, nothing printed; (2) a section
# declared — variables exported and printed correctly; (3)
# check-workspace-isolation.sh failing on a malformed row — stops the script
# before any export, non-zero exit.
#
# Bash 3.2 is the floor, as test-check-finish-preflight.sh's header records for
# this repository: indexed arrays only, no associative arrays, and this
# harness uses none.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SCRIPT="$SCRIPT_DIR/prepare-workspace.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# An indexed array, not a space-separated string: mktemp paths under TMPDIR may
# contain spaces, and word-splitting a string would leak a sandbox whose path
# split and `rm -rf` the fragments.
REPOS=()
cleanup() {
  [ "${#REPOS[@]}" -eq 0 ] && return 0
  for repo in "${REPOS[@]}"; do
    rm -rf "$repo"
  done
}
trap cleanup EXIT

# new_repo <change-name> -> sets REPO to a freshly initialized git repository,
# checked out on `openspec/<change-name>` — the branch every apply worktree is
# created on, per section 2 of skills/myflow-do/SKILL.md, and the one place
# prepare-workspace.sh reads the change name from.
new_repo() {
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/prepare-workspace-test.XXXXXX")"
  REPOS+=("$REPO")
  git -C "$REPO" init -q
  git -C "$REPO" config user.email "test@example.com"
  git -C "$REPO" config user.name "Test"
  printf 'seed\n' > "$REPO/README.md"
  git -C "$REPO" add -A
  git -C "$REPO" commit -q -m "seed"
  git -C "$REPO" checkout -q -b "openspec/$1"
}

write_config() {
  mkdir -p "$REPO/.myflow"
  printf '%s\n' "$1" > "$REPO/.myflow/project.md"
  git -C "$REPO" add -A
  git -C "$REPO" commit -q -m "configure"
}

run_script() {
  set +e
  OUT="$("$SCRIPT" "$@" 2>"$ERRFILE")"
  RC=$?
  set -e
  ERR="$(cat "$ERRFILE")"
}
ERRFILE="$(mktemp "${TMPDIR:-/tmp}/prepare-workspace-test-err.XXXXXX")"
REPOS+=("$ERRFILE")

# ===========================================================================
# Case 1: No `## workspace isolation` section — no-op, exit 0, nothing printed.
# ===========================================================================

# 1a. This repository itself: it declares no section, exactly mirroring
#     check-workspace-isolation.sh's own no-op case for a project with none.
run_script "$REPO_ROOT"
[ "$RC" -eq 0 ] && pass "1a: this repository exits 0" \
  || fail "1a: expected exit 0, got rc=$RC out=$OUT err=$ERR"
[ -z "$OUT" ] && pass "1a: this repository prints nothing" \
  || fail "1a: expected no stdout, got: $OUT"

# 1b. A fixture project.md with sections but none named `## workspace
#     isolation`.
new_repo "kan-1-no-isolation"
write_config "# fixture

## test

\`\`\`bash
scripts/test-setup.sh
\`\`\`
"
run_script "$REPO"
[ "$RC" -eq 0 ] && pass "1b: no isolation section exits 0" \
  || fail "1b: expected exit 0, got rc=$RC out=$OUT err=$ERR"
[ -z "$OUT" ] && pass "1b: no isolation section prints nothing" \
  || fail "1b: expected no stdout, got: $OUT"

# 1c. No `.myflow/project.md` at all.
new_repo "kan-2-no-config"
run_script "$REPO"
[ "$RC" -eq 0 ] && pass "1c: no .myflow/project.md exits 0" \
  || fail "1c: expected exit 0, got rc=$RC out=$OUT err=$ERR"
[ -z "$OUT" ] && pass "1c: no .myflow/project.md prints nothing" \
  || fail "1c: expected no stdout, got: $OUT"

# ===========================================================================
# Case 2: A section declared — variables exported and printed correctly.
#
# The change name is kan-15-parallel-myflow-do-task-lanes, the exact worked
# example under "The workspace id" (skills/myflow-contracts/workspace-
# isolation.md): id kan-15-55a6, id_underscored kan_15_55a6, digest 55a6, and
# offset 3270 (the file's own worked value for that digest) — so every
# expected value below is asserted against that canonical worked example
# rather than against this script's own arithmetic.
# ===========================================================================
new_repo "kan-15-parallel-myflow-do-task-lanes"
write_config "## workspace isolation

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| \`database\` | \`DB_URL\` | \`jdbc:postgresql://localhost:5432/appdb\` | \`jdbc:postgresql://localhost:5432/appdb_<id_underscored>\` |
| \`bucket\` | \`MEDIA_BUCKET\` | \`appdb-media\` | \`appdb-media-<id>\` |
| \`port\` | \`API_PORT\` | \`8080\` | \`+<offset>\` |
| \`url\` | \`MEDIA_BASE_URL\` | \`http://localhost:9000/appdb-media\` | \`http://localhost:9000/<value:MEDIA_BUCKET>\` |
| \`url\` | \`WEB_URL\` | \`http://localhost:8080\` | \`http://localhost:<value:API_PORT>\` |
| \`cache index\` | \`CACHE_INDEX\` | \`0\` | \`probed\` |

| Command | Runs |
|---------|------|
| \`create\` | \`./scripts/workspace create\` |
| \`remove\` | \`./scripts/workspace remove\` |
| \`survivors\` | \`./scripts/workspace survivors\` |
"
run_script "$REPO"
[ "$RC" -eq 0 ] && pass "2: a declared section exits 0" \
  || fail "2: expected exit 0, got rc=$RC out=$OUT err=$ERR"

case "$OUT" in
  *"DB_URL=jdbc:postgresql://localhost:5432/appdb_kan_15_55a6"*) pass "2: database row substitutes <id_underscored>" ;;
  *) fail "2: DB_URL not resolved as expected, got: $OUT" ;;
esac
case "$OUT" in
  *"MEDIA_BUCKET=appdb-media-kan-15-55a6"*) pass "2: bucket row substitutes <id>" ;;
  *) fail "2: MEDIA_BUCKET not resolved as expected, got: $OUT" ;;
esac
case "$OUT" in
  *"API_PORT=11350"*) pass "2: port row adds the offset (8080 + 3270)" ;;
  *) fail "2: API_PORT not resolved as expected, got: $OUT" ;;
esac
case "$OUT" in
  *"MEDIA_BASE_URL=http://localhost:9000/appdb-media-kan-15-55a6"*) pass "2: url row resolves <value:MEDIA_BUCKET>" ;;
  *) fail "2: MEDIA_BASE_URL not resolved as expected, got: $OUT" ;;
esac
case "$OUT" in
  *"WEB_URL=http://localhost:11350"*) pass "2: url row resolves <value:API_PORT>" ;;
  *) fail "2: WEB_URL not resolved as expected, got: $OUT" ;;
esac
LINES="$(printf '%s\n' "$OUT" | wc -l | tr -d ' ')"
[ "$LINES" -eq 5 ] && pass "2: exactly one KEY=value line per declared row" \
  || fail "2: expected 5 lines, got $LINES: $OUT"

case "$OUT" in
  *"CACHE_INDEX"*) fail "2: the cache index row leaked a KEY=value line onto stdout: $OUT" ;;
  *) pass "2: the cache index row is not among the exported KEY=value lines" ;;
esac
case "$ERR" in
  *"CACHE_INDEX"*"cache index"*) pass "2: the cache index row is reported by name on stderr" ;;
  *) fail "2: expected the cache index row reported on stderr, got: $ERR" ;;
esac

# ===========================================================================
# Case 3: check-workspace-isolation.sh failing on a malformed row — stops before
#    any export, non-zero exit.
# ===========================================================================
new_repo "kan-3-malformed"
write_config "## workspace isolation

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| \`port\` | \`API_PORT\` | \`notanumber\` | \`+<offset>\` |

| Command | Runs |
|---------|------|
| \`create\` | \`./scripts/workspace create\` |
| \`remove\` | \`./scripts/workspace remove\` |
| \`survivors\` | \`./scripts/workspace survivors\` |
"
run_script "$REPO"
[ "$RC" -ne 0 ] && pass "3: a malformed row is a non-zero exit" \
  || fail "3: expected a non-zero exit, got rc=0 out=$OUT"
case "$OUT" in
  *"API_PORT"*"not a bare integer"*) pass "3: the guard's own violation is relayed" ;;
  *) fail "3: expected the guard's violation on stdout, got: $OUT" ;;
esac
case "$OUT" in
  *"="*) fail "3: a KEY=value line reached stdout despite the malformed row: $OUT" ;;
  *) pass "3: no KEY=value line reached stdout" ;;
esac

# ===========================================================================
# Case 4: a `port` row's `Default` carries a leading zero. Bash arithmetic
# reads a leading-zero literal as octal unless forced to base 10, so this
# guards the fix rather than just the happy path: `0070 + 20` read as octal is
# `76`, not `90`, and `0080` is not even valid octal, which crashes the whole
# script rather than exporting a wrong value.
# ===========================================================================
new_repo "kan-4-leading-zero-port"
write_config "## workspace isolation

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| \`port\` | \`API_PORT\` | \`0070\` | \`+<offset>\` |

| Command | Runs |
|---------|------|
| \`create\` | \`./scripts/workspace create\` |
| \`remove\` | \`./scripts/workspace remove\` |
| \`survivors\` | \`./scripts/workspace survivors\` |
"
run_script "$REPO"
[ "$RC" -eq 0 ] && pass "4a: a leading-zero port default exits 0" \
  || fail "4a: expected exit 0, got rc=$RC out=$OUT err=$ERR"
case "$OUT" in
  *"API_PORT="*) pass "4a: API_PORT is exported" ;;
  *) fail "4a: expected API_PORT in output, got: $OUT" ;;
esac
case "$OUT" in
  *"API_PORT=76"*) fail "4a: 0070 was read as octal (76) instead of decimal (90): $OUT" ;;
  *) pass "4a: 0070 is not read as octal" ;;
esac

# 4b. A leading-zero default that is not even valid octal (`8` and `9` are not
#     octal digits) — this must not crash the script.
new_repo "kan-5-invalid-octal-port"
write_config "## workspace isolation

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| \`port\` | \`API_PORT\` | \`0080\` | \`+<offset>\` |

| Command | Runs |
|---------|------|
| \`create\` | \`./scripts/workspace create\` |
| \`remove\` | \`./scripts/workspace remove\` |
| \`survivors\` | \`./scripts/workspace survivors\` |
"
run_script "$REPO"
[ "$RC" -eq 0 ] && pass "4b: an invalid-octal port default does not crash" \
  || fail "4b: expected exit 0, got rc=$RC out=$OUT err=$ERR"

# ===========================================================================
# Case 5: check-workspace-isolation.sh present but not executable — a defined
# failure (exit 2, a clear message) rather than a raw exec "Permission denied"
# (exit 126) escaping this script's stated 0/1/2 contract.
# ===========================================================================
new_repo "kan-6-guard-not-executable"
write_config "## workspace isolation

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| \`port\` | \`API_PORT\` | \`8080\` | \`+<offset>\` |

| Command | Runs |
|---------|------|
| \`create\` | \`./scripts/workspace create\` |
| \`remove\` | \`./scripts/workspace remove\` |
| \`survivors\` | \`./scripts/workspace survivors\` |
"
mkdir -p "$REPO/nonexec-scripts"
cp "$SCRIPT_DIR/check-workspace-isolation.sh" "$REPO/nonexec-scripts/check-workspace-isolation.sh"
chmod -x "$REPO/nonexec-scripts/check-workspace-isolation.sh"
cp "$SCRIPT" "$REPO/nonexec-scripts/prepare-workspace.sh"
set +e
OUT="$("$REPO/nonexec-scripts/prepare-workspace.sh" "$REPO" 2>"$ERRFILE")"
RC=$?
set -e
ERR="$(cat "$ERRFILE")"
[ "$RC" -eq 2 ] && pass "5: a non-executable guard exits 2" \
  || fail "5: expected exit 2, got rc=$RC out=$OUT err=$ERR"
case "$ERR" in
  *"cannot find a runnable check-workspace-isolation.sh"*) pass "5: a clear message names the problem" ;;
  *) fail "5: expected a clear message on stderr, got: $ERR" ;;
esac

# ===========================================================================
# Case 6: a `url` row's `<value:...>` reference resolves — by variable name —
# to a row this script never gives a WSVAL: a `cache index` row. A real
# check-workspace-isolation.sh already refuses a file shaped this way (a
# `<value:...>` reference may only name a `database`, `bucket` or `port` row),
# so this exercises prepare-workspace.sh's OWN defense of that same rule,
# independent of the guard — a stub guard stands in for it, reporting success
# with exactly the `#ROW` lines a buggy or future guard might let through, so
# this case still fails loudly rather than silently exporting the empty
# string that a `cache index` row's WSVAL starts as and is never given.
# ===========================================================================
new_repo "kan-7-value-ref-wrong-kind"
mkdir -p "$REPO/.myflow" "$REPO/fake-scripts"
printf '%s\n' "## workspace isolation

(fixture only — the stub guard below supplies the rows this script reads)
" > "$REPO/.myflow/project.md"
git -C "$REPO" add -A
git -C "$REPO" commit -q -m "configure"

cp "$SCRIPT" "$REPO/fake-scripts/prepare-workspace.sh"
cat > "$REPO/fake-scripts/check-workspace-isolation.sh" <<'STUB'
#!/usr/bin/env bash
# Stub standing in for check-workspace-isolation.sh: reports success
# unconditionally, with `#ROW` lines naming a `url` row whose `<value:...>`
# resolves to a `cache index` row — the shape a real guard already refuses,
# used here to exercise prepare-workspace.sh's own independent defense of
# the same rule.
printf 'ISOLATION-OK: stub — 2 resource row(s) and 3 command row(s) validated\n'
printf '#ROW\tcache index\tCACHE_INDEX\t0\tprobed\n'
printf '#ROW\turl\tWEB_URL\thttp://localhost/0\thttp://localhost/<value:CACHE_INDEX>\n'
exit 0
STUB
chmod +x "$REPO/fake-scripts/check-workspace-isolation.sh"

set +e
OUT="$("$REPO/fake-scripts/prepare-workspace.sh" "$REPO" 2>"$ERRFILE")"
RC=$?
set -e
ERR="$(cat "$ERRFILE")"
[ "$RC" -ne 0 ] && pass "6: a url row referencing a cache index row is a non-zero exit" \
  || fail "6: expected a non-zero exit, got rc=0 out=$OUT"
case "$OUT" in
  *"CACHE_INDEX"*) fail "6: an empty/stale substitution reached stdout: $OUT" ;;
  *) pass "6: no substituted value reached stdout" ;;
esac
case "$ERR" in
  *"cache index"*) pass "6: the wrong-kind reference is reported by name on stderr" ;;
  *) fail "6: expected the wrong-kind reference reported on stderr, got: $ERR" ;;
esac

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'prepare-workspace: all cases pass\n'
