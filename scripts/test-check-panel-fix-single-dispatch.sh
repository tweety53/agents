#!/usr/bin/env bash
# Assertion harness for check-panel-fix-single-dispatch.sh.
#
# Every case builds a worktree-shaped sandbox and a stub `flow` binary placed
# ahead of the real one on PATH inside that sandbox: the stub answers
# `record dispatches -change <name> [-C <dir>]` and -- since kan-499's
# chunked contract -- `record findings -change <name> [-C <dir>]` with canned
# JSON arrays (the shape the real verbs print), or exits non-zero to simulate
# a store the guard could not reach. Follows
# test-check-panel-findings-closed.sh's own shape for these helpers.
#
# `-e` as well as `-u`/`pipefail`, matching this repository's sibling
# harnesses: without `-e`, a failed `mktemp` would leave a stale sandbox path
# in play and the harness would keep going anyway.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-panel-fix-single-dispatch.sh"
FAILED=0

WORKTREES=()
cleanup() {
  [ "${#WORKTREES[@]}" -eq 0 ] && return 0
  for wt in "${WORKTREES[@]}"; do
    rm -rf "$wt"
  done
}
trap cleanup EXIT

# dispatch_json <key> <role> <token> [<key> <role> <token> ...] -- builds a
# compact JSON array of dispatch-row objects from key/role/token triples,
# via `jq -n --args` rather than hand-quoted string interpolation.
dispatch_json() {
  jq -nc '
    [$ARGS.positional as $a
     | range(0; ($a | length) / 3)
     | {key: $a[. * 3], role: $a[. * 3 + 1], sessionToken: $a[. * 3 + 2]}]
  ' --args -- "$@"
}

# make_worktree_json <json-array> [<findings-json-array>] -- a worktree-shaped
# sandbox carrying a stub `flow` on its own bin/, which prints <json-array> on
# stdout for `record dispatches` and, since kan-499's chunked contract, the
# optional second array (default `[]`) for `record findings`; any other verb
# exits 2. Prints the worktree path.
make_worktree_json() {
  local wt json="$1" findings="${2:-[]}"
  wt="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-fix-single-dispatch-test.XXXXXX")" || {
    printf 'make_worktree_json: mktemp failed -- aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  WORKTREES+=("$wt")
  mkdir -p "$wt/bin"
  printf '%s' "$json" > "$wt/bin/dispatches.json"
  printf '%s' "$findings" > "$wt/bin/findings.json"
  cat > "$wt/bin/flow" <<'STUB'
#!/usr/bin/env bash
dir="$(dirname -- "$0")"
case "${2:-}" in
  dispatches) cat "$dir/dispatches.json" ;;
  findings)   cat "$dir/findings.json" ;;
  *) printf 'stub flow: unexpected invocation: %s\n' "$*" >&2; exit 2 ;;
esac
exit 0
STUB
  chmod +x "$wt/bin/flow"
  printf '%s' "$wt"
}

# findings_json <count> <round> -- a canned findings array of <count> fixed
# rows all raised in <round>, the shape stats/internal/records' Finding prints.
findings_json() {
  jq -nc --argjson n "$1" --argjson r "$2" \
    '[range(0; $n) | {ref: ("F" + (. + 1 | tostring)), round: $r, status: "fixed"}]'
}

# make_worktree_store_unreachable -- a worktree-shaped sandbox whose stub
# `flow` exits non-zero and prints nothing useful to stdout, simulating a
# store `record dispatches` could not reach.
make_worktree_store_unreachable() {
  local wt
  wt="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-fix-single-dispatch-test.XXXXXX")" || {
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

# make_worktree_findings_unreachable -- a sandbox whose stub answers
# `record dispatches` from a canned `[]` but exits non-zero for
# `record findings`, simulating kan-499's rule: a findings read the guard
# cannot establish is "cannot answer" (exit 2), never a clean verdict.
make_worktree_findings_unreachable() {
  local wt
  wt="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-fix-single-dispatch-test.XXXXXX")" || {
    printf 'make_worktree_findings_unreachable: mktemp failed -- aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  WORKTREES+=("$wt")
  mkdir -p "$wt/bin"
  printf '%s' '[]' > "$wt/bin/dispatches.json"
  cat > "$wt/bin/flow" <<'STUB'
#!/usr/bin/env bash
case "${2:-}" in
  dispatches) echo '[]' ;;
  findings)   echo "flow: connect: connection refused" >&2; exit 1 ;;
  *)          exit 2 ;;
esac
exit 0
STUB
  chmod +x "$wt/bin/flow"
  printf '%s' "$wt"
}

