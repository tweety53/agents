Confirmed consistent with existing pattern (line 247 is the prior 8b case, 286 is new 8c). Both test suites pass in full, isolation/cleanup pattern matches existing conventions, and both shim reproducers target only the intended git subcommand (verified against `check-finish-preflight.sh` and `resolve-remote-base.sh` for stray `merge-base`/`rev-list` calls — none found).

No defects found.

VERDICT: clean
