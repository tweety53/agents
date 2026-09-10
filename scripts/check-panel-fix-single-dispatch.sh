#!/usr/bin/env bash
# check-panel-fix-single-dispatch.sh <worktree> <change-name> <session-token>
#
# THE CHANGE NAME IS A SECOND ARGUMENT and the session token a third, for
# the same reason the name is one on check-panel-findings-closed.sh: a
# change's dispatches are rows in the store, keyed by change name, not a
# file at a fixed path, and the token scopes the count to THIS run --
# `flow record dispatches -change <name> -C <worktree>` answers with the
# decoded JSON array of that change's dispatch rows (every role, every
# session token, seq order), and this guard reads only that array.
#
# This is the gate KAN-482 asked for: on the kan-449 run the conductor
# launched four background fix subagents, one per reviewer's findings,
# against the contract review-panel.md already stated -- give the surviving
# findings to ONE fix subagent as the combined list. Prose alone failed
# once, so the obligation is checked here, at the panel's own close,
# immediately before `flow stage end -command '/flow' -stage
# flow.review-panel`, beside check-panel-findings-closed.sh. Dispatch rows
# only exist once dispatches have happened, so a pre-dispatch check has
# nothing to read; the close gate is the last point where the run can still
# be stopped before handoff.
#
# THE COUNT IS SCOPED to the session token named on the command line --
# a change's rows span every run that ever touched it, and only this run's
# dispatches are this round's business. Within that token, every row whose
# role is `panel-fix` must satisfy:
#
#   - its key matches `^panel-fix-[0-9]+(-retry)?$` -- the canonical shape
#     review-panel.md's fix step declares; a conductor that invents keys
#     (panel-fix-f1, one per finding) would otherwise read as four clean
#     single-dispatch rounds, which is exactly the KAN-482 shape;
#   - its round (the key with a trailing `-retry` stripped) carries at most
#     one non-retry dispatch, at most two rows total, and two only when
#     exactly one of them is the `-retry` variant -- the handshake's one
#     allowed second dispatch, per implement.md's **The handshake**;
#   - a `-retry` row never stands alone: a retry with no original has no
#     round it could be a retry of.
#
# Rows of other roles and rows of other session tokens never count.
#
# THE GUARD NEVER READS THE JOURNAL and never consults anything but the
# store's answer, matching check-panel-findings-closed.sh: a read that
# failed is a question this guard cannot answer, not a clean verdict.
#
# THE CHANGE-NAME CONTAINMENT CASE IS DUPLICATED, on purpose, from
# check-panel-findings-closed.sh's own copy -- the change name arrives from
# a pull-request-editable state file and is passed to `flow record
# dispatches -change`, so `../../../planted` and a glob metacharacter are
# hazards here exactly as they are there. The worktree is canonicalised
# before it is ever passed to `flow` as `-C`, and refused rather than
# proceeded with if it vanished between the `-d` check and here, for the
# same reasons that guard's header names.
#
# Exit codes:
#   0  every panel-fix row of this session token is shape- and count-clean
#   1  at least one violation -- each offending round or key named on stderr
#   2  cannot answer at all -- missing arguments, a non-directory worktree,
#      a change name outside the allowlist, the store unreachable (the flow
#      call exits non-zero), or jq missing or failing
set -euo pipefail

export LC_ALL=C

WORKTREE="${1:-}"
NAME="${2:-}"
TOKEN="${3:-}"
[[ -n "$WORKTREE" && -d "$WORKTREE" ]] || { echo "check-panel-fix-single-dispatch: not a directory: ${WORKTREE:-<missing>}" >&2; exit 2; }
[[ -n "$NAME" ]] || { echo "check-panel-fix-single-dispatch: usage: check-panel-fix-single-dispatch.sh <worktree> <change-name> <session-token>" >&2; exit 2; }
[[ -n "$TOKEN" ]] || { echo "check-panel-fix-single-dispatch: session token is required and must be this run's own literal token" >&2; exit 2; }

# CONTAINMENT, identical to check-panel-findings-closed.sh's own copy and
# its own comment canonical for why the six-line `case` block stays
# duplicated rather than centralized: the change name arrives from a
# pull-request-editable state file and is passed to `flow record
# dispatches -change`, so `../../../planted` and a glob metacharacter are
# hazards here exactly as they are there.
case "$NAME" in
  [!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789]* \
  | *[!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-]*)
    echo "check-panel-fix-single-dispatch: change name '$NAME' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'" >&2
    exit 2
    ;;
esac

