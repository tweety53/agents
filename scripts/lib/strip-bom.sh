# scripts/lib/strip-bom.sh — strip_bom_cat, defined once.
#
# Sourced by scripts/check-visual-verification.sh, scripts/check-visual-trigger.sh
# and scripts/resolve-visual-screenshots.sh. Task 15 had already extracted
# this trio's other shared parsing helpers (split_cells/trimcell/foldcell in
# visual-table-cells.awk, sanitize_display, git_clean) into this same
# scripts/lib/ directory; task 16b's BOM fix arrived one commit later and
# landed in check-visual-verification.sh alone instead of beside them. The
# other two guards share the identical heading-parse convention this
# function feeds and did not get the fix: check-visual-trigger.sh — the
# guard that decides whether flow.visual-verify runs at all — read a
# BOM-prefixed `## visual verification` heading as ABSENT and exited 2,
# which the stage's step 1 treats as "not configured";
# resolve-visual-screenshots.sh made the identical misread. A project that
# declares the section correctly, in a file a BOM happens to sit on, got no
# visual verification and no warning — the exact silent-skip class this
# whole change exists to prevent, materialising from a copied fix inside its
# own change, one commit after the extraction that was supposed to make a
# copied fix impossible. This is the argument for the shared lib, not
# against it: the hazard task 15's own header predicted — "the next fix to a
# shared helper must reach every copy, and nothing checks that it did" — is
# what happened here. One definition, sourced by every guard that can safely
# reach it, is what stops the next copy from being missed a third time.
#
# "SAFELY REACH IT" IS THE OPERATIVE PHRASE — see scripts/lib/resolve-file.sh's
# header for the criterion this file follows too: a guard that ships through
# the skills/*/scripts/ symlink farm can assume a sibling `lib/` travels
# with it; a guard reached only by hand-copying a single file into an
# unrelated project's own tooling cannot. All three callers above ship
# through the farm — each carries its own `lib` symlink into scripts/lib/
# beside it — so each sources this file rather than carrying its own copy.
#
# Not meant to be executed directly — a caller sources it and calls
# strip_bom_cat; it sets no `set -euo pipefail` of its own and relies on the
# sourcing script's.

# strip_bom_cat <file> -> writes <file>'s content to stdout, with a leading
# UTF-8 BOM (EF BB BF) — if the file's own literal first three bytes carry
# one — stripped before anything downstream reads it. A `.flow/project.md`
# some editor or Windows tooling saved with a BOM, with `## visual
# verification` as the file's own first line, puts those three bytes before
# the `#` every caller's `^##` heading anchor expects at column one — under
# some invocations' locale handling the anchor then never matches, and a
# section that is genuinely present reads as absent. Detected once per call,
# at the byte level, rather than left to whichever locale a given invocation
# happens to run under. This fails SAFE when a caller treats "absent" as
# "not configured" rather than "grant nothing" — but a silent misparse of a
# heading that is present is still closed here rather than left open.
strip_bom_cat() {
  local file="$1"
  if [ "$(LC_ALL=C head -c 3 "$file" 2>/dev/null)" = "$(printf '\xef\xbb\xbf')" ]; then
    tail -c +4 "$file"
  else
    cat "$file"
  fi
}
