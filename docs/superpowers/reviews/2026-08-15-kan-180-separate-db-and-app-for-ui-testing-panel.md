# Final review panel — kan-180-separate-db-and-app-for-ui-testing

## Roster

**Preset:** `light` (recorded `reviewPanelRoster`). Three required slots, all on `sonnet`
(`models.reviewPanel`): Primary (plan alignment + code quality), Principles (`[LENS]` = Merged),
Code review (low).

**Optional slots — all four triggers fired against the diff, all four declined by the operator.**
Declined is recorded distinctly from a trigger that never fired.

| Slot | Trigger that fired | Disposition |
|------|--------------------|-------------|
| 4 — Security | database-name construction, temp-path handling, a changed config file | declined by operator |
| 5 — Adversarial | >300 changed lines; behaviour changes to tested code; a modified test file | declined by operator |
| 6 — Lens B (simplicity & state) | >200 changed lines; 3+ new modules | declined by operator |
| 6 — Lens C (robustness & ops) | error handling; external integrations; config/env | declined by operator |

Declining optional slots does not move the handoff bar.

## Panel diff size

`scripts/check-panel-diff-size.sh <worktree> 33f77be` measured **1075** changed lines at pass 1,
**under cap**, exit 0. No operator question required. The diff grew to 1935 lines across six fix
rounds.

## Passes

Six panel passes and seven fix rounds. Pass 1 ran the full roster; every later pass was a **Full**
re-run, escalated automatically because each fix altered a guard's behaviour.

**No finding was bounced back to its raising slot on any pass.**

| Pass | Primary | Principles | Code review (low) |
|------|---------|-----------|-------------------|
| 1 | clean | F2 | F1 (Critical), F3, F4 |
| 2 | clean | F1, F2 | clean |
| 3 | F1 | clean | F1 (Critical), F2–F8 |
| 4 | F1, F2 (Critical), F3, F4 | clean | clean (narrower probe) |
| 5 | F1, F2, F3 (guard holes), F4 | F1 | F1, F2, F3 |
| 6 | F1 (Critical) | F1 (Minor, non-blocking) | F1 (Critical), F2, F3 |

### The defect that recurred, and how it was finally closed

The branch's defining defect appeared six times in three shapes. Each early fix closed the reported
instance and left the next one open:

1. **Pass 1** — `ui-test-up` force-dropped `$(UITEST_DB)`, which a make command-line assignment can
   override, reaching the **live database**. Fixed with `override UITEST_DB`.
2. **Pass 3** — the destructive line had since moved to `$(UITEST_ID)`, unprotected. Fixed with
   `override UITEST_ID`.
3. **Pass 4** — six further `UITEST_*` variables were unprotected and reached `rm -rf`, a daemon DSN
   and a `kill`. Fixed **structurally**: all nine variables `override`, a rule-not-list comment, and
   `scripts/check-uitest-overrides.sh` to enforce it mechanically.
4. **Pass 5** — the new guard's own regex had holes (leading whitespace, `?=`, `+=`, `=`, `::=`) and
   was the only `check-*.sh` with no test harness, which is why its holes went unseen. Fixed, and
   `scripts/test-check-uitest-overrides.sh` added.
5. **Pass 6** — three more forms evaded: the safety macro `UITEST_SAFE_RMRF` was itself a plain
   `define` and so overridable, plus `export UITEST_X` and `$(eval UITEST_X := …)`. Fixed in a final
   bounded round.

**What made the difference was moving from fixing named variables to enforcing a rule**, and then
giving that enforcement its own tests. A guard reporting OK while the rule is violated is worse than
no guard, because it licenses the belief that the rule is enforced.

## Findings

Every finding raised across six passes, with its final status. Locations are as recorded by the
raising slot.

