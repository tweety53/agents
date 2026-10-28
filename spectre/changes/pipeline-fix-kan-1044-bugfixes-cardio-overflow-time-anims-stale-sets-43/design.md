## Context

Loop 41 (`e0d2ff81`) text: `rules/fix-determinism-at-the-source.mdc` § "Read the flake before touching
anything" (lines ~32-39), `skills/flow/implement.md` `> **PROOF RUNS:**`, `skills/flow/verify-fix-loop.md`
item 2 (~line 106), `skills/flow/primary-reviewer-prompt.md` (~line 57), `skills/flow/unfinished-work-gate.md`
(~line 24); guard `stats/internal/guard/dispatchparagraphs.go` entries `proofruns` (line 84) and
`noloadrepro` (line 86) with sites at lines 136-138.

Central executor of `## test`/`## lint`: `stats/internal/guard/runverifylist.go` — every command is run as
`bash -c <cmd>` (argvs built at ~lines 367, 382, 384), optionally behind `flow suite record ... --`.
Dev-stack start: `skills/flow/verify-and-handoff.md` bullet "Every start command comes from ... `## run`"
(~line 361); `skills/flow/implement.md` ~line 620 hands `## run` to implementers.

Blast radius: files the fix edits or adds = 13 (run-capped.sh, its test, runverifylist.go, its test,
project-configuration.md, the five wording files above, verify-and-handoff.md, dispatchparagraphs.go,
check_dispatch_paragraphs_test.go). Citing files reference rule/section names that do not change.
Not `big`.

## Decisions

- **One bash wrapper, no cgroups/ulimit.** macOS has no cgroups and `ulimit -v` is ignored by darwin;
  polling group RSS with `ps` works on both macOS and Linux. Ceiling: 1 s poll granularity, a spike
  inside one second can overshoot (`# ponytail:` comment in the script).
- **Interface:** `run-capped.sh [-m <MB>] -- <cmd...>`. Without `-m` the cap is
  `project-get.sh "$PWD" 'resource cap'` first integer, else 4096. On exceeding: kills the group
  (TERM, then KILL after 2 s), prints `CAPPED: <cap>MB (group RSS <n>MB)` to stderr, exits 137.
  Otherwise exits with the command's own status. Runs the child via `set -m` / `setsid`-free
  background job so it gets its own process group (bash job control); traps INT/TERM to kill the group.
- **Wire once, centrally.** `runverifylist.go` prefixes every argv `bash -c <cmd>` with
  `<scriptDir>/run-capped.sh --` (scriptDir already known there). No per-call-site edits elsewhere.
- **Dev stack and load repro by wording** — the parent starts them by Bash, so the contract text says
  to prefix with `<agents repo>/scripts/run-capped.sh --`; no guard beyond the phrase guard.
- **Bounded workers stated in the wording**, sized under the cap: Gradle `--max-workers`/`maxParallelForks`
  and `-Xmx`, Playwright `--workers`, vitest `--maxWorkers`.
- **Unreproducible → straight to code.** One capped reproduction attempt series; if the cause does not
  show within the cap, no further reproduction loops: fix on deeper code analysis, commit body says so,
  mechanism-removed run waived and stated.
