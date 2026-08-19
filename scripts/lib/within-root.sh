# scripts/lib/within-root.sh — within_root, defined once.
#
# Sourced by scripts/gather-dispatch-context.sh and
# scripts/gather-self-review-context.sh, both of which used to carry their
# own byte-for-byte copy of this function — gather-dispatch-context.sh's
# copied verbatim from gather-self-review-context.sh's (that script's own
# comment said so). This is a security-relevant path BOUNDARY check — the
# containment test that decides whether a change directory's own content
# source escapes
# <change-root> — and a check like that must not be correctable in one
# script and not its sibling, the same reasoning that moved resolve_file
# into scripts/lib/resolve-file.sh. One definition, sourced by every guard
# that can safely reach it, is what stops that drift from happening again.
#
# "SAFELY REACH IT" IS THE OPERATIVE PHRASE — see
# scripts/lib/resolve-file.sh's header for the criterion this file follows
# too: a guard that ships through the skills/*/scripts/ symlink farm can
# assume a sibling `lib/` travels with it; a guard reached only by
# hand-copying a single file into an unrelated project's own tooling cannot.
# gather-self-review-context.sh DOES ship through the farm —
# skills/myflow-fast/scripts/ and skills/myflow-finish/scripts/ both carry
# it, each alongside its own `lib` symlink into scripts/lib/ — so it sources
# this file rather than carrying its own copy.
#
# Not meant to be executed directly — a caller sources it and calls
# within_root; it sets no `set -euo pipefail` of its own and relies on the
# sourcing script's.

# within_root <resolved-path> <root> — true iff <resolved-path> is <root>
# itself or lives under it. A path BOUNDARY test, not a string prefix test
# (a bare prefix would wrongly accept "/foo/bar-evil" against "/foo/bar").
within_root() {
  case "$1" in
    "$2" | "$2"/*) return 0 ;;
    *) return 1 ;;
  esac
}
