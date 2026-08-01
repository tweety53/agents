## MODIFIED Requirements

### Requirement: The state file carries only fields with a live consumer

The state file SHALL contain `state`, `branch`, `worktrees`, `artifactUrl`, `jiraIssue`,
`planningEffort`, `models`, `prUrl`, `updatedAt` and `updatedBy`, and SHALL NOT contain `gates`,
`tested`, `originStage`, `REVIEWED_TREE`, `fastPath`, separate `worktree` and `MERGE_BASE` fields,
or `effort`. That is a rule about what a **write** emits: a file this pipeline produces never carries
the retired key, which is a different question from what a read does with one that already does.

`worktrees` SHALL be an object keyed by the **absolute path** of each affected worktree, whose
value is that worktree's merge base, or `null` where no merge base is recorded for that path. This
key set SHALL be the authoritative list of worktrees for a change spanning more than one repository,
and a `null` value SHALL be read exactly as a missing merge base is — never as a licence to infer
one.

`models` SHALL be an object carrying `implementation`, `reviewPanel` and `panelFix`, each naming the
model chosen for that role. Its live consumer is `/myflow-do`, which dispatches on those values;
`myflow-model-policy` is canonical for the roles, their defaults and how an override applies.

`planningEffort` SHALL carry the level chosen for the change's planning. Its live consumer is
`/myflow-start`, which reuses it on a revision round; `myflow-planning-effort` is canonical for the
levels and the default.

`planningEffort` and `models` SHALL each read as "not recorded" when absent, rather than making the
file unparseable. This extends the exception the effort field already carried, and for the same
reason: routing every file written before a field existed through self-heal announces unrecovered
fields and rewrites from artifact inference, which is a loud correction for a value nobody had the
opportunity to set. The exception covers a key that is **absent**, which is a different thing from a
key that is present and null.

A file carrying the retired `effort` key SHALL be read as recording the equivalent level rather than
as unparseable, and SHALL be rewritten under `planningEffort` on the next write it receives. Where
both keys are present, `planningEffort` SHALL win; a value outside the mapped three SHALL read as
"not recorded" and SHALL NOT make the file unparseable. The mapping is
`openspec/specs/myflow-planning-effort/spec.md`'s, and is not restated here.

**A self-heal rebuild SHALL recover the `worktrees` keys rather than emptying the map.** When the
prior file cannot be read, the paths SHALL be recovered by scanning each affected repository for the
worktrees on the change's branch, and each recovered entry SHALL carry `null` for the merge base,
which the scan cannot produce. An emptied map is not a neutral loss: `/myflow-finish`'s preflight and
its unfinished-work gate are both defined as once per recorded worktree, so an empty map makes both
pass having examined none — for a change that may still hold an unmerged worktree. The rebuild SHALL
name `worktrees` in the correction announcement as recovered without merge bases.

The file SHALL continue to live outside the repository at
`/Users/tweety53/Agents/myflow/state/<project-key>/<name>.json`, resolved via `--git-common-dir`,
and SHALL never be staged, committed or archived.

#### Scenario: Writing state carries forward unowned fields

- **WHEN** any command writes the state file
- **THEN** it renders the whole object, re-emitting `artifactUrl`, `jiraIssue`, `prUrl`,
  `planningEffort`, `models` and `worktrees` as read, rather than resetting a field it does not own

#### Scenario: Multi-repo worktrees are enumerable

- **WHEN** a change affects two repositories
- **THEN** `worktrees` has one absolute-path key per repository, and `/myflow-finish` cleans up
  every key

#### Scenario: A rebuild recovers the worktree paths

- **WHEN** self-heal rebuilds a state file it could not read, for a change that still holds a
  worktree on its branch
- **THEN** the rewritten `worktrees` map carries that worktree's absolute path with a `null` merge
  base, rather than being emptied
- **AND** the correction announcement names `worktrees` as recovered without merge bases
- **AND** the next `/myflow-finish` refuses on that worktree rather than passing with none examined

#### Scenario: No field records testing

- **WHEN** the state file is read after `/myflow-do` has produced a test guide
- **THEN** there is no `tested` field, because no command observes whether the human ran the apps

#### Scenario: An absent optional field is legal

- **WHEN** a command reads a state file carrying neither `planningEffort` nor `models`
- **THEN** the file parses normally with both read as not recorded
- **AND** no self-heal correction is announced on that account

#### Scenario: The retired key does not make the file unparseable

- **WHEN** a command reads a state file carrying `effort`
- **THEN** the file parses and the level is read as the equivalent under the current key
- **AND** the next write carries `planningEffort` and no `effort` key
- **AND** no correction is announced on that account
