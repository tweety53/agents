#!/usr/bin/env bash
# check-installed-citations.sh — wrapper.
#
# The classifier — deriving the installed set, the fence-aware token
# scanner, the root judgment — lives entirely in check-installed-citations.py
# (Python 3, standard library only), per task 2 step 1's own instruction to
# copy scripts/check-plan-provenance.sh's wrapper shape rather than invent
# a second one, including the python3-actually-runs probe below: `command
# -v` proves the file exists, not that it runs, and the realistic failure
# (macOS without the Command Line Tools) is a `/usr/bin/python3` stub that
# exits 1 before ever reaching the guard.
#
# WHY THIS FILE IS NOT A THIN `exec` (unlike check-plan-provenance.sh):
# task 2 step 5 requires the same per-member coverage discipline
# scripts/lib/coverage.sh already gives check-guard-symlinks.sh and
# check-references.sh — an undeclared zero is itself a violation, not a
# silent pass — and that library is Bash, sourced by a caller running under
# its own `set -u`. check-installed-citations.py cannot source it. So the
# Python script classifies and hands back two things behind sentinel
# prefixes neither of which collides with a real repository path (see its
# own CIC_ROOT_PREFIX/CIC_COVERAGE_PREFIX and main()'s closing comment): a
# real citation violation line, printed as-is, and one coverage line per
# scanned member. THIS file — not the Python — sources coverage.sh, calls
# coverage_record per member, declares the members that legitimately cite
# no path at all (below), and is what actually decides clean vs. violation
# and the final exit code. The Python script's own exit code is narrower:
# 2 for a refusal (unchanged), 0 otherwise — never 1, because whether an
# undeclared zero turns a run non-clean is this file's decision to make,
# not something Python can answer without reinventing coverage.sh's own
# declared-vs-undeclared logic a second time.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD_NAME="check-installed-citations"

command -v python3 >/dev/null 2>&1 || {
  echo "$GUARD_NAME: python3 not found on PATH — cannot run the guard" >&2
  exit 2
}

if ! python3 -c 'import sys; sys.exit(0)'; then
  echo "$GUARD_NAME: python3 is present but failed to run a trivial program (see above) — cannot run the guard" >&2
  exit 2
fi

# coverage_record / coverage_declare / coverage_report / coverage_verdict —
# per-member coverage reporting and the declared-vs-undeclared-zero
# decision, owned once in lib/coverage.sh rather than reinvented here. See
# that file's header for why (KAN-197) and check-guard-symlinks.sh /
# check-references.sh for the pattern this guard follows.
source "$SCRIPT_DIR/lib/coverage.sh"

WORK="$(mktemp -d "${TMPDIR:-/tmp}/check-installed-citations.XXXXXX")"
cleanup() { rm -rf "$WORK"; }
trap cleanup EXIT

# Run the classifier. `set +e` around it: a real citation violation is an
# ordinary, expected outcome here, never a shell failure, and the script's
# own exit code is narrower than this wrapper's (see the header above) —
# only 2 (a refusal, passed straight through with nothing on stdout, per
# the Python script's own contract) is ever treated as this wrapper's own
# early exit.
set +e
PY_OUT="$(python3 "$SCRIPT_DIR/check-installed-citations.py")"
PY_RC=$?
set -e

if [ "$PY_RC" -eq 2 ]; then
  exit 2
fi
if [ "$PY_RC" -ne 0 ]; then
  echo "$GUARD_NAME: check-installed-citations.py exited $PY_RC unexpectedly — cannot answer" >&2
  exit 2
fi

VIOLATIONS_FILE="$WORK/violations"
MEMBERS_FILE="$WORK/members"
: > "$VIOLATIONS_FILE"
: > "$MEMBERS_FILE"
ROOT=""

# Parse the Python script's protocol: a CIC-ROOT line (once), a
# CIC-COVERAGE line per scanned member, and every other non-empty line is
# a real citation violation, passed through byte-for-byte.
while IFS= read -r line || [ -n "$line" ]; do
  case "$line" in
    "CIC-ROOT"$'\t'*)
      ROOT="${line#CIC-ROOT$'\t'}"
      ;;
    "CIC-COVERAGE"$'\t'*)
      rest="${line#CIC-COVERAGE$'\t'}"
      member="${rest%%$'\t'*}"
      count="${rest#*$'\t'}"
      printf '%s\n' "$member" >> "$MEMBERS_FILE"
      if ! coverage_record "$member" "$count"; then
        echo "$GUARD_NAME: coverage_record failed for $member (see stderr above)" >&2
        exit 2
      fi
      ;;
    "")
      ;;
    *)
      printf '%s\n' "$line" >> "$VIOLATIONS_FILE"
      ;;
  esac
