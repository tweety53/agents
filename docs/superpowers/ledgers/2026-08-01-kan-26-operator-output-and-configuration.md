# SDD ledger — plan: openspec/changes/kan-26-operator-output-and-configuration/tasks.md

**Change:** kan-26-operator-output-and-configuration
**Worktree:** /Users/tweety53/Projects/agents-worktrees/openspec-kan-26-operator-output-and-configuration
**Merge base:** 1797b84c9574c279351457b73dd7ef5733cf06b7
**Git boundary:** MYFLOW — no per-task commits. Diffs are taken against a recorded TASK_BASE tree, not a commit range.

## Pre-flight scan

Controller-resolved ambiguity, carried into Task 11's dispatch:

- **Task 11 Step 3's positive control uses `git checkout -- skills/myflow-contracts/state-file.md` to
  revert.** Under myflow's no-commit boundary every earlier task's work is *uncommitted*, so that
  command would restore the file to its `HEAD`/index state and destroy Task 1's edits — and, through
  the guard re-run, mask the whole rename. The block is tagged `unverified:`, so it is a hypothesis
  rather than text to transcribe. Resolution: the implementer takes a byte-for-byte backup outside
  the repository, appends the literal, runs the guard, then restores from that backup and diffs to
  prove the restore was exact. `git checkout`/`git restore`/`git stash` are forbidden in this task.
  The control still proves exactly what the plan asked it to prove.

No other conflict found between tasks, or between a task and the Global Constraints.

## Surviving `effort` references — routing map (controller, after Task 2)

Taken from `grep -rn 'effort' skills rules commands commands-claude scripts README.md AGENTS.md CLAUDE.md`,
excluding `check-vocabulary.sh` and the two protected false positives. **No task's own verify grep
catches the prose `` `effort` `` form** — they match `"effort":` or `.effort` only — so each is routed
here by hand.

Correct as they stand (the retired-key rule and the new concept name; do **not** rename):
`state-self-heal.md:73,78` · `state-file.md:58,79,151,155,172` · `pipeline.md:67,91,427,992`

| Owner | Location | What must change |
|---|---|---|
| Task 4 | `myflow-start/SKILL.md:63` | heading `## Ask the effort — creating runs only` |
| Task 4 | `myflow-start/SKILL.md:72-75` | question text "How much effort should planning this change take?" **and its offered levels, which still read Medium/High/Low** — confirmed by Task 2's reviewer as a live mismatch against `state-file.md`'s `low`/`default`/`detailed` table |
| Task 4 | `myflow-start/SKILL.md:81` | **reads the retired `effort` field** — flagged by Task 2, **not** caught by Task 4's grep |
| Task 4 | `myflow-start/SKILL.md:184` | `"effort": "<low\|medium\|high, or null>"` — caught by Task 4's grep |
| Task 4 | `myflow-start/SKILL.md:232,233,243` | guardrail wording "effort level" |
| Task 5 | `myflow-do/SKILL.md:369` | carry-forward list — caught by Task 5's grep |
| Task 6 | `myflow-finish/SKILL.md:261` | carry-forward list — caught by Task 6's grep |
| Task 7 | `myflow-status/SKILL.md:43` | jq `.effort` — caught by Task 7's grep |
| Task 7 | `myflow-status/SKILL.md:96,101,103` | prose `` `effort` `` — **not** caught by Task 7's grep |
| Task 10 | `CLAUDE.md:106`, `AGENTS.md:152` | "Asks the effort level once", "records the effort" |

**Tasks 4, 7 and 10 must not re-rename the `**Planning effort**` citation** on those same lines —
Task 1 already fixed it to keep `check-references.sh` green.

## OPEN CROSS-TASK ISSUE — handoff block divergence (raised at Task 4)

`pipeline.md`'s `### The block each state renders` says the block is defined "here and nowhere else",
but the per-state templates do not match what the commands actually print:

- `STARTED` template: `Proposal artifact:`, `Jira:`, `Planning effort:`, IntelliJ, `Next:`
- `myflow-start/SKILL.md` actually prints: `**Change:**`, `**Artifact:**`, `**Decisions recorded:**`,
  `**Jira:**`, `**Jira description (pre-edit):**`, `**Planning effort:**`, `**Models:**`, IntelliJ, `Next:`

