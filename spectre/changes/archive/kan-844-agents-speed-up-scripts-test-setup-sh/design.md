# Design — kan-844-agents-speed-up-scripts-test-setup-sh

## Context

Measured at `75410c1c` (post-KAN-843 `main`), 2026-09-28, this machine (10 cores):

- `scripts/test-setup.sh`: 580 assertions in 26 groups, all passing; one description (`the abort
  names the offending rule`) occurs in two groups.
- Wall time across four runs: 40.0 s, 40.5 s, 31.9 s, 12.7 s — load average 16–17 from other
  sessions throughout. A single run is noise.
- Profile of one 21.1 s run: 75 `setup.sh` calls took 17.5 s. Real-repo `global` ≈ 1.0 s each,
  fixture `global` 0.25–0.7 s, `claude-code` ≈ 0.16 s; each fingerprint pair ≈ 0.4 s.
- `setup.sh` never writes into the source tree it installs from, so groups may share a read-only
  fixture repo.
- Only one pair of groups shares state: "Every guard in a command skill's scripts/ directory
  reaches the install" and "check-panel-reproducers.sh runs through the installed path…" — the
  latter runs the guard out of the former's HOME.
- Effort vocabulary: three agent definitions (`agents/flow-{low,medium,high}.md`, `effort:` in
  frontmatter, since the Agent tool has no dispatch-time effort parameter); `flow record`'s closed
  `-effort` set is `low`/`medium`/`high`/`default` (`stats/cmd/flow/record.go` `recordEfforts`);
  `dispatches.effort` is unconstrained `TEXT` (migration 0020). Claude Code subagent frontmatter
  accepts `xhigh` and falls back to `high` on models without it (code.claude.com docs: sub-agents,
  statusline).

## Decisions

### Port the harness to Go and run its groups concurrently

**ID:** setup-harness-go-parallel
**Status:** active
**Chosen:** `stats/internal/setuptest/`, `_test.go` files only; one top-level `Test…` per group,
each `t.Parallel()`, each with its own sub-sandbox under one `/tmp/flow-test-setup.*` sandbox;
`TestMain` reads the delimiters out of `setup.sh`, creates the sandbox, takes both fingerprints,
builds the shared fixture repo, runs the tests, then compares the fingerprints (the close-out
group) and removes the sandbox unless `KEEP_SANDBOX` is set. `scripts/test-setup.sh` becomes a
`go test ./internal/setuptest/ -count=1` shim, the `scripts/test-go-guards.sh` pattern. The
dependent pair above is one test. Groups that mutate a fixture build their own; the rest share the
base fixture (the bash harness's later groups ran against the prune fixture, whose content equals
the base). Fingerprints and assertions run in-process; every assertion description is kept byte for
byte and logged `✓ <desc>` on pass.
**Considered:** bash restructure through `scripts/lib/parallel.sh` self-reinvocation — keeps
≈ 600 per-assertion forks and needs per-group count files; splitting into several
`test-setup-*.sh` harnesses — fragments the containment check and the standalone run no longer
covers everything; a sequential trim (shared installs, cheaper fingerprint) — ≈ 3 s of ≈ 21 s,
misses 2×.

### Every group keeps a fresh HOME

**ID:** setuptest-no-shared-installs
**Status:** active
**Chosen:** no install is shared between groups.
**Considered:** one fixture `global` and one real-repo `global` shared by read-only groups —
under parallelism it saves CPU, not wall, and a group that mutates a shared install breaks another.

### Prove containment on fake roots

**ID:** setuptest-leak-detection
**Status:** active
**Chosen:** the fingerprint functions take a root; a new group fingerprints a fake home and a fake
source tree under the sandbox, writes into each sampled location, and asserts each fingerprint
moved. Its descriptions start `leak detection: `.
**Considered:** a one-off manual mutation writing into the real HOME — never acceptable; relying on
the ported logic being unchanged — the acceptance asks for proof.

