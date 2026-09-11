#!/usr/bin/env bash
# check-visual-verify-dispatched.sh — report whether a change whose diff
# touched a declared `## visual verification` UI path actually dispatched
# `flow.visual-verify`'s own verifier, for /flow's run-1 finish gate.
#
# Usage: check-visual-verify-dispatched.sh <worktree> <change-name> <merge-base>
#
# WHY THIS GUARD EXISTS. `flow.visual-verify` (skills/flow/verify-and-handoff.md)
# is the one stage built to catch a defect that is obvious the moment a page
# is opened but invisible to a diff, a review panel and both test suites —
# see that file's own **Blocking** paragraph. Nothing stopped an agent from
# writing `flow stage begin`/`flow stage end -outcome completed` around
# `flow.visual-verify` with NOTHING inside — no verifier dispatched, no screenshot
# read, no mockup composed — and reaching this handoff anyway. kan-30's own
# fix round 4 did exactly that: the stage's marks were written, the operator
# was told "no open finding", and a five-minute manual pass immediately
# after found the shipped screens badly wrong. A stage mark is not evidence
# the stage ran; only a recorded dispatch is. This guard makes the gap loud
# rather than silent, the same way check-panel-fix-single-dispatch.sh did
# for a different flavour of "the mark says one thing, the dispatches say
# another."
#
# Prints ONE verdict line to stdout:
#   VISUAL-VERIFY-OK: <reason>          nothing outstanding
#   VISUAL-VERIFY-MISSING: <reason>     UI paths touched, no verifier dispatch found
#
# Exit 0 whenever a verdict was reached; exit 2 when it cannot answer at
# all — a non-directory worktree, a change name outside the allowlist, an
# empty merge-base, jq missing, or a store call that failed outright (never
# read as "no dispatches", exactly as check-panel-fix-single-dispatch.sh's
# own header explains for the identical hazard: an outage that read as zero
# rows would pass every round it was blind to).
#
# THE TRIGGER QUESTION IS DELEGATED, NOT RE-ASKED. Whether this change's
# diff touched a declared `ui paths` glob is exactly
# check-visual-trigger.sh's own job (flow.visual-verify step 2's own guard);
# this script pipes the same `git diff --name-only <merge-base>..HEAD`
# through it and reads its three exit codes as-is:
#   2  the project declares no `## visual verification` section at all —
#      VISUAL-VERIFY-OK: not configured. Nothing here to require.
#   1  the section is declared but this diff touched none of its `ui
#      paths` — VISUAL-VERIFY-OK: no UI paths touched.
#   0  at least one touched path matched — continue to the dispatch check
#      below. check-visual-trigger.sh reads `.flow/project.md` from the
#      worktree itself, matching flow.visual-verify's own per-worktree
#      resolution; this guard is called once per worktree in the run's
#      resolved set, exactly as that stage is.
#
# THE DISPATCH CHECK READS THE STORE, NOT A RENDERED FILE — the same
# posture check-panel-fix-single-dispatch.sh and check-panel-reproducers.sh
# already take for the identical reason: `flow record dispatches -change
# <name> -C <worktree>` answers with the decoded JSON array of every
# dispatch row this change has ever recorded, across every run. A row
# counts when its `role` is `verifier`, its `key` starts with
# `visual-verify` (`visual-verify`, `visual-verify-<worktree-basename>`,
# `visual-verify-2`, `visual-verify-retry`, and combinations — this guard
# does not validate the key's exact shape, only that the stage's own
# canonical prefix was used; a malformed key is a different guard's job,
# the same division check-panel-fix-single-dispatch.sh draws between "did
# it happen" here and "did it happen the right number of times" there),
# and its `outcome` is exactly `completed` — a row still open (no `end`
# ever recorded), or one closed `aborted` or `fallback` with no completed
# retry beside it, is not evidence the verifier's report was ever read; see
# **The return** (skills/flow/implement.md) for why only `completed` closes
# a dispatch's own story.
#
# NO PROJECT NEEDS A ROW PER WORKTREE. A cross-repo change resolves more
# than one worktree, and flow.visual-verify dispatches once per worktree in
# its own resolved set (**Visual verification**, verify-and-handoff.md) —
# so this guard is called once per worktree too, by its own caller, exactly
# as check-base-moved.sh and check-unfinished-work.sh already are; it never
# tries to resolve the whole set itself.
#
# THE CHANGE-NAME CONTAINMENT CASE IS DUPLICATED, on purpose, from
# check-panel-fix-single-dispatch.sh's own copy — see that guard's header
# for why: the name arrives from a pull-request-editable state file and is
# passed to `flow record dispatches -change`, so `../../../planted` and a
# glob metacharacter are hazards here exactly as they are there.
set -euo pipefail

