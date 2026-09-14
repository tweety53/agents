#!/usr/bin/env bash
# Assertion harness for check-model-keys.sh. Builds fixture project roots
# under a sandboxed TMPDIR and asserts the guard's exit status and output.
# Never touches the real repository tree.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-model-keys.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

ROOTS=()
cleanup() {
  [ "${#ROOTS[@]}" -eq 0 ] && return 0
  for r in "${ROOTS[@]}"; do
    rm -rf "$r"
  done
}
trap cleanup EXIT

# new_root -> sets ROOT to a fresh project root directory (no .flow yet).
new_root() {
  ROOT="$(mktemp -d "${TMPDIR:-/tmp}/check-model-keys-test.XXXXXX")"
  ROOTS+=("$ROOT")
}

run_guard() {
  set +e
  OUT="$("$GUARD" "$@" 2>&1)"
  RC=$?
  set -e
}

# run_guard_path <PATH> [args...] — run the shipped guard under an explicit PATH, so
# a case controls exactly which `flow` (if any) the guard resolves.
run_guard_path() {
  local path_override="$1"
  shift
  set +e
  OUT="$(PATH="$path_override" "$GUARD" "$@" 2>&1)"
  RC=$?
  set -e
}

# run_sandbox_guard <PATH> <guard-path> [args...] — run a sandboxed copy of
# the guard (see cases 11-14) under an explicit PATH.
run_sandbox_guard() {
  local path_override="$1" guard="$2"
  shift 2
  set +e
  OUT="$(PATH="$path_override" "$guard" "$@" 2>&1)"
  RC=$?
  set -e
}

# new_sandbox — a directory holding a runnable copy of the guard and its
# lib/, whose SETTINGS_GO resolves inside the sandbox. Cases 6 and 11-14
# write whatever stats/internal/store/settings.go they need beside it.
new_sandbox() {
  SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/check-model-keys-sbx.XXXXXX")"
  ROOTS+=("$SANDBOX")
  mkdir -p "$SANDBOX/scripts"
  cp "$GUARD" "$SANDBOX/scripts/"
  cp -R "$SCRIPT_DIR/lib" "$SANDBOX/scripts/lib"
}

# stub_flow <var-name> <script body> — writes a stub `flow` executable and
# stores its directory in <var-name>. The stub dir joins ROOTS so EXIT
# cleanup removes it; the variable is set by name (printf -v) rather than
# captured, because command substitution would run this function in a
# subshell and lose the ROOTS append.
stub_flow() {
  local var="$1" script="$2" dir
  dir="$(mktemp -d "${TMPDIR:-/tmp}/check-model-keys-stub.XXXXXX")"
  ROOTS+=("$dir")
  printf '#!/usr/bin/env bash\n%s\n' "$script" > "$dir/flow"
  chmod +x "$dir/flow"
  printf -v "$var" '%s' "$dir"
}

# 1. A valid `## self review model` key passes.
new_root
mkdir -p "$ROOT/.flow"
printf '## self review model\n\n`fable`\n' > "$ROOT/.flow/project.md"
run_guard "$ROOT"
[ "$RC" -eq 0 ] && pass "valid self review model key passes" \
  || fail "valid key: rc=$RC out=$OUT"

# 2. An invalid model value fails with a named violation.
new_root
mkdir -p "$ROOT/.flow"
printf '## self review model\n\n`bogus-model`\n' > "$ROOT/.flow/project.md"
run_guard "$ROOT"
[ "$RC" -eq 1 ] && pass "invalid model value fails" \
  || fail "invalid value: rc=$RC out=$OUT"
case "$OUT" in
  *"bogus-model"*"is not a ValidModels member"*) pass "invalid value is named" ;;
  *) fail "invalid value not named: out=$OUT" ;;
esac

# 3. A project root with no .flow/project.md at all is a pass, not a
# violation — absence is a supported, valid state.
new_root
run_guard "$ROOT"
[ "$RC" -eq 0 ] && pass "missing .flow/project.md passes" \
  || fail "missing project.md: rc=$RC out=$OUT"

