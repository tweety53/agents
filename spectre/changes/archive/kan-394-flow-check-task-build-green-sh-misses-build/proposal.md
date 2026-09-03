# kan-394-flow-check-task-build-green-sh-misses-build

## Why

KAN-29's self-review found `check-task-build-green.sh` failing eight `**Build:**` lines in its
`tasks.md` for the change's whole life — six `**Build:** green — <prose>`, one
`**Build:** green (unchanged; …)`, and one with no keyword at all. The tag grammar,
`^\*\*Build:\*\*\s+(green|red)\s*$` in `scripts/lib/plan_grammar.py`, accepts only the bare
keyword, and every miss is reported as `task <id> has no **Build:** tag` — the consequence, with the
cause (a tag line whose value the grammar does not read) hidden, the same shape fix F22 removed for
unclosed fences.

## What changes

- The keyword is read as a word-boundary prefix of the value after `**Build:**`; trailing prose on
  the line is ignored. `**Build:** green — desktopTest unchanged` is green, `**Build:** red — see below` is red.
- The first column-0 `**Build:**` line in a task's body is the tag. A first line whose value does
  not open with `green` or `red` is a new violation naming the line and its value:
  `task <id> has a **Build:** line reading "<value>", which is neither green nor red`. "No tag"
  stays for a body with no `**Build:**` line at all.
- One grammar change in `scripts/lib/plan_grammar.py` reaches both readers,
  `check-task-build-green.py` and `check-task-commit-fields.py`; the latter changes no logic.
- `skills/flow-contracts/build-green.md` states the prefix grammar and the malformed-tag class in
  place of "treated the same as no tag"; test case 7 and six new build-green cases plus one
  commit-fields case pin it.

Design, decisions and rejected alternatives: `design.md` beside this file.
