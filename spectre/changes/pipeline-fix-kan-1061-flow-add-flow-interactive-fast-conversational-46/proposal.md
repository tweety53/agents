# pipeline-fix-kan-1061-flow-add-flow-interactive-fast-conversational-46

## Why

`skills/flow/implement.md` tells the parent to run `check-plan-provenance.sh` when transcribing a
`Correction:` paragraph, but `skills/flow/scripts/` carries no such symlink, so in run kan-1061 the
call through `~/.claude/skills/flow/scripts/` failed with "no such file or directory".
`scripts/check-guard-symlinks.sh` still printed GUARD-SYMLINKS-OK: its rule 2 reads a backtick
basename as an invocation only when "Run" precedes it, and the sentence put "run" after it. Behind
that sits a second defect: the shim derives its root as `$(dirname BASH_SOURCE)/..`, so called
through a skill's `scripts/` symlink it scans `skills/flow` and exits 2 with
"spectre/changes not found under …/skills/flow".

## What changes

- `skills/flow/scripts/check-plan-provenance.sh` symlinks `../../../scripts/check-plan-provenance.sh`.
- The `implement.md` sentence reads "Run `check-plan-provenance.sh` when transcribing it", an
  invocation rule 2 detects, so the guard now requires the symlink.
- `scripts/check-plan-provenance.sh` derives its root through `flow_guard_root`
  (`scripts/lib/flow-guard.sh`), as `check-plan-shape.sh` does.
