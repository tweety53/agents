#!/usr/bin/env bash
# check-task-reviewer-single-dispatch.sh <worktree> <change-name> <session-token>
#
# THE CHANGE NAME IS A SECOND ARGUMENT and the session token a third, for
# the same reason the name is one on check-panel-fix-single-dispatch.sh: a
# change's dispatches are rows in the store, keyed by change name, not a
# file at a fixed path, and the token scopes the count to THIS run --
# `flow record dispatches -change <name> -C <worktree>` answers with the
# decoded JSON array of that change's dispatch rows, and this guard reads
# only that array plus <worktree>/spectre/changes/<name>/tasks.md (the
# canonical worktree's copy -- the only one that exists on a cross-repo
# change, per skills/flow-contracts/worktree-resolution.md).
#
# This is the gate KAN-527 asked for, the gated-per-task-reviewer sibling
# of check-panel-fix-single-dispatch.sh: on that run the conductor launched
# nine one-task reviewer dispatches against nine gate-fired tasks that
# belonged to three implementer groups, before the operator stopped the
# run and implement.md's "The gated per-task reviewer" paragraph was
# rewritten to require ONE bundled dispatch per implementer group whose
# gate fired (or one bundle for the whole run on `small`/`regular`). Prose
# alone already failed once for the panel's own fix step (KAN-482); this
# guard exists so the same failure mode is caught here too, at the stage's
# own close, immediately before `flow stage end -command '/flow' -stage
# flow.sdd-tdd`, the same position check-panel-fix-single-dispatch.sh holds
# at `flow.review-panel`'s close.
#
# THE COUNT IS SCOPED to the session token named on the command line --
# a change's rows span every run that ever touched it, and only this run's
# dispatches are this run's business. Within that token, every row whose
# role is `reviewer` and whose key matches the gated-per-task-reviewer
# family (`^task-[0-9]+(\+[0-9]+)*-reviewer(-fix-[0-9]+)?(-retry)?$`) must
# satisfy:
#
#   - its key is in that canonical shape -- a key naming one task
#     (`task-5-reviewer`) or several `+`-joined in plan order
#     (`task-13+14+15-reviewer`), optionally a fix round
#     (`task-13+14-reviewer-fix-1`) and optionally the handshake's one
#     retry (`task-13+14-reviewer-fix-1-retry`); a conductor that invents
#     one key per task (`task-5-reviewer`, `task-6-reviewer`, ... for tasks
#     that share an implementer group) is exactly the KAN-527 shape;
#   - its base (the key with a trailing `-retry` stripped) carries at most
#     one non-retry dispatch, at most two rows total, and two only when
#     exactly one of them is the `-retry` variant -- the handshake's one
#     allowed second dispatch, per implement.md's **The handshake**;
#   - a `-retry` row never stands alone: a retry with no original has no
#     bundle it could be a retry of.
#
# BEYOND the per-key shape (which is everything check-panel-fix-single-
# dispatch.sh's own key family checks), this guard also checks the
# BUNDLING ITSELF, since a KAN-527-shaped run can produce well-formed keys
# that are still one-per-task: no two ORIGINAL (non-fix, non-retry) keys
# may name tasks belonging to the same implementer group. Group membership
# is read from `<worktree>/spectre/changes/<name>/tasks.md` via
# `plan-dispatch-bundles.sh` (task -> bundle) composed with
# `plan-dispatch-groups.sh` (bundle -> group) -- the same source of truth
# `skills/flow/brainstorm-planner.md`'s Decide step and implement.md's own
# wave-grouping already use, never re-derived by this guard from the raw
# markdown. On the decision's `class` `small` or `regular` (read via `flow
# record decisions -change <name> -C <worktree>`, the last entry's
# `.decision.class`; absent or unreadable is treated as `big`, the
# stricter posture, so a decision-read failure never loosens the check),
# implement.md requires every gate-fired task of the run to join ONE
# bundle at the last boundary -- so on those classes, more than one
# original key is itself a violation, independent of group membership.
#
# Rows of other roles and rows of other session tokens never count. A task
# id that never appears in any `reviewer`-role key of this token was never
# gated, or its gate hasn't fired yet, or its bundle is still in flight --
# this guard says nothing about coverage, only about the shape of what was
# already dispatched.
#
# THE GUARD NEVER READS THE JOURNAL and never consults anything but the
# store's answer for dispatch rows, matching check-panel-fix-single-
# dispatch.sh: a read that failed is a question this guard cannot answer,
# not a clean verdict. The tasks.md read is a plain file read, not a store
# read; a missing or unparsable tasks.md is likewise "cannot answer" (exit
# 2), never a clean verdict, since without it group membership cannot be
# established at all.
#
# THE CHANGE-NAME CONTAINMENT CASE IS DUPLICATED, on purpose, from
# check-panel-fix-single-dispatch.sh's own copy (itself duplicated from
# check-panel-findings-closed.sh) -- the change name arrives from a
# pull-request-editable state file and is passed to `flow record
# dispatches -change`, so `../../../planted` and a glob metacharacter are
# hazards here exactly as they are there. The worktree is canonicalised
# before it is ever passed to `flow` as `-C`, and refused rather than
# proceeded with if it vanished between the `-d` check and here, for the
# same reasons that guard's header names.
#
# Exit codes:
#   0  every gated-per-task reviewer row of this session token is
#      shape-clean, retry-clean, and bundled per implementer group (and,
#      on `small`/`regular`, bundled into one dispatch for the whole run)
#   1  at least one violation -- each offending key or group named on
#      stderr
#   2  cannot answer at all -- missing arguments, a non-directory
#      worktree, a change name outside the allowlist, the store
#      unreachable, tasks.md missing or unreadable, plan-dispatch-bundles.sh
#      or plan-dispatch-groups.sh failing, or jq missing or failing
set -euo pipefail

