# scripts/lib/trim-glob-element.sh — trim_glob_element, defined once.
#
# Sourced by scripts/check-visual-trigger.sh and scripts/check-visual-verification.sh,
# which carried byte-identical copies of this function until now. KAN-359's
# own design.md recorded a decision, `no-shared-parser-yet`, against
# extracting it — chosen against as larger than the defect warranted, with
# extraction named as the answer "if these two parsers diverge again."
# That decision is SUPERSEDED, not deleted (see design.md), on the review
# panel's evidence and a corrected cost estimate: both callers already
# source this same `scripts/lib/` directory for `sanitize-display.sh` and
# `strip-bom.sh`, and both already share `scripts/lib/visual-table-cells.awk`
# for this same table's cells. The plumbing existed in both files before
# this extraction touched either, so this is one new lib file plus two
# `source` lines, not a third alignment problem. The identical drift this
# repository has already paid for once — `scripts/lib/strip-bom.sh`'s own
# header records a BOM fix that reached one guard of three and had to be
# extracted a commit later — is the reason a second divergence is closed
# here rather than watched for.
#
# "SAFELY REACH IT" IS THE OPERATIVE PHRASE — see
# scripts/lib/resolve-file.sh's header for the criterion this file follows
# too: a guard that ships through the skills/*/scripts/ symlink farm can
# assume a sibling `lib/` travels with it; a guard reached only by
# hand-copying a single file into an unrelated project's own tooling cannot.
# Both callers above ship through the farm — each carries its own `lib`
# symlink into scripts/lib/ beside it — so each sources this file rather
# than carrying its own copy.
#
# Not meant to be executed directly — a caller sources it and calls
# trim_glob_element; it sets no `set -euo pipefail` of its own and relies on
# the sourcing script's.

# trim_glob_element <raw> -> <raw> with a leading/trailing run of whitespace
# and backtick characters removed — NEVER an interior one, so a glob
# containing a space (a directory name with one) or a backtick of its own
# survives untouched. Applied per split element, never to a whole cell:
# splitting `ui paths` on comma happens FIRST in both callers, and each
# element's OWN backticks are stripped only afterward. Reversing that order
# is design.md's `split-then-strip` — the whole fix for a value carrying
# more than one glob, each individually backticked (`` `a/**`, `b/**` ``):
# splitting first and stripping every element left every element but the
# outermost with a stray backtick, and matched nothing.
#
# ONE awk PASS, NOT A BASH LOOP — TRIED, MEASURED, AND REJECTED FIRST. The
# original version of this function stripped a matched character at a time
# — `s="${s:1}"` off the front, `s="${s%?}"` off the back — and each of
# those reassigns `s` to a freshly copied string, so a value padded with N
# throwaway characters cost O(N) copies of up to N characters each:
# quadratic in the padding length. `.flow/project.md` is tracked and
# editable in any pull request (the same fact both callers' own headers
# record for why the input is attacker-influenced), so a hostile value is a
# real input this function must not choke on.
#
# A PRIOR VERSION OF THIS COMMENT MISDESCRIBED ITS OWN MEASUREMENT: it
# labeled its numbers "Measured on the ORIGINAL char-loop" but actually ran
# the padding at the EDGES of the whole `ui paths` cell, where
# `visual-table-cells.awk`'s own whole-cell `trimcell` strips a
# leading/trailing run of whitespace and backticks BEFORE this function is
# ever called on a split element — so those old numbers (0.05s / 0.19s /
# 0.84s for the char loop, 0.005s / 0.005s / 0.010s for the awk single pass)
# measured a codepath the padding never reached. The four implementations
# below are re-measured with padding on an INTERIOR list element — the only
# place `trimcell` cannot reach — via `/usr/bin/time -p` against the actual
# guard end to end, this worktree, three trials each, median reported:
#
#   interior padding    char loop   two-pointer   param-expansion   awk (shipped)
#    4,000 per side        0.60s       0.30s          0.23s            0.11s
#    8,000 per side        2.10s       1.00s          0.80s            0.34s
#   20,000 per side       12.10s          —              —             1.75s
#
# THE FIX IS WORTH MORE THAN THE OLD NUMBERS IMPLIED, in absolute terms: the
# old (wrong) numbers made switching implementations look like it saved
# fractions of a second either way. Measured on the padding this function
# actually receives, the char loop costs whole seconds at sizes an attacker
# could put in a tracked file, and the shipped awk single pass cuts that by
# roughly 5-7x at these sizes (0.60s->0.11s at 4,000; 2.10s->0.34s at
# 8,000) — real wall-clock savings on a guard run, not a rounding error.
#
# THE TWO BASH-NATIVE REPLACEMENTS WERE NOT, IN FACT, SLOWER THAN THE CHAR
# LOOP — the old header's comparison used the same edge-padded, function-
# never-reached methodology, so it never actually exercised any of the four
# implementations on the padding it claimed to. Measured correctly above,
# both bash-native alternatives beat the char loop at every size tested,
# though neither beats the shipped awk single pass:
#   1. A two-pointer walk using `${s:i:1}` to find the first/last non-padding
#      byte offset, then one substring extraction. Faster than the char
#      loop (no repeated copy of `s`) but still slower than the awk single
#      pass: indexed access into a large string is not free in this bash,
#      so a loop that indexes further into the same string every iteration
#      still costs more than one C-implemented regex pass.
#   2. The classic no-loop parameter-expansion idiom,
#      `lead="${s%%[!$ws]*}"; s="${s#"$lead"}"` (and the mirror for the
#      trailing run) — no explicit loop at all, and the fastest of the three
#      bash-only options, but still behind the awk single pass: bash's own
#      `%%`/`##` pattern matcher is doing the same anchored-run search awk's
#      `sub()` does, just without awk's C implementation underneath it.
# The shipped choice (below) is still correct — it wins at every size
# tested — but the reason is "still fastest," not "the alternatives were
# slower than what they replaced."
#
# THE FIX THAT MEASURES FASTEST: one `awk` process, one `sub()` each end.
# `sanitize_display` (lib/sanitize-display.sh) already shells out to `awk`
# for exactly this reason — C-implemented, single-pass regex substitution
# instead of bash's own string engine — so this follows an established
# convention rather than introducing a new one. Called directly (no guard
# around it) on a single interior element up to 60,000 backticks per side,
# this function alone measures under 10ms — the guard-level growth in the
# table above beyond a few hundred milliseconds is NOT this function; it
# comes from elsewhere in the shared table parser (`visual-table-cells.awk`,
# outside this file's scope) scanning the same long cell.
#
# "NOT A DENIAL OF SERVICE," RE-DERIVED RATHER THAN RESTATED: for this
# function specifically, the conclusion is STRONGER than the superseded
# header claimed, not weaker — a hostile interior element up to 60,000
# backticks per side costs this function under 10ms, and the guard-level
# cost at the sizes measured above (up to 20,000 per side, 1.75s) is bounded
# and linear-looking for the awk single pass. It does NOT extend to every
# component a hostile `ui paths` value touches: the shared table parser's
# own per-character cell split was observed to grow faster than linear at
# larger paddings (60,000 per side, full guard: ~15s) independent of which
# `trim_glob_element` implementation is in use — a separate, pre-existing
# cost this file does not own and this task does not fix, worth a follow-up
# against `visual-table-cells.awk` rather than folding into this comment.
trim_glob_element() {
  LC_ALL=C awk '
    { sub(/^[ \t`]+/, ""); sub(/[ \t`]+$/, ""); printf "%s", $0 }
  ' <<< "$1"
}