# 4. F5 regression: every project root examined counts toward the summary,
# including one with no .flow/project.md. Mixing one root that has a valid
# project.md with one that has none must report "2 project(s) checked", not
# undercount the one with no file.
new_root
NO_FILE_ROOT="$ROOT"
new_root
mkdir -p "$ROOT/.flow"
printf '## self review model\n\n`fable`\n' > "$ROOT/.flow/project.md"
WITH_FILE_ROOT="$ROOT"
run_guard "$NO_FILE_ROOT" "$WITH_FILE_ROOT"
[ "$RC" -eq 0 ] && pass "mixed roots exit 0" || fail "mixed roots: rc=$RC out=$OUT"
case "$OUT" in
  *"2 project(s) checked"*) pass "a root with no .flow/project.md still counts toward the summary" ;;
  *) fail "summary undercounts: out=$OUT" ;;
esac

# 5. A non-directory project root exits 2.
run_guard "${TMPDIR:-/tmp}/check-model-keys-does-not-exist-$$"
[ "$RC" -eq 2 ] && pass "a nonexistent project root exits 2" \
  || fail "nonexistent root: rc=$RC out=$OUT"

# 6. The guard ASKS `flow settings models` first and uses its answer where
# no source exists to cross-check against: a sandbox copy of the guard with
# no stats/ tree at all, and a stub CLI whose set is {alpha, bravo}, makes
# `bravo` — a name in no settings.go anywhere — valid. Fails while the
# guard still regexes Go source only.
STUB=""
stub_flow STUB $'[ "$1" = settings ] && [ "$2" = models ] || { echo "stub: unexpected args: $*" >&2; exit 3; }\nprintf \'alpha\\nbravo\\n\''
new_sandbox
new_root
mkdir -p "$ROOT/.flow"
printf '## self review model\n\n`bravo`\n' > "$ROOT/.flow/project.md"
run_sandbox_guard "$STUB:/usr/bin:/bin" "$SANDBOX/scripts/check-model-keys.sh" "$ROOT"
[ "$RC" -eq 0 ] && pass "with no source, the CLI's set governs" \
  || fail "no-source CLI governing: rc=$RC out=$OUT"

# 7. And on drift the source set governs: a stub set {sonnet, gamma}
# disagrees with the real settings.go, so `gamma` — a CLI-only name the
# checkout does not declare — is a violation. This is the operator-chosen
# round-2 governance, the inverse of the drift direction F1 pinned.
STUB=""
stub_flow STUB $'[ "$1" = settings ] && [ "$2" = models ] || { echo "stub: unexpected args: $*" >&2; exit 3; }\nprintf \'sonnet\\ngamma\\n\''
new_root
mkdir -p "$ROOT/.flow"
printf '## self review model\n\n`gamma`\n' > "$ROOT/.flow/project.md"
run_guard_path "$STUB:/usr/bin:/bin" "$ROOT"
[ "$RC" -eq 1 ] && pass "on drift a CLI-only name fails despite the CLI" \
  || fail "drift governance: rc=$RC out=$OUT"
case "$OUT" in
  *"gamma"*"is not a ValidModels member"*) pass "the CLI-only value is named" ;;
  *) fail "CLI-only value not named: out=$OUT" ;;
esac

# 8. A `flow` that cannot answer (non-zero exit) falls back to the
# settings.go parse: `fable` is valid again through the real source.
STUB=""
stub_flow STUB 'echo "stub: store down" >&2; exit 1'
new_root
mkdir -p "$ROOT/.flow"
printf '## self review model\n\n`fable`\n' > "$ROOT/.flow/project.md"
run_guard_path "$STUB:/usr/bin:/bin" "$ROOT"
[ "$RC" -eq 0 ] && pass "a failing CLI falls back to the source parse" \
  || fail "failing CLI fallback: rc=$RC out=$OUT"
case "$OUT" in
  *"no answer from flow settings models"*"stub: store down"*)
    pass "the fallback announcement carries the CLI's own diagnostic" ;;
  *) fail "diagnostic missing from fallback line: out=$OUT" ;;
esac

