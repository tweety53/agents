# Subagent-facing prompts + `/flow-fast` — slimming analysis

> **Audit snapshot at `700e184` (2026-09-29).** A read-only model pass that followed [`rubric.md`](rubric.md); every row is a candidate to verify, not a verified fact. Line numbers refer to `700e184`. Main has since retired the bugbot and security slots (`ebdfdde1`..`4dafab4a`), so re-locate each row by its quoted first words, and treat bugbot/security rows as obsolete. Tracked in KAN-851 → KAN-858 (and the Drift items in KAN-853 / KAN-854). Plan: [`README.md`](README.md).

Read-only analysis of `skills/flow/{primary,principles,bugbot,security}-reviewer-prompt.md`,
`skills/flow/experimental/failure-modes.md`, `skills/flow/engineering-principles.md`,
`rules/agent-baseline.md`, `skills/flow-fast/SKILL.md` and `agents/flow-*.md`. Every file read in
full; every byte count is `sed -n 'A,Bp' FILE | wc -c` or, where marked *(substr)*, the exact byte
length of the quoted span (leading separator space included). Nothing in the repository
was modified (`git status --porcelain` empty; guards run with the scratch build cache).

**Load facts this analysis rests on (verified, not assumed):**

- **Templates are read by the slot subagent, not the parent.** `review-panel.md:152-157`: "A
  subagent-facing file is passed by absolute path, never read into this context … the dispatcher
  resolves their paths … and names them in the prompt." So the dispatcher-facing parts of a
  template (intro, Agent-call header, Placeholders) are paid by the subagent. The one exception is
  `review-panel.md:667-670`, which routes the parent into the principles template's
  `[STANDARDS_PATHS]` step (see D3).
- **One bundled subagent reads several templates.** Default grouping puts `primary`,
  `principles`, `security` in one dispatch (`review-panel.md:330-336`); every decided class's
  floor bundle is `primary+principles` (`brainstorm-planner.md:462-499`). So `principles` and
  `primary` load on essentially every panel round; `bugbot`/`security` on `big` (or a custom store
  list); `failure-modes.md` when `experimental_roll < 30` on a decided panel (it is the only file
  in `experimental/`).
- **Every dispatched subagent reads `rules/agent-baseline.md` first** (hook-enforced pointer).
- **The dispatch paragraphs (NO DELEGATION, MODEL HANDSHAKE, REPRODUCE DON'T READ, TOOLS,
  FOREGROUND BUILDS, REPORT FILE, WORKTREES, CITATION CHECK, ENTRY CONTEXT, CONTEXT BUNDLE) are in
  none of these files.** They live in `review-panel.md:361-364, 504-588` and are typed once per
  bundle prompt. `check-dispatch-paragraphs` (`stats/internal/guard/dispatchparagraphs.go:89-117`)
  pins sites only in `review-panel.md`, `implement.md`, `visual-verify.md` — **no candidate below
  is forbidden by it.** Its bash header says as much: panel slots carry READ-ONLY "inside their
  prompt files' own Read-Only Review sections, which this table does not pin."
- **Guard side effects of moving template text.** Every slash-bearing citation in the five
  templates sits *outside* the prompt fence — in the intro or the Placeholders list (verified per
  line). `check-references` coverage today: primary 1 (line 3), principles 1 (line 4), security 1
  (line 2), bugbot 3 (2, 5, 100), failure-modes 2 (5, 121). `check-installed-citations`: 6/11/7/5/5,
  all from the same regions. Removing intros zeroes `check-references` coverage for
  primary/principles/security. Removing intros *and* Placeholders zeroes both guards for all five.
  Each zero needs an entry in `crExpectedZero` (`stats/internal/guard/references.go:103`, the
  existing "reviewer-prompt file, deliberately self-contained" category) and in `cicExpectedZero`
  (`stats/internal/guard/installedcitations.go:92-109`). That is a small Go edit, not prose. The
  comment at `references.go:61-65` shows the templates *were* citation-free once.
- **No `MUST`/`SHALL` in any assigned file** (grep, whole-word), so `check-normative-inventory.sh`
  output is unchanged by any move or cut below.

## Totals

Conservative: only high/med-confidence rows are counted; each byte is counted once. Low-confidence
rows (E3, A5, A6, FF8, FF10, FF13) are listed under Candidates but excluded here.

| file | bytes | RATIONALE | DUPLICATE | STALE | LAZY-SPLIT | MECHANICS | MISPLACED |
|---|---:|---:|---:|---:|---:|---:|---:|
| `skills/flow/primary-reviewer-prompt.md` | 5,816 | 0 | 19 | 0 | 0 | 569 | 668 |
| `skills/flow/principles-reviewer-prompt.md` | 9,464 | 0 | 198 | 144 | 0 | 259 | 2,646 |
| `skills/flow/bugbot-reviewer-prompt.md` | 5,792 | 0 | 0 | 0 | 0 | 684 | 834 |
| `skills/flow/security-reviewer-prompt.md` | 4,628 | 0 | 19 | 0 | 0 | 532 | 564 |
| `skills/flow/experimental/failure-modes.md` | 6,818 | 172 | 19 | 0 | 0 | 445 | 462 |
| `skills/flow/engineering-principles.md` | 8,487 | 191 | 0 | 0 | 0 | 0 | 0 |
| `rules/agent-baseline.md` | 5,795 | 539 | 90 | 0 | 0 | 0 | 0 |
| `skills/flow-fast/SKILL.md` | 23,367 | 266 | 666 | 0 | 0 | 2,582 | 79 |
| `agents/flow-{low,medium,high,xhigh}.md` | 1,436 | 0 | 0 | 0 | 0 | 0 | 0 |
| **total** | **71,603** | **1,168** | **1,011** | **144** | **0** | **5,071** | **5,253** |

