# Slimming analysis — `skills/flow/review-panel.md` + `skills/flow/review-panel-optional-slots.md`

> **Audit snapshot at `700e184` (2026-09-29).** A read-only model pass that followed [`rubric.md`](rubric.md); every row is a candidate to verify, not a verified fact. Line numbers refer to `700e184`. Main has since retired the bugbot and security slots (`ebdfdde1`..`4dafab4a`), so re-locate each row by its quoted first words, and treat bugbot/security rows as obsolete. Tracked in KAN-851 → KAN-857 (and the Drift items in KAN-853 / KAN-854). Plan: [`README.md`](README.md).

Read-only analysis. Both files read in full. Every byte count below is `sed -n 'A,Bp' | wc -c`, or the exact
UTF-8 length of the quoted span for sentence-level cuts. Every guard claim was checked by **simulation**: I
copied the scanned corpus into a scratch directory and ran the real guards against the copy through their
`CHECK_*_ROOT` overrides. Nothing in the repository was modified (`git status` is clean).

Two facts shape every move below:

- **No `review-panel-rationale.md` exists.** A verbatim RATIONALE move from either file goes to
  `skills/flow/SKILL-rationale.md`, into the existing section *"Moved by the 2026-09-22 prompt audit (pass A) —
  review-panel.md and implement.md"* (lines 264–507). An earlier audit already moved the incident narratives
  there, so the RATIONALE left in these files is residual.
- **Measured frequencies used for the LAZY-SPLIT conditions:**
  - Archived panel records (`spectre/changes/archive/*/panel.md`): 11 of 11 raised at least one finding, and
    9 of 11 raised a Critical or Important.
  - Archived proposals: 5 of 106 record a fix run.
  - `brainstorm-planner.md` rolls a compact roster when `compact_roll < 90`, which gives `primary+principles`
    in one dispatch. It rolls an `exp-` slot when `experimental_roll < 30`, and that slot joins only a second
    dispatch that has room.
  - The base branch took about 38 first-parent commits a day in September 2026.

## Totals

Totals are conservative. Only high and medium confidence candidates are counted, and each byte is counted
once: sentences cut from inside a split block are subtracted from that split. Guard-blocked candidates
(RP02 7,397 B and RP06 1,213 B) are excluded.

| file | bytes | RATIONALE | DUPLICATE | STALE | LAZY-SPLIT | MECHANICS | MISPLACED |
|---|---|---|---|---|---|---|---|
| `skills/flow/review-panel.md` | 99,985 | 2,191 | 2,968 | 484 | 38,842 | 4,622 | 916 |
| `skills/flow/review-panel-optional-slots.md` | 7,968 | 2,084 | 0 | 0 | 2,627 | 1,119 | 0 |

**The cut/move set was simulated and every guard stayed green.** The set is every RATIONALE, DUPLICATE, STALE and
MISPLACED row in the totals, plus RP04 and RP05 moved into optional-slots.

- It removes **7,488 B** from `review-panel.md`. Of that, 929 B moves into optional-slots.
- `review-panel-optional-slots.md` goes from 7,968 B to 6,814 B.
- The guards run were: `check-references.sh`, `check-dispatch-paragraphs.sh`, `check-mutation-reproducer-pin.sh`
  and `check-markdown-integrity.py`.
- The `check-normative-inventory.sh` output (the MUST/SHALL set) was byte-identical before and after.

## Candidates

`RP` = `skills/flow/review-panel.md`, `OS` = `skills/flow/review-panel-optional-slots.md`. The **LAZY-SPLIT
"false in"** column states how often the load condition does not hold, which is when the bytes are saved.

