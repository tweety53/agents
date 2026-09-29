# Proposal — kan-854-flow-fix-audit-behaviour-decisions

**Jira:** KAN-854 (epic KAN-851). **Plan:** `docs/prompt-audit-2026-09-29/README.md`.

## Why

The 2026-09-29 prompt audit found 25 defects in the `/flow` prompts that are not wording drift:
following the text as written makes a run fail a guard, skip a step, load files it does not need,
or leave an artifact behind. Each fix changes what a run does, so each needs a decision rather
than a copy-edit. KAN-853 settled the contradicting copies; this change settles these.

## What changes

Each item is either fixed, or recorded as a decision with its rejected alternative in
`design.md`. Grouped as the issue groups them:

- **Pipeline and dispatch (1–15):** the panel-fix retry key, the fix run's re-Decide, project
  hazards and the paragraph binding on an inline run, the gated reviewer's pair on
  `micro`/`small`/`regular`, the zcode handshake, the conditional SDD skill, `## worktree setup`
  for peer and app worktrees, the guard-presence `lib/` sibling, the rules that lived only in
  rationale files, the fix subagent's `fix-mutation:` shape, the placeholder fill table, reviewer
  slots re-reading bundled files, README's subagent `CLAUDE.md` claim, and `/flow-fast`'s re-read.
- **Finish (16–20):** `SELF_REVIEW_MODEL` dropped, check 4's single ask, step 6's legacy file,
  the archive-scope fallback root, and the explain-before-asking rule at the gate that offers
  filing.
- **Other (21–25):** `<repos>`, `flow jira transition`, wave-group copies, port collisions, and
  `cost-status`.

Code changes, each with tests and a mutation check (the new test must fail against the old code):

- `check-dispatch-paragraphs` pins the `fix-mutation:` shape (case 91).
- `check-task-reviewer-single-dispatch` covers `micro` (a new case).
- `check-guard-symlinks` reads the `$(dirname …)` spelling of a sibling
  (`TestEveryFlowGuardShimRequiresLib`, `TestSelfDirSiblingSpellings`).
- `check-cleanup-complete` reports a surviving wave-group copy (cases 2d, 2e).
- `check-model-resolution-shell.sh` checks only `DEFAULT_MODEL`/`MODEL_SOURCE`, and refuses a
  re-added `SELF_REVIEW_MODEL` (its harness has new cases).

Every run-loaded sentence reworded on purpose is listed in `verbatim-moves.txt`.

## Out of scope

- Retiring the store's `selfReviewModel` field, its flag, API and `## self review model` key end to
  end (a follow-up; see design.md D16).
- The placeholder render script (the separate mechanics issue).
- Removing `agent-baseline.md`'s rule table for Claude Code, pending operator verification (D14).
- Any trim (KAN-855..859).
