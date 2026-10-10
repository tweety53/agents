#!/usr/bin/env bash
# integrate-change.sh — the mechanical half of a bare /flow at IN_PROGRESS:
# run 1 (integrate) and run 2 (cleanup), sequenced through the existing
# guards so the parent agent does only what needs judgement — the commit
# subjects, the narrative paragraph, the self-review reasoning pass, the Jira
# calls, and every stop that needs the operator.
#
# skills/flow-contracts/finish-contract-run1.md and finish-contract-run2.md
# are canonical for every step this script runs, and each guard's own header
# is canonical for its verdicts: this script re-implements none of them. It
# calls them in the contracts' order, relays their lines unchanged, and stops
# where a verdict needs a decision. skills/flow/integrate.md and
# skills/flow/cleanup.md say what the parent does at each stop.
#
# Usage:
#   integrate-change.sh prepare <main-checkout> <name> [--accept-foreign] [--accept-outstanding]
#   integrate-change.sh commit  <main-checkout> <name> [<worktree>=<impl subject>]... [--plan-message <msg>]
#   integrate-change.sh land    <main-checkout> <name> [--route <route>] [--pr-url <worktree>=<url|none>]...
#   integrate-change.sh cleanup <main-checkout> <name> [--proceed <main-checkout>]...
# Every subcommand also takes --harness <harness> and --session-token
# <literal token> (else $FLOW_HARNESS and $FLOW_SESSION_TOKEN) for its stage
# marks; a mark the store cannot take is a warning on stderr, never a stop.
#
# prepare  resolves the worktree set (the state file's `worktrees` map, else
#          the porcelain scan of <main-checkout>), runs check-foreign-staged
#          and check-main-checkout-drift per distinct main checkout,
#          migrate-worktrees, then check-finish-preflight per worktree. Every
#          worktree RUN2 -> NEXT is `cleanup`. Otherwise run 1: the
#          uncommitted-archive undo, the archived decision, the unfinished-work
#          gate (check-unfinished-work, check-visual-verify-dispatched against
#          `git merge-base HEAD origin/$BASE`), check-base-moved and
#          sync-onto-base on MOVED (running each GUARD-TEST it names), and the
#          default landing route (project-get.sh). Prints per worktree its
#          reshape base and the diffstat the commit subjects are written from,
#          and records the reshape base in the worktree's own git dir
#          (`git rev-parse --git-path flow-reshape-base`) for `commit`.
#          --accept-foreign / --accept-outstanding: the operator chose
#          Continue at that stop; its lines still print.
# commit   reshape-branch per worktree, the ledger render, commit-split per
#          worktree in `## Merge order`, `spectre archive` (each
#          <name>-fix-N too) unless already archived, commit-archive. An
#          archived change without new work skips straight to commit-archive.
#          The parent appends its narrative before calling this.
# land     the route (merge and push: the --force-with-lease branch push then
#          `push origin HEAD:<base>` per worktree in merge order, one re-sync
#          and retry after a rejection; pull request: the branch push and
#          `gh pr create --fill`, or that worktree's --pr-url when there is
#          no usable gh; manual: the branch push), the self-review report's
#          sha re-read before each landing push and its `fixed` rows after,
#          and the IN_PROGRESS write. Refuses to push anything until the
#          self-review report is committed in the canonical worktree. Merge
#          and push continues into `cleanup`.
# cleanup  run 2: verify the merge (merged when gh reports the recorded PR
#          MERGED or `merge-base --is-ancestor` holds), remove-change-worktrees
#          per repository (--proceed passed only to each repository named by
#          one), the workspace `remove` row, the
#          proposal artifact, check-cleanup-complete per repository, the
#          FINISHED write (clearing every worktree no longer on disk) and
#          refresh-main-checkout per repository.
#
# Prints the guards' own lines, one verdict line per step of its own, and
# ends on exactly one machine-readable line:
#   NEXT: <what the parent does next>          exit 0
#   STOP: <kind> — <what the parent does>      exit 1
#   JIRA: transition <KEY> to <status>         before NEXT, when an issue is linked
#
# Exit 0 the subcommand ran to its end; 1 a stop — a guard verdict, a
# refusal or a failed git step needs the parent or the operator, and the
# stage mark open at that point is closed `stopped` (`not-run-2` and
# `leftover` where run 2's contract names those); 2 usage, a guard it
# sequences missing beside it (named on stderr, nothing run), or the state
# record cannot be read.
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/spec-root.sh
. "$SCRIPT_DIR/lib/spec-root.sh" || { echo "integrate-change: cannot load lib/spec-root.sh" >&2; exit 2; }

die() { printf 'integrate-change: %s\n' "$*" >&2; exit 2; }
say() { printf '%s\n' "$*"; }

missing=""
for g in check-base-moved check-cleanup-complete check-finish-preflight check-foreign-staged \
  check-main-checkout-drift check-unfinished-work check-visual-verify-dispatched commit-archive \
  commit-split land-self-review-report migrate-worktrees project-get refresh-main-checkout \
  remove-change-worktrees reshape-branch resolve-base-branch sync-onto-base; do
  [ -x "$SCRIPT_DIR/$g.sh" ] || missing="$missing $g.sh"
