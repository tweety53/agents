#!/usr/bin/env bash
# gather-dispatch-context.sh — deterministically collect one change's
# planning context into a single bundle, so every dispatch in a /myflow-do
# stage (each panel slot, each implementer, each fix subagent) is given
# demonstrably identical inputs instead of separately locating and reading
# proposal.md, design.md, tasks.md and the engineering principles on its
# own.
#
# Usage: gather-dispatch-context.sh <worktree> <change-root> <name> <principles-path>
#
# <worktree> and <change-root> are absolute paths; <change-root> is expected
# to sit under <worktree> (spectre/changes/<name>/, but this script does not
# assume that literal shape — it only requires containment). <principles-path>
# is the absolute path of engineering-principles.md, resolved by the CALLER
# (skills/myflow-do/SKILL.md's own [PRINCIPLES_PATH] rule) and never derived
# here — it may legitimately sit OUTSIDE <worktree> entirely, e.g. under a
# global install (~/.claude/skills/myflow-do/engineering-principles.md), so
# it is validated but never checked for containment under <worktree>.
#
# Prints one bundle to stdout: a header naming the change, the generating
# instant and the worktree's HEAD sha; a found/skipped/refused census line;
# one "refused: <src> (resolves outside the change directory)" line per source
# whose leaf resolves outside <change-root> (see LEAF VALIDATION below), one
# "skipped: <src> (absent)" line per missing source; then each found source
# under its own "## " heading, in this fixed order: proposal.md, design.md,
# tasks.md, then the principles file.
#
# THERE IS NO SPEC SOURCE, AND ADDING ONE BACK WOULD BE A MISTAKE. This
# script used to also carry every <change-root>/specs/*/spec.md — the
# OpenSpec delta specs. Under spectre a change has no specs/ subdirectory at
# all: spec edits go straight into <project>/spectre/specs/<capability>.md on
# the change's own branch, so there was nothing left for that glob to find
# and it reported "skipped: specs/*/spec.md (absent)" on every run of every
# change. The capability specs themselves are NOT bundled in its place —
# which ones a change touches is a judgement from its proposal, and this
# script is deterministic collection, not judgement; skills/myflow-do/SKILL.md
# names them for the reader instead.
#
# `## standards` entries are NEVER read or carried here — the spec's own
# Requirement forbids it: they resolve through the entry-form table and
# containment rule in skills/flow-contracts/project-configuration.md,
# belong to the principles slot alone, and re-implementing that containment
# logic here would duplicate the contract it depends on. This script globs
# nothing anywhere under <change-root>: it reads three leaves by exact name,
# so any other file placed there (a standards file, a stray note) never
# reaches the bundle.
#
# Exit 2 on a malformed invocation: a missing argument, a change name outside
# the allowlist below, or any of the three paths failing validation (see
# below). Exit 0 in every other case, including every missing source — a
# change may legitimately carry no design.md, and this script performs no
# judgement about that.
#
# VALIDATION. Modelled on scripts/gather-self-review-context.sh's own
# mechanism — read that script's header before touching this one. Each of
# <worktree>, <change-root> and <principles-path> is:
#   1. lexically normalized into an absolute path by pure string
#      manipulation (a component-stack walk: push each `/`-separated
#      segment, pop on `..`, never past the root, skip `.` and empty
#      segments) — touching the filesystem nowhere, so it cannot itself be
#      fooled by a symlink;
#   2. semantically resolved by following every symlink at every component
#      (the leaf, via a bounded readlink loop, and every ancestor directory,
#      via `cd -P`), the resolver always given the LEXICAL form from step 1
#      rather than the raw argument, so a bare "." or a trailing slash —
#      neither of which step 1 ever produces — can never itself masquerade
#      as a divergence;
#   3. compared for EXACT equality against the lexical form.
# For <worktree> and <change-root>, any divergence found in step 3 — an
# ancestor symlink at any depth, the leaf itself being a symlink, a `..`
# component, in any combination — means the invocation is refused, exactly
# as gather-self-review-context.sh's header documents: one general
# mechanism rather than several bounded, individually-bypassable checks.
# They must additionally resolve to an existing directory (not merely pass
# the lexical/real comparison), and <change-root> must resolve INSIDE
# <worktree> — a change-root outside the repository is a malformed
# invocation, not a legitimate absence.
# <principles-path> is normalized and resolved the same way but is NEVER
# refused for a divergence found in step 3, and carries no existence or
# containment requirement: it may be absent (reported as a skipped source,
# exit 0) or sit outside <worktree> (the global-install case above) or be
# reached through a symlink at any depth — a global install reaches every
# skill, and therefore every principles-path, through one by construction.
# It is a trusted argument the CALLER resolves (skills/myflow-do/SKILL.md's
# own [PRINCIPLES_PATH] rule), never attacker-influenced content the way
# <change-root>'s leaves are, so it is simply resolved and read.
#
# THE CHANGE-NAME ALLOWLIST is copied verbatim from
# gather-self-review-context.sh: one leading alphanumeric, then letters,
# digits, '.', '_' and '-'.
#
# LEAF VALIDATION. The three CONTENT sources (proposal.md, design.md and
# tasks.md) are never passed as arguments, so the
# exit-2 VALIDATION above does not apply to them — a change legitimately
# carries no design.md, and a symlinked leaf must not abort a whole
# dispatching stage the way a malformed invocation does. Each is instead
# resolved and boundary-checked against <change-root> the same way
# gather-self-review-context.sh's check_boundary() validates its own leaves
# (LEDGER_FILE, PANEL_FILE, TASKS_FILE) before reading them: a leaf that
# resolves outside <change-root> is refused per-source — omitted from the
# bundle, reported as "refused", but the run still exits 0 — never treated
# as a missing source, so an attack is never indistinguishable from a
# legitimate absence. See add_fixed_source() below.
set -euo pipefail

