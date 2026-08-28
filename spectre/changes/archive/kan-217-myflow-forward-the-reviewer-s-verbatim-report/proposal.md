# kan-217-myflow-forward-the-reviewer-s-verbatim-report

Forward the reviewer's own report to the panel's fix agent, and keep the dispatcher's framing as
direction rather than fact.

**Jira:** KAN-217 — self-review finding 2.2 from KAN-189.

## Why

A panel slot reports to the dispatcher, the dispatcher distills that report into a one-line
`flow record finding -note`, and the rest is discarded. Nothing in `skills/` persists a reviewer's
own words. At fix time, **Carry each surviving finding to the fix subagent as a structured block**
(`skills/flow/review-panel.md`) hands the fix subagent exactly that distillation — `F<n>`, slot,
severity, `file:line`, theme, reproducer, bounces — and forbids inlining source alongside it.

Two costs follow, both measured on KAN-189:

- The fix agent re-derives context the reviewer already established, because the block carries the
  verdict without the evidence behind it.
- The dispatcher's summary asserts facts the dispatcher never checked. On KAN-189 the summary
  itself introduced two findings, F20 and F23, and the implementer built and documented them as
  specified. Measured context for the visible part of that change: ~795k subagent tokens and
  ~59 min agent time, ~475k of it in the two fix rounds.

## What changes

- Each dispatched panel slot's returned report is written verbatim to
  `<abs-worktree>/.superpowers/sdd/panel-report-<round>-<id>.md`, one file per slot per round.
- The fix-subagent dispatch prompt carries a new **VERBATIM REPORT — THE FACT** paragraph naming
  those files, and the per-finding structured block gains its slot's report path as a field. The
  block is stated to be direction; the report is the fact, and wins where the two disagree.
- `flow record finding -note` carries the reviewer's own sentence naming the defect, not a
  dispatcher restatement — closing the same hole at the point the paraphrase is created.
- `scripts/check-reproduce-not-read.sh` is generalized from one hardcoded paragraph to a table of
  required dispatch paragraphs and renamed `scripts/check-dispatch-paragraphs.sh`, so the new
  paragraph is guarded by the mechanism that already guards REPRODUCE, DON'T READ.
- `skills/flow-contracts/artifacts-registry.md` gains the row the new files require.

Nothing changes in `stats/` — no CLI flag, no schema, no migration. The report files are
`.superpowers/`-local, gitignored, and removed with the worktree at finish run 2.