done
[ -z "$missing" ] || die "missing beside this script:$missing — run the procedure by hand"

SUB="${1:-}"; MAIN="${2:-}"; NAME="${3:-}"
[ -n "$SUB" ] && [ -n "$MAIN" ] && [ -n "$NAME" ] || die "usage: integrate-change.sh prepare|commit|land|cleanup <main-checkout> <name> [flags]"
shift 3
MAIN="$(cd "$MAIN" 2>/dev/null && pwd -P)" || die "$MAIN is not a directory"
case "$NAME" in *[!a-z0-9._-]*|-*|'') die "change name '$NAME' is not a plain change name" ;; esac

HARNESS="${FLOW_HARNESS:-}"; TOKEN="${FLOW_SESSION_TOKEN:-}"
ACCEPT_FOREIGN=""; ACCEPT_OUTSTANDING=""; PLAN_MSG="chore(spectre): plan $NAME"
ROUTE=""; PR_ARGS=(); PROCEED=(); SUBJECTS=()
while [ $# -gt 0 ]; do
  case "$1" in
    --harness) HARNESS="${2:-}"; shift ;;
    --session-token) TOKEN="${2:-}"; shift ;;
    --accept-foreign) ACCEPT_FOREIGN=1 ;;
    --accept-outstanding) ACCEPT_OUTSTANDING=1 ;;
    --plan-message) PLAN_MSG="${2:-}"; shift ;;
    --route) ROUTE="${2:-}"; shift ;;
    --pr-url) case "${2:-}" in ?*=?*) PR_ARGS+=("$2") ;; *) die "--pr-url takes <worktree>=<url|none>" ;; esac; shift ;;
    --proceed)
      p="$(cd "${2:-}" 2>/dev/null && pwd -P)" || die "--proceed takes the main checkout its disclosure stop named"
      PROCEED+=("$p"); shift ;;
    *=*) SUBJECTS+=("$1") ;;
    *) die "unknown argument '$1'" ;;
  esac
  shift
done
[ -n "$HARNESS" ] && [ -n "$TOKEN" ] || die "--harness and --session-token are required (or FLOW_HARNESS / FLOW_SESSION_TOKEN)"
FLAGS="--harness $HARNESS --session-token $TOKEN"
# Outside every worktree: remove-change-worktrees' check 6 counts this
# process's own cwd.
cd / || die "cannot cd /"

OPEN_STAGE=""
mark_begin() {
  OPEN_STAGE="$1"
  flow stage begin -C "$MAIN" -command /flow -stage "$1" -harness "$HARNESS" -session-token "$TOKEN" "$NAME" >/dev/null ||
    echo "integrate-change: warning: flow stage begin $1 failed" >&2
}
mark_end() {
  flow stage end -C "$MAIN" -command /flow -stage "$1" -outcome "$2" "$NAME" >/dev/null ||
    echo "integrate-change: warning: flow stage end $1 failed" >&2
  OPEN_STAGE=""
}
STOP_OUTCOME=stopped
stop() {
  [ -n "$OPEN_STAGE" ] && mark_end "$OPEN_STAGE" "$STOP_OUTCOME"
  say "STOP: $1 — $2"
  exit 1
}

# run <guard> <args...>: OUT its stdout, relayed; RC its exit code. Its
# stderr goes straight through. Each guard is spelled $SCRIPT_DIR/<name> at
# its call, where check-guard-symlinks rule 2 reads a sibling.
OUT=""; RC=0
run() {
  OUT="$("$@")"; RC=$?
  [ -n "$OUT" ] && printf '%s\n' "$OUT"
  return 0
}
verdict() { grep -m1 -E "^($1):" <<<"$OUT" || true; }

main_of() {
  local d
  d="$(git -C "$1" rev-parse --path-format=absolute --git-common-dir 2>/dev/null)" || return 1
  (cd "$d/.." && pwd -P)
}

# main_of_gone <worktree>: the main checkout of a worktree no longer on
# disk, derived from the layout every worktree is created in
# (<main>-worktrees/<name>, git-boundaries.md) — accepted only when it is the
# top of a git repository that is not itself a linked worktree.
main_of_gone() {
  local p c d
  p="$(dirname -- "$1")"
  case "$p" in *?-worktrees) ;; *) return 1 ;; esac
  c="$(cd "${p%-worktrees}" 2>/dev/null && pwd -P)" || return 1
  d="$(git -C "$c" rev-parse --path-format=absolute --git-dir 2>/dev/null)" || return 1
  [ "$d" = "$(git -C "$c" rev-parse --path-format=absolute --git-common-dir)" ] || return 1
  [ "$(git -C "$c" rev-parse --show-toplevel)" = "$c" ] || return 1
  printf '%s\n' "$c"
}

