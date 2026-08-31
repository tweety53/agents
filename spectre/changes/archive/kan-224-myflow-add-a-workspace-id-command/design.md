## Context

`stats/cmd/flow` is the daemon-agnostic CLI every `/flow` skill shells out to. `scripts/prepare-workspace.sh`
already carries a Bash port of the same four-step derivation (lines 143-152) for the apply-worktree
case, where the change name is read off the branch. Run 2 has no worktree to read a branch from by
the time it needs the id (the worktree is what's being removed), and does its own hand-derivation
today with only the change name in hand — that hand-derivation is the piece this change replaces.
This is one change because the command and its one caller-facing consumer (the finish contract) are
inseparable: a command with no wired caller is dead code, and rewiring the caller without the command
does not exist.

## Decisions

### Add a `flow workspace-id <name>` command rather than a shell script

**ID:** cli-command-not-shell-script
**Status:** active
**Chosen:** a Go subcommand on the existing `flow` CLI — `stats/cmd/flow/workspaceid.go` — following
the `tick.go`/`tasks.go` split (pure function + tests, thin cmd file + tests) already established for
`flow tasks tick`.
**Considered:** the ticket's own alternative, `scripts/derive-workspace-id.sh <change-name>` — ruled
out because the ticket itself says "the Go app already owns the derivation," and a second
implementation (shell) alongside the CLI's existing state/stage/settings/tasks commands would be the
"two implementations that must agree" problem `workspace-isolation.md` already goes out of its way to
avoid (`LC_ALL=C`, byte-wise definition) — a Go implementation has no locale to get wrong in the
first place.

### Print the bare id only — no `_underscored` or port-offset flags

**ID:** id-only-no-derived-values
**Status:** active
**Chosen:** the command's entire output is the id on one line (e.g. `kan-15-55a6`), nothing else.
**Considered:** also printing the underscored spelling and/or the port offset — ruled out because the
ticket's stated caller (run 2's `remove` command substitution) needs only the bare id, and
`workspace-isolation.md` already treats the underscored spelling and the port offset as separate,
narrower derivations (KAN-236/KAN-249 are the tickets for those). Adding flags nobody asked for is
exactly the speculative-configurability this repo's build-the-simplest-thing rule forbids.

### Leave `scripts/prepare-workspace.sh`'s own inline derivation untouched

**ID:** no-touch-prepare-workspace
**Status:** active
**Chosen:** `prepare-workspace.sh` keeps its own Bash derivation (lines 143-152) exactly as it is.
**Considered:** having it shell out to `flow workspace-id` instead, for one implementation total —
ruled out as scope creep for this ticket: `prepare-workspace.sh` already has the change name from the
worktree's branch and its own working, tested Bash block; rewiring it is a separate, unrequested
refactor with its own blast radius (bash 3.2 floor, `set -euo pipefail`, its own test suite) that
KAN-224 does not ask for.

### Wire `finish-contract-run2.md` step 5 to call the new command

**ID:** run2-calls-new-command
**Status:** active
**Chosen:** step 5's text changes from "run 2 re-derives it [by hand, from the contract]" to instruct
calling `flow workspace-id <name>`, matching the ticket's "Run 2 then calls it; the contract stays
canonical and is loaded only by whoever edits it."
**Considered:** leaving the contract's prose as-is and only shipping the command — ruled out because
an unused command does not fix the ticket's actual complaint (the hand-transcription step immediately
before a destructive removal); the fix is only real once run 2's own procedure calls it.

## Open questions

<!-- none -->
