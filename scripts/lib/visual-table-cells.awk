# scripts/lib/visual-table-cells.awk — split_cells, trimcell and foldcell,
# defined once.
#
# Sourced (via `awk -f`) by scripts/check-visual-verification.sh,
# scripts/check-visual-trigger.sh and scripts/resolve-visual-screenshots.sh,
# which used to carry three byte-identical copies of this trio. All three
# parse the same `| Setting | Value |` / `| Command | Runs |` table shape
# out of `.flow/project.md`, a file tracked in the repository and editable
# in any pull request — the identical drift hazard
# scripts/lib/resolve-file.sh's and scripts/lib/sanitize-display.sh's own
# headers record for their own functions. One definition, sourced by every
# guard that can safely reach it, is what stops that drift from happening
# here too.
#
# "SAFELY REACH IT" IS THE OPERATIVE PHRASE — see
# scripts/lib/resolve-file.sh's header for the criterion this file follows
# too: a guard that ships through the skills/*/scripts/ symlink farm can
# assume a sibling `lib/` travels with it; a guard reached only by
# hand-copying a single file into an unrelated project's own tooling cannot.
# All three callers above ship through the farm — each carries its own
# `lib` symlink into scripts/lib/ beside it — so each sources this file
# rather than carrying its own copy.
#
# check-workspace-isolation.sh's OWN copy of this trio (widened to four
# columns) is deliberately left alone: that guard's own header records why
# ITS parser stays a copy rather than a sourced helper — it is the same
# parser check-cleanup-complete.sh carries, and THAT guard is hand-copied
# into a project's own tooling standalone, which cannot assume a sibling
# `lib/` travels with it. This task's scope is the three visual guards,
# which carry no such standalone copy anywhere, not a repository-wide
# consolidation of every table parser.
#
# AN awk PROGRAM MADE OF FUNCTION DEFINITIONS ONLY, meant to be combined
# with a caller's own `-f` file via a second `-f` flag on the same
# invocation — awk concatenates every `-f` source into one program before
# parsing, so a file holding only function bodies contributes nothing but
# those functions. Not meant to be run alone; it defines no BEGIN, END or
# pattern-action rule of its own.

# split_cells(line, out) -> the number of cells, populating `out` (an array
# passed by reference) with each cell's raw text. A leading `|` is required;
# a `\|` inside a cell is unescaped and does not split the row; a missing
# trailing `|` is tolerated by dropping one final empty cell.
#
# SCANS BY DELIMITER, NOT BY CHARACTER — TRIED, MEASURED, AND REJECTED FIRST.
# The original version of this function built each cell with `cur = cur ch`,
# one character at a time, and every one of those reassigns `cur` to a
# freshly copied string — quadratic in cell length. This is the SAME defect
# class `scripts/lib/trim-glob-element.sh` already fixed once (see that
# file's own header): a hostile `.flow/project.md` — tracked and editable in
# any pull request, the same fact that file's header records — pads an
# interior cell and the per-character loop pays for every byte of padding
# again on every subsequent byte. This function ALONE, isolated from the
# rest of the guard (called directly on one 120 KB row, old vs new body):
#
#   char loop (old)     ~0.65s
#   delimiter walk (new) ~0.01s
#
# `check-visual-trigger.sh`'s own end-to-end baseline for the same fixture —
# 120 KB config, 60,000 interior padding per side -> 15.44s — is NOT mostly
# this function's own cost, and the task that produced this fix originally
# assumed it was. Measured (isolated from split_cells, old split_cells left
# in place) the OTHER ~14.8s of that 15.44s was already `trimcell`'s own
# cost, for an UNRELATED reason — see that function's header immediately
# below. Both were fixed here because both are in this same shared file and
# both are needed to make the end-to-end number move at all: split_cells's
# own fix alone leaves the guard at ~15s, unchanged within measurement
# noise.
#
# The fix walks `rest` (what remains of the line) with `index()` to find the
# next backslash and the next `|`, and takes ONE `substr()` per run of plain
# text between them — so a long run of non-special bytes (the padding case
# above) costs one substr/concat proportional to ITS OWN length, not one
# copy per byte accumulated over the whole cell. The three-way branch on
# each iteration reproduces the original character loop's escape semantics
# exactly, just applied to a chunk instead of to one byte:
#   - no backslash before the next `|` (or no backslash at all) -> that `|`
#     splits the row; everything before it, taken in one `substr()`, is the
#     completed cell.
#   - a backslash sits at or before the next `|` (or no `|` remains) and is
#     followed by `|` or `\` -> that pair unescapes to the single following
#     character and does NOT split; scanning continues past it.
#   - a backslash followed by anything else -> kept as a literal backslash;
#     the character after it is left for the next iteration to process
#     normally, exactly as the original loop's `continue`-without-`i++` did.
function split_cells(line, out,   rest, bs, pi, nc, n, cur, k) {
  for (k in out) delete out[k]
  gsub(/\r/, "", line)
  sub(/^[[:space:]]+/, "", line)
  sub(/[[:space:]]+$/, "", line)
  if (substr(line, 1, 1) != "|") return 0
  rest = substr(line, 2)
  n = 0
  cur = ""
  while (length(rest) > 0) {
    bs = index(rest, "\\")
    pi = index(rest, "|")
    if (bs == 0 && pi == 0) {
      cur = cur rest
      rest = ""
      break
    }
    if (bs == 0 || (pi > 0 && pi < bs)) {
      # The next special byte is an unescaped `|` — split here.
      cur = cur substr(rest, 1, pi - 1)
      n++; out[n] = cur; cur = ""
      rest = substr(rest, pi + 1)
      continue
    }
    # The next special byte is a backslash, at or before any remaining `|`.
    cur = cur substr(rest, 1, bs - 1)
    nc = substr(rest, bs + 1, 1)
    if (nc == "|" || nc == "\\") {
      cur = cur nc
      rest = substr(rest, bs + 2)
    } else {
      cur = cur "\\"
      rest = substr(rest, bs + 1)
    }
  }
  n++
  out[n] = cur
  if (n > 1 && cur == "") { delete out[n]; n-- }
  return n
}

