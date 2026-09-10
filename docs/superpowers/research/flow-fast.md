# flow-fast — research notes

Source: KAN-490

Brainstormed on 2026-09-10 against `skills/flow/`, `skills/flow-contracts/` and the prior notes
`flow-remove-resumed-subagents.md`, `flow-speedup.md` and `flow-gymie-implementation-speedup.md`.
The operator's opening list, verbatim: "1. minimum context overhead 2. brainstorm stage -> do it
but auto pick recommended 3. full flow go app interaction as usual flow run 4. full inline except
review, review bundle always the same - primary + simple reviewer bundled 5. less guards 6.
targeted test/lint reruns, full ignored or before the handoff if it is REALLY needed 7. ALL minor
findings are ignored. Criticals and majors always fixed. 8. Faster cleanup when finishing." One
shape question and eight design questions were put to the operator; every answer below is the
operator's decision, not a default left for a later reader.

Two facts about the current `/flow` bounded the design. Minor findings already defer by default
(commit 722341a), `execution: inline` and a dynamic panel roster already exist as project toggles,
and the docs-only reduction already shrinks the panel — so several list items are partly present.
And `/myflow-fast` existed before (KAN-111) as a composite chaining command with the full rigor of
each stage; it was folded into `/flow`. This note's `/flow-fast` is a different thing: the same
pipeline with less ceremony, not the same ceremony in one command.

## 1. Shape — thin router plus slim phase files

The `/flow` phase files total about 10k lines; loading them is the context cost, not the guards.
That pulls against the repository's non-repetition rule. Three shapes were weighed:

- **Thin router + slim phase files** — chosen. Own `skills/flow-fast/SKILL.md` and short phase
  files. Shares `skills/flow-contracts/*` as canonical, the reviewer prompts under `skills/flow/`
  and `skills/flow/scripts/`. Restates nothing from `skills/flow/*.md`; each fast file states the
  procedure it runs and names the contract it follows.
- **Profile inside `/flow`** — rejected: no duplication, but every run still loads the full phase
  files, so the context goal is lost.
- **Fully standalone skill** — rejected: copies and trims the contracts too; smallest context,
  largest drift, and it would need an exemption from non-repetition.

Files: `skills/flow-fast/SKILL.md`, `brainstorm.md`, `implement.md`, `review.md`, `finish.md`.
Edits outside the new skill: `skills/flow-contracts/finish-contract-run2.md` gains one clause
allowing a `/flow-fast` run to skip self-review and verify-cleanup; `CLAUDE.md`'s skill index
gains a row; and the `/rename` / `/color` removal in section 10.

## 2. Router

Same announce line shape ("Using flow-fast for change `<name>`."), loads
`skills/flow-contracts/pipeline.md` first, same three states and transition table, same task-list
registration, one session token per run. Every stage mark carries `-command '/flow-fast'`. Model
resolution is a single `flow settings get` for `DEFAULT_MODEL`, with the same unreachable-store
fallback `/flow` uses. No planning model, self-review model or verify model is resolved; no project
toggle (`## execution mode`, `## implementer model`, `## review panel`) is read. No flags; the
argument rules are `/flow`'s.

## 3. State interop

One change can move between `/flow` and `/flow-fast` mid-life (start fast, land with full, or the
reverse). So fast writes the same state record and the same `flow.*` stage keys; it simply marks
fewer of them. `/flow-status` shows both. The recorded decision is fixed (execution `inline`,
roster `primary` + `simple-reviewer`) so a `/flow` run resuming the change reads a valid record.

Rejected: a `mode: fast` field that makes `/flow` refuse the change, and a separate record shape
invisible to `/flow-status`.

## 4. Brainstorm — inline, auto-pick, no design gate

