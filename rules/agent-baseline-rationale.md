# Agent baseline — rationale

The reasoning behind `rules/agent-baseline.md`, moved here verbatim. `setup.sh` does not install
this file, and no dispatched agent reads it.

## Preamble

It carries **one line per rule plus a pointer to that rule's own file** — never a copy of the rule, so there is nothing here to drift out of sync with the source.

## Propagate this

You cannot know from a dispatch prompt whether the agent will end up touching production, adding a dependency, or pushing a branch. An agent that inherits nothing and is told nothing has no rules at all.

## Project rules come on top

The flow pipeline is deliberately absent from the table above — it is command-triggered, and summarising a state machine is exactly the staleness its own rule forbids.

## Reporting back

An unchecked "possible bug" costs the operator a round trip to learn what the agent could have
learned in the minutes it already had the code open. Checking turns a guess into a verdict; fixing closes it. Outside the task the fix is held to cheap and
fast, so closing a side defect never costs more than the task itself. A small open choice is named cheap.
