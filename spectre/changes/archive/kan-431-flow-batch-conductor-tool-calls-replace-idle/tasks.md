# kan-431-flow-batch-conductor-tool-calls-replace-idle

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

**Spec:** `design.md` beside this file. Three tasks, all `Build: green`, documentation only, one
phase file each. Every quoted anchor below was read from `main` at `41688f0`; an anchor that has
moved is found by its quoted text, never by its line number. `docs/superpowers/research/` and
this change's own directory are never in a task commit. No script changes: the three scripts the
issue names are what the batched calls chain, unchanged.

- [x] 1. State turn discipline and the batched task boundary in `skills/flow/implement.md`

**Build:** green
**Files:** `skills/flow/implement.md`
**Tests:** none — documentation; the shipped guards in step 5 are the check
**Regression:** none — no test is added; `scripts/check-dispatch-paragraphs.sh` and `scripts/check-stage-mark-calls.sh` are the mechanical checks and pass before and after
**Baseline:** before=0 after=0
<!-- measured: no test suite covers skills/ prose; the count is the number of tests this task touches @ 41688f0 -->
<!-- predicted: unchanged after this task, on branch spectre/kan-431-flow-batch-conductor-tool-calls-replace-idle -->
**Commit:** `docs(implement): batch the conductor's boundary calls and wait on report files`

  - [x] **Step 1: Turn discipline.** In `skills/flow/implement.md` **4. Execute (SDD + TDD)**,
    replace the paragraph opening `**Never end a turn with a child in flight** — wait for every
    implementer and reviewer launched` (line 424 on `main`) with:

````markdown verified:authored in-tree for this change; the harness facts are in design.md's "The wait"
**Never end a turn with a child in flight** — wait for every implementer and reviewer launched
before reporting a stage boundary or asking the operator anything.

**Turn discipline.** A turn is spent only where an output must be read before the next action is
chosen. Calls that do not depend on one another share one Bash call — every verdict printed, each
read afterwards — and launches that do not depend on one another share one message. A stage's
`flow stage begin` rides its first command and its `flow stage end` its last, in the same Bash
call. **A wait on a child is one foreground call, never a chain of idle calls:**

```bash verified:a foreground sleep and a bounded loop ran in this harness's Bash tool while planning this change
for i in $(seq 1 110); do test -s <report> && break; sleep 5; done
test -s <report> && echo ready || echo still-running
```

`<report>` is the file the child's REPORT FILE paragraph names — every child kind writes one as
its last act, after its commit and its final test run, so the file's presence is the child's
completion. The loop is bounded under the Bash tool's ten-minute cap; `still-running` re-issues
the wait, and a ceiling (**No forking, and a wall-clock ceiling on every slot**,
`skills/flow/review-panel.md`) is tracked across the calls. `skills/flow/review-panel.md` and
`skills/flow/verify-and-handoff.md` state their own batches under this paragraph and restate
none of it.
````

  - [x] **Step 2: The implementer's REPORT FILE paragraph.** Directly after the blockquote opening
    `> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary` that follows `Every
    implementer dispatch **must** also carry:` (lines 366–376 on `main`), add:

```markdown verified:shape copied from review-panel.md's REPORT FILE paragraph at lines 299-305 on main
> **REPORT FILE:** write your report to
> `<abs-worktree>/.superpowers/sdd/implementer-report-<k>.md` as your **last** act — after your
> commit and your final test run — carrying each commit's sha, the failing RED output you saw,
> and anything the plan's `unverified:` tags asked you to establish. The dispatcher waits on that
> file's presence; a resumed fix writes `implementer-report-<k>-fix-<n>.md` instead.
```

  - [x] **Step 3: The batched boundary.** Replace the four numbered items under `**Review
    overlaps the next implementer.**` (`1. **Bundle N+1's implementer commits**` through the end
    of `4. **Dispatch together:** …`, lines 382–404 on `main`) with the same four steps in their
    batched shape — the substance of each is unchanged, only which call it rides:

````markdown verified:authored in-tree for this change; the substance of each step is the text it replaces at implement.md:382-404 on main
1. **Bundle N+1's implementer commits** and writes its report; the wait above ends.
2. **A pending fix for bundle N folds in first.** If bundle N's review raised a fix, resume that
   bundle's implementer (`SendMessage`; record the resumption as its own pair under
   `task-<n>-implementer-fix-<k>`, `-agent-id` the implementer's own id) with the reviewer's
   report path; it commits `git commit --fixup=<task-sha>`, runs `git rebase --autosquash` and
   writes `implementer-report-<k>-fix-<n>.md`. A conflict there is between two of the branch's own
   commits and the implementer resolves it — `skills/flow/integrate.md`'s never-auto-resolve rule
   concerns the operator's base branch.
