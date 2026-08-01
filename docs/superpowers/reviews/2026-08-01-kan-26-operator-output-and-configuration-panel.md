# Final review panel — kan-26-operator-output-and-configuration

**Diff reviewed:** `.superpowers/sdd/final-review.diff` — 730 changed lines, 13 files, from merge base
`1797b84`, with `openspec/`, `docs/manual-test/` and `docs/superpowers/` excluded so the panel read
exactly what will be staged.

## Pass 1 — full roster

**Mode:** full. Every slot in the roster ran; this is pass 1, which always runs the complete roster
selected for the change.

| Slot | Included | Model | Why |
|---|---|---|---|
| 0 Primary | required | sonnet | — |
| 1 Bug hunt | required | sonnet | — |
| 2 Principles (Merged) | required | sonnet | — |
| 3 Security | **selected** | sonnet | the Jira join reads externally-authored issue titles and descriptions, then writes back to that issue and unions its labels; plus a shell-guard edit |
| 4 Adversarial | **selected** | sonnet | 730 changed lines against a >~300 trigger; plus a state-file key rename with a compatibility read, which is a migration in all but name |
| 5 Principles lens B (simplicity & state) | **selected** | sonnet | >~200 changed lines |
| 5 Principles lens C (robustness & ops) | **selected** | sonnet | external integration (Jira) and a state-schema change |

**No optional slot was excluded.** Every trigger fired.

**Harness deviation.** This harness exposes no `bugbot` or `security-review` agent type, so slots 1
and 3 ran as general-purpose agents executing the prompts the skill itself ships for that purpose
(`bug-hunter-reviewer-prompt.md`, `security-reviewer-prompt.md`). Their model is therefore known
rather than agent-defined, and is recorded as `sonnet` rather than `unknown (agent-defined)`.

`[PRINCIPLES_PATH]` resolved to `/Users/tweety53/.claude/skills/myflow-do/engineering-principles.md`
(verified present, passed as a path and never pasted). `[STANDARDS_PATHS]` resolved from
`.myflow/project.md`'s `## standards`: both entries are bare filenames containing no `/` (form 2),
resolving to `<worktree>/CLAUDE.md` and `<worktree>/AGENTS.md`, each verified to exist with its
normalized parent exactly the project root. **No standards entry was dropped.**

## Findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Adversarial + Lens C | Critical | `skills/myflow-status/SKILL.md:43` | The retired-key compatibility read is promised by two contracts but implemented nowhere; the jq expression reads `.planningEffort` with no `.effort` fallback, so an unmigrated file reports "not recorded — planned at default" when it recorded a real level. Two slots found this independently. |
| F2 | Principles (Merged) | Critical | `skills/myflow-contracts/pipeline.md:410` | "Defined here and nowhere else" is asserted, not achieved: the three producing skills still carry independently authored blocks with no citation to the template, and divergence is already realized — `pipeline.md:448` has `<artifactUrl, or "missing">` where `myflow-start/SKILL.md:243` has `<artifactUrl>`, and `pipeline.md:495` has `<prUrl, or why there is none…>` where `myflow-finish/SKILL.md:232` enumerates three literal alternatives. |
| F3 | Lens B + Principles (Merged) | Critical | `skills/myflow-contracts/state-file.md:63` | Three files make mutually exclusive canonicality claims for the model roles and defaults: `state-file.md` names the OpenSpec spec canonical and warns against a second copy, `pipeline.md:1088-1118` is that second copy, and `CLAUDE.md:106` / `AGENTS.md:152` name `pipeline.md` as the single source. Runtime consumers read the file not named canonical. |
| F4 | Bug hunt | Critical | `skills/myflow-contracts/pipeline.md:1058` | "Every review-panel reviewer runs on Sonnet" survives as an unqualified absolute in seven places — `pipeline.md:1058`, `pipeline.md:1075`, `CLAUDE.md:92`, `AGENTS.md:138`, `README.md:326`, `skills/README.md:31`, `commands/myflow-do.md:8` — contradicting the per-change override the same files now describe. Task 12 fixed this wording in one spot; the sweep never generalised, and Task 9's command-tree sweep grepped only `effort`, never `Sonnet`. |
| F5 | Security | Important | `skills/myflow-contracts/jira-integration.md:229` | The join selects a write target by label + title + status — all attacker-settable — across an unnarrowed project-wide search, then writes to it and echoes its unreviewed description into the operator's handoff, with no confirmation step. The same file's **Resolution** section requires explicit confirmation for any non-literal candidate match. **Conflicts with a recorded design decision — see the handback below.** |
| F6 | Security | Important | `skills/myflow-contracts/jira-integration.md:268` | The reused pre-write assertion compares the payload against the same run's own read, never against live state, so it is not a compare-and-swap; two concurrent finish runs joining one issue can silently lose one's append. The join design deliberately fans multiple changes onto one issue, which widens this window. |
| F7 | Lens C | Important | `skills/myflow-contracts/jira-integration.md:280` | The degradation rule names a failed creation and a failed join; a failed **search** is a third, unnamed case. The natural reading — "no match, so create" — reintroduces on every transient failure exactly the duplicate proliferation the feature exists to prevent. |
| F8 | Lens C | Important | `skills/myflow-contracts/jira-integration.md:272` | Only the description append is bound to the pre-write assertion; the retitle and label union are stated as separate unconditional writes with no guard, and the binary "failed join" vocabulary cannot express a partial join — so the exact invisibility the label-union rule exists to prevent can occur and be reported as success. |
| F9 | Lens C + Adversarial | Important | `skills/myflow-contracts/pipeline.md:472` | On the merge-and-push and manual routes the wrong `IN_PROGRESS` rendering is not merely sparse: it prints `git -C <worktree> diff --cached`, a literal command that returns nothing for an already-committed branch, and the `Git:` field enumerates only `staged and uncommitted` or `committed and pushed to the PR branch`, neither of which is true. |
| F10 | Lens C | Important | `skills/myflow-contracts/pipeline.md:551` | The tab commands are mandated for all three pipeline commands in all harnesses, but the reason they are printed rather than invoked is stated in Claude-Code-only terms with no fallback — unlike the sibling **Progress visibility** section in the same file, which explicitly answers the identical multi-harness question. |
| F11 | Adversarial | Important | `skills/myflow-contracts/state-file.md:79` | No precedence rule for a file carrying both `effort` and `planningEffort`. Both keys are individually excepted from the closed-schema rule, so such a file parses cleanly with no defined answer for which value governs. |
| F12 | Adversarial | Important | `skills/myflow-contracts/state-self-heal.md:73` | The retired-key exception is unconditional, but the mapping covers exactly three values; a file carrying `"effort": "urgent"` is declared parseable with no defined target level. |
| F13 | Adversarial | Important | `skills/myflow-contracts/pipeline.md:470` | The `Panel:` field is not marked `(run-only)`, so by the section's own rule it must be regenerable — but `/myflow-status` is given no source for it, and never mentions the panel record that is the only on-disk trace of which optional slots ran. |
| F14 | Lens B | Important | `skills/myflow-contracts/state-self-heal.md:56` | The pre-existing "**There is no legacy-value migration**… a mapping table would be a second, drifting definition" sits directly above the mapping table this change adds, with no reconciliation. The rule that should have governed this decision does not actually rule it in or out. |
| F15 | Lens B | Important | `skills/myflow-contracts/jira-integration.md:257` | The join is not idempotent under retry: a re-entered run 1 choosing the filing course again finds its own prior follow-up and appends a second identical dated section. |
| F16 | Bug hunt | Important | `openspec/changes/kan-26-operator-output-and-configuration/tasks.md:764` | Task 11 Step 3 still carries `git checkout -- skills/myflow-contracts/state-file.md`. Under the no-per-task-commit rule that discards uncommitted work and then reports a green guard over the loss. The implementer avoided executing it but the plan text was never corrected, leaving the landmine live for any re-run or reuse. |
| F17 | Bug hunt | Important | `skills/myflow-contracts/pipeline.md:450` | The `STARTED` template promises a Jira issue **URL**; no file stores or computes one, `jiraIssue` holds only the key, and the block `/myflow-start` actually prints carries no URL. The template claims a field no renderer emits and `/myflow-status` cannot regenerate. |
| F18 | Bug hunt | Minor | `skills/myflow-finish/SKILL.md:92` | The gate's option is still labelled "File a Jira task, then continue", but the usual outcome is now joining an existing follow-up rather than filing anything. |
| F19 | Primary | Important | `openspec/changes/kan-26-operator-output-and-configuration/specs/myflow-handoff-output/spec.md:21` | The spec's contents table omits fields the templates carry — `Decisions recorded` and `Models` for `STARTED`, `Panel` / `Progress` / `Git` for the no-`prUrl` `IN_PROGRESS` row. Task 12 Step 5 split the row but never brought the field lists with it, so a contributor reading the SHALL-table as normative would "correct" `pipeline.md` by deleting them, reopening the exact drift the capability exists to close. |
| F20 | Primary | Important | `skills/myflow-contracts/pipeline.md:517` | For a change stopped at a run-2 cleanup leftover, one `/myflow-status <name>` invocation contradicts itself: its table correctly reports "branch merged → it will archive" from its own ancestor test, while the detail block, keyed on `prUrl`, prints `## Branch integrated — waiting on the merge`. This is **not** the accepted `prUrl` ambiguity — the merge is a proven fact from the same command's step 2, so the discriminator is being used where a better signal is already in hand. |
| F21 | Primary | Minor | `skills/myflow-contracts/state-file.md:62` | The `models` bullet cites `openspec/specs/myflow-model-policy/spec.md` as a bare backtick path with no adjacent bold token, so `check-references.sh` never fires on it and it rots silently. Note the bind: Task 10 established that *adding* a bold token to that path fails the guard permanently, because the file's headings are all `### Requirement: …` — so an OpenSpec spec cannot currently be cited in any checkable shape at all. |


