#!/usr/bin/env bash
# check-self-review-report.sh — every report under docs/self-review/ carries
# every self-review angle the canonical table serves, each section carrying
# either a finding line
# or the none-marker, each finding line parseable with its label matching
# its section, a filed finding naming an issue key and a fixed finding
# naming a sha.
#
# Usage: scripts/check-self-review-report.sh [dir]
#
# With no argument it scans `docs/self-review/`, resolved relative to the
# repo root — the one tree self-review reports live under. A caller may pass
# an explicit directory instead (the Go tests do, against t.TempDir()
# fixtures).
#
# WHAT THIS CATCHES. KAN-200's own report shape — one `##` section per
# angle, one
# parseable line per finding, an explicit none-marker for an empty angle —
# was, before this guard, checked by nothing but a reviewer's prose reading.
# An angle silently omitted from a report read exactly like an angle that
# produced nothing, which is how KAN-73's cost angle passed unnoticed while
# its section existed. This guard makes an incomplete or malformed report a
# loud, mechanical fact instead.
#
# THE REPORT SHAPE (the kan-200 self-review-filing tasks, "The report shape
# this change defines"):
#
#   ## <prose> — `<label>`
#
#   - **[<label>]** <text> — fixed: <sha>
#   - **[<label>]** <text> — filed: <KEY>
#   - **[<label>]** <text> — declined
#
# An angle with no findings carries `_none — this angle produced no
# findings._` in place of finding lines, never both a marker and finding
# lines together. The label on a finding line must match the label in its
# own section's heading. `<KEY>` is an uppercase project key, a hyphen, and
# digits — e.g. `KAN-201`; anything else (empty, `yes`, a bare number, a
# trailing hyphen) is a violation naming the malformed key. `<sha>` is the
# landed commit of a finding the self-review pass fixed (KAN-875): 7–40
# lowercase hex characters; an empty or non-hex sha is a violation naming it.
#
# The `##` sections appear in the order the canonical table states,
# each exactly once — an out-of-order or duplicate section is a named
# violation. A heading at ANY level (`#`, `###`, …) ends the current
# section, not only `##`. A finding-shaped line appearing before any
# recognized section heading is a named violation rather than a silent drop.
#
# THE ANGLE LABELS ARE SERVED, NOT COPIED (kan-585). The canonical angle
# table — step 2's numbered table in
# skills/flow-self-review/SKILL.md, the one source
# jira-integration.md already cites instead of copying — is parsed at run
# time into ANGLE_LABELS, in table order. A label renamed or added there
# moves this guard first: every report still carrying the old spelling
# reports a missing section for the new label, and the human decides
# whether the rename or the reports give way. The guard can never lag the
# table the way a hardcoded copy would. (The original five came from the
# kan-200 self-review spec's "One combined reasoning pass" angle table.)
#
# FIVE-ANGLE REPORTS ARE A FROZEN LIST. `five-angle-reports.txt` in the
# scanned directory names, one basename per line, every report written
# before angle 6 (`flow-speed`) existed. A listed report is checked against
# the first five angles of the table only; every report not on the list
# must carry every angle the table serves. The list is frozen: it never
# gains an entry, so a new report cannot be exempted from a later angle. A
# directory with no list exempts nothing. A list that exists but cannot be
# read is "cannot answer" (exit 2), never an empty exemption set.
#
# PER-REPORT COVERAGE, via coverage.go (scripts/lib/coverage.sh's Go twin).
# Each report's recorded
# count is the number of section-level checks this guard actually performed
# on it: one per angle heading it FOUND (not one per angle attempted), plus
# one per finding line it found under a recognized heading. A report that is
# not recognizable as a self-review report at all — none of the angle
# headings present anywhere in it — gets NO per-section "missing" findings,
# because there was nothing in it to check against; it is instead flagged
# entirely through coverage.sh's undeclared-zero mechanism, so a report
# checked for nothing is a named, failing fact rather than a silent pass
# (KAN-73's own regression shape, reproduced as case 8 of this guard's Go
# test). A report that IS recognizable — at least one angle
# heading present — but is missing one or more of the others gets an
# explicit "missing section" finding for each absent one, on top of whatever
# non-zero count its present sections contribute.
#
# ADOPTED, NOT INVENTED: every guard in this repository holds to three
# disciplines — `-a` on every grep, the `rc > 1` split between "no match"
# and a real error, and `--` before every path. This guard uses neither
# `grep` nor `awk` against a report — each report is read whole and every
# line is classified at the point it is read — so the first and third
# disciplines have no call site on the report path, exactly as
# scripts/lib/coverage.sh's own header states for the same reason. The
# canonical-table parse reads a generated-shaped Markdown table, not
# untrusted report prose. The posture behind them still applies: the
# directory walk and each report file's own read are the two remaining I/O
# operations this guard has to face, and each is checked by its own error
# rather than folded into a reassuring empty result. Neither is checked by a
# precondition standing in for it — the status of the operation, or nothing.
# A read that fails partway through a file it opened successfully — recorded
# under KAN-211 as a known limit of the all-bash shape — is now that read's
# own error too.
#
# SELF-REVIEW CONTEXT BUNDLES ARE NOT REPORTS (KAN-512). Run 1 commits
# `docs/self-review/<name>-context.md` on every run (canonical: **Save the
# self-review context bundle**, `skills/flow-contracts/finish-contract-run1.md`), and
# `/flow-self-review <name>` deletes it once it runs the deferred pass. It is
# a bundle awaiting a reasoning pass, never a report of one — the directory
# walk excludes `*-context.md` outright so a pending bundle is neither scanned for
# the angle shape (which it does not carry) nor flagged as an
# undeclared-zero coverage violation naming it.
#
# NO RECORD PROTOCOL, DELIBERATELY (KAN-211). This guard used to classify
# each report in one `awk` pass, serialize the result as records whose fields
# were joined with ASCII Unit Separator (0x1F, never tab), and re-parse those
# records back in bash. Three of KAN-200's findings — G2, G3 and H2 — were
# taxes on that one seam, each a manifestation of the same fact: a value had
# to survive a hand-written encoding across two languages with different
# escaping and splitting rules. H2 reopened the class one fix round after G3
# supposedly closed it, which is why KAN-211 collapsed the boundary rather
# than patching a fourth manifestation. Do not reintroduce one: the
# classification body reads each line directly and is already correct, and
# an intermediate encoding has nothing to do there but re-encode what it can
# read directly. A raw 0x1F byte in a report is ordinary content now — it was
# rejected only because it was that protocol's field delimiter, and the rule
# was retired with the protocol that created it.
#
# Exit codes: 0 clean, 1 violations found (including an undeclared-zero
# report or an empty corpus), 2 cannot answer at all (a missing or
# unreadable target directory, an unreadable report file inside it, or an
# internal coverage record failing).
#
# Ported to Go (KAN-873): the body's reasoning for every step is in
# stats/internal/guard/selfreviewreport.go, beside the code. The default
# directory and the canonical table resolve from the repository root this
# script sits in, so this shim exports it as FLOW_GUARD_REPO_ROOT, derived as
# the bash's `$SCRIPT_DIR/..` was. flow-guard is built from this checkout,
# never taken from PATH: scripts/lib/flow-guard.sh derives it, and exits 2
# (this guard's cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-self-review-report: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-self-review-report 2 "check-self-review-report:" "$@"
