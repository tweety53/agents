#!/usr/bin/env bash
# test-lib-flow-guard.sh — assertion harness for scripts/lib/flow-guard.sh,
# the helper every Go-ported guard's shim sources to run flow-guard built
# from its own checkout (KAN-760, design.md's
# guard-binary-built-from-checkout; panel round 0, F1/F2/F8).
#
# Every case runs a real shim through real bash against a real source tree
# and a real `go build` — the defect this helper closes is a binary that is
# not the checkout's source, which only a real build can show. Two kinds of
# tree are used:
#   - this repository itself, reached through its real shims (and through a
#     skills/flow/scripts/ symlink, the route every installed skill takes);
#   - a fixture repository under WORK: a copy of the helper at
#     scripts/lib/flow-guard.sh, a shim beside it, and a tiny stats/ module
#     whose flow-guard prints a version word, so a rebuild after a source
#     edit is visible in the guard's own output.
#
# The build cache is always FLOW_GUARD_CACHE_DIR under WORK: no case writes
# into the operator's real ~/.cache.
#
# No `set -e`: each case is a condition list read by `check` through `$?`,
# and under -e a failing last condition would end the harness silently
# instead of printing its FAIL line.
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO="$(cd "$SCRIPT_DIR/.." && pwd)"
LIB="$SCRIPT_DIR/lib/flow-guard.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

if [ ! -r "$LIB" ]; then
  echo "test-lib-flow-guard: cannot read $LIB" >&2
  exit 2
fi
if ! command -v go >/dev/null 2>&1; then
  echo "test-lib-flow-guard: go is not on PATH — every case needs a real build" >&2
  exit 2
fi

WORK="$(mktemp -d "${TMPDIR:-/tmp}/test-lib-flow-guard.XXXXXX")"
trap 'rm -rf "$WORK"' EXIT
export FLOW_GUARD_CACHE_DIR="$WORK/cache"
NOGO_PATH=/usr/bin:/bin

# run <cmd...> — OUT (stdout+stderr) and RC of the command.
run() {
  OUT="$("$@" 2>&1)"
  RC=$?
}

# check <label> <condition-rc> — pass when the condition held.
check() {
  if [ "$2" -eq 0 ]; then pass "$1"; else fail "$1 — rc=$RC out=$OUT"; fi
}

# new_fixture <main-body-word> — FIX, a fixture repository whose
# scripts/check-x.sh shim runs a flow-guard printing "<word> <guard-name>".
new_fixture() {
  FIX="$(mktemp -d "$WORK/fixture.XXXXXX")"
  mkdir -p "$FIX/scripts/lib" "$FIX/stats/cmd/flow-guard" "$FIX/stats/internal/guard"
  cp "$LIB" "$FIX/scripts/lib/flow-guard.sh"
  cat > "$FIX/scripts/check-x.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
. "${BASH_SOURCE[0]%/*}/lib/flow-guard.sh"
flow_guard_exec check-x 4 "check-x:" "$@"
EOF
  chmod +x "$FIX/scripts/check-x.sh"
  printf 'module fixture\n\ngo 1.21\n' > "$FIX/stats/go.mod"
  : > "$FIX/stats/go.sum"
  printf 'package guard\n' > "$FIX/stats/internal/guard/guard.go"
  printf 'package guard\n' > "$FIX/stats/internal/guard/guard_test.go"
  write_main "$1"
}

# write_main <word> — the fixture's flow-guard prints "<word> <argv[1]>".
write_main() {
  printf 'package main\n\nimport (\n\t"fmt"\n\t"os"\n)\n\nfunc main() { fmt.Println("%s", os.Args[1]) }\n' "$1" \
    > "$FIX/stats/cmd/flow-guard/main.go"
}

# ---------------------------------------------------------------------------
# 1. A stale flow-guard first on PATH never answers for the real shim: the
#    shim builds this checkout's own binary instead (F1/F8).
STALE="$WORK/stale-bin"
mkdir -p "$STALE"
printf '#!/bin/sh\necho STALE-BINARY-ANSWERED\nexit 0\n' > "$STALE/flow-guard"
chmod +x "$STALE/flow-guard"
run env PATH="$STALE:$PATH" "$REPO/scripts/check-task-commit-fields.sh"
[ "$RC" -eq 2 ] && [[ "$OUT" != *STALE-BINARY-ANSWERED* ]] && [[ "$OUT" == *usage* ]]
check "case 1: a flow-guard on PATH is ignored; the checkout-built guard answers" $?