## Fix round 1

One fix subagent took the union of all 21 findings — not one fixer per finding. Model: **opus**, per
the panel-fix role's default (the role applies fixes, so it is an implementer and sits at the
ceiling). Diff: `.superpowers/sdd/fix-round-1.diff`, **512 changed lines** in the implementation
alone.

**F5 was decided by the operator, not by the panel.** It collided with `design.md`'s
`join-search-two-todo-statuses`, which recorded that asking before each join had been considered and
rejected. Put to the operator at the handback; they chose to **add the confirmation**, reversing that
earlier call. The decision record was updated to show the reversal rather than deleting the rejected
alternative, and the delta spec's scenarios were amended so a declined confirmation is a defined
outcome.

**Three things the fix wave established that the panel had not:**

- **F1's defect had two more instances.** `/myflow-do` and `/myflow-finish` both instructed carrying
  `planningEffort` forward *verbatim*, which for a file recording the level under the retired key
  means either preserving the old key forever or writing `planningEffort: null` and erasing the
  operator's choice. `/myflow-finish` step 7 is the terminal write, so that one was permanent.
- **F21's premise was wrong, and it was measured rather than argued.** An OpenSpec
  `### Requirement: …` heading *is* citable in a fully checked shape, by naming the requirement in
  full as the bold token. Task 10 had concluded such a file "cannot be cited in any checkable shape at
  all"; that conclusion is retracted. Both surviving spec pointers are now checked rather than
  documented as gaps.
- **Two findings interacted.** F13 marking `Panel:` run-only means F19's contents table must *not*
  list the panel roster among what a regenerated block carries; and F5's confirmation prompt broke a
  "single carve-out from Never blocking" claim in the same file, which is now "one of exactly two",
  both bounded.

**Two deliberate departures from a finding's literal wording**, both argued rather than silently
taken: F2's `prUrl` placeholder stays on both sides, because moving the finish command's enumeration
into the template would contradict the template's own describe-don't-reproduce policy — the
divergence was closed instead by making the description↔enumeration relationship explicit and fixing
the one place a skill was genuinely *narrower* (`artifactUrl` gained `| missing`); and F19's `Panel`
row, per the interaction above.

## Pass 2 — full re-run, escalated automatically

**Mode: full, not targeted.** Escalation was automatic and is recorded rather than chosen: the fix
diff is **512 changed lines against a ~150 threshold**, and it **altered delta specs and public
contracts** (`myflow-handoff-output`, `myflow-jira-projection`, `myflow-planning-effort`, plus
`design.md` and four contract files). Either trigger alone forces a full re-run; both fired. Every
slot in this change's roster therefore re-runs against a **rewritten** `final-review.diff` (1146
changed lines from the merge base), not against the fix diff.

### Pass 2 result

