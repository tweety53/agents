# Planning session — slimming analysis

> **Audit snapshot at `700e184` (2026-09-29).** A read-only model pass that followed [`rubric.md`](rubric.md); every row is a candidate to verify, not a verified fact. Line numbers refer to `700e184`. Main has since retired the bugbot and security slots (`ebdfdde1`..`4dafab4a`), so re-locate each row by its quoted first words, and treat bugbot/security rows as obsolete. Tracked in KAN-851 → KAN-856 (and the Drift items in KAN-853 / KAN-854). Plan: [`README.md`](README.md).

Scope: the PLANNING session's own load set — `skills/flow/brainstorm.md` (16,982 B),
`skills/flow/brainstorm-planner.md` (44,477 B), `skills/flow-contracts/jira-integration.md` (15,037 B),
`plan-provenance.md` (5,833 B), `build-green.md` (5,942 B), `handoff-blocks.md` (15,945 B) — 104,216 B.
Read-only; nothing in the repository was modified.

Method: whole-line spans measured with `sed -n 'A,Bp' FILE | wc -c`; spans marked *(sub)* are an exact
sub-line substring measured in UTF-8 bytes (start and end phrase located and verified unique). Line
numbers are those of HEAD `700e184`.

## Totals

Conservative: **high and med confidence only**; low-confidence rows are listed in the Candidates
table but excluded here. Each byte is counted under one lever only.

| file | bytes | RATIONALE | DUPLICATE | STALE | LAZY-SPLIT | MECHANICS | MISPLACED |
|---|---|---|---|---|---|---|---|
| skills/flow/brainstorm.md | 16,982 | 907 | 2,325 | 62 | 6,005 | 2,163 | 376 |
| skills/flow/brainstorm-planner.md | 44,477 | 813 | 1,000 | 351 | 5,298 | 5,224 | 0 |
| skills/flow-contracts/jira-integration.md | 15,037 | 1,295 | 498 | 0 | 2,845 | 0 | 0 |
| skills/flow-contracts/plan-provenance.md | 5,833 | 547 | 0 | 0 | 0 | 0 | 1,003 |
| skills/flow-contracts/build-green.md | 5,942 | 810 | 416 | 0 | 0 | 0 | 0 |
| skills/flow-contracts/handoff-blocks.md | 15,945 | 0 | 0 | 0 | 0 | 0 | 14,637 |
| **total** | **104,216** | **4,372** | **4,239** | **413** | **14,148** | **7,387** | **16,016** |

What a typical planning run would actually stop paying for (every lazy-split condition false,
no code change): **≈ 39.9 KB**. That is handoff-blocks.md dropped from the load entirely (−15,945 B)
with the 604-byte STARTED template carried in brainstorm.md instead (net −15.3 KB), plus
LAZY-SPLIT 14.1 KB, RATIONALE 4.4 KB, DUPLICATE 4.2 KB, the other MISPLACED 1.4 KB and STALE 0.4 KB.
MECHANICS adds 7.4 KB more, but needs code and parity tests.

