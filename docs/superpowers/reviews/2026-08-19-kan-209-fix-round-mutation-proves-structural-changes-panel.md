# Review panel record — kan-209-fix-round-mutation-proves-structural-changes

Roster: `light`. Panel model: `sonnet` (recorded `models.reviewPanel`).

Diff read: `.superpowers/sdd/final-review.diff`, 108 insertions across 3 files, merge base `c6bd1cb`.

## Pass 1 — full roster

Mode: full (pass 1 always runs the full roster selected for this change).

**Diff size:** `check-panel-diff-size.sh` exit 0, under the 2000-line cap. No operator prompt was needed.

**Slots dispatched.** Required, per the `light` preset: Primary (slot 0), Principles (slot 2,
Merged lens), Code review at effort low (slot 3). Optional: Adversarial (slot 5) — its trigger
*any test modified or deleted* fired on `scripts/test-check-unfinished-work.sh`, and the operator
selected it from the single multi-select prompt the lighter presets use. Security (slot 4),
Principles lens B and lens C: **not run — no trigger fired**. Every slot ran on `sonnet`, named
explicitly; no slot was dispatched by `subagent_type`, so no ledger entry reads
`unknown (agent-defined)` for this pass.

**Slot verdicts.** Primary: clean. Principles: 1. Code review (low): 3. Adversarial: 3.

### Findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Major | skills/myflow-do/SKILL.md:721 | the rule's scope is never stated in the operative text — the delta spec scopes it to a panel fix round, but the added text says only the fix round, a term section 4 already uses for the per-task review's fixup |
| F2 | Adversarial | Major | scripts/test-check-unfinished-work.sh:913 | a record line whose free text carries the literal marker or total label is counted as a malformed marker, so an otherwise clean record reads OUTSTANDING and the inert claim is false |
| F3 | Adversarial | Major | skills/myflow-do/SKILL.md:723 | the obligation is anchored to the list the fix subagent reports, and nothing requires the parent to check that list against the fix diff for completeness |
| F4 | Adversarial | Critical | scripts/test-check-unfinished-work.sh:915 | the case pins only the correct placement, so the guard's sensitivity to a misplaced line rests on a developer's review-time mutation rather than on a committed case |
| F5 | Code review (low) | Minor | skills/myflow-do/SKILL-rationale.md:161 | the no-guard argument rebuts only a diff-classifying guard and never the cheaper shape-and-count guard that check-panel-reproducers.sh already precedents |
| F6 | Code review (low) | Minor | skills/myflow-do/SKILL-rationale.md:157 | the record is called the identical question to the reproducer record, but a reproducer is re-runnable and guarded while this record is unverifiable prose |
| F7 | Code review (low) | Minor | skills/myflow-do/SKILL.md:723 | the parent classifies what counts as executable behaviour with no stated route for an ambiguous case |
| F8 | Principles | Critical | scripts/check-contract-budget.sh:87 | the budget-row change was exempted from proof on a reason the guard's own harness contradicts, so the fix round failed the bar it was landing |
| F9 | Adversarial | Major | skills/myflow-do/SKILL.md:769 | the label constraint is scoped to one line's free text, but the guard counts the labels across the whole file, so a fenced example anywhere in the record still flips it |
| F10 | Adversarial | Minor | skills/myflow-do/SKILL.md:783 | the hunk walk has no way to validate an exemption for a hunk that weakens or deletes test code, since the deleted assertion is the thing that would have failed |
| F11 | Adversarial | Major | skills/myflow-do/SKILL.md:769 | the rescoped constraint generalised its sentence to the whole record but not its enumeration, missing a third label that the sibling reproducer guard also counts unanchored |
| F12 | Adversarial | Minor | skills/myflow-do/SKILL.md:797 | the obligation for a hunk that removes coverage is asserted and never checked, so a round can name a replacement that does not cover it and satisfy every sentence |
| F13 | Principles | Minor | skills/myflow-do/SKILL.md:770 | the label constraint binds the whole record but sits inside the mutation subsection, where a writer of another slot's free text has no reason to look for it |
| F14 | Principles | Minor | skills/myflow-do/SKILL-rationale.md:161 | the no-guard argument rebuts two guard shapes and never the one the observed recurrence argues for, a lint catching a reserved label outside its marker position |
| F15 | Primary | Critical | scripts/test-check-contract-budget.sh:185 | fix round 1's budget-row raise broke the guard's own harness — its over-budget fixture is a fixed 20000 bytes, which no longer exceeds the raised row, so a case that must fail now passes |
| F16 | Adversarial | Critical | scripts/test-check-contract-budget.sh:197 | the midpoint fixture stops discriminating when its two source rows converge, so the case cannot fail even when the bug its own name describes is present |
| F17 | Adversarial | Major | scripts/test-check-contract-budget.sh:19 | the helper's missing-row error branch is new shared behaviour with no coverage — replacing it with a zero default leaves every case green, a demonstrated surviving mutant |
| F18 | Adversarial | Minor | scripts/test-check-contract-budget.sh:19 | the helper takes the first of a duplicated row silently, so a table with two rows for one path is read without a warning in either direction |
| F19 | Code review (low) | Minor | scripts/test-check-contract-budget.sh:21 | the helper uses the awk -v idiom the guard's own budget_for documents as deliberately avoided, in the harness sitting directly beside it |
| F20 | Code review (low) | Major | scripts/test-check-contract-budget.sh:95 | the duplicate case asserts only the exit status and the path, so it cannot tell which refusal fired and a swapped diagnosis survives as a mutant |
| F21 | Code review (low) | Minor | scripts/test-check-contract-budget.sh:37 | the read scans the guard's whole source rather than its table, so any stray line whose first two fields match a path and a number aborts as a false duplicate |
| F22 | Code review (low) | Minor | scripts/test-check-contract-budget.sh:38 | a table whose final line carries no newline loses that row entirely, an undocumented precondition on a helper now taking a caller-supplied table |
| F23 | Code review (low) | Minor | scripts/test-check-contract-budget.sh:90 | the duplicate-row temp table is created before the cleanup trap is installed and is never added to one, so an interrupt leaks it |
| F24 | Code review (low) | Minor | scripts/test-check-contract-budget.sh:140 | the unterminated-row case omits the guard its two sibling cases carry, so on the regression it exists to catch the harness dies with no failure line and most checks unrun |

