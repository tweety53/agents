# Design — KAN-297

## Context

A review panel finding is a row in the store, written `open` and never moved off it. The fix
round repairs the defect and the parent verifies the repair, but no step records the outcome on
the finding, so `check-unfinished-work.sh` reports every repaired finding as outstanding at
`/flow`'s integrate gate. `proposal.md` carries the observed incident and the scope boundary —
KAN-297's second defect is already fixed and is not touched here.

## The mechanism

A finding's lifecycle today: `flow record finding … -status open` writes the row; the fix round
repairs it; `check-unfinished-work.sh` reads the store at integrate and counts anything that is
neither `fixed` nor `withdrawn …` as outstanding. Nothing sits between the repair and that count.

This change puts two things there.

### 1. The close, at the verify point

`skills/flow/review-panel.md`'s fix round already computes the exact predicate a close needs:

> Once the fix subagent reports, re-run every dispatched finding's reproducer under the same
> constraints and require it now to exit **0**. **The flip alone does not close a finding — the
> fix's diff must also touch at least one path the finding named, with a non-comment,
> non-whitespace change.**

When both conditions hold for a finding, the parent records it there and then:

```bash verified:flow record status's flag set, against `flow record`'s own usage output
flow record status -change <name> -ref F<n> -status fixed
```

Three properties of that placement, each load-bearing:

- **The parent records it, never the fix subagent.** The parent holds the evidence — it ran the
  reproducer and walked the diff. A subagent that overstates its work would otherwise close a
  finding nothing verified, which is the failure the parent's verification exists to catch.
- **Per finding, as the verdict is reached — not batched at the round's end.** A round that aborts
  part-way then still leaves every already-verified finding closed. A batch at the end loses them.
- **A finding that fails either condition is untouched.** It stays `open` and goes to the operator
  through the existing handback, exactly as today.

`review-panel.md:277`'s "Fix subagents record `fixed` when they fixed it, and leave `open` when they
did not" is replaced by the statement above: left standing it would contradict the parent's
ownership.

### 2. The gate, at panel close

`scripts/check-panel-findings-closed.sh <worktree> <change-name>` — a sibling of
`check-panel-reproducers.sh`, reading the same array:

- `export LC_ALL=C`, the change-name containment allowlist, and the `-d "$WORKTREE"` check, all in
  the shape both sibling guards already carry.
- One store read: `flow record findings -change "$NAME" -C "$WORKTREE"`. **stdout and stderr are
  captured separately.** `flow` writes diagnostics to stderr — most reliably the
  `flow: using FLOW_ADDR=…` line it prints whenever the address is overridden — and folding them
  into the payload with `2>&1` puts that line at the head of the JSON, so `jq` fails on a run that
  actually succeeded. `check-unfinished-work.sh` carries this same separation and its own comment is
  canonical for why.
- A jq count of findings whose `status` is neither `fixed` nor a `withdrawn …` value.
- Exit **0** when the count is zero; exit **1** naming each still-open ref on stderr; exit **2**
  when it cannot answer at all — no worktree, no change name, a name outside the allowlist, the
  store unreachable, or jq failing.

`review-panel.md` invokes it immediately before `flow stage end -command '/flow' -stage
flow.review-panel`. That file already claims the stage ends only at "zero open findings at any
severity"; nothing checked it.

## Decisions

### Where a finding gets closed

**ID:** close-at-verify-point
**Status:** active
**Chosen:** the parent, per finding, at the fix round's existing verification step — it is the only
place the evidence for "this was actually repaired" exists, and recording there survives an aborted
round.
**Considered:** the fix subagent closing what it reports fixed — cheaper and closest to the current
prose, but it lets an overstating subagent close a finding nothing verified, which is precisely what
the parent's verification exists to prevent. The parent closing everything verified in one call as
the round closes — fewer calls, but a round that stops part-way loses every close it had earned.

### No bulk close operation

