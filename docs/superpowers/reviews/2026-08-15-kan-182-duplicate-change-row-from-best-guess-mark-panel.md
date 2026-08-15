# Final review panel — kan-182-duplicate-change-row-from-best-guess-mark

**Roster:** `light` (recorded). Required slots: Primary, Principles, Code review (low). Panel model:
`sonnet` for every slot the panel dispatched directly; no slot was dispatched by `subagent_type`, so
no entry reads `unknown (agent-defined)`.

**Diff:** `.superpowers/sdd/final-review.diff`, `git diff c07fed7` — 123 changed lines.
`scripts/check-panel-diff-size.sh` measured **123**, under the cap, exit 0. No operator answer was
needed on size.

**Optional slots.** The trigger table was evaluated against the diff before dispatch. One trigger
fired: **Adversarial** — an existing test file was modified (`scripts/test-check-stage-mark-calls.sh`
gained cases) and a guard's behaviour changed. Under the `light` preset the fired trigger went to the
operator in one multi-select prompt; the operator selected **Adversarial** and nothing else.
**Security**, **Lens B** and **Lens C** are recorded as *trigger never fired* — no auth, crypto,
secrets, query construction, deserialization or dependency change; 123 lines, under the ~200
threshold, with no new modules; no error handling, retries, schedulers, external integrations, config
or migrations. None was declined.

## Pass 1 — full roster

| Slot | Model | Result |
|------|-------|--------|
| Primary — plan alignment + code quality | sonnet | clean, no findings |
| Principles — Merged lens | sonnet | 1 Minor (F5) |
| Code review (low) | sonnet | clean, no findings |
| Adversarial | sonnet | 3 Major, 1 Minor (F1–F4) |

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Adversarial | Major | `scripts/check-stage-mark-calls.sh:222` | a quoted guess placeholder evades the check, because the match requires the token to begin with a literal `<` |
| F2 | Adversarial | Major | `scripts/check-stage-mark-calls.sh:221` | a trailing inline comment after the placeholder evades the check, because last-token extraction picks up the comment word |
| F3 | Adversarial | Minor | `scripts/check-stage-mark-calls.sh:223` | a placeholder containing "guess" is rejected regardless of context, so a hypothetical negating placeholder would false-positive |
| F4 | Adversarial | Major | `skills/myflow-fast/SKILL.md` State gate | the creating-run firing point is stated only in this file, and the cited section A says nothing about a foreign `do.state-gate` mark, so an executing agent may drop the mark instead of firing it |
| F5 | Principles | Minor | `scripts/check-stage-mark-calls.sh:220` | the change-argument extraction is inlined rather than following the file's own named `*_value` helper convention |

**Pass 1's five findings carry their status and reproducer in the single marker block below**, which
covers every finding this record holds rather than one block per pass.

**On the three rejected reproducer shapes.** Each was recorded rather than rewritten to satisfy the
guard, per the panel contract: a rejected shape is a refusal, not something to reshape until it
passes. None of the three lines was ever executed. The findings were nonetheless established as real
by the parent constructing equivalent fixtures itself, so none of them reached the operator as
unverifiable.

## Pass 2 — Full re-run

**Escalated automatically, and why:** the pass-1 fix round altered a guard's behaviour. That clause
of the escalation ladder fires on its own, so the operator was not asked about the escalation itself.
Size did not escalate it — the fix-round diff measured 138 changed lines, under the threshold, and
adds no new file. `scripts/check-panel-diff-size.sh` measured the rewritten branch diff at 239, under
the cap, exit 0.

**Slots.** Every required slot re-ran against a rewritten `final-review.diff` (311 changed lines).
Conditional slots were re-evaluated against `fix-round-1.diff`: **Adversarial** still fires (a test
file was modified again) and re-ran; **Lens B** does not fire against the fix-round diff (138 lines,
under ~200; no new modules) and is recorded *trigger did not fire — not re-run*; **Security** and
**Lens C** never fired. The operator was asked which conditional slots to include and took the
recommended answer: Adversarial alone.

| Slot | Model | Result |
|------|-------|--------|
| Primary | sonnet | 1 Major (F6) |
| Principles — Merged lens | sonnet | clean; F5 confirmed resolved |
| Code review (low) | sonnet | clean |
| Adversarial | sonnet | 2 Critical (F7, F8); confirmed no regression on the three original checks, and that cases 16–18 fail with the fix reverted |

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F6 | Primary | Major | `scripts/check-stage-mark-calls.sh:202` | a quoted change argument containing an internal space is truncated by the space-splitting last-token extraction and never matches |
| F7 | Adversarial | Critical | `scripts/check-stage-mark-calls.sh:182` | the awk quote-tracker closes a quote on an escaped literal quote, desyncing so that a later `#` discards the rest of the line, including the guessed argument |
| F8 | Adversarial | Critical | `scripts/check-stage-mark-calls.sh:201` | the extraction splits on a literal space only, so a tab before the change argument leaves the whole tail as the "last token", which never matches |