findings-total: 24
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
finding-status: F18 fixed
finding-status: F19 fixed
finding-status: F20 fixed
finding-status: F21 fixed
finding-status: F22 fixed
finding-status: F23 fixed
finding-status: F24 fixed

finding-reproducer: F1 none — a prose scoping finding; the absent phrase is established by reading the added section, with no runnable check
finding-reproducer: F2 none — reproduced by the parent by building a panel record whose record line free text carries the literal marker label; the guard returned OUTSTANDING naming a malformed marker, which needs a multi-line fixture rather than one command
finding-reproducer: F3 none — a process gap in prose, with no runnable check
finding-reproducer: F4 none — the claim is about what the committed case omits, established by reading it, with no runnable check
finding-reproducer: F5 none — a rationale-completeness finding in prose, with no runnable check
finding-reproducer: F6 none — a prose-comparison finding, with no runnable check
finding-reproducer: F7 none — a design-gap finding in prose, with no runnable check
finding-reproducer: F8 none — a record-accounting defect: the entry claimed an exemption whose stated reason the guard's own harness header contradicts, established by reading both, with no single command to run
finding-reproducer: F9 none — reproduced by the parent by building a panel record carrying a fenced example that quotes the record grammar; the guard returned OUTSTANDING, which needs a multi-line fixture rather than one command
finding-reproducer: F10 none — a structural gap in prose, not present in this diff, with no runnable check
finding-reproducer: F11 none — reproduced by the parent by building a well-formed record plus a fenced example quoting the third label's grammar; the sibling guard exited 1, which needs a multi-line fixture rather than one command
finding-reproducer: F12 none — a process-design gap in prose; showing it needs a fabricated round's claim and human judgement, not a runnable check
finding-reproducer: F13 none — a documentation-placement finding, established by reading the section headings, with no runnable check
finding-reproducer: F14 none — a rationale-completeness finding, established by reading the rejected-guard arguments against the pattern of the earlier findings, with no runnable check
finding-reproducer: F15 scripts/test-check-contract-budget.sh
finding-reproducer: F16 none — reproduced by the parent in a scratch copy by converging the two source rows; the case still reported ok, which needs an edited copy of the guard rather than one command
finding-reproducer: F17 none — reproduced by the parent in a scratch copy by replacing the error branch with a zero default; all cases still passed, which needs an edited copy of the harness rather than one command
finding-reproducer: F18 none — reproduced by the raising slot in a scratch copy carrying a duplicated row; the helper returned the first value with no warning
finding-reproducer: F19 scripts/check-contract-budget.sh
finding-reproducer: F20 none — reproduced by the parent in a scratch copy by rewording the duplicate refusal to the missing-row wording; the case still reported ok and the suite passed, which needs an edited copy rather than one command
finding-reproducer: F21 none — reproduced by the parent in a scratch copy by appending a stray line outside the table whose first two fields match a real row; the harness aborted on a false duplicate, which needs an edited copy rather than one command
finding-reproducer: F22 none — established by the raising slot against a table whose last line carries no newline; needs a constructed table rather than one command
finding-reproducer: F23 none — observing the leak needs an interrupt signal mid-run, established by reading the trap's position
finding-reproducer: F24 none — reproduced by the parent in a scratch copy by reverting the read guard; the harness stopped after two passing lines with no failure line, which needs an edited copy rather than one command

reproducers-total: 24

### Parent's note on F4

F4 is recorded as its raising slot stated it. The parent's own reading is that a characterization
case is *supposed* to commit the correct arrangement — the mutations that establish it can fail are
run at review time and reverted, which is what mutation-proof means, and both the Primary slot and
the raising slot independently confirmed the committed fixtures do flip. What survives scrutiny is
the narrower point: nothing *committed* pins the guard's sensitivity, so it is re-established by
hand on every future read. That is a real gap and the fix round addresses it by committing the
negative cases, rather than by treating the case as broken.


## Fix round 1

