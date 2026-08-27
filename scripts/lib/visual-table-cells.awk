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
function split_cells(line, out,   i, ch, nxt, n, cur, len, k) {
  for (k in out) delete out[k]
  gsub(/\r/, "", line)
  sub(/^[[:space:]]+/, "", line)
  sub(/[[:space:]]+$/, "", line)
  if (substr(line, 1, 1) != "|") return 0
  line = substr(line, 2)
  len = length(line)
  n = 0
  cur = ""
  for (i = 1; i <= len; i++) {
    ch = substr(line, i, 1)
    if (ch == "\\" && i < len) {
      nxt = substr(line, i + 1, 1)
      if (nxt == "|" || nxt == "\\") { cur = cur nxt; i++; continue }
      cur = cur ch
      continue
    }
    if (ch == "|") { n++; out[n] = cur; cur = ""; continue }
    cur = cur ch
  }
  n++
  out[n] = cur
  if (n > 1 && cur == "") { delete out[n]; n-- }
  return n
}

# trimcell(c) -> c with leading/trailing whitespace and backticks removed.
function trimcell(c) {
  gsub(/^[[:space:]`]+|[[:space:]`]+$/, "", c)
  return c
}

# foldcell(c) -> trimcell(c), lowercased, with internal whitespace runs
# collapsed to one space — the shape a header or a `Setting`/`Command` name
# is compared in.
function foldcell(c) {
  c = tolower(trimcell(c))
  gsub(/[[:space:]]+/, " ", c)
  return c
}
