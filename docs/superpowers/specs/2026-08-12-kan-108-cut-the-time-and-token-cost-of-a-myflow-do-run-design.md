# Cut the time and token cost of a `/myflow-do` run — design

**Jira:** KAN-108
**Date:** 2026-08-12

## Problem

KAN-108 records a measurement taken on the KAN-107 run, which spent roughly 3.44M subagent tokens.
The review panel accounted for ~1.82M of that (53%), the eight implementers plus their fix rounds
~1.06M, and the per-task reviewers ~0.56M. The ticket isolates two structural causes, ranked by
measured saving.

**The escalation trigger `the fix altered a public contract` is vacuous in this repository.** This
repository *is* contracts, so the clause fires on every fix round and never discriminates, forcing a
full re-run each time. Pass 3 of the KAN-107 run cost ~430k tokens; a targeted re-run of the primary
slot plus the one slot holding an open finding would have caught the same two defects at ~190k.

**A wrong fix instruction is dispatched as readily as a right one.** Two findings bounced for three
rounds each because the instruction itself was wrong. One told the fixer to pass an absolute path
that the guard's name allowlist rejects before any path is built, so the repaired test still proved
nothing. The other accepted a fixture teardown that ran before the guard was invoked, leaving the
case a duplicate of its neighbour. Both instructions would have been caught by running the check they
described and observing that it passed.

KAN-109, which landed on 2026-08-12, touched the same section of `skills/myflow-do/SKILL.md` but left
both of these open: its proposal states that the escalation ladder's trigger conditions are untouched,
and it added no verification step ahead of the fix dispatch. What KAN-109 did change is what a *Full*
escalation costs once it fires — a conditional slot now re-runs only when its own trigger still fires
against the fix diff — so part of the first item's measured gap is already collected. The remaining
gap is that the trigger forces Full rather than Targeted on every fix round in this repository.

## Decisions

### The vacuous escalation clause is reworded, not dropped

