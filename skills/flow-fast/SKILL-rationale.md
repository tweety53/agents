# flow-fast — rationale

Reasoning behind `skills/flow-fast/SKILL.md`: what pins its tables and the incidents its rules were
written against. Moved here verbatim from the run-loaded file; **no `/flow-fast` run loads this
file.** Each heading names the source file and the section the passage came from.

## SKILL.md — Stage keys

> This table is what `TestStageKeysMatchFlowFastSkillTable`
> (`<agents repo>/stats/internal/stages/names_test.go`) pins the `/flow-fast` vocabulary to, so a key
> added to a step above without a matching `stages.Table` row — or the reverse — is a test failure
> rather than a mark rejected mid-run.
