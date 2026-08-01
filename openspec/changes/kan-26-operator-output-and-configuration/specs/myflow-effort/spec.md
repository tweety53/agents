## REMOVED Requirements

### Requirement: `/myflow-start` asks for an effort level once per change

**Reason**: The capability is renamed to `myflow-planning-effort`. The requirement is not deleted —
it moves under the new capability, where it is restated with the concept named "planning effort" and
the question wording following from that.

**Migration**: Read
`openspec/specs/myflow-planning-effort/spec.md` §"`/myflow-start` asks for a planning effort level
once per change". The rule is unchanged: the creating run asks, a revision round reuses, "creates"
means the state file is absent, and the level is never a command argument.

### Requirement: Effort scales the reasoning spent inside the gates, never the gates themselves

**Reason**: The capability is renamed, and the three level names change with it — `medium` becomes
`default` and `high` becomes `detailed`, so that a level says what it means to an operator choosing
between them. `low` is unchanged.

**Migration**: Read `openspec/specs/myflow-planning-effort/spec.md` §"Planning effort scales the
reasoning spent inside the gates, never the gates themselves". Every constraint carries over
unchanged: every level runs brainstorming, holds the design approval gate, runs writing-plans, and
leaves `tasks.md` at plan quality; no level may switch a gate off; and the review panel's breadth is
never scaled from this field. The operational table moves from the **Effort** section of
`skills/myflow-contracts/state-file.md` to that file's **Planning effort** section.

### Requirement: The chosen effort is recorded in the state file, and its absence is legal

**Reason**: The capability is renamed and the state-file key changes from `effort` to
`planningEffort`.

**Migration**: Read `openspec/specs/myflow-planning-effort/spec.md` §"The chosen planning effort is
recorded in the state file, and its absence is legal". No migration pass is run over existing state
files. A file still carrying `effort` is read as recording the equivalent level — `medium` as
`default`, `high` as `detailed`, `low` as `low` — and is rewritten under `planningEffort` on the
next write it receives, with no announced correction. Where both keys are present the current one
wins; a value outside the mapped three reads as "not recorded" rather than making the file
unparseable. The absent-key exception carries over unchanged. The compatibility read is not
defensive padding: state files on this machine do record a non-null value under the old key, and one
of them belongs to a change that is still open rather than `FINISHED`.