# 9. No `flow` on PATH at all falls back the same way.
new_root
mkdir -p "$ROOT/.flow"
printf '## self review model\n\n`fable`\n' > "$ROOT/.flow/project.md"
run_guard_path "/usr/bin:/bin" "$ROOT"
[ "$RC" -eq 0 ] && pass "no flow on PATH falls back to the source parse" \
  || fail "no-flow fallback: rc=$RC out=$OUT"

# 10. A CLI that answers exit 0 with nothing falls back too — an empty
# answer is not an empty valid set.
STUB=""
stub_flow STUB 'exit 0'
new_root
mkdir -p "$ROOT/.flow"
printf '## self review model\n\n`fable`\n' > "$ROOT/.flow/project.md"
run_guard_path "$STUB:/usr/bin:/bin" "$ROOT"
[ "$RC" -eq 0 ] && pass "an empty CLI answer falls back to the source parse" \
  || fail "empty-answer fallback: rc=$RC out=$OUT"

# 11. The redefined exit-2 contract's first branch, pinned: no CLI on PATH
# and no settings.go at all inside the sandbox copy's own tree — exit 2,
# naming the unreadable file.
new_sandbox
run_sandbox_guard "/usr/bin:/bin" "$SANDBOX/scripts/check-model-keys.sh" "$SANDBOX"
[ "$RC" -eq 2 ] && pass "no CLI and no settings.go exits 2" \
  || fail "missing-source branch: rc=$RC out=$OUT"
case "$OUT" in
  *"cannot read"*settings.go*) pass "the missing-source refusal is named" ;;
  *) fail "missing-source not named: out=$OUT" ;;
esac

# 12. The second branch: settings.go readable but with no ValidModels map
# — exit 2, naming the empty parse.
mkdir -p "$SANDBOX/stats/internal/store"
printf 'package store\n\nvar Other = map[string]bool{\n\t"x": true,\n}\n' \
  > "$SANDBOX/stats/internal/store/settings.go"
run_sandbox_guard "/usr/bin:/bin" "$SANDBOX/scripts/check-model-keys.sh" "$SANDBOX"
[ "$RC" -eq 2 ] && pass "a memberless settings.go exits 2" \
  || fail "memberless-source branch: rc=$RC out=$OUT"
case "$OUT" in
  *"no ValidModels members found"*) pass "the empty-parse refusal is named" ;;
  *) fail "empty-parse not named: out=$OUT" ;;
esac

# 13. The cross-check announces a stale install: a CLI whose set is a
# subset of the real source's answers drifts, so the source set governs
# the verdict — `sonnet` sits in both sets, which is why the key still
# validates; what the case pins is the announcement.
STUB=""
stub_flow STUB $'[ "$1" = settings ] && [ "$2" = models ] && printf \'sonnet\\n\'\nexit 0'
new_root
mkdir -p "$ROOT/.flow"
printf '## self review model\n\n`sonnet`\n' > "$ROOT/.flow/project.md"
run_guard_path "$STUB:/usr/bin:/bin" "$ROOT"
[ "$RC" -eq 0 ] && pass "a subset CLI answer still validates its own members" \
  || fail "drift case verdict: rc=$RC out=$OUT"
case "$OUT" in
  *"disagrees with"*settings.go*) pass "the stale-install drift is announced" ;;
  *) fail "drift not announced: out=$OUT" ;;
esac

# 14. A readable source that parses empty while the CLI answers turns the
# cross-check off — announced, not silent (the map literal's shape may
# have changed). Self-contained: builds its own sandbox and its own
# memberless settings.go rather than reusing case 12's fixture.
new_sandbox
mkdir -p "$SANDBOX/stats/internal/store"
printf 'package store\n\nvar Other = map[string]bool{\n\t"x": true,\n}\n' \
  > "$SANDBOX/stats/internal/store/settings.go"
new_root
mkdir -p "$ROOT/.flow"
printf '## self review model\n\n`sonnet`\n' > "$ROOT/.flow/project.md"
STUB2=""
stub_flow STUB2 $'[ "$1" = settings ] && [ "$2" = models ] && printf \'sonnet\\n\'\nexit 0'
run_sandbox_guard "$STUB2:/usr/bin:/bin" "$SANDBOX/scripts/check-model-keys.sh" "$SANDBOX"
[ "$RC" -eq 0 ] && pass "a shape-changed source keeps the CLI verdict" \
  || fail "shape-drift case verdict: rc=$RC out=$OUT"
