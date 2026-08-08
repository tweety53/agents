#!/usr/bin/env bash
# gather-self-review-context.sh — deterministically collect the SDD ledger,
# the review-panel record, tasks.md and the relevant git log for a finished
# change, so the self-review reasoning pass (/myflow-finish run 2, step 8)
# judges a bundle instead of re-reading files itself.
#
# Usage: gather-self-review-context.sh <archived-change-path> <name> <state-dir>
#
# <archived-change-path> is the ARCHIVED change directory
# (openspec/changes/archive/<date>-<name>/), never a worktree path — by the
# time step 8 runs, run 2 has already removed the worktree. <state-dir> is
# accepted for CLI parity with this script's usage line and with
# preserve-session-records.sh's own <worktree> <name> <state-dir> shape; none
# of the four sources below currently read from it.
#
# Prints one bundle to stdout: a header line naming which sources were found
# vs. skipped, an explicit "skipped: <src> (absent)" line per missing source
# (1-3) — matching preserve-session-records.sh's own skipped:/preserved:
# vocabulary — and each found source's content under its own subheading.
# Exits 2 on a missing argument or an invalid change name (a malformed
# invocation); otherwise ALWAYS exits 0 — a missing source is never fatal,
# since a change may legitimately have no review-panel record, but a
# malformed invocation is. This script performs no judgement and makes no
# pass/fail determination based on what it finds.
#
# NOTE on <archived-change-path> itself being a symlink, or any ancestor of
# it, at any depth (F12, F13, F14, F20, F22): git tracks symlinks as
# committed blobs, so a merged PR could make openspec/changes/archive/<date>-
# <name> itself a symlink to anywhere, or make ANY ancestor component
# (openspec/, openspec/changes/, openspec/changes/archive/, or a deeper
# nesting level not even part of the documented shape) a symlink instead.
# Four rounds of bounded, lexical, fixed-depth patches — a leaf-only check, a
# trailing-slash strip, an exactly-3-hop ancestor `dirname` walk, a literal
# `..`-component rejection — each closed the specific bypass the previous
# round's reviewer found, but the underlying design stayed fragile to any
# path shape one step outside what each patch anticipated: most simply, one
# extra directory nesting level under openspec/changes/archive/ defeats a
# fixed hop count outright (F22).
#
# validate_archived_path() below replaces all of that with one general
# mechanism: resolve once, then compare exactly.
#
#   1. TRUSTED_REPO_ROOT is derived from this script's own process cwd via
#      `git rev-parse --git-common-dir`, NEVER `-C "$ARCHIVED_PATH"`, which
#      would trust the very path being validated. `--git-common-dir` is used
#      rather than `--show-toplevel` for the same reason this repository's own
#      `skills/myflow-contracts/state-file.md` already documents: `--show-
#      toplevel` returns a *worktree's* root when run inside a worktree, not
#      the main repository's root, which would let a worktree-cwd invocation
#      derive the wrong trust boundary entirely. `--git-common-dir` always
#      points at the main repo's `.git` — including from inside a worktree —
#      so its parent directory is the correct, invocation-independent trust
#      anchor. This script's documented caller (skills/myflow-finish/SKILL.md
#      step 8) always invokes it with process cwd at the repository root,
#      exactly like this repository's other pipeline scripts, so this remains
#      a safe trust anchor there too. Empty (not inside a git repository at
#      all) is handled the same way as every other invalid-path case below.
#   2. $ARCHIVED_PATH is normalized into an absolute, lexically-collapsed
#      path — `.` and `..` components resolved by pure string manipulation (a
#      component-stack walk: push each `/`-separated segment, pop on `..`,
#      never popping past the root, skip `.` and empty segments) — WITHOUT
#      touching the filesystem at all, so it cannot itself be fooled by a
#      symlink. Call this ARCHIVED_LEXICAL.
#   3. ARCHIVED_LEXICAL must sit exactly at
#      "$TRUSTED_REPO_ROOT/openspec/changes/archive/<leaf>" — one path
#      segment past the trusted archive root, no more and no less. Checked as
#      a real path BOUNDARY via `within_root()`, not a string prefix.
#      Anything else (missing, extra nesting, wrong location entirely) is an
#      invalid invocation.
#   4. $ARCHIVED_PATH is ALSO resolved semantically, following every symlink
#      at every component (the `cd -P`-based real resolution this script
#      already used). Call this ARCHIVED_REAL.
#   5. ARCHIVED_LEXICAL and ARCHIVED_REAL are compared for EXACT equality.
#      Step 2 never resolves anything; step 4 resolves everything; so any
#      divergence between them — the leaf, any ancestor at any depth, a
#      trailing slash, a `..` component, in any combination — means a
#      symlink sat somewhere between the trusted repo root and the leaf, and
#      the invocation is refused. This single comparison subsumes every one
#      of the four prior, bounded checks (F12 leaf symlink, F13 trailing
#      slash, F14 ancestor symlink at any hop, F20 `..` desync) with no
#      assumption about how many ancestor levels exist, because it isn't
#      looking for a specific bypass shape — it's looking for ANY divergence
#      between "what the path string says" and "what the filesystem actually
#      resolves to".
#
# Once step 5 passes, REPO_ROOT = TRUSTED_REPO_ROOT and ARCHIVED_REAL is
# already known-safe from step 4 — nothing downstream needs another `git -C
# "$ARCHIVED_PATH"` call.
#
# All failure modes (no enclosing repo, wrong shape/location, or a lexical/
# real divergence) are treated the same invocation-error way: a distinct
# note, every source reported skipped, still exit 0 — this is deliberately
# ONE code path (ARCHIVED_PATH_INVALID below), not several: they are all "the
# invocation looks wrong," just for different reasons.
#
# The four sources:
#   1. docs/superpowers/ledgers/<name>.md
#   2. docs/superpowers/reviews/<name>-panel.md
#   3. <archived-change-path>/tasks.md
#   4. git log --stat for the change's two finish-run-1 commits (the
#      implementation commit and the "plan and session records" commit) plus
#      the archive commit. Both that current subject and the wording finish
#      run 1 used before the rename ("plan, test guide and session records")
#      are matched, so changes committed before the rename keep resolving.
#
# NOTE on sources 1 and 2: preserve-session-records.sh writes these under
# docs/superpowers/{ledgers,reviews}/ with a LEADING DATE, e.g.
# "2026-08-01-demo.md", never literally "<name>.md" — this script's messages
# still name the source using the plain "<name>.md" / "<name>-panel.md" form,
# matching the wording the delta spec
# (openspec/changes/kan-23-myflow-self-review/specs/myflow-self-review/spec.md)
# states verbatim in its scenarios, but the SEARCH below uses the real,
# date-prefixed filename shape so a source that exists is actually found.
#
# NOTE on the "skipped:" stream: the design doc originally said stderr; this
# was corrected to stdout to match the delta spec
# (openspec/changes/kan-23-myflow-self-review/specs/myflow-self-review/spec.md,
# which has always said stdout) and to mirror preserve-session-records.sh's
# own convention of printing "skipped:" on stdout, since this script's own
# bundle is a single stdout document by design.
#
# THE CHANGE-NAME ALLOWLIST mirrors preserve-session-records.sh's own: one
# leading alphanumeric, then letters, digits, '.', '_' and '-'. A name
# containing '/' or a glob metacharacter must not be used to build a path or
# to build the `find -name` pattern used to locate the dated ledger/panel
# files, for the same reasons that script's header states.
set -euo pipefail

