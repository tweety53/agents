---
name: myflow-contracts
description: The myflow pipeline's contract definitions — state file shape, state self-heal, project configuration, and Jira integration. Load the one file you need; each is canonical for its own contract. Referenced by the stubs in rules/myflow-manual-review.mdc.
allowed-tools: Bash(jq:*), Bash(git:*)
license: MIT
metadata:
  author: gymie
  version: "1.0"
---

# myflow contracts

myflow's contract definitions, split out of `rules/myflow-manual-review.mdc` so the always-on rule
layer carries only the judgment an agent needs without being asked to load anything.

**Load the one file you need — not this whole directory.**

## Index

| File | Load it when you need to |
|------|--------------------------|
| [state-file.md](state-file.md) | Read or write a change's state file: its path, its full shape, monotonic gates, carry-forward |
| [state-self-heal.md](state-self-heal.md) | Validate a state file against on-disk artifacts, or handle a missing/contradicted one |
| [project-configuration.md](project-configuration.md) | Resolve `.myflow/project.md` — apps, run, test, lint, standards, jira — including standards-entry resolution and containment |
| [jira-integration.md](jira-integration.md) | Resolve a linked issue, transition it, or sync its description |

Each file is **canonical** for its own contract. Where a skill and one of these files disagree, the
file wins.

The headings inside these files keep the names they had in the rule, so an existing reference to a
section by name still resolves — now to the file that holds it.