**F1–F21 all hold.** Every slot re-verified its own pass-1 findings against the live tree rather than
trusting the fix report; the primary and Merged-principles slots additionally reproved the load-bearing
citations by sandbox mutation. **Pass 2 raised 10 new findings, 2 Critical**, most of them in text the
fix wave itself wrote.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F22 | Adversarial | Critical | `skills/myflow-contracts/pipeline.md:562` | **The F20 fix is a regression.** Making merge status authoritative over `prUrl` relies on `git merge-base --is-ancestor`, which returns true for a branch with **no commits of its own** — the ordinary `IN_PROGRESS` state, since `/myflow-do` stages without committing. So `/myflow-status` prints "merged and waiting on run 2" for a change never once through finish. Verified empirically. Pre-fix, `prUrl: null` rendered this correctly. This repository's own `scripts/check-finish-preflight.sh:69-79` documents the identical trap and guards it with a `HEAD == recorded merge base` pre-check that `myflow-status/SKILL.md:61`'s ancestor test lacks. |
| F23 | Adversarial | Critical | `skills/myflow-contracts/state-self-heal.md:90` | **F12's promise is implemented nowhere — the same shape as the original F1.** The unmapped-value rule declares the file unparseable, but the contradiction table has no row for it, `/myflow-status`'s five artifact checks never inspect the value, and `/myflow-do` / `/myflow-finish` have no defined value to write. `myflow-status/SKILL.md:119` asserts a `⚠` its own procedure cannot produce. |
| F24 | Security | Important | `skills/myflow-contracts/jira-integration.md:270` | The join confirmation interpolates the candidate's externally-authored title **unescaped and unbounded** into the bolded question the operator must read to decide. The same file caps a summary-derived slug at `[a-z0-9-]+`/48 chars because it reaches a path; nothing constrains untrusted text reaching a human decision. The data-never-instructions clause protects the agent, not the operator. Compounded by the title predicate being stated semantically, so a JQL `~` match would widen the candidate space. |
| F25 | Security + Lens C + Adversarial | Important | `skills/myflow-contracts/jira-integration.md:355` | The idempotency guard runs *before appending* and says "write nothing" on a match, with no carve-out for "appended but retitle/label-union never succeeded". A partial join is therefore **permanently abandoned** on every later run, and the report **downgrades** from `⚠ partially joined` to an unmarked "no write" line — reproducing, via the retry path, exactly the invisibility F8 existed to prevent. Three slots converged. |
| F26 | Adversarial + Primary | Important | `skills/myflow-contracts/jira-integration.md:356` | The same guard matches on `## From <KEY> — <date>` **for this date**. A retry the next day finds no section for today, and appends a second content-identical one — reproducing F15's original bug, gated on a date rollover. A partial-join failure is exactly the kind an operator fixes and retries later. |
| F27 | Bug hunt | Important | `skills/myflow-contracts/pipeline.md:1159` | The citation names `**Requirement: Implementer subagents run on the strongest available model**`, which is **not a heading** in the target spec — that phrase appears only as prose inside another requirement's body. Green today only because an unresolved path is silently skipped. Proven by copying the tree, simulating finish run 2's spec sync, and re-running the guard: it fails. Contradicts the fix report's own citation-proof table. |
| F28 | Bug hunt | Important | `skills/myflow-finish/SKILL.md:253` | New text claims run 1 "ends with the branch pushed and unmerged, so *waiting on the merge* is true for every run of it". False for the **Merge and push** route, which merges into the base branch within run 1's own step 1.3, before this block prints. The operator is told to wait for a merge that already happened, beside a `Route:` line reading "merged and pushed". |
| F29 | Primary | Important | `openspec/changes/kan-26-operator-output-and-configuration/specs/myflow-contract-distribution/spec.md:26` | An **eighth** unqualified "every slot runs on Sonnet", inside the requirement that defines what a level-2 expansion must contain — the exact text whose `pipeline.md` counterpart F4 fixed. Survived because `openspec/` is excluded from the reviewed diff, so no panel pass had read it. At finish run 2 it syncs as permanent normative text describing pre-F4 behaviour. |
| F30 | Bug hunt (disputed by Primary) | Important | `skills/myflow-start/SKILL.md:302` | Bug hunt read "**The one carve-out**" as contradicting `jira-integration.md`'s new "one of exactly two". Primary judged it correctly scoped — the join confirmation never occurs during `/myflow-start`. **Controller ruling: the primary is right on substance, but two independent readers reached opposite readings of the same sentence, which is itself the defect.** Fix by scoping it explicitly rather than by argument. |
| F31 | Lens B | Important | `skills/myflow-contracts/state-file.md:79` | The compatibility apparatus — fallback read, two-key precedence, bounded mapping, unmapped-value carve-out — spans six files, where the retired `stage` field in the same contract uses one existing mechanism (unparseable → self-heal → announce). Three pass-1 findings (F11, F12, F14) and F23 above are all holes this apparatus created. **Conflicts with an approved requirement — see the handback below.** |


## Pass 3 — full re-run, escalated automatically

Fix round 2 was 306 changed lines and altered delta specs and the proposal — either trigger forces a
full re-run. Diff rewritten to 1200 changed lines from the merge base.

**Clean slots:** Security (no defect; verified F24/F25/F26 on the merits and showed the F25xF26
collision cannot occur) and Principles-Merged (no Critical, no Important; both its pass-1 Criticals
still closed). **F22, F28, F30 and the F31 removal's mechanics all verified holding** by multiple
slots, several with empirical git sandboxes.

**But pass 3 found the F31 decision itself rested on a false premise the controller had supplied.**
That is this pass's headline result.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F32 | Lens C | Critical | `openspec/changes/kan-26-operator-output-and-configuration/design.md:125` | **The accepted cost of the F31 reversal is empirically false.** It claims the loss is "one field that only affects a revision round's reasoning depth" and that only two `FINISHED` files are affected. Verified against the real state directory: **four** files carry the retired key, and one is **this change's own file at `STARTED`**, holding a live `artifactUrl` and `jiraIssue: KAN-26`. `state-self-heal.md` infers `state` **only**, so a rebuild drops `branch`, `worktrees`, `artifactUrl`, `jiraIssue`, `models` and `prUrl` — real operator-set data, not reasoning depth. |
| F33 | Adversarial | Critical | `skills/myflow-contracts/state-self-heal.md:71` | **The "undocumented key implies unparseable" detection is implemented nowhere** — the same promised-but-unimplemented shape as F1 and F23, one level down. `/myflow-do` and `/myflow-finish` never mention self-heal (zero grep hits); `/myflow-status`'s literal jq projection silently ignores unnamed keys, proved empirically. The loud announcement the whole reversal depends on never fires on the two commands most likely to touch such a file first. |
| F34 | Lens B + Primary + Adversarial | Critical | `openspec/changes/kan-26-operator-output-and-configuration/design.md:360` | `### The old key is read, never migrated` still asserts the reversed compatibility-read design as current, contradicting the reversal recorded at `design.md:110` in the same file. Three slots found it. It survived round 2's straggler grep because the passage uses none of the searched literals, and `design.md` is archived verbatim and read by no guard, so it would stay wrong permanently. |
| F35 | Lens C | Important | `skills/myflow-contracts/state-self-heal.md:35` | A rebuild nulls `worktrees` though the path is mechanically recoverable from `git worktree list`, which `pipeline.md` itself documents. `/myflow-finish`'s preflight and unfinished-work gates both **iterate that map**, so an emptied map makes both pass vacuously — zero worktrees examined — for a change that may hold a real unmerged worktree. |
| F36 | Adversarial + Lens C | Important | `skills/myflow-status/SKILL.md:57` | The merge-base pre-check names only two conditions, equal and absent. `check-finish-preflight.sh:65-68` and the Finish contract name a **third**: present but **unresolvable**. Verified in a sandbox — an unresolvable value reads as "not equal", falls through to the bare ancestor test, and reports merged for a branch with no commits of its own: F22's exact regression through a narrower trigger. |
| F37 | Adversarial | Important | `skills/myflow-contracts/jira-integration.md:413` | The append guard matches on "the item list" but never defines the comparison — no ordering, case, whitespace, or exact-vs-set semantics. Read strictly it double-appends when the outstanding list shifts between attempts; read loosely it silently drops a newly-added item, and no outcome row reports "new items merged into an existing section". |
| F38 | Principles (Merged) | Minor | `skills/myflow-contracts/pipeline.md:576` | The "why the recorded-merge-base check precedes the ancestor test" reasoning is authored at near-equal length in both `pipeline.md` and `myflow-status/SKILL.md`. Each correctly claims not to re-derive the *script's* reasoning, but each re-derived it relative to the other, with no note of why two copies were kept. |
| F39 | Adversarial | Minor | `skills/myflow-status/SKILL.md:54` | No combination rule for a multi-repo change: the Finish contract says proceed only when **every** worktree returns `RUN2`, but `/myflow-status`'s procedure is singular throughout and never says how to render two worktrees that disagree on merge status. |

### Handback — the F31 decision was re-put to the operator, and reversed again

F32 established that the controller's own summary to the operator — "no state file records a non-null
effort except two already `FINISHED`" — was **false**. The decision to remove the compatibility
apparatus had been taken on that basis. The corrected facts (four files, one of them this change's own
open file; and F33 showing the replacement mechanism unimplemented) were put back to the operator, who
chose to **restore the compatibility read** — the design `proposal.md` originally approved.

F31 is therefore **withdrawn**: its removal is itself reverted. The apparatus returns, and with it the
pass-1 fixes that closed its real gaps (F11 precedence, F12 unmapped value, F14 reconciliation) —
except that an unmapped value now reads as *not recorded* rather than *unparseable*, because declaring
unparseable requires exactly the detection mechanism F33 proves nothing implements.

