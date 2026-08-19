# Final review panel — kan-236-base-branch-resolution-retyped-by-hand

**Roster:** `light` — required slots: Primary, Principles (Merged lens), Code review (low).
**Panel model:** sonnet (`models.reviewPanel`), named explicitly on all three slots. No slot was
dispatched by `subagent_type`, so no entry reads `unknown (agent-defined)`.

**Diff-size guard:** `check-panel-diff-size.sh <worktree> db3e25d` measured **539** changed lines
against the cap in force, exit 0 — under cap, so no operator question was raised.

## Optional slots

Four triggers fired against `final-review.diff`, and under the `light` preset all four went into one
multi-select prompt rather than being auto-included:

| Slot | Trigger that fired | Outcome |
|---|---|---|
| Security | path/file handling; a config file changed beyond a version bump | **declined by the operator** |
| Adversarial | a test was modified; >~300 changed lines | **declined by the operator** |
| Lens B — simplicity & state | >~200 changed lines (539) | **declined by the operator** |
| Lens C — robustness & ops | error handling; external integrations | **declined by the operator** |

All four are recorded as **declined**, distinct from a slot whose trigger never fired.

## Pass 1 findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Minor | `scripts/resolve-base-branch.sh:126` | the `HEAD@{upstream}` rationale is written out in full in both the guard header and the contract prose |
| F2 | Primary | Minor | `openspec/changes/kan-236-base-branch-resolution-retyped-by-hand/proposal.md:70` | the Impact table omits `scripts/test-check-guard-symlinks.sh`, which the diff touches |
| F3 | Code review (low) | Major | `skills/myflow-contracts/finish-contract.md:392` | run 2 reads `origin/$BASE` with no assignment in scope, falling through to the `@{upstream}` branch |
| F4 | Code review (low) | Minor | `scripts/resolve-base-branch.sh:92` | the unguarded `branch --show-current` assignment aborts under `set -e` with git's raw status, outside the documented exit contract |

Every finding's state is recorded once, in the single marker block at the end of this record — the
one place a finding's status is written. Look an `F<n>` up there.

## Notes on the findings

**F3 was pre-existing.** `git show db3e25d:skills/myflow-contracts/finish-contract.md` shows `BASE`
already assigned only inside run 1's block before this change. It was fixed here because this
change's own delta spec requires it — the requirement **Every consumer invokes the resolver against
the apply worktree** names run 2 — not because this change introduced it.

**F4's fix chose the explicit refusal.** An unreadable current-branch ref now refuses with exit `2`
("cannot answer") rather than folding into the exit-1 detached-HEAD case, because reporting a
corrupt ref as "detached" states a different and misleading fact, and exit `2` is already how the
script treats a tree it cannot read.

## Fix round 1

`FIX_BASE` per finding: F4 folded into `986bcd7` (tasks 1-2), F3 folded into `1da0842` (task 4), each
via `git commit --fixup=<sha>` then `git rebase --autosquash`. No `fixup!` commit remains; the branch
carries five commits, one per task unit.

**Mutation proof of what the fix round changed.** F4 altered executable behaviour, so its change was
mutation-proved: with the fix reverted, the new case 13 failed (`rc=128`, empty stderr); restored,
all 39 assertions across 13 cases pass. F3 changed contract prose only and falls under the explicit
prose exemption.

## Full re-run 1 findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F5 | Primary | Minor | `skills/myflow-contracts/finish-contract.md:164` | the exit-2 enumeration still lists only the causes that existed before fix round 1 widened it |
| F6 | Code review (low) | Major | `skills/myflow-contracts/finish-contract.md:155` | the deleted fenced block was the guard's hand-run fallback, and three places still promise one exists |
| F7 | Code review (low) | Major | `skills/myflow-contracts/finish-contract.md:404` | a base resolved once is consumed inside a per-worktree loop, so a multi-repo change tests the wrong ref |

## Full re-run 2 findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F8 | Code review (low) | Minor | `skills/myflow-contracts/finish-contract.md:163` | the exit-1 clause and the hand-run fallback each list three of the guard's four exit-1 causes, omitting name validation |
| F9 | Code review (low) | Minor | `scripts/resolve-base-branch.sh:170` | the wrapped fetch runs before the cheap current-branch check that can already doom the call |

findings-total: 14
finding-status: F1 withdrawn — the two texts serve different readers, the contract's caller-facing and the guard header's implementer-facing, and the delta spec's own scenario requires the contract keep the warning
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 withdrawn — the refusal order is deliberate, matching the fenced block this guard replaced, and reordering would change which exit code wins when several conditions hold; the cost withdrawn is one bounded, already-wrapped fetch on a path that was doomed anyway
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 withdrawn — a wording nit in a comment inside a test harness, with no behaviour and no drift risk beyond that file, and the substantive half of the comment (why case 12 pins /bin/bash) is accurate and intact; withdrawn against the cost of a sixth fix round and a seventh full pass