# 2. The binary case 1 built is cached: the next call needs no go at all (F2 —
#    nothing to install, and a cached call is only a hash away).
run env PATH="$NOGO_PATH" "$REPO/scripts/check-task-commit-fields.sh"
[ "$RC" -eq 2 ] && [[ "$OUT" == *usage* ]] && [[ "$OUT" != *"cannot build flow-guard"* ]]
check "case 2: a cached flow-guard answers with no go on PATH" $?

# 3. Reached through a skills/flow/scripts/ symlink, the shim still finds the
#    checkout's stats/ (the lib symlink resolves physically) and the cache.
run env PATH="$NOGO_PATH" "$REPO/skills/flow/scripts/check-task-commit-fields.sh"
[ "$RC" -eq 2 ] && [[ "$OUT" == *usage* ]] && [[ "$OUT" != *"cannot build flow-guard"* ]]
check "case 3: a shim reached through a skill's scripts/ symlink runs the checkout's cached flow-guard" $?

# 4. A source edit rebuilds: the guard answers with the edited source, never
#    the binary cached for the old one.
new_fixture v1
run "$FIX/scripts/check-x.sh"
[ "$RC" -eq 0 ] && [ "$OUT" = "v1 check-x" ]
check "case 4a: the fixture's first call builds and runs its flow-guard" $?
write_main v2
run "$FIX/scripts/check-x.sh"
[ "$RC" -eq 0 ] && [ "$OUT" = "v2 check-x" ]
check "case 4b: after a source edit the rebuilt flow-guard answers, not the cached one" $?

# 5. The key covers exactly the build's sources: a _test.go edit keeps it, an
#    internal/guard or go.mod edit changes it.
# shellcheck source=lib/flow-guard.sh
source "$LIB"
k0="$(flow_guard_key "$FIX/stats")"
printf '// edit\n' >> "$FIX/stats/internal/guard/guard_test.go"
k1="$(flow_guard_key "$FIX/stats")"
printf '// edit\n' >> "$FIX/stats/internal/guard/guard.go"
k2="$(flow_guard_key "$FIX/stats")"
printf '// edit\n' >> "$FIX/stats/go.mod"
k3="$(flow_guard_key "$FIX/stats")"
[ -n "$k0" ] && [ "$k0" = "$k1" ] && [ "$k1" != "$k2" ] && [ "$k2" != "$k3" ]
check "case 5: the cache key ignores _test.go and tracks internal/guard and go.mod (k0=$k0 k1=$k1 k2=$k2 k3=$k3)" $?

# 6. No go and nothing cached: the guard's own cannot-answer code, the cause
#    named, never a verdict. A word no earlier case built, since two trees
#    with the same sources share one cached binary.
new_fixture v6
run env PATH="$NOGO_PATH" "$FIX/scripts/check-x.sh"
[ "$RC" -eq 4 ] && [[ "$OUT" == "check-x: cannot build flow-guard from "*"no go on PATH" ]]
check "case 6: no go and no cached binary exits the cannot-answer code naming the missing go" $?

# 7. A build that fails exits the cannot-answer code with go's output on
#    stderr, and caches nothing.
printf 'package main\n\nfunc main() { undefined() }\n' > "$FIX/stats/cmd/flow-guard/main.go"
OUT="$("$FIX/scripts/check-x.sh" 2>&1 >/dev/null)"
RC=$?
key="$(flow_guard_key "$FIX/stats")"
[ "$RC" -eq 4 ] && [[ "$OUT" == *undefined* ]] && [[ "$OUT" == *"check-x: could not build flow-guard from"* ]] \
  && [ ! -e "$FLOW_GUARD_CACHE_DIR/$key/flow-guard" ]
check "case 7: a failed build exits the cannot-answer code, shows go's error on stderr, caches nothing" $?

# 8. No cache location at all: the cannot-answer code, naming the variables.
run env -i PATH="$PATH" "$FIX/scripts/check-x.sh"
[ "$RC" -eq 4 ] && [[ "$OUT" == *"no cache location"* ]]
check "case 8: with no FLOW_GUARD_CACHE_DIR, XDG_CACHE_HOME or HOME the shim exits the cannot-answer code" $?

# 9. Without a cache override the cache is ${XDG_CACHE_HOME}/flow-guard.
write_main v3
run env -u FLOW_GUARD_CACHE_DIR XDG_CACHE_HOME="$WORK/xdg" "$FIX/scripts/check-x.sh"
key="$(flow_guard_key "$FIX/stats")"
[ "$RC" -eq 0 ] && [ -x "$WORK/xdg/flow-guard/$key/flow-guard" ]
check "case 9: the default cache is \${XDG_CACHE_HOME}/flow-guard/<key>/flow-guard" $?