findings-total: 8
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 withdrawn — narrow literal matching is the recorded design decision (design.md, guard-matches-guess-placeholders), no such placeholder exists in the skill tree, and context-sensitive matching is the parsing complexity that produced F6-F8
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed

reproducers-total: 8
finding-reproducer: F1 none — the slot supplied a reproducer rejected by shape (shell metacharacters), so it was never run; the parent confirmed the bypass directly with an equivalent fixture, exit 0 where exit 1 was required
finding-reproducer: F2 none — the slot supplied a reproducer rejected by shape (shell metacharacters), so it was never run; the parent confirmed the bypass directly with an equivalent fixture, exit 0 where exit 1 was required
finding-reproducer: F3 none — the slot supplied a reproducer rejected by shape (shell metacharacters), so it was never run; the parent confirmed the behaviour directly with an equivalent fixture, exit 1 on a placeholder that negates being a guess
finding-reproducer: F4 none — a text-execution judgment about skill prose, with no runnable check
finding-reproducer: F5 none — a structural consistency judgment, with no runnable check
finding-reproducer: F6 none — the fixture needs quotes and an internal space, which no shape-clean command can carry; the slot verified it directly and reported exit 0 where exit 1 was required
finding-reproducer: F7 none — the fixture needs an embedded escaped quote, which no shape-clean command can carry; the slot verified it directly and reported exit 0 where exit 1 was required
finding-reproducer: F8 none — the fixture needs an embedded literal tab, which no shape-clean command can carry; the slot verified it directly and reported exit 0 where exit 1 was required

**Nothing bounced this pass.** No finding's reproducer was executed, because none was supplied in a
runnable, shape-clean form; each was established by its own slot or by the parent building the
fixture directly, and the disposition is recorded per finding above.

**One root cause across F6, F7 and F8, and what it changed about the fix.** All three are the same
defect: identifying "the last token" requires parsing a shell command line, and pass 1's fix bought
quote-awareness that introduced three new evasions of its own. Fix round 2 therefore deletes the
extraction entirely and matches the bracketed guess placeholder anywhere in the assembled command —
strictly stronger than the extraction, and simpler. The one behaviour that changes is that a guessed
placeholder written inside a trailing comment is now reported too, which is correct.

## Fix round 2

F6, F7 and F8 share one root cause: identifying "the last token" requires parsing a shell command
line, and pass 1's fix bought quote-awareness that introduced three new evasions of its own. The fix
therefore deleted `change_arg_value` entirely — the check matches a bracketed `<...guess...>`
placeholder anywhere in the assembled command text and reasons about shell syntax not at all. The
guard is a net 15 lines shorter than before that helper existed. Cases 20-22 cover the three
bypasses; several earlier cases had their rationale comments rewritten, with no assertion changed.

The parent independently confirmed all five evasion shapes are now caught (quoted, trailing comment,
internal space, escaped quote, tab), that a compliant `<name>` still passes, and that `stage end`
remains exempt.

**F3 was withdrawn by the operator at the handback**, with the reason recorded on its marker line
above. It was never fixed and never silently closed.

## Pass 3 — Full re-run

**Escalated automatically:** fix round 2 again altered a guard's behaviour, and three fix rounds is
itself an escalation clause. `scripts/check-panel-diff-size.sh` measured the rewritten branch diff at
287, under the cap, exit 0.

**Slots.** The three required slots re-ran against a rewritten `final-review.diff` (359 changed
lines). **Adversarial**'s trigger fires again against `fix-round-2.diff` (a test file was modified
again), and the operator was asked which conditional slots to include and chose **none** — so
Adversarial is recorded **declined by the operator**, distinctly from a slot whose trigger never
fired. **Security**, **Lens B** and **Lens C** did not fire.

| Slot | Model | Result |
|------|-------|--------|
| Primary | sonnet | clean — F6, F7 and F8 each reproduced against the live guard and now caught; no new evasion found |
| Principles — Merged lens | sonnet | clean — the simplification judged proportionate, and no test comment describes the deleted mechanism as current |
| Code review (low) | sonnet | clean |
| Adversarial | sonnet | **declined by the operator for this pass — its pass-2 result is therefore stale** |

Primary additionally neutered the new matcher in a `mktemp` copy and re-ran the harness: all eleven
guess-related assertions fail without it, while the cases covering the three original checks still
pass — so the new cases are not vacuous and nothing was lost from the checks that predate this change.

**Zero open findings.** Eight findings, seven fixed and one withdrawn by the operator with its reason
recorded on its marker line.

**One qualification, stated rather than glossed.** The handoff bar asks for a non-stale clean result
from every slot in the roster. Adversarial raised F7 and F8 on pass 2 and was declined for pass 3 by
the operator, so its clean-ness is the pass-2 result and does not cover fix round 2's rewrite. Primary
covered that rewrite adversarially in its place — reproducing all three findings against the live
guard and probing for new evasions — but that is a substitution, not the same thing, and the operator
made the call knowing it.