reproducers-total: 14
finding-reproducer: F1 none — a documentation-consistency observation, not a runnable check
finding-reproducer: F2 none — a planning-artifact completeness gap, verified by reading the Impact table against the diff's file list
finding-reproducer: F3 none — reading run 2's procedure shows no assignment between its start and the use
finding-reproducer: F4 none — reproduced by making the current branch's ref unreadable, which the guard rejects as an absolute path
finding-reproducer: F5 none — a documentation-consistency gap between the guard header and the contract paraphrase
finding-reproducer: F6 none — an absence, shown by the contract carrying a hand-run fallback for one guard and not the other
finding-reproducer: F7 none — reproduced with two scratch repositories whose bases differ, which needs paths this record may not carry
finding-reproducer: F8 none — a documentation-completeness gap between the guard header and two contract passages
finding-reproducer: F9 none — reproduced with a git shim logging call order, which needs paths this record may not carry
finding-reproducer: F10 none — a false pointer, verified by reading the guard's header against the body comment that actually holds the rule
finding-reproducer: F11 none — reproduced by comparing a stale local branch against its remote-tracking ref, which needs paths this record may not carry
finding-reproducer: F12 none — reproduced by running the harness as root in a container, which needs an image this record may not carry
finding-reproducer: F13 none — structural duplication, shown by diffing the two helper bodies against each other
finding-reproducer: F14 none — a comment-precision nit, not a runnable defect

**A further candidate was dismissed at full re-run 2.** The `code-review` skill flagged
`.myflow/project.md`'s documented 118.63s guard-test runtime as stale now that a test is added to
that list. The slot measured the suite at 143.96s **without** the new test and 159.18s with it — so
the documented figure was already wrong before this change, and the overrun cannot be attributed to
it. Left unchanged and raised to the operator in the handoff rather than corrected here, since
editing a number that was stale beforehand is outside this change.

