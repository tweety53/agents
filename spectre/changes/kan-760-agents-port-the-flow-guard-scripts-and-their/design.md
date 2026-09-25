## Context

- Scope is the first slice of KAN-760: the five guards the issue names. Porting them alone cannot
  reach "seconds" — the next-slowest harness (`test-check-panel-reproducers.sh`, 59s under load)
  becomes the ceiling — so the rest are follow-up slices (see **Follow-ups**).
  <!-- measured: per-harness lines of /usr/bin/time -p scripts/run-guard-tests.sh @ 0747740 -->
- Skills resolve a guard as `<skill-dir>/scripts/<name>.sh`, a relative symlink into this
  repository's `scripts/` (**Guard resolution**, `skills/flow-contracts/pipeline.md`).
- Other bash scripts call these guards as `"$SCRIPT_DIR/<name>.sh"` — `run-reproducer.sh` alone is
  called by `mutate-and-verify.sh`, `prove-reproducer.sh`, `break-and-prove.sh`,
  `check-panel-reproducers.sh`, `check-mutation-reproducer-pin.sh`.
- `check-task-commit-fields.py` is loaded as a module by `check-task-records.py` and
  `check-plan-shape.py`, so it stays until a later slice ports those.

## Architecture

- `stats/cmd/flow-guard/main.go` — `flow-guard <name> [args…]`; a table maps `<name>` to its
  `Run`; an unknown name is exit 2 with one stderr line.
- `stats/internal/guard/` — one file per guard, each exposing
  `Run(args []string, env Env, stdout, stderr io.Writer) int`. `Env` carries the process
  environment lookup, the working directory, and the injectable seams: deadlines as
  `time.Duration`, the `flow record findings` source, the reproducer runner.
- Helpers the bash guards `source` (`reproducer-metachars.sh`, `lib/spec-root.sh`,
  `lib/sha256-hex.sh`, `lib/resolve-file.sh`, `lib/change-plan.sh`) are reimplemented in
  `internal/guard` only as far as the five guards use them.
- Subprocesses: `git` stays an `exec.Command` (no new dependency); reproducers run under
  `exec.CommandContext` in their own process group, the deadline and grace enforced by the
  context and a timer — no poll loop.

## Measurements

**Before** — `0747740`, this machine (10 cores), 2026-09-25:

| run | real | user | sys |
|---|---|---|---|
| full suite, 86 harnesses | 120.7s | 191.3s | 249.3s |

<!-- measured: /usr/bin/time -p scripts/run-guard-tests.sh @ 0747740 -->

| harness | under suite load | standalone real | standalone user | standalone sys |
|---|---|---|---|---|
| test-check-panel-reproducer-exit-contract.sh | 102s | 24.2s | 1.5s | 2.0s |
| test-check-cleanup-complete.sh | 84s | 45.6s | 5.7s | 6.6s |
| test-check-task-commit-fields.sh | 63s | 44.2s | 16.7s | 15.0s |
| test-run-reproducer.sh | 57s | 31.9s | 2.5s | 2.3s |
| test-gather-dispatch-context.sh | 32s | 17.8s | 3.6s | 6.4s |

<!-- measured: suite per-harness lines, then /usr/bin/time -p scripts/test-<name>.sh one at a time @ 0747740 -->

**After** — filled by the live-verification task, same machine, same commands plus
`go test ./internal/guard/...`.

## Decisions

### Scope: the five named guards, rest as follow-up slices

**ID:** scope-five-named-guards
**Status:** active
**Chosen:** port `check-cleanup-complete`, `check-panel-reproducer-exit-contract`,
`check-task-commit-fields`, `run-reproducer`, `gather-dispatch-context`; file follow-up KAN
slices for the rest at integrate — the operator's choice.
**Considered:** porting until suite wall < 10s in one change — ~15–20 guards and >20k lines, too
large for one reviewable change; five with no follow-ups — leaves the issue's "seconds" goal with
no owner.

### Invocation: a separate flow-guard binary behind basename shims

