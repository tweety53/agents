# kan-876-flow-speed-up-the-e2e-fidelity-live-verification

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** end-to-end spec and fidelity-capture tasks run as one parallel wave on one stack the
parent starts once, the live-verification task folds into `flow.verify`, and specs stop waiting out
production timings.

**Architecture:** Task 1 adds the one-stack rule to `skills/flow/implement.md`. Task 2 adds the live
check to `flow.verify` and makes run-instructions reuse a stack already serving the worktree's build. Task 3
rewrites the planner rules in `skills/flow/brainstorm-planner.md` that cite both, plus the rationale
and README lines that describe the old live-verification task.

**Spec:** `design.md` beside this file is canonical for every scope decision. Tasks cite its
`## Decisions` by ID and never restate them.

## Global Constraints

- A commit's scope names the module it moved (`rules/commit-scope-is-the-module.mdc`).
- Task order is load-bearing: `scripts/check-references.sh` resolves every `**Section**`
  (`path`) citation to a heading in that file, so Task 2 cites Task 1's new heading and Task 3 cites
  both tasks' new headings.
- Every run-loaded sentence a task adds, deletes or rewords is listed, exactly as
  `scripts/check-verbatim-moves.sh` prints it after `::`, in
  `spectre/changes/kan-876-flow-speed-up-the-e2e-fidelity-live-verification/verbatim-moves.txt`,
  under a first line `# Operator, 2026-10-05: KAN-876 — verification tasks fan out, one stack, live
  check in verify.` The file is edited in the worktree and **left uncommitted**: it is a planning
  path, carried by the next `chore(spectre): plan` commit (**Planning commits**,
  `skills/flow-contracts/git-boundaries.md`).
- Every task's verify step runs `scripts/check-references.sh`, `scripts/check-verbatim-moves.sh`,
  `scripts/check-markdown-integrity.py` and `scripts/check-guard-symlinks.sh`; no task adds a test —
  this change is pipeline prose.
- Never stop or restart `flowd`, `flow-postgres` or the `flow` database (`CLAUDE.md`).

## Review Focus

- **Fix run after a stack start.** A fix that changes code after the stack started must still get a
  restart: run-instructions skips its start only when `check-dev-stack-fresh.sh` exits 0, never on
  exit 1 or 2.
- **No UI path, runtime touched.** The live check runs in `flow.verify`, so a change with no
  `ui paths` match still gets it.
- **Inline execution.** The one-stack rule binds the parent on an `inline` run too, not only the
  `sdd` waves.
- **Verification-only group pin.** Dropping the live-verification record from the
  `opus`/`low` verification-only group leaves that pin intact for spec and capture groups.

---

- [x] 1. Implement: one stack for the verification tasks

**Files:** `skills/flow/implement.md`
**Tests:** **none** — pipeline prose.
**Regression:** none — prose.
**Baseline:** before=0 after=0
<!-- predicted: no test files in this task -->
**Commit:** `feat(flow): start the verification tasks' stack once`
**After:** none
**Build:** green

**Decision:** stack-started-once

  - [x] **Step 1: Add the heading.** In `skills/flow/implement.md` section `## 4. Execute (SDD + TDD)`,
    directly after the paragraph that opens `**Load \`skills/flow/sdd-dispatch.md\` only when**`,
    insert:

```markdown unverified:new prose — check-references.sh and check-verbatim-moves.sh pass once listed
### One stack for the verification tasks

**Before the first task whose `**Files:**` only add end-to-end specs, fidelity captures or their
baselines goes out** — dispatched under `sdd`, or started by the parent on an `inline` run — the
parent starts the canonical worktree's stack once: the project's `## stop` command when it
declares one, then its `## run`, then any build the plan's Global Constraints name for those
tasks, then `check-dev-stack-fresh.sh <canonical-worktree>`. Every such task, and every wave copy
it runs in, runs against that stack's worktree-resolved URLs, which its prompt names, and
starts, stops and rebuilds nothing. The stack stays up for `flow.verify`. A start that fails
twice ends the turn with `## Question` naming the command and its output, verbatim. Never the
flow dev stack (`<project>/CLAUDE.md`).
```

  - [x] **Step 2: Acknowledge.** Run `scripts/check-verbatim-moves.sh`; append every sentence it
    prints after `::` to `verbatim-moves.txt` (Global Constraints), creating the file with its first
    line.
  - [x] **Step 3: Verify.** Run `scripts/check-references.sh`, `scripts/check-verbatim-moves.sh`,
    `scripts/check-markdown-integrity.py` and `scripts/check-guard-symlinks.sh`; each exits 0.
  - [x] **Step 4: Commit** `skills/flow/implement.md` alone, with the `**Commit:**` subject.

Correction (2026-10-05): the plan placed the new `###` heading directly after the `sdd-dispatch.md` load directive, mid section 4; there it would have become the parent heading of the rest of section 4 (about 520 lines). It shipped as section 4's last subsection, after the `flow.sdd-tdd` stage close, so no existing content changes heading. Its opening clause ("Before the first task … goes out") locates it in time, and phase files are read in full at stage start.

