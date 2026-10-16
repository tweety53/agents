# pipeline-fix-kan-954-user-row-loaded-twice-per-request-13

## Why

`TestMutateAndVerify/project_command_timeout_counts_as_killed`
(`stats/internal/guard/mutate_and_verify_test.go`) failed under load during the gymie `/flow` run
of `kan-954-user-row-loaded-twice-per-request`: rc=4, "harness ./project-cmd.sh timed out after
3s on the baseline pass". It and `mutated pass timeout counts as killed` set
`MUTATE_AND_VERIFY_TIMEOUT_SECONDS=3` and rely on the trivial baseline run finishing inside the
same 3s wall-clock bound the mutated run must hit. Host load decides whether it does.

## What changes

- `Env.MutateExpire` (`stats/internal/guard/guard.go`): when non-nil, it replaces
  mutate-and-verify's per-run timer, as `Env.SurvivorsExpire` already does for the survivors
  bound. Production leaves it nil, so `MUTATE_AND_VERIFY_TIMEOUT_SECONDS` behaves as before.
- The two "timeout counts as killed" cases fire the bound on readiness — a marker the harness
  touches only on the mutated pass — never on the wall clock, so the baseline can never be the
  run that times out.
