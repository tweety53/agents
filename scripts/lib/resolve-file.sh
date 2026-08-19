# scripts/lib/resolve-file.sh — resolve_file, defined once.
#
# Sourced by scripts/check-guard-symlinks.sh, scripts/plan-dispatch-bundles.sh
# and scripts/check-workspace-isolation.sh, which used to carry three
# near-identical copies of this function (plus two more, left alone below).
# The copies had already drifted before this file existed:
# check-guard-symlinks.sh's copy put `--` before every `readlink`, `dirname`
# and `cd` argument; plan-dispatch-bundles.sh's and check-workspace-isolation.sh's
# did not. KAN-73's own review panel caught the drift (see that change's
# final-review-panel.md, F1) — the same shape of defect
# scripts/lib/panel-record.sh's header records for the two guards it was
# extracted from. One definition, sourced by every guard that can safely
# reach it, is what stops that drift from happening a second time.
#
# "SAFELY REACH IT" IS THE OPERATIVE PHRASE, and not every guard carrying a
# copy of this function qualifies. scripts/preserve-session-records.sh keeps
# its own inline copy — untouched by this file — because a guard reached
# only by hand-copying a single file into an unrelated project's own tooling
# cannot assume a sibling `lib/` travels with it; a guard shipped through the
# farm can, because the farm already symlinks `lib` as a directory beside
# every guard that needs it (KAN-73's design.md, "The guard-to-skill map") —
# the identical argument KAN-153's F7 accepted for check-unfinished-work.sh
# sourcing scripts/lib/panel-record.sh, on the evidence that setup.sh
# distributes skills, not scripts/, so a guard and the library it sources
# always travel together in this repository. preserve-session-records.sh
# DOES ship through the farm too (skills/myflow-do/scripts/,
# skills/myflow-fast/scripts/ and skills/myflow-finish/scripts/ all carry
# it, each alongside its own `lib` symlink), so it qualifies under this same
# criterion and staying uninlined here is a gap, not a deliberate exemption
# — out of this change's scope to close, since none of this change's own
# callers exercise that script. scripts/gather-self-review-context.sh
# previously kept its own inline copy of resolve_file too, on the same
# false premise that it did not ship through the farm; it now sources
# lib/within-root.sh for within_root, but still carries its own inline
# resolve_file (a differently-shaped, non-hardened variant than this file's
# own — see that script's header), left untouched here since this change's
# scope was the within_root duplication, not resolve_file's. Follow-up
# question, still open: whether preserve-session-records.sh's and
# gather-self-review-context.sh's own resolve_file copies should ever join
# this file.
#
# Not meant to be executed directly — a caller sources it and calls
# resolve_file; it sets no `set -euo pipefail` of its own and relies on the
# sourcing script's.
#
# `--` BEFORE EVERY readlink, dirname AND cd ARGUMENT — the hardened form,
# carried here rather than the un-hardened one four of the five original
# copies used, so a path beginning with `-` is never read as an option. This
# is the same discipline scripts/lib/panel-record.sh's header states for
# every guard in this repository.
#
# resolve_file <path> -> prints the path's resolved PHYSICAL location on
# stdout, or returns 1 if it cannot. macOS /bin/sh has no `readlink -f`, so
# this walks symlinks by hand, capped at 40 hops against a cycle.
#
# NORMALIZES ITS OWN INPUT (F9, task 9's post-commit review, pass 2): every
# caller in this repository used to be responsible for handing this function
# an already-normalized path (no `.`, no trailing slash) itself, with that
# obligation recorded only in the callers' own comments (gather-dispatch-context.sh's
# own header) rather than enforced here, where a caller that forgot it —
# check-guard-symlinks.sh:250 passes $entry, a path straight from a directory
# scan, not a caller-normalized one — got silently wrong output instead of a
# refusal. The dirname/basename split below assumes p's final path component
# is a real leaf name; a trailing slash or a bare/buried `.` or `..`
# component breaks that assumption, so both are normalized away before the
# split ever runs, making every caller's own workaround redundant
# belt-and-braces rather than the only thing standing between this function
# and the bug.
resolve_file() {
  local p="$1" hops=0 target dir

  # Trailing slashes: strip every one so "/tmp/" resolves identically to
  # "/tmp" instead of losing its leaf entirely (the dirname/basename split
  # below reads "" as the leaf of any path ending in "/"). The root "/"
  # itself is left alone, and IS handled below (a bare "/" or a run of
  # slashes that collapses to it) -- see the dir == "/" guards further
  # down, both in this loop and at the function's final assembly.
  while [ "$p" != "/" ] && [ "${p%/}" != "$p" ]; do
    p="${p%/}"
  done

  while [ -L "$p" ]; do
    hops=$((hops + 1))
    [ "$hops" -le 40 ] || return 1
    target="$(readlink -- "$p")" || return 1
    case "$target" in
      /*) p="$target" ;;
      *)
        # A relative symlink target is joined onto its own link's parent
        # directory. When that parent is root, plain "$dir/$target"
        # concatenation produces a doubled leading slash ("//private/tmp"
        # rather than "/private/tmp") -- and under bash (unlike this
        # repository's interactive zsh), a leading "//" is not collapsed
        # by a later `cd -P`/`pwd -P` round-trip, so a root-parented
        # symlink (e.g. macOS's /tmp -> private/tmp) never lost the
        # doubled slash once introduced here (F16, pass 3 of KAN-201's
        # own review panel).
        dir="$(dirname -- "$p")"
        case "$dir" in
          /) p="/$target" ;;
          *) p="$dir/$target" ;;
        esac
        ;;
    esac
  done

  # A path whose final component is "." or ".." -- a bare "." or "..", or
  # one buried at the end of a longer path like "foo/." -- names a
  # directory, not a leaf inside one. Splitting it into dirname/basename
  # the ordinary way (below) would keep that "." or ".." as a spurious
  # leaf, appended to its own parent's resolved path. Resolving the whole
  # thing as a directory instead, with nothing appended, is what makes
  # `resolve_file "."` answer the process's own resolved working directory
  # rather than that path plus a trailing "/.".
  case "${p##*/}" in
    . | ..)
      dir="$(cd -P -- "$p" 2>/dev/null && pwd -P)" || return 1
      printf '%s\n' "$dir"
      return 0
      ;;
  esac

  dir="$(cd -P -- "$(dirname -- "$p")" 2>/dev/null && pwd -P)" || return 1
  # Root ("/") is the one directory whose own path already ends in the
  # separator this line's usual "$dir/$leaf" join would otherwise add a
  # second one of -- so a root-parented leaf (including "/" itself, whose
  # leaf is empty) is assembled without it, matching this file's own
  # header comment ("the root / itself is left alone").
  case "$dir" in
    /) printf '%s\n' "/${p##*/}" ;;
    *) printf '%s\n' "$dir/${p##*/}" ;;
  esac
}
