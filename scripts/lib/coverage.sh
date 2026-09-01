# scripts/lib/coverage.sh — per-member coverage, defined once.
#
# A guard exits 0 when it finds no violation — the same result a guard
# produces when it examined nothing at all. KAN-73 hit exactly this:
# check-guard-symlinks.sh's rule 2 computed an empty required set for
# skills/myflow-fast/, and deleting a symlink from that skill still reported
# GUARD-SYMLINKS-OK. Three reviewers read that guard and missed it, because
# the guard's own output gave them nothing to notice. This library exists so
# that four guards (check-guard-symlinks.sh, check-references.sh,
# check-vocabulary.sh, check-stage-mark-calls.sh) make "nothing was checked
# for this member" a visible, failing fact on a healthy tree, and so that
# they cannot disagree with each other about how — the exact five-copy
# `resolve_file` drift KAN-73's own review raised as a Critical, and that
# scripts/lib/resolve-file.sh's header records happening a second time
# before it existed. Written once, ahead of any of the four guards adopting
# it, rather than extracted after the fact.
#
# It owns exactly three things, so no guard sourcing it can invent a fourth
# way to do any of them:
#
#   RECORDING a per-member count       — coverage_record
#   RENDERING the members-and-counts   — coverage_report
#     fragment a verdict line carries
#   DECIDING declared vs. undeclared   — coverage_verdict
#     zero (fine vs. a named, non-zero-exit violation)
#
# Declaration is a plain function call the guard's own source makes
# (coverage_declare), never inferred from the tree — inferring it would
# restate the assumption the zero already encodes, which is exactly the case
# this library exists to fail instead of pass.
#
# ADOPTED, NOT INVENTED: every guard in this repository holds to three
# disciplines — `-a` on every grep, the `rc > 1` split between "no match"
# and a real error, and `--` before every path. This library reads no file
# and takes no path argument — it only keeps in-process state a guard feeds
# it — so the first and third have no call site here. The posture behind
# all three still does: reject anything that is not unambiguously valid
# rather than fold it into a reassuring default. coverage_record's count
# argument is the one place this library validates input from a caller, and
# it is rejected outright (a non-zero return, not a coerced 0) the moment it
# is not a plain non-negative integer — the same "never fail toward
# reassurance" rule behind a grep's `rc > 1` split.
#
# Not meant to be executed directly — a caller sources it and calls its
# functions; it sets no `set -euo pipefail` of its own and relies on the
# sourcing script's (or, for scripts/test-lib-coverage.sh, the test
# harness's).
#
# BASH 3.2 IS THE FLOOR (macOS's own /bin/bash), so every array below is
# walked by its indices — "${!ARRAY[@]}" — and never by its values —
# "${ARRAY[@]}" — directly. On bash 3.2, "${ARRAY[@]}" on an EMPTY array is
# unbound-variable-under-`set -u`; "${!ARRAY[@]}" and "${#ARRAY[@]}" on the
# same empty array are not. A guard sourcing this library runs under its own
# `set -u`, so this file must be safe under that, not merely under whatever
# bash happens to run these functions during development. No associative
# arrays and no `mapfile` either — both are bash-4-only.

# coverage_reset — clear every member, count and declaration this library is
# holding. Called once at source time below so the arrays exist and are
# typed before any guard calls into this library, and callable again by a
# caller — such as scripts/test-lib-coverage.sh — that runs more than one
# independent coverage pass in the same shell process.
coverage_reset() {
  COVERAGE_MEMBERS=()
  COVERAGE_COUNTS=()
  COVERAGE_DECLARED_MEMBERS=()
  COVERAGE_DECLARED_REASONS=()
}