ARCHIVED_PATH="${1:-}"
NAME="${2:-}"
STATE_DIR="${3:-}"

if [ -z "$ARCHIVED_PATH" ] || [ -z "$NAME" ] || [ -z "$STATE_DIR" ]; then
  echo "usage: gather-self-review-context.sh <archived-change-path> <name> <state-dir>" >&2
  exit 2
fi

case "$NAME" in
  [!A-Za-z0-9]* | *[!A-Za-z0-9._-]*)
    echo "gather-self-review-context: change name '$NAME' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'" >&2
    exit 2
    ;;
esac

# within_root <resolved-path> <root> — true iff <resolved-path> is <root>
# itself or lives under it. Compared as a path BOUNDARY, not a string prefix
# (preserve-session-records.sh's own protections make the same distinction):
# a bare prefix test would accept "/foo/bar-evil" against "/foo/bar". Defined
# ahead of validate_archived_path() below, which reuses it for the same
# boundary test rather than a bespoke string comparison.
within_root() {
  case "$1" in
    "$2" | "$2"/*) return 0 ;;
    *) return 1 ;;
  esac
}

# validate_archived_path — the ONE mechanism described in the header NOTE
# above (F22): lexically normalize, semantically resolve, compare, refuse on
# divergence. Sets:
#   ARCHIVED_PATH_INVALID — 0 or 1
#   ARCHIVED_PATH_REASON  — one of no-repo | shape | symlink, set only when
#                           INVALID=1
#   REPO_ROOT             — TRUSTED_REPO_ROOT, set only when INVALID=0
#   ARCHIVED_REAL         — the fully-resolved archived path, set only when
#                           INVALID=0
# Purely cosmetic normalization of $ARCHIVED_PATH (e.g. a trailing-slash
# strip) is NOT this function's job and does not belong inside it (F24) —
# this function is security-named and does only security-relevant work; see
# the dedicated cosmetic-only step immediately before the call below.
#
# Called BEFORE anything else derives a trust boundary from $ARCHIVED_PATH,
# so an invalid path — of any shape, at any depth — is never resolved into
# one. An archived path invalid for any reason here is an invalid-invocation
# shape (a caller passed the wrong path, run 2's archive step did not
# actually run, a tracked symlink was planted somewhere at or above the
# archive location, or the caller's own trusted argument construction was
# somehow bypassed), not a legitimate "this change has no records" state.
validate_archived_path() {
  ARCHIVED_PATH_INVALID=0
  ARCHIVED_PATH_REASON=""
  REPO_ROOT=""
  ARCHIVED_REAL=""

  # Step 1: TRUSTED_REPO_ROOT, derived from this script's own process cwd,
  # independent of $ARCHIVED_PATH — never `-C "$ARCHIVED_PATH"`. Resolved via
  # `--git-common-dir` (see header NOTE, F23) rather than `--show-toplevel` so
  # this is correct whether cwd is the main checkout or a worktree.
  local trusted_root common_dir
  common_dir="$(git rev-parse --git-common-dir 2>/dev/null || true)"
  if [ -n "$common_dir" ]; then
    trusted_root="$(cd "$(dirname "$common_dir")" && pwd -P 2>/dev/null || true)"
  else
    trusted_root=""
  fi
  if [ -z "$trusted_root" ]; then
    ARCHIVED_PATH_INVALID=1
    ARCHIVED_PATH_REASON="no-repo"
    return 0
  fi

  # Step 2: lexically normalize $ARCHIVED_PATH into an absolute path — a
  # relative path is joined to $trusted_root (this script's documented
  # caller always runs it with cwd AT the repo root, so $trusted_root is
  # cwd's own trusted, canonical representation; joining to it here, rather
  # than to a separately-read $PWD, sidesteps a spurious mismatch when cwd
  # sits under an OS-level path alias such as macOS's /tmp -> /private/tmp,
  # since `git rev-parse` resolves that alias and a raw $PWD read would not).
  # `.`/`..` components are then collapsed by a pure component-stack walk —
  # push each `/`-separated segment, pop on `..` (never past the root), skip
  # `.` and empty segments — touching the filesystem nowhere in this step, so
  # it cannot be fooled by a symlink. The split into segments uses
  # `IFS=/ read -ra` (F26), never an unquoted `for part in $abs_input`: bash
  # applies pathname (glob) expansion as well as word-splitting to an
  # unquoted expansion, so a `*`/`?`/`[...]` in $ARCHIVED_PATH would expand
  # against the process cwd and corrupt this "pure string manipulation" step
  # with real filesystem entries; `read` only field-splits on $IFS and never
  # globs, closing that gap entirely.
  local abs_input
  case "$ARCHIVED_PATH" in
    /*) abs_input="$ARCHIVED_PATH" ;;
    *) abs_input="$trusted_root/$ARCHIVED_PATH" ;;
  esac

  local -a stack=() parts=()
  local part
  IFS=/ read -ra parts <<< "$abs_input"
  for part in "${parts[@]}"; do
    case "$part" in
      "" | ".") continue ;;
      "..")
        if [ "${#stack[@]}" -gt 0 ]; then
          unset "stack[$((${#stack[@]} - 1))]"
        fi
        ;;
      *) stack+=("$part") ;;
    esac
  done

  local archived_lexical="" seg
  for seg in "${stack[@]:-}"; do
    [ -n "$seg" ] && archived_lexical="$archived_lexical/$seg"
  done
  [ -n "$archived_lexical" ] || archived_lexical="/"

  # Step 3: ARCHIVED_LEXICAL must sit exactly one segment past
  # $trusted_root/openspec/changes/archive/ — a real path boundary
  # (within_root), not a string prefix, and exactly one segment: no deeper
  # nesting (F22), no shallower.
  local archive_root="$trusted_root/openspec/changes/archive"
  if ! within_root "$archived_lexical" "$archive_root"; then
    ARCHIVED_PATH_INVALID=1
    ARCHIVED_PATH_REASON="shape"
    return 0
  fi
  local remainder="${archived_lexical#"$archive_root"/}"
  case "$remainder" in
    */* | "")
      ARCHIVED_PATH_INVALID=1
      ARCHIVED_PATH_REASON="shape"
      return 0
      ;;
  esac

  # Step 4: resolve $ARCHIVED_PATH semantically, following every symlink at
  # every component.
  local archived_real
  archived_real=""
  if [ -d "$ARCHIVED_PATH" ]; then
    archived_real="$(cd -P "$ARCHIVED_PATH" 2>/dev/null && pwd -P)" || archived_real=""
  fi
  if [ -z "$archived_real" ]; then
    # A leaf that IS a symlink but resolves to something other than a
    # directory (e.g. a symlink to a plain file) fails the `-d` test above
    # just like a missing path does — but it is not a shape problem, it's
    # exactly the same trust-boundary problem as a symlink pointing outside
    # the repo (F25): report it as such rather than as "shape".
    if [ -L "$ARCHIVED_PATH" ]; then
      ARCHIVED_PATH_INVALID=1
      ARCHIVED_PATH_REASON="symlink"
      return 0
    fi
    ARCHIVED_PATH_INVALID=1
    ARCHIVED_PATH_REASON="shape"
    return 0
  fi

  # Step 5: exact-compare. Any divergence, at any depth, means a symlink sat
  # somewhere between the trusted repo root and the leaf.
  if [ "$archived_lexical" != "$archived_real" ]; then
    ARCHIVED_PATH_INVALID=1
    ARCHIVED_PATH_REASON="symlink"
    return 0
  fi

  REPO_ROOT="$trusted_root"
  ARCHIVED_REAL="$archived_real"
  return 0
}

