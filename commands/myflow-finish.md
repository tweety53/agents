---
name: /myflow-finish
id: myflow-finish
category: myflow
description: Finish — verify the PR merged (Gate D), sync delta specs, archive the OpenSpec change
---

**Model:** Sonnet (or your default) is fine here — Opus is reserved for `/myflow-start`'s brainstorming stage. Cursor doesn't yet support a per-command model frontmatter field, so this is a recommendation, not an enforced switch.

Use the **openspec-archive-superpowers** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

Follow that skill exactly. Requires stage `review-done` — the stage `/myflow-review-done` writes once the human has reviewed and merged the PR (or that `/myflow-review automerge` writes directly). On mismatch, stop with the standard mismatch handoff. Runs **after `/myflow-review`** (which committed, pushed, and opened the PR, then deliberately stopped) **and after the human has reviewed and merged the PR on the forge (Gate D)**: verify the PR actually merged → OpenSpec delta sync (offer, recommended) → archive the change. **Never commits, tests, merges, or pushes anything itself** — it only verifies the PR merged and archives.

Also follow the myflow manual-review rule (`myflow-manual-review.mdc`) — installed globally, so let your harness resolve it rather than assuming a project-local path.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted: run `openspec list --json`; if exactly one active change is at stage `review-done`, use it automatically; if multiple, ask which.

**If the change is still at `awaiting-pr-review`:** the PR review has not been confirmed yet — run `/myflow-review-done <name>` first (a pure state write; it is what moves the change to `review-done`). This command does not accept `awaiting-pr-review`.

**If the PR is still open:** this is the normal, expected block — Gate D hasn't happened yet. Stop and tell the user to review and merge the PR themselves, then re-run; never offer to merge it for them, and never silently archive a change whose code never reached the base branch. Only proceed unmerged on an explicit user override (AskUserQuestion, default No).

**If nested `<name>-fix-N` sub-changes exist** (from `/myflow-do-fix`): archives them together with `<name>` in the same operation — never one without the other.
