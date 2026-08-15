## Why

KAN-172 shipped session attribution and it does not work. **Every stage run from the finish sequence
that landed KAN-172 itself — 14 of them, correct keys, correct harness, correct token — is unbound.**

`commandBeginsWithMyflow` strips one leading `cd <path> &&` and then requires the very next field to
be `myflow`. Real marks never look like that: they arrive inside multi-line shell blocks with
variable assignments — `N=…; T=…; cd …` and then `myflow stage begin` on a later line — and every one
is rejected.

<!-- measured: run against the merged code on 2026-08-15 -- the two clean shapes match, the two real shapes do not -->

| Command shape | Matches |
|---|---|
| `myflow stage begin … -session-token …` | yes |
| `cd /repo && myflow stage begin …` | yes |
| `N=kan-172; T=mf-x; cd /repo` ⏎ `myflow stage begin …` | **no** |
| `WT=/repo` ⏎ `cd "$WT"` ⏎ `myflow stage begin …` | **no** |

**How it got here.** KAN-172's F4 fixed the opposite defect — bare `strings.Contains`, so a command
merely *mentioning* a token counted as its session. F5 then anchored on a leading `myflow` to close a
residual echoed-example gap, and over-anchored. **It traded a false positive for a false negative,
and a false negative here is total and silent.**

Neither the review panel nor the dispatcher's own verification caught it. The verification probe used
four clean invocation shapes, none resembling what is actually emitted — a test that could only pass,
then cited as evidence that over-anchoring was absent.

## What Changes

- **The `myflow` anchor is removed.** A command is a mark when it contains `stage begin`/`stage end`
  **and** binds the token as `-session-token`'s value. Every real shape matches again.

  **This deliberately reopens the echoed-example gap** that two reviewers called Important, and the
  asymmetry is the reason: a false negative means *nothing ever binds*, which is the current state —
  silent and total. A false positive needs a mark-shaped string carrying a currently-pending token,
  and where two sessions match, the ambiguity rule already refuses to bind. Trading a certain total
  failure for a rare narrow one is the right direction, and the residual is documented rather than
  hidden.
- **`/myflow-fast`'s state gate reads the state before it marks.** Marking first auto-creates a
  synthetic `STARTED` record, and `/myflow-fast` accepts only *no state* or `IN_PROGRESS` — so **the
  gate manufactures the state that makes it refuse.** Observed on this change's own creating run.
  A record whose only author is a synthetic mark also stops counting as state for gating.

## Capabilities

### Modified Capabilities

- `myflow-run-telemetry`: what makes a command a stage mark, and the ordering rule that stops a
  mark's own side effect from failing the gate that emits it.

## Impact

**Code** — `stats/internal/harvest/watcher.go` and its tests; `stats/cmd/myflow`'s synthetic-create
path only if the gating rule needs it.

**Skills** — `skills/myflow-fast/SKILL.md`'s state-gate ordering.

**Not changing** — the token mechanism, the binding, the bounded give-up, the ambiguity refusal, the
stage-key vocabulary, or anything else KAN-172 established. This is a matcher and an ordering.
