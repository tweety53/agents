## ADDED Requirements

### Requirement: A skill directory SHALL carry symlinks only under its `scripts/` directory

A skill directory SHALL NOT contain a symlink directly under `skills/<skill>/`. Every symlink a
skill carries SHALL sit under `skills/<skill>/scripts/`, where **A guard a command invokes SHALL be
reachable from that command's installed skill** already governs it.

This closes the one placement that requirement does not reach. A file a skill's text names but does
not carry — most concretely `engineering-principles.md` and the reviewer-prompt files, which
`skills/myflow-do/SKILL.md` resolves **beside itself**, in `skills/myflow-do/`, including when
`/myflow-fast` is the command running that section — is resolved by reading where the text says it
lives, never by symlinking a copy into the command's own skill directory. That symlink makes a wrong
reading of the resolution rule work, which is what keeps the wrong reading alive.

The repository's own lint SHALL report such a symlink by path and target, and fail. Prose SHALL NOT
be the only countermeasure: `skills/myflow-do/SKILL.md` already states in terms that symlinking a
file in is the wrong fix, and a session created three such symlinks in
`skills/myflow-fast/` roughly five hours after that sentence was committed.

#### Scenario: A symlink appears directly under a skill directory

- **WHEN** `skills/<skill>/<name>` is a symlink, at the skill directory's top level rather than
  under its `scripts/` directory
- **THEN** the repository's own lint reports its path and its target, and fails

#### Scenario: A skill's scripts directory is unaffected

- **WHEN** `skills/<skill>/scripts/<name>` is a symlink into the repository's root `scripts/`
- **THEN** this requirement reports nothing, because that placement is the one the guard-reachability
  requirement above requires

#### Scenario: A skill carrying no symlink at its top level passes

- **WHEN** every entry directly under `skills/<skill>/` is a regular file or a directory
- **THEN** this requirement reports nothing
