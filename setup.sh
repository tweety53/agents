#!/usr/bin/env bash
# setup.sh — Install agents-data into a project for a specific harness
#
# Usage: ./setup.sh <harness> [project-dir] [--register-hooks]
#
# Harnesses: claude-code | global
#
# Examples:
#   ./setup.sh claude-code                    # current directory
#   ./setup.sh claude-code /path/to/other-project
#   ./setup.sh global                         # user-level install (~/.claude)
#   ./setup.sh global --register-hooks        # …and merge the hook registrations into ~/.claude/settings.json
#
# The installer is `flow setup` (stats/cmd/flow/setup.go). This file is only its bootstrap: it
# builds `flow` from this checkout into a fresh temporary directory — never the installed CLI,
# which `make -C stats restart` alone writes — and execs `flow setup`, which removes that directory
# first thing. Nothing is read from PATH but `go`, so a machine with no `flow` yet installs too.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
command -v go >/dev/null 2>&1 || { echo "ERROR: no go on PATH — setup.sh builds its installer from $SCRIPT_DIR/stats" >&2; exit 1; }
build_dir="$(mktemp -d "${TMPDIR:-/tmp}/flow-setup.XXXXXX")"
(cd "$SCRIPT_DIR/stats" && env -u GOOS -u GOARCH -u GOFLAGS go build -o "$build_dir/flow" ./cmd/flow) || {
  rm -rf "$build_dir"
  echo "ERROR: could not build the installer from $SCRIPT_DIR/stats" >&2
  exit 1
}
FLOW_SETUP_REPO_ROOT="$SCRIPT_DIR" FLOW_SETUP_ARGV0="$0" exec "$build_dir/flow" setup "$@"