## Pass 4 — full re-run, and the point at which the run stops

Fix round 3 was 306 changed lines and altered delta specs — full re-run again. Diff rewritten to
1372 changed lines. **F32–F39 all verified holding**, several by mutation probe and one by a real
git sandbox exercising all six merge-base conditions. Security and Lens C both judged their own
areas closed; Lens C declared the branch robust.

**But pass 4 raised 9 more findings, 2 Critical — and the finding count has stopped converging:
21 → 10 → 8 → 9.**

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F40 | Bug hunt | Critical | `skills/myflow-contracts/jira-integration.md:384` | **Carryover, never fixed — a controller tracking failure.** The pass-3 Critical that the partial-join retry is scoped to run 1 was folded into round 3's brief as prose under F37 instead of being recorded as its own finding. `fix-round-3.diff`'s only hunk in this file changed the item-comparison semantics and left `:384-385` and `:404-412` byte-identical. The unqualified claim "a partial join is re-attempted, not abandoned, and keeps its ⚠ until it is complete" is still false past the merge boundary: once preflight returns `RUN2` the retry path is unreachable, and no durable record names which follow-up was partially joined. |
| F41 | Adversarial | Critical | `skills/myflow-contracts/state-self-heal.md:112` | The `planningEffort` exemption from being named among unrecovered fields is **unconditional**, but its justification only argues the absent-or-unmapped case. A file holding a real operator-set level that is unparseable for an *unrelated* reason is rebuilt, the level is nulled, and the announcement never names it — contradicting the same file's "name every field that could not be recovered, so the loss is visible at the moment it happens". Corroborated independently: `design.md:139-142`, written this round to fix F32's incomplete loss-accounting, **omits `planningEffort` from its own loss list**. Fifth recurrence of the promised-but-unimplemented class, this time inside the mechanism built to stop it. |
| F42 | Merged + Lens B | Important | `skills/myflow-contracts/state-file.md:114` | The full justification for "unmapped reads as *not recorded*, not *unparseable*" — the same three empirical claims — is independently authored in both `state-file.md` and `state-self-heal.md` rather than stated once and cited. The same round applied the opposite discipline to fix F38, and said why. |
| F43 | Security | Important | `skills/myflow-contracts/jira-integration.md:422` | The append guard's "already present" set is built from the target issue's **live description**, which any project member can edit, with no provenance check. Items are deterministic pipeline output derived from repo-visible state, so a forged `## From <KEY>` section makes a run skip appending real outstanding work and report clean success. The confirmation gate cannot catch it — it shows title and status only, deliberately never the description. |
| F44 | Lens B | Important | `skills/myflow-contracts/state-self-heal.md:25` | Line 25 now says self-heal infers `state` **and**, on a rebuild, the keys of `worktrees`; line 30, five lines later and untouched, still says its "only owned field is `state`" and that `worktrees` is "re-emitted exactly as read". The canonical self-heal contract contradicts itself. |
| F45 | Lens B | Important | `skills/myflow-contracts/state-self-heal.md:87` | The file restates the `medium`/`high`/`low` mapping to make its `stage`-vs-`effort` argument, then twenty lines later claims the mapping is "stated once" elsewhere — false of this very file. The argument needs only "renamed 1:1", not the literal pairs. |
| F46 | Primary | Important | `openspec/changes/kan-26-operator-output-and-configuration/proposal.md:99` | The Impact bullet still reads "none in flight records a non-null effort, so the rename affects no change currently open" — **the exact claim disproved at pass 3**, and the one the controller relayed to the operator. `design.md` was corrected; this was never touched by any fix round, and it archives verbatim at finish. |
| F47 | Adversarial | Important | `openspec/changes/kan-26-operator-output-and-configuration/tasks.md:568` | Task 7's jq block omits the `// .effort` fallback restored in round 3, under a provenance tag that explicitly certifies "the jq expression itself is exact". Its companion grep at `:597` expects no output but now returns two hits. A ticked plan step, tagged exact, that no longer matches what the diff delivers. |
| F48 | Lens C | Minor | `skills/myflow-contracts/state-self-heal.md:43` | The worked example of the correction announcement was not updated to demonstrate the `worktrees` clause the paragraph beneath it now mandates — the one template an implementer is most likely to copy literally. |

### Why the run stops here rather than opening fix round 4

The escalation ladder allows five rounds; this would be the fourth. It is being stopped early and
deliberately, and the reasoning is recorded so the operator can overrule it:

1. **The finding count is no longer converging** — 21, 10, 8, 9.
2. **One defect class accounts for the Criticals in every pass**: a mechanism promised in prose and
   absent from the place that executes it (F1 → F23 → F33 → F41). F41 is that class occurring
   *inside the mechanism built to stop it*, which is the signal that another round of the same
   treatment is unlikely to end it.
3. **Two findings this pass (F46, F47) are stale false claims in the planning artifacts**, both
   left by earlier rounds' sweeps. Those artifacts archive verbatim, so each round that edits
   contracts without re-sweeping them adds durable, wrong text.
4. **F40 is a controller failure, not a fixer failure.** A Critical was dropped from this record and
   went unfixed for a full round. That is a reason to hand the record to a human, not to dispatch
   another agent against it.


## Pass 5 — full roster, escalated automatically

**Mode:** full. Escalated **automatically, without asking**, on the standing trigger *three or more
fix rounds have already run* — this pass follows fix round 4, the fourth. Diff rewritten from the
merge base to 1724 changed lines across 15 files, with the three planning paths excluded.

`[PRINCIPLES_PATH]` resolved to `/Users/tweety53/.claude/skills/myflow-do/engineering-principles.md`
(verified present, passed as a path, never pasted). `[STANDARDS_PATHS]` resolved from
`.myflow/project.md`'s `## standards`: `CLAUDE.md` and `AGENTS.md`, both bare filenames containing no
`/` (form 2), each verified to exist with its normalized parent exactly the project root. **No
standards entry was dropped.** The pass-1 harness deviation still applies: this harness exposes no
`bugbot` or `security-review` agent type, so slots 1 and 3 ran the prompts the skill ships for that
purpose, and their model is recorded as `sonnet` rather than `unknown (agent-defined)`.

| Slot | Included | Model | Result |
|---|---|---|---|
| 0 Primary | required | sonnet | F56, F57 |
| 1 Bug hunt | required | sonnet | F53, F54, F55 |
| 2 Principles (Merged) | required | sonnet | F49 |
| 3 Security | selected | sonnet | F50, F51, F52 |
| 4 Adversarial | selected | sonnet | F53 (dup), F58, F59, F60 |
| 5 Principles lens B | selected | sonnet | clean |
| 5 Principles lens C | selected | sonnet | clean |