# COSMETIC ONLY, not part of the security check below (F24): strips any
# trailing slash(es) from $ARCHIVED_PATH purely so this script's own
# note/label text (which quotes $ARCHIVED_PATH verbatim) reads consistently.
# Has no effect on validate_archived_path()'s trust decision — its step-2
# normalization collapses a trailing slash on its own regardless.
while [ "${ARCHIVED_PATH%/}" != "$ARCHIVED_PATH" ] && [ "$ARCHIVED_PATH" != "/" ]; do
  ARCHIVED_PATH="${ARCHIVED_PATH%/}"
done

validate_archived_path

LEDGER_LABEL="docs/superpowers/ledgers/$NAME.md"
PANEL_LABEL="docs/superpowers/reviews/$NAME-panel.md"
TASKS_LABEL="$ARCHIVED_PATH/tasks.md"
GITLOG_LABEL="git log --stat"

# find_dated <dir> <suffix> — the most recent "<date>-<name><suffix>" file
# under <dir>, or empty. Anchored digit-by-digit exactly as
# preserve-session-records.sh's own search is: a bare "*-${NAME}${suffix}"
# would also match a DIFFERENT change whose name ends in this one.
find_dated() {
  local dir="$1" suffix="$2"
  [ -n "$dir" ] && [ -d "$dir" ] || return 0
  find "$dir" -maxdepth 1 \
    -name "[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]-${NAME}${suffix}" \
    2>/dev/null | sort | tail -1
}