case "$OUT" in
  *"cross-check is off"*) pass "the disabled cross-check is announced" ;;
  *) fail "cross-check-off not announced: out=$OUT" ;;
esac

# 15. The F1 flip, pinned in the harness: a stale-but-working install whose
# set LACKS a member the checkout declares cannot fail that member — drift
# hands the verdict to the source set, so `sonnet` validates against a CLI
# that has never heard of it, and the drift is still announced.
STUB=""
stub_flow STUB $'[ "$1" = settings ] && [ "$2" = models ] && printf \'fable\\nhaiku\\nopus\\n\'\nexit 0'
new_root
mkdir -p "$ROOT/.flow"
printf '## self review model\n\n`sonnet`\n' > "$ROOT/.flow/project.md"
run_guard_path "$STUB:/usr/bin:/bin" "$ROOT"
[ "$RC" -eq 0 ] && pass "a source member the stale CLI lacks still validates" \
  || fail "drift source-governs: rc=$RC out=$OUT"
case "$OUT" in
  *"disagrees with"*settings.go*) pass "the stale-install drift is announced beside the flip" ;;
  *) fail "drift not announced beside the flip: out=$OUT" ;;
esac

# 16. Duplicate names in the CLI's answer are not drift: comm compares
# SETS, so a CLI printing one member twice alongside the rest of the real
# set still agrees with a source declaring each member once, and no
# announcement fires.
STUB=""
stub_flow STUB $'[ "$1" = settings ] && [ "$2" = models ] && printf \'sonnet\\nsonnet\\nfable\\nhaiku\\nopus\\n\'\nexit 0'
new_root
mkdir -p "$ROOT/.flow"
printf '## self review model\n\n`sonnet`\n' > "$ROOT/.flow/project.md"
run_guard_path "$STUB:/usr/bin:/bin" "$ROOT"
[ "$RC" -eq 0 ] && pass "a duplicated CLI member is not drift" \
  || fail "duplicate-member case: rc=$RC out=$OUT"
case "$OUT" in
  *"disagrees with"*) fail "false drift announced on duplicate member: out=$OUT" ;;
  *) pass "no drift announced for a duplicated member" ;;
esac

# 17. A CLI that answers names AND writes to stderr has its diagnostic
# relayed — the answer governs, the warning is not swallowed.
STUB=""
stub_flow STUB $'echo "stub: deprecation warning" >&2\n[ "$1" = settings ] && [ "$2" = models ] && printf \'fable\\n\'\nexit 0'
new_root
mkdir -p "$ROOT/.flow"
printf '## self review model\n\n`fable`\n' > "$ROOT/.flow/project.md"
run_guard_path "$STUB:/usr/bin:/bin" "$ROOT"
[ "$RC" -eq 0 ] && pass "an answering CLI with stderr still validates" \
  || fail "warn-and-answer case: rc=$RC out=$OUT"
case "$OUT" in
  *"flow settings models reported:"*"stub: deprecation warning"*)
    pass "the answering CLI's diagnostic is relayed" ;;
  *) fail "answering CLI diagnostic swallowed: out=$OUT" ;;
esac

# 18. The diagnostic appears exactly ONCE per stream: on a failed ask the
# fallback announcement carries it and the relay stays silent, so the
# operator reading a stale-install debug never sees the same line twice.
STUB=""
stub_flow STUB 'echo "stub: store down" >&2; exit 1'
new_root
mkdir -p "$ROOT/.flow"
printf '## self review model\n\n`fable`\n' > "$ROOT/.flow/project.md"
run_guard_path "$STUB:/usr/bin:/bin" "$ROOT"
[ "$RC" -eq 0 ] && pass "a failed ask still falls back cleanly" \
  || fail "single-emission case verdict: rc=$RC out=$OUT"
N="$(printf '%s\n' "$OUT" | grep -c 'stub: store down')"
[ "$N" -eq 1 ] && pass "the failed ask's diagnostic is emitted exactly once" \
  || fail "diagnostic emitted $N times, expected 1: out=$OUT"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-model-keys: all cases pass\n'
