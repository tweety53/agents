# Review panel — kan-107-remove-manual-test-guide-and-gate

## Pass 1 — full roster

Diff read by every slot: `.superpowers/sdd/final-review.diff` (44 files, 247 insertions, 2290
deletions against merge base `58bff0d`).

| Slot | Ran? | Model | Why |
|------|------|-------|-----|
| 0 Primary | yes | Sonnet | required |
| 1 Bugbot | yes | unknown (agent-defined) | required |
| 2 Principles — Merged | yes | Sonnet | required |
| 3 Security | yes | unknown (agent-defined) | selected: the change edits guard scripts that build paths from a change name, and a traversal fixture was edited |
| 4 Adversarial | yes | Sonnet | selected: test cases were deleted, and the diff exceeds ~300 changed lines |
| 5 Principles — Lens B, simplicity & state | yes | Sonnet | selected: the diff exceeds ~200 changed lines |
| 5 Principles — Lens C, robustness & ops | yes | Sonnet | selected: the unfinished-work guard's error paths and exit contract were touched |

Slots 1 and 3 were dispatched by `subagent_type` and carry their own agent definitions; no model
override was passed to either, so their model is recorded as `unknown (agent-defined)`.

### Findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Bugbot | Important | `scripts/check-unfinished-work.sh:177` | `ids_of`'s comment still says "see signal three"; the signals were renumbered to one and two everywhere else in the file |
| F2 | Bugbot | Important | `scripts/gather-self-review-context.sh:101` | The header's source list still names the `"plan, test guide and session records"` commit, while the body comment and both greps carry the new subject |
| F3 | Primary, Bugbot | Important | `.gitignore:6-8` | The `.worktrees/` ignore entry is undocumented in the plan, and contradicts `skills/myflow-contracts/finish-contract.md:309`, which states this repository keeps its worktrees in a sibling directory |
| F4 | Lens C, Bugbot | Important | `skills/myflow-contracts/handoff-blocks.md:126-129` | `Run it:` is declared on-disk and `/myflow-status` is said to regenerate it, but `skills/myflow-status/SKILL.md` carries no resolution procedure and never cites the one in `/myflow-do` section 6 |
| F5 | Bugbot, Adversarial | Important | `scripts/test-check-unfinished-work.sh` | Old case 8d was deleted with the guide, leaving `unreadable()` and `count_matching`'s `rc > 1` discipline with no test at all |
| F6 | Bugbot, Adversarial | Important | `scripts/test-check-unfinished-work.sh` | Old case 8 exercised two signals firing at once; nothing replaced it for the two surviving signals, while both file headers still assert the behaviour |
| F7 | Bugbot, Security, Primary, Lens B, Adversarial | Critical | `scripts/test-check-unfinished-work.sh` | Case 8f passed a name the guard could not resolve, so it proved nothing beyond the allowlist. Pass 1's fix matched the planted tree to the name but left the name absolute, which fails as `OUTSTANDING` rather than `CLEAR`; reopened on pass 2. Pass 2's fix switched to a relative traversal name and dropped the unreachable planted panel record. Verified on pass 3 by four slots, three of which reproduced the mutation independently |
| F8 | Lens C | Minor | `skills/myflow-do/SKILL-rationale.md:121` | Heading still reads "Why the three planning paths are left unstaged" |
| F9 | Adversarial | Minor | `scripts/gather-self-review-context.sh:470` | The `IMPL_SHA` exclusion matches the plan-commit phrase as an unanchored substring, so an implementation commit whose subject contains it is wrongly excluded. Pre-existing in shape, but the diff touches the line |

### Raised and not carried as findings

Lens B noted that the two-wording alternation is necessarily written at both grep sites, since Bash
offers no shared pattern variable, and judged the cross-referencing comments an adequate mitigation.
Recorded here rather than as a finding, at that slot's own recommendation.

Security noted that dropping `docs/manual-test/` from the exclusion pathspec means a leftover file
at that path in a consumer repository would now be staged into a checkpoint. That is the intended
consequence of the change: the path no longer exists in any repository the pipeline manages.

## Pass 2 — full roster (escalated)

