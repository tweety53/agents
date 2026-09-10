# kan-463-flow-planner-splits-desktoptest-ui-tests-into

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

## Global constraints

- The paragraph's text is fixed by `design.md` `## Change` and was approved by the operator.
  Reproduce it exactly — its wording, its four clauses and its backticked names are the
  deliverable, not a sketch of one.
- The paragraph deliberately uses "carries / names / is needed" rather than `MUST` or `SHALL`, so
  `scripts/check-normative-inventory.sh`'s inventory is unchanged by this edit. Do not "strengthen"
  the wording.
- Nothing else in `skills/flow/brainstorm-planner.md` moves. No other file is edited.

- [x] 1. Add the UI-test follow-on-task paragraph to `skills/flow/brainstorm-planner.md` section D

**Build:** green

**Files:**
- Modify: `skills/flow/brainstorm-planner.md`

**Tests:** **none** — one paragraph of instruction prose in a skill file; `design.md`
`## Verification` records why no guard row is added and what runs instead.
**Regression:** n/a — no test declared. Reverting this commit restores the gap KAN-463 reports:
section **D** again says nothing about splitting a feature's UI tests from its source files.
**Baseline:** n/a — no test declared.
**Commit:** `docs(flow): plan a feature's UI tests as their own follow-on task`

  - [x] **Step 1: Capture the normative-inventory baseline before editing**

  From the worktree root:

  ```bash verified:run at fb8f38a on the change branch — exits 0, payload on stdout, counts on stderr
  scripts/check-normative-inventory.sh > /tmp/kan-463-inventory-before.txt
  ```

  This guard writes its payload to stdout and its file and sentence counts to stderr, and has no
  violation exit code — it reports a set, and comparing two of them is the caller's act, per
  `.flow/project.md`'s `## lint` note on it. Capture it now, before any edit; Step 3 diffs against
  this file.

  - [x] **Step 2: Insert the paragraph**

  In `skills/flow/brainstorm-planner.md`, section **D. Basic Workflow #3 — Writing plans**, insert
  the block below as its own paragraph. It goes **immediately after** the verify-step blockquote —
  the one beginning "> **Write each task's verify step as its own lint commands plus targeted
  tests.**" and ending "> lint alone and no test command." — and **immediately before** the
  paragraph beginning "**Load `skills/flow-contracts/plan-provenance.md`.**", separated from each by
  one blank line.

  ```markdown verified:trial-inserted at this exact position on the change branch at fb8f38a, guards green, then reverted
  > **Write a feature's UI tests as their own follow-on task.** When a feature's tests live in a
  > UI-test source set of their own — Compose Multiplatform's `desktopTest`, and the like — the plan
  > carries two tasks: the feature task, whose `**Files:**` names the source files and the unit-test
  > (`commonTest`) files, then one follow-on task per feature task whose `**Files:**` names only that
  > feature's UI-test file(s), all of them. The two are file-disjoint, so `plan-dispatch-bundles.sh`
  > keeps them separate bundles and the UI-test iteration starts in a fresh implementer at a small
  > context instead of the one that just wrote the feature. No `**After:**` field is needed — the
  > serial default already runs the follow-on after its feature task — and the last bundle's FULL
  > SUITE run still covers the pair.
  ```

  Reflow the blockquote's line breaks to this repository's prose width if your editor differs, but
  change no word, no backtick and no punctuation. The four clauses — the UI-test-source-set
  trigger with `desktopTest` as the example, the two-task split with all UI-test files in the one
  follow-on, the file-disjoint/fresh-context rationale, and the `**After:**`/FULL SUITE
  non-changes — must all survive.

  - [x] **Step 3: Verify**

  Run exactly these, from the worktree root — the lint commands this task's own `**Files:**` need,
  and no others:

  ```bash verified:run on the trial insert at fb8f38a — all four guards exit 0 and the diff printed nothing
  scripts/check-references.sh
  scripts/check-vocabulary.sh
  scripts/check-markdown-integrity.py
  scripts/check-contract-budget.sh
  scripts/check-normative-inventory.sh > /tmp/kan-463-inventory-after.txt
  diff /tmp/kan-463-inventory-before.txt /tmp/kan-463-inventory-after.txt
  ```

  All four guards exit 0. `check-contract-budget.sh` has headroom: the file is 28392 bytes against
  a declared budget of 31664, and the block takes it to 29217.
  <!-- measured: wc -c skills/flow/brainstorm-planner.md before and after the trial insert, and the budgets() row in scripts/check-contract-budget.sh @ branch spectre/kan-463-flow-planner-splits-desktoptest-ui-tests-into -->

  The `diff` must print nothing. A difference means the wording drifted into `MUST`, `MUST NOT`,
  `SHALL` or `SHALL NOT`; reword back to the design's text rather than accepting the new inventory,
  per that guard's own convention.

  Do not run the whole `## lint` list, `scripts/run-guard-tests.sh`, `go test ./...` or
  `npm test` here: no Go, TypeScript or script file is touched, and the full lists run once at the
  last bundle and again in `flow.verify`.
