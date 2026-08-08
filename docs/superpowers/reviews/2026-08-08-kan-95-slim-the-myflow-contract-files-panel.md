# Review panel — kan-95-slim-the-myflow-contract-files

**Pass 1, full roster.** Diff: `.superpowers/sdd/final-review.diff` — 30 files, 1473 insertions,
1506 deletions (2,979 changed lines).

## Roster and why each slot ran

| # | Slot | Required? | Model | Selected because |
|---|------|-----------|-------|------------------|
| 0 | Primary | always | sonnet | — |
| 1 | Bugbot | always | unknown (agent-defined) | — |
| 2 | Principles (Merged) | always | sonnet | — |
| 3 | Security | conditional | unknown (agent-defined) | path handling and file resolution: the `## standards` containment rule, the agents-repo derivation, workspace-id derivation, and a Bash guard's symlink refusal |
| 4 | Adversarial | conditional | sonnet | >~300 changed lines, and a test fixture modified |
| 5 | Lens B — simplicity & state | conditional | sonnet | >~200 changed lines, and 3 new files created |
| 6 | Lens C — robustness & ops | conditional | sonnet | config resolution, guard scripts, failure-path handling |

No slot was excluded. Every conditional trigger fired.

## Findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Bugbot | Minor | `scripts/test-check-contract-budget.sh` | over-budget fixture at 10,000 B sits only ~3.5% above the 9,665 B budget; a fragile margin that could stop exercising the over-budget path if the budget drifts |
| F2 | Principles (Merged) | Minor | `skills/myflow-contracts/project-configuration-rationale.md` | one `##` heading serves ~9 distinct citation targets; the sibling appendices scope one heading per target, so a reader following a citation lands in unrelated text. Resolves mechanically, reaches nothing useful |
| F3 | Primary | Critical | `openspec/changes/kan-95-slim-the-myflow-contract-files/specs/myflow-command-surface/spec.md:3` | MODIFIED requirement title still reads "plus **two** read-only ones" while its body says one; archives a false title. This change's own `myflow-state-machine` delta set the REMOVE+ADD precedent for exactly this |
| F4 | Primary | Important | `skills/myflow-contracts/pipeline-rationale.md:95` | `### The block each state renders` rationale orphaned — Task 2 moved the core section to `handoff-blocks.md`; nothing cites these ~86 lines now. `check-references.sh` checks citations resolve, never that a heading is cited |
| F5 | Security | Important | `skills/myflow-start/SKILL.md:235`, `:448` | citations repointed to a bare `(README.md)`, which `setup.sh` never installs. `/myflow-start` runs with the target project as cwd, so it resolves to that project's own tracked, PR-editable README — attacker-authored prose read as canonical gate procedure |
| F6 | Security | Minor | `skills/myflow-contracts/project-configuration-rationale.md:37` | the containment prohibition "never search the filesystem for a checkout that might be one" moved to the appendix; the core keeps treat-a-miss-as-absence, so a weakening rather than a gap |
| F7 | Security | Minor | `skills/myflow-contracts/finish-contract.md:277` | the worktree scan falls through only when the `worktrees` map is *absent*, not when empty. `{}` makes the preflight and the unfinished-work gate each pass having examined zero worktrees. This is the one thing that genuinely depended on deleted self-heal |
| F8 | Lens B | Important | `skills/myflow-contracts/project-configuration-rationale.md` | same defect as F1's sibling F2, rated higher with fuller evidence: one heading serving ~18 citation sites across ~15 unrelated topics. **Deduped with F2; carried at Important** |
| F9 | Lens B | Important | `skills/myflow-contracts/handoff-blocks.md` | 14,063 B of templates *and* design-history prose, loaded by `/myflow-status` on every detail view. **CONTESTED — collides with `A section reachable from only one command lives in its own file`, which says such a file is not core/appendix split.** Operator decision |
| F10 | Adversarial | Important | `skills/myflow-contracts/pipeline-rationale.md:83` | says "a **five**-command surface"; the pipeline now has four. `pipeline.md` cites this heading "for why", so the citation resolves and reaches a wrong fact. Task 1 fixed the same phrase in the passage it moved and missed this one |
| F11 | Adversarial | Minor | `skills/myflow-do/SKILL-rationale.md:35,40,97` | evicted passages begin mid-sentence under their headings, an artifact of verbatim eviction. Fixable by adding a lead-in *to the appendix* — the moved text itself must not be altered |

