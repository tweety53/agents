# Known bugs

- `scripts/test-check-cleanup-complete.sh` — the fixture survivors commands' timing bounds (the
  5-second inside-bound cases and the 10-second escaper-race case) lose their races when the
  machine is under heavy external load — observed at load average ~24 on 10 cores, where
  `./survivors.sh` spawns slower than the bound and the group kill can beat a fork — failing 3-4
  timing cases that pass on an idle machine; the guard under test and every assertion are
  correct, the bounds are simply not load-proof (KAN-376 raised one bound for the same reason) —
  introduced by 18feb597 (refactor(flow): work only in worktrees, back the branch up remotely,
  drop the myflow legacy), an ancestor of main.

## Deferred review findings

- `stats/internal/guard/panelreproducers.go:140` — F6, Minor, kan-778-agents-port-the-next-five-slowest-guards-to-the — four hand-written refusals
  are untested: a null state record, a finding with `status: null`, and `FLOW_GUARD_REPO_ROOT`
  unset in check-references and check-installed-citations — breaks: a regression goes unseen, and
  losing the null-state check fails open (`REPRODUCERS-OK`, exit 0) — fix: one subtest per
  refusal — deferred: coverage-gap.
- `.flow/project.md:17` — F7, Minor, kan-778-agents-port-the-next-five-slowest-guards-to-the — lists `check-task-commit-fields.py` as a running
  Python guard; the guard is a `flow_guard_exec` shim, the `.py` only a module
  `check-plan-shape.py` imports — breaks: misleads a reader about what runs — fix: reword the
  line — deferred: doc-only.
- `skills/flow/SKILL.md:156` — F8, Minor, kan-778-agents-port-the-next-five-slowest-guards-to-the — names 2 of the 7 `flow-guard` shims in
  `skills/flow/scripts/` as needing `lib/flow-guard.sh` — breaks: an installer following it
  misses five sibling dependencies — fix: list all seven — deferred: doc-only.
- `scripts/reproducer-metachars.sh:25` — F9, Minor, kan-778-agents-port-the-next-five-slowest-guards-to-the — this change removed the last bash
  consumer of `reproducer-metachars.sh` and `scripts/lib/change-plan.sh`, which remain as a
  second copy of logic now in Go — breaks: two copies drift — fix: delete both, or record why they
  stay — deferred: out-of-scope.
- `stats/internal/guard/guard.go:17` — F10, Minor, kan-778-agents-port-the-next-five-slowest-guards-to-the — `Env.LookupEnv` sits beside
  `Getenv` instead of one deriving the other — breaks: an `Env` built with only `Getenv`
  panics in the three guards calling `LookupEnv` — fix: derive `Getenv` from `LookupEnv` —
  deferred: other.
- `stats/internal/guard/check_panel_reproducers_test.go:160` — F11, Minor, kan-778-agents-port-the-next-five-slowest-guards-to-the — case 20 is named
  "cannot-answer, not violations-found" but asserts exit 1 — breaks: misleads a reader of the
  test — fix: rename the case — deferred: cosmetic.
- `stats/internal/guard/installedcitations.go:347` — F12, Minor, kan-778-agents-port-the-next-five-slowest-guards-to-the — TMPDIR comes from the
  injected `Env` but `setup.sh` gets `os.Environ()` — breaks: a test injecting an
  environment does not reach `setup.sh` — fix: build `setup.sh`'s environment from `Env` —
  deferred: other.
- `skills/flow-contracts/jira-followups.md:17` — F13, Minor, kan-778-agents-port-the-next-five-slowest-guards-to-the
  — still frames integrate run 1 as one of several filing sites ("every site that files a
  follow-up", the heading `### The filing site's outstanding items`, "that site's list") —
  breaks: a reader looks for a second filing site that no longer exists — fix: reduce the three
  passages to the integrate run's outstanding items — deferred: doc-only.
