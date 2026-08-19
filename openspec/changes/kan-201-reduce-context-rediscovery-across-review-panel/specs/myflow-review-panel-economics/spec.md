## ADDED Requirements

### Requirement: Every panel slot is dispatched with the shared context bundle

Each review-panel slot's dispatch prompt SHALL name the absolute path of
`<worktree>/.superpowers/sdd/dispatch-context.md` and SHALL state that the slot need not go looking
for the change's proposal, design, plan, delta specs or engineering principles.

The same prompt SHALL require the slot to read the review diff itself. The bundle carries planning
context and never the diff's content, so no slot may treat the bundle as evidence about the code
under review.

This requirement SHALL NOT change which slots a roster dispatches, that slots are dispatched
separately and never merged into one prompt, the model each slot runs on, the reproducer each finding
must carry, or the zero-open-findings handoff bar.

#### Scenario: Three slots, one bundle, three dispatches

- **WHEN** a `light` roster dispatches its three required slots
- **THEN** each is a separate dispatch naming the same bundle path, and no two slots are merged into
  one prompt

#### Scenario: A slot dispatched with no bundle available

- **WHEN** the gather script could not produce a bundle
- **THEN** the slot is dispatched with the prompt shape used before this capability existed, and the
  panel is neither reduced nor skipped

### Requirement: The fix dispatch carries each finding's located evidence

The dispatch that repairs panel findings SHALL carry, for every surviving finding, that finding's
identifier, the slot that raised it, its severity, its `file:line` taken verbatim from the findings
table's Location column, its theme, the text of its `finding-reproducer:` line, and any bounce
recorded against its defect identity. It SHALL also name the shared context bundle.

The dispatch prompt SHALL state that these locations were established by the slot that raised them
and are not to be re-derived. The evidence already exists on the other side of the dispatch boundary;
this requirement is what carries it across.

Source excerpts around a finding's `file:line` SHALL NOT be inlined into the dossier. The fix round
edits the code it is given, so an inlined excerpt would be invalidated by the fixer's own work, and a
fixer reasoning from a stale excerpt is a correctness risk that the location alone does not carry.

This requirement SHALL NOT change the reproducer verification that precedes dispatch, the
re-run-and-materiality condition that closes a finding, the bounce accounting, the operator handback,
or the model the fix subagent runs on.

#### Scenario: A surviving finding reaches the fixer located

- **WHEN** a finding at `scripts/foo.sh:42` survives its review pass
- **THEN** the fix dispatch names that location, its theme and its reproducer, rather than a bare
  restatement of the finding's prose

#### Scenario: The fixer reads the real file

- **WHEN** the fix subagent is dispatched
- **THEN** its prompt carries no source excerpt for any finding, and it opens the named file itself

#### Scenario: A bounced finding carries its bounce

- **WHEN** a finding has already been bounced once against its defect identity
- **THEN** the dossier states that bounce alongside the finding