# coverage_record <member> <count> — record how many items a guard checked
# for <member>. Call once per member; a second call for the same member is a
# caller bug and is rejected (return 2) rather than silently overwriting or
# accumulating, since either would let a guard's own double-counting or
# double-scanning bug hide behind a plausible total. <count> must be a plain
# non-negative integer — anything else (empty, negative, non-digit) is
# rejected the same way, never coerced to 0.
coverage_record() {
  if [ "$#" -ne 2 ]; then
    printf 'coverage_record: want 2 arguments (member, count), got %s\n' "$#" >&2
    return 2
  fi
  local member="$1" count="$2" i
  if [ -z "$member" ]; then
    printf 'coverage_record: member name must not be empty\n' >&2
    return 2
  fi
  case "$count" in
    ''|*[!0-9]*)
      printf "coverage_record: count must be a non-negative integer, got '%s'\n" "$count" >&2
      return 2
      ;;
  esac
  for i in "${!COVERAGE_MEMBERS[@]}"; do
    if [ "${COVERAGE_MEMBERS[$i]}" = "$member" ]; then
      printf "coverage_record: '%s' was already recorded — call coverage_record once per member\n" "$member" >&2
      return 2
    fi
  done
  COVERAGE_MEMBERS+=("$member")
  COVERAGE_COUNTS+=("$count")
}

# coverage_declare <member> [reason] — declare that <member> legitimately
# checks nothing, so its zero is fine rather than a violation. <reason>, when
# given, is carried through to coverage_report's rendering of that member —
# see check-stage-mark-calls.sh's flow-status case, which is a declared
# zero BECAUSE it is a read-only report that marks no stage runs, and whose
# declaration should say so rather than just list the name. A second call
# for the same member is a caller bug and is rejected (return 2), same as
# coverage_record.
coverage_declare() {
  if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
    printf 'coverage_declare: want 1 or 2 arguments (member, [reason]), got %s\n' "$#" >&2
    return 2
  fi
  local member="$1" reason="${2:-}" i
  if [ -z "$member" ]; then
    printf 'coverage_declare: member name must not be empty\n' >&2
    return 2
  fi
  for i in "${!COVERAGE_DECLARED_MEMBERS[@]}"; do
    if [ "${COVERAGE_DECLARED_MEMBERS[$i]}" = "$member" ]; then
      printf "coverage_declare: '%s' was already declared — call coverage_declare once per member\n" "$member" >&2
      return 2
    fi
  done
  COVERAGE_DECLARED_MEMBERS+=("$member")
  COVERAGE_DECLARED_REASONS+=("$reason")
}

# coverage_is_declared <member> -> exit 0 if <member> was declared expected-
# zero, 1 otherwise. Used by coverage_report and coverage_verdict; exposed
# because a guard may want the same answer for its own reporting.
coverage_is_declared() {
  local member="$1" i
  for i in "${!COVERAGE_DECLARED_MEMBERS[@]}"; do
    [ "${COVERAGE_DECLARED_MEMBERS[$i]}" = "$member" ] && return 0
  done
  return 1
}

# coverage_is_recorded <member> -> exit 0 if <member> has a coverage_record
# entry (of any count), 1 otherwise. Used by coverage_verdict's declared-but-
# never-recorded check (KAN-197 F3) below; exposed for the same reason
# coverage_is_declared is.
coverage_is_recorded() {
  local member="$1" i
  for i in "${!COVERAGE_MEMBERS[@]}"; do
    [ "${COVERAGE_MEMBERS[$i]}" = "$member" ] && return 0
  done
  return 1
}

# coverage_declared_reason <member> -> prints the reason <member> was
# declared with (may be empty), exit 0, if it is declared; exit 1 with no
# output if it is not declared at all. Distinct from an empty reason, which
# prints an empty string at exit 0.
coverage_declared_reason() {
  local member="$1" i
  for i in "${!COVERAGE_DECLARED_MEMBERS[@]}"; do
    if [ "${COVERAGE_DECLARED_MEMBERS[$i]}" = "$member" ]; then
      printf '%s' "${COVERAGE_DECLARED_REASONS[$i]}"
      return 0
    fi
  done
  return 1
}

