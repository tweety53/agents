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
commits. A trailered commit touching the plan is what `flow-guard check-task-commit-planning-paths`
reports as a sweep and what `flow-guard check-task-records` reports as an undeclared file, and `flow-guard check-fast-route-record` rejects any trailer on a `/flow-fast` branch,
where the fixes ride `<name>` itself. No guard needs the trailer here — `flow-guard check-task-records`
finds a task's commit by its `**Commit:**` subject, and `flow-guard close-task` never runs on this loop — and
on `spectre/<name>` a `Task-Id: <n>` would name the change's own task of that number. The ticks ride
the archive commit, made before item 4's lint lines because `flow-guard check-task-records` reads a live
unticked task whose subject a commit carries as a violation.

## SKILL.md — 2. Read the bundle and run the eight angles

The harness refuses a subagent's Write of `findings.md` and `report-draft.md` ("Subagents should
return findings as text, not write report files" — kan-1051 self-review F6, kan-1073 self-review
F15), while the pass's case files were written, so the pass returns the two files' text and this
session writes them (decision pass-hands-back-file-text).

### Angle 8 — prompt tuning

The angle borrows from `/claude-api prompt-audit` and its eval-hillclimb guide: describe the
behaviour, never paste the incident's content into the prompt; scale a change to how many cases
show the behaviour, never to one vivid failure; treat a prompt change as a hypothesis proven by a
before/after behavioural probe; one idea per change; and keep context and reasons, which are never
cruft — re-baselining can add text as well as remove it (kan-1032).

- **The four-part picking bar, zero cases a normal result** (decision `picking-bar`) — rejected:
  picking any finding a prompt line could address, because it would turn every round into an
  incident-specific prompt patch (prompt-audit's overgeneralization and patch-accretion rows).
- **The angle asks nothing, every case `noted`** (decision `never-ask-in-run`) — rejected: offering
  cases in step 4's filing prompt, by the operator's instruction.
- **Cases ride the pass's fix branch** (decision `cases-ride-fix-branch`) — rejected: a separate
  `prompt-tuning-<name>` branch, because it adds a second landing per run.

