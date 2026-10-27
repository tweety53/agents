# kan-1039-end-to-end-autonomy-rule

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

**Goal:** one global rule — work end to end, stop only to plan or to test by hand — and the `/flow`
contracts aligned to it; `xhigh` agent definitions for later use, refused until wired.

**Spec:** `design.md` beside this file.

## Global constraints

- Work in `/Users/tweety53/Projects/agents-worktrees/kan-1039-end-to-end-autonomy-rule`; every
  `**Files:**` path is relative to its root. Never touch the main checkout.
- Every replacement below is literal: find the "Replace" block exactly once in the named file and
  substitute the "with" block, byte for byte — no rewording, no reflow. Every block was applied to a
  detached copy of `origin/main` at plan time and passed the prose guards; `verbatim-moves.txt`
  beside this file already lists every sentence the blocks add, reword or delete, so a sentence
  that differs from its block fails `scripts/check-verbatim-moves.sh`.
- `scripts/check-installed-rules.sh` skips a linked worktree; the reinstall that makes it see the
  new rule runs from the main checkout after landing (`CLAUDE.md`, **Reinstall after a change to
  this repository is fully done**).

- [x] 1. Add the global rule and its agent-baseline row
  - [x] **Step 1: Create `rules/work-end-to-end.mdc`** with exactly this content:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references, check-markdown-integrity and check-ask-silence exited 0
---
description: The goal is apps and tooling that work end to end — fix and re-run as often as it takes; interactive only for planning and for manual testing the user asked to stop for
alwaysApply: true
---

# Work end to end — stop only to plan or to test by hand

<!-- core -->
**The one thing that matters is apps and tooling that work end to end.** However many fix rounds,
re-runs, tooling updates or tooling fixes it takes, keep going: a red check, a failing guard, a
broken script or a stale environment is the next thing to fix, never a report or a question.
Tokens, time, round count and context length are never a reason to stop. Always on, every project
and every subagent.

**Only two parts are interactive: planning, and manual testing of the implementation with the fixes
it finds — the latter only when the user explicitly asked to stop after implementation.** Everywhere
else a question with a recommended option is never asked: take that option, record it, and continue
(**What still stops**, `skills/flow-contracts/operator-prompts-auto-resolution.md`). Still stop for
a choice no option can honestly be recommended for, an outward-facing or irreversible action another
rule requires confirmed, and a cause outside your reach — production, a credential only the user
holds, a repository you were not given. Stay inside your dispatch's remit: a read-only role reports,
and its dispatcher fixes.

<!-- /core -->

## How `/flow` applies it

The user's stop after implementation is the change record's `reviewBeforeLanding` on `/flow` and
`## handoff` `required` on `/flow-fast`; with neither set, nothing stops for testing.
**Fewest operator actions** (`skills/flow-contracts/pipeline.md`) is this rule's `/flow`
statement, and **What still stops** (`skills/flow-contracts/operator-prompts-auto-resolution.md`)
its list of exceptions.

## The incidents behind it

- kan-1037 (2026-10-27): the user spent about 380k uncached tokens answering mid-run questions with
  the option already marked recommended.
- kan-1038 (2026-10-27): asked whether to fix all eleven self-review findings or only the blocker,
  the user answered "recommended — don't ask such things again".
- KAN-1039 (2026-10-27), the user's own statement: the only thing they care about is well-working
  apps and tooling end to end, no matter how many fix rounds or tooling updates and fixes are
  needed; the only interactive parts are planning and manual testing of the implementation with the
  fixes that follow, and that only when they explicitly asked to stop after implementation.
