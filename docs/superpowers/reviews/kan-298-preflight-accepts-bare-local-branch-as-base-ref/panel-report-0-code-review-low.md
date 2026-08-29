Everything checks out: heredoc terminates correctly under `set -euo pipefail` (quoted `'EOF'` prevents expansion of `<`/`/` chars; none present anyway), stdout stays empty, exit code stays 2, both test harness assertions are reachable and pass, `grep -qF --` handles the pattern correctly, and both `$SCRIPT_DIR`s resolve to `scripts/` where the respective guard lives. Full test suites pass.

No high-confidence defects found.

PASS
