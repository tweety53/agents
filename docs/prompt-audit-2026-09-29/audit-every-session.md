# Router + contracts: slimming analysis (read-only)

> **Audit snapshot at `700e184` (2026-09-29).** A read-only model pass that followed [`rubric.md`](rubric.md); every row is a candidate to verify, not a verified fact. Line numbers refer to `700e184`. Main has since retired the bugbot and security slots (`ebdfdde1`..`4dafab4a`), so re-locate each row by its quoted first words, and treat bugbot/security rows as obsolete. Tracked in KAN-851 → KAN-855 (and the Drift items in KAN-853 / KAN-854). Plan: [`README.md`](README.md).

Repository at `700e1841`, working tree clean. Nothing in the repository was modified.

Files analysed, read in full: `skills/flow/SKILL.md`, `skills/flow-contracts/pipeline.md`, `commands-claude/flow.md`,
`CLAUDE.md`, the rendered always-on block (`~/.claude/CLAUDE.md` from a sandboxed `HOME=$(mktemp -d) ./setup.sh global`, rendered by `setup.sh`
from each `rules/*.mdc` H1 + `<!-- core -->` section + a generated `Full rule:` line; setup.sh:668-681),
`skills/flow-contracts/operator-prompts.md`, `model-policy.md`, `state-file.md`, `project-configuration.md`.

Method:
- Bytes are exact: `sed -n 'A,Bp' FILE | wc -c` for whole-line ranges. For a sub-paragraph span, a
  helper script (not kept) measures the verbatim text between two phrases. That is why some
  ranges share a line with a neighbour.
- Citers were found with a whitespace-normalised `**Heading**` + path matcher (a helper script, not kept).
  Unlike the line-based `check-references.sh`, it also catches bold spans that cross a line break.
- Totals count high- and medium-confidence rows only, and count each byte once. Low-confidence rows are
  listed but kept out of the totals.
- The per-session figures assume this operator's configuration from `.flow/project.md`:
  - `## decisions: recommended`
  - `## self review: defer`
  - `## default landing route: merge and push`
  - `## workspace isolation` declared
  - a typical run: no `ui paths` touched and no gate verdict fires.

## Totals

| file | bytes | RATIONALE | DUPLICATE | STALE | LAZY-SPLIT | MECHANICS | MISPLACED | sum |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| skills/flow/SKILL.md | 15449 | 274 | 2963 | 609 ¹ | 699 | 365 | 1320 | 6230 |
| skills/flow-contracts/pipeline.md | 28567 | 1343 | 1223 | 346 | 3199 | 1876 | 1035 | 9022 |
| commands-claude/flow.md | 2915 | 0 | 2186 | 0 | 0 | 0 | 0 | 2186 |
| CLAUDE.md | 6636 (≈5883 as loaded ²) | 114 | 335 | 126 | 0 | 0 | 0 | 575 |
| always-on block (rendered) | 13296 (≈12655 as loaded ²) | 1006 | 0 | 0 | 0 | 0 | 2292 | 3298 |
| skills/flow-contracts/operator-prompts.md | 5907 | 276 | 0 | 0 | 4901 | 0 | 149 | 5326 |
| skills/flow-contracts/model-policy.md | 7335 | 0 | 1918 | 26 | 828 | 0 | 0 | 2772 |
| skills/flow-contracts/state-file.md | 21258 | 445 | 754 | 63 | 0 | 0 | 5799 | 7061 |
| skills/flow-contracts/project-configuration.md | 53831 | 0 | 1349 | 0 | 39475 | 0 | 0 | 40824 |
| **total** | **155194** | **3458** | **10728** | **1170** | **49102** | **2241** | **10595** | **77294** |

¹ 499 B of it (K8, K9) is stale text that needs a rewrite, not a cut. See the rows.
² HTML comments are stripped at load: this session's own system-reminder copy of `CLAUDE.md` lacks
lines 54-63. The rendered block's `<!-- … -->` lines (641 B) cost nothing either.

**Saving per session** excludes MECHANICS (needs code) and K8/K9. A move counts as a saving only in
sessions that stop loading the text.

| session | saves | saves if state-file.md and project-configuration.md are not in fact loaded |
|---|---:|---:|
| planning | 66,783 B | 18,898 B |
| implementation | 64,431 B | 22,156 B |
| finish (merge-and-push, run 1 chained into run 2) | 69,049 B | 23,897 B |

The first column follows the literal `load it before …` directives (pipeline.md:440-441, 447-448). They put both
contracts in every session; see **B** below. The rubric's per-session load lists omit both files. No `/flow`
transcript exists on this machine to show which is true: `~/.claude/projects/` holds only this analysis
session. So both columns are given.

## Candidates

Row-id prefixes:
- K = SKILL.md
- P = pipeline.md
- C = commands-claude/flow.md
- L = CLAUDE.md
- A = always-on, attributed to the source `.mdc` line with the rendered line in brackets
- O = operator-prompts.md
- M = model-policy.md
- F = state-file.md
- G = project-configuration.md

O3's row shows its whole span (4302 B). 276 B of that span are O4-O6, so the totals count O3 as 4026 B.

