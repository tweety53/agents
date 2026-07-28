#!/usr/bin/env bash
# state-advance.sh — the mechanical half of myflow's pure state write.
#
# Handles the happy path: state file readable and structurally complete, current
# stage in --accepted, worktree (if any) still present in ITS OWN repository.
# Anything needing judgment — resolving a name across multiple candidates, the
# stage-mismatch override prompt, artifact-based self-heal — is NOT done here.
# The script exits with a distinct code and the calling command loads the
# myflow-state-advance skill, which behaves exactly as it always has.
#
#   0  wrote the new stage
#   2  usage error, or jq is not installed
#   3  state file missing, unparseable, not a JSON object, missing a required
#      key, or a non-null worktree that its own repository no longer lists
#   4  current stage not in --accepted
#   5  --target originStage but originStage is null/missing
#   6  --target originStage but originStage is outside the six legal origins
#
# Exit statuses are OWNED by this script: jq's own statuses (2 usage, 3 compile,
# 5 runtime) are never allowed to reach the caller, because they collide with
# three of the codes above and would route a corrupt state file to "originStage
# is null" instead of to self-heal. Every jq invocation is therefore captured
# and translated.
#
# MYFLOW_STATE_FILE overrides path resolution (used by the test harness).
set -euo pipefail

NAME=""; TARGET=""; ACCEPTED=""; BY=""

while [ $# -gt 0 ]; do
  case "$1" in
    --name|--target|--accepted|--by)
      if [ $# -lt 2 ]; then
        printf 'usage error: %s requires a value\n' "$1" >&2
        exit 2
      fi
      flag="$1"; val="$2"; shift 2
      case "$flag" in
        --name)     NAME="$val" ;;
        --target)   TARGET="$val" ;;
        --accepted) ACCEPTED="$val" ;;
        --by)       BY="$val" ;;
      esac
      ;;
    *) printf 'usage error: unknown argument %s\n' "$1" >&2; exit 2 ;;
  esac
done

for required in NAME TARGET ACCEPTED BY; do
  eval "value=\$$required"
  if [ -z "$value" ]; then
    printf 'usage error: --%s is required\n' "$(printf '%s' "$required" | tr '[:upper:]' '[:lower:]')" >&2
    exit 2
  fi
done

# --name is interpolated into a filesystem path, so it is validated here rather
# than trusted from prose. The change-naming contract's slug rule is
# [a-z0-9-] (see skills/myflow-contracts/jira-integration.md); enforcing it in
# code is what stops `--name ../../elsewhere` from writing outside the state
# directory.
case "$NAME" in
  *[!a-z0-9-]*|-*|*-|"")
    printf 'usage error: --name must match [a-z0-9-] and may not start or end with "-": %s\n' "$NAME" >&2
    exit 2 ;;
esac

# The twelve pipeline stages, plus the literal `originStage` for the dynamic
# target. An unvalidated --target would write a typo'd stage and exit 0.
VALID_STAGES="awaiting-proposal-review proposal-done awaiting-do-review do-review-started \
do-done awaiting-fix-review fix-review-started awaiting-manual-test manual-test-done \
awaiting-pr-review review-done finished"

stage_is_valid() {
  local candidate="$1" s
  for s in $VALID_STAGES; do
    [ "$s" = "$candidate" ] && return 0
  done
  return 1
}

if [ "$TARGET" != "originStage" ] && ! stage_is_valid "$TARGET"; then
  printf 'usage error: --target %s is not a pipeline stage\n' "$TARGET" >&2
  exit 2
fi

# --accepted is a comma-separated stage list. Whitespace around an element is
# tolerated (`a, b` is the natural way to write it) but every element must still
# be a real stage.
ACCEPTED_LIST=""
for candidate in $(printf '%s' "$ACCEPTED" | tr ',' ' '); do
  if ! stage_is_valid "$candidate"; then
    printf 'usage error: --accepted contains %s, which is not a pipeline stage\n' "$candidate" >&2
    exit 2
  fi
  ACCEPTED_LIST="$ACCEPTED_LIST $candidate"