# resolve_file <path> — print <path> with every symlink resolved, both on the
# final component and on each directory component. Ported verbatim from
# preserve-session-records.sh's own helper of the same name: `readlink -f`
# would do this in one call on GNU but is not portable to the BSD readlink
# macOS ships, so the final component is walked here and the directory
# components are left to `cd -P`. Fails on a symlink loop rather than
# spinning.
resolve_file() {
  local p="$1" hops=0 target dir
  while [ -L "$p" ]; do
    hops=$((hops + 1))
    [ "$hops" -le 40 ] || return 1
    target="$(readlink "$p")" || return 1
    case "$target" in
      /*) p="$target" ;;
      *) p="$(dirname "$p")/$target" ;;
    esac
  done
  dir="$(cd -P "$(dirname "$p")" 2>/dev/null && pwd -P)" || return 1
  printf '%s\n' "$dir/${p##*/}"
}

# check_boundary <candidate> <root> — Protections 2/3 ported from
# preserve-session-records.sh: a candidate file under a tracked, PR-editable
# directory (docs/superpowers/ledgers/, docs/superpowers/reviews/, or
# <archived-change-path>/tasks.md) may be a symlink planted to make `cat`
# read an arbitrary file outside the repository. Resolve it and verify the
# resolved path stays under <root>; refuse (not skip) it if not — refusing is
# reported separately so an attack is never indistinguishable from a source
# that is legitimately absent. Sets globals BOUNDARY_RESOLVED (the resolved
# path, on success) and BOUNDARY_REFUSED (1 iff refused). Called as a plain
# statement, never inside "$(...)" — a command substitution runs in a
# subshell, and array/variable writes there would not survive back to the
# caller.
check_boundary() {
  local candidate="$1" root="$2" resolved
  BOUNDARY_RESOLVED=""
  BOUNDARY_REFUSED=0
  if [ -z "$root" ]; then
    BOUNDARY_REFUSED=1
    return 0
  fi
  resolved="$(resolve_file "$candidate")" || { BOUNDARY_REFUSED=1; return 0; }
  if within_root "$resolved" "$root"; then
    BOUNDARY_RESOLVED="$resolved"
  else
    BOUNDARY_REFUSED=1
  fi
}