**ID:** separate-flow-guard-binary
**Status:** superseded by flow-guard-binary-rationale-in-go
**Chosen:** a new `stats/cmd/flow-guard` binary; each ported `scripts/<name>.sh` keeps its header
comment verbatim (contracts cite gate guards' headers as canonical) and its body becomes
`command -v flow-guard` check + `exec flow-guard <name> "$@"` — the operator's choice.
**Considered:** a `flow guard <name>` subcommand of the existing CLI — one binary fewer, but it
couples guard releases to the state CLI; the operator chose the separate binary.
**Superseded because:** "the body becomes the shim" deleted every comment below each header — the
reasoning for each branch, one comment that called itself canonical (`check-cleanup-complete.sh`'s
Protection 1) and ~50 citations to them across `scripts/`, `scripts/lib/` and
`skills/flow-contracts/`, which `check-references.sh` does not see; `check-mutation-reproducer-pin.sh`
(in `## lint`) went red reading a line the shim removed. Kept headers also went on describing
plumbing the ports dropped (the python3 wrapper, `dispatch_python_guard`).

### The flow-guard binary; each guard's rationale moves into its Go file

**ID:** flow-guard-binary-rationale-in-go
**Status:** active
**Chosen:** the separate `stats/cmd/flow-guard` binary stands. Each ported `scripts/<name>.sh`
keeps its header (contracts cite gate guards' headers as canonical), corrected only where the port
made a statement false, and its body is the shim. The old body's comments move into the guard's
Go file beside the code they explain, updated where the mechanism changed; every citation of a
body comment or a deleted harness case is repointed to the Go file or test; guards that read a
shimmed script's body read the Go source instead — the operator's choice.
**Considered:** restoring the bash body comments verbatim into each shim — every citation resolves
unchanged, but the comments would describe bash code that no longer exists beside them; fixing
lint only and filing a follow-up — leaves ~50 stale citations on main.

### A missing flow-guard is exit 2, never a fallback

**ID:** missing-binary-exits-2
**Status:** superseded by missing-binary-cannot-answer-code
**Chosen:** a shim finding no `flow-guard` on PATH prints
`<name>: flow-guard not on PATH — run make -C <agents repo>/stats install-guard` to stderr and
exits 2, every guard's existing "cannot answer" code.
**Considered:** keeping the bash body as a fallback — two implementations to keep in parity, which
the port exists to remove.

**Superseded because:** run-reproducer's own contract gives 2 as "refused" and 4 as "cannot
answer", so a flat exit 2 made its callers (`prove-reproducer.sh`, `mutate-and-verify.sh`) read a
missing binary as a refused reproducer.

### A missing flow-guard exits the guard's own cannot-answer code

**ID:** missing-binary-cannot-answer-code
**Status:** active
**Chosen:** a shim finding no `flow-guard` on PATH prints
`<name>: flow-guard not on PATH — run make -C <agents repo>/stats install-guard` to stderr and
exits that guard's existing "cannot answer" code: 4 for `run-reproducer`, 2 for the other four.
**Considered:** exit 2 everywhere (missing-binary-exits-2) — reads as "refused" to run-reproducer's
callers.

### Installed by make and setup.sh global

**ID:** install-via-make-and-setup
**Status:** active
**Chosen:** `make build` writes `bin/flow-guard`; a new `make install-guard` builds it to
`~/.local/bin/flow-guard` without touching the dev daemon; `restart` depends on `install-guard`;
`setup.sh global` builds it too — the operator's choice.
**Considered:** make only — a fresh `setup.sh global` would install shims whose binary is absent.
Installing via `restart` alone — rejected because `restart` stops flowd, which no agent may do.

### Go tests replace the bash harnesses

**ID:** go-tests-replace-harnesses
**Status:** active
**Chosen:** every case of the five harnesses ported to `stats/internal/guard/<name>_test.go`,
`t.Parallel()`, git fixtures built once in `TestMain` and copied per case; the five
`scripts/test-<name>.sh` deleted; a new `scripts/test-go-guards.sh` runs
`go test ./internal/guard/...` as one harness; `run-guard-tests.sh`'s companion-presence rule
accepts `stats/internal/guard/<name>_test.go` for a check guard whose script is a shim — the
operator's choice.
**Considered:** keeping the bash harnesses as a parity check — keeps the slow suite this change
exists to remove.

### Parity is case count plus byte-identical verdicts

**ID:** parity-by-case-count
**Status:** active
**Chosen:** each Go test file carries at least as many cases as its harness's `ok:` lines on a
green run at `0747740`, each asserting exit code and the verdict/stderr substrings the bash case
asserted.
**Considered:** differential testing against the bash guard — needs the bash guard kept, see
**missing-binary-exits-2**.

### Deadlines are injected in-process, CLI knobs unchanged

**ID:** inject-deadlines-in-process
**Status:** active
**Chosen:** `RUN_REPRODUCER_BOUND_SECONDS`, `RUN_REPRODUCER_GRACE_SECONDS`,
`RUN_REPRODUCER_BOUND_FILE`, `CHECK_CLEANUP_SURVIVORS_TIMEOUT` keep their integer-seconds parsing
and defaults on the CLI; tests set sub-second `time.Duration`s on `Env` directly. Shortening a
deadline is monotone in the safe direction (more timeouts, never fewer), per the bash guards' own
headers.
**Considered:** changing the env knobs to accept fractions — a CLI contract change nobody asked
for.

### Shared data helpers get a parity test against their bash/Python source

**ID:** helper-parity-tests
**Status:** active
**Chosen:** where a reimplemented helper's output is data other bash/Python guards still read —
the banned metachar set in `scripts/reproducer-metachars.sh`, and the task-field parse of
`scripts/check-task-commit-fields.py` — a Go test reads that source (or runs `python3` once over
the Go tests' plan fixtures) and fails when the two differ.
**Considered:** no parity test — silent drift between the Go port and the still-live bash/Python
copies.

### Follow-ups filed at integrate

**ID:** follow-ups-at-integrate
**Status:** active
**Chosen:** the integrate run files follow-up KAN slices per `jira-followups.md`, ordered by the
**Before** harness times; the first slice is `check-panel-reproducers`, `check-unfinished-work`,
`check-references`, `check-plan-provenance`, `check-installed-citations`, continuing until the
slowest remaining harness is under 10s.
<!-- predicted: a target threshold set at planning, confirmed by each follow-up slice's own suite measurement -->
**Considered:** filing now — the after-measurement should size the slices.

## Open questions
