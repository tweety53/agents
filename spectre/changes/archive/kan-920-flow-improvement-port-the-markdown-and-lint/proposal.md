# kan-920-flow-improvement-port-the-markdown-and-lint

## Why

- Every new guard and every extension of one is Go in `flow-guard` (`.flow/project.md`), so a guard
  still in bash must be ported before it can be extended; each port also takes a bash harness off
  `scripts/run-guard-tests.sh`.
- Three `## lint` guards over the Markdown corpus and the lint policy are still bash (KAN-920).

## What changes

- Three scripts run as Go ports inside `flow-guard`: `check-normative-inventory`,
  `check-mutation-reproducer-pin`, `check-python-suppressions` — each CLI contract unchanged:
  arguments, environment overrides, output lines (the inventory's `LC_ALL=C` sort order included),
  stdout/stderr split, exit codes.
- Each `scripts/<name>.sh` becomes a `flow_guard_exec` shim keeping its header; its bash harness is
  replaced by an in-process Go test, every case ported.
- `scripts/lib/owned-corpus.sh` loses its last bash consumer and is deleted; the Go corpus walk
  `check-ask-silence` already carries becomes the one definition, shared by both Go guards.
