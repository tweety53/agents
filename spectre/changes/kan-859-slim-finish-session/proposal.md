# Proposal — kan-859-slim-finish-session

**Jira:** KAN-859 (epic KAN-851). **Audit:** `docs/prompt-audit-2026-09-29/audit-finish.md`,
`## Candidates`. **Blocked by:** KAN-852 (guard) and KAN-853 (reconcile). Both are done.

## Why

- This operator's default finish session is merge-and-push, chaining run 1 into run 2. It loaded
  `integrate.md`, `archive.md`, `finish-contract-run1.md` and `finish-contract-run2.md` whole:
  106,969 B.
- Some of that text applies only in rare cases:
  - when a guard is missing;
  - when the unfinished-work gate fires;
  - when a worktree's base moved.

  It still loaded on every run.
- `integrate.md` and `archive.md` restated their contracts, and the contracts carried
  editor-facing rationale.
- `jira-followups.md` (35 KB) loads only after the operator picks "File or join a Jira follow-up".
  Even then, two thirds of it is the join path and nearly half is rationale.

## What changes

Only verbatim moves and declared duplicate cuts. The only rewordings are the lines listed in
`verbatim-moves.txt`.

1. **Hand fallbacks.** R1c, R1d, R1f, R1h, R1q, R1z, R2d, R2e and R2q, plus
   `project-configuration-isolation.md`'s load directive, move to
   `skills/flow-contracts/finish-hand-fallbacks.md`. A08 moves there too, and A01 is cut as a
   duplicate of R2e.
   - Run 1 and run 2 each gain a directive: load the file only when the guard presence check named
     one of their scripts missing, or a call finds the script absent.
   - No heading moves.
2. **Self-review `run` branch:** nothing to do. KAN-854 already made self-review always-deferred,
   so A14, A20, R2t and R2u no longer exist. The `ANGLE_CONTRACT` default already names
   `skills/flow-self-review/SKILL.md`.
3. **Rationale moved verbatim.**
   - To `finish-contract-rationale.md`: R1b, R1e, R1g, R1j, R1l, R1m, R1x, R1y, R1aa, R1ae, R2b,
     R2f, R2g, R2h, R2i, R2n, R2o, R2p, R2x, R2y, R2z, R2ac, R2ad, R2ae, R2af, R2ai, R2aj, R2ak,
     R2al, R2am, R2an and R2ao.
   - To `SKILL-rationale.md`: I06a+b and A05.
4. **Duplicates cut, one copy kept per pair.**
   - From `integrate.md`: I02a/b, I03, I06c, I07, I09, I11a/b, I12 and I15.
   - From `archive.md`: A02a/b, A04, A06, A07, A09, A10, A17a/b, A18 and A19b.
   - From run 1: R1i, R1k, R1r1, R1s and R1ab.
   - From run 2: R2j, R2k and R2l.
   - R2r and R2s: `/flow-fast` text that no `/flow-fast` run loads.
5. **The unfinished-work gate apparatus** moves to `skills/flow/unfinished-work-gate.md`: I16,
   I17, I04, the filing-ask paragraph, R1u1, R1u2, R1n, R1o and R1p. It loads only when a
   worktree reports `OUTSTANDING` or `VISUAL-VERIFY-MISSING`.
   - The explain-before-asking rule moves with the prompt, so it is in force at the gate (D14).
6. **The base-moved rebase block** moves to `skills/flow/sync-onto-base.md`: I05a's aside, I05b,
   and R1r2a–d. The **Sync the branch onto the base** heading moves with it.
   - Seven citers are repointed: `implement.md`, `integrate.md`, `review-panel-fix-round.md`,
     `flow-fast/SKILL.md`, three places in run 1, and `SKILL-rationale.md`.
7. **`jira-followups.md`.**
   - The join path moves to `skills/flow-contracts/jira-followups-join.md`, loaded only when the
     join search returned a candidate.
   - Rows J01–J35 (high and medium confidence) move to `jira-integration-rationale.md`.
   - Citers of the join are repointed.
8. **Support updates:**
   - `stage-keys.md` gains a row for each new `skills/flow/` sibling.
   - The contracts index gains rows for the three new contract files.
   - `load-sets.sh` counts the new conditional files.
   - Citations in `guard-verdict-verification.md`, `AGENTS.md` and `templates/AGENTS.md` follow
     the moves, and so do the location notes in `finish-contract-rationale.md`.

## Out of scope

- **Rows not done:** I01, I10, I13, I14, A03, A15, A21, R1t, R1w, R1ad, R1pr, R1man, R2c, R2w1–3,
  R2aa, R2ab, R2ag, R2ah and J03b/J26/J36. Design.md D5 gives the reasons.
- **MECHANICS rows** (MI1, MA1, MA2, MA3, MR1, MR2): they need code and parity tests.
- **Drift items** that need a decision rather than a trim (D2–D13, D15–D24): KAN-853 and KAN-854
  own them.
