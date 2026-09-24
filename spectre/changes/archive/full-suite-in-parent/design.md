# The parent runs the full suite — design

**Change:** `full-suite-in-parent`
**Jira:** none — run unlinked, by the operator's choice
**Date:** 2026-09-24

## Context

The problem is `proposal.md`'s `## Why`; this file fixes the mechanism and its decisions.

## Mechanism

- **The FULL SUITE paragraph leaves the implementer dispatch.** No implementer runs the resolved
  `## test` list any more, in any wave shape.
- **The parent runs it, once, after the last group's guard passes** — on the canonical worktree, in
  the foreground, its output through `tail`, in the order the context bundle carries it. This is
  the rule that already governed the shared-wave case; it now governs the singleton case too, so
  the shared-wave carve-out is folded into one rule.
- **A failure in a file the plan-last group's tasks declare in `**Files:**` is fixed by the parent
  itself** — the same way the inline mode's parent applies panel fixes: edit, re-run the failing
  test, commit on the route the branch's push state dictates (one new commit on top of a pushed
  branch), then re-run the resolved `## test` list. Recorded as one `dispatches` row
  `-role panel-fix -agent-id inline` under the key `full-suite-fix-<n>`, like inline fix rounds.
- **Any other failure** is the existing `## Question` stop — the failing command and its output
  verbatim — before `final-review.diff` is written. Unchanged from today.
- **No dispatch is ever resumed.** Every Agent-tool dispatch is one-shot; the parent never sends a
  finished child a `SendMessage`. The four sites that re-woke a finished implementer — a wave pick
  conflict or guard failure on a pick, a `check-task-commit-fields.sh` exit 1, a
  `check-task-commit-planning-paths.sh` exit 1, and a gated reviewer's `fix` verdict — become the
  parent's own inline fix, with the commit mechanics and records the inline mode already uses
  (`-agent-id inline` under the task's fix key). The re-review after a gated fix is still a fresh
  reviewer dispatch.
- **The last-boundary paragraph** reads the parent's own suite verdict instead of the last
  implementer report's `## Full suite` heading.
- **TARGETED TESTS stays verbatim.** "The full `## test` list runs once per worktree at the last
  bundle" is still true, and the paragraph is reproduced byte for byte in `implement.md`,
  `review-panel.md` and `scripts/test-check-dispatch-paragraphs.sh`'s fixtures.
  <!-- verified: git grep 'the last bundle, and again in' @ f7a825b — implement.md:654, review-panel.md:1131, test-check-dispatch-paragraphs.sh:315-338 -->

## Decisions

### Where the full suite runs

**ID:** full-suite-runs-in-parent
**Status:** active
**Chosen:** the parent, in every wave shape — the main session writes its prompt cache with a
1-hour TTL, subagents with 5 minutes, and the suite outlasts 5 minutes.
<!-- measured: usage.cache_creation in ~/.claude/projects/-Users-tweety53-Projects-agents transcripts, and flow suite list @ 2026-09-24 — see the comments under this entry -->
**Considered:**
- Keep it in the implementer, with a mid-run keep-alive poll — rejected: needs a background run
  plus a polling loop inside the subagent, more machinery for the same result, and depends on the
  subagent being allowed a wait command.
- Set `subagentPromptCacheTtl: 1h` harness-wide — rejected: every subagent write then costs 2×
  instead of 1.25×; over the last 40 subagents that is ~4% more input cost than the misses it
  saves (19.8M vs 19.0M base-input-token units).

<!-- measured: ~/.claude/projects/-Users-tweety53-Projects-agents transcripts @ 2026-09-24 — subagent usage writes only ephemeral_5m_input_tokens (0 × 1h across 3 sampled subagents), main session only ephemeral_1h_input_tokens; one implementer re-wrote 204911 tokens after a 765s gap and 351188 after a 307s gap -->
<!-- measured: flow suite list @ 2026-09-24 — guard-tests 2m04s–5m58s on Yuriys-MacBook-Pro -->
<!-- verified: code.claude.com/docs/en/prompt-caching.md @ 2026-09-24 — "Subagents fall outside the main-conversation TTL bucket, so they get five minutes even on a subscription" -->

### Who fixes a red suite in the last group's own files

**ID:** parent-fixes-own-file-failure
**Status:** active
**Chosen:** the parent fixes it inline — the operator's answer; no subagent resumes, so no cache
rewrite is paid even on failure.
**Considered:**
- Resume the last group's implementer via `SendMessage` — rejected by the operator; the resumed
  implementer's cache has expired by then and it pays one full rewrite.
- `## Question` stop for every failure — rejected: a red suite in the task's own files would no
  longer self-heal, a regression from today.

### No subagent is ever resumed

**ID:** no-subagent-resume
**Status:** active
**Chosen:** every dispatch one-shot; whatever a finished child's work still needs is the parent's
own inline fix — the operator's requirement, added mid-implementation: a resumed child's cache has
expired by the time it is woken, so each resume re-writes its whole context.
**Considered:**
- A fresh one-shot fixer subagent per fix, like panel-fix — rejected by the operator: each fixer
  pays a cold cache write, and the parent already holds the context.

## Open questions