Outside this session: splitting out **Resuming at `STARTED`** (BR-08) would also save the
IMPLEMENTATION session **≈ 14.9 KB**. Every post-`/clear` implementation entry is routed through
brainstorm.md today (Drift #8).

## Candidates

`(sub)` = sub-line span. Rows marked *low* are excluded from Totals.

| id | file:startline-endline | bytes | lever | condition or canonical location or evidence | first ~10 words verbatim | confidence | behaviour-risk note |
|---|---|---|---|---|---|---|---|
| BR-01 | brainstorm.md:24 (sub) | 67 | DUPLICATE | canonical jira-integration.md:58-59 ("When only a key is supplied, derive the slug from the issue summary"). That file is loaded at brainstorm.md:9 | "Derive the slug from the issue summary when only a key was given." | high | none |
| BR-02 | brainstorm.md:30-31 (sub) | 156 | DUPLICATE | canonical pipeline.md:477-480: one match → use and announce; several → AskUserQuestion (name, state, last modified); zero → "a creating `/flow` run asks what to build" | "Exactly one match → resume it, announcing which; multiple → **AskUserQuestion**" | med | keep the preceding "restricted to changes with incomplete planning artifacts". pipeline.md does not state it |
| BR-03 | brainstorm.md:61 (sub) | 151 | RATIONALE | the reason for the null fields. Already verbatim, with decision ids, in SKILL-rationale.md:54-58; the guardrail itself is SKILL.md:204-205 | ": `/flow` asks no planning-effort or model question on a creating run, and" | high | none. "written `null` and stay `null` for the life of the change" stays |
| BR-04 | brainstorm.md:61-62 (sub) | 70 | DUPLICATE | canonical SKILL.md:207-208 ("`artifactUrl` is written `null` and stays `null` for the life of the change") | "`artifactUrl` stays `null` — `/flow` publishes no proposal artifact." | high | none |
| BR-05 | brainstorm.md:65-67 (sub) | 144 | RATIONALE | why STARTED is written first. The ordering rule (38-39, 64-65) stays. Target: SKILL-rationale.md "## brainstorm.md — A." | "The operator sees `STARTED` recorded the moment they invoke `/flow`, whether or" | high | none |
| BR-06 | brainstorm.md:88-92 (sub) | 395 | RATIONALE | why the worktree is persisted before the step returns. The imperative at 84-88 stays. Target: SKILL-rationale.md "## brainstorm.md — the worktree is created inside `flow.kickoff`" | "The record written at **A** carries `"worktrees": {}`, and the next durable" | med | reads as a "must not", but only restates the consequence the 84-88 imperative already prevents |
| BR-07 | brainstorm.md:98-99 (sub) | 110 | RATIONALE | why a failed worktree-setup command ends the turn | "— a worktree that cannot be set up fails `flow.verify` later anyway" | high | none |
| BR-08 | brainstorm.md:108-130 + 257-261 (sub; the resume half of "Resume and fix runs") | 1,851 | LAZY-SPLIT | load when `flow state get` says STARTED (the router already branches there, SKILL.md:140-141), or when A's key-prefixed candidate lookup resumes an existing change (brainstorm.md:18-22). False on a fresh creating run, the usual planning session. **True on every implementation-session entry** | "### Resuming at `STARTED`" / "A run finding `"state": "STARTED"` already recorded is resuming" | high | verbatim move, and route SKILL.md's STARTED bullet to the new file. The moved half still carries the stale `design.md` pointer (Drift #1) |
| BR-09 | brainstorm.md:132-192 | 4,154 | LAZY-SPLIT | load when the reachability check ended the run (brainstorm-planner.md:24-26) or on a planless STARTED resume (`total == 0`, brainstorm.md:119-123). False in nearly every run; the route's first real instance is the still-open `withdraw-changes-abandoned-before-planning` change | "### The withdrawal route" / "A change abandoned before planning — nothing implemented, nothing merged" | high | put a one-line load directive at both offer points. The ask wording moves verbatim. Pairs with BP-02 in one file |
| BR-10 | brainstorm.md:196-200 | 438 | DUPLICATE | canonical brainstorm-planner.md:3-6 ("no dispatched subagent, no relay. Every "you" below addresses that session directly") and brainstorm.md:206-207, which is the load-and-follow directive | "Sections **B**, **C** and **D** of `skills/flow/brainstorm-planner.md` are the running" | med | "on its own model" is implied by "no dispatched subagent". Keep 206-209 |
| BR-11 | brainstorm.md:211-213 | 298 | DUPLICATE | canonical brainstorm-planner.md:52-56 (batched into one AskUserQuestion call) and 104-108 / 136-139 (the confirm and the offer as named-option asks) | "**Questions are the session's own direct `AskUserQuestion` calls**, batched per" | med | none |
| BR-12 | brainstorm.md:218-219 (sub) | 107 | RATIONALE | why the summary prose is shown. The imperative (215-218) stays | "The operator approving or answering the question is approving against the summary" | high | none |
| BR-13 | brainstorm.md:221-229 | 492 | DUPLICATE | canonical brainstorm-planner.md:152-161: the same three commands. brainstorm-planner.md:10-13 names Convergence as their home | "Mark `flow.brainstorm` end and `flow.design-approval` begin/end around the merged" | high | none. The copies agree; only the `#` comment differs |
| BR-14 | brainstorm.md:231-236 | 545 | DUPLICATE | canonical brainstorm-planner.md:165-167 (create-artifacts begin), 240-242 (end), 246-248 (writing-plans begin), 613-615 (end), 538 (the decision.json path), 620-621 ("no answer, no commit") | "After the `flow.design-approval` mark above closes, mark `flow.create-artifacts` begin and" | med | the marks come in the same order in both files |
| BR-15 | brainstorm.md:239 (sub) | 62 | STALE | `design.md` resolves to no file at the repo root or in skills/flow/. The only heading "The `## Decision` block" is spectre/changes/archive/kan-472-…/design.md:119, which shows an obsolete one-table shape; the live shape is brainstorm-planner.md:562-584. check-references skips paths that do not resolve | "— the shape **The `## Decision` block** (`design.md`) shows" | med | cutting leaves "then run the record sequence that section states" pointing at nothing. The fix needs the dispatcher's approval (Drift #1) |
| BR-16 | brainstorm.md:248-251 (sub) | 259 | DUPLICATE | canonical brainstorm-planner.md:626-638: summary, block, "Proceed to implementation?" Yes / No, No → revise and re-decide, "in `/flow`, the `flow.decide` record sequence again" | "— the prose summary of the logic to be implemented, the `## Decision` block," | med | keep "**The gate comes next**: **Plan review gate** (…)" and the On-Yes sentence |
| BR-17 | brainstorm.md:261-265 (sub) | 376 | MISPLACED | the fix-run half. A fix run never loads brainstorm.md: SKILL.md:142-145 routes it to implement.md §3, and implement.md:336-452 never runs `plan-class.sh` or the Decide step | "; a fix run's `flow.document-fix` (`skills/flow/implement.md`) hands the appended plan" | med | moving it into implement.md §3 would **start** fix runs re-deciding, which is a behaviour change (Drift #6). The dispatcher decides |
| BR-M1 | brainstorm.md:75-102, minus BR-06 and BR-07 | 2,163 | MECHANICS | kickoff steps 1–5 become one `kickoff-worktree.sh <project> <name>` call; precedent: `prepare-workspace.sh` replaced implement-phase prose. It would also serve /flow-plan (flow-plan/SKILL.md:173-174) and implement.md §2's recipe for additional worktrees (implement.md:293-297) | "1. `check-worktree-location.sh <project>` — exit 1 or 2 stops the run" | med | needs code and parity tests. The exit contract must keep step 4's stop on a non-zero command and step 3's persist-before-return |
| BR-M2 | brainstorm.md:164-184 | 1,757 | MECHANICS | withdrawal steps 1–5 become one idempotent script | "1. Per worktree of the resolved set (**Resolving a change's worktrees**," | *low* | BR-09 (lazy split) is cheaper |
| BR-M3 | brainstorm.md:115-127 | 1,136 | MECHANICS | resume-point detection: `git worktree list`, then `spectre list --json`, then `check-plan-shape.sh` exit 0 as the "plan quality" test | "- **Does the worktree exist** — `git worktree list` naming `<project>/.worktrees/<name>`." | *low* | "writing-plans quality" is partly a judgment call |
| BP-01 | brainstorm-planner.md:10-13 | 286 | DUPLICATE | a placement note for editors ("stay exactly where they are"). The marks themselves are brainstorm.md:202-204 and brainstorm-planner.md:152-161 | "The `flow.brainstorm` begin mark lives in **Run brainstorming and planning directly**" | med | none |
| BP-02 | brainstorm-planner.md:17-28 | 1,044 | LAZY-SPLIT | load when the linked issue's labels carry `flow-fix` or `flow-cost`. A's `getJiraIssue` knows this before B opens. The archive has 5 of 109 changes named flow-fix/flow-cost (4 of the last ~15) | "**A run on a filed fix/cost finding verifies the defect still exists before planning.**" | high | pair with BR-09. The directive must fire "before any design question" |
| BP-03 | brainstorm-planner.md:40-47 | 652 | LAZY-SPLIT | load when the design is specified by a mockup frame | "- **A frame-specified design needs its handoff assets reachable from the tree.**" | *low* | detected mid-dialogue |
| BP-04 | brainstorm-planner.md:58-85 | 2,219 | LAZY-SPLIT | load when the request arrives as a seeded research note (a fully-worked issue, a capture or a handoff package). Decidable when the checklist opens; a minority of runs | "**The seeded-note path is legitimate, never a bypass to prevent.** When the ask" | med | includes 75-78 (the retired capture layout, which is RATIONALE) and 83-84 (Drift #14). Its Files clause duplicates 338-342 |
| BP-05 | brainstorm-planner.md:110-114, 124-131, 141-143 | ~1,400 | LAZY-SPLIT | these silence-default rules never fire under `## decisions: recommended` (116-118, 145-147) — this repo's own `.flow/project.md` setting | "End the stage only on an explicit choice of **approve the design and move on**." | *low* | other projects need them. The question and option names must stay loaded for the auto-resolution record |
| BP-06 | brainstorm-planner.md:236-238 | 277 | STALE | the STARTED handoff is printed when the plan gate answers Yes (brainstorm.md:251-254), after design.md exists, and its template counts decisions and open questions (handoff-blocks.md:68; handoff-blocks-rationale.md:18-31). The IN_PROGRESS block (verify-and-handoff.md:374-400) has no such count | "`STARTED` is written before this section exists (per **A** above), so no `STARTED`" | med | cutting removes a contradiction with the block the run actually prints. Confirm the intent (Drift #3) |
| BP-07 | brainstorm-planner.md:251-252 (sub) | 52 | DUPLICATE | lists writing-plans' own self-review items | "(spec coverage, placeholder scan, type consistency)" | *low* | none |
| BP-08 | brainstorm-planner.md:254-256 (sub) | 175 | DUPLICATE | canonical build-green.md:22-31 (Placement), which this paragraph itself calls "the rule in full". build-green.md is loaded in D (333) | "A task is a column-0 checkbox line, `- [ ] <n>. <title>`, whose `<n>`" | med | writing-plans is invoked at 250, before the load directive at 333. Keep the pointer sentence |
| BP-09 | brainstorm-planner.md:267-276 | 936 | LAZY-SPLIT | applies only to projects with a UI-test source set (Compose `desktopTest`) | "> **Write a feature's UI tests as their own follow-on task.**" | *low* | project-specific |
| BP-10 | brainstorm-planner.md:309-311 (sub) | 168 | DUPLICATE | canonical plan-provenance.md:8-11 and 15-20 (block and assumption tags), 60-62 (an `unverified` block stays) | ": code that cannot be verified is tagged `unverified:` and **kept**, and an" | med | none |
| BP-11 | brainstorm-planner.md:319-320 (sub) | 135 | RATIONALE | an analogy; no action follows from it | "— the same verify-the-premise move the filed-finding reachability check makes at checklist open" | high | none |
| BP-12 | brainstorm-planner.md:329-331 (sub) | 172 | RATIONALE | an incident narrative (KAN-750). Target: SKILL-rationale.md "### brainstorm-planner.md — D. Writing plans" | "KAN-750's task 11 was blocked for 35 minutes on a plan that assumed" | high | none |
| BP-13 | brainstorm-planner.md:356-357 (sub) | 111 | RATIONALE | the consequence of "each task's delta its own" | ", so two tasks dispatched in parallel count independently — neither's baseline depends" | med | none |
| BP-14 | brainstorm-planner.md:366-367 (sub) | 131 | RATIONALE | the consequence of "names no ref of its own" | ", so a command naming a ref measures that one tree at both points," | med | none |
| BP-15 | brainstorm-planner.md:373-374 | 122 | DUPLICATE | canonical build-green.md:14-18 (a red task carries `Squash-with`, naming the partner its commit lands in) | "A task tagged `Build: red` additionally carries `**Squash-with:** Task <N>`, naming" | high | none |
| BP-16 | brainstorm-planner.md:387-388 (sub) | 78 | RATIONALE | already in SKILL-rationale.md:159 | "Restated decision prose drifts from its entry the first time either is edited." | high | none |
| BP-17 | brainstorm-planner.md:391-392 (sub) | 122 | RATIONALE | already in SKILL-rationale.md:162, in its KAN-636 form | ": a citation with no blank line rides in the subject's continuation in every" | high | none |
| BP-18 | brainstorm-planner.md:408-409 (sub) | 64 | RATIONALE | the reason the header line is "required and explicit" | ", per this repository's "missing rather than dropped" convention" | med | none |
| BP-19 | brainstorm-planner.md:410 (sub) | 74 | STALE | "a script this change adds elsewhere": `scripts/generate-relocation-comparison.sh` exists and runs at review-panel.md:192 | "(generated later in the pipeline, by a script this change adds elsewhere)" | high | none |
| BP-20 | brainstorm-planner.md:450-455 + 475-490 | 2,035 | LAZY-SPLIT | load when Decide step 1 comes out `sdd` (class `big` after the override). 2 of 106 archived plans had ≥22 tasks, so ~2–5% of runs | "2. **implementer and fixer model + effort** — only when step 1 came out `sdd`" / "4. **implementer groups** — on every run whose step 1 came out `sdd`" | med | leave numbered stubs. flow-fast/SKILL.md:81-83 cites "Decide steps 1–4" and must name the new file |
| BP-21 | brainstorm-planner.md:529-534 (sub) | 586 | MECHANICS | the experimental prompt pick is `sorted(ls experimental/*.md)[experimental_roll mod count]` plus the line-1 description. `plan-class.sh` already prints the roll | "**Which prompt, when experimental rolled true:** `ls <agents repo>/skills/flow/experimental/*.md`, sorted," | med | needs code and a parity test. Keep the `none available` contract |
| BP-22 | brainstorm-planner.md:492-500 + 524-527 (sub) | 863 | MECHANICS | the tree table and the compact/experimental thresholds are a lookup on class and rolls. A script prints execution, roster, compact and static grouping | "**The tree**, one row per `class`:" / "Compact when `compact_roll < 90` (small, regular, big);" | med | the override comes after the script, so the script needs a class-override input. Keep 501-503 (micro never skips verify or self-review) |
| BP-23 | brainstorm-planner.md:559-611 (559-560 sub, 562-584, 586-603, 605-611) | 3,775 | MECHANICS | the `## Decision` block is a pure function of decision.json plus DEFAULT_MODEL, MODEL_SOURCE, REVIEWERS and the session's model. Replace with `flow decision render -file …`, or have `flow record decision` print it | "Print this exact shape as the run's own output once the Decide step completes" | med | the worked format moves into the renderer's tests. The gate prints its output verbatim |
| BP-24 | brainstorm-planner.md:538-559 | 2,004 | MECHANICS | the mechanical JSON fields (`classMechanical`, `inputs`, `rolls`, `resolved`) could be emitted by the script | "Write the decision JSON to `<abs-worktree>/.superpowers/sdd/decision.json` — on a first creating" | *low* | the definitions of the judgment fields stay |
| BP-25 | brainstorm-planner.md:643-645 | 250 | DUPLICATE | canonical brainstorm.md:251-255 | "What happens once this section's plan enrichment completes is stated in **Run brainstorming" | med | also wrong for /flow-plan readers (Drift #16) |
| JI-01 | jira-integration.md:79-82 (sub) | 314 | DUPLICATE | canonical implement.md:432-435 (a fix run syncs the description and "Never transition the issue here"). Neither planning nor finish acts on it | "In particular `/flow`'s implement phase touches Jira's **status** not at all." | med | keep "No other command transitions the issue." |
| JI-02 | jira-integration.md:84-85 | 185 | DUPLICATE | canonical brainstorm.md:33-36, the table row at 75 and Never blocking (150-162). Also stale: "writes its state at the end", but STARTED is written at the start (brainstorm.md:38-39) | "The creating run's transition fires before brainstorming. A failed transition is one line" | med | planning-only text in a file both sessions load, so both would benefit |
| JI-03 | jira-integration.md:87-91 | 470 | LAZY-SPLIT | finish run 1 only (integrate.md:258-261); planning never acts on it. The "because…" clause at 89-90 is RATIONALE and moves too | "bare `/flow`'s In Review transition is **not** conditioned on a pull request existing." | med | means splitting the shared file by who loads it |
| JI-04 | jira-integration.md:112-117 (sub) | 461 | RATIONALE | why not `statusCategory`. The rule sentence (111) and the To Do names (103-104) stay. Target: jira-integration-rationale.md "### Transitions" | "That field groups a custom `TO DO URGENT` with `In Progress` under `indeterminate`," | high | none |
| JI-05 | jira-integration.md:141-142 (sub) | 108 | RATIONALE | a reason, in the retired "proposal" vocabulary | "Nothing about the proposal depends on the answer, which is what keeps the guardrail's" | med | none |
| JI-06 | jira-integration.md:144-148 | 400 | LAZY-SPLIT | the join confirmation happens only in finish run 1, with jira-followups.md loaded. Move it next to the join itself | "**The other is the join confirmation**, stated under **Follow-up issues** (`jira-followups.md`)," | med | the "exactly two carve-outs" count (135) must still resolve |
| JI-07 | jira-integration.md:182-183 (sub) | 95 | RATIONALE | the reason for append-only | "— a bad paraphrase must be able to add a line, never to destroy the reporter's original ask." | med | none |
| JI-08 | jira-integration.md:187-188 (sub) | 180 | RATIONALE | why the assertion exists | "A truncated read, a lossy ADF↔Markdown round-trip, or a summarising paraphrase would" | med | the failure modes are restated at 195-197 |
| JI-09 | jira-integration.md:208-211 (sub) | 344 | RATIONALE | why the description is re-read just before the write. The imperatives (206-207, 211-212) stay | "Between such a read and the write, anything may have edited the description:" | med | none |
| JI-10 | jira-integration.md:215-216 (sub) | 107 | RATIONALE | why the pre-edit text is echoed. "Every append is likewise reported…" stays | "The transcript is then the recovery path: the original is recoverable even if" | med | none |
| JI-11 | jira-integration.md:219-221 | 260 | LAZY-SPLIT | join-only (finish run 1). Move it into jira-followups.md, which already states the join echo | "**A join is the one write this echo does not cover**, because the description it" | med | a scope qualifier; it must stay reachable wherever a join write runs |
| JI-12 | jira-integration.md:223-249 | 1,715 | LAZY-SPLIT | `/flow`'s creating run never creates an issue. Used by finish run 1 (finish-contract-run1.md:178), run 2 (finish-contract-run2.md:287-288), /flow-plan (158-159), /flow-self-review (67) and jira-followups.md:14 | "### Labels on issues the pipeline creates" / "An issue any `/flow*` command creates carries **every label" | med | many citers have to follow the heading |
| JI-13 | jira-integration.md:93-111 | 1,345 | MECHANICS | `flow jira transition` (KAN-571; stats/internal/jira/jira.go:7-9, 62-65) already implements the same four-position table in the daemon | "**Resolve transitions by name, never by identifier.** Transition IDs are not portable" | *low* | off unless the `FLOWD_JIRA_*` variables are set. It handles an unrecognised status differently (Drift #15) |
| JI-14 | jira-integration.md:61-66 | 513 | MECHANICS | folding and capping the slug is a pure string transform | "A slug derived from an issue summary is derived from untrusted, externally-authored text" | *low* | the security text has to stay |
| JI-15 | jira-integration.md:71-149 + 167-249 | 10,112 | LAZY-SPLIT | a run with no key (`jiraIssue` null; 162: "no Jira call is attempted at all") | "### Transitions" … "### Follow-up issues" | *low* | overlaps JI-01 to JI-12. Rare in this project (`## jira: KAN`) |
| JI-16 | jira-integration.md:169-221 | ~3,100 | LAZY-SPLIT | load only when the operator added scope the issue does not describe | "`/flow`'s creating run and `/flow`'s implement phase are the only commands that write" | *low* | the fix run and jira-followups.md:278 need it too |
| PV-01 | plan-provenance.md:26-28 (sub) | 255 | RATIONALE | why the merge base is the wrong ref. The "Name the branch…" imperatives stay | "A plan under `/flow`'s implement phase sits on a branch whose commits do not exist yet," | med | none |
| PV-02 | plan-provenance.md:62-64 (sub) | 167 | RATIONALE | why numbers get no escape hatch. The rule at 60-62 stays. Target: plan-provenance-guard-rationale.md "## plan-provenance.md — The asymmetry rule" | "A number has no such escape hatch: an untagged number is not "unverified"," | med | none |
| PV-03 | plan-provenance.md:72-73 (sub) | 125 | RATIONALE | why a false tag is worse than an honest one | "— it tells the next reader a check happened when it did not, which is exactly" | med | none |
| PV-04 | plan-provenance.md:82-94 | 1,003 | MISPLACED | written for the implementer; the implementer dispatch prompt cites it (implement.md:719-720). The planner's own rule for a disproved premise is brainstorm-planner.md:313-331 | "## When a measurement contradicts the plan" / "A plan is an argument, not a script." | med | needs a new home the implementer reads — a small sibling contract, not implement.md itself |
| BG-01 | build-green.md:18-20 (sub) | 182 | RATIONALE | a detail of guard behaviour | "The field's own grammar still admits a dotted id, so a partner written `2.1`" | *low* | could help diagnose a guard hit |
| BG-02 | build-green.md:39-44 (sub) | 511 | RATIONALE | why fields sit at column 0, plus a "Measured on…" provenance note. Its "nothing catches kindly" is now false: check-plan-shape.py:222 (F3a) catches it, verified by running the guard. Target: build-green-rationale.md "## build-green.md — The build-green tag" | "Indenting the fields along with the steps is the natural reading of "the body sits" | high | none |
| BG-03 | build-green.md:49-50 (sub) | 104 | RATIONALE | a reason in parentheses | "(which would name the consequence and hide the cause, exactly as the unclosed-fence" | med | none |
| BG-04 | build-green.md:70-73 | 416 | DUPLICATE | the creating-run half is brainstorm-planner.md:413-415 (same session); the panel half is review-panel.md:1153-1154 and 1431-1432 | "`/flow`'s creating run runs this guard, when the project declares one, at the same point" | med | none. The "The guard's scope" heading, which review-panel.md cites, stays |
| BG-05 | build-green.md:78-81 (sub) | 195 | RATIONALE | a cross-reference to a limit stated in a file the run does not load. The duty sentence (81-82) stays | "This is the same accepted limit as **What the guard does not do**" | med | the moved citation still resolves: the heading exists in plan-provenance-guard.md |
| HB-01 | handoff-blocks.md:1-60 + 89-262 | 14,637 | MISPLACED | the planning run loads this /flow-status contract only for the STARTED template (61-88), because brainstorm.md:252-253 cites the file instead of carrying the block. That goes against handoff-blocks.md:5-6, flow-contracts/SKILL.md:25 and pipeline.md:301-302. verify-and-handoff.md:374 and integrate.md:280 already carry their own blocks | "# Handoff blocks" / "The per-state handoff block templates and the rules governing their regeneration." | high | carry the 604-byte template (run-only markers dropped) in brainstorm.md: net −15.3 KB. Resolve Drift #2 and #3 in the same change |

## Top 5

1. **HB-01**: brainstorm.md carries the STARTED template (604 B) and stops citing
   handoff-blocks.md for it. **Planning, −15.3 KB net.** This is what the contract already requires.
2. **BR-08**: move **Resuming at `STARTED`** plus the resume half of "Resume and fix runs" into their
   own file, routed from SKILL.md:140-141. **Implementation, −14.9 KB** (it loads all of brainstorm.md
   today to read this one section). **Planning, −1.85 KB.**
3. **BR-09 + BP-02**: load the withdrawal route and the flow-fix/flow-cost reachability check from one
   file, only when needed. **Planning, −5.2 KB** in the typical run.
4. **BP-23 + BP-21 + BP-22 (MECHANICS)**: script the Decide defaults (extend `plan-class.sh`) and the
   `## Decision` render (`flow decision render`). **Planning, −5.2 KB.** Needs code and parity tests.
5. **In-session duplicate cluster**: BR-10, BR-11, BR-13, BR-14, BR-16 and BP-25, where "Run
   brainstorming and planning directly" repeats brainstorm-planner.md. **Planning, −2.3 KB**, all plain
   cuts. Next after this: BP-04 (seeded note, −2.2 KB) and BP-20 (sdd-only Decide steps, −2.0 KB).

## Answers to the dispatch questions

### Blocks that apply only under one condition

| Condition | Blocks | Bytes | When it can be decided | How often the condition is true |
|---|---|---|---|---|
| A resumed STARTED run | brainstorm.md:108-130 (**Resuming at `STARTED`**) and 257-261 (resume half). Also the resume offer at 146-149 (inside BR-09) and the clause at brainstorm-planner.md:172-173 | 1,851 | at load time: the router's `flow state get` (SKILL.md:140-141), or A's candidate lookup (brainstorm.md:18-22) | planning session: rare (only when a capture or an interrupted run exists). Implementation session: **every run** |
| The withdrawal route | brainstorm.md:132-192 | 4,154 | at the offer point: the reachability check ended the run, or a planless STARTED resume | very rare |
| A run with no Jira key | jira-integration.md:71-149 and 167-249 (unused once Resolution yields null, per 162). Also brainstorm.md:33-36 (the transition step, 320 B) | 10,112 (+320) | inside A's Resolution, after the file has loaded. A directive keyed on `## jira: none` or on having no Atlassian tooling would catch only the silent-null case | rare in this project (`## jira: KAN`); every run in a project with no tracker |
| A fix run's re-plan | brainstorm.md:261-265 | 376 | never, because the fix run never loads brainstorm.md. The text does nothing where it sits (Drift #6) | — |
| A big-class plan | contiguous: brainstorm-planner.md:450-455 (step 2) and 475-490 (step 4, including "Each group carries its own `model` and `effort`"). Inline, and not movable verbatim: 499 (tree row), 512-513 (`xhigh` clause), 551-553 (groups JSON fields), 577 and 582-583 (block rows), 597-601 (row rules) | 2,035 contiguous (+~700 inline) | after `plan-class.sh` and the override (Decide step 1) | ~2–5% |
| A No at the plan review gate | brainstorm-planner.md:634-638, plus the brainstorm.md:250-251 restatement (BR-16) | 407 | from the answer | never under this project's `## decisions: recommended` (639-641). Not worth splitting: it carries "**Yes** is the only exit." |

### What /flow needs from handoff-blocks.md

- Only the **STARTED** template (63-80, 604 B once the run-only markers are dropped), the
  none-when-zero rule for the two counts (82-83), and two facts the carried copy can spell out
  literally: `(run-only)` is an annotation and is never printed (55), and a null `artifactUrl` is
  rendered as "missing" (44-46, 67). `/flow` always writes `artifactUrl`, `planningEffort` and
  `models.default` as null, so those cells are fixed strings.
- Nothing else in the file is used by any `/flow` session. The implementation and finish phases carry
  their own blocks: verify-and-handoff.md:374 and integrate.md:280. integrate.md:302-303 cites the
  merge-status test only when the landing route is uncertain.
- Before copying the template, fix Drift #2 (the IntelliJ path) and #3 (whether the STARTED block
  shows the Recorded counts).

### What brainstorm-planner.md restates

- **superpowers skills** (only explicit restatements): 251-252 lists writing-plans' own
  self-review items (BP-07, 52 B, low). Every other mention of the skills is an invocation or a
  scoped override and must stay: 30-31, 33-36 ("even where the brainstorming skill says to"), 48-56,
  250-251 ("plan quality" is itself the resume criterion at brainstorm.md:125-126).
- **build-green.md**: 254-256, the Placement rule (BP-08); 373-374, red requires Squash-with (BP-15).
- **plan-provenance.md**: 309-311, the tag duties (BP-10). 80-85 applies the evidence rule to seeding
  (keep that application; 83-84's attribution is wrong, Drift #14).
- **pipeline.md**: 110-114 and 130-131 partly restate Stage exit (pipeline.md:60-63). Keep them: they
  add the call-site default and the exact ⚠ marker text that operator-prompts.md requires each call
  site to state. 643-645 restates brainstorm.md:251-255 and pipeline.md:239-240's `/clear` rule
  (BP-25).
- **Within the file**: the Files clause in 66-71 restates 338-342 (inside BP-04).

### jira-integration.md by session

| Section (lines, bytes) | Planning (creating run) | Fix run (implement) | Finish run 1 | Finish run 2 |
|---|---|---|---|---|
| Header 1-15 (830) | yes | yes | yes | yes |
| Resolution 17-51 (2,292) | **yes, the only resolver in /flow** | no | only when jira-followups.md is loaded (candidate-key prefix rule) | no |
| Change naming 53-69 (1,030) | **yes** | no | only via jira-followups.md:149 (slug cap) | no |
| Transitions: table and by-name / forward-only rules 71-78, 93-117 | In Progress | no | In Review | Done |
| — 79-82 (implement phase touches no status; writes description) | no | yes (duplicated at implement.md:432-435) | no | no |
| — 84-85 (creating-run timing) | **planning-only** (duplicate) | no | no | no |
| — 87-91 (In Review not tied to a PR) | no | no | **finish-only** | no |
| Unrecognised statuses 119-142 | rare | no | rare | rare |
| — 144-148 (join confirmation) | no | no | **finish-only** (with jira-followups.md) | no |
| Never blocking 150-165 (770) | yes | yes | yes | yes |
| Description sync 167-218 | yes, only if the operator added scope | yes | only for a join (jira-followups.md:278) | no |
| — 219-221 (join echo exception) | no | no | **finish-only** | no |
| Labels 223-243 (1,514) | no (/flow planning never creates an issue; /flow-plan does) | no | **yes** (follow-ups) | **yes** (self-review filing) |
| Follow-up pointer 245-249 (200) | no | no | **finish-only** | no |

- Only planning needs: Resolution, Change naming and 84-85, about 3.5 KB. Description sync (3.2 KB) is
  shared with the fix run.
- Only finish needs: 87-91, 144-148, 219-221 and 223-249, about 2.85 KB (JI-03, JI-06, JI-11, JI-12).
- A finish session that files no follow-up never uses Resolution, Change naming or Description sync,
  about 6.5 KB. That is a saving for the finish-session analysis.

### MECHANICS candidates — derivations done by hand that a script does or could do

1. The experimental prompt pick (BP-21). `plan-class.sh` already prints `experimental_roll`.
2. The compact and experimental verdicts, execution mode, static roster and grouping, and bundle-cap
   placement (BP-22). All are table lookups on class and rolls.
3. The micro decision (442-445, 501-503). It is fully defaulted, so a script can emit the whole
   decision.json. Keep 501-503's "never skips verify / self-review".
4. The `## Decision` block and its `planning:`/`models:` preamble (BP-23). A pure function of
   decision.json.
5. decision.json's mechanical fields: `classMechanical`, `inputs`, `rolls`, `resolved` (BP-24, low).
6. `<repos>`. It could be derived from the plan's `**Files:**` and `## apps`, which would also fix
   Drift #7.
7. The kickoff worktree steps (BR-M1). One script shared by /flow, /flow-plan and implement.md §2.
8. Jira transitions (JI-13). `flow jira transition` already exists in the daemon but no skill calls
   it; the dispatcher has to decide on the unrecognised-status difference and on the
   `FLOWD_JIRA_*` gating.
9. Lower value: the change-name slug (JI-14), the withdrawal steps (BR-M2), resume-point detection
   (BR-M3), and a key-prefix filter on `flow state resolve` (brainstorm.md:18-22).

### Citers of each moved section

`check-references.sh` checks only a bold token and a path **on the same line**. Split-line citers
("unchecked" below) still need updating by hand.

- **Resuming at `STARTED`** (brainstorm.md:108):
  - SKILL.md:141 (checked)
  - pipeline.md:80 (checked)
  - flow-plan/SKILL.md:222-223 (split, unchecked)
  - brainstorm.md:255 ("above", no path)
  - plain mentions: implement.md:4, :216 ("resume at `skills/flow/brainstorm.md`")
- **The withdrawal route** (brainstorm.md:132):
  - brainstorm-planner.md:25 (checked)
  - brainstorm.md:121 ("below")
  - pipeline.md:81 (checked)
  - pipeline.md:246-247 (split)
  - handoff-blocks.md:257-258 (split)
  - plain mentions: operator-prompts.md:86-87 ("the withdrawal offers in `skills/flow/brainstorm.md`"), state-file.md:188
- **Resume and fix runs** (brainstorm.md:257): self only (259).
- **Reachability check** (brainstorm-planner.md:17-28, no heading): plain mentions at pipeline.md:80,
  brainstorm-planner.md:319-320 and brainstorm.md:143.
- **Seeded note** (brainstorm-planner.md:58-85, no heading): plain mention at implement.md:398.
- **Decide steps 2 and 4** (the `**Decide**` heading stays): flow-fast/SKILL.md:81-83 ("**Decide**
  steps 1–4 and **The tree**") must name the new file. Unaffected: SKILL.md:91, flow-plan/SKILL.md:183,
  README.md:131, model-policy.md:37 (**Model and effort**).
- **Labels on issues the pipeline creates** (jira-integration.md:223):
  - finish-contract-run1.md:178 (checked)
  - flow-self-review/SKILL.md:67 (checked)
  - jira-followups.md:14-15 (split)
  - flow-plan/SKILL.md:158-159 (split)
  - finish-contract-run2.md:287-288 (split)
  - jira-integration-rationale.md:51 (the same heading)
- **Join paragraphs** (jira-integration.md:144-148, 219-221): no inbound citers; they cite jira-followups.md.
- **Description sync** (only if JI-16 were taken):
  - implement.md:434
  - jira-followups.md:278
  - jira-integration.md:81
  - jira-integration-rationale.md:40 (heading)
- **When a measurement contradicts the plan** (plan-provenance.md:82):
  - implement.md:719-720 (split)
  - flow-contracts/SKILL.md:30 (description text)
- **The block each state renders** (handoff-blocks.md:12, stays): HB-01 removes only the brainstorm.md:253 citer.
- The **The `## Decision` block** / **Resume and fix runs** `design.md` citers (Drift #1):
  - brainstorm.md:239, 259
  - implement.md:34, 143
  - review-panel.md:6
  - verify-and-handoff.md:415-416

### Guards that constrain the trim (checked)

- `check-normative-inventory.sh` (a bulk prose trim must keep the MUST/SHALL inventory
  byte-identical, duplicates included): **none of the six files contains a whole-word MUST or SHALL**,
  so no candidate here changes the inventory.
- `check-references.sh`: every RATIONALE move keeps its citations resolvable (BG-05's heading exists in
  plan-provenance-guard.md). Moving a section needs the citer updates listed above.

## Drift / bugs found

1. **Stale `design.md` citations.**
   - Where: brainstorm.md:239 and 259, implement.md:34 and 143, review-panel.md:6,
     verify-and-handoff.md:415-416.
   - The headings they cite, "The `## Decision` block" and "Resume and fix runs", exist only in
     `spectre/changes/archive/kan-472-flow-dynamic-review-panel-roster-repo-scoped/design.md:119,157`.
     No skill file carries either heading.
   - That archived shape is one `Setting / Toggle / Result` table followed by `class:`/`inputs:`/`rolls:`
     lines. The live shape is two tables (brainstorm-planner.md:562-584).
   - In an installed project, "`design.md`" reads as the change's own design.md.
   - `check-references` skips paths that resolve to nothing, so this was never caught.
   - brainstorm.md:239 tells the run to print "the shape `design.md` shows".
2. **STARTED IntelliJ path.** handoff-blocks.md:73 opens `<absolute main-checkout path>`. pipeline.md:343-345
   says STARTED opens the apply worktree root, and pipeline.md:253-255 forbids main-checkout paths.
3. **STARTED counts.** brainstorm-planner.md:236-238 says no STARTED handoff counts decisions or open
   questions and that the IN_PROGRESS handoff does. The truth is the opposite: handoff-blocks.md:68
   (STARTED) carries `Recorded: <N> decisions · <N> open questions`, and so does
   handoff-blocks-rationale.md:18-31. verify-and-handoff.md:374-400 (IN_PROGRESS) has no count.
4. **Wrong file cited for Open questions.** handoff-blocks.md:84-85 cites **Open questions** in
   `skills/flow/brainstorm.md`; the heading is brainstorm-planner.md:217. The citation is split across
   two lines, so the guard never checks it.
5. **brainstorm.md:252-253 sends `/flow` to handoff-blocks.md for its block.** This contradicts
   handoff-blocks.md:5-6 ("Loaded by `/flow-status` and no other command"), flow-contracts/SKILL.md:25
   and pipeline.md:301-302. It costs the planning session 15.9 KB.
6. **A fix run's re-Decide cannot be reached.**
   - It is stated only in brainstorm.md:261-265.
   - A fix run is routed to implement.md §3 (SKILL.md:142-145), and implement.md:336-452 never runs
     `plan-class.sh` or records a second decision row.
   - Appended tasks can change the class without the decision following.
   - Moving the text would change behaviour, so this needs a decision.
7. **`<repos>` means different things.**
   - /flow (brainstorm-planner.md:432-433, brainstorm.md:207-208): the size of the resolved worktree
     set. At plan time that is always 1, because extra worktrees are created only in
     `flow.isolate-workspace` (implement.md:263-297).
   - /flow-plan (flow-plan/SKILL.md:183-185): the distinct repo roots of the plan's `**Files:**`, per
     `## apps`.
   - So `plan-class.sh`'s `repos>1 and tasks>=11 → big` can fire only in /flow-plan.
   - worktree-resolution.md:5-6 does not list the creating run among its loaders, yet brainstorm.md:208
     cites it.
8. **The implementation session loads brainstorm.md.**
   - SKILL.md:140-141 routes every STARTED run to **Resuming at `STARTED`**. Per brainstorm.md:254-255
     and pipeline.md:80-81, that includes the normal post-`/clear` implementation entry.
   - The rubric's IMPLEMENTATION load set leaves this out; it costs about 17 KB a run.
   - implement.md:296 also points to brainstorm.md step 3 when extra `## apps` worktrees are needed.
9. **Where the STARTED write happens.**
   - jira-integration.md:84-85 says "the run still writes its state at the end", but STARTED is
     written first (brainstorm.md:38-39).
   - brainstorm.md:64-65 says "No further command runs before this point", but the In Progress
     transition (33-36) runs after the name is fixed and before the write.
10. **Key resolution is not /flow-only.** jira-integration.md:19 says "Only `/flow`'s creating run
    resolves a key". /flow-plan (flow-plan/SKILL.md:154-156) and /flow-fast (commands-claude/flow-fast.md:10)
    resolve keys with the same Resolution section.
11. **Retired "proposal" wording.** jira-integration.md:135-142 says "alter the proposal" and "Nothing
    about the proposal…", but `/flow` publishes no proposal (SKILL.md:207).
12. **build-green.md:40-44 is out of date.** It says an indented field is "wrong in a way nothing catches
    kindly". check-plan-shape.py F3a now names it. Verified by running the guard on a scratch plan: "task 1
    carries a **Build:** line indented past column 0…", exit 1. The guard runs unconditionally at
    brainstorm-planner.md:413.
13. **Relocation script already exists.** brainstorm-planner.md:410 says "a script this change adds
    elsewhere", but `generate-relocation-comparison.sh` exists and runs at review-panel.md:192.
14. **Wrong attribution to plan-provenance.md.** brainstorm-planner.md:83-84 says plan-provenance.md "also
    names why: seeding is how…". plan-provenance.md:76-80 says nothing about seeding.
15. **`flow jira transition` is unused.**
    - It exists (stats/cmd/flow/jira.go, KAN-571), but no skill calls it.
    - It treats an unrecognised status as exit 1 → skipped. jira-integration.md:127-142 asks the
      operator once instead.
    - That mismatch matters if it is ever wired in.
16. **Wrong ending for /flow-plan.** brainstorm-planner.md:643-645 sits in the Plan review gate, which
    /flow-plan also runs, and says "the run ends with a `/clear` handoff". /flow-plan commits and ends
    with its own report (flow-plan/SKILL.md:207-230).
17. **Possible hidden load (outside these six files).**
    - brainstorm.md:99-100 and brainstorm-planner.md:45-47 cite project-configuration.md (53.8 KB), and
      pipeline.md:447-448 says to "load it before resolving project configuration".
    - The planning session may therefore pull in 54 KB, even though `project-get.sh` resolves the key.
    - It is not in the rubric's planning load set. Worth checking against a real transcript.

## Not slimmable

- **brainstorm.md A: name, STARTED JSON, marks, kickoff steps 1–5 (7-106, less BR-01…07).** The
  first durable write, exit contracts, ordering, and crash-safe persistence. Only MECHANICS (BR-M1)
  could shrink it.
- **brainstorm.md record sequence and Yes path (242-246, 251-255).** Stated nowhere else in `/flow`.
- **Convergence (brainstorm-planner.md:90-161).** Operator-prompt wording, silence defaults, exact ⚠
  markers, and the exceptions under the recommended-decisions mode. Never cut. BP-05 is at most a
  low-confidence split.
- **Decisions / Open questions (190-234).** Worked entry shapes; ID immutability and the
  supersede / never-delete rules.
- **D's plan-writing rules (259-305), field family (338-371), After (376-382), Decision citation
  (384-392), headers (394-411), guard exits (413-421).** Guards parse these byte for byte, or they are
  exit-code contracts.
- **Decide steps 1 and 3, Model and effort (505-520), decision.json schema (538-559).** The record's
  schema. Only MECHANICS could shrink them.
- **Plan review gate (617-641).** Operator-prompt wording, the No loop, the mode exception.
- **jira-integration.md Resolution (17-51) and Change naming (53-69).** Security: untrusted keys and
  slugs, explicit confirmation, prompt wording.
- **jira-integration.md Never blocking (150-165) and the Description sync pre-write assertion
  (185-204).** The exact ⚠ lines and the data-loss guard.
- **Unrecognised statuses ask (119-139).** Prompt wording and the limits of the carve-out.
- **plan-provenance.md four tags, examples, asymmetry and evidence rule (6-80, less PV-01…03).** The
  canonical vocabulary the guards parse.
- **build-green.md tag, Placement, column-0 rule, first-line rule, guard scope (7-68, less BG-01…03).**
  Parsed rules and the guard's failure list.
- **handoff-blocks.md STARTED template (63-80).** Must be carried verbatim: its labels must match what
  /flow-status regenerates.
