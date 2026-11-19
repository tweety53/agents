---
name: flow-visual-tooling-analysis
description: Use when a visual verification missed a defect and you want the sweeps that would have caught its class. Runs inline in your session on a change whose verify-report-*.md exist, writes sweeps-<n>.md. Standalone, not a pipeline stage.
---

# Visual verification — tooling analysis

Runs inline in the operator's session; it dispatches nothing and is no pipeline stage.

**Input.** The verifier's own dispatch (**The verifier dispatch**, `skills/flow/visual-verify.md`) produced the reports. The missed defects, as the operator states them; the change's
`<abs-worktree>/.superpowers/sdd/verify-report-*.md`; `<changeRoot>/visual-verification.md`; the
captures they cite; and `skills/flow/visual-verify-verifier.md`.

**Task.** For each missed defect, name the sweep in steps 8–10 of
`skills/flow/visual-verify-verifier.md` that should have caught it and why it did not, then name the
defect's class — the property, element kind or state no sweep read. Write the sweeps that close each
class to `<changeRoot>/sweeps-<n>.md` (`<n>` the next unused ordinal): a new sweep, or a sharpened copy
of an existing one under its own name, each in step 10's sweep shape — what it reads from the capture,
and from the frame where one is composed, and what makes it fail. A sweep targets the class, never the
instance: it must be able to find another defect of the class in a view the fix instructions do not
name. Fix nothing, edit no source and run no capture.

**Finish.** Name the sweeps file by absolute path and list each sweep with the miss it closes, so the
operator can decide whether its sweeps join step 10 for every later change. This skill never edits
`visual-verify-verifier.md`.