*(pass 1 complete — all 7 slots reported. F3 fixed by the controller: `openspec/` is closed to implementers.)*


## Operator decision on F9

Split `handoff-blocks.md`. `A section reachable from only one command lives in its own file`
amended: a single-consumer file is still not split merely for being large, but **is** split where
the resulting appendix would not be loaded at all — the "extra file" the rule declines to pay for
is then not paid. This also gives F4's orphaned rationale a correct home.

## Pass 2 — full roster (escalated: the fixes altered delta specs)

Fix diff: `.superpowers/sdd/fix-round-2.diff` — 8 files, 190 insertions, 308 deletions.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F12 | Security + Principles | **Critical** | `CLAUDE.md:86,122,124,126`, `AGENTS.md:132,168,170,172` | eight bare `(README.md)` citations in the two files `setup.sh:206,222` **copies** into a target project. `CLAUDE.md` is auto-loaded every session there, so the resolution needs no command invocation — a strictly larger surface than F5, which the F5 fix did not cover because the controller's sweep scope omitted the repo-root files. **Found independently by two slots**; Principles rates it Critical, noting `CLAUDE.md` states of itself that Claude Code loads it into every session in the project |
| F13 | Security | Important | `finish-contract.md:46,60`, `skills/myflow-finish/SKILL.md:56,81` | F7's fix corrected the two fall-through sites but not the four consumers, which read the raw map "once per worktree recorded in the state file's `worktrees` map". Empty `{}` still iterates zero times, so `RUN2` and `CLEAR` are both vacuously true and run 2 archives with neither gate having run |
| F14 | Security | Minor | `skills/myflow-start/SKILL-rationale.md:25` | bare `(README.md)` in a file that *is* installed (it sits inside a symlinked skill directory). Same shape as F5, one gate weaker — no run loads an appendix |
| F15 | Security | Minor | `skills/myflow-contracts/handoff-blocks.md:~176` | core states the merge base has "three conditions, not two" but no longer says what they are; the operative *resolve-then-compare, never compare alone* imperative is appendix-only. Not exploitable today — both consumers carry the rule themselves — but a future renderer written from the core alone would miss it |
| F16 | Primary | Minor | `.superpowers/sdd/` | fix rounds 1 and 2 performed 5 rule-extractions and a whole-section move and emitted no per-move ledger. `A move or eviction is recorded in a per-move ledger` binds *any* task that moves prose from a loaded file, not only the numbered ones. Controller omission: the fix dispatches never asked for it |
| F17 | Lens B | Important | `openspec/changes/kan-95-slim-the-myflow-contract-files/specs/myflow-contract-economy/spec.md:103-127` | the amended split test is vacuous: "SHALL be split where the resulting appendix would not be loaded at all" is true of **every** appendix, since no run loads any of them — so it never says no, and would license splitting `jira-followups.md`, which the same requirement forbids. The operative necessity test sits after it rather than leading. Controller wrote this amendment |
| F18 | Lens B | Minor | `docs/superpowers/specs/2026-08-08-kan-95-slim-the-myflow-contract-files-design.md:218` | design record says `Resolving a change's worktrees` "stays in `pipeline.md`"; Task 5's fix round moved it to `finish-contract.md` with controller approval and the design was never squared. Plan and implementation disagree |
| F19 | Principles | Minor | `skills/myflow-contracts/handoff-blocks.md` | all 7 pointers cite one `### The block each state renders` heading covering ~14 distinct sub-facts — the same coarseness F2 fixed in the sibling appendix, milder because it is rationale-only |
| F20 | Lens C | Important | `openspec/specs/myflow-finish-cleanup/spec.md:196-197` | live SHALL still reads "when that is absent"; no `myflow-finish-cleanup` delta exists in this change, so it archives licensing the vacuous empty-map pass F7 identified. Same class as the three planning gaps found during implementation |
| F21 | Lens C | Minor | `scripts/check-cleanup-complete.sh:302`, `scripts/test-check-cleanup-complete.sh:317` | comments cite the contract as finding worktrees "when the map is absent"; the contract now says "absent or empty". No functional bug — the script scans unconditionally — but a guard and its documenting prose must not disagree |
| F22 | Adversarial | Important | `README.md:119-127` vs `skills/myflow-contracts/pipeline.md:45-54` | F5's fix restored the rule to `pipeline.md` but left `README.md`'s original prose in place uncited, so the corpus now holds two independently-authored copies of one rule where it held one. Caused by the controller's instruction to "leave README.md's human-facing prose alone". `README.md:129-130` already uses the correct citation pattern nine lines later |

