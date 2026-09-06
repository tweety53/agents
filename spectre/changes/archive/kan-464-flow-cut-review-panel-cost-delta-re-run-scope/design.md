# Design — cut review-panel cost

## Context

Three edits to `skills/flow/review-panel.md`, one of them carrying a new required dispatch
paragraph into `scripts/check-dispatch-paragraphs.sh` and its harness. Why each is worth making is
`proposal.md`'s; the brainstorming round that chose between the options is
`docs/superpowers/specs/2026-09-06-kan-464-flow-cut-review-panel-cost-delta-re-run-scope-design.md`.

## Part 1 — which slots re-run after a fix round

In **Panel re-runs**, the paragraph beginning "**When the round raised anything above Minor,
re-run on deltas**" keeps its opener, its per-slot-per-worktree held-sha rule and its delta-path
definition unchanged. Its three bullets become:

- **a slot re-runs only when it raised a finding in the previous round, or the previous round
  raised a new Critical** — every slot in the resolved roster, Primary included, and every
  operator-added slot already dispatched in an earlier pass of this run, on that one rule. A slot
  that raised nothing keeps the result it has; the round's own mutation-proof (below) covers what
  the fix changed;
- **a diff-reading slot that re-runs reads its delta**; Bugbot, Mutation and Security read no diff
  file and re-run in their pass-1 shape, throwaway worktree included. **A diff-reading slot whose
  delta is empty in every worktree is not dispatched**, and the record states `not re-run —
  nothing new since its last read`;
- **a slot the operator has not named for this run is never added here** — unchanged, verbatim.

The first bullet is today's Bugbot/Mutation/Security clause, widened; the "every diff-reading
slot … re-runs on its delta" bullet is deleted, and what it said about Primary, about
operator-added slots and about an empty delta is carried into the two bullets above rather than
lost.

**The stale-result carve-out widens with it.** The sentence "A non-Minor fix Bugbot or Mutation
did not raise leaves that slot's result current: the round's own mutation-proof (below) covers
what the fix changed" becomes "A fix against which a slot raised no finding leaves that slot's
result current: the round's own mutation-proof (below) covers what the fix changed." This edit is
not optional decoration on Part 1: the preceding sentence fails a slot's clean result when "any
commit or working-tree change to source landed after that slot's last read", so without the wider
carve-out the handoff blocks on precisely the slots the new rule declines to re-run.

Untouched by Part 1, and deliberately: the Minor-only rule that triggers no re-run at all, the
docs-only reclassify clause, `-diff-base`, and the re-run cap check — the latter measures once per
worktree per **distinct** held sha among the diff-reading slots dispatched this round, which needs
no edit to mean the right thing when fewer slots are dispatched.

## Part 2 — the fix subagent performs the mutation-proof

### The section