# The worktree set: WT[i] its path, MB[i] its recorded merge base (- when
# none), REPO[i] its main checkout (main_of_gone's for one already gone,
# empty when that cannot be derived), MAINS the distinct main checkouts,
# <main-checkout> always among them.
load_set() {
  local p m state
  state="$(flow state get -C "$MAIN" "$NAME")" || die "flow state get $NAME failed"
  PR_URL_STATE="$(printf '%s' "$state" | jq -r '.prUrl // empty')"
  JIRA_KEY="$(printf '%s' "$state" | jq -r '.jiraIssue // empty')"
  WT=(); MB=(); REPO=(); MAINS=("$MAIN")
  while IFS=$'\t' read -r p m; do
    [ -n "$p" ] && WT+=("$p") && MB+=("$m")
  done < <(printf '%s' "$state" | jq -r '(.worktrees // {}) | to_entries[] | "\(.key)\t\(.value // "-")"')
  if [ ${#WT[@]} -eq 0 ]; then
    while IFS= read -r p; do
      [ -n "$p" ] && WT+=("$p") && MB+=("-")
    done < <(git -C "$MAIN" worktree list --porcelain |
      awk -v b="refs/heads/spectre/$NAME" '/^worktree /{w=substr($0, 10)} /^branch /{if ($2==b) print w}')
  fi
  local i r k seen
  for i in "${!WT[@]}"; do
    r="$(main_of "${WT[i]}" 2>/dev/null)" || r="$(main_of_gone "${WT[i]}")" || r=""
    REPO+=("$r")
    [ -n "$r" ] || continue
    seen=""
    for k in "${MAINS[@]}"; do [ "$k" = "$r" ] && seen=1; done
    [ -n "$seen" ] || MAINS+=("$r")
  done
}

# CANON: the worktree whose change directory holds tasks.md; SPEC its tree.
find_canonical() {
  local w s
  CANON=""; SPEC=""
  for w in "${WT[@]}"; do
    [ -d "$w" ] || continue
    s="$(spec_root_leaf "$w" 2>/dev/null)" || continue
    if [ -f "$w/$s/changes/$NAME/tasks.md" ] || [ -f "$w/$s/changes/archive/$NAME/tasks.md" ]; then
      CANON="$w"; SPEC="$s"; return 0
    fi
  done
  return 1
}
idx_of() { local i; for i in "${!WT[@]}"; do [ "${WT[i]}" = "$1" ] && { printf '%s\n' "$i"; return 0; }; done; return 1; }

# ORDER: worktree indexes in the canonical link.md's `## Merge order` (`.`
# the canonical repository, a peer through <spec>/peers), every worktree it
# does not name after them in map order.
merge_order() {
  local link item rel root i j have
  ORDER=()
  link="$CANON/$SPEC/changes/$NAME/link.md"
  [ -f "$link" ] || link="$CANON/$SPEC/changes/archive/$NAME/link.md"
  if [ -f "$link" ]; then
    while IFS= read -r item; do
      if [ "$item" = "." ]; then
        root="$(main_of "$CANON")"
      else
        rel="$(awk -v n="$item" '$1==n {print $2; exit}' "$CANON/$SPEC/peers" 2>/dev/null)"
        [ -n "$rel" ] || continue
        root="$(cd "$(main_of "$CANON")/$rel" 2>/dev/null && pwd -P)" || continue
      fi
      for i in "${!WT[@]}"; do
        [ "${REPO[i]}" = "$root" ] && ORDER+=("$i")
      done
    done < <(awk '/^## /{on=($0=="## Merge order")} on && /^[0-9]+\. `/ {s=$0; sub(/^[^`]*`/,"",s); sub(/`.*/,"",s); print s}' "$link")
  fi
  for i in "${!WT[@]}"; do
    have=""
    for j in ${ORDER[@]+"${ORDER[@]}"}; do [ "$j" = "$i" ] && have=1; done
    [ -n "$have" ] || ORDER+=("$i")
  done
}

# base_of <i>: BASE the worktree's base branch, resolve-base-branch.sh's
# answer (it fetches).
base_of() {
  BASE="$("$SCRIPT_DIR/resolve-base-branch.sh" "${WT[$1]}")"; local rc=$?
  [ "$rc" -eq 3 ] && stop no-remote "${WT[$1]}: this repository has no remote, so there is nothing to push to or merge into"
  [ "$rc" -eq 0 ] || stop base "${WT[$1]}: no base branch resolved (resolve-base-branch exit $rc) — ask the operator"
}

# arch_base <worktree> <merge-base>: an archived re-run's base.
arch_base() {
  local s
  s="$(git -C "$1" log --format='%H %s' "$2..HEAD" | awk -v a="docs(self-review): $NAME self-review report" \
    -v b="docs(self-review): $NAME self-review context bundle" '{h=$1; sub(/^[^ ]+ /,""); if ($0==a || $0==b) {print h; exit}}')"
  [ -n "$s" ] || s="$(git -C "$1" log --format='%H %s' "$2..HEAD" |
    awk -v a="chore(spectre): archive $NAME" '{h=$1; sub(/^[^ ]+ /,""); if ($0==a) {print h; exit}}')"
  printf '%s\n' "${s:-$2}"
}

reshape_file() { git -C "$1" rev-parse --path-format=absolute --git-path flow-reshape-base; }

write_state() { # <state> [<keep-json-array>] [<pr-url>]
  local j
  j="$(flow state get -C "$MAIN" "$NAME")" || return 1
  j="$(printf '%s' "$j" | jq --arg s "$1" --argjson keep "${2:-null}" --arg pr "${3:-}" '
    .state = $s | .updatedBy = "/flow"
    | (if $pr != "" then .prUrl = $pr else . end)
    | (if $keep != null then .worktrees |= with_entries(select(.key as $k | any($keep[]; . == $k))) else . end)')" || return 1
  printf '%s' "$j" | flow state set -C "$MAIN" "$NAME"
}

# guard_tests: run each GUARD-TEST the last sync-onto-base named; a failure
# stops with the test's own output above the stop line.
guard_tests() {
  local t out
  while IFS= read -r t; do
    [ -n "$t" ] || continue
    t="${t##* — }"
    out="$(bash "$t" 2>&1)" || { printf '%s\n' "$out"; stop guard-test "the scoped re-verification $t failed — report its output above; the change stays IN_PROGRESS"; }
    say "GUARD-TEST-PASSED: $t"
  done < <(printf '%s\n' "$OUT" | grep '^GUARD-TEST: ')
}

cmd_prepare() {
  local i m v ask=0 run1=0 rb vline sha t
  load_set
  [ ${#WT[@]} -gt 0 ] || stop preflight "the resolved worktree set is empty — report it and ask the operator"

  for m in "${MAINS[@]}"; do
    run "$SCRIPT_DIR/check-foreign-staged.sh" "$m"
    { [ "$RC" -eq 0 ] && [ -n "$(verdict STAGED-CLEAN)" ]; } || ask=1
    run "$SCRIPT_DIR/check-main-checkout-drift.sh" "$m"
    { [ "$RC" -eq 0 ] && [ -n "$(verdict DRIFT-CLEAN)" ]; } || ask=1
  done
  [ "$ask" -eq 1 ] && [ -z "$ACCEPT_FOREIGN" ] &&
    stop foreign-staged "ask the operator the foreign-staged-work question; Continue re-runs prepare with --accept-foreign"

  run "$SCRIPT_DIR/migrate-worktrees.sh" "${MAINS[@]}"
  [ "$RC" -eq 0 ] || stop migrate "migrate-worktrees exit $RC — report its lines"
  load_set

  for i in "${!WT[@]}"; do
    [ -d "${WT[i]}" ] || stop preflight "${WT[i]} is not on disk — report it and ask the operator"
    base_of "$i"; BASES[i]="$BASE"
    run "$SCRIPT_DIR/check-finish-preflight.sh" "${WT[i]}" "origin/$BASE" "${MB[i]}"
    v="$(verdict 'RUN1|RUN2|REFUSE')"
    case "$v" in
      RUN1:*) run1=1 ;;
      RUN2:*) ;;
      *) stop preflight "relay the verdict and its hand-verification procedure, then ask the operator" ;;
    esac
  done
  if [ "$run1" -eq 0 ]; then
    say "NEXT: run 2 — integrate-change.sh cleanup $MAIN $NAME $FLAGS"
    exit 0
  fi

  mark_begin flow.preflight; mark_end flow.preflight completed
  mark_begin flow.unfinished-work-gate
  find_canonical || stop preflight "no worktree in the set holds $NAME's tasks.md — ask the operator"
  local ci live arch d newwork=0
  ci="$(idx_of "$CANON")"
  live="$CANON/$SPEC/changes/$NAME"; arch="$CANON/$SPEC/changes/archive/$NAME"
  if [ -d "$arch" ] && ! grep -qxF "chore(spectre): archive $NAME" <<<"$(git -C "$CANON" log --format=%s "${MB[ci]}..HEAD" 2>/dev/null)"; then
    [ -d "$live" ] && stop archive-undo "both $live and $arch exist — ask the operator to check the live copy holds nothing the archived one lacks, remove it, and re-run"
    for d in "$arch" "$arch"-fix-[0-9]*; do
      [ -d "$d" ] || continue
      git -C "$CANON" mv "${d#"$CANON/"}" "$SPEC/changes/${d##*/}" || stop archive-undo "git mv of $d failed"
      say "UNARCHIVED: $d"
    done
  fi
  ARCHIVED=0
  [ -d "$arch" ] && [ ! -d "$live" ] && ARCHIVED=1
  if [ "$ARCHIVED" -eq 1 ]; then
    [ "$(git -C "$CANON" rev-parse HEAD)" != "$(git -C "$CANON" rev-parse "$(arch_base "$CANON" "${MB[ci]}")")" ] && newwork=1
    for i in "${!WT[@]}"; do
      [ -n "$(git -C "${WT[i]}" status --porcelain --untracked-files=no)" ] && newwork=1
    done
    say "ARCHIVED: $NAME — $([ "$newwork" -eq 1 ] && echo 'with new work' || echo 'without new work')"
  fi

  if [ "$ARCHIVED" -eq 1 ] && [ "$newwork" -eq 0 ]; then
    mark_end flow.unfinished-work-gate completed
  else
    ask=0
    # A guard that reached no verdict is not OUTSTANDING: its own stop,
    # which --accept-outstanding never passes.
    for i in "${!WT[@]}"; do
      run "$SCRIPT_DIR/check-unfinished-work.sh" "${WT[i]}" "$NAME" "$CANON"
      case "$RC:$(verdict 'CLEAR|OUTSTANDING')" in
        0:CLEAR:*) ;;
        0:OUTSTANDING:*) ask=1 ;;
        *) stop no-verdict "check-unfinished-work reached no verdict for ${WT[i]} (exit $RC) — report its output and ask the operator; --accept-outstanding does not pass this" ;;
      esac
      run "$SCRIPT_DIR/check-visual-verify-dispatched.sh" "${WT[i]}" "$NAME" "$(git -C "${WT[i]}" merge-base HEAD "origin/${BASES[i]}")"
      case "$RC:$(verdict 'VISUAL-VERIFY-OK|VISUAL-VERIFY-MISSING')" in
        0:VISUAL-VERIFY-OK:*) ;;
        1:VISUAL-VERIFY-MISSING:*) ask=1 ;;
        *) stop no-verdict "check-visual-verify-dispatched reached no verdict for ${WT[i]} (exit $RC) — report its output and ask the operator; --accept-outstanding does not pass this" ;;
      esac
    done
    [ "$ask" -eq 1 ] && [ -z "$ACCEPT_OUTSTANDING" ] &&
      stop unfinished-work "load skills/flow/unfinished-work-gate.md and offer its three courses; Continue or File-or-join re-runs prepare with --accept-outstanding"
    mark_end flow.unfinished-work-gate completed
  fi

  mark_begin flow.landing-question
  local moved=()
  for i in "${!WT[@]}"; do
    run "$SCRIPT_DIR/check-base-moved.sh" "${WT[i]}" "origin/${BASES[i]}" "${MB[i]}"
    v="$(verdict 'CLEAR|MOVED|REFUSE')"
    case "$v" in
      CLEAR:*merge\ base\ now\ *) RB[i]="${v##*merge base now }" ;;
      CLEAR:*) RB[i]="${MB[i]}" ;;
      MOVED:*) moved+=("$i") ;;
      *) stop base-moved "report the verdict and ask the operator" ;;
    esac
  done
  for i in ${moved[@]+"${moved[@]}"}; do
    run "$SCRIPT_DIR/sync-onto-base.sh" "${WT[i]}" "${BASES[i]}" "${MB[i]}"
    case "$RC" in
      0) ;;
      1) stop sync-conflict "load skills/flow/sync-onto-base.md and resolve in place, run '$SCRIPT_DIR/sync-onto-base.sh' --resume, the whole ## lint and ## test lists, then re-run prepare" ;;
      *) stop sync "sync-onto-base exit $RC — report its output and ask the operator" ;;
    esac
    vline="$(verdict REBASED)"; sha="${vline##*onto }"
    RB[i]="$sha"
    guard_tests
  done

  local route=""
  route="$("$SCRIPT_DIR/project-get.sh" "$MAIN" "default landing route" --enum "pull request" "merge and push" manual)" || route=""
  if [ -n "$route" ]; then
    say "ROUTE: $route — from this project's configured default, not asked"
  else
    say "ROUTE: ask — no configured default; ask the landing question, then pass --route to land"
  fi
  [ -n "$PR_URL_STATE" ] && say "PR: $PR_URL_STATE"
  mark_end flow.landing-question completed

  local reshape=0 f
  for i in "${!WT[@]}"; do
    f="$(reshape_file "${WT[i]}")"; rm -f "$f"
    [ "$ARCHIVED" -eq 1 ] && [ "$newwork" -eq 0 ] && continue
    rb="${RB[i]}"
    [ "$ARCHIVED" -eq 1 ] && rb="$(arch_base "${WT[i]}" "$rb")"
    printf '%s\n' "$rb" >"$f"; reshape=1
    say "WORKTREE: ${WT[i]} — base ${BASES[i]}, reshape base $rb"
    git -C "${WT[i]}" diff --no-renames --stat "$rb" -- . ":(exclude)$SPEC/changes/"
    git -C "${WT[i]}" ls-files --others --exclude-standard -- . ":(exclude)$SPEC/changes/" | sed 's/^/ new: /'
  done
  local narrative="$live/narrative.md"
  [ "$ARCHIVED" -eq 1 ] && narrative="$arch/narrative.md"
  if [ "$reshape" -eq 1 ]; then
    mark_begin flow.preserve-sessions
    say "NEXT: append this run's narrative to $narrative, then integrate-change.sh commit $MAIN $NAME <worktree>=<impl subject>... $FLAGS"
  else
    say "NEXT: integrate-change.sh commit $MAIN $NAME $FLAGS"
  fi
}

