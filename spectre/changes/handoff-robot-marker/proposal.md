# handoff-robot-marker

## Why

Operator, 2026-10-02: the handoff marks an auto-picked decision with an emoji, not `⚠`.

## What changes

- The `⚠` marking an auto-picked or auto-resolved decision in a handoff `**Decisions:**` line becomes `🤖` in `skills/flow/verify-and-handoff.md`, `skills/flow/SKILL.md`, `skills/flow/document-fix.md`, `skills/flow/SKILL-rationale.md`, `skills/flow-contracts/handoff-blocks.md`, `skills/flow-contracts/operator-prompts.md`, `skills/flow-contracts/operator-prompts-auto-resolution.md`.
- Left as `⚠`: warnings and CLI output (`⚠ flow:`, `⚠ Jira: …`, setup.sh, `⚠ GUARDS MISSING`, `⚠ roster:`), the Jira partial-join marker, the brainstorm planner's printed lines, withdrawal's invalid-entry line.
