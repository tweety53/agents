# Review panel — kan-82-cut-myflow-per-command-token-overhead

Panel model: **Sonnet**, from the state file's `models.reviewPanel`. Fix waves: **Sonnet**, from
`models.panelFix`. Both are operator choices recorded at the creating `/myflow-start` run, overriding
the pipeline defaults of Sonnet / Opus.

## Roster, and why each optional slot fired

| Slot | Lens | Required? | Model | Selected because |
|------|------|-----------|-------|------------------|
| 0 | Primary — plan alignment + code quality | always | sonnet | — |
| 1 | Defect hunt | always | sonnet *(see substitution)* | — |
| 2 | Principles — Merged | always | sonnet | — |
| 3 | Security | conditional | sonnet *(see substitution)* | the new guard does path handling, reads an env override, and globs a directory; `.myflow/project.md` is tracked and PR-editable |
| 4 | Adversarial | conditional | sonnet | diff is ~1800 changed lines, far past the ~300 trigger |
| 5 | Principles — Lens B, simplicity & state | conditional | sonnet | past the ~200 changed-line trigger |
| 6 | Principles — Lens C, robustness & ops | conditional | sonnet | new error handling, an environment-variable override, and a config edit |

**Substitution, recorded rather than passed off as a clean pass.** Slots 1 and 3 are specified to be
dispatched by `subagent_type: bugbot` and `subagent_type: security-review`, carrying their own agent
definitions and taking no model override. **Neither agent type exists in this harness.** Both slots
were therefore run as `general-purpose` subagents reading the corresponding prompt files
(`bug-hunter-reviewer-prompt.md`, `security-reviewer-prompt.md`) on the panel's model. The lens is
preserved; the agent definition is not. Their ledger entries read `sonnet (prompt-file substitute)`
rather than `unknown (agent-defined)`, because in this run the dispatcher *did* choose the model.

## Pass 1 — full roster

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Adversarial, Primary, Defect hunt | Major | `skills/myflow-contracts/pipeline.md` | the term "the pre-check" used three times in the core, its defining paragraph moved to the appendix, which no run loads |
| F2 | Primary, Defect hunt, Adversarial | Major | `skills/myflow-contracts/jira-integration.md:155,233` | two citations never repointed; the content they promise moved to `jira-followups.md` |
| F3 | Adversarial | Major | `skills/myflow-contracts/jira-followups.md` | ten citations back into `jira-integration.md` kept a now-false `above`; line 139 names a rule with no citation at all |
| F4 | Lens C | Major | `scripts/check-contract-budget.sh` | `wc -c` assigned bare under `set -e`, so an unreadable file exits 1 — the code reserved for over-budget — instead of the documented 2 |
| F5 | Lens C | Major | `scripts/test-check-contract-budget.sh` | every exit-2 case was a directory-shape variant; the file-level read failure had no fixture, so F4 would have shipped through the harness |
| F6 | Lens B | Major | `rules/myflow-manual-review.mdc:39` | "Eight narrower contracts" is a hand-maintained count nothing checks, in a change that touches a file documenting this exact anti-pattern |
| F7 | Adversarial | Minor | `skills/myflow-contracts/pipeline-rationale.md` | six more stale cross-file position words, editor-facing |
| F8 | Security | Minor | `scripts/check-contract-budget.sh` | a filename containing a newline breaks `awk -v`, aborting the whole run instead of reporting one bad name |
| F9 | Principles, Primary | Minor | `scripts/check-contract-budget.sh` | two budget rows drifted from achieved-size + 25%, the table having been computed before a later fix changed those files |
| F10 | Lens C | Minor | `scripts/check-contract-budget.sh` | the over-budget message states no remedy while its sibling message does |
| F11 | Lens C, Defect hunt | Minor | `scripts/check-contract-budget.sh` | `[ -e ]` is false for a broken symlink, so one dangling `*.md` reported "no .md files" while real ones sat unexamined |
| F12 | Primary | Minor | `openspec/changes/.../tasks.md` | groups 2, 3 and 6 unticked for verified-complete work |
| F13 | Defect hunt | Minor | `skills/myflow-contracts/jira-followups.md:1,8` | `# Follow-up issues` and `### Follow-up issues` duplicate each other |
| F14 | Lens B | Minor | `skills/myflow-contracts/SKILL.md`, `rules/myflow-manual-review.mdc` | the mirrored-heading rule is new unguarded maintenance surface, asymmetric with the byte budget that did get tooling |

findings-total: 27
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
finding-status: F14 withdrawn — the rule is already disclosed as unenforced in SKILL.md, matching how this corpus treats its other judgment rules; drift costs a rationale paragraph filed under a stale heading, which is an editor-facing annoyance and never a runtime defect
finding-status: F15 fixed
finding-status: F16 fixed
finding-status: F17 withdrawn — pre-existing duplication carried across by a verbatim move, not introduced by this change; fixing it would edit a moved line, which the partition rule forbids
finding-status: F18 fixed
finding-status: F19 fixed
finding-status: F20 fixed
finding-status: F21 fixed
finding-status: F22 fixed
finding-status: F23 fixed
finding-status: F24 fixed
finding-status: F25 fixed
finding-status: F26 fixed
finding-status: F27 fixed