run_guard() {
  local wt="$1" name="${2:-demo}" token="${3-mf-tok}"   # ${3-...}, not ${3:-...}: an
  # explicitly empty token must reach the guard, which is what case 7a tests.
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
  errfile="$(mktemp "${TMPDIR:-/tmp}/check-panel-fix-single-dispatch-test-stderr.XXXXXX")"
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
# 1. One canonical panel-fix row in each of two rounds -- exit 0, no stderr.
# ===========================================================================
wt="$(make_worktree_json "$(dispatch_json panel-fix-1 panel-fix mf-tok panel-fix-2 panel-fix mf-tok)")"
expect_exit_and_no_stderr 'case 1: one canonical dispatch per round exits 0 with no stderr' 0 run_guard "$wt"

# ===========================================================================
# 2. Original plus -retry in one round -- the handshake's one allowed
#    second dispatch -- exit 0.
# ===========================================================================
wt="$(make_worktree_json "$(dispatch_json panel-fix-1 panel-fix mf-tok panel-fix-1-retry panel-fix mf-tok)")"
expect_exit 'case 2: original plus -retry exits 0' 0 run_guard "$wt"

# ===========================================================================
# 3. Two non-retry rows in one round -- the KAN-482 shape, one dispatch per
#    reviewer -- exit 1, naming panel-fix-1.
#
#    PROVED BY MUTATION: two originals in one round are caught by TWO
#    rules -- the per-round original-count branch and the
#    total==2-without-retry clause -- so flipping this case takes both:
#    with `elif [ "$orig" -gt 1 ]; then` -> `elif false; then` AND
#    `[ "$retry" -ne 1 ]` -> `false` in a scratch copy, this case exits 0
#    (confirmed by hand before this comment was written); either mutation
#    alone leaves it caught by the other rule.
# ===========================================================================
wt="$(make_worktree_json "$(dispatch_json panel-fix-1 panel-fix mf-tok panel-fix-1 panel-fix mf-tok)")"
expect_exit_and_names 'case 3: two originals in one round exit 1, naming the round' 1 'panel-fix-1' run_guard "$wt"

# ===========================================================================
# 4. Invented key panel-fix-f1 -- one dispatch per finding under keys no
#    contract declared -- exit 1, naming the key verbatim.
# ===========================================================================
wt="$(make_worktree_json "$(dispatch_json panel-fix-f1 panel-fix mf-tok)")"
expect_exit_and_names 'case 4: invented key exits 1, naming it verbatim' 1 'panel-fix-f1' run_guard "$wt"

# ===========================================================================
# 5. A -retry row with no original -- a retry is never a round's only
#    dispatch -- exit 1, naming panel-fix-1.
# ===========================================================================
wt="$(make_worktree_json "$(dispatch_json panel-fix-1-retry panel-fix mf-tok)")"
expect_exit_and_names 'case 5: retry-only round exits 1, naming the round' 1 'panel-fix-1' run_guard "$wt"

# ===========================================================================
# 6. Other roles and this token's clean rounds never count, and the foreign
#    token's rows DO: a foreign-token panel-fix row carries two originals,
#    so a guard that drops its session-token scoping counts them and exits 1
#    (mutant M15, kan-482 pass 1) -- while the correct guard ignores them
#    and exits 0. Implementer and reviewer rows of this token are present
#    and must never count either.
# ===========================================================================
wt="$(make_worktree_json "$(dispatch_json task-1-implementer implementer mf-tok slot-primary reviewer mf-tok panel-fix-1 panel-fix mf-tok panel-fix-1 panel-fix mf-other panel-fix-1 panel-fix mf-other)")"
expect_exit 'case 6: other roles ignored, foreign token rows ignored' 0 run_guard "$wt"

# ===========================================================================
# 7. Cannot answer -- exit 2 for every one: missing session-token argument,
#    non-directory worktree, store unreachable, bad change name.
# ===========================================================================
wt="$(make_worktree_json "$(dispatch_json panel-fix-1 panel-fix mf-tok)")"
expect_exit 'case 7a: missing session-token argument exits 2' 2 run_guard "$wt" demo ""
expect_exit 'case 7b: non-directory worktree exits 2' 2 run_guard "$wt/no-such-dir"
wt="$(make_worktree_store_unreachable)"
expect_exit 'case 7c: store unreachable exits 2' 2 run_guard "$wt"
expect_exit 'case 7d: change name outside the allowlist exits 2' 2 run_guard "$(make_worktree_json '[]')" '../evil' mf-tok

# ===========================================================================
# 8. kan-499's chunked contract -- the legal shape. 24 round-0 findings chunk
#    into three round-1 dispatches: the bare key carries chunk 1, `-2` and
#    `-3` continue it, contiguous, and 3 <= ceil(24/10). Exit 0, no stderr.
#    RED DRIVER: today's guard rejects ANY second dispatch per round.
# ===========================================================================
wt="$(make_worktree_json "$(dispatch_json panel-fix-1 panel-fix mf-tok panel-fix-1-2 panel-fix mf-tok panel-fix-1-3 panel-fix mf-tok)" "$(findings_json 24 0)")"
expect_exit_and_no_stderr 'case 8: chunked-legal-round-passes' 0 run_guard "$wt"

# ===========================================================================
# 9. A chunk's handshake retry -- the one allowed second dispatch lands on
#    the CHUNK's own key. 20 round-0 findings bound 3 chunks; the retry does
#    not count as a chunk. Exit 0. RED DRIVER alongside case 8.
# ===========================================================================
wt="$(make_worktree_json "$(dispatch_json panel-fix-1 panel-fix mf-tok panel-fix-1-2 panel-fix mf-tok panel-fix-1-2-retry panel-fix mf-tok)" "$(findings_json 20 0)")"
expect_exit 'case 9: chunk-retry-pair-passes' 0 run_guard "$wt"

# ===========================================================================
# 10. The legacy single-dispatch round still passes with the findings read
#     in the loop -- empty findings, one original plus its retry. Exit 0.
# ===========================================================================
wt="$(make_worktree_json "$(dispatch_json panel-fix-1 panel-fix mf-tok panel-fix-1-retry panel-fix mf-tok)")"
expect_exit 'case 10: legacy-round-with-findings-verb-passes' 0 run_guard "$wt"

# ===========================================================================
# 11. A chunked round with no bare key -- chunk numbering must start at the
#     round's own panel-fix-<round> dispatch. Exit 1, naming the round.
# ===========================================================================
wt="$(make_worktree_json "$(dispatch_json panel-fix-1-2 panel-fix mf-tok)" "$(findings_json 24 0)")"
expect_exit_and_names 'case 11: chunk-without-bare-key-fails, naming the round' 1 'panel-fix-1' run_guard "$wt"

# ===========================================================================
# 12. A gap in the chunk numbering -- 2 then 4 with no 3 is not a chunking
#     of any finding list. Exit 1, naming panel-fix-1-4.
# ===========================================================================
wt="$(make_worktree_json "$(dispatch_json panel-fix-1 panel-fix mf-tok panel-fix-1-2 panel-fix mf-tok panel-fix-1-4 panel-fix mf-tok)" "$(findings_json 24 0)")"
expect_exit_and_names 'case 12: chunk-gap-fails, naming the gapped chunk' 1 'panel-fix-1-4' run_guard "$wt"

# ===========================================================================
# 13. Over the bound -- 15 round-0 findings bound the round to
#     ceil(15/10)=2 chunks; three chunk dispatches is the per-finding abuse
#     in well-formed clothing (KAN-482). Exit 1, naming the round.
# ===========================================================================
wt="$(make_worktree_json "$(dispatch_json panel-fix-1 panel-fix mf-tok panel-fix-1-2 panel-fix mf-tok panel-fix-1-3 panel-fix mf-tok)" "$(findings_json 15 0)")"
expect_exit_and_names 'case 13: chunk-count-over-bound-fails, naming the round' 1 'panel-fix-1' run_guard "$wt"

# ===========================================================================
# 14. A chunk's retry with no original on that chunk's key -- the retry
#     rule holds per chunk, not merely per round. Exit 1, naming the chunk.
# ===========================================================================
wt="$(make_worktree_json "$(dispatch_json panel-fix-1 panel-fix mf-tok panel-fix-1-2-retry panel-fix mf-tok)" "$(findings_json 20 0)")"
expect_exit_and_names 'case 14: chunk-retry-alone-fails, naming the chunk' 1 'panel-fix-1-2' run_guard "$wt"

# ===========================================================================
# 15. Two originals on one chunk key -- the one-dispatch-per-base rule
#     applies to each chunk key. Exit 1, naming panel-fix-1-2.
# ===========================================================================
wt="$(make_worktree_json "$(dispatch_json panel-fix-1 panel-fix mf-tok panel-fix-1-2 panel-fix mf-tok panel-fix-1-2 panel-fix mf-tok)" "$(findings_json 24 0)")"
expect_exit_and_names 'case 15: two-originals-on-chunk-fails, naming the chunk' 1 'panel-fix-1-2' run_guard "$wt"

# ===========================================================================
# 16. The KAN-482 key shape survives chunking -- panel-fix-1-f2 is one
#     dispatch per finding under an invented key, out of shape still.
#     Exit 1, naming the key verbatim.
# ===========================================================================
wt="$(make_worktree_json "$(dispatch_json panel-fix-1 panel-fix mf-tok panel-fix-1-f2 panel-fix mf-tok)" "$(findings_json 24 0)")"
expect_exit_and_names 'case 16: invented-chunk-key-shape-fails, naming it verbatim' 1 'panel-fix-1-f2' run_guard "$wt"

# ===========================================================================
# 17. A findings read the guard cannot establish is "cannot answer" (exit
#     2), never a clean verdict -- same posture as the dispatches read.
# ===========================================================================
wt="$(make_worktree_findings_unreachable)"
expect_exit 'case 17: findings-store-unreachable-exits-2' 2 run_guard "$wt"

if [ "$FAILED" -ne 0 ]; then
  printf 'FAIL: test-check-panel-fix-single-dispatch had failures\n'
  exit 1
fi
printf 'PASS: test-check-panel-fix-single-dispatch\n'