## Pass 2 outcome

All 7 slots reported. Eleven pass-1 findings verified ADDRESSED. Eleven new findings (F12-F22), of
which **seven originate in the controller's own fixes or planning** rather than in the implementers'
work. F17 and F20 fixed by the controller directly (`openspec/` is closed to implementers).

## Fix round 3 — controller adjudication of one raised concern

The round-3 fixer flagged F6's fix as leaving "a duplicate in `project-configuration-rationale.md`,
same shape as F22". **Examined and rejected — not a defect.** `project-configuration.md:114-117`
carries the prohibition *and* cites **Confirm the derived root before using it**
(`project-configuration-rationale.md`), where the original passage lives. That is exactly the
sanctioned rule-extraction shape: the original moves to the appendix byte-for-byte, the core gains
the rule plus a citation to it. F22 was a defect because `README.md` is not an appendix and no
citation linked the two copies. Recorded rather than silently discarded.

## Pass 3 — full roster (escalated: three fix rounds have run)

Fix diff: `.superpowers/sdd/fix-round-3.diff` — 12 files, 117 insertions, 229 deletions.

**Primary:** F12, F22, F13, F21, F14, F15, F19, F16 all **ADDRESSED**, verified against files rather
than the fix report. No regressions against rounds 1-2. All four integration checks pass: delta specs
match the tree, and `myflow-finish-cleanup`'s four scenarios are each satisfied by the contract text.
No new findings at any severity.

**Deferred, pre-existing, out of scope for this change:** `myflow-contract-economy`'s scenario
"Follow-up issues moves to its own file" asserts `jira-integration.md` "does not contain the
`Follow-up issues` section", but a redirect stub carrying that heading exists at
`jira-integration.md:252`. The content did move; the scenario's literal wording did not follow.
Predates this branch. For a follow-up, not for pass 3.