**ID:** reword-not-drop
**Status:** active
**Chosen:** Replace `a public contract` with `a guard's behaviour` in the auto-escalate trigger list —
the trigger keeps a real coverage reason (a fix that changes a script's exit codes or output) while
ceasing to fire on every contract edit in a repository made of contracts.
**Considered:**
- *Drop the clause entirely*, leaving delta spec and migration alone — cheapest, but a fix that
  changes a guard's behaviour without touching a spec would then re-run Targeted only, which is a
  real coverage loss in the one repository where guards are the product.
- *Define a contract path set per project*, declared in `.myflow/project.md`, so `public contract`
  becomes mechanically checkable — more machinery than the saving justifies, and it puts an
  attacker-editable path list in front of an escalation decision.

### A fix instruction must be verified by a failing reproducer before dispatch

**ID:** failing-reproducer-required
**Status:** active
**Chosen:** Before the fix subagent is dispatched, the parent runs each finding's reproducer command
in the worktree and requires a non-zero exit. A reproducer that passes means the instruction proves
nothing, which is exactly the defect both KAN-107 failures had.
**Considered:**
- *A parent self-check with no command* — the parent re-reads each instruction against the file it
  names and records why it is sound. Cheap and adds no machinery, but it is the same judgment that
  failed for three rounds on each of the two findings.
- *A cheap verifier subagent* over the whole instruction list — adds a dispatch to save fix rounds,
  and reintroduces the judgment problem one level out, where the verifier has less context than the
  slot that raised the finding.

### A finding with no runnable reproducer is exempt, with its reason stated

**ID:** exempt-with-reason
**Status:** active
**Chosen:** The slot writes `none — <why>` in place of a command. The parent dispatches such a finding
without a run. The rule therefore bites only on findings claiming a mechanical defect, which is where
both measured failures were.
**Considered:**
- *Exempt only named categories* (principles, prose, naming) — tighter, but the category list becomes
  a maintained artifact and an argument surface, for no measured gain.
- *The parent constructs a reproducer when one is missing* — strongest coverage, and the most parent
  work; it also relocates the failed judgment into the parent rather than removing it.

### A reproducer that passes bounces once, then reaches the operator

**ID:** one-bounce-then-operator
**Status:** active
**Chosen:** The finding goes back to the slot that raised it, carrying the reproducer's passing
output. If the slot's second reproducer also passes, the run stops and puts the finding to the
operator through the handback prompt already in section 5 — take another round, withdraw with a
reason, or stop the run.
**Considered:**
- *Unbounded bounces to the same slot* — simple, and a slot stuck on a wrong finding spins with no
  exit.
- *No bounce; straight to the operator* — fastest exit, but it discards the cheap case where the slot
  merely wrote the wrong command and would correct it in one retry.

### A bounce is guard-class and consumes no fix round

**ID:** bounce-is-guard-class
**Status:** active
**Chosen:** A bounce does not count toward the escalation ladder's fix-round threshold, following the
precedent section 4 already sets for `check-task-commit-fields.sh`: a guard failure is not a review
finding, so it does not consume one of the review loop's fix-round slots.
**Considered:** *Counting a bounce as a round* — it bounds a slot producing unreproducible findings,
but escalates the panel to Full for what is a reviewer defect rather than a risk in the code, which
spends exactly the tokens this change exists to save.

### The reproducer field is enforced by a guard, not by prose alone

**ID:** guard-over-panel-record
**Status:** active
**Chosen:** A new `scripts/check-panel-reproducers.sh` reads the panel record and fails when a finding
carries no reproducer line at all. This repository's own history is that an unguarded rule drifts, and
the ticket's complaint is precisely a rule nobody checked.
**Considered:**
- *Prose in `SKILL.md` alone* — cheaper to land and unenforced.
- *Also guarding the reworded escalation trigger* — `a fix altered a guard's behaviour` is not
  mechanically decidable from a diff, so such a guard could only check that the record states a
  verdict, not that the verdict is right; that is a check that reads as enforcement without being it.

### A finding whose slot supplied no reproducer is recorded as such and dispatched

**ID:** not-supplied-dispatches-unverified
**Status:** active
**Chosen:** The parent writes `none — not supplied by <slot>` and dispatches the finding unverified.
The omission is visible in the panel record, and the guard passes because the field is present.
**Considered:**
- *Re-asking the slot once* before accepting the not-supplied form — one cheap extra exchange, and
  one more state in a loop that already has a bounce in it.
- *Rejecting the not-supplied form in the guard* — strongest, and it can wedge a run on the output of
  a third-party agent whose definition this pipeline does not control.

### The reproducer lines form their own marker block

**ID:** separate-marker-block
**Status:** active
**Chosen:** Reproducer lines sit in a block of their own, with their own count line, separate from the
`finding-status:` block.
**Considered:** *A `Reproducer` column in the findings table* — the table exists for the reader and no
guard parses it, so a guard-read field does not belong there. *Interleaving the lines with the
`finding-status:` markers* — `check-unfinished-work.sh` requires those markers to occupy one unbroken
span of lines, so interleaving breaks a guard that currently passes.

## Open questions

*(none — every question this design raised was answered during brainstorming)*

## What changes

### 1. The escalation trigger

`skills/myflow-do/SKILL.md`, in the auto-escalate sentence under **Panel re-runs**: `a public
contract` becomes `a guard's behaviour`. The resulting set is — the fix touched a file outside the set
named in the findings; the fix diff exceeds ~150 changed lines; the fix altered a delta spec, a
migration, or a guard's behaviour; a targeted re-run surfaced a new Critical finding; three or more
fix rounds have already run.

That set currently appears in no capability spec, so it also becomes a requirement under
`myflow-review-panel-economics`. Without one, the next reword has nothing to check it against.

### 2. Reproducer-verified fix dispatch

Every slot's dispatch prompt gains a reproducer brief: for each finding, supply either a runnable
command that demonstrates the defect, or `none — <reason>`. Slots dispatched by `subagent_type`
(Bugbot, Security) receive it as prompt text, the same way Bugbot already receives the
mutation-testing brief — no agent definition is edited.

The panel record `.superpowers/sdd/final-review-panel.md` gains a second marker block:

```text unverified:the format this change introduces; no panel record carries it until the change lands
reproducers-total: 2
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh
finding-reproducer: F2 none — prose-only, no runnable check
```

Before dispatching the fix subagent, the parent runs each supplied reproducer in the worktree and
requires a non-zero exit. On exit 0 the finding is bounced once to the slot that raised it, carrying
the passing output; a second passing reproducer stops the run at the existing operator handback. A
bounce consumes no fix round. Findings recorded `none — <reason>` are dispatched without a run.

### 3. The guard

`scripts/check-panel-reproducers.sh <worktree>`:

- exit 0 — every `F<n>` named in the `finding-status:` block has exactly one well-formed
  `finding-reproducer:` line, and `reproducers-total` equals the number of those lines;
- exit 1 — violations found, each reported with its line;
- exit 2 — cannot answer (no record, unreadable path).

It reads the marker block through anchored regexes and never parses the findings table, matching
`check-unfinished-work.sh`'s construction. `scripts/test-check-panel-reproducers.sh` covers it and
joins `.myflow/project.md`'s `## test` list. Neither joins `## lint`: both need a change in flight and
a real worktree, the same reason `check-panel-diff-size.sh` is excluded.

## Impact

- `skills/myflow-do/SKILL.md` — the trigger clause, the reproducer brief on the slot dispatches, the
  reproducer marker block, the pre-dispatch run, and the bounce rule.
- `skills/myflow-do/principles-reviewer-prompt.md` and
  `skills/myflow-do/adversarial-reviewer-prompt.md` — the reproducer requirement.
- New `scripts/check-panel-reproducers.sh` and `scripts/test-check-panel-reproducers.sh`.
- `.myflow/project.md` — `## test` gains the new harness.
- `openspec/specs/myflow-review-panel-economics/` — three new requirements, via the change's delta
  spec.
- No change to the state file shape, the three-state machine, the roster presets, the optional-slot
  trigger table, the panel record's findings table or `finding-status:` rules, the zero-open-findings
  bar, or any `/myflow-start` / `/myflow-finish` / `/myflow-status` behaviour. `/myflow-fast` inherits
  everything through the `/myflow-do` sections it cites, with no edit of its own.
- No budget re-anchor: `skills/myflow-do/SKILL.md` is at 46,899 bytes against a 58,623-byte budget.
