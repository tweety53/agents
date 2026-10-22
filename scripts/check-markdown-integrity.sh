#!/usr/bin/env bash
# check-markdown-integrity.sh — fail when this repository's own skill and
# rule Markdown carries structural damage a rendering-agnostic reader would
# mis-read: a code span broken by a misunderstood backslash escape, a
# blockquote whose first line lost its `>` marker, a paragraph torn apart
# mid-sentence, or a paragraph that promises content with a trailing colon
# and then supplies none.
#
# Usage: scripts/check-markdown-integrity.sh [project-root]
#
# Exit codes:
#   0  clean — every scanned file is structurally sound.
#   1  violations found — one or more of the four signals fired. Each is
#      reported on stdout as `file:line: <signal>: <message>`. Fix the named
#      line; there is no suppression marker and no mechanism for one.
#   2  cannot answer — the project root does not exist, an argument this
#      guard does not accept was given, or a file or the rules/ directory
#      inside the scanned scope exists but cannot be read (permission
#      denied, not valid UTF-8).
#
# Scope: `skills/**/*.md` and `rules/*.mdc` under the given project root
# (default: the repository this script lives in) — the files every `/flow*`
# run loads, and where torn text changes an agent's behaviour rather than
# merely reading badly. Nothing else is scanned: not `docs/`, not the
# repository root's own Markdown.
#
# The block model, the four signals and why signal 4 is structural rather
# than phrase-based are the mechanism's doc comment,
# stats/internal/guard/markdownintegrity.go; its table test
# (stats/internal/guard/check_markdown_integrity_test.go) stands in the
# test-check-*.sh companion's place.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-markdown-integrity: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-markdown-integrity 2 "check-markdown-integrity:" "$@"
