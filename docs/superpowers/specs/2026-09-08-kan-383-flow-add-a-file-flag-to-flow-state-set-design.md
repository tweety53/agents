# Design — flow: add a -file flag to `flow state set` (KAN-383)

Date: 2026-09-08
Change: kan-383-flow-add-a-file-flag-to-flow-state-set
Linked issue: KAN-383 ("flow: add a -file flag to `flow state set` so the canonical write is a single non-compound command")

## Problem

A compound heredoc-piped `flow state set` command was blocked by the harness's
permission classifier mid-run (self-review finding from KAN-380), requiring a
workaround: write the JSON to a scratch file, then `flow state set ... < file`.

The canonical form the state contract itself documents is already a pipe —

```bash
printf '%s' "$RECORD_JSON" | flow state set "$NAME" -C "$DIR"
```

— so the canonical write carries the same classifier friction on every run. A
`-file <path>` flag makes the canonical form a single non-compound command
(`flow state set -file <path> <name>`), which a `Bash(flow state set *)`
allowlist rule can cover cleanly. This is the same shape as KAN-224, which
moved the workspace-id shell derivation into the CLI as `flow workspace-id`.

## Approach

Add a `state set`-only flag `-file <path>` that replaces stdin as the record
source when present. Without the flag, stdin remains the source, exactly as
today. Nothing in the write pipeline after the body is read changes: the same
`isJSONObject` check, the same `updatedAt` stamping, the same
`validateWorktreeMergeBases` refusal, the same store put, the same fallback
path, the same exit codes.

## Components

### CLI — `stats/cmd/flow/state.go`

- A `state set`-only flag layer over the shared `parseStateFlags`: `-file` is
  registered only when parsing for `state set`, so `state get -file` remains a
  usage error (exit 2). The flag's variable is captured on the passed
  `flag.FlagSet` before parse and read after, so the shared parser's
  `-addr`/`-timeout`/`-C` semantics are untouched.
- With `-file`, the body is read from the named file under the same rules
  stdin has today:
  - the `maxStdinBytes` (1 MiB) cap applies; an oversize file is reported and
    exits 2;
  - the `isJSONObject` check applies; a file that is not a JSON object is
    reported and exits 2;
  - a missing or otherwise unreadable file is reported and exits 2.
- All three are **local input errors** in the sense
  `skills/flow-contracts/state-file.md` ("The pipeline never blocks") already
  defines for non-JSON stdin and the oversize cap: reported on stderr, exit 2,
  never the fallback path, never a network call. A caller mistake is not a
  store outage, and journalling a payload that was never read would record
  nothing.
- No `-file -` stdin alias: `-` is just a filename. The flag is unnecessary
  configurability, and stdin remains reachable by omitting the flag.
- The file path resolves against the process's own working directory, like any
  other path argument. `-C` keeps its documented meaning — resolving the
  project key — and is never consulted for the file path.
- `stateUsage` and `main.go`'s `usage` line are updated:
  `flow state set [-addr url] [-timeout dur] [-C dir] [-file path] <name>`,
  noting the record is read from the file when `-file` is given and from
  stdin otherwise.

### Tests — `stats/cmd/flow/state_test.go`

In the file's existing drive-`run`/`runStateSet` style:

- `-file` with a valid record file writes through the same path stdin does —
  success exits 0 silently.
- Missing file exits 2; non-JSON-object file exits 2; oversize file exits 2.
- `state get -file` exits 2 (unknown flag — the flag is set-only).
- The existing stdin tests already pin the no-flag behaviour and stay
  untouched.

### Contract — `skills/flow-contracts/state-file.md`

- The usage comment at the top and the **Read it, write it** section: the
  canonical form becomes `flow state set -file <path> "$NAME" -C "$DIR"`,
  stated as a single non-compound command. The
  `printf '%s' "$RECORD_JSON" | flow state set` piped form remains documented
  as the supported stdin alternative.
- The canonical-form section names `Bash(flow state set *)` as the allowlist
  rule the file form is designed for: harness permission classifiers block
  compound piped/heredoc commands, and the single-command form is coverable
  (KAN-383's finding). No config file is touched — the operator's
  `.claude/settings.local.json` is machine-local and uncommitted, so the
  contract is the durable record of the intended rule.
- **The pipeline never blocks**: the local-input-error sentence gains "a
  `-file` path that cannot be read" alongside non-JSON stdin and the oversize
  cap.

## Out of scope

- `flow settings set` and any other stdin-reading subcommand — the issue names
  `state set`, and nothing else has reported classifier friction.
- A spectre spec edit: `spectre/specs/` holds no capability for the flow CLI
  (the same was true for KAN-224), so the contract markdown is the only doc
  surface.
- The operator's actual allowlist entry — an operator action, outside the
  repository.

## Decisions

Recorded canonically in the change's `design.md` (`## Decisions`); summarised
here:

1. **Add the `-file` flag** rather than the status quo (canonical write stays
   a compound pipe, classifier friction recurs) or doc-wording-only change (no
   CLI change leaves the pipe canonical).
2. **Contract names the allowlist rule** rather than editing machine-local
   settings (uncommitted, nothing to review) or leaving it out entirely (the
   issue's motivation is allowlist coverage; recording it keeps the why
   durable).
3. **Stdin stays the default** — backwards compatible; making `-file`
   mandatory would break every existing caller for no reported benefit.
4. **File read errors are local input errors (exit 2, never fallback)** —
   same family as non-JSON stdin; a caller mistake is not a store outage.

## Testing

Go cmd tests in `stats/cmd/flow/state_test.go` per the task plan's `**Tests:**`
fields; lint via the project's configured Go checks (`gofmt`, `go vet`, and the
repo's `scripts/check-*.sh` guards per `.flow/project.md`).