3. **One Bash call: the implementer's `record dispatch end`, the guard on every commit whose sha
   is new** — bundle N+1's tasks and every task the rebase rewrote — **and bundle N+2's gather.**
   The guard takes the canonical worktree's absolute path (the worktree created or resumed in
   step **2** above) as its fifth argument and this run's resolved `<name>` as its sixth:

   ```bash verified:copied from implement.md:392 on main
   check-task-commit-fields.sh <worktree> <task-id> <task-sha> <task-base> <canonical-worktree> <name>
   ```

   The guard reads git objects and `tasks.md` only, so it is safe while the tree changes. Its
   verdict is read before anything launches: a nonzero exit is a guard failure, not a review
   finding — it does **not** consume a fix-round slot; send the task back to the **same
   implementer**, then re-run the guard, before anything below.
4. **One message launches together:** bundle N's re-reviewer if step 2 ran (the task's full range,
   `git diff <task-base>..<new-task-sha>`), bundle N+1's reviewers, and bundle N+2's implementer.
   **The next Bash call records every launch's `begin`** — the very next action after the launches
   return, which is what "recorded immediately after the launch returns" above requires. The
   reviewer's `record end` and `flow tasks tick` share one call once its report file exists.
````

  - [x] **Step 4: The last bundle.** In the paragraph opening `**The last bundle's reviewers run
    alone.**` (line 420 on `main`), replace `Overlap them only with the review panel's pre-work —
    its citation check, bundle rebuild, relocation comparison and diff-size check;` with `Overlap
    them only with the review panel's pre-work, in its one call;` — the panel's own file lists what
    that call holds.
  - [x] **Step 5: Verify.** From the repository root run `scripts/check-vocabulary.sh`,
    `scripts/check-references.sh`, `scripts/check-dispatch-paragraphs.sh`,
    `scripts/check-stage-mark-calls.sh`, `scripts/check-normative-inventory.sh` (diff its output
    against a capture taken before step 1 — the only difference is the new paragraph's own
    sentences, none lost) and `scripts/check-contract-budget.sh` — each exits 0. Commit.

- [x] 2. Batch the review panel's calls in `skills/flow/review-panel.md`

**Build:** green
**Files:** `skills/flow/review-panel.md`, `scripts/check-contract-budget.sh`
**Tests:** none — documentation; the shipped guards in step 6 are the check
**Regression:** none — no test is added; `scripts/check-contract-budget.sh` and `scripts/check-dispatch-paragraphs.sh` are the mechanical checks and pass before and after
**Baseline:** before=0 after=0
<!-- measured: no test suite covers skills/ prose; the count is the number of tests this task touches @ 41688f0 -->
<!-- predicted: unchanged after this task, on branch spectre/kan-431-flow-batch-conductor-tool-calls-replace-idle -->
**Commit:** `docs(review-panel): batch the panel's pre-work, records, findings and reproducer runs`

  - [x] **Step 1: The pre-work call.** Directly before the paragraph opening `**Run the citation
    pre-check before rebuilding the dispatch context bundle below**` (line 67 on `main`), add:

```markdown verified:authored in-tree for this change; the batches are design.md's "The batches"
**Everything from the citation pre-check to the throwaway worktrees and the `[PRINCIPLES_PATH]`
and `[STANDARDS_PATHS]` resolution below is one Bash call**, under **Turn discipline**
(`skills/flow/implement.md`): every verdict printed, and only the dispatch depends on them — an
over-cap exit, a `REFUSE`, an absent principles file or an operator prompt is read off that call's
output and handled before any slot launches. The base-movement check above is the one exception:
its rebase branch changes what every later step reads, so it runs and is read first.
```

  - [x] **Step 2: Launches, records and the round's wait.** Directly after the `flow record
    dispatch begin … end` fence under `**Every slot's dispatch is recorded**` (lines 245–254 on
    `main`), before the paragraph opening `` `-slot` names the slot ``, add:

```markdown verified:authored in-tree for this change
Every slot of a round launches in one message; every `begin` is recorded in the next Bash call,
one call for all; the round's wait is one call whose condition is `test -s` on every launched
slot's report file; every `end` is recorded in one call once they all exist (**Turn
discipline**, `skills/flow/implement.md`). A slot whose report never appears within its ceiling
takes the breach path under **No forking, and a wall-clock ceiling on every slot** below.
```

  - [x] **Step 3: Findings and reproducers.** In the paragraph opening `**Every finding is a row
    in the store.**` (line 444 on `main`), replace `As each slot raises a finding, record it — one
    call, as it is raised:` with `Every finding a round raised is recorded in one Bash call, one
    `flow record finding` per finding:`. In the paragraph opening `**For each open finding whose
    record carries a runnable `finding-reproducer:` command**` (line 604 on `main`), replace
    `, run` and the fence that follows it with `, run — every finding's run and every throwaway
    worktree removal in one Bash call, each run followed by `; echo "F<n>: exit $?"` so every exit
    code stays readable:` and keep the `run-reproducer.sh` fence line itself unchanged.
  - [x] **Step 4: The per-finding sentence.** In the paragraph opening `**The parent records it,
    never the fix subagent**` (lines 632–635 on `main`), replace `**Record it per finding, as its
    verdict is reached, not batched at the round's end** — an aborted round still leaves every
    already-verified finding closed.` with `**Record every verdict this turn reached in one call,
    never deferred to the round's end** — the reproducer re-runs and the fix diff are read in one
    call, each finding judged, then every `status fixed` recorded together, so an aborted round
    still leaves every already-verified finding closed.`
  - [x] **Step 5: The fix subagent's REPORT FILE paragraph.** Directly after the FOREGROUND
    BUILDS blockquote under `**Every fix subagent's dispatch prompt also carries the FOREGROUND
    BUILDS paragraph**` (lines 710–714 on `main`), add:

```markdown verified:shape copied from review-panel.md's REPORT FILE paragraph at lines 299-305 on main
**Every fix subagent's dispatch prompt also carries the REPORT FILE paragraph**:

> **REPORT FILE:** write your report to
> `<abs-worktree>/.superpowers/sdd/panel-fix-report-<round>.md` as your **last** act — after
> the rebase and your final test run — naming each finding you addressed, the executable
> behaviours your fix changed, and the task commit each fixup folded into. The dispatcher waits
> on that file's presence.
```

  - [x] **Step 6: The budget row, then verify.** `skills/flow/review-panel.md` measured 45207 bytes
    against a 45213-byte row before this task.
    <!-- measured: wc -c skills/flow/review-panel.md; grep -n 'skills/flow/review-panel.md' scripts/check-contract-budget.sh @ 41688f0 -->
    After steps 1–5, run `wc -c skills/flow/review-panel.md` and set the `skills/flow/review-panel.md`
    row in `scripts/check-contract-budget.sh`'s `budgets()` table (line 196 on `main`) to that
    size plus 25%, rounded up — the guard's own documented rule for a genuine addition. Then from
    the repository root run `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
    `scripts/check-dispatch-paragraphs.sh`, `scripts/check-stage-mark-calls.sh`,
    `scripts/check-normative-inventory.sh` (diffed against a pre-edit capture: the one changed
    sentence is step 4's, and nothing is lost) and `scripts/check-contract-budget.sh` — each
    exits 0. Commit.

- [x] 3. Batch the verify and handoff tail in `skills/flow/verify-and-handoff.md`

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** none — documentation; the shipped guards in step 3 are the check
**Regression:** none — no test is added; `scripts/check-stage-mark-calls.sh` and `scripts/check-references.sh` are the mechanical checks and pass before and after
**Baseline:** before=0 after=0
<!-- measured: no test suite covers skills/ prose; the count is the number of tests this task touches @ 41688f0 -->
<!-- predicted: unchanged after this task, on branch spectre/kan-431-flow-batch-conductor-tool-calls-replace-idle -->
**Commit:** `docs(verify-and-handoff): batch the verifier's calls and the tail's stage marks`

  - [x] **Step 1: The verifier's report file and wait.** In the paragraph opening `**The relay
    contract.** The verifier has no operator channel` (line 79 on `main`), replace `Its turn ends
    with a single `## Report` block.` with `Its turn ends with a single `## Report` block, and its
    last act before that is writing the same block to
    `<abs-worktree>/.superpowers/sdd/verify-report-<key>.md` — `<key>` this dispatch's own key
    from **Recording** below — which is what the conductor waits on (**Turn discipline**,
    `skills/flow/implement.md`).` Then replace the sentence `The conductor shows the report as this
    stage's output.` (line 118 on `main`) with `The conductor shows the report as this stage's
    output. `prepare-workspace.sh`, both `project-get.sh` calls and this stage's `begin` mark are
    one Bash call; the wait, the verifier's `record dispatch end` and the ledger render below are
    one more.`
  - [x] **Step 2: The tail's marks.** Directly after the sentence ending `never a raw read of the
    state file's `worktrees` map.` under **Stage, excluding the planning paths** — the paragraph
    opening `Confirm every intended task checkbox is `[x]`` (line 261 on `main`) — add one
    sentence at the end of that paragraph: `From here to the handoff, each stage's `begin` mark
    rides its first command and its `end` mark its last (**Turn discipline**,
    `skills/flow/implement.md`); the `## Stage` returns are unchanged.`
  - [x] **Step 3: Verify.** From the repository root run `scripts/check-vocabulary.sh`,
    `scripts/check-references.sh`, `scripts/check-stage-mark-calls.sh`,
    `scripts/check-normative-inventory.sh` (diffed against a pre-edit capture: nothing lost) and
    `scripts/check-contract-budget.sh` — each exits 0. Commit.
