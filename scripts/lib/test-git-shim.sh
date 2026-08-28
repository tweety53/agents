# scripts/lib/test-git-shim.sh — shim_failing_git, shared by the
# finish-preflight and base-moved test harnesses (test-check-finish-preflight.sh,
# test-check-base-moved.sh) to build a `git` that fails one invocation and
# passes every other one through to the real git. KAN-88 fix round 2:
# extracted after this exact block was copied three times (fix round 1's
# panel-report-1-principles.md, case 8c vs. case 8b vs.
# test-check-base-moved.sh's case 9b).
#
# TEST-ONLY, DELIBERATELY IN scripts/lib/ (KAN-88 fix round 3, F10). Every
# other file in this directory is production logic with a `check-*.sh`
# consumer and `test-*.sh` files as secondary readers of that same logic;
# this file inverts that — its only two consumers are
# test-check-finish-preflight.sh and test-check-base-moved.sh, both test
# harnesses. Considered moving it to a directory name that says so
# (`scripts/test-lib/`), but nothing else in this repository enumerates
# `scripts/lib/`'s membership or its file count — not check-guard-symlinks.sh
# (which validates skills/*/scripts/ symlink targets, never scripts/lib's own
# contents), not owned_corpus_files (scripts/lib/owned-corpus.sh, which scopes
# `.md`/`.mdc` files only), not check-installed-citations.py's installed-set
# classifier, not .flow/project.md's test-file list (which names
# `scripts/test-*.sh`, never anything under scripts/lib/) — so a new directory
# would be a bigger, unforced move for a Minor finding: a fresh path nothing
# currently governs, next to an existing one everything already does. Staying
# in scripts/lib/ and making the file's status impossible to miss — this
# paragraph, the filename's own `test-` prefix already matching this
# repository's scripts/test-*.sh convention for test code, and the
# "Not meant to be executed directly" line below — is the lower-risk fix for
# the same finding.
#
# Not meant to be executed directly — a caller sources it, has already
# defined an indexed REPOS array for sandbox cleanup (bash 3.2 has no
# associative arrays), and calls shim_failing_git.

# The real git, resolved once — on the first source of this file — before
# any caller has built a shim or touched PATH. KAN-88 fix round 3, F14's
# hardening: a shim built with `real_git="$(command -v git)"` evaluated
# per-call would resolve to whatever shim directory a caller had already
# prefixed onto PATH, chaining shim-inside-shim instead of shim-inside-real —
# a caller mistake this library should make impossible rather than merely
# document. Every caller sources this file before it builds its first shim
# (see shim_failing_git's own header), so the FIRST source always runs while
# PATH still names only the real git. KAN-88 fix round 4, F18: the capture
# itself is guarded to run only once — re-sourcing this file (directly, or
# via a second caller in the same shell) while a shim already sits on PATH
# must not overwrite the value the first, correct source captured. See
# test-lib-test-git-shim.sh for the direct proof.
: "${TEST_GIT_SHIM_REAL_GIT:=$(command -v git)}"

# shim_failing_git <prefix> <match-arg> <message> — creates a
# `mktemp -d`-backed directory containing an executable `git` that, when
# invoked with <match-arg> anywhere in its argument list, prints
# "fatal: <message>" to stderr, records that it fired (see assert_shim_fired
# below), and exits 128; every other invocation execs the real git resolved
# when this library was sourced. <match-arg> is usually a subcommand name
# ("status", "merge-base") but MUST be a more specific argument — e.g. a
# literal ref expression like "HEAD^{commit}" — whenever the guard under test
# calls that subcommand more than once, so the shim fails only the one call
# under test. The match is exact equality against a single argv element
# (KAN-88 fix round 3, F14): an argument that merely CONTAINS <match-arg> as
# a substring must not fire the shim, which is what keeps a match-arg like
# "status" from also catching an unrelated argument such as
# "some/status/path". See test-check-finish-preflight.sh's exact-match case
# for the direct proof. Appends the directory to REPOS so the caller's EXIT
# trap removes it, and sets SHIM_DIR to that directory so the caller can
# prefix PATH with it (PATH="$SHIM_DIR:$PATH").
shim_failing_git() {
  local prefix="$1" match_arg="$2" message="$3"
  SHIM_DIR="$(mktemp -d "${TMPDIR:-/tmp}/${prefix}.XXXXXX")"
  REPOS+=("$SHIM_DIR")
  {
    printf '#!/usr/bin/env bash\n'
    printf 'for a in "$@"; do\n'
    printf '  if [ "$a" = "%s" ]; then\n' "$match_arg"
    printf '    : > "%s/.fired"\n' "$SHIM_DIR"
    printf '    echo "fatal: %s" >&2\n' "$message"
    printf '    exit 128\n'
    printf '  fi\n'
    printf 'done\n'
    printf 'exec %s "$@"\n' "$TEST_GIT_SHIM_REAL_GIT"
  } > "$SHIM_DIR/git"
  chmod +x "$SHIM_DIR/git"
}

# assert_shim_fired <shim-dir> <label> — asserts that the shim built at
# <shim-dir> (by shim_failing_git, or by a bespoke inline shim following the
# same "touch .fired on the matching branch" convention — see
# test-check-base-moved.sh's cases 9c and 9f) actually intercepted a call
# during the run just made. KAN-88 fix round 3 hardening: without this, a
# shim whose match-arg no longer matches — a typo, or a guard whose argument
# shape changed after the case was written — lets every call fall through to
# the real git, and the guard then returns whatever verdict the untouched
# repository produces. That reads exactly like the guard behaving correctly;
# the case has silently stopped testing the failure path it names. Calls
# `pass`/`fail`, defined by the sourcing test file rather than a private copy
# here, so a firing failure counts toward that file's own FAILURES total the
# same way every other assertion in it does.
assert_shim_fired() {
  local shim_dir="$1" label="$2"
  if [ -f "$shim_dir/.fired" ]; then
    pass "$label: shim actually intercepted the call"
  else
    fail "$label: shim never intercepted the call — this case tested nothing"
  fi
}