cmd_commit() {
  local i s subj f rb reshape=0 o
  load_set
  find_canonical || die "no worktree in the set holds $NAME's tasks.md"
  merge_order
  for i in "${!WT[@]}"; do [ -f "$(reshape_file "${WT[i]}")" ] && reshape=1; done
  if [ "$reshape" -eq 1 ]; then
    OPEN_STAGE=flow.preserve-sessions
    for o in "${ORDER[@]}"; do
      f="$(reshape_file "${WT[o]}")"; [ -f "$f" ] || continue
      rb="$(cat "$f")"
      run "$SCRIPT_DIR/reshape-branch.sh" "${WT[o]}" "$NAME" "$rb"
      [ "$RC" -eq 0 ] || stop reshape "reshape-branch exit $RC in ${WT[o]} — report its lines"
    done
    flow record render -C "$MAIN" -change "$NAME" -kind ledger -repo "$CANON" ||
      say "LEDGER-RENDER-FAILED: a destination was refused or could not be written — report it"
    mark_end flow.preserve-sessions completed
    mark_begin flow.commit-two
    for o in "${ORDER[@]}"; do
      f="$(reshape_file "${WT[o]}")"; [ -f "$f" ] || continue
      subj=""
      for s in ${SUBJECTS[@]+"${SUBJECTS[@]}"}; do
        case "${s%%=*}" in "${WT[o]}"|"${REPO[o]}") subj="${s#*=}" ;; esac
      done
      [ -n "$subj" ] || stop commit-split "no implementation subject given for ${WT[o]} — pass ${WT[o]}=<subject>"
      run "$SCRIPT_DIR/commit-split.sh" "${WT[o]}" "$NAME" "$subj" "$PLAN_MSG"
      [ "$RC" -eq 0 ] || stop commit-split "commit-split exit $RC in ${WT[o]} — report its lines"
      rm -f "$f"
    done
    mark_end flow.commit-two completed
  fi

  if [ ! -d "$CANON/$SPEC/changes/archive/$NAME" ] || [ -d "$CANON/$SPEC/changes/$NAME" ]; then
    mark_begin flow.sync-archive
    local d
    for d in "$CANON/$SPEC/changes/$NAME" "$CANON/$SPEC/changes/$NAME"-fix-[0-9]*; do
      [ -d "$d" ] || continue
      (cd "$CANON" && spectre archive "${d##*/}") || stop archive "spectre archive ${d##*/} refused — report it; never --force"
    done
    mark_end flow.sync-archive completed
  fi
  mark_begin flow.commit-archive
  run "$SCRIPT_DIR/commit-archive.sh" "$CANON" "$NAME"
  [ "$RC" -eq 0 ] || stop commit-archive "commit-archive exit $RC — report its lines"
  mark_end flow.commit-archive completed
  mark_begin flow.self-review
  if git -C "$CANON" cat-file -e "HEAD:docs/self-review/$NAME-self-review.md" 2>/dev/null; then
    say "NEXT: the self-review report is already committed — skip the pass; integrate-change.sh land $MAIN $NAME $FLAGS"
  else
    say "NEXT: run the self-review pass over \`flow self-review bundle -C $CANON -change $NAME\` and land its report in $CANON, then integrate-change.sh land $MAIN $NAME $FLAGS"
  fi
}

