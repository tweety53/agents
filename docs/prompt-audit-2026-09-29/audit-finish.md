# Finish session slimming: integrate.md, archive.md, finish-contract-run1/run2.md, jira-followups.md

> **Audit snapshot at `700e184` (2026-09-29).** A read-only model pass that followed [`rubric.md`](rubric.md); every row is a candidate to verify, not a verified fact. Line numbers refer to `700e184`. Main has since retired the bugbot and security slots (`ebdfdde1`..`4dafab4a`), so re-locate each row by its quoted first words, and treat bugbot/security rows as obsolete. Tracked in KAN-851 → KAN-859 (and the Drift items in KAN-853 / KAN-854). Plan: [`README.md`](README.md).

Scope: the FINISH session. That is a bare `/flow` at `IN_PROGRESS`: integrate run 1, which on this operator's merge-and-push route chains straight into archive run 2 in the same session. This operator's own project settings (`<agents repo>/.flow/project.md`) are:

- `## default landing route: merge and push`
- `## self review: defer`
- `## jira: KAN`
- a `## stop` key that declares no command

The four finish files add up to 111,911 bytes on the chained route. A PR-route run-1-only session loads integrate.md plus run1, 48,886 bytes.

All byte counts are exact (`sed -n 'A,Bp' | wc -c` for whole lines, the exact substring for partial rows). Nothing in the repository was modified.

## Totals

Only high- and med-confidence rows are counted, and each byte is counted once. LAZY-SPLIT bytes are deferred to a sibling that loads on a condition, not deleted. MECHANICS bytes are the procedure text a script call would replace; they overlap nothing counted elsewhere. The `sum` column leaves MECHANICS out.

| file | bytes | RATIONALE | DUPLICATE | STALE | LAZY-SPLIT | MECHANICS | MISPLACED | sum (excl. MECHANICS) |
|---|---|---|---|---|---|---|---|---|
| integrate.md | 16838 | 160 | 2764 | 0 | 4026 | 460 | 0 | 6950 (41%) |
| archive.md | 22413 | 594 | 3422 | 142 | 3246 | 1100 | 0 | 7404 (33%) |
| finish-contract-run1.md | 32048 | 3551 | 4690 | 0 | 8446 | 0 | 0 | 16687 (52%) |
| finish-contract-run2.md | 40612 | 6468 | 2056 | 562 | 7789 | 3992 | 409 | 17284 (42%) |
| jira-followups.md | 35541 | 16435 | 0 | 0 | 10546 | 0 | 0 | 26981 (75%) |
| **total** | 147452 | 27208 | 12932 | 704 | 34053 | 5552 | 409 | 75306 |

Low-confidence rows are listed below but left out of the totals:

| file | lever | bytes |
|---|---|---|
| integrate.md | DUPLICATE | 136 |
| integrate.md | LAZY-SPLIT | 187 |
| archive.md | DUPLICATE | 612 |
| archive.md | LAZY-SPLIT (defer-only, which is always true here) | 987 |
| archive.md | MECHANICS | 858 |
| run1 | LAZY-SPLIT | 756 |
| run1 | MECHANICS | 289 |
| run2 | LAZY-SPLIT | 379 |
| run2 | RATIONALE | 589 |
| run2 | MISPLACED | 336 |
| jira-followups.md | RATIONALE | 689 |
| jira-followups.md | STALE | 58 |

**This operator's default FINISH session** is merge and push chained into run 2, `defer`, Jira configured, guards installed, a single repository, the unfinished-work gate `CLEAR`, and cleanup `COMPLETE`. Of its 111,911 bytes from the four finish files, the counted rows would take out:

| portion | bytes |
|---|---|
| RATIONALE, DUPLICATE, STALE and MISPLACED, always applicable | 24,818 |
| Hand fallbacks | 6,713 |
| Self-review run branch plus skip prompt | 6,544 |
| Standalone-reach rows | 1,674 |
| PR and manual route text | 455 |
| Unfinished-work gate apparatus | 2,923 |
| Leftover handoff | 446 |
| **Total when a base moved** | **43,573 (39%)** |
| Base-moved blocks, when no base moved | +4,752 |
| **Total when no base moved** | **48,325 (43%)** |

The high-confidence floor alone is 13,127 bytes (12%):

| lever | bytes |
|---|---|
| RATIONALE | 4,912 |
| DUPLICATE | 1,898 |
| STALE | 704 |
| MISPLACED | 409 |
| Hand-fallback LAZY-SPLIT | 5,204 |

## Candidates

Row ids:
- `I` = integrate.md
- `A` = archive.md
- `R1` = finish-contract-run1.md
- `R2` = finish-contract-run2.md
- `J` = jira-followups.md
- `M*` = MECHANICS

"(partial)" means the cut is the exact substring that starts with the quoted first words. The rest of the lines stay.

RATIONALE moves go verbatim to:
- `skills/flow-contracts/finish-contract-rationale.md` for run1 and run2
- `skills/flow/SKILL-rationale.md` for integrate.md and archive.md
- `skills/flow-contracts/jira-integration-rationale.md` for jira-followups.md; that file already has a "Moved from jira-followups.md" section