**All nine round-4 findings (F40–F48) verified closed by three slots independently** — Lens B, Lens C
and Bug hunt each re-checked them against the live file text rather than the round-4 diff's own
commentary, and the controller re-verified all nine separately. F47's upgraded `verified:` tag was
confirmed by byte-comparing the plan's `jq` block against `myflow-status/SKILL.md:43`.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F49 | Merged | Important | `skills/myflow-contracts/pipeline.md:43` | The stage table claims to be "the only copy in this repository" and that "no skill carries a second copy". `CLAUDE.md:107` and `AGENTS.md:155` each carry an ordered arrow-form stage sequence, and the copies have **drifted**: `state gate` and `document the fix` appear 0 times in either file while `pipeline.md`'s `/myflow-do` row lists both. Both files already apply the citation discipline twice in those same rows. |
| F50 | Security | Important | `skills/myflow-contracts/jira-integration.md:259` | The follow-up join search is never scoped to a Jira project. Every disclosure says "any project member", but no JQL project constraint exists anywhere — a repo-wide grep returns only two prose mentions. The real population is anyone who can create an issue in any project the connection can query. `design.md:298-313` built its risk-acceptance on the single-project premise. |
| F51 | Security | Important | `skills/myflow-contracts/jira-integration.md:309` | The round-4 title sanitisation folds control characters but not fence delimiters, which are ordinary printable ASCII. A title that is a bare backtick run closes the ` ```text ` block early; the template's own closing fence then opens a dangling block that swallows the `<n>` of `<m>` count — the one signal round 4 added to make forged evidence visible before the write — and the Yes/No options. A bypass of the F43 mitigation in its own rendering. |
| F52 | Security | Minor | `skills/myflow-contracts/jira-integration.md:313` | "Every control character" plausibly reads as category `Cc` only, leaving `Cf` — bidi overrides, zero-width joiners — uncovered in a paragraph that otherwise takes care over terminal escapes. |
| F53 | Bug hunt + Adversarial | Critical | `.superpowers/sdd/final-review-panel.md:246`, `docs/manual-test/kan-26-operator-output-and-configuration.md:115` | The panel record's marker block and the guide's `## Known incomplete` both still declared F40–F48 open after round 4 closed them. Reproduced against the live tree: `check-unfinished-work.sh` returned `OUTSTANDING: … 9 open finding(s)`. Raised independently by two slots. **This is the controller's own bookkeeping, not a defect in the diff** — both surfaces are refreshed by this run's remaining steps, which is what closes it. |
| F54 | Bug hunt | Minor | `skills/myflow-contracts/jira-integration.md:546` | The join outcome table has no row for a retry that appends only the newly-missing items and *then* fails the retitle: `joined — new items only` implies the later writes succeeded, `partially joined` is silent on whether the append was full. A reporting-string gap; the safe direction is preserved either way. |
| F55 | Bug hunt | Minor | `skills/myflow-contracts/state-self-heal.md:56` | The rebuild recovers `worktrees` keys by scanning "each affected repository", but for a multi-repo change the only record of which repositories participate is the `worktrees` key set — the very thing that is unparseable on the path triggering a rebuild. Net-neutral against the old behaviour, so not a regression, but now load-bearing for a feature this change adds. |
| F56 | Primary | Important | `skills/myflow-contracts/pipeline.md:565`, `:616`, `skills/myflow-status/SKILL.md:216` | `prUrl` is claimed to be the tiebreaker "only where the stronger signal is genuinely absent", and the one-way gap to apply "only to the inconclusive rows above". The table at `:568-575` also splits `not merged (proven)` by `prUrl`. The row is reachable: the *handle it manually* route completes run 1, commits, pushes and leaves `prUrl` null, after which an unmerged branch renders "Implementation staged — review and test" for work already committed and past the human gate. |
| F57 | Primary | Minor | `skills/README.md:30` | A pasted line was not rewrapped to the ~90-character width the rest of the file uses. Cosmetic. |
| F58 | Adversarial | Important | `skills/myflow-contracts/pipeline.md:467` | The `STARTED` template's `**Jira:**` line carries no `(run-only)` marker, though it reports "the transition made" — which the state file never records and which `/myflow-status` may not obtain, being forbidden to call Jira. Two of its three literal alternatives are structurally unreproducible by the second renderer, and the delta spec lists the Jira line among what a *regenerated* block carries while omitting it from the run-only subtractions. A SHALL-level requirement its mechanism cannot deliver. |
| F59 | Adversarial | Important | `scripts/check-vocabulary.sh` | The new `"effort":` literal matches too narrowly. Measured with the plan's own Task 11 Step 3 sandbox recipe: `"effort": null` is caught (exit 1), while `"effort" : null` and `'effort': null` both pass clean. Within the guard's documented limits, but every real JSON example in this repository uses the caught form, so a hand-typed regression in either variant slips through. |
| F60 | Adversarial | Minor | `skills/myflow-contracts/jira-integration.md` | The retitle fires even when the matched candidate is already titled `<same-KEY> follow-up` — the change joining its own prior follow-up — renaming a correctly-named issue for no reason. |

**Two controller errors in this pass, recorded because the run acted on them.**
1. **The fixer for round 5 was dispatched one slot early.** The controller stated "all 7 slots in"
   after six had reported; the adversarial slot was still running and returned F58–F60 afterwards.
   Corrected by sending the three findings to the running fixer rather than opening a sixth round.
2. **A false-negative probe was nearly reported as a catch.** The first F59 probe ran the guard in a
   sandbox missing its `DEFAULT_TARGETS`, so all three variants returned exit 2 — a missing-path hard
   error — which reads as non-zero and therefore as "caught". Re-run with the plan's own recipe, two
   of the three are genuine false negatives. Exit codes are checked against the guard's contract, not
   for mere non-zeroness.

## Fix round 5

Findings F49–F52 and F54–F60 dispatched as a single fix subagent on **opus**, per the union rule and
the model policy. F53 is the controller's own and is closed by this run's remaining steps, not by the
fixer. F58–F60 were relayed to the same subagent mid-flight rather than deferred to a sixth round.


## Pass 6 — full roster, escalated automatically

**Mode:** full, on the same standing trigger. Diff rewritten to 1894 changed lines across 15 files.
`[PRINCIPLES_PATH]` and `[STANDARDS_PATHS]` resolved exactly as in pass 5; no standards entry was
dropped. The pass-1 harness deviation still applies to slots 1 and 3.

| Slot | Model | Result |
|---|---|---|
| 0 Primary | sonnet | F62, F63 |
| 1 Bug hunt | sonnet | F67, F68 |
| 2 Principles (Merged) | sonnet | F61 |
| 3 Security | sonnet | F64, F65, F66 |
| 4 Adversarial | sonnet | F69, F70 + the bookkeeping recurrence |
| 5 Principles lens B | sonnet | clean |
| 6 Principles lens C | sonnet | clean |

**All twelve round-5 findings (F49–F60) verified closed**, several independently: Lens B grepped for
leftover arrow-form stage chains and confirmed `stateDiagram-v2` occurs exactly once; Merged and
Adversarial each re-probed the widened vocabulary guard in sandboxes, Adversarial also confirming the
two known-legitimate `effort` matches still pass clean; Security re-derived the title-sanitisation
ordering and could not construct a bypass. The controller verified each separately.