done <<<"$PY_OUT"

if [ -z "$ROOT" ]; then
  echo "$GUARD_NAME: check-installed-citations.py reported no root — cannot answer" >&2
  exit 2
fi

# declare_if_present <member> <reason> — declares <member> expected-zero
# ONLY when it is actually part of THIS run's corpus (present in
# MEMBERS_FILE — which may be a sandboxed CHECK_INSTALLED_CITATIONS_ROOT
# fixture, not this repository), exactly as check-guard-symlinks.sh's own
# declare_if_present does: declaring a name outside the current corpus
# would make it a KAN-197 F3 "declared but never recorded" violation for
# every fixture that does not happen to carry that file.
declare_if_present() {
  local member="$1" reason="$2" rc
  set +e
  grep -aqxF -- "$member" "$MEMBERS_FILE"
  rc=$?
  set -e
  if [ "$rc" -ge 2 ]; then
    echo "$GUARD_NAME: grep exited $rc while checking $MEMBERS_FILE for '$member'" >&2
    exit 2
  fi
  [ "$rc" -eq 0 ] || return 0
  if ! coverage_declare "$member" "$reason"; then
    echo "$GUARD_NAME: coverage_declare failed for '$member' (see stderr above)" >&2
    exit 2
  fi
}

# EXPECTED-ZERO MEMBERS — established by reading each file's own content
# (2026-08-20, at this change's own HEAD), never guessed: none of these
# carries a single backticked path (or, inside a bash/sh/zsh fence
# comment, a bare shell-word path) in any shape this guard's classifier
# recognises, so it genuinely verifies nothing inside them. Declared here,
# once, rather than inferred from the tree — inferring it would restate
# the very assumption a silently uncovered file already encodes.
declare_if_present "rules/be-brief.mdc" \
  "always-on rule body — cites no .md/.mdc path at all, backticked or bare"
declare_if_present "rules/build-the-simplest-thing.mdc" \
  "always-on rule body — cites no .md/.mdc path at all, backticked or bare"
declare_if_present "rules/context7.mdc" \
  "always-on rule body — cites no .md/.mdc path at all, backticked or bare"
declare_if_present "rules/dependency-versions.mdc" \
  "always-on rule body — cites no .md/.mdc path at all, backticked or bare"
declare_if_present "rules/design-mockups-are-specs.mdc" \
  "always-on rule body — cites no .md/.mdc path at all, backticked or bare"
declare_if_present "rules/never-touch-production.mdc" \
  "always-on rule body — cites no .md/.mdc path at all, backticked or bare"
declare_if_present "rules/no-direct-pushes-to-main.mdc" \
  "always-on rule body — cites no .md/.mdc path at all, backticked or bare"
declare_if_present "skills/myflow-do/bug-hunter-reviewer-prompt.md" \
  "reviewer-prompt file, deliberately self-contained — cites no .md/.mdc path anywhere"
declare_if_present "skills/myflow-do/engineering-principles.md" \
  "reviewer-prompt file, deliberately self-contained — cites no .md/.mdc path anywhere"
declare_if_present "skills/myflow-do/security-reviewer-prompt.md" \
  "reviewer-prompt file, deliberately self-contained — cites no .md/.mdc path anywhere"

COVERAGE_VERDICT_FILE="$WORK/coverage_verdict"
if ! coverage_verdict > "$COVERAGE_VERDICT_FILE"; then
  while IFS= read -r cvline; do
    [ -n "$cvline" ] || continue
    cvmember="${cvline%%:*}"
    cvmsg="${cvline#*: }"
    printf '%s:0: %s\n' "$cvmember" "$cvmsg" >> "$VIOLATIONS_FILE"
  done < "$COVERAGE_VERDICT_FILE"
fi

if [ -s "$VIOLATIONS_FILE" ]; then
  N_FOUND="$(wc -l < "$VIOLATIONS_FILE" | tr -d ' ')"
  cat "$VIOLATIONS_FILE"
  printf 'INSTALLED-CITATIONS-INVALID: %s — %s violation(s)\n' "$ROOT" "$N_FOUND"
  COVERAGE_FRAGMENT="$(coverage_report)"
  [ -n "$COVERAGE_FRAGMENT" ] && printf '  %s\n' "$COVERAGE_FRAGMENT"
  exit 1
fi

FILE_COUNT="$(wc -l < "$MEMBERS_FILE" | tr -d ' ')"
printf 'INSTALLED-CITATIONS-OK: %s — %s file(s) scanned\n' "$ROOT" "$FILE_COUNT"
COVERAGE_FRAGMENT="$(coverage_report)"
[ -n "$COVERAGE_FRAGMENT" ] && printf '  %s\n' "$COVERAGE_FRAGMENT"
exit 0
