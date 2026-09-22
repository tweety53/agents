# flow-plan — rationale

Reasoning behind `skills/flow-plan/SKILL.md`: why its lists are shaped as they are and the incidents
its rules were written against. Moved here verbatim from the run-loaded file; **no `/flow-plan` run
loads this file.** Each heading names the source file and the section the passage came from.

## SKILL.md — Capturing a new change

> Each basename above is written with its usage
> arguments so `check-guard-symlinks.sh` reads it as a citation; its rule 6 then fails any symlink
> here that no citation names, so dropping a basename from this list is a lint failure, not a moved
> coverage count (KAN-532).
