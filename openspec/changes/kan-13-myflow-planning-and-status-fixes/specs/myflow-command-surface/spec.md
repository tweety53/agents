## ADDED Requirements

### Requirement: Change name resolution enumerates every state-tracked change

The shared change name resolution used by every `/myflow-*` command SHALL enumerate open changes
from the union of `openspec list --json` and the names of the files in the project's state
directory (`/Users/tweety53/Agents/myflow/state/<project-key>/*.json`), minus any name whose
`openspec/changes/<name>/` directory has reached `archive/`.

This resolution SHALL be defined once, in the shared `## Change name resolution` section every
`/myflow-*` command reads, so no command's enumeration can drift from another's.

A state-directory file that cannot be read SHALL be reported and skipped from the union, and SHALL
NOT be silently dropped without being named.

#### Scenario: A worktree-only change is discovered

- **WHEN** a change's `openspec/changes/<name>/` directory exists only in a worktree, staged but
  never committed to the main checkout, and its state file exists in the project's state directory
- **THEN** `/myflow-status` lists the change

#### Scenario: Every command shares the same enumeration

- **WHEN** `/myflow-do` or `/myflow-finish` is invoked with no change name and more than one match
  exists after the union
- **THEN** the operator is asked which change, exactly as `/myflow-status` would enumerate the same
  set

#### Scenario: An archived change is not re-surfaced

- **WHEN** a change's state file still exists in the state directory but its
  `openspec/changes/<name>/` directory has moved to `archive/`
- **THEN** it is not included in the enumeration

#### Scenario: An unreadable state file is reported, not silently dropped

- **WHEN** a file in the project's state directory cannot be parsed
- **THEN** it is skipped from the union and named in the report, rather than causing that change to
  be invisible with no explanation