Fix subagent: one combined dispatch, model: `sonnet` (recorded `models.panelFix`). All seven
findings fixed; none disputed by the fixer. Folded into the two task commits via
`git commit --fixup` and `git rebase --autosquash` — Task 1 `26e44cc` → `cb4d763`, Task 2
`50e8522` → `7478906`. Diff read for the re-run: `.superpowers/sdd/fix-round-1.diff`.

**Bounced findings this pass:** none.

### Mutations this round proved

The round changed executable behaviour in two files, so the rule this change lands binds it. Each
mutation altered one mechanism, run against the cases as committed, restored afterwards.

fix-mutation: scripts/test-check-unfinished-work.sh — disabled the marker-block contiguity comparison in scripts/check-unfinished-work.sh — case 12a's two assertions failed, expected OUTSTANDING got CLEAR
fix-mutation: scripts/test-check-unfinished-work.sh — disabled the row-versus-marker identifier comparison in scripts/check-unfinished-work.sh — case 12b's two assertions failed the same way
fix-mutation: scripts/check-contract-budget.sh — reverted the rationale file's budget row to its pre-change value of 15721 — the guard exited 1 reporting 16744 bytes over budget, and exited 0 again once restored
fix-mutation: skills/myflow-do/SKILL.md — none — prose only
fix-mutation: skills/myflow-do/SKILL-rationale.md — none — prose only
fix-mutations-total: 5

Both mutations also failed pre-existing cases exercising the same two mechanisms (the 4s-0/4s-0b
and 4q/4r families). That is expected and is not cross-contamination: each mutation still altered
exactly one mechanism, and the new case failed on its own account.

### Escalation — automatic, and why

Pass 2 runs the **full** roster rather than a targeted re-run. The trigger is the ladder's first
clause: the fix touched a file outside the set the findings named — `scripts/check-contract-budget.sh`,
whose budget row for `skills/myflow-do/SKILL-rationale.md` was raised from 15721 to 20930 as the
file grew to 16744 bytes. That edit changes a guard's configuration, so it is not read as
incidental. No operator prompt is involved; the ladder escalates on its own and the reason is
recorded here.

## Pass 2 — full roster

Escalated automatically, for the reason recorded above. Slots re-run: Primary, Principles, Code
review (low) — the `light` preset's three required slots — plus Adversarial, whose trigger *any test
modified or deleted* still fires against `fix-round-1.diff`. Security, Lens B and Lens C: **not
re-run — no trigger fired**. Every slot ran on `sonnet`, named explicitly. The operator's pass 1
selection of Adversarial carries forward; it was not re-asked.

**Slot verdicts.** Primary: clean. Code review (low): clean. Principles: 1 (F8). Adversarial: 2
(F9, F10).

### F8 — the parent's own accounting, and what it cost

F8 is a finding against the **parent**, not against either implementer. The fix-round entry above
originally recorded the budget-row change as an exemption, reasoning that the guard's behaviour on
that row was covered by its own harness. That reason is contradicted by the harness's own header,
which states that every case runs against a sandboxed fixture tree so it never depends on the real
repository's sizes — so no test exercised the real row at all.

A budget row is executable behaviour under this change's own definition: the guard reads the table
and gates its exit code on the comparison. The mutation was one command away and was not run. The
parent has since run it — reverting the row to 15721 made the guard exit 1, naming the file as
16744 bytes over budget, and restoring it returned exit 0 with a clean working tree — and the
recorded line above now states that mutation instead of the exemption.

This is the second time in this run that the rule being landed caught its own implementation. The
first was Task 1's case being proved against a fixture it did not commit. Both are recorded rather
than quietly corrected, because a change arguing that rounds skip proof where it is merely implied
is worth exactly as much as its own compliance record.

### F9 and F10 — dispatched to fix round 2

F9 is confirmed: the parent reproduced a clean record turning `OUTSTANDING` on a fenced example that
quoted the record grammar. The constraint's scope is the defect, not its reasoning.

## Fix round 2

Fix subagent: one combined dispatch, model: `sonnet` (recorded `models.panelFix`). F9 and F10 fixed;
neither disputed. F8 was fixed by the parent before this dispatch and was not given to the fixer.
Folded into Task 2's commit via `git commit --fixup` and `git rebase --autosquash` — `7478906` →
`0189ecc`. Task 1's commit `cb4d763` is unchanged, since the fixup targeted Task 2 alone. Diff read
for the re-run: `.superpowers/sdd/fix-round-2.diff`, 20 insertions and 8 deletions in one file.

**Bounced findings this pass:** none.

### Mutations this round proved

fix-mutation: skills/myflow-do/SKILL.md — none — prose only, no executable behaviour changed
fix-mutations-total: 1

The fixer noted, correctly, that it used the two literal labels in its own prose and that doing so is
safe: the guard's record variable points only at the panel record, never at a skill file, so the
constraint being described does not bind the file describing it. Recorded because a later reader
will otherwise wonder.

### Re-run mode — targeted, and why

