# scripts/lib/parallel.sh — bounded-concurrency spawn, per-job output
# capture, deterministic replay and the aggregate exit code, defined once.
#
# KAN-362 exists because the guard test suite's 40 harnesses ran
# sequentially at 119s, past the tool timeout. Two callers need the same
# answer to "run N independent jobs concurrently, and when one fails, show
# me exactly what it printed, in the order I listed the jobs, never the
# order they happened to finish": scripts/run-guard-tests.sh (task 2, one
# job per `test-*.sh` harness) and scripts/test-check-installed-citations.sh
# (task 3, one job per fixture case). If those two carried separate copies
# of the `xargs` spawn-and-capture logic, they could drift on exactly the
# concern that matters most here — replay ORDER — the same five-copy
# `resolve_file` drift scripts/lib/coverage.sh's own header records this
# repository already suffered when a shared concern was not owned once. So
# the resolution lives here, in one implementation both source, per
# design.md's shared-parallel-lib decision.
#
# BASH 3.2 IS THE FLOOR (macOS's own /bin/bash), per
# test-check-installed-citations.sh's own header, citing
# test-check-finish-preflight.sh: indexed arrays only, no associative
# arrays, and therefore no `wait -n` (bash 4.3+) — which is why this file is
# built on `xargs -P`, the same mechanism proposal.md's 49.6s measurement
# used, rather than a second mechanism nobody has timed. Every array below
# is walked by its indices — "${!ARRAY[@]}" — never by its values —
# "${ARRAY[@]}" — directly, for the same empty-array-under-`set -u` reason
# scripts/lib/coverage.sh's header records.
#
# WHY A SCRIPT FILE, NOT ONLY A SOURCED LIBRARY: `xargs -P` spawns each job
# as a genuinely separate process, and bash cannot export an array or a
# function's own local state into that process's environment (arrays are
# never exportable, and `export -f` only carries function bodies, not the
# indexed-by-list-position command strings parallel_run is handed). So each
# job's command string is staged to its own file under the run's mktemp -d
# BEFORE spawning, and `xargs -P` invokes this file a second time, in a
# fresh `bash`, as `bash parallel.sh __run_one <tmpdir> <index>` — the one
# piece of this library's own contract that runs as a script rather than a
# sourced function, because nothing else can reach across that process
# boundary. `parallel_run` is the only caller of this dispatch path; nothing
# above it should invoke `__run_one` directly.
#
# CAPTURE: each job's stdout and stderr are merged into one file,
# `<tmpdir>/<index>.out` — merged because REPLAY's job is to reproduce
# exactly what a reader would have seen watching that job run alone in a
# terminal, and a terminal does not keep two separate scrollback buffers.
#
# REPLAY IS IN LIST ORDER, NEVER COMPLETION ORDER: parallel_replay_failures
# walks PARALLEL_RC by ascending index — list order, by construction, since
# a job's index is fixed at the position `parallel_run` was handed it in —
# never by when its output file was written.
#
# JOB COUNT (design.md's jobs-from-core-count): `sysctl -n hw.ncpu` (macOS),
# then `nproc` (Linux), then a fallback of 4; `JOBS=` overrides. A `JOBS=`
# that is SET but not a positive integer is refused at exit 2, never
# coerced into the fallback — this repository's guards reject rather than
# fold bad input into a reassuring default, exactly as
# scripts/lib/coverage.sh's header records for coverage_record's own count
# argument.
#
# CLEANUP: every mktemp -d this library creates is removed when the
# sourcing process exits, including on SIGINT/SIGTERM — a trap installed
# once, at source time, on EXIT/INT/TERM. This is intentionally global to
# the sourcing process rather than scoped per parallel_run call, because a
# caller reads PARALLEL_OUTFILES (or replays failures) only AFTER
# parallel_run returns; removing the tmp dir any earlier than process exit
# would race the caller's own read of it. Installing these traps means this
# library --- unlike scripts/lib/coverage.sh and scripts/lib/owned-corpus.sh
# --- is NOT trap-neutral: a caller that needs its own EXIT/INT/TERM trap
# must install it before sourcing this file and chain to
# `_parallel_cleanup` itself, or install it after sourcing and call
# `_parallel_cleanup` from within it. scripts/run-guard-tests.sh installs no
# competing trap of its own, so it needs none of this — but
# scripts/test-check-installed-citations.sh does install its own EXIT/INT/TERM
# trap (guarding its own fixture-repo SANDBOXES) and sources this file into
# the same process, so it chains to `_parallel_cleanup` per the recipe above.
#
# Not meant to be executed directly for anything other than the internal
# `__run_one` dispatch documented above — a caller sources it and calls
# parallel_job_count / parallel_run / parallel_replay_failures.