```

  - [x] **Step 2: Add the row to `rules/agent-baseline.md`**, directly after the row that opens
    `| **Build the simplest thing that meets the requirement.**`, as its own table line:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references exited 0
| **Work end to end — stop only to plan or to test by hand.** Fix red checks, failing guards and broken tooling and keep going, however many rounds it takes; size is never a stop. Interactive only in planning and in manual testing the user explicitly asked to stop for — anywhere else take the recommended option, record it and continue. Stop only for a choice with no honest recommendation, an action another rule requires confirmed, or a cause out of your reach. A read-only role reports; its dispatcher fixes. | `~/.claude/rules/work-end-to-end.md` |
```

  - [x] **Step 3: Check the render** — the sandboxed install renders the new core:
    `H=$(mktemp -d) && HOME=$H ./setup.sh global >/dev/null 2>&1; grep -c 'rule: work-end-to-end.mdc' $H/.claude/CLAUDE.md`
    prints `1`, and `test -L $H/.claude/rules/work-end-to-end.md` succeeds.
    <!-- predicted: the Step 3 command after Step 1; the same command printed 1 for rule: be-brief.mdc on the worktree at origin/main 1be0e995 -->
  - [x] **Step 4: Verify:** `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `scripts/check-ask-silence.sh`, `scripts/check-markdown-integrity.sh` and
    `scripts/check-verbatim-moves.sh` exit 0.

**Files:** `rules/work-end-to-end.mdc`, `rules/agent-baseline.md`
**Tests:** none — prose; the guards in Step 4 and the sandboxed render in Step 3 check it.
**Regression:** none — no test is declared.
**Build:** green
**After:** none
**Commit:** `feat(rules): work end to end — stop only to plan or to test by hand`

**Decision:** new-rule-work-end-to-end

- [x] 2. Widen auto-resolution past planning and cure tooling failures in the run
  - [x] **Step 1: `skills/flow-contracts/operator-prompts-auto-resolution.md`** — four replacements.

    Replace:

```markdown verified:present once in skills/flow-contracts/operator-prompts-auto-resolution.md at origin/main 1be0e995
**During implementation and every fix run, a prompt with a recommended option is never asked: the
run takes the recommended option itself, at once.** The phases are everything from `flow.implement`
to the `IN_PROGRESS` handoff of a creating run — `skills/flow/implement.md`,
`skills/flow/review-panel.md`, `skills/flow/visual-verify.md`, `skills/flow/verify-and-handoff.md`
— and every fix run (`/flow <fix instructions>`, or a plain message run as a fix). The operator
asked for this once, for every flow run: a mid-run question they would answer with the recommended
option costs them a stop for nothing.
```

    with:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references exited 0
**Past planning, a prompt with a recommended option is never asked: the run takes the recommended
option itself, at once.** Past planning is everything from `flow.implement` to the run's final
handoff — `skills/flow/implement.md`, `skills/flow/review-panel.md`,
`skills/flow/visual-verify.md`, `skills/flow/verify-and-handoff.md`, `skills/flow/integrate.md`,
`skills/flow/cleanup.md` and the self-review pass (`skills/flow-self-review/SKILL.md`) — every fix
run (`/flow <fix instructions>`, or a plain message run as a fix), `/flow-fast` from its
implementation on, and the prompts that open a run. The operator asked for this for every flow run
(**Work end to end — stop only to plan or to test by hand**, `rules/work-end-to-end.mdc`): a
question they would answer with the recommended option costs them a stop for nothing.
```

    Replace:

```markdown verified:present once in skills/flow-contracts/operator-prompts-auto-resolution.md at origin/main 1be0e995
are auto-resolved exactly as the implementation phases are. The key absent, or a head that
resolves to nothing, leaves the scope exactly as the paragraph above states it. The remaining
**What still stops** bullets are the mode's limits: a prompt with no recommended option, nothing
to choose, anything outward-facing or irreversible, and every prompt outside the run's own phases
are asked under the mode exactly as off it.
```

    with:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references exited 0
are auto-resolved exactly as the phases past planning are. The key absent, or a head that
resolves to nothing, leaves the scope exactly as the paragraph above states it. The remaining
**What still stops** bullets are the mode's limits: the manual testing the operator asked for, a
prompt with no recommended option, a cause out of the run's reach and anything outward-facing or
irreversible stop under the mode exactly as off it.
```

    Replace the line `These are asked, or stop with `## Question`, exactly as their call sites state:`
    under `### What still stops` with:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references exited 0
These are asked, or stop with `## Question`, exactly as their call sites state — and nothing else
past planning is (**Work end to end — stop only to plan or to test by hand**,
`rules/work-end-to-end.mdc`):
```

    Insert this bullet directly before the line opening `- **No recommended option.**`:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references exited 0
- **Manual testing the operator asked for.** A `/flow` record whose `reviewBeforeLanding` is `true`
  stops at its `IN_PROGRESS` handoff (**Review-before-landing instruction**, `skills/flow/SKILL.md`),
  and `/flow-fast` with `## handoff` `required` stops after its change summary; the fixes that
  testing finds are the operator's fix runs. Neither stops unless the operator set it.