Pass 3 is **targeted**, not escalated. The ladder's clauses were checked one by one against this fix:
the diff touches only `skills/myflow-do/SKILL.md`, a file the findings named, so nothing sits outside
that set; no delta spec, migration or guard behaviour was altered — this round changed no script at
all; no new Critical arose from a targeted re-run, since pass 2 was itself a full roster; two fix
rounds have run, short of the three-or-more clause; and the diff is 28 changed lines and adds no
file, so the size-plus-new-file clause cannot fire.

Slots re-run: **Primary** (always, as the integration check), **Principles** (raised F8) and
**Adversarial** (raised F9 and F10). **Code review (low)** raised nothing in pass 2 and is *not
re-run — subject unchanged*; its clean pass 2 result stands, and not re-running it neither closes
nor softens anything, since the zero-open-findings bar still governs every slot in the roster.

### Budget note carried to the next reader

`skills/myflow-do/SKILL.md` is now 70267 bytes against its 71317 row — roughly 1050 bytes of
headroom, down from 2200 before this round. The next change that adds a section here will trip the
ratchet and should raise the row rather than trim the rule.

## Pass 3 — targeted

Slots re-run: Primary (integration check), Principles (raised F8), Adversarial (raised F9 and F10).
Code review (low): *not re-run — subject unchanged*; its clean pass 2 result stands.

**Slot verdicts.** Primary: clean. Principles: clean — it independently reproduced F8's mutation,
reverting the budget row and observing the guard exit 1, then restoring it. Adversarial: 2 (F11,
F12).

### F11 — the same defect one guard over

F9 rescoped the label constraint from one line to the whole record. It generalised the *sentence*
without generalising the *enumeration*: it names the two labels the finding guard counts unanchored,
and misses a third that the sibling reproducer guard counts the same way. The parent reproduced it —
a fully well-formed record plus a format example quoting that third label's grammar makes the
sibling guard exit 1.

The adversarial slot's precision here is worth recording: the reproducer *count* line is anchored in
that guard and is therefore safe, so exactly one label was missed rather than the whole family. The
repair is to derive the list from both guards rather than to extend it by hand, which is the
enumeration-versus-inference distinction this repository already applies to Jira status positions.

### F12 — an accepted-risk question, not a defect to paper over

F10's fix turned an unvalidated exemption into a mandatory statement, and left the validation gap
where it was. Its failure direction is the opposite of F9's and F11's: those block a clean round,
while this one can let a real gap through. That asymmetry is why it goes to a fix round rather than
being accepted silently.

Three fix rounds will have run once round 3 completes, so **pass 4 escalates to the full roster
automatically** under the ladder's three-or-more clause, regardless of what round 3 touches.

## Fix round 3

Fix subagent: one combined dispatch, model: `sonnet` (recorded `models.panelFix`). F11 and F12 fixed,
neither disputed. Folded into Task 2's commit — `0189ecc` → `a1c39f1`. Task 1's `cb4d763` unchanged.
Diff: `.superpowers/sdd/fix-round-3.diff`, 17 insertions and 14 deletions in one file.

**Bounced findings this pass:** none.

### Mutations this round proved

fix-mutation: skills/myflow-do/SKILL.md — none — prose only, no executable behaviour changed
fix-mutations-total: 1

No budget row was raised this round — the file grew to 70530 bytes against its unchanged 71317 row —
so the executable-behaviour case that made F8 a Critical does not arise here. The parent checked that
rather than accepting it: the guard's row is byte-identical to its pre-round value.

### What F11's repair actually changed

The list of hazardous labels is now **derived** from the two guards rather than extended by hand, and
the text says so, pointing a later reader at re-reading the guards instead of guessing. It also names
the one label that is counted but **anchored**, so a reader knows why it is excluded rather than
assuming it was forgotten. That distinction — enumerate what is true, state what is deliberately
absent — is the same discipline this repository already applies to Jira status positions.

### Pass 4 — full roster, escalated automatically

Three fix rounds have now run, which fires the ladder's three-or-more clause on its own, regardless
of what round 3 touched. Every required slot re-runs.

**Adversarial re-runs as well, and that is a deliberate widening rather than the table's default.**
Its trigger — *any test modified or deleted* — does **not** fire against `fix-round-3.diff`, which is
prose in one skill file, so the table alone would leave it out with *not re-run — subject unchanged*.
It is re-run anyway because its last recorded result was two findings rather than a clean pass, and
the handoff bar wants a non-stale clean result from every slot that has run. Not re-running never
closes a finding it raised; running it is never a violation. Given that this change's whole argument
is that a round's own account of its success is not evidence, accepting the fixer's report on F11 and
F12 without the raising slot looking again would be the change contradicting itself.

## Pass 4 — full roster

Escalated by the ladder's three-or-more-fix-rounds clause, which fires on its own regardless of what
round 3 touched. Adversarial was re-run as a **deliberate widening** beyond the table, for the reason
recorded under fix round 3.

**Slot verdicts.** Primary: clean. Code review (low): clean. Adversarial: clean. Principles: 2 (F13,
F14).

Three of the four slots independently re-derived the hazardous-label list from the guards rather than
accepting the text's word for it, and all three reached the same three-label set with the fourth
label correctly excluded as anchored. That is the check F11's fix was written to make possible, and
it worked.

