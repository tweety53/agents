#!/usr/bin/env bash
# run-reproducer.sh <worktree> <reproducer-command-line>
#
# Runs one finding's reproducer — the command text that follows
# `finding-reproducer: F<n> ` in a panel record, already validated
# lexically by check-panel-reproducers.sh — and reports whether it
# demonstrates the defect it names.
#
# WHY THIS SCRIPT EXISTS. Every rule about *running* a reproducer used to be
# prose in skills/myflow-do/SKILL.md, with no code behind it: the
# containment of each token, the ban on shell metacharacters, the direct
# exec, the bound, and the timed-out and surviving-process dispositions.
# Eight review findings across three panel passes were defects in that
# prose, because a rule nothing implements can only be re-read rather than
# tested — most plainly F44, whose fix for an earlier finding cited
# `ps -o sid=`, a keyword this platform's `ps` does not have:
#
#   $ ps -o sid= -p $$
#   ps: sid: keyword not found
#
# This script turns those rules into something with its own test harness,
# so the constraints are enforced where they were only described.
#
# THE COMMAND IS DISPATCHED AS AN ARGUMENT VECTOR, NEVER A STRING THROUGH A
# SHELL. The path token is resolved to a real file on disk and executed
# directly, with its arguments passed as a bash array — `exec "$path"
# "${args[@]}"` inside a subshell — so nothing in the reproducer's own text
# is ever handed to a shell for interpretation. That is the guarantee the
# prose this script replaces claimed and had nothing behind; the
# shell-metacharacter checks below still run first, but they are now the
# second layer of a real argv-exec barrier rather than the only one.
#
# CONTAINMENT IS RESOLVED, NOT MERELY LEXICAL. check-panel-reproducers.sh
# checks a panel record's reproducer lines lexically only — no leading `-`
# on the path token, no absolute token, no `..` path segment, no shell
# metacharacter, no URL — because that guard's record may be read in a
# context where the worktree that wrote it does not exist to resolve
# against. This script always has a real worktree, so it repeats those
# lexical checks (a record can be edited after the guard last ran) and then
# goes further: every token is resolved to its physical path — following
# `..`, `.` and symlinks — and required to stay inside the worktree. A
# relative path with no `..` segment can still point outside the worktree
# through a symlink, which is exactly the shape the record-format guard's
# own header says it deliberately cannot decide; this script is where that
# shape is decided.
#
# Exit codes:
#   0  defect demonstrated — the command ran to completion inside the
#      bound, as a direct exec, and exited non-zero
#   1  defect not demonstrated — the command ran to completion inside the
#      bound and exited 0; the instruction built on this reproducer cannot
#      be verified as a fix
#   2  refused before execution — the command line failed a lexical or
#      resolved containment/shape check and was never run at all; the
#      reason is named on stderr
#   3  unverifiable — the command was still running at the bound (or a
#      detached child of it survived the kill sequence) and was killed;
#      its exit status is read as neither a pass nor a fail. A surviving
#      child is additionally named on stderr with its pid.
#   4  cannot answer at all — bad usage, no such worktree, an
#      environment failure (no writable temporary directory, and similar), or
#      a plumbing failure in the exec subshell itself (the worktree vanished
#      before it could be entered, or exec could not run the resolved path) —
#      never the reproducer's own exit status
set -euo pipefail

# The banned-character set is single-sourced with check-panel-reproducers.sh's
# own copy of the same check — see scripts/reproducer-metachars.sh's header
# for why: this set has already drifted between the two scripts twice.
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
# Readability is checked before the source, not left to `set -e` to catch a
# failed source. Without this, a missing or unreadable
# reproducer-metachars.sh made `source` itself fail, and `set -e` aborted the
# script at exit 1 — this script's own "defect not demonstrated" code. The
# truthful answer to a missing dependency is "cannot answer at all", this
# script's own exit 4, not a verdict on a reproducer that was never run.
if [ ! -r "$SCRIPT_DIR/reproducer-metachars.sh" ]; then
  echo "run-reproducer: cannot read $SCRIPT_DIR/reproducer-metachars.sh — cannot determine the banned-character set" >&2
  exit 4
