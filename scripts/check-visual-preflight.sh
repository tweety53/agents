#!/usr/bin/env bash
# check-visual-preflight.sh — the four workspace checks `flow.visual-verify`
# step 3 runs before any verifier is dispatched (skills/flow/visual-verify.md).
#
# Usage: prepare-workspace.sh <worktree> \
#          | check-visual-preflight.sh <worktree> <app-root>=<resolved-url> [...]
#
# stdin is prepare-workspace.sh's `KEY=value` lines; one argument per app
# `ui paths` matched — its package root, relative to <worktree>, and its
# worktree-resolved URL. An app root inside a different git repository than
# <worktree> is checked as its own worktree: checks 2–4 run against that
# repository's worktree root, its own `.flow/project.md` rows and `start`
# command, and with that repository's apps only; check 1 is unchanged. The
# checks:
#   1. ports    — every `port` row's exported value and every app URL's port:
#                 a listener `lsof -nP -iTCP:<port> -sTCP:LISTEN` reports fails
#                 only when its probe goes unanswered (an HTTP GET of the app
#                 URL the port came through, any response within 5 s; else a
#                 TCP connect to 127.0.0.1 or ::1).
#   2. base-url — for each `## workspace isolation` row whose exported value
#                 differs from its Default, a non-Markdown text file under an
#                 app root naming the Default (a `port` row: `:<default>` not
#                 followed by a digit) on a line not naming the row's Variable
#                 fails, unless the `## visual verification` `start` command
#                 names the Variable or the resolved value, or the Default
#                 appears only inside `${VAR:fallback}` references whose VAR
#                 the `start` command names or stdin exports.
#   3. origins  — a configuration-shaped file (`.env*`, `*.json`, `*.y*ml`,
#                 `*.toml`, `*.ini`, `*.properties`, `*.conf`, `*.config.*`;
#                 never a production `*.prod.*` file)
#                 whose `allowed[_-]?origins` line or bracketed list names a
#                 URL must name every app's origin, unless the line holds a
#                 `${VAR:fallback}` reference whose VAR the `start` command
#                 names or stdin exports.
#   4. playwright — `@playwright/test` resolved by `node` from each app root
#                 must sit inside <worktree>; unresolvable or no `node` is an
#                 INFO line and passes (setup installs it).
# Skipped by checks 2–3: .git node_modules dist build coverage test-results
# playwright-report testdata vendor .superpowers, a nested checkout, binaries,
# files over 1 MiB. design-verify.md § VV-18 records why.
#
# Prints any INFO: and FAIL: lines, then one verdict line:
#   PREFLIGHT-OK:     <worktree> — ports, base URL, origins and Playwright all pass
#   PREFLIGHT-FAILED: <worktree> — <n> failing check line(s)
# Exit 0 every check passed; 1 at least one FAIL: line; 2 cannot answer
# (usage, a missing directory, lsof absent or failing, a malformed stdin
# line, a declared row with no exported value, an invalid or duplicated
# section) — the cause on stderr.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-visual-preflight: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-visual-preflight 2 "check-visual-preflight:" "$@"