**Adversarial explicitly declined to raise** a flaky-test edge case against F12's new verification
step, judging it a generic test-quality concern rather than one this rule creates. Recorded because a
declined finding is evidence about the review, and because the alternative — a slot manufacturing a
finding to look thorough — costs a real fix round.

### F14 and the scope line

F14 is correct that the recurrence of one defect class across F2, F9 and F11 is evidence for a cheap
lint over reserved labels. It is **not** evidence against the recorded `no-guard` decision, and the
two are different guards: the operator declined a guard that would enforce *mutation-proof*, while
F14 describes a guard over the *record's well-formedness*, which belongs to the family
`check-unfinished-work.sh` and `check-panel-reproducers.sh` already form.

Building it inside this change would widen a deliberately prose-only ticket late in its life, and the
decision is the operator's rather than a fix round's. Fix round 4 therefore completes the rationale —
naming that guard shape and conceding what the recurrence shows — and the guard itself is carried to
the operator as a follow-up candidate at integration.

## Fix round 4

Fix subagent: one combined dispatch, model: `sonnet` (recorded `models.panelFix`). F13 and F14 fixed,
neither disputed. Folded into Task 2's commit — `a1c39f1` → `b79b59c`. Task 1's `cb4d763` unchanged.
Diff: `.superpowers/sdd/fix-round-4.diff`, 16 insertions across the two skill files.

**Bounced findings this pass:** none.

### Mutations this round proved

fix-mutation: skills/myflow-do/SKILL.md — none — prose only, no executable behaviour changed
fix-mutation: skills/myflow-do/SKILL-rationale.md — none — prose only, no executable behaviour changed
fix-mutations-total: 2

No budget row was raised: `skills/myflow-do/SKILL.md` is 70768 against its unchanged 71317 row, and
`skills/myflow-do/SKILL-rationale.md` is 17763 against 20930. The parent checked both rows are
byte-identical to their pre-round values rather than accepting the report, since a raised row would
make these exemptions wrong — the F8 failure.

### F13's repair is a pointer, not a copy

The cross-reference sits beside the record-format block's existing rule against quoting the format,
which is where a slot's author is already reading, and it points forward to the full statement rather
than restating the label list. A copy would have been the Single Source of Truth violation this
repository's own reference guard exists to catch.

### Pass 5 — full roster, escalated; Adversarial not re-run

The three-or-more-fix-rounds clause still fires, so every **required** slot re-runs: Primary,
Principles, Code review (low).

**Adversarial is *not* re-run — subject unchanged.** Its trigger, *any test modified or deleted*,
does not fire against `fix-round-4.diff`, which is prose in two skill files. Unlike pass 4, there is
no reason to widen beyond the table this time: its pass 4 result was **clean**, so nothing stale is
being carried, and not re-running it neither closes nor softens anything. The widening at pass 4 was
justified by a non-clean previous result; that justification is absent here, and repeating the
widening without it would be ceremony rather than review.

## Pass 5 — full roster (Adversarial not re-run)

**Slot verdicts.** Principles: clean — it verified F14's paragraph states the evidence plainly rather
than minimising it, checking the framing against `design.md`'s actual decision. Code review (low):
clean, and deliberately short, having verified the two files it owns are unchanged since it cleared
them twice. Primary: **1 Critical (F15)**.

### F15 — found by running the suite, not by reading the record

Every prior pass, and every fix round, ran the project's `## lint` list. **None of them ran the
project's `## test` list**, and the parent did not either — which is where this was hiding.

Fix round 1 raised `skills/myflow-do/SKILL-rationale.md`'s budget row from 15721 to 20930. The guard's
own harness builds an over-budget fixture for that same path at a hardcoded 20000 bytes — comfortably
over the old row, and **under** the new one. The case that must exit 1 now exits 0, and the harness
reports `FAIL a SKILL-rationale.md over budget fails: exit 0, wanted 1`.

The parent confirmed this is a regression this branch introduced, not a pre-existing failure: the same
harness passes at the merge base `c6bd1cb` in a scratch worktree, and fails deterministically at HEAD.
The parent then ran the project's entire `## test` list — twenty-six harnesses — and this is the only
failure among them.

**This is the third time the change's own subject has caught the change's own work**, and the sharpest
of the three. Fix round 1's mutation for that row was recorded and real: it reverted the row and
watched the production guard exit 1. But it mutated the row against the **production tree only**, and
never ran the guard's **own harness** — so a proof that looked complete missed the one place the row
is also depended on. That is precisely the failure the rule names: proving what was asked, and not
what was implied.

The rule as written already reaches this case — it says to establish that *an existing test* fails,
and the existing test here is the guard's harness. What failed was the practice, not the text.

## Fix round 5

Fix subagent: one dispatch, model: `sonnet` (recorded `models.panelFix`). F15 fixed, not disputed.
Task 1 `cb4d763` → `7d5d92d`; Task 2 `b79b59c` → `9eb999b` (rewritten by the rebase, content
unchanged).

**Bounced findings this pass:** none.

### The fix went wider than the finding, correctly

