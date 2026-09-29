# Proposal — kan-855-slim-every-session-load

**Jira:** KAN-855 (epic KAN-851). **Audit:** `docs/prompt-audit-2026-09-29/audit-every-session.md`,
`## Candidates`. **Blocked by:** KAN-852 (guard), KAN-853 (reconcile) — both done.

## Why

With `/clear` between phases, planning, implementation and finish each load the router
(`commands-claude/flow.md`, `skills/flow/SKILL.md`, `pipeline.md`) fresh. `pipeline.md`'s
**State file** and **Project configuration** directives put `state-file.md` (21 KB) and
`project-configuration.md` (52 KB) in every session too. Most of those bytes serve one session,
one condition, or no run at all.

## What changes

Verbatim moves and declared-duplicate cuts only; nothing reworded but the lines in
`verbatim-moves.txt`.

1. `project-configuration.md` → a 11.2 KB core. Standards and `<agents repo>` → `-standards.md`
   (principles dispatch); workspace-isolation authoring/validator spec → `-isolation.md` (only when
   `prepare-workspace.sh` is missing); visual verification → `-visual.md` (with
   `visual-verify.md`); the `create`/`remove`/`survivors` command table and token rule → run 2
   step 5; `## review panel citation check` cut (duplicate of its key row and `review-panel.md`).
2. `state-file.md` → 14.4 KB. Fallback paths, key derivation, `mainCheckoutPath`, write ordering,
   journal replay and the worked example → `state-file-internals.md` (no run loads it); `flow state
   dir` and the withdrawal daemon note cut (duplicates); two rationale passages →
   `state-file-rationale.md`; `journal_test.go` repointed.
3. `pipeline.md`: Hand-verifying → `guard-verdict-verification.md` behind a directive; Stage exit
   → `brainstorm-planner.md`; per-task granularity → `implement.md`; the two `/flow-status`
   lines → `flow-status/SKILL.md`; guard-prose authoring rule → `skills/README.md`; the re-entrant
   bullets and the Finish contract pointer cut; seven rationale clauses → `pipeline-rationale.md`.
4. `SKILL.md`: Stage keys → `skills/flow/stage-keys.md`; the `VERIFY_MODEL` paragraph →
   `visual-verify.md`; K1, K7, K10, K12–K17 cut; K14's re-check sentence → review-panel's
   **Bundled dispatch**.
5. `commands-claude/flow.md`: C1–C5 cut, leaving the accepted states and the input rule.
6. `operator-prompts.md`: auto-resolution → `operator-prompts-auto-resolution.md` behind a
   directive; multi-select → `flow-self-review/SKILL.md` (its one call site). `model-policy.md`:
   three duplicate paragraphs cut. `CLAUDE.md` and `templates/CLAUDE.md`: rationale and two
   duplicates cut.
7. `check-verbatim-moves`: a leading `\` in `verbatim-moves.txt` escapes a heading, which `#`
   otherwise comments out (with tests).

## Out of scope

The planning, implement, verify and finish trims (KAN-856..859); mechanics to code (KAN-860);
the always-on `flow-manual-review.mdc` core and the "See … for why" pointers (kept — design.md D1,
D2).