Labels differ and the skill carries four fields the template lacks. `myflow-handoff-output/spec.md`
requires one definition rendered by two commands, so on its face this is a violation. The plan
schedules **no** reconciliation. Task 7 (`/myflow-status`) renders from the template, so if this is
left, `/myflow-status` and `/myflow-start` print different blocks for the same state — the exact
drift the capability exists to prevent. The same divergence probably exists for `IN_PROGRESS`
(`/myflow-do`, `/myflow-finish`).

**Status: RESOLVED — scope authorised by the operator, scheduled as Task 12.**

Task 4's reviewer confirmed it is a genuine SSOT violation (the spec's per-state content table is
exact contents, not a floor), that **neither Task 2 nor Task 4 is at fault** — the divergence
predates the whole plan, since all three skills already printed richer bold-label blocks — and that
**no task in the 11-task plan was scoped to reconcile them**. A plan-breakdown gap, not an
implementation defect.

Reviewer also confirmed the divergence spans all three commands, and that `/myflow-finish` run 1's
block is a *fourth* shape carrying none of the `IN_PROGRESS` template's fields.

Put to the operator via AskUserQuestion; they chose **"Fix now, document the scope"**. Recorded in
`proposal.md` (What Changes) and `tasks.md` (**Task 12**, four steps, `pipeline.md` only). Task 12
runs **before Task 7**. Numbered 12 rather than inserted so no earlier task's number moves.
Controller addition beyond the reviewer's proposal: run 1 needs its **own** `IN_PROGRESS` template —
one block cannot be correct for both a diff awaiting review and a branch awaiting a merge, which
would make the requirement unsatisfiable rather than merely unsatisfied.

## CARRY INTO TASK 7 — two `IN_PROGRESS` splitting signals to reconcile

Raised by Task 12's reviewer. `/myflow-status` will end up with **two different splits of the same
state**, and Task 7 must make them coherent:

- `skills/myflow-status/SKILL.md:110-112` already splits its **table** view on **merge status**
  (`branch not merged` / `branch merged`), because `/myflow-finish` integrates before the merge and
  archives after.
- Task 12 split the **detail block** on **`prUrl`**, per the operator-authorised design direction.

They are orthogonal, not contradictory — merge status separates "archive next" from everything else;
`prUrl` separates "waiting on a merge" from "waiting on a diff review". But one command carrying two
signals for one state needs stating deliberately, not by accident.

**Known limitation to carry, not to hide:** `prUrl` is one-way. Only the pull-request route records
it, so non-null proves run 1 happened but null proves nothing — wrong on "merge and push" and
"handle it manually". The reviewer priced the cost honestly: on those routes `/myflow-status` would
offer `git diff --cached` for a branch already committed and pushed, and that command returns
nothing, which is confusing rather than merely sparse.

**Complete alternative, flagged for Task 7's designer and deliberately not adopted in Task 12:** all
three landing routes push the branch as their first action and nothing before run 1 ever pushes, so
a remote-ref check (`git ls-remote` / `origin/openspec/<name>`) discriminates perfectly. Its cost is
a network call from a read-only command the contract elsewhere works hard to keep cheap — a real
tradeoff, not a free win. Task 12's brief mandated `prUrl` verbatim, so this was never the
implementer's call to make.

## Progress

Task 1: implementer dispatched (model: opus) — returned DONE_WITH_CONCERNS.
Task 1: base tree 8b9cab5, head tree dbd972a; diff `.superpowers/sdd/task-1.diff` (6 files, +69/-39).
Task 1: scope escape accepted by controller — renaming `## Effort` broke four live
  `check-references.sh` citations (`skills/myflow-start/SKILL.md:78`, `skills/myflow-status/SKILL.md:98`,
  `CLAUDE.md:106`, `AGENTS.md:152`). One bold-token rename per line, verified minimal by reading the
  diff. The plan's Task 1 Files list omitted them; the guards-green constraint forced the edit.
  **Tasks 4, 7 and 10 rewrite those same lines and must not re-rename the citation.**
Task 1: task reviewer dispatched (model: sonnet) — spec ✅, quality Approved, 0 Critical, 0 Important.
  Reviewer independently reproduced the forced-scope claim: reverting the `AGENTS.md` token rename
  alone reproduces the predicted `check-references.sh` failure. Guards re-run clean by the reviewer.
Task 1: minor (deferred): task-1-report overstates why `myflow-start/SKILL.md:78` was rewrapped —
  the bold token and path were already on one physical line; the reflow was cosmetic. Diff is fine.