# resolve_file <path> -> the path's resolved PHYSICAL location, following
# every symlink (leaf and every ancestor). Sourced from lib/resolve-file.sh,
# and within_root <resolved-path> <root> -> the path-boundary test, sourced
# from lib/within-root.sh — both rather than carried as inline copies, since
# this script ships through the skills/*/scripts/ symlink farm
# (skills/myflow-do/scripts/ and skills/myflow-fast/scripts/ both carry it,
# alongside their own `lib` symlink into scripts/lib/), which is exactly the
# criterion each library file's own header states for when a guard may
# safely source a sibling instead of carrying its own copy. SCRIPT_DIR is
# derived from this script's own location (which may itself be one of those
# symlinks) rather than assumed, so sourcing resolves correctly from either
# entry point.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/resolve-file.sh"
source "$SCRIPT_DIR/lib/within-root.sh"
source "$SCRIPT_DIR/lib/lexical-normalize.sh"

WORKTREE="${1:-}"
CHANGE_ROOT="${2:-}"
NAME="${3:-}"
PRINCIPLES_PATH="${4:-}"

if [ -z "$WORKTREE" ] || [ -z "$CHANGE_ROOT" ] || [ -z "$NAME" ] || [ -z "$PRINCIPLES_PATH" ]; then
  echo "usage: gather-dispatch-context.sh <worktree> <change-root> <name> <principles-path>" >&2
  exit 2
fi

case "$NAME" in
  [!A-Za-z0-9]* | *[!A-Za-z0-9._-]*)
    echo "gather-dispatch-context: change name '$NAME' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'" >&2
    exit 2
    ;;
esac

