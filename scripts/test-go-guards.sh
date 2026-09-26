#!/usr/bin/env bash
# test-go-guards.sh — the companion harness for every guard ported to Go
# (kan-760). A ported scripts/<name>.sh is a shim that execs
# `flow-guard <name>`; its cases live in
# stats/internal/guard/<name, - replaced by _>_test.go, and run-guard-tests.sh
# accepts that file as the guard's companion in place of a
# test-<name>.sh. This harness runs the whole package once, so the suite
# runs every ported guard's cases as one harness.
#
# Exit codes are `go test`'s: 0 every case passed, non-zero otherwise.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../stats" && go test ./internal/guard/... -count=1
