#!/usr/bin/env bash
# gather-self-review-context.sh — deterministically collect the SDD ledger,
# the review-panel record, tasks.md and the relevant git log for a finished
# change, so the self-review reasoning pass (/myflow-finish run 2, step 9)
# judges a bundle instead of re-reading files itself.
#
# Usage: gather-self-review-context.sh <archived-change-path> <name> <state-dir> [<repo-root>]
#
# <archived-change-path> is the ARCHIVED change directory
# (spectre/changes/archive/<date>-<name>/), never a worktree path — by the
# time step 9 runs, run 2 has already removed the worktree. <state-dir> is
# accepted for CLI parity with this script's usage line and with the
# <worktree> <name> <state-dir> shape the /myflow-finish record helper took
# when this script was written — that helper has since been retired and its
# work moved into `myflow record render`, but the argument stays for callers
# that already pass it; none of the four sources below currently read from
# it. <repo-root> is OPTIONAL
# (KAN-239): see the dedicated NOTE below for what it is, why it exists, and
# why accepting it does not weaken this script's trust argument.
#
# Prints one bundle to stdout: a header line naming which sources were found
# vs. skipped, an explicit "skipped: <src> (absent)" line per missing source
# (1-3) — one distinct OUTCOME WORD per source on stdout rather than a
# distinct exit status, which is the convention the pipeline's own record
# steps follow; see the outcome table under **Rendering the session records**
# (`skills/myflow-contracts/session-records.md`) — and each found source's content
# under its own subheading.
# Exits 2 on a missing argument, an invalid change name, or a malformed
# <repo-root> (a malformed invocation, in every case); otherwise ALWAYS exits
# 0 — a missing source is never fatal, since a change may legitimately have
# no review-panel record, but a malformed invocation is. This script
# performs no judgement and makes no pass/fail determination based on what
# it finds.
#
# NOTE on <repo-root> (KAN-239): by the time /myflow-finish run 2 reaches
# step 9, run 2 has already removed the worktree that implemented the
# change (see the note on <archived-change-path> above), so this script's
# own process cwd can no longer be trusted to sit inside any git repository
# at all — on KAN-201 it did not, which made this script structurally
# unrunnable from step 9's actual cwd at that point. <repo-root> lets the
# caller pass the main checkout's own path explicitly instead of relying on
# process cwd to derive it. When absent, TRUSTED_REPO_ROOT is derived
# exactly as before (see validate_archived_path() step 1 below); when
# present, it BECOMES TRUSTED_REPO_ROOT directly, once validated by the same
# tests <archived-change-path> itself receives: absolute, lexically-
# collapsed form identical to its symlink-resolved form, and — reusing this
# script's own existing git-common-dir derivation, just anchored at the
# supplied path instead of at process cwd — the root of a git repository.
# Accepting a caller-supplied root does NOT weaken the containment argument
# validate_archived_path() makes below: the prohibition that argument
# defends is deriving the trust anchor from <archived-change-path> itself —
# the untrusted input under test, which a merged PR can shape freely — never
# accepting one from the caller, who is this script's own trusted invoker
# (skills/myflow-finish/SKILL.md step 9), not the untrusted change content.
# A future reader who does not find this stated may read the fourth argument
# as a regression of that argument and delete it; it is not one.
#
# NOTE on <archived-change-path> itself being a symlink, or any ancestor of
# it, at any depth (F12, F13, F14, F20, F22): git tracks symlinks as
# committed blobs, so a merged PR could make spectre/changes/archive/<date>-
# <name> itself a symlink to anywhere, or make ANY ancestor component
# (spectre/, spectre/changes/, spectre/changes/archive/, or a deeper
# nesting level not even part of the documented shape) a symlink instead.
# Four rounds of bounded, lexical, fixed-depth patches — a leaf-only check, a
# trailing-slash strip, an exactly-3-hop ancestor `dirname` walk, a literal
# `..`-component rejection — each closed the specific bypass the previous
# round's reviewer found, but the underlying design stayed fragile to any
# path shape one step outside what each patch anticipated: most simply, one
# extra directory nesting level under spectre/changes/archive/ defeats a
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
#      step 9) always invokes it with process cwd at the repository root,
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
#      "$TRUSTED_REPO_ROOT/spectre/changes/archive/<leaf>" — one path
#      segment past the trusted archive root, no more and no less. Checked as
#      a real path BOUNDARY via `within_root()`, not a string prefix.
#      Anything else (missing, extra nesting, wrong location entirely) is an
#      invalid invocation.
#   4. The SAME absolute path built in step 2 (relative $ARCHIVED_PATH
#      already joined onto TRUSTED_REPO_ROOT) is ALSO resolved
#      semantically, following every symlink at every component (the
#      `cd -P`-based real resolution this script already used) — never raw
#      process cwd, which could differ from TRUSTED_REPO_ROOT when the
#      optional <repo-root> override is in play (F3, this change's own
#      review panel). Call this ARCHIVED_REAL.
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
# NOTE on sources 1 and 2: `myflow record render` writes these under
# docs/superpowers/{ledgers,reviews}/ with a LEADING DATE, e.g.
# "2026-08-01-demo.md", never literally "<name>.md" — this script's messages
# still name the source using the plain "<name>.md" / "<name>-panel.md" form,
# matching the wording this repository's archived kan-23-myflow-self-review
# change's delta spec states verbatim in its scenarios, but the SEARCH below
# uses the real, date-prefixed filename shape so a source that exists is
# actually found.
#
# NOTE on the "skipped:" stream: the design doc originally said stderr; this
# was corrected to stdout to match that same kan-23-myflow-self-review delta
# spec (which has always said stdout) and to match the pipeline's own convention of
# printing every outcome word on stdout — `myflow record render` prints
# `rendered:`, `MISSING:` and `journalled:` there — since this script's own
# bundle is a single stdout document by design.
#
# THE CHANGE-NAME ALLOWLIST mirrors records.Destination's own
# (stats/internal/records/render.go): one leading alphanumeric, then letters,
# digits, '.', '_' and '-'. A name containing '/' or a glob metacharacter must
# not be used to build a path or to build the `find -name` pattern used to
# locate the dated ledger/panel files, for the same reasons that function's
# comment states.
set -euo pipefail