# lexical_normalize <path> — print <path> as an absolute, lexically-collapsed
# path, built by pure string manipulation with no filesystem access. A
# relative <path> is joined to this process's own cwd first; the collapse
# itself (removing `.` and `..` components) is lexically_collapse(), sourced
# from lib/lexical-normalize.sh (F34, this change's own review panel) rather
# than carried here, since gather-self-review-context.sh's
# validate_archived_path() needs the identical collapse, just joined to a
# different root first (see that file's own header for the full
# explanation of why only the "make it absolute" half stays local to each
# caller).
lexical_normalize() {
  local input="$1" abs
  case "$input" in
    /*) abs="$input" ;;
    *) abs="$(pwd)/$input" ;;
  esac
  lexically_collapse "$abs"
}

# validate_path <path> — the one mechanism the header above documents:
# normalize, resolve, compare. The "resolve" half is resolve_file(), sourced
# above from lib/resolve-file.sh — it already walks the leaf's symlinks (a
# bounded readlink loop) and every ancestor directory (`cd -P`, following the
# physical filesystem structure regardless of how many ancestor components
# exist), and it is `--`-hardened against a path beginning with `-` where the
# inline copy this replaced was not. Sets:
#   VALIDATE_LEXICAL  — the lexically-normalized path (always set)
#   VALIDATE_REAL     — the fully-resolved path, or "" when an ancestor in
#                        the chain does not exist (a plain absence)
#   VALIDATE_MISMATCH — 1 iff VALIDATE_REAL is non-empty and differs from
#                        VALIDATE_LEXICAL, i.e. a symlink sits somewhere
#                        between an existing ancestor and the leaf
validate_path() {
  local path="$1"
  VALIDATE_LEXICAL="$(lexical_normalize "$path")"
  VALIDATE_REAL=""
  # resolve_file() is given the already lexically-normalized form, not the
  # raw argument (F2): resolve_file's dirname/basename split
  # (lib/resolve-file.sh:56) mishandles a bare "." (dirname "." -> ".",
  # basename "." -> ".", so the result gains a spurious trailing "/.") and a
  # trailing-slash path (basename "" -> the result gains a spurious trailing
  # "/") — neither of which lexical_normalize() ever produces, so either
  # form was wrongly compared as a "divergence" against a clean lexical
  # path even with no symlink anywhere in the chain. VALIDATE_LEXICAL is
  # always absolute with no "." segment and no trailing slash, so
  # resolve_file() never sees either shape here. resolve_file() itself now
  # normalizes both shapes on its own input too (F9, lib/resolve-file.sh's
  # own "NORMALIZES ITS OWN INPUT" comment), so this pre-normalization is
  # redundant belt-and-braces here, not the only thing standing between
  # this function and the bug.
  VALIDATE_REAL="$(resolve_file "$VALIDATE_LEXICAL" 2>/dev/null)" || VALIDATE_REAL=""
  VALIDATE_MISMATCH=0
  if [ -n "$VALIDATE_REAL" ] && [ "$VALIDATE_REAL" != "$VALIDATE_LEXICAL" ]; then
    VALIDATE_MISMATCH=1
  fi
}

# --- worktree: must resolve to an existing, non-symlinked directory ---
validate_path "$WORKTREE"
if [ -z "$VALIDATE_REAL" ] || [ ! -d "$VALIDATE_REAL" ]; then
  echo "gather-dispatch-context: worktree '$WORKTREE' does not resolve to an existing directory" >&2
  exit 2
fi
if [ "$VALIDATE_MISMATCH" -eq 1 ]; then
  echo "gather-dispatch-context: worktree '$WORKTREE' resolves through a symlink somewhere between the repository root and the leaf" >&2
  exit 2
fi
WORKTREE_REAL="$VALIDATE_REAL"

# --- change-root: must resolve to an existing, non-symlinked directory
# inside the (already-validated) worktree ---
validate_path "$CHANGE_ROOT"
if [ -z "$VALIDATE_REAL" ] || [ ! -d "$VALIDATE_REAL" ]; then
  echo "gather-dispatch-context: change-root '$CHANGE_ROOT' does not resolve to an existing directory" >&2
  exit 2
fi
if [ "$VALIDATE_MISMATCH" -eq 1 ]; then
  echo "gather-dispatch-context: change-root '$CHANGE_ROOT' resolves through a symlink somewhere between the repository root and the leaf" >&2
  exit 2
fi
CHANGE_ROOT_REAL="$VALIDATE_REAL"
if ! within_root "$CHANGE_ROOT_REAL" "$WORKTREE_REAL"; then
  echo "gather-dispatch-context: change-root '$CHANGE_ROOT' resolves outside the worktree '$WORKTREE'" >&2
  exit 2
fi

# --- principles-path: normalized and resolved the same way, but never
# required to exist, never checked for containment under the worktree (it
# may be a global install path outside it entirely), and NEVER refused for
# a symlink divergence (F1). setup.sh global installs every skill as a
# symlink (~/.claude/skills/myflow-do -> the repository's skills/myflow-do/),
# so [PRINCIPLES_PATH] — resolved by the CALLER from its own installed
# skill directory (skills/myflow-do/SKILL.md's own rule), never derived
# here — ALWAYS diverges lexically from its real path in that install
# shape. The lexical/real divergence refusal exists to stop a change
# directory's own (repo-tracked, pull-request-editable, attacker-
# influenced) content escaping <change-root>; principles-path is not that —
# it is a trusted argument the dispatching skill resolves for itself, so it
# is simply resolved (following every symlink, leaf and ancestor alike) and
# read. Containment and divergence refusal are UNCHANGED for <worktree> and
# <change-root> above, and for the five leaf content sources below. ---
validate_path "$PRINCIPLES_PATH"
PRINCIPLES_REAL="$VALIDATE_REAL"

# ===========================================================================
# The bundle itself
# ===========================================================================

PROPOSAL_FILE="$CHANGE_ROOT_REAL/proposal.md"
DESIGN_FILE="$CHANGE_ROOT_REAL/design.md"
TASKS_FILE="$CHANGE_ROOT_REAL/tasks.md"

FOUND_LABELS=()
FOUND_PATHS=()
SKIPPED_LABELS=()
REFUSED_LABELS=()
# REFUSED_REASONS runs parallel to REFUSED_LABELS (bash 3.2 has no
# associative arrays) — "outside" for a genuine within_root containment
# breach, "unresolvable" for a resolve_file() failure (a symlink loop past
# its own 40-hop cap, or a component that could not be cd'd into, e.g. a
# race or a permissions failure) (F36, pass 7 of this change's own review
# panel). Before this, both cases were folded into one REFUSED_LABELS
# bucket and every one was reported with the same "(resolves outside the
# change directory)" text below — true for the first case, and simply
# wrong for the second: a source resolve_file() could not even walk was
# never compared against CHANGE_ROOT_REAL at all, so nothing about it says
# "outside". Both dispositions still omit the source from the bundle and
# still let the run exit 0 — only the printed diagnostic differs.
REFUSED_REASONS=()

# add_fixed_source <path> <label> — unlike <worktree>/<change-root>/
# <principles-path> above, these three content sources are never passed as
# arguments, so validate_path()'s exit-2-on-mismatch contract does not apply
# to them: a change legitimately carries no design.md, and "this leaf is a
# symlink" must not abort a whole dispatching stage. Instead this mirrors
# gather-self-review-context.sh's check_boundary() + is_refused() pair: `[ -f
# "$path" ]` first (a missing target, including a dangling symlink, is a
# plain absence — skipped, not refused); then resolve_file() (leaf and every
# ancestor symlink) and a within_root() boundary check against
# CHANGE_ROOT_REAL, the same containment root that script uses for its own
# change-relative leaf (tasks.md). A leaf that resolves inside change-root —
# including a symlink to another file inside it — is read normally. A leaf
# that resolve_file() cannot even walk (REFUSED_REASONS' own "unresolvable"),
# or that resolves OUTSIDE change-root ("outside"), is refused per-source:
# omitted from the bundle, counted separately from "skipped", and the run
# still exits 0 — the spec's "a missing bundle never stops a run"
# requirement, and the same disposition gather-self-review-context.sh's
# header documents for exactly this shape of problem (a per-source
# trust-boundary violation is not a malformed invocation).
add_fixed_source() {
  local path="$1" label="$2" resolved
  if [ ! -f "$path" ]; then
    SKIPPED_LABELS+=("$label")
    return 0
  fi
  resolved="$(resolve_file "$path")" || {
    REFUSED_LABELS+=("$label")
    REFUSED_REASONS+=("unresolvable")
    return 0
  }
  if within_root "$resolved" "$CHANGE_ROOT_REAL"; then
    FOUND_LABELS+=("$label")
    FOUND_PATHS+=("$resolved")
  else
    REFUSED_LABELS+=("$label")
    REFUSED_REASONS+=("outside_change")
  fi
}

add_fixed_source "$PROPOSAL_FILE" "proposal.md"
add_fixed_source "$DESIGN_FILE" "design.md"
add_fixed_source "$TASKS_FILE" "tasks.md"

if [ -n "$PRINCIPLES_REAL" ] && [ -f "$PRINCIPLES_REAL" ]; then
  FOUND_LABELS+=("$PRINCIPLES_PATH")
  FOUND_PATHS+=("$PRINCIPLES_REAL")
else
  SKIPPED_LABELS+=("$PRINCIPLES_PATH")
fi

# --- project commands: the ## lint, ## test and ## run sections of
# <worktree>/.flow/project.md, so a dispatched subagent already carries this
# project's lint/test/run commands and never needs to open project.md itself
# (CLAUDE.md's lint-fix-priority rule sends every subagent there). Extracted,
# not the whole 20+ KB file, per design.md's
# scoped-dispatch-bundle-not-full-project-md decision. Appended after the
# principles file section, as the last "found" entry — see FOUND_LABELS
# ordering above.
#
# extract_project_section <file> <key> -> prints the body of "## <key>"
# (everything after that heading line up to, but not including, the next
# top-level "## " heading, or EOF), or nothing if <key> is not a heading in
# <file>.
extract_project_section() {
  local file="$1" key="$2"
  awk -v key="## $key" '
    $0 == key { grab = 1; next }
    /^## / { if (grab) exit }
    grab { print }
  ' "$file"
}

PROJECT_FILE="$WORKTREE_REAL/.flow/project.md"
PROJECT_COMMANDS_BODY=""
if [ -f "$PROJECT_FILE" ]; then
  PROJECT_FILE_RESOLVED="$(resolve_file "$PROJECT_FILE")" || PROJECT_FILE_RESOLVED=""
  if [ -n "$PROJECT_FILE_RESOLVED" ] && within_root "$PROJECT_FILE_RESOLVED" "$WORKTREE_REAL"; then
    for key in lint test run; do
      section="$(extract_project_section "$PROJECT_FILE_RESOLVED" "$key")"
      if [ -n "$(printf '%s' "$section" | tr -d '[:space:]')" ]; then
        PROJECT_COMMANDS_BODY="${PROJECT_COMMANDS_BODY}### ${key}

${section}

"
      fi
    done
    if [ -n "$PROJECT_COMMANDS_BODY" ]; then
      FOUND_LABELS+=("project commands")
      FOUND_PATHS+=("")
    else
      SKIPPED_LABELS+=("project commands")
    fi
  elif [ -n "$PROJECT_FILE_RESOLVED" ]; then
    REFUSED_LABELS+=("project commands")
    REFUSED_REASONS+=("outside_worktree")
  else
    REFUSED_LABELS+=("project commands")
    REFUSED_REASONS+=("unresolvable")
  fi
else
  SKIPPED_LABELS+=("project commands")
fi

echo "# Dispatch context bundle for $NAME"
echo "generated: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "head: $(git -C "$WORKTREE_REAL" rev-parse --short HEAD 2>/dev/null || echo unknown)"
echo
echo "found: ${#FOUND_LABELS[@]} source(s); skipped: ${#SKIPPED_LABELS[@]} source(s); refused: ${#REFUSED_LABELS[@]} source(s)"
if [ "${#REFUSED_LABELS[@]}" -gt 0 ]; then
  i=0
  while [ "$i" -lt "${#REFUSED_LABELS[@]}" ]; do
    case "${REFUSED_REASONS[$i]}" in
      outside_change) echo "refused: ${REFUSED_LABELS[$i]} (resolves outside the change directory)" ;;
      outside_worktree) echo "refused: ${REFUSED_LABELS[$i]} (resolves outside the worktree)" ;;
      *) echo "refused: ${REFUSED_LABELS[$i]} (could not be resolved: a symlink loop or other failure walking its path)" ;;
    esac
    i=$((i + 1))
  done
fi
if [ "${#SKIPPED_LABELS[@]}" -gt 0 ]; then
  i=0
  while [ "$i" -lt "${#SKIPPED_LABELS[@]}" ]; do
    echo "skipped: ${SKIPPED_LABELS[$i]} (absent)"
    i=$((i + 1))
  done
fi
echo

if [ "${#FOUND_LABELS[@]}" -gt 0 ]; then
  i=0
  while [ "$i" -lt "${#FOUND_LABELS[@]}" ]; do
    echo "## ${FOUND_LABELS[$i]}"
    echo
    if [ -n "${FOUND_PATHS[$i]}" ]; then
      cat "${FOUND_PATHS[$i]}"
    else
      printf '%s' "$PROJECT_COMMANDS_BODY"
    fi
    echo
    i=$((i + 1))
  done
fi

exit 0
