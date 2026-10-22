# kan-922-flow-improvement-port-the-bash-python-tooling

## Why

KAN-922: the bash → Go migration inventory (2026-10-07) left five non-guard helpers in bash or
Python. Two of them — `plan-dispatch-groups` and `generate-relocation-comparison` — were ported by
KAN-923 (`d8045f82`). Three remain bash: `scripts/workspace.sh`, `scripts/load-sets.sh` and
`scripts/installer-sandbox-diff.sh`, two of them with bash harnesses.

## What changes

- `flow-guard workspace`, `flow-guard load-sets` and `flow-guard installer-sandbox-diff` carry the
  three helpers' behaviour, with byte-for-byte parity of stdout, stderr and exit codes.
- Each `scripts/<name>.sh` becomes a thin `flow_guard_exec` shim at the same path, so every caller
  — `.flow/project.md`'s `## workspace isolation` triple, `stats/Makefile`'s `ui-test-up` — is
  unchanged.
- Go tests carry every case of `scripts/test-workspace.sh` and `scripts/test-installer-sandbox-diff.sh`,
  and both harnesses are deleted; `load-sets` gains a Go test (it had no harness).
- `setup.sh` stays bash (decided 2026-10-07).