```

    Replace the two-line bullet opening `- **Nothing to choose.** A guard's exit 2 (it cannot answer)`
    (ending `every other `## Question` handback that carries no options.`) with:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references exited 0
- **Out of the run's reach.** A failure whose cure lies outside the run — production, a credential
  or account only the operator holds, a repository the run was not given — and a `## Question`
  handback that carries no options. A guard's exit 2, a failing command or tool, a stack that will
  not start and an environment fault the run's own commands can repair are not this bullet: they
  are cured in the run (**Fewest operator actions**, `skills/flow-contracts/pipeline.md`).
```

    Delete the last bullet of the file, the two lines opening
    `- **Outside implementation and fix runs.** The wrong-state override, the plain-message ambiguity`.
  - [x] **Step 2: `skills/flow-contracts/operator-prompts.md`** — in the `## Auto-resolution`
    load line, replace `only when a prompt arises during implementation or a fix run, or during planning when`
    with `only when a prompt arises past planning — implementation, review, verify, integrate, cleanup, the self-review pass or a fix run — or during planning when`.
  - [x] **Step 3: `skills/flow-contracts/pipeline.md`** — in `## Fewest operator actions`, after
    the paragraph that ends `as open items, each of which had a recommended fix).`, insert a blank
    line and:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references and check-markdown-integrity exited 0
**A failing tool, guard or environment is cured in the run, never handed off.** A guard's exit 2,
a script that errors, a stack that will not start, a stale environment and an out-of-date tool are
defects to fix, in the project's dev tooling or in `<agents repo>`
(**Pipeline defects found mid-run**, below), and the step runs again, however many rounds that
takes. Where a phase file says such a failure stops the run, asks, or ends its turn — and where a
prompt's recommended option is itself a stop — the stop holds only for a cause out of the run's
reach (**What still stops**, `skills/flow-contracts/operator-prompts-auto-resolution.md`);
otherwise this paragraph wins, as **Work end to end — stop only to plan or to test by hand**
(`rules/work-end-to-end.mdc`) requires.
```

  - [x] **Step 4: `skills/flow-contracts/project-configuration.md`** — in the `## decisions` row,
    replace `to the planning asks — never the pivot, an outward-facing or irreversible action, or an integrate or cleanup prompt (**Auto-resolution**`
    with `to the planning asks — never an outward-facing or irreversible action (**Auto-resolution**`.
  - [x] **Step 5: `skills/flow/SKILL-rationale.md`** — append at the end of the file, after a
    blank line:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references exited 0 once the rule was cited by its heading
### Work end to end — stops only for planning and asked-for testing (KAN-1039, 2026-10-27)

The operator, after kan-1037 and kan-1038 spent their stops on recommended picks: the only thing
that matters is apps and tooling that work end to end, however many fix rounds or tooling fixes it
takes, and the only interactive parts are planning and the manual testing they ask to stop for. The
rule is **Work end to end — stop only to plan or to test by hand** (`rules/work-end-to-end.mdc`);
**Fewest operator actions** (`skills/flow-contracts/pipeline.md`) and **What still stops** (`skills/flow-contracts/operator-prompts-auto-resolution.md`) carry it into
`/flow`.

- **Auto-resolution covers everything past planning** — integrate, cleanup, the self-review pass
  and the prompts that open a run, not only implementation and fix runs.
- **A tooling failure is cured in the run**, by one precedence paragraph in
  **Fewest operator actions**. **Rejected — rewording each call site:** about sixty sentences
  reworded for one rule.
- **A prompt whose recommended option was a stop now recommends going on** — the review panel's
  moved-base, bundle-failure and wall-clock prompts; a self-review product defect continues into
  its fix run; `/flow-fast` with no `## handoff` key lands. **Rejected — keeping the
  product-defect handoff as the fix run's planning boundary:** the fix run's own planning asks
  still reach the operator where the project asks them, so the handoff added a `/clear` and a
  retyped command and nothing else.
- **Kept as stops:** asks with no recommended option, outward-facing or irreversible actions, and
  causes out of the run's reach.
```

  - [x] **Step 6: Verify:** `scripts/check-references.sh`, `scripts/check-ask-silence.sh`,
    `scripts/check-markdown-integrity.sh` and `scripts/check-verbatim-moves.sh` exit 0.

**Files:** `skills/flow-contracts/operator-prompts-auto-resolution.md`, `skills/flow-contracts/operator-prompts.md`, `skills/flow-contracts/pipeline.md`, `skills/flow-contracts/project-configuration.md`, `skills/flow/SKILL-rationale.md`
**Tests:** none — prose; the guards in Step 6 check it.
**Regression:** none — no test is declared.
**Build:** green
**After:** Task 1
**Commit:** `docs(flow-contracts): resolve every stop past planning under the end-to-end rule`

**Decision:** auto-resolution-past-planning
**Decision:** tooling-failure-precedence
**Decision:** outward-facing-still-asked

- [x] 3. Make the review panel's stop-recommended prompts go on
  - [x] **Step 1: `skills/flow/review-panel.md`, the moved-base prompt.** Replace:

```markdown verified:present once in skills/flow/review-panel.md at origin/main 1be0e995
(`skills/flow-contracts/operator-prompts.md`), which that contract's **Auto-resolution** resolves on
its recommended **Stop**; **Rebase** and **Continue** run only on an explicit operator instruction —
silence is not one, and an unanswered ask takes **Stop** (**Unanswered mid-run asks**,
`skills/flow-contracts/pipeline.md`):

> **The base branch has moved and touches paths this change also touched — how should the
> panel proceed?**
> - **Stop — I'll rebase or reorder first** *(recommended)*
> - **Rebase onto `<base>` now, then continue**
> - **Continue — review as is**
```

    with:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references, check-ask-silence and check-dispatch-paragraphs exited 0
(`skills/flow-contracts/operator-prompts.md`), which that contract's **Auto-resolution** resolves on
its recommended **Continue**: integrate's sync onto the moved base settles the overlap before
anything lands (`skills/flow/sync-onto-base.md`). **Rebase** and **Stop** run only on an explicit
operator instruction, and an unanswered ask takes **Continue** (**Unanswered mid-run asks**,
`skills/flow-contracts/pipeline.md`).

> **The base branch has moved and touches paths this change also touched — how should the
> panel proceed?**
> - **Continue — review as is** *(recommended)*
> - **Rebase onto `<base>` now, then continue**
> - **Stop — I'll rebase or reorder first**
```

  - [x] **Step 2: the context-bundle failure.** Replace:

```markdown verified:present once in skills/flow/review-panel.md at origin/main 1be0e995
> - **Stop — resolve the build, then re-run the panel** *(recommended; silence defaults here)* —
>   closes `flow.review-panel` with `-outcome stopped`
```

    with (the five phrases `check-dispatch-paragraphs.sh` pins in this block stay):

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-dispatch-paragraphs exited 0
> - **Fix the build, then dispatch** *(recommended; silence defaults here)* — the run cures the
>   cause itself (**Fewest operator actions**, `skills/flow-contracts/pipeline.md`), re-runs the
>   gather and dispatches with the bundle; only a cause out of its reach closes
>   `flow.review-panel` with `-outcome stopped`
```

  - [x] **Step 3: the second wall-clock breach.** Replace:

```markdown verified:present once in skills/flow/review-panel.md at origin/main 1be0e995
(`skills/flow-contracts/operator-prompts.md`), resolved per that contract's **Auto-resolution** on
its recommended **Stop the run**:

> **Slot `<slot>` breached the wall-clock ceiling a second time. How should this proceed?**
> - **Re-dispatch it again**
> - **Proceed without the slot**
> - **Stop the run** *(default, recommended)* — ends at `IN_PROGRESS` with the implementation
>   committed on the branch
```

    with:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references and check-markdown-integrity exited 0
(`skills/flow-contracts/operator-prompts.md`), resolved per that contract's **Auto-resolution** on
its recommended **Re-dispatch it again**; a third breach of the same slot is asked, under that
contract's never-repeats clause.

> **Slot `<slot>` breached the wall-clock ceiling a second time. How should this proceed?**
> - **Re-dispatch it again** *(default, recommended)*
> - **Proceed without the slot**
> - **Stop the run** — ends at `IN_PROGRESS` with the implementation committed on the branch
```

  - [x] **Step 4: Verify:** `scripts/check-dispatch-paragraphs.sh`, `scripts/check-references.sh`,
    `scripts/check-ask-silence.sh`, `scripts/check-markdown-integrity.sh` and
    `scripts/check-verbatim-moves.sh` exit 0.

**Files:** `skills/flow/review-panel.md`
**Tests:** none — prose; the guards in Step 4 check it.
**Regression:** none — no test is declared.
**Build:** green
**After:** none
**Commit:** `docs(flow): review-panel prompts recommend going on, not stopping`

**Decision:** moved-base-continue
**Decision:** bundle-failure-fix-then-dispatch
**Decision:** wall-clock-redispatch

- [x] 4. Continue a self-review product defect into its fix run
  - [x] **Step 1: `skills/flow-contracts/known-bugs.md`** — three replacements. Replace:

```markdown verified:present once in skills/flow-contracts/known-bugs.md at origin/main 1be0e995
fixed. The `flow.self-review` mark closes `-outcome stopped`, nothing lands, the change keeps its
state, and the run ends on the block below.
```

    with:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references exited 0
fixed. The `flow.self-review` mark closes `-outcome stopped`, nothing lands, the change keeps its
state, and the run continues straight into the fix run that defect names, in the same invocation —
never a handoff for the operator to start it (KAN-1039).
```

    Replace:

```markdown verified:present once in skills/flow-contracts/known-bugs.md at origin/main 1be0e995
- **`/flow`'s integrate run 1** stops before step 5's route. The fix run writes its plan above the
  archive commit, so the next bare `/flow <name>` is an archived re-run with new work, which runs
  the pass (**Run 1 — the branch is not merged**, `skills/flow-contracts/finish-contract-run1.md`).
- **`/flow-fast`'s verify** stops before `flow.stage-diff`; the fix re-run lands normally.
```

    with:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references and check-markdown-integrity exited 0
- **`/flow`'s integrate run 1** stops before step 5's route and continues as
  `/flow <name> <the defect, one line>` would: the fix run writes its plan above the archive
  commit and, once it writes its state, continues into integrate as an archived re-run with new
  work, which runs the pass again (**Run 1 — the branch is not merged**,
  `skills/flow-contracts/finish-contract-run1.md`).
- **`/flow-fast`'s verify** stops before `flow.stage-diff` and continues as
  `/flow-fast <name> <the defect, one line>` would; the fix re-run lands normally.
```

    Delete everything after the line `operator to start that fix run.` — the blank line, the
    fenced block opening `## Landing stopped — a product defect to fix first`, and the line
    `` `/flow-fast` prints `/flow-fast <name> <the defect, one line>` as the last line instead. ``
    — so `operator to start that fix run.` ends the file.

  - [x] **Step 2: `skills/flow/integrate.md`** — replace:

```markdown verified:present once in skills/flow/integrate.md at origin/main 1be0e995
`-outcome stopped` instead, and the run stops before step 5 on the handoff block **the sweep**
(`skills/flow-contracts/known-bugs.md`) defines:
```

    with:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references exited 0
`-outcome stopped` instead, and the run continues straight into the fix run **the sweep**
(`skills/flow-contracts/known-bugs.md`) names, in the same invocation, before step 5:
```

  - [x] **Step 3: `skills/flow-fast/SKILL.md`** — replace:

```markdown verified:present once in skills/flow-fast/SKILL.md at origin/main 1be0e995
`-outcome stopped` instead, and ends the run on the handoff block **the sweep**
(`skills/flow-contracts/known-bugs.md`) defines.
```

    with:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references exited 0
`-outcome stopped` instead, and continues straight into the fix re-run **the sweep**
(`skills/flow-contracts/known-bugs.md`) names, in the same invocation.
```

  - [x] **Step 4: Verify:** `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh`
    and `scripts/check-verbatim-moves.sh` exit 0, and
    `git grep -n -F -e 'Landing stopped' -e 'the run ends on the block below' -- skills` prints
    nothing.

**Files:** `skills/flow-contracts/known-bugs.md`, `skills/flow/integrate.md`, `skills/flow-fast/SKILL.md`
**Tests:** none — prose; the guards and the absence sweep in Step 4 check it.
**Regression:** none — no test is declared.
**Build:** green
**After:** none
**Commit:** `docs(flow-contracts): a self-review product defect continues into its fix run`

**Decision:** product-defect-continues

- [x] 5. Land `/flow-fast` without a `## handoff` key
  - [x] **Step 1: `skills/flow-fast/SKILL.md`** — replace:

```markdown verified:present once in skills/flow-fast/SKILL.md at origin/main 1be0e995
`skills/flow-contracts/project-configuration.md`): `none` continues to section 6 in this same
invocation; `required`, or the key absent, ends the run here — worktree and branch kept, the
summary naming the branch and `Re-run /flow-fast <name> to land it` — so the operator reviews
the branch before anything is pushed. The remaining marks below are then made by that bare
re-run, not by this one.
```

    with:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references exited 0
`skills/flow-contracts/project-configuration.md`): `none`, or the key absent, continues to
section 6 in this same invocation; `required` ends the run here — worktree and branch kept, the
summary naming the branch and `Re-run /flow-fast <name> to land it` — so the operator, who asked
to test by hand, reviews the branch before anything is pushed. The remaining marks below are then
made by that bare re-run, not by this one.
```

    and, in the landing-route paragraph, replace
    `recommended; with `## handoff` `none`, take `pull request` without asking.` with
    `recommended; otherwise take `pull request` without asking.`
  - [x] **Step 2: `commands-claude/flow-fast.md`** — replace
    `then — unless the project's `## handoff` is `none` — stop for the` with
    `then — when the project's `## handoff` is `required` — stop for the`.
  - [x] **Step 3: `skills/flow-contracts/project-configuration.md`** — in the `## handoff` row,
    replace
    `` `none` lands the change in the same invocation, straight after the change summary; `required`, or absent, stops after the summary so the operator reviews the branch first and re-runs `/flow-fast <name>` bare to land it. `/flow` always hands off at `IN_PROGRESS` and never reads this key. ``
    with
    `` `none`, or absent, lands the change in the same invocation, straight after the change summary; `required` stops after the summary so the operator, who asked to test by hand, reviews the branch first and re-runs `/flow-fast <name>` bare to land it. `/flow` never reads this key: its stop is the change record's `reviewBeforeLanding`. ``
  - [x] **Step 4: Verify:** `scripts/check-references.sh`, `scripts/check-installed-citations.sh`,
    `scripts/check-ask-silence.sh`, `scripts/check-markdown-integrity.sh` and
    `scripts/check-verbatim-moves.sh` exit 0.