F15 named one fixture. The fixer checked the rest of the harness for the same coupling and found
**three more** hardcoded sizes measured against real rows — for `skills/myflow-contracts/build-green.md`,
`skills/myflow-status/SKILL.md`, and `skills/myflow-do/SKILL.md` — none of which had broken yet, each
of which would break on some future row change. All four now derive their size from the guard's own
row via a `budget_row_bytes` helper that reads the table with `awk` rather than sourcing the guard.

No other `scripts/test-*.sh` harness carries the coupling; the fixer checked every hardcoded `head -c`
size and found the remaining one unrelated.

### Mutations this round proved

fix-mutation: scripts/test-check-contract-budget.sh — changed the fixture's derived size from row-plus-margin to row-minus-margin — the over-budget case failed with exit 0 where 1 was wanted, reproducing F15 exactly
fix-mutation: scripts/test-check-contract-budget.sh — raised the guard's row for that path to 500000 — the harness still passed, which is what over-budget-by-construction means
fix-mutations-total: 2

**The parent re-ran the second mutation itself rather than accepting it**, since F15 exists precisely
because a recorded mutation was real but incomplete. Row set to 500000, harness passed all cases, row
restored, `git diff --stat` empty. The property holds.

**The parent also ran the project's entire `## test` list** — twenty-six harnesses, in batches,
because the full list exceeds the tool timeout the project's own configuration documents — plus all
eleven `## lint` guards. Everything green. That run is what should have happened before pass 1.

### Pass 6 — full roster

The three-or-more-fix-rounds clause fires, so every required slot re-runs. **Adversarial re-runs
too**, and this time by the table rather than by widening: `fix-round-5.diff` modifies
`scripts/test-check-contract-budget.sh`, so its trigger *any test modified or deleted* fires on its
own row.

## Pass 6 — full roster

**Slot verdicts.** Primary: clean — it verified the helper fails loudly and ran a scratch-copy
counterfactual proving the derived size, not coincidence, closes F15's coupling. Principles: clean.
Code review (low): 1 (F19). Adversarial: 3 (F16, F17, F18).

### Principles settled the F15 text question on the merits

Asked whether F15 needed a text change, it derived — rather than accepted — that the rule already
reaches the case: applying the rule's own prescribed mutation to the registered *test* would have
produced a surviving mutation, which the rule's "add the test that catches it" clause already
obligates. The failure was execution, not wording, so **no text was added**. Patching prose against a
discipline lapse is the move KAN-197 already rejected in this repository.

It also found direct precedent for a harness parsing a guard's source text
(`scripts/test-check-cleanup-complete.sh`), which settles the awk-extraction shape as house style
rather than an invention — F19 is about the *idiom*, not the approach.

### F16 and F17 — and the accounting failure that hid them

The parent reproduced both in scratch copies. Converging the two rows leaves the midpoint case
reporting `ok`; replacing the helper's error branch with a zero default leaves every case green.

**Fix round 5's record is where this went wrong, and that record is the parent's.** It declared
`fix-mutations-total: 2` against a diff carrying five behaviour-bearing hunks — the new helper and
four fixture derivations. Three hunks carried neither a mutation line nor a stated exemption, and two
of those three were hiding these findings. The rule's per-hunk walk is exactly the check that would
have caught it, and the parent accepted a count instead of performing the walk.

**This is the fourth time this change's own subject has caught this change's own work**, and the
second time the failure is the parent's rather than a subagent's. The pattern across all four is
identical: proof was produced where it was asked for, and skipped where it was implied.

F16 is also F15's own defect class relocated — a fixture that stops discriminating when its premise
erodes — introduced by the round that fixed F15, in the same file, unproved.

## Fix round 6

Fix subagent: one dispatch, model: `sonnet` (recorded `models.panelFix`). F16, F17, F18 and F19
fixed, none disputed. Task 1 `7d5d92d` → `dc8df7d`; Task 2 `9eb999b` → `e32ba7e`. Harness now carries
22 cases. Diff: `.superpowers/sdd/fix-round-6.diff`, one file.

**Bounced findings this pass:** none.

### Mutations this round proved — and the parent re-ran every one

fix-mutation: scripts/test-check-contract-budget.sh — converged the two source rows in a scratch copy — the new midpoint assertion fired, aborting with a message naming both rows, where the case previously reported ok
fix-mutation: scripts/test-check-contract-budget.sh — replaced the helper's missing-row branch with a zero default — the new missing-path case failed, reporting exit 0 with output 0
fix-mutation: scripts/test-check-contract-budget.sh — reverted duplicate detection to first-wins — the new duplicated-row case failed, reporting exit 0 with output 100
fix-mutation: scripts/test-check-contract-budget.sh — broke the table read comparison — the derived-size cases aborted rather than passing silently
fix-mutations-total: 4

**One line per behaviour-bearing hunk, which is what fix round 5 failed to produce.** The parent did
not accept the count this time: it re-ran the first three mutations itself, each in its own scratch
copy, one mechanism per mutation, and observed each named case fail. The fourth is the fixer's,
covering the read rewrite.

### What each finding's repair actually established

- **F16** — the midpoint case's premise is now asserted rather than assumed. A row pair too close to
  yield a strict midpoint aborts the harness with both values named, instead of passing vacuously.
  This is the repair F15's own fix should have carried.
