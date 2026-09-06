# kan-465-flow-planner-must-write-targeted-tests

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

## Global constraints

- The paragraph's text is fixed by `design.md` `## Change` and was approved verbatim by the
  operator. Reproduce it exactly — its wording, its three clauses and its backticked selectors are
  the deliverable, not a sketch of one.
- The paragraph deliberately uses "names" and "never" rather than `MUST` or `SHALL`, so
  `scripts/check-normative-inventory.sh`'s inventory is unchanged by this edit. Do not "strengthen"
  the wording.
- Nothing else in `skills/flow/brainstorm-planner.md` moves. No other file is edited.

- [x] 1. Add the verify-step paragraph to `skills/flow/brainstorm-planner.md` section D

**Build:** green

**Files:**
- Modify: `skills/flow/brainstorm-planner.md`

**Tests:** **none** — one paragraph of instruction prose in a skill file; `design.md`
`## Verification` records why no guard row is added and what runs instead.
**Regression:** n/a — no test declared. Reverting this commit restores the gap KAN-465 reports:
section **D** again says nothing about what a task's verify step may run.
**Baseline:** n/a — no test declared.
**Commit:** `docs(flow): plan a task's verify step as targeted tests, not the suite`

  - [x] **Step 1: Capture the normative-inventory baseline before editing**

  From the worktree root:

  ```bash verified:run at main 3865c4b — exits 0, payload on stdout, counts on stderr
  scripts/check-normative-inventory.sh > /tmp/kan-465-inventory-before.txt
  ```

  This guard writes its payload to stdout and its file and sentence counts to stderr, and has no
  violation exit code — it reports a set, and comparing two of them is the caller's act, per
  `.flow/project.md`'s `## lint` note on it. Capture it now, before any edit; Step 3 diffs against
  this file.

  - [x] **Step 2: Insert the paragraph**

  In `skills/flow/brainstorm-planner.md`, section **D. Basic Workflow #3 — Writing plans**, insert
  the block below as its own paragraph. It goes **immediately after** the task-shape paragraph —
  the one beginning "**Tell it the task shape, because it is spectre's and not writing-plans'
  own.**" and ending "... the `Placement` paragraph under **The build-green tag**
  (`skills/flow-contracts/build-green.md`)." — and **immediately before** the paragraph beginning
  "**Load `skills/flow-contracts/plan-provenance.md`.**", separated from each by one blank line.

  ```markdown unverified:to be written on the change branch
  > **Write each task's verify step as its own lint commands plus targeted tests.** A verify step
  > names the lint commands the task's own `**Files:**` actually need — never the project's whole
  > `## lint` list — and the build tool's own selector for each `**Tests:**` entry (`--tests
  > '<class>'` for Gradle, `-run '<name>'` for `go test`, `-t '<name>'` for vitest), never the bare
  > module or repository suite: that run belongs to the last bundle's FULL SUITE paragraph
  > (`skills/flow/implement.md`) and to `flow.verify`. A task whose `**Tests:**` is `none` names
  > lint alone and no test command.
  ```

  Reflow the blockquote's line breaks to this repository's prose width if your editor differs, but
  change no word, no backtick and no punctuation. The three clauses — targeted selector per
  `**Tests:**` entry, lint bounded to the task's own `**Files:**`, and lint alone for a
  `**Tests:** none` task — must all survive.

  - [x] **Step 3: Verify**

  Run exactly these, from the worktree root — the lint commands this task's own `**Files:**` need,
  and no others, which is the paragraph this task adds applied to itself:

  ```bash unverified:to be run on the change branch, after Step 2's edit
  scripts/check-references.sh
  scripts/check-vocabulary.sh
  scripts/check-markdown-integrity.py
  scripts/check-contract-budget.sh
  scripts/check-normative-inventory.sh > /tmp/kan-465-inventory-after.txt
  diff /tmp/kan-465-inventory-before.txt /tmp/kan-465-inventory-after.txt
  ```

  All four guards exit 0. `check-references.sh` is the one that can genuinely fail here — it
  resolves the new `skills/flow/implement.md` citation. `check-contract-budget.sh` has headroom:
  the file is 19188 bytes against a declared budget of 23248, and the block adds roughly 600.
  <!-- measured: wc -c skills/flow/brainstorm-planner.md and the budgets() row in scripts/check-contract-budget.sh @ main 3865c4b -->

  The `diff` must print nothing. A difference means the wording drifted into `MUST`, `MUST NOT`,
  `SHALL` or `SHALL NOT`; reword back to "names / never" rather than accepting the new inventory,
  per that guard's own convention.

  Do not run the whole `## lint` list, `scripts/run-guard-tests.sh`, `go test ./...` or
  `npm test` here: no Go, TypeScript or script file is touched, and the full lists run once at the
  last bundle and again in `flow.verify`.