# is_refused <label> — true iff check_boundary already recorded this source
# as refused. A refused source must never also be reported "skipped: ...
# (absent)" — that would hide an attack as a normal, legitimate absence.
is_refused() {
  local label="$1" r
  [ "${#REFUSED[@]}" -eq 0 ] && return 1
  for r in "${REFUSED[@]}"; do
    [ "$r" = "$label" ] && return 0
  done
  return 1
}

REFUSED=()

LEDGER_FILE=""
PANEL_FILE=""
if [ -n "$REPO_ROOT" ]; then
  ledger_candidate="$(find_dated "$REPO_ROOT/docs/superpowers/ledgers" ".md")"
  if [ -n "$ledger_candidate" ]; then
    check_boundary "$ledger_candidate" "$REPO_ROOT"
    if [ "$BOUNDARY_REFUSED" -eq 1 ]; then
      REFUSED+=("$LEDGER_LABEL")
    else
      LEDGER_FILE="$BOUNDARY_RESOLVED"
    fi
  fi
  panel_candidate="$(find_dated "$REPO_ROOT/docs/superpowers/reviews" "-panel.md")"
  if [ -n "$panel_candidate" ]; then
    check_boundary "$panel_candidate" "$REPO_ROOT"
    if [ "$BOUNDARY_REFUSED" -eq 1 ]; then
      REFUSED+=("$PANEL_LABEL")
    else
      PANEL_FILE="$BOUNDARY_RESOLVED"
    fi
  fi