Jira resolution and the In Progress transition follow `skills/flow-contracts/jira-integration.md`
exactly. Planning runs in the parent session; there is no planner dispatch (consistent with
`flow-remove-resumed-subagents.md`, which moves `/flow`'s own brainstorming inline). Every options
round picks the recommended option. The parent asks the operator only when no option is
recommended or the request is unusable without an answer. Writes `proposal.md`, `design.md` and
`tasks.md` in `/flow`'s shape, so `check-unfinished-work.sh` and a later `/flow` resume both read
them. Marks `flow.kickoff`, `flow.brainstorm`, `flow.create-artifacts`, `flow.writing-plans`,
`flow.decide`; skips `flow.design-approval` entirely. The IN_PROGRESS staged-diff review is the
only human gate before landing.

Rejected: keeping one design gate (a short summary and a yes before implementing).

## 5. Implement — inline, TDD, targeted runs

Workspace isolation via `prepare-workspace.sh` as today. The parent implements task by task from
`tasks.md`, failing test first, one commit per task carrying the same commit fields as `/flow`
(the `check-task-commit-fields.sh` guard is dropped, the format is kept). After each task: the
touched package's tests and lint on touched files only. A full test or lint run happens only when
the operator's instruction text asks for it — never automatically, not even before handoff. SDD
ceremony (spec deltas, plan-provenance guard) is dropped.

Rejected: a full run before handoff when the diff touches shared surface (build config, shared
libs, a module without targeted tests); one full run always at handoff; dropping TDD; dropping
`tasks.md` (which would break `check-unfinished-work.sh`).

## 6. Review — fixed bundle, Critical/Major only

The bundle is always `primary` on `DEFAULT_MODEL` and `simple-reviewer` on `haiku`, briefed with
`skills/flow/principles-reviewer-prompt.md`'s sibling prompts as `/flow` briefs them
(`primary` against `proposal.md`, `design.md` and `tasks.md`; `simple-reviewer` by
`skills/flow/simple-reviewer-prompt.md`). The store's reviewer list and the docs-only reduction
are ignored: the roster never changes. Critical and Major findings are fixed inline by the parent.
Every Minor is deferred and recorded, never fixed — no per-Minor "trivially easy" judgment. After a
fix, both slots re-run on the delta until no Critical or Major is open.
`check-panel-findings-closed.sh` gates the handoff. The panel record keeps `/flow`'s shape.

Rejected: re-running `primary` alone on the delta; one round with no re-run; `simple-reviewer` on
`DEFAULT_MODEL`; honouring the project toggles.

## 7. Handoff and fix runs

No visual verification. Stage the diff, print run instructions and the IN_PROGRESS block per
`skills/flow-contracts/handoff-blocks.md`, write IN_PROGRESS. A fix run at IN_PROGRESS fixes
inline, runs targeted tests, re-runs the panel on the delta and hands off the same way; a fix never
moves the state.

## 8. Finish — the contracts, minus two steps

A bare IN_PROGRESS run follows `skills/flow-contracts/finish-contract-run1.md` and
`finish-contract-run2.md` as `/flow` does: preflight, unfinished-work gate, landing question or
`## default landing route`, session preservation, the two commits, the landing routes, then verify
merge, sync and commit the archive, cleanup, write FINISHED, push. Run 2 skips the self-review
subagent and the verify-cleanup pass.

Rejected: dropping self-review only; also dropping session preservation.

## 9. Guards

Presence-checked and run: `check-unfinished-work.sh`, `check-base-moved.sh`,
`check-finish-preflight.sh`, `check-cleanup-complete.sh`, `check-workspace-isolation.sh`,
`check-worktree-processes.sh`, `check-panel-findings-closed.sh`. Every other guard in
`skills/flow/scripts/` is dropped, not hand-run: panel citation-trigger, diff-size, docs-only,
reproducers, fix-single-dispatch, plan-shape, spec-reach, visual-trigger, visual-verification,
task-commit-fields.

Rejected: state and git safety only (without `check-panel-findings-closed.sh`), which would let a
Critical slip past unreviewed; keeping the full set.

## 10. Drop `/rename` and `/color` everywhere

The two operator-paste lines `/rename <change-name>` and `/color cyan` that `/flow` prints after
its announce line are deprecated. Remove them from `skills/flow/SKILL.md`, from **Handoff output**
in `skills/flow-contracts/pipeline.md`, and the paragraph in `pipeline-rationale.md` that records
why they were printed rather than invoked. `/flow-fast` never prints them. The archived KAN-26
design and artifact under `docs/superpowers/` are history and stay as they are.

## 11. Verifying the change

`scripts/check-references.sh` and `scripts/check-vocabulary.sh` (and every other
`scripts/check-*.sh` guard named in `.flow/project.md`'s `## lint`) pass, and one docs-only change
runs through `/flow-fast` end to end in this repository — the same cutover check commit e83c9f4
ran for `/myflow-fast`.

## 12. Raise the dynamic plan-class thresholds

The class `plan-class.sh` assigns drives every dynamic decision: small and regular run inline,
big runs sdd via the conductor on opus/high with the six-slot roster and a full rerun. The
operator's read: the current thresholds send too much work to big and to the full roster. Raise
them so far more changes land inline and compact. The experimental roll and the bundle roll are
unchanged.

| Value | Now | New |
|---|---|---|
| small | tasks ≤ 5 and files ≤ 12 and repos = 1, no migration, no spec | tasks ≤ 10 and files ≤ 25, other conditions unchanged |
| big | tasks ≥ 15 or files ≥ 40 or (repos > 1 and tasks ≥ 8) or (migration and tasks ≥ 8) | tasks ≥ 30 or files ≥ 80 or (repos > 1 and tasks ≥ 15) or (migration and tasks ≥ 15) |
| regular | everything else | everything else |
| compact roster roll | < 70 on small, < 30 on regular and big | < 90 on small, < 60 on regular and big |
| experimental slot roll | < 30 on every class | unchanged |
| static grouping roll | bundle < 30 | unchanged |

Touches `skills/flow/scripts/plan-class.sh` and its header comment, `test-plan-class.sh`, the
class and roll paragraphs of `skills/flow/brainstorm-planner.md`, and the "Inputs and the class"
and "The rolls" sections of the archived kan-472 design those files cite. In scope of KAN-490 by
the operator's decision, though it is independent of the fast skill itself.

## Step-by-step breakdown

### Router (`skills/flow-fast/SKILL.md`)

**What:** Resolves state, registers tasks, resolves `DEFAULT_MODEL`, dispatches into the fast
phase file for the state in force, marks with `-command '/flow-fast'`.
**Why:** Same state machine as `/flow` so the two commands are interchangeable on one change, with
one model call instead of `/flow`'s resolution block.
**Uses:** `skills/flow-contracts/pipeline.md`, `flow settings get`, `flow stage begin/end`.

### Brainstorm (`skills/flow-fast/brainstorm.md`)

**What:** Inline planning that auto-picks the recommended option, writes `proposal.md`,
`design.md`, `tasks.md`, records the fixed decision, no design gate.
**Why:** The brainstorm's value is the plan, not the option rounds; the operator reviews the
staged diff instead.
**Uses:** `skills/flow-contracts/jira-integration.md`, `spectre new`, `flow record decision`.

### Implement (`skills/flow-fast/implement.md`)

**What:** Inline TDD per task, one commit per task, targeted tests and lint on touched packages.
**Why:** No conductor or implementer dispatch to resume; full runs are the slow part and are opt-in.
**Uses:** `prepare-workspace.sh`, `check-workspace-isolation.sh`, the project's `## test` and
`## lint` commands.

### Review (`skills/flow-fast/review.md`)

**What:** Fixed two-slot panel, Critical/Major fixed inline, Minors deferred, both slots re-run on
the delta until closed.
**Why:** A constant roster removes the roster resolution and the docs-only branch; deferring every
Minor removes the per-finding judgment.
**Uses:** `skills/flow/simple-reviewer-prompt.md`, the primary briefing from
`skills/flow/review-panel.md`, `check-panel-findings-closed.sh`, `flow record panel`.

### Handoff (in `skills/flow-fast/review.md` or `implement.md`)

**What:** Stage the diff, print run instructions and the IN_PROGRESS block, write IN_PROGRESS.
**Why:** The one human gate; the block shape is shared so `/flow-status` renders it.
**Uses:** `skills/flow-contracts/handoff-blocks.md`, `skills/flow-contracts/state-file.md`.

### Finish (`skills/flow-fast/finish.md`)

**What:** Runs the two finish contracts, skipping self-review and verify-cleanup in run 2.
**Why:** Landing and archive are git and record work that cannot be cut without losing the record;
the two skipped steps are the slow, optional tail.
**Uses:** `skills/flow-contracts/finish-contract-run1.md`, `finish-contract-run2.md`,
`check-finish-preflight.sh`, `check-unfinished-work.sh`, `check-base-moved.sh`,
`check-cleanup-complete.sh`, `check-worktree-processes.sh`.

### Guard set

**What:** Seven guards presence-checked and run; the rest dropped.
**Why:** State and git safety plus the one panel guard that keeps an unfixed Critical out of a
handoff; the dropped guards check ceremony fast does not perform.
**Uses:** `skills/flow/scripts/`, `scripts/lib/change-plan.sh`.

### `/rename` and `/color` removal

**What:** Delete the two printed operator lines and their rationale from `/flow` and the pipeline
contract.
**Why:** Deprecated; `/flow-fast` must not inherit them.
**Uses:** `skills/flow/SKILL.md`, `skills/flow-contracts/pipeline.md`,
`skills/flow-contracts/pipeline-rationale.md`.

### Plan-class thresholds

**What:** Raise the small and big class limits and the compact roll cut-offs per section 12.
**Why:** More changes run inline with the compact roster; the conductor and full roster are the
cost the operator wants to reach less often.
**Uses:** `skills/flow/scripts/plan-class.sh`, `skills/flow/scripts/test-plan-class.sh`,
`skills/flow/brainstorm-planner.md`.