| id | file:startline-endline | bytes | lever | condition / canonical location / evidence | first ~10 words verbatim | confidence | behaviour-risk note |
|---|---|---:|---|---|---|---|---|
| K1 | skills/flow/SKILL.md:22-24 | 227 | DUPLICATE | canonical pipeline.md:150-160 (Progress visibility); every session loads both | One entry per brainstorming checklist item and artifact on the | high | none; fix pipeline.md:152 drift first (see Drift #12) |
| K2 | skills/flow/SKILL.md:24-25 | 110 | STALE | flow-fast/SKILL.md:167-168 registers one entry per file or logical unit, not per branch item | — the same granularity `/flow-fast` used for the branch it | high | none |
| K3 | skills/flow/SKILL.md:30-45 | 1320 | MISPLACED | editor index: each phase file marks its own keys with literal `-stage` commands; optional-file load conditions are stated at the call sites (review-panel.md:185, verify-and-handoff.md:172); a resume reads artifacts, not keys (brainstorm.md:108-128). Target: README.md beside the Level 1 table | ## Stage keys Every stage `/flow` marks uses a `flow.*` | med | repoint README.md:138, pipeline.md:88, flow-plan/SKILL.md:38 |
| K4 | skills/flow/SKILL.md:105-107 | 274 | RATIONALE | explains why SELF_REVIEW_MODEL is absent; archive.md step 9 resolves it and is canonical. Target SKILL-rationale.md 'SKILL.md — Model resolution' | **`SELF_REVIEW_MODEL` is not resolved here.** It governs no dispatch and | med | none |
| K5 | skills/flow/SKILL.md:109-115 | 620 | LAZY-SPLIT | implementation only, and only when visual-verify.md loads (a diff matched `ui paths`); sole consumer visual-verify.md:15. False in most runs. Target visual-verify.md beside its dispatch. Trailing clause L114-115 (117 B) is RATIONALE | **`VERIFY_MODEL` governs the one verifier dispatch** — `flow.visual-verify`'s (**Visual verification**, | med | keep `VERIFY_MODEL=opus` in the bash block (check-model-resolution-shell.sh pins the block); repoint visual-verify.md:15 |
| K6 | skills/flow/SKILL.md:117-121 | 417 | DUPLICATE | each dispatch site names its own model: implement.md:502-503, review-panel.md:159-160 and 414-419, visual-verify.md:94-95; model-policy.md:50-51. Implementation-only content | **`DEFAULT_MODEL` is the model for all four roles this run | med | low; alternative: move verbatim to implement.md:31 |
| K7 | skills/flow/SKILL.md:182-183 | 140 | DUPLICATE | pipeline.md:374-376, 388-392 | A complete set prints nothing; any absence prints that section's | high | none |
| K8 | skills/flow/SKILL.md:185-188 | 326 | STALE | names 2 flow-guard shims; 26 of skills/flow/scripts/* source lib/flow-guard.sh (grep) | `check-unfinished-work.sh` and `check-task-commit-fields.sh` are `flow-guard` shims: each also requires `<agents | high | FIX, NOT CUT: pipeline.md:397's `$SCRIPT_DIR/<name>` grep misses `$(dirname "${BASH_SOURCE[0]}")/lib/flow-guard.sh`, so this is the only lib/ sibling check |
| K9 | skills/flow/SKILL.md:192-194 | 173 | STALE | `flow.state-gate` exists nowhere else (grep: README, stats/internal/stages); cited section is not in pipeline.md, only pipeline-rationale.md:89; bold spans L193-194 so check-references.sh cannot see it | — defer `flow.state-gate`-equivalent bookkeeping into that section, per **The `<change>` | high | FIX, NOT CUT: the deferral is normative; only the key and the citation target are wrong |
| K10 | skills/flow/SKILL.md:207-208 | 116 | DUPLICATE | brainstorm.md:60-62 (planning, the only writer); state-file.md:164-165 | - Never publish a proposal artifact. `artifactUrl` is written `null` | high | none |
| K11 | skills/flow/SKILL.md:209-209 | 79 | LAZY-SPLIT | planning only; move verbatim to brainstorm-planner.md (design approval / writing-plans) | - Never skip brainstorming's design gate, or leave `tasks.md` a | med | none |
| K12 | skills/flow/SKILL.md:210-217 | 666 | DUPLICATE | review-panel.md:168-174 (operator-named id, checked at stage start and every fix round) and 241-245 (only automatic reductions, only remove; nothing added by size/area/trigger); implementation-only | - Never add a slot beyond the resolved roster automatically, | high | none |
| K13 | skills/flow/SKILL.md:218-220 | 274 | DUPLICATE | implement.md:550-558 (Waves: at most three in flight; a fourth queues in plan order); implementation, `big` class only | - Never run more than three implementer dispatches in flight | high | none |
| K14 | skills/flow/SKILL.md:221-224 | 354 | DUPLICATE | review-panel.md:326-328 (at most two dispatches, one to three roles) and 175-181 (per decision's panel.dispatches) | - Never dispatch review-panel roles as separate parallel `Agent` calls. | med | 're-check before any round' is not verbatim there: move that sentence into review-panel.md's Bundled dispatch rather than delete it |
| K15 | skills/flow/SKILL.md:225-227 | 237 | DUPLICATE | review-panel.md:928-931 (zero open findings, no stale result), 1351 (a deferred Minor blocks nothing); 'no preset' is stale (no presets exist: grep) | - Never hand off with an open finding of any | med | none |
| K16 | skills/flow/SKILL.md:228-229 | 157 | DUPLICATE | pipeline.md:257 (never stages spectre/changes); git-boundaries.md (routes own push/merge/PR) | - Never commit `<project>/spectre/changes/` in a task or fixup commit. | med | none |
| K17 | skills/flow/SKILL.md:233-237 | 375 | DUPLICATE | implement.md:19-24 (The parent orchestrates directly); implementation-only | - `implement.md` sections 1, 2 and 4, <!-- refs-guard:allow --> | high | none |
| K18 | skills/flow/SKILL.md:178-182 | 365 | MECHANICS | with pipeline.md:371-403 (P17): a script prints the block; the run keeps one call + exit contract | **Check guard presence.** Per **Guard presence check** (`skills/flow-contracts/pipeline.md`), confirm every | med | needs code + parity tests |
| K19 | skills/flow/SKILL.md:9-11 | 247 | RATIONALE | descriptive; 'Nothing here duplicates that file's own content' is false (K10-K17) | `/flow` is one command, one state file, three states, with | low (not in totals) | — |
| K20 | skills/flow/SKILL.md:197-200 | 335 | DUPLICATE | pipeline.md:187-189 — but SKILL.md adds placement ('right here') and scope ('inside every phase file') | **Generate this run's session token once, right here, before the | low (not in totals) | ordering constraint: keep |
| K21 | skills/flow/SKILL.md:230-232 | 219 | DUPLICATE | pipeline.md:122-123; drift: 'only run 2 writes FINISHED' vs the withdrawal route | - Never advance the state past what the phase in | low (not in totals) | fix drift, keep |
| K22 | skills/flow/SKILL.md:27-28 | 187 | DUPLICATE | pipeline.md:70-74, flow.md:33-37; SKILL.md's copy is the correct one | **No flags.** The only argument is the optional change name/description | low (not in totals) | fix pipeline.md instead |
| P1 | skills/flow-contracts/pipeline.md:35-36 | 122 | STALE | one pipeline command since the myflow→/flow rework (commit 72f4f0e6); pipeline.md:68 | **Each command ends in the state named after it**, so | med | none |
| P2 | skills/flow-contracts/pipeline.md:41-42 | 83 | RATIONALE | why no *-done command | This is why no `*-done` command exists — there would | high | none |
| P3 | skills/flow-contracts/pipeline.md:54-64 | 894 | LAZY-SPLIT | planning only (the looping brainstorm stage); sole citer brainstorm-planner.md:150. Target brainstorm-planner.md beside **Convergence** | ## Stage exit — never the command's own judgment Within | med | repoint brainstorm-planner.md:150 |
| P4 | skills/flow-contracts/pipeline.md:68-70 | 224 | STALE | inventory omits /flow-fast and /flow-self-review (ls commands-claude/: 6 commands) | One command, `/flow`, drives the whole pipeline, plus one read-only | med | heading stays for flow-settings/SKILL.md:20 |
| P5 | skills/flow-contracts/pipeline.md:73-74 | 82 | RATIONALE | why arguments are reported | — a silently ignored word is indistinguishable from a flag | med | none |
| P6 | skills/flow-contracts/pipeline.md:94-105 | 826 | DUPLICATE | pipeline.md:78-86 (table), SKILL.md:136-176, implement.md §3; 'test guide' occurs nowhere else (grep) | Re-invoking `/flow` is the supported way to revise its output. | med | none; fix-doc ordering is canonical in implement.md §3 |
| P7 | skills/flow-contracts/pipeline.md:155-160 | 487 | LAZY-SPLIT | implementation only; target implement.md (SKILL.md:22-23 also states per-task). Contains RATIONALE clause L157-159 (168 B) | **On the implementation branch the granularity is per task.** The | med | none |
| P8 | skills/flow-contracts/pipeline.md:162-163 | 158 | MISPLACED | /flow-status-only; flow-status/SKILL.md loads pipeline.md but states no task-list rule. Target flow-status/SKILL.md | `/flow-status` is read-only and **registers nothing**. Registering steps for a | high | none |
| P9 | skills/flow-contracts/pipeline.md:167-168 | 75 | RATIONALE | why tasks.md stays the single source | — a second source of completion state would be one | med | none |
| P10 | skills/flow-contracts/pipeline.md:190-193 | 341 | RATIONALE | why the literal token still matters | On Claude Code the mark also binds the stage run's | high | none (the MUST sentence before it stays) |
| P11 | skills/flow-contracts/pipeline.md:197-199 | 253 | RATIONALE | why -harness is a placeholder | because one skill source installs into `~/.claude/skills/` and `~/.zcode/skills/` alike, | med | none |
| P12 | skills/flow-contracts/pipeline.md:217-219 | 155 | RATIONALE | why stage end carries no token | — attribution happens once, at `begin`, and the harness recorded | med | none |
| P13 | skills/flow-contracts/pipeline.md:221-222 | 172 | MISPLACED | /flow-status-only (+ rationale clause). Target flow-status/SKILL.md | `/flow-status` marks nothing — the Level 1 section of `<agents | high | none |
| P14 | skills/flow-contracts/pipeline.md:240-242 | 207 | RATIONALE | why /clear is recommended | Every invocation re-enters from the state file and the change's | med | none |
| P15 | skills/flow-contracts/pipeline.md:354-355 | 147 | RATIONALE | why basenames | — such a path resolves only when the project being | med | none |
| P16 | skills/flow-contracts/pipeline.md:362-369 | 705 | MISPLACED | authoring rule for skill prose, enforced at lint (check-guard-symlinks.sh); no run writes skill prose outside agents-repo changes | **Prose describing this repository's own guard is not an invocation.** | med | none |
| P17 | skills/flow-contracts/pipeline.md:371-403 | 1876 | MECHANICS | guard-set derivation + presence + sibling grep: guardsymlinks.go already derives rule 2's set at lint; a run-time mode could print the block | ## Guard presence check `/flow` checks, once at the start | med | needs code + parity tests; cited by handoff-blocks.md:128, flow-status/SKILL.md:22, SKILL.md:178, integrate.md:13 |
| P18 | skills/flow-contracts/pipeline.md:405-427 | 1818 | LAZY-SPLIT | load only when a gate verdict fires (STAGED-FOREIGN/REFUSE/OUTSTANDING/MOVED/LEFTOVER); planning never; false in most runs. Contains RATIONALE L419-421 (271 B) | ## Hand-verifying a guard verdict **When a gate guard's verdict | med | repoint 5 citers; add the directive at review-panel.md (MOVED) and finish-contract-run2.md step 7 (LEFTOVER) |
| P19 | skills/flow-contracts/pipeline.md:429-435 | 397 | DUPLICATE | integrate.md:7 and archive.md:9 name their finish contract as canonical; also always-on flow-manual-review.mdc:38-41 | ## Finish contract **Finish contract** (`skills/flow-contracts/finish-contract-run1.md`, `skills/flow-contracts/finish-contract-run2.md`) governs the preflight | med | none |
| P20 | skills/flow-contracts/pipeline.md:15-22 | 408 | MISPLACED | discovery aid; phase files cite each sibling at point of use | **Five sections that reach fewer than every command live beside | low (not in totals) | — |
| P21 | skills/flow-contracts/pipeline.md:125-126 | 130 | RATIONALE | state-file-rationale.md 'no fix origin' | No field records where a fix was raised. Whether the | low (not in totals) | — |
| C1 | commands-claude/flow.md:10-19 | 794 | DUPLICATE | pipeline.md:78-86, SKILL.md:136-147, pipeline.md:83. STALE: says a creating run continues to IN_PROGRESS; it ends at STARTED (pipeline.md:80) | On a creating run it writes `STARTED` immediately, then runs | high | removes a contradiction; keep L9-10's accepted-states sentence (pipeline-rationale.md: a command file must state its row's states) |
| C2 | commands-claude/flow.md:21-25 | 426 | DUPLICATE | SKILL.md:204-217, brainstorm.md:60-62, review-panel.md:168-174; drift: roster 'from the settings store's reviewer list' (true only for micro) | Publishes no proposal artifact — the operator is present for | high | none |
| C3 | commands-claude/flow.md:27-31 | 415 | DUPLICATE | SKILL.md:15-18, always-on flow-manual-review.mdc:35-41; stale: pipeline.md no longer holds git boundaries or the finish contract | Also follow the flow rule (`flow-manual-review.mdc`) — installed globally, so | high | none |
| C4 | commands-claude/flow.md:34-35 | 163 | DUPLICATE | SKILL.md:27-28, pipeline.md:457-480; drift: `spectre list --json` contradicts `flow state resolve` and misses worktree-only changes (pipeline-rationale.md:247-250) | **This command takes no flags.** If omitted at `IN_PROGRESS`, run | high | removes a wrong candidate source |
| C5 | commands-claude/flow.md:39-42 | 388 | DUPLICATE | pipeline.md:237-249 + handoff-blocks.md; drift: 'run the apps' vs 'the stack is running' | **When done:** at `IN_PROGRESS` with a fresh staged diff — | med | none |
| L1 | CLAUDE.md:28-29 | 126 | STALE | CONTRIBUTING.md absent from the tree and its git history; restates always-on lint-fix-priority.mdc:13 | Pre-approved suppressions and documented deviations live in `CONTRIBUTING.md`. Do not | high | none |
| L2 | CLAUDE.md:13-14 | 114 | RATIONALE | why the policy is not restated | It is not restated here — one source of truth, | med | none |
| L3 | CLAUDE.md:88-91 | 170 | DUPLICATE | CLAUDE.md:69-71 | ### How to invoke a skill Type its slash command. | med | none |
| L4 | CLAUDE.md:100-101 | 165 | DUPLICATE | CLAUDE.md:69-71 | After install, general skills auto-trigger from their descriptions. Project-specific `/flow` | med | none |
| L5 | CLAUDE.md:95-98 | 235 | MISPLACED | install instructions for the operator | The Superpowers plugin provides general-purpose workflow skills (brainstorming, TDD, subagent-driven-development, | low (not in totals) | — |
| L6 | CLAUDE.md:76-86 | 2178 | DUPLICATE | harness skill listing (frontmatter descriptions); L86 carries a normative rule | ### Skill index \| Skill directory \| Trigger \| Purpose | low (not in totals) | — |
| A1 | rules/flow-manual-review.mdc:8-44 (rendered L184-222) | 2292 | MISPLACED | flow-only core paid by every non-flow session in every project; in flow sessions DUPLICATE of SKILL.md:15-18, pipeline.md:28-52 and 429-435, flow.md:27-31 (with drift). Rendered bytes incl. heading and generated 'Full rule' line | # flow — /flow, start to finish When running `/flow`, | med | design decision (pipeline-rationale.md 'States'); edit the .mdc and re-install (check-installed-rules.sh) |
| A2 | rules/dispatch-carries-the-baseline.mdc:9-13 (rendered same text) | 428 | RATIONALE | why the pointer must be carried | This file is injected into **this** session's system prompt only. | med | none |
| A3 | rules/dispatch-carries-the-baseline.mdc:20-22 (rendered same text) | 221 | RATIONALE | how the baseline file works | `~/.claude/rules/agent-baseline.md` lists every rule in a line and points at | high | none |
| A4 | rules/dispatch-carries-the-baseline.mdc:24-25 (rendered same text) | 111 | RATIONALE | why unconditionally | : you cannot tell from the prompt whether the agent | med | none |
| A5 | rules/fix-determinism-at-the-source.mdc:12-14 (rendered same text) | 138 | RATIONALE | why not widen | A wider tolerance still passes while the defect is present: | high | none |
| A6 | rules/dependency-versions.mdc:10-11 (rendered same text) | 54 | RATIONALE | why not remembered versions | — it is a guess about the past, and it | med | none |
| A7 | rules/context7.mdc:13-13 (rendered same text) | 54 | RATIONALE | why Context7 | — your training data may not reflect recent changes. | med | none |
| A8 | rules/be-brief.mdc:14-25 (rendered same text) | 945 | MISPLACED | scope is agents-repo Markdown + /flow artifacts only ('A consuming project's pre-existing documentation is not subject to it') | **A written file owes completeness and non-repetition — both at | low (not in totals) | the rule this trim itself relies on |
| A9 | rules/build-the-simplest-thing.mdc:10-11 (rendered same text) | 144 | RATIONALE | framing | If something more elaborate is wanted, the user will say | low (not in totals) | — |
| O1 | skills/flow-contracts/operator-prompts.md:16-30 | 875 | LAZY-SPLIT | sole call site: run 2 step 9's self-review filing ask (finish-contract-run2.md:277-286; /flow-self-review copies that shape). False whenever `## self review` is defer/skip (this repo: defer). Target finish-contract-run2.md step 9 | ## The multi-select variant Some prompts ask the operator to | med | 0 heading citers |
| O2 | skills/flow-contracts/operator-prompts.md:32-35 | 149 | MISPLACED | editor doctrine for call-site authors | ## The doctrine Every call site cites this contract for | med | 0 citers |
| O3 | skills/flow-contracts/operator-prompts.md:37-96 | 4302 | LAZY-SPLIT | needed in implementation/fix runs and in planning only under `## decisions: recommended`; never in finish (integrate/archive prompts stay asked, L95-96). Row shows the whole span; 276 B inside it are O4-O6, so the totals count 4026 B | ## Auto-resolution **During implementation and every fix run, a prompt | med | fix project-configuration.md:35 drift FIRST; repoint 14 **Auto-resolution** citers; no guard pins the citation (dispatchparagraphs.go pins phrases only) |
| O4 | skills/flow-contracts/operator-prompts.md:43-45 | 149 | RATIONALE | why auto-resolve | The operator asked for this once, for every flow run: | high | none |
| O5 | skills/flow-contracts/operator-prompts.md:83-84 | 86 | RATIONALE | history of the scope | The operator scoped auto-resolution to implementation and fix options, not | med | none |
| O6 | skills/flow-contracts/operator-prompts.md:94-94 | 41 | RATIONALE | why | The global rules require these confirmed. | med | none |
| M1 | skills/flow-contracts/model-policy.md:6-6 | 26 | STALE | no creating run asks a model question (brainstorm.md:60-61) | at the model questions and | high | none |
| M2 | skills/flow-contracts/model-policy.md:41-48 | 704 | DUPLICATE | SKILL.md:51-80 (order + MODEL_SOURCE reporting), always loaded alongside | **Where `DEFAULT_MODEL` comes from: the project key, then the store, | med | none |
| M3 | skills/flow-contracts/model-policy.md:50-53 | 335 | DUPLICATE | SKILL.md:105-121 | The governed roles are every role above that reads `DEFAULT_MODEL` | med | none |
| M4 | skills/flow-contracts/model-policy.md:57-60 | 312 | DUPLICATE | SKILL.md:123-128 (near-verbatim 'Record the instruction with the dispatch') | **An explicit operator instruction overrides either default, in either direction** | med | cut this copy, not SKILL.md's: SKILL.md is in every session |
| M5 | skills/flow-contracts/model-policy.md:62-66 | 452 | DUPLICATE | SKILL.md:49, brainstorm.md:60-62; STALE: 'records only a model an operator explicitly chose' — nothing writes it | **The model each role runs on is resolved per run | med | none |
| M6 | skills/flow-contracts/model-policy.md:71-72 | 115 | DUPLICATE | SKILL.md:123 ('for this run only') | **A session instruction governs the run in which it is | med | none |
| M7 | skills/flow-contracts/model-policy.md:105-115 | 828 | LAZY-SPLIT | harness zcode only; every citer already says 'On harness zcode' | ## Harness mapping **On harness `zcode`, every model a dispatch | med | repoint 8 **Harness mapping** citers |
| M8 | skills/flow-contracts/model-policy.md:96-102 | 579 | RATIONALE | why frontmatter enforces nothing; carries a normative fragment stated at dispatch sites | a command's `model:` frontmatter (`commands-claude/*.md`) applies only to the turn | low (not in totals) | — |
| F1 | skills/flow-contracts/state-file.md:18-25 | 409 | MISPLACED | fallback-file role + this machine's hardcoded paths; runs never read/write them (L307-310 is the run rule) | **The on-disk JSON file at the path below is the | med | none |
| F2 | skills/flow-contracts/state-file.md:27-31 | 454 | DUPLICATE | finish-contract-run2.md:155-157 carries the `flow state dir` use at its only call site | **The directory those paths live under is resolvable, never guessed: | med | none |
| F3 | skills/flow-contracts/state-file.md:33-56 | 2086 | MISPLACED | project-key derivation the CLI performs ('stated here because ...'); reference for CLI/daemon authors | `<project-key>` = `<basename of main checkout>-<first 8 hex of sha1 | med | stats/internal/fallback/journal_test.go:49 reads MAIN_CHECKOUT=/PROJECT_KEY= from state-file.md — move the test's contractPath too |
| F4 | skills/flow-contracts/state-file.md:62-64 | 237 | RATIONALE | why never-block | — this is deliberately broader than the Jira contract's "never | high | none |
| F5 | skills/flow-contracts/state-file.md:118-121 | 365 | MISPLACED | CLI transport detail; 'no skill supplies it' | **`state set` also accepts, and the CLI itself injects, a | med | none |
| F6 | skills/flow-contracts/state-file.md:172-173 | 63 | STALE | /flow never reads models.default (SKILL.md:51-72; verify-and-handoff.md:338-340 'always null') | Its live consumer is `/flow`, which dispatches on that value. | high | none |
| F7 | skills/flow-contracts/state-file.md:180-182 | 208 | RATIONALE | why absent keys are valid | Without this exception, `state get` would hand back a record | high | none |
| F8 | skills/flow-contracts/state-file.md:194-197 | 300 | DUPLICATE | near-verbatim brainstorm.md:180-184 (the withdrawal route, the only writer) | The route's state write needs a daemon that knows the | high | none |
| F9 | skills/flow-contracts/state-file.md:219-226 | 640 | MISPLACED | store ordering internals / journal retirement | The instant this ordering rests on comes from a single | med | none |
| F10 | skills/flow-contracts/state-file.md:238-265 | 2134 | MISPLACED | daemon replay behaviour; no run step acts on it | ## The journal is replayed, never merged The on-disk journal | high | in-file mentions at L86, L128, L226 repoint |
| F11 | skills/flow-contracts/state-file.md:293-298 | 165 | MISPLACED | worked example of the map: move verbatim with F3/F10, never cut | ```json "worktrees": { "/Users/tweety53/Projects/agents/.worktrees/<name>": "5ee4c9a…", "/Users/tweety53/Projects/other/.worktrees/<name>": "b31f7c2…" } ``` | med | none |
| G1 | skills/flow-contracts/project-configuration.md:135-242 | 9023 | LAZY-SPLIT | workspace-isolation authoring + validator spec; runs read it only when prepare-workspace.sh cannot be located (verify-and-handoff.md:49). Target project-configuration-authoring.md or a new sibling | **How a `## workspace isolation` section is written.** One table | med | repoint workspace-isolation.md:113,261,275,329,342,351, artifacts-registry.md:62,69,80; scripts/prepare-workspace.sh:225 message |
| G2 | skills/flow-contracts/project-configuration.md:273-342 | 5647 | LAZY-SPLIT | `survivors` output/exit contract — executed by check-cleanup-complete.sh; the run reads the guard's verdict table (finish-contract-run2.md:160-178) | **What `survivors` prints, and what its exit code means.** Both | med | cleanupcomplete.go:64,803,830 comments cite it |
| G3 | skills/flow-contracts/project-configuration.md:350-454 | 8296 | LAZY-SPLIT | row validation / enforcement / left-to-the-agent: enforced by check-workspace-isolation.sh via prepare-workspace.sh (exit 1 = dropped row, verify-and-handoff.md:34-37) | **A project declares only the ports it can actually move.** | med | none |
| G4 | skills/flow-contracts/project-configuration.md:243-272 | 2168 | LAZY-SPLIT | finish only: run 2 reads the command table + working directory (finish-contract-run2.md:133-135). Target finish-contract-run2.md step 5 | The command table has three rows and two columns: \| | med | none |
| G5 | skills/flow-contracts/project-configuration.md:343-349 | 565 | LAZY-SPLIT | finish only: token substitution for `remove` (same target as G4) | The tokens `<id>` and `<id_underscored>` are substituted in a command's | med | none |
| G6 | skills/flow-contracts/project-configuration.md:455-528 | 8166 | LAZY-SPLIT | visual-verify stage only (a diff matched `ui paths`) or planning with mockup frames (brainstorm-planner.md:42-47); false in most runs. Target a sibling loaded with visual-verify.md | ## visual verification **How a `## visual verification` section is | med | repoint visual-verify.md:675 (heading citer), 242, 285; verify-and-handoff.md:151-152 |
| G7 | skills/flow-contracts/project-configuration.md:529-547 | 1349 | DUPLICATE | key-table row L40 + review-panel.md:90-107 (procedure, never-blocks); the section itself says review-panel.md is canonical | ## review panel citation check **This key declares one command, | med | 0 citers |
| G8 | skills/flow-contracts/project-configuration.md:43-88 | 3834 | LAZY-SPLIT | implementation only: resolving [STANDARDS_PATHS] for the principles slot (review-panel.md:666-671, principles-reviewer-prompt.md:166-175) | **How a `## standards` entry resolves to a file.** Every | med | none (no heading) |
| G9 | skills/flow-contracts/project-configuration.md:97-130 | 1776 | LAZY-SPLIT | `<agents repo>` resolution — only consumer at run time is a form-1 standards entry (no phase file executes an `<agents repo>/` path) | ## Where the agents repository is `<agents repo>` is the | med | heading has 0 citers |

## Top 5

1. **project-configuration.md split** (every session, per pipeline.md:447-448).
   - Move lines 135-546 out:
     - G1-G3 (22,966 B), the workspace-isolation authoring and validator spec, to `project-configuration-authoring.md` or a sibling. A run reads it only when `prepare-workspace.sh` cannot be found.
     - G6 (8,166 B), the visual-verification spec, to a file loaded with `visual-verify.md`.
     - G4-G5 (2,733 B) to finish-contract-run2.md step 5.
   - Cut G7 (1,349 B), a duplicate.
   - Load G8-G9 (5,610 B) only at the principles dispatch.
   - Run-time core left: **13,007 B of 53,831 B**. Saves 35.2-40.8 KB per session.
2. **state-file.md split** (every session, per pipeline.md:440-441).
   - Move the reference text (F1, F3, F5, F9, F10, F11 = 5,799 B) to a sibling.
   - Cut the duplicates F2 and F8 (754 B) and the stale F6 (63 B).
   - Move the rationale F4 and F7 (445 B) to `state-file-rationale.md`.
   - Run-time part left: **14.2 KB of 21.3 KB**; 7,061 B saved per session.
   - `stats/internal/fallback/journal_test.go:49` pins the key-derivation lines. Its path must move too.
3. **pipeline.md** (all three sessions, 5.6-6.5 KB each).
   - P18, Hand-verifying (1,818 B): load it only when a gate verdict fires.
   - P3, Stage exit (894 B): move to brainstorm-planner.md.
   - P7 (487 B): move to implement.md.
   - P8 and P13 (330 B): move to flow-status.
   - Cut or move the rest: P16 misplaced (705 B), P6 and P19 duplicates (1,223 B), P2-P15 rationale (1,343 B), P1 and P4 stale (346 B).
4. **SKILL.md** (all three sessions, ≈5.3 KB each).
   - Cut the Guardrails and Model-resolution bullets its phase files already state: K1, K6, K7, K10, K12-K17 (2,963 B).
   - Move out of the router: the K3 Stage-keys index (1,320 B), K5 `VERIFY_MODEL` (620 B, to visual-verify.md), K4 (274 B, to rationale).
   - Cut K2 (110 B, stale).
5. **commands-claude/flow.md** (all three sessions).
   - Cut C1-C5: 2,186 B of 2,915 B.
   - These passages restate SKILL.md and pipeline.md, and four of them contradict the contract (Drift 1-3, 5).
   - High confidence. It is also a bug fix.

Next in line:
- operator-prompts.md O1+O3 (4.9 KB; for this operator, finish sessions only).
- The always-on `flow-manual-review.mdc` core, A1 (2,292 B). Every non-flow session in every project pays it; removing it is your design decision.
- model-policy.md duplicates M2-M6 (1.9 KB; implementation sessions).

## Drift / bugs found

1. **The command file contradicts the pipeline** on what a creating run does.
   - commands-claude/flow.md:10-14 says the run continues through implementation to `IN_PROGRESS`.
   - pipeline.md:80 and SKILL.md:136-139 say it ends at `STARTED`, at the plan gate, with a `/clear` handoff.
   - flow.md is the first file a `/flow` session reads. pipeline-rationale.md:33-36 states the rule this breaks: when a command and its skill disagree, "whichever the agent reads first wins".
2. **The command file resolves the change name the wrong way.**
   - flow.md:34-35 says to use `spectre list --json`.
   - pipeline.md:461-466 requires `flow state resolve` and forbids "a directory listing of your own".
   - pipeline-rationale.md:247-250 records why: `spectre list` misses a change staged only in a worktree, which is the normal state at `IN_PROGRESS`.
3. **flow.md:22-24 says the roster comes from "the settings store's reviewer list".** It is the recorded decision's roster; the store list applies only to a `micro` decision (SKILL.md:204-205, review-panel.md:7).
4. **flow.md:27-31 has two stale claims.**
   - It calls `flow-manual-review.mdc` a file the harness resolves. It is rendered into `~/.claude/CLAUDE.md`.
   - It says pipeline.md is canonical for git boundaries and the finish contract. Both have moved out (pipeline.md:18, 429-435).
5. **"Run the apps" versus "the stack is running".**
   - `flow-manual-review.mdc`:15-16 (rendered 191-192) and flow.md:40 tell the operator to run the apps.
   - pipeline.md:30-31, 38-39 and 47 say the handoff has already started the stack.
6. **The always-on diagram adds a line pipeline.md's diagram lacks.** `flow-manual-review.mdc`:14 has `(reachability end) → FINISHED (withdrawn)`; pipeline.md:28-33 does not (its table at :80 does).
7. **"Open a PR by default" ignores the configured landing route.**
   - `flow-manual-review.mdc`:25-26 and flow.md:18 say the run asks how to land the branch.
   - integrate.md:138-146 skips the question when `## default landing route` resolves. This repo's value is `merge and push`, so it is never asked.
8. **pipeline.md forbids arguments that the router accepts.**
   - pipeline.md:70-74 says the only argument is the change name, and anything else is "reported".
   - SKILL.md:27-28 and flow.md:33-37 accept a description, a Jira key, or fix instructions.
   - pipeline.md:9-10 says pipeline.md wins any disagreement. Read literally, every `/flow <fix>` would be reported.
9. **SKILL.md:230-232 says only run 2 writes `FINISHED`.** The withdrawal route also writes `FINISHED` with `withdrawn: true` (brainstorm.md:179-181, pipeline.md:48, 80-81).
10. **SKILL.md:191-194 has two broken references that the guard cannot see.**
    - `flow.state-gate` is a stage key nowhere: not in README Level 1, not in the Stage keys table, not in stats/internal/stages.
    - The cited section **The `<change>` argument is always a resolved change name** (pipeline.md) does not exist. The sentence lives only in pipeline-rationale.md:89, which no run loads.
    - `check-references.sh` misses it because the bold span crosses lines 193-194.
11. **The guard-presence check misses `lib/flow-guard.sh` for 24 guards.**
    - SKILL.md:185-188 names two flow-guard shims. 26 of the 59 entries in skills/flow/scripts/ source `lib/flow-guard.sh`.
    - pipeline.md:396-399 derives sibling files by grepping for `$SCRIPT_DIR/<name>`. The shims use `$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh`, which that grep does not match.
12. **pipeline.md contradicts itself on implementation-branch task entries.** pipeline.md:150-153 says "stages on the implementation branch". pipeline.md:155 and SKILL.md:23 say one entry per task.
13. **SKILL.md:24-25 says `/flow` uses "the same granularity `/flow-fast` used".** /flow-fast registers one entry per file or logical unit (flow-fast/SKILL.md:167-168).
14. **pipeline.md:35-36 says "Each command ends in the state named after it".** That is left over from the old start/do/finish commands; there is one pipeline command now.
15. **pipeline.md:68-70 lists the commands and omits `/flow-fast` and `/flow-self-review`.** commands-claude/ holds six.
16. **pipeline.md:98-100 refers to "the test guide".** No such term exists anywhere else in the repo.
17. **pipeline.md:468 and 472 say `state list`** where L461 prescribes `flow state resolve`. `resolve` wraps `list`, so this is naming only.
18. **state-file.md:172-173 says `models.default`'s "live consumer is `/flow`, which dispatches on that value".**
    - `/flow` never reads it: SKILL.md:51-72 resolves the model from the project key and the store.
    - verify-and-handoff.md:338-340 says it is "always `null`".
19. **model-policy.md assumes a creating run still asks a model question.**
    - model-policy.md:6 says the file is loaded "at the model questions".
    - model-policy.md:64-66 says `models.default` records "only a model an operator explicitly chose".
    - No run asks (brainstorm.md:60-61), so nothing writes it.
20. **The `## decisions` row claims a wider auto-resolve scope than the operator-prompts contract allows.**
    - project-configuration.md:35 says `recommended` covers "every prompt the run would otherwise ask, whatever phase it sits in".
    - operator-prompts.md:53-61 and 75-96 lift the planning asks only. The pivot, outward-facing or irreversible actions, and every integrate or archive prompt stay asked.
    - This drift points toward auto-resolving pushes and merges. Fix it before any split that leaves a finish session holding this row but not **Auto-resolution**.
21. **CLAUDE.md:28-29 cites a file that does not exist.**
    - `CONTRIBUTING.md` is absent from the tree and from all git history.
    - The same paragraph, at lines 13-14, says the policy "is not restated here", then lines 28-29 restate it.
22. **CLAUDE.md doubles as a template but carries agents-repo-only content.**
    - `setup.sh:223-224` copies this file verbatim into any project that lacks a `CLAUDE.md`.
    - It carries this repo's lint commands (L19-26) and `## stop` / `stats/README.md` pointers (L45-48).
    - Its own template comment (L54-63) says it "ships generic on purpose".
23. **flow-plan/SKILL.md:38 cites **Stage keys** (SKILL.md) for the `mf-<literal-token>` convention.** The token rule is at SKILL.md:197-200 and pipeline.md:187-210.
24. **Some rules live only in rationale files, which no run loads.**
    - pipeline-rationale.md:107-110: "Do not branch on `flow stage`'s exit code". pipeline.md:212-215 does not say this.
    - pipeline-rationale.md:85: "The token is per run … never reused across runs". SKILL.md:197-200 partly covers it.
    - pipeline-rationale.md:101: what `<literal-token>` means. SKILL.md:197-200 partly covers it.
25. **SKILL.md:225 mentions a "preset".** No preset exists anywhere in skills/ (grep).
26. **Stale citations outside the run load set:**
    - `scripts/prepare-workspace.sh:266`: "the registry in skills/flow-contracts/pipeline.md" (it is in artifacts-registry.md).
    - `stats/internal/store/settings.go:16,59,75`: SKILL.md's self-review-model resolver (it is now archive.md step 9).
    - `stats/internal/records/render.go:34`: marker format "stated in skills/flow/SKILL.md" (it is in review-panel.md).
    - `stats/internal/harvest/watcher_test.go:2227`: "as skills/flow/SKILL.md emits" (the router emits no marks, per stagemarkcalls.go:39).

## Not slimmable

- **SKILL.md:130-152, Reading the state (1,348 B).** The router's dispatch table; every session.
- **SKILL.md:154-176, A plain message at IN_PROGRESS (1,450 B).**
  - A later plain prompt triggers it with no `/flow`, so the text must already be in context. It matters in implementation and finish sessions.
  - Moving it would need `hooks/flow-active-change.py` to inject a load pointer. The no-hook path (SKILL.md:159-160) would then lose it.
- **SKILL.md:47-103 and 123-128, the Model resolution block and source rules.**
  - Planning uses them for the Decide `resolved` object and the `models:` line, which prints `REVIEWERS` (brainstorm-planner.md:610). Implementation dispatches on them; finish prints them in the summary line.
  - `check-model-resolution-shell.sh` extracts the bash block from under `## Model resolution`.
  - 17 citers, including /flow-fast and /flow-plan.
- **pipeline.md core sections:** States (less P1 and P2), State transitions, Wrong state, Handoff output (less P14), Summary and live-stack, Change name resolution. Every session's router and handoff use them.
- **pipeline.md:107-120, the bare-invocation rule (1,270 B).** It is what the `IN_PROGRESS` human gate rests on.
- **pipeline.md:187-189 and 209-210, the two `MUST` sentences, and the mark example at 201-207.**
  - These are the corpus's only `MUST` sentences in these files; the normative-inventory guard tracks them.
  - pipeline.md is in `check-stage-mark-calls`' candidate set (stagemarkcalls.go:72-76).
- **pipeline.md:257-269.** git-boundaries.md:87 delegates the definition of "the planning path" to it.
- **pipeline.md:304-330, Artifact brevity.** Its list of fields never compressed is parsed byte for byte by guards.
- **The state-file.md run-time part (14.2 KB).** CLI exit semantics, record shape, closed vocabulary, omission-clears, the field rules, carry-forward and multi-repo keys. Every write site depends on it.
- **project-configuration.md:21-41, the key table (10,455 B).**
  - It is the canonical key list, and every row is normative.
  - Some rows serve one session or one command: `## handoff` (675 B) is read by /flow-fast alone; `## known failures` (1,126 B) only in implementation. Splitting the table by row is not recommended.
- **operator-prompts.md:1-14, The shape.** Every session.
- **Always-on block: the other rules.** never-touch-production, lint-fix-priority, commit-scope, design-mockups, context7, dependency-versions, fix-determinism and build-the-simplest are general and normative, apart from the rationale rows A2-A9.
- **The "See **X** (…-rationale.md) for why …" pointer sentences.**
  - pipeline.md has 9 (1,461 B), model-policy.md 7 (743 B), project-configuration.md 16 (2,443 B, mostly inside G1-G3).
  - They are the repo's breadcrumb convention. Deleting them is a convention change, not a trim; your decision.

## Answers to the specific questions

### A. Blocks in SKILL.md and pipeline.md that only one session uses

Sessions: P = planning, I = implementation, F = finish.

| block | file:lines | bytes | only session | disposition / target |
|---|---|---:|---|---|
| `VERIFY_MODEL` paragraph | SKILL.md:109-115 | 620 | I, and only when `ui paths` matched | move to visual-verify.md (its sole consumer, :15) |
| `DEFAULT_MODEL` four roles | SKILL.md:117-121 | 417 | I | cut: each dispatch site states its model; or move to implement.md:31 |
| Never publish a proposal artifact | SKILL.md:207-208 | 116 | P | cut: brainstorm.md:60-62 |
| Design gate / thin scaffold | SKILL.md:209 | 79 | P | move to brainstorm-planner.md |
| No slot beyond the roster | SKILL.md:210-217 | 666 | I | cut: review-panel.md:168-174, 241-245 |
| At most three implementers in flight | SKILL.md:218-220 | 274 | I (`big` only) | cut: implement.md:550-558 |
| Bundled panel dispatch | SKILL.md:221-224 | 354 | I | cut; move its "re-check before any round" sentence to review-panel.md's Bundled dispatch |
| Open finding / stale result | SKILL.md:225-227 | 237 | I | cut: review-panel.md:928-931, 1351 |
| Parent orchestrates | SKILL.md:233-237 | 375 | I | cut: implement.md:19-24 |
| `SELF_REVIEW_MODEL` not resolved here | SKILL.md:105-107 | 274 | none (F reads archive.md) | move to SKILL-rationale.md |
| Stage exit | pipeline.md:54-64 | 894 | P | move to brainstorm-planner.md beside **Convergence** |
| Per-task granularity | pipeline.md:155-160 | 487 | I | move to implement.md |
| Finish contract pointer | pipeline.md:429-435 | 397 | F | cut: integrate.md:7 and archive.md:9 already name their contracts |
| /flow-status registers or marks nothing | pipeline.md:162-163, 221-222 | 330 | none (/flow-status only) | move to flow-status/SKILL.md |

These blocks are used by two sessions, so they cannot move into one phase file:
- pipeline.md:405-427, Hand-verifying (1,818 B), used by I (the `MOVED` verdict at review-panel.md:28) and F. Instead, load it only when a gate verdict fires.
- SKILL.md:154-176, the plain message (I+F), must stay resident.
- pipeline.md:107-120, the bare-invocation rule (I+F).
- pipeline.md:332-348, IntelliJ (P+I).
- pipeline.md:304-330, Artifact brevity (P+I).

### B. state-file.md and project-configuration.md: run-time versus reference

**Yes: the "load it before …" directives force a full load in every session.**
- state-file.md:
  - pipeline.md:440-441 says to load it "before reading or writing a state file".
  - Every session reads state in the router (SKILL.md:133).
  - Every session writes state: brainstorm.md:38-52 (`STARTED`); implement.md:242-252 and verify-and-handoff.md:331 (`IN_PROGRESS`); integrate.md:255-256 and archive.md:157 (`FINISHED`).
- project-configuration.md:
  - pipeline.md:447-448 says to load it "before resolving project configuration".
  - Every run's Model resolution calls `project-get.sh … 'model'` (SKILL.md:54) before any dispatch.
- Both directives are unconditional and whole-file, sit in pipeline.md's always-loaded tail, and fire at the router before any phase file.
- rules/flow-manual-review.mdc:53-56 repeats them. That part is outside the rendered core, so it reaches only readers of the full rule.

**Every write site already spells out what to write and what to carry forward:**
- brainstorm.md:44-62 and 174-184
- implement.md:242-252
- verify-and-handoff.md:331-342
- integrate.md:255-256
- archive.md:157-160 and finish-contract-run2.md:208-211

**state-file.md, what a run needs (14.2 KB of 21.3 KB):**
- lines 1-16: CLI, `-C`, and the record key
- lines 58-91: never-block and the exit codes, less F4's 237 B
- lines 93-116: the record
- lines 123-135: closed vocabulary and omission-clears
- lines 137-209: the fields, less F6, F7 and F8
- lines 211-217: monotonic refusal
- lines 228-236: carry-forward
- lines 267-291: the multi-repo record
- lines 300-310: read and write

A tighter cut, moving the reference halves of paragraphs 76-86, 146-157, 188-197, 198-208 and 267-277 sentence by sentence, gets to about 11.5-12 KB, at more risk.

**state-file.md, reference only (5.8 KB):**
- lines 18-25: this machine's fallback paths
- lines 33-56: key derivation, which the CLI performs itself
- lines 118-121: `mainCheckoutPath`
- lines 219-226: ordering internals
- lines 238-265: journal replay (daemon behaviour)
- lines 293-298: the worked example

**project-configuration.md, what a run needs:**
- **Core, every session (13,007 B):**
  - lines 1-42: intro and key table
  - lines 89-96: single-line literal rule
  - lines 131-134: Roots in `## apps`
  - lines 548-556: the file is optional
- **Implementation adds:**
  - G8+G9 (5,610 B): standards, containment and `<agents repo>`, loaded at the principles dispatch (review-panel.md:666-671)
  - G6 (8,166 B), only when `ui paths` matched
- **Finish adds:** G4+G5 (2,733 B) for run 2's `remove` (finish-contract-run2.md:133-135).
- **Read by no run:**
  - G1-G3 (22,966 B), the workspace-isolation authoring and validator spec, except when `prepare-workspace.sh` cannot be located (verify-and-handoff.md:49). `prepare-workspace.sh`, `check-workspace-isolation.sh` and `check-cleanup-complete.sh` execute it.
  - G7 (1,349 B) duplicates the table row at line 40 and review-panel.md:90-107.
- **The existing `project-configuration-authoring.md` (1,971 B) is the natural home for G1-G3.** It already says "nothing here is consulted by any run".

### C. The always-on block, by core section

| section | source | rendered | bytes (rendered) | classification |
|---|---|---|---:|---|
| flow — /flow, start to finish | flow-manual-review.mdc:8-44 | 184-222 | 2,292 | **flow-only** (A1): every non-flow session in every project pays it. In flow sessions it duplicates SKILL.md:15-18, pipeline.md:28-52 and 429-435, and flow.md:27-31, with drift (Drift 5-7). Sub-parts: diagram :10-17 (576 B), gate prose :20-23 (393 B, including rationale :21-22, 86 B), landing :25-28 (308 B), load-pipeline :35-41 (548 B) |
| Be brief — non-repetition + cut-never-paraphrase | be-brief.mdc:14-25 | 15-26 | 945 | scope limited to agents-repo Markdown and /flow artifacts (A8, low: it is the rule this trim relies on) |
| Every rule here applies at every depth | dispatch-carries-the-baseline.mdc:9-13, 20-22, 24-25 (partial) | 134-138, 145-147, 149-150 | 427 + 220 + 111 | RATIONALE (A2-A4) |
| Fix determinism | fix-determinism-at-the-source.mdc:12-14 (partial) | 168-170 | 138 | RATIONALE (A5) |
| Dependency versions | dependency-versions.mdc:10-11 (partial) | 101-102 | 54 | RATIONALE (A6) |
| Context7 | context7.mdc:13 (partial) | 87 | 54 | RATIONALE (A7) |
| Build the simplest thing | build-the-simplest-thing.mdc:10-11 (partial) | 54-55 | 144 | RATIONALE, low (A9) |
| Lint Fix Priority | lint-fix-priority.mdc:9, 13 (partial) | 229, 233 | 36 + 49 | DUPLICATE (within the rule) and RATIONALE, low; not counted |

Everything else in the block is general and normative. Edits go in the `.mdc` files, followed by `setup.sh global` (`check-installed-rules.sh` detects drift).

### D. Duplicates within the every-session load set

| # | copy | canonical | drift |
|---|---|---|---|
| 1 | always-on flow-manual-review.mdc:10-17 (diagram) | pipeline.md:28-33 | yes: "run the apps"; extra withdrawn line |
| 2 | always-on flow-manual-review.mdc:20-23 | pipeline.md:38-42, 50-52, 92, 122-123 | no |
| 3 | always-on flow-manual-review.mdc:25-28; flow.md:18-19 | pipeline.md:84, integrate.md:138-146 | yes: ignores the configured landing route |
| 4 | always-on flow-manual-review.mdc:35-41; flow.md:27-31 | SKILL.md:15-18; pipeline.md:429-435 | flow.md is stale (Drift 4) |
| 5 | SKILL.md:27-28; flow.md:33-37 | pipeline.md:66-74 | yes (Drift 8) |
| 6 | flow.md:9-19 | pipeline.md:24-90; SKILL.md:130-176 | yes (Drift 1) |
| 7 | flow.md:21-25 | SKILL.md:204-217, brainstorm.md:60-62 | yes (Drift 3) |
| 8 | flow.md:39-42 | pipeline.md:237-249 | yes (Drift 5) |
| 9 | SKILL.md:20-25 | pipeline.md:146-160 | yes (Drift 12-13) |
| 10 | SKILL.md:197-200 | pipeline.md:187-189 | no; SKILL.md adds placement and scope, so keep both |
| 11 | SKILL.md:178-183 | pipeline.md:371-403 | no |
| 12 | SKILL.md:228-229 | pipeline.md:257; git-boundaries.md | no |
| 13 | SKILL.md:230-232 | pipeline.md:122-123 | yes (Drift 9) |
| 14 | CLAUDE.md:28-29 | always-on lint-fix-priority.mdc:13 | the file it names does not exist |
| 15 | CLAUDE.md:88-91 and 100-101 | CLAUDE.md:69-71 | no |
| 16 | pipeline.md:94-105 | pipeline.md:78-86 | slight: omits merge-and-push chaining; "test guide" |
| 17 | pipeline.md:1-5 ("See **Finish contract**") | pipeline.md:429-435 | no |

Rows 1, 2 and 4 are also paid by every non-flow session, because they sit in the always-on block.

### E. Citers and pins for each proposed move

`check-references.sh` checks a bold heading next to a path on the same line. Cross-line spans are flagged "×"; the guard cannot see them, but they should be repointed anyway.

| move | citers to repoint (`**Heading**` + path) | other pins |
|---|---|---|
| P3 Stage exit → brainstorm-planner.md | brainstorm-planner.md:150 | pipeline-rationale.md:13 heading follows by convention |
| P18 Hand-verifying → new lazy file | finish-contract-run1.md:88 ×, :167, :209; integrate.md:32, :63 | add the load directive at review-panel.md:28-64 (`MOVED`) and finish-contract-run2.md:160-178 (`LEFTOVER`). The new file cites no bold+path pair, so declare it in `crExpectedZero` (references.go:69-121) |
| P7, P8, P13 | none (the heading **Progress visibility** stays; SKILL.md:21) | — |
| K3 Stage keys → README.md | README.md:138, pipeline.md:88, flow-plan/SKILL.md:38 (already mis-cited) | stats/internal/stages/names_test.go parses README's Level 1 table from its heading onward; keep the moved table outside it, or use SKILL-rationale.md |
| K5 `VERIFY_MODEL` → visual-verify.md | visual-verify.md:15 (heading **Model resolution** stays, so the guard passes; repoint semantically) | check-model-resolution-shell.sh: `VERIFY_MODEL=opus` stays in the block |
| K4 → SKILL-rationale.md | none | — |
| G6 visual verification → sibling | visual-verify.md:675 (heading citer); :242, :285 and brainstorm-planner.md:46 cite the H1 | verify-and-handoff.md:151-152 (plain text); check-visual-verification.sh header; visualverification.go:272 comment |
| G1-G3 → authoring file | H1 citers that mean this content: workspace-isolation.md:113, 261, 275, 329, 342, 351; artifacts-registry.md:62, 69, 80 | scripts/prepare-workspace.sh:225 (error text cites **What a `url` row may reference, and what it may not**); cleanupcomplete.go:64, 803, 830 comments |
| G4-G5 → finish-contract-run2.md step 5 | finish-contract-run2.md:134 | — |
| G8-G9 → lazy standards file | principles-reviewer-prompt.md:166-175 (prose); review-panel.md:666-671 | none |
| G7 cut | 0 citers | — |
| F-rows → state-file reference sibling | state-file-rationale.md:8 (**The pipeline never blocks**) stays valid; in-file "see … below" at :86, :128, :226 | journal_test.go:49 `contractPath` (MAIN_CHECKOUT=/PROJECT_KEY=); stats/internal/fallback/statefile.go:9-47 comments |
| O3 Auto-resolution → new file | 14: project-configuration.md:35; flow-fast/SKILL.md:146; brainstorm-planner.md:116, 146, 641; brainstorm.md:155; implement.md:367, 390, 1141, 1156; review-panel.md:127, 1422; verify-and-handoff.md:424; SKILL-rationale.md:441 | dispatchparagraphs.go:67-69 pins phrases only, not the citation. Fix Drift 20 first |
| O1 multi-select → finish-contract-run2.md step 9 | 0 heading citers | — |
| M7 Harness mapping → zcode file | archive.md:206 ×; brainstorm-planner.md:520; implement.md:117, 507; review-panel.md:181, 1335; visual-verify.md:17, 96 | — |

**Checks that apply to every move:**
- Each new split file needs one bold-adjacent citation or a `crExpectedZero` entry. Otherwise `check-references.sh` fails it for zero coverage.
- Verbatim moves leave the sorted normative inventory unchanged. No proposed cut contains a `MUST` or `SHALL` sentence; the only two in these files, pipeline.md:187 and 209, stay.
- Re-run `check-guard-symlinks.sh`. P18 names five gate guards and would move them to another contract file.
