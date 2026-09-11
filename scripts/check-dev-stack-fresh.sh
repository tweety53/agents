#!/usr/bin/env bash
# check-dev-stack-fresh.sh — verify, before a handoff prints a running dev
# stack's URL, that the stack is serving the worktree's own build rather
# than one committed before HEAD.
#
# Usage: check-dev-stack-fresh.sh <worktree>
#
# ONE ARGUMENT, THE WORKTREE — the directory holding `.flow/project.md` and
# the build the stack ought to be serving, matching check-visual-verification.sh's
# own project-root convention.
#
# KAN-334's failure, measured on kan-333: a handoff printed `localhost:3000`
# while the dev server behind it was serving a bundle eleven commits stale —
# the watcher had stopped recompiling, the log had stopped growing, and
# nothing in the pipeline looked. The start rule in the run-instructions
# stage rebuilds and restarts what it can, but a stack it only skips or
# merely starts (a start command that does not rebuild, a watcher that has
# died) is trusted rather than checked. This guard is the check.
#
# THE CHECK IS THE PROJECT'S OWN DECLARED FINGERPRINT. A project tells the
# pipeline once, in `.flow/project.md`'s `## visual verification` section,
# how to prove a served bundle is the worktree's build: the `fingerprint`
# command row, which rebuilds first and then compares the served document
# against the rebuild (this repository's own row is the worked example).
# Reusing that row is the whole design: no second per-project mechanism is
# invented, no new project-configuration vocabulary is added, and a project
# that already declared the row for visual verification needs to write
# nothing. The row's own exit code is the verdict.
#
# THREE EXIT CODES, project-get.sh's convention:
#   0  fresh — a `fingerprint` row is declared and its command exited 0.
#   1  stale (or not serving) — the command is declared and exited non-zero.
#      The restart action is the same for both, so the caller words one line.
#   2  cannot answer — usage; `<worktree>` is not a directory; no
#      `.flow/project.md`; no `## visual verification` section; the section
#      carries no readable `fingerprint` row. The caller prints its
#      freshness-unverified gap line from this exit, so the reason goes to
#      stderr and the exit is never a silent pass.
#
# Exit 1 carries the fingerprint command's own output, streamed verbatim —
# it is the evidence, and a diff or a curl failure names itself.
#
# THE DECLARED COMMAND IS EXECUTED — that is this guard's purpose, and the
# one thing that separates it from check-visual-verification.sh, which
# validates the same section and deliberately runs nothing. The command
# comes from the project's own tracked `.flow/project.md`, the same trust
# every `## run` command the pipeline executes already carries.
#
# THE TABLE PARSER IS THE SOURCED TRIO, not a fresh regex: split_cells,
# trimcell and foldcell come from scripts/lib/visual-table-cells.awk via a
# second `-f` on the same awk invocation, for the same reason
# check-visual-verification.sh, check-visual-trigger.sh and
# resolve-visual-screenshots.sh source it — one parser for the
# `| Command | Runs |` shape, or three drifting ones. A leading UTF-8 BOM is
# stripped by lib/strip-bom.sh's strip_bom_cat before the awk reads anything,
# for the reason that lib's own header records; code fences are skipped so a
# worked example inside the section can never declare a row.
#
# This guard ships through the skills/*/scripts/ symlink farm, so sourcing a
# sibling lib/ is safe by the criterion scripts/lib/resolve-file.sh's header
# states; reached by hand-copy alone, it is not.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ "$#" -ne 1 ]; then
  echo "check-dev-stack-fresh: usage: check-dev-stack-fresh.sh <worktree>" >&2
  exit 2
fi
WORKTREE="$1"

if [ ! -d "$WORKTREE" ]; then
  echo "check-dev-stack-fresh: $WORKTREE is not a directory" >&2
  exit 2
fi

CFG="$WORKTREE/.flow/project.md"
if [ ! -e "$CFG" ] && [ ! -L "$CFG" ]; then
  echo "check-dev-stack-fresh: $CFG does not exist — no fingerprint is declared, so freshness cannot be checked" >&2
  exit 2
fi
if [ ! -f "$CFG" ]; then
  echo "check-dev-stack-fresh: $CFG is not a regular file" >&2
  exit 2
fi

# strip_bom_cat — see lib/strip-bom.sh's header: a leading BOM would make the
# present section read as absent, and the same false-absence hazard this lib
# exists for applies to every guard that greps this file.
source "$SCRIPT_DIR/lib/strip-bom.sh"

# The row-extraction program is written to a private temp file so the shared
# trio and this caller's own rules reach awk as two `-f` sources on ONE
# invocation — an inline program cannot be mixed with `-f` portably, and the
# heredoc-as-stdin form would steal the data stream the pipe is already
# feeding awk. Removed by one EXIT trap, on failure and signal paths too.
AWK_PROG="$(mktemp "${TMPDIR:-/tmp}/check-dev-stack-fresh.XXXXXX")"
cleanup() { rm -f "$AWK_PROG"; }
trap cleanup EXIT

cat > "$AWK_PROG" <<'AWK'
BEGIN { insec = 0; infence = 0; incmd = 0; found = 0 }
/^```/ { infence = !infence; next }
infence { next }
!insec && /^##[[:space:]]+visual verification[[:space:]]*$/ { insec = 1; next }
!insec { next }
/^##[[:space:]]+/ { exit }
/^[[:space:]]*$/ { incmd = 0; next }
!/^\|/ { incmd = 0; next }
{
  n = split_cells($0, cells)
  if (n < 2) next
  key = foldcell(cells[1])
  if (key == "command" && foldcell(cells[2]) == "runs") { incmd = 1; next }
  if (!incmd) next
  if (key == "fingerprint") {
    print trimcell(cells[2])
    found = 1
    exit
  }
}
END { if (!found) exit 3 }
AWK

# THREE ANSWERS, READ AS THREE (the check-visual-verification.sh pattern):
# a set -e read would end this script before RC was ever read, collapsing
# "row found" and "could not look" into one early exit.
set +e
FPR="$(strip_bom_cat "$CFG" | awk -f "$SCRIPT_DIR/lib/visual-table-cells.awk" -f "$AWK_PROG")"
RC=$?
set -e
case "$RC" in
  0) ;;
  3) echo "check-dev-stack-fresh: $CFG declares no readable \`fingerprint\` row under \`## visual verification\` — freshness cannot be checked" >&2; exit 2 ;;
  *) echo "check-dev-stack-fresh: could not read the fingerprint row from $CFG (awk exited $RC)" >&2; exit 2 ;;
esac
if [ -z "$FPR" ]; then
  echo "check-dev-stack-fresh: the fingerprint row in $CFG resolves to an empty command — freshness cannot be checked" >&2
  exit 2
fi

# Run the declared command from the worktree, streaming its output — the
# evidence on a stale or dead stack is the command's own diff or curl noise,
# not a verdict line.
echo "check-dev-stack-fresh: fingerprint: $FPR"
if (cd "$WORKTREE" && bash -c "$FPR"); then
  echo "stack-fresh: the fingerprint exited 0 — the stack is serving the worktree's build"
  exit 0
else
  echo "stack-stale: the fingerprint exited non-zero — the stack is not serving the worktree's build (stale bundle, or nothing serving)" >&2
  exit 1
fi