fi

TASKS_FILE=""
if [ "$ARCHIVED_PATH_INVALID" -eq 0 ]; then
  tasks_candidate="$ARCHIVED_PATH/tasks.md"
  if [ -f "$tasks_candidate" ]; then
    check_boundary "$tasks_candidate" "$ARCHIVED_REAL"
    if [ "$BOUNDARY_REFUSED" -eq 1 ]; then
      REFUSED+=("$TASKS_LABEL")
    else
      TASKS_FILE="$BOUNDARY_RESOLVED"
    fi
  fi
fi

# The three commits: the implementation commit and the "plan and session
# records" commit (finish run 1's own two-commit chain, per Git boundaries
# in skills/myflow-contracts/pipeline.md), plus the archive commit.
# Resolved by walking git log for the most recent commit matching
# each subject shape, never by a fixed commit count back from HEAD, so this
# still finds the right commits regardless of how much history has landed
# since. Grep is anchored on the subject line; NAME's only regex metacharacter
# a real allowlisted name can carry is '.', escaped below.
#
# This --max-count=5 grep-and-exclude approach assumes at most a small number
# of intervening commits between the implementation commit and the archive
# commit; a change with several /myflow-do fix-round commits may not resolve
# to exactly the "first" implementation commit. Accepted as low-impact: this
# git log content is advisory input to a reasoning pass, not load-bearing.
GITLOG_CONTENT=""
if [ -n "$REPO_ROOT" ]; then
  NAME_RE="$(printf '%s' "$NAME" | sed 's/\./\\./g')"
  ARCHIVE_SHA="$(git -C "$REPO_ROOT" log -E --grep="^chore\(${NAME_RE}\): .*archive" \
    --max-count=1 --format='%H' 2>/dev/null || true)"
  # The alternation matches both the current subject ("plan and session
  # records") and the subject finish run 1 used before it was renamed
  # ("plan, test guide and session records"), so changes committed under
  # either wording keep resolving. Do not drop the old branch: doing so
  # would leave a pre-rename change's plan commit unmatched here while
  # IMPL_SHA's own exclusion below (which must stay in sync with this
  # grep) still needs it — see that comment for the resulting failure mode.
  PLAN_SHA="$(git -C "$REPO_ROOT" log -E \
    --grep="^chore\(${NAME_RE}\): plan(, test guide and| and) session records" \
    --max-count=1 --format='%H' 2>/dev/null || true)"
  # Excluded only when the SUBJECT matches the same dedicated archive-commit
  # shape ARCHIVE_SHA itself searches for above — never a bare substring
  # check for "archive" — so a genuine implementation commit that happens to
  # mention "archive" (e.g. "feat(name): archive stale records") is not
  # wrongly excluded from IMPL_SHA candidacy.
  # This exclusion must stay in sync with PLAN_SHA's --grep above, alternation
  # for alternation: the plan commit is always the most recent chore(...) commit
  # ahead of the implementation commit, so if a wording matches PLAN_SHA but is
  # missing here, this exclusion fails to filter it out, `head -1` picks the
  # plan commit itself (it outranks the true implementation commit by recency),
  # and IMPL_SHA resolves to the plan commit too — a wrong answer, not a
  # missing one, since the implementation commit is then never reached at all.
  IMPL_SHA="$(git -C "$REPO_ROOT" log -E --grep="^(feat|fix|chore)\(${NAME_RE}\): " \
    --max-count=5 --format='%H %s' 2>/dev/null \
    | grep -Ev "^[0-9a-f]+ chore\(${NAME_RE}\): plan(, test guide and| and) session records|^[0-9a-f]+ chore\(${NAME_RE}\): .*archive" \
    | head -1 | cut -d' ' -f1 || true)"

  for sha in "$IMPL_SHA" "$PLAN_SHA" "$ARCHIVE_SHA"; do
    if [ -n "$sha" ]; then
      GITLOG_CONTENT="$GITLOG_CONTENT$(git -C "$REPO_ROOT" log --stat -1 "$sha" 2>/dev/null || true)