| id | file:startline-endline | bytes | lever | condition / canonical location / evidence | first ~10 words verbatim | confidence | behaviour-risk note |
|---|---|---|---|---|---|---|---|
| RP01 | RP:757-824, 852-927, 940-1172, 1193-1211, 1301-1391 (fix-round block, unpinned text only) | 33,803 (36,473 minus 2,670 of RP13–RP17 and RP26–RP30 counted in their own rows) | LAZY-SPLIT | **Condition:** load once a round has recorded a Critical or Important finding. This is decidable at the round's single `flow record finding` Bash call (690-691), because the parent types `-severity` itself. **False in:** every pass-1 turn of every run, and the whole remaining session for Minor-only or clean rounds (2 of 11 archived panels). | "**Every round this stage dispatches after pass 1 — each fix-round re-run —" | med | **Keep in main:** the `## Panel re-runs` heading, 825-851 (triage, deferral, mid-run rule), 928-938 (stale definition, cited by SKILL.md:226 and verify-and-handoff.md:59) and the `### Deferred findings…` section. **Re-point the citers of moved content:** implement.md:797 and 956, verify-and-handoff.md:189, brainstorm-planner.md:472 and 535, primary-reviewer-prompt.md:107, bugbot-reviewer-prompt.md:97, failure-modes.md:121, and internal lines 66, 179 and 371. **check-stage-mark-calls** stops scanning the panel-fix `flow record dispatch` pair (1342-1347); checked sites fall from 67 to 65 in simulation. Either keep that fenced block in main or add the new basename to `smcCandidates` (stagemarkcalls.go:72-76). **`/flow-fast`** consumes "Check base movement first through Panel re-runs" as written (flow-fast SKILL.md:92-97), so the load directive must sit inside that range. **Precedent edits:** a row in SKILL.md's Stage keys table and a mention in implement.md:1121-1122. **Simulated:** refs, dispatch-paragraphs, pin and integrity all green; the new file carries 5 checked citations. |
| RP02 | RP:1173-1192, 1212-1237, 1243-1300 | 7,397 | LAZY-SPLIT (guard-blocked, not in totals) | Same condition as RP01. | "**Every fix subagent's dispatch prompt also carries the VERBATIM REPORT — THE FACT" | low | **Forbidden by `check-dispatch-paragraphs.sh` as it stands.** `dpSites` (dispatchparagraphs.go:89-117) pins these paragraphs to `review-panel.md`: VERBATIM, FINDINGS, TARGETED, OUTPUT BUDGET, MUTATION PROOF and PIXEL PROBE (min 1 each), plus the second copies of FOREGROUND, TOOLS, HANDSHAKE and NO DELEGATION (min 2 each). Simulated result: exit 1, 10 violations. Moving them needs the dpSites paths, the shell header table and `check_dispatch_paragraphs_test.go` re-pointed. |
| RP03 | RP:271-323, plus the sentence at 934-938 | 4,110 (4,131 + 412 − RP21 − RP22) | LAZY-SPLIT | **Condition:** fix run only. The section itself says "never on a creating run"; decided at stage open. **False in:** every creating run's implementation session (fix runs appear in 5 of 106 archived proposals). | "**A fix run whose delta is small against an already-clean, already-verified stage" | high | **Must re-point** SKILL.md:217 and implement.md:408; simulation showed check-references fails without it. **The moved section carries no checked citation:** its new file needs a same-line `**Heading** (`path`)` citation in its header or a `crExpectedZero` entry (references.go); simulated without either, check-references exits 1 with coverage 0. **Unchecked mentions to update:** optional-slots:30-31, verify-and-handoff.md:413-414, RP:241 and 768. Contains M3 (280-299). |
| RP04 | RP:598-607 | 674 | LAZY-SPLIT (into optional-slots) | Roster carries `mutation`, which is optional-slots' existing load condition. False in about 90% of decided panels (compact roll). | "**The mutation slot's dispatch prompt also carries the MUTATION ENTRY CONTEXT paragraph**," | high | No citers. All guards green in simulation. |
| RP05 | RP:353-356 (sentence) | 255 | LAZY-SPLIT (into optional-slots) | Roster carries `bugbot` or `mutation`. | "Mutating roles (`mutation`, `bugbot`) are always the last passes of a bundle" | med | Carries a checked citation: RP coverage goes from 15 to 14, and OS gains one. |
| RP06 | RP:641-656 | 1,213 | LAZY-SPLIT (guard-blocked, not in totals) | Roster carries `bugbot` or `mutation`. | "### The mutation-testing brief" | low | **Forbidden by `check-mutation-reproducer-pin.sh`.** Line 653 is the file's only single-line "first 10 lines" statement, and the script hard-codes `PANEL` (line 92). Simulated result: exit 1. |
| RP07 | RP:163-166 | 310 | DUPLICATE | Canonical: implement.md:76-88 (the `flow-<effort>` `tools:` list omits `Agent`, and NO DELEGATION is backed by that capability). The `general-purpose` clause is rationale and moves verbatim to SKILL-rationale.md. | "The `flow-<effort>` definitions are this repository's own, whose `tools:` allowlist omits" | high | None. |
| RP08 | RP:241-246 | 327 | DUPLICATE | Canonical: SKILL.md:210-217 (Guardrails, loaded in every session); also RP:168-174. The copies agree. | "**This and the late-fix reduction below are the only automatic reductions, and" | med | None. |
| RP09 | RP:373-375 | 176 | DUPLICATE | Canonical: verify-and-handoff.md:377 and 415-419 for the `Panel:` line groups. The pass log is rendered from the dispatch rows. | "The rendered panel record's pass-log section and the `IN_PROGRESS` handoff's `Panel:`" | med | None. |
| RP10 | RP:407-408 (clause) | 75 | DUPLICATE | Canonical: implement.md:1057-1058 ("the subagent's own clean-state claim never answers for the tree"). | "— a slot's clean report is never the answer to what happened" | med | Keep the sentence's final period. |
| RP11 | RP:538-542 | 307 | DUPLICATE (drifted) | Canonical: implement.md:94-113. That section is "stated once here, cited everywhere else" and names panel slots explicitly. | "The dispatcher compares that line against the model this slot was given and" | high | The cut also removes the drifted `## Question` wording (see B2). |
| RP12 | RP:752-753 (sentence) | 115 | DUPLICATE | Canonical: RP:1069 ("The parent records it, never the fix subagent."). | "A finding is recorded `fixed` by the parent at the fix round's" | med | None. |
| RP13 | RP:768-771 (bold sentence) | 324 | DUPLICATE | Canonical: RP:220-224, 301-302 and 168-172. | "**Pass 1 runs the roster **The docs-only reduction** or **The late-fix reduction** chose" | med | Inside RP01. |
| RP14 | RP:973-980 | 600 | DUPLICATE | Canonical: RP:438-446 (the declaration and the guard's audit) and 967-968. | "Before anything runs, the guard audits the instrument itself: each runnable reproducer" | med | Keep 980-983: it is the only statement that the premise audit tolerates missing lines. The hand-run fallback survives in 443-446. Inside RP01. |
| RP15 | RP:983-987 | 334 | DUPLICATE | Canonical: RP:447-449, 452-453 and 967-968. | "Findings at any other status claim nothing about the current tree and are" | med | Inside RP01. |
| RP16 | RP:1091-1094 (second sentence) | 296 | DUPLICATE | Canonical: RP:1269-1273 (MUTATION PROOF, same file). | "The mutation flips the line the fix changed — the code the added" | med | The bold sentence and "The flip is recorded…" stay, so "The flip" keeps its antecedent. Inside RP01. |
| RP17 | RP:1351-1352 | 104 | DUPLICATE | Canonical: RP:928-929 and SKILL.md:225-227. | "A Minor either fixed or deferred blocks nothing; a Minor left `open` blocks" | med | Inside RP01. |
| RP18 | RP:686-688 | 174 | STALE | Commit `ebaea04fa` removed the `code-review-low` slot and the qualifier "Primary and Code review (low) overlap" with it. The rule now contradicts RP:341-342 (each finding records its single role) and RP:366 (no de-duplication across roles). | "When two independently dispatched slots raise the same defect, the dispatcher records" | med | This is a live normative sentence. Deleting it settles the contradiction in favour of 341 and 366, so the dispatcher decides. |
| RP19 | RP:1238-1242 | 310 | STALE | The retry key `<round>-fix-retry` contradicts RP:1321-1322 and the guard's `pfdKeyRE` `^panel-fix-[0-9]+(-[0-9]+)?(-retry)?$` (panelfixsingledispatch.go:37). The rest duplicates implement.md:94-113. The model is stated at 1332-1335. | "Dispatched on `DEFAULT_MODEL` (below); the dispatcher compares that line against it and" | high | The cut removes a bug (see B1). |
| RP20 | RP:264-266 | 203 | RATIONALE | Explains why the first line exists; the rest of this passage is already at SKILL-rationale.md:301-305. | "The first line is the file's own answer to the reviewer who reads" | high | None; the `printf` carries the semantics. |
| RP21 | RP:273-278 | 250 | RATIONALE | Why the late-fix path exists. | "An appended fix that a full pass would bury under re-covered ground" | med | Inside RP03. |
| RP22 | RP:297-299 | 183 | RATIONALE | Condition 5 is fully stated before this passage. | "A consuming project typically carries none of those paths and the condition" | med | Inside RP03. |
| RP23 | RP:398-399 (clause) | 105 | RATIONALE | Why the plan-tree verify exists. | "— the slots read those artifacts, and a flight that changed them" | med | Keep the sentence's final period. |
| RP24 | RP:401-403 (clause) | 143 | RATIONALE | Defense-in-depth rationale. | "the slots' own read-only briefs are the first line of defense, this" | med | Keep the sentence's final period. |
| RP25 | RP:487-490 (clause) | 295 | RATIONALE | Describes the internals of `prove-reproducer.sh`; line 484 already says "against a scratch worktree at the pre-fix commit". | "— the detached scratch worktree it materializes at the pre-fix commit, with" | med | Hand-run fallback retained by 484. |
| RP26 | RP:762-766 | 316 | RATIONALE | Why the base is re-checked at each round boundary. | "The entry check ran once, before pass 1; a base that moves" | high | Inside RP01. |
| RP27 | RP:907-908 | 107 | RATIONALE | Why the rounds are scoped. | "The scoping exists because a round that re-reads a growing fix diff" | high | Inside RP01. |
| RP28 | RP:1075-1078 | 259 | RATIONALE | Incident narrative (KAN-582 F1) and why the guard class exists; the ordering rule itself stays. | "A `fixed` recorded in the same call as its re-run stands verified" | high | Inside RP01. |
| RP29 | RP:1144-1146 | 152 | RATIONALE | Why exit 2 stops the run, and what the guard catches. | ": a handback cannot repair an inability. This catches an undeclared file" | med | Inside RP01. |
| RP30 | RP:1160-1162 (clause) | 178 | RATIONALE | Keeps "binds the fix round every run — the obligation is the round's, not a slot's". | ", so a run where neither Bugbot nor Mutation is in the resolved" | med | Inside RP01. |
| RP31 | RP:1432-1434 (clause) | 112 | RATIONALE (not counted) | Why the build-green close check exists. | ", so a pass that raised nothing and closed without a fix round" | low | Excluded from totals. |
| RP32 | 7 parenthetical tags `(KAN-658)` / `(KAN-839)` at 457, 474, 958, 971, 983, 1010 and 1057 | ~80 | RATIONALE (not counted) | Provenance tags; the prior audit kept parentheticals. | "(KAN-658)" | low | Excluded from totals. |
| RP33 | RP:723-751 | 916 | MISPLACED | The renderer produces the table and marker blocks (stats/internal/records/render.go:151-179). Downstream reads the store, not the record (verify-and-handoff.md:364-371). The parent never writes this format. | "The record carries a findings table, one row per finding:" | med | Keep 720-721 (render even when zero findings). The comment at installedcitations.go:775 names this example row and would go stale (cosmetic only). |
| RP34 | RP:23-40 + 61-68 | 1,909 | MECHANICS | A single call could run `resolve-base-branch` → `check-base-moved` → the automatic rebase when MOVED with no overlap → re-check, and print the new merge base. The prompt and the Stop, Continue, Rebase and Conflict prose stay. | "Once per worktree in this run's resolved set (**Resolving a change's worktrees**," | med | Needs code and parity tests. It is a mutating step, so its exit contract must name the conflict outcome. |
| RP35 | RP:247-262 + 590-596 | 1,554 | MECHANICS | One writer could produce `final-review.diff` and `[TOUCHED_FILES]`, and also `slot-delta-*` (852-862), `late-fix.diff` (301-305) and `fix-round-N.diff` (771-773), which sit inside RP01 and RP03. | "Write `<abs-worktree>/.superpowers/sdd/final-review.diff` (the canonical worktree's) once per round from" | med | Needs code. |
| RP36 | RP:332-337 | 513 | MECHANICS | Deterministic default-panel grouping and truncation. | "On a `default` panel, the settings-store roster is grouped deterministically by the" | low-med | Default (micro) panels only. |
| RP37 | RP:206-210 + 237-240 | 646 | MECHANICS | The guards could write their own `flow record pass` rows (count, cap, proceed decision; verdict, printed path, roster). | "Exit 0 proceeds. Exit 1 proceeds too, unasked: the panel dispatches reading" | low-med | Needs code. |
| RP38 | RP:220-229 | 785 | LAZY-SPLIT (not recommended) | Docs-only exit 0. Docs-only branches are common in this Markdown repo and rare in code projects. SKILL.md:215 cites the heading. | "**Exit 0 from every worktree in the resolved set — every path this branch" | low | Small; the directive's cost is comparable. |
| RP39 | RP:36 (after "Report every worktree's verdict:") to 79 | ~3.2 KB | LAZY-SPLIT (not recommended) | Any verdict other than CLEAR. The base took about 38 commits a day, so MOVED is nearly always true. | "`MOVED` with no overlap is confirmed conflict-free and rebases that worktree automatically" | low | Overlaps RP34. |
| RP40 | RP:176-183 (decided panel only) / RP:332-337 (default panel only) | 969 / 513 | LAZY-SPLIT (not recommended) | Decided panels cover every class except micro; the rest is sentence-level. | "**On a decided panel** (the decision's `panel` an object), model and effort belong" | low | Not contiguous enough to split. |
| RP41 | RP:783-812 | 2,487 | LAZY-SPLIT (alternative to RP01, not counted) | Condition: the fixup target is not on the remote. Branch backup (git-boundaries.md:79-85) pushes every commit, so the condition is nearly always false. implement.md:961-989 carries the same route. | "**Rewrite-based folding is for unpushed history only**: the fixup — stage first" | med | Nested in RP01; use only if RP01 is rejected. |
| RP42 | RP:429-502 | ~6.5 KB | MISPLACED (structural, not counted) | Could become a template passed by path, per this file's own rule at 152. | "**Every slot must supply, per finding, a reproducer**: a runnable command that" | low | The parent also acts on it (bounces, inline repair). The bundle prompt lists it as a pasted shared paragraph. |
| RP43 | RP:1080-1081 (clause) | 89 | DUPLICATE (skip) | Canonical: implement.md:1113-1115. | ": the specific hunks a finding names, never the whole fix diff" | low | The `<!-- refs-guard:allow -->` on 1080 must survive, or check-references flags **Read discipline** (it is not a heading). |
| RP44 | RP:1165-1166 | 135 | DUPLICATE (skip) | Canonical: RP:118-119. | "Report the script's stderr line (`bundle unchanged — reusing …` or" | low | None. |
| OS01 | OS:80-94 | 1,357 | RATIONALE | Why `git diff HEAD --binary`, `--allow-empty`, the `-z`/NUL loop and the scaffold lines. It records rejected alternatives (bare `git diff`, the `awk` form). Moves verbatim to SKILL-rationale.md under "review-panel.md — The throwaway worktree". | "`git diff HEAD --binary` — against `HEAD`, not a bare `git diff --binary` —" | high | None; the recipe is code. |
| OS02 | OS:116-120 | 376 | RATIONALE | Explains what the fold-back recipe does. | "It rescues the slot's reproducers beside its reports — both are recorded as" | high | None. |
| OS03 | OS:46-50 | 351 | RATIONALE | Why a throwaway worktree is needed (the collision). | "Bugbot and Mutation both mutate code in place to run their brief;" | med | The next sentence's "therefore" still reads. |
| OS04 | OS:7-43 | 2,627 | LAZY-SPLIT | Its own file, loaded only when the roster carries `exp-`. That is at most 30% of optional-slots loads, because `exp-` rides only regular-class full rosters (brainstorm-planner.md:463-469 and 526-527). | "## Experimental slot" | med | Re-point failure-modes.md:5 (a checked citation) and RP:186. Line 11 is this file's only checked citation, so the remaining throwaway half needs one; RP05's moved sentence supplies it. OS:28-35 (560 B) inside it restates RP:220-229, 301-309 and 920-926. |
| OS05 | OS:64-78 + 100-112 | 1,119 | MECHANICS | A second caller exists: implement.md:562-572 is the same create recipe. One `throwaway-worktree.sh create\|remove` with parity tests would serve both. | "```bash git -C <worktree> worktree add --detach <worktree>-<slot>-<round> HEAD" | med | Needs code. |

## Top 5

1. **RP01, the fix-round lazy split: 33.8 KB, or 41.2 KB if `dpSites` is re-pointed.**
   - This is the implementation session's review-panel stage.
   - It comes off every pass-1 turn of every run, and off the entire panel-plus-verify tail for Minor-only or
     clean runs.
   - Cost: 8 citers to re-point, and one stage-mark call-site decision.
2. **RP03, the late-fix lazy split: 4.1 KB.** It comes off every creating run's implementation session. Fix
   runs are the minority.
3. **The validated cut/move set, RP07–RP33 plus RP04 and RP05: 7.5 KB off `review-panel.md`.**
   - 4.1 KB of it is text read on every panel-stage turn.
   - Every guard stays green and the normative inventory is unchanged.
   - It also removes the stale `<round>-fix-retry` key and the drifted `## Question` handshake clause.
   - This is the safest move and needs no split machinery.
4. **Optional-slots: 2.1 KB of RATIONALE (OS01–OS03) plus a 2.6 KB Experimental-slot split (OS04).** This
   saving applies only in sessions that load the file, which is about 10% of panels.
5. **MECHANICS scripts.**
   - Base-movement call: 1.9 KB.
   - Diff writer: 1.6 KB, plus 2.9 KB inside RP01 and RP03.
   - Late-fix trigger guard: 1.6 KB inside RP03.
   - Throwaway-worktree script: 1.1 KB in OS plus 0.4 KB in implement.md.
   - This is the biggest change: it needs code and parity tests.

## Drift / bugs found

**B1. Bug: the stale retry key.**
- RP:1240 retries the fix subagent's handshake under `<round>-fix-retry`.
- RP:1321-1322 and the guard require `panel-fix-<round>[-<n>]-retry` (`pfdKeyRE`, panelfixsingledispatch.go:37).
- Following 1240 makes `check-panel-fix-single-dispatch.sh` exit 1 with "out-of-shape panel-fix key".
- RP19 removes the stale clause.

**B2. Handshake second mismatch.**
- RP:541 and 1241 say "a fallback plus `## Question`".
- The canonical text at implement.md:109-112 asks through **AskUserQuestion**, with the options Continue on
  `<model>` or Stop the run.
- operator-prompts.md:88 lists this case under "No recommended option".
- RP11 and RP19 remove both drifted copies.

**B3. Contradictory de-duplication rules.**
- RP:686-687 records one row with a `+`-joined `-slot`.
- RP:341-342 says "Every finding still records its own single role in `-slot`".
- RP:366 says "No de-duplication across roles … two `F<n>` rows".
- Cause: commit `ebaea04fa` dropped the qualifier that confined the rule to the Primary / Code review (low)
  overlap. See RP18.

**B4. `begin` order.**
- RP:387-388 says "every `begin` is recorded in the next Bash call" after the one-message launch.
- RP:422-423 says "record `begin` immediately before its launch".
- implement.md carries both orders too: 489-493 ("`begin` must precede the launch") and 855-856 ("The next
  Bash call records every launch's `begin`").

**B5. The two fix-commit routes have diverged.**
- **What the two copies carry:**
  - RP:775-812 and implement.md:956-989 both carry "**A branch the remote already holds takes the fix as one
    new commit on top, never a rewrite**".
  - Only implement.md wraps the fold in `aside-planning-artifacts.sh <aside|restore>` (KAN-628; added in
    `6773d527`, "cite … at both pipeline rebase sites"). Only implement.md states how a conflict is handled
    (resolve by hand, keeping both sides).
  - Only RP carries the empty-fold drop (807-812) and "a clean autosquash is not evidence" (800-805).
- **Why the missing aside matters:** RP's PLAN FIELDS paragraph (1195-1201) leaves uncommitted `tasks.md`
  edits in the tree at exactly the moment the fold's rebase runs.
- **Related gap:** RP's base-movement rebases (37, 56) also lack the aside step that integrate.md:117 carries
  for the same base rebase.
- **Dispatcher to decide** whether these differences are intentional.

**B6. A misdirected cross-reference.**
- RP:1408 says "the class **Recording findings**' ordering above added".
- The fixed-after-re-run ordering was added under **Panel re-runs** (1072-1075), in the same commit
  `ae39a985`, not under "Recording findings, and the record's format".

**B7. A broken reference the fix subagent cannot follow.**
- MUTATION PROOF (1283-1284) says "in the shape the review-panel contract's fenced block gives".
- The fix subagent's REPORT FILE (1307-1308) says "the lines this round's contract requires".
- The fix subagent never receives `review-panel.md`, so it never sees the `fix-mutation:` shape (1101-1105)
  unless the parent pastes it.

**B8. The inline panel-fix record is incomplete.**
- RP:1328-1330 says "records the pass with `-role panel-fix -agent-id inline`".
- implement.md:162-167 requires `-model <parent model> -effort <parent effort>`, with `begin` recorded before
  the first edit and `end` once the commit lands.

**B9. `-agent-id` is both forbidden and required.**
- RP:421 says "never typed, never invented".
- RP:1330 and implement.md:163 type `-agent-id inline`.
- implement.md:493-494 accepts a typed id "for a caller that knows the id".

**B10. Verify timing.**
- The plan-tree and marker verify runs "before any finding is recorded" (RP:397 and 407-408).
- The general rule at implement.md:1052 and 1065-1066 says "before the report is read or acted on".
- RP's wording lets the parent read slot reports before the verify.

**B11. Throwaway-copy removal timing.**
- RP:1000-1001 removes "every throwaway worktree … in one Bash call" together with the reproducer runs.
- OS:97-98 and RP:679-680 remove each copy, with its fold-back, as that slot's dispatch closes.
- By the time the reproducers run, nothing is left to remove.

**B12. Experimental-slot placement (cross-session).**
- OS:40-42 says the slot "joins whichever group has room".
- brainstorm-planner.md:463-469 says "nothing else ever joins the floor bundle"; `exp-` joins only a second
  dispatch that has room, otherwise it is `skipped — bundle cap` in the decision JSON rather than a pass note.

**B13. The non-convergence prompt.**
- RP:1363-1366 says "Only then does the run ask".
- The prompt has a recommended option, so Auto-resolution takes it without asking the first time
  (operator-prompts.md:39-45) and asks only on a repeat (72-73). RP's wording omits this.

**B14. TARGETED TESTS, fix-subagent copy (RP:1245-1250).** It keeps implementer wording ("this task's
`**Tests:**` field", "at the last bundle"). The pinned phrases limit how it can be edited. Low severity.

**B15. Bundle failure policy.** implement.md:614-615 says "the context bundle never gates a run", but RP:121-135
auto-resolves a bundle failure to Stop. The difference is scope (implementer versus panel), but the
implement.md sentence reads as run-wide. Low severity.

**B16. Bundle key naming.** RP:381 and 383 use `-key panel-<round>-<that slot>` for a bundle; implement.md:51
uses `panel-<round>-<slot+slot>`. Low severity.

**B17. Guard documentation drift (outside these files).**
- The header table in `scripts/check-dispatch-paragraphs.sh` gives `visual-verify.md` TOOLS, HANDSHAKE and
  NO DELEGATION a minimum of 1, where `dpSites` requires 2 (dispatchparagraphs.go:101, 104 and 108).
- The header table has no READ-ONLY REVIEW row, which the Go table has.

## Not slimmable

- **RP:512-597, the slot dispatch paragraphs.** Each one is pasted into every slot's prompt. FOREGROUND,
  TOOLS, NO DELEGATION, HANDSHAKE, REPRODUCE (reviewer variant) and ENTRY CONTEXT carry pinned minimum
  counts in `review-panel.md`.
- **RP:1173-1300, the fix-subagent paragraphs.** They are pinned (RP02); only a change to the Go `dpSites`
  table would move them.
- **RP:121-135, CONTEXT BUNDLE FAILURE, and RP:359-364, INDEPENDENT PASSES.** Both are pinned.
- **RP:429-502, the reproducer authoring rule.**
  - It is pasted into every slot's prompt and is the exit-code contract.
  - Lines 436 and 440 carry the pinned `# mutation-reproducer` marker.
  - The parent bounces and repairs against it.
- **RP:641-655, the mutation-testing brief.** It is pinned by the pin guard.
- **RP:21-35, 1399-1445 and 11-13.** These are the base check, the close guards and the stage marks. The
  marks are scanned by `check-stage-mark-calls.sh`.
- **RP:928-938, the stale definition.** SKILL.md:226 and verify-and-handoff.md:59 cite it, and `flow.verify`
  depends on it.
- **RP:825-851, triage, deferral and the mid-run rule.** This text decides whether a fix round opens at all,
  so it must be read before RP01 would load.
- **RP:137-150, the roster table.** It exhausts the `ValidReviewers` vocabulary; 9 files cite **The roster**.
- **RP:689-721, the finding and status record calls and the render.** They are needed whenever any finding
  exists, and the render happens every time.
- **OS:64-78, 95-112, 114-116 (first sentence) and 122-126.** These are the recipes and the isolation rules;
  the recipes stay until they are scripted (OS05).

## Specific questions

### Blocks needed only once a round has produced findings

**Decidability.** Yes, the condition can be decided at that point. The parent records every finding of a round
"in one Bash call, one `flow record finding` per finding" (690-691), after every report exists and before
triage (825-838) acts. It types `-severity` itself, so it knows the count and the maximum severity in that
turn. A sharper condition is also decidable: "the round recorded a Critical or Important." Only that condition
opens a fix round; a Minor-only round never reads the fix machinery (825-838).

**Frequency.** Archived records: "≥1 finding" held in 11 of 11 runs, and "≥1 Critical or Important" held in 9
of 11. Because the split loads lazily, it still saves the pass-1 turns in every run.

| block | lines | bytes | needed when |
|---|---|---|---|
| round-boundary base re-check | 757-767 | 909 | fix round (≥1 C/I) |
| pass-1 restatement + `FIX_BASE` / fix-round-N.diff | 768-774 | 542 | fix round (restatement = RP13) |
| fix commit routes (on-top + unpushed fold + autosquash + empty fold) | 775-813 | 3,203 | fix round (fold: unpushed only, RP41) |
| which slots re-run + Critical primary-source confirmation | 814-824 | 774 | fix round / ≥1 Critical |
| triage + Minor deferral + reason rule | 825-842 | 1,435 | ≥1 finding, any severity (stays in main) |
| re-run on deltas, FIX-ROUND SCOPE, third-round scoping, re-run cap, docs-only re-check | 852-927 | 6,141 | fix round |
| re-run staleness definition | 928-939 | 1,016 | **always**; cited, and read at `flow.verify` (stays) |
| dedupe + `check-panel-reproducers` + exit-contract guard + per-finding reproducer runs | 940-1020 | 5,344 | fix round (Minor-only rounds run no reproducer) |
| post-fix re-runs, flip + path condition, 3 close shapes, fixed-after-re-run ordering | 1021-1082 | 4,899 | fix round |
| mutation-proving the fix, the fix-diff walk, task-field guard, build-green round close, bundle rebuild, structured block | 1083-1172 | 6,490 | fix round |
| fix-subagent paragraphs (7,397 B pinned + PLAN FIELDS / ROUND SCOPE 1,328 B + handshake 310 B) | 1173-1300 | 9,035 | fix round (the paragraphs also bind the parent inline) |
| fix REPORT FILE, chunks, fixer model, dispatch record, non-convergence, closes on clean re-run | 1301-1391 | 6,843 | fix round |
| deferral to KNOWN-BUGS.md | 1392-1398 | 384 | ≥1 deferred Minor (stays) |
| finding / status record calls | 689-711 | 1,106 | ≥1 finding (stays) |

- **Total needed only when a fix round opens:** 44,180 B, of which 7,397 B is pinned.
- **Needed at any finding and staying in main:** 2,925 B.
- **The reproducer authoring rules are not findings-only.** Every slot's pass-1 prompt carries them (429-502).

### Blocks needed only under a condition

- **Docs-only reduction.** The exit-0 branch (220-229, 785 B) and the re-run reclassification (920-926, 627 B,
  inside RP01). Not worth splitting (RP38).
- **Late-fix reduction (fix runs).** Lines 271-323 (4,131 B) plus the staleness carve-out sentence at 934-938
  (412 B). These are needed only on a fix run and never on a creating run. **Recommended (RP03).**
- **Two dispatches versus one.** No contiguous two-dispatch-only block exists. The bundle prompt (349-366) is
  needed whenever any dispatch carries two or more roles, which includes the compact floor bundle
  `primary+principles` in about 90% of decided panels, and INDEPENDENT PASSES is pinned. Two dispatches occur
  only on regular or big full rosters.
- **Roster carries bugbot or mutation** (the existing optional-slots condition; false in about 90% of decided
  panels):
  - Movable: RP04 (674 B) and RP05 (255 B).
  - Blocked: RP06, the brief (1,213 B, pin guard).
  - Mentions too small to move: 467-469 and 870-871.
- **Default (micro) panel versus object panel.** The decided-only paragraph 176-183 (969 B) is needed for every
  class except micro. The default-only text is the grouping at 332-337 (513 B) plus sentence-level fragments
  (7-8, 139-142, 160-163, 306-308, 414-418). Not worth splitting (RP40).
- **Harness `zcode`.** Only two sentences (181 and 1335, about 150 B each), both verbatim copies of
  implement.md:507 and model-policy.md:107. The model-policy file is cited, not loaded. No block to split.
- **Moved base.** The non-CLEAR handling at 36-79 (about 3.2 KB) plus the round-boundary re-check (757-766,
  inside RP01). The base takes about 38 commits a day, so MOVED is nearly always true. Not worth splitting
  (RP39); mechanize the check instead (RP34).
- **Inline versus sdd execution.** The fix-dispatch mechanics are sentence-level and interleaved (1313-1335),
  and inline fixes are bound by the same paragraphs. No split.
- **Unpushed branch.** The rewrite and fold route (783-812, 2,487 B) is nearly always dead under Branch backup
  (RP41).

### Restatements of same-session files

| review-panel copy | canonical copy | state |
|---|---|---|
| RP:163-166 | implement.md:76-88 | consistent → cut (RP07) |
| RP:241-245 | SKILL.md:210-217 | consistent → cut (RP08) |
| RP:373-374 | verify-and-handoff.md:377, 415-419 | consistent → cut (RP09) |
| RP:387-391 | implement.md:1082-1101 (Turn discipline allows each file's own batch) | **drift**: begin ordering (B4) |
| RP:393-408 | implement.md:1044-1072 | **drift**: verify timing (B10); clauses cut (RP10, RP23, RP24) |
| RP:421-423 | implement.md:489-495 | **drift** (B4, B9) |
| RP:538-541, 1238-1241 | implement.md:94-113 | **drift** (B1, B2) → cut (RP11, RP19) |
| RP:344-347 | implement.md:513-525 | consistent (the added scope "a re-run, or the panel-fix subagent" is kept) |
| RP:775-812 | implement.md:956-989 | **drift** (B5); implement.md cites RP as canonical, then restates it |
| RP:1080-1081 | implement.md:1113-1115 | consistent; implement.md:1106 says "never restated" (RP43; refs-guard marker) |
| RP:1328-1330 | implement.md:162-167 | **drift** (B8) |
| RP:118-119, 1165-1166 | implement.md:618-619 | consistent, scoped per stage |
| RP:1351-1352, 928-929 | SKILL.md:225-227 | consistent → cut (RP17) |
| RP:159-163 | SKILL.md:117-121 | consistent (pointer) |
| RP:514-549, 1214-1260 (FOREGROUND, TOOLS, NO DELEGATION, HANDSHAKE, REPRODUCE, TARGETED, OUTPUT BUDGET) | implement.md:727-767, 995-1017, 1130-1138 | verbatim; one copy per dispatch site is pinned → not cuttable |
| RP prompts (42-50, 121-135, 628-636, 1368-1371, 1424-1427) | operator-prompts.md | cite-only, per the doctrine; wording gap B13 |
| RP:11-13, 1439-1441 | pipeline.md:202-204 (example) | pipeline restates RP, not the reverse |

### MECHANICS candidates

All five are candidates for code; each needs the code plus parity tests.

- **Base-movement check (RP34, 1,909 B).** One call would resolve the base, check it, rebase automatically when
  MOVED with no overlap, re-check, and rewrite the working-notes merge base. The prompt and the conflict prose
  stay as text.
- **Diff and entry-context writer (RP35, 1,554 B, plus about 2.9 KB inside RP01 and RP03).** It would produce
  `final-review.diff`, `[TOUCHED_FILES]`, `slot-delta-<round>-<slot>.diff`, `late-fix.diff` and
  `fix-round-N.diff`, all in the same per-worktree sectioned shape.
- **Late-fix trigger guard (280-299, 1,567 B inside RP03).** The five conditions become an exit 0 (reduce) or
  exit 1 (full path), printing `late-fix reduction: <n> changed lines since <sha>`.
- **Bundle planning.**
  - Default-panel grouping (RP36, 513 B).
  - Fix chunk plan and key minting (1313-1327, 1,412 B inside RP01). `check-panel-fix-single-dispatch`
    already checks this after the fact.
  - The re-run cap derivation (910-918, 817 B inside RP01).
- **Record calls (RP37, 646 B).** The guards could write their own pass rows. The dispatch `begin`/`end` pairs
  (376-391) could be recorded in one batch call from `decision.json`'s `panel.dispatches`; that saving is small.
- **Throwaway-worktree create and remove (OS05, 1,119 B).** A second caller exists at implement.md:562-572.

### Guard constraints

**`check-references.sh`**: citers of each moved heading, found by grepping the whole scanned corpus.

- **The late-fix reduction (RP03).**
  - Same-line citers, which the guard checks: SKILL.md:217 and implement.md:408. Both fail in simulation until
    re-pointed.
  - Unchecked mentions: optional-slots:30-31, verify-and-handoff.md:413-414, RP:241 and 768.
- **Panel re-runs (RP01).** The heading stays in main. Its moved content is cited by implement.md:797 and 956,
  brainstorm-planner.md:472 and 535, primary-reviewer-prompt.md:107, bugbot-reviewer-prompt.md:97,
  failure-modes.md:121 and verify-and-handoff.md:189. These are semantic re-points; the guard passes because
  the heading remains. SKILL.md:226, verify-and-handoff.md:59 and flow-fast SKILL.md:93 cite content that
  stays.
- **The fix round mutation-proves what it changed.** No `.md` citer; only the header prose of
  `scripts/mutate-and-verify.sh`.
- **The mutation-testing brief and MUTATION ENTRY CONTEXT.** No citers.
- **Experimental slot (OS04).** Cited by failure-modes.md:5 (checked) and RP:186.
- **Any new split file** needs at least one same-line `**Heading** (`path`)` citation or a `crExpectedZero`
  entry. The late-fix section has none of its own; the fix-round block has 5.

**`check-dispatch-paragraphs.sh`** (dispatchparagraphs.go). This guard **forbids** RP02, the 10 pinned
fix-subagent paragraphs; simulation gave 10 violations. It permits every other candidate; the RP01, RP03 and
cut simulations all stayed OK.

**Other guards found in the course of this analysis:**

- **`check-mutation-reproducer-pin.sh`** forbids RP06 (simulated exit 1).
- **`check-stage-mark-calls.sh`** uses the basename allowlist `smcCandidates`. A new file is not scanned, so
  moving 1342-1347 silently drops 2 checked call sites.
- **`check-markdown-integrity.py`** is why sentence cuts keep the terminating period, and why every lead-in
  "…paragraph**:" moves together with its blockquote.