export LC_ALL=C

WORKTREE="${1:-}"
NAME="${2:-}"
MERGE_BASE="${3:-}"
[[ -n "$WORKTREE" && -d "$WORKTREE" ]] || { echo "check-visual-verify-dispatched: not a directory: ${WORKTREE:-<missing>}" >&2; exit 2; }
[[ -n "$NAME" ]] || { echo "check-visual-verify-dispatched: usage: check-visual-verify-dispatched.sh <worktree> <change-name> <merge-base>" >&2; exit 2; }
[[ -n "$MERGE_BASE" ]] || { echo "check-visual-verify-dispatched: a merge-base is required — pass the state file's recorded value for this worktree" >&2; exit 2; }

case "$NAME" in
  [!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789]* \
  | *[!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-]*)
    echo "check-visual-verify-dispatched: change name '$NAME' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'" >&2
    exit 2
    ;;
esac

WORKTREE="$(cd "$WORKTREE" && pwd -P)" || { echo "check-visual-verify-dispatched: worktree vanished: $WORKTREE" >&2; exit 2; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TRIGGER_GUARD="$SCRIPT_DIR/check-visual-trigger.sh"
[ -x "$TRIGGER_GUARD" ] || { echo "check-visual-verify-dispatched: required sibling module not found or not executable: $TRIGGER_GUARD" >&2; exit 2; }

command -v jq >/dev/null 2>&1 || { echo "check-visual-verify-dispatched: jq is required but was not found" >&2; exit 2; }

# Every git invocation whose failure would otherwise be read as an answer is
# captured and checked on its own line, never piped straight into the
# trigger guard — the same discipline check-base-moved.sh's own comment
# documents for the identical hazard.
if ! CHANGED_PATHS="$(git -C "$WORKTREE" diff --name-only "$MERGE_BASE"..HEAD 2>&1)"; then
  echo "check-visual-verify-dispatched: git diff failed against merge-base '$MERGE_BASE' in $WORKTREE — cannot answer: $CHANGED_PATHS" >&2
  exit 2
fi

TRIGGER_EXIT=0
printf '%s\n' "$CHANGED_PATHS" | "$TRIGGER_GUARD" "$WORKTREE" >/dev/null 2>/tmp/check-visual-verify-dispatched.trigger.$$ || TRIGGER_EXIT=$?
TRIGGER_STDERR="$(cat /tmp/check-visual-verify-dispatched.trigger.$$ 2>/dev/null || true)"
rm -f /tmp/check-visual-verify-dispatched.trigger.$$

case "$TRIGGER_EXIT" in
  2)
    echo "VISUAL-VERIFY-OK: not configured"
    exit 0
    ;;
  1)
    echo "VISUAL-VERIFY-OK: no UI paths touched"
    exit 0
    ;;
  0)
    ;;
  *)
    echo "check-visual-verify-dispatched: check-visual-trigger.sh exited $TRIGGER_EXIT, neither 0, 1 nor 2 — cannot answer: $TRIGGER_STDERR" >&2
    exit 2
    ;;
esac

# A store the call could not reach is "cannot answer", never "no rows" —
# the identical posture check-panel-fix-single-dispatch.sh takes for the
# same reason: an outage that read as zero dispatches would pass every
# round it was blind to.
if ! ROWS="$(flow record dispatches -change "$NAME" -C "$WORKTREE" 2>/dev/null)"; then
  echo "check-visual-verify-dispatched: flow record dispatches failed for '$NAME' — cannot answer" >&2
  exit 2
fi

if ! printf '%s' "$ROWS" | jq empty >/dev/null 2>&1; then
  echo "check-visual-verify-dispatched: dispatch rows were not readable JSON — cannot answer" >&2
  exit 2
fi

MATCH_COUNT="$(printf '%s' "$ROWS" | jq \
  '[.[] | select((.role // "") == "verifier" and ((.key // "") | startswith("visual-verify")) and (.outcome // "") == "completed")] | length')"

if [ "$MATCH_COUNT" -ge 1 ]; then
  echo "VISUAL-VERIFY-OK: UI paths touched and a completed verifier dispatch is recorded for '$NAME'"
  exit 0
fi

echo "VISUAL-VERIFY-MISSING: this change's diff touched a declared UI path but no completed 'verifier' dispatch (key starting 'visual-verify') is recorded for '$NAME' — flow.visual-verify's stage marks were written with no verifier ever dispatched, or its report was never read to completion"
exit 1