**Files:** `skills/flow-fast/SKILL.md`, `commands-claude/flow-fast.md`, `skills/flow-contracts/project-configuration.md`
**Tests:** none — prose; the guards in Step 4 check it.
**Regression:** none — no test is declared.
**Build:** green
**After:** Task 2, 4
**Commit:** `docs(flow-fast): land without a handoff key; only required stops`

**Decision:** flow-fast-handoff-absent-lands

- [x] 6. Point the unfinished-work stop at finishing the work
  - [x] **Step 1: `skills/flow/integrate.md`** — replace:

```markdown verified:present once in skills/flow/integrate.md at origin/main 1be0e995
- **`STOP: unfinished-work`** → **Load `skills/flow/unfinished-work-gate.md`** — the
  breakdown, the three-course prompt and what each course does are there. **Continue** and
  **File or join…** re-run `prepare` with `--accept-outstanding`; **Stop** leaves the mark the
  script closed `stopped`.
```

    with:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references exited 0
- **`STOP: unfinished-work`** → **Load `skills/flow/unfinished-work-gate.md`** — the run
  finishes the unfinished work itself, asks nothing, and re-runs `prepare`, per its
  **Finish it — never a prompt**.
```

  - [x] **Step 2: `skills/flow-contracts/finish-contract-run1.md`** — replace the table row
    `` | `OUTSTANDING` | show the breakdown and offer the three courses of **The three courses** (`skills/flow/unfinished-work-gate.md`) | ``
    with
    `` | `OUTSTANDING` | finish the outstanding work and re-run the guard, per **Finish it — never a prompt** (`skills/flow/unfinished-work-gate.md`) | ``.
  - [x] **Step 3: `stats/internal/guard/integratechange_prepare.go`** — replace the stop message:

```go verified:present once in stats/internal/guard/integratechange_prepare.go at origin/main 1be0e995
			c.stop("unfinished-work", "load skills/flow/unfinished-work-gate.md and offer its three courses; Continue or File-or-join re-runs prepare with --accept-outstanding")