_parallel_cleanup() {
  local i
  for i in "${!PARALLEL_CLEANUP_DIRS[@]}"; do
    rm -rf "${PARALLEL_CLEANUP_DIRS[$i]}"
  done
}

_parallel_install_traps() {
  trap '_parallel_cleanup' EXIT
  trap '_parallel_cleanup; trap - INT; kill -s INT "$$"' INT
  trap '_parallel_cleanup; trap - TERM; kill -s TERM "$$"' TERM
}

# parallel_job_count -> prints the job count to stdout, exit 0. Honours
# JOBS when it is set at all (including set-but-empty, which is refused,
# not treated as unset) — see the header for why a bad value is refused
# rather than coerced. Falls back through sysctl -n hw.ncpu, then nproc,
# then a literal 4 when JOBS is unset.
parallel_job_count() {
  if [ "${JOBS+set}" = set ]; then
    case "$JOBS" in
      ''|*[!0-9]*)
        printf "parallel_job_count: JOBS must be a positive integer, got '%s'\n" "$JOBS" >&2
        return 2
        ;;
    esac
    if [ "$JOBS" -eq 0 ]; then
      printf 'parallel_job_count: JOBS must be a positive integer, got 0\n' >&2
      return 2
    fi
    printf '%s\n' "$JOBS"
    return 0
  fi

  local n
  if command -v sysctl >/dev/null 2>&1; then
    n="$(sysctl -n hw.ncpu 2>/dev/null || true)"
    case "$n" in
      ''|*[!0-9]*) : ;;
      *) [ "$n" -gt 0 ] && { printf '%s\n' "$n"; return 0; } ;;
    esac
  fi
  if command -v nproc >/dev/null 2>&1; then
    n="$(nproc 2>/dev/null || true)"
    case "$n" in
      ''|*[!0-9]*) : ;;
      *) [ "$n" -gt 0 ] && { printf '%s\n' "$n"; return 0; } ;;
    esac
  fi
  printf '%s\n' 4
}

# _parallel_run_one <tmpdir> <index> — internal. Runs the command staged at
# <tmpdir>/cmd.<index> in a fresh bash, merging stdout+stderr into
# <tmpdir>/<index>.out, and records its exit code at <tmpdir>/<index>.rc.
# Invoked only via this file's own __run_one dispatch below, one process per
# job, so it never shares in-process state (arrays, functions) with
# parallel_run's caller — see the header's "WHY A SCRIPT FILE" note.
_parallel_run_one() {
  local tmp="$1" idx="$2" cmd rc
  cmd="$(cat "$tmp/cmd.$idx")"
  set +e
  bash -c "$cmd" >"$tmp/$idx.out" 2>&1
  rc=$?
  set -e
  printf '%s\n' "$rc" > "$tmp/$idx.rc"
  return "$rc"
}

