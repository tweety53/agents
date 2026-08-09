#!/usr/bin/env bash
# prepare-workspace.sh — validate a worktree's `## workspace isolation`
# declaration and export the workspace variables it derives.
#
# Usage: prepare-workspace.sh <worktree>
#
# This is exactly what `myflow-do` section 7 used to do by hand, in prose:
# run check-workspace-isolation.sh against the worktree first — fail loudly
# and stop if it fails — then derive and export the workspace variables the
# project's `## workspace isolation` section declares, resolved against the
# workspace id. Every rule this script applies is stated once, and canonical,
# elsewhere: the id derivation and what it derives under **The workspace id**
# and **What the id derives** (skills/myflow-contracts/workspace-isolation.md),
# and the `## workspace isolation` row shapes — the four `In a workspace`
# forms, the `<id>`, `<id_underscored>` and `<value:VARIABLE>` tokens — under
# **Project configuration** (skills/myflow-contracts/project-configuration.md).
# This script re-derives none of that reasoning; it applies the rule.
#
# Prints one `KEY=value` line per exported variable to stdout, one per
# declared row that carries a value (see the cache-index exception below), so
# a caller reads exactly what was exported without re-deriving anything
# itself.
#
# THE WORKSPACE ID NEEDS THE CHANGE NAME, AND THIS SCRIPT TAKES ONLY A
# WORKTREE. The id is derived from the change name and from nothing else, per
# **The workspace id**, so this script reads the name from the worktree's own
# branch: every apply worktree is created on `openspec/<name>`, per section 2
# of skills/myflow-do/SKILL.md, and that branch is the one place the name is
# already recorded where a script can read it without being handed it
# separately. A worktree not on such a branch — detached HEAD, or checked out
# on something else — cannot be resolved, and this script refuses rather than
# guessing at a name.
#
# THE CACHE INDEX IS THE ONE ROW THIS SCRIPT DOES NOT EXPORT. Per **The cache
# index** (skills/myflow-contracts/workspace-isolation.md), a `cache index`
# row is claimed by probing the project's own cache, not derived from the id —
# and the registry in skills/myflow-contracts/pipeline.md names `/myflow-do`,
# by probing, as what claims it "when it exports the workspace's variables".
# Probing means holding a client for whatever cache technology the project
# actually runs, which is exactly the kind of project-specific knowledge a
# script shipped into every project alike cannot carry. So a declared
# `cache index` row is reported by name on stderr — never silently dropped —
# and exports nothing; claiming it stays the operator's or the agent's own
# step, done against the real service, same as it was before this script
# existed.
#
# Exit 0 when the section is absent (no-op: nothing exported, nothing
# printed) or when isolation resolved and exported cleanly. When
# check-workspace-isolation.sh itself fails, this script's exit code is the
# guard's own — 1 (a dropped row) or 2 (the guard could not answer at all) —
# and the guard's stdout is relayed verbatim before exiting; its stderr is
# inherited, not captured, so it is already in front of the operator by the
# time this script's own exit code lands. Exit 2 additionally covers a
# worktree this script cannot derive a change name from, and one whose
# check-workspace-isolation.sh cannot be found.
#
# A project declaring no `## workspace isolation` section is the ordinary
# case for a repository with no runnable application — this repository is
# one — and is never reported as a misconfiguration, per **The empty id**
# (skills/myflow-contracts/workspace-isolation.md).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ "$#" -ne 1 ]; then
  echo "Usage: prepare-workspace.sh <worktree>" >&2
  exit 2
fi

WORKTREE="$1"

if [ ! -d "$WORKTREE" ]; then
  echo "prepare-workspace: $WORKTREE is not a directory" >&2
  exit 2
fi

GUARD="$SCRIPT_DIR/check-workspace-isolation.sh"
# `-x` as well as `-f`: a present-but-non-executable guard is not runnable
# either, and letting the exec below hit it directly would fail with a raw
# exit 126 ("Permission denied") — outside this script's own 0/1/2 exit
# contract and undocumented as such. Treated the same as "cannot find it at
# all", since neither case leaves this script able to validate anything.
if [ ! -f "$GUARD" ] || [ ! -x "$GUARD" ]; then
  echo "prepare-workspace: cannot find a runnable check-workspace-isolation.sh in $SCRIPT_DIR" >&2
  exit 2
fi