# 11. Sources edited while the build runs: the binary is discarded rather
#     than cached under a key it was not built from. A `go` wrapper first on
#     PATH edits main.go, then runs the real go.
new_fixture v11
GOWRAP="$WORK/gowrap"
mkdir -p "$GOWRAP"
printf '#!/bin/sh\nprintf "// edited mid-build\\n" >> "%s"\nexec "%s" "$@"\n' \
  "$FIX/stats/cmd/flow-guard/main.go" "$(command -v go)" > "$GOWRAP/go"
chmod +x "$GOWRAP/go"
key="$(flow_guard_key "$FIX/stats")"
run env PATH="$GOWRAP:$PATH" "$FIX/scripts/check-x.sh"
[ "$RC" -eq 4 ] && [[ "$OUT" == *"changed during its build"* ]] && [ ! -e "$FLOW_GUARD_CACHE_DIR/$key/flow-guard" ]
check "case 11: sources edited mid-build exit the cannot-answer code and cache nothing under the old key" $?

# 12. A call from a cross-compiling shell (GOOS/GOARCH/GOFLAGS exported)
#     still builds for the host: it answers, and so does a later plain call,
#     instead of both exec'ing a cached foreign binary (panel round 2,
#     F17/F22).
new_fixture v12
case "$(uname -s)" in Darwin) other=linux ;; *) other=darwin ;; esac
if [ "$(uname -m)" = x86_64 ]; then otherarch=arm64; else otherarch=amd64; fi
run env GOOS="$other" GOARCH="$otherarch" GOFLAGS=-trimpath "$FIX/scripts/check-x.sh"
rc1=$RC out1=$OUT
run env -u GOOS -u GOARCH -u GOFLAGS "$FIX/scripts/check-x.sh"
[ "$rc1" -eq 0 ] && [ "$out1" = "v12 check-x" ] && [ "$RC" -eq 0 ] && [ "$OUT" = "v12 check-x" ]
check "case 12: a cross-compile environment neither breaks its own call nor poisons the cache (first rc=$rc1 out=$out1)" $?

# 13. A failed rename into the cache exits the cannot-answer code with the
#     cause, never set -e's 1 (panel round 2, F21). A `go` wrapper builds,
#     then makes the key directory unwritable so the rename fails.
new_fixture v13
GOWRAP13="$WORK/gowrap13"
mkdir -p "$GOWRAP13"
printf '#!/bin/sh\n"%s" "$@" || exit $?\nwhile [ $# -gt 0 ]; do [ "$1" = -o ] && chmod 555 "$(dirname "$2")"; shift; done\n' \
  "$(command -v go)" > "$GOWRAP13/go"
chmod +x "$GOWRAP13/go"
key="$(flow_guard_key "$FIX/stats")"
run env PATH="$GOWRAP13:$PATH" "$FIX/scripts/check-x.sh"
chmod u+w "$FLOW_GUARD_CACHE_DIR/$key" 2>/dev/null
[ "$RC" -eq 4 ] && [[ "$OUT" == *"check-x: could not install flow-guard into"* ]] && [ ! -e "$FLOW_GUARD_CACHE_DIR/$key/flow-guard" ]
check "case 13: a failed rename exits the cannot-answer code naming the install" $?

# 14. A shim run by bare filename from its own directory still finds its lib
#     (panel round 2, F20). Every real shim, each from its cached binary.
for shim in check-cleanup-complete check-panel-reproducer-exit-contract check-task-commit-fields \
  gather-dispatch-context run-reproducer; do
  run bash -c 'cd "$1" && bash "$2.sh"' _ "$REPO/scripts" "$shim"
  [[ "$OUT" == "$shim: "* || "$OUT" == usage:* ]] && [[ "$OUT" != *"cannot load lib/flow-guard.sh"* ]] \
    && [[ "$OUT" != *"cannot build flow-guard"* ]]
  check "case 14: scripts/$shim.sh run by bare filename loads its lib" $?
done

# 10. A shim with no stats/ beside its checkout exits the cannot-answer code.
rm -rf "$FIX/stats"
run "$FIX/scripts/check-x.sh"
[ "$RC" -eq 4 ] && [[ "$OUT" == *"no stats/ source tree"* ]]
check "case 10: a checkout without stats/ exits the cannot-answer code" $?

# ---------------------------------------------------------------------------
if [ "$FAILURES" -eq 0 ]; then
  printf '\n✓ PASS\n'
  exit 0
fi
printf '\n✗ FAIL — %s failure(s)\n' "$FAILURES" >&2
exit 1