# Canonicalise before the worktree is ever passed to `flow` as `-C`, and
# refuse rather than proceed if it vanished between the `-d` check above and
# here -- see check-panel-findings-closed.sh's own comment on both hazards.
WORKTREE="$(cd "$WORKTREE" && pwd -P)" || { echo "check-panel-fix-single-dispatch: worktree vanished: $WORKTREE" >&2; exit 2; }

command -v jq >/dev/null 2>&1 || { echo "check-panel-fix-single-dispatch: jq is required but was not found" >&2; exit 2; }

# A store the call could not reach is "cannot answer", never "no rows": an
# outage that read as zero rows would pass every round it was blind to.
if ! rows="$(flow record dispatches -change "$NAME" -C "$WORKTREE" 2>/dev/null)"; then
  echo "check-panel-fix-single-dispatch: flow record dispatches failed for '$NAME' -- cannot answer" >&2
  exit 2
fi

# Each panel-fix row of THIS token, one compact object per line. jq failing
# here means the payload was not the shape the verb documents -- the same
# cannot-answer posture as a failed call.
if ! filtered="$(printf '%s' "$rows" | jq -c --arg tok "$TOKEN" \
    '.[] | select((.sessionToken // "") == $tok and (.role // "") == "panel-fix")')"; then
  echo "check-panel-fix-single-dispatch: dispatch rows were not readable JSON -- cannot answer" >&2
  exit 2
fi

violations=()

# Parallel plain-indexed arrays, not associative arrays: this machine's
# /bin/bash is 3.2, where a string subscript on an indexed array is
# evaluated arithmetically -- originals["panel-fix-1"] would look up a
# variable named `panel` and die under set -u ("panel: unbound variable").
base_list=()
orig_counts=()
retry_counts=()

# base_index <base> -- echoes the position of <base> in base_list, or
# nothing when absent. Linear scan is the whole job; a run has a handful of
# fix rounds at most.
base_index() {
  local want="$1" i=0 b
  for b in ${base_list[@]+"${base_list[@]}"}; do
    [ "$b" = "$want" ] && { echo "$i"; return 0; }
    i=$(( i + 1 ))
  done
  return 1
}

CANONICAL_KEY_RE='^panel-fix-[0-9]+(-retry)?$'

while IFS= read -r row; do
  [ -n "$row" ] || continue
  key="$(printf '%s' "$row" | jq -r '.key // ""')"
  if [[ ! "$key" =~ $CANONICAL_KEY_RE ]]; then
    violations+=("out-of-shape panel-fix key '$key' -- the canonical shape is panel-fix-<round> or panel-fix-<round>-retry")
    continue
  fi
  if [[ "$key" == *-retry ]]; then
    base="${key%-retry}"
    field=retry_counts
  else
    base="$key"
    field=orig_counts
  fi
  if idx="$(base_index "$base")"; then
    eval "$field[$idx]=\$(( \${$field[$idx]} + 1 ))"
  else
    base_list+=("$base")
    if [ "$field" = retry_counts ]; then
      orig_counts+=(0)
      retry_counts+=(1)
    else
      orig_counts+=(1)
      retry_counts+=(0)
    fi
  fi
# '%s\n', not '%s': read drops a final line with no trailing newline, and
# jq -c's last row must reach the loop like every other one.
done < <(printf '%s\n' "$filtered")

for idx in "${!base_list[@]}"; do
  base="${base_list[$idx]}"
  orig="${orig_counts[$idx]}"
  retry="${retry_counts[$idx]}"
  total=$(( orig + retry ))
  if [ "$retry" -gt 0 ] && [ "$orig" -eq 0 ]; then
    violations+=("round $base carries $retry -retry dispatch(es) and no original -- a retry is never a round's only dispatch")
  elif [ "$orig" -gt 1 ]; then
    violations+=("round $base carries $orig panel-fix dispatches -- the fix goes to ONE subagent as the combined list, never one per reviewer, slot or finding")
  elif [ "$total" -gt 2 ] || { [ "$total" -eq 2 ] && [ "$retry" -ne 1 ]; }; then
    violations+=("round $base carries $total panel-fix dispatches -- at most the original plus the handshake's one -retry")
  fi
done

if [ "${#violations[@]}" -gt 0 ]; then
  for v in "${violations[@]}"; do
    echo "check-panel-fix-single-dispatch: $v" >&2
  done
  exit 1
fi

echo "PANEL-FIX-SINGLE-DISPATCH-OK: every panel-fix dispatch of this run is one per round (retry included)"
