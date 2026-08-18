## ADDED Requirements

### Requirement: A corpus-scanning guard SHALL report what it checked, per member

A guard that discovers its corpus from the tree — rather than being handed a single explicit target —
SHALL report, in its verdict line, how many items it checked **for each member of that corpus**, and
SHALL name every member whose count is zero.

A clean exit reporting nothing about coverage is indistinguishable from a guard that examined nothing,
and a guard that examines nothing reports clean forever. The count is what separates the two, and it
SHALL be present on a **passing** run: a fact visible only when something is already broken cannot
reveal a rule that was never doing anything.

**A guard handed an explicit single target is not in scope** and SHALL NOT be required to report a
count. It has no member it could silently skip, so a constant would be ceremony.

#### Scenario: A rule covers nothing for one corpus member

- **WHEN** a corpus-scanning guard runs against a tree it considers clean
- **AND** one of its rules resolves to an empty required set for one member of the corpus
- **THEN** the verdict names that member and reports its coverage as zero

#### Scenario: A guard takes an explicit single target

- **WHEN** a guard is invoked with an explicit target rather than discovering a corpus
- **THEN** it reports no per-member coverage, and this requirement does not apply to it

### Requirement: Zero coverage SHALL be declared or SHALL be a violation

Each corpus-scanning guard SHALL carry an explicit set of corpus members that legitimately check
nothing. A member whose coverage is zero and which is **absent** from that set SHALL be a violation:
named, with its reason, and a non-zero exit.

A member's presence in that set SHALL be a written declaration in the guard, never inferred from the
tree. Inferring it would restate the same assumption the zero already encodes, and would pass exactly
the case this requirement exists to fail.

**Reporting the zero without failing SHALL NOT satisfy this requirement.** The defect that motivated
it survived three reviewers reading the guard, so a rule whose enforcement is a human noticing a line
has no enforcement.

#### Scenario: A member legitimately checks nothing

- **WHEN** a corpus member's coverage is zero
- **AND** that member is declared in the guard's expected-zero set
- **THEN** the run reports the zero as declared and does not fail on it

#### Scenario: A member checks nothing and was never declared

- **WHEN** a corpus member's coverage is zero
- **AND** that member is not in the guard's expected-zero set
- **THEN** the guard reports a violation naming that member and exits non-zero
- **AND** it does so on a tree that is otherwise clean

#### Scenario: A rule silently stops covering a member

- **WHEN** a change causes a rule's required set to become empty for a member that previously had
  coverage, and that member is not declared expected-zero
- **THEN** the guard fails at the moment the coverage disappears, rather than at the moment someone
  deliberately breaks that member
