## ADDED Requirements

### Requirement: A project is named on screen, not keyed

Wherever the interface names a project for a human to read, it SHALL show the project's **display
name** — the project key with its derivation suffix removed — and SHALL NOT show the raw key.

A project key is `<basename of the main checkout>-<first 8 hex of sha1 of that checkout's absolute
path>`. The suffix exists so two same-named repositories in different directories address different
records; it identifies, and identity is not a label.

The display name SHALL be derived by removing a trailing `-` followed by exactly eight lowercase
hexadecimal characters. A key not matching that shape SHALL be shown **unchanged** — a key derived
some other way stays readable rather than losing its last segment to a pattern that was never about
it.

The full key SHALL remain reachable from every surface that shows a display name: as the row
identity, as the link target where the surface links, and as a tooltip on the named element. Where
two projects share a display name, what distinguishes them is therefore still available without
being the thing every reader has to parse.

#### Scenario: A project is named in a table

- **WHEN** a view renders a column naming the project a row came from
- **THEN** the cell reads the display name, and the full key is available from that cell

#### Scenario: A key that does not carry the derivation suffix

- **WHEN** a project key does not end in `-` plus eight hexadecimal characters
- **THEN** it is shown exactly as it is, with nothing removed

#### Scenario: Two projects share a display name

- **WHEN** two project keys differ only in their suffix
- **THEN** both are named identically on screen, and each one's full key is still reachable from its
  own surface

### Requirement: The project parameter accepts a display name

Every endpoint accepting a `project` parameter SHALL accept either the full project key or a display
name, so that a value read off the screen is a value the interface accepts.

Resolution SHALL happen on the server, where the set of projects is known. The interface SHALL NOT
reconstruct the mapping from whichever rows it happens to be showing — a project with no rows in the
selected period is absent from the interface and present in the store, so a client-side mapping
would be wrong exactly when a filter is most needed.

A display name matching **exactly one** project SHALL resolve to that project's key. A display name
matching **more than one** SHALL be rejected with a client error that names the ambiguity and lists
the candidate keys; it SHALL NOT be resolved to any one of them, because a filter whose value means
two projects is not a filter, and choosing silently answers a question the caller did not ask. A
value matching **none** SHALL be passed through unchanged, yielding no rows exactly as an unknown key
does today.

The view queries themselves SHALL continue to filter on an exact project key. Resolution sits above
them, so the rule is stated once rather than copied into every view's query.

#### Scenario: A display name is filtered on

- **WHEN** a view is requested with a `project` value that is a display name matching one project
- **THEN** the result is that project's rows, exactly as if its full key had been sent

#### Scenario: A display name matching two projects

- **WHEN** a view is requested with a `project` value matching more than one project
- **THEN** the request is rejected, and the error names the ambiguity and the candidate keys

#### Scenario: A full key is filtered on

- **WHEN** a view is requested with a `project` value that is an exact project key
- **THEN** it is used as-is, with no resolution attempted
