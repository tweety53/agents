# kan-487-flow-panel-review-slots-can-fork-sub-agents

> **Execution:** `/flow-fast` implements this plan inline. Mark a task's own checkbox once its
> commit lands.
> **Relocation:** no

- [x] 1. Strip `Agent` from the `flow-<model>-<effort>` tool allowlist

**Build:** green

**Files:**
- Modify: `agents/flow-haiku-high.md`
- Modify: `agents/flow-haiku-low.md`
- Modify: `agents/flow-haiku-medium.md`
- Modify: `agents/flow-opus-high.md`
- Modify: `agents/flow-opus-low.md`
- Modify: `agents/flow-opus-medium.md`
- Modify: `agents/flow-sonnet-high.md`
- Modify: `agents/flow-sonnet-low.md`
- Modify: `agents/flow-sonnet-medium.md`

**Tests:** **none** — a `tools:` frontmatter allowlist has no automated harness on this repo;
verified by inspection that every file carries the identical line and omits `Agent`
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `fix(agents): strip Agent-tool access from flow-<model>-<effort> family`

  - [ ] **Step 1: add `tools: Read, Glob, Grep, Bash, Write, Edit, ToolSearch` to each file**

  Insert the line directly after each file's `effort:` line, before the closing `---`. Confirm with
  `grep -L '^tools:' agents/flow-*.md` (must print nothing) and `grep -h '^tools:' agents/flow-*.md
  | sort -u` (must print exactly one distinct line, and it must not contain `Agent`).

- [x] 2. Document the structural enforcement in `implement.md`'s dispatch-tree section

**Build:** green

**Files:**
- Modify: `skills/flow/implement.md`

**Tests:** **none** — prose-only documentation change
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow): note structural Agent-tool exclusion on the closed dispatch list`
**After:** Task 1

  - [ ] **Step 1: extend the "whole run's dispatch tree" paragraph**

  Add one sentence noting that on `REVIEW_PANEL_TOGGLE: dynamic`, every one of the four closed-list
  dispatch rows runs on the `flow-<model>-<effort>` family, whose `tools:` allowlist omits `Agent`
  — the NO DELEGATION prose is now backed by a capability the dispatched agent structurally lacks,
  not only by prompt text — and that `default`'s `general-purpose` reviewer dispatch is unaffected,
  since it is a harness-provided type this repository does not own.

- [x] 3. Close the "architectural decision" ask-loophole in `/flow-fast`'s auto-pick rule

**Build:** green

**Files:**
- Modify: `skills/flow-fast/brainstorm.md`

**Tests:** **none** — prose-only documentation change
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow-fast): auto-pick recommended options regardless of architectural weight`
**After:** none

  - [ ] **Step 1: state the two grounds to ask are the whole test**

  Add a paragraph after the two bulleted ask-grounds in section B stating that a round with a
  recommended default is auto-picked even when the decision is architectural, touches a
  shared/core file, or is hard to reverse — closing the loophole this run's own operator correction
  named (KAN-487 session).
