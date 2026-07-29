# myflow-plan-provenance Specification

## Purpose

Define how a plan states the provenance of its claims about the world, and what an
unverified claim obliges of the implementer who reads it. The guard enforces that provenance
is *stated*, which is what a script can check — it does not make plans correct.

## Requirements

### Requirement: A plan states the provenance of every claim it makes about the world

A plan's fenced code blocks and numeric claims SHALL carry a tag recording how the claim was
established. Four tags exist, and no others.

On a fenced code block, in its info string:

- `verified:<how>` — the symbol and its signature were checked against the real artifact. `<how>`
  names the evidence concretely, e.g. `verified:javap intellij.platform.diff.jar`,
  `verified:compiled in-tree`, `verified:PlatformActions.xml:452`.
- `unverified:<what-to-check>` — a hypothesis, naming the check that would settle it.

On a numeric or factual claim, as an HTML comment on the following line:

- `measured:<command> @ <ref>` — the number came from running that command at that revision.
- `predicted:<what-confirms-it>` — a forecast, naming the command that settles it.

Every fenced code block in a plan SHALL carry `verified:` or `unverified:`. In-repo code satisfies
this cheaply with `verified:compiled in-tree`; the cost falls on claims about things outside the
repository.

#### Scenario: A verified block names its evidence

- **WHEN** a plan contains a code block calling an API outside the repository
- **AND** the symbol and signature were checked against the real artifact
- **THEN** the block is tagged `verified:<how>` and `<how>` names the artifact or command consulted

#### Scenario: A block that could not be checked is marked, not omitted

- **WHEN** a plan needs a code block whose API could not be verified while planning
- **THEN** the block is tagged `unverified:<what-to-check>` and remains in the plan
- **AND** the tag names the check that would settle it

#### Scenario: An untagged code block is a defect

- **WHEN** a plan contains a fenced code block with neither tag
- **THEN** the plan is not ready to publish

### Requirement: A number with no provenance may not appear in a plan

A numeric claim SHALL carry `measured:` or `predicted:`. A number carrying neither SHALL NOT appear
in the plan at all.

Stating an unsourced number is itself the violation. Labelling alone would not have helped the case
this rule exists for: a baseline of "194 tests" was not mislabelled, it was invented, and the plan
that carried it was wrong at every task that read it.

#### Scenario: A measured baseline names how it was measured

- **WHEN** a plan states a test-count or other measured baseline
- **THEN** it carries `measured:<command> @ <ref>` naming the command and the revision

#### Scenario: A forecast is distinguishable from a measurement

- **WHEN** a plan states a number that could not be measured while planning, such as a count after a
  change that has not been made
- **THEN** it carries `predicted:<what-confirms-it>` rather than `measured:`
- **AND** the reader can tell the two apart without consulting anything else

#### Scenario: An unsourced number is removed rather than tagged

- **WHEN** a number cannot be measured and no command would confirm it
- **THEN** it does not appear in the plan

### Requirement: Provenance is enforced mechanically, and only where it applies

A guard SHALL fail when a plan contains an untagged fenced code block or an untagged numeric claim,
reporting `file:line`.

The guard SHALL check only that provenance is **stated**, never whether it is true. No script can
confirm that a verification was performed; a guard claiming otherwise would be both unimplementable
and dishonest about what it proves.

The guard's scope SHALL be limited to a change's `tasks.md`, excluding archived changes. A guard that
fires outside the file type its rule is about produces false failures, and false failures are
answered with suppression markers that switch off the real checks sharing those lines.

#### Scenario: An untagged block fails the guard

- **WHEN** the guard runs over a plan containing a fenced block with no provenance tag
- **THEN** it exits non-zero and reports the file and line

#### Scenario: A correctly tagged plan is silent

- **WHEN** the guard runs over a plan whose blocks and numbers all carry tags
- **THEN** it exits zero and reports no failures

#### Scenario: Archived plans are not scanned

- **WHEN** the guard runs in a repository containing archived changes with untagged plans
- **THEN** those plans are not scanned and do not cause a failure

#### Scenario: Documentation outside a plan is not scanned

- **WHEN** a skill file, contract or README contains an untagged fenced code block
- **THEN** the guard does not fail — its scope is a change's `tasks.md` only

### Requirement: An unverified block is a hypothesis, not code to transcribe

An implementer given a plan SHALL treat a block tagged `unverified:` as a hypothesis: establish the
real behaviour before writing code against it, and report what was found.

A block tagged `verified:<how>` was checked as stated. If it nevertheless fails to compile or
behave as described, the implementer SHALL report that rather than contorting the code to match it.

#### Scenario: The implementer verifies before writing

- **WHEN** an implementer is dispatched with a task whose plan contains an `unverified:` block
- **THEN** it establishes the real API before writing against it
- **AND** its report states what it found

#### Scenario: A wrong verified tag is reported, not worked around

- **WHEN** a block tagged `verified:` does not compile against the real artifact
- **THEN** the implementer reports the discrepancy
- **AND** does not reshape the code to preserve the plan's claim