**ID:** no-bulk-close
**Status:** active
**Chosen:** none. The per-finding `flow record status` call the CLI already has is the whole
mechanism.
**Considered:** the "close every finding for round N" operation KAN-297's own Fix bullet names — new
CLI verb, HTTP endpoint and store method. Rejected on two grounds. It is the wrong shape: a fix
round routinely leaves findings surviving (the handback loop exists for them), so closing by round
would close findings nobody repaired. And it solves a cost that stops existing: the thirty by-hand
calls on KAN-295 were the consequence of nothing closing findings at all, not of the calls being
expensive — once the round closes them as it verifies them, the operator makes zero calls, not one.

### The gate is a shipped guard, not skill prose

**ID:** gate-is-a-guard
**Status:** active
**Chosen:** a new `scripts/check-panel-findings-closed.sh` with its own mutation-tested harness.
**Considered:** inline prose in `review-panel.md` telling the parent to run `flow record findings`
and count — about six lines and no new files, but nothing would check that the agent did it, and an
unenforced prose obligation going unperformed is the entire defect this change repairs. Adding the
rule to `check-panel-reproducers.sh` — no new file or harness, but that guard's name and stated
contract are about reproducers, so it would silently check two unrelated things.

### The open-finding predicate is duplicated, not shared

**ID:** duplicate-the-predicate
**Status:** active
**Chosen:** the new guard carries its own copy of `status != "fixed" and (status | startswith
("withdrawn") | not)` rather than sourcing it from `scripts/lib/`.
**Considered:** a shared `scripts/lib/` helper. Rejected on this repository's own recorded
precedent: `check-unfinished-work.sh`'s containment comment states that its six-line `case` block
stays duplicated *on purpose*, because both harnesses then assert the same shape and that is what
keeps the two copies from drifting silently. A one-line jq predicate under two mutation-tested
harnesses is the same trade.

### The guard never consults the journal

**ID:** no-journal-excuse
**Status:** active
**Chosen:** the guard reads the store alone. A close that was journalled rather than stored — `flow
record status` never blocks, so a store outage journals the write and exits 0 — leaves the finding
reading `open`, and the guard reports exit 1.
**Considered:** consulting `flow record journal-count` and treating a pending journal entry as a
close. Rejected: it would clear a finding on the strength of a write that has not landed, failing
toward the reassuring answer. `check-unfinished-work.sh`'s header names that direction as the one
failure the whole guard family exists to remove. Exit 1 is the honest verdict, and the round's
existing handback is where the operator resolves it.

## Open questions

None.

## Testing

`scripts/test-check-panel-findings-closed.sh`, built to `test-check-panel-reproducers.sh`'s shape
and discovered automatically by `scripts/run-guard-tests.sh`'s `scripts/test-*.sh` glob — no
`.flow/project.md` edit is needed.

Cases:

| Case | Store answers | Expected |
|------|---------------|----------|
| every finding closed | all `fixed` | exit 0, no stderr |
| one still open | `F1 open`, rest `fixed` | exit 1, `F1` named on stderr |
| several still open | `F1 open`, `F3 open` | exit 1, both named |
| withdrawn counts as closed | `withdrawn <reason>` | exit 0 |
| no findings at all | `[]` | exit 0 |
| store unreachable | `flow record findings` exits non-zero | exit 2 |
| change name outside the allowlist | — | exit 2, no store call |
| missing arguments | — | exit 2 |
| worktree is not a directory | — | exit 2 |

Mutation tests, per KAN-197's requirement that every guard carry one:

- Remove the `startswith("withdrawn")` clause → the withdrawn-counts-as-closed case must fail.
- Remove the open-count branch that adds the violation → the one-still-open case must fail.
- Replace the separated stderr capture with `2>&1` → a case whose store read emits a diagnostic line
  on stderr must fail.

## Verification

- `scripts/run-guard-tests.sh` — the new harness plus every existing one.
- `scripts/check-guard-symlinks.sh` — rule 2 requires the symlink at `skills/flow/scripts/` the
  moment `review-panel.md` invokes the guard by name in a fence; rule 1 requires it to be relative
  and to resolve.
- `scripts/check-references.sh` and `scripts/check-vocabulary.sh` — the edited skill prose.
- `cd stats && go vet ./... && gofmt -l .` — untouched by this change, run to confirm it stays clean.
