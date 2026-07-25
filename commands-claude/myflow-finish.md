---
model: sonnet
description: Finish — verify the PR merged (Gate D), sync delta specs, archive the OpenSpec change
---

Use the **openspec-archive-superpowers** skill (`.claude/skills/openspec-archive-superpowers/SKILL.md`).

Follow that skill exactly. Requires stage `awaiting-pr-review`. Runs **after `/myflow-code-review`** (which committed, pushed, and opened the PR, then deliberately stopped) **and after the human has reviewed and merged the PR on the forge (Gate D)**: verify the PR actually merged → OpenSpec delta sync (offer, recommended) → archive the change. **Never commits, tests, merges, or pushes anything itself** — it only verifies the PR merged and archives.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted: run `openspec list --json`; if exactly one active change has been through code review, use it automatically; if multiple, ask which.

**If the PR is still open:** this is the normal, expected block — Gate D hasn't happened yet. Stop and tell the user to review and merge the PR themselves, then re-run; never offer to merge it for them, and never silently archive a change whose code never reached the base branch. Only proceed unmerged on an explicit user override (AskUserQuestion, default No).

**If nested `<name>-fix-N` sub-changes exist** (from `/myflow-do-fix`): archives them together with `<name>` in the same operation — never one without the other.