Mode: **full**, not targeted. Escalated automatically because the fix wave altered
`skills/myflow-contracts/handoff-blocks.md`, the canonical definition of the `IN_PROGRESS` handoff
template and this change's public interface. The fix diff itself was 105 changed lines, under the
size trigger, and no new Critical finding was raised — the contract edit is the sole reason.

Diff read by every slot: `.superpowers/sdd/final-review.diff`, rewritten against merge base
`58bff0d` after the fix wave (44 files, 337 insertions, 2295 deletions). The fix-only diff is at
`.superpowers/sdd/fix-round-1.diff`.

### Pass 2 findings

F7 is **reopened**: the pass-1 fix changed what case 8f planted but not the shape of the name it
passes, so the case still cannot reach the containment logic. Four slots examined it and their
accounts reconcile — Security's mutation shows the case is load-bearing (removing the allowlist
makes it fail), while Primary, Bugbot and Adversarial show it fails for the wrong reason: with an
**absolute** planted name the guard concatenates a nonsense path and reports `OUTSTANDING`, never
`CLEAR`. Bugbot settled it by running both shapes against a mutated guard: the absolute name yields
`OUTSTANDING: … no plan at …`, while a **relative** name yields `CLEAR: … every plan item is
checked`. So a relative traversal name is what makes the case demonstrate containment rather than
path-concatenation garbage.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F10 | Bugbot | Important | `skills/myflow-contracts/plan-provenance.md:279,285` | The counts `Twelve lines` / `Eleven are counter-examples` / `(eleven lines)` were left unchanged while the source list above them was rewritten. The deleted guide carried three such lines. Bugbot's re-derivation gives 10 and 9; the fix round's own claim of 12 and 11 disagrees, so the number must be re-derived with the guard's real functions rather than either estimate |
| F11 | Bugbot | Minor | `skills/myflow-contracts/plan-provenance.md:285` | The table row names only "this file and the delta spec" as homes for the counter-examples, while `check-plan-provenance.py:591-593` also names archived copies of either — and three of them live in `openspec/changes/archive/2026-07-30-kan-20-…/` |
| F12 | Lens C | Minor | `skills/myflow-status/SKILL.md:219-224` | The citation restates the resolution's two deltas but drops `/myflow-do`'s own "never the project's declared base" — the single invariant the citation exists to protect |
| F13 | Bugbot | Minor | `scripts/test-check-unfinished-work.sh:196` | `rm -rf "$WT/docs/manual-test"` is a no-op: `new_fixture` never creates that directory, so case 2 is byte-identical in effect to case 1 |


Pass 2 also confirmed, by mutation rather than inspection: F5's restored read-failure case fails
when `count_matching`'s `rc > 1` discipline is disabled, F6's combined case fires two structurally
separate signals, and F9's new case fails against the unanchored pattern.

### Raised and not carried as findings, pass 2

Lens C observed that nothing forces `/myflow-status` to load the cited section, so a skipped load
fails silently. That is a property of every citation in this prose-driven skill set rather than a
regression from this change, and the slot itself declined to press it.

Adversarial observed that case 5b's `chmod 644` restore is a following statement rather than a trap,
and then verified no cross-case contamination is possible: each case gets a fresh `mktemp` fixture,
and `rm -rf` unlinks a mode-000 file without needing read permission.

## Pass 3 — full roster (escalated)

Mode: **full**, not targeted. Fix wave 2 touched `scripts/test-check-unfinished-work.sh`,
`skills/myflow-contracts/plan-provenance.md` and `skills/myflow-status/SKILL.md` — files that sit in
every lens's territory, so no slot's pass-2 clean result is still fresh. A targeted re-run would have
handed off on stale clean results from Principles, Lens B and Security.

F10 was settled in that wave by re-derivation with the guard's own imported functions rather than a
reimplementation: 12 hits, 11 counter-examples, confirming the existing prose and overturning pass
2's lenient estimate of 10 and 9. No numeric text changed.

Diff read by every slot: `.superpowers/sdd/final-review.diff`, rewritten against merge base
`58bff0d`. The fix-only diff is at `.superpowers/sdd/fix-round-2.diff`.

### Pass 3 findings

Pass 3 confirmed F7 and F10 as genuinely closed — three slots reproduced F7's mutation
independently, and two re-derived F10's counts with the guard's own imported functions, both
arriving at 12 and 11. Bugbot withdrew its pass-2 count of 10 and 9 as the product of a lenient
reimplementation rather than the guard.