export LC_ALL=C

WORKTREE="${1:-}"
NAME="${2:-}"
TOKEN="${3:-}"
[[ -n "$WORKTREE" && -d "$WORKTREE" ]] || { echo "check-task-reviewer-single-dispatch: not a directory: ${WORKTREE:-<missing>}" >&2; exit 2; }
[[ -n "$NAME" ]] || { echo "check-task-reviewer-single-dispatch: usage: check-task-reviewer-single-dispatch.sh <worktree> <change-name> <session-token>" >&2; exit 2; }
[[ -n "$TOKEN" ]] || { echo "check-task-reviewer-single-dispatch: session token is required and must be this run's own literal token" >&2; exit 2; }

# CONTAINMENT, identical to check-panel-fix-single-dispatch.sh's own copy --
# see that script's header for why this six-line `case` block stays
# duplicated rather than centralized.
case "$NAME" in
  [!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789]* \
  | *[!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-]*)
    echo "check-task-reviewer-single-dispatch: change name '$NAME' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'" >&2
    exit 2
    ;;
esac

# Canonicalise before the worktree is ever passed to `flow` as `-C`, and
# refuse rather than proceed if it vanished between the `-d` check above and
# here -- see check-panel-fix-single-dispatch.sh's own comment on both hazards.
WORKTREE="$(cd "$WORKTREE" && pwd -P)" || { echo "check-task-reviewer-single-dispatch: worktree vanished: $WORKTREE" >&2; exit 2; }