- **F17** — the helper's missing-row branch has a case, so the shared helper's defensive path is no
  longer a surviving mutant.
- **F18** — a path matching more than one row is refused rather than silently resolved. The guard's
  own `budget_for` keeps its first-match tolerance; that is the guard's decision and was left alone.
- **F19** — the table is read with the `while read -r` idiom `budget_for` uses, not `awk -v`, matching
  the convention the guard states in its own comment. The extraction *approach* was kept: reading a
  guard's source text is established house style here.

### Pass 7 — full roster

The three-or-more-fix-rounds clause fires. Adversarial's trigger fires on its own row as well, since
`fix-round-6.diff` modifies a test file.

## Pass 7 — full roster

**Slot verdicts.** Primary: clean. Principles: clean. Adversarial: clean. Code review (low): 4 (F20
through F23).

### The scope question, put to Principles and answered

The parent asked Principles directly whether rounds 5 and 6 — spent entirely on a test harness for an
unrelated guard — were scope creep the panel should have stopped, with an instruction to say so
plainly even at pass 7. It ruled **legitimate**, tracing one causal chain: the ticket's prose forced a
budget-row raise; this change's rule required proving that raise; proving it exposed a genuine
pre-existing harness bug; fixing that required a shared helper; the helper is new shared behaviour,
so the same rule required proving it too. No unrelated work entered the branch.

It flagged separately, not as a finding, that **the ticket's own framing — "small, prose-only, no new
mechanism" — is now inaccurate**, since the branch carries a behaviour-bearing test helper. That is a
description-versus-reality gap for the operator, and it is carried to the handoff rather than written
to Jira: **Description sync** (`skills/myflow-contracts/jira-integration.md`) writes the issue
description only when *the operator* added scope during the run, and here the scope grew from the
panel's own rounds.

### The whole `## lint` and `## test` lists were run, finally

Primary reported `go vet` and `npx tsc -b` failing in the worktree. The parent established the branch
touches nothing under `stats/`, that both failures were missing build output in a fresh worktree, then
**built it** — `npm ci`, `make build` — and re-ran everything: `gofmt -l` clean, `go vet` 0,
`tsc -b` 0, `npm test` 163 passing, `go test ./... -race -count=1` 0, all twenty-six guard harnesses
and all eleven lint guards green. The build outputs are gitignored, so nothing can stage them.

### F20 — a surviving mutant inside the case that was added to kill a surviving mutant

F20 is the sharpest of the four and the parent reproduced it: rewording the duplicate refusal to the
missing-row wording leaves the case reporting `ok` and the suite passing. The case asserts an exit
status and a substring, never which of the helper's two refusals fired — so it cannot tell a correct
diagnosis from a wrong one. F17's case has the same shape and the same gap.

F21 was reproduced too: a stray line appended anywhere in the guard's source, whose first two fields
match a real row, aborts the harness on a false duplicate — because the read walks the whole file
rather than the table.

## Fix round 7

Fix subagent: one dispatch, model: `sonnet` (recorded `models.panelFix`). F20 through F23 fixed, none
disputed. Task 1 `dc8df7d` → `1a8d19d`; Task 2 `e32ba7e` → `4e01b4e`. Harness now carries 24 cases.

**Bounced findings this pass:** none.

### Mutations this round proved

fix-mutation: scripts/test-check-contract-budget.sh — reworded only the duplicate refusal's message to the missing-row wording — the duplicated-row case failed, naming the wrong diagnosis in its output
fix-mutation: scripts/test-check-contract-budget.sh — reworded only the missing-row refusal's message to the duplicate wording — the missing-path case failed the same way
fix-mutation: scripts/test-check-contract-budget.sh — appended a stray matching line outside the table, inside an inert heredoc so the guard stayed valid — the harness passed, where the whole-file read had aborted on a false duplicate
fix-mutation: scripts/test-check-contract-budget.sh — removed the unterminated-final-line handling from the read — the new case aborted the harness rather than passing
fix-mutation: scripts/test-check-contract-budget.sh — none — the temp-table cleanup change is purely structural: both tables now live under the directory the existing trap already removes, leaving no separate creation or removal to mutate
fix-mutations-total: 5

### The parent re-ran these, and got its own first two mutations wrong

The parent's first attempt at both key mutations was invalid, and that is recorded rather than
quietly re-run. The F20 mutation used a substitution that rewrote the assertion string as well as the
message it was supposed to break, so the mutation cancelled itself and reported a false survival. The
F21 mutation appended a bare line to a shell script, which the shell then tried to execute — exit
127 — so the harness failed for a reason unrelated to the mechanism under test.

Both were redone surgically: the F20 message was rewritten alone, and the F21 stray line was placed
inside an inert heredoc that left the guard syntactically valid. Only then did each mutation
demonstrate what it claimed. **A mutation that alters more than the one mechanism proves nothing** —
which is this change's own isolation rule, met from the wrong side by the parent enforcing it.

A **positive control** was also run, since a scoping fix can pass by disabling the check rather than
narrowing it: a real duplicate row inside the table still aborts. The scoping is narrow enough to
ignore a stray and wide enough to catch a genuine duplicate.

### Verification after the round