# trimcell(c) -> c with leading/trailing whitespace and backticks removed.
#
# NO `$`-ANCHORED MATCH AGAINST A COMPUTED STRING — A SECOND, LARGER
# QUADRATIC-LOOKING COST FOUND WHILE MEASURING split_cells's OWN FIX ABOVE.
# The previous body was one `gsub(/^[[:space:]\`]+|[[:space:]\`]+$/, "", c)`.
# That is correct, but on THIS awk (the "one true awk" this repository's
# guards run under; `awk --version` -> `awk version 20200816`) a `$`-anchored
# quantified match against any string that is NOT the just-read input
# record — a `substr()`/concatenation result, a function parameter built
# from one, even `$0` after a plain reassignment — costs seconds at sizes a
# hostile `.flow/project.md` can reach, independent of how the match turns
# out. Confirmed empirically, same 120 KB fixture as split_cells's own
# header, isolated from split_cells entirely (the string handed to trimcell
# built by plain `substr($0, 1, length($0))`, byte-identical to `$0`):
#
#   sub(/^[[:space:]`]+/, "", x)   (leading, `^`-anchored)     ~0.005s
#   sub(/[[:space:]`]+$/, "", x)   (trailing, `$`-anchored)   ~15s
#   match(x, /[[:space:]`]+$/)     (`match`, not `sub`/`gsub`) ~15s
#   match(x, /[[:space:]`]+/)      (SAME class, no `$`)        ~0.005s
#
# So the cost tracks the `$` anchor specifically, not the character class,
# not `sub` vs `match`, and not cell length on its own — a trailing run that
# reaches cleanly to the true end of the string (nothing left to check after
# it) is fast even anchored; a trailing run competing against non-matching
# bytes past it (this cell always has some — the closing backtick, or
# whitespace before the next `|`) is what triggers it. THIS PRE-DATES THIS
# TASK: run against the guard AS IT STOOD BEFORE THIS COMMIT (old split_cells
# and old trimcell together), the SAME 120 KB fixture costs the same ~15s —
# so the char-loop `split_cells` fix above recovers only a small fraction
# (~0.65s of the ~15.44s baseline) of the total; this is the rest of it, and
# it was already there before split_cells was ever quadratic.
#
# THE FIX AVOIDS `$` RATHER THAN WORKING AROUND THIS awk's ENGINE: strip the
# leading run with the (fast) `^`-anchored `sub`, unchanged, then find the
# trailing run's start by walking forward through NON-anchored matches of
# the COMPLEMENT class (`[^[:space:]\`]+`, proven fast above) and keeping
# the end position of the LAST one — everything after that position, to the
# true end, is the trailing run, however long, without ever asking the
# engine "does this reach the end". One or a few `match()`/`substr()` calls
# per cell (one per run of real content), the same shape split_cells's own
# fix above uses for the identical reason.
function rtrim(s,    last_end, off, rest) {
  last_end = 0
  off = 0
  rest = s
  while (match(rest, /[^[:space:]`]+/) > 0) {
    last_end = off + RSTART + RLENGTH - 1
    off += RSTART + RLENGTH - 1
    rest = substr(rest, RSTART + RLENGTH)
  }
  return substr(s, 1, last_end)
}
function trimcell(c) {
  sub(/^[[:space:]`]+/, "", c)
  return rtrim(c)
}

# foldcell(c) -> trimcell(c), lowercased, with internal whitespace runs
# collapsed to one space — the shape a header or a `Setting`/`Command` name
# is compared in. The final `gsub` here is never `$`-anchored, so it is not
# subject to the cost trimcell's own header describes.
function foldcell(c) {
  c = tolower(trimcell(c))
  gsub(/[[:space:]]+/, " ", c)
  return c
}
