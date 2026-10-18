## Context

- Five Backlog Flow findings planned as one change per item 1 of **3. Fix every finding that is
  not `big`** (`skills/flow-self-review/SKILL.md`). Brainstorm design:
  `.superpowers/sdd/2026-10-18-self-review-backlog-flow-1-design.md` in this worktree. Every
  question was auto-resolved to its recommended option (`skills/flow-contracts/operator-prompts-auto-resolution.md`).
- **KAN-397 — disproved, no task.** The refusal it names came from `prepare-archive-branch.sh`'s
  flat count of the main checkout's untracked entries. Commit 9721e67e (`fix(finish-contract):
  archive and save the self-review bundle at integrate on the change branch; run 2 only cleans
  up`) deleted `scripts/prepare-archive-branch.sh` and `scripts/classify-untracked.sh` with the
  landing worktree and the archive branch; no script in `scripts/` reads the main checkout's
  untracked entries any more — `check-main-checkout-drift.sh` and `check-foreign-staged.sh` pin
  `--untracked-files=no` — so stray captures in the main checkout refuse nothing.
- **KAN-539 — first half disproved, second half Task 1.** "Require the mockup source located
  before fidelity/icon tasks are planned" is already the rule: commit 32d25c2c (`docs(flow):
  require mockup handoff assets reachable before frame-spec tasks`) made a frame-specified design
  unapprovable, and forbade any task whose specification is that frame, until the handoff assets
  are committed or their location recorded (**The checklist**, `skills/flow/brainstorm-planner.md`).
  The commit-message half is live: **Revert and back-out commits**
  (`skills/flow-contracts/git-boundaries.md`) covers only a landed-then-backed-out commit's body.
  Blast radius: the two edited files, plus the five files citing the section name — under 60.
- **KAN-544 — disproved, no task.** Each fix it asks for is in the tree. c4d262f's sweeps run on
  every `flow.visual-verify` verifier dispatch: the report carries a `sweeps:` line per frame with
  each sweep `done` or `n/a — <why>`, and a sweep in any other state makes the report incomplete
  before **Blocking** is applied (`skills/flow/visual-verify.md`). A self-captured baseline is
  never the comparison reference: "A `toHaveScreenshot` passing over a baseline this change wrote
  is the app agreeing with itself, never with the mockup — step 10 is the only comparison"
  (`skills/flow/visual-verify-verifier.md`, step 8, from 6adb3dc2). The frame list is the
  change's, never the round's, and the `frames:` line counts against it (same commit). Its
  mechanical reconciliation is KAN-549's, Task 3.
- **KAN-545 — Task 2.** No rule names locators: the verifier's step 4 warns only about its own
  scripted navigation. Blast radius: three edited files plus the five citing the blockquote's
  name — under 60.
- **KAN-549 — Task 3.** The parent's reconciliation of the `frames:` line against `design.md`'s
  own frame list is prose the parent reads (`skills/flow/visual-verify.md`); no guard compares
  them. Blast radius: seven files (shim, its skill symlink, Go source, Go test,
  `skills/flow/visual-verify.md`, `skills/flow/SKILL-rationale.md`, `.flow/project.md`'s list of
  ported guards) — under 60.
- Every task's diff was applied in order to a scratch clone at `7aea62c2` (this worktree's base):
  `gofmt -l .` and `go vet ./...` clean, `go test ./internal/guard/` and
  `scripts/run-guard-tests.sh` pass, `TestCheckFrameList` fails with the count check removed, and
  `check-references.sh`, `check-guard-symlinks.sh`, `check-dispatch-paragraphs.sh`,
  `check-installed-citations.sh`, `check-stage-mark-calls.sh`, `check-hand-notes-in-step.sh`,
  `check-ask-silence.sh`, `check-markdown-integrity.py` and `check-verbatim-moves.sh 7aea62c2`
  (with this change's `verbatim-moves.txt`) exit 0.

## Decisions

### A commit's message describes its net tree

**ID:** net-tree-message
**Status:** active
**Chosen:** a paragraph in **Revert and back-out commits** (`skills/flow-contracts/git-boundaries.md`) — an attempt undone before it lands, in the working tree or by a fixup folded into the commit, is never what the commit's subject or body describes; the body names it only as abandoned, with why, or the attempt lands and is backed out as two commits. The implementer prompt in `skills/flow/implement.md` cites it beside the withdrawn-task sentence it already carries.
**Considered:** a guard comparing a commit's message to its diff — no mechanical reading tells whether a message describes a tree; a rule in `implement.md` alone — the fix-round fixup fold is not the implementer's, and the section is where commit-message rules already live.

### Structural locators in end-to-end and fidelity specs

**ID:** structural-locators
**Status:** active
**Chosen:** a writing-plans blockquote in section D (`skills/flow/brainstorm-planner.md`) beside **Write end-to-end spec and fidelity-capture tasks to run side by side** — every click and capture reaches its control by role and accessible name, label, text or test id, never a page coordinate — so the rule rides each such task's steps into the implementer's prompt; the verifier's capture step cites the section, since the verifier also extends capture specs.
**Considered:** a re-anchor probe after every task that shifts layout — it detects the drift a locator removes, and adds a step to every such task; a lint over spec files for `page.mouse.click(` — the specs live in consuming projects' checkouts, outside this repository, and a coordinate is legitimate in `page.mouse.move(0, 0)`'s pointer park.

### The frame list is reconciled by a guard

**ID:** frame-list-guard
**Status:** active
**Chosen:** a new `flow-guard check-frame-list` behind `scripts/check-frame-list.sh <design.md> <mockups dir> <report file>`, its header the contract, run by `flow.visual-verify`'s parent whenever `mockups` is declared, before the rest of the reconciliation. A `no sidecar line` gap blocks under **Blocking**; every other gap makes the report incomplete like the by-hand reconciliation's list.
**Considered:** folding it into `check-verify-report.sh` — that guard's exit 1 dispatches a verifier remainder, which cannot cure a missing sidecar line a capture task owes; a new section in `design.md` listing frame ids — a second copy of the list to drift, where the mockups directory already names every frame.

### The declared list is derived from the mockups directory

**ID:** frame-list-derived
**Status:** active
**Chosen:** a frame is declared when `<mockups dir>/<frame id>.png` exists and `design.md` names `<frame id>` as a whole token (not preceded or followed by a letter, digit, `_` or `-`); sidecars are `<spec>.mockups` beside every `- spec:` path the report names; a declared frame is reported when a `- <frame id>:` or `- <frame id> <…>:` line exists or the `frames:` line names it; a `frames:` total below the declared count is a gap, one above it is not, since the verifier's list also takes frames the diff changed and `moved:` frames.
**Considered:** reading `tasks.md` too — the ticket and `skills/flow/visual-verify.md` name `design.md`'s own list; failing a `frames:` total above the declared count — it would refuse the verifier's legitimate additions.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — two prose paragraphs and one pure-addition guard on `check-verify-report`'s shape, its table cases pinning each gap and the token boundary, every diff applied and checked in a scratch clone.
**Considered:** `flow-medium` — the guard gates `flow.visual-verify`, but a false gap shows at once in the stage's verdict line and the derivation reads only files the stage already has.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-backlog-flow-1-fix`, Tasks 1, 2 and 3 in that order.
**Considered:** one fixer per task — each task is one `git apply` of a checked diff, and Tasks 2 and 3 both edit `skills/flow/SKILL-rationale.md`, so they run in series anyway.

## Open questions