# Validate first, with the guard, not by eye — and before resolving a single
# row. Its stdout is captured so a violation report can be relayed after this
# script's own exit code is decided; its stderr is inherited, so a refusal
# it prints is already visible.
#
# CHECK_WORKSPACE_ISOLATION_PRINT_ROWS=1 asks the guard to append each
# validated resource row to its own report, as a `#ROW` line, once it has
# finished validating — see check-workspace-isolation.sh's own comment on
# that variable. This is the ONE parse of `.myflow/project.md`'s resource
# table this script needs: the guard already walked the file's cells to
# validate them, and the rows below are extracted from that same walk rather
# than a second one this script would otherwise have to run itself.
set +e
GUARD_OUT="$(CHECK_WORKSPACE_ISOLATION_PRINT_ROWS=1 "$GUARD" "$WORKTREE")"
GUARD_RC=$?
set -e

if [ "$GUARD_RC" -ne 0 ]; then
  [ -n "$GUARD_OUT" ] && printf '%s\n' "$GUARD_OUT"
  exit "$GUARD_RC"
fi

CFG="$WORKTREE/.myflow/project.md"
ISO_HEADING='^##[[:space:]]+workspace isolation[[:space:]]*$'

# No `.myflow/project.md`, or one declaring no `## workspace isolation`
# section, is a no-op: nothing exported, nothing printed. The guard above
# already confirmed this file — if present — is well formed, so the only
# question left here is whether the section exists at all, exactly mirroring
# the guard's own no-op case for a project with no such section.
if [ ! -f "$CFG" ]; then
  exit 0
fi
if ! grep -qiE "$ISO_HEADING" "$CFG"; then
  exit 0
fi