done
if [ -z "$ACCEPTED_LIST" ]; then
  printf 'usage error: --accepted names no stage\n' >&2
  exit 2
fi

command -v jq >/dev/null 2>&1 || { printf 'jq is required\n' >&2; exit 2; }

# Resolve the state file exactly as the State file contract specifies (see
# skills/myflow-contracts/state-file.md): via --git-common-dir, so a worktree
# and the main checkout agree on one path.
if [ -n "${MYFLOW_STATE_FILE:-}" ]; then
  STATE_FILE="$MYFLOW_STATE_FILE"
else
  MAIN_CHECKOUT="$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd)"
  PROJECT_KEY="$(basename "$MAIN_CHECKOUT")-$(printf '%s' "$MAIN_CHECKOUT" | shasum | cut -c1-8)"
  STATE_FILE="$HOME/Agents/myflow/state/$PROJECT_KEY/$NAME.json"
fi

if [ ! -f "$STATE_FILE" ]; then
  printf 'escalate: no state file at %s\n' "$STATE_FILE" >&2
  exit 3
fi

# jq_test <filter> — true when the filter yields true. Any other outcome
# (false, or ANY jq failure status) is false, so a jq usage/compile/runtime
# status can never surface as one of this script's own exit codes.
jq_test() {
  jq -e "$1" "$STATE_FILE" >/dev/null 2>&1
}

# jq_read <filter> — prints the string result. Returns non-zero on any jq
# failure, which callers translate into exit 3.
jq_read() {
  jq -r "$1" "$STATE_FILE" 2>/dev/null
}

# A top-level array, string, or number parses as JSON and passes `jq -e .`, but
# `.stage` against it is a jq RUNTIME error — status 5, which is this script's
# "originStage is null" code. Require an object explicitly, before any field is
# read.
if ! jq_test 'type == "object"'; then
  printf 'escalate: state file is not a JSON object: %s\n' "$STATE_FILE" >&2
  exit 3
fi

# A partial state file must escalate, not be perpetuated. The contract's write
# rule is "always write every field, never a partial merge" — merging a new
# stage into `{"stage": "..."}` and exiting 0 would ship a three-field file that
# every later reader treats as authoritative.
#
# Only the STRUCTURAL keys are required: `stage`, and `gates` carrying all four
# gate values. The contract explicitly allows an optional field to be absent
# rather than null (`fastPath` — "`null`/absent"), so demanding the full
# twelve-key set would escalate a file the contract calls legitimate. Anything
# present is carried forward by the jq merge below untouched; anything the
# contract permits to be missing stays missing.
REQUIRED_KEYS='["stage","gates"]'
REQUIRED_GATES='["reviewed","tested","prOpened","prMerged"]'
if ! jq -e --argjson req "$REQUIRED_KEYS" --argjson gk "$REQUIRED_GATES" \
      '(.gates | type) == "object"
       and (($req - (keys_unsorted)) | length) == 0
       and (($gk - (.gates | keys_unsorted)) | length) == 0' \
      "$STATE_FILE" >/dev/null 2>&1; then
  printf 'escalate: state file is missing required keys (or gates is not an object): %s\n' \
    "$STATE_FILE" >&2
  exit 3
fi

CURRENT="$(jq_read '.stage // empty')" || {
  printf 'escalate: could not read stage from %s\n' "$STATE_FILE" >&2
  exit 3
}
if [ -z "$CURRENT" ]; then
  printf 'escalate: state file has no stage: %s\n' "$STATE_FILE" >&2
  exit 3
fi