| id | file:startline-endline | bytes | lever | condition / canonical location / evidence | first ~10 words verbatim | confidence | behaviour-risk note |
|---|---|---|---|---|---|---|---|
| I01 | integrate.md:13-14 | 136 | DUPLICATE | SKILL.md:178-183 + pipeline.md:373-393 (guard presence check, run once at the top of every invocation) | **Check guard presence** per **Guard presence check** (`skills/flow-contracts/pipeline.md`), already run | low | Keep: the handoff `Guards:` line (integrate.md:286) points at "the guard presence check above"; SKILL.md files its guard-presence paragraph under the "A plain message at IN_PROGRESS" subsection, so this line is the only unambiguous pointer for a bare integrate run. |
| I02a | integrate.md:16-22 (partial) | 77 | DUPLICATE | finish-contract-run1.md:73-76 | resolve each affected repository's main checkout from the worktree set, | med | Keep the clause that runs both guards: integrate.md is the only skills/flow/ file citing check-foreign-staged.sh and check-main-checkout-drift.sh (check-guard-symlinks rules 2/6). |
| I02b | integrate.md:16-22 (partial) | 258 | DUPLICATE | finish-contract-run1.md:78-82, 103-107 | , and on any `STAGED-FOREIGN`, `DRIFT-BRANCH` or `DRIFT-DIRTY` stop and | med | Same guard-citation caveat as I02a. The final "." stays. |
| I03 | integrate.md:49-52 (partial) | 280 | DUPLICATE | finish-contract-run1.md:134-137 | `<canonical-worktree>` is the one member of the resolved set whose | med | The invocation line (L48) keeps `<canonical-worktree>`; verify-and-handoff.md:123 cites section 1 for "the same member", which still resolves. |
| I16 | integrate.md:61-72 | 645 | LAZY-SPLIT | Any worktree returned OUTSTANDING or VISUAL-VERIFY-MISSING (known right after the two step-1 guards). Usually false at integrate, because the implement phase ticks every task first. | - **`OUTSTANDING:`** → show the breakdown — and the guard's | med | Operator-prompt wording: move it verbatim. The CLEAR, empty-set and no-verdict bullets (L59-60, L73) and the mark lines (L91-93) stay. |
| I04 | integrate.md:75-77 (partial) | 185 | DUPLICATE | finish-contract-run1.md:162-163 (course table) | **Stop** exits leaving the change at `IN_PROGRESS` with nothing staged, | med | None. |
| I17 | integrate.md:76-89 (partial) | 915 | LAZY-SPLIT | Same condition as I16. | When the operator's answer says the verdict was verified structural | med | Carries the `flow record verdict false-positive` call, which pipeline.md:423-427 names as the one recording site. That citer is prose, not a heading. The **Follow-up issues** citer moves with it. |
| I05a | integrate.md:112-124 (partial) | 1179 | LAZY-SPLIT | Any check-base-moved.sh verdict is MOVED (known right after the step-2 check, before the rebase). False whenever no worktree's base moved. | Every `MOVED` worktree is then rebased — no prompt, conflicts | med | Move verbatim to a skills/flow/ sibling (rule 2 scans skills/flow/ only). The `stopped` mark block L108-110 and the no-MOVED sentence L124-125 stay. This is the canonical copy of the aside rule (run1:233-238 is the duplicate, R1r1). |
| I05b | integrate.md:127-136 | 832 | LAZY-SPLIT | A clean rebase whose check-base-moved.sh re-run reported an `overlaps:` set (a subset of MOVED). | **Scoped re-verification**: for each path `check-base-moved.sh` reported under `overlaps:`, look | med | finish-contract-run1.md:390-393 names "integrate.md's step 2" as canonical for it; update that pointer if moved. |
| MI1 | integrate.md:138-142 | 460 | MECHANICS | `project-get.sh` enum mode (first non-blank line, trim, strip backticks, byte-for-byte literal match, report and drop). The same parse appears at archive.md:209-213 and in SKILL.md's model block. | Run `project-get.sh <main-checkout> "default landing route"` (exit 1: absent), take | med | Needs code and parity tests. Until then this is the only in-session statement of the parse, because project-configuration.md is not loaded. |
| I14 | integrate.md:155-156 | 187 | LAZY-SPLIT | The landing question is actually asked (no resolved `## default landing route`). | **Report an existing PR before asking.** If a PR exists | low | Small. |
| I06a | integrate.md:183-184 (partial) | 155 | DUPLICATE | finish-contract-run1.md:294-301 | This keeps every planning commit as its own commit on | high | Cut together with I06b (one sentence spans both). |
| I06b | integrate.md:184-186 (partial) | 160 | RATIONALE | Why the rebased merge base is used. The instruction ("never the state file's now-stale pre-rebase value") stays at L181-182. | using the stale value here would also collapse in the | med | None. |
| I06c | integrate.md:186-188 (partial) | 148 | DUPLICATE | finish-contract-run1.md:302-304 | **When the script cannot be located**, do not fall back | high | None. |
| I07 | integrate.md:190-193 (partial) | 277 | DUPLICATE | finish-contract-run1.md:306-317; also integrate.md:212-228 in the same file | All three routes then commit — implementation, then the `<project>/spectre/changes/` | med | None. |
| I09 | integrate.md:249-250 | 172 | DUPLICATE | finish-contract-run1.md:380-385 (same quoted sentence, plus the exit-3 detail) | If there is no remote at all, say exactly that | high | None. |
| I10 | integrate.md:252-253 | 163 | LAZY-SPLIT | Route is pull request and no PR CLI is usable (route known at step 2). | **Human confirmation is a legitimate substitute for an API probe** | med | This operator lands by merge and push, so the text is never used. |
| I11a | integrate.md:260-261 (partial) | 75 | DUPLICATE | jira-integration.md:76, 87-89 | , whichever route was taken — pull request, merge and | med | Keep "after the state write, never before, never blocking" (specific to this site). |
| I11b | integrate.md:262-263 (partial) | 60 | DUPLICATE | jira-integration.md:90-91 | A run that stopped on a failed push does **not** | high | None. |
| I12 | integrate.md:269-275 | 409 | DUPLICATE | finish-contract-run1.md:387-393 | ## No verification gate **Run no tests, no linters, and | high | Also removes the drifted copy: this section says "One exception", run1 says two (drift D1). |
| I13 | integrate.md:310-314 | 292 | LAZY-SPLIT | Route is pull request or manual. | ## After open PR or manual specifically Stop after the | med | This operator's configured route is merge and push. The routing heading moves with it into a route file. |
| I15 | integrate.md:316-327 | 668 | DUPLICATE | Per bullet: run1:129; run1:306 + git-boundaries.md:118-132; run1:10-12 and run2:9-11; run1:387-393; run1:341, 355; integrate.md:245-247; git-boundaries.md:144-150; jira-integration.md:150-162 | ## Guardrails - **Never** ask how the branch should land | med | Bullet L322 ("Never run tests, linters…") contradicts run1's two exceptions (D1), so cutting it also removes a contradiction. The heading has no citers. |
| A18 | archive.md:15-17 | 300 | DUPLICATE | SKILL.md:197-200 (one token per run, reused in every phase file it dispatches into) | **Generate this run's own session token here, before this first | med | "Generate … here" can read as minting a second token on a standalone run 2, where the router already generated one. This is a mild drift. |
| A19a | archive.md:23-23 (partial) | 63 | DUPLICATE | finish-contract-run2.md:9-11 | a PR CLI when usable, otherwise `git merge-base --is-ancestor`. Fetch | low | Tiny. |
| A19b | archive.md:24-25 (partial) | 154 | DUPLICATE | finish-contract-run2.md:12-15 | On the merge-and-push continuation the merge is still local, so | med | None. |
| A02a | archive.md:40-42 (partial) | 208 | DUPLICATE | finish-contract-run2.md:23-24, 31-32 | Exit `0` → `<landing-worktree>` is on `chore/archive-<name>`, cut from a | med | Keep "The four exit codes are **Run 2 …**, step 2." It is a citer. |
| A02b | archive.md:43-45 (partial) | 197 | DUPLICATE | finish-contract-run2.md:17-19, 31-32 | The main checkout itself is never read, checked out, or | med | The copies differ: this one says "never read", run2 says "never checked out, staged or committed" (D17). |
| A01 | archive.md:47-48 | 123 | LAZY-SPLIT | The guard presence check named prepare-archive-branch.sh missing. That is essentially never true on an installed machine. | **When the guard is absent**, perform the same positioning by | high | Also duplicates run2:51-69. |
| A03 | archive.md:55-56 | 157 | DUPLICATE | finish-contract-run2.md:74-78 | **One call per change, parent and each `<name>-fix-N` sibling alike** | low | Carries the literal `spectre archive "<name>-fix-N"` form. |
| MA1 | archive.md:74-86 | 667 | MECHANICS | A `commit-archive.sh <landing> <canonical-wt> <name>` script: copy loop, branch assert, add, scope check, commit. | ```bash bash -c 'for pair in "ledgers/<name>.md ledger.md" "reviews/<name>-panel.md panel.md"; | med | This would also retire A04, A05 and A07. Needs code and tests. |
| A04 | archive.md:88-89 (partial) | 145 | DUPLICATE | finish-contract-run2.md:104-109 | **The copy loop runs before the `add -A`, so the | med | None. |
| A05 | archive.md:89-92 (partial) | 335 | RATIONALE | KAN-816 incident plus a rejected alternative (zsh-only splitting). Move verbatim to SKILL-rationale.md. | The loop is invoked through `bash -c` (KAN-816): its `set | med | The `bash -c` wrapper itself stays in the shell block. |
| A06 | archive.md:93-96 (partial) | 287 | DUPLICATE | finish-contract-run2.md:109-112, which is itself RATIONALE (R2i) | Why the preservation exists — step 5 destroys the worktree | high | Move together with R2i. Contains a **Run 2 — the branch is merged** citer. |
| A07 | archive.md:98-103 (partial) | 502 | DUPLICATE | finish-contract-run2.md:91-102 | **`check-archive-scope.sh`** refuses a `git add -A` that staged more than | med | The shell block (L83) keeps the invoking citation of check-archive-scope.sh. |
| A08 | archive.md:103-105 (partial) | 150 | LAZY-SPLIT | The guard presence check named check-archive-scope.sh missing. | **When absent**, run `git -C <landing-worktree> diff --cached --name-only` by | high | Carries the path bug D10. |
| A09 | archive.md:139-144 (partial) | 459 | DUPLICATE | finish-contract-run2.md:160-163, 177-180 | `COMPLETE:` → report the cleanup as verified, **relay every clause | med | Keep the invocation sentence L138-139, the only skills/flow/ citation of check-cleanup-complete.sh. Keep the mark-outcome note L150-151. |
| A10 | archive.md:174-181 (partial) | 667 | DUPLICATE | finish-contract-run2.md:219-229 | What is specific to *executing* it here: `flow self-review bundle | high | None. |
| A11 | archive.md:183-184 (partial) | 133 | DUPLICATE | SKILL.md:105-107 | — `/flow`'s **Model resolution** (`skills/flow/SKILL.md`) deliberately does not, since no | med | The bold marker "Resolve `SELF_REVIEW_MODEL` here, where it is consumed" must stay verbatim: check-model-resolution-shell.sh anchors on it. |
| MA3 | archive.md:186-200 | 858 | MECHANICS | SELF_REVIEW_MODEL "governs nothing" and is recorded nowhere (D12). Either a `flow settings` subcommand or deletion. | ```bash MAIN_CHECKOUT="${MAIN_CHECKOUT:-$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd -P)}" SELF_REVIEW_MODEL="$(flow | low | Extracted by check-model-resolution-shell.sh. Deleting it drops an exit-2 stop on a broken project.md, so this needs a decision, not a trim. |
| A12 | archive.md:202-205 (partial) | 392 | DUPLICATE | The code block L186-200 plus project-configuration.md:36 | `<project>/.flow/project.md`'s `## self review model` key, when present and a | low | Carries "named as a fallback rather than a resolved value", a reporting instruction. |
| A13 | archive.md:206-207 (partial) | 142 | STALE | The step-9 pass is inline, "no subagent, no dispatch" (archive.md:247-248; run2:242-245). Harness mapping (model-policy.md:105-113) governs dispatches only. | On harness `zcode` the dispatch runs on `glm-5.3-flash` / `high` | high | None. |
| MA2 | archive.md:209-213 | 433 | MECHANICS | The same enum parse as MI1. | Run `project-get.sh <main-checkout> "self review"` (exit 1: absent), take the | med | Needs code and tests. |
| A20 | archive.md:218-227 | 455 | LAZY-SPLIT | The `## self review` key is absent (resolved at L209-216, just before). This operator's key is `defer`. | When the key is absent, the skip prompt fires first: | med | Operator-prompt wording: move it verbatim, never cut it. |
| A21 | archive.md:229-245 | 987 | LAZY-SPLIT | Self review resolves to `defer`. | **On `defer`** — by key or by the prompt's third | low | Always true for this operator, so it saves nothing here. It only helps projects on run or skip. |
| A14 | archive.md:247-286 | 2072 | LAZY-SPLIT | Self review resolves to `run` (key `run`, or Yes at the prompt). Never true for this operator (`## self review: defer`). | **On `run` (or the skip prompt's explicit Yes), this session | med | Move verbatim, together with run2:242-297, to a sibling. It has no stage marks, and land-self-review-report.sh stays cited at L238. Its internal restatements of run2 (L247-260, L274-280) then collapse to one copy. |
| A15 | archive.md:341-355 | 446 | LAZY-SPLIT | check-cleanup-complete.sh returned LEFTOVER or no verdict at step 7. | On a leftover — or on no verdict at all | med | Handoff template: move verbatim. |
| A16 | archive.md:364-368 (partial) | 259 | RATIONALE | Why the override is safe. | ; it is safe here because the records worth keeping | high | None. |
| A17a | archive.md:376-378 | 218 | DUPLICATE | finish-contract-run2.md:9-11 (not merged means not run 2); run2:160-163, 177-180, 189 | - **Never** merge the change branch in run 2; step | med | Bullet 1 inherits drift D6. |
| A17b | archive.md:380-381 | 152 | DUPLICATE | jira-integration.md:152-162; finish-contract-run2.md:214-216 | - **Never** let a Jira call block the archive — | med | Keep L379 (the state file is never moved into the archive, which is unique here) and L382-383. |
| R1b | finish-contract-run1.md:20-24 (partial) | 344 | RATIONALE | Why the bare local name is wrong. The instruction stays at L19-20. | A bare local branch is wrong here because the ancestor | high | None. |
| R1c | finish-contract-run1.md:32-50 | 1509 | LAZY-SPLIT | check-finish-preflight.sh is missing. The signal list is what the hand fallback (R1d) executes. | 1. **`HEAD` against the merge base recorded in the state | med | pipeline.md:431-432 says the finish contract "governs the preflight signals". A REFUSE relay takes its procedure from the guard header (pipeline.md:405-418), not from this list. |
| R1d | finish-contract-run1.md:58-63 | 525 | LAZY-SPLIT | check-finish-preflight.sh is missing. | **When the script is absent** — a harness whose repository | high | None. |
| R1e | finish-contract-run1.md:68-71 (partial) | 334 | RATIONALE | Why the surfacing runs before the preflight. The KAN-546 incident is already in finish-contract-rationale.md. | It is pre-run on purpose: which run this is is | high | None. |
| R1f | finish-contract-run1.md:89-94 (partial) | 414 | LAZY-SPLIT | check-foreign-staged.sh is missing. | **When the script is absent** — a repository that does | high | None. |
| R1g | finish-contract-run1.md:101-102 (partial) | 102 | RATIONALE | KAN-647 incident provenance. | — the shapes KAN-647's reverse-image incidents took, which the staged-only | high | None. |
| R1h | finish-contract-run1.md:115-120 | 478 | LAZY-SPLIT | check-main-checkout-drift.sh is missing. | **When the drift script is absent** — a repository that | high | None. |
| R1i | finish-contract-run1.md:124-127 | 361 | DUPLICATE | pipeline.md:108-122 (always loaded) | **Run 1 itself only starts from a fresh bare `/flow` | med | Its bold citation spans two lines, so check-references never verified it. The target is bold prose, not a heading. |
| R1j | finish-contract-run1.md:135-137 (partial) | 146 | RATIONALE | Why `[canonical-worktree]` is passed. Cross-repo only. | Without it, a satellite worktree's call falls back to resolving | med | None. |
| R1k | finish-contract-run1.md:137-138 (partial) | 102 | DUPLICATE | worktree-resolution.md:20-24 | A resolved set that comes back empty stops the run | med | None. |
| R1l | finish-contract-run1.md:153-156 (partial) | 198 | RATIONALE | Why the guard exists. | This guard exists because a stage mark is not evidence | high | None. |
| R1u1 | finish-contract-run1.md:158-169 | 1128 | LAZY-SPLIT | Same condition as I16. | On `OUTSTANDING` — from either guard — the operator is | med | The course table is canonical here. It cites **1. Check for unfinished work** (integrate.md), which stays a heading. |
| R1m | finish-contract-run1.md:170-174 (partial) | 473 | RATIONALE | Why Stop is the recommended course. | The gate only fires because something really is unfinished, and | high | Keep the bold sentence at L169. |
| R1u2 | finish-contract-run1.md:176-178 | 235 | LAZY-SPLIT | Same condition as I16. | There is no fourth course, and in particular none that | med | None. |
| R1n | finish-contract-run1.md:180-187 | 659 | DUPLICATE | jira-integration.md:152-162, 237-242 (names run 1's option explicitly) | **A filing that fails is one skipped-with-reason line, and the | med | Reachable only on course 3. |
| R1o | finish-contract-run1.md:189-195 | 575 | DUPLICATE | jira-followups.md:34-39, 211-213, 229-241, 420-450 (loaded on course 3) | **Three more outcomes of that course behave the same way**, | med | Contains a **Follow-up issues** citer. Reachable only on course 3. |
| R1p | finish-contract-run1.md:197-199 | 254 | DUPLICATE | finish-contract-run1.md:163 (Continue row) | **What the operator integrated over is recorded where a transcript | med | Resolve drift D2 first. |
| R1q | finish-contract-run1.md:218-220 | 248 | LAZY-SPLIT | check-base-moved.sh is missing. | **When the script is absent** — a harness whose repository | high | None. |
| R1r2a | finish-contract-run1.md:222-232 | 400 | LAZY-SPLIT | Any MOVED verdict. | #### Sync the branch onto the base **Runs after the | med | Citers of **Sync the branch onto the base** (run1) must follow the move: implement.md:210, integrate.md:112, flow-fast/SKILL.md:351-352, and in-file run1:214, 294, 331, 390; also implement.md:977 **Conflict**. |
| R1r1 | finish-contract-run1.md:233-238 | 493 | DUPLICATE | integrate.md:116-124, the executing copy with paths and the `git stash list` recovery | Uncommitted planning artifacts are set aside before this rebase and | med | If I05a moves, the canonical copy moves with it. |
| R1r2b | finish-contract-run1.md:240-248 | 651 | LAZY-SPLIT | Any MOVED verdict. | `origin/$BASE` is current: `resolve-base-branch.sh` fetched when the caller resolved the | med | Same citers as R1r2a. |
| R1r2c | finish-contract-run1.md:249-265 | 1543 | LAZY-SPLIT | The rebase exited non-zero (conflict). Rarer than MOVED. | - **Conflict** (non-zero exit): **resolve it in place, automatically.** For | med | /flow-fast applies this bullet "as written" (flow-fast/SKILL.md:350-353). |
| R1r2d | finish-contract-run1.md:266-267 | 147 | LAZY-SPLIT | Any MOVED verdict. | - **The handoff names every file that conflicted and what | med | None. |
| R1s | finish-contract-run1.md:269-280 (partial) | 534 | DUPLICATE | integrate.md:138-153 (same prompt wording, byte-identical) | Read `<project>/.flow/project.md`'s `## default landing route` (canonical in **Project configuration**, | med | Keep "Only then decide, **before any git action**, how the branch should land." and "The answer is never remembered between runs." Both are unique here. |
| R1t | finish-contract-run1.md:284-288 (partial) | 405 | LAZY-SPLIT | The resolved set has more than one worktree (known after the preflight). | docs/links.md in the `spectre` repository is canonical for that section's | low | The ordering instruction at L282-284 stays. |
| R1w | finish-contract-run1.md:319-326 | 803 | DUPLICATE | git-boundaries.md:87-92, 118-132 (loaded at integrate step 3) and commit-split.sh | Those two planning paths are cleared from the index before | med | Already stale: the chain clears and excludes only spectre/changes/, not "two planning paths" (D3). |
| R1pr | finish-contract-run1.md:330-330 | 248 | LAZY-SPLIT | Route is pull request. | \| **Open a pull request** \| push `--force-with-lease` (**Branch backup**, | low | This is a table row, so a route split breaks up the table. |
| R1man | finish-contract-run1.md:332-332 | 103 | LAZY-SPLIT | Route is manual. | \| **Handle it manually** \| push the branch `--force-with-lease` only; | low | Table row. |
| R1x | finish-contract-run1.md:355-362 (partial) | 625 | RATIONALE | Why never `HEAD@{upstream}`. | bare `/flow` runs inside the apply worktree, where `HEAD` *is* | high | Keep the bold instruction at L355. |
| R1y | finish-contract-run1.md:364-365 (partial) | 105 | RATIONALE | Reason clause. | An unresolvable base is an honest unknown; a guessed one | high | Keep "If no base branch resolves, **stop and ask**." |
| R1z | finish-contract-run1.md:367-378 | 1168 | LAZY-SPLIT | resolve-base-branch.sh is missing. | **When the script is absent** — a harness whose repository | high | finish-contract-rationale.md explains why the character rule is written in full. That reason still holds in a sibling contract, which ships with the contracts, not the guards. |
| R1aa | finish-contract-run1.md:388-390 (partial) | 220 | RATIONALE | Why there is no gate. | Correctness was established during `/flow`'s implement phase — TDD per | med | Keep L387 and the exceptions at L390-393. |
| R1ab | finish-contract-run1.md:397-402 (partial) | 462 | DUPLICATE | worktree-resolution.md:15-31 | This is bare `/flow`'s own application of the rule stated | med | Keep the first sentence (L397). |
| MR1 | finish-contract-run1.md:404-410 | 289 | MECHANICS | A `flow worktrees <name>` / resolve-worktrees script using the same parse as stats/internal/guard/cleanupcomplete.go. | The set of worktrees is the **keys of the state | low | This would also retire R1ae. |
| R1ac | finish-contract-run1.md:412-414 (partial) | 261 | RATIONALE | A layout note. It is also stale against git-boundaries.md:28-29, since every project's worktrees are `<project>/.worktrees/<name>`. | Worktree layout differs per repository — this repo keeps worktrees | med | Contains a citer of **2. Isolate the workspace** (implement.md). The heading exists, so the moved line stays valid. |
| R1ad | finish-contract-run1.md:417-422 (partial) | 447 | DUPLICATE | worktree-resolution.md:20-28 | Per **Resolving a change's worktrees** (`skills/flow-contracts/worktree-resolution.md`), that is a state | med | Keep the definition sentence L416-417. |
| R1ae | finish-contract-run1.md:424-432 (partial) | 743 | RATIONALE | A worked example plus an editor note ("must not disagree, or the wrong one gets copied next"). | `worktree list --porcelain` emits it raw, so a field reference | med | Keep "The path is taken with `substr`, never `$2`." |
| R2b | finish-contract-run2.md:10-11 (partial) | 114 | RATIONALE | Editor-facing reason. | That fallback must stay reachable on its own — it | med | None. |
| R2c | finish-contract-run2.md:44-48 (partial) | 379 | LAZY-SPLIT | classify-untracked.sh reported an asset (rare). | an asset stays untracked, so the guard's refusal stands until | low | run1:331 relies on the same asset rule. |
| R2d | finish-contract-run2.md:48-49 (partial) | 99 | LAZY-SPLIT | classify-untracked.sh is missing. | When the script is absent, classify by hand to the | high | None. |
| R2e | finish-contract-run2.md:51-69 | 1694 | LAZY-SPLIT | prepare-archive-branch.sh is missing. | **When the script is absent** — a harness whose repository | high | finish-contract-rationale.md step-2 note ("the guard being absent takes its header with it") still holds in a sibling contract. |
| R2f | finish-contract-run2.md:71-73 (partial) | 115 | RATIONALE | Reason clause. | , because `spectre archive` adds none: a prefix re-added here | high | None. |
| R2g | finish-contract-run2.md:79-81 (partial) | 131 | RATIONALE | Reason clause. | : a change edits that tree directly on its own | med | None. |
| R2h | finish-contract-run2.md:94-97 (partial) | 286 | RATIONALE | Why the scope check exists. | `add -A` stages the whole landing worktree, not only the | high | None. |
| R2i | finish-contract-run2.md:109-112 (partial) | 262 | RATIONALE | Why the renders are preserved. | The store's rows are the terminal record, but rows that | med | Move together with A06. |
| R2j | finish-contract-run2.md:120-129 (partial) | 857 | DUPLICATE | finish-contract-run2.md:379-384 (check 3's code and comment) and :495 | A multi-repo change has one `origin` and one default branch | med | Keep L119-120 and the "Anything but exit 0 …" sentence. |
| R2k | finish-contract-run2.md:135-141 (partial) | 483 | DUPLICATE | archive.md:120-122 (`<id>` from `flow workspace-id <name>`, never handed to this run) | **Run 2 is not handed that id and does not | med | None. |
| R2l | finish-contract-run2.md:143-149 | 535 | DUPLICATE | archive.md:122-125 | **A project declaring no `## workspace isolation` section, or no | med | None. |
| R2n | finish-contract-run2.md:169-171 (partial) | 154 | RATIONALE | Reason. | A run that reported only "cleanup verified" would have told | high | None. |
| R2o | finish-contract-run2.md:182-187 (partial) | 456 | RATIONALE | Reason. Relevant only when `## stop` declares a command, which the agents repo's does not. | Once check 5 in **Worktree cleanup** below has stopped the | med | Keep the bold first sentence. |
| R2p | finish-contract-run2.md:190-194 (partial) | 489 | RATIONALE | Reason. | `FINISHED` is terminal: bare `/flow` stops at it and `/flow-status` | high | None. |
| R2q | finish-contract-run2.md:202-204 | 305 | LAZY-SPLIT | check-cleanup-complete.sh is missing. | **When the script is absent** — a repository that does | high | None. |
| R2r | finish-contract-run2.md:206-207 | 182 | MISPLACED | /flow-fast never loads this file (run2:5). flow-fast/SKILL.md:377-417 does its own cleanup with no verification stage. | **A `/flow-fast` run skips this step entirely** — cleanup itself | high | None. |
| R2s | finish-contract-run2.md:212-214 (partial) | 227 | MISPLACED | Same. The canonical text is flow-fast/SKILL.md:256. | **A `/flow-fast` run runs no reasoning pass and no prompt | high | The following "a skip" would need a capital letter, a one-character edit: flag it to the editor. |
| R2aq | finish-contract-run2.md:236-240 (partial) | 336 | MISPLACED | Text for the deferred pass in /flow-self-review. | The pass then runs in `/flow-self-review <name>` (`skills/flow-self-review/SKILL.md`), canonical for | low | flow-self-review/SKILL.md:51-52 cites it, so moving it needs that citer changed. |
| R2t | finish-contract-run2.md:242-297 | 4017 | LAZY-SPLIT | Self review resolves to `run`. Never true for this operator. | **On `run`, this same step-9 session runs the reasoning pass | med | check-self-review-report.sh reads the angle table from this file (ANGLE_CONTRACT default), so its default path must change. Citers: flow-self-review/SKILL.md:11, 48-49, 59; jira-integration.md:233; archive.md:254. Bytes exclude R2u. |
| R2u | finish-contract-run2.md:276-278 (partial) | 215 | RATIONALE | Reason. | A prompt's option text cannot carry that explanation, so the | high | None. |
| R2w1 | finish-contract-run2.md:304-304 | 739 | LAZY-SPLIT | archive.md was reached by a standalone invocation, not the merge-and-push chain. This is known at load time. | \| a standalone invocation \| in `<landing-worktree>`: push `chore/archive-<name>`; open | med | Table row. |
| R2w2 | finish-contract-run2.md:317-321 (partial) | 378 | LAZY-SPLIT | Standalone reach. | On the standalone row, a PR opened but not yet | med | None. |
| R2w3 | finish-contract-run2.md:324-329 | 557 | LAZY-SPLIT | Standalone reach. The recognition rule becomes the load condition itself. | **This split reads no persisted field.** The merge-and-push row is | med | None. |
| R2x | finish-contract-run2.md:343-347 (partial) | 444 | RATIONALE | Why the refresh exists. | No step above checks out, stages or commits the main | med | None. |
| R2y | finish-contract-run2.md:353-356 (partial) | 325 | RATIONALE | Reason. Keep the bold ordering sentence. | Per **Jira integration** (`skills/flow-contracts/jira-integration.md`)'s own timing — the issue moves | med | Contains a **Jira integration** citer. |
| MR2 | finish-contract-run2.md:358-541 | 3992 | MECHANICS | A `remove-change-worktrees.sh <repo> <name>` guard would run checks 1-3, check 5 with its 60-second bound, check 6 through check-worktree-processes.sh, the check-4 bucket classification by path (L447-458), the remove/prune/`branch -d`, and the remote delete that tells failures apart by git's message. It would print verdict lines with an exit contract. Counted: the code at L370-417, 477-481 and 506-523 (less the comment bytes already counted as RATIONALE) plus the bucket list at L447-458. | ```bash # 1. no uncommitted tracked changes — must be empty | med | Needs code and parity tests. The run keeps only the disclosure relay and the ask. It would also retire most of archive.md:357-372. |
| R2z | finish-contract-run2.md:388-393 | 520 | RATIONALE | Comment lines inside check 3's code (the squash-merge reasoning). | # Step 1 already proved the branch is an ancestor | med | The code is unchanged. |
| R2aa | finish-contract-run2.md:403-406 (partial) | 322 | RATIONALE | Also restated at L483-489. | `--exclude-standard` in check 2 hides # everything matched by .gitignore, | med | None. |
| R2ab | finish-contract-run2.md:412-415 (partial) | 218 | RATIONALE | Reason. | : a project that declares no stop command can still | med | None. |
| R2ac | finish-contract-run2.md:420-423 (partial) | 328 | RATIONALE | Reason. The instruction to cd out stays. | `check-worktree-processes.sh`'s own header treats a process whose working directory is | med | None. |
| R2ad | finish-contract-run2.md:428-429 (partial) | 133 | RATIONALE | Restates the bold sentence plus a reason. | Neither is a pass: an inability that proceeded to removal | med | None. |
| R2ae | finish-contract-run2.md:437-441 (partial) | 276 | RATIONALE | Reason. | That disclosure is safe to confirm because the operator can | high | Keep "The remedy is to clear the process and re-run…". |
| R2af | finish-contract-run2.md:460-463 (partial) | 277 | RATIONALE | Reason. | — every entry in it is reproduced identically by the | high | None. |
| R2ag | finish-contract-run2.md:468-473 | 562 | STALE | flow-fast/SKILL.md has no "Guardrails" heading (only a bold paragraph at L29). Its cleanup (L403-405) runs `git worktree remove` without --force and `branch -D`, with no disclosure. The citation spans L468-469, so the line-based check-references guard misses it. | **`/flow-fast` overrides the ask a level further, and only the | high | Also MISPLACED: /flow-fast never loads run2. |
| R2ah | finish-contract-run2.md:483-489 | 589 | RATIONALE | A caveat. It says "the operator confirms", which contradicts archive.md's override (D8). | - **`--force` destroys every ignored file in the worktree, and | low | Never compress a caveat; resolve D8 first. |
| R2ai | finish-contract-run2.md:490-492 | 200 | RATIONALE | A known limit that drives no action. | - **Neither check sees a file whose `assume-unchanged` bit is | med | None. |
| R2aj | finish-contract-run2.md:501-502 (partial) | 122 | RATIONALE | Reason. | Writing `worktrees: {}` regardless would drop it from the only | med | None. |
| R2ak | finish-contract-run2.md:507-511 | 421 | RATIONALE | Code comment with "Measured … 2026-08-31" provenance. | # `push --delete` exits non-zero BOTH when the branch was | med | The branch logic in the code is self-evident. |
| R2al | finish-contract-run2.md:525-528 (partial) | 355 | RATIONALE | Reason. | Run 2 is reached only by proving the branch is | high | None. |
| R2am | finish-contract-run2.md:531-533 (partial) | 177 | RATIONALE | Reason. | A bare `\|\| true` would make an expired credential indistinguishable | high | None. |
| R2an | finish-contract-run2.md:534-535 (partial) | 128 | RATIONALE | Reason. | Gating it would leave the remote branch behind whenever anything | high | None. |
| R2ao | finish-contract-run2.md:539-541 (partial) | 181 | DUPLICATE | finish-contract-run2.md:412-415 | — check 6 among them, which is what makes an | med | None. |
| J36 | jira-followups.md:17-17 (partial) | 58 | STALE | Frames integrate as one of several filing sites. The review panel now uses KNOWN-BUGS.md (L6-8; KNOWN-BUGS.md:41, F13). | **This naming governs every site that files a follow-up.** | low | Known and already tracked. |
| J01 | jira-followups.md:42-46 (partial) | 404 | RATIONALE | Why the project clause is needed. | `searchJiraIssuesUsingJql` searches whatever the session's Atlassian connection can reach, which | med | The following "therefore" loses its antecedent. That is a cut, not a rewrite. |
| J02 | jira-followups.md:53-58 (partial) | 527 | RATIONALE | Threat reasoning. | `<project>/.flow/project.md` is tracked in the repository and editable in any | med | The following "So each key…" keeps the rule. |
| J03a | jira-followups.md:88-90 (partial) | 210 | RATIONALE | A worked reason for the parenthesis rule, which stays. | : JQL binds `AND` tighter than `OR`, so a query | med | None. |
| J03b | jira-followups.md:90-96 (partial) | 539 | RATIONALE | Tells the implementer not to add escaping; half instruction, half reason. | **No escaping step is specified, because the shape required above | low | None. |
| J04 | jira-followups.md:98-102 | 453 | RATIONALE | Reason. | **Searching where the follow-up would be filed is what keeps | high | None. |
| J05 | jira-followups.md:108-111 (partial) | 231 | RATIONALE | Reason. | This has to be said because JQL's `~` operator is | med | None. |
| JLZ | jira-followups.md:114-466 | 10546 | LAZY-SPLIT | The join search returned a candidate. Covers L114-227 (confirm, title sanitising) and L266-466 (append, echo exception, idempotency, guard, outcomes); bytes are net of the RATIONALE rows inside those ranges. The create path (L1-112, L229-264, L467-468) stays. | **Confirm before joining.** The rule and the shape are the | med | The file is loaded rarely anyway (see the jira-followups section). Citers of **Follow-up issues** that mean the join (jira-integration.md:145, 221; run1:190; integrate.md:87) would need the new path. |
| J06 | jira-followups.md:134-144 | 990 | RATIONALE | Why the count is shown. | **That count is shown because the guard's evidence is forgeable, | high | Keep L144-146 (an unreadable candidate is not offered). |
| J07 | jira-followups.md:149-151 (partial) | 210 | RATIONALE | Reason. | **Change naming** (`jira-integration.md`) already caps a summary-derived slug because it | med | Contains a **Change naming** citer. |
| J08 | jira-followups.md:154-167 (partial) | 1206 | RATIONALE | Explains the Unicode categories. The fold rule stays. | `Cc` is the control class — newlines, carriage returns, tabs, | med | None. |
| J09 | jira-followups.md:170-173 (partial) | 312 | RATIONALE | Reason. | Stated as a category because "whitespace" left undefined is read | med | None. |
| J10 | jira-followups.md:174-175 (partial) | 51 | RATIONALE | Pointer to J13. | Why this is not cosmetic is the paragraph below. | med | Move together with J13. |
| J11 | jira-followups.md:178-189 | 1186 | RATIONALE | Why the order holds. "In this order" stays at L152. | **The order is a requirement, and what each dependency actually | high | None. |
| J12 | jira-followups.md:192-195 (partial) | 309 | RATIONALE | Reason. | so no run of attacker-chosen text can be read as | med | None. |
| J13 | jira-followups.md:197-209 | 1204 | RATIONALE | Why step 3 exists. | **The fenced block is the isolation, so a title able | med | None. |
| J14 | jira-followups.md:215-223 | 777 | RATIONALE | Reason. | **Why the confirmation exists.** The search selects a **write target** | high | None. |
| J15 | jira-followups.md:225-227 | 265 | RATIONALE | Reason. | **The cost of the confirmation is bounded, and it is | high | None. |
| J16a | jira-followups.md:231-234 (partial) | 242 | RATIONALE | Reason. | That reading files a new follow-up on every transient failure, | med | None. |
| J16b | jira-followups.md:236-241 (partial) | 457 | RATIONALE | Reason. Also carries a stale second filing site ("the round's deferred findings", KNOWN-BUGS F13). | What that costs is one tracker entry, and the cost | med | None. |
| J17 | jira-followups.md:244-246 (partial) | 176 | RATIONALE | Reason. | The project clause is the only narrowing there is, and | med | None. |
| J18 | jira-followups.md:249-250 (partial) | 117 | RATIONALE | Editor note. | The set is cited from that one statement rather than | high | None. |
| J19 | jira-followups.md:255-258 (partial) | 342 | RATIONALE | Reason. | **This does not reopen the unrecognised-status rule.** That rule governs | med | Keep "An urgent To Do is joined exactly where it sits…". |
| J20a | jira-followups.md:281-286 (partial) | 471 | RATIONALE | Reason. | The echo exists as a recovery path for text this | med | None. |
| J20b | jira-followups.md:288-290 (partial) | 235 | RATIONALE | Reason. | The recovery path for that text is the issue's own | med | None. |
| J21 | jira-followups.md:300-302 (partial) | 240 | RATIONALE | Reason. | Without that check a re-entered run finds its own prior | med | None. |
| J22 | jira-followups.md:315-321 (partial) | 510 | RATIONALE | Reason. | The case this rule exists for is the run that | med | Keep the last sentence (the retry reports partially joined again). |
| J23 | jira-followups.md:325-342 | 1616 | RATIONALE | Post-merge window. Nothing in it is actionable at integrate. | **That window closes at the merge, and the `⚠` does | med | None. |
| J24 | jira-followups.md:344-348 (partial) | 372 | RATIONALE | Reason. | The section heading carries the date it was written, and | med | None. |
| J25 | jira-followups.md:353-362 (partial) | 757 | RATIONALE | Rejected alternatives with reasons. Move verbatim. | The two obvious readings fail in opposite directions: - **Comparing | med | None. |
| J26 | jira-followups.md:367-369 (partial) | 150 | RATIONALE | Reason for "case included". | The items are this pipeline's own output rather than operator | low | None. |
| J27 | jira-followups.md:373-378 (partial) | 512 | RATIONALE | Reason. | Nothing distinguishes a section a previous run of this change | med | None. |
| J28 | jira-followups.md:394-404 | 843 | RATIONALE | Residual-risk statement. Also carries the stale F13 filing site. | **What remains.** An operator who confirms a candidate whose forged | high | None. |
| J30 | jira-followups.md:415-418 (partial) | 217 | RATIONALE | Reason. | , because every description write this contract makes is a | med | None. |
| J31a | jira-followups.md:428-431 (partial) | 330 | RATIONALE | Reason. | The second is the ordinary shape of a retry — | med | None. |
| J31b | jira-followups.md:432-434 (partial) | 137 | RATIONALE | Reason. | Without that, an issue holding a second change's outstanding work | med | None. |
| J34 | jira-followups.md:452-455 (partial) | 252 | RATIONALE | Reason. | Those two are independent — the append has the three | med | None. |
| J35 | jira-followups.md:461-464 (partial) | 274 | RATIONALE | Reason. | The label union exists to stop an issue holding a | med | None. |

## Top 5

1. **Hand-fallback LAZY-SPLIT** (R1c, R1d, R1f, R1h, R1q, R1z, R2d, R2e, R2q, A01, A08).
   - Size: 5,204 B at high confidence, plus 1,509 B med (R1c).
   - Where it lands: 4,342 B off every run-1 session and 2,371 B off every run-2 session.
   - Load condition: the guard presence check printed `GUARDS MISSING` naming that guard. That is essentially never on an installed machine.
   - Mechanics: move them verbatim into one `skills/flow-contracts/finish-hand-fallbacks.md` behind a "Load … only when the guard presence check named a missing guard" line in run1 and run2.
   - No heading moves, so no check-references citers are affected.
2. **Self-review `run`-branch LAZY-SPLIT** (A14 2,072 + R2t 4,017 + A20 455).
   - Size: 6,544 B off every run-2 session for this operator, whose key is `defer`.
   - Mechanics: move the rules from run2 and the literal prompt from archive.md into one shared sibling. `/flow-self-review` loads that sibling too, so its own copy of the angle table and prompt shape can go.
   - Needs `check-self-review-report.sh`'s `ANGLE_CONTRACT` default to point at the new file.
   - Needs five citer edits (see Dispatch questions, part 5).
3. **RATIONALE moves.**
   - Size: 10,773 B across the four finish files (4,912 B high).
   - Where it lands: 3,711 B off every run-1 session and 7,062 B off every run-2 session.
   - About 50 verbatim moves. Largest: R1ae 743, R1x 625, R2z 520, R2p 489, R1m 473, R2o 456, R2x 444.
4. **DUPLICATE removal, keeping one copy per pair** (see the pairing tables under Dispatch questions).
   - Size: 12,932 B (1,898 high).
   - Where it lands: 7,454 B off integrate.md plus run1, and 5,478 B off archive.md plus run2.
   - Cutting I12 and the L322 bullet of I15 also deletes the drifted copies behind D1.
   - Cutting R1w removes the stale "two planning paths" statement (D3).
5. **Unfinished-work gate LAZY-SPLIT** (I16, I17, R1u1, R1u2).
   - Size: 2,923 B off every run-1 session in which no worktree reports `OUTSTANDING`/`VISUAL-VERIFY-MISSING`, which is the usual case. Another 2,146 B of gate text sits in DUPLICATE/RATIONALE rows (I04, R1m, R1n, R1o, R1p).

Runner-up: the **base-moved LAZY-SPLIT** (I05a, I05b, R1r2a–d).
- It saves 4,752 B when no worktree's base moved, and the conflict-only bullet saves 1,543 B whenever the rebase is clean.
- It is ranked sixth because how often a base moves is unknown, and seven citers have to follow the section.

## Drift / bugs found

All of the following are co-loaded in the FINISH session unless noted.

**D1. The no-verification-gate exceptions disagree.**
- integrate.md:273-275 says "One exception" (the scoped re-verification).
- The integrate.md:322 guardrail says "Never run tests, linters, or a coverage check."
- run1:390-393 names two exceptions: a rebase that needed conflict resolution runs the project's whole `## lint`/`## test`.
- integrate.md:112-115 itself cites "the after-resolution lint and test run".
- run1 is right. The integrate copies are the drifted ones, and I12 plus the I15 bullet remove them.

**D2. The planning commit message is both a fixed literal and a carrier of the outstanding list.**
- integrate.md:215-221 passes `"chore(spectre): plan"` to `commit-split.sh` and calls it "a **fixed literal**". git-boundaries.md:137 says the same.
- integrate.md:227-228, run1:163 and run1:197-199 require "the second commit's message lists anything the operator chose to integrate over".
- A run that follows the literal drops the outstanding list from the one durable record run1 says must carry it.
- `commit-split.sh` takes the message as an argument, so this is a documentation contradiction, not a script limit.

**D3. run1:319-323 describes two excluded paths; the chain excludes one.**
- run1 says "Those two planning paths are cleared from the index before the first `add` and excluded from it by pathspec".
- `commit-split.sh:62-74` and git-boundaries.md:124-126 clear and exclude only `spectre/changes/`. `docs/superpowers/`, the other path `aside-planning-artifacts.sh` sets aside, is not excluded.
- git-boundaries.md:137-138 ("stages the same two trees") is similarly stale.

**D4. git-boundaries.md:24 contradicts run2 step 10.**
- git-boundaries.md:24 says run 2 "**Pushes** `chore/archive-<name>` once … and opens its pull request — never pushes `<base>`".
- The merge-and-push row of run2:303 pushes `<base>` and does not push the archive branch.
- The standalone row of run2:304 merges its PR immediately.

**D5. artifacts-registry.md:29 and :74 contradict run2 step 10.**
- The registry says nothing removes the archive branch "on `origin`" and "the pull request outlives it".
- run2:304 runs `gh pr merge --merge --delete-branch`, which deletes the remote branch.
- On merge-and-push (run2:303) the archive branch is never pushed at all.

**D6. run2:83-84 contradicts run2:303.**
- run2:83-84 says "Run 2 never merges anything into the base branch".
- run2:303 runs `git merge --ff-only chore/archive-<name>` on `<base>` and pushes it.
- The guardrail at archive.md:376 inherits the same wording problem.

**D7. The comment on check 2 contradicts a bullet two screens later.**
- run2:375-376 comments that check 2 "is the check that makes `--force` safe".
- run2:487-489 says "Claiming the checks make `--force` safe would be false".

**D8. Two co-loaded triggers for the check-4 ask.**
- run2:444-466 asks for confirmation on any non-empty unclassified bucket, and run2:487 says "the operator confirms".
- archive.md:362-372, the only consumer of run2, never asks except for "genuinely irreplaceable and *unpreserved*" entries.
- The base ask therefore applies to no command.
- This needs a decision; a trim cannot settle it.

**D9. run2:468-473 is stale.**
- It cites "Its own **Guardrails** (`skills/flow-fast/SKILL.md`)". No such heading exists; flow-fast has only a bold paragraph at L29.
- It describes a `--force` disclosure override that flow-fast no longer has: flow-fast/SKILL.md:403-405 runs `worktree remove` without `--force` and `branch -D`.
- check-references passes because the citation spans lines 468-469 (see D23).

**D10. Wrong root for the archive scope path.**
- archive.md:105 and run2:92 use `<agents repo>/spectre/changes/`.
- The scope is the landing worktree's `spectre/changes/`, a `<project>` path; the guard call at archive.md:83 correctly passes `"spectre/changes/"`.
- In any consuming project, the hand fallback would refuse paths by the wrong prefix. This looks like a citation-roots pass that over-applied.

**D11. Step 6 claims to be a no-op, which conflicts with the registry and the cleanup guard.**
- archive.md:126-129 says step 6 is "unconditionally a no-op skip".
- run2:150-152 and artifacts-registry.md:52-56 say run 2 deletes the proposal artifact source "whether or not it exists", since legacy changes may hold one.
- `cleanupcomplete.go:293-298` reports a surviving `<state-dir>/<name>-proposal-artifact.html` as a leftover.
- So a legacy change would stop at step 7 with a leftover that no instruction removes.

**D12. `SELF_REVIEW_MODEL` governs nothing and is recorded nowhere.**
- archive.md:249-252 and run2:243-245 say it "governs nothing"; no report, handoff or store record carries it.
- archive.md:183-207 still resolves it: 1,599 B, two subprocesses, and an exit-2 stop when `project-get.sh` fails.
- archive.md:249 says it resolves "purely as a recorded value".
- Needs a decision: either record it, or drop it and update `check-model-resolution-shell.sh`, project-configuration.md:36, flow-settings/SKILL.md:12-14 and 85, and model-policy.md.

**D13. archive.md:206-207 describes a dispatch that no longer exists.** It says that on zcode "the dispatch runs on `glm-5.3-flash`", but the pass is inline, with no dispatch.

**D14. jira-followups.md:27-32 is not in context when it applies.**
- The rule ("Every filing ask explains before it asks … what breaks … what the fix would be") governs the unfinished-work gate's prompt, because that prompt offers the filing course.
- The file is loaded only after that course is chosen.
- integrate.md:61-64 relays only "the breakdown", so the rule is not in force at the gate.

**D15. The "only command that loads this file" claims are inaccurate.**
- run1:6, run2:5 and pipeline.md:434 say only bare `/flow` loads the finish contracts.
- flow-fast/SKILL.md:350-353 applies run1's **Conflict** bullet "as written".
- flow-self-review/SKILL.md:11, 48-49, 52 and 59 read run2 step 9.
- flow-status/SKILL.md:99 and 118 read run1.

**D16. A stale second loading site for jira-followups.md.**
- skills/flow-contracts/SKILL.md:29 and rules/flow-manual-review.mdc:58 (outside its core block) list "the review panel's deferred-findings close" as a loading site.
- Deferred findings go to KNOWN-BUGS.md (known-bugs.md:53-58; review-panel.md:1392).
- Following either index could load 35.5 KB into an IMPLEMENTATION session for nothing.
- Related: jira-followups.md:17 and :237-240, and :401-402, still frame multiple filing sites. This is KNOWN-BUGS.md:41, F13.

**D17. archive.md:43-45 and run2 describe the main checkout differently.**
- archive.md:43-45 says the main checkout is "never read" by step 2.
- run2:17-19 and 31-32 say "never checked out, staged or committed / untouched".
- The hand fallback at run2:52-53 runs `git -C <main-checkout> worktree add`.

**D18. archive.md:15-17 can read as a second token.** "Generate this run's own session token here" could mean minting a second token on a standalone run 2. The router already generates one per run (SKILL.md:197-200).

**D19. archive.md:359-360 points at checks that are not there.** It says "run **every** check below", but the checks live in run2, not below.

**D20. run1:412-414 is out of date on worktree location.**
- It says worktree layout "is not where every repository … keeps them".
- Since kickoff creates them, every project's worktrees are `<project>/.worktrees/<name>` (git-boundaries.md:28-29).

**D21. SKILL-rationale.md:67-72 still records `rebase-is-a-confirmed-choice`.**
- It says the rebase runs "only after the operator picks this option".
- integrate.md:112 rebases with no prompt.
- The rationale file is not loaded; this is editor-facing staleness only.

**D22. A stale DECLARED_RULE6 entry.**
- `stats/internal/guard/guardsymlinks.go:65` declares `check-visual-verify-dispatched.sh` as invoked only via run1, "which rule 2's scope deliberately does not scan".
- integrate.md:54 invokes it inside skills/flow/.
- The declaration is redundant and harmless.

**D23. check-references is line-based and misses split citations.**
- A bold heading and its path split across lines are never verified.
- Hits in this set: run2:468-469 (broken, D9), run1:125-126 (target is bold prose, not a heading), integrate.md:16-17, flow-fast/SKILL.md:351-352.

**D24. archive.md:157-160 annotates carried fields as `(null)`.**
- It says to carry "`planningEffort` (`null`), `models.default` (`null`)".
- A legacy record may hold non-null values (state-file.md:167), and "(null)" can be read as "write null".
- Low severity.

## Not slimmable

These blocks must stay. Sub-spans listed as candidates above are the only exceptions; everything else in each range is kept.

- **run1:14-30, 52-56.** The preflight exit contract and verdict table, and the REFUSE and every-worktree rule. flow-status/SKILL.md:118 cites the combine rule.
- **run1:96-113, 129-147, 201-216, 341-353.** The drift surfacing and its prompt, the unfinished-work, base-moved and base-resolution script contracts, and their exit codes.
- **run1:290-311, 328-337, 380-393.** Reshape, the two commits, the route table (the merge-and-push row is this operator's route), no remote, and the no-verification gate. These are the canonical copies once the integrate duplicates go.
- **run2:16-35.** Step 2 and the `prepare-archive-branch.sh` exit contract.
- **run2:82-108.** The branch assertion, the scope check and the preservation instruction.
- **run2:153-180, 196-200.** The cleanup verdicts, the no-verdict rule and re-entrancy.
- **run2:214-236, 298-303, 306-316, 330-342, 348-351.** Step 9's skippability, key, bundle and defer (this operator's path), and steps 10 and 11 for merge and push.
- **run2:358-481, 493-530.** The cleanup checks, the check-6 gate, the check-4 buckets (archive.md's override depends on them), removal, the remote delete and its bullets. They stay until the MECHANICS work lands.
- **archive.md: every `flow stage` fence.** `check-stage-mark-calls.sh` scans archive.md by basename.
- **archive.md:74-86.** The step-4 shell.
- **archive.md:183-200.** The `SELF_REVIEW_MODEL` block, which `check-model-resolution-shell.sh` extracts. It stays until D12 is decided.
- **archive.md:218-227 and 262-272.** Prompt wording: they may move verbatim but never be cut.
- **archive.md:325-339.** The `Finished` handoff template. The phase file carries the producing copy; handoff-blocks.md is loaded by `/flow-status` only. The leftover template at 341-355 may move verbatim (A15).
- **archive.md:357-372.** The override, until D8 is decided.
- **archive.md:379 and 382-383.** Guardrails that exist only here.
- **integrate.md:29-44, 91-110, 158-168, 207-218, 230-247, 255-258, 265-267.** Routing, stage marks, the route pointer, the failure rule, the state write and the Jira load.
- **integrate.md:67-70 and 148-151.** Prompt wording.
- **integrate.md:277-308.** The handoff template and the chain directive.
- **Every guard-invoking sentence cited only in integrate.md or archive.md.** These keep `check-guard-symlinks.sh` rules 2 and 6 green:
  - only in integrate.md: `check-foreign-staged`, `check-main-checkout-drift`, `check-visual-verify-dispatched`, `reshape-branch`
  - only in archive.md: `classify-untracked`, `prepare-archive-branch`, `check-archive-scope`, `check-cleanup-complete`, `land-self-review-report`, `refresh-main-checkout`
- **jira-followups.md:41-96, 104-112, 120-132, 152-176, 211-213, 229-235, 409-450.** The instructions of the security rules: JQL project scoping, whole-key validation, tokenization and its worked example, clause assembly, the exact title match, the join confirmation prompt, the four sanitising steps, the explicit-Yes rule, the failed-search rule, and the append, outcome and write tables. Only their RATIONALE clauses move.

## Dispatch questions

### 1. How much integrate.md/archive.md and the run contracts restate each other

**integrate.md ↔ run1.**
- integrate.md restates 2,629 B of run1, plus 135 B of jira-integration.md: 16% of the file.
- run1 restates 1,027 B of integrate.md (R1r1, R1s), 254 B of itself (R1p), and 3,409 B of other in-session contracts:
  - pipeline.md (R1i)
  - worktree-resolution.md (R1k, R1ab, R1ad)
  - jira-integration.md (R1n)
  - jira-followups.md (R1o)
  - git-boundaries.md (R1w)

| # | rule | integrate.md | run1 | keep | cut | drift |
|---|---|---|---|---|---|---|
| 1 | Foreign-staged and drift surfacing | 16-22 | 65-120 | run1 | I02a+I02b 335 (keep the clause that runs both guards) | none |
| 2 | Canonical-worktree definition | 49-52 | 134-137 | run1 | I03 280 | none |
| 3 | Stop/Continue outcomes | 75-77 | 162-163 | run1 | I04 185 | none |
| 4 | Planning-artifact aside around the rebase | 116-124 | 233-238 | integrate (paths, `git stash list`, restore refusal) | R1r1 493 | run1 is only less specific |
| 5 | Landing default and prompt | 138-153 | 269-280 | integrate (executing parse, handoff line) | R1s 534 (keep "before any git action" and "never remembered between runs") | prompt byte-identical |
| 6 | Reshape keeps planning commits; no `reset --soft` | 183-188 | 290-304 | run1 | I06a+I06c 303 (+I06b RATIONALE 160) | none |
| 7 | Two commits; session records uncommitted | 190-193 | 306-317 | run1 | I07 277 | none |
| 8 | Outstanding list in the planning commit | 227-228 | 163, 197-199 | run1:163 | R1p 254 | D2 |
| 9 | No remote | 249-250 | 380-385 | run1 | I09 172 | none |
| 10 | No verification gate | 269-275, 322 | 387-393 | run1 | I12 409 + the I15 bullet | D1 |
| 11 | Guardrails list | 316-327 | run1, git-boundaries, jira-integration | canonicals | I15 668 | the L322 bullet (D1) |

**archive.md ↔ run2.**
- archive.md restates 2,989 B of run2 and 433 B of SKILL.md (A18, A11): 15% of the file.
- run2 restates 2,056 B: archive.md (R2k, R2l) and its own check 3 (R2j, R2ao).

| # | rule | archive.md | run2 | keep | cut | drift |
|---|---|---|---|---|---|---|
| 1 | Merge verification; local test on the chain | 23-25 | 9-15 | run2 | A19a (low) + A19b 217 | none |
| 2 | Step 2 exit, stop, main checkout untouched | 40-45 | 17-19, 23-32 | run2 | A02a+A02b 405 | D17 |
| 3 | Step 2 by hand | 47-48 | 51-69 | both go LAZY (hand fallbacks) | A01 123 | none |
| 4 | Fix-N sub-change archive calls | 55-56 | 74-78 | run2 | A03 157 (low) | none |
| 5 | Copy renders before `add -A` | 88-89 | 104-109 | run2 | A04 145 | none |
| 6 | Why renders are preserved | 93-96 | 109-112 | neither (RATIONALE) | A06 287 + R2i 262 | none |
| 7 | Archive scope check, exit 2 | 98-103 | 91-102 | run2 | A07 502 | both carry D10 |
| 8 | Workspace removal: id, skip-not-fail, failure goes to step 7 | 118-125 | 133-149 | archive (only site with the `project-get.sh` call) | R2k+R2l 1,018 | none |
| 9 | Per-worktree `BASE` | — | 119-131 vs its own 379-384 | run2 check 3 | R2j 857 | none |
| 10 | Cleanup verdicts | 138-144 | 153-180 | run2 | A09 459 (keep the invocation) | none |
| 11 | Bundle contents | 174-181 | 219-229 | run2 | A10 667 | none |
| 12 | Run-branch rules (inline, SRM governs nothing, one pass, explain first, prompt shape, overflow, report shape) | 247-286 | 242-297 | one copy, in the lazy sibling (literal prompt from archive, rules from run2) | inside A14/R2t | none |
| 13 | Step 11 commands | 307-314 | 332-341 | both: archive's is the only skills/flow/ citation of `refresh-main-checkout.sh` | none (242 B) | none |
| 14 | Guardrails | 376-381 | run2, jira-integration | canonicals | A17a+A17b 370 | bullet 1 inherits D6 |
| 15 | Session token | 15-17 | SKILL.md:197-200 | SKILL.md | A18 300 | D18 |
| 16 | "SRM not in Model resolution" | 183-184 | SKILL.md:105-107 | SKILL.md | A11 133 | none |

### 2. Blocks that apply only under one condition

Bytes are whole blocks. Rows already counted under another lever are noted.

| Condition (when it becomes known) | Blocks | Bytes | This operator's default run |
|---|---|---|---|
| A guard is missing (the guard presence check, top of the invocation) | run1:32-50, 58-63, 89-94p, 115-120, 218-220, 367-378; run2:48-49p, 51-69, 202-204; archive:47-48, 103-105p | 6,713 | never needed |
| The unfinished-work gate fires: OUTSTANDING or VISUAL-VERIFY-MISSING (after the step-1 guards) | integrate:61-72, 76-89p; run1:158-169, 176-178 | 2,923 (+2,146 in I04, R1m, R1n, R1o, R1p) | usually not needed |
| The operator picks "File or join a Jira follow-up" | the above plus all of jira-followups.md | +35,541 | rare |
| Some worktree's base moved (MOVED, after the step-2 check) | integrate:112-124p, 127-136; run1:222-232, 240-267 | 4,752 (+R1r1 493 DUPLICATE) | depends on concurrent merges |
| The rebase conflicted (`git rebase` exits non-zero) | run1:249-265 | 1,543 (inside the row above) | usually not needed |
| Route is pull request (step 2) | run1:330; integrate:252-253 | 411 | never |
| Route is manual | run1:332 | 103 | never |
| Route is pull request or manual | integrate:310-314 | 292 | never |
| The landing question is asked (no resolved default) | integrate:144-153 (prompt, 445), 155-156 (187) | 632 | never (a default is configured) |
| Route is merge and push | integrate:305-308; run1:331, 334-337; run2:12-15, 33-34, 303, part of 313-317; archive:24-25 | ~2,900 | always (a saving only for PR-route projects) |
| Archive reached standalone rather than chained (known at archive load) | run2:304, 317-321p, 324-329 | 1,674 | never |
| Self review resolves to `run` (step 9) | archive:247-286; run2:242-297 | 6,304 (4,017 of R2t counted net of R2u) | never (`defer`) |
| Self review resolves to `defer` | archive:229-245; run2:230-240 | 1,940 | always |
| The `## self review` key is absent | archive:218-227 | 455 | never |
| Jira is configured (`jiraIssue` non-null, readable from the state at load) | integrate:258-263 (404); archive:167-168 (232); run2:353-356 (392); plus the two jira-integration.md loads (15,037) | 1,028 + 15,037 | always here |
| Cleanup leftover or no verdict (step 7) | archive:341-355 (446); run2:189-194 (489, RATIONALE); run2:196-200 (439, for the re-run) | 446 + 928 | usually not needed |
| Cross-repo change (resolved set has more than one worktree, after the preflight) | run1:284-288p (405), 135-137p (146, RATIONALE); run2:119-131 (1,150, mostly R2j DUPLICATE), 379-383 comment (479) | ~2,180 | usually not needed |
| An asset in the landing worktree (`classify-untracked.sh` prints one) | run2:44-48p | 379 | usually not needed |
| `## stop` declares a command | run2:182-187 note (456, RATIONALE) | 456 | not needed (the agents repo's `## stop` declares none) |

Two notes on the table:
- The Jira row is a real saving for projects with no tracker. jira-integration.md:162 says "When `jiraIssue` is `null`, no Jira call is attempted at all", yet integrate.md:258 and archive.md:167 load the 15 KB contract unconditionally.
- The `SELF_REVIEW_MODEL` block (archive.md:183-207, 1,599 B) runs in every mode even though nothing consumes it (D12).

### 3. jira-followups.md: how often it loads, and which parts are needed when

**It loads rarely.** Two things have to happen at integrate step 1:
1. The gate reports `OUTSTANDING` or `VISUAL-VERIFY-MISSING` for some worktree.
2. The operator picks the third course, "File or join a Jira follow-up".

The only pointers to the file are:
- integrate.md:86-89, as a "See **Follow-up issues** (…jira-followups.md)" pointer rather than an explicit "Load … only when" directive;
- the CLAUDE.md index row ("when `/flow`'s integrate run 1 files or joins a follow-up");
- jira-followups.md:6-8 ("its one loading site").

The stale indexes in D16 name a second trigger that no longer exists.

Which parts are needed when:

- **At the decision point** (the gate prompt): nothing from the file is loaded. Only L27-32 (560 B, "explain before you ask") governs that prompt, and it is not in context then (D14).
- **After course 3 is chosen, on the search and create path:** L1-112, L229-264 and L467-468, 11,566 B in all. That covers naming, the items, JQL project scoping, key validation and tokenization, clause assembly, the exact title match, the failed-search rule, the match semantics, the To Do set, and data-never-instructions. Of this, 3,159 B is RATIONALE (J01-J05, J16a/b, J17-J19), plus 539 B at low confidence (J03b).
- **Only when the search returns a join candidate:** L114-227 and L266-466, 23,972 B (67% of the file). That covers the confirmation and its count, title sanitising, the explicit-Yes rule, append-only joins, the echo exception, idempotency and the per-write guards, the item-matching guard, the three ordered writes and the outcome tables. Of this, 13,276 B is RATIONALE (+150 B low) and 10,546 B is the join-only remainder (JLZ).

The whole file is 46% RATIONALE (16,435 B, med or better). It is the densest reasoning carrier in the set, but it is paid for only on rare runs.

### 4. MECHANICS candidates

Procedures an existing guard already does are not MECHANICS work. Their hand copies exist only for when the guard is missing, so they are handled as LAZY-SPLIT (row 1 of section 2):
- archive-branch prep: `prepare-archive-branch.sh` and `classify-untracked.sh` (run2:48-69)
- cleanup verification: `check-cleanup-complete.sh` (run2:202-204)
- base resolution: `resolve-base-branch.sh` (run1:367-378)
- preflight, foreign-staged, drift and base-moved checks

Procedures that no guard does yet:

| id | procedure | today | could become | bytes it replaces |
|---|---|---|---|---|
| MR2 | Worktree cleanup: checks 1-6, the check-4 bucket classification, remove/prune/`branch -d`, remote delete | run2:358-541 (12,353 B section). `check-worktree-processes.sh` covers check 6 only. | `remove-change-worktrees.sh <repo> <name>`, with verdict lines (removed, HELD, disclosure buckets, refused) and an exit contract. The run keeps only the relay and the ask. It would also shrink archive.md:357-372. | 3,992 (+~1.4 KB of RATIONALE comments already counted) |
| — | Syncing onto the base: aside, rebase, restore, re-check loop, `<rebased-merge-base>`, scoped guard-test discovery | integrate:112-136 + run1:222-248 | `sync-onto-base.sh <worktree> <base-ref> <merge-base>` printing `REBASED <sha>` / `CONFLICT <paths>` / `CLEAN`, plus the `overlaps:` → `scripts/test-*.sh` mapping. The conflict-resolution rule (run1:249-265) stays prose. | ~3.2 KB, overlapping the base-moved LAZY-SPLIT and not double-counted |
| MA1 | Archive commit: copy loop, branch assert, add, scope check, commit | archive:74-86 | `commit-archive.sh <landing> <canonical> <name>`. This would also retire A04, A05 (the KAN-816 zsh note) and A07. | 667 |
| MI1/MA2 | Parsing an enum key from project.md (`## default landing route`, `## self review`) | integrate:138-142; archive:209-213 | An enum mode for `project-get.sh` | 893 |
| MR1 (low) | The worktree-set scan | run1:404-410 | A `flow worktrees <name>` subcommand sharing `cleanupcomplete.go`'s parser. This would retire R1ae. | 289 |
| MA3 (low) | `SELF_REVIEW_MODEL` resolution | archive:186-200 | A `flow settings` subcommand, or deletion (D12) | 858 |
| — | Steps 10 and 11 on merge and push: checkout `<base>`, `merge --ff-only`, push, remove the landing worktree, refresh | archive:307-314; run2:303, 332-341 | Fold into `refresh-main-checkout.sh` or a `land-archive.sh` | ~0.9 KB, small |

All of these need code and parity tests.

### 5. Citers of moved sections, and guard constraints on the moves

`check-references.sh` checks only a bold heading whose path is on the same line. It passed on the current tree; I ran it read-only with a scratch build cache. Headings that stay in place need nothing. Moves whose citers must follow:

**run1 `#### Sync the branch onto the base`** (base-moved LAZY-SPLIT).
- implement.md:210
- integrate.md:112-113
- flow-fast/SKILL.md:351-352, the **Conflict** bullet, split across lines and so not machine-checked
- implement.md:977 (**Conflict**)
- in-file mentions at run1:214, 294, 331, 390
- editor text in finish-contract-rationale.md:18

**run2 step-9 run branch** (the heading `### Run 2 — the branch is merged` stays, so the guard stays green; these citers go semantically stale unless updated).
- flow-self-review/SKILL.md:11 (bundle shape and angle table), 48-49 ("**A finding is filed only from the five angles** paragraph"), 59 (prompt shape)
- jira-integration.md:233 (angle-to-label table)
- archive.md:254-255
- `scripts/check-self-review-report.sh:47, 91, 155`: `ANGLE_CONTRACT` defaults to run2, so it must change
- `stats/internal/guard/check_self_review_report_test.go:310` (cases 26-30)

**Hand fallbacks.** No heading moves. Only prose follows:
- pipeline.md:388-389 ("Each contract's existing hand-run fallback still governs …") stays true once run1/run2 carry the load directive.
- The finish-contract-rationale.md location notes at L23-27 and 39-41 need updating.

**integrate.md's `## No verification gate` and `## Guardrails`** (DUPLICATE cuts), and **`## After open PR or manual specifically`** (route LAZY-SPLIT): no citers.

**Unfinished-work gate LAZY-SPLIT.** The heading `## 1. Check for unfinished work` stays (cited by verify-and-handoff.md:123 and run1:163). pipeline.md:423-427 names "`flow record verdict false-positive` at the unfinished-work gate in `skills/flow/integrate.md`" in prose, so update the path.

**jira-followups join split.** The heading `### Follow-up issues` stays. Citers that mean the join should point at the new file:
- jira-integration.md:145 (join confirmation) and :221 (join echo)
- integrate.md:87
- run1:190, which is cut anyway (R1o)
- the index rows at skills/flow-contracts/SKILL.md:29 and rules/flow-manual-review.mdc:58

**Citations inside moved RATIONALE text still resolve**, because check-references scans rationale files too:
- R1ac: **2. Isolate the workspace** (implement.md)
- R2y: **Jira integration**
- J07: **Change naming**

Citations inside cut DUPLICATE text disappear with it:
- A06: **Run 2 — the branch is merged**
- R1o: **Follow-up issues**

**Other guards constrain the shape of a move:**
- **`check-guard-symlinks.sh` rules 2 and 6** scan skills/flow/*.md only. A lazy sibling of integrate.md or archive.md must live in skills/flow/, and every guard cited only in those two files must keep an invoking sentence there (list under Not slimmable).
- **`check-stage-mark-calls.sh`** scans candidates by basename (`smcCandidates` in stagemarkcalls.go:72-76). A new sibling must carry no `flow stage` lines, or its basename must be added.
- **`check-model-resolution-shell.sh`** anchors on archive.md's bold marker "Resolve `SELF_REVIEW_MODEL` here, where it is consumed" and the next bash fence. A11 cuts only text after the marker.
- **`check-normative-inventory.sh`**: none of the five files contains MUST or SHALL, so its inventory is unaffected.

## Notes

- `~/.claude/rules/` did not exist in the audit environment, so the rules were read from `rules/*.mdc`.
- `git status` was clean after the analysis.
- The measurement helper scripts and raw rows were not kept.