**Zero Critical findings against the diff.** The one Critical raised was the controller's bookkeeping,
below.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F61 | Merged | Minor | `skills/myflow-contracts/pipeline.md:709` | The tab-commands passage cites the progress-visibility passage as answering "the identical question", then restates the general claim almost verbatim — the pattern the same file argues against at `:601-604`. |
| F62 | Primary | Important | `commands/myflow-finish.md:14`, `commands-claude/myflow-finish.md:10` | Both still describe the unfinished-work gate's third course as "file a Jira task"; the skill and `pipeline.md` now read "File or join a Jira follow-up, then continue", joining being the usual outcome since F18. `README.md:383-385` states the rule: a command contradicting its skill is a defect. Task 9's sweep was `grep -rni 'effort'`, scoped to the rename, so no task ever looked at this wording. |
| F63 | Primary | Minor | `specs/myflow-handoff-output/spec.md:29` | The contents table claims to list "the template's fields minus its run-only ones" but omits `**Change:**`, which is neither run-only nor missing-capable. |
| F64 | Security | Important | `skills/myflow-contracts/jira-integration.md:270` | The mandatory JQL project clause is built from `## jira`, a PR-editable value that `project-configuration.md:28` constrains in no way, and the contract never states how the clause is assembled or quoted. The fallback path *is* pinned to `[A-Z]{2,10}`. A crafted value carrying JQL syntax widens the very scoping F50 established — F50's failure mode reopened through the one input round 5 left unconstrained. |
| F65 | Security | Minor | `skills/myflow-contracts/jira-integration.md:332` | The `Cc`/`Cf` fold does not cover `Zl`/`Zp` (U+2028, U+2029), so the load-bearing "step 1 guarantees the title is exactly one line" is not fully justified for a renderer honouring them as hard breaks. Does not reopen the fence-escape vector. |
| F66 | Security | Minor | `openspec/changes/kan-26-operator-output-and-configuration/design.md` | No decision entry records the mandatory JQL project clause; the existing entries still read as though the project-wide scan was implemented behaviour rather than a gap just closed. This file archives verbatim and has drifted this way before (F34, F46). |
| F67 | Bug hunt | Important | `skills/myflow-contracts/state-self-heal.md:31` | The carry-forward list names `planningEffort` among fields "re-emitted exactly as read", without the non-byte-copy caveat `state-file.md:196-202` makes prominent and that `/myflow-do` and `/myflow-finish` each received. An `effort`-keyed file routed through self-heal for an *unrelated* contradiction could be re-emitted unmigrated, or worse have `planningEffort` written `null` — the erasure the pass-3 handback restored the compatibility apparatus to prevent. |
| F68 | Bug hunt | Minor | `scripts/check-vocabulary.sh:265` | The widened pattern matches quoted-word-plus-colon with no structural JSON anchor. Measured: `Two things determine 'effort': the level and the model.` trips it, exit 1. Dormant — nothing in the tree hits it today — but by the guard's own recorded reasoning for excluding `medium` and `high`, a pattern that can hit truthful prose invites a `vocab-guard:allow` marker on a line telling the truth. |
| F69 | Adversarial | Important | `skills/myflow-info/SKILL.md:47` | The file hardcodes the three-line pipeline shape while adding a guardrail at `:80-82` reading "**Never** present a remembered diagram or stage table". The frozen copy has already drifted: it reads `terminal (it integrates on its first run)` where `pipeline.md:20` reads `terminal (second run — see the finish contract)`. A live self-contradiction inside the one command whose mandate is never to answer from memory. |
| F70 | Adversarial | Minor | `CLAUDE.md:66`, `AGENTS.md:112`, `skills/README.md:7` | Three of four non-skill entry points keep the three-line digest that `README.md`'s equivalent was replaced with a citation for. Asymmetric treatment, worth a deliberate note or a follow-up. |

**The controller's bookkeeping went stale a second time, and was caught by two slots.** When the pass-5
record was written, F49–F60 were genuinely open — round 5 had not run. After it landed, the markers
were never flipped, and the guide's `## Known incomplete` was deliberately deferred to after this
pass. The pass-6 dispatch briefs nevertheless told reviewers the F53 bookkeeping was "now fixed — the
panel record and guide were resynced", which was true only of the F40–F48 markers. Bug hunt and
Adversarial both reproduced the consequence with `check-unfinished-work.sh`. The lesson is the one
F53 already taught and this run failed to apply: **the record is refreshed when the fix lands, not
when the run is ready to hand off**, because every consumer in between reads it as current. Deferring
it also fed a false premise to seven reviewers, which is the more expensive half.

## Fix round 6 — operator override of the ladder

The ladder allows five fix rounds; this is the sixth. At the pass-6 handback the operator was offered
four courses — fix the four Important and withdraw the six Minor, fix all ten and re-run the panel,
stop with all ten open, or fix all ten without a further pass — and chose **fix all ten, then re-run
the full panel**. The override is recorded here and in `proposal.md`; nothing about it was inferred.

Dispatched as a single fix subagent on **opus**, per the union rule and the model policy.

**Two findings were not closed the way they were written, and the reasoning is recorded rather than
folded into a claim of closure.**

- **F70 turned out to be two classes, not one.** `CLAUDE.md` and `AGENTS.md` are copied into a project
  root by `setup.sh` and auto-loaded at session start — the same layer as the always-on rule, whose
  own text says it carries *only the trigger* and not the pipeline. Deleting their digest would remove
  the one thing that layer exists to hold, so it **stays**, with the reason now stated in each file,
  and all three copies were brought byte-identical to the rule's trigger block (they previously read
  `terminal (it integrates on its first run)`). `skills/README.md` is read on demand beside the
  contract, exactly like `README.md`, so it got the round-5 treatment: digest deleted, citation added.
  The symmetry the finding assumed would have been the wrong fix for two of the four files.
- **F68 was closed by tightening the guard's anchor, which narrows it by one case, disclosed rather
  than hidden.** A quoted key mid-prose whose value is *also* unquoted — `the state file's "effort":
  low entry` — is no longer caught. Verified independently: all five realistic reintroduction
  spellings still trip the guard, the measured prose false positive no longer does, and the disclosed
  gap is exactly the case named. This is a judgment call the panel should test: the Lint Fix Priority
  rule forbids narrowing a guard to make a check pass, and the defence is that this narrowing removes
  a false positive rather than admitting a real hit — the guard's own recorded reasoning for excluding
  `checkpoint`, `medium` and `high` is that a pattern able to fire on a truthful line invites a lying
  suppression marker.

**One out-of-scope drift was reported, not fixed:** `commands/myflow-info.md:23` and its
`commands-claude` twin describe the input as a change name, while `skills/myflow-info/SKILL.md:81`
says the only argument is an optional *topic*. Outside F61–F70; carried here so it is not lost.

**A residual divergence pass 7 should judge:** the trigger-block copies now read
`terminal (second run — it integrates first)` while `pipeline.md`'s own States block reads
`terminal (second run — see the finish contract)`. Both are accurate, and a dangling citation would be
useless in a file that deliberately does not load `pipeline.md` — but the two renderings are not
identical and nothing records that they are allowed to differ.

## Pass 7 — full roster

**Three slots returned clean** (primary, Lens B, Lens C), and the primary answered **"Ready for the
human gate? Yes"**. Seven findings after deduplication, **zero Critical**, and the trend has turned:
12 → 10 → 7 across passes 5–7.

**Two adjudications the panel was asked for, both delivered.** Merged ruled round 6's guard tightening
a **permitted tightening, not a forbidden weakening** — it was not made to unblock a red check, its
true-positive coverage was verified preserved, it applies the guard's own documented rationale for
excluding `checkpoint`/`medium`/`high`, its cost is disclosed, and its one new `vocab-guard:allow` line
is a pattern split rather than a new suppression. Merged and Adversarial both upheld round 6's **F70
"two classes" reasoning**: the trigger files are auto-loaded before any command can read `pipeline.md`,
so a citation there would be unfollowable exactly when needed.