**A third candidate was dismissed, not recorded as a finding.** The Code review slot's own
`code-review` skill proposed replacing the guard's charset check with `git check-ref-format
--branch`. The delta spec mandates the narrower rule verbatim, so the implementation is
spec-compliant and `check-ref-format` would be the deviation; the one case where the two disagree
(`..`) was confirmed to fail safe downstream.

## Fix round 2

All three fixes fold into `af7d7d7` (task 4, the finish contract), now `8901740`, via
`git commit --fixup` and `git rebase --autosquash`. The branch still carries five commits and no
`fixup!`.

F5 and F6 are prose and fall under the prose exemption. **F7 changed a documented procedure and was
proved by reproduction:** with two repositories whose bases are `main` and `master`, the old text
resolved `main` from the first and tested the second against `origin/main` — `fatal: Not a valid
object name` — refusing cleanup for a worktree genuinely merged into `origin/master`; the new text
resolves `master` in that worktree and correctly reports it merged.

**The parent separately reconciled the planning artifacts**, which no implementer may touch: the
delta spec's exit-2 row and its unreadable-target scenario, a new scenario for the per-worktree rule
F7 exposed, the requirement text stating that a resolved base belongs to the worktree it came from,
and `design.md`'s order-of-operations step 5 and call-site table.

## Fix round 3

F8 alone. Folded into `8901740` (task 4, the finish contract), now `50f78e6`, via
`git commit --fixup` and `git rebase --autosquash`. The branch still carries five commits and no
`fixup!`. F9 was withdrawn rather than fixed, so `scripts/resolve-base-branch.sh` was not touched in
this round.

F8 is prose and falls under the prose exemption — no mutation was required and no test changed.

## Full re-run 2 aftermath — fix round 4

Three findings, and **two slots raised the same one**: Principles' F10 and the Code review slot's own
F12 are one defect at one location on one theme — the hand-run fallback attributing the
character-shape rule to the guard's header — and are deduped to **F10** by defect identity. The Code
review slot's other two are recorded as F11 and F12.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F10 | Principles + Code review (low) | Important | `skills/myflow-contracts/finish-contract.md:189` | the fallback attributes the character rule to a header that does not state it, and paraphrases it incompletely |
| F11 | Code review (low) | Major | `skills/myflow-contracts/finish-contract.md:15` | nothing tells a caller to compose `<base-ref>` as `origin/$BASE`, so a stale local branch can feed the preflight verdict |
| F12 | Code review (low) | Major | `scripts/test-resolve-base-branch.sh:271` | case 13 `chmod 000`s a ref with no root or packed-ref guard, though its own comment says it needs one |

**F10 had a deeper cause than the false pointer.** The paragraph's precondition is *the script's
absence*, so deferring the operator to any part of that file — header or body — is useless whichever
lines it names. The rule is now stated in the contract in full, and the guard's header gained the
shape too, so the header is genuinely the authority for a reader who does have the script.

**F11 is the third instance of one defect class in this change** — a consumer of the resolver reading
the wrong ref — after F3 (run 2 had no `BASE` in scope) and F7 (one `BASE` reused across worktrees).
All three were found by review rather than by the design, which is worth knowing about this change's
blast radius.

**F12 was proved on both arms rather than argued.** The fix round ran the harness as root in an
`alpine:3.20` container and saw case 13 skip instead of assert, and probed the packed-ref arm
separately. The parent had flagged the packed-ref condition, which the repository's own `skip()`
helper documents alongside the root one and which the finding named only in part.

## Fix round 4

F10's script half and F12 fold into `60218c8` (tasks 1-2), now `e0fcc8e`; F10's contract half and
F11's contract half fold into `50f78e6` (task 4), now `cf8acbc`; F11's skill half folds into
`fcc56ec` (task 5), now `e6f4f3f`. `skills/myflow-status/SKILL.md` needed no change — it never
invokes the preflight and already composed `origin/<base>` correctly. Five commits, no `fixup!`.

F10 and F11 are prose. **F12 changed executable test code and was proved on both arms**, as recorded
above.

## Full re-run 3 — after fix round 4

Primary and Code review both returned **clean**. Principles returned one Minor.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F13 | Principles | Minor | `scripts/test-resolve-base-branch.sh:52` | two harness helpers duplicate the capture logic, so a fix to it has to be applied twice |

**Principles judged F10's remaining duplication justified rather than a violation**, adding a second
argument the fix round had not made: the delta spec is a planning artifact this project archives once
the change ships, so pointing the long-lived contract at it would rot on archive — inlining beats
citing on that ground independently of the absent-script one.

**F13 was fixed rather than withdrawn** because the duplicated helpers are code *this change
authored*, so cleaning them is inside its scope rather than an unrelated wart.

## Fix round 5

F13 alone, folded into `e0fcc8e` (tasks 1-2, which owns the harness). It changes executable test
code, so the round had to prove that case 12 still genuinely runs under `/bin/bash` — a merge that
silently stopped pinning the interpreter would leave the locale defect unproved while the case still
went green.

## Full re-run 4 — after fix round 5, the closing pass

**Primary: clean. Code review (low): clean. Principles: one Minor, F14, withdrawn by the operator.**

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F14 | Principles | Minor | `scripts/test-resolve-base-branch.sh:55` | the comment calls `GUARD_INTERP` the same style as `PATH=`/`LC_ALL=`, but those are read by the exec'd process while this one needs a branch inside the function |

**Both clean slots proved the interpreter pin rather than reading for it**, which was the one property
fix round 5 could have silently broken: Primary traced `bash -x` and saw `++ /bin/bash …`, and the
Code review slot ran a `$BASH_VERSION` probe through the same branch — 5.3.15 unpinned, 3.2.57
pinned. Both independently confirmed a temporary assignment on a shell-function call does not leak to
later calls, so case 13 is unaffected.

**The Code review slot's own `code-review` skill rediscovered a settled item and it declined to
re-raise it** — the charset validation admitting git-invalid sequences such as `..`, dismissed at
full re-run 2 because the delta spec mandates the narrower rule verbatim. The skill is launched
without the panel record's memory, so rediscovery is expected; catching it is what stops a settled
argument restarting.

## Panel outcome

**Fourteen findings across six passes and five fix rounds: eleven fixed, three withdrawn by the
operator (F1, F9, F14).** No finding is open at any severity, and the closing pass is non-stale clean
for Primary and Code review, with Principles' only open item withdrawn.

**Three of the fourteen were one defect class** — a consumer of the resolver reading the wrong ref
(F3, F7, F11). All three were found by review rather than by the design, which is the most useful
thing this panel recorded about the change.

## Re-run mode

**Full after every fix round**, escalated automatically rather than chosen. Round 1 altered a
guard's behaviour (`resolve-base-branch.sh` gained an exit-2 path); round 2 altered a delta spec;
and by round 3 three fix rounds had run. Each is its own clause on the escalation ladder — the third
fires on the count alone, independently of what round 3 changed, which was prose only. Every **required** slot re-runs against a rewritten
`final-review.diff`. No conditional slot re-runs — all four were declined by the operator at pass 1,
and a declined slot is not re-evaluated.
