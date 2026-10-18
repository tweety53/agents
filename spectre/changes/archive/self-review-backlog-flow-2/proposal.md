# self-review-backlog-flow-2

## Why

Five tickets in KAN's "Backlog Flow" status, filed by earlier self-review passes, are planned here
as one self-review fix change. One half of one ticket is already fixed (`design.md`, **Context**);
the rest name gaps the tree still has:

- **KAN-558 / KAN-565** — kan-469's `design.md` and `tasks.md` cited mockup frame `P6.png`, which no
  file in the mockups root held; nothing at plan time checks that a cited frame exists.
- **KAN-569** — kan-485's copy-removal tasks each ended on an absence sweep, which made "done"
  mechanical; no planning rule asks for it.
- **KAN-570** — kan-485 proved its test failures pre-dated the change by reverting its own touches
  in the failing area and re-running; no contract prescribes that isolation.
- **KAN-603** — the capture-scope report line carries only a sweeps slot, so the frame-bound
  readings promised n/a have nowhere to land, and no **Blocking** item covers the capture-scope
  sweeps line.

## What changes

- A new guard, `check-mockup-citations.sh` (Go, `flow-guard check-mockup-citations`), fails on any
  `<mockups root>/<frame id>.png` token in a plan file that names no file. Writing plans
  (`skills/flow/brainstorm-planner.md`, section D) requires frames cited in that path form and runs
  the guard whenever the project declares `mockups`.
- Writing plans gains **End a removal task on an absence sweep**.
- **The sweep** (`skills/flow-contracts/known-bugs.md`) gains the isolation step for a failure
  suspected to pre-date the change.
- The verifier's report template (`skills/flow/visual-verify-verifier.md`) gains a per-view
  `bands | seams | matrix` n/a line, and **Blocking** (`skills/flow/visual-verify.md`) gains the
  capture-scope sweeps line.