**Two pairs were deduplicated by theme**, per the union rule: Adversarial's trigger-digest finding is
the sharper form of Merged's and is recorded as F71; its guide-wording finding is the consequence of
F77 and is recorded there.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F71 | Merged + Adversarial | Important | `CLAUDE.md:73`, `AGENTS.md:119`, `rules/myflow-manual-review.mdc:14` vs `skills/myflow-contracts/pipeline.md:20` | The three trigger copies are byte-identical to each other and assert they are "kept identical", but read `terminal (second run — it integrates first)` where the canonical block reads `terminal (second run — see the finish contract)`, and nothing records that they may differ. Sharper than a wording nit because F69 — the same round — made `/myflow-info` read that block **live**: an operator now sees both renderings in one session, from the session-start file and from the command whose mandate is never to answer from memory. |
| F72 | Primary | Minor | `specs/myflow-handoff-output/spec.md:45` | The F63 edit left a 125-character line against the file's ~90–100 wrap. Same class as the already-fixed F57. |
| F73 | Primary | Minor | `docs/manual-test/kan-26-operator-output-and-configuration.md:135` | The controller's carried-forward note cites `commands/myflow-info.md:23`; the line is 18, and 14 in the `commands-claude` twin. Transcribed from the fixer's report without checking. Inert — not a citation `check-references.sh` validates. |
| F74 | Security | Important | `skills/myflow-contracts/jira-integration.md:276` | The whole "no escaping needed" argument rests on `[A-Z]{2,10}` being a **whole-value** match, and the contract never says so. Three things make it a real gap: the same file uses *search* language for a near-identical pattern at `:28` ("Look … for `[A-Z]{2,10}-\d+`"); it demonstrably knows how to say otherwise at `:316` ("an exact match on two shapes, never a substring search"); and the fix's **own** counter-example `KAN" OR project != "KAN` begins with `KAN`, so it would survive a search reading. No tokenization rule is given for splitting the `## jira` body into candidate keys, where `## standards` gets an explicit normalization procedure. |
| F75 | Security | Minor | `skills/myflow-contracts/jira-integration.md:390` | The worked example justifying the step order does not hold. Step 2 collapses whitespace runs **to a single space, never zero**, so two backtick pairs separated by a space stay separated and cannot join into a run of four. The adopted order is independently sound — fold and collapse never touch delimiters and never delete a separator entirely, so they can only preserve or split adjacency, and truncation is suffix-only — but the illustration is false. The controller repeated this rationale to the operator without checking it against step 2's own semantics. |
| F76 | Bug hunt | Important | `scripts/check-vocabulary.sh:282` | Round 6's comma anchor reopens the false-positive class F68 was meant to close. Measured: `Three settings exist, "effort": low, medium, high are the names.` and `As for scope, "effort": that is what this section is about.` both trip the guard; the original reproducer is now clean. Dormant — zero occurrences in the tree — but this repository's comma-heavy prose is exactly the style that would hit it, and its own policy forbids silencing a truthful line with a marker. |
| F77 | Adversarial | Important | `scripts/check-vocabulary.sh:282`, `docs/manual-test/kan-26-operator-output-and-configuration.md:150` | Both pattern alternatives require the key to be **quoted**, so every unquoted spelling evades the guard — measured: `effort: low`, `- effort: low`, `` `effort: null` ``, `` `effort`: gone now. ``, and the exact sentence this diff deletes (`` Carry `artifactUrl`, `jiraIssue`, `effort` and `prUrl` forward verbatim. ``). The repository's own idiom for naming a field is the bare-backtick form, so a hand-typed reintroduction in house style is invisible. **This is not a round-6 regression** — the round-5 pattern required quotes too, so the coverage never existed; the defect is the **claim**. The controller wrote "every realistic reintroduction spelling still caught (verified independently)" into the guide after testing five *quoted* variants only. Fix the guard or narrow the claim, not both silently. |

**A third controller error, recorded like the first two.** F75 and F77 are both the controller
asserting more than it measured — repeating a fixer's ordering rationale without checking it against
the stated transform, and generalising a five-case quoted-only probe into "every realistic spelling".
That is the same overclaiming class this panel has been catching in the contracts since pass 1, and
it appeared in the run's own output twice in one pass.

## Fix round 7 — second operator override

The ladder allows five rounds; this is the seventh, authorised at the pass-7 handback. The operator
directed that the two vocabulary-guard findings be closed by **narrowing the guard's claim rather than
widening its pattern** — the entry had produced a finding in three consecutive passes (F59 → F68 →
F76/F77) because a fixed literal list was being asked to prove a rename complete, which it cannot.

**Two closures took a different shape than the finding proposed, and the reasoning is recorded.**

- **F71 was closed by disclosure, not alignment.** Making the four copies byte-identical would have
  made "kept identical" true *today* with **no mechanism keeping it true** — any later edit to
  `pipeline.md:20` silently falsifies it and no guard checks digest identity, which is the same defect
  class this round exists to close. The trigger files now state that the copy reproduces the states and
  transitions **but not the wording**, that the canonical block's third line points at the finish
  contract (a pointer that cannot resolve in a file read before it loads), that the two are
  deliberately **not** byte-identical, and that both renderings are expected in one session. That
  disclosure cannot be falsified by a future edit to either side. It preserves Lens B's point and
  removes the claim Merged and Adversarial objected to.
- **F76/F77 were closed by removing the comma anchor and correcting the claim in both places.** No
  widening, no suppression marker. The unquoted gap is **unchanged and now stated**: the comment and the
  guide both say plainly that an unquoted key is matched by neither alternative and never was. A
  widening was considered and declined for a stated reason — unquoted, `effort` is the ordinary English
  word, and every candidate pattern fires on truthful prose, which is exactly why `checkpoint` and the
  level values are excluded.

**Two defects were found while fixing, beyond the findings as written.**

1. **The false ordering illustration had propagated into the delta spec.**
   `specs/myflow-jira-projection/spec.md` carried it as requirement text plus a scenario whose THEN
   asserted behaviour that cannot occur. Both rewritten to the true invariant. **This is why pass 8 is
   a full re-run rather than the targeted one the operator was offered: the fix altered a delta spec,
   which is a standing automatic-escalation trigger.**
2. **A checklist item in the manual test guide was already false.** Bullet 45 tells the operator to run
   `grep -rn '"effort":' …` and expect no output; the guard's own comment carried the literal.
   Verified by the controller against both trees: **1 hit before round 7, 0 after**. The item would
   have failed the moment it was exercised, and had been passing on inspection rather than measurement
   since round 6.

**Controller verification of round 7** (independent, sandboxed, worktree never mutated): four positive
controls still caught; F76's two measured false positives now clean; F77's gap confirmed unchanged and
disclosed; `[A-Z]{2,10}` now qualified "in its entirety" with an explicit tokenization rule and the
counter-example walked through it; "kept identical" absent from all three trigger files; the delta
spec's false claim gone; three guards and seven harnesses exit 0.

## Pass 8 — full roster, escalated automatically

Escalated because round 7 altered two delta specs. **Nine findings after deduplication: one Critical,
four Important, four Minor.** Merged returned clean and ruled both requested judgments in round 7's
favour (the comma-anchor removal is a permitted correction; F71's disclosure-over-alignment is the
right close, since aligning would embed a dead pointer in files that load before `pipeline.md`).

