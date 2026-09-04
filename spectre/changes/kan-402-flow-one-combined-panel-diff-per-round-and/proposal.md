# kan-402-flow-one-combined-panel-diff-per-round-and

**Jira:** KAN-402

## Why

KAN-402 carries two findings from KAN-29's self-review, cost angle. Both are real; the second is
mis-diagnosed, and this proposal records the corrected diagnosis.

**Finding 1 — every panel slot ran twice.** KAN-29 spans two repositories, so its change had two
worktrees. The store record for `gymie-7c1f238a/kan-29-final-verification-public-profiles-friends`
shows dispatches 72–79 as `panel-3-<slot>-backend` (diff base `22cae21`) and
`panel-3-<slot>-frontend` (diff base `7295fb5`) for each of Primary, Principles, Code review (low)
and Bugbot. That is not a bug in the run: `skills/flow/review-panel.md` says *"dispatch
**separate** review subagents — one per included slot, in **every** affected worktree."* The
doubling is the rule, and it doubles again with every further repository a change touches. Each
slot also had to be told the *other* worktree's diff does not exist, so cross-repository seams —
a backend contract and the frontend that consumes it — were never in one reviewer's view.

**Finding 2 — cache reads of 30–90M tokens per implementer dispatch.** The 87.7M read is dispatch
65 (task 29, 35 minutes on `sonnet`).
<!-- measured: curl http://127.0.0.1:4173/api/v1/records/gymie-7c1f238a/kan-29-final-verification-public-profiles-friends @ the dev flow store, 2026-09-04 — a live store read, not re-runnable at a git ref --> The issue proposes *"scoping the gathered context to the
task's declared `**Files:**`"*. The gathered context — `gather-dispatch-context.sh`'s bundle —
carries **no source files at all**: it is `proposal.md`, `design.md`, `tasks.md`, the engineering
principles and the project's lint/test/run sections. There is nothing in it to scope by `Files:`.
What it does carry is the *whole plan*: on KAN-29 the bundle was 210 KB, of which `tasks.md` (46
tasks) was 156 KB — roughly 40K tokens re-read as cached prefix on every one of the implementer's
turns, for a task that needed one of those 46 blocks.
<!-- measured: wc -c proposal.md design.md tasks.md in gymie's spectre/changes/archive/kan-29-final-verification-public-profiles-friends/ plus skills/flow/engineering-principles.md @ gymie main and agents main, 2026-09-04; the token figure is 156 KB at roughly 4 bytes per token --> The scoping that exists to be done is the
plan section, to the task(s) the implementer was actually dispatched with.

## What changes

- `skills/flow/review-panel.md` — `final-review.diff` becomes one combined diff per round across
  every worktree in the change's resolved set, and each included slot is dispatched **once** per
  round against it, in the canonical worktree. The per-slot delta bookkeeping, the diff-size cap
  and the docs-only reduction are restated for the multi-worktree case. Bugbot gets one throwaway
  worktree per repository and one dispatch naming them all; Security gets one dispatch naming every
  worktree. When the set holds more than one worktree, findings qualify `file:line` with the
  worktree and reproducers run from it. KAN-29's eight pass-3 dispatches would have been four.
- `scripts/gather-dispatch-context.sh` — an optional sixth argument, a comma-separated list of task
  ids. When present, the `## tasks.md` section carries the plan header and only the named tasks'
  blocks; a named id absent from the plan exits 2. The five-argument call is byte-identical to
  today.
- `scripts/test-gather-dispatch-context.sh` — cases for the scoped section, document ordering, a
  fenced task-line lookalike, the unknown-id exit and the unchanged five-argument output.
- `skills/flow/implement.md` — the bundle is gathered once per dispatch bundle, immediately before
  that bundle's implementer, to a per-bundle path, with the ids `plan-dispatch-bundles.sh` printed;
  the CONTEXT BUNDLE paragraph names that path and says the plan section is scoped. The panel and
  fix subagents keep the full plan through the five-argument call.

No guard script's verdict or exit code changes. `check-panel-diff-size.sh` and
`check-panel-docs-only.sh` are untouched — the panel runs each once per worktree and combines the
answers in prose.
