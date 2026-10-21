## Context

The guard ports (KAN-778 and its successors) moved each guard's logic into
`stats/internal/guard/` and left the bash source beside it. For the files below, a grep on the
base (`ef6f87cd`) finds no live bash or Python consumer — only comments citing them and Go tests
that source the live file for parity. The brainstorm design is
`.superpowers/sdd/2026-10-21-kan-887-flow-fix-delete-the-retained-bash-twins-and-design.md` in this
worktree.

## Decisions

### Keep spec-root.sh for KAN-923

**ID:** keep-spec-root-for-kan-923
**Status:** active
**Chosen:** leave `scripts/lib/spec-root.sh` in place — `check-plan-shape.sh`,
`check-task-build-green.sh`, `integrate-change.sh` and `plan-dispatch-bundles.sh` still source it,
the ticket's own condition for leaving a file to KAN-923.
**Considered:** deleting it with the rest — rejected: four live consumers would break.

### Delete project-section.sh

**ID:** delete-project-section
**Status:** active
**Chosen:** delete `scripts/lib/project-section.sh` — the ticket named `project-get.sh` as a
consumer, but `project-get.sh` is now a flow-guard shim that no longer sources it; only
`strip-bom.sh`'s comment and Go tests name it.
**Considered:** leaving it to KAN-923 as the ticket's text suggests — rejected: its premise
(a live consumer) no longer holds.

### Parity tests read the pinned historical copy

**ID:** parity-at-historical-ref
**Status:** active
**Chosen:** a Go parity test that sources a deleted bash file reads it through `historicalFile` at
`ef6f87cd` (the base this change was cut from), the existing parity-ref-on-main pattern
`project_get_test.go` and `aside_planning_artifacts_test.go` already use.
**Considered:** deleting the parity tests — rejected: they are the only proof the Go port still
behaves as the bash did; keeping them live against the file — impossible once it is deleted.

### Drop the metachars bash pin

**ID:** drop-metachars-bash-pin
**Status:** active
**Chosen:** delete `TestMetacharsMatchBashSource` and the bash half of case 19 — case 19's literal
`want` already pins `reproducerMetachars`, so a historical copy adds nothing.
**Considered:** repointing to the historical copy — rejected: a literal compared with a frozen
literal proves nothing the `want` constant does not.

### Record the deletion in the Go twin's header

**ID:** record-in-go-header
**Status:** active
**Chosen:** each Go twin's header ("stays the source" / "Go twin of") and `pipeline.md`'s
`changePlanDir` sentence say the bash twin was deleted by KAN-887, without naming the deleted
path, so each task's absence sweep stays clean; every other comment citing a deleted file names
the Go twin instead.
**Considered:** a central ledger file — rejected: no such file exists; the port decisions live in
those headers.

### Harnesses are audited case by case

**ID:** harness-case-audit
**Status:** active
**Chosen:** each duplicated harness is deleted only after every case it asserts is matched to a Go
test case; a missing one is added to the Go test in the same task.
**Considered:** deleting on the ticket's word — rejected: the ticket asks for the audit.

## Open questions

## Live check

None — the change deletes dead scripts and test code; no running service or persistent state is
touched.