- [x] 2. Verify: the live check, and reuse of the running stack

**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** **none** — pipeline prose.
**Regression:** none — prose.
**Baseline:** before=0 after=0
<!-- predicted: no test files in this task -->
**Commit:** `feat(flow): run the live check in verify and reuse the running stack`
**After:** Task 1
**Build:** green

**Decision:** live-check-in-verify

**Decision:** stack-started-once

  - [x] **Step 1: `create` paragraph.** In `skills/flow/verify-and-handoff.md` `## Verify`, replace
    `and this step starts none of them — it exports, lints, tests, and hands off.` with `and this
    step starts them only for **Live check** below, through the project's own \`## run\`.`
  - [x] **Step 2: Live check section.** Insert after `### Inline verify`'s `**Recording.**`
    paragraph, before `**Load \`skills/flow-contracts/session-records.md\`**`:

```markdown unverified:new prose — check-references.sh and check-verbatim-moves.sh pass once listed
### Live check

**Runs only when `design.md` carries a `## Live check` section naming something to exercise**
(**D. Basic Workflow #3 — Writing plans**, `skills/flow/brainstorm-planner.md`); a one-line
no-runtime section, or none, skips it. After the lint and test list, with the parent's own Bash
calls: when the stack's worktree-resolved URLs do not answer, or `check-dev-stack-fresh.sh
<worktree>` exits 1, start it as **One stack for the verification tasks**
(`skills/flow/implement.md`) does; then exercise what the section names and write each
before/after figure, beside the failure look the section states, to
`<changeRoot>/live-verification.md`, a planning path the write-in-progress planning commit
carries. The `## Report` gains one line, `live check — <matched | failure look>`. A figure that
shows the failure look is the branch's own defect and takes **The loop**
(`skills/flow/verify-fix-loop.md`); a stack that will not start ends the turn as a failed
command's environment cause does. The stack stays up for **Visual verification** and **Resolve
the run instructions**.
```

  - [x] **Step 3: Run-instructions start.** In `## Resolve the run instructions`, after the sentence
    ending `from the project's \`## run\` commands.`, insert: `**A stack already serving this
    worktree's build is left running**: when its URLs answer and \`check-dev-stack-fresh.sh
    <worktree>\` exits 0 before the start, the start is skipped — the stack the verification tasks
    or **Live check** started is the one handed off. Exit 1 or 2 starts it as above.`
    `skills/flow/visual-verify.md` is not edited: its step 5 already starts nothing when the stack
    answers, and its step 13 stops only a stack it started itself.
  - [x] **Step 4: Acknowledge.** Run `scripts/check-verbatim-moves.sh`; append every sentence it
    prints after `::` to `verbatim-moves.txt`.
  - [x] **Step 5: Verify.** Run `scripts/check-references.sh`, `scripts/check-verbatim-moves.sh`,
    `scripts/check-markdown-integrity.py` and `scripts/check-guard-symlinks.sh`; each exits 0.
  - [x] **Step 6: Commit** the file, with the `**Commit:**` subject.

- [ ] 3. Planner: fan-out, live check and spec-timing rules

**Files:** `skills/flow/brainstorm-planner.md`, `skills/flow/SKILL-rationale.md`, `README.md`
**Tests:** **none** — pipeline prose.
**Regression:** none — prose.
**Baseline:** before=0 after=0
<!-- predicted: no test files in this task -->
**Commit:** `feat(flow): plan verification tasks to fan out and fold the live walk into verify`
**After:** Task 2
**Build:** green

**Decision:** verification-tasks-fan-out

**Decision:** live-check-in-verify

**Decision:** no-production-waits-in-specs

  - [ ] **Step 1: Replace the live-verification paragraph.** In `skills/flow/brainstorm-planner.md`
    section D, replace the whole quoted paragraph opening `> **Write a live-verification task when
    the change touches a running service or persistent` with:

```markdown unverified:new prose — check-references.sh and check-verbatim-moves.sh pass once listed
> **Write a live check when the change touches a running service or persistent state — never a
> live-verification task.** When the change under plan touches a long-running service, a daemon,
> a store, a scheduler or anything else holding runtime state, `design.md` carries a
> `## Live check` section naming what to exercise against the real thing — the running system
> against its real state, not its fakes — the before/after figures to record, and what the
> change **not** working would look like, so a null result is recognisable as failure rather
> than read as success: a check that only asserts a command exited 0 cannot tell a repaired
> system from an untouched one. `flow.verify` runs it (**Live check**,
> `skills/flow/verify-and-handoff.md`); no task in `tasks.md` does. When the planner judges the
> change has no runtime to verify against, the section is one line saying so — the
> justification is required, the check is not.
```

  - [ ] **Step 2: Add the fan-out and timing rules** directly after the quoted paragraph opening
    `> **Write a feature's UI tests as their own follow-on task.**`:

```markdown unverified:new prose — check-references.sh and check-verbatim-moves.sh pass once listed
> **Write end-to-end spec and fidelity-capture tasks to run side by side.** Each such task's
> `**After:**` names only the feature tasks whose behaviour it exercises — never another
> verification task, since one spec file never needs another's commit — and each seeds its own
> accounts and data, so the tasks share one running stack (**One stack for the verification
> tasks**, `skills/flow/implement.md`) without touching each other's state. Global Constraints
> names the stack's start row and any build the tasks run against; no task's steps start, stop
> or rebuild the stack.

> **Write a spec so it never waits out a production timing.** A scenario that depends on a
> production timeout, retry backoff or polling interval plans a test-only override that
> shortens it — the production value unchanged — in the task that adds the spec, and a suite
> whose specs are independent runs on the test runner's parallel workers.
```

  - [ ] **Step 3: Decide step 4 split.** In `### Decide` step 4, after the sentence ending `would
    collapse a real parallel wave.`, insert: `**A mechanical group holding two or more
    verification-only bundles** — end-to-end specs or fidelity captures, none named in another's
    after-set — **is split, one group per such bundle**, \`groups_override\` naming the parallel
    wave, since the chain merge joins each to the group holding the feature task it waits on and
    runs them in series.`
  - [ ] **Step 4: Model and effort.** In `#### Model and effort`, replace `fidelity captures or
    baselines against design frames, or a live-verification record of an already-built feature,`
    with `fidelity captures or baselines against design frames of an already-built feature,`.
  - [ ] **Step 5: Rationale and README.** In `skills/flow/SKILL-rationale.md`, retitle `### brainstorm-planner.md — D. Writing plans (live-verification task)` to `### brainstorm-planner.md — D. Writing plans (live check)`, keep its bullet, and add a bullet: `- *…never a live-verification task …* — (KAN-876) KAN-754's live task ran after the spec and capture tasks in one serial group and repeated, against a stack of its own, what \`flow.verify\` and visual verify already do against the running app.` Add under the same heading a bullet for the fan-out rule: `- *Write end-to-end spec and fidelity-capture tasks to run side by side* — (KAN-876) KAN-754 chained its fidelity task after its spec task with no file dependency, and the chain merge then ran them in series.` In the `## brainstorm-planner.md — Model and effort, the verification-only group (opus-low-verification)` section, drop `, a live-verification record` from its first sentence and add a closing sentence `KAN-876 folded the live-verification record into \`flow.verify\`; the pin now covers spec and capture groups.` In `README.md`'s class table, `(end-to-end specs, fidelity captures, a live-verification record)` becomes `(end-to-end specs, fidelity captures)`.
  - [ ] **Step 6: Acknowledge.** Run `scripts/check-verbatim-moves.sh`; append every sentence it
    prints after `::` to `verbatim-moves.txt`.
  - [ ] **Step 7: Verify.** Run `scripts/check-references.sh`, `scripts/check-verbatim-moves.sh`,
    `scripts/check-markdown-integrity.py`, `scripts/check-guard-symlinks.sh` and
    `scripts/check-plan-shape.sh`; each exits 0.
  - [ ] **Step 8: Commit** the three files, with the `**Commit:**` subject.