# within_root <resolved-path> <root> -> the path-boundary test, sourced from
# lib/within-root.sh rather than carried as an inline copy, since this
# script ships through the skills/*/scripts/ symlink farm
# (skills/myflow-fast/scripts/ and skills/myflow-finish/scripts/ both carry
# it, alongside their own `lib` symlink into scripts/lib/), which is
# exactly the criterion that file's own header states for when a guard may
# safely source it instead of carrying its own copy. SCRIPT_DIR is derived
# from this script's own location (which may itself be one of those
# symlinks) rather than assumed, so sourcing resolves correctly from either
# entry point.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/within-root.sh"
source "$SCRIPT_DIR/lib/lexical-normalize.sh"

ARCHIVED_PATH="${1:-}"
NAME="${2:-}"
STATE_DIR="${3:-}"
REPO_ROOT_ARG="${4:-}"

if [ -z "$ARCHIVED_PATH" ] || [ -z "$NAME" ] || [ -z "$STATE_DIR" ]; then
  echo "usage: gather-self-review-context.sh <archived-change-path> <name> <state-dir> [<repo-root>]" >&2
  exit 2
fi

case "$NAME" in
  [!A-Za-z0-9]* | *[!A-Za-z0-9._-]*)
    echo "gather-self-review-context: change name '$NAME' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'" >&2
    exit 2
    ;;
esac

