# scripts/lib/sanitize-display.sh — sanitize_display, defined once.
#
# Sourced by scripts/check-visual-verification.sh, scripts/check-visual-trigger.sh
# and scripts/resolve-visual-screenshots.sh, which used to carry three
# byte-identical copies of this function. All three read `.flow/project.md`,
# a file tracked in the repository and editable in any pull request, and
# echo cells out of it back to the operator's terminal on a violation or a
# cannot-answer exit — a copy that missed the next escape a terminal control
# sequence needs would be a forged-verdict hazard in exactly one of the
# three and not the other two, the same drift `scripts/lib/resolve-file.sh`
# and `scripts/lib/within-root.sh` were written to close. One definition,
# sourced by every guard that can safely reach it, is what stops that drift
# from happening here too.
#
# "SAFELY REACH IT" IS THE OPERATIVE PHRASE — see
# scripts/lib/resolve-file.sh's header for the criterion this file follows
# too: a guard that ships through the skills/*/scripts/ symlink farm can
# assume a sibling `lib/` travels with it; a guard reached only by
# hand-copying a single file into an unrelated project's own tooling cannot.
# All three callers above ship through the farm — each carries its own
# `lib` symlink into scripts/lib/ beside it — so each sources this file
# rather than carrying its own copy. `check-cleanup-complete.sh` and
# `check-workspace-isolation.sh` carry their own, separate copies of this
# same function and are deliberately left alone: KAN-73's guard-to-skill map
# does not change what either of those two already does, and this task's
# scope is the three visual guards, not a repository-wide consolidation.
#
# Not meant to be executed directly — a caller sources it and calls
# sanitize_display; it sets no `set -euo pipefail` of its own and relies on
# the sourcing script's.

# sanitize_display — copy stdin to stdout with every C0 control byte, DEL and
# backslash rendered as visible text. Every violation or cannot-answer line a
# caller prints through this function quotes a cell straight out of
# `.flow/project.md`, so a cell holding an escape sequence would otherwise
# write that sequence to the operator's terminal — a forged-verdict hazard.
# ESCAPED, NOT STRIPPED AND NOT REFUSED: stripping rewrites the text the
# operator is sent to go and fix, and refusing hands whoever edited the file
# a way to suppress a real finding.
sanitize_display() {
  LC_ALL=C awk '
    BEGIN {
      for (j = 1; j < 32; j++) esc[sprintf("%c", j)] = sprintf("\\x%02x", j)
      esc[sprintf("%c", 127)] = "\\x7f"
      esc["\\"] = "\\\\"
    }
    {
      out = ""
      n = length($0)
      for (i = 1; i <= n; i++) {
        c = substr($0, i, 1)
        out = out ((c in esc) ? esc[c] : c)
      }
      print out
    }
  '
}
