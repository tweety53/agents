# Agent baseline — rationale

The reasoning behind `rules/agent-baseline.md`, moved here verbatim. `setup.sh` does not install
this file, and no dispatched agent reads it.

## Preamble

It carries **one line per rule plus a pointer to that rule's own file** — never a copy of the rule, so there is nothing here to drift out of sync with the source.

## Propagate this

You cannot know from a dispatch prompt whether the agent will end up touching production, adding a dependency, or pushing a branch. An agent that inherits nothing and is told nothing has no rules at all.

## Project rules come on top

The flow pipeline is deliberately absent from the table above — it is command-triggered, and summarising a state machine is exactly the staleness its own rule forbids.
