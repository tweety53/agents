# Review panel — kan-374-flow-run-planning-stages-on-fable-opus-fallback

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Code review (low) | Minor | commands/flow.md:16 | This diff updated the Claude-Code twin's sentence but not the Cursor twin's, leaving the two files describing the same command inconsistently: commands/flow.md:16 still says brainstorming runs (unchanged, fully interactive) where commands-claude/flow.md:11 says it now runs in a planner subagent on the configured planning model. |
| F2 | Principles | Important | skills/flow/brainstorm.md:143 | The second-mismatch branch treats the second begin call's -model opus as the record of what was attempted, so the persisted dispatch row keeps -model opus even though the handshake has just proven the planner did not run on opus — contradicting model-policy's every-dispatch-records-the-model-it-used rule and the design's own opus-fallback-verified decision; the same gap propagates to implement.md's fix-run dispatch. |
| F3 | Principles | Important | skills/flow-contracts/project-configuration.md:89 | The new planning-model paragraph repeats the default-landing-route matching rule near-verbatim, swapping only the noun, instead of citing it — the same validation rule written in two places. |
| F4 | Principles | Minor | skills/flow/implement.md:118 | Introduces a planner-fix-<n> key ordinal with no stated derivation — a reader cannot tell how <n> is computed from any existing convention. |
| F5 | Primary | Critical | skills/flow/brainstorm.md:137 | The opus-fallback re-dispatch reuses the same -key planner under the same session token, so the store treats the second begin as an idempotent replay of the first: the persisted row keeps model fable and the later end overwrites outcome fallback with completed, silently erasing the fallback and recording a false model; implement.md's fix-run dispatch inherits it by citation. |
| F6 | Bugbot | Minor | scripts/check-stage-mark-calls.sh:312 | The guard never scans skills/flow/brainstorm.md or implement.md, so a planted shell-substitution session token in the new planner-dispatch line was not caught; the restriction predates this branch but this diff's new call sites fall through it. |
| F7 | Bugbot | Minor | stats/internal/store/settings.go:57 | DefaultPlanningModel is declared but referenced by no Go code and asserted by no test; changing its value compiles clean and every test passes, so it is dead decoration that can drift from the skill's own literal. |
| F8 | Code review (low) | Minor | commands/flow.md:18 | The F1 fix merged two sentences onto one line without rewrapping, breaking the file's own hard-wrap convention that every other body line and the commands-claude twin follow. |
| F9 | Bugbot | Minor | scripts/lib/coverage.sh:185 | A declared-zero member whose scanned count becomes non-zero silently drops its declared reason and reports the real count with no violation and exit 0, so the guard never notices a declared reason has gone false; pre-existing mechanism, adjacent to this diff's corpus widening. |

findings-total: 9
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed

reproducers-total: 9
finding-reproducer: F1 none — documentation-consistency defect; compare commands/flow.md:16 with commands-claude/flow.md:11
finding-reproducer: F2 grep -n attempted skills/flow/brainstorm.md
finding-reproducer: F3 grep -n byte-for-byte skills/flow-contracts/project-configuration.md
finding-reproducer: F4 grep -n planner-fix skills/flow/implement.md
finding-reproducer: F5 none — prose defect; a same-key begin is an idempotent replay in the store, verified against live Postgres during review
finding-reproducer: F6 none — mutation survives; the guard's find matches only files named SKILL.md or pipeline.md, so it never scans the skills/flow phase files
finding-reproducer: F7 none — mutation survives; no Go test or caller references store.DefaultPlanningModel
finding-reproducer: F8 none — line-wrap defect; compare commands/flow.md:18 against its neighbouring lines
finding-reproducer: F9 none — mutation survives; coverage_report checks a declared-zero member only when its count is 0, never when it becomes non-zero
