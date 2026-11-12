# Reviewer calibration

The severity scale the panel's **Primary** slot (`skills/flow/primary-reviewer-prompt.md`) and the
gated per-task reviewer (**4. Execute (SDD + TDD)**, `skills/flow/implement.md`) categorize findings
by. The reader is the dispatched reviewer.

Categorize by actual severity — not everything is Critical, and a nitpick is never one.

A lens with nothing to report says so in one line; a finding manufactured to fill a section is a
defect of the review.

- **Critical** — bugs, security issues, data-loss risks, broken functionality; a
  planned piece missing outright.
- **Important** — architecture problems, poor error handling, test gaps; a task field
  that no longer reflects the diff; a defect that produces a wrong result under a
  realistic input.
- **Minor** — code style, optimization opportunities, documentation polish.
