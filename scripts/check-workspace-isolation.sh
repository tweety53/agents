#!/usr/bin/env bash
# check-workspace-isolation.sh — validate a project's `## workspace isolation`
# declaration against the contract that defines it.
#
# Usage: check-workspace-isolation.sh [<project root> ...]
#
# With no arguments it checks the repository this script ships in. Each argument
# is a PROJECT ROOT — the directory holding `.flow/project.md` — never the
# configuration file itself, because that is the path every caller already has.
#
# WHO RUNS IT, AND WHY THAT IS TWO CALLERS RATHER THAN ONE. `/flow`'s implement
# phase runs it against each apply worktree before it resolves or exports a
# single row — skills/flow/implement.md — and that is the call that reaches
# every project flow is installed into, because the read happens in the
# project rather than here. This repository ALSO lists it in its own `## lint`,
# which is a self-check on this repository's own `## workspace isolation`
# section and nothing more: a green lint run says nothing about any other
# project. Reading the lint entry as the enforcement was exactly the shape of
# false pass every branch below is written against, one level up.
#
# Prints one violation line per finding, then ONE verdict line per project:
#   ISOLATION-OK:      <path> — <what was checked>
#   ISOLATION-INVALID: <path> — <n> violation(s)
#
# Exit 0 when every project checked is well formed, 1 when any violation was
# found, 2 when it cannot answer at all — a project root that is not a
# directory, a `.flow/project.md` that is not a regular file (a directory, a
# fifo, a dangling symlink) or cannot be read, or a scan of it that failed
# rather than found nothing. The violations
# are on stdout and the refusals on stderr, so a caller reading stdout is
# reading findings and never an excuse for their absence.
#
# WHY THE VERDICT WORDS DIFFER FROM check-cleanup-complete.sh's. That guard says
# COMPLETE/LEFTOVER about the same section of the same file, and the two
# vocabularies are deliberately not shared. It asks whether the removal
# HAPPENED; this one asks whether the declaration is WELL FORMED. A reader who
# carried COMPLETE across to OK would be carrying an answer about resources into
# an answer about text.
#
# NEVER FAIL OPEN — the rule every branch below is written to. A project that
# declares no section passes SILENTLY, and that is the overwhelmingly common
# case, so every OTHER way of ending up with nothing to check has to be told
# apart from it: an unreadable configuration refuses, a duplicated heading is a
# violation, a table whose header this guard does not recognise is a violation,
# and a row whose column count does not match its header is a violation. Reading
# "declares nothing" out of any of those would report every malformed project
# valid, which is the failure this guard exists to prevent.
#
# A COMMAND'S FAILURE IS NEVER READ AS ITS NEGATIVE ANSWER, which is the same
# invariant one level down. `-r` is true of a directory and `grep -q` then fails
# with the same exit status it uses for "no match", so a `.flow/project.md`
# that is a directory once answered "declares no section" and exited 0. Every
# test below therefore says which question it is answering: the file type is
# tested before its readability, and each `grep` exit status is read as three
# outcomes — matched, did not match, could not look — rather than as two.
#
# WHAT IS CANONICAL, AND WHAT THIS FILE IS. Every rule enforced here is stated
# by **Project configuration — workspace isolation** (`skills/flow-contracts/project-configuration-isolation.md`), which wins on
# any disagreement — under its "How a `## workspace isolation` section is
# written" paragraph and the two bullets that follow "An isolation row resolves
# under the same rules this file applies to everything else it consumes". This
# file restates none of the reasoning behind a rule; it cites the paragraph and
# enforces the rule. Which rules are mechanical and which remain the agent's to
# apply is recorded in that same file, under "Which of these rules a script
# checks, and which are left to the agent", so the contract does not imply that
# all of them are checked.
#
# THE PARAGRAPH NAMES ABOVE ARE PROSE, NOT CITATIONS, AND THAT IS DELIBERATE.
# check-references.sh resolves a bold token against the REAL HEADINGS of the
# file it names, and project-configuration.md is structured almost entirely as
# bold lead-in paragraphs rather than headings. A bold token naming one of those
# would be a stale reference the moment anything checked it. The paragraph names
# here are unbolded for that reason, and the only bold tokens are ones that
# resolve to a heading of that file. No COUNT of its headings is written here:
# one was — "exactly one, its title" — and it went stale the first time a
# heading was added to that file.
#
# THE INPUT IS ATTACKER-INFLUENCED. `.flow/project.md` is tracked in the
# repository and editable in any pull request. NOTHING read here is executed:
# this guard never runs a project's `create`, `remove` or `survivors` command,
# never resolves a path out of the file, and never interpolates a cell into a
# shell. The whole of the parsing happens inside one function (wiValidate in
# stats/internal/guard/workspaceisolation.go) whose only input is the file's
# text — which is also why its cell splitter is a character walk rather than a
# field split.
#
# THE PARSER IS THE CLEANUP GUARD'S, WIDENED FROM TWO COLUMNS TO FOUR. Its
# hazards are the same because its input is: a `|` inside a cell, written `\|`,
# must not split the row; a row may omit its trailing `|`; a file may arrive
# with CRLF line endings; and a row may be indented. THE PARSER ITSELF stays a
# copy of its own (wiParse.splitCells, not the cleanup guard's ccWorkspaceRow);
# the heading rule, the token shape and the display escaping are the cleanup
# guard's own Go values (ccHeading, ccUnknownToken, ccSanitize in
# stats/internal/guard/cleanupcomplete.go), in the one flow-guard binary both
# guards run in.
#
# `lib/flow-guard.sh` IS SOURCED (in the bash it was `resolve_file`, now
# resolveFile in stats/internal/guard/resolvefile.go), and that is deliberate
# rather than accidental. Unlike the containment
# parser above — which a project can and does copy into its OWN tooling,
# standalone, with nothing else from this repository — this guard's own
# distribution is the KAN-73 symlink farm: it is never hand-copied anywhere,
# only ever reached at `<repo>/scripts/check-workspace-isolation.sh` or
# through a `skills/<name>/scripts/check-workspace-isolation.sh` symlink that
# KAN-73's install carries, and both directories carry a `lib` symlink
# beside it as part of that same farm (KAN-73's design.md, "The guard-to-skill
# map"). "A sourced helper would make a guard present but unrunnable" is
# exactly the concern that fails to apply here: the helper travels WITH the
# guard, unconditionally, wherever the guard is reachable at all — never
# absent, so never a way to be present-but-unrunnable. This is the identical
# argument KAN-153's review (F7) accepted for check-unfinished-work.sh
# sourcing a library of its own (since folded into the store rewrite): a
# sourced helper travels WITH its guard, unconditionally, whenever
# `setup.sh` installs skills rather than `scripts/` on their own — see that
# review's Disposition section for the full reasoning, carried forward here
# rather than re-litigated. Left
# genuinely single-file, out of this change's scope: the since-retired
# the retired self-review gather (kan-526), which was copied into OTHER projects' own
# tooling standalone and could not assume any sibling travels with it.
#
# A FENCED EXAMPLE INSIDE THE SECTION IS PARSED AS A REAL TABLE, and that is a
# known limitation rather than an oversight. This guard carries no fence
# tracker, exactly as the cleanup guard's parser carries none, so a project that
# writes a specimen table inside the section will have it validated. The failure
# direction is the safe one — a specimen is reported rather than silently
# skipped — and the remedy is to put the example outside the section, where
# nothing reads it.
#
# Ported to Go (KAN-842): the body's reasoning for every step is in
# stats/internal/guard/workspaceisolation.go, beside the code. Resolving the
# no-argument root needs the path this script was invoked by, before any
# symlink is followed, so this shim exports it as FLOW_GUARD_SELF; the guard
# resolves it only when no project root was given.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
. "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-workspace-isolation: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_SELF="${BASH_SOURCE[0]}"
export FLOW_GUARD_SELF
flow_guard_exec check-workspace-isolation 2 "check-workspace-isolation:" "$@"