fi
# shellcheck source=reproducer-metachars.sh
source "$SCRIPT_DIR/reproducer-metachars.sh"

usage_fail() {
  echo "run-reproducer: usage: run-reproducer.sh <worktree> <reproducer-command-line>" >&2
  exit 4
}

WORKTREE="${1:-}"
CMD_TEXT="${2:-}"
[ -n "$WORKTREE" ] && [ -n "$CMD_TEXT" ] || usage_fail

[ -d "$WORKTREE" ] || { echo "run-reproducer: not a directory: $WORKTREE" >&2; exit 4; }
# Canonicalise before anything is concatenated onto it, for the same reason
# check-panel-reproducers.sh does: a relative worktree argument beginning
# with `-` would otherwise make every path built from it look like an
# option to whatever reads it. `pwd -P`, not the bare builtin: bash's `pwd`
# is logical by default and does not resolve a symlink in the path it
# printed, so a TMPDIR that is itself a symlink (`/tmp` on Darwin, to
# `/private/tmp`) left $WORKTREE and realpath's own physical answer for the
# same file disagreeing on every case below — every reproducer refused as
# "outside the worktree" against a worktree it was always inside.
WORKTREE="$(cd -- "$WORKTREE" && pwd -P)" || { echo "run-reproducer: worktree vanished before it could be resolved: ${WORKTREE}" >&2; exit 4; }

refuse() {
  echo "run-reproducer: refused — $1 — never executed" >&2
  exit 2
}

# An embedded newline is refused outright: every check below reads
# $CMD_TEXT as one line, and a second line hidden inside the argument would
# ride along unchecked by every one of them.
case "$CMD_TEXT" in
  *$'\n'*) refuse "the command line carries an embedded newline" ;;
esac

# ---------------------------------------------------------------------
# Lexical checks — mirrors check-panel-reproducers.sh's own reproducer-line
# rules, repeated here because a record can be edited after that guard last
# ran, and because this script must refuse on its own even when it is
# handed a command line from somewhere other than a validated record.
# ---------------------------------------------------------------------

# Split on whitespace via `read -ra`, not an unquoted `for tok in
# $CMD_TEXT` — the latter is subject to bash's own pathname expansion,
# which would stat the filesystem looking for matches, itself a filesystem
# touch ahead of the deliberate one below. Done before the path-token check,
# not after: `${CMD_TEXT%% *}` splits on a literal space only, while this
# tokenizer (and check-panel-reproducers.sh's own) splits on space AND tab,
# so a tab-separated command line previously yielded a path token containing
# the tab and everything after it — refused as "does not exist" rather than
# resolved correctly. Deriving PATH_TOKEN from the tokenized array's first
# element keeps the two splits in agreement.
IFS=$' \t' read -ra CMD_TOKENS <<< "$CMD_TEXT"
[ "${#CMD_TOKENS[@]}" -gt 0 ] || refuse "the command line names no path token"
PATH_TOKEN="${CMD_TOKENS[0]}"

case "$PATH_TOKEN" in
  -*) refuse "the path token '$PATH_TOKEN' begins with '-' — a reproducer names a path, never an option" ;;
esac

case "$CMD_TEXT" in
  *'://'*) refuse "the command names a URL — a reproducer is a bare path inside the worktree, never a network location" ;;
esac

# The same banned-character set check-panel-reproducers.sh enforces,
# including both quote characters: a direct exec never expands any of
# these, so their presence only ever means the author expected a shell to
# see them. Sourced from scripts/reproducer-metachars.sh, the single source
# of truth for this set.
metachars="$REPRODUCER_METACHARS"
mc_i=0
while [ "$mc_i" -lt "${#metachars}" ]; do
  mc_c="${metachars:$mc_i:1}"
  case "$CMD_TEXT" in
    *"$mc_c"*) refuse "the command carries the shell metacharacter '$mc_c' — a reproducer is a bare path optionally followed by plain arguments, never a shell command line" ;;
  esac
  mc_i=$((mc_i + 1))
done