Task 1: minor (deferred): task-1-report says "five call sites" then lists six. Count nit, no diff effect.
Task 1: complete (uncommitted, review clean, model: opus; reviewer model: sonnet). 6/6 checkboxes ticked.

Task 2: implementer dispatched (model: opus) — returned DONE.
Task 2: base tree 637b5b8, head tree 8b5995f; diff `.superpowers/sdd/task-2.diff` (pipeline.md, +310/-0).
Task 2: task reviewer dispatched (model: sonnet) — spec ✅ with one Important exception,
  quality Approved. Reviewer independently confirmed: no tuned threshold copied into pipeline.md,
  diagram byte-identical to README's, all eight expansions present, guards clean, no commits.
Task 2: fix round 1/5 dispatched (resumed original implementer, model: opus) — 2 findings.
  F-Important `pipeline.md:69-71`: level-1 table gives `/myflow-finish` three rows against the spec's
  and the brief's "one row per command"; `/myflow-status`+`/myflow-info` share a row (converse
  deviation). Not a plan conflict — the plan's own Step 1 says one row per command.
  F-Minor `pipeline.md:338`: "the change's task list" for the spec's literal `tasks.md`. Promoted into
  this round rather than deferred, because the same section declares the harness's task list a
  "view, never a record" — the ambiguity sits exactly on the distinction the capability is built on.
Task 2: fix round 1/5 (2 addressed, 0 open; fix base 8b5995f → 1b3ec4f, diff
  `.superpowers/sdd/task-2-fix-round-1.diff`, +22/-19, pipeline.md only).
  Scoped re-reviewer (model: sonnet) confirmed both ADDRESSED, no new breakage, four interface
  headings intact, new `**Command surface**` citation resolves, guards 0, no commits. Scope proven
  with `git apply --check --reverse`. Implementer withdrew its own three-row justification.
Task 2: minor (deferred): `pipeline.md` level-1 intro says the five commands are named "exactly as
  **Command surface** below names them", but that section gives counts rather than listing all five.
  Wording looseness, not incorrect.
Task 2: minor (deferred): brief Step 7's grep `panelFix|reviewPanel|implementation` is
  non-discriminating — `implementation` matches ~15 lines of ordinary prose. Affects the plan's
  verification text, not the delivered file.
Task 2: note for Tasks 8 and 10: `▸` is a new in-file marker convention, scoped to `## Pipeline flow`.
Task 2: complete (uncommitted, review clean after 1 fix round, model: opus; reviewer/re-reviewer: sonnet).
  7/7 checkboxes ticked.

Task 3: implementer dispatched (model: opus) — returned DONE_WITH_CONCERNS (5 minor judgment calls).
Task 3: base tree d397b3a, head tree 6c0c638; diff `.superpowers/sdd/task-3.diff`
  (jira-integration.md, +66/-0 — pure addition, no existing line altered).
Task 3: task reviewer dispatched (model: sonnet) — spec ✅, quality Approved, 0 Critical, 0 Important.
  All five flagged judgment calls upheld. The one worth recording: naming the Atlassian MCP tool
  `searchJiraIssuesUsingJql` does **not** conflict with this change's own
  `myflow-progress-visibility` spec forbidding a named harness tool — Atlassian MCP is
  protocol-level and reachable from Claude Code, Cursor and Codex alike, whereas a task-list API
  differs per harness. Precedent confirmed: `createJiraIssue` is already named at
  jira-integration.md:209. Reviewer verified all five new citations resolve and that the
  "does not reopen Unrecognised statuses" paragraph is genuinely consistent, not merely asserted.
Task 3: minor (deferred): task-3-report claims the dated heading mirrors Description sync's style;
  the mirroring is actually inverted. Report-accuracy nit, file text is correct.
Task 3: minor (deferred): repeated joins from the same `<KEY>` on different days yield multiple
  `## From <KEY> — YYYY-MM-DD` headings rather than one persistent heading. Spec is silent; edge
  case flagged for a future reviewer.
Task 3: complete (uncommitted, review clean, model: opus; reviewer model: sonnet). 5/5 checkboxes ticked.

Task 4: implementer (model: opus) → DONE_WITH_CONCERNS; reviewer (model: sonnet) → spec ✅,
  quality Approved, 0 Critical. The one Important is the handoff SSOT gap above — not attributable
  to Task 4, which could not have closed it inside its one-file scope. Parked and routed to Task 12.
  Base 6c0c638 → a78da86, diff `.superpowers/sdd/task-4.diff` (myflow-start/SKILL.md, +63/-19).
