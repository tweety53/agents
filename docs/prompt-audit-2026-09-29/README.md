# Prompt audit, 2026-09-29: slim the /flow prompts without changing behaviour

An audit of every Markdown file a `/flow` session loads. It measures each session's load and plans
how to cut it through verbatim moves and duplicate deletions only. The work is tracked in
**KAN-851** (Epic).

- **Base.** The audit ran at `700e184`. Main has since retired the bugbot and security slots
  (`ebdfdde1`..`4dafab4a`). Line numbers in the audit files refer to `700e184`, so re-locate each
  row by its quoted first words.
- **Workflow assumed.** The operator runs `/clear` between planning, implementation and finish, so
  each phase is a fresh session. Each byte it loads is paid again on every later turn, as cache
  reads.
- **Units.** Bytes are `wc -c`. Tokens are about bytes/4 (the KAN-378 convention); real counts run
  somewhat higher.
- **Confidence.** Savings count high and medium confidence together. High confidence alone is lower:
  the finish files' high-only floor is 12%, against 39% for high plus medium.

## Issues

| Issue | What | Audit files | Order |
|---|---|---|---|
| KAN-852 | Safety net: the `check-verbatim-moves` guard, split-citation checks, and per-session load reports and measurement | [`tools.md`](tools.md) | first |
| KAN-853 | Reconcile rule copies that contradict each other | every file's `## Drift / bugs found` | first |
| KAN-854 | Defects that need a behaviour decision | every file's `## Drift / bugs found` | in parallel |
| KAN-855 | The every-session layer: router, `pipeline.md`, the command body, and the `state-file.md` / `project-configuration.md` run-time cores | [`audit-every-session.md`](audit-every-session.md) | after 852, 853 |
| KAN-856 | The planning session; stop implementation sessions loading `brainstorm.md` | [`audit-planning.md`](audit-planning.md) | after 852, 853 |
| KAN-857 | `implement.md` and `review-panel.md` lazy splits | [`audit-implement.md`](audit-implement.md), [`audit-review-panel.md`](audit-review-panel.md) | after 852, 853 |
| KAN-858 | `verify-and-handoff.md`, `visual-verify.md`, the workspace contracts, the reviewer prompts | [`audit-verify-and-contracts.md`](audit-verify-and-contracts.md), [`audit-subagent-prompts.md`](audit-subagent-prompts.md) | after 852, 853 |
| KAN-859 | The finish session | [`audit-finish.md`](audit-finish.md) | after 852, 853 |
| KAN-860 | Move hand-written procedures into scripts | the MECHANICS rows of every audit file | last |

## Where the bytes are

| Session | Loads today (directives + cited) | Cut with no code | Share |
|---|---|---|---|
| Planning (creating run → `STARTED`) | ~190 KB (~48k tok) | ~61 KB | ~32% |
| Implementation (`STARTED` → `IN_PROGRESS`, inline) | ~348 KB (~87k tok); worst case ~453 KB with UI, isolation and a big roster | ~90 KB, plus ~34 KB loaded only after the first Critical/Important finding | ~26% (+10% deferred) |
| Finish (merge-and-push, run 1 → run 2) | ~229 KB (~57k tok) | ~73 KB | ~32% |

`pipeline.md:440-448` says to load `state-file.md` (21 KB) "before reading or writing a state
file" and `project-configuration.md` (54 KB) "before resolving project configuration". Every
session does both (`flow state get`, and `project-get.sh … model`).

If runs obey that literally, add 75 KB to every row above. The run-time cores of the two files are
14.2 KB and 13.0 KB, and the cuts rise to:
- planning: ~109 KB, ≈41%;
- implementation: ~132 KB, ≈31% (+8% deferred);
- finish: ~119 KB, ≈39%.

KAN-852 measures which case holds.

**Growth since the last trim** (KAN-378/379, 2026-09-03), counting phase files plus load directives
only:
- planning: 56.6 → 88.3 KB (+56%);
- implementation: 110.6 → 231.0 KB (+109%);
- finish: 118.7 → 149.3 KB (+26%).

`implement.md` was 18.5 KB on 2026-09-01; it is 81.7 KB now. The growth is spread across
incident-driven fix commits, one paragraph at a time. `check-contract-budget.sh` was removed on
2026-09-28 (`085b9ce0`); before that, its rows were routinely raised to admit growth (e.g.
`375452a8`).

## The allowed edit shapes

Only three shapes are behaviour-neutral, per `rules/be-brief.mdc` and the KAN-378 precedent.
Nothing is reworded.

1. **Verbatim move to a lazily loaded file.** The directive reads "**Load `X`** only when Y".
   - Y must be decidable at that point, and must cover every place the moved text applies.
   - Precedents: `review-panel-optional-slots.md`, `visual-verify.md`, and the conditional
     `workspace-isolation.md` load.