### The Go test also runs under `stats-go`

**ID:** setuptest-in-stats-go
**Status:** active
**Chosen:** no build tag; `go test ./...` runs the package, as it already runs the guard tests that
`scripts/test-go-guards.sh` also runs.
**Considered:** a build tag excluding it from `./...` — a knob nobody asked for, against precedent.

### Measure as the median of interleaved runs

**ID:** setuptest-timing-interleaved-median
**Status:** active
**Chosen:** A = the bash harness from a detached `75410c1c` worktree, B = the Go shim on the branch;
one warm-up each, then five A/B pairs; record real, user+sys and load average per run; target
median(A) / median(B) ≥ 2.
**Considered:** one run each — a single run swings 3× on load alone; quiet-machine runs only —
needs scheduling nobody asked for.

### `xhigh` for implementers only

**ID:** xhigh-implementers-only
**Status:** active
**Chosen:** `agents/flow-xhigh.md` (`effort: xhigh`, the same `tools:` allowlist without `Agent`);
the implementer pair and each implementer group choose from `low`/`medium`/`high`/`xhigh`; the
fixer, panel dispatches and rerun pair are unchanged; a `big` plan's gated per-task reviewer, which
otherwise takes its group's pair, runs at `high` for an `xhigh` group (operator, 2026-09-28, panel
round 0 F4); `recordEfforts` gains `xhigh`, which also
lets an inline row record an `xhigh` parent. zcode's mapping stays `high`.
**Considered:** adding the fixer (implementer work per model policy) — not asked for; every chosen
pair — not asked for.

## Open questions

## Measurements

2026-09-28, this machine. A = bash harness at `75410c1c` (detached worktree), B = the Go shim on the
branch; one warm-up each (A 12.58 s, B 3.53 s, not counted), then A, B interleaved. Every run exited 0.

| Run | Harness | real (s) | user+sys (s) | load average (1/5/15 min) |
|-----|---------|---------:|-------------:|---------------------------|
| 1 | A | 11.60 | 9.36 | 3.13 3.16 4.69 |
| 1 | B | 3.35 | 15.38 | 2.95 3.12 4.66 |
| 2 | A | 11.50 | 9.35 | 3.20 3.17 4.66 |
| 2 | B | 3.41 | 14.79 | 3.17 3.17 4.64 |
| 3 | A | 11.61 | 9.38 | 3.31 3.20 4.65 |
| 3 | B | 3.59 | 15.01 | 3.27 3.19 4.63 |
| 4 | A | 11.74 | 9.56 | 3.97 3.34 4.67 |
| 4 | B | 3.87 | 15.74 | 3.81 3.32 4.65 |
| 5 | A | 11.90 | 9.63 | 4.55 3.48 4.70 |
| 5 | B | 3.54 | 15.32 | 4.38 3.48 4.68 |

- median(A) = 11.61 s; median(B) = 3.54 s; median(A) / median(B) = **3.28** — target ≥ 2 met.
- B spends more CPU (≈ 15.3 s vs ≈ 9.4 s user+sys, compile included) to finish in under a third of
  the wall time — the parallel groups trade CPU for wall, as **Decision:** setuptest-no-shared-installs
  expected.
- Load was 3–4.5 throughout, against 16–17 in the `## Context` runs, so A's 11.6 s here is the quiet
  figure; the ratio is measured on the same load, run for run.
- Parity on the head: 580 assertions, empty diff against `setup-harness-baseline.tsv`; `✓ leak
  detection: ` lines = 5.
<!-- measured: /usr/bin/time -p <A|B> and uptime per run, interleaved A,B ×5 after one warm-up each @ branch spectre/kan-844-agents-speed-up-scripts-test-setup-sh (A: @ 75410c1c) -->
<!-- measured: the tasks.md parity check with every group title, RUN='.'; grep -c '✓ leak detection: ' @ branch spectre/kan-844-agents-speed-up-scripts-test-setup-sh -->