Task 4: implementer found two stale `Effort`-capitalised references invisible to a case-sensitive
  grep. **Later tasks must sweep with `grep -in`.**
Task 4: minor (deferred): section heading widened to "Ask the planning effort and the models" —
  disclosed and justified, the section now asks four questions.
Task 4: minor (deferred): the `**Models:**` handoff line has no "planned at <default>" fallback
  framing analogous to the Planning effort line. No spec requires one.
Task 4: complete (uncommitted, 1 Important parked with ruling + routed to Task 12, model: opus).
  7/7 checkboxes ticked.

Task 5: implementer (model: opus) → DONE_WITH_CONCERNS; reviewer (model: sonnet) → spec ✅,
  quality Approved, 0 Critical, 0 Important.
  Base a78da86 → head; diff `.superpowers/sdd/task-5.diff` (myflow-do/SKILL.md, +73/-21).
  Reviewer upheld both scope judgments: the two guardrail edits were necessary (leaving "stay on
  Sonnet" absolute would self-contradict the new `models.reviewPanel` resolution two paragraphs
  away), and editing the slot table's Model column was correct — roster, Required and Spawn columns
  and the whole trigger table verified byte-identical.
Task 5: minor (deferred): task-5-report calls the section-4 clause edit out-of-scope; it is actually
  inside one of the five enumerated locations. Overcautious self-report, not a defect.
Task 5: reviewer finding routed to Task 12 as its Step 4 — `pipeline.md`'s level-2 panel expansion
  still states Sonnet as an absolute while the same file's `## Model policy` now carries the
  override. One file disagreeing with itself.
Task 5: complete (uncommitted, review clean, model: opus; reviewer model: sonnet). 6/6 ticked.