## What pass 1 changed about the change itself

**F1, F2, F3 and F7 are one defect, not four.** A split moves passages across a file boundary that
did not exist before, so a position word true inside one file becomes false across two. The planning
decision permitting only *path* repointing is what left them; the operator reversed it on the
panel's evidence, and a passage may now have a stale `above` or `below` **deleted** — a deletion,
never a substitution. The reversal is recorded as `position-words-deletable` in `design.md`, as a
requirement with two scenarios in the delta spec, and in the plan's global constraints.

**F1 is the one that mattered most.** Every other instance was editor-facing or cost a reader one
hop; F1 left a **core** term undefined, in a file every `/myflow-*` command loads, with its
definition in a file none of them do. That is precisely the failure a contract split exists to
avoid, reached by obeying two individually sound rules.

**F2 had been withdrawn once and was re-raised by three independent slots.** The operator withdrew it
at the task-1.1 review on the reasoning that a real section by that name is still below in the file,
so nothing false is stated. Pass 1's primary, defect-hunt and adversarial slots each reached it
independently with a sharper argument — the citations promise *specific content* now reduced to a
five-line redirect stub — and the operator overturned their own withdrawal on that evidence.

**Two guards were shown to be blind here, and neither blindness is a bug in them.**
`scripts/check-references.sh` passes a citation whose section has been hollowed out, because a
heading of that name still exists — and the mirroring rule guarantees one always will, in both
files. The line-multiset check sorts and drops blanks, so it cannot see reordering, duplication, or a
paragraph whose meaning depended on its neighbours. Both facts are why F1 through F3 survived to the
panel, and both are stated here so the next reader does not mistake a green guard for a verified
claim.

## Pass 2 — full roster, escalated

**Escalated automatically rather than targeted, and the trigger is recorded:** the fix round changed
**441 lines across 11 files** — 375 insertions, 66 deletions — which is past the ~150-line threshold
that forces a full re-run. It also altered a delta spec, which is independently a full-escalation
trigger. No operator was asked, per the escalation rule.

Diff: `.superpowers/sdd/final-review.diff`, rewritten after the fix round. The fix round's own diff
is `.superpowers/sdd/fix-round-1.diff`.

**Slot collapse, recorded as a deviation.** Pass 2's primary and defect-hunt lenses were dispatched
in **one** prompt rather than two. The panel contract says two slots are never merged into one
prompt, and this run broke that. The remaining five slots were dispatched separately. It is recorded
here rather than left to be inferred from the agent count.

### Pass 2 findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F15 | Lens B | Major | `skills/myflow-contracts/SKILL.md:38` | "**Two** contracts are large enough…" — a hand-maintained count above the two-row table that already names both, written *while fixing* F6, which is the identical defect |
| F16 | Lens B | Minor | `skills/myflow-contracts/pipeline.md:8` | "the **four** contract files beside it" — pre-existing stale count that this change made wronger by adding a ninth file |
| F17 | Lens B | Minor | `skills/myflow-contracts/pipeline-rationale.md` | "a cache offers sixteen indices" duplicates `workspace-isolation.md`; pre-existing, carried across by the move rather than introduced |
| F18 | Lens C | Minor | `scripts/check-contract-budget.sh` | `2>/dev/null` cannot suppress bash's own failed-open message, so the operator saw a doubled, format-inconsistent diagnostic |
| F19 | Lens C | Minor | `scripts/check-contract-budget.sh` | the over-budget message advertised only raising the budget, never trimming the file — the wrong default for a ratchet |
| F21 | Security | Major | `scripts/check-contract-budget.sh` | a `.md`-named symlink was followed, so `wc -c` sized whatever it pointed at anywhere readable — an existence-and-size oracle, with the target's byte count printed when the name matched a budget row |
| F22 | Security | Minor | `scripts/check-contract-budget.sh` | a name that is a budget entry plus trailing whitespace never matches `*.md`, and the sentinel then misreported a populated directory as empty |
| F24 | Primary | Minor | `openspec/changes/.../tasks.md` task 6.1 | "Expected: no `<` lines at all" was never updated after the position-word reversal, so the box was ticked on a command that cannot produce its stated output |
| F25 | Primary | Minor | `openspec/changes/.../tasks.md` task 4.1 | the index-agreement check used a substring grep that matches `myflow-review.md` inside a `.mdc` path and `project.md` inside `.myflow/project.md`, so its "Expected: empty" was unreachable |
| F26 | Adversarial | Major | `skills/myflow-contracts/pipeline-rationale.md` | the Self-review "Which file to change first" paragraph was relocated core→appendix, which is neither of the two permitted edits |
| F27 | Adversarial | Minor | `skills/myflow-contracts/pipeline-rationale.md:104` | an `above` was deleted from a bare reference with no citation behind it, leaving "the missing-rather-than-dropped rule" unlocatable |
| F23 | Principles | Major | `scripts/check-contract-budget.sh` | the `jira-followups.md` budget row went stale again — recomputed mid-round, then the F13 title fix changed the file |
| F20 | self | Minor | `skills/myflow-contracts/SKILL.md` | "the appendices are absent from the table above" became self-contradictory once an appendix table was added directly above it |


