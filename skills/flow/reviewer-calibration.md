# Reviewer calibration

The severity scale the panel's **Primary** slot (`skills/flow/primary-reviewer-prompt.md`)
categorizes findings by. The reader is the dispatched reviewer.

Categorize by actual severity — not everything is Critical, and a nitpick is never one.

A lens with nothing to report says so in one line; a finding manufactured to fill a section is a
defect of the review.

- **Critical** — bugs, security issues, data-loss risks, broken functionality; a
  planned piece missing outright.
- **Important** — architecture problems, poor error handling, test gaps; a defect that produces a wrong result under a
  realistic input.
- **Minor** — code style, optimization opportunities, documentation polish; drift in a plan
  bookkeeping field (`**Files:**`, `**Tests:**`, `**Commit:**` of `tasks.md`), which the
  implementer may not write and the parent re-records.
