#!/usr/bin/env bash
# check-spec-reach.sh — fail when a Playwright spec in a project's
# `regression checkout` is reached by no package.json script.
#
# Usage: check-spec-reach.sh <project-root>
#
# WHICH CHECKOUT IS READ: the declared `regression checkout` — or, when the
# project root is on a branch and a worktree of that checkout's repository is
# on the same branch, that worktree, exactly as resolve-visual-screenshots.sh
# resolves it.
#
# Why this exists: seven route A–F capture suites in gymie-playwright never
# ran once — `test:e2e` reached `tests/`, `test:e2e:full` reached
# `tests-full/`, and the root `kan-*.spec.ts` files flow.visual-verify's
# `capture` writes matched neither. One passed green against a 404 page for
# its whole life (KAN-29 self-review, KAN-390).
#
# HOW REACHABILITY IS DECIDED: by Playwright itself, never by reading the
# scripts' arguments as globs. Every package.json script whose command
# contains `playwright test` is run as `npm run -s <script> -- --list
# --reporter=json` from the checkout, and the union of `config.rootDir` +
# `suites[].file` is what those scripts reach. A static reading is wrong on
# the first real case: `playwright test kan-*.spec.ts` works only because npm
# runs the script through `sh`, which expands the glob — Playwright treats a
# positional argument as a regular expression against the absolute path, and
# the same argument quoted lists 0 files.
#
# WHAT IS ENUMERATED: every `*.spec.ts` under the checkout, `node_modules`,
# `.git` and `.worktrees` pruned — a `/flow` apply worktree not yet migrated
# to the sibling `<repo>-worktrees/` layout, under `.worktrees/<change>/`,
# carries its own full copy of every spec (KAN-30's
# own gymie-playwright worktree, the first one ever created under a
# checkout this guard reads, is what surfaced this). Not Playwright's wider
# default testMatch (`.test.ts`,
# `.spec.js`, `.spec.mjs`) — a vitest `*.test.ts` file beside a Playwright
# suite would otherwise become an orphan of a runner it never belonged to.
#
# STATED LIMIT: a script reaching Playwright indirectly (`npm run other`) is
# not followed; the files only it reaches count as orphans. Name
# `playwright test` in the script that reaches them.
#
# THREE EXIT CODES:
#   0  every enumerated spec is reached — or the project is NOT CONFIGURED
#      (no .flow/project.md, no `## visual verification` section, or no
#      `regression checkout` row), which prints `Spec reach: not configured`.
#      Both call sites (flow.verify, flow.visual-verify) run this guard as one
#      more command in a list where any non-zero exit blocks, so "nothing to
#      check" must be 0 — unlike check-visual-trigger.sh's exit 2, whose
#      caller distinguishes its exits by hand.
#   1  at least one orphan — one `<path>: reached by no package.json script`
#      line each.
#   2  cannot answer: not exactly one argument, the root is not a directory,
#      the section is declared twice (ambiguous — neither is read), the
#      checkout is not a directory, it holds no package.json, no script
#      contains `playwright test`, `node` or `npm` is absent, or a `--list`
#      run exits non-zero or prints unparsable JSON (its output is relayed).
#
# THE INPUT IS ATTACKER-INFLUENCED, exactly as check-visual-verification.sh's
# header states: .flow/project.md and the checkout's package.json are tracked
# and editable in any pull request. This guard runs `npm run` on scripts the
# project already runs through `verify` and `capture` — the same trust level,
# no widening — and every interpolated value reaching an error message passes
# through sanitizeDisplay first.
#
# THE HEADING SCAN AND TABLE PARSER ARE check-visual-verification.sh's OWN
# CONVENTIONS, reused rather than reinvented — vvHeadingCount,
# vvSectionLines and vtSplitCells/vtTrimCell/vtFoldCell, with every
# interpolated value reaching an error message passed through
# sanitizeDisplay. The mechanism is stats/internal/guard/specreach.go; its
# test is stats/internal/guard/check_spec_reach_test.go.
#
# Derives no repository root from its own location (check-guard-symlinks.sh
# rule 4) — the project root is always the argument.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-spec-reach: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
flow_guard_exec check-spec-reach 2 "check-spec-reach:" "$@"
