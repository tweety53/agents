# Slimming analysis — `skills/flow/implement.md` (IMPLEMENTATION session)

> **Audit snapshot at `700e184` (2026-09-29).** A read-only model pass that followed [`rubric.md`](rubric.md); every row is a candidate to verify, not a verified fact. Line numbers refer to `700e184`. Main has since retired the bugbot and security slots (`ebdfdde1`..`4dafab4a`), so re-locate each row by its quoted first words, and treat bugbot/security rows as obsolete. Tracked in KAN-851 → KAN-857 (and the Drift items in KAN-853 / KAN-854). Plan: [`README.md`](README.md).

Read in full (1170 lines, 81 740 bytes). Read-only: nothing in the repository was modified
(`git status` clean). Byte counts are exact (`sed -n 'A,Bp' | wc -c` for line ranges; exact
substring byte length for sub-paragraph fragments, measured with a script against the file). Sibling
rationale file for `skills/flow/*.md` is `skills/flow/SKILL-rationale.md` (pass A precedent:
"## Moved by the 2026-09-22 prompt audit (pass A) — review-panel.md and implement.md").

## Totals

Conservative: only med/high-confidence candidates counted; every byte counted once (a RATIONALE
fragment inside a LAZY-SPLIT block is counted under RATIONALE and subtracted from the block).
STALE items are citation/wording fixes, not byte savings (0). MECHANICS is passage size, flagged
separately — it needs code + parity tests and leaves a one-line call behind (est. net ≈ 3.3 KB).

| file | bytes | RATIONALE | DUPLICATE | STALE | LAZY-SPLIT | MECHANICS | MISPLACED |
|---|---|---|---|---|---|---|---|
| skills/flow/implement.md | 81 740 | 4 734 | 626 | 0 (8 fix items) | 18 894 | 5 219 (passage; est. net ~3.3 K) | 44 |

- Behaviour-neutral total (RATIONALE + DUPLICATE + LAZY-SPLIT + MISPLACED) = **24 298 B ≈ 29.7 %** of the file.
- A typical run (Claude Code, first implementation run, `inline`, single repo, no gated `fix` pass)
  would stop loading all of it (minus ~0.8 KB of new "Load X only when Y" directives).
- LAZY-SPLIT breakdown: L1 fix-run 7 201 · L2 sdd-only 4 601 · L4 cross-repo 3 696 · L3 gated-fix 2 834 · L5 zcode 562.

## Candidates

`implement.md` = `skills/flow/implement.md`. "Move→R" = move verbatim to `skills/flow/SKILL-rationale.md`
under a `### implement.md — <section>` sub-heading (pass A shape). "cut-only" = the text already sits
verbatim in SKILL-rationale.md.