command -v jq >/dev/null 2>&1 || { echo "check-task-reviewer-single-dispatch: jq is required but was not found" >&2; exit 2; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TASKS_MD="$WORKTREE/spectre/changes/$NAME/tasks.md"
[[ -f "$TASKS_MD" ]] || { echo "check-task-reviewer-single-dispatch: no tasks.md at $TASKS_MD -- cannot answer" >&2; exit 2; }

# A store the call could not reach is "cannot answer", never "no rows": an
# outage that read as zero rows would pass every run it was blind to.
if ! rows="$(flow record dispatches -change "$NAME" -C "$WORKTREE" 2>/dev/null)"; then
  echo "check-task-reviewer-single-dispatch: flow record dispatches failed for '$NAME' -- cannot answer" >&2
  exit 2
fi

if ! printf '%s' "$rows" | jq empty >/dev/null 2>&1; then
  echo "check-task-reviewer-single-dispatch: dispatch rows were not readable JSON -- cannot answer" >&2
  exit 2
fi

# Each reviewer row of THIS token whose key is in the gated-per-task-
# reviewer family, one compact object per line.
CANONICAL_KEY_RE='^task-[0-9]+(\+[0-9]+)*-reviewer(-fix-[0-9]+)?(-retry)?$'
if ! filtered="$(printf '%s' "$rows" | jq -c --arg tok "$TOKEN" \
    '.[] | select((.sessionToken // "") == $tok and (.role // "") == "reviewer" and ((.key // "") | test("^task-[0-9]+(\\+[0-9]+)*-reviewer")))')"; then
  echo "check-task-reviewer-single-dispatch: dispatch rows were not readable JSON -- cannot answer" >&2
  exit 2
fi

# Group membership: task id -> group number, from this run's own plan --
# never re-derived from the raw markdown by this guard.
if ! bundle_lines="$($SCRIPT_DIR/plan-dispatch-bundles.sh "$TASKS_MD" 2>&1)"; then
  echo "check-task-reviewer-single-dispatch: plan-dispatch-bundles.sh failed on $TASKS_MD -- cannot answer" >&2
  echo "$bundle_lines" >&2
  exit 2
fi
if ! group_lines="$($SCRIPT_DIR/plan-dispatch-groups.sh "$TASKS_MD" 2>&1)"; then
  echo "check-task-reviewer-single-dispatch: plan-dispatch-groups.sh failed on $TASKS_MD -- cannot answer" >&2
  echo "$group_lines" >&2
  exit 2
fi

# bundle <i> -> "id id id" ; group <g> -> "<bundle ids>". Compose into
# task_id -> group, a plain-indexed parallel-array map (bash 3.2 has no
# associative arrays reliably usable under `set -u` with numeric-looking
# keys -- same constraint check-panel-fix-single-dispatch.sh's own comment
# names for its base_list arrays).
bundle_task_ids=()   # bundle_task_ids[i] = "id id id" for bundle i+1
while IFS= read -r line; do
  [[ "$line" =~ ^bundle\ ([0-9]+):\ (.*)$ ]] || continue
  idx=$(( ${BASH_REMATCH[1]} - 1 ))
  bundle_task_ids[$idx]="${BASH_REMATCH[2]}"
done <<< "$bundle_lines"

task_group=()   # parallel to task_id_list
task_id_list=()
task_group_index() {
  local want="$1" i=0 t
  for t in ${task_id_list[@]+"${task_id_list[@]}"}; do
    [ "$t" = "$want" ] && { echo "$i"; return 0; }
    i=$(( i + 1 ))
  done
  return 1
}
while IFS= read -r line; do
  [[ "$line" =~ ^group\ ([0-9]+):\ (.*)$ ]] || continue
  g="${BASH_REMATCH[1]}"
  for bid in ${BASH_REMATCH[2]}; do
    bidx=$(( bid - 1 ))
    for tid in ${bundle_task_ids[$bidx]:-}; do
      task_id_list+=("$tid")
      task_group+=("$g")
    done
  done
done <<< "$group_lines"

group_of_task() {
  local want="$1" idx
  if idx="$(task_group_index "$want")"; then
    echo "${task_group[$idx]}"
    return 0
  fi
  return 1
}

# The decision's class -- absent, unreadable, or non-`sdd` reads as `big`,
# the stricter posture (more bundles tolerated), so a decision-read
# failure never loosens this check.
CLASS="big"
if decisions_json="$(flow record decisions -change "$NAME" -C "$WORKTREE" 2>/dev/null)" \
    && printf '%s' "$decisions_json" | jq empty >/dev/null 2>&1; then
  last_class="$(printf '%s' "$decisions_json" | jq -r '[.[] | .decision.class // empty] | last // empty')"
  [ -n "$last_class" ] && CLASS="$last_class"
fi

violations=()

base_list=()
orig_counts=()
retry_counts=()
base_task_ids=()   # parallel: the key's own `+`-joined task ids, unsplit

base_index() {
  local want="$1" i=0 b
  for b in ${base_list[@]+"${base_list[@]}"}; do
    [ "$b" = "$want" ] && { echo "$i"; return 0; }
    i=$(( i + 1 ))
  done
  return 1
}

while IFS= read -r row; do
  [ -n "$row" ] || continue
  key="$(printf '%s' "$row" | jq -r '.key // ""')"
  if [[ ! "$key" =~ $CANONICAL_KEY_RE ]]; then
    violations+=("out-of-shape gated-reviewer key '$key' -- the canonical shape is task-<n[+n+n...]>-reviewer[-fix-<k>][-retry]")
    continue
  fi
  if [[ "$key" == *-retry ]]; then
    base="${key%-retry}"
    field=retry_counts
  else
    base="$key"
    field=orig_counts
  fi
  ids_part="${base#task-}"
  ids_part="${ids_part%-reviewer*}"
  if idx="$(base_index "$base")"; then
    eval "$field[$idx]=\$(( \${$field[$idx]} + 1 ))"
  else
    base_list+=("$base")
    base_task_ids+=("$ids_part")
    if [ "$field" = retry_counts ]; then
      orig_counts+=(0); retry_counts+=(1)
    else
      orig_counts+=(1); retry_counts+=(0)
    fi
  fi
done < <(printf '%s\n' "$filtered")

for idx in "${!base_list[@]}"; do
  base="${base_list[$idx]}"
  orig="${orig_counts[$idx]}"
  retry="${retry_counts[$idx]}"
  total=$(( orig + retry ))
  if [ "$retry" -gt 0 ] && [ "$orig" -eq 0 ]; then
    violations+=("bundle $base carries $retry -retry dispatch(es) and no original -- a retry is never a bundle's only dispatch")
  elif [ "$orig" -gt 1 ]; then
    violations+=("bundle $base carries $orig reviewer dispatches -- the bundle goes to ONE dispatch, never one per task, slot or finding")
  elif [ "$total" -gt 2 ] || { [ "$total" -eq 2 ] && [ "$retry" -ne 1 ]; }; then
    violations+=("bundle $base carries $total reviewer dispatches -- at most the original plus the handshake's one -retry")
  fi
done

# ---- bundling-itself checks (never one dispatch per gate-fired task) -----
# Only ORIGINAL bundles (a fix round's own re-review key, `-fix-<k>`, is its
# own bundle by construction and is exempt: implement.md re-dispatches one
# bundle per group's fixed tasks, and two groups fixing in the same round
# are two legitimately separate re-review bundles, never a KAN-527 shape).

orig_base_idxs=()
for idx in "${!base_list[@]}"; do
  [ "${orig_counts[$idx]}" -ge 1 ] || continue
  [[ "${base_list[$idx]}" == *-fix-* ]] && continue
  orig_base_idxs+=("$idx")
done

# Same group, two separate original bundles: exactly the KAN-527 shape,
# regardless of class.
seen_groups=()
seen_group_bases=()
for idx in ${orig_base_idxs[@]+"${orig_base_idxs[@]}"}; do
  for tid in ${base_task_ids[$idx]//+/ }; do
    g="$(group_of_task "$tid" || true)"
    [ -n "$g" ] || continue
    gi=-1
    for j in "${!seen_groups[@]}"; do
      [ "${seen_groups[$j]}" = "$g" ] && { gi=$j; break; }
    done
    if [ "$gi" -ge 0 ]; then
      if [ "${seen_group_bases[$gi]}" != "${base_list[$idx]}" ]; then
        violations+=("tasks in group $g are split across reviewer bundles '${seen_group_bases[$gi]}' and '${base_list[$idx]}' -- every gate-fired task of one implementer group goes out in ONE reviewer dispatch")
      fi
    else
      seen_groups+=("$g")
      seen_group_bases+=("${base_list[$idx]}")
    fi
  done
done

# On `small`/`regular`, implement.md requires every gate-fired task of the
# whole run to join ONE bundle at the last boundary -- more than one
# original bundle is itself a violation on those classes.
if { [ "$CLASS" = "small" ] || [ "$CLASS" = "regular" ]; } && [ "${#orig_base_idxs[@]}" -gt 1 ]; then
  names=""
  for idx in "${orig_base_idxs[@]}"; do names="$names ${base_list[$idx]}"; done
  violations+=("class '$CLASS' carries ${#orig_base_idxs[@]} original reviewer bundles ($names) -- every gate-fired task of the run joins ONE bundle at the last boundary on small/regular")
fi

if [ "${#violations[@]}" -gt 0 ]; then
  for v in "${violations[@]}"; do
    echo "check-task-reviewer-single-dispatch: $v" >&2
  done
  exit 1
fi

echo "TASK-REVIEWER-SINGLE-DISPATCH-OK: every gated-per-task reviewer dispatch of this run is bundled by implementer group (bounds and retries included)"