| ID | Pass | Slot | Severity | Location | Note |
|---|---|---|---|---|---|
| F1 | 1 | Code review | Critical | `stats/Makefile` | `ui-test-up` force-dropped an overridable `$(UITEST_DB)`, reaching the live database |
| F2 | 1 | Principles | Minor | `stats/Makefile` | `dropdb`/`createdb` duplicated from `scripts/workspace.sh` with no justifying comment |
| F3 | 1 | Code review | Minor | `stats/Makefile` | fixed sleep before rebinding, where the target polls elsewhere |
| F4 | 1 | Code review | Minor | `stats/cmd/uitest-seed/guard_test.go` | two byte-identical DSN subtests |
| F5 | 2 | Principles | Important | `stats/Makefile` | partial reuse still re-hardcoded container and superuser |
| F6 | 2 | Principles | Minor | `stats/Makefile` | runtime guard unreachable behind `override` |
| F7 | 3 | Primary | Important | `stats/Makefile` | the `override` comment's stated mechanism and reproducer were stale |
| F8 | 3 | Code review | Critical | `stats/Makefile` | `UITEST_ID` unprotected; `override` was on the wrong variable |
| F9 | 3 | Code review | Important | `stats/cmd/myflow/state.go`, `stats/README.md` | exported `MYFLOW_ADDR` silently reroutes every later write for the session |
| F10 | 3 | Code review | Minor | `stats/Makefile` | `ui-test-down` dropped without waiting for shutdown |
| F11 | 3 | Code review | Minor | `scripts/workspace.sh` | `remove --force` had no target check of its own |
| F12 | 3 | Code review | Minor | `stats/cmd/uitest-seed/seed_test.go` | the per-test-database helper duplicated for roughly the seventh time |
| F13 | 3 | Code review | Minor | `stats/cmd/uitest-seed/seed_test.go` | a hardcoded DSN literal where `config.DefaultDSN` exists |
| F14 | 3 | Code review | Minor | `stats/cmd/uitest-seed/seed.go` | a duplicated metrics-map literal |
| F15 | 3 | Code review | Minor | `stats/cmd/myflow/state.go` | `resolveDefaultAddr` restating the empty-is-unset policy `config.FromEnv` encodes |
| F16 | 4 | Primary | Critical | `stats/Makefile` | `UITEST_DSN` unprotected — a second daemon could migrate and seed the live database |
| F17 | 4 | Primary | Critical | `stats/Makefile` | four unprotected variables reached `rm -rf` |
| F18 | 4 | Primary | Important | `stats/Makefile` | `UITEST_PIDFILE` unprotected, feeding `kill` |
| F19 | 5 | Primary | Important | `scripts/check-uitest-overrides.sh` | regex anchored at column 0, missing leading-whitespace assignments |
| F20 | 5 | Primary | Important | `scripts/check-uitest-overrides.sh` | regex matched only `:=`, missing `?=`, `+=`, `=`, `::=` |
| F21 | 5 | Primary | Important | `scripts/check-uitest-overrides.sh` | the only `check-*.sh` with no test harness |
| F22 | 5 | Code review | Important | `stats/Makefile` | guard wired only into `test`, bypassed by `ui-test-up` itself |
| F23 | 5 | Primary | Minor | `stats/Makefile` | `UITEST_SAFE_RMRF`'s comment overclaimed what a prefix match guarantees |
| F24 | 5 | Principles | Minor | `scripts/check-uitest-overrides.sh` | verdict line used the refusal idiom, not the sibling verdict convention |
| F25 | 6 | Code review | Critical | `stats/Makefile` | `UITEST_SAFE_RMRF` was a plain `define`, so the safety macro itself was overridable |
| F26 | 6 | Primary | Important | `scripts/check-uitest-overrides.sh` | `export UITEST_X := …` evaded the guard |
| F27 | 6 | Code review | Important | `scripts/check-uitest-overrides.sh` | `$(eval UITEST_X := …)` evaded the guard |
| F28 | 6 | Principles | Minor | `scripts/check-uitest-overrides.sh` | include-following is generality for a case that does not exist yet |

### Withdrawals

Three findings were **withdrawn by the operator**, each with the reason recorded here. No run wrote
a withdrawal for itself.

- **F12** — the per-test-database helper is duplicated roughly seven times repository-wide;
  extracting it is a separate refactor well beyond this change's scope.
- **F15** — `resolveDefaultAddr`'s two-line env-or-default is an idiom already used identically by
  `resolveHarness`; extracting a helper for it is the abstraction this project's simplicity rule
  forbids.
- **F28** — a knowingly-taken, documented tradeoff in a guard with a three-time recurrence history,
  which the principles file treats as a valid answer rather than a defect.

### Scope decision on further evasion classes

After pass 6 the operator decided: **fix the three named forms, then stop.** Any further Make
evasion class is out of scope, on a threat model stated explicitly rather than assumed:

- **No evasion found after pass 4 can reach the database.** `scripts/workspace.sh` refuses any
  `--force` drop of a database not ending `_uitest`, and `stats/cmd/uitest-seed/guard.go` refuses a
  non-`_uitest` DSN. Neither is a Make variable, so neither is reachable by a Makefile override.
