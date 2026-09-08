# kan-383-flow-add-a-file-flag-to-flow-state-set

## Context

The change adds a `-file <path>` flag to `flow state set`
(`stats/cmd/flow/state.go`, `runStateSet`) so the state contract's canonical
write becomes a single non-compound command. Everything constraining it is
already in the tree:

- `runStateSet` reads the whole record from stdin only, capped at
  `maxStdinBytes` (1 MiB), then runs `isJSONObject` → stamp → validate → put
  → fallback. Only the read of the body changes; everything after it is
  untouched.
- Flag parsing is shared by `state get`/`state set` via `parseStateFlags`
  (`-addr`, `-timeout`, `-C`). `-file` must therefore be registered at the
  `state set` layer only, so `state get -file` stays a usage error.
- `skills/flow-contracts/state-file.md` is canonical for the write contract:
  its **Read it, write it** section carries the piped canonical form, and its
  **The pipeline never blocks** section defines the local-input-error family
  (reported, exit 2, never fallback) the new file errors join.
- `spectre/specs/` holds no capability for the flow CLI, so there is no spec
  edit — the contract markdown is the only doc surface (the same held for
  KAN-224).
- The allowlist rule the flag exists for lives in operator-side harness
  settings; the project's `.claude/settings.local.json` is machine-local and
  uncommitted, so no config file is part of this change.

It is one change because the flag, its error semantics and the contract's
canonical form are one contract movement: landing the flag without the doc
update leaves the pipe canonical, and updating the doc without the flag
documents a command that does not exist.

The full design, including alternatives considered, is
`docs/superpowers/specs/2026-09-08-kan-383-flow-add-a-file-flag-to-flow-state-set-design.md`
in this repository.

## Decisions

### Add a `-file <path>` flag to `flow state set`

**ID:** state-set-file-flag
**Status:** active
**Chosen:** a `state set`-only `-file <path>` flag replacing stdin as the record source when present — the canonical write becomes one non-compound command a `Bash(flow state set *)` allowlist rule can cover (KAN-383's finding)
**Considered:** the status quo (the canonical write stays a compound pipe and the classifier friction recurs on every run); a docs-only rewording (no CLI change leaves the pipe canonical); a generic `-file` across every stdin-reading subcommand (`flow settings set` has no reported friction — speculative scope)

### The contract names the allowlist rule

**ID:** contract-names-allowlist-rule
**Status:** active
**Chosen:** `state-file.md`'s canonical-form section names `Bash(flow state set *)` as the allowlist rule the file form is designed for; no config file is touched
**Considered:** editing the project's `.claude/settings.local.json` (machine-local and uncommitted — nothing in the change would be reviewable); leaving the allowlist motivation unrecorded (the contract is the durable record of why the flag exists)

### Stdin remains the default source

**ID:** stdin-remains-default
**Status:** active
**Chosen:** without `-file`, `state set` reads stdin exactly as today, and the piped form stays documented as the supported alternative — backwards compatible with every existing caller
**Considered:** making `-file` mandatory (breaks every existing caller for no reported benefit); a `-file -` stdin alias (unrequested configurability; omitting the flag already reaches stdin)

### File read errors are local input errors

**ID:** file-errors-exit-2
**Status:** active
**Chosen:** a missing, unreadable, oversize or non-object file is reported on stderr and exits 2, never the fallback path, never a network call — joining the local-input-error family `state-file.md` already defines for non-JSON stdin and the oversize cap
**Considered:** taking the fallback on an unreadable file (a caller mistake is not a store outage, and journalling a payload that was never read would record nothing)

## Open questions

None — the brainstorming round closed with every question answered and the
design approved.