# reread_fix_shas <i>: before a landing push from the canonical worktree,
# rewrite each report `fixed:` sha a rebase changed, found by its subject,
# and commit the rewrite (step 5 of /flow-self-review).
reread_fix_shas() {
  local w="${WT[$1]}" rpt="docs/self-review/$NAME-self-review.md" old subj new changed=0 tmp
  [ "$w" = "$CANON" ] && [ -f "$w/$rpt" ] || return 0
  while IFS= read -r old; do
    git -C "$w" merge-base --is-ancestor "$old" HEAD 2>/dev/null && continue
    subj="$(git -C "$w" log -1 --format=%s "$old" 2>/dev/null)" || continue
    new="$(git -C "$w" log --format='%h %s' HEAD --not "origin/$BASE" |
      awk -v s="$subj" '{h=$1; sub(/^[^ ]+ /,""); if ($0==s) {print h; exit}}')"
    [ -n "$new" ] || continue
    tmp="$(mktemp)"
    sed "s/fixed: $old/fixed: $new/g" "$w/$rpt" >"$tmp" && mv "$tmp" "$w/$rpt"
    say "FIXED-SHA: $old -> $new"; changed=1
  done < <(grep -o 'fixed: [0-9a-f]\{7,40\}' "$w/$rpt" | awk '{print $2}' | sort -u)
  [ "$changed" -eq 0 ] && return 0
  run "$SCRIPT_DIR/land-self-review-report.sh" "$w" "spectre/$NAME" "docs(self-review): $NAME self-review report" "$rpt"
  [ "$RC" -eq 0 ] || stop self-review-report "land-self-review-report exit $RC — report its line"
}

