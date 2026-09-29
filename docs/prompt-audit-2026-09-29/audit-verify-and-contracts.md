# Slimming analysis — verify-and-handoff, visual-verify and five flow contracts

> **Audit snapshot at `700e184` (2026-09-29).** A read-only model pass that followed [`rubric.md`](rubric.md); every row is a candidate to verify, not a verified fact. Line numbers refer to `700e184`. Main has since retired the bugbot and security slots (`ebdfdde1`..`4dafab4a`), so re-locate each row by its quoted first words, and treat bugbot/security rows as obsolete. Tracked in KAN-851 → KAN-858 (and the Drift items in KAN-853 / KAN-854). Plan: [`README.md`](README.md).

Read-only analysis. Nothing in the repository was modified (`git status` clean).
Files: `skills/flow/verify-and-handoff.md`, `skills/flow/visual-verify.md`,
`skills/flow-contracts/{workspace-isolation,git-boundaries,artifacts-registry,session-records,worktree-resolution}.md`.
Byte counts: `sed -n 'A,Bp' | wc -c` for whole-line ranges, and an exact UTF-8 substring count for
partial-line passages (start words → end words, as the table's "first words" column shows).
Rationale targets: contract files move text to their existing `-rationale.md` siblings; the two
`skills/flow/` phase files have no own sibling, and their precedent target is
`skills/flow/SKILL-rationale.md` (it already holds `### verify-and-handoff.md — …` and
`### visual-verify.md — …` sections under "Moved by the 2026-09-22 prompt audit").

## Totals

A byte is counted once. Where passages overlap, precedence is LAZY-SPLIT/MISPLACED (the relocation)
over the finer lever inside it, except that RATIONALE inside VV-05 is split out (it would move
either way).

| file | bytes | RATIONALE | DUPLICATE | STALE | LAZY-SPLIT | MECHANICS | MISPLACED |
|---|---:|---:|---:|---:|---:|---:|---:|
| skills/flow/verify-and-handoff.md | 26,951 | 626 | 2,688 | 0 | 2,661 | 2,002 | 225 |
| skills/flow/visual-verify.md | 61,248 | 2,675 | 638 | 0 | 4,093 | 3,365 | 37,978 |
| skills/flow-contracts/workspace-isolation.md ¹ | 25,118 | 2,589 | 381 | 0 | 14,593 | 1,144 | 417 |
| skills/flow-contracts/git-boundaries.md | 11,565 | 0 ² | 0 | 0 | 5,721 | 0 | 0 |
| skills/flow-contracts/artifacts-registry.md | 7,010 | 93 | 1,154 | 0 ³ | 0 | 0 | 5,763 ⁴ |
| skills/flow-contracts/session-records.md | 1,879 | 0 | 0 | 183 | 0 | 0 | 0 |
| skills/flow-contracts/worktree-resolution.md | 1,877 | 0 | 303 | 0 | 0 | 0 | 0 |
| **total** | 135,648 | 5,983 | 5,164 | 183 | 27,068 | 6,511 | 44,383 |

¹ Loaded only conditionally (see "workspace-isolation.md — exactly when it loads"). LAZY-SPLIT here
is the part no load trigger ever needs (L18-181 + L310-360); WI-07…WI-11 lie inside it and are
counted there.
² GB-03/GB-04 (214 B RATIONALE/STALE) lie inside GB-01 and are counted under LAZY-SPLIT.
³ AR-06's stale design.md pointer is counted under DUPLICATE (same passage).
⁴ 684 B of in-file editor-only text (AR-03, AR-04) plus the 5,079 B remainder of AR-01: the whole
file is loaded by the implementation session, which acts on none of it. The finish session still
needs the file; AR-02…AR-08 (1,931 B) are the cuts that also apply there.

Practical effect per session (definite = every run of that kind):
- Implementation, ordinary run (no UI path, verify green, no `prUrl`, no cache-index row):
  ≈ 6.2 KB of verify-and-handoff.md cuts (RATIONALE + DUPLICATE + VH-04/VH-11/VH-19 not loaded + MISPLACED),
  + 7.0 KB (AR-01), + 2.2 KB (GB-01), + 0.5 KB (SR-01, WR-01) ≈ **16 KB**.
- Implementation, UI run: + ≈ 40.6 KB (VV-05) + 4.1 KB (VV-04, needs a Go change) + 0.6 KB (VV-01…03)
  ≈ **+45 KB** (up to ~2× VV-05 if the parent currently pastes steps 7–11 into the verifier prompt).
- Implementation, project with a `cache index` row: + ≈ 20.4 KB (WI-06).
- Planning session (git-boundaries.md cited at point of use): 2.2 KB (GB-01).
- Finish session: 3.5 KB (GB-02) + 1.9 KB (AR-02…08) + 0.5 KB (SR-01, WR-01) ≈ 5.9 KB.

## Candidates

`verify-and-handoff.md` = VH, `visual-verify.md` = VV, `workspace-isolation.md` = WI,
`git-boundaries.md` = GB, `artifacts-registry.md` = AR, `session-records.md` = SR,
`worktree-resolution.md` = WR. Paths below are repository-relative.

| id | file:startline-endline | bytes | lever | condition / canonical location / evidence | first ~10 words verbatim | confidence | behaviour-risk note |
|---|---|---:|---|---|---|---|---|
| VH-01 | skills/flow/verify-and-handoff.md:8-10 | 106 | DUPLICATE | Same Load directive already executed at skills/flow/implement.md:260-261 in the same session | **Load `skills/flow-contracts/worktree-resolution.md`** before resolving this run's worktree set, below. | low | implement.md:230's heading says "(first run only)", so a fix run may skip that load; only safe once that heading drift is fixed |
| VH-02 | skills/flow/verify-and-handoff.md:26-28 (from "— never a raw read" to "do not proceed.") | 195 | DUPLICATE | worktree-resolution.md:15-23 (loaded two lines above) | — never a raw read of the state file's `worktrees` | low | Local copy says "do not proceed", contract says "a gate stops and asks the operator" — cut would move behaviour to the contract's wording |
| VH-03 | skills/flow/verify-and-handoff.md:60-63 (from "`## lint`, `## test` and" to "the ledger render.") | 270 | DUPLICATE | implement.md:56-61 (closed list: `prepare-workspace.sh`, `## lint` and `## test` are the parent's own Bash); SKILL.md:109-111; implement.md:160-161 | `## lint`, `## test` and `check-spec-reach.sh <worktree>` **are the parent's | med | The bold is a reminder against a historic subagent dispatch for tests (KAN-389); canonical rule is still in the load set |
| VH-04 | skills/flow/verify-and-handoff.md:91-109 | 1,644 | LAZY-SPLIT | Condition: a `## lint`/`## test`/`check-spec-reach.sh` command exits non-zero (runtime, before the next action; precedent: L39's workspace-isolation load on a non-zero exit). False in most runs (est. 70–90%: the full suite already ran green after the last group; lint first runs here) | **Inline verify — a failing command.** When a command exits | low-med | Citers must be repointed by hand (guard will not catch): project-configuration.md:41, known-bugs.md:47, in-file L76-77 |
| VH-05 | skills/flow/verify-and-handoff.md:133-134 (from "`journalled: ledger`" to "exits above.") | 146 | DUPLICATE | session-records.md:17-28 (loaded at L119) | `journalled: ledger` and a non-zero exit are reported the same | low-med | "unlike the lint and test exits above" contrast is lost; the table already says none of them stops |
| VH-06 | skills/flow/verify-and-handoff.md:221-223 (from "`MISSING: <kind>` means" to "committing the fix.") | 223 | DUPLICATE | session-records.md:20-24 | `MISSING: <kind>` means the store holds no rows of that | med | Moves with VH-11 if VH-11 is done |
| VH-07 | skills/flow/verify-and-handoff.md:151-153 (sentence) | 76 | RATIONALE | Editor meta ("nothing else restates it") — no action | This stage owns its procedure — nothing else in this | high | none |
| VH-08 | skills/flow/verify-and-handoff.md:169-170 (sentence) | 176 | RATIONALE | Editor meta; the guard's header is canonical for glob semantics | `check-visual-trigger.sh` owns the glob semantics (`**` spanning directories, a leading | med | Slight: it tells the parent not to re-judge globs; the "with the guard, not by eye" heading of step 2 already does |
| VH-09 | skills/flow/verify-and-handoff.md:191-193 (" From here to the handoff, …" through `<!-- refs-guard:allow -->.`) | 181 | DUPLICATE | implement.md:1084-1086 (Turn discipline: begin rides the first command, end the last) | From here to the handoff, each stage's `begin` mark rides | med-high | none; Turn discipline binds every stage already |
| VH-10 | skills/flow/verify-and-handoff.md:202-204 (first two blockquote sentences; keep "This step only confirms nothing slipped in.") | 194 | DUPLICATE | SKILL.md:228; pipeline.md:257-262; git-boundaries.md:94-97 | **`<project>/spectre/changes/` is never part of a task commit.** `<project>/spectre/specs/` | med-high | Keep the third sentence so the blockquote still scopes the `git status`/`git log` check |
| VH-11 | skills/flow/verify-and-handoff.md:206-207 + 212-221 (Load git-boundaries.md; "On that path only …" through "The render overwrites in place.") | 654 | LAZY-SPLIT | Condition: the state file records a `prUrl` (known at run start). False on every merge-and-push-route change (operator default) → ~always false | **Load `skills/flow-contracts/git-boundaries.md`** before committing below. **The PR-branch split.** Every task, | med | Keep L208-212's first two sentences (they state the branch). Prose citers only: finish-contract-run1.md:325-326, session-records.md:5,14, state-file.md:185-187, stats/cmd/flow/record.go:2412 (comment) |
| VH-12 | skills/flow/verify-and-handoff.md:237-239 | 205 | DUPLICATE | pipeline.md:253-256 ("Every path is absolute — … in run instructions … never a main-checkout path … Resolve app roots from `git worktree list` or the state file's `worktrees` keys") | - **Every app root is absolute**, resolved from `git worktree | high | none — the canonical bullet covers every clause |
| VH-13 | skills/flow/verify-and-handoff.md:250-253 (sentence "A fix run's motivation … runs another.") | 256 | RATIONALE | Motivation for "start the stack on every run"; its measured provenance already sits at SKILL-rationale.md:207-209 | A fix run's motivation is the sharpest example: it hands | low | The 2026-09-22 audit quoted it as the rule's attribution, i.e. treated it as rule text |
| VH-14 | skills/flow/verify-and-handoff.md:260-261 (sentence) | 74 | RATIONALE | Design history for editors ("no new key is added") | No new project-configuration key is added for this, and none | high | none |
| VH-15 | skills/flow/verify-and-handoff.md:333-336 (from "`flow.kickoff` writes the first" to "deferring to here — ") | 267 | DUPLICATE | implement.md:242-258 (persist each worktree the moment it exists; kickoff entry at brainstorm.md step 3) | `flow.kickoff` writes the first worktree's entry the moment it is | med | Cut leaves "…and its merge base — so this step re-reads…", grammatical |
| VH-16 | skills/flow/verify-and-handoff.md:339-341 | 269 | DUPLICATE | state-file.md:164-185, 228-236 (carry every unowned field forward verbatim) | Carry `artifactUrl` (always `null` under `/flow`), `jiraIssue`, `planningEffort` (always `null`), | low | Copies disagree (see D7): cutting leaves state-file.md's stale "`/flow` dispatches on `models.default`" as the only word — fix state-file.md first |
| VH-17 | skills/flow/verify-and-handoff.md:419-421 (from " — the same fields and shape" to "the same state") | 152 | MISPLACED | `/flow-status`/editor sync note; the implementation session never loads handoff-blocks.md | — the same fields and shape `skills/flow-contracts/handoff-blocks.md`'s `Panel:` line carries | high | none |
| VH-18 | skills/flow/verify-and-handoff.md:428-430 (sentence "Every screenshot path …") | 178 | DUPLICATE | pipeline.md:253 (every path absolute, in handoffs) | Every screenshot path in it is absolute, per **Handoff output** | med | none |
| VH-19 | skills/flow/verify-and-handoff.md:433-436 | 363 | LAZY-SPLIT | Condition: a UI fix run dispatched a tooling analyst (the template line itself always prints "none — no miss"). False in ~95%+ of runs. Could ride VV-04's file | **The `Tooling analysis:` line reports **A missed defect — the | low | Keep "the run itself never edits `visual-verify.md`" wherever it lands |
| VH-20 | skills/flow/verify-and-handoff.md:438-440 | 206 | DUPLICATE | jira-integration.md:214-216 ("Echo the pre-edit description into the handoff, verbatim in a fenced block"), loaded by implement.md:432-433 exactly on the runs this applies to | The pre-edit description line is present only on a fix | med | none |
| VH-21 | skills/flow/verify-and-handoff.md:442-444 | 248 | DUPLICATE | SKILL.md:233-237; implement.md:21-24, 123-124 | **The parent prints this block directly, as this stage's own | med | "in the same turn" nuance is only implied canonically |
| VH-22 | skills/flow/verify-and-handoff.md:446-448 | 73 | MISPLACED | No step in this file creates a worktree; creation is brainstorm.md step 3 and implement.md §2 | ## Guardrails - **Never** create a second worktree for the | low | Move verbatim to implement.md §2 (same session → no byte saving there) |
| VH-23 | skills/flow/verify-and-handoff.md:79 (fence info " verified:design.md section 2 of this change") | 44 | RATIONALE | Authoring provenance; "this change" is an archived change; no guard reads skill fence tags (check-plan-provenance scans change dirs only) | verified:design.md section 2 of this change | low | none |
| VH-24 | skills/flow/verify-and-handoff.md:348-372 | 853 | MECHANICS | Three per-worktree `flow record` calls + the deferred filter/format could be one `flow record handoff-lines -change <name> -C <wt>` printing `Records:`/`Costs:`/`Deferred:` and `### Deferred minors` | **Produce the handoff's `Records:` count**, one call per affected worktree: | med | Needs CLI code + parity tests |
| VH-25 | skills/flow/verify-and-handoff.md:156-158 | 265 | MECHANICS | `check-visual-trigger.sh` exit 2 already covers "section/ui paths absent" (its header L14-23); step 1's by-hand read could be its exit-2 stderr | 1. **Resolve the section** — read that worktree's own `<project>/.flow/project.md` | low-med | Guard's exit 2 conflates "not configured" and "cannot answer"; needs a distinct stderr token |
| VH-26 | skills/flow/verify-and-handoff.md:257-261 (less VH-14) + 297-302 | 884 | MECHANICS | Stop/run resolution and `Running:` extraction (http lines + stop command) are scriptable (`resolve-run-stack.sh`-style) | Resolve the start from the two keys the project already | low | Protected-service and refused-start rules must stay prose or be encoded exactly |
| VV-01 | skills/flow/visual-verify.md:18-22 (" Its prompt carries, verbatim:" + baseline blockquote) | 211 | DUPLICATE | Always-on global block, from rules/dispatch-carries-the-baseline.mdc:15-18; enforced by hooks/enforce-agent-baseline.py. No other `/flow` dispatch site (implement.md, review-panel.md) carries it | Its prompt carries, verbatim: > Before anything else, read `~/.claude/rules/agent-baseline.md` | med | Must cut the colon lead-in with the quote (check-markdown-integrity's trailing-colon signal) |
| VV-02 | skills/flow/visual-verify.md:28-29 (sentence) | 90 | DUPLICATE | visual-verify.md:48-49 (MODEL HANDSHAKE paragraph, same file) | The first line of its first reply is `Model: <the | high | none |
| VV-03 | skills/flow/visual-verify.md:75-78 (after "unchanged:") | 337 | DUPLICATE | implement.md:94-113 ("The handshake — stated once here, cited everywhere else") | a first mismatch closes `<key>` `-outcome fallback` and re-dispatches once | med | Copies differ (D1): cut switches the second-mismatch stop from `## Question` to implement.md's AskUserQuestion; decide the canonical form first |
| VV-04 | skills/flow/visual-verify.md:94-155 | 4,093 | LAZY-SPLIT | Condition: fix run with ≥1 miss (classified by L85-92 before the verifier dispatch). False on every first run and most fix runs (~95%+) | `subagent_type: flow-high` (`agents/flow-high.md`, effort `high`), the Agent tool's `model` parameter | med | Blocked by dpSites (visual-verify.md min 2 for TOOLS/MODEL HANDSHAKE/NO DELEGATION) → Go + test change. Heading citers: implement.md:54, SKILL.md:121, verify-and-handoff.md:433-434 |
| VV-05 | skills/flow/visual-verify.md:210 + 228-694 + 728-755, less L425-428 and L565-568 parent clauses (292 B) and VV-06…VV-16 (2,631 B) | 37,978 | MISPLACED | Verifier-only text the parent loads: step 4, steps 7–11, the `## Report` template (gross 40,609 B). Move verbatim to a verifier-read file whose absolute path the prompt carries (L162-168 "run steps 4 and 7–11 below as written" → names that file) | 4. **Run `setup`, if declared.** A non-zero exit blocks, printing | med | Parent keeps steps 3, 5, 6, 12, 13, Blocking + the two reconciliation clauses. Update analyst task text (L128, L131), verify-and-handoff.md:379, :436, implement.md:424, :1121-1124 |
| VV-06 | skills/flow/visual-verify.md:241 (" (KAN-747)") | 10 | RATIONALE | Ticket provenance | (KAN-747). | high | none |
| VV-07 | skills/flow/visual-verify.md:243-247 (sentence "A clip is the right tool …") | 372 | RATIONALE | Why captures are never clipped | A clip is the right tool for an implementer's own | low | Carries a scope note (clips fine for an implementer's own assertion) |
| VV-08 | skills/flow/visual-verify.md:341-344 (": a verification that … exists to replace") | 270 | RATIONALE | Why the state sweeps are default | : a verification that transcribes no text run > (sweep | med | Leaves "…composed or not." — clean |
| VV-09 | skills/flow/visual-verify.md:436-437 (": two captures … show the drift") | 166 | RATIONALE | Why states are diffed against each other | : two captures can each match their own frame and | low | In-sentence reason (audit precedent keeps these) |
| VV-10 | skills/flow/visual-verify.md:450-451 (sentence) | 120 | RATIONALE | Editor cross-reference to rules/design-mockups-are-specs.mdc | `rules/design-mockups-are-specs.mdc` puts the same question to the implementer; this is | med | none |
| VV-11 | skills/flow/visual-verify.md:499-502 (sentence "Nothing else here sees it …") | 306 | RATIONALE | Why sweep 9 exists; whole sentence of reasoning | Nothing else here sees it — the composite's ratio barely | med | none |
| VV-12 | skills/flow/visual-verify.md:507-510 (two sentences) | 364 | RATIONALE | Why sweep 10 exists | Every sweep above starts from something the capture shows or | med | none |
| VV-13 | skills/flow/visual-verify.md:528-535 (two sentences) | 652 | RATIONALE | Why the element × property matrix exists | Every sweep above is anchored on one kind of element | low-med | The next sentence was quoted as the rule by the 2026-09-22 audit; keep it |
| VV-14 | skills/flow/visual-verify.md:641-643 (sentence) | 170 | RATIONALE | Why gaps are measured per side | Gaps compound where sizes do not: an outer container's content | low | none |
| VV-15 | skills/flow/visual-verify.md:656-658 (parenthetical) | 157 | RATIONALE | Test provenance; `TestMeasureVisualProperties` exists at stats/internal/guard/measure_visual_properties_test.go:149, unreadable from a consuming project | (both shown on the incident's own geometry by `TestMeasureVisualProperties` in | high | none |
| VV-16 | skills/flow/visual-verify.md:728 (fence info " verified:design.md section 3 of this change") | 44 | RATIONALE | Authoring provenance (as VH-23) | verified:design.md section 3 of this change | low | none |
| VV-17 | skills/flow/visual-verify.md:708-709 (" — see `no-automatic-push` (design.md)") | 44 | RATIONALE | Decision-id provenance; the id lives only in spectre/changes/archive/kan-171-generic-visual-verification-step/design.md — "design.md" in a run means the current change's (also STALE) | — see `no-automatic-push` (design.md): a file inside a repository cannot | high | Leaves "**Never push**: a file inside…" |
| VV-18 | skills/flow/visual-verify.md:170-209 | 3,365 | MECHANICS | Step 3's four checks (port probe via lsof/URL, base-URL literal, allowed origins, Playwright resolution) → a `check-visual-preflight.sh` guard with a call + exit contract | 3. **Pre-flight the workspace, before anything is dispatched.** The verifier's | med | Checks 2–3 need project-specific file discovery; parity tests required |
| WI-01 | skills/flow-contracts/workspace-isolation.md: 31 sentences at L5, 9, 14, 28, 59, 67, 71, 75, 85, 94, 98, 101, 124, 141, 155, 166, 172, 178, 198, 220, 223, 232, 238, 265, 302, 305, 313, 320, 330, 352, 358 | 4,772 (1,576 counted here; 131 inside WI-05; the rest inside WI-06's ranges) | RATIONALE | Editor-only "See **…** (`…workspace-isolation-rationale.md`) for why …" pointers; the file lacks the one-line "a `/flow*` run never loads it" boilerplate every sibling contract has, so the pointers invite a run to open the rationale | See **The workspace id** (`skills/flow-contracts/workspace-isolation-rationale.md`) for why a second | med | Deletion, not a move (the reasons already live in the rationale file); a repo convention question for the dispatcher |
| WI-02 | skills/flow-contracts/workspace-isolation.md:189-196 | 664 | RATIONALE | Rejected alternative (derive the index; the six-percent argument) → rationale "## The cache index" | **The reason is the size of the space.** A cache | med | L197 "That argument holds only…" then reads without its antecedent (readability only) |
| WI-03 | skills/flow-contracts/workspace-isolation.md:225-228 (two sentences) | 349 | RATIONALE | Why no expiry | Identity replaces it: a claim naming its holder can be | med | none |
| WI-04 | skills/flow-contracts/workspace-isolation.md:241-245 | 417 | MISPLACED | Advice to project authors; states "None of that is required here" — no run action | **A project whose claim is visible can do better than | med | none |
| WI-05 | skills/flow-contracts/workspace-isolation.md:237-240 | 381 | DUPLICATE | artifacts-registry.md:32, 78-83 (loaded in the same session by implement.md:232) | **This pipeline releases nothing at finish, and the claimed index | low-med | none |
| WI-06 | skills/flow-contracts/workspace-isolation.md:18-181 + 310-360 (whole file 25,118; each trigger needs one section) | 14,593 | LAZY-SPLIT | Section-scoped load: cache-index stderr → L182-247 only; exit 1 → L248-309 only; exit 2 → nothing. No trigger needs The workspace id / What the id derives / Creation and cleanup (the id is `flow workspace-id`/prepare-workspace.sh; cleanup is run 2, which does not load it) | ## The workspace id **An apply worktree has a workspace | med | Change only the directive at verify-and-handoff.md:39-41 (a `grep -n` + `sed -n` section read, as Read discipline already does). Do NOT split the file physically (test reads its fenced blocks; ~20 citers) |
| WI-07 | skills/flow-contracts/workspace-isolation.md:131-133 (two sentences) | 211 | RATIONALE | Rejected alternative (replace only the joiner) | Replacing only the joiner would leave `kan-15_fb13`, which still needs | med | Inside WI-06 range |
| WI-08 | skills/flow-contracts/workspace-isolation.md:79-81 (sentence) | 204 | RATIONALE | Why the boundary case is written down | flow change names are `<lowercased-jira-key>-<slug>` in practice, so this is | med | Inside WI-06 range |
| WI-09 | skills/flow-contracts/workspace-isolation.md:338-340 (sentence "Why a third verb …") | 238 | RATIONALE | Rejected alternative (read the removal's result) — must be moved, never deleted | Why a third verb rather than two, rather than reading | high | Inside WI-06 range |
| WI-10 | skills/flow-contracts/workspace-isolation.md:120-121 (sentence) | 115 | DUPLICATE | Same file L12-14 | What is namespaced is the logical resource *inside* each service, | med | Inside WI-06 range |
| WI-11 | skills/flow-contracts/workspace-isolation.md:44, 137, 149, 160 (fence info strings after the language) | 574 | RATIONALE | "measured" provenance of the verification environment; ccCanonicalIDCases matches fences by "```" prefix, not the tag | verified:run in this worktree on macOS (Darwin 25.5.0) with | med | Inside WI-06 range; keep the language token |
| WI-12 | skills/flow-contracts/workspace-isolation.md:205-218 | 1,144 | MECHANICS | Claim/verify-empty/give-back/refuse could be a project-declared `claim` verb beside create/remove/survivors, run by prepare-workspace.sh | **The claim is taken atomically, and an index is verified | low | Design change to the `## workspace isolation` command table; prepare-workspace.sh's header explains why it carries no cache client |
| GB-01 | skills/flow-contracts/git-boundaries.md:118-151 | 2,239 | LAZY-SPLIT | Two-commit chain. Needed: finish run 1 (always); implementation only on the `prUrl` path (≈never on merge-and-push); planning never | **Both commits are guarded, and an empty one is skipped | med | Repoint integrate.md:224 and finish-contract-run1.md:324 ("**Git boundaries**" meaning the chain); Load at integrate.md:164 and in VH-11's block |
| GB-02 | skills/flow-contracts/git-boundaries.md:34-76 | 3,482 | LAZY-SPLIT | Planning commits. Needed by planning and implementation; never by the finish session (reshape-branch.sh/commit-split.sh run the location guard themselves) | ## Planning commits **Planning artifacts are committed in their own | low-med | 13 citer lines (see citers list); high churn for 3.5 KB in the finish session only |
| GB-03 | skills/flow-contracts/git-boundaries.md:137-138 (" — every planning commit stages the same two trees … varies") | 110 | RATIONALE | Reason clause, and STALE: commit-split.sh:63-64 handles one planning dir | — every planning commit stages the same two trees in | med | Inside GB-01 range; git-boundaries-rationale.md is declared expected-zero — the moved text carries no bold+path, fine |
| GB-04 | skills/flow-contracts/git-boundaries.md:135-136 (sentence) | 104 | RATIONALE | Editor cross-reference to the writing-plans rule | That is the same rule the creating run's writing-plans stage applies to each task's `**Commit:**` field. | low | Inside GB-01 range |
| AR-01 | skills/flow-contracts/artifacts-registry.md:1-84 (whole file, via implement.md:232-233) | 7,010 | MISPLACED | Implementation session loads it every first run; no implementation step creates-then-removes a governed artifact by this table (throwaway copies: review-panel-optional-slots.md:59; render target: verify-and-handoff.md:121-125) | # Temporary artifacts registry Every artifact the pipeline creates, with | med | Change is in implement.md:232-233 (not this file); finish session keeps the load (archive.md:7) |
| AR-02 | skills/flow-contracts/artifacts-registry.md:36-39 (after the bold sentence) | 273 | DUPLICATE | git-boundaries.md:94-97 (both sessions load both files) | the implement phase's implementer writes them directly into `<project>/spectre/specs/<capability>.md` on | low-med | Keep the bold "not an artifact and carry no row" |
| AR-03 | skills/flow-contracts/artifacts-registry.md:41-45 (after the bold first sentence) | 375 | MISPLACED | Editor authoring rule + rationale pointer | Everything else that mentions a removal points here rather than | low-med | "Worktree cleanup is the procedure …" is already how run 2 is routed (archive.md) |
| AR-04 | skills/flow-contracts/artifacts-registry.md:47-50 | 309 | MISPLACED | Editor authoring rule (a run never edits the registry) | **An artifact no row accounts for is a defect in | med | none for a run; keep for editors in the rationale file |
| AR-05 | skills/flow-contracts/artifacts-registry.md:66-72 | 637 | DUPLICATE | finish-contract-run2.md:146-149, 165-175; workspace-isolation.md:333-337 | **This is the one row whose removal is verified by | med | none |
| AR-06 | skills/flow-contracts/artifacts-registry.md:74-76 | 244 | DUPLICATE (+STALE) | Table row L29 says the same; "design.md's open question `archive-branch-cleanup`" — the id exists only in artifacts-registry.md and its rationale (grep), no design.md | **Nothing removes the archive branch either, on `origin` or in | med-high | none |
| AR-07 | skills/flow-contracts/artifacts-registry.md:25 (" (`publish-proposal-removed`)") | 29 | RATIONALE | Decision-id provenance (spectre/changes/archive/kan-326-…/design.md:71) | (`publish-proposal-removed`), so the file is absent on every change it | high | Table cell edit; ccRegistryCoupling reads only column 1 |
| AR-08 | skills/flow-contracts/artifacts-registry.md:58-59 (" — which is why it names no database, no bucket and no service") | 64 | RATIONALE | Reason clause | — which is why it names no database, no bucket | low | none |
| SR-01 | skills/flow-contracts/session-records.md:14-16 | 183 | STALE | "implement phase reads this table on its `prUrl` commit path" — verify-and-handoff.md:119 loads it on every implementation run (ledger render, L121-135); also restates L5 | `/flow`'s implement phase reads this table on its `prUrl` commit | med | L5 carries the same stale clause and needs an editor's rewrite (not a cut) |
| WR-01 | skills/flow-contracts/worktree-resolution.md:26-29 (first sentence) | 303 | DUPLICATE | Same file L15-18 ("Any step in any command that needs 'the worktrees' … resolves the set first") | This binds every command that iterates a change's worktrees: the | low | Enumerated call sites are scope examples; keep the second sentence (delegation to finish-contract-run1.md) |

## Top 5

1. **VV-05** — move step 4, steps 7–11 and the `## Report` template of `visual-verify.md`
   (40.6 KB gross) into a verifier-read file; the parent keeps steps 3, 5, 6, 12, 13 and Blocking.
   Implementation session, every UI run (up to ~2× if the parent pastes the steps into the prompt today).
2. **WI-06** — narrow `verify-and-handoff.md:39-41`'s load of `workspace-isolation.md` to the one
   section each trigger needs: −20.4 KB per conditional load; implementation session, every run of a
   project that declares a `cache index` row.
3. **AR-01** — drop the `artifacts-registry.md` load (7.0 KB) from the implementation session
   (`implement.md:232-233`); the finish session keeps it.
4. **VH duplicates** — VH-09, VH-10, VH-12 (high/med-high, 580 B) plus VH-03, VH-05, VH-06, VH-15,
   VH-18, VH-20, VH-21 (med, 1.5 KB): ≈2.1 KB off every implementation session.
5. **GB-01 + VH-11** — the two-commit chain (2.2 KB) into its own file, loaded by `integrate.md:164`
   and by `verify-and-handoff.md` only on the `prUrl` path, with VH-11 (0.65 KB): ≈2.9 KB off every
   implementation session, 2.2 KB off the planning session.

Runner-up: VV-04 (4.1 KB off UI first runs), which needs a `dpSites` change first.

## Drift / bugs found

- **D1 — handshake second mismatch disagrees across three copies.** `implement.md:109-112` (canonical,
  "stated once here"): the parent asks through **AskUserQuestion**. `visual-verify.md:76-78` and
  `review-panel.md:539-541`: "ends the turn with `## Question`". The visual-verify copy also omits the
  single-model-harness (`zcode`) exception (`implement.md:115-121`), though it says "unchanged".
- **D2 — port collisions: contract vs. run.** `workspace-isolation.md:158-169, 176-180`: every port is
  checked free, and a bound port abandons the whole block for free-port discovery ("the block
  rediscovered"). `verify-and-handoff.md:304-311`: a held port is relayed, the stage "does not pick
  another port". `prepare-workspace.sh` does no `lsof` check and always exports Default+offset. The
  ordinary exit-0 run never loads `workspace-isolation.md`, so nothing executes L158-169.
  `docs/self-review/kan-302-…md:34` filed this as KAN-314.
- **D3 —** `implement.md:424` cites "(**10** in `skills/flow/verify-and-handoff.md`)". Step 10 lives in
  `visual-verify.md`.
- **D4 —** `implement.md:73-74` says the NO DELEGATION paragraph is in `verify-and-handoff.md`. It
  moved to `visual-verify.md` (verify-and-handoff.md carries no block, and dpSites no longer names it).
- **D5 —** `implement.md:496-497` lists `-role` as implementer/reviewer/panel-fix/verifier. The
  tooling analyst records `-role planner` (`visual-verify.md:141`), which the CLI accepts
  (`stats/cmd/flow/record.go:54`).
- **D6 — `session-records.md:5, 14-15`** say the implement phase reads the table only "on its `prUrl`
  commit path". `verify-and-handoff.md:119` loads it on every run. `session-records-rationale.md`
  ("renders on its `prUrl` commit path alone") is stale the same way.
- **D7 — `state-file.md:170-173`** says "`models` … Its live consumer is `/flow`, which dispatches on
  that value". `verify-and-handoff.md:340-341` says `/flow` "never records a value into the
  per-change state", and SKILL.md's Model resolution uses the settings store. VH also says
  `planningEffort` is "always `null`", while state-file.md:167 allows a legacy level.
- **D8 — `git-boundaries.md:92`** says "through **Planning commits** below". The section is above
  (L34). The rest of the file's positional words are correct.
- **D9 — "two trees" / "two planning paths" are relics.** `git-boundaries.md:137` and
  `finish-contract-run1.md:318` say this, but `commit-split.sh:63-64` and the chain at
  git-boundaries.md:124-126 handle one planning dir (the spec-root's `changes/`).
- **D10 — registry gap.** `implement.md:563-581` creates `<worktree>-wave-group-<g>` sibling
  worktrees and removes them after the pick. `artifacts-registry.md` has no row for them, and its own
  L47 calls that a registry defect. A run that dies mid-wave leaves a worktree that
  `check-cleanup-complete.sh` never looks for. Adding a row also needs a marker decision
  (ccRegistryCoupling).
- **D11 — "Loaded by" lines are incomplete.** `git-boundaries.md:6` (and `flow-contracts/SKILL.md:34`)
  omit the creating run (brainstorm.md:102, :252) and `/flow-plan` (flow-plan/SKILL.md:217).
  `worktree-resolution.md:5` omits the creating run (brainstorm.md:165, :208).
- **D12 — guard header vs. code.** The `check-dispatch-paragraphs.sh` header table (L137-166) says
  visual-verify.md needs min **1** TOOLS/HANDSHAKE/DELEGATION block and has no READ-ONLY REVIEW row.
  `dispatchparagraphs.go` dpSites requires **2** and carries `readonly` for implement.md. The header
  claims to be the contract.
- **D13 — `verify-and-handoff.md:293`** says "A refused start above already ends the run". That
  bullet is below (L304).
- **D14 — `verify-and-handoff.md:265`** says "`<project>/CLAUDE.md` states that prohibition". Only
  this repo's CLAUDE.md/AGENTS.md carry it; setup.sh renders nothing like it into other projects.
- **D15 — `verify-and-handoff.md:354-357`** runs `flow record cost-status -change <name>` "one call
  per affected worktree" with no `-C`, which gives N identical calls.
- **D16 — `implement.md:230`** heads §2 "(first run only)", but implement.md:9, :25-29 and :329-330
  run it on every run ("isolate (resume)"). Both VH-01's premise and verify-and-handoff.md:25-26
  ("the same set **2. Isolate the workspace** resolved") depend on it running.
- **D17 — `verify-and-handoff.md:81`** allows "its last 40 lines". implement.md:1117-1120 has the
  parent pipe its own flow.verify lint/test runs through `| tail -20`.
- **D18 — `brainstorm-planner.md:237-238`** cites "**Verify and hand off**
  (`skills/flow/verify-and-handoff.md`)". No such heading exists (the H1 is "Verify, stage, and hand
  off"). It passes check-references only because the bold spans two lines.
- **D19 — hidden finish-session load.** `integrate.md:170-173` cites `verify-and-handoff.md` for the
  narrative append, so a finish session following citations may read 27 KB for a 610 B rule
  (VH L323-329).
- **D20 —** `verify-and-handoff.md:39-41` loads the whole 25 KB contract on a `prepare-workspace.sh`
  exit 2, though nothing in it governs exit 2. L34-37 already says to stop.

## Not slimmable

- **visual-verify.md:31-50 and 102-121** (2 × 1,161 B, TOOLS / NO DELEGATION / MODEL HANDSHAKE):
  pinned by `stats/internal/guard/dispatchparagraphs.go` dpSites (min 2 in visual-verify.md). The
  handshake and tools rules also act before the first tool call, so they must sit in the prompt, not
  in a file the verifier reads.
- **visual-verify.md:170-209** (step 3 pre-flight): parent-only evidence rules. Only VV-18
  (MECHANICS) could shrink it.
- **visual-verify.md:756-777** (Blocking + end mark): the parent's closed list of blocking conditions
  and the one `stopped` outcome.
- **visual-verify.md steps 7–11 body** (≈36 KB net of VV-06…16): normative verifier rules. The
  2026-09-22 audit already moved every incident record (SKILL-rationale.md:139-262). Only relocation
  (VV-05) is possible, not deletion.
- **verify-and-handoff.md:19-37, 65-89, 110-117**: the `prepare-workspace.sh` exit contract, the
  inline-verify order and report template, and the dispatch-row recording.
- **verify-and-handoff.md:373-406**: the handoff template. Operator-facing field names;
  handoff-blocks.md mirrors it.
- **verify-and-handoff.md:249-311** (net of VH-12/13/14/26): the start / protect / freshness /
  refused-start rules. "Never the flow dev stack" is the only statement of the prohibition in any
  project whose CLAUDE.md lacks it (D14).
- **workspace-isolation.md:32-56, 137-139**: the derivation steps and the two fenced
  `id=`/`id_underscored=` blocks with their `# <name> -> <id>` lines are executed by
  `check_cleanup_complete_test.go` ccCanonicalIDCases. They must stay in this file (so no physical
  split of it).
- **workspace-isolation.md:271-300**: the dropped-row refusal asymmetry. Normative; it is what exit 1
  loads the file for.
- **artifacts-registry.md:13-34**: the table. ccRegistryCoupling
  (check_cleanup_complete_test.go:856-916) parses rows under `## Temporary artifacts registry`.
- **git-boundaries.md**: the table, the planning-commit bash (L50-56) and the chain block
  (L123-132). Canonical; commit-split.sh, reshape-branch.sh and check-planning-commit-location.sh
  cite them. Only relocation (GB-01/02) is possible.
- **session-records.md:17-28, worktree-resolution.md:15-24**: short canonical rules, each loaded by
  both sessions.

## Answers to the dispatch questions

### visual-verify.md — who needs what (61,248 B)

| part | lines | bytes | needed by |
|---|---|---:|---|
| preamble | 1-8 | 436 | parent (orientation) |
| The verifier dispatch | 9-82 | 5,000 | parent, every UI run (prompt paragraphs are pasted verbatim) |
| miss classification | 83-93 | 679 | parent, UI fix runs only |
| tooling analyst dispatch, prompt, recording, re-run, abort | 94-156 | 4,093 | parent, UI fix run with ≥1 miss only |
| steps intro + verifier prompt contract | 157-169 | 1,117 | parent |
| step 3 pre-flight | 170-209 | 3,365 | parent |
| step 4 | 210 | 88 | verifier only |
| steps 5–6 | 211-227 | 1,542 | parent |
| steps 7–11 | 228-694 | 38,405 | verifier only — except two parent-duty clauses: L425-428 "the parent reconciles it against `design.md`'s own frame list…" (160 B) and L565-568 "the parent opens the frame beside the matrix…" (132 B), which Blocking (L763-768) already names |
| steps 12–13 | 695-727 | 2,464 | parent |
| `## Report` template (the only fenced "prompt" block) | 728-755 | 2,408 | verifier writes it; the parent reads the actual report, not the template |
| Blocking + end mark | 756-778 | 1,652 | parent |

The parent needs ≈15.6 KB on every UI run, +0.7 KB on fix runs, +4.1 KB only with a miss. About
40.9 KB is verifier-only.

**Can the verifier prompt live in its own file the verifier reads?**
- **No, not the fixed paragraphs.** The baseline pointer, relay contract, TOOLS, NO DELEGATION and
  MODEL HANDSHAKE (~2.3 KB) must be in the prompt itself: TOOLS and the handshake act before any
  tool call, and dpSites pins ≥2 copies of each in visual-verify.md.
- **Yes, the verifier's work.** Step 4, steps 7–11 and the `## Report` template (VV-05) can move
  verbatim to e.g. `skills/flow/visual-verify-verifier.md`. The prompt contract at L162-168 then
  names that file's absolute path instead of "steps 4 and 7–11 below", as the analyst prompt
  already carries "the path of this file".
- **What changes with it (pointer edits, not rewording):**
  - the analyst's task text (L128 "steps 8–10 of `skills/flow/visual-verify.md`", L131);
  - verify-and-handoff.md:379 and :436;
  - implement.md:424 (already stale, D3);
  - implement.md:1121-1124 (the Read-discipline list).
- **Guard impact:** no `**Heading** (path)` citer exists for the steps, and the new file keeps ≥1
  verified reference (L675 `**visual verification** (…project-configuration.md)`), so
  check-references coverage holds.
- **Risk:** today "run steps 4 and 7–11 below as written" is ambiguous. If the parent pastes the
  steps into the Agent call, it carries them twice (the file read plus the tool input). After the
  split it carries neither.

### verify-and-handoff.md — conditional blocks

- **Fix-run-only:**
  - the pre-edit Jira line explanation (L438-440; also a DUPLICATE, VH-20);
  - the Tooling-analysis line explanation (L433-436, only with a miss);
  - the PR-branch split (L206-223, only a fix run on a change whose PR is open).
- **First-run-only:** no whole block. Only the "IN_PROGRESS from STARTED" clause (L331) and the
  "staged and uncommitted" spelling for a run resuming before any task commit (L408-410).
- **Route-specific:** PR route only — the PR-branch split (L206-223), which is `prUrl`-gated and so
  never reached on merge-and-push.
- **Failure-specific:** "Inline verify — a failing command" + one re-run (L91-108); the blocked
  outcome in Recording (L114-117); the refused-start relay (L304-311).
- **Stack/project-specific:**
  - workspace isolation: load, cache-index claim, hand-apply (L39-51);
  - visual steps 1–2 (L141-179) — every project still prints "Visual: not configured";
  - "no runnable application" (L247-248, L282-283);
  - the protected flow dev stack + "Not started" template (L263-278, agents repo only);
  - freshness check (L285-295, needs a `fingerprint` row and URL lines).
- **Every run:** Verify core (L6-89, L110-139), stage-diff confirm (L181-205), run instructions
  (L229-302), write IN_PROGRESS + handoff (L317-437).

### git-boundaries.md — per-session need

| section | lines | bytes | planning | implementation | finish |
|---|---|---:|---|---|---|
| preamble | 1-12 | 477 | ✓ | ✓ | ✓ |
| table | 13-27 | 2,081 | row L17 | rows L18-20 | rows L21-24 |
| no command writes the main checkout | 28-33 | 440 | ✓ | ✓ | ✓ |
| Planning commits | 34-76 | 3,482 | ✓ (plan gate) | ✓ (link, merge order, reviewer, document-fix, write-in-progress) | ✗ |
| Branch backup ¶1 | 77-86 | 600 | ✓ (push -u at kickoff) | ✓ (push after each commit) | ✓ (force-with-lease after reshape) |
| planning-path clearing | 87-93 | 507 | ✗ | ✓ | ✓ (finish-contract-run1.md:318-321 cites it) |
| capability spec / link.md / .gitignore | 94-110 | 1,226 | spec only | ✓ | context (commit-split.sh enforces) |
| pathspec-scoped default | 111-117 | 514 | ✗ | ✓ (fix commits, visual step 12) | ✓ |
| two-commit chain, skip/fail, symlink | 118-151 | 2,239 | ✗ | `prUrl` path only | ✓ |

The clean splits are GB-01 (chain → finish and `prUrl`) and GB-02 (Planning commits → planning and
implementation). Table rows could split too, but saving ~1 KB is not worth the churn.

### Restatements of pipeline.md / SKILL.md / implement.md / review-panel.md

- **Cut candidates:**
  - VH-03 → implement.md:56-61, SKILL.md:109-111
  - VH-09 → implement.md:1084-1086
  - VH-10 → SKILL.md:228, pipeline.md:257-262
  - VH-12 → pipeline.md:253-256
  - VH-15 → implement.md:242-258
  - VH-18 → pipeline.md:253
  - VH-21 → SKILL.md:233-237, implement.md:21-24
  - VV-03 → implement.md:94-113 (drift D1)
  - VV-02 → same file L48-49
- **Kept on purpose:**
  - verify-and-handoff.md:58-60 → review-panel.md:929-932 — the reason for "edits no source".
  - verify-and-handoff.md:187-190 → review-panel.md:775 — defines the expected log shape.
  - verify-and-handoff.md:411-421 (Panel: field definitions) → review-panel.md docs-only / late-fix /
    Bundled dispatch.
  - visual-verify.md:14-17 → SKILL.md:109-115 — ~95 B of the VERIFY_MODEL rule at its own dispatch
    site.
  - visual-verify.md:80-81 → implement.md:131-134 — adds the "no `## Report`" case and "blocks this
    handoff".
  - visual-verify.md:159-160 → implement.md:56-61 (partial).
- **Duplicates owned by the other side:**
  - pipeline.md:255-256 restates git-boundaries.md:28-32.
  - pipeline.md:257-264 restates git-boundaries.md:87-99.
  - review-panel.md:539-541 restates the handshake (D1).
  - implement.md:263-266 restates worktree-resolution.md.

### workspace-isolation.md — exactly when it loads

- **Directive at verify-and-handoff.md:39-41 — `prepare-workspace.sh <worktree>` exited non-zero:**
  - **exit 1:** check-workspace-isolation.sh found a malformed (dropped) `## workspace isolation`
    row; its stdout is relayed.
  - **exit 2:** usage error; the argument is not a directory; check-workspace-isolation.sh is
    missing or not executable; that guard itself exits 2; the worktree is not on a `spectre/<name>`
    branch; or a `<value:VAR>` reference names a non-database/bucket/port row.
  - Every non-zero case stops before `## lint`/`## test`.
- **Directive at verify-and-handoff.md:39-47 — exit 0 whose stderr names a `cache index` row**
  (prepare-workspace.sh Pass 1, "`VAR` (cache index) is claimed by probing…"). The project's table
  declares a `cache index` resource; the parent then claims one per "The cache index".
- **verify-and-handoff.md:49-51 — `prepare-workspace.sh` cannot be located:** hand-apply from it
  and project-configuration.md.
- **Cited at point of use, not directed:**
  - visual-verify.md:185 and :213 cite **What the id derives** for the `lsof` probe and the
    worktree-resolved URL. Both facts are already inline or in the printed `KEY=value` lines.
- **Never loaded:**
  - the ordinary exit-0 run without a cache-index row (always, in this repo — its table has no such
    row);
  - implement.md, which uses `flow workspace-id <name>`;
  - finish run 2, which also uses `flow workspace-id` and only cites "Creation and cleanup".
- **Sections each trigger actually needs:**
  - exit 1 → The empty id L285-291 (report the row, the cell and the declined shared value; stop);
  - exit 2 → none;
  - cache index → L182-247;
  - hand-apply → The workspace id + What the id derives (or just `flow workspace-id`).

### Citers of each LAZY-SPLIT candidate (grep of the repo)

- **VH-04 (failing command):**
  - project-configuration.md:41 and known-bugs.md:47 (both `**Verify**` → still resolve, so
    repointing is manual);
  - in-file L76-77;
  - SKILL-rationale.md:173-176 (attribution).
- **VH-11 (PR-branch split):** no heading citer. Prose citers:
  - finish-contract-run1.md:325-326;
  - session-records.md:5, 14;
  - state-file.md:185-187;
  - stats/cmd/flow/record.go:2412 (comment).
- **VH-19:** none external. visual-verify.md:153-155 points at the handoff line.
- **VV-04 (heading "A missed defect — the tooling analysis"):**
  - implement.md:54, SKILL.md:121, verify-and-handoff.md:433-434;
  - in-file L13-14, L167;
  - stats/cmd/flow/record.go:31 (comment), SKILL-rationale.md:227.
  - Guard: dpSites plus check_dispatch_paragraphs_test.go (cases 50/50a, `dpCleanVisualVerify`).
- **VV-05 (steps):** no heading citers. Step-number prose:
  - implement.md:424, :471 (step 6 stays), :1121-1124;
  - verify-and-handoff.md:267 (step 13 stays), :293 (step 6 stays), :379, :436;
  - visual-verify.md:128, :131, :147-151, :162-168;
  - scripts/resolve-visual-screenshots.sh:7-8 (comment, step 9);
  - SKILL-rationale.md:178-262 (attribution headings);
  - .flow/project.md:59 (start/stop steps, which stay).
- **WI-06 (section-scoped read):** no citer changes. If split physically instead, ~20 citers break:
  - "The workspace id": finish-contract-run2.md:138; project-configuration.md:165, 236.
  - "What the id derives": artifacts-registry.md:64; project-configuration.md:167, 239;
    visual-verify.md:185, 213.
  - "The cache index": project-configuration.md:177.
  - "The empty id": project-configuration.md:438, 453.
  - "Creation and cleanup": finish-contract-run2.md:149, 175; artifacts-registry.md:68;
    project-configuration.md:252, 285, 289, 337, 348.
  - Script comments in prepare-workspace.sh and check-cleanup-complete.sh, plus the test that reads
    the file.
- **GB-01 (chain):**
  - "**Git boundaries**" citers meaning the chain: integrate.md:224, finish-contract-run1.md:324;
  - loads: integrate.md:164, verify-and-handoff.md:206;
  - citers meaning the rest, which stay: pipeline.md:259, finish-contract-run1.md:320.
- **GB-02 (Planning commits):**
  - implement.md:277, 326, 446; review-panel.md:394; brainstorm.md:252;
  - verify-and-handoff.md:328; finish-contract-run1.md:296; pipeline.md:261;
    flow-plan/SKILL.md:217;
  - in-file L17-18, L92;
  - comments in check-planning-commit-location.sh:16, check-task-commit-planning-paths.sh:21,
    reshape-branch.sh:15.

### Guard constraints on these moves

- **check-references** (stats/internal/guard/references.go):
  - Every `**Heading** (path)` citer of a moved heading must be repointed; the lists are above.
  - A new file needs ≥1 verified reference, or a `crExpectedZero` declaration (L69-121).
  - `git-boundaries-rationale.md`, `session-records-rationale.md` and
    `worktree-resolution-rationale.md` are declared expected-zero, so text moved into them must not
    add a bold-adjacent path citation, or the declaration becomes false.
  - `skills/flow/SKILL-rationale.md`, `workspace-isolation-rationale.md` and
    `artifacts-registry-rationale.md` are not declared.
- **check-dispatch-paragraphs:** dpSites pins visual-verify.md at min 2 for TOOLS, MODEL HANDSHAKE
  and NO DELEGATION. No site names verify-and-handoff.md, so no required paragraph lives there any
  more. VV-04 needs dpSites (new file, min 1; visual-verify.md → 1) and test changes. VV-05 does
  not touch them.
- **check_cleanup_complete_test.go:**
  - workspace-isolation.md's fenced id blocks and published ids stay put (WI-11 edits only info
    strings, which the test ignores);
  - artifacts-registry.md's table stays under its heading.
- **check-markdown-integrity.py:**
  - no cut may leave a trailing-colon paragraph (VV-01 takes the "verbatim:" lead-in with it);
  - no cut may leave a torn paragraph or a blockquote without its `>`.
- **check-installed-citations:** a new installed file needs non-zero coverage or a declaration.
- **check-normative-inventory:** none of the seven files carries MUST/SHALL (grep), so there is no
  inventory impact.
- **check-stage-mark-calls:** visual-verify.md is not a corpus candidate, and a verifier file
  carries no mark, so nothing changes.