**Principles:** F12 **ADDRESSED**. One new Important — F23.
**Bugbot:** round 3 otherwise clean; one new Minor — F24.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F23 | Principles | Important | `README.md:187-188,191,198` | round 3's own F13 fix changed how worktrees are resolved in `finish-contract.md`; README's Level-2 expansion still restates the old mechanism ("once per worktree recorded in the state file", "every recorded worktree returns `RUN2`"). Pre-round-3 the two matched; now the README is silently false. Same SSOT class as F22, caused by round 3. README's own text at L108-110 warns against exactly this |
| F24 | Bugbot | Minor | `CLAUDE.md:85-86,124,127,130`, `AGENTS.md` equivalents | F12's repointing named files without sections — "spelled out in its own `SKILL.md`" — where the old citations named `**Level 1 — the stages of each command**`. `scripts/check-references.sh` verifies only named-section citations, so these are now unverified by any guard |
**Lens C:** F20 and F21 **CLOSED**. Zero findings. The empty-worktree-set path — which failed twice —
verified correct by tracing both gates: all three consumers resolve through
**Resolving a change's worktrees**, none reads the raw map, an empty resolved set stops. Every pass-1
failure path re-verified undisturbed. Noted without raising: 6 of the 14 new rationale sub-headings
have no inbound citation, which is supporting material in a file no run loads rather than a broken link.
**Lens B:** F17 **SOUND — defect closed.** The separability test is content-dependent and can say no;
both worked examples verified against the actual files; the never-loaded property correctly demoted to
explaining why the saving is one-sided rather than doing the deciding. Round 3's own changes hold up
under both lenses, and the worktree state model is sound: existence is derived by git-scan when the map
cannot be trusted, merge-base value is only ever read and never guessed.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F25 | Lens B (+ round-3 fixer) | Important — **CONTESTED** | `project-configuration.md:112-114` vs `project-configuration-rationale.md:40-46` | the bold rule sentence, the "Check that the path…" sentence and the "never a licence… never a reason…" clause appear verbatim in both files. Raised independently by the round-3 fixer and by Lens B; **the controller adjudicated it as sanctioned once already** and that adjudication was too quick. It is the shape rule extraction produces — original to the appendix byte-for-byte, core gains the rule — so it violates no rule; but the overlap is larger than the carve-out contemplates. **The proposed fix (trim the appendix restatement) would violate the byte-for-byte clause of the same requirement.** Operator decision |
**Security:** F12, F13, F15 **RESOLVED** (verified against files). F14 **half-fixed**. Four protected
surfaces intact. Four new findings:

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F26 | Security | **Critical** | `skills/myflow-contracts/finish-contract.md:269`, `skills/myflow-finish/SKILL.md:318` | both cite `openspec/specs/myflow-self-review/spec.md` as **canonical for the procedure** of run 2 step 8. `setup.sh` never installs `openspec/`, so it resolves to the target project's own tracked, PR-editable openspec tree. Worse than F12: cited as canonical procedure, executed straight after run 2's destructive cleanup with `FINISHED` already written. **New on this branch** — at base the procedure lived in `pipeline.md` (installed); Task 1 deleted that section with the stage tables and repointed here |
| F27 | Security | Important | `skills/myflow-do/SKILL.md:387`, `skills/myflow-status/SKILL.md:58` | two more raw-`worktrees`-map iterations. `{}` makes `/myflow-do` run the workspace-isolation guard zero times — a real safety guard silently skipped. **Structural:** F13's fix lives in `finish-contract.md`, which only `/myflow-finish` loads, so neither command can cite it. Fix by hoisting the empty-set rule into `pipeline.md` |
| F28 | Security | Minor | `skills/myflow-start/SKILL-rationale.md:23-25` | F14's repoint went to `**Convergence**` (`skills/myflow-start/SKILL.md`), which never defines a planning-stage exchange, never states the convergence test, and never argues one-test-vs-rule-per-gate. Security half closed, correctness half not; also circular — the appendix now cites the section it annotates. Should point at `**Stage exit**` (`pipeline.md`) |
| F29 | Security | Minor — **pre-existing, out of scope** | `CLAUDE.md:128`, `AGENTS.md:174`, `finish-contract.md:15,60,129,221`, `myflow-finish/SKILL.md:56,85,145,295`, `myflow-do/SKILL.md:213,370,375,475` | bare `scripts/*.sh` citations resolve project-relative in an installed project. A PR adding `scripts/check-unfinished-work.sh` gets it **executed** by `/myflow-finish` run 1 — arbitrary code execution, strictly worse than a read. Present at base `f763481`; only `myflow-do/SKILL.md:384` uses the `<agents repo>/` prefix. **Not this change's defect — for a follow-up ticket** |
**Adversarial:** F22 **FIXED** (no duplicate remains). F12's 8 citations, F13's 3 execution sites and
stop condition, F19's 14 headings, F15's move, cross-round interference — all verified clean. Two new:

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F30 | Adversarial | Minor | `skills/myflow-finish/SKILL.md:48` | "the four preflight checks" undercounts: round 3's F13 fix inserted a fifth outcome bullet two lines below it |
| F31 | Adversarial | Important | `README.md:98,188,191,198`, `CLAUDE.md:128`, `AGENTS.md:174`, `commands/myflow-finish.md:14`, `commands-claude/myflow-finish.md:10` | **supersedes F23 with a wider site list.** The description layer still says the gates run "once per *recorded* worktree", contradicting the scan-fallback F13 installed. Execution layer is correct; this is the same layer F21 was about. `CLAUDE.md`/`AGENTS.md` weigh more than README by F12's own precedent — installed and auto-loaded |

