# self-review-kan-1034-spectre-optional-bundled-planning-prompts-that

## Why

The self-review pass over `kan-1034-spectre-optional-bundled-planning-prompts-that`
(`.superpowers/sdd/self-review-pass.md`) found ten defects, none `big`: guards that let a missing
panel report, a missing reproducer and an unreadable `measured:` comment through, or wrongly refuse
a bold `**none**` repair form and a symlinked relocation path; Markdown layout no guard reads;
spectre declaring no verify lists; and four planning and recording rules the run had to supply by
judgement. Each cost the run a stop, a hand step or a fix round.

## What changes

- `check-panel-findings-closed` refuses a panel close where a completed reviewer dispatch left a
  role's report file absent or empty, or a `fixed` finding's runnable reproducer resolves in no
  worktree of the change; `review-panel.md` has the parent confirm the reproducer path at recording.
- `check-plan-shape` reads `**Tests:** **none** — repairs existing tests:` as the repair form, and
  refuses (new F20) a Baseline `<!-- measured: … -->` comment the parser cannot read.
- `generate-relocation-comparison` skips a scoped path that is a symlink on either side, with a
  stderr note, instead of refusing the whole comparison.
- `check-markdown-integrity` gains two signals — a missing final newline, and consecutive blank
  lines outside a fence — and the three sites in the tree they flag are fixed.
- spectre's `.flow/project.md` declares `## lint` and `## test`, so `run-verify-list.sh` runs and
  records spectre's checks.
- `brainstorm-planner.md` gains two planning rules (a relocation's citations resolve in what its
  destination ships; committed content never cites a change-branch sha); `brainstorm.md` says when
  a planning dispatch's `begin` and `end` are recorded; the `skills/spectre-planning/` index row
  names the re-vendor duty.