# parallel_run <cmd...> — runs each argument as a shell command string,
# bounded by parallel_job_count concurrent workers. Populates, in list
# order (index 0 = the first argument):
#   PARALLEL_RC[i]        the exit code of job i
#   PARALLEL_OUTFILES[i]  path to job i's merged stdout+stderr capture
# Returns 0 when every job exited 0, 1 when at least one did not, 2 when
# JOBS was set but invalid or a temporary directory could not be created —
# in either exit-2 case, PARALLEL_RC/PARALLEL_OUTFILES are left untouched.
#
# SAFE UNDER `set -euo pipefail` (F3): parallel_run neutralizes `set -e`/
# `pipefail` around its own internal `xargs -P` spawn pipeline, so its
# INTERNAL machinery never aborts a caller mid-function — PARALLEL_RC and
# PARALLEL_OUTFILES are always fully populated, in list order, by the time
# this function returns or is in the process of returning, regardless of
# how many jobs failed or how the caller invoked it. What this does NOT
# change: a caller that invokes parallel_run as a bare, unguarded statement
# under `set -e` is still aborted the instant it returns non-zero — that is
# ordinary bash semantics for ANY command returning non-zero as a bare
# statement, not something a library function can opt out of while still
# honestly reporting failure via its own return code. A caller that needs
# to read the result — `if`/`while`/`&&`/`||`, or its own `set +e ...
# set -e` bracket (every current call site already does the latter) —
# always reaches it correctly; see scripts/test-lib-parallel.sh case 9 for
# the proof, including the generic-bash control that isolates this from a
# parallel.sh-specific defect.
parallel_run() {
  if [ "$#" -eq 0 ]; then
    printf 'parallel_run: want at least 1 job\n' >&2
    return 2
  fi

  local jobs
  jobs="$(parallel_job_count)" || return $?

  local tmp
  tmp="$(mktemp -d "${TMPDIR:-/tmp}/parallel-lib.XXXXXX")" || {
    printf 'parallel_run: cannot create a temporary directory\n' >&2
    return 2
  }
  PARALLEL_CLEANUP_DIRS+=("$tmp")

  local n=$# i=0 cmd
  for cmd in "$@"; do
    printf '%s' "$cmd" > "$tmp/cmd.$i"
    i=$((i + 1))
  done

  local self
  self="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/$(basename "${BASH_SOURCE[0]}")"

  # SAFE UNDER `set -euo pipefail` (F3): xargs exits non-zero the instant any
  # job it dispatched exited non-zero, and with `pipefail` that non-zero
  # status propagates through the pipe into this statement's own exit
  # status. A caller running under `set -e` would then be aborted HERE, at
  # the call to parallel_run, before ever reaching the return below that
  # reports rc_overall — the exact failure this comment's own reproducer
  # demonstrates. `set +e` neutralizes that for this one statement only;
  # parallel_run computes rc_overall itself from each job's own recorded
  # PARALLEL_RC below regardless of the pipeline's own exit status, so
  # nothing here depends on the discarded value.
  #
  # RESTORE, NEVER UNCONDITIONALLY `set -e`: a caller who already brackets
  # its own call in `set +e ... parallel_run ...; rc=$?; set -e` (every
  # current call site does exactly this) has errexit OFF, process-wide,
  # for the whole duration of this call — `set -e` and `set +e` are shell
  # options, not scoped per function call. Turning it back ON here
  # unconditionally would re-enable errexit WHILE STILL INSIDE that
  # caller's own bracket, aborting this function's own later bare
  # `return "$rc_overall"` before the caller ever reaches its `rc=$?` —
  # trading the finding's hazard for a worse one for the callers that
  # already do the right thing. So the errexit bit is captured via `$-`
  # before touching it, and restored only if it was actually on.
  local had_errexit=0
  case "$-" in *e*) had_errexit=1 ;; esac
  i=0
  set +e
  {
    while [ "$i" -lt "$n" ]; do
      printf '%s\n' "$i"
      i=$((i + 1))
    done
  } | xargs -P "$jobs" -n 1 bash "$self" __run_one "$tmp"
  [ "$had_errexit" -eq 1 ] && set -e

  PARALLEL_RC=()
  PARALLEL_OUTFILES=()
  local rc_overall=0 j jrc
  j=0
  while [ "$j" -lt "$n" ]; do
    jrc="$(cat "$tmp/$j.rc" 2>/dev/null || printf 127)"
    PARALLEL_RC[$j]="$jrc"
    PARALLEL_OUTFILES[$j]="$tmp/$j.out"
    [ "$jrc" -eq 0 ] || rc_overall=1
    j=$((j + 1))
  done
  return "$rc_overall"
}

# parallel_replay_failures [label...] -> prints, for every job whose
# PARALLEL_RC was non-zero, its captured output verbatim, in list order and
# contiguously (one job's file fully, then the next) — never interleaved,
# since each file is cat'd whole before moving to the next index. Prints
# nothing when every job passed. Must be called after parallel_run.
#
# LABELS (optional, positional, same list order as parallel_run's own
# arguments): when given, one per job, a failing job's block is preceded by
# a `---- <label> ----` header and followed by a blank line — the per-job
# name this library cannot derive on its own, since it only ever sees
# opaque command strings, never a caller's name for one — and
# scripts/run-guard-tests.sh needs exactly that header to say which harness
# a block belongs to. Called with no labels at all, output is unchanged
# from this function's original header-less form (no header, no trailing
# blank line), so an existing caller keeps its old output exactly.
parallel_replay_failures() {
  local n=${#PARALLEL_RC[@]}
  [ "$n" -eq 0 ] && return 0
  local labels=("$@")
  local i
  for i in "${!PARALLEL_RC[@]}"; do
    if [ "${PARALLEL_RC[$i]}" -ne 0 ]; then
      if [ "$#" -gt 0 ]; then
        printf -- '---- %s ----\n' "${labels[$i]}"
      fi
      cat "${PARALLEL_OUTFILES[$i]}"
      [ "$#" -gt 0 ] && printf '\n'
    fi
  done
  return 0
}

if [ "${BASH_SOURCE[0]:-$0}" = "$0" ]; then
  set -u
  case "${1:-}" in
    __run_one)
      shift
      _parallel_run_one "$@"
      exit $?
      ;;
    *)
      printf 'parallel.sh: source this file and call parallel_run — do not execute it directly (except the internal __run_one dispatch)\n' >&2
      exit 2
      ;;
  esac
fi

PARALLEL_CLEANUP_DIRS=()
_parallel_install_traps