12,647 B in candidates (17.7 %). 7,576 B of that is behaviour-neutral without new code, i.e.
everything except MECHANICS.

Outside the assigned set but in the same `/flow-fast` session: `commands-claude/flow-fast.md`
(2,435 B) has 1,232 B DUPLICATE (row X1).

**Share of each template that is dispatcher-facing text the subagent reads.** Compare the file
against its prompt body (the fence minus the Agent-call header):

| template | file | prompt body | dispatcher-facing |
|---|---:|---:|---:|
| primary | 5,816 | 4,552 (14-102) | 1,264 (22 %) |
| principles | 9,464 | 6,388 (22-150) | 3,076 (33 %) |
| bugbot | 5,792 | 4,266 (15-92) | 1,526 (26 %) |
| security | 4,628 | 3,505 (13-85) | 1,123 (24 %) |
| failure-modes | 6,818 | 5,770 (14-115) | 1,048 (15 %), of which line 1 must stay |

## Candidates

| id | file:startline-endline | bytes | lever | condition or canonical location or evidence | first ~10 words verbatim | confidence | behaviour-risk note |
|---|---|---:|---|---|---|---|---|
| P1 | primary-reviewer-prompt.md:1-4 | 233 | MISPLACED | Dispatcher-facing; the parent never reads the file (review-panel.md:152). Parent-side canonical: review-panel.md:146 (roster row) and :218-230 (docs-only → `primary`). | Use this template for the panel's **Primary** slot — plan | med | Subagent cannot act on it. Its line 3 is the file's only `check-references` check, so add `crExpectedZero`. "on every roster" is also inexact: a micro store list without `primary` (SKILL.md:99-103). |
| P2 | primary-reviewer-prompt.md:5-6 | 19 | DUPLICATE | Same file :43-46 (`## Read-Only Review`, inside the prompt) | Read-only review. | high | None |
| P3 | primary-reviewer-prompt.md:8-13 | 435 | MISPLACED | Agent-call params (subagent_type, description, model, `prompt: \|`). Canonical: review-panel.md:161-166, :176-181. No code or doc consumes the `description:` strings (grep). | Subagent (<the dispatch's subagent_type>):  # flow-low on a `default` panel; | med | Keep the fence around the indented prompt. The subagent cannot act on these lines. A parent that read the template against :152 would lose a restatement only. |
| P4 | primary-reviewer-prompt.md:104-112 | 569 | MECHANICS | Needs `render-slot-prompt.sh` (see Q2). `[DIFF_PATH]`→review-panel.md:854-858; `[CONTEXT_BUNDLE_PATHS]`→:504-508; `[ARTIFACT_PATHS]` is defined nowhere parent-side (D2). | **Placeholders:** - `[DIFF_PATH]` — `<abs-worktree>/.superpowers/sdd/final-review.diff`, or on | med (as script) / low (plain cut) | The fill protocol is ambiguous (D2): the subagent may be the one resolving from this list. Do not cut without the script. It carries the file's `check-installed-citations` coverage. |
| R1a | principles-reviewer-prompt.md:1-2 | 39 | MISPLACED | Title heading; no citer (grep) | # Principles Reviewer Prompt Template | med | None |
| R1b | principles-reviewer-prompt.md:3-5 | 144 | STALE | "required on every `/flow` run" is false. The docs-only reduction drops it (review-panel.md:218-230); a micro run dispatches exactly the store list (SKILL.md:99-103); the fixed roster was superseded (SKILL-rationale.md:270-272). | Use this template for the panel's **Principles** slot — required on | med-high | Only `check-references` check of the file (line 4), so add `crExpectedZero`. |
| R1c | principles-reviewer-prompt.md:6-12 | 551 | MISPLACED | Addressed to "the dispatcher". Canonical: review-panel.md:657-665 ("never a pasted copy", absolute `[PRINCIPLES_PATH]`). Its reason clause ("the subagent's working directory is the project worktree…") moves verbatim to skills/flow/SKILL-rationale.md. | The principle list itself is **not** restated here — the reviewer | med | The subagent's "read the principles file FIRST" stays at :44. |
| R2 | principles-reviewer-prompt.md:13-14 | 19 | DUPLICATE | Same file :61-64 | Read-only review. | high | None |
| R3 | principles-reviewer-prompt.md:16-21 | 440 | MISPLACED | As P3. The description label "Principles review (Merged)" has no current meaning. | Subagent (<the dispatch's subagent_type>):  # flow-low on a `default` panel; | med | As P3 |
| R4 | principles-reviewer-prompt.md:66-71 | 179 | DUPLICATE | engineering-principles.md:15-16 ("All three groups below always apply … No group may be left uncovered"), which the same subagent reads FIRST (principles:44). Also review-panel.md:147. | ## Your Scope  Apply all three principle groups: `## Structure`, | med | Heading is inside the fence and cited by nobody. |
| R5 | principles-reviewer-prompt.md:152-155 + :176 | 259 | MECHANICS | As P4. `[DIFF_PATH]` here names `fix-round-N.diff`, drifted from the other four (D4). `[GLOBAL_CONSTRAINTS]` has no parent-side producer (D2). | **Placeholders:** - `[DIFF_PATH]` — `<abs-worktree>/.superpowers/sdd/final-review.diff`, or on | med (script) / low | As P4 |
| R6 | principles-reviewer-prompt.md:156-165 | 782 | MISPLACED | Parent-filled (review-panel.md:663). "Verify … exists … stop" duplicates review-panel.md:664-665: delete. The install-path worked example and "never hardcode a repo-relative `skills/…` path" have no parent-side copy: move verbatim to review-panel.md **Principles**. | - `[PRINCIPLES_PATH]` — the **absolute** path of `engineering-principles.md` inside the | med | Worked example must be moved, never dropped. Net for the parent: about +0.5 KB in review-panel.md. The MECHANICS script makes it about 0. |
| R7 | principles-reviewer-prompt.md:166-175 | 834 | MISPLACED | Parent-facing: review-panel.md:667-670 has the parent resolve standards "per … the `[STANDARDS_PATHS]` step of `skills/flow/principles-reviewer-prompt.md`". Move verbatim to review-panel.md **Principles**; retarget :669. Entry-form table canonical: project-configuration.md:43-84. | - `[STANDARDS_PATHS]` — the project's own written standards. Resolve in | med | **Resolve D3 first:** the step's auto-detect fallback contradicts review-panel.md:667-670. Upgrade path: `resolve-standards.sh` (MECHANICS). |
| B1 | bugbot-reviewer-prompt.md:1-7 | 408 | MISPLACED | Intro plus "Read-write review … how those copies are made and removed". Canonical: review-panel.md:148 and review-panel-optional-slots.md **The throwaway worktree**. The subagent's own obligation is already at bugbot:27-30. | Use this template for the panel's **Bugbot** slot — a defect | med | Coverage drops 3→1 (line 100 remains). |
| B2 | bugbot-reviewer-prompt.md:9-14 | 426 | MISPLACED | As P3 | Subagent (<the dispatch's subagent_type>):  # flow-low on a `default` panel; | med | As P3 |
| B3 | bugbot-reviewer-prompt.md:94-103 | 684 | MECHANICS | As P4. Its re-run delta contradicts review-panel.md:870-871 (D4). `[REPO_COPIES]` lives parent-side in review-panel-optional-slots.md. | **Placeholders:** - `[DIFF_PATH]` — `<abs-worktree>/.superpowers/sdd/final-review.diff`, or on | med (script) / low | As P4 |
| S1 | security-reviewer-prompt.md:1-3 | 143 | MISPLACED | Canonical review-panel.md:149 | Use this template for the panel's **Security** slot — dispatched like | med | Only `check-references` check (line 2), so add `crExpectedZero`. |
| S2 | security-reviewer-prompt.md:4-5 | 19 | DUPLICATE | Same file :42-45 | Read-only review. | high | None |
| S3 | security-reviewer-prompt.md:7-12 | 421 | MISPLACED | As P3 | Subagent (<the dispatch's subagent_type>):  # flow-low on a `default` panel; | med | As P3 |
| S4 | security-reviewer-prompt.md:87-94 | 532 | MECHANICS | As P4. `[PLAN_OR_REQUIREMENTS]` has no parent-side definition (D2). The delta contradicts review-panel.md:870-871 (D4). | **Placeholders:** - `[DIFF_PATH]` — `<abs-worktree>/.superpowers/sdd/final-review.diff`, or on | med (script) / low | As P4 |
| F1 | experimental/failure-modes.md:2-6 | 238 | MISPLACED | Canonical review-panel-optional-slots.md:7-45 (**Experimental slot**) | Use this template for the panel's experimental `exp-failure-modes` slot, dispatched | med | **Keep line 1** (`description:`, read by brainstorm-planner.md:529-533). Coverage 2→1. |
| F2 | experimental/failure-modes.md:7-8 | 19 | DUPLICATE | Same file :67-70 | Read-only review. | high | None |
| F3 | experimental/failure-modes.md:10-13 | 224 | MISPLACED | As P3. Canonical review-panel-optional-slots.md:9-12 and review-panel.md:176-181. | Subagent (flow-<effort>):  # decided-panel-only slot | med | As P3 |
| F4 | experimental/failure-modes.md:117-123 | 445 | MECHANICS | As P4 | **Placeholders:** - `[DIFF_PATH]` — `<abs-worktree>/.superpowers/sdd/final-review.diff`, or on | med (script) / low | As P4 |
| F5 | experimental/failure-modes.md:19-21 | 172 | RATIONALE | Why the slot exists. The job is already fixed by the role statement at :14-17. Move to skills/flow/SKILL-rationale.md. | No persistent slot systematically asks what a changed function does | med | "That is this slot's whole job" restates :16-17. |
| E1 | engineering-principles.md:4-5 *(substr)* | 191 | RATIONALE | Who reads the file. Implementers already get it from REQUIRED READING (implement.md:699-701); the reviewer from its template. Move to skills/flow/SKILL-rationale.md (relative link still resolves there). | Implementer dispatches in `/flow` require this file as reading; the | med | Loaded by every implementer and slot (via the bundle), and twice by the principles slot (D1). `cicExpectedZero`'s reason string mentions this link, so it goes stale without failing. |
| E3 | engineering-principles.md:18-19 | 92 | DUPLICATE | Same file :15-16 | The groups are headings because they organise the reading; they | low | "do not partition the work" reads as normative; keep unless the owner agrees. |
| A1 | agent-baseline.md:3-5 *(substr)* | 165 | RATIONALE | Design reason ("never a copy … nothing here to drift"). Destination: new `rules/agent-baseline-rationale.md`, which is not installed by setup.sh (install_rules_claude links only always-on `.mdc` + this file) but needs a `crExpectedZero` entry. Alternative: dispatch-carries-the-baseline.mdc "## Why a pointer rather than the rules themselves" (non-core). | It carries **one line per rule plus a pointer to | high | Read by every dispatched subagent. |
| A2 | agent-baseline.md:15-17 *(substr)* | 204 | RATIONALE | Reason for "Unconditionally" (kept). A near-copy sits in the always-on core of dispatch-carries-the-baseline. | You cannot know from a dispatch prompt whether the agent | high | None |
| A3 | agent-baseline.md:19 *(substr)* | 90 | DUPLICATE | agent-baseline.md:38 "A hook denies a dispatch that omits them." | This is enforced: a `PreToolUse` hook denies any dispatch whose | med | The carve-outs sentence that follows stays. |
| A4 | agent-baseline.md:54-56 *(substr)* | 170 | RATIONALE | Why the pipeline is absent from the table. The instruction ("any `/flow*` step loads its own contract file first") stays. | The flow pipeline is deliberately absent from the table above — | med | None |
| A5 | agent-baseline.md:60 *(substr)* | 59 | DUPLICATE | Row :32 plus be-brief's mandatory dispatch sentence | Findings first, bullets over prose, no preamble, no recap. | low | The dispatch copy depends on the dispatcher complying; keep. |
| A6 | agent-baseline.md:28-40 | 3,014 | DUPLICATE | Managed-block cores, if the always-on block reaches Claude Code subagents (D6) | \| Rule \| Full text \| … **Never touch production.** No SSH, | low | Needs harness verification; ZCode and claudeMd-less agents need it. Do not act yet. |
| FF1 | flow-fast/SKILL.md, 12 empty begin/end pairs (110-111, 195-196, 202-203, 286-287, 304-305, 311-312, 332-335, 395-398, 414-417) | 2,582 | MECHANICS | A `flow stage` verb that marks a list of keys begin+end in one call (net about −1.4 to −2.4 KB). Needs the CLI, `check-stage-mark-calls` (smcCandidates has SKILL.md) and stats parity. Stage-keys table untouched (TestStageKeysMatchFlowFastSkillTable). | flow stage begin -command '/flow-fast' -stage flow.isolate-workspace -harness <harness> | low-med | "Every mark is a literal call" is a deliberate design (SKILL.md:23-27). The 54 `flow stage` lines total 5,788 B. |
| FF3 | flow-fast/SKILL.md:87-89 *(substr)* | 204 | DUPLICATE | implement.md **4. Execute (SDD + TDD)**, loaded on `sdd` and cited "as written" | — one implementer per decided group on that group's model | med | Leaves "as written except that …". |
| FF4 | flow-fast/SKILL.md:93-97 *(substr, final period kept)* | 294 | DUPLICATE | review-panel.md:176-181, :876-889, :1327-1335 (fix routing inline→parent, sdd→`fixer`), loaded whenever this branch runs | — the decision's roster, grouping and dispatches on their own | med | Keep "A `default` panel (the `micro` class) runs no panel": that is an override. |
| FF5 | flow-fast/SKILL.md:107 *(substr)* | 68 | DUPLICATE | jira-integration.md:93-103, :150-160, already cited "per **Transitions** there" | (by name, forward-only, one line on failure per **Never blocking**) | med | None |
| FF6 | flow-fast/SKILL.md:217-218 *(substr)* | 100 | DUPLICATE | Always-on commit-scope core in the managed block (every session) | with the scope naming the module the commit moved | high | Leaves "Conventional Commits form, no attribution trailer". |
| FF7 | flow-fast/SKILL.md:221-224 *(substr, trailing comma kept)* | 266 | RATIONALE | Why no task-fields guard reads a flow-fast commit. The guard name stays at SKILL.md:89. Move to skills/flow-fast/SKILL-rationale.md (declared zero; moved text has no bold, so it stays zero). | : `check-task-commit-fields.sh`, the guard `/flow`'s implement phase closes | med | The instruction "no task-fields guard reads it, and section 5's lint and tests are the only close" stays. |
| FF8 | flow-fast/SKILL.md:281-282 *(substr)* | 113 | RATIONALE | What `/flow-self-review` does later | `/flow-self-review <name>` then runs the pass on `<default-branch>` | low | May stop the run deleting the bundle itself; keep. |
| FF9b | flow-fast/SKILL.md:359-360 *(substr)* | 79 | MISPLACED | Project-configuration advice the run cannot act on. The run's own fallback is at :357-359. | ; a project whose default branch is protected declares `pull request` | med | None |
| FF10 | flow-fast/SKILL.md:258-282 | 1,650 | LAZY-SPLIT | Condition `## self review` = `defer`. This operator's `.flow/project.md:247-249` is `defer`, so it is almost never false. | With `## self review` `defer` (**Project configuration**, | low | No real saving for this operator; not worth a file. |
| FF13 | flow-fast/SKILL.md:260-261 *(substr)* | 62 | RATIONALE | Reason clause | , since `/flow-fast` asks no review question and runs no pass. | low | Tiny |
| X1 | commands-claude/flow-fast.md:9-19 (minus "Follow that skill exactly.") + :21-23 | 1,232 | DUPLICATE | Summary of skills/flow-fast/SKILL.md, which is canonical ("Follow that skill exactly") | One invocation runs from the Jira key to the landed change: | med-low | Outside the assigned set. Line 10 is the stub's only coverage for both guards, so a declaration is needed. It also mis-cites **Transitions** for name resolution (D9). |

**Cross-file MECHANICS the caller asked about (not counted above; the bytes live in review-panel.md).**
The parent types these dispatch paragraphs into every bundle prompt:

- review-panel.md:361-364, 514-516, 520-523, 527-531, 535-536, 545-549, 554-557, 562-568,
  573-576, 582-588, plus implement.md:704-709: **4,325 B per bundle dispatch**.
- The reproducer rule (review-panel.md:429-488, 5,398 B), if pasted.

All of it stays in the implementation parent's context as tool-call input. A render script can
write them into the rendered prompt file from their pinned blocks in review-panel.md. The blocks
stay there, so `check-dispatch-paragraphs` still passes. Two things must stay typed in the Agent
call: the MODEL HANDSHAKE (answered before any tool call) and the baseline pointer (the hook
checks the prompt).

## Top 5

1. **principles-reviewer-prompt.md dispatcher text out: 2,988 B.** Covers R1a/b/c, R2, R3, R4,
   R6, R7. The subagent saves it on every panel round (floor bundle). The implementation parent
   stops being routed into a 9.5 KB subagent file by review-panel.md:669. R6/R7 land in
   review-panel.md **Principles**, or become one script call. Resolve D3 first.
2. **Intro + Agent-call header + "Read-only review." out of the other four templates: 2,585 B.**
   primary 687 (every round), failure-modes 481 (~30 % of decided panels), bugbot 834 and
   security 583 (`big` bundles). Needs `crExpectedZero` entries for primary/principles/security.
3. **`render-slot-prompt.sh` (MECHANICS): 2,489 B.** Removes the remaining Placeholders lists
   (P4, R5, B3, S4, F4) from slot subagents and turns the undefined fill protocol (D2) into code.
   Extended to the shared paragraphs, it also cuts ≥ 4.3 KB of typed prompt per bundle dispatch
   from the implementation parent.
4. **agent-baseline.md rationale + duplicate: 629 B** (A1, A2, A4 to a rationale file; A3 cut).
   Saved in every dispatched subagent: implementers, slots, fixers, verifier, per-task reviewers.
5. **flow-fast/SKILL.md neutral trims: 1,011 B** (FF3, FF4, FF5, FF6, FF7, FF9b), for the whole
   `/flow-fast` parent session. Next step: FF1's stage verb, 2,582 B (MECHANICS).

## Drift / bugs found

- **D1: duplicate runtime reads. This is the biggest cost here, but fixing it is a behaviour
  change, not a trim.**
  - Every slot's prompt carries the CONTEXT BUNDLE paragraph (review-panel.md:504-508). The
    panel bundle is built with `<principles-path>` (review-panel.md:112-114) and holds, in order,
    proposal.md, design.md, tasks.md and the principles file
    (`scripts/gather-dispatch-context.sh:33-37`).
  - The templates still send the reviewer to the same sources separately:
    - primary:26-28 reads `proposal.md`, `design.md`, `tasks.md` via `[ARTIFACT_PATHS]`;
    - security:23 reads proposal + design via `[PLAN_OR_REQUIREMENTS]`;
    - principles:40-44 "Read the principles file FIRST" via `[PRINCIPLES_PATH]`.
  - Size, from medians of 106 archived changes: proposal 2.1 KB, design 6.1 KB, tasks 13.2 KB;
    principles 8.5 KB. That is about 21 KB duplicated per primary pass, 8.5 KB per principles
    pass and ~8 KB per security pass.
  - Separately, primary, security, bugbot and mutation each carry 8.5 KB of principles in their
    bundle that their angle never uses (primary:63-66 "Do not duplicate the Principles … angles").
  - The separate paths are still needed on the CONTEXT BUNDLE FAILURE "Continue" path
    (review-panel.md:131-136), so the fix is conditional wording plus a panel-bundle option.
    The panel slots also carry no OUTPUT BUDGET ("never re-read a file already in your context").
- **D2: placeholder fill protocol is undefined.**
  - review-panel.md:152-157 forbids the parent reading templates. Yet only `[PRINCIPLES_PATH]`,
    `[STANDARDS_PATHS]` (:663-671) and `[TOUCHED_FILES]` (:590) are defined parent-side.
  - `[DIFF_PATH]`, `[ARTIFACT_PATHS]`, `[REPO_COPIES]`, `[PLAN_OR_REQUIREMENTS]`,
    `[GLOBAL_CONSTRAINTS]` and `[CONTEXT_BUNDLE_PATHS]` exist only in the templates' Placeholders
    (grep across skills/, commands-claude/, rules/: no other hit). Either the parent reads the
    templates anyway, or the subagent self-resolves.
  - `[GLOBAL_CONSTRAINTS]` ("verbatim constraints from design/specs", principles:176) needs
    judgment, and no parent step produces it.
- **D3: principles `[STANDARDS_PATHS]` routing and content conflict.**
  - review-panel.md:667-670 sends the parent into principles:166-175, contradicting :152.
  - The two disagree when no `## standards` is declared. review-panel.md says "(exit 1: none
    declared) … Pass an **empty** value when none resolve". The template says "otherwise
    auto-detect: `<project>/CLAUDE.md`, `<project>/AGENTS.md` and `CONTRIBUTING.md`".
  - That fallback hands the reviewer both harness renderings, against agent-baseline.md:49-51
    ("never read it as well"). CLAUDE.md is also already in a Claude Code subagent's context.
    This repo's own `## standards` lists both (.flow/project.md:230-233; 6,636 + 11,514 B).
- **D4: `[DIFF_PATH]` re-run semantics drifted.**
  - principles:154-155 says `fix-round-N.diff`. primary:106-108, bugbot:96-98, security:89-91
    and failure-modes:119-121 say `slot-delta-<round>-<id>.diff`.
  - review-panel.md itself says:
    - :854-858: a default panel reads slot deltas;
    - :876-880: a decided panel reads `fix-round-N.diff` plus finding sites;
    - :870-871: "Bugbot, Mutation and Security read no diff file and re-run in their pass-1
      shape". That contradicts bugbot's and security's slot-delta placeholder.
  - review-panel-optional-slots.md:28 calls Mutation "a diff-reading slot", contradicting :870.
  - Every template body hard-codes "Read `final-review.diff` in full" even when `[DIFF_PATH]` is
    a delta.
- **D5: Bugbot's copy of the mutation brief is incomplete.** review-panel-optional-slots.md:54
  names bugbot-reviewer-prompt.md as Bugbot's copy of the mutation-testing brief. bugbot:21-25
  lacks review-panel.md:643-655's "confirm the edit landed … a refusal, never a surviving mutant
  … never buys a test" and the `# mutation-reproducer` declaration. Either the parent also pastes
  the brief (then bugbot:21-25 is a partial duplicate) or Bugbot runs without those rules.
- **D6: harness premise conflict (verify on the operator's machine).**
  - dispatch-carries-the-baseline's always-on core says "An agent you dispatch starts from an
    empty context: it has never read this file". agent-baseline.md:49 says "its instruction file
    is already in your context".
  - On Claude Code, subagents receive CLAUDE.md content: this analysis subagent received
    `CLAUDE.md` in its first message. `~/.claude/CLAUDE.md` (the managed block)
    rides the same mechanism.
  - If confirmed, agent-baseline.md:28-40 (3,014 B) duplicates the managed block for Claude Code
    subagents (A6).
- **D7: stale roster claims.** principles:3 "required on every `/flow` run" is false (R1b).
  primary:1-2 "on every roster" is false for a micro store list without `primary`.
- **D8: /flow-fast re-reads instruction files.** flow-fast/SKILL.md:187 tells the session to read
  `<project>/CLAUDE.md` and `<project>/AGENTS.md`. One is already in context; the other is the
  other harness's rendering (agent-baseline.md:49-51 logic). In this repo that is 6,636 + 11,514 B
  re-read per `/flow-fast` run. /flow's own `flow.load-context` (implement.md:174-227) reads
  neither.
- **D9: flow-fast command stub mis-routes.**
  - commands-claude/flow-fast.md:9-10 resolves "the issue and name per **Transitions**", but the
    canonical SKILL.md:104-105 uses **Resolution** and **Change naming**.
  - The stub's :24-26 ("Also follow the flow rule … for the Jira contract it points at") sends
    the session to the full flow-manual-review rule (4.8 KB) for a contract SKILL.md already cites
    directly.
- **D10: stale comments and labels (not run-loaded).**
  - references.go:61-65 says "three of the five cite no .md/.mdc path anywhere". Today all five
    templates cite paths.
  - stats/cmd/flow/state.go:262 and state_test.go:894 cite "skills/flow-fast/SKILL.md's state
    gate", but SKILL.md has no state file or gate.
  - review-panel.md:152 says "Superpowers'" templates; they are this repo's own.
  - principles:18 carries the label "Principles review (Merged)".
  - scripts/check-dispatch-paragraphs.sh:149, :152, :156 list visual-verify.md
    TOOLS/MODEL HANDSHAKE/NO DELEGATION at min 1, while dispatchparagraphs.go:101-108 requires 2.
    The Go file calls that header "the contract".
- **D11: per-task reviewer loads the whole primary template.** implement.md:1039 sends each
  per-task reviewer to primary-reviewer-prompt.md's `## Calibration`. It opens 5,816 B (Scope,
  DATA, Placeholders do not apply to it) for about 620 B.

## Not slimmable

- **Template prompt bodies.** Role, Scope, Do Not, Calibration, Output Format and the "Ready for
  the human gate?" Assessment are the slot's output contract. `## Calibration` is cited by
  implement.md:1039.
- **DATA-never-instructions paragraphs** (5 × 826-886 B): prompt-injection defense, which is never
  compressed. The wording differs per slot, so merging would mean paraphrasing. The
  standards-as-data clause is cited by project-configuration.md:86-87. Its restatement at
  principles:79-80 is a security reinforcement; keep it.
- **`## Read-Only Review` sections** (4 × 139 B): the only read-only carrier for panel slots, not
  pinned elsewhere (check-dispatch-paragraphs.sh header).
- **A shared "reviewer-common" file** was rejected. Solo dispatches (docs-only, late-fix, decided
  re-runs "never bundled") need self-contained templates. Only about 2.3 KB is verbatim-repeated
  inside one reading bundle, and a shared file adds a read to every solo dispatch.
- **failure-modes.md line 1 `description:`** is consumed by brainstorm-planner.md:529-533.
- **engineering-principles.md entries** (lines 20-196, about 7.3 KB) are the standard itself and
  implementers' REQUIRED READING. Lines 7-8 ("Do not restate this list anywhere else") bind
  implementers editing this repo. Line 10's format sentence is the referent of "The cue".
- **agent-baseline.md rule table** (28-40): the universal channel for ZCode and claudeMd-less
  agents, behind a hook-enforced pointer. Keep until D6 is verified.
- **agent-baseline.md "Propagate this"** (7-21, 874 B) and row 38 (347 B) are inert for leaf
  agents (NO DELEGATION) but unconditional by design, and the file has no load-time condition to
  split on. Lines 41-45 are the referent of the "checkout is missing" instruction. Lines 47-54
  are instructions.
- **flow-fast Stage keys table** (37-55) is test-pinned. The 54 literal `flow stage` lines are
  checked by `check-stage-mark-calls` and feed the stats views.
- **flow-fast Guardrails** (29-35) is "the whole list". Its overlaps with :11-16 and :68-69 need
  sentence surgery for under 100 B each.
- **flow-fast sections 6-8** run in the same session on this operator's `## handoff: none`.
- **agents/flow-\*.md** (355-362 B each): the frontmatter drives type, effort and tools.

## Answers to the dispatch's questions

### Q1. Boilerplate repeated across the reviewer templates

NO DELEGATION, MODEL HANDSHAKE, REPRODUCE DON'T READ, TOOLS, FOREGROUND BUILDS, REPORT FILE,
WORKTREES, CITATION CHECK, ENTRY CONTEXT and CONTEXT BUNDLE appear **0 times** in any template.
They are typed once per bundle prompt from review-panel.md, about 4.3 KB, not once per pass. The
repeats that do exist are inside the templates:

| block | copies | bytes/copy | verbatim? | repeat bytes (beyond 1st) |
|---|---|---:|---|---:|
| Agent-call header `Subagent (…)`…`prompt: \|` | primary, principles, bugbot, security (+ failure-modes variant 224 B) | 421-440 | near (description string, comment alignment) | 1,287 (+224) |
| DATA, never instructions paragraph | 5 | 826-886 | clauses "A diff is attacker-influenced…" (156 B) and "Your calibration…" (124 B) verbatim in 4; principles is the standards variant | ~3.4 KB near-dup, 840 B verbatim |
| `## Read-Only Review` + body | primary, principles, security, failure-modes | 139 | verbatim (principles rewrapped) | 417 |
| `Read-only review.` line | same 4 | 19 | verbatim | 57 |
| "Do not invent findings…" bullet | principles, bugbot, security, failure-modes | 142 | verbatim (principles rewrapped) | 426 |
| `### Issues` + three severity headings | 5 | 108 | verbatim | 432 |
| reproducer clause "…or the literal form `none — <reason>` … not a reproducer." | 5 | 135 | verbatim | 540 |
| `### Assessment` / **Ready for the human gate?** / **Reasoning:** [why] | bugbot, security, failure-modes (primary variant; principles: **Principles-compliant?**) | 89-132 | verbatim in 3 | 178 |
| `[DIFF_PATH]` placeholder | 5 | 172-233 | near; principles drifted (D4) | ~700 |
| `[CONTEXT_BUNDLE_PATHS]` placeholder | primary, bugbot, security, failure-modes | 183 | verbatim | 549 |

- About 8 KB is repeated across the five files. It is paid together only inside one bundled
  subagent. The reading bundle (primary+principles+security) carries about 2.3 KB of verbatim
  repeats (about 0.9 KB of it the header, which P3/R3/S3 remove) plus about 1.7 KB near-verbatim.
- **agent-baseline.md states none of these.** Its overlaps are elsewhere:
  - "Reporting back" (:58-61) matches be-brief's dispatch sentence.
  - "Propagate this" is the mirror of NO DELEGATION: it tells an agent what to put in prompts it
    writes, which a leaf never writes.

### Q2. Dispatcher-only text the subagent receives, and the reverse

**Subagent receives dispatcher-only text.**

| template | intro | "Read-only review." | Agent-call header | Placeholders | total |
|---|---:|---:|---:|---:|---:|
| primary | 1-4 (233) | 5-6 (19) | 8-13 (435) | 104-112 (569) | 1,256 B |
| principles | 1-12 (734) | 13-14 (19) | 16-21 (440) | 152-176 (1,875) | 3,068 B |
| bugbot | 1-7 (408), incl. the read-write note | none | 9-14 (426) | 94-103 (684) | 1,518 B |
| security | 1-3 (143) | 4-5 (19) | 7-12 (421) | 87-94 (532) | 1,115 B |
| failure-modes | 2-6 (238), line 1 stays | 7-8 (19) | 10-13 (224) | 117-123 (445) | 926 B |

The Placeholders lists are ambiguous-audience (D2); all the other parts are dispatcher-only
outright.

**Reverse: subagent-only text the parent carries.** By rule, none: review-panel.md:152 says the
parent never reads a template. Three exceptions:

- review-panel.md:669 sends the parent into principles:166-175. If it reads the whole file, it
  carries 6,388 B of prompt body it never acts on.
- If the parent reads templates to learn placeholder names (D2), it carries each prompt body:
  4,552 / 6,388 / 4,266 / 3,505 / 5,770 B.
- implement.md:1039 has the per-task reviewer (a subagent) load all of primary for
  `## Calibration` (D11).

**Can the parent fill placeholders without reading the template? Yes.**

- **Neutral route.** Put `[STANDARDS_PATHS]` and the unique parts of `[PRINCIPLES_PATH]` in
  review-panel.md **Principles** (R6, R7), and define the other six placeholders there too.
- **MECHANICS route: `render-slot-prompt.sh`.**
  - Arguments: `<slot-id> <round> <canonical-worktree> <changeRoot> <principles-path>
    [<standards>…]`.
  - It extracts the fenced prompt body and substitutes every placeholder from values the parent
    already computes: diff/delta path, changeRoot artifacts, per-worktree bundle paths, throwaway
    copies, principles path, and standards resolved per project-configuration.md:43-84.
  - It writes `<abs-worktree>/.superpowers/sdd/slot-prompt-<round>-<id>.md` and prints the path.
  - Exit codes: 0 written; 1 a placeholder left unresolved, named; 2 cannot answer.
  - Parity tests: the rendered file equals the template body with substitutions, and every
    template placeholder has a resolver.
  - `[GLOBAL_CONSTRAINTS]` is the one judgment placeholder. It stays parent-filled or is dropped
    by decision.
  - The dispatch prompt then carries the baseline pointer, the MODEL HANDSHAKE and "read
    <rendered file>". The shared paragraphs can be rendered from their pinned blocks too.

### Q3. agent-baseline.md

- **Rationale:**
  - :3-5 second sentence (A1, 165 B);
  - :15-17 two sentences after "Unconditionally…" (A2, 204 B);
  - :54-56 last sentence (A4, 170 B);
  - :41 first sentence is mechanism, but it is the referent of :43-45's instruction, so keep it.
- **Duplicate within the file:**
  - :19 "This is enforced…" duplicates :38 (A3, 90 B);
  - :60 first sentence duplicates :32 (A5, low).
- **Instruction, keep:**
  - :9-13 the two propagated sentences;
  - :15 "Unconditionally — not only when a rule looks relevant";
  - :20-21 the carve-outs;
  - :25-26 "read the linked file before acting";
  - :43-45 the dangling-pointer handling;
  - :49-54 project rules win, never read the other harness file, load the contract first;
  - :60-61 "state what you did not finish; say so rather than deciding".
- **Restated from the always-on rules:**
  - Every table row is a one-line restatement of an always-on rule's core. That is by design,
    but it is a duplicate for any subagent that also receives the managed block (D6, A6).
  - A2 is near-verbatim of dispatch-carries-the-baseline's core ("you cannot tell from the prompt
    whether the agent will end up in production…").
  - The :12-13 quote is verbatim from that core. It must stay: the hook matches it.

### Q4. flow-fast/SKILL.md: restated vs. its own overrides

**Restated from files the same session loads:**

- :87-89 restates implement.md §4 (FF3).
- :93-97 restates review-panel.md's decided-panel, re-run and fix routing (FF4).
- :107 restates jira-integration.md **Transitions** / **Never blocking** (FF5).
- :217-218 restates the always-on commit-scope core (FF6).
- :225-226 lint repeats the always-on lint-fix-priority, but adds the scope "on the files you
  touched". Keep.
- :31-32 "Never ask a model, planning-effort or review question" is restated by the stub
  (X1, :21-22).
- The stub's :9-19 restates the whole skill (X1).

**Its own overrides (not duplicates):**

- The substitution list (:62-70): `<changeRoot>`, `ff-` token, one worktree, merge-base from
  `origin/<default-branch>`, no render.
- Prohibitions: no spectre, no state file, no workspace isolation, no planner or verifier
  (:11-16, :29-35).
- Git-only worktree creation at kickoff (:119-133).
- Brainstorm asks only with `## handoff: required` (:141-148).
- Plan via the harness task list plus a `tasks.md` shape (:72-80); decide roll always runs
  (:81-85).
- `check-task-commit-fields.sh` not run and no ticks (:89, :220-225).
- A default panel runs no panel (:97).
- Self-review is `defer`-only (:256-282).
- Lint plus targeted tests as the close (:245-249).
- `## handoff` gate (:295-300); landing-route selection (:324-328); rebase before landing
  (:344-355).
- Route commands (:357-363); cleanup (:385-410); Jira Done (:420).

### Q5. Citers of every moved section, and guard side effects

- **Template intros and Placeholders.** No `**Heading** (`file`)` citer (grep over skills/,
  rules/, commands-claude/, root .md). Removal zeroes coverage: `check-references` for
  primary/principles/security once the intro goes, all five once Placeholders go too;
  `check-installed-citations` for all five once both go. Add entries to `crExpectedZero`
  (references.go:103) and `cicExpectedZero` (installedcitations.go:92-109).
- **principles `[STANDARDS_PATHS]` step.** Textual citer review-panel.md:669 must be retargeted.
  project-configuration.md:86-87 cites the standards-as-data clause, which stays.
  gather-dispatch-context.sh:23-27 cites review-panel.md's `[PRINCIPLES_PATH]` rule, which is
  unaffected.
- **principles `## Your Scope`.** No citer. The heading is inside the fence, which
  `check-references` ignores.
- **primary `## Calibration`** stays; cited by implement.md:1039.
- **agent-baseline.md sections.** No bold-heading citer. The pointer text is cited by
  be-brief.mdc:80, dispatch-carries-the-baseline.mdc:17-20, README.md:186/299 and
  hooks/enforce-agent-baseline.py:28-31, all unaffected. A new `rules/agent-baseline-rationale.md`
  is scanned by `check-references` (rules/ is a target) and needs a `crExpectedZero` entry.
  setup.sh does not install it, so the installed-rules and installed-citations guards never see it.
- **flow-fast/SKILL.md.** **5. Verify** is cited by project-configuration.md:34,
  finish-contract-run2.md:214 and flow-self-review/SKILL.md:12; not moved. `## Stage keys` is
  pinned by stats/internal/stages/names_test.go:323-330; not moved. FF7's destination,
  skills/flow-fast/SKILL-rationale.md, is declared expected-zero; the moved text has no bold, so
  the declaration holds.
- **Rationale moves into skills/flow/SKILL-rationale.md** (R1c clause, F5, E1). That file is not
  declared zero (coverage 3) and the moved text adds no bold-path pair.
- **`check-dispatch-paragraphs`** pins nothing in these files. `check-normative-inventory` is
  unaffected (no MUST/SHALL).
