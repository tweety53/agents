#!/usr/bin/env bash
# test-setup.sh — regression harness for setup.sh.
#
# The harness is the Go test package stats/internal/setuptest/: its assertion
# groups run concurrently, each against its own sandboxed HOME under
# /tmp/flow-test-setup.*, and the real ~/.claude, ~/.zcode and the source tree
# are fingerprinted around all of them. Its package doc says what a green run
# does and does not prove. This shim keeps the one name run-guard-tests.sh's
# test-*.sh glob and every caller use.
#
# Usage: scripts/test-setup.sh
# Set KEEP_SANDBOX=1 to leave the sandbox behind for inspection.
# Exit codes are `go test`'s: 0 every assertion passed, non-zero otherwise.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../stats" && exec go test ./internal/setuptest/ -count=1