**The fix round mutation-proves what it changed** keeps its opening two lines ("**This binds the
review panel's fix round.**" and "**Every executable behaviour the fix changed is mutation-proved,
not only the test cases the round adds.**") and states that the **fix subagent** performs the
proof and reports it, per the MUTATION PROOF paragraph its dispatch carries; the parent runs no
build of its own here. A survivor the fix subagent cannot judge real or equivalent goes to the
operator through the same handback the section already names.

The `fix-mutation:` / `fix-mutations-total:` fenced block stays exactly as it is, with the parent
transcribing the fix subagent's reported lines into the pass log entry rather than producing them;
"**These lines go in the pass log entry and never inside the marker block**" and the
marker-label ban paragraph after it are unchanged.

The section's operational sentences — mutate the mechanism, one mechanism per mutation, split a
revert that would change state a second check reads, `scripts/mutate-and-verify.sh` mechanizes
backup/apply/run/report/restore, a surviving mutation is repaired in this round — **relocate into
the MUTATION PROOF blockquote below**, which is the fix subagent's own copy of them. They are
moved, not dropped; the section no longer restates them.

The closing sentence "This binds the fix round every run — the obligation is the round's, not a
slot's …" stays verbatim: with Part 1 cutting re-runs it is more load-bearing than before, not
less.

### The parent's check

"**The parent checks the reported list against the fix diff before the round can close**" and its
hunk walk are unchanged, including the clause on a hunk that removes or weakens a test. One
sentence is added to it: **the same walk holds the fix subagent to its PLAN FIELDS obligation** —
a hunk that adds a test case, adds a file, or changes what a task's `**Baseline:**` counts, whose
task's `**Tests:**`, `**Baseline:**` or `**Files:**` field in `<changeRoot>/tasks.md` does not
reflect it, does not close the round; it goes to the handback. With Part 1 no longer re-running a
slot that raised nothing, this walk is where a stale plan field is caught — four of the seven
re-run findings across the gymie changes were exactly that, raised by `primary`.

### The paragraph

A new blockquote joins the fix subagent's dispatch paragraphs, beside PLAN FIELDS, FOREGROUND
BUILDS, TARGETED TESTS and REPORT FILE:

> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you
> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a
> scratch tree, or flip the single value it turns on, confirm an existing test fails, and restore.
> `<agents repo>/scripts/mutate-and-verify.sh` mechanizes backup, apply, run, report and restore
> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to
> mutate is your judgment, not the script's. Each mutation alters one mechanism — where a single
> revert would also change state a second check reads, split it into surgical mutations, one per
> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before
> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a
> `fix-mutations-total:` count, in the shape the review-panel contract's fenced block gives. Where
> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather
> than deciding it yourself.

The fix subagent's REPORT FILE paragraph gains those lines in what it names, alongside the
findings addressed, the behaviours changed and the task commit each fixup folded into.

### The guard row

`scripts/check-dispatch-paragraphs.sh` gains one entry, following the shape KAN-217, KAN-263 and
KAN-441 each used:

| Array | Value |
|---|---|
| `ENTRY_LABEL[mutation]` | `**MUTATION PROOF:**` |
| `ENTRY_SHARED_PHRASES[mutation]` | `mutation-proved before you end your turn`, `confirm an existing test fails, and restore`, `a surviving mutant` |
| `ENTRY_VARIANTS[mutation]` | empty — every block carrying the label is equivalent |
| site | `skills/flow/review-panel.md`, min 1 block, no variants |

and its header's paragraph table and per-entry phrase notes gain the matching rows. The guard's
own machinery is untouched: this is a table row, which is what that generalization was for.

`scripts/test-check-dispatch-paragraphs.sh` gains a `MUTATION_PROOF_BLOCK` constant plus three
one-phrase-dropped variants, adds the block to every `review-panel.md` fixture the harness asserts
clean, and adds four cases: the label absent from `review-panel.md`, and one per required phrase
dropped. Fixtures written deliberately deficient for another label are left alone — a case
asserting a REPRODUCE, DON'T READ failure does not need this block.

## Part 3 — a reviewer may write its reproducer as a script

The paragraph "**Every slot must supply, per finding, a reproducer**" keeps its two forms (a
runnable command, or the literal `none — <reason>` exemption) and gains the script form:

- a demonstrating command needing a pipe, a quote, a glob or any other shell metacharacter is
  written as a script rather than abandoned — the guards refuse a metacharacter in the recorded
  line, never one inside a script;
- the slot writes it to `<abs-worktree>/.superpowers/sdd/reproducers/<round>-<id>-<n>.sh` —
  `<round>` this round's number, `<id>` the slot's own resolved reviewer id, `<n>` that slot's own
  1-based finding index — gives it a shebang and `chmod +x`;
- it records `.superpowers/sdd/reproducers/<round>-<id>-<n>.sh`, the path **relative to the
  worktree**: `run-reproducer.sh` refuses an absolute token, and requires the resolved path to be
  a regular file inside the worktree with execute permission;
- **Bugbot and Mutation write theirs into the canonical worktree**, never their own
  `<worktree>-<slot>-<round>` copy, which is removed the moment their dispatch closes;
- the parent records the path the slot supplied verbatim — there is no rename step, which is why
  the filename is slot-and-round derived rather than `<ref>.sh`.

The paragraph's existing closing instruction — carry this requirement on every slot's dispatch
prompt — carries the script form with it. No guard changes: `check-panel-reproducers.sh` is purely
lexical and does not check existence, `run-reproducer.sh` accepts a relative, contained,
executable path, and `.superpowers/` is gitignored, so no reproducer script is ever committed.

## Budget

`skills/flow/review-panel.md` measures about 47.7 KB against the 58732-byte row
`scripts/check-contract-budget.sh` declares for it; Part 1 is net deletion and Parts 2 and 3 add
roughly 2 KB between them. No budget row moves.

## Decisions

### Where a reviewer-written reproducer script lives, and what names it

**ID:** repro-script-slot-round-name
**Status:** active
**Chosen:** `.superpowers/sdd/reproducers/<round>-<id>-<n>.sh`, written by the slot and recorded
verbatim — unique by construction across slots and rounds, and it costs the parent nothing.
**Considered:** the research note's `<ref>.sh`, which a reviewer cannot produce because `F<n>`
refs are assigned by the dispatcher after every slot has reported — it would need a parent rename
plus a rewritten reproducer field per finding, which is the conductor glue this change exists to
cut; and a free-form path under `reproducers/`, rejected because every slot writes into the one
canonical worktree and two slots choosing the same name would collide silently.

### What makes a slot re-run after a fix round

**ID:** rerun-clause-reused-for-every-slot
**Status:** active
**Chosen:** the clause already in the file for Bugbot, Mutation and Security — raised a finding in
the previous round, or the previous round raised a new Critical — widened to every slot, so lever
A is a deletion plus a widening with no second rule to keep in sync.
**Considered:** re-running only slots whose finding the round actually *fixed*, matching the
research note's wording, rejected for the extra qualifier and the withdrawn-finding case it forces
a reader to reason about, at the cost of one unnecessary delta dispatch in a rare case; and
dropping the new-Critical disjunct as well, rejected because it removes the one trigger that makes
an unrelated slot re-read a branch after a severe defect was patched.

### The stale-result carve-out widens with the re-run rule

**ID:** stale-carveout-generalised
**Status:** active
**Chosen:** generalise "A non-Minor fix Bugbot or Mutation did not raise leaves that slot's result
current" to any slot that raised no finding against the fix, in the same commit as the re-run rule.
**Considered:** leaving the stale-result definition alone, rejected as incoherent — the handoff
requires no stale result, so every slot the new rule declines to re-run would block the handoff and
the change would deliver nothing.

### Who performs the fix round's mutation-proof

**ID:** fix-subagent-owns-mutation-proof
**Status:** active
**Chosen:** the fix subagent mutates, repairs a survivor and reports; the parent transcribes the
reported lines and keeps only its existing fix-diff hunk walk, running no build.
**Considered:** leaving the proof on the parent and cutting cost elsewhere, rejected because it is
the single largest conductor cost measured (seven minutes of Gradle on a 200k+ token context in
kan-455); and splitting it — the subagent mutates, the parent re-runs each mutation to confirm —
rejected as paying the conductor cost twice for a check the hunk walk already approximates.

### The plan-field net moves into the parent's hunk walk

**ID:** parent-walk-checks-plan-fields
**Status:** active
**Chosen:** one sentence added to the walk the parent already performs, holding the fix subagent to
the PLAN FIELDS obligation its dispatch already carries.
**Considered:** keeping `primary` re-running purely to catch stale plan fields, rejected because
that is the cost Part 1 exists to remove and the walk reads the same diff anyway; and trusting the
PLAN FIELDS paragraph alone, rejected because kan-454 and kan-455 both show it ignored with no
check behind it.

### The MUTATION PROOF paragraph is guarded like every other required one

**ID:** mutation-proof-paragraph-guarded
**Status:** active
**Chosen:** add the table row in `check-dispatch-paragraphs.sh` and its harness cases in the same
change that adds the paragraph.
**Considered:** folding the instruction into the existing REPORT FILE paragraph, which is
unguarded, rejected because KAN-289's root cause was exactly an instruction living in no table; and
adding the paragraph without a guard row, rejected for the same reason — with Part 1 cutting
re-runs, the round's own proof is the only mutation reasoning some runs will do at all.

## Open questions

None. Both rounds' questions were answered by the operator and are recorded above.