Two duplicate rows for F7 were merged by the controller into the single row above, and the reopened
status folded into its note: the guard requires each `F<n>` to name exactly one finding, and the
second row made the panel record itself report `OUTSTANDING`.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F14 | Bugbot | Important | `openspec/changes/kan-107-remove-manual-test-guide-and-gate/design.md:191`, `docs/superpowers/specs/2026-08-08-kan-107-remove-manual-test-guide-and-gate-design.md:177` | The design states the guard keeps exit codes `0` clear, `1` outstanding, `2` cannot determine. The guard never exits 1 — it prints `OUTSTANDING:` and exits 0, because the verdict line carries the answer and the exit status does not. The word "unchanged" is wrong too: it was never that. A caller written from this text reads every `OUTSTANDING` as clear |
| F15 | Bugbot | Minor | `skills/myflow-contracts/plan-provenance.md:286`, `scripts/check-plan-provenance.py:588-596` | The row names "this file, the delta spec, or an archived copy of either" as the homes of the eleven counter-examples, but two of them are neither: an SDD ledger at `docs/superpowers/ledgers/2026-07-30-kan-20-…:22`, and an archived design document at `openspec/changes/archive/2026-07-30-kan-20-…/design.md:112` |
| F16 | Bugbot, Adversarial | Minor | `scripts/test-check-unfinished-work.sh:192-202` | F13's fix does not work. The `rm -rf` runs before `run_guard`, so the guard sees a tree identical to case 1's. The "once had a guide" state leaves no trace the guard could observe, so the case still costs a fixture and proves nothing its neighbour does not |
| F17 | Lens B | Important | `skills/myflow-status/SKILL.md:222` | F12's added clause restates a rule stated verbatim at `skills/myflow-do/SKILL.md:338`, two lines above a sentence forbidding exactly that. Lens C and Principles both judged the clause necessary and the citation still a pointer, so the defect is the contradiction rather than the clause |

findings-total: 17
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 fixed
finding-status: F16 fixed
finding-status: F17 fixed

F13 is reopened: F16 is the same defect under a second diff, so the two are one finding and F13
carries it.

## Pass 4 — full roster (escalated)

Mode: **full**. The trigger is the escalation ladder's own "three or more fix rounds have already
run", which fires on this round regardless of what the fix touched. Three rounds without convergence
is what the rule is for, and the panel follows it rather than reasoning its way past a
non-convergence guard on cost grounds.

Diff read by every slot: `.superpowers/sdd/final-review.diff`, rewritten against merge base
`58bff0d`. The fix-only diff is at `.superpowers/sdd/fix-round-3.diff`.

### Pass 4 result — clean

All seven slots reported zero open findings. Every pass-3 fix was verified against the code rather
than against another document: the exit-code contract read from `scripts/check-unfinished-work.sh`'s
own header and exit paths and cross-checked against all three callers, the counter-example homes
re-derived twice more with the guard's imported functions and matching to the line, case 2 confirmed
load-bearing by two slots, and `scripts/check-plan-provenance.py` proven comment-only by token
stream — 4221 tokens on each side of the merge base.

Handoff condition met: zero open findings at any severity, with a non-stale clean result from every
slot in the roster on the same diff.

### Dissent recorded, not carried as a finding

Lens B maintains that F17's clause "never the project's declared base" in
`skills/myflow-status/SKILL.md` is a literal duplicate of `skills/myflow-do/SKILL.md`'s wording with
nothing enforcing that the two stay in step, and that the citation alone would carry the invariant.
Lens C, the principles reviewer and the primary reviewer judged the clause necessary and the
paragraph now internally consistent. Lens B recorded the disagreement, judged it non-blocking, and
explicitly recommended against a fourth fix round over it. Recorded here so the position is visible
rather than resolved by silence.

### Raised and out of scope

Bugbot noted that `skills/myflow-contracts/workspace-isolation.md:171` claims the ports actually
bound are printed in the `/myflow-do` handoff, while `/myflow-do` starts no application, so nothing
is bound at handoff time. The old text made the same claim about the guide, so the defect predates
this change and was only carried through the rename. Not fixed here.