## Pass 3 outcome

All 7 slots reported. Of pass 2's findings, **all verified ADDRESSED**. F17's rewritten requirement
judged **sound** by the slot that found it defective. Eight new findings: F26 (Critical), F27 and F31
(Important), F24, F28, F30 (Minor), plus F25 **contested — operator decision** and F29 **pre-existing,
out of scope — needs its own ticket**.

## Pass 4 findings (targeted: Primary, Security, Adversarial, Bugbot)

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F32 | Security | **Critical** | `skills/myflow-do/SKILL.md:387` | round 4's own fix converted a silent skip into a guaranteed block: `/myflow-start` writes `worktrees: {}` and `/myflow-do` populates the map only at its final state write, *after* the guard that now stops on an empty resolved set. Every first run would halt before lint, test, staging and the state write. Fixed in round 5 by giving `/myflow-do` its own resolution — the worktree it created or resumed — which `pipeline.md` already delegates to each command |
| F33 | Security | Minor | `skills/myflow-contracts/state-file.md:147`, `skills/myflow-do/SKILL.md:80` | "the key set of `worktrees` is the authoritative list" read as contradicting `pipeline.md`'s "never loops over the map directly". Both true of different things — the map is the record, the resolved set is what a step iterates — now stated so |
| F34 | Bugbot | Minor | `CLAUDE.md:128`, `AGENTS.md:174` | cited **Run 1 — integrate** for the stages of *both* runs; Run 2's stages are a separate section. Both now named |
| F35 | Bugbot | Minor | `skills/myflow-contracts/finish-contract.md:312` | "Per the pipeline-wide rule above" — the rule is in `pipeline.md`, not above in that file. Replaced with the citation |
| F36 | Primary + Adversarial | Important | `openspec/changes/.../specs/myflow-finish-cleanup/spec.md:11-12` | the controller's own delta still carried the retired "once per *recorded* worktree" phrasing that round 4 eliminated everywhere else — and this is the text run 2 syncs into the live spec at archive. Fixed by the controller |
| F37 | Adversarial | Minor | `docs/superpowers/specs/2026-08-08-…-design.md:218-222` | the design record said `/myflow-do` has nothing to do with the worktree-resolution section; round 4's own hoist made that false. Fixed by the controller, recording the hoist rather than hiding it |

## Final state — pass 4 (targeted) and fix round 5

Pass 4 ran Primary, Security, Adversarial and Bugbot over round 4. All six of round 4's findings
verified ADDRESSED. Five new findings, one **Critical**: round 4's own fix for the silent-skip
defect turned it into a guaranteed hard block on `/myflow-do`'s first run, because `/myflow-start`
writes `worktrees: {}` and `/myflow-do` does not populate the map until after the guard that now
stopped on empty.

Fix round 5 — the cap — closed all five. The Critical was verified by tracing a first run end to
end: section 2 now defines `/myflow-do`'s resolved set as the worktree it created or resumed,
non-empty by construction, and section 7 cites that resolution.

Operator decisions applied: F25 trimmed (appendix keeps the argument alone; the exception is recorded
in the requirement with a test), `.serena/` gitignored, F29 filed as **KAN-101** at Highest.

findings-total: 37
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 fixed
finding-status: F16 fixed
finding-status: F17 fixed
finding-status: F18 fixed
finding-status: F19 fixed
finding-status: F20 fixed
finding-status: F21 fixed
finding-status: F22 fixed
finding-status: F23 fixed
finding-status: F24 fixed
finding-status: F25 fixed
finding-status: F26 fixed
finding-status: F27 fixed
finding-status: F28 fixed
finding-status: F29 withdrawn — pre-existing, not introduced by this change; filed as KAN-101 at Highest by operator decision
finding-status: F30 fixed
finding-status: F31 fixed
finding-status: F32 fixed
finding-status: F33 fixed
finding-status: F34 fixed
finding-status: F35 fixed
finding-status: F36 fixed
finding-status: F37 fixed
