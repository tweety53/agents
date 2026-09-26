# scripts/lib/flow-guard.sh — flow_guard_exec, how every guard ported to Go
# (KAN-760) runs: its scripts/<name>.sh shim sources this file and hands its
# arguments to flow_guard_exec, which execs `flow-guard <name>` built from the
# very checkout the shim lives in (design.md's
# guard-binary-built-from-checkout).
#
# WHY BUILD FROM THE CHECKOUT. A bash guard was its own source: the skills'
# scripts/ symlink into the checkout, so a guard edit merged to main was in
# force on the next call. A flow-guard installed once to a PATH directory is
# not — it keeps enforcing whatever it was built from until someone rebuilds
# it by hand, and nothing reports the skew, so a check added since silently
# does not run (panel round 0, F1/F8). An install step also has no place to
# run where a merge lands, so a machine that never ran it has guards that
# cannot answer at all (F2). Deriving the binary at the point of use closes
# both: there is no installed copy to go stale and no install step to miss.
#
# THE KEY. The cache key is the SHA-256 of the per-file SHA-256 lines of
# exactly the sources the binary is built from — stats/go.mod, stats/go.sum,
# and the non-test .go files of stats/cmd/flow-guard/ and
# stats/internal/guard/ (the only package it imports from this module) —
# named relative to stats/, so two checkouts with the same sources share one
# binary and a _test.go edit never forces a rebuild. The binary lives at
# <cache>/<key>/flow-guard, where <cache> is FLOW_GUARD_CACHE_DIR if set (the
# test suites point it at a temporary directory, so no test writes into the
# operator's real cache), else ${XDG_CACHE_HOME:-$HOME/.cache}/flow-guard.
# Old keys are never pruned: each is one binary of a few MB, and removing the
# directory is always safe.
#
# THE BUILD. A missing key is built with `go build` to a pid-named temporary
# file in the key's directory and moved into place with `mv -f`, a rename
# within one directory — atomic, so concurrent first calls each build and the
# last rename wins with an identical binary, and no caller ever execs a
# half-written file. The key is recomputed after the build; if the sources
# changed while it ran, the binary is discarded rather than cached under a key
# it was not built from. The build runs with GOOS, GOARCH and GOFLAGS removed
# from its environment, so it always targets the host and the artifact
# depends only on the hashed sources: a call from a cross-compiling shell
# would otherwise cache a foreign binary every later call execs (exit 126).
# A failed rename exits the cannot-answer code like any other build failure.
#
# CANNOT ANSWER. Anything that stops the binary from being derived — no
# SHA-256 tool, no cache location, no stats/ beside this file's checkout, no
# `go` on PATH, a failed build — prints "<opening> <cause>" to stderr and exits
# the calling guard's own cannot-answer code: 4 for run-reproducer (whose 2 is
# "refused"), 2 for the other four. Never a verdict. `go build`'s own output
# goes to stderr too, so a guard's stdout carries only its verdict lines.
#
# Bash 3.2 is the floor (macOS /bin/bash). Not meant to be executed directly —
# a shim sources it; it sets no shell options of its own.

# flow_guard_key <stats-dir> — print the cache key of the flow-guard sources
# under <stats-dir>, or return non-zero if they cannot be hashed.
flow_guard_key() {
  (
    cd "$1" || exit 1
    files=(go.mod go.sum)
    for f in cmd/flow-guard/*.go internal/guard/*.go; do
      case "$f" in *_test.go) ;; *) files+=("$f") ;; esac
    done
    if command -v sha256sum >/dev/null 2>&1; then
      sum() { sha256sum "$@"; }
    elif command -v shasum >/dev/null 2>&1; then
      sum() { shasum -a 256 "$@"; }
    else
      exit 1
    fi
    set -o pipefail
    key="$(sum "${files[@]}" | sum)" || exit 1
    printf '%s\n' "${key%% *}"
  )
}

# flow_guard_exec <name> <cannot-answer-code> <opening> [guard arguments...]
# — exec the checkout-built flow-guard as guard <name>, or print
# "<opening> <cause>" to stderr and exit <cannot-answer-code>.
flow_guard_exec() {
  local name="$1" code="$2" opening="$3" stats cache key bin tmp
  shift 3
  stats="$(cd -P "$(dirname -- "${BASH_SOURCE[0]}")" && cd ../../stats 2>/dev/null && pwd)" || {
    echo "$opening cannot build flow-guard — no stats/ source tree beside ${BASH_SOURCE[0]}" >&2
    exit "$code"
  }
  cache="${FLOW_GUARD_CACHE_DIR:-${XDG_CACHE_HOME:-${HOME:+$HOME/.cache}}}"
  if [ -z "$cache" ]; then
    echo "$opening cannot build flow-guard — no cache location (set FLOW_GUARD_CACHE_DIR, XDG_CACHE_HOME or HOME)" >&2
    exit "$code"
  fi
  [ -n "${FLOW_GUARD_CACHE_DIR:-}" ] || cache="$cache/flow-guard"
  key="$(flow_guard_key "$stats")" || {
    echo "$opening cannot build flow-guard — could not hash its sources in $stats (needs sha256sum or shasum)" >&2
    exit "$code"
  }
  bin="$cache/$key/flow-guard"
  if [ ! -x "$bin" ]; then
    if ! command -v go >/dev/null 2>&1; then
      echo "$opening cannot build flow-guard from $stats — no go on PATH" >&2
      exit "$code"
    fi
    tmp="$bin.$$"
    if ! mkdir -p "$cache/$key" || ! (cd "$stats" && env -u GOOS -u GOARCH -u GOFLAGS go build -o "$tmp" ./cmd/flow-guard) >&2; then
      rm -f "$tmp"
      echo "$opening could not build flow-guard from $stats" >&2
      exit "$code"
    fi
    if [ "$(flow_guard_key "$stats")" != "$key" ]; then
      rm -f "$tmp"
      echo "$opening flow-guard's sources in $stats changed during its build — run the guard again" >&2
      exit "$code"
    fi
    if ! mv -f "$tmp" "$bin"; then
      rm -f "$tmp" || :
      echo "$opening could not install flow-guard into $cache/$key" >&2
      exit "$code"
    fi
  fi
  exec "$bin" "$name" "$@"
}
