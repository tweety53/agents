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
reason: a file written before a field existed would otherwise be reported as unparseable and its
change skipped, which is a loud failure for a value nobody had the opportunity to set. The exception
covers a key that is **absent**, which is a different thing from a key that is present and null.

A file carrying the retired `effort` key SHALL be read as recording the equivalent level rather than
as unparseable, and SHALL be rewritten under `planningEffort` on the next write it receives. Where
both keys are present, `planningEffort` SHALL win; a value outside the mapped three SHALL read as
"not recorded" and SHALL NOT make the file unparseable. The mapping is
`openspec/specs/myflow-planning-effort/spec.md`'s, and is not restated here.

**A state file that is missing or unparseable SHALL be reported and skipped, never rebuilt from
inference.** No command infers a state file's contents, so an unreadable file names itself in the
reading command's own output and its change is omitted from that command's report. The `worktrees`
map is therefore never reconstructed by scanning, and a change whose state file cannot be read is
visible as unreadable rather than silently replaced by a guess.

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

#### Scenario: An unreadable state file is named rather than rebuilt

- **WHEN** a command reads a state file that is missing or unparseable
- **THEN** it names that file in its own output and omits the change from its report
- **AND** it writes nothing to that file and reconstructs no `worktrees` map

#### Scenario: No field records testing

- **WHEN** the state file is read after `/myflow-do` has produced a test guide
- **THEN** there is no `tested` field, because no command observes whether the human ran the apps

#### Scenario: An absent optional field is legal

- **WHEN** a command reads a state file carrying neither `planningEffort` nor `models`
- **THEN** the file parses normally with both read as not recorded
- **AND** the file is not reported as unparseable on that account

#### Scenario: The retired key does not make the file unparseable

- **WHEN** a command reads a state file carrying `effort`
- **THEN** the file parses and the level is read as the equivalent under the current key
- **AND** the next write carries `planningEffort` and no `effort` key
- **AND** the file is not reported as unparseable on that account

## ADDED Requirements

### Requirement: State writes are monotonic

No command SHALL write a state earlier than the one it read. There is no exception.

`prUrl` SHALL NOT be cleared by any command. The one carve-out this rule previously carried allowed
`/myflow-status` to clear a recorded `prUrl` when a PR CLI conclusively answered that no pull request
existed for the branch; that command no longer makes any network call, so the evidence the carve-out
required can no longer be gathered, and a rule whose precondition can never be met is removed rather
than left to read as live guidance.

#### Scenario: A recorded PR URL is never cleared

- **WHEN** any command reads a state file recording a `prUrl`
- **THEN** it carries that value forward verbatim
- **AND** no command clears it, whatever the pull request's actual state

#### Scenario: An earlier state is never written

- **WHEN** a command that accepts `IN_PROGRESS` completes for a change at `IN_PROGRESS`
- **THEN** it writes `IN_PROGRESS` back, and never `STARTED`

## REMOVED Requirements

### Requirement: Self-heal validates state against artifacts

**Reason**: State self-heal is removed from the pipeline. `/myflow-status` was the only command that
performed it, and it no longer validates a state file against artifacts — it reads the state file
plus local git and reports what it finds. A requirement obliging every command to validate, rewrite
and announce describes behaviour no command has.

**Migration**: A state file that is missing or unparseable is **reported and skipped**, per the
scenario `An unreadable state file is named rather than rebuilt` in
**The state file carries only fields with a live consumer** above. Artifacts remain the source of
truth in the sense that matters — the pipeline still derives its answers from the worktree, the plan
and git — but nothing rewrites the cache from them, and nothing announces a correction.

**The cost is stated rather than left implicit.** A state file that goes stale, is contradicted by
the artifacts on disk, or records a `prUrl` for a pull request that no longer exists is now never
corrected by anything, and never flagged with a `⚠`. This was raised during planning and the
decision was to delete the mechanism rather than move it to another command.

### Requirement: State writes are monotonic with one exception

**Reason**: Renamed and narrowed. The exception in its title was the `prUrl` clear, which depended on
a conclusive PR-CLI probe that `/myflow-status` no longer performs. With the exception gone the title
states something false about the rule beneath it.

**Migration**: **State writes are monotonic** above carries the monotonicity rule unchanged and
states explicitly that `prUrl` is never cleared. No behaviour is lost that a command still performs.