# The change name, read from the worktree's own branch. Every apply worktree
# is created on `openspec/<name>` per section 2 of skills/myflow-do/SKILL.md,
# and the id is derived from the change name and nothing else, per **The
# workspace id** (skills/myflow-contracts/workspace-isolation.md).
BRANCH="$(git -C "$WORKTREE" symbolic-ref --quiet --short HEAD 2>/dev/null || true)"
case "$BRANCH" in
  openspec/*)
    NAME="${BRANCH#openspec/}"
    ;;
  *)
    echo "prepare-workspace: $WORKTREE is not on an openspec/<name> branch (got '${BRANCH:-detached HEAD}') — cannot derive a workspace id without the change name" >&2
    exit 2
    ;;
esac

# The workspace id — prefix and digest joined by '-' — derived exactly as
# **The workspace id** (skills/myflow-contracts/workspace-isolation.md)
# states, under LC_ALL=C so the normalisation is locale-independent.
PREFIX="$(printf '%s' "$NAME" | LC_ALL=C tr 'A-Z' 'a-z' | LC_ALL=C tr -c 'a-z0-9' '-')"
while [ "${#PREFIX}" -gt 12 ] && [ "${PREFIX%-*}" != "$PREFIX" ]; do PREFIX="${PREFIX%-*}"; done
PREFIX="$(printf '%s' "${PREFIX:0:12}" | LC_ALL=C sed 's/-*$//')"
DIGEST="$(printf '%s' "$NAME" | shasum -a 256 | cut -c1-4)"
ID="$PREFIX-$DIGEST"
ID_UNDERSCORED="${ID//-/_}"
OFFSET=$(( (16#$DIGEST % 400 + 1) * 10 ))

# The resource table's rows, already parsed and validated — extracted from
# the guard's own report rather than re-reading `.myflow/project.md`. Every
# line of $GUARD_OUT that starts `#ROW\t` is one resource row the guard's awk
# program walked while validating it (see check-workspace-isolation.sh's
# check_resource_row and its CHECK_WORKSPACE_ISOLATION_PRINT_ROWS comment);
# stripping the `#ROW\t` marker leaves exactly the four tab-separated cells
# (`Resource`, `Variable`, `Default`, `In a workspace`) the loop below expects.
ROWS="$(printf '%s\n' "$GUARD_OUT" | awk -F'\t' '$1 == "#ROW" { print $2 "\t" $3 "\t" $4 "\t" $5 }')"

# Parallel indexed arrays — bash 3.2 is the floor here, as
# test-check-finish-preflight.sh's header records for this repository, so no
# associative array carries the rows by name; a `<value:VARIABLE>` reference
# resolves with a linear scan of VAR below instead.
RES=()
VAR=()
DEF=()
CELL=()
WSVAL=()

while IFS=$'\t' read -r r v d c; do
  [ -z "$v" ] && continue
  RES+=("$r")
  VAR+=("$v")
  DEF+=("$d")
  CELL+=("$c")
  WSVAL+=("")
done <<EOF
$ROWS
EOF

N=${#VAR[@]}

# substitute_id_tokens <cell> — <id> and <id_underscored>, the two tokens
# every `database`, `bucket` and `url` cell may carry, per **What the id
# derives** (skills/myflow-contracts/workspace-isolation.md). The two tokens
# never overlap as literal substrings, so the order they are replaced in does
# not matter.
substitute_id_tokens() {
  local cell="$1"
  cell="${cell//<id_underscored>/$ID_UNDERSCORED}"
  cell="${cell//<id>/$ID}"
  printf '%s' "$cell"
}

# resolve_value_refs <cell> — every `<value:VARIABLE>` reference in a `url`
# cell, resolved against the already-computed WSVAL of the row named
# VARIABLE. The guard above has already confirmed every reference names a
# `database`, `bucket` or `port` row in this same table, so this is
# resolution, not re-validation — but it is checked again here rather than
# trusted blindly, because a resolution bug that only ever checked the NAME
# would be silent rather than loud: `database`, `bucket` and `port` rows are
# the only ones Pass 1 (below) gives a WSVAL before this runs, so a reference
# resolving to any other kind — a `url` row (forbidden per **What a `url` row
# may reference, and what it may not**,
# skills/myflow-contracts/project-configuration.md — a `url` row may never
# reference another `url` row, so this table read cleanly by definition never
# produces one) or a `cache index` row (never given a WSVAL at all, per **The
# cache index** below) — would still find the VARIABLE by name and substitute
# whatever WSVAL currently holds for it: the empty string every row starts
# with. That is a wrong value printed and exported with no error at all, so
# the kind is checked here too, and a reference this script cannot resolve
# stops the run rather than silently exporting an empty substitution.
resolve_value_refs() {
  local cell="$1" rest name val i found
  while [[ "$cell" == *'<value:'*'>'* ]]; do
    rest="${cell#*<value:}"
    name="${rest%%>*}"
    val=""
    found=0
    for ((i = 0; i < N; i++)); do
      if [ "${VAR[$i]}" = "$name" ]; then
        case "${RES[$i]}" in
          database|bucket|port) found=1 ;;
          *)
            echo "prepare-workspace: <value:$name> resolves to \`${VAR[$i]}\`, a \`${RES[$i]}\` row — a reference may only name a \`database\`, \`bucket\` or \`port\` row, per **What a \`url\` row may reference, and what it may not** (skills/myflow-contracts/project-configuration.md) — refusing rather than substituting an empty or stale value" >&2
            exit 2
            ;;
        esac
        val="${WSVAL[$i]}"
        break
      fi
    done
    if [ "$found" -eq 0 ]; then
      # The guard already refused a reference naming no row, so this branch
      # is unreachable on a validated file — guarded anyway rather than
      # looping forever on a token that never resolves.
      cell="${cell/<value:$name>/}"
    else
      cell="${cell/<value:$name>/$val}"
    fi
  done
  printf '%s' "$cell"
}

# Pass 1 — every row a `url` row may reference: `database`, `bucket`, `port`.
# `cache index` is resolved in this same pass in the sense that it is
# recognised and reported; it never receives a WSVAL, and no `url` row may
# reference it (the guard already refuses one that tries).
for ((i = 0; i < N; i++)); do
  case "${RES[$i]}" in
    database|bucket)
      WSVAL[$i]="$(substitute_id_tokens "${CELL[$i]}")"
      ;;
    port)
      # `10#` forces base-10 interpretation of `Default`. Without it, bash
      # arithmetic reads a leading-zero literal as octal: a declared default
      # of `0070` would silently become the wrong port (`0070 + 20` is `76`,
      # not `90`), and a default like `0080` is not even valid octal — it
      # crashes the whole script with "value too great for base". The guard's
      # own `Default` rule for a `port` row is only "a bare integer"
      # (`^[0-9]+$`), which admits a leading zero, so this script — not the
      # table's author — is the one place that has to force the base.
      WSVAL[$i]=$(( 10#${DEF[i]} + OFFSET ))
      ;;
    "cache index")
      echo "prepare-workspace: \`${VAR[$i]}\` (cache index) is claimed by probing the project's own cache, not derived from the workspace id — per the registry in skills/myflow-contracts/pipeline.md, \`/myflow-do\` claims it, by probing, when it exports the workspace's variables. This script does not carry a client for the project's cache, so this row is reported rather than exported; claim it against the real service before anything reads \`${VAR[$i]}\`." >&2
      ;;
  esac
done

# Pass 2 — `url` rows, which may reference the values pass 1 just resolved.
for ((i = 0; i < N; i++)); do
  if [ "${RES[$i]}" = "url" ]; then
    WSVAL[$i]="$(resolve_value_refs "$(substitute_id_tokens "${CELL[$i]}")")"
  fi
done

# Export and print, in declaration order — the same order the project wrote
# the table in, so a KEY=value line's position in the section is the same
# position it prints in.
for ((i = 0; i < N; i++)); do
  case "${RES[$i]}" in
    "cache index") continue ;;
  esac
  export "${VAR[$i]}=${WSVAL[$i]}"
  printf '%s=%s\n' "${VAR[$i]}" "${WSVAL[$i]}"
done