2. **Verbatim move of editor-facing reasoning to the sibling `-rationale.md`.**
   - This covers rejected alternatives, design history, incident narrative and provenance.
   - A clause that conditions, scopes or breaks a tie stays: it is instruction even when it says
     "because".
3. **Deletion of a copy restated in the same session's load set, or of stale text.** This is only
   neutral when both copies say the same thing. Where they disagree, KAN-853 or KAN-854 picks the
   behaviour first.

## The moves, ranked by value per session

The per-file audits carry every candidate row with its line range, bytes, citers and guard
constraints. This section is the overview.

### A. Stop loading what the session never acts on

| Move | Saves | Session(s) | Issue |
|---|---|---|---|
| Split `project-configuration.md` to a 13.0 KB run-time core (lines 135-546 move out) | 35–41 KB each, if loaded | all 3 | KAN-855 |
| Split `state-file.md` to a 14.2 KB run-time part | 7.1 KB each, if loaded | all 3 | KAN-855 |
| Move **Resuming at `STARTED`** out of `brainstorm.md`. Every post-`/clear` implementation session currently loads all 17 KB of it to read 1.6 KB | ~15 KB | implementation | KAN-856 |
| Carry the 604-byte `STARTED` block in `brainstorm.md` instead of loading `handoff-blocks.md` | ~15 KB | planning | KAN-856 |
| Drop the `artifacts-registry.md` load at `implement.md:232-233` | 7 KB | implementation | KAN-858 |
| Scope the `workspace-isolation.md` load to the section each trigger needs | ~20 KB per load | implementation (conditional) | KAN-858 |

### B. Lazy splits by condition

| Block | Saves | Condition | Issue |
|---|---|---|---|
| `review-panel.md` fix-round block | 33.8 KB | a round records a Critical/Important finding (9 of 11 archived panels did, so this mostly defers the cost past pass 1) | KAN-857 |
| `visual-verify.md` text only the verifier acts on, moved to a file the verifier reads | 40.6 KB | UI runs | KAN-858 |
| `implement.md` §3, fix-run documentation | 7.2 KB | fix run | KAN-857 |
| `implement.md` sdd-only blocks | 4.6 KB | `execution: sdd` | KAN-857 |
| `implement.md` cross-repo text; the gated-reviewer `fix` path | 3.7 KB; 2.8 KB | more than one repo; a gated pass returns `fix` | KAN-857 |
| `review-panel.md` late-fix reduction | 4.1 KB | fix run | KAN-857 |
| Finish hand fallbacks | 5.2–6.7 KB | the guard-presence check reported a missing guard | KAN-859 |
| Self-review `run` branch | 6.5 KB | `## self review` is not `defer` | KAN-859 |
| Unfinished-work gate apparatus; base-moved rebase block | 2.9 KB; 4.8 KB | gate reports OUTSTANDING; base moved | KAN-859 |
| Withdrawal route plus reachability check | 5.2 KB | issue labelled flow-fix/flow-cost, or `STARTED` with no plan | KAN-856 |
| `git-boundaries.md` per session; `jira-integration.md` per session | 2–3.5 KB; 2.8 KB | per session | KAN-858, KAN-856 |
| `operator-prompts.md` auto-resolution and multi-select | 4.9 KB | finish sessions only, for this operator | KAN-855 |

### C. In-session duplicates

| File | Cut | Issue |
|---|---|---|
| `commands-claude/flow.md` | 2.2 KB of its 2.9 KB, in all 3 sessions. It contradicts the contract in four places | KAN-853, then KAN-855 |
| `SKILL.md` | ~5.3 KB per session | KAN-855 |
| `pipeline.md` | 5.6–6.5 KB per session | KAN-855 |
| integrate/archive against run 1/run 2 | 12.9 KB | KAN-859 |
| `verify-and-handoff.md` | 2.1 KB | KAN-858 |
| `review-panel.md` | 3.0 KB | KAN-857 |
| `brainstorm.md` | 2.3 KB | KAN-856 |

### D. Rationale moved to `-rationale.md`

| File | Moves | Issue |
|---|---|---|
| `implement.md` | 4.7 KB | KAN-857 |
| finish files | 10.8 KB | KAN-859 |
| `review-panel.md` and optional slots | 4.3 KB | KAN-857 |
| planning files | 4.4 KB | KAN-856 |
| `visual-verify.md` and `workspace-isolation.md` | 5.3 KB | KAN-858 |
| `agent-baseline.md` | 0.6 KB, saved in every subagent | KAN-858 |

KAN-378 already did one pass, so the rationale that remains is smaller than the other levers.

### E. Subagent prompts

- **Reviewer templates** (KAN-858):
  - Dispatcher-only text that every slot receives: 3.0 KB in `principles-reviewer-prompt.md`.
  - Template intros and headers: about 1.2 KB, after the bugbot and security templates were retired.