```

    with:

```go verified:applied to a detached origin/main 1be0e995 copy at plan time; gofmt left it unchanged
			c.stop("unfinished-work", "load skills/flow/unfinished-work-gate.md: finish the unfinished work, asking nothing, then re-run prepare")
```

  - [x] **Step 4: Verify:** `cd stats && gofmt -l .` prints nothing;
    `cd stats && make -s web-dist && go vet ./...` and
    `cd stats && go test ./internal/guard -run 'TestIntegrateChangePrepare' -count=1` exit 0;
    `scripts/check-references.sh` and `scripts/check-verbatim-moves.sh` exit 0; and
    `git grep -n -F -e 'offer its three courses' -e 'the three-course prompt' -e 'offer the three courses' -- skills stats` prints nothing.

**Files:** `skills/flow/integrate.md`, `skills/flow-contracts/finish-contract-run1.md`, `stats/internal/guard/integratechange_prepare.go`
**Tests:** none — the stop's message text is pinned by no test; `TestIntegrateChangePrepare` keeps checking its `STOP: unfinished-work` kind.
**Regression:** none — no test is declared.
**Build:** green
**After:** Task 4
**Commit:** `fix(flow): point the unfinished-work stop at finishing the work`

**Decision:** stale-three-courses

- [x] 7. Drop the self-review's rating-only and xhigh asks
  - [x] **Step 1: `skills/flow-self-review/SKILL.md`** — replace:

```markdown verified:present once in skills/flow-self-review/SKILL.md at origin/main 1be0e995
pair's `flow-planner-medium` or `flow-planner-low`. Where `high` still cannot produce a
design that holds, the next step is effort `xhigh` — and an `xhigh` dispatch is never made
without the operator's approval, asked first with the finding, the objection and what `high`
returned; a refusal leaves the finding offered for filing in step 4. **This session plans,
```

    with:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-references exited 0
pair's `flow-planner-medium` or `flow-planner-low`. Where `high` still cannot produce a
design that holds, the next re-plan changes approach again under **Fewest operator actions**
(`skills/flow-contracts/pipeline.md`) — never a stop and never an ask; `flow-planner-xhigh` is
defined for later and no step dispatches it yet. **This session plans,
```

    and replace:

```markdown verified:present once in skills/flow-self-review/SKILL.md at origin/main 1be0e995
With nothing to offer, the call
carries the rating alone.
```

    with:

```markdown verified:applied to a detached origin/main 1be0e995 copy at plan time; check-ask-silence exited 0
With nothing to offer, no call
is made: the rating is asked only beside a filing or mining ask.
```

  - [x] **Step 2: Verify:** `scripts/check-references.sh`, `scripts/check-ask-silence.sh`,
    `scripts/check-markdown-integrity.sh` and `scripts/check-verbatim-moves.sh` exit 0.

**Files:** `skills/flow-self-review/SKILL.md`
**Tests:** none — prose; the guards in Step 2 check it.
**Regression:** none — no test is declared.
**Build:** green
**After:** none
**Commit:** `docs(flow-self-review): no rating-only ask and no xhigh approval ask`

**Decision:** rating-only-beside-an-ask
**Decision:** xhigh-replan-ask-struck

- [x] 8. Define the xhigh agents and refuse dispatches on them
  - [x] **Step 1: RED first.** In `stats/internal/repotest/enforce_agent_baseline_hook_test.go`,
    in `TestEnforceAgentBaselineHook`'s case table, add after the row named
    `2e pipeline-fix loop escalation on flow-high`:

```go verified:the row shape copies the 2b row of enforce_agent_baseline_hook_test.go at origin/main 1be0e995
		{"2f hand-picked flow-xhigh", hook, payload(baseline+" implement task 3", "flow-xhigh", "opus", ""), []string{"defined for later use"}},
		{"2g hand-picked flow-planner-xhigh", hook, payload(baseline+" re-plan", "flow-planner-xhigh", "opus", ""), []string{"defined for later use"}},
```

    In `stats/internal/setuptest/agents_test.go`, add `"xhigh"` to both effort lists in
    `TestSetupAgents` (`[]string{"low", "medium", "high", "xhigh"}`), change the count assertion
    to `g.assertEq("eight flow-* symlinks present under "+agentsDir, 8, len(found))`, and replace
    <!-- predicted: cd stats && go test ./internal/setuptest -run 'TestSetupAgents$' after Step 2 -->
    the doc comment's first three lines with:

```go verified:authored for this change; replaces the stale "three" wording at agents_test.go:9-11
// TestSetupAgents asserts setup.sh installs the flow-<effort> and flow-planner-<effort> agent
// definitions on `./setup.sh global`, run against a sandboxed HOME so it never touches the real
// ~/.claude. Every symlink must exist under $HOME/.claude/agents/ and resolve; each flow-<effort>
```

    Run `cd stats && go test ./internal/repotest -run 'TestEnforceAgentBaselineHook$' -count=1`
    and `cd stats && go test ./internal/setuptest -run 'TestSetupAgents$' -count=1`: both fail —
    rows 2f/2g are allowed, and `flow-xhigh.md` is missing.
  - [x] **Step 2: Create the definitions.** `agents/flow-xhigh.md`:

```markdown verified:shape of 43125acf^:agents/flow-xhigh.md, which carried effort: xhigh before 2026-09-30
---
name: flow-xhigh
description: Flow role at xhigh effort — defined for later use; no contract dispatches it yet and the agent-baseline hook refuses a dispatch on it; pass `model` at dispatch time.
effort: xhigh
tools: Read, Glob, Grep, Bash, Write, Edit, ToolSearch
---
General-purpose flow role at xhigh effort; the dispatch prompt carries every instruction.
```

    `agents/flow-planner-xhigh.md`:

```markdown verified:shape of agents/flow-planner-high.md at origin/main 1be0e995
---
name: flow-planner-xhigh
description: Flow planner at xhigh effort — defined for later use; no contract dispatches it yet and the agent-baseline hook refuses a dispatch on it; it may dispatch fresh Explore and Plan agents; pass `model` at dispatch time.
effort: xhigh
tools: Read, Glob, Grep, Bash, Write, Edit, ToolSearch, Agent
---
Flow planner at xhigh effort; it may dispatch fresh Explore and Plan agents, never a fork; the dispatch prompt carries every instruction.
```

    In `stats/internal/guard/installedcitations.go`'s `cicExpectedZero`, after the
    `agents/flow-planner-high.md` row, add:

```go verified:the row shape copies installedcitations.go:113 at origin/main 1be0e995
	{"agents/flow-xhigh.md", "generic dispatch-target agent definition — cites no .md/.mdc path at all"},
	{"agents/flow-planner-xhigh.md", "generic dispatch-target agent definition — cites no .md/.mdc path at all"},
```

  - [x] **Step 3: The hook.** In `hooks/enforce-agent-baseline.py`, after the `FLOW_HIGH_MARKERS`
    tuple, add:

```python verified:mirrors flow_high_reason in hooks/enforce-agent-baseline.py at origin/main 1be0e995
# flow-xhigh and flow-planner-xhigh are defined for later use (KAN-1039): no contract dispatches
# them yet, so a dispatch on either was hand-picked.
FLOW_XHIGH_TYPES = ("flow-xhigh", "flow-planner-xhigh")
```

    after `flow_high_reason`, add:

```python verified:mirrors flow_high_reason in hooks/enforce-agent-baseline.py at origin/main 1be0e995
def flow_xhigh_reason(tool_input: dict) -> "str | None":
    subagent_type = tool_input.get("subagent_type")
    if subagent_type not in FLOW_XHIGH_TYPES or "/.zcode/" in __file__:
        return None
    return (
        f"Dispatch blocked: {subagent_type} is defined for later use and no contract "
        "dispatches it yet (KAN-1039). Dispatch on the pair the recorded decision or the "
        "loop's pairs table names."
    )
```

    and in `main()`, add `or flow_xhigh_reason(tool_input)` on its own line directly after
    `or flow_high_reason(prompt, tool_input)`.
  - [x] **Step 4: GREEN.** Re-run both tests from Step 1: they pass.
  - [x] **Step 5: Verify:** `cd stats && gofmt -l .` prints nothing;
    `cd stats && make -s web-dist && go vet ./...` exits 0; `scripts/check-installed-citations.sh`,
    `scripts/check-references.sh` and `scripts/check-python-suppressions.sh` exit 0.

**Files:** `agents/flow-xhigh.md`, `agents/flow-planner-xhigh.md`, `stats/internal/guard/installedcitations.go`, `stats/internal/setuptest/agents_test.go`, `hooks/enforce-agent-baseline.py`, `stats/internal/repotest/enforce_agent_baseline_hook_test.go`
**Tests:** `2f hand-picked flow-xhigh`, `2g hand-picked flow-planner-xhigh`
**Regression:** reverting the hook lets rows 2f and 2g through, failing `TestEnforceAgentBaselineHook`; reverting the definitions fails `TestSetupAgents` on the missing `flow-xhigh.md` and the symlink count.
**Build:** green
**After:** none
**Commit:** `feat(agents): define xhigh flow agents for later use, refused until wired`

**Decision:** xhigh-agents-defined-refused