# REPO_ROOT_OVERRIDE — the optional fourth argument (KAN-239; see the header
# NOTE on <repo-root>), validated here as an invocation error: exit 2, the
# same class as a missing argument or a malformed change name above — never
# the exit-0/skipped-source treatment validate_archived_path() below gives
# an invalid <archived-change-path>, because a caller-supplied repo-root is
# this script's own trusted input (skills/myflow-finish/SKILL.md step 9),
# not untrusted change content.
#
# Reuses the two mechanisms the script already owns rather than writing new
# ones: lexically_collapse() for the lexical half, and the same `cd -P`/
# `pwd -P` real-resolution used throughout this script for the symlink half.
# "root of a git repository" reuses this script's own git-common-dir
# derivation — identical to validate_archived_path() step 1 below, just
# anchored at the supplied path instead of at process cwd — so a worktree
# root is refused here exactly as it would be if it were process cwd,
# keeping both derivation paths agreeing on what "the repository root"
# means.
REPO_ROOT_OVERRIDE=""
if [ -n "$REPO_ROOT_ARG" ]; then
  case "$REPO_ROOT_ARG" in
    /*) : ;;
    *)
      echo "gather-self-review-context: repo-root '$REPO_ROOT_ARG' is not an absolute path" >&2
      exit 2
      ;;
  esac
  repo_root_lexical="$(lexically_collapse "$REPO_ROOT_ARG")"
  repo_root_real=""
  if [ -d "$REPO_ROOT_ARG" ]; then
    repo_root_real="$(cd -P "$REPO_ROOT_ARG" 2>/dev/null && pwd -P)" || repo_root_real=""
  fi
  if [ -z "$repo_root_real" ] || [ "$repo_root_lexical" != "$repo_root_real" ]; then
    echo "gather-self-review-context: repo-root '$REPO_ROOT_ARG' does not exist or resolves through a symlink" >&2
    exit 2
  fi
  repo_root_common_dir="$(git -C "$repo_root_real" rev-parse --git-common-dir 2>/dev/null || true)"
  repo_root_from_git=""
  if [ -n "$repo_root_common_dir" ]; then
    case "$repo_root_common_dir" in
      /*) repo_root_from_git="$(cd "$(dirname "$repo_root_common_dir")" && pwd -P 2>/dev/null || true)" ;;
      *) repo_root_from_git="$(cd "$repo_root_real/$(dirname "$repo_root_common_dir")" && pwd -P 2>/dev/null || true)" ;;
    esac
  fi
  if [ -z "$repo_root_from_git" ] || [ "$repo_root_from_git" != "$repo_root_real" ]; then
    echo "gather-self-review-context: repo-root '$REPO_ROOT_ARG' is not the root of a git repository" >&2
    exit 2
  fi
  REPO_ROOT_OVERRIDE="$repo_root_real"
fi

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

  # Step 1: TRUSTED_REPO_ROOT. When the caller supplied <repo-root>
  # (REPO_ROOT_OVERRIDE, already validated above — KAN-239, see the header
  # NOTE), that path IS TRUSTED_REPO_ROOT; no derivation from process cwd
  # happens at all in that case. Otherwise it is derived from this script's
  # own process cwd, independent of $ARCHIVED_PATH — never
  # `-C "$ARCHIVED_PATH"`. Resolved via `--git-common-dir` (see header NOTE,
  # F23) rather than `--show-toplevel` so this is correct whether cwd is the
  # main checkout or a worktree.
  local trusted_root common_dir
  if [ -n "$REPO_ROOT_OVERRIDE" ]; then
    trusted_root="$REPO_ROOT_OVERRIDE"
  else
    common_dir="$(git rev-parse --git-common-dir 2>/dev/null || true)"
    if [ -n "$common_dir" ]; then
      trusted_root="$(cd "$(dirname "$common_dir")" && pwd -P 2>/dev/null || true)"
    else
      trusted_root=""
    fi
  fi
  if [ -z "$trusted_root" ]; then
    ARCHIVED_PATH_INVALID=1
    ARCHIVED_PATH_REASON="no-repo"
    return 0
  fi

  # Step 2: lexically normalize $ARCHIVED_PATH into an absolute path — a
  # relative path is joined to $trusted_root, never to a separately-read
  # $PWD. This is NOT merely because this script's documented caller
  # happens to run it with cwd at the repo root (F3, this change's own
  # review panel: that was once this step's stated justification, but the
  # optional <repo-root> override — see the header NOTE — exists precisely
  # so a caller can invoke this script from a cwd that is NOT inside the
  # repository at all, which makes that justification false). The real
  # reason is that step 4 below resolves this SAME relative $ARCHIVED_PATH
  # from this SAME $abs_input, not from raw process cwd — so both the
  # lexical and the semantic resolution always share one base, and can
  # never diverge just because cwd and $trusted_root differ. Joining to
  # $trusted_root also sidesteps a spurious mismatch when cwd sits under an
  # OS-level path alias such as macOS's /tmp -> /private/tmp, since `git
  # rev-parse` resolves that alias and a raw $PWD read would not.
  # `.`/`..` components are then collapsed by lexically_collapse(), sourced
  # from lib/lexical-normalize.sh (F34, this change's own review panel)
  # rather than carried here — the walk itself (push each `/`-separated
  # segment, pop on `..` never past the root, skip `.` and empty segments,
  # via `IFS=/ read -ra` per F26 rather than an unquoted `for` that would
  # let a `*`/`?`/`[...]` in $ARCHIVED_PATH glob against the process cwd)
  # was byte-for-byte identical to gather-dispatch-context.sh's own
  # lexical_normalize(); only the "what does a relative path join onto"
  # half above is genuinely specific to this script, and that half stays
  # here.
  local abs_input
  case "$ARCHIVED_PATH" in
    /*) abs_input="$ARCHIVED_PATH" ;;
    *) abs_input="$trusted_root/$ARCHIVED_PATH" ;;
  esac
  local archived_lexical
  archived_lexical="$(lexically_collapse "$abs_input")"

  # Step 3: ARCHIVED_LEXICAL must sit exactly one segment past
  # $trusted_root/spectre/changes/archive/ — a real path boundary
  # (within_root), not a string prefix, and exactly one segment: no deeper
  # nesting (F22), no shallower.
  local archive_root="$trusted_root/spectre/changes/archive"
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
  # every component. Resolved from $abs_input (step 2's absolute form), NOT
  # from raw $ARCHIVED_PATH (F3, this change's own review panel): a relative
  # $ARCHIVED_PATH was joined onto $trusted_root above, so the semantic
  # resolution here must start from that same base. Resolving raw
  # $ARCHIVED_PATH instead would follow the script's actual process cwd
  # rather than $trusted_root — identical only while cwd truly sits at the
  # repository root, which the optional <repo-root> override (see the
  # header NOTE) exists precisely to let the caller NOT be true. When they
  # diverge, ARCHIVED_REAL comes back empty and every source is silently
  # reported skipped instead of found. $abs_input is already absolute
  # either way, so an already-absolute $ARCHIVED_PATH resolves exactly as
  # before.
  local archived_real
  archived_real=""
  if [ -d "$abs_input" ]; then
    archived_real="$(cd -P "$abs_input" 2>/dev/null && pwd -P)" || archived_real=""
  fi
  if [ -z "$archived_real" ]; then
    # A leaf that IS a symlink but resolves to something other than a
    # directory (e.g. a symlink to a plain file) fails the `-d` test above
    # just like a missing path does — but it is not a shape problem, it's
    # exactly the same trust-boundary problem as a symlink pointing outside
    # the repo (F25): report it as such rather than as "shape".
    if [ -L "$abs_input" ]; then
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
# records.existingDatedFile's own search is (stats/internal/records/render.go,
# datedFilePrefix): a bare "*-${NAME}${suffix}" would also match a DIFFERENT
# change whose name ends in this one.
find_dated() {
  local dir="$1" suffix="$2"
  [ -n "$dir" ] && [ -d "$dir" ] || return 0
  find "$dir" -maxdepth 1 \
    -name "[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]-${NAME}${suffix}" \
    2>/dev/null | sort | tail -1
}

# resolve_file <path> — print <path> with every symlink resolved, both on the
# final component and on each directory component. Ported verbatim from the
# helper of the same name in the record-copying script this repository has
# since retired: `readlink -f`
# would do this in one call on GNU but is not portable to the BSD readlink
# macOS ships, so the final component is walked here and the directory
# components are left to `cd -P`. Fails on a symlink loop rather than
# spinning.
#
# Kept inline here rather than sourced from scripts/lib/resolve-file.sh
# (unlike within_root above) -- not because this script fails that file's
# own sourcing criterion (it ships through the farm exactly like
# within_root does, per that file's header), but because migrating it was
# left out of scope for the change that migrated within_root. See
# scripts/lib/resolve-file.sh's own header for the full explanation and
# the open follow-up question it records (F37, pass 7 of this change's
# own review panel).
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

# check_boundary <candidate> <root> — the READ half of the path-boundary pair
# this script inherited from the record-copying step, kept after that step
# became a render: a candidate file under a tracked, PR-editable
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

# is_real_impl_commit <PLAN_PARENT> — true iff PLAN_SHA's first parent
# qualifies as the implementation commit (pass 2, finding G). Reads globals
# REPO_ROOT (set above) and PLAN_SUBJECT_RE_NEW / PLAN_SUBJECT_RE_OLD /
# ARCHIVE_SUBJECT_RE (assigned once, just before this function's caller).
# Extracted so a future added condition does not have to be threaded into an
# already-long inline `if` — each condition below was added in response to a
# real, demonstrated bug and is kept exactly as it was, comments included.
# THREE OF THESE FOUR CONDITIONS ARE DELIBERATELY DEAD TODAY. Mutation
# testing confirms PARENT_TOUCHES_OUTSIDE alone decides every case
# reachable through this pipeline's own commits: dropping the merge gate,
# any one subject rejection, or all four together fails no test. That is
# the documented, expected state -- NOT a coverage gap, and NOT a licence
# to delete them. Each is retained as a named guard against a specific,
# plausible future edit to code this function does not control:
#
#   - the merge gate guards this function's OWN diff-tree call below.
#     Adding -m, -c or --cc to it -- the natural way to extend merge or
#     root-commit handling later -- would make diff-tree report real paths
#     for a merge commit, at which point a merge could satisfy
#     PARENT_TOUCHES_OUTSIDE and this gate becomes load-bearing again.
#   - the plan-subject rejections guard against a change to
#     commit-split.sh's staging order or exclusion pathspec letting the
#     plan commit's `git add -A` pick up a file outside the planning
#     trees. They also guard a collision this very change created: the
#     planning subject is now the fixed literal
#     "chore(openspec): plan and session records", IDENTICAL across every
#     change, so a skipped implementation commit can leave PLAN_PARENT
#     pointing at a DIFFERENT change's planning commit, with no name left
#     in the subject to tell them apart.
#   - the archive-subject rejection guards against a change to run 2's
#     archive step touching a path outside spectre/.
#
# Every one of those guarantees lives in another file, with no test
# coupling this function to it, so a future edit there would silently
# reactivate the confident-wrong-answer failure this function exists to
# prevent. Do not delete these on the strength of a mutation pass alone.
is_real_impl_commit() {
  local PLAN_PARENT="$1"
  local PARENT_SUBJECT
  PARENT_SUBJECT="$(git -C "$REPO_ROOT" log -1 --format='%s' "$PLAN_PARENT" 2>/dev/null || true)"

  # Non-merge: commit-split.sh's own commits always carry exactly one
  # parent. "${PLAN_PARENT}^2" resolving at all means a second parent
  # exists, i.e. PLAN_PARENT is a merge commit.
  local PARENT_IS_MERGE=0
  if git -C "$REPO_ROOT" rev-parse --verify -q "${PLAN_PARENT}^2" >/dev/null 2>&1; then
    PARENT_IS_MERGE=1
  fi

  # Touches at least one path outside the planning paths — what
  # commit-split.sh guarantees of a real implementation commit, and
  # what an unrelated commit that merely happens to sit just ahead of
  # the plan commit (a merge, or anything else) is not guaranteed to
  # do. Skipped entirely for a merge commit, which is refused on its
  # own above regardless of what it touches.
  #
  # Resolved via `git diff-tree --no-commit-id --name-only -r --root`,
  # NOT `git diff --name-only "${PLAN_PARENT}^..${PLAN_PARENT}"`: the
  # latter fails with "fatal: ambiguous argument" when PLAN_PARENT is
  # the repository's own root commit (a root commit has no `^` parent
  # to diff against), and the `2>/dev/null` on that call swallowed the
  # failure silently, leaving PARENT_TOUCHES_OUTSIDE at its default 0
  # and dropping a genuine implementation commit — a wrong answer
  # produced by a silenced error, the same failure class this file
  # keeps getting bitten by. `--root` makes diff-tree diff a root
  # commit against the empty tree instead of erroring, while leaving a
  # normal (non-root) commit's diff against its actual parent
  # unchanged, so one form is correct for both cases.
  local PARENT_TOUCHES_OUTSIDE=0
  if [ "$PARENT_IS_MERGE" -eq 0 ] \
    && git -C "$REPO_ROOT" diff-tree --no-commit-id --name-only -r --root "$PLAN_PARENT" 2>/dev/null \
      | grep -Ev '^(spectre/|docs/superpowers/)' | grep -q .; then
    PARENT_TOUCHES_OUTSIDE=1
  fi

  # The repeated "$PARENT_IS_MERGE" -eq 0 check below is a deliberate,
  # redundant no-op: PARENT_TOUCHES_OUTSIDE can only be 1 when that same
  # condition already held above, so this line's copy can never itself
  # flip the outcome (confirmed by mutation testing — removing it here
  # alone changes nothing). Kept anyway as defence-in-depth against a
  # future edit to the PARENT_TOUCHES_OUTSIDE computation above that
  # stops implying PARENT_IS_MERGE=0 — without this copy, such an edit
  # would silently let a merge commit's parent become IMPL_SHA. See the
  # block above this function for why the subject rejections below are
  # likewise dead today and likewise kept.
  if [ "$PARENT_IS_MERGE" -eq 0 ] \
    && [ "$PARENT_TOUCHES_OUTSIDE" -eq 1 ] \
    && [[ ! "$PARENT_SUBJECT" =~ $PLAN_SUBJECT_RE_NEW ]] \
    && [[ ! "$PARENT_SUBJECT" =~ $PLAN_SUBJECT_RE_OLD ]] \
    && [[ ! "$PARENT_SUBJECT" =~ $ARCHIVE_SUBJECT_RE ]]; then
    return 0
  fi
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
  # Built from $ARCHIVED_REAL, NOT raw $ARCHIVED_PATH (F3, this change's own
  # review panel): $ARCHIVED_PATH may still be relative here, and testing
  # it directly (`-f "$ARCHIVED_PATH/tasks.md"`) would resolve against this
  # script's actual process cwd rather than $ARCHIVED_REAL's already-
  # trusted, already-resolved base — the same divergence
  # validate_archived_path() itself was fixed against, just one level
  # further downstream. $ARCHIVED_REAL is absolute and known-safe whenever
  # ARCHIVED_PATH_INVALID=0.
  tasks_candidate="$ARCHIVED_REAL/tasks.md"
  if [ -f "$tasks_candidate" ]; then
    check_boundary "$tasks_candidate" "$ARCHIVED_REAL"
    if [ "$BOUNDARY_REFUSED" -eq 1 ]; then
      REFUSED+=("$TASKS_LABEL")
    else
      TASKS_FILE="$BOUNDARY_RESOLVED"
    fi
  fi
fi

# The three commits: the implementation commit and the planning commit
# (finish run 1's own two-commit chain, per Git boundaries in
# skills/myflow-contracts/git-boundaries.md), plus the archive commit.
#
# PLAN_SHA is resolved by PATH **AND** SUBJECT SHAPE together (pass 2,
# finding E): the plan commit is the most recent commit that both (a)
# touched THIS change's own spectre/changes/<name>/ directory and (b)
# carries a plan-commit subject — PLAN_SUBJECT_RE_NEW, the module-scope
# convention's fixed literal "chore(openspec): plan and session records",
# or PLAN_SUBJECT_RE_OLD, either wording finish run 1 used before that
# rename. Path alone is not commit-specific: verified directly, a LATER
# commit that merely touches the same directory (e.g. a typo fix landed
# after archiving) outranks the real planning commit by recency and wins a
# path-only `head -1`, making the real planning commit disappear from the
# bundle entirely. Subject alone is not change-specific either, since the
# module-scope convention made every change's planning commit read the
# identical literal subject (a subject-only match can return a DIFFERENT
# change's planning commit). Together they are both.
#
# Only the LIVE pathspec (spectre/changes/<name>/) is searched (pass 2,
# finding F) — NOT ALSO the archived location
# (spectre/changes/archive/<date>-<name>/, where run 2 later moves it):
# `git log -- <path>` filters each commit by its OWN historical tree, so
# the live pathspec alone finds the planning commit even after run 2's
# `git mv` renames the directory — verified directly. A second, archived
# pathspec earned nothing and only widened what a later, unrelated commit
# touching the archived directory could match instead — exactly finding
# E's own reproduction. Combined with the subject filter above, the prior
# archive-shape EXCLUSION grep (needed only when path-only matching could
# select the archive commit itself) is now unnecessary too: the archive
# commit's own subject never matches a plan subject, so it was never a
# candidate under the new filter in the first place.
#
# IMPL_SHA is then DERIVED from PLAN_SHA, never searched for independently:
# commit-split.sh makes the implementation commit and the planning commit
# BACK TO BACK, in that order, so the implementation commit is the planning
# commit's first parent. Either commit is skipped, not failed, when nothing
# is staged (commit-split.sh's own header) — when the implementation commit
# was skipped, the planning commit's parent is not one, and accepting it
# anyway is a confident wrong answer waiting to happen (pass 2, finding G:
# verified, a merge commit — "Merge pull request #42 from someone/some-
# other-change" — was reported as this change's own implementation commit,
# since it matched none of the three reserved subject shapes this guard
# used to check alone). The parent is now accepted as IMPL_SHA only when it
# is BOTH a non-merge commit (exactly one parent — every commit-split.sh
# commit is) AND touches at least one path outside spectre/ and
# docs/superpowers/ (what a real implementation commit is guaranteed to do,
# per commit-split.sh's own boundary, and what some unrelated commit
# sitting just ahead of the plan commit is not), on top of the existing
# reserved-subject-shape rejections. This is still a heuristic, not a
# certainty — a hand-crafted commit could still slip past it — but it
# closes the "anything not named is accepted" gap the merge-commit case
# exploited.
#
# There is deliberately NO subject-only fallback for a change finished
# before this landed. One was tried and removed: it assumed a pre-KAN-202
# planning commit's own historical tree "may not even touch"
# spectre/changes/<name>/, which is false — that IS commit-split.sh's own
# planning commit, in either era, so it always touches that path. Verified
# directly against this repository's own archived history: the planning
# commits for kan-201-, kan-209-, kan-102- and kan-236- all resolve through
# the PRIMARY path+subject query above, because PLAN_SUBJECT_RE_OLD is
# already one of that query's --grep alternatives and the live pathspec
# matches. A subject-only, no-path-filter fallback was therefore
# unreachable for any genuine historical commit — the only way to reach it
# was a commit carrying an old-era subject while touching no path at all,
# which is not what a real planning commit looks like; it was only ever
# exercised by a test fixture built with `git commit --allow-empty` and
# nothing staged.
#
# ARCHIVE_SHA is UNCHANGED by any of this: run 2's archive commit has always
# used, and still uses, `chore(<name>): ... archive ...` — a subject grep
# stays exact and change-specific there, so it is left exactly as it was.
# NAME's only regex metacharacter a real allowlisted name can carry is '.',
# escaped below.
#
# ARCHIVE_SUBJECT_RE / PLAN_SUBJECT_RE_NEW / PLAN_SUBJECT_RE_OLD are each
# assigned ONCE, immediately below, and reused at every site that needs
# one of these shapes (pass 2, finding H) — each was previously hand-copied
# at several call sites, which this file's own comments already warned had
# to be kept in sync by hand; assigning once removes the chance to drift.
#
GITLOG_CONTENT=""
if [ -n "$REPO_ROOT" ]; then
  NAME_RE="$(printf '%s' "$NAME" | sed 's/\./\\./g')"

  ARCHIVE_SUBJECT_RE="^chore\\(${NAME_RE}\\): .*archive"
  PLAN_SUBJECT_RE_NEW='^chore\(openspec\): plan and session records'
  PLAN_SUBJECT_RE_OLD="^chore\\(${NAME_RE}\\): plan(, test guide and| and) session records"

  ARCHIVE_SHA="$(git -C "$REPO_ROOT" log -E --grep="$ARCHIVE_SUBJECT_RE" \
    --max-count=1 --format='%H' 2>/dev/null || true)"

  # PLAN_SHA: path AND subject shape together (findings E, F above). Only
  # the LIVE pathspec is searched; --grep is repeated deliberately — git
  # ORs multiple --grep patterns together by default (no --all-match) — so
  # either the new fixed literal or an old-era wording qualifies, whichever
  # this particular commit carries.
  PLAN_PATH_LIVE="spectre/changes/$NAME"
  PLAN_SHA="$(git -C "$REPO_ROOT" log -E \
    --grep="$PLAN_SUBJECT_RE_NEW" --grep="$PLAN_SUBJECT_RE_OLD" \
    --max-count=1 --format='%H' \
    -- "$PLAN_PATH_LIVE" 2>/dev/null || true)"

  # IMPL_SHA derived from PLAN_SHA's first parent (finding G): accepted
  # only when that parent is a non-merge commit touching at least one path
  # outside spectre/ and docs/superpowers/, and its subject matches none
  # of the three reserved plan-/archive-shapes above. Anything else
  # resolves NOTHING rather than a confident wrong answer. The four
  # conditions themselves live in is_real_impl_commit, defined above.
  IMPL_SHA=""
  if [ -n "$PLAN_SHA" ]; then
    PLAN_PARENT="$(git -C "$REPO_ROOT" rev-parse "${PLAN_SHA}^" 2>/dev/null || true)"
    if [ -n "$PLAN_PARENT" ] && is_real_impl_commit "$PLAN_PARENT"; then
      IMPL_SHA="$PLAN_PARENT"
    fi
  fi

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
      echo "note: archived-change-path '$ARCHIVED_PATH' does not resolve to the expected spectre/changes/archive/<leaf> location — every source below is reported skipped for that reason"
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