for tok in "${CMD_TOKENS[@]}"; do
  case "$tok" in
    /*) refuse "the token '$tok' is an absolute path — a reproducer names a path relative to the worktree, on the path token or any argument" ;;
  esac
  case "$tok" in
    ..|../*|*/..|*/../*) refuse "the token '$tok' carries a '..' path segment — a reproducer must stay inside the worktree, on the path token or any argument" ;;
  esac
done

# ---------------------------------------------------------------------
# Resolved containment — the check check-panel-reproducers.sh's own header
# says it deliberately leaves to whatever runs the reproducer against a
# real worktree. Every token is resolved to its physical path and required
# to stay inside $WORKTREE; `realpath` follows `..`, `.` and symlinks in
# one step, so a relative path with no lexical `..` segment that escapes
# through a symlink is caught here.
# ---------------------------------------------------------------------

CANDIDATE_PATH="$WORKTREE/$PATH_TOKEN"
[ -e "$CANDIDATE_PATH" ] || refuse "the path token '$PATH_TOKEN' does not exist inside the worktree"
RESOLVED_PATH="$(realpath -- "$CANDIDATE_PATH" 2>/dev/null)" || refuse "the path token '$PATH_TOKEN' could not be resolved"
case "$RESOLVED_PATH" in
  "$WORKTREE" | "$WORKTREE"/*) : ;;
  *) refuse "the path token '$PATH_TOKEN' resolves to '$RESOLVED_PATH', outside the worktree — a symlink escape" ;;
esac
[ -f "$RESOLVED_PATH" ] || refuse "the path token '$PATH_TOKEN' does not resolve to a regular file"
[ -x "$RESOLVED_PATH" ] || refuse "the path token '$PATH_TOKEN' resolves to a file with no execute permission"

# Every argument after the path token is resolved and checked the same way,
# but only when it names something that exists: a plain flag such as
# `--strict` is not a path at all, and resolving a token that names nothing
# on disk answers a question this script was never asked.
ARGS=("${CMD_TOKENS[@]:1}")
for arg in "${ARGS[@]}"; do
  candidate_arg="$WORKTREE/$arg"
  [ -e "$candidate_arg" ] || continue
  resolved_arg="$(realpath -- "$candidate_arg" 2>/dev/null)" || refuse "the argument '$arg' could not be resolved"
  case "$resolved_arg" in
    "$WORKTREE" | "$WORKTREE"/*) : ;;
    *) refuse "the argument '$arg' resolves to '$resolved_arg', outside the worktree — a symlink escape" ;;
  esac
done

# ---------------------------------------------------------------------
# Execution, bounded, with a direct exec — the guarantee this script
# exists to make true rather than merely claim.
# ---------------------------------------------------------------------

# The bound and grace are this rule's own numbers (skills/myflow-do/SKILL.md
# section 5): 20 seconds, plus a 2-second SIGTERM-to-SIGKILL grace matching
# check-cleanup-complete.sh's own SURVIVORS_KILL_GRACE. RUN_REPRODUCER_*
# overrides exist for the same reason CHECK_CLEANUP_SURVIVORS_TIMEOUT does
# in that script: driving a real 20-second wait through a test harness costs
# real wall clock per case, and the override is monotone in the safe
# direction — it can only ever shrink the wait, producing MORE timeouts,
# never fewer, so no value of it can turn an unverified run into a verified
# one. Never set either for a normal invocation.
BOUND="${RUN_REPRODUCER_BOUND_SECONDS:-20}"
GRACE="${RUN_REPRODUCER_GRACE_SECONDS:-2}"
case "$BOUND" in *[!0-9]*|'') BOUND=20 ;; esac
case "$GRACE" in *[!0-9]*|'') GRACE=2 ;; esac

OUT_TMP=""
ERR_TMP=""
cleanup_tmp() {
  [ -n "$OUT_TMP" ] && rm -f "$OUT_TMP"
  [ -n "$ERR_TMP" ] && rm -f "$ERR_TMP"
  return 0
}
trap cleanup_tmp EXIT

OUT_TMP="$(mktemp "${TMPDIR:-/tmp}/run-reproducer-out.XXXXXX" 2>/dev/null)" || {
  echo "run-reproducer: no writable temporary directory — cannot run '$PATH_TOKEN' at all" >&2
  exit 4
}
ERR_TMP="$(mktemp "${TMPDIR:-/tmp}/run-reproducer-err.XXXXXX" 2>/dev/null)" || {
  echo "run-reproducer: no writable temporary directory — cannot run '$PATH_TOKEN' at all" >&2
  exit 4
}
# The sentinel is a private, unnamed pipe, not a path under $TMPDIR (F55).
# The earlier `mktemp -u ".../run-reproducer-sentinel.XXXXXX"` sentinel was
# a signal in name only: this script runs with $TMPDIR inherited by the
# reproducer it launches, and the sentinel's own literal prefix is sitting
# in this file's own source, readable by anyone deciding what a reproducer
# fixture should do — `rm -f "$TMPDIR"/run-reproducer-sentinel.*` as the
# fixture's first act deleted it before this script ever looked, and a
# genuine "defect demonstrated" then read as "cannot answer", silencing the
# one answer this tool exists to produce. It never forged the opposite
# direction — a deleted sentinel was never read as a pass — but it defeated
# the purpose on exactly the finding it was protecting.
#
# The standard "unlinked FIFO" trick makes the fix: create a real FIFO,
# open it read-write onto one fd (a FIFO opened `<>`, read-write, never
# blocks waiting for a second opener — the usual FIFO open-blocks-until-both-
# ends problem this sidesteps), then `rm -f` the FIFO's directory entry
# immediately. The open file descriptor keeps the underlying pipe alive —
# closing an fd is what frees a pipe, not removing its name — but from that
# `rm -f` onward there is no path in the filesystem at all: nothing for a
# reproducer to glob, stat, or unlink, even if it somehow guessed the exact
# random suffix. A single fd (not a coproc pair spawning a `cat` process to
# babysit) is enough because both ends of a read-write FIFO fd live in THIS
# process and are inherited by the subshell forked below like any other
# open fd — a bare `coproc` pipe was tried first and rejected: bash tracks
# coproc-created fds specially and closes them in every forked child
# (subshell included) to stop them leaking into unrelated pipelines, so the
# subshell below could never have reached them at all.
SENTINEL_FIFO="$(mktemp -u "${TMPDIR:-/tmp}/run-reproducer-sentinel.XXXXXX" 2>/dev/null)" || {
  echo "run-reproducer: no writable temporary directory — cannot run '$PATH_TOKEN' at all" >&2
  exit 4
}
mkfifo -m 600 -- "$SENTINEL_FIFO" 2>/dev/null || {
  echo "run-reproducer: could not create a sentinel pipe — cannot run '$PATH_TOKEN' at all" >&2
  exit 4
}
exec {SENTINEL_FD}<>"$SENTINEL_FIFO"
rm -f -- "$SENTINEL_FIFO"

# A sub-second poll where the platform's sleep accepts one, matching
# check-cleanup-complete.sh's own run_survivors. Probing beats assuming — a
# `sleep 0.2` that errors on some minimal shell would busy-loop instead.
POLL=1
if sleep 0.05 2>/dev/null; then POLL=0.2; fi

RC=0
TIMED_OUT=0
SURVIVOR_PIDS=""

# collect_descendants <root-pid> — emits "<pid> <depth>" for every LIVE
# descendant of <root-pid>, walking the WHOLE tree breadth-first rather than
# the one level `pgrep -P "$root"` alone sees. F54: a real double fork —
# child forks a grandchild, then the child exits immediately — orphans the
# grandchild one hop below what a single-level `pgrep -P "$CMD_PID"` can
# ever name; a snapshot taken only at that first level reported a clean
# verdict while the grandchild kept running, unreported, forever. Depth 1 is
# a direct child, depth 2 a grandchild, and so on; the caller uses the depth
# to kill the deepest processes first, so a mid-sweep kill of a parent can
# never re-orphan a child this same sweep has not reached yet. The walk is
# capped at 64 levels, purely defensively — this repository's fixtures never
# nest that deep — so a pathological or cyclic `pgrep -P` answer cannot spin
# this loop forever.
collect_descendants() {
  local root="$1" depth=1 frontier="$1" next pid children c
  while [ -n "$frontier" ] && [ "$depth" -le 64 ]; do
    next=""
    for pid in $frontier; do
      children="$(pgrep -P "$pid" 2>/dev/null || true)"
      for c in $children; do
        printf '%s %s\n' "$c" "$depth"
        next="$next $c"
      done
    done
    frontier="${next# }"
    depth=$((depth + 1))
  done
}

# `set -m` is what puts the backgrounded subshell in a process group of its
# own, so the group kill below reaches every process the reproducer forks —
# not merely the one pid this script holds. stdin is /dev/null: this script
# is non-interactive, and a reproducer that prompts for input must fail
# rather than wait on a terminal nobody is watching.
set -m
( cd -- "$WORKTREE" && printf x >&$SENTINEL_FD && exec "$RESOLVED_PATH" "${ARGS[@]}" ) </dev/null >"$OUT_TMP" 2>"$ERR_TMP" &
CMD_PID=$!
set +m

DESC_LOG=""
{
  DEADLINE=$((SECONDS + BOUND))
  while kill -0 "$CMD_PID" 2>/dev/null; do
    # Snapshotted on EVERY iteration, not only when the bound fires: a
    # reproducer that forks a child, detaches it (e.g. via os.setsid()) and
    # then exits well inside the bound would otherwise leak that child
    # entirely — the child is reparented the instant $CMD_PID dies, so a
    # walk taken only after that has nothing left to find. Polling here,
    # while the parent is still alive, is the only window in which the
    # relationship is visible at all. The walk is the FULL descendant tree
    # (`collect_descendants`, above), not one level: a double fork — a child
    # that forks a grandchild and exits immediately — orphans the
    # grandchild one hop below what a single `pgrep -P "$CMD_PID"` can ever
    # see (F54).
    seen="$(collect_descendants "$CMD_PID")"
    [ -n "$seen" ] && DESC_LOG="$DESC_LOG
$seen"
    if [ "$SECONDS" -ge "$DEADLINE" ]; then
      TIMED_OUT=1
      # Captured HERE too, right as the bound fires, rather than relying on
      # an earlier poll alone: a reproducer that double-forks or detaches
      # may not have spawned that child yet at launch time or at an earlier
      # poll, and this is the last moment before the kill sequence that a
      # walk can still see it as this script's own descendant. `ps -o sid=`
      # is not used to find it — it is not a valid ps(1) keyword on this
      # platform, the citation F44 recorded as wrong — and `ps -o sess=`
      # was measured on this machine to report 0 for every process tried, a
      # real setsid() session leader included, so it cannot be trusted to
      # name one either. `pgrep -P <pid>` — "children of this pid" — was
      # measured instead to correctly name a detached child before the kill
      # and to keep naming it, alive, after the parent was killed with
      # SIGKILL, which is the mechanism this script relies on; walking it
      # recursively is what extends that to every depth, not only the
      # first.
      seen="$(collect_descendants "$CMD_PID")"
      [ -n "$seen" ] && DESC_LOG="$DESC_LOG
$seen"
      kill -TERM -"$CMD_PID" 2>/dev/null || kill -TERM "$CMD_PID" 2>/dev/null || true
      GRACE_DEADLINE=$((SECONDS + GRACE))
      while kill -0 "$CMD_PID" 2>/dev/null && [ "$SECONDS" -lt "$GRACE_DEADLINE" ]; do sleep "$POLL"; done
      kill -KILL -"$CMD_PID" 2>/dev/null || kill -KILL "$CMD_PID" 2>/dev/null || true
      break
    fi
    sleep "$POLL"
  done
  set +e
  wait "$CMD_PID"
  RC=$?
  set -e
} 2>/dev/null

# KIDS is the union of every "<pid> <depth>" line logged above, reduced to
# one entry per pid (the deepest depth it was ever seen at — a pid does not
# change its place in the tree while it stays a descendant, but taking the
# max is defensive against a reused pid observed at different depths across
# polls) and ordered DEEPEST FIRST. That ordering is what the kill sweep
# below relies on: killing a parent before a still-uncollected child would
# re-orphan that child mid-sweep, the exact defect this walk exists to stop
# happening by omission.
if [ -n "$DESC_LOG" ]; then
  KIDS="$(printf '%s\n' "$DESC_LOG" | awk 'NF==2 { d=$2+0; p=$1; if (!(p in maxd) || d>maxd[p]) maxd[p]=d } END { for (p in maxd) print maxd[p], p }' | sort -rn -k1,1 -k2,2n | awk '{print $2}')"
else
  KIDS=""
fi

# RESIDUAL GAP, stated honestly rather than implied away: a grandchild that
# re-parents (e.g. via a double fork whose middle process exits) to pid 1
# BEFORE any poll iteration above ever observes it is invisible to this
# technique. `pgrep -P` only ever answers "children of this still-tracked
# pid right now" — once a process is reparented to init/launchd with no
# poll having caught it as a descendant first, nothing run by this script
# can walk back to it. `pgrep -P` remains the only tool available here:
# `ps -o sid=` is not a valid keyword on this platform (F44) and `ps -o
# sess=` was measured to report 0 for every process tried, real session
# leader included, so neither can name a detached tree by session instead.
# Polling on every iteration (not only at the bound) narrows this window as
# far as it can be narrowed without a tool this platform does not have; it
# does not close it.

# A detached child that left the reproducer's own process group is
# untouched by the group kill above (on the timeout path) and was never
# killed at all (on the path below, where the parent exited under the bound
# on its own). Retry the same SIGTERM-then-grace-then-SIGKILL sequence
# against every pid any poll above named, then report whichever of them is
# still alive — checked UNCONDITIONALLY, on every run, not only when
# $TIMED_OUT is 1: a child detached before the parent exits normally is
# exactly as orphaned and exactly as unaccounted-for as one detached before
# a kill, and the disposition below is the same either way.
if [ -n "${KIDS:-}" ]; then
  STILL=""
  for kid in $KIDS; do
    kill -0 "$kid" 2>/dev/null && STILL="$STILL $kid"
  done
  STILL="${STILL# }"
  if [ -n "$STILL" ]; then
    for kid in $STILL; do kill -TERM "$kid" 2>/dev/null || true; done
    SURVIVOR_GRACE_DEADLINE=$((SECONDS + GRACE))
    for kid in $STILL; do
      while kill -0 "$kid" 2>/dev/null && [ "$SECONDS" -lt "$SURVIVOR_GRACE_DEADLINE" ]; do sleep "$POLL"; done
    done
    for kid in $STILL; do kill -KILL "$kid" 2>/dev/null || true; done
    # Checked with NO added delay, deliberately. Measured on Darwin 25.5.0:
    # a descendant reparented to launchd that ignores SIGTERM and is then
    # sent SIGKILL is reaped within roughly 0.2s in the ordinary case, but
    # `kill -0` immediately after the SIGKILL call — no sleep at all —
    # reliably still reports it present, five runs out of five, because the
    # signal has been delivered but the process table entry has not yet
    # been reclaimed by its new parent. Sleeping here even briefly closes
    # exactly the window this check exists to observe, and would read a
    # genuinely-surviving process as gone before naming it. A process that
    # really has already been reaped by the time this line runs is still
    # correctly read as gone; the check costs nothing when there is nothing
    # left to find.
    for kid in $STILL; do
      kill -0 "$kid" 2>/dev/null && SURVIVOR_PIDS="$SURVIVOR_PIDS $kid"
    done
    SURVIVOR_PIDS="${SURVIVOR_PIDS# }"
  fi
fi

# A reproducer killed at the bound, or one whose child survived the retry
# above, may already have written to the worktree before it was killed, so
# the worktree is re-checked here, before this script's own run continues
# any further. `/myflow-do` re-checks it again itself when the operator
# resumes, since nothing between this exit and that resume observed what a
# still-alive survivor went on to write — that second check is the caller's
# responsibility and is out of this single invocation's scope.
git_status_note() {
  local status
  status="$(git -C "$WORKTREE" status --porcelain 2>/dev/null)" || return 0
  if [ -n "$status" ]; then
    printf 'run-reproducer: worktree status after the kill:\n%s\n' "$status" >&2
  else
    printf 'run-reproducer: worktree status after the kill: clean\n' >&2
  fi
}

# The bounce built on a reproducer's result carries "the reproducer's
# passing output" back to whichever review-panel slot raised the finding —
# but until this ran, the caller had nothing to carry: $OUT_TMP and $ERR_TMP
# were read by nothing and deleted, unread, in the EXIT trap. Emitted here,
# to the same stream $1 names, wrapped in a fixed pair of delimiter lines so
# a caller can tell where the reproducer's own output starts and ends and
# never mistake it for one of this script's own `run-reproducer:` messages
# printed around it. Silent when the reproducer produced nothing on either
# stream — an empty pair of delimiters would say less than saying nothing.
emit_captured_output() {
  local stream="$1" out err
  out="$(cat "$OUT_TMP" 2>/dev/null || true)"
  err="$(cat "$ERR_TMP" 2>/dev/null || true)"
  [ -n "$out" ] || [ -n "$err" ] || return 0
  {
    echo "run-reproducer: --- captured reproducer output begin ---"
    [ -n "$out" ] && printf '%s\n' "$out"
    if [ -n "$err" ]; then
      echo "run-reproducer: --- captured reproducer stderr ---"
      printf '%s\n' "$err"
    fi
    echo "run-reproducer: --- captured reproducer output end ---"
  } >&"$stream"
}

if [ -n "$SURVIVOR_PIDS" ]; then
  git_status_note
  emit_captured_output 2
  if [ "$TIMED_OUT" -eq 1 ]; then
    echo "run-reproducer: unverifiable — '$CMD_TEXT' was still running at the ${BOUND}s bound and was killed, and a detached child survived the process-group kill; surviving process pid(s): $SURVIVOR_PIDS — find and kill it manually, and re-check the worktree again when you resume" >&2
  else
    echo "run-reproducer: unverifiable — '$CMD_TEXT' exited $RC inside the ${BOUND}s bound, but it forked a detached child that outlived it and survived the kill retry; surviving process pid(s): $SURVIVOR_PIDS — find and kill it manually, and re-check the worktree again when you resume" >&2
  fi
  exit 3
fi

if [ "$TIMED_OUT" -eq 1 ]; then
  git_status_note
  emit_captured_output 2
  echo "run-reproducer: unverifiable — '$CMD_TEXT' was still running at the ${BOUND}s bound and was killed; its exit status is read as neither a pass nor a fail" >&2
  exit 3
fi

# No byte on the sentinel pipe means the subshell never reached `exec` at
# all — `cd -- "$WORKTREE"` failed, almost certainly because the worktree
# vanished between this script's own earlier resolution of it and the
# backgrounded subshell actually starting. Whatever $RC came back as is
# bash's own report of that failure, never the reproducer's, so it is read
# as "cannot answer" rather than as a pass or a fail. By the time this line
# runs the subshell has already finished (the `wait` above returned), so a
# byte written before `exec` is already sitting in the pipe and `read`
# returns immediately; a short timeout — never a long wait — is only a
# guard against this check itself hanging if the coprocess were ever in a
# state it should not be.
if ! IFS= read -r -t 1 -n 1 -u "$SENTINEL_FD" _sentinel_byte 2>/dev/null; then
  echo "run-reproducer: cannot answer — the worktree vanished before '$CMD_TEXT' could be started" >&2
  exit 4
fi

# The sentinel IS present, so `cd` succeeded and the subshell reached the
# `exec` line — but exit codes 126 and 127 are bash's own conventional
# reports of "found but not executable" and "not found" for a command *it*
# tried to run, emitted only when exec itself never replaced the process
# image (a race after this script's own executability check: the file was
# removed, replaced, or its permissions changed in the gap). A reproducer
# that actually ran can and occasionally does exit with either number on
# its own, so this is a heuristic, not a certainty — but a false "cannot
# answer" on a real 126/127 exit costs a re-run, while reading a genuine
# plumbing failure as "defect demonstrated" costs a fix built on nothing.
case "$RC" in
  126|127)
    echo "run-reproducer: cannot answer — exec of '$RESOLVED_PATH' itself failed (exit $RC), which is never read as the reproducer's own exit status" >&2
    exit 4
    ;;
esac

if [ "$RC" -ne 0 ]; then
  emit_captured_output 1
  echo "run-reproducer: defect demonstrated — '$CMD_TEXT' exited $RC"
  exit 0
fi

emit_captured_output 1
echo "run-reproducer: defect not demonstrated — '$CMD_TEXT' exited 0"
exit 1
