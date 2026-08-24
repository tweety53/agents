## ADDED Requirements

### Requirement: The build target produces the binaries it documents

The stats project's build target SHALL write every runnable binary the project documents to a fixed,
named path, so that running the documented build command refreshes what the documented run command
executes.

A build target that compiles a main package without naming an output path SHALL be treated as a
defect: with more than one main package the compiler discards the result, leaving whatever binary
was there before in place, and an operator who follows the documented build then runs stale code
with no indication that they are doing so.

The target SHALL additionally compile every package in the project, including main packages whose
binary it does not emit, so that widening the set of emitted binaries is not the only way to keep a
package compiling.

#### Scenario: The documented build refreshes the daemon

- **WHEN** an operator runs the documented build command after changing the daemon's source
- **THEN** the daemon binary at its documented path reflects that change
- **AND** starting it runs the changed code

#### Scenario: A main package with no emitted binary

- **WHEN** the project holds a main package the build target emits no binary for
- **AND** that package stops compiling
- **THEN** the build target fails

#### Scenario: The documented build and the working build are one command

- **WHEN** the project's documentation tells an operator how to build a binary before running it
- **THEN** it names the build target alone, and no separate explicit compile step beside it