**The Critical was introduced by round 7's own fix**, and is the third consecutive instance of a
*justification* rather than a *rule* being the defect — F75 was a false illustration, round 7 replaced
it, and the replacement fix for F74 introduced another.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F80 | Primary + Security + Bug hunt | **Critical** | `skills/myflow-contracts/jira-integration.md:280`, `:283-286` vs `:307-310` | `:280` names the attack — "a bare `OR` between two keys". `:283-286` is **per-candidate**: each failing key is dropped, and only "if **every** named key is dropped" does the section name none. Measured: `OR` fullmatches `[A-Z]{2,10}`, so it survives and is accepted as a project key — the named attack is not defeated by the mechanism named as defeating it. The worked example claiming "not one of which is uppercase letters and nothing else" is false. `:307-310` frames the same rule as all-or-nothing, contradicting `:283-286`. Twelve JQL keywords pass the shape test. **No injection is achievable** — a letters-only value cannot escape its quotes — so the failure is scope integrity, gated behind the join confirmation. Fourth round on this clause (F50 → F64 → F74 → F80). |
| F79 | Lens C + Adversarial | Important | `rules/myflow-manual-review.mdc` | F71's disclosure reached two of the three trigger files the finding named. Measured: `git diff HEAD -- rules/myflow-manual-review.mdc` is **empty across all seven rounds**. Adversarial ran `setup.sh global` into a throwaway `$HOME` and confirmed the **shipped** global managed block still carries the pre-fix, undisclosed text — that file is `setup.sh`'s source for the global `~/.claude/CLAUDE.md` and `~/.codex/AGENTS.md` blocks and is Cursor's **only** rule layer. The round-7 closure narrative in this record says "the trigger files now state…", which overclaims. |
| F81 | Bug hunt | Important | `scripts/check-vocabulary.sh:258-269`, `:301` | Alternative 2 prefix-matches `null` and `<` rather than matching a JSON value exactly, contradicting round 7's new claim that the guard matches a quoted key carrying a JSON value "and nothing more". Measured: `"effort": nullify the previous value`, `"effort": nullable in some configs` and `"effort": <5` all trip it. |
| F83 | Adversarial | Important | `specs/myflow-jira-projection/spec.md:50` | The delta spec's normative fold still names only `Cc` and `Cf`; the contract has named `Cc`, `Cf`, `Zl` and `Zp` since round 6's F65 fix. Round 7 rewrote adjacent sentences in the same paragraph without correcting this one. The spec archives verbatim, so an implementer following it alone reopens the line-separator gap. |
| F84 | Adversarial | Important | `skills/myflow-contracts/jira-integration.md:291-309` | A `## jira` body mixing the literal `none` with a real key (`KAN, none`) is undefined: `none` is described both as a per-candidate value compared before the shape test and as a whole-body alternative to keys. The two readings scope the search to different projects. Exactly the ambiguity class the paragraph was written to close. |
| F78 | Lens B | Minor | `CLAUDE.md:80-86`, `AGENTS.md:126-132` | The F71 disclosure runs five near-duplicated sentences where one or two would carry it. Its supporting example is weaker than stated — the controller and Merged both found `CLAUDE.md:99`'s use of "the finish contract" is a resolvable in-sentence citation, not a bare forward pointer — but the verbosity point stands. |
| F82 | Bug hunt | Minor | `scripts/check-vocabulary.sh:266-267` | Alternative 1's backtick anchor matches after **any** backtick, including one closing an unrelated span, not "the backtick that opens an inline-code span" as round 7's comment newly claims. Measured: `` `code` "effort": x `` trips it. |
| F85 | Adversarial | Minor | `tasks.md:1342` | The Step 5 fence is tagged `verified: … 24 payloads`; the table beneath lists 23. The measurements themselves reproduce exactly — an arithmetic slip in a `verified:` annotation. |
| F86 | Adversarial | Minor | `tasks.md:1367` | The Step 6 fence claims `openspec validate --strict` and a sandboxed installer run exited 0, but the block runs neither. `openspec validate --all --strict` exits **1** on a pre-existing untouched spec; the change-scoped form passes. The claim is plausible under the natural reading but not reproducible as written. |

**The trend, stated plainly for the operator's decision.** Findings per pass: 21, 10, 8, 9, 12, 10, 7,
9. Criticals: 2, 2, 1, 2, 0, 0, 0, **1**. Seven fix rounds have run against a ladder that allows five,
overridden twice. The contracts pass every mechanical check and three slots returned clean — but each
round's newly written *justification prose* has become the next round's finding, and this pass's
Critical was manufactured by the previous round's fix.

findings-total: 86

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
finding-status: F27 withdrawn — premise disproved by measurement: the cited heading exists at openspec/specs/myflow-model-policy/spec.md:11, and renaming it in a sandbox makes check-references.sh fail on pipeline.md:1176 and exit 1, which it cannot do for a path it skips; the reviewer conflated the delta spec with the live spec. Operator withdrew it at the handback.
finding-status: F28 fixed
finding-status: F29 fixed
finding-status: F30 fixed
finding-status: F31 withdrawn — removal itself reverted at the operator's handback after F32 showed the controller's supporting premise was false (four affected state files, not two, one of them this change's own open file) and F33 showed the replacement mechanism unimplemented; the compatibility read that proposal.md originally approved is restored.
finding-status: F32 fixed
finding-status: F33 fixed
finding-status: F34 fixed
finding-status: F35 fixed
finding-status: F36 fixed
finding-status: F37 fixed
finding-status: F38 fixed
finding-status: F39 fixed
finding-status: F40 fixed
finding-status: F41 fixed
finding-status: F42 fixed
finding-status: F43 fixed
finding-status: F44 fixed
finding-status: F45 fixed
finding-status: F46 fixed
finding-status: F47 fixed
finding-status: F48 fixed
finding-status: F49 fixed
finding-status: F50 fixed
finding-status: F51 fixed
finding-status: F52 fixed
finding-status: F53 fixed
finding-status: F54 fixed
finding-status: F55 fixed
finding-status: F56 fixed
finding-status: F57 fixed
finding-status: F58 fixed
finding-status: F59 fixed
finding-status: F60 fixed
finding-status: F61 fixed
finding-status: F62 fixed
finding-status: F63 fixed
finding-status: F64 fixed
finding-status: F65 fixed
finding-status: F66 fixed
finding-status: F67 fixed
finding-status: F68 fixed
finding-status: F69 fixed
finding-status: F70 fixed
finding-status: F71 fixed
finding-status: F72 fixed
finding-status: F73 fixed
finding-status: F74 fixed
finding-status: F75 fixed
finding-status: F76 fixed
finding-status: F77 fixed
finding-status: F78 open
finding-status: F79 open
finding-status: F80 open
finding-status: F81 open
finding-status: F82 open
finding-status: F83 open
finding-status: F84 open
finding-status: F85 open
finding-status: F86 open

- **The `prUrl` discriminator's existence** was judged by the Merged principles lens to be an
  honestly-disclosed tradeoff done right, not a defect: the gap is named plainly, the rejected
  alternative is recorded with its reason, and the cost is correctly bounded because both renderings
  end in the same next command. F9 is the narrower, real defect inside that accepted design — the
  block prints a command that cannot work and a field with no true option.
- **`default` appearing in four places** was judged acceptable by two lenses: it is one word that must
  address a human directly in a prompt, a fallback sentence and a canonical table, which is forced WET
  rather than a duplicated ruleset.
- **The vocabulary guard's new literal** was attacked by three slots and held: correctly anchored to
  the quote-and-colon shape, no false positives against the two protected matches, no false negatives
  in scope, and its positive control reproduced independently.
- **The deferred items** (partial `models` object, three-line text-shape drift, README command table)
  were re-examined and none was reclassified.
