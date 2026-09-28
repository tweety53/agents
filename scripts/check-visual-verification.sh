#!/usr/bin/env bash
# check-visual-verification.sh — validate a project's `## visual
# verification` declaration against the contract that defines it.
#
# Usage: check-visual-verification.sh <project root>
#
# ONE ARGUMENT, THE PROJECT ROOT — never the configuration file itself,
# matching check-workspace-isolation.sh's own convention: every caller
# already has the root, not the path underneath it.
#
# A project with no `.flow/project.md`, or one that carries it but declares
# no `## visual verification` section, is CLEAN, not a failure — the section
# is optional per design.md's "The contract", and a guard that failed on its
# absence would make every project in the world declare one.
#
# Prints one violation line per finding as `file:line: <message>`, then ONE
# verdict line:
#   VISUAL-OK:      <path> — <what was checked, or why nothing was>
#   VISUAL-INVALID: <path> — <n> violation(s) in the `## visual verification`
#                    section
#
# Exit 0 when clean, 1 when any violation was found, 2 when the guard cannot
# answer at all: the project root is not a directory, `.flow/project.md`
# exists but cannot be read or is not a regular file, the heading scan itself
# failed to look, or `git remote get-url origin` against a declared
# `regression checkout` failed for a reason other than "no such remote" (that
# one IS a finding — see below). A `2` is never a silent skip: a guard
# pointed at a moved file that reports `VISUAL-OK` is a vacuous pass.
#
# THE TWO ON-DISK TABLES, AND WHERE THEIR SHAPE COMES FROM. design.md's "The
# contract" section documents the settings and commands as
# `Setting | Required | Meaning` and `Command | Required | Meaning` — that is
# documentation prose, describing what each row MEANS, not the literal
# heading text of the table a project writes. The literal on-disk shape is
# established from two other places instead: task 8 of tasks.md's worked
# example for Gymie writes the settings table as `| Setting | Value |`, and
# `## workspace isolation` (skills/flow-contracts/project-configuration.md,
# `check-workspace-isolation.sh`) already establishes `| Command | Runs |` as
# this repository's one convention for a table of commands a project
# declares — reused here rather than invented a second time, so a project
# author who has already written one `## workspace isolation` section is not
# asked to learn a second table shape for `## capture`. Task 2, not yet
# written when this guard was, is canonical for the section's prose once it
# lands; this guard's own tests pin the two headers it actually
# parses, so a disagreement between that landing and this file fails loudly
# rather than silently.
#
# WHAT IS CANONICAL, AND WHAT THIS FILE IS, for both checks below — matching
# check-workspace-isolation.sh's own citation of the identical shape for its
# `Resource` word. **Project configuration**
# (`skills/flow-contracts/project-configuration.md`), under its "## visual
# verification" heading, states the `Setting` and `Command` vocabularies as
# closed, and states that `regression repo` is an identity assertion, not an
# authorisation, checked whenever `regression checkout` and `regression
# repo` are both declared — design.md's `no-automatic-push` decision, which
# superseded an earlier design where the same check ran only behind a
# `push to default branch: allowed` row this contract no longer carries: a
# panel slot demonstrated that both fields of that row lived in this same
# pull-request-editable file and so authorised nothing. This file restates
# neither decision's reasoning; it cites the paragraph and enforces the rule.
#
# `git remote get-url origin` FAILING FOR "NO SUCH REMOTE" IS A FINDING, NOT
# A `2`. A checkout that is a real, valid git repository but carries no
# `origin` remote at all cannot honour the identity assertion either — that
# is a fact about the declaration, not a failure to look, so it is reported
# and the run continues. Any OTHER failure of that command (measured against
# git 2.50.1: none reproduces once `git rev-parse --git-dir` has already
# accepted the checkout as valid, which is why this guard's own tests
# reach this branch through a stubbed `git` rather than a real one) is read
# as "cannot determine whether the identity assertion holds" and escalates
# the whole run to exit 2, matching this guard's own "never fail open" rule
# for every other unreadable input below.
#
# THE INPUT IS ATTACKER-INFLUENCED, the same fact check-workspace-isolation.sh
# is written against: `.flow/project.md` is tracked and editable in any pull
# request. NOTHING read here is executed — this guard never runs `setup`,
# `verify`, `capture`, `fingerprint`, `start` or `specs`, and never interpolates a
# cell into a shell. All of
# the table parsing happens in flow-guard
# (stats/internal/guard/visualverification.go), whose only input is the
# file's text, and every violation line, plus every cell this guard
# interpolates into a message afterward, passes through sanitizeDisplay —
# the Go twin of lib/sanitize-display.sh — before it reaches the terminal;
# see stats/internal/guard/visualsection.go for why escaping and not
# stripping or refusing.
#
# THE PARSER'S split_cells/trimcell/foldcell TRIO IS SHARED — the Go twins
# of scripts/lib/visual-table-cells.awk in
# stats/internal/guard/visualsection.go, not check-workspace-isolation's own
# (wider, four-column; stats/internal/guard/workspaceisolation.go) copy of
# the same shape. Its hazards are the same because its input is: a `|`
# inside a cell, written `\|`, must not split the row; a row may omit its
# trailing `|`; a file may arrive with CRLF line endings.
#
# A LEADING UTF-8 BOM IS STRIPPED BEFORE ANY OF THE ABOVE READS THE FILE, via
# stripBOM, the Go twin of scripts/lib/strip-bom.sh — see that file's
# header for why a BOM would otherwise make a present section read as
# absent: the fix arrived after check-visual-trigger.sh and
# resolve-visual-screenshots.sh already existed, sharing this same
# heading-parse convention, and stayed a single guard's own logic for one
# commit before moving into a shared helper.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
. "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-visual-verification: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-visual-verification 2 "check-visual-verification:" "$@"
