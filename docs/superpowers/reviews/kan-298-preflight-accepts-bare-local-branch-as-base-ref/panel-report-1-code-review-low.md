Working tree clean (only the expected untracked `spectre/changes/...` directory). Both suites pass consistently (rc=0, "all cases pass") across repeated runs. The single earlier failed run was a one-off shell artifact from my own investigation (a stray `(eval)` error appeared in that same batched command), not reproducible in five follow-up isolated runs — not attributable to the diff.

**Verdict: PASS**

No high-confidence defects found in the delta. Both new test cases (9b in `test-check-finish-preflight.sh`, 11b in `test-check-base-moved.sh`):
- Correctly bracket `set +e`/`set -e` around the command substitution that returns exit 2.
- `EXPECTED_USAGE` matches the guards' actual heredoc byte-for-byte (verified directly against `scripts/check-finish-preflight.sh:54-60` and `scripts/check-base-moved.sh:50-56`).
- Trailing-newline stripping by `$()` is symmetric between `cat "$MISSING_ARGS_ERR"` and the literal comparison string, so it cannot mask an appended second line — confirmed the guards emit `exit 2` immediately after the heredoc, no fallthrough.
- `MISSING_ARGS_ERR` (a `mktemp` file) is pushed onto `REPOS` and cleaned by the existing `rm -rf` trap loop, same pattern already used for `NOT_A_REPO` (a directory) elsewhere in the file.
- No variable-name collisions: `MISSING_ARGS_OUT`, `MISSING_ARGS_ERR`, `ACTUAL_USAGE`, `EXPECTED_USAGE` are each unique to their block.
- Correctly placed immediately after the existing case 9 / case 11, before the next numbered case; `OUT`/`RC` are reassigned by `run_guard` in every subsequent case, so nothing leaks forward.
