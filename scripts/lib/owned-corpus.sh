# scripts/lib/owned-corpus.sh — the Markdown this repository owns, defined once.
#
# Two guards need the same answer to "which Markdown files are this
# repository's own content?": check-normative-inventory.sh, which inventories
# the normative sentences in them, and check-contract-budget.sh, which will
# measure them once its own widening task lands. If those two disagreed about
# ownership, a file could be over budget in one guard's view and invisible to
# the other's inventory — and the inventory is the only thing standing between a
# corpus-wide prose trim and a silently deleted requirement. So the resolution
# lives here, in one implementation both call, rather than in two copies that
# drift the way resolve_file's five copies did (see scripts/lib/resolve-file.sh).
#
# THE CORPUS is every `.md` and `.mdc` file under these scope roots:
#
#   skills/  rules/  spectre/specs/  commands/  commands-claude/  .myflow/
#
# plus the `.md`/`.mdc` files sitting directly at the repository root
# (README.md, AGENTS.md, CLAUDE.md today). The root is deliberately NOT scanned
# recursively: doing that would sweep in spectre/changes/, whose session
# records and panel reports a `/myflow-*` run writes DURING a change, so an
# inventory captured before a change's first edit could never match one captured
# after its last, however faithfully the prose was preserved.
#
# `openspec/` NAMES NO SCOPE ROOT HERE, deliberately. It is this repository's
# frozen, pre-spectre tree — history, not something a `/myflow-*` run edits
# again — so it is never linted, budgeted or inventoried, exactly like
# `openspec/changes/archive/` was already excluded below before this file's
# corpus moved: a frozen tree gets the same treatment an already-archived
# change got, for the same reason. Nothing needs to name it as an exclusion
# for that to hold; it simply is not one of the roots above, and the two
# callers this file serves only ever look under the roots they are given.
#
# THE EXCLUSIONS are structural — a path component, or a path prefix, never a
# list of filenames:
#
#   any component named `node_modules`   vendored third-party text
#   any component named `.superpowers`   a run's own scratch records
#   spectre/changes/archive/             archived changes, frozen by definition
#   docs/superpowers/                    a vendored copy of another project's skills
#
# The last three sit outside every scope root above, so today they are already
# unreachable; they are stated anyway because this function is the one place
# ownership is defined, and a later widening of the scope roots must not silently
# acquire them. A filename list would be the wrong shape for all four: it would
# pass the moment a vendored tree renamed a file.
#
# SYMLINKS ARE NOT OWNED CONTENT. Enumeration uses `find -type f`, which neither
# follows a symlinked directory nor reports a symlinked file, so a `.md` symlink
# contributes nothing. That is the honest answer for both callers: a symlink has
# no size of its own to ratchet and no sentences of its own to inventory — its
# target has both, and is inventoried in its own right if this repository owns it.
#
# A SCOPE ROOT THAT IS A SYMLINK IS REFUSED, not resolved and not skipped. The
# distinction matters because the two behaviours are indistinguishable on stdout
# and only one of them is honest: `[ -d ]`, `[ -r ]` and `[ -x ]` all follow
# symlinks, so a link to a directory passes every presence test below, while the
# `find` that follows runs in physical mode and does not descend into a
# symlinked starting point — the entire scope root then contributes zero files,
# with no error and no warning. That is the quietly smaller corpus the
# missing-root refusal exists to prevent, arrived at from the other direction,
# and it is worse there: the caller's before/after inventory comparison silently
# stops covering whatever the link swallowed. Following the link is not the fix.
# A corpus root that is a symlink is a question about this repository's shape,
# which a caller has to answer; it is not something a guard may paper over.
#
# A SYMLINKED DIRECTORY NESTED INSIDE A SCOPE ROOT IS REFUSED WHEN FOLLOWING IT
# WOULD REACH ANY `.md` OR `.mdc` FILE, AND SKIPPED OTHERWISE. This is the same
# defect as the scope-root refusal above, one level deeper: the `-L` test above
# guards `$root/$dir` and nothing beneath it, and the `find` below runs in
# physical mode, so it walks past a symlinked subdirectory without descending
# into it. Whatever the link points at is then absent from the corpus with no
# error and no warning — invisible to the inventory a bulk prose trim is checked
# against, and outside the ratchet a covered file is measured by.
#
# The predicate is deliberately the coarse one: ANY reachable `.md` or `.mdc`,
# not "any file the corpus does not already contain by another route". A link
# pointing at a directory inside the same scope root loses nothing, because the
# physical walk reaches the real target independently — so the coarse rule can
# refuse a link that was in fact harmless. That direction is safe and it is
# cheap to answer: a refusal is loud, names the path, and is resolved by a human
# deciding what this repository's shape should be. The precise rule would have
# to compare two file sets keyed on identity rather than path, which is a second
# enumeration and a second thing to get wrong, for the sole benefit of tolerating
# a shape nothing here needs.
#
# The other half of the rule is what keeps the real repository green. This
# repository carries three deliberate nested directory links today —
# skills/myflow-do/scripts/lib, skills/myflow-fast/scripts/lib and
# skills/myflow-finish/scripts/lib, each pointing at ../../../scripts/lib so a
# guard script is installed beside the skill that invokes it. They hold `.sh`
# files and no Markdown, so not descending into them loses nothing, and refusing
# them would fail every run of both guards against the real repository — a worse
# outcome than the bug, since a guard nobody can run green is a guard nobody
# runs. Following a link is still not on the table: what is owned is decided by
# where a file physically lives.
#
# Not meant to be executed directly — a caller sources it and calls its
# functions; it sets no `set -euo pipefail` of its own and relies on the sourcing
# script's.