# A worktree the state file names but ITS OWN repository no longer knows about
# is the contradiction State self-heal exists for — hand it to the skill.
#
# The probe must run against the repository that owns the recorded path, not
# against cwd's repository. A change spanning two repos is the documented normal
# case, and `git worktree list` in cwd's repo simply does not know about another
# repo's worktree: probing cwd would report a perfectly valid worktree as stale
# and instruct self-heal to null out a live path.
#
# And a probe that cannot be performed is not a contradiction (the contract says
# so in as many words): if the path exists but git cannot answer for it — not a
# repository, git unavailable — the check is UNKNOWN and the script continues.
# Only two outcomes escalate: the path is gone, or its own repository answered
# and did not list it.
WORKTREE="$(jq_read '.worktree // empty')" || {
  printf 'escalate: could not read worktree from %s\n' "$STATE_FILE" >&2
  exit 3
}
if [ -n "$WORKTREE" ] && [ "$WORKTREE" != "null" ]; then
  if [ ! -d "$WORKTREE" ]; then
    printf 'escalate: state file names a worktree that does not exist: %s\n' "$WORKTREE" >&2
    exit 3
  fi
  if WT_LIST="$(git -C "$WORKTREE" worktree list --porcelain 2>/dev/null)"; then
    WORKTREE_PHYS="$(cd "$WORKTREE" && pwd -P)"
    if ! printf '%s\n' "$WT_LIST" | grep -qxF "worktree $WORKTREE" \
       && ! printf '%s\n' "$WT_LIST" | grep -qxF "worktree $WORKTREE_PHYS"; then
      printf 'escalate: %s is not listed as a worktree by its own repository\n' "$WORKTREE" >&2
      exit 3
    fi
  fi
fi

stage_accepted=0
for candidate in $ACCEPTED_LIST; do
  [ "$candidate" = "$CURRENT" ] && stage_accepted=1
done
if [ "$stage_accepted" -eq 0 ]; then
  printf 'escalate: stage %s is not in accepted set %s\n' "$CURRENT" "$ACCEPTED" >&2
  exit 4
fi

CLEAR_ORIGIN=0
RESOLVED_TARGET="$TARGET"
if [ "$TARGET" = "originStage" ]; then
  ORIGIN="$(jq_read '.originStage // empty')" || {
    printf 'escalate: could not read originStage from %s\n' "$STATE_FILE" >&2
    exit 3
  }
  if [ -z "$ORIGIN" ] || [ "$ORIGIN" = "null" ]; then
    printf 'escalate: --target originStage but originStage is null/missing\n' >&2
    exit 5
  fi
  case "$ORIGIN" in
    awaiting-do-review|do-review-started|do-done|awaiting-manual-test|manual-test-done|awaiting-pr-review) ;;
    *) printf 'escalate: originStage %s is not one of the six legal origins\n' "$ORIGIN" >&2
       exit 6 ;;
  esac
  if [ "$ORIGIN" = "do-review-started" ]; then
    RESOLVED_TARGET="awaiting-do-review"
  else
    RESOLVED_TARGET="$ORIGIN"
  fi
  CLEAR_ORIGIN=1
fi

NOW="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
TMP="$STATE_FILE.tmp.$$"
trap 'rm -f "$TMP"' EXIT

if [ "$CLEAR_ORIGIN" -eq 1 ]; then
  FILTER='.stage = $s | .originStage = null | .updatedAt = $t | .updatedBy = $b'
else
  FILTER='.stage = $s | .updatedAt = $t | .updatedBy = $b'
fi
if ! jq --arg s "$RESOLVED_TARGET" --arg t "$NOW" --arg b "$BY" "$FILTER" \
      "$STATE_FILE" > "$TMP" 2>/dev/null; then
  printf 'escalate: could not render the new state for %s\n' "$STATE_FILE" >&2
  exit 3
fi
mv "$TMP" "$STATE_FILE"
trap - EXIT

printf '## Stage advanced\n\n'
printf '**Change:** %s\n' "$NAME"
printf '**Stage:** %s → %s\n' "$CURRENT" "$RESOLVED_TARGET"
printf '**State:** %s\n' "$STATE_FILE"