- **The residual exposure is `rm -rf` of an operator-named path via a deliberately hostile `make`
  invocation.** That is self-sabotage, not a vulnerability: anyone able to type
  `make ui-test-down 'UITEST_SAFE_RMRF=…'` can type `rm -rf` directly.
- **This is distinct from the pass-1 Critical, which was a plausible accident.**
  `make ui-test-up UITEST_DB=myflow` is something a person might genuinely type believing it
  redirects the target. That class is closed.
- **The guard can never be complete without parsing Make itself**, so "every slot returns clean" is
  not a reachable termination condition for it.

## Verification of the final round

The final bounded round was verified by the dispatcher rather than re-reviewed by the panel, and
that is stated here rather than left implicit. Measured directly against the final tree:

- `make -n ui-test-down 'UITEST_SAFE_RMRF=@echo INJECTED $(1)'` — **0** occurrences of the injected
  text; the safety `case` block is intact.
- `define UITEST_EVIL`, `export UITEST_EVIL := x`, `$(eval UITEST_EVIL := x)` — each exits **1** with
  `file:line`.
- `scripts/test-check-uitest-overrides.sh` — 16/16 cases pass.
- `scripts/check-uitest-overrides.sh` against the real tree — `UITEST-OVERRIDES-OK`, exit 0.

## Recorded for follow-up, not fixed

**GNU Make 3.81 — the `make` on this machine's PATH — does not honour `override` combined with
`export`**, in either order; a command-line reassignment still wins. GNU Make 4.4.1 handles it
correctly. `stats/Makefile` has no `export UITEST_*` line today, so nothing is exploitable, but
`override export` must not be adopted here on the assumption that it works — it has to be tested
against whichever `make` actually runs.

<!-- measured: both Make versions exercised on macOS during the final fix round, 2026-08-15 -->

findings-total: 28
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 withdrawn — repository-wide test-helper extraction is a separate refactor beyond this change's scope
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 withdrawn — a two-line env-or-default idiom already used identically elsewhere does not warrant an abstraction
finding-status: F16 fixed
finding-status: F17 fixed
finding-status: F18 fixed
finding-status: F19 fixed
finding-status: F20 fixed
finding-status: F21 fixed
finding-status: F22 fixed
finding-status: F23 fixed
finding-status: F24 fixed
finding-status: F25 fixed
finding-status: F26 fixed
finding-status: F27 fixed
finding-status: F28 withdrawn — a documented tradeoff in a guard with a three-time recurrence history, which the principles file treats as a valid answer

reproducers-total: 28
finding-reproducer: F1 scripts/check-uitest-overrides.sh
finding-reproducer: F2 none — a duplication finding with no runnable check
finding-reproducer: F3 none — timing-dependent, not reliably reproducible
finding-reproducer: F4 scripts/test-check-uitest-overrides.sh
finding-reproducer: F5 none — a duplication finding with no runnable check
finding-reproducer: F6 none — a reachability observation with no runnable check
finding-reproducer: F7 none — a comment-accuracy finding with no runnable check
finding-reproducer: F8 scripts/check-uitest-overrides.sh
finding-reproducer: F9 none — established by grep over the skills tree, not a runnable check
finding-reproducer: F10 none — timing-dependent, not reliably reproducible
finding-reproducer: F11 scripts/test-workspace.sh
finding-reproducer: F12 none — a duplication finding with no runnable check
finding-reproducer: F13 none — a duplication finding with no runnable check
finding-reproducer: F14 none — a duplication finding with no runnable check
finding-reproducer: F15 none — a duplication finding with no runnable check
finding-reproducer: F16 scripts/check-uitest-overrides.sh
finding-reproducer: F17 scripts/check-uitest-overrides.sh
finding-reproducer: F18 scripts/check-uitest-overrides.sh
finding-reproducer: F19 scripts/test-check-uitest-overrides.sh
finding-reproducer: F20 scripts/test-check-uitest-overrides.sh
finding-reproducer: F21 scripts/test-check-uitest-overrides.sh
finding-reproducer: F22 scripts/check-uitest-overrides.sh
finding-reproducer: F23 none — a comment-accuracy finding with no runnable check
finding-reproducer: F24 scripts/test-check-uitest-overrides.sh
finding-reproducer: F25 scripts/test-check-uitest-overrides.sh
finding-reproducer: F26 scripts/test-check-uitest-overrides.sh
finding-reproducer: F27 scripts/test-check-uitest-overrides.sh
finding-reproducer: F28 none — a judgment call on documented generality, not a mechanical defect
