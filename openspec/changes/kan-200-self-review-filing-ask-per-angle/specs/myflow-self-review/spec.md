## MODIFIED Requirements

### Requirement: One combined reasoning pass, not four separate dispatches

Self-review SHALL answer all **five** angles below in a single reasoning pass, fed the gathered
bundle plus the live session's own context — never as five separate subagent dispatches:

| # | Angle | Label |
|---|-------|-------|
| 1 | Problems encountered, and what pipeline change would avoid them | `myflow-fix` |
| 2 | Token/time cost, and what would reduce it without quality loss | `myflow-cost` |
| 3 | What went well, and how to reproduce it | `myflow-improvement` |
| 4 | What could be automated or moved to a script | `myflow-automation` |
| 5 | What could move to the Go app or its persistent storage | `myflow-stats-app` |

Angle 5 SHALL cover both **records** — which artifacts the pipeline writes to files today would be
better held in the stats daemon's database and queryable across runs — and **logic** — which
derivation work now performed by a Bash guard or by the agent itself could move into the `myflow`
CLI or the daemon. It SHALL NOT extend to what the SPA should display.

Each angle SHALL produce zero or more findings. An angle that produces none SHALL say so
**explicitly**, the way `## Decisions` and `## Open questions` are present-but-empty rather than
absent. An angle that is silent SHALL NOT be treated as an angle that yielded nothing: the two are
indistinguishable to a reader, and that is how a write-only angle passes unnoticed.

#### Scenario: One pass produces all five angles

- **WHEN** self-review runs
- **THEN** a single reasoning pass produces findings for all five angles, not five separate
  dispatches

#### Scenario: An angle with nothing to report says so

- **WHEN** one angle produces no findings at all
- **THEN** that angle's section is present and carries an explicit statement that it produced none,
  rather than being omitted or left empty

### Requirement: Each actionable finding gets its own Jira filing ask

The filing ask SHALL cover the findings of **every** angle, not only the problems angle.

Before any prompt fires, self-review SHALL explain **each** finding in the message body, stating
three things: what was observed, what breaks because of it, and what the fix would be. A prompt's
option text SHALL NOT be relied on to carry that explanation. The operator SHALL have the
explanation in front of them before the decision is recorded, because a filed issue is durable and
already on the board — an explanation that arrives after the filing describes something the operator
did not agree to.

The decision SHALL then be recorded as **one multi-select prompt per angle**, listing that angle's
findings, defaulting to filing none of them. A bare observation with no concrete change implied
SHALL NOT appear as an option.

A filed issue SHALL carry every label on the change's linked issue, plus `AI-generated`, **plus the
label naming the angle that produced it** from the table above, and SHALL link to that issue when
one exists, per **Labels on issues the pipeline creates**
(`skills/myflow-contracts/jira-integration.md`). A filing failure SHALL degrade to the standard
`⚠ Jira: skipped — <reason>` line and SHALL NOT stop self-review, per **Never blocking**
(`skills/myflow-contracts/jira-integration.md`).

This explain-before-filing rule SHALL bind every filing ask the pipeline offers, not only
self-review's — `/myflow-finish` run 1's follow-up filing included.

#### Scenario: A cost or what-went-well finding is offered for filing

- **WHEN** an angle other than problems produces a finding naming a concrete pipeline or script
  change
- **THEN** that finding is explained in full and appears as an option in its angle's filing prompt,
  exactly as a problems-angle finding does

#### Scenario: The explanation precedes the prompt

- **WHEN** self-review has findings to offer
- **THEN** every finding's observation, consequence and proposed fix are stated in the message body
  before the first filing prompt is presented

#### Scenario: A filed issue carries its angle's label

- **WHEN** the operator selects a finding from the automation angle for filing
- **THEN** the created issue carries `myflow-automation` in addition to the linked issue's labels
  and `AI-generated`

#### Scenario: Operator declines every finding in an angle

- **WHEN** the operator selects nothing in an angle's filing prompt
- **THEN** no issue is created for that angle, and self-review continues to the next angle

## ADDED Requirements

### Requirement: The report records each finding in a parseable form

`docs/self-review/<name>-self-review.md` SHALL carry one section per angle, all five present, in the
order the angle table states.

Each finding SHALL be recorded as a single line naming, in a fixed form, the label of the angle that
produced it, the finding itself, and its disposition — the issue key when it was filed, or an
explicit declined marker when it was not:

```markdown
- **[myflow-cost]** Every panel slot gathers the same context independently — filed: KAN-201
- **[myflow-fix]** Preflight compares against a stale local base ref — declined
```

An angle that produced no findings SHALL carry an explicit none-marker in place of finding lines.

The angle label on a finding line SHALL match the section the line sits under. A finding recorded as
filed SHALL name an issue key.

#### Scenario: A filed finding names its issue

- **WHEN** a finding was filed as KAN-201
- **THEN** its line in the report names `myflow-cost` and `filed: KAN-201`

#### Scenario: An empty angle carries a marker

- **WHEN** the automation angle produced no findings
- **THEN** its section is present and carries the none-marker rather than being empty or omitted

### Requirement: A guard checks every self-review report

The repository SHALL carry a guard that checks every report under `docs/self-review/` against the
report shape above: all five angle sections present; each section carrying either at least one
finding line or the none-marker; each finding line parseable, with its angle label matching its
section and a filed finding naming an issue key.

The guard SHALL report per-member coverage and SHALL exit 0 when clean, 1 when it finds violations,
and 2 when it cannot answer at all — the exit-code contract every guard in this repository carries.

Reports written before this requirement existed SHALL be **declared by name, with a reason**, in the
guard's own source, exactly as a legitimately-zero corpus member is declared under **Requirement:
Zero coverage SHALL be declared or SHALL be a violation** (`openspec/specs/agents-repo-verification/spec.md`).
A report absent from that declared set SHALL satisfy every rule. The guard SHALL NOT infer that a
report predates the rule from the report's own content — a marker a new report can forget to write
is a mechanism for passing without being checked, which is the outcome the declaration exists to
prevent.

#### Scenario: A new report omits an angle

- **WHEN** a report not on the declared pre-rule list is missing the angle-5 section
- **THEN** the guard names the report and the missing section, and exits 1

#### Scenario: A pre-rule report is declared

- **WHEN** the guard runs against the reports that predate this requirement
- **THEN** each is reported as declared, with its reason, and the guard exits 0

#### Scenario: A finding claims to be filed but names no issue

- **WHEN** a finding line is marked filed and carries no issue key
- **THEN** the guard names that line and exits 1