"
    fi
  done
fi

FOUND=()
SKIPPED=()

if is_refused "$LEDGER_LABEL"; then :
elif [ -n "$LEDGER_FILE" ]; then FOUND+=("$LEDGER_LABEL")
else SKIPPED+=("$LEDGER_LABEL"); fi

if is_refused "$PANEL_LABEL"; then :
elif [ -n "$PANEL_FILE" ]; then FOUND+=("$PANEL_LABEL")
else SKIPPED+=("$PANEL_LABEL"); fi

if is_refused "$TASKS_LABEL"; then :
elif [ -n "$TASKS_FILE" ]; then FOUND+=("$TASKS_LABEL")
else SKIPPED+=("$TASKS_LABEL"); fi

if [ -n "$GITLOG_CONTENT" ]; then FOUND+=("$GITLOG_LABEL"); else SKIPPED+=("$GITLOG_LABEL"); fi

n_found=${#FOUND[@]}
n_skipped=${#SKIPPED[@]}
n_refused=${#REFUSED[@]}

echo "# Self-review context bundle for $NAME"
echo
# F18: one lookup mapping ARCHIVED_PATH_REASON to its note message, rather
# than N near-duplicate if/elif branches each with their own echo. Three
# reasons now (F22 collapsed dotdot/symlink/missing/ancestor-symlink into
# "shape" or "symlink" — see validate_archived_path() above).
if [ "$ARCHIVED_PATH_INVALID" -eq 1 ]; then
  case "$ARCHIVED_PATH_REASON" in
    no-repo)
      echo "note: archived-change-path '$ARCHIVED_PATH' could not be validated — no enclosing git repository was found from this script's own working directory; every source below is reported skipped for that reason"
      ;;
    shape)
      echo "note: archived-change-path '$ARCHIVED_PATH' does not resolve to the expected openspec/changes/archive/<leaf> location — every source below is reported skipped for that reason"
      ;;
    symlink)
      echo "note: archived-change-path '$ARCHIVED_PATH' resolves through a symlink somewhere between the repository root and the leaf — refusing to use it as a trust boundary; every source below is reported skipped for that reason"
      ;;
  esac
  echo
fi
echo "found: $n_found of 4 sources; skipped: $n_skipped of 4 sources; refused: $n_refused of 4 sources"
if [ "$n_refused" -gt 0 ]; then
  for src in "${REFUSED[@]}"; do
    echo "refused: $src (resolves outside the repository)"
  done
fi
if [ "$n_skipped" -gt 0 ]; then
  for src in "${SKIPPED[@]}"; do
    echo "skipped: $src (absent)"
  done
fi
echo

if [ -n "$LEDGER_FILE" ]; then
  echo "## $LEDGER_LABEL"
  echo
  cat "$LEDGER_FILE"
  echo
fi

if [ -n "$PANEL_FILE" ]; then
  echo "## $PANEL_LABEL"
  echo
  cat "$PANEL_FILE"
  echo
fi

if [ -n "$TASKS_FILE" ]; then
  echo "## $TASKS_LABEL"
  echo
  cat "$TASKS_FILE"
  echo
fi

if [ -n "$GITLOG_CONTENT" ]; then
  echo "## $GITLOG_LABEL"
  echo
  printf '%s\n' "$GITLOG_CONTENT"
fi

exit 0