| id | file:startline-endline | bytes | lever | condition / canonical location / evidence | first ~10 words verbatim | conf | behaviour-risk note |
|---|---|---|---|---|---|---|---|
| L1 | implement.md:336-451 (whole `## 3.`) | 7 201 (7 500 − R10/R11/R12) | LAZY-SPLIT | Fix run only (state `IN_PROGRESS` + argument, or plain-message trigger). Line 341: "**Fix runs only** — a first run … marks nothing here". False on every first implementation run (one per change); a same-session plain-message fix already has it loaded either way. Est. false in ≥60 % of implementation sessions | `## 3. Documenting a fix, before implementing it **Parent work, run` | high | Carries `flow stage begin … flow.document-fix` (l.345): the new file's basename must be added to `smcCandidates` in `stats/internal/guard/stagemarkcalls.go:72-76` or the call silently leaves check-stage-mark-calls' corpus. Carries two operator prompts (moved verbatim — fine). Citers to repoint: SKILL.md:39 (Stage keys row), SKILL.md:144-145, SKILL.md:171-172, verify-and-handoff.md:438-439, operator-prompts.md:81-82, brainstorm.md:261-262, implement.md:25-29 & 1121-1124 (phase-file list). None is a check-references-checked token (`**3. …**` contains "."). |
| L2 | implement.md:541-548, 550-582, 584-592, 599-601, 612-619, 621-624 | 4 601 (4 771 − D39) | LAZY-SPLIT | `execution` = `sdd` only (class `big`: tasks≥22 or files≥60 or repos>1∧tasks≥11 or migration∧tasks≥11, per `scripts/plan-class.sh`). Inline never dispatches an implementer, never forms groups (`groups` null), never runs waves (implement.md:150-151 "Waves are not parallel inline"). Est. false in 85-95 % of runs | `**Dispatch one implementer per group, not per bundle.** The unit is` / `**Waves — concurrent dispatch of ready groups.** A group is ready` / `**Gather one context bundle per group, immediately before that group's implementer` / `where <changeRoot> is …` / `A non-zero exit — including the guard being absent — is` / `The sixth argument scopes the group's ## tasks.md section to the` | high (med for 599-601) | Keep in implement.md: `<shape>` (594-597) and `<canonical-worktree>` (603-610) — review-panel.md:114-115 & 1139 use them on every run. 599-601 (`<principles-path>`) is also defined verbatim at review-panel.md:663-665 ([PRINCIPLES_PATH]); review-panel.md:114 uses the `<principles-path>` token 550 lines before its own definition — med. No pinned dispatch labels, no `flow stage begin`/`flow record dispatch` lines → check-dispatch-paragraphs and check-stage-mark-calls unaffected. Citers: SKILL.md:218-220 ("Waves paragraph of **4. Execute (SDD + TDD)**"), flow-fast/SKILL.md:86-90 (sdd: "…the context bundle gathered…, waves"), scripts/plan-dispatch-groups.py:78 ("implement.md's **Waves**"), implement.md:1113-1116 Read-discipline bullet ('"Never read the bundle back" (**4**, above)' → repoint), implement.md:626-628 ("the gather above"). |
| L4 | implement.md:268-285, 294(sentence "Every other entry…")-313, 315-327 | 3 696 (4 378 − R7/R8/R9) | LAZY-SPLIT | Needed only when the change links a peer, an `## apps` entry names a repository holding no worktree for this change, or `## visual verification` names a `regression checkout`. Text itself: "A change with no linked peers runs nothing here" (284-285), "A change with one repository runs nothing here" (327). For this repo (`.flow/project.md` `## apps` = two rows, same repo; no regression checkout) false unless a spectre/gymie peer is linked. Est. false 80-95 % | `**After creating each additional worktree, run spectre link --root <abs-worktree>/spectre <canonical-peer>:<name>` / `Every other entry gets the kickoff recipe in its own repository` / `**Then make the merge-order record cover the whole set.** The canonical` | med | Keep 287-294 up to "…records that worktree as its root and creates nothing." (every run resolves `## apps`). Directive must name all three triggers. Citers: git-boundaries.md:44-45 (planning-commit table rows naming `flow.isolate-workspace (implement.md)`), state-file.md:289-290, finish-contract-run1.md:285-286, project-configuration.md:28. None checked by check-references. |
| L3 | implement.md:950-989 ("**The parent applies the fix itself**…" to "`git diff <task-sha>^..<new-task-sha>`.") | 2 834 (3 214 − R25/R26a/R26b) | LAZY-SPLIT | Only when a gated per-task reviewer bundle closes with any pass `fix` (Critical/Important). Est. false 65-85 % of runs | `**The parent applies the fix itself**, never resuming the group's implementer:` | med | Split point is a sentence boundary mid-line 950 (after "of its bundle-mates."). Directive at the mixed-verdict sentence (947-950). 956-973 inside it restates review-panel.md:775-798 (**Panel re-runs**) — see ROUTE. No pinned labels, no stage-mark/dispatch command lines. No heading citers. |
| L5 | implement.md:115-121 | 562 (671 − R3) | LAZY-SPLIT | Harness `zcode` only; never true on Claude Code. Natural home: `skills/flow-contracts/model-policy.md` **Harness mapping** (already the zcode section; cited from l.116-117) | `**On a single-model harness the recorded mapping satisfies the handshake.** Where` | med | Conflicts with model-policy.md:113 ("The handshake compares against `glm-5.3-flash`") — reconcile when moving (drift #6). No citers besides l.116's own pointer. |
| R1 | implement.md:67-68 | 123 | RATIONALE | Why dispatches are one-shot; rule stated before and after | `A subagent's prompt cache lives five minutes, so a child woken` | high | Move→R. |
| R2 | implement.md:78-80 | 135 | RATIONALE | Why agent definitions carry `effort:` and no `model:` (design of `agents/flow-*.md`) | `, since the Agent tool's dispatch-time model parameter overrides a definition's` | high | Move→R; leaves "each carrying `effort:` and no `model:`) carries". |
| R3 | implement.md:119-120 | 109 | RATIONALE | Why the zcode rule holds (inside L5) | `The mapping already fixes what the handshake exists to establish, and` | high | Move→R. |
| R4 | implement.md:133-134 | 87 | RATIONALE | Why an aborted dispatch is not retried; rule in the bold lead | `, and the operator should see the death rather than have` | med | Move→R. |
| R5 | implement.md:197-199 | 162 | RATIONALE | Why the plan is refreshed (pass A moved only the kan-579 parenthetical) | `The plan and design were written against a snapshot of the` | high | Move→R. |
| R6 | implement.md:246-249 | 234 | RATIONALE | Why persisting early matters | `— that mismatch is exactly what **Reading the state** (skills/flow/SKILL.md) uses` | med-high | Move→R; removes one checked citation (harmless — many remain). The preceding "must not leave the state record looking like a creating run…" stays. |
| R7 | implement.md:270-274 | 423 | RATIONALE | Why cwd = primary checkout and why `--root` (inside L4) | `The working directory is what resolves the peers file's relative entries` | med | Move→R. `check-unfinished-work.sh` named here is named elsewhere in skills/flow (guard-symlinks rule 2 unaffected). |
| R8 | implement.md:283-284 | 139 | RATIONALE | Why a refused link is a hard failure (inside L4) | `— a change whose cross-repo link cannot be established lands at` | high | Move→R. |
| R9 | implement.md:308-309 | 120 | RATIONALE | Why a failed app worktree add is a hard failure (inside L4) | `: a declared app left unresolved is the commit destination a` | high | Move→R. |
| R10 | implement.md:352-353 | 104 | RATIONALE | Why the re-plan budget exists (inside L1) | `: an append past this budget is how a change outgrows` | high | Move→R. |
| R11 | implement.md:425-426 | 105 | RATIONALE | Why a disputed visual is measured (inside L1) | `The glance that passed the control is what the operator is` | high | Move→R. |
| R12 | implement.md:439-440 | 90 | RATIONALE | Why plan growth is recorded (pass A moved only "(KAN-415)") (inside L1) | `— gate-time re-planning visible as a trend in the app rather` | high | Move→R; leaves "plan-growth series:". |
| R13 | implement.md:465-466 | 108 | RATIONALE | Example behind "however file-disjoint two tasks look on paper" (pass A moved its KAN-30 incident) | `: UI fixes routinely touch shared files — icon sets, shared` | med-high | Move→R. |
| R14 | implement.md:490-493 | 269 | RATIONALE | Harvester mechanism behind "begin before the launch" (rule already stated at 489-490) | `— Claude Code writes it into the parent transcript's own launch` | med | Move→R; contains "must precede" but only restates the rule it follows. |
| R15 | implement.md:526-528 | 163 | RATIONALE | kan-527 incident narrative | `: kan-527's chunk 1 worked under one and its git incident` | high | Move→R; leaves "…is not a carry." |
| R16 | implement.md:529-530 | 100 | RATIONALE | Provenance "(KAN-643's rule)" | `— the same reason the content markers assert by grep, never` | high | Move→R. |
| R18 | implement.md:793-795 | 187 | RATIONALE | Why the parent (not an implementer) runs the suite | `A subagent's prompt cache lives five minutes, the parent's an hour:` | high | Move→R. Same fact as R1. |
| R19 | implement.md:868 | 76 | RATIONALE | Already verbatim in SKILL-rationale.md "### implement.md — The review gate" | `Forty changed lines is the boundary below which a diff still` | low | cut-only; leaves a lowercase sentence start ("apply it as stated…") — cosmetic; excluded from totals. |
| R20 | implement.md:869-871 | 176 | RATIONALE | Why the undeclared-path arm exists; arm itself stated at 867-868 | `The undeclared-path arm is the gate's risk half: a commit reaching` | med-high | Move→R. |
| R21 | implement.md:893-894 | 151 | RATIONALE | Why the disclosure cannot disarm the gate | `— the pre-correction declaration lives in the refusal, not in any` | med | Move→R; "reads the paths the refusal named" (the instruction) stays. |
| R22 | implement.md:903-905 | 119 | RATIONALE | Why the dated Correction paragraph | `The archived plan then reads as what actually shipped, and the` | high | Move→R. |
| R23 | implement.md:934-935 | 75 | RATIONALE | Already verbatim in SKILL-rationale.md "### implement.md — The gated per-task reviewer" | `(the review-dispatch count tracks the change's size, never its task count).` | med-high | cut-only. |
| R24 | implement.md:942 | 60 | RATIONALE | Why per-pass verdicts | `, so per-task review yield stays measurable against the gate. **A` | med | Move→R. |
| R25 | implement.md:964-966 | 163 | RATIONALE | Why `<task-sha>^` is explicit (inside L3; same reason at review-panel.md:787-790) | `— the explicit base is load-bearing: a bare git rebase --autosquash` | med | Move→R; guards against "simplifying" the command — keep if in doubt. |
| R26a | implement.md:968-969 | 127 | RATIONALE | Describes guard-autosquash (its header is canonical) (inside L3) | `— a target that merely resolves hands the autosquash a spurious` | med | Move→R. |
| R26b | implement.md:971-972 | 90 | RATIONALE | Same (inside L3) | `, so a rewritten branch cannot pass while tasks.md still names` | med | Move→R. |
| R27 | implement.md:1025-1027 | 145 | RATIONALE | gymie KAN-635 incident inside the READ-ONLY REVIEW prompt (pass A moved analogous incidents out of FLOW — COMMIT-PER-TASK) | `: gymie KAN-635's > reviewer ran git checkout <sha> -- .` | med | Move→R. Subagent-facing: a vivid example may affect reviewer compliance. check-dispatch-paragraphs' three READ-ONLY REVIEW phrases stay. |
| R28 | implement.md:1047-1049 | 148 | RATIONALE | Why the reviewer-dispatch planning commit (incident) | `(task commits never carry plan paths, so uncommitted edits are the` | med-high | Move→R; leaves "…and ticks, then runs". |
| R29 | implement.md:1054-1055 | 159 | RATIONALE | Restates the lead-in's "asserted by git, never by the reviewer's prose" as a why | `A reviewer that closed clean over a tree it changed has` | med-high | Move→R. |
| R30 | implement.md:1069-1072 | 278 | RATIONALE | Why markers beside the plan-tree guard (KAN-579) | `: that guard asserts git's view of one directory, while markers` | med-high | Move→R; "The markers sit beside the plan-tree guard, never in its place." stays. |
| R31 | implement.md:1095-1097 | 249 | RATIONALE | Rejected alternative (10-minute cap) + cache reasoning; the loop code fixes 48×5 s | `The loop is bounded at 240 s rather than the Bash` | med-high | Move→R. |
| R32 | implement.md:1103-1104 | 136 | RATIONALE | Why read discipline | `The parent does a large amount of reading across one long-lived` | high | Move→R. |
| R37 | implement.md:400-401 | 79 | RATIONALE | Why the evidence rule bites at append (inside L1) | `The append is where a fix round is most tempted to` | low | Excluded from totals. |
| D17 | implement.md:637-639 | 190 | DUPLICATE | Canonical implement.md:510-511 ("A record write never blocks. An unreachable store journals the intent…") + a rationale tail | `The write journals on store failure like every record write and` | med-high | Cut; paragraph ends at "…verbatim." |
| D39 | implement.md:546-548 | 170 | DUPLICATE | Canonical implement.md:672-676 ("A `Build: red` task is dispatched with its `Squash-with:` partner, in one bundle… **one** commit for the pair") | `— a red task and its partner make one commit between` | med | Cut (inside L2); leaves "…own `Task-Id:` trailer." |
| D40 | implement.md:136-138 | 266 | DUPLICATE | Canonical review-panel.md:144-166 (**The roster** table: every slot `flow-low` + its prompt file; "Every slot in this table, Bugbot and Security included…"); also SKILL.md:117-119 (every session) and model-policy.md:76 | `**Bugbot and Security are prompt-driven roles like every other panel slot,` | med-high | Cut. Panel dispatch happens under review-panel.md, which states it. "never a fixed bugbot or security-review type" is history. |
| ROUTE | implement.md:956-973 | 1 392 (inside L3; not added) | DUPLICATE | Canonical review-panel.md:775-798 (**Panel re-runs**: on-top commit, pathspec staging, `--fixup`, `<task-sha>^` upstream, guard-autosquash targets/after). implement.md:955-956 already cites it; the full-suite fix at 796-797 relies on the citation alone | `**A branch the remote already holds takes the fix as one` | med-low | Copies have drifted (drift #10); cutting needs a reconcile first and a reword of "A conflict there" (975). Prefer L3. Timing: review-panel.md is not yet loaded during `flow.sdd-tdd`, so the parent would `grep -n`+`sed -n` that section. |
| FLOWEFF | implement.md:76-88 (minus R2) | 912 (not added) | DUPLICATE | review-panel.md:160-166 & 176-181, visual-verify.md:14, SKILL.md:109-115, implement.md:504-506 | `**The flow-<effort> family (agents/flow-low.md, agents/flow-medium.md, agents/flow-high.md, agents/flow-xhigh.md — four definitions, one` | low | Only place implying the gated reviewer's subagent type; that type is itself unspecified on small/regular (drift #13). Keep until fixed. |
| R34 | implement.md:94 | 44 | MISPLACED | Editor meta inside the bold lead "**The handshake — stated once here, cited everywhere else.**" | `— stated once here, cited everywhere else.** Every dispatched role in` | med | Cut; citers use "**The handshake**". |
| M38 | implement.md:1021-1023 | 116 | MISPLACED | READ-ONLY REVIEW goes only to the gated per-task reviewer; the panel-slot exception never applies to its recipient (panel slots carry their own Read-Only Review sections) | `— the panel's > mutating slots are the one declared exception,` | low | It is an exception clause (rubric: never cut exceptions) — excluded from totals. |
| MX1 | implement.md:862-868 | 574 | MECHANICS | Review-gate computation (numstat sum >40, name-only ⊄ Files ∪ Allowed-collateral) → e.g. `check-review-gate.sh <worktree> <task-id> <task-sha>` (Go: check-task-commit-fields already parses both fields) printing FIRE/QUIET + reason | `**The review gate.** After the guard passes a task's commit, the` | med | check-dispatch-paragraphs.sh header names implement.md "the one statement of the gate's threshold … and the one site to re-tune" — mechanizing moves the constant into Go; update that header. |
| MX3 | implement.md:806-847 | 3 271 | MECHANICS | Task-close boundary Bash call (dispatch end, check-task-commit-fields single/map form, gate, tick, push, check-task-commit-planning-paths) → `flow task close …` with a 0/1/2 contract | `2. **One Bash call: the implementer's record dispatch end, the guard` | low-med | Big: exit-2-stops / exit-1-recommit semantics must survive as the command's contract. |
| MX4 | implement.md:249-255 | 576 | MECHANICS | State-file worktree read-merge-write (`flow state get` → merge → `flow state set`, never drop peers, keep `state`) → `flow state add-worktree <name> <abs-path> <merge-base>` (no such verb in stats/cmd/flow/state.go today); brainstorm.md step 3 runs the same recipe | `Read the current record with flow state get, merge in this` | med | Timing sentence ("once per worktree, immediately after…") stays. |
| MX6 | implement.md:199-209 | 798 | MECHANICS | Base refresh (fetch, compare merge base, `git show` specs at moved base, `git diff --name-only … -- <Files: paths>`) → script printing MOVED/UNMOVED + intersecting paths | `Before the first task dispatches — inline or as an implementer` | med | Collision handling (pivot) and "never rebases" stay. |
| MX2 | implement.md:562-572 | 448 (in L2) | MECHANICS | Throwaway-worktree snapshot recipe, duplicated (plus a `.superpowers/sdd` copy) at review-panel-optional-slots.md:65-76 → `snapshot-worktree.sh <src> <dest> [--sdd]` | `git -C <worktree> worktree add --detach <worktree>-wave-group-<g> HEAD` | med | Counted under L2. |
| MX5 | implement.md:315-327 | 1 216 (in L4) | MECHANICS | Merge-order record extension/creation rules → script | `**Then make the merge-order record cover the whole set.** The canonical` | low | Counted under L4. |
| MX7 | implement.md:961-983 | (in L3) | MECHANICS | guard-autosquash targets → aside → `--fixup` → `rebase --autosquash <sha>^` → restore → guard-autosquash after; same fold at review-panel.md:783-798 → `fold-fixup.sh` | `**Rewrite-based folding is for unpushed history only**: the fix commits` | med | Counted under L3; would also end drift #10. |

## Top 5

1. **L1** — move `## 3. Documenting a fix` (336-451) to a fix-run-only file: **7 201 B** net, every first implementation session (IMPLEMENTATION). Needs one Go edit: add the basename to `smcCandidates`.
2. **L2** — move the sdd-only blocks (per-group implementer, Waves, per-group gather, sixth-argument scoping): **4 601 B**, every `inline` run (IMPLEMENTATION). No guard edits; repoint 4 citers.
3. **RATIONALE batch** — move R1-R32 verbatim to SKILL-rationale.md: **4 734 B** (3 264 B outside the lazy blocks), every IMPLEMENTATION session.
4. **L4** — move cross-repo/apps worktree creation, `spectre link`, merge-order write: **3 696 B**, every single-repo run (IMPLEMENTATION).
5. **L3** — move the gated-reviewer `fix` path (950-989): **2 834 B**, runs where no gated pass returns `fix` (IMPLEMENTATION). It absorbs the 1 392 B restated Panel-re-runs route.

## Condition map — answers to the task's specific questions

### Blocks an `inline` run never acts on (execution `sdd` only)

- **Movable as-is (L2): 4 771 B** — 541-548 (746), 550-582 Waves incl. the snapshot recipe (2 267), 584-592 gather lead + command (481), 599-601 `<changeRoot>`/`<principles-path>` (220), 612-619 gather failure / `test -f` / never-read-back / stderr report (722), 621-624 sixth-argument scoping (335).
- **sdd-only fragments inside mixed paragraphs, not cleanly movable: ≈ 2 117 B**:
  - 502-507 implementer `-model`/`subagent_type` (540) and 507 zcode pair (147);
  - 730-731 implementer handshake compare (141);
  - 802-805 "next implementer overlaps the guard" framing and step 1 (273);
  - 855-856 step 3 (187);
  - 927-930 "The bundle is the implementer group … xhigh → high" (311);
  - 462-464 SDD / dispatching-parallel-agents override (168);
  - table row 11 (SDD skill, 86);
  - table rows 49 and 52 (implementer, panel-fix: 264; the Inline section defines itself as "this same table minus" these two rows, so they stay).
- **Dispatch-mechanics paragraphs the inline parent cannot act on as written: ≈ 2 264 B**, all guard-pinned or cited. They are CONTEXT BUNDLE 704-709 (554), PROJECT HAZARDS 711-714 (319), MODEL HANDSHAKE block 725-728 (228), TOOLS 737-740 (392), NO DELEGATION 742-746 (419) and REPORT FILE 775-779 (352).
- **Bind the parent inline "in the same words" (152-155), so they are NOT sdd-only:** FLOW — COMMIT-PER-TASK (641-670), the Build: red rule (672-676), both REQUIRED SUB-SKILLs, PROVE THE GUARD BITES, PIN BEFORE REFACTOR, REQUIRED READING, PLAN PROVENANCE, FOREGROUND BUILDS, TARGETED TESTS, PROOF RUNS, OUTPUT BUDGET, REPRODUCE (implementer), REPORT DON'T DECIDE. In any case check-dispatch-paragraphs pins them in implement.md.
- **Other paragraphs that look sdd-only but also serve inline:**
  - dispatch-record semantics 474-511 — cited by review-panel.md:376 and visual-verify.md:53/143; inline uses the same calls with `-agent-id inline`;
  - `<shape>` 594-597 and `<canonical-worktree>` 603-610 — review-panel.md:114-115;
  - full-suite 789-800 — TARGETED TESTS says the full list runs "at the last bundle";
  - task-close step 2 at 806-847 — Inline: "runs `check-task-commit-fields.sh` and ticks the task exactly as section 4 states";
  - mutator rule 459-472 — may bind the verifier, a writer; keep.
- **Total an inline run never acts on: ≈ 9.2 KB**, of which 4.77 KB is movable without code changes.

### Fix run only (section 3)

- 336-451 (7 500 B) — L1. The rest of the file's fix-run text is routing (25-29) plus the tooling-analyst table row (54). Both stay.

### First run only (section 2)

- **The heading "(first run only)" is stale.** Section 2 runs on every run: table l.9 "every run"; l.28 fix-run order "document-fix → load-context → isolate (resume) → sdd-tdd"; l.254 "on a fix or resumed run"; l.329-330 "on a fix run exactly as on the first".
- **The parts that act only when a worktree is created or the set grows** are 255-258 (286, informational) and L4's 268-285, 294-313 and 315-327. On a single-repo change they are no-ops on every run.

### Harness `zcode` only

- 115-121 (671 B) — L5.
- 507, one sentence (147 B). It is the implementer-site pointer (also sdd-only). Too small to move.

### Only when a gated per-task reviewer fires

- 924-950 (2 411 B): bundle rules, keys, outcomes, Minor → KNOWN-BUGS.
- 950-989 (3 214 B): the `fix` path only — L3.
- 991-1042 (3 658 B): the bundle prompt. It is guard-pinned (READ-ONLY REVIEW; REPRODUCE reviewer variant; one each of FOREGROUND ×3, TOOLS ×2, HANDSHAKE ×2 and NO DELEGATION ×2 minimums), so check-dispatch-paragraphs forbids moving it.
- Plus "A gated task's tick defers…" (876-877) and step 3 (855-856).
- The gate is a poor lazy condition: >40 changed lines or any undeclared path fires on most non-micro runs. Only L3 is worth splitting.
- **Look gated-only but are not:**
  - plan-tree guard 1044-1055 and content markers 1057-1072 — review-panel.md:393-407 cites both for every panel round;
  - check-task-reviewer-single-dispatch 1144-1164 — runs at every stage close.

### Restatements of review-panel.md / verify-and-handoff.md / SKILL.md / pipeline.md (both locations)

**review-panel.md**

| implement.md | review-panel.md | what is restated | verdict |
|---|---|---|---|
| 956-973 | 775-798 (**Panel re-runs**) | on-top vs rewrite fold, pathspec staging, `--fixup`, `<task-sha>^`, guard-autosquash | drifted; see ROUTE / L3 |
| 136-138 | 144-166 | Bugbot/Security prompt-driven (also SKILL.md:117-119) | D40 |
| 76-88 | 160-166, 176-181 | `flow-<effort>` allowlist omits Agent; flow-low on a default panel | FLOWEFF, keep |
| 507 | 181, 1335 | the zcode pair sentence, per site | keep, pointer |
| 599-601 | 663-665 | `<principles-path>` ≡ [PRINCIPLES_PATH], verbatim | goes with L2 |
| 618-619 | 117-118 | report the gather's stderr line, per site | keep |
| 1046-1052 | 393-407 | plan-tree snapshot/verify | per-site application; review-panel cites implement.md for the stop |
| 1147-1164 | 1413-1428 | single-dispatch close + prompt | parallel shape, different guard and wording — not a duplicate |
| MODEL HANDSHAKE / TOOLS / FOREGROUND / NO DELEGATION / OUTPUT BUDGET / TARGETED TESTS / REPRODUCE | each file's own dispatch sites | verbatim paragraphs | required per site by check-dispatch-paragraphs — not cuttable |

**verify-and-handoff.md**

- Nothing substantive. It cites implement.md for Turn discipline (193), the resolved set (25-26, 333-335), Inline records (111-113, see drift #7) and **3. Documenting a fix** (438-439).

**SKILL.md** — canonical is implement.md; the copy sits in SKILL.md, every session.

- SKILL.md:218-220 ↔ implement.md:553-560 (wave cap).
- SKILL.md:233-237 ↔ implement.md:19-24 (parent runs 1/2/4 + panel + verify).
- SKILL.md:117-119 ↔ implement.md:136-138 (Bugbot/Security).
- SKILL.md:109-115 ↔ implement.md:86-88 (verifier on flow-low).
- SKILL.md:11 claims "Nothing here duplicates that file's own content".

**pipeline.md**

- pipeline.md:209-210 ("session token MUST be a literal") ↔ implement.md:498-499 ("`-session-token` takes a literal, never a shell substitution"). Different command (dispatch record vs stage mark) — keep.
- pipeline.md:212 ("A mark never blocks") ↔ implement.md:112-113 / 510-511. The latter is canonical for records.
- pipeline.md:257-264 (never stage spectre/changes) ↔ FLOW — COMMIT-PER-TASK. The subagent prompt must be self-contained — keep.

**git-boundaries.md** (same session)

- Its 86-88 ↔ implement.md:651-653 (":(exclude) governs what an add adds…"), inside the implementer prompt — keep.

**Internal duplicates**

| implement.md | implement.md | what | verdict |
|---|---|---|---|
| 546-548 | 672-676 | Build: red partner, one commit | D39 |
| 637-638 | 510-511 | record journals / never blocks | D17 |
| 1130-1138 | 1003-1011 | FOREGROUND + REPRODUCE (reviewer), verbatim | required, min 3 each — keep |
| 67-68 | 793-795 | the 5-minute prompt cache | both RATIONALE |

### MECHANICS candidates

- **Standalone:** MX1 (review gate, 574), MX3 (task-close call, 3 271), MX4 (`flow state add-worktree`, 576; shared with brainstorm.md step 3), MX6 (base refresh, 798).
- **Inside lazy blocks:** MX2 (snapshot recipe, shared with review-panel-optional-slots.md:65-76), MX5 (merge order), MX7 (fixup fold, shared with review-panel.md:783-798).

### Guard constraints

- **check-references.sh.** Headings reached by checked citations must stay resolvable in implement.md:
  - "Dispatch sites — the parent's closed list" — visual-verify.md:12;
  - "Inline — the parent implements" — review-panel.md:1328;
  - "The parent orchestrates directly" — visual-verify.md:74, review-panel.md:539 and 1239.
  - No candidate moves them.
  - L1 moves `## 3. …`. Its citers write `**3. …**`, which contains "." and is therefore unchecked (crLooksLikeSection), so the guard stays green. Repoint them anyway for correctness.
- **check-dispatch-paragraphs.sh** (dpSites, stats/internal/guard/dispatchparagraphs.go).
  - Forbids, as-is: moving 641-787 or 991-1042; cutting either parent copy at 1128-1138; any edit that drops a listed phrase.
  - None of L1-L5 / R / D / M touch a pinned label. R27 and M38 edit inside READ-ONLY REVIEW but keep its three phrases.
- **check-stage-mark-calls.sh.** Only L1 moves a `flow stage begin` line. Its corpus is a basename allowlist (`smcCandidates`), so the new file needs adding. Otherwise the call is silently unchecked rather than failed.
- **check-guard-symlinks rule 2.** Scans every `*.md` in the skill dir, so new sibling files are covered.
- **check-normative-inventory.sh.** No cut removes an uppercase MUST/SHALL sentence. Moves keep the set identical.
- **Every new lazy file** also needs listing in implement.md:1121-1124 ("Phase files read once per run") and a SKILL.md Stage-keys-style note, following the precedent of review-panel-optional-slots.md / visual-verify.md.

## Drift / bugs found

1. **Stale heading.** `## 2. Isolate the workspace (first run only)` contradicts l.9, l.28, l.254 and l.329-330: section 2 runs on fix runs too. The worktree has been created at `flow.kickoff` since 18feb597. The same name is mirrored in stats/internal/stages/names.go:77 and README.md:149, so a fix is a coordinated rename, not a trim.
2. **Stale wording at l.341-342.** "a first run creates the worktree instead, per **2** above" — section 2 resumes the worktree that brainstorm.md step 3 created.
3. **Wrong file at l.73-75.** It says every row's prompt carries NO DELEGATION in `verify-and-handoff.md`, but that file has 0 such paragraphs. The verifier dispatch moved to visual-verify.md, which carries 2 `**NO DELEGATION:**` blocks (per the check-dispatch-paragraphs.sh header).
4. **Tooling analyst missing from the inline list (l.90-92).** "This same table minus the implementer and panel-fix rows" is 4 rows. The next clause says "the parent's only permitted dispatches inline are the panel-bundle, gated per-task-reviewer and verifier rows", which is 3. The tooling-analyst row (e231d5e5, 2026-09-28) was not added. So on an inline fix run with a missed visual defect, the analyst dispatch is both allowed and forbidden.
5. **Role lists omit `planner`.** l.496-497 gives `-role` ∈ {implementer, reviewer, panel-fix, verifier}, but the tooling analyst records `-role planner` (visual-verify.md:141; `recordRoles` in stats/cmd/flow/record.go:54). l.95's handshake role list also omits the tooling analyst, which carries MODEL HANDSHAKE (visual-verify.md, 2 blocks).
6. **zcode handshake conflict.** l.115-121 says a missing or other `Model:` line is not a mismatch. model-policy.md:113 says "The handshake compares against `glm-5.3-flash`." The two disagree.
7. **Suffix convention cited but never stated.** verify-and-handoff.md:110-113 says the `-<worktree basename>` suffix is "the same convention **Inline — the parent implements** (implement.md) uses for implementer and panel-fix rows". implement.md has no "basename" anywhere, and `git log -S"worktree basename"` on it is empty.
8. **Loose argument count at l.621-624.** "The panel's and the fix subagent's bundles … keep the five-argument call" — but review-panel.md:114-115 passes 8 arguments (`"" <canonical-worktree> <shape>`), and l.603 says the seventh argument is "passed on every call". review-panel.md:506 repeats the "five-argument" wording.
9. **Direction error at l.575-576.** It says "the task-close step above", but that step is below (806-816).
10. **Duplicated fold route has diverged.** implement.md:956-989 and review-panel.md:775-798 state the same fold, but:
    - only implement.md wraps the rebase in `aside-planning-artifacts.sh` and states the keep-both-sides conflict rule;
    - only review-panel.md says a guard-autosquash refusal "stops the round", while implement.md never says what a refusal does;
    - only review-panel.md has the empty-fold drop and the "clean autosquash is not evidence" rules.
11. **Inline runs never see project hazards.** PROJECT HAZARDS (711-714) "binds the parent in the same words" (152-155). But `## hazards` travels only in the per-group bundle, which inline never gathers (584: "immediately before that group's implementer goes out"), and l.617 forbids reading a bundle back. No other flow file surfaces hazards (grep); `flow hazards` exists but nothing calls it.
12. **"Binds the parent in the same words" is broader than intended (152-155).** It sweeps in NO DELEGATION ("Never call the `Agent` tool"), which contradicts the inline parent's permitted panel, gated-reviewer and verifier dispatches (156-161). MODEL HANDSHAKE, TOOLS, CONTEXT BUNDLE and REPORT FILE also cannot apply to the parent.
13. **Gated reviewer's type/effort undefined on small/regular.** Runs with no groups dispatch "on `DEFAULT_MODEL`/`default`" (932-933). But `default` means "no effort set" (record.go:56-62), and no `flow-<effort>` definition exists for it. The "structurally fork-free on both panel shapes" claim (82-86) covers only the group-pair and `default`-panel cases. On the typical run the reviewer could go out on `general-purpose`, which review-panel.md:166 says cannot be tool-restricted.
14. **`micro` has no gated-reviewer rule.** It is not named in the bundling rule (930-934, table row 50). check-task-reviewer-single-dispatch treats it like `big` (taskreviewersingledispatch.go:143, 226).
15. **A fix run's re-Decide is unreachable.** brainstorm.md:259-265 says `flow.document-fix` (implement.md) "hands the appended plan through the same Decide step". Section 3 never mentions Decide, and a fix run loads neither brainstorm.md nor brainstorm-planner.md.
16. **SDD skill named unconditionally (table row 4, l.11).** It names `superpowers:subagent-driven-development` with no condition, and nothing says an inline run skips it. A typical run may load that external skill for nothing.
17. **Wrong section in a citation.** review-panel.md:344-346 cites the incident rule under **The parent orchestrates directly**; it lives in `## 4. Execute`.
18. **Guard location misattributed.** project-configuration.md:403-406 ("per **Isolate the workspace** in implement.md") and worktree-resolution.md ("the implement phase's workspace-isolation gate") both say implement.md runs the workspace-isolation guard. It doesn't — `prepare-workspace.sh` runs it in verify-and-handoff.md's Verify.
19. **`## worktree setup` never runs for additional worktrees.** project-configuration.md:28 says it runs "immediately after **2. Isolate the workspace** (implement.md) creates it". Section 2 creates app/peer worktrees but never runs it; only the kickoff (brainstorm.md step 4) and wave copies (l.557) do.
20. **Low: `flow-task-commit-fields` (l.858)** names the capability spec deleted in 18feb597 (`openspec/specs/myflow-task-commit-fields/spec.md`); spectre/specs/ is empty. It survives only as a field-family label (brainstorm-planner.md:336, build-green.md:39).
21. **Guard doc drift (not in this file).** The check-dispatch-paragraphs.sh header table gives visual-verify.md min 1 for TOOLS / MODEL HANDSHAKE / NO DELEGATION; `dpSites` in Go says 2.

## Not slimmable

- **Implementer dispatch paragraphs, 641-787 (~10.2 KB).** They bind the inline parent verbatim (152-155) and are pinned in implement.md by dpSites. Pass A already moved their incident narratives.
- **Gated reviewer bundle prompt, 991-1042 (3.66 KB).** Pinned: READ-ONLY REVIEW, the REPRODUCE reviewer variant, and the FOREGROUND / TOOLS / HANDSHAKE / NO DELEGATION minimums.
- **Parent FOREGROUND/REPRODUCE copies, 1128-1138 (743 B).** They are verbatim copies of 1003-1011, but both are needed to meet the min-3 counts.
- **Dispatch sites table, self-check, one-shot rule, handshake, return, dies (42-134).** This is the closed list, and three headings here are targets of checked citations.
- **Inline section (140-169).** It is the typical run's own contract.
- **Section 1 (171-228) and section 2's core (230-266, 287-294, 329-334).** They run on every run.
- **Dispatch-record semantics (474-511).** Cited by review-panel.md:376 and visual-verify.md:53/143.
- **`<shape>` and `<canonical-worktree>` (594-597, 603-610).** review-panel.md's gather uses them.
- **Incident carry (513-530, minus R15/R16).** Cited by review-panel.md:344-346, and must be known before an incident happens.
- **Substitution record (626-636).** Must be known at the moment of substitution.
- **Task-close boundary, gate, tick, corrections, Correction paragraph, pivot (802-922).** Every run's task close; MECHANICS is the only lever.
- **Gated reviewer rules (924-950).** The gate fires on most non-micro runs, so a lazy split would rarely save anything.
- **Plan-tree guard and content markers (1044-1072, minus R28-R30).** review-panel.md applies both to every round.
- **Turn discipline and Read discipline (1082-1126, minus R31/R32).** The canonical copies, cited from review-panel.md, verify-and-handoff.md and visual-verify.md.
- **check-task-reviewer-single-dispatch close and its prompt (1144-1170).** Runs at every stage close, and contains operator-prompt wording.
- **Mutator rule (459-472).** Mostly about sdd dispatches, but may bind the verifier, a writer, on inline runs.
- **The `flow-<effort>` paragraph (76-88, minus R2).** Keep until drift #13 is fixed.
