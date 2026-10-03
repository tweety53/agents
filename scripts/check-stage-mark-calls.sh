#!/usr/bin/env bash
# check-stage-mark-calls.sh — every `flow stage begin` in a skill must carry
# a session token and a harness, the session token must be a literal, and the harness must
# NOT be — the two required flags are wrong in opposite directions, and this
# script rejects both. Every `flow record dispatch` must carry a literal
# session token too, for the same reason and by the same two checks.
#
# Usage: scripts/check-stage-mark-calls.sh [path ...]          # dirs or files
#
# With no arguments it scans `skills/`, resolved relative to the repo root —
# the one tree the marking skills live under. A caller may pass explicit
# paths instead (its Go tests do, against t.TempDir() fixtures).
#
# WHAT THIS CATCHES. KAN-172 task 1 gave `flow stage begin` two required
# flags: `-session-token`, a literal token the daemon later finds in the calling
# session's own transcript to bind that stage run's session_id, and
# `-harness`, the harness recording the mark. Neither is enforced by prose:
# four skills carried a `flow stage begin` invocation with neither flag for
# the whole of KAN-16's life before this guard existed (tasks.md, "This guard
# is the deliverable of the task, not the edits"). This script makes that
# defect loud rather than silent, forever.
#
# It also enforces design.md's second decision, "the session token is a literal,
# never a shell substitution": `-session-token "mf-$(date +%s)-$$"` is recorded by
# the transcript *before* the shell expands it, so every session would record
# the identical string and it would identify nothing. A `-session-token` value
# containing `$(`, a backtick, or `$` immediately followed by a shell
# variable-name character is rejected here, in the skill source, before a
# reader ever has the chance to run it — the same three shapes
# `stats/cmd/flow/stage.go`'s `validateSessionToken` rejects at the CLI, and
# `internal/api/stages.go`'s `validateSessionTokenShape` rejects again at the store,
# as defence in depth.
#
# `flow record dispatch` IS SCANNED FOR THE SESSION-TOKEN RULES TOO
# (KAN-258). The record verb takes `-session-token` for precisely the reason
# `stage begin` does: the daemon finds that literal token in the calling
# session's own transcript and binds the dispatch row to the session that
# wrote it. A substituted value there lands in every transcript as the same
# unexpanded string and so discriminates between no two sessions — the whole
# mechanism, silently gone. A guard that enforced the literal on one verb and
# not the other would be a guard with a hole, and the record verb's call
# sites are numerous (four in `skills/flow/review-panel.md` alone). So both
# session-token rules — the flag is present, and its value is a literal —
# are applied to `flow stage begin` and `flow record dispatch` alike.
#
# The other rules stay with the verb that owns them. `-harness` is a `stage
# begin` flag and `record dispatch` has none (a dispatch row inherits its
# harness from the stage run it belongs to), so requiring one on a record
# call would demand a flag the CLI would reject. The guessed-change-name
# rule likewise stays on `stage begin`, whose begin handler is the one that
# bootstraps a change row from a name it has never seen.
#
# Only `flow record dispatch` is scanned: `record finding`, `record
# status`, `record render` and `record journal-count` take no session token
# at all, so there is nothing on them for these rules to say.
#
# `-harness` FAILS THE OPPOSITE WAY: a value that IS a literal is the defect.
# These skill files are one source installed into `~/.claude/skills/` and
# `~/.zcode/skills/` alike (CLAUDE.md's "installed alongside the skills in
# every harness"), so a hardcoded `-harness claude-code` in the shared source
# mislabels every ZCode run as Claude Code — masking the very thing the
# harness field exists to record: which harness ran the stage. The skill source must instead carry a
# placeholder the agent fills in at call time, exactly as `-session-token`'s literal
# token is filled in at call time — this guard requires the `-harness` value
# to be written as a bracketed placeholder (`<harness>`) and rejects any bare
# value, the same way it rejects a substituted session token.
#
# `flow stage end` is deliberately NOT checked: pipeline.md's "Stage marks"
# section states plainly that `stage end` carries no harness of its own (the
# harness is recorded once at `begin` and is immutable), and it takes no
# `-session-token` either — attribution happens once, at `begin`. Only lines that
# invoke `stage begin` are examined.
#
# `/flow-status` marks nothing, per README.md's Level 1 section
# so a `flow stage begin` appearing in its SKILL.md would itself be
# a defect this guard would (correctly) catch — there is no exemption for it.
#
# It also enforces that a `stage begin`'s change argument is never a guess.
# The guard cannot know whether `<name>` is resolved at a given call site,
# but it can know that a bracketed placeholder whose own text says "guess"
# is not — `<name-or-best-guess>` and any restatement of it. KAN-182: on
# 2026-08-15 `/flow-fast`'s state gate marked its stage with
# `<name-or-best-guess>` before its change name was resolved; the daemon's
# begin handler bootstraps a change row for any name it has never seen
# (`ApplyBeginStageMark`, `stats/internal/api/stages.go`), so the guess got
# a row of its own — `kan-175` and `kan-175-more-ui-ux-fixes` appeared as
# two open changes, 29 seconds apart. A mark writes, so a guessed name
# bootstraps a change row that outlives the run.
#
# The check does not try to identify "the change argument" at all. Finding
# the last token on a shell command line requires parsing shell syntax, and
# this guard kept getting that parsing wrong: KAN-182 round 1 added
# quote-awareness to a naive last-token extraction so a quoted or
# commented-out placeholder would still be caught, and that parsing
# introduced three more bypasses (round 2) — a quoted value with an internal
# space that a space-only split truncated before the placeholder, an escaped
# quote inside a session token that desynced the quote-tracker so a later
# `#` was read as an unquoted comment and discarded the rest of the line,
# and a tab before the change argument that a space-only split never saw as
# a token boundary. Extraction is deleted rather than patched a fourth time.
# Instead the check searches the WHOLE assembled command text for a
# bracketed placeholder whose own text says "guess" (`<...guess...>`,
# case-insensitive) — strictly stronger than extracting a token, and immune
# to quoting, spacing and comments because it never tries to reason about
# any of them. One behaviour is worth stating plainly: a guessed placeholder
# written inside a trailing comment is reported too — a `stage begin` line
# that mentions `<name-or-best-guess>` anywhere is a call site to fix.
#
# MULTI-LINE INVOCATIONS. A real call is frequently wrapped across several
# lines with a trailing `\` continuation, exactly as an ordinary shell
# command would be. This guard joins a `stage begin` line with every
# following line while the previous physical line ends in `\`, then checks
# the assembled logical command as one string — a check against the raw
# physical lines alone would miss `-session-token`/`-harness` written on a
# continuation line, which is precisely how these calls are written today.
#
# STAGE KEYS ARE SERVED, NOT TRANSCRIBED. A `stage begin`'s `-stage <key>`
# names a key of the documented stage vocabulary, which this guard does not
# keep a copy of: the key set is served by `go run ./cmd/flow stage keys`
# from $REPO_ROOT's own stats module (internal/stages' table), so a key the
# checked-out tree serves is exactly a key the tree's Go side accepts --
# one source, no third transcription that could drift from it (KAN-533).
# A key the served vocabulary does not list -- a typo, a rename that missed
# a call site -- is a violation here rather than a caller-mistake exit at
# run time, when the mark silently prints one line and the stage goes
# unrecorded. A serve that fails or yields no key at all is a guard that
# cannot answer (exit 2).
#
# Exit codes: 0 every scanned call is compliant; 1 at least one
# violation was found; 2 the guard cannot answer at all (a bad path, a
# rejected coverage record, or no stage key served by flow stage keys — never
# read as "no findings").
#
# Ported to Go (KAN-841): the body's reasoning for every step is in
# stats/internal/guard/stagemarkcalls.go. The binary does not live in this
# checkout, so this shim exports FLOW_GUARD_REPO_ROOT — this script's parent
# directory, derived as the bash guard derived its REPO_ROOT — as the root
# whose skills/ is scanned by default and whose stats module serves the keys.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-stage-mark-calls: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-stage-mark-calls 2 "check-stage-mark-calls:" "$@"
