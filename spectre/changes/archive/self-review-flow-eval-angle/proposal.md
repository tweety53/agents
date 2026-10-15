# self-review-flow-eval-angle

## Why

Operator, 2026-10-15: a seventh `/flow-self-review` angle that lists the run's work that would
make good flow-eval cases — implementer, reviewer and visual — and offers to mine them. flow-eval
measures each flow role on real cases mined from flow history (`flow-eval case add
impl|review|visual`), and today nothing in a run surfaces which of its tasks, review findings or
visual pairs would make one: the cases are found by hand, after the fact, if at all.

The operator fixed the shape:

- Each candidate is presented in its kind's template, every field mandatory — impl: the task as
  recorded, the resulting code as an excerpt under its worktree-resolved `path:line`, why it is a
  good candidate; review: the finding(s) with severity, location and the reviewer's sentence, the
  bad code, the fixed code with its fixing commit, why; visual: the actual app screenshot and the
  mockup screenshot by absolute path, the findings, why.
- Mining happens only on the operator's explicit approval, per case — never automatically.
- The angle is as explicit as the other six: an empty flow-eval angle says so. Paths are
  absolute; a screenshot that exists only in git history is extracted to a durable path first.
- Every "six angles" site, the report guard's all-angles check, the filing/rating prompt shape and
  their tests move to seven.

## What changes

- `skills/flow-self-review/SKILL.md`: angle 7, `flow-eval`, in the canonical table, with its remit,
  the per-kind template, absolute and durable paths, and its carve-outs (product code is in scope
  for it; it may read the commits and files the bundle names). Step 4's one `AskUserQuestion`
  call gains mining questions beside the filing ones — never mixed in one question, default
  **None — mine nothing**. Step 5 mines each chosen candidate with `case add` in a `flow-eval`
  worktree once the change's commits are final on the default branch, one commit per case, never
  merged or pushed. A candidate is recorded `noted`; the report carries it as a `noted` line with
  its template beneath. Every "six" becomes "seven".
- `scripts/check-self-review-report.sh` (Go, `stats/internal/guard/selfreviewreport.go`): a second
  frozen list, `docs/self-review/six-angle-reports.txt`, holds every report written before angle 7
  to the first six angles; every other report must carry all seven.
- `flow self-review bundle` (`stats/internal/selfreview/git.go`) serves the archived
  `visual-verification.md`, the record a visual candidate is read from.
- `skills/flow-contracts/finish-contract-run1.md`, `skills/flow-fast/SKILL.md`,
  `commands-claude/flow-self-review.md`: seven angles; run 1 names the mining after a merge-and-push
  landing; two citations of "step 5" for the landing push's shas name item 5 of step 3, where that
  rule lives.
- `skills/flow-contracts/operator-prompts-auto-resolution.md`: the mining is an outward-facing act
  that auto-resolution never takes.