All twenty-six guard harnesses and all eleven lint guards green. `stats/` is untouched by this round
and its Go and SPA suites were verified green earlier in the run.

## Pass 8 — full roster

**Slot verdicts.** Primary: clean, unhedged. Adversarial: clean. Principles: clean. Code review
(low): 1 (F24).

Primary added evidence the parent had not: F18's duplicate case runs against a *caller-supplied*
table, which takes the plain-read branch and never reaches the heredoc extraction — so F21's scoping
could not have narrowed it, independently of the parent's positive control. Two lines of evidence
agreeing from different directions.

### Considered and declined, by the slots themselves

Both are recorded because a declined finding is evidence about the review, and because a later reader
should not have to rediscover that they were examined.

- **Adversarial** examined two attack scenarios on the heredoc extraction — a second heredoc sharing
  the delimiter, and a reformatted `<<-` opener — and judged neither blocking: the first is
  contingent on a future guard edit and would abort loudly rather than mispass, the second makes
  every lookup fail loudly under `set -e`. It noted the extraction's comment claims slightly more
  than the mechanism guarantees, and called that a follow-up rather than a fix round.
- **Principles** examined the same delimiter coupling and declined it for the same reason, adding
  that it extends the read-the-guard's-source house style this panel already endorsed.

### F24 — the operator's call, taken

Three slots reported ready. The parent had committed to returning the continue-or-stop decision to
the operator if pass 8 surfaced anything, and did. The operator chose to fix F24 and re-run the
raising slot plus Primary, rather than carry it as a follow-up or withdraw it.

The finding is real and the parent reproduced it: line 140 lacks the `|| rc=$?` its siblings at 111
and 126 both carry, so on the very regression the case exists to catch, the harness stops after two
passing lines with no failure line and twenty-two checks unrun.

## Fix round 8

Fix subagent: one dispatch, model: `sonnet` (recorded `models.panelFix`). F24 fixed, not disputed.
Task 1 `1a8d19d` → `99e8102`; Task 2 `4e01b4e` → `458b8c2` (rewritten by the rebase, content
unchanged). Diff: 4 insertions, 3 deletions, one file — scoped deliberately, since unrelated edits
would have restarted a review that was otherwise finished.

**Bounced findings this pass:** none.

### Mutations this round proved

fix-mutation: scripts/test-check-contract-budget.sh — reverted the read's unterminated-line handling — the unterminated-row case printed a failure line and the run continued to its tally, twenty-three passing and one failing across all twenty-four cases, where it had previously stopped after two
fix-mutation: scripts/test-check-contract-budget.sh — changed the case's expected value while leaving the handling correct — the case failed, showing the assertion still discriminates rather than having been loosened into always passing
fix-mutations-total: 2

The parent re-ran both itself, one mechanism per mutation, in separate scratch copies. The first is
the one that matters: it demonstrates the repair achieved what F24 asked for — the case now *reports*
the regression it exists to catch, instead of killing the run silently.

### Pass 9 — targeted, by the operator's decision

Slots re-run: **Code review (low)**, which raised F24, and **Primary** as the standing integration
check. Principles and Adversarial are *not re-run — subject unchanged*: both reported clean at pass 8
against a branch this round changed by four lines in a file neither had a finding open against, and
not re-running them neither closes nor softens anything.

This is narrower than the escalation ladder's three-or-more-fix-rounds clause would give. The
operator was shown the state after pass 8 — three slots clean, one low finding — and chose to fix
F24 and re-run the raising slot plus Primary rather than carry it or withdraw it. That is the
operator resolving the scope of the final pass, which is theirs to resolve.

## Pass 9 — targeted, and the panel closes

**Slot verdicts.** Code review (low): clean — the slot that raised F24 confirmed its own finding
closed, by running both mutations rather than reading the fix. Primary: clean, unhedged — ready for
the human review gate, nothing blocking.

**Principles** and **Adversarial**: *not re-run — subject unchanged*. Both were clean at pass 8, and
this round changed four lines in a file neither had an open finding against. Their pass 8 results
stand; not re-running closes, softens and expires nothing.

### Final state

Twenty-four findings across nine passes, every one fixed, none withdrawn and none disputed at the
close. Convergence by pass: 7, 3, 2, 2, 1, 4, 4, 1, 0.

The two upticks are worth naming rather than smoothing over. Pass 6's four came from the first pass
that attacked the fix-of-a-fix; pass 7's four came from the first pass after the parent finally ran
the project's whole `## test` list. Neither was drift — both were the review reaching somewhere it
had not reached before.

### What this change's own rule caught in this change, four times

1. Task 1's case was mutation-proved against a fixture it did not commit.
2. The parent recorded a budget-row edit as exempt from proof, on a reason the guard's own harness
   header contradicts.
3. Fix round 5 declared a mutation count instead of walking its hunks; two of the three unaccounted
   hunks were hiding a Critical and a Major.
4. The parent's own first two verification mutations each altered more than one mechanism, and one
   reported a false survival.

Every one is the same shape the ticket predicted: proof produced where it was asked for, skipped
where it was merely implied. They are recorded here rather than quietly corrected, because a change
arguing that rounds skip implied proof is worth exactly what its own compliance record is worth.