# record_fixed_rows: once the landing push succeeded, one `fixed` row per
# report line whose sha this repository carries and the store does not list.
record_fixed_rows() {
  local rpt="$CANON/docs/self-review/$NAME-self-review.md" line known re
  [ -f "$rpt" ] || return 0
  known="$(flow self-review findings -C "$MAIN" -change "$NAME" 2>/dev/null)"
  re='^- \*\*\[([a-z-]+)\]\*\* (.*) — fixed: ([0-9a-f]+)'
  while IFS= read -r line; do
    [[ "$line" =~ $re ]] || continue
    git -C "$CANON" cat-file -e "${BASH_REMATCH[3]}^{commit}" 2>/dev/null || continue
    grep -q "${BASH_REMATCH[3]}" <<<"$known" && continue
    flow self-review finding -C "$MAIN" -change "$NAME" -angle "${BASH_REMATCH[1]}" -disposition fixed \
      -ref "${BASH_REMATCH[3]}" -note "${BASH_REMATCH[2]}" ||
      echo "integrate-change: warning: recording the fixed row for ${BASH_REMATCH[3]} failed" >&2
  done <"$rpt"
}

push_branch() { git -C "$1" push --force-with-lease -u origin "spectre/$NAME" || stop push "the branch push from $1 was rejected — report git's output; the change stays IN_PROGRESS"; }

