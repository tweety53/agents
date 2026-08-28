# scripts/lib/resolve-remote-base.sh — resolve_remote_base, defined once.
#
# KAN-88 half 1 hardening (design.md: preflight-resolves-remote-tracking).
# check-base-moved.sh (task 2) needs the same resolution
# check-finish-preflight.sh applies here, so it is a shared library function
# rather than two copies that could drift the way `git_clean` and
# `within_root` record their own callers once did before extraction.
#
# THE LOOKUP IS A PREFERENCE, NOT A REWRITE (design.md:
# remote-lookup-is-a-preference-not-a-rewrite). The correct caller already
# composes `origin/$BASE` itself; feeding that back in must be a no-op.
# `origin/main` looks up `refs/remotes/origin/origin/main`, finds nothing,
# and passes through unchanged — so the contract's own call site, and every
# existing test built against a repository with no `origin` remote, keep
# their current behaviour.
#
# Not meant to be executed directly — a caller sources it and calls
# resolve_remote_base; it sets no `set -euo pipefail` of its own and relies
# on the sourcing script's.

# resolve_remote_base <worktree> <base-ref> — print to stdout
# `origin/<base-ref>` when `refs/remotes/origin/<base-ref>` resolves to a
# commit in <worktree>, and <base-ref> unchanged otherwise.
#
# Never fails the caller: an unreadable worktree, or a ref that does not
# resolve either way, still prints a value. The caller's own next step
# already refuses a ref that does not resolve, by name — this function only
# ever narrows which name that check tests.
#
# `--end-of-options` on the `rev-parse` so a ref beginning with `-` is read
# as a ref rather than parsed as a git option, matching every other ref
# resolution in this repository's guards.
resolve_remote_base() {
  local worktree="$1" base_ref="$2"
  if git -C "$worktree" rev-parse --verify --end-of-options \
    "refs/remotes/origin/${base_ref}^{commit}" >/dev/null 2>&1; then
    printf '%s\n' "origin/${base_ref}"
  else
    printf '%s\n' "$base_ref"
  fi
}
