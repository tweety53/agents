# Proposal — kan-858-slim-verify-contracts-and-reviewer-prompts

**Jira:** KAN-858 (epic KAN-851). **Audits:** `docs/prompt-audit-2026-09-29/audit-verify-and-contracts.md`
and `audit-subagent-prompts.md`, `## Candidates`. **Blocked by:** KAN-852 (guard) and KAN-853
(reconcile). Both are done.

## Why

- The implementation session loads `verify-and-handoff.md` and the contracts it cites on every run.
  It also loaded `artifacts-registry.md`, which no implementation step acts on.
- On a UI run the parent loaded all of `visual-verify.md` (61.6 KB). About 40 KB of it is text only
  the dispatched verifier acts on.
- Every reviewer slot received template text addressed to the dispatcher. Every dispatched subagent
  read `agent-baseline.md`'s rationale.

## What changes

This change makes verbatim moves and declared-duplicate cuts only. The only rewording is in the
lines listed in `verbatim-moves.txt`.

1. **VV-05:** step 4, steps 7–11 and the `## Report` template move to
   `skills/flow/visual-verify-verifier.md`. The verifier reads it at the absolute path its prompt
   carries. The parent keeps steps 3, 5, 6, 12 and 13, **Blocking**, and the two reconciliation
   clauses. MODEL HANDSHAKE, TOOLS and NO DELEGATION stay in the prompt.
2. **VV-04:** the tooling analyst's dispatch, prompt, recording, re-run and abort move to
   `skills/flow/visual-verify-tooling-analysis.md`. It loads only on a fix run with a miss.
   `dpSites` now pins each of the two files at min 1 (it was visual-verify.md at min 2). The tests
   and the shell header follow.
3. **WI-06:** `verify-and-handoff.md` loads `workspace-isolation.md` by section:
   - **The cache index** for a `cache index` row;
   - **The empty id** for exit 1;
   - nothing for exit 2.

   The file is not split physically.
4. **AR-01:** implement.md no longer loads `artifacts-registry.md`. The finish session keeps it.
5. **Duplicates cut in `verify-and-handoff.md`:** VH-03, VH-05, VH-06, VH-09, VH-12, VH-15, VH-18
   and VH-20.
6. **GB-01 + VH-11:** `git-boundaries.md`'s two-commit chain moves to
   `skills/flow-contracts/git-boundaries-commit-chain.md`. `integrate.md` loads it.
   `verify-and-handoff.md` loads it only when a `prUrl` is recorded.
7. **Rationale moved verbatim:**
   - from the verifier's steps: VV-06, VV-08, VV-10, VV-11, VV-12, VV-15 and VV-16;
   - VV-17 from `visual-verify.md`;
   - from `workspace-isolation.md`: WI-02, WI-03, WI-04, WI-07, WI-08, WI-09 and WI-11.
8. **Reviewer templates:**
   - **Primary, failure-modes and principles:** the intro, the Agent-call header and the
     "Read-only review." line are cut.
   - **Principles:** also loses `## Your Scope` (R4), `[PRINCIPLES_PATH]` (R6) and
     `[STANDARDS_PATHS]` (R7). R6's install-path worked example moves verbatim into
     `review-panel.md` **Principles**. R1c's reasoning moves to `SKILL-rationale.md`.
9. **D11:** `## Calibration` moves to `skills/flow/reviewer-calibration.md`.
   - The per-task reviewer's REPORT FILE cites that file.
   - The primary prompt reads it through a new `[CALIBRATION_PATH]` placeholder, filled from
     `review-panel.md`'s placeholder table.
10. **`agent-baseline.md`:** A1, A2 and A4 move to a new `rules/agent-baseline-rationale.md`, which
    `setup.sh` does not install. A3 is cut.
11. **`flow-fast/SKILL.md`:**
    - FF3–FF6 and FF9b are cut;
    - FF7's reason clause moves to its `SKILL-rationale.md`.
12. **Guard and support updates:**
    - `crExpectedZero` gains the calibration file and the baseline rationale;
    - `load-sets.sh`, `stage-keys.md`, Read discipline's phase-file list and the contracts index
      follow the moves;
    - comments in `record.go` and `resolve-visual-screenshots.sh` are repointed.

## Out of scope

- VV-01…VV-03, VV-07, VV-09, VV-13, VV-14, VH-01, VH-02, VH-04, VH-07, VH-08, VH-13, VH-14,
  VH-16, VH-17, VH-19, VH-21–VH-26, WI-01, WI-05, WI-10, WI-12, GB-02–GB-04, AR-02…AR-08, SR-01
  and WR-01. KAN-858's scope does not name them, or the audit scores them low, or design.md D6
  explains why they are not done.
- **MECHANICS rows:** P4, R5, F4, FF1, VV-18, VH-24–VH-26 and WI-12. They need code and parity
  tests.
- Bugbot and security rows. Those slots were retired on main.