# pr_arg <i>: the --pr-url answer given for worktree <i>, by its path or
# its main checkout.
pr_arg() {
  local s
  for s in ${PR_ARGS[@]+"${PR_ARGS[@]}"}; do
    case "${s%%=*}" in "${WT[$1]}"|"${REPO[$1]}") printf '%s\n' "${s#*=}"; return 0 ;; esac
  done
  return 1
}

cmd_land() {
  local o w pr="" url fork
  load_set
  find_canonical || die "no worktree in the set holds $NAME's tasks.md"
  merge_order
  git -C "$CANON" cat-file -e "HEAD:docs/self-review/$NAME-self-review.md" 2>/dev/null ||
    stop self-review-report "docs/self-review/$NAME-self-review.md is not committed in $CANON — run the self-review pass and land its report, then re-run land; nothing was pushed"
  [ -n "$ROUTE" ] || ROUTE="$("$SCRIPT_DIR/project-get.sh" "$MAIN" "default landing route" --enum "pull request" "merge and push" manual)" || ROUTE=""
  case "$ROUTE" in
    "pull request"|"merge and push"|manual) ;;
    *) stop landing-route "no route — ask the landing question, then integrate-change.sh land $MAIN $NAME --route <route> $FLAGS" ;;
  esac
  mark_end flow.self-review completed
  mark_begin flow.landing-routes
  for o in "${ORDER[@]}"; do
    w="${WT[o]}"
    base_of "$o"
    reread_fix_shas "$o"
    push_branch "$w"
    case "$ROUTE" in
      "merge and push")
        if ! git -C "$w" push origin "HEAD:$BASE"; then
          say "PUSH-REJECTED: $w — re-syncing onto origin/$BASE once"
          base_of "$o"
          fork="$(git -C "$w" merge-base HEAD "origin/$BASE")"
          run "$SCRIPT_DIR/sync-onto-base.sh" "$w" "$BASE" "$fork"
          case "$RC" in
            0) ;;
            1) stop sync-conflict "load skills/flow/sync-onto-base.md, resolve in place, --resume, run the whole ## lint and ## test lists, then re-run land — that push is the one retry" ;;
            *) stop sync "sync-onto-base exit $RC — report its output" ;;
          esac
          guard_tests
          reread_fix_shas "$o"
          push_branch "$w"
          git -C "$w" push origin "HEAD:$BASE" || stop push "the second push of $w to origin/$BASE was rejected — report git's output; the change stays IN_PROGRESS"
        fi
        say "PUSHED: $w — spectre/$NAME onto origin/$BASE"
        ;;
      "pull request")
        if [ -n "$PR_URL_STATE" ]; then
          pr="$PR_URL_STATE"; say "PR: $pr — updated by the push"
        elif url="$(pr_arg "$o")" || { url="$(cd "$w" && gh pr create --fill --base "$BASE" --head "spectre/$NAME" 2>&1 | tail -1)" && [ "${url#http}" != "$url" ]; }; then
          [ "$url" = none ] && continue
          say "PR: $url"
          { [ "$w" = "$CANON" ] || [ -z "$pr" ]; } && pr="$url"
        else
          stop pr "no usable gh for $w ($url) — print the forge's create-PR URL, ask whether it was opened, then re-run land adding --pr-url $w=<url|none> to every --pr-url already given"
        fi
        ;;
      manual) say "PUSHED: $w — spectre/$NAME; the rest is the operator's" ;;
    esac
  done
  record_fixed_rows
  write_state IN_PROGRESS "" "$pr" || stop state "the IN_PROGRESS write failed — report it"
  mark_end flow.landing-routes completed
  [ -n "$JIRA_KEY" ] && say "JIRA: transition $JIRA_KEY to In Review"
  if [ "$ROUTE" = "merge and push" ]; then
    cmd_cleanup
  else
    say "NEXT: print the run 1 handoff — waiting on the merge"
  fi
}

