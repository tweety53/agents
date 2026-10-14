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
commits. A trailered commit touching the plan is what `check-task-commit-planning-paths.sh`
reports as a sweep and what `check-task-records.sh` reports as an undeclared file, and `check-fast-route-record.sh` rejects any trailer on a `/flow-fast` branch,
where the fixes ride `<name>` itself. No guard needs the trailer here — `check-task-records.sh`
finds a task's commit by its `**Commit:**` subject, and `close-task.sh` never runs on this loop — and
on `spectre/<name>` a `Task-Id: <n>` would name the change's own task of that number. The ticks ride
the archive commit, made before item 4's lint lines because `check-task-records.sh` reads a live
unticked task whose subject a commit carries as a violation.
