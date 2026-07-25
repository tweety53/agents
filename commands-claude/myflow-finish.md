---
model: sonnet
description: Finish — verify the branch merged into main/develop, sync delta specs, archive the OpenSpec change
---

Use the **openspec-archive-superpowers** skill (`.claude/skills/openspec-archive-superpowers/SKILL.md`).

Follow that skill exactly. Runs **after `/myflow-code-review`** (which already committed and ran Basic Workflow #7): validate the branch actually landed on `main`/`develop` → OpenSpec delta sync (offer, recommended) → archive the change. **Never commits, tests, or runs #7 itself** — it only verifies that already happened.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted: run `openspec list --json`; if exactly one active change has been through code review, use it automatically; if multiple, ask which.

**If not yet merged:** stop or get explicit confirmation before archiving — never silently archive a change whose code never reached the base branch.

**If nested `<name>-fix-N` sub-changes exist** (from `/myflow-do-fix`): archives them together with `<name>` in the same operation — never one without the other.