cmd_cleanup() {
  local i m mb out rc row id sd v leftover=0 keep base p pf
  load_set
  mark_begin flow.verify-merge
  for i in "${!WT[@]}"; do
    [ -d "${WT[i]}" ] || continue
    base_of "$i"
    v=""
    if [ -n "$PR_URL_STATE" ] && command -v gh >/dev/null; then
      v="$(gh pr view "$PR_URL_STATE" --json state -q .state 2>/dev/null)" || v=""
    fi
    [ "$v" = MERGED ] && continue
    git -C "${WT[i]}" merge-base --is-ancestor HEAD "origin/$BASE" && continue
    STOP_OUTCOME=not-run-2
    stop not-merged "${WT[i]} is not merged into origin/$BASE — this is not run 2; remove nothing and run prepare"
  done
  mark_end flow.verify-merge completed

  mark_begin flow.cleanup
  for m in "${MAINS[@]}"; do
    mb=-
    for i in "${!WT[@]}"; do [ "${REPO[i]}" = "$m" ] && mb="${MB[i]}" && break; done
    pf=""
    for p in ${PROCEED[@]+"${PROCEED[@]}"}; do [ "$p" = "$m" ] && pf=--proceed; done
    run "$SCRIPT_DIR/remove-change-worktrees.sh" "$m" "$NAME" "$mb" $pf
    case "$RC" in
      0) ;;
      3) stop disclose "relay $m's UNCLASSIFIED and DISCLOSE lines, ask only what they ask, then re-run integrate-change.sh cleanup $MAIN $NAME adding --proceed $m to every --proceed already given $FLAGS" ;;
      *) stop worktree-cleanup "remove-change-worktrees exit $RC for $m — report its lines; every worktree it refused is left alone" ;;
    esac
  done
  id="$(flow workspace-id "$NAME")" || id=""
  for m in "${MAINS[@]}"; do
    row="$("$SCRIPT_DIR/project-get.sh" "$m" "workspace isolation" 2>/dev/null |
      awk -F'|' '{c=$2; gsub(/[ `]/,"",c); if (c=="remove") {r=$3; sub(/^[ ]*`/,"",r); sub(/`[ ]*$/,"",r); print r; exit}}')"
    if [ -z "$row" ] || [ -z "$id" ]; then
      say "WORKSPACE: $m — skipped, no remove command declared"
      continue
    fi
    row="${row//<id_underscored>/${id//-/_}}"; row="${row//<id>/$id}"
    (cd "$m" && bash -c "$row"); rc=$?
    if [ "$rc" -eq 0 ]; then say "WORKSPACE-REMOVED: $m"; else say "WORKSPACE-REMOVE-FAILED: $m — exit $rc; the cleanup check decides"; fi
  done
  sd="$(flow state dir -C "$MAIN")"
  if [ -f "$sd/$NAME-proposal-artifact.html" ]; then
    rm -f "$sd/$NAME-proposal-artifact.html"; say "PROPOSAL-ARTIFACT: removed"
  else
    say "PROPOSAL-ARTIFACT: none to remove"
  fi
  mark_end flow.cleanup completed

  mark_begin flow.verify-cleanup
  for m in "${MAINS[@]}"; do
    run "$SCRIPT_DIR/check-cleanup-complete.sh" "$m" "$NAME" "$(flow state dir -C "$m")"
    [ "$RC" -eq 0 ] && [ -n "$(verdict COMPLETE)" ] || leftover=1
  done
  STOP_OUTCOME=leftover
  [ "$leftover" -eq 1 ] && stop leftover "name what remains; FINISHED is not written — re-run cleanup once the operator clears it"
  STOP_OUTCOME=stopped
  mark_end flow.verify-cleanup completed

  mark_begin flow.write-finished
  keep="$(for i in "${!WT[@]}"; do [ -d "${WT[i]}" ] && printf '%s\n' "${WT[i]}"; done | jq -R . | jq -s -c .)"
  write_state FINISHED "$keep" || stop state "the FINISHED write failed — report it"
  mark_end flow.write-finished completed
  [ -n "$JIRA_KEY" ] && say "JIRA: transition $JIRA_KEY to Done"

  mark_begin flow.refresh-main-checkout
  for m in "${MAINS[@]}"; do
    base="$(git -C "$m" symbolic-ref --short refs/remotes/origin/HEAD 2>/dev/null)"
    run "$SCRIPT_DIR/refresh-main-checkout.sh" "$m" "${base#origin/}"
  done
  mark_end flow.refresh-main-checkout completed
  say "NEXT: print the Finished handoff"
}

case "$SUB" in
  prepare) cmd_prepare ;;
  commit) cmd_commit ;;
  land) cmd_land ;;
  cleanup) cmd_cleanup ;;
  *) die "unknown subcommand '$SUB'" ;;
esac
