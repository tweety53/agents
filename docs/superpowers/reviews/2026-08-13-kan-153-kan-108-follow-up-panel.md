# Review panel — kan-153-kan-108-follow-up

## Configuration

- **Roster:** `light` (recorded), operator confirmed at dispatch.
- **Required slots:** Primary (slot 0), Principles (slot 2), Code review low (slot 3).
- **Models:** every slot on Sonnet, from `models.reviewPanel`. No slot inherited the parent model.
- **Optional slots:** none. The operator was asked before dispatch and declined all of them for this
  pass — recorded as **declined**, not as triggers that never fired. Adversarial's trigger (over 300
  changed lines; behaviour changes to code with existing tests) and Lens B's (over 200 changed lines,
  a new module) would both have fired on this diff.
- **Diff cap:** `scripts/check-panel-diff-size.sh` measured **1821** changed lines against a cap it
  reported as not exceeded; exit 0. No operator prompt was required.

## Pass 1

**Mode:** full roster, whole-branch diff `.superpowers/sdd/final-review.diff`, merge base `a6d2855`.

| Slot | Model | Result |
|------|-------|--------|
| 0 — Primary | sonnet | 5 findings (F1–F5) |
| 2 — Principles | sonnet | clean, no findings |
| 3 — Code review low | sonnet | 3 findings (F6–F8) |

**Slot 3, not the principles slot, is the one that runs the `code-review` skill** — an earlier
version of this record said otherwise, which was this record's error and not a fact about the run.
In pass 1 that slot reported running the skill at effort `low` without substitution. In pass 2 it
could not confirm the skill was present and fell back to a general-purpose reviewer briefed to report
high-confidence defects only, and said so. The substitution is recorded rather than smoothed over:
the slot was not dropped, and the fallback is what its dispatch prescribes when the harness offers no
such skill.

### Findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Primary | Major | `openspec/changes/kan-153-kan-108-follow-up/proposal.md:69` | proposal still described F43 as verified-absent after design.md superseded that decision |
| F2 | Primary | Major | `openspec/specs/myflow-review-panel-economics/spec.md` | the code-span repair was owned by no task, tracked by no field, and named in no commit |
| F3 | Primary | Minor | `openspec/changes/kan-153-kan-108-follow-up/tasks.md` | Task 3's baseline declared 16 cases against a harness carrying 21 |
| F4 | Primary | Minor | `scripts/test-check-unfinished-work.sh` | the spec scenario for a missing library being a refusal had no test in either harness |
| F5 | Primary | Minor | `scripts/check-markdown-integrity.py` | the three-layer special-case chain in the unterminated-paragraph signal is the piece most likely to need a fourth layer |
| F6 | Code review low | Major | `scripts/check-markdown-integrity.py:588` | signal 2 excluded only blockquote and heading predecessors, so a tilde fence or tight list item above a blockquote was a false positive |
| F7 | Code review low | Major | `scripts/check-unfinished-work.sh` | the file's own comment declared these guards single-file by design while the same file sourced a library |
| F8 | Code review low | Minor | `scripts/lib/panel-record.sh` | the digit-bound pattern stayed copy-pasted at four call sites, the exact drift the library was extracted to close |

### Disposition

F1, F2 and F3 were plan-artifact defects and were repaired by the parent directly, since a subagent
may not touch `openspec/`. F4, F6, F7 and F8 went to one fix subagent on `models.panelFix`.

F5 was put to the operator at the handback, since the two slots examined the same code and reached
opposite conclusions about whether its accreted special cases are justified. The operator withdrew it
with the reason recorded on its marker line: every layer traces to a false positive that was actually
observed, the harness pins the behaviour in both directions, and no reviewer proposed a simpler rule
that keeps the coverage. It is accepted complexity and a watch item, not a defect fixed or ignored.

F7's resolution was an operator decision. The library is kept and the contradicting comment corrected,
on the evidence that `setup.sh` installs skills, rules, commands and the managed block and does not
distribute `scripts/` at all — so a guard and the library it sources travel together in the agents
repository. The two other guards that cite the same single-file rationale to justify duplicating code
were deliberately left untouched as a separate question.

### Reproducers

reproducers-total: 8
finding-reproducer: F1 none — a plan-consistency defect in prose, with no runnable check
finding-reproducer: F2 none — an absence of tracking, established by reading the commit range and the plan's fields
finding-reproducer: F3 scripts/test-check-markdown-integrity.sh
finding-reproducer: F4 none — missing test coverage; the behaviour itself was manually confirmed correct
finding-reproducer: F5 none — a maintainability judgement, not a runtime defect
finding-reproducer: F6 none — needs a constructed fixture with a tilde fence above a blockquote, and no such file exists in the worktree to point at
finding-reproducer: F7 none — a contradiction between a comment and the code beside it, with no runtime behaviour to demonstrate inside this checkout
finding-reproducer: F8 none — a duplication observation across four call sites, not a runtime defect

### Findings received outside the panel

A peer session sent two reuse findings unprompted. They are recorded here because they were acted on,
but they are **not** panel findings and carry no marker line.

Its digit-bound duplication finding was the same defect as F8, reached independently, and is closed by
the same fix. Its second finding — that the new guard defines its own block-classifier regexes rather
than sharing `check-plan-provenance.py`'s — was assessed and declined: the peer's own text concedes
the coarsened classifier is legitimately not that one, sharing only the base regex constants would
couple two standalone guards' classifiers through a third file, and the principles slot examined the
same duplication and did not raise it. Recorded rather than actioned.

## Pass 2

**Mode:** Targeted, by operator decision, against `.superpowers/sdd/fix-round-1.diff` (158 changed
lines, 7 files).

**An escalation trigger fired and the operator overrode it.** Fix round 1 altered a guard's behaviour,
which the auto-escalate set requires escalate to Full without asking. The operator was shown that and
chose a Targeted re-run of Primary and Code review low — the slot that raised every fixed finding,
plus Primary as integration check. Principles reviewed the same code clean in pass 1 and the fix
introduced no new abstraction, reusing an existing constant and an existing kind-set. Recorded here
because an override nobody wrote down is indistinguishable from a trigger that failed to fire.

| Slot | Model | Result |
|------|-------|--------|
| 0 — Primary | sonnet | clean — F1 through F4 confirmed closed, no new findings |
| 3 — Code review low | sonnet | clean — F6, F7 and F8 confirmed closed, no new findings |

Primary confirmed F1 through F4 closed and additionally verified commit coherence commit by commit:
six commits, one per task, each commit's file list matching that task's declared `Files` field with no
cross-task contamination, and no `fixup!` left unsquashed after the history was rewritten twice during
the fix round. Slot 3 verified each fix rather than accepting the fix report: it re-checked `setup.sh` itself for
whether `scripts/` is distributed before accepting F7's rewritten comment, confirmed
`PANEL_RECORD_TOTAL_DIGITS` is byte-identical to the literal it replaced at all four call sites and
that each site quotes the interpolation, and confirmed the F6 narrowing did not silence signal 2 —
un-marked prose above a blockquote still fires. It also checked that the new missing-library cases
clean up through an `EXIT` trap and never write into the real library. All four harnesses run twice
where touched.

findings-total: 8
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 withdrawn every layer of the signal traces to a false positive that was actually observed, all 25 harness cases pin the behaviour in both directions, and no reviewer proposed a simpler rule that keeps the coverage — accepted as a watch item rather than a defect to fix now
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
