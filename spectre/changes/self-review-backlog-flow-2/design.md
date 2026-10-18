## Context

- Five Backlog Flow findings planned as one change per item 1 of **3. Fix every finding that is
  not `big`** (`skills/flow-self-review/SKILL.md`). Brainstorm design:
  `.superpowers/sdd/2026-10-18-self-review-backlog-flow-2-design.md` in this worktree. Every
  question was auto-resolved to its recommended option (`skills/flow-contracts/operator-prompts-auto-resolution.md`).
- **KAN-558 — Task 1.** `check-frame-list` (`stats/internal/guard/framelist.go`, KAN-549) derives
  the declared frames from files that exist in the mockups root, so a frame `design.md` cites with
  no file is silently undeclared; no plan-time check exists. Blast radius: seven files (Go source,
  Go test, shim, its skill symlink, `skills/flow/brainstorm-planner.md`,
  `skills/flow/SKILL-rationale.md`, `.flow/project.md`'s list of ported guards); the planner
  blockquote it extends has no other citer — under 60.
- **KAN-565 — first half Task 1 (the same defect as KAN-558); second half already fixed, no
  task.** An always-zero generic reproducer reads "defect not demonstrated", and
  `check-panel-reproducer-exit-contract.sh` (KAN-554) refuses an open finding's reproducer that
  does not demonstrate, before the panel relies on it; `prove-reproducer.sh` (KAN-663) requires the
  pre-fix leg demonstrated and the fix leg not, so an always-zero `# mutation-reproducer` fails its
  fix leg. The ticket itself records `run-reproducer.sh` refusing kan-469's always-zero script.
- **KAN-569 — Task 2.** `grep -rn "absence sweep" skills` finds nothing. Blast radius: two files.
- **KAN-570 — Task 3.** **The sweep** (`skills/flow-contracts/known-bugs.md`) is canonical for a
  pre-existing test failure and says it blocks, but nothing says how a suspected one is told from
  the change's own. Blast radius: one file; ten files cite `known-bugs.md` and none needs a change.
- **KAN-603 — Task 4.** The per-view `sweeps:` line is the template's only capture-scope slot
  (`skills/flow/visual-verify-verifier.md`), while the prose above it promises bands, seams and the
  matrix n/a — no frame; **Blocking** (`skills/flow/visual-verify.md`) names no capture-scope
  duty, and `check-verify-report.sh` admits step 10 as n/a. Blast radius: two files.
- Every task's diff was applied in order to a scratch clone at `7d7a1b40` (this worktree's base):
  `gofmt -l` and `go vet ./internal/guard` clean, `go test ./internal/guard/` and
  `scripts/run-guard-tests.sh` (43/43) pass, `TestCheckMockupCitations` fails with the
  file-existence check disabled, and `check-references.sh`, `check-guard-symlinks.sh`,
  `check-dispatch-paragraphs.sh`, `check-installed-citations.sh`, `check-stage-mark-calls.sh`,
  `check-hand-notes-in-step.sh`, `check-ask-silence.sh`, `check-markdown-integrity.py` and
  `check-verbatim-moves.sh 7d7a1b40` (with this change's `verbatim-moves.txt`) exit 0.

## Decisions

### A plan-time guard over path-form frame citations

**ID:** mockup-citations-guard
**Status:** active
**Chosen:** a new `flow-guard check-mockup-citations` behind `scripts/check-mockup-citations.sh <project root> <mockups root> <file>...`, its header the contract: every `<mockups root>/<frame id>.png` token in `design.md` and `tasks.md` must name a regular file, else exit 1 naming each `<file>:<line> <citation>`. Section D of `skills/flow/brainstorm-planner.md` requires each frame cited in that path form and runs the guard before continuing whenever `mockups` is declared; its exit 1 is the missing handoff asset **The checklist** already refuses to plan against. Measured on gymie's archive @ 3c57055d: its 100 path-form citations all resolve, and kan-469's `P6.png` rewritten in path form is refused.
**Considered:** judging bare `<frame id>.png` tokens too — gymie's archived plans carry 178 bare `.png` tokens naming no mockup file, nearly all capture-screenshot names, so the guard would refuse legitimate plans; judging ids by shape (an uppercase letter) — `screens-A.png` and `A4-settings-darwin.png` defeat it; folding the check into `check-frame-list` — that guard runs at visual verification, after the plan has been built on the missing frame.

### The reproducer half of KAN-565 is already closed

**ID:** reproducer-validator-exists
**Status:** active
**Chosen:** no task — `check-panel-reproducer-exit-contract.sh` and `prove-reproducer.sh` already refuse a reproducer whose exit cannot demonstrate the defect.
**Considered:** a static scan for scripts with no non-zero exit path — the two runners already decide it by running the script, which a static scan can only approximate.

### Absence sweeps are a writing-plans rule

**ID:** absence-sweep-rule
**Status:** active
**Chosen:** a section-D blockquote, **End a removal task on an absence sweep**, beside the other per-task-shape rules, so it reaches each removal task's steps and through them the implementer.
**Considered:** a guard over `tasks.md` requiring a `git grep` step in removal tasks — no mechanical reading tells a removal task from any other.

### Pre-existing-failure isolation lives in the sweep

**ID:** isolate-before-classifying
**Status:** active
**Chosen:** a paragraph in **The sweep** (`skills/flow-contracts/known-bugs.md`), the contract canonical for pre-existing failures: revert the change's touches in the failing area to merge-base content, re-run, restore, then classify; the classification never decides whether it blocks.
**Considered:** a rule in `skills/flow/verify-fix-loop.md` — the sweep is cited from every stage that meets a pre-existing failure, the loop from verify alone.

### The capture-scope report gets its slots and its tripwire

**ID:** capture-scope-report
**Status:** active
**Chosen:** one template line, `- per view, no frame composed: <view id> bands | seams | matrix — n/a — no frame`, under the capture-scope sweeps line, and a **Blocking** item for a captured view with no frame composed whose report carries no capture-scope `sweeps:` line or leaves a sweep neither done nor n/a with its reason.
**Considered:** extending `check-verify-report.sh` to parse per-view lines — the ticket asks for the contract change, and the guard reads none of the report's per-view lines today.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — one pure-addition guard whose table cases pin the token boundary and the existence check, and three prose additions, every diff applied and checked in a scratch clone.
**Considered:** `flow-medium` — the guard gates planning, but a false refusal shows at once in its verdict line, and it reads only the plan files and the mockups root.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-backlog-flow-2-fix`, Tasks 1, 2, 3 and 4 in that order.
**Considered:** one fixer per task — each task is one `git apply` of a checked diff, and Tasks 1 and 2 both edit `skills/flow/brainstorm-planner.md` and `skills/flow/SKILL-rationale.md`, so they run in series anyway.

## Open questions
