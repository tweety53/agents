# kan-382-flow-review-panel-reproducer-shape-rejected — design

## Context

The accepted reproducer shape already exists as code — `check-panel-reproducers.sh` enforces it
lexically at read time, against rows `flow record finding` validates only for emptiness and the
`none` form. One contract, two halves: make it enforceable at write time and visible to every slot
at dispatch time. `spectre/specs/` carries no capability specs, so no spec edit is planned.

## Design

### 1. Write-time validation — `stats/cmd/flow/record.go`

`validateFindingReproducer` extends to mirror the guard's lexical rules exactly, before the store
is contacted:

- first word `none` → must be exactly `none — <reason>` (em dash, non-empty reason); already
  enforced, unchanged.
- any other value is a runnable command and must carry:
  - no shell metacharacter anywhere in the text — the `REPRODUCER_METACHARS` set
    (`| & ; $ \` ( ) < > { } ~ * ? [ ] # \ ' "`);
  - no leading `-` on the path token (the first whitespace-separated token);
  - no `://` anywhere;
  - no absolute token and no `..` path segment on any token.
- each refusal names the broken shape, so the bounce tells the raising slot what to fix.
- the validator's doc comment is rewritten — it currently reads "this validator has nothing
  further to say about its shape".

`flow record finding` exits non-zero and writes nothing; the panel parent bounces the finding to
the raising slot.

### 2. The brief states the shape — `skills/flow/review-panel.md`

The "Every slot must supply, per finding, a reproducer" paragraph — carried on every slot's
dispatch prompt — states the accepted form verbatim: the bare relative path of the slot's
reproducer script under `.superpowers/sdd/reproducers/`, optionally followed by plain arguments,
subject to the same five prohibitions; or exactly `none — <reason>`; and that `flow record
finding` refuses anything else at write time.

### Testing

- `stats/cmd/flow/record_test.go`: per-rule refusals (representative metacharacters including both
  quote characters, leading-dash path token, URL, absolute token, `..` segment, bare `none`,
  hyphen-instead-of-em-dash exemption), accepted shapes (bare script path with plain arguments,
  `none — <reason>`), and a drift-sync test parsing `REPRODUCER_METACHARS` out of
  `scripts/reproducer-metachars.sh` and asserting equality with the Go const.
- Guard and runner unchanged: `check-panel-reproducers.sh` stays as read-time defense in depth for
  rows written by older CLIs.

## Decisions

### Metachar set stays single-sourced in bash, mirrored in Go by a sync test

**ID:** reproducer-sourcing
**Status:** active
**Chosen:** const in Go + unit test parsing `scripts/reproducer-metachars.sh` — the set has
drifted twice (that file's header), and one assertion closes that permanently.
**Considered:** comment-only duplicate — rejected, it re-opens the recorded drift; generating the
Go constant from the bash file — rejected as overkill for one string.

### Write-time validation mirrors the guard's whole lexical shape set

**ID:** full-shape-mirror
**Status:** active
**Chosen:** all five prohibitions at write time — a rejected "shape" is the whole accepted form,
and a partial mirror would leave some bad reproducers bouncing only at read time, after the store
landing this change exists to prevent.
**Considered:** metachar + none-form only (the ticket's literal minimum) — rejected for the same
reason.

### Refused reproducers bounce at record time

**ID:** bounce-at-record-time
**Status:** active
**Chosen:** `flow record finding` exits non-zero before any store contact; the panel parent
re-dispatches to the raising slot.
**Considered:** accepting the write and repairing later — rejected, it is today's defect.

## Open questions