Task 12: added to `tasks.md` (5 steps) and `proposal.md` after operator authorisation. Brief at
  `.superpowers/sdd/tasks/task-12-brief.md`. Base tree 0739165. Implementer dispatched (model: opus).
  Controller design direction given: template is the definition and adopts the commands' bold-label
  style; `Jira description (pre-edit)` named **run-only** (structurally unregeneratable, distinct
  from the spec's "reported as missing"); `IN_PROGRESS` gets a second template for finish run 1; and
  the `prUrl` discriminator must be stated with its real limitation — it is `null` on the
  "merge and push" and "handle it manually" routes, so it cannot cleanly separate the two renderings.
Task 12: implementer → DONE_WITH_CONCERNS. Base 0739165 → c7f4979; diff `.superpowers/sdd/task-12.diff`
  (pipeline.md, +111/-17). Field-by-field mismatch list across the five emitted blocks: 31 → 0.
Task 12: reviewer (model: sonnet) → **spec ❌ Critical**, quality Approved. The delta spec's single
  `IN_PROGRESS` contents row contradicted the two-rendering split, at SHALL level. Reviewer
  independently reverified the 31→0 claim against the live files rather than recomputing arithmetic,
  and supplied exact rewording.
Task 12: fix round 1/5 (1 addressed, 0 open; fix base c7f4979 → head, diff
  `.superpowers/sdd/task-12-fix-round-1.diff`). Scope widened to `specs/myflow-handoff-output/spec.md`.
  Re-reviewer confirmed ADDRESSED row by row against pipeline.md, no new breakage, pipeline.md
  untouched that round, no commits.
Task 12: implementer corrected **my brief** — no lint guard scans `openspec/specs/`
  (`check-plan-provenance.py` `SCANNED_FILENAMES = ("tasks.md","design.md","proposal.md")`; the other
  two exclude `openspec/`). Verified. So "three guards green" was never evidence about that file;
  the real evidence is `openspec validate --strict` plus a row-by-row read.
Task 12: implementer added a normative `SHALL` beyond the wording I supplied, and it was necessary —
  splitting the table exposed a second contradiction: the requirement's existing "a value the state
  file does not carry SHALL be reported as missing" would have mandated `Route: missing` on every
  regenerated run-1 block, the exact lie the run-only concept exists to prevent. Re-reviewer judged
  the carve-out correctly scoped (an on-disk-existence test, not "absent in this instance") and the
  enumeration kept in one place.
Task 12: controller fixed two bookkeeping items in `tasks.md` afterwards — the Files list now names
  the spec file, and the divergence table's run-1 row now includes `Change`.
Task 12: complete (uncommitted, review clean after 1 fix round, model: opus; reviewer/re-reviewer:
  sonnet). 6/6 checkboxes ticked.

Task 6: implementer (model: opus) → DONE; reviewer (model: sonnet) → spec ✅, quality Approved,
  0 Critical, 0 Important. Base 7816586 → head; diff `.superpowers/sdd/task-6.diff`
  (myflow-finish/SKILL.md, +26/-6).
Task 6: **validated Task 12's premise.** Both implementer and reviewer independently compared the
  run-1 handoff block against Task 12's corrected template field by field and found full agreement —
  no skill edit needed. That was the theory Task 12 was built on (correct the contract, the skills
  already conform), and it now has evidence behind it rather than assertion.
Task 6: implementer corrected the brief's framing — the enumerated carry-forward list at old line 261
  is **run 2's** step 7, not run 1's. Run 1's write is generic ("every other field carried forward")
  and correctly needed no edit. Reviewer confirmed.
Task 6: reviewer upheld dropping the `**Labels on issues the pipeline creates**` citation in favour of
  `**Follow-up issues**` — the latter is a strict superset, stating the join-path label union directly
  and reaching the create-path rule via its own onward pointer, which the reviewer verified exists.
Task 6: minor (deferred): task-6-report's self-reported diff stat is off by one line each way.
Task 6: minor (deferred): task-6-report justifies a marker placement with a wrong claim about
  `check-references.sh` — the guard pairs bold tokens only against `\.mdc?$` paths, so a `.sh` path is
  never checked at all. Placement harmless, reasoning wrong.
Task 6: complete (uncommitted, review clean, model: opus; reviewer model: sonnet). 4/4 ticked.

Task 7: implementer (model: opus) → DONE_WITH_CONCERNS; reviewer (model: sonnet) → spec ✅, quality
  Approved, 0 Critical, 0 Important. Base 171c973 → head; diff `.superpowers/sdd/task-7.diff`
  (myflow-status/SKILL.md, +43/-8). Reviewer traced all four merge-status × `prUrl` combinations and
  confirmed the two-splits reconciliation is accurate and reads as design, not as a bug.
Task 7: **guard blind spot found — repo-wide, and worth keeping.** `check-references.sh` associates a
  bold token with a path only on the same *physical* line, so a **wrapped citation is silently
  skipped, not failed**. A green guard is not evidence a given citation was checked. Established by
  the implementer's probe and independently re-established by the reviewer, which mutated all six
  section names in a sandbox and confirmed all 15 citations in that file are genuinely examined.
  Controller swept all eight modified files for both wrap patterns: 4 hits, all benign (emphasis
  bold, or a non-section path). **No wrapped-citation defect survives in this change.**
Task 7: minor (deferred): `not recorded — planned at default` is a fourth site naming the recommended
  level. Brief-mandated; justified as the rendered output line, which both renderers must match.
  Nothing mechanical re-checks it if the level is ever renamed.
Task 7: **FOLLOW-UP, deferred on the reviewer's explicit recommendation** — `state-file.md` and
  `state-self-heal.md` carve out an absent `models` **object** as legal but say nothing about a
  **partial** one (an absent sub-key). Real gap. Deferred because no writer in this repository ever
  produces a partial object — `/myflow-start` writes all three keys, each nullable but never absent —
  so it has no live path through the pipeline's own commands. The implementer drafted a rule and then
  deleted it rather than invent contract law inside a consumer file. Goes in the guide's
  `## Known incomplete`.
Task 7: complete (uncommitted, review clean, model: opus; reviewer model: sonnet). 5/5 ticked.

Task 8: implementer (model: opus) → DONE; reviewer (model: sonnet) → spec ✅, quality Approved,
  **zero issues at any severity**. Base 0c7f2ca → head; diff `.superpowers/sdd/task-8.diff`
  (myflow-info/SKILL.md, +12/-0). Both citation lines proved examined by the implementer's sandbox
  mutation probe and again by the reviewer's independent one.
Task 8: reviewer upheld leaving the file's pre-existing remembered prose alone — the spec's SSOT
  requirement is scoped textually to the diagram and the stage table, not to the three-line state
  shape or to prose about landing routes, and widening it would invent scope.
Task 8: **FOLLOW-UP, pre-existing, outside this change's spec** — the three-line pipeline text shape
  has drifted. `pipeline.md:20` (canonical) reads `terminal (second run — see the finish contract)`;
  four copies — `CLAUDE.md:73`, `AGENTS.md:119`, `README.md:71`, `myflow-info/SKILL.md:47` — still
  read `terminal (it integrates on its first run)`, which misdescribes the two-run finish. Task 10
  deletes the README's copy as a side effect but owns none of the others, and no task owns this
  wording. **Not absorbed** — this change already took one authorised scope addition (Task 12), and a
  second unprompted one is the operator's call. Goes in the guide's `## Known incomplete`.
Task 8: complete (uncommitted, review clean, model: opus; reviewer model: sonnet). 3/3 ticked.

Task 9: implementer (model: opus) → DONE_WITH_CONCERNS. **Genuine no-op — nothing changed in either
  command tree.** No subagent review dispatched: the diff is empty, so there is nothing for a
  reviewer to read, and the substance is the sweep evidence. **Controller verified all four claims
  directly instead**: `grep -rni 'effort'` empty, `grep -rniE '\b(medium|high)\b'` empty, no change
  since base tree 0b5730e, and neither tree touched by the whole change.
Task 9: the brief's `verified:` Step 1 claim still holds after eight tasks, and structurally rather
  than by luck — command files describe state transitions and delegation, never the questions a skill
  asks at runtime, so the rename had no surface in either tree to reach.
Task 9: two-tree agreement confirmed — all five paired files identical in normalised body text; the
  only differences are the by-design harness adaptations (`commands-claude` uses `model:` frontmatter,
  `commands` an equivalent prose paragraph, Cursor having no such field).
Task 9: implementer adjudicated and left `commands/myflow-do.md:8` alone (it asserts Opus implementers
  / Sonnet panel). Checked the merge base first and found the "explicit operator instruction overrides
  either default" paragraph already existed pre-change — so the sentence was always a default under an
  override regime, and this change makes the override durable rather than inverting it.
Task 9: minor (deferred): both `myflow-status.md` files now under-describe the report by not
  mentioning handoff regeneration. An omission, not a contradiction — the command-surface spec
  requires description agreement and matching accepted states, both of which hold.
Task 9: correction to the controller's own framing — `/myflow-status`'s handoff block is **additive**,
  not a replacement. The skill says "**Then** regenerate…", the next-command column survives, and the
  no-argument report is unchanged. The trees' existing description therefore remains true.
Task 9: complete (uncommitted, no-op verified by controller, model: opus). 3/3 ticked.

Task 10: implementer (model: opus) → DONE_WITH_CONCERNS; reviewer (model: sonnet) → spec ✅, quality
  Approved, 0 Critical, 0 Important. Base c669537 → head; diff `.superpowers/sdd/task-10.diff`
  (README.md, CLAUDE.md, AGENTS.md; +4/-21). Two instruction rows verified byte-identical.
Task 10: **deviation from the controller's dispatch, and the reviewer strengthened it.** I told it to
  cite `openspec/specs/myflow-model-policy/spec.md`; it cited `**Model policy**
  (`skills/myflow-contracts/pipeline.md`)` instead, because that spec's headings are all
  `### Requirement: …` form so no bold token can bind. The reviewer's own sandbox probe showed the
  instructed citation would have been **checked and permanently failing** (exit 1), not merely
  unchecked — a stronger justification than the implementer gave. My instruction was wrong.
Task 10: fifth copy of the drifted text shape found at `skills/README.md:10`, beyond the three
  enumerated. Left untouched, as instructed.
Task 10: complete (uncommitted, review clean, model: opus; reviewer model: sonnet). 3/3 ticked.

Task 11: implementer (model: opus) → DONE; reviewer (model: sonnet) → spec ✅, quality Approved,
  0 Critical, 0 Important. Base 82705c4 → head; diff `.superpowers/sdd/task-11.diff`
  (check-vocabulary.sh, +14/-0).
Task 11: **the destructive step flagged in this ledger's pre-flight scan was real and was avoided.**
  The implementer ran no `git checkout`/`restore`/`stash`/`reset`, probing against a scratchpad copy
  instead. Positive control proven both directions and independently reproduced by the reviewer:
  RED — the pre-change guard misses a reintroduced `"effort":` (exit 0); GREEN — the new guard
  catches it at `state-file.md:175` (exit 1). Reviewer also re-ran all seven harnesses itself.
Task 11: minor (deferred): the new marker's `#` sits at column 75, matching 11 of 13 pattern lines but
  not its two immediate neighbours at column 79. Cosmetic.
Task 11: complete (uncommitted, review clean, model: opus; reviewer model: sonnet). 5/5 ticked.

**ALL 12 TASKS COMPLETE — 60/60 checkboxes ticked, 0 unchecked.**

## Final review panel

Implementation diff `.superpowers/sdd/final-review.diff` — 730 changed lines, 13 files, taken from
merge base 1797b84 with the three planning paths excluded (matching what will be staged).

**Roster: full — all three required slots plus all four conditional ones.** Every optional trigger
fired against the diff, and the selection is recorded rather than assumed:

| Slot | Included? | Why |
|---|---|---|
| 0 Primary | required | — |
| 1 Bug hunt | required | — |
| 2 Principles (Merged) | required | — |
| 3 Security | **yes** | the Jira join reads externally-authored issue titles/descriptions, then appends to that issue and unions its labels — an attacker-influenceable path feeding an agent; plus a shell-guard edit |
| 4 Adversarial | **yes** | 730 changed lines (trigger is >~300); plus a state-file key rename with a compatibility read rule, which is a migration in all but name |
| 5 Lens B (simplicity & state) | **yes** | >~200 changed lines |
| 5 Lens C (robustness & ops) | **yes** | external integration (Jira) and a state-schema change |

**Harness deviation, recorded:** this harness exposes no `bugbot` or `security-review` agent type, so
slots 1 and 3 were dispatched as general-purpose agents running the prompts the skill itself ships
for that purpose (`bug-hunter-reviewer-prompt.md`, `security-reviewer-prompt.md`). Because the model
is therefore known rather than agent-defined, the ledger records **sonnet** for those two slots rather
than `unknown (agent-defined)`.

**Every slot on sonnet.** `[PRINCIPLES_PATH]` resolved to
`/Users/tweety53/.claude/skills/myflow-do/engineering-principles.md` (verified present; passed as a
path, never pasted). `[STANDARDS_PATHS]` resolved from `.myflow/project.md`'s `## standards` — both
entries are form 2 (bare filename, no `/`), resolving to the worktree root and passing containment:
`<worktree>/CLAUDE.md` and `<worktree>/AGENTS.md`, both verified to exist. No entry was dropped.


## Run stopped at the panel handback — NOT a clean handoff

**12/12 tasks complete, 60/60 checkboxes.** Lint (3 guards) and tests (7 harnesses) all exit 0.
Staged: 15 implementation files, +1449/-169. No commits. `openspec/` and `docs/manual-test/` left
unstaged per the git boundary.

**The review panel ran four full passes and three fix rounds. 48 findings: 37 fixed, 2 withdrawn by
the operator, 9 open (2 Critical).** The run was stopped before a fourth fix round.

**Panel model policy:** every slot on sonnet across all four passes; implementer and fix-wave
subagents on opus. This harness exposes no `bugbot` or `security-review` agent type, so those two
slots ran the prompts the skill ships for that purpose — model therefore known and recorded as
`sonnet`, not `unknown (agent-defined)`.

**Two operator decisions, both recorded in `design.md` and `proposal.md`:**
- Task 12 (handoff-template reconciliation) — added scope, authorised at the Task 4 handback.
- The `effort` compatibility apparatus — removed at the pass-2 handback, then **restored** at the
  pass-3 handback when the controller's supporting premise proved false.

**Two controller errors, recorded because they cost real work:**
1. **A false premise relayed to the operator.** `design.md`'s measurement ("two `FINISHED` files")
   was summarised as fact without re-running it. Five files carry the key; four non-null; one is
   this change's own open file. The operator decided on it, and pass 3 had to catch it.
2. **A Critical dropped from this record.** Pass 3's partial-join run1/run2 finding was folded into
   round 3's brief as prose instead of being given an ID. It went unfixed for a round and returned
   as F40.

**State file written with care:** it carried `effort: medium` with a live `artifactUrl` and
`jiraIssue`. Written as `planningEffort: default` with both preserved; `models` omitted, since
`/myflow-start` never asked for it and absent is the contract's "not recorded". Backup at
`<scratchpad>/state-backup.json`.
