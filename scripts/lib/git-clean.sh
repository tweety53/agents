# scripts/lib/git-clean.sh — git_clean, defined once.
#
# Sourced by scripts/check-visual-verification.sh, which used to carry this
# function inline. It was written once there and never copied elsewhere in
# this repository until this file existed; this is its only current caller,
# extracted anyway rather than left in place so the next guard that needs
# ambient-`GIT_*` neutralisation sources it instead of re-deriving it —
# `scripts/lib/resolve-file.sh`'s and `scripts/lib/within-root.sh`'s own
# headers record the same shape of drift for a function copied only after a
# second caller needed it.
#
# "SAFELY REACH IT" IS THE OPERATIVE PHRASE — see
# scripts/lib/resolve-file.sh's header for the criterion this file follows
# too: a guard that ships through the skills/*/scripts/ symlink farm can
# assume a sibling `lib/` travels with it; a guard reached only by
# hand-copying a single file into an unrelated project's own tooling cannot.
# check-visual-verification.sh ships through the farm — it carries its own
# `lib` symlink into scripts/lib/ beside it — so it sources this file rather
# than carrying its own copy.
#
# Not meant to be executed directly — a caller sources it and calls
# git_clean; it sets no `set -euo pipefail` of its own and relies on the
# sourcing script's.

# git_clean <git args...> — run git with every AMBIENT `GIT_*` environment
# variable neutralized, so a caller's own repo-selection decisions (which
# checkout `-C` points at, and what its `origin` resolves to) cannot be
# overridden by whatever invoked it. `GIT_DIR` is the demonstrated vector:
# git honours it over `-C`'s repo selection, so `GIT_DIR=/anywhere/.git`
# makes both a `rev-parse --git-dir` and a `remote get-url origin` call
# answer about that repository instead of the one `-C` named — including
# reporting a `regression checkout` that is not a git repository at all as
# one. Git hooks set `GIT_DIR` automatically, so a caller invoked from one
# runs pre-bypassed without this. `GIT_WORK_TREE`, `GIT_COMMON_DIR`,
# `GIT_OBJECT_DIRECTORY`, `GIT_ALTERNATE_OBJECT_DIRECTORIES`,
# `GIT_INDEX_FILE`, `GIT_CEILING_DIRECTORIES` and every `GIT_CONFIG*`
# variable (including the numbered `GIT_CONFIG_KEY_<n>`/`GIT_CONFIG_VALUE_<n>`
# pairs) redirect repo or config resolution the same way, so ALL currently-set
# `GIT_`-prefixed variables are cleared here — one mechanism applied to every
# git call a caller makes, rather than a patch aimed at the one named vector.
# Only variables actually present in the environment are unset (`env -u` on
# an absent name is a harmless no-op), so this reads the environment rather
# than guessing its shape from a fixed list that the next git release could
# add to.
git_clean() {
  local unset_args=() var
  while IFS= read -r var; do
    [ -n "$var" ] && unset_args+=(-u "$var")
  done < <(env | LC_ALL=C awk -F= '/^GIT_/ { print $1 }')
  env "${unset_args[@]}" git "$@"
}