**F15 is the finding worth keeping.** It is the same defect as F6 — a hand-maintained count beside
the table that makes it redundant — written into `SKILL.md` *by the fix for F6*, two files away, in
the same round. A reviewer caught it; nothing mechanical would have. That is the argument for the
panel in one example.

**F4 and F5 were verified by reverting the fix.** Lens C put the bare `wc -c` assignment back and
re-ran the harness: the unreadable-file and dangling-symlink cases both failed. That is evidence the
new fixtures exercise the guarded path rather than passing vacuously — a check on the check.

**Two hypotheses were tested and disproved rather than assumed.** Lens C predicted the classic
herestring gotcha would silently drop the last budget row, and found that `<<<` appends its own
trailing newline so every row is read; and a directory named `*.md` was found to fail closed at exit
2 rather than being sized. Both are recorded because a disproved hypothesis is worth as much here as
a confirmed one.

**F21 was introduced by the fix for F11.** Adding `[ -L "$path" ]` to the empty-glob sentinel made a
dangling symlink reachable, and a *valid* one readable — turning a size check into a path oracle. The
guard now refuses any symlink outright, before reading it, because a contract is a regular file
tracked in git and a symlink has no size of its own to ratchet. The harness gained a case that
creates a symlink to a 9000-byte file outside the fixture tree and asserts that number never appears
in the output.

**F23 is the budget table drifting for the third time, and the cause is structural rather than
careless.** It is recomputed from the tree, and every later fix changes the tree. The table is
therefore regenerated **once, as the last action before handoff**, after every other finding is
closed — which is what task 5.1 meant by placing the guard last, applied to the fix rounds as well as
to the tasks.

**The Merged-lens slot was asked whether the four plan amendments lowered a bar to fit finished
work**, and reported that three of them *narrow* what is permitted — paragraph granularity preserves
the line-multiset check's power, the struck heading-promotion clause resolves a contradiction in
favour of the spec, and position-word deletion stays visible in the same check — while the fourth's
cost, the unreachable token target, is stated plainly rather than absorbed. That question is the one
worth asking of a change that amended its own plan four times, and it is recorded here with its
answer.

**F24 and F25 are the same fault: a checkbox ticked against a stale expectation.** Both tasks'
verification commands ran and both were marked complete, but neither command produced the output its
"Expected" line promised — 6.1's because the position-word reversal made a non-zero `<` count
correct, 4.1's because its substring grep matched two paths that are not files in the directory it
was checking. Neither hid a real defect: the index genuinely agrees with the directory, and every
`<` line is genuinely a permitted edit. What they hid was that **nobody had re-run the command as
written**, which is the thing a ticked box is supposed to assert.

The corrected commands and the corrected expectations are now in the plan, and the box for 6.1 now
states the real counts — 14 for the pipeline pair, 13 for the jira triple — rather than zero.

**F27 was fixed by applying this change's own newest rule to itself.** A passage that depends on
another by reference is classified with that passage — and this one named "the missing-rather-than-
dropped rule" with no citation to carry the reference across a file boundary, so deleting its
`above` left a reader in the appendix with no way to find the rule at all. The paragraph moved back
into `pipeline.md`, where the rule it names lives, and its `above` is true again.

**F26 went to the operator and was decided against my reading.** I argued the two-permitted-edits
rule governs what may change *inside a line*, not which file a paragraph is classified into — every
paragraph in an appendix got there by being moved — and that nothing in the core references this one,
so it is unlike F1, where a core term was left undefined. The operator chose to move it back anyway,
which also removes the inconsistency the reviewer noticed: both "Which file to change first"
paragraphs now sit in `pipeline.md`. Recorded as fixed, and the disagreement is recorded with it.

## Outcome

**Two passes, seven slots each, 27 findings: 26 fixed, 1 withdrawn by the operator.** No finding is
open at any severity, and every guard and harness this repository declares passes on the final tree.

**Three findings were introduced by fixes to earlier findings** — F21 by the fix for F11, F15 by the
fix for F6, F23 by the fix for F13. That is the argument for re-running the panel after a fix round
rather than trusting the fix, and it is recorded here because the count is the evidence.

**What a `/myflow-start` run loads from these two files fell from 156109 bytes to 102945 — a 34%
cut.** The linked issue's target for `pipeline.md` alone was roughly 8k tokens; it is not reached,
for the reason `paragraph-granularity-partition` records, and the shortfall is stated in the manual
test guide's `## Known incomplete` rather than absorbed.