# coverage_report -> prints ONE line: every recorded member and its count,
# in the order coverage_record saw them, joined by " · " — the fragment a
# guard's own verdict line carries (the guard supplies its own prefix, e.g.
# "GUARD-SYMLINKS-OK: <repo> — 53 guard(s) across 6 skill(s) validated").
# A declared zero is suffixed " (declared)", or " (declared: <reason>)" when
# a reason was given. Prints nothing for an empty corpus — coverage_verdict,
# not this function, is what turns that into a failure; this function only
# renders.
coverage_report() {
  local n=${#COVERAGE_MEMBERS[@]}
  [ "$n" -eq 0 ] && return 0
  local i member count reason line=""
  for i in "${!COVERAGE_MEMBERS[@]}"; do
    member="${COVERAGE_MEMBERS[$i]}"
    count="${COVERAGE_COUNTS[$i]}"
    [ -n "$line" ] && line="${line} · "
    if [ "$count" -eq 0 ] && coverage_is_declared "$member"; then
      reason="$(coverage_declared_reason "$member")"
      if [ -n "$reason" ]; then
        line="${line}${member} ${count} (declared: ${reason})"
      else
        line="${line}${member} ${count} (declared)"
      fi
    else
      line="${line}${member} ${count}"
    fi
  done
  printf '%s\n' "$line"
}

# coverage_verdict -> the expected-zero comparison. Prints one line per
# undeclared-zero member, "<member>: 0 checked, and not declared
# expected-zero (coverage)", and returns 1 if it printed anything. An empty
# corpus — coverage_record was never called at all — is itself printed as a
# violation and returns 1, per this change's own requirement: the most
# extreme form of "nothing was checked" is checking for no member whatever,
# and that SHALL NOT read as a vacuous pass. Prints nothing and returns 0
# when every zero-count member is declared (or there are no zero-count
# members at all).
#
# KAN-197 F3: a SECOND pass over COVERAGE_DECLARED_MEMBERS, printing one line
# per declared member that was never coverage_record'd at all — "<member>:
# declared expected-zero but never recorded (coverage)". The first pass above
# only ever reasons over members that WERE recorded; a coverage_declare for a
# member nobody scanned is otherwise silently inert forever, and the
# declaration list can only grow, never self-prune, as a renamed or deleted
# member leaves a stale entry nobody notices. A legitimate declared-zero
# member IS always also recorded (with count 0) by every guard's own scan —
# see check-guard-symlinks.sh, check-references.sh and check-vocabulary.sh,
# each of which records every corpus member it enumerates before checking
# whether that member's declaration applies — so this second pass costs
# nothing on a healthy guard and only fires on a genuinely stale declaration.
#
# KAN-374 F9: the first pass above only ever reasons about a member recorded
# at count 0 — a member that was declared expected-zero and whose recorded
# count later becomes non-zero falls into neither branch (it is not 0, so
# the undeclared-zero check never runs; coverage_report's own "(declared:
# ...)" suffix is gated on the same count-is-0 condition, so it silently
# stops rendering the reason too) and the guard kept exiting 0 with the
# stale reason just quietly dropped. A declared zero is a claim ("nothing
# will ever be checked here, because ...") and the moment the count is
# non-zero that claim is false — exactly the case a declaration mechanism
# exists to make loud, the same way an undeclared zero already is.
coverage_verdict() {
  local n=${#COVERAGE_MEMBERS[@]}
  if [ "$n" -eq 0 ]; then
    printf 'coverage: no member was ever recorded — this guard checked nothing (empty corpus)\n'
    return 1
  fi
  local i member count reason rc=0
  for i in "${!COVERAGE_MEMBERS[@]}"; do
    member="${COVERAGE_MEMBERS[$i]}"
    count="${COVERAGE_COUNTS[$i]}"
    if [ "$count" -eq 0 ] && ! coverage_is_declared "$member"; then
      printf '%s: 0 checked, and not declared expected-zero (coverage)\n' "$member"
      rc=1
    elif [ "$count" -ne 0 ] && coverage_is_declared "$member"; then
      reason="$(coverage_declared_reason "$member")"
      if [ -n "$reason" ]; then
        printf '%s: %s checked, but declared expected-zero (%s) — declaration is now false (coverage)\n' \
          "$member" "$count" "$reason"
      else
        printf '%s: %s checked, but declared expected-zero — declaration is now false (coverage)\n' \
          "$member" "$count"
      fi
      rc=1
    fi
  done
  for i in "${!COVERAGE_DECLARED_MEMBERS[@]}"; do
    member="${COVERAGE_DECLARED_MEMBERS[$i]}"
    if ! coverage_is_recorded "$member"; then
      printf '%s: declared expected-zero but never recorded — stale declaration, or this member was never scanned (coverage)\n' "$member"
      rc=1
    fi
  done
  return "$rc"
}

coverage_reset