# The scope roots, relative to the repository root. A bash array rather than a
# function printing lines, so a caller iterates it without word-splitting.
OWNED_CORPUS_SCOPE_DIRS=(skills rules spectre/specs commands commands-claude .myflow)

# owned_corpus_excluded <path-relative-to-repo-root> -> exit 0 when the path is
# excluded from the corpus, 1 otherwise. Every test is a path-component or
# path-prefix test; none matches a bare basename, so a file is never excluded
# for what it is called, only for where it lives.
owned_corpus_excluded() {
  case "$1" in
    node_modules|node_modules/*|*/node_modules|*/node_modules/*) return 0 ;;
    .superpowers|.superpowers/*|*/.superpowers|*/.superpowers/*) return 0 ;;
    spectre/changes/archive|spectre/changes/archive/*) return 0 ;;
    docs/superpowers|docs/superpowers/*) return 0 ;;
  esac
  return 1
}

# owned_corpus_files <repo-root> — print every owned Markdown file's absolute
# path, NUL-terminated, so a filename containing a newline stays one entry.
# Returns 2, with a message on stderr, when it cannot answer at all: a root that
# is not a readable directory, a scope root that is missing or unreadable, or a
# `find` that failed. A missing scope root is refused rather than skipped —
# skipping it would hand the caller a quietly smaller corpus that still looks
# like a complete answer.
owned_corpus_files() {
  if [ "$#" -ne 1 ]; then
    printf 'owned_corpus_files: want 1 argument (repo root), got %s\n' "$#" >&2
    return 2
  fi
  local root="$1" dir work rc=0 path rel found
  if [ -z "$root" ]; then
    printf 'owned_corpus_files: repo root must not be empty\n' >&2
    return 2
  fi
  if [ ! -d "$root" ] || [ ! -r "$root" ] || [ ! -x "$root" ]; then
    printf 'owned_corpus_files: not a readable directory: %s\n' "$root" >&2
    return 2
  fi
  for dir in "${OWNED_CORPUS_SCOPE_DIRS[@]}"; do
    # The symlink test comes FIRST because every test after it follows links: a
    # symlinked scope root would pass all three and then be silently skipped by
    # the physical-mode `find` below. See the header.
    if [ -L "$root/$dir" ]; then
      printf 'owned_corpus_files: scope root is a symlink: %s\n' "$root/$dir" >&2
      return 2
    fi
    if [ ! -d "$root/$dir" ] || [ ! -r "$root/$dir" ] || [ ! -x "$root/$dir" ]; then
      printf 'owned_corpus_files: scope root missing or unreadable: %s\n' "$root/$dir" >&2
      return 2
    fi
  done

  if ! work="$(mktemp -d)"; then
    printf 'owned_corpus_files: cannot create a temporary directory\n' >&2
    return 2
  fi

  # The nested-symlinked-directory sweep the header describes. It runs BEFORE
  # the corpus enumeration below, so a scope root that hides Markdown behind a
  # link never produces a partial answer at all. `-type l` in physical mode
  # reports every symlink at its own location, including one nested many levels
  # down, which is exactly the set the enumeration below cannot see through.
  {
    for dir in "${OWNED_CORPUS_SCOPE_DIRS[@]}"; do
      find "$root/$dir" -type l -print0 || rc=1
    done
  } > "$work/links"
  if [ "$rc" -ne 0 ]; then
    printf 'owned_corpus_files: could not enumerate symlinks under %s\n' "$root" >&2
    rm -rf "$work"
    return 2
  fi

  while IFS= read -r -d '' path; do
    rel="${path#"$root"/}"
    if owned_corpus_excluded "$rel"; then
      continue
    fi
    # A symlink to a FILE is not this rule's subject: it hides nothing, since a
    # link standing where a covered file could stand is already the budget
    # guard's own violation. A DANGLING link is not a directory either, so it
    # falls out here rather than being refused for a shape it does not have.
    [ -d "$path" ] || continue
    # `-L` is what makes this see through the link; a non-zero status is a
    # refusal rather than an empty answer, because an unreadable target and a
    # symlink loop both land here and neither is "there is no Markdown under
    # it". The result is captured rather than counted, so an empty capture is
    # the whole test.
    if ! found="$(find -L "$path" -type f \( -name '*.md' -o -name '*.mdc' \) -print 2>/dev/null)"; then
      printf 'owned_corpus_files: cannot look through the symlinked directory: %s\n' \
        "$path" >&2
      rm -rf "$work"
      return 2
    fi
    if [ -n "$found" ]; then
      printf 'owned_corpus_files: symlinked directory hides Markdown from the corpus: %s\n' \
        "$path" >&2
      rm -rf "$work"
      return 2
    fi
  done < "$work/links"

  # A brace group with a redirection runs in THIS shell, so an `rc` set by a
  # failing find inside it survives — which a pipeline or a process substitution
  # would not, and a find that failed halfway would otherwise read as a smaller
  # corpus rather than as an error.
  {
    find "$root" -maxdepth 1 -type f \( -name '*.md' -o -name '*.mdc' \) -print0 || rc=1
    for dir in "${OWNED_CORPUS_SCOPE_DIRS[@]}"; do
      find "$root/$dir" -type f \( -name '*.md' -o -name '*.mdc' \) -print0 || rc=1
    done
  } > "$work/all"

  if [ "$rc" -ne 0 ]; then
    printf 'owned_corpus_files: could not enumerate the corpus under %s\n' "$root" >&2
    rm -rf "$work"
    return 2
  fi

  while IFS= read -r -d '' path; do
    rel="${path#"$root"/}"
    if owned_corpus_excluded "$rel"; then
      continue
    fi
    printf '%s\0' "$path"
  done < "$work/all"

  rm -rf "$work"
}
