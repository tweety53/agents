## REMOVED Requirements

### Requirement: The guide is a behaviour checklist at capability scope

**Reason**: The generated guide is removed from the pipeline. `/myflow-do` no longer writes
`docs/manual-test/<name>.md`, and that directory is deleted.

**Migration**: The operator still runs the apps at the `IN_PROGRESS` gate. What they run comes from
the run instructions the `/myflow-do` handoff prints, governed by the new requirement **The
`IN_PROGRESS` handoff carries run instructions** in `myflow-handoff-output`. Guides already
committed remain in git history.

### Requirement: The guide keeps the shapes its guards read

**Reason**: The two guards that read those shapes no longer read them.
`check-unfinished-work.sh` drops its guide signals, and no other guard reads the guide.

**Migration**: Unfinished work is recorded where it already was — unticked checkboxes in
`openspec/changes/<name>/tasks.md`, and open findings in `.superpowers/sdd/final-review-panel.md`.
Those two signals are unchanged and are what `check-unfinished-work.sh` now reads.

### Requirement: A repository with no runnable application states its checks as commands

**Reason**: The guide this requirement shaped no longer exists.

**Migration**: The rule survives, applied to the handoff rather than to a file: see the scenario
**A repository with no runnable application lists its checks** under the new requirement **The
`IN_PROGRESS` handoff carries run instructions** in `myflow-handoff-output`.