- **Reviewer slots re-read their bundle's files** (KAN-854): they re-read `proposal.md`, `design.md`,
  `tasks.md` and the principles file their context bundle already carries. That is about 21 KB per
  primary pass. It is a behaviour change, so it sits in the defects issue.

### F. Mechanics to code (KAN-860)

Each of these needs code and parity tests:
- Decide rendering;
- base-movement and base-refresh checks;
- one diff writer;
- the late-fix trigger;
- throwaway-worktree and fold-fixup scripts;
- a `flow task close` verb;
- `flow state add-worktree`;
- cleanup checks 1–6;
- `render-slot-prompt.sh`;
- a multi-stage `flow stage` verb for `/flow-fast`.

Together: about 25–30 KB.

## How to prove "no behaviour change"

- **`check-normative-inventory.sh` cannot be the safety net.**
  - It matches only `MUST`/`SHALL`: 7 sentences corpus-wide, against about 960 "never"s.
  - `.flow/project.md` still says it prints "roughly a thousand lines".
  - A `MUST` moved into a `-rationale.md` leaves its output unchanged.
- **`verbatim_check.py`** ([`tools.md`](tools.md); the guard version is KAN-852):
  - It works sentence by sentence over `skills/`, `commands-claude/` and `rules/`.
  - Every sentence that leaves the run-loaded files must still be stated in one, or land verbatim in
    a `-rationale.md`. It reports REVIEW when such a sentence carries an imperative marker.
  - Any new run-loaded sentence must be a load directive or a citation.
  - It cannot judge whether a load condition is right; review each directive by hand.
- **Guards a move must keep green:**
  - `check-references`: citers of every moved heading. It checks single-line citations only, until
    KAN-852 extends it.
  - `check-dispatch-paragraphs` (`dpSites`): fix-subagent paragraphs are pinned to `review-panel.md`,
    implementer paragraphs to `implement.md`, and two copies each to `visual-verify.md`.
  - `check-stage-mark-calls`: `smcCandidates` must list any new file that carries `flow stage` lines.
  - `check-mutation-reproducer-pin`: `review-panel.md:653`.
  - `check-guard-symlinks`: lazy siblings of phase files stay in `skills/flow/`.
  - `check-model-resolution-shell`: `archive.md`'s `SELF_REVIEW_MODEL` marker line.
  - `check-self-review-report`: run 2's angle table.
  - `journal_test.go:49` and `check_cleanup_complete_test.go`: they read `state-file.md`,
    `workspace-isolation.md` and `artifacts-registry.md`.
  - `crExpectedZero` entries in `references.go` / `installedcitations.go`, for any file that loses
    its last citation.
- **Measure before and after.**
  - `loadsets.sh`: static bytes per session.
  - `loaded_files.py` over real transcripts: what each session actually Read or Skill-loaded, when,
    re-reads, and turn-weighted cost.
  - flowd's per-stage tokens.

## Keeping it slim afterwards

This avoids reinstating the per-file ratchet that was removed on purpose:
- New incident narrative goes straight to `-rationale.md` at write time.
- A non-blocking per-session load report (KAN-852) keeps regrowth visible.

## Files in this directory

| File | What |
|---|---|
| [`rubric.md`](rubric.md) | The brief every audit pass followed: levers, allowed moves, report format |
| [`audit-every-session.md`](audit-every-session.md) | `SKILL.md`, `pipeline.md`, `commands-claude/flow.md`, `CLAUDE.md`, the always-on block, `operator-prompts.md`, `model-policy.md`, `state-file.md`, `project-configuration.md` |
| [`audit-planning.md`](audit-planning.md) | `brainstorm.md`, `brainstorm-planner.md`, `jira-integration.md`, `plan-provenance.md`, `build-green.md`, `handoff-blocks.md` |
| [`audit-implement.md`](audit-implement.md) | `implement.md` |
| [`audit-review-panel.md`](audit-review-panel.md) | `review-panel.md`, `review-panel-optional-slots.md` |
| [`audit-verify-and-contracts.md`](audit-verify-and-contracts.md) | `verify-and-handoff.md`, `visual-verify.md`, `workspace-isolation.md`, `git-boundaries.md`, `artifacts-registry.md`, `session-records.md`, `worktree-resolution.md` |
| [`audit-subagent-prompts.md`](audit-subagent-prompts.md) | the reviewer templates, `engineering-principles.md`, `agent-baseline.md`, `failure-modes.md`, `flow-fast/SKILL.md` |
| [`audit-finish.md`](audit-finish.md) | `integrate.md`, `archive.md`, `finish-contract-run1.md`, `finish-contract-run2.md`, `jira-followups.md` |
| [`tools.md`](tools.md) | The prototypes `loadsets.sh`, `loaded_files.py` and `verbatim_check.py` |
