# flow-self-review — rationale

Reasoning behind `skills/flow-self-review/SKILL.md`: the incidents its rules were written against.
Moved here verbatim from the run-loaded file; **no `/flow-self-review` run loads this file.** Each
heading names the source file and the section the passage came from.

## SKILL.md — 7. Write the report, delete the bundle, land it

The exit-2 rule — the guard cannot answer at all is not a violation — was written for:

> (KAN-601)

## SKILL.md — 3. Fix every finding that is not `big`

The pass and the planning run as one-shot `flow-medium` dispatches on `opus`, and every fix is
planned in full — brainstorming, artifacts and a writing-plans `tasks.md`, every question
auto-resolved to its recommended option — at the operator's explicit request (2026-10-14). The
planned `tasks.md` is what lets the fixer run at `flow-low` on `opus`: the flow-eval implementation
runs showed a cheap implementer succeeding on detailed writing-plans tasks. Re-reviews read a delta
to confirm a fix, so they run `flow-low` on `sonnet`. The pairs are stated in this file, and
**Pipeline defects found mid-run** cites them for its in-run loop, which matches this one at the
operator's request (2026-10-14).

The fixer's commits carry no `Task-Id:` trailer and never write
`<agents repo>/spectre/changes/`, following `/flow`'s own split of task commits from planning
commits. A trailered commit touching the plan is what the since-deleted `check-task-commit-planning-paths`
reports as a sweep and what the since-deleted `check-task-records` reports as an undeclared file,
where the fixes ride `<name>` itself. No guard needs the trailer here — the since-deleted `check-task-records`
finds a task's commit by its `**Commit:**` subject, and `flow-guard close-task` never runs on this loop — and
on `spectre/<name>` a `Task-Id: <n>` would name the change's own task of that number. The ticks ride
the archive commit, made before item 4's lint lines because the since-deleted `check-task-records` reads a live
unticked task whose subject a commit carries as a violation.

## SKILL.md — 2. Read the bundle and run the four angles

The harness refuses a subagent's Write of `findings.md` and `report-draft.md` ("Subagents should
return findings as text, not write report files" — kan-1051 self-review F6, kan-1073 self-review
F15), while the pass's case files were written, so the pass returns the two files' text and this
session writes them (decision pass-hands-back-file-text).

### Angle 3 — prompt tuning

The angle borrows from `/claude-api prompt-audit` and its eval-hillclimb guide: describe the
behaviour, never paste the incident's content into the prompt; scale a change to how many cases
show the behaviour, never to one vivid failure. The replay/apply skill and its case files were
removed (2026-11-20): a finding is filed as one Jira task whose note opens with its trace — prompt
file, review slot and round, the change's fix sha — which is all a later prompt change needs to be
tied back to the review result that prompted it.

## Moved by the 2026-11-13 prompt audit — incident records

Each passage below sat inline in a run-loaded file beside the rule it motivated; the rule and its
in-sentence reason stay there. The quoted fragment before each dash is the end of that rule, for
attribution; the text after it is the passage, verbatim.

### SKILL.md — 2. Read the bundle and run the four angles

- *action is `big`, reaches production (`rules/never-touch-production.mdc`) or deletes what cannot be restored* — (kan-995: "feature/cardio lacks develop's fix 98da4aad; a branch decision for the operator" was recorded `noted` and relayed as an open decision, and the operator then had the run merge the branch)

### SKILL.md — 3. Fix every `flow-fix` finding that is not `big`

- *the objection as a constraint, never offered for filing in step 4* — (kan-982: the `Task-Id:` guard and the scope check were dropped as imprecise and offered for filing; the operator answered "fix them now")
- *never left as a forward-only fix with the repair an operator's option* — (kan-995: decision `k995-rows-left` left three mis-stamped ledger rows wrong, and `flowd restamp-dispatch`, `b2555b98`, repaired them once the operator asked)
- *and its exit code and `red:` lines are read as this verify's* — (kan-1051 self-review, F18: the suite ran twice at one HEAD, about 5 min a loop)
