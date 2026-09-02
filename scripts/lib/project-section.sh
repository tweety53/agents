# scripts/lib/project-section.sh — project_section, defined once.
#
# Sourced by scripts/project-get.sh, scripts/gather-dispatch-context.sh and
# scripts/check-model-keys.sh — three inline copies of the identical
# heading-to-next-heading awk used to live one in each, the same drift
# hazard scripts/lib/within-root.sh's own header records for
# gather-dispatch-context.sh's and gather-self-review-context.sh's copy of
# within_root. One definition, sourced by every caller that can safely
# reach it, is what stops that drift from happening a second time here.
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
# project_section; it sets no `set -euo pipefail` of its own and relies on
# the sourcing script's. It sources scripts/lib/strip-bom.sh relative to its
# own ${BASH_SOURCE[0]} directory, so a leading UTF-8 BOM in
# `.flow/project.md` never hides a `## <key>` heading on the file's own
# first line — see that file's header for why.

PROJECT_SECTION_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$PROJECT_SECTION_LIB_DIR/strip-bom.sh"

# project_section <file> <key> -> prints the body of the first "## <key>"
# heading: every line after it up to, not including, the next "^## " line
# or EOF ("### " subheadings included), leading and trailing blank lines
# removed, nothing else normalised. Prints nothing when <key> is absent.
project_section() {
  local file="$1" key="$2"
  strip_bom_cat "$file" | awk -v key="## $key" '
    $0 == key { grab = 1; next }
    /^## / { if (grab) exit }
    grab { lines[++n] = $0 }
    END {
      s = 1; e = n
      while (s <= e && lines[s] ~ /^[[:space:]]*$/) s++
      while (e >= s && lines[e] ~ /^[[:space:]]*$/) e--
      for (i = s; i <= e; i++) print lines[i]
    }'
}
