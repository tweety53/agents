# kan-843-agents-remove-cursor-codex-harnesses-myflow-era

## Why

Only Claude Code and zcode are in use. The repository still installs into Cursor and Codex, keeps a
myflow-era vocabulary ("banned words") guard, carries stats code that reads data shapes written by
long-retired versions, and special-cases individual KAN tickets in guards and tests. None of it is
used now, and all of it is read, maintained and tested on every change. Keep only what matters and
is used now (KAN-843).

## What changes

- `setup.sh` installs for Claude Code and zcode only — modes `claude-code`, `zcode`, `global`; the
  Cursor command set `commands/`, every `~/.cursor`/`~/.codex` install and the `~/.codex/AGENTS.md`
  managed block are gone, and so is every Cursor/Codex mention in live docs, rules, skills,
  scripts and stats.
- `scripts/check-vocabulary.sh` and its harness are deleted, with every reference to them.
- myflow leftovers go: the self-review guard's legacy labels and 17-report exemption (the 17
  pre-shape reports are deleted, 14 reports relabelled `myflow-<angle>` → `flow-<angle>`), myflow
  prose, and the `kan-15-…myflow…` worked example.
- The stats store migrates its legacy rows (synthetic `updated_by`, bare-array decision groups,
  the collapsed pricing column, dropped without a backfill) and the code that read the old shapes
  is removed.
- Code that special-cases a KAN ticket is removed or named by behaviour.

Out of scope: renaming `rules/*.mdc`, `.idea/`, `KNOWN-BUGS.md`, historical `docs/self-review/`
reports beyond the two edits above, inline `(KAN-NNN)` comment citations, speeding up
`test-setup.sh` (KAN-844).
