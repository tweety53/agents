# Review panel — kan-108-cut-the-time-and-token-cost-of-a-myflow-do-run

## Pass 1

**Roster:** `light` (recorded default for this change; `/myflow-fast` records defaults rather than asking).
**Required slots:** Primary, Principles (Merged), Code review (low) — all dispatched on `models.reviewPanel` = sonnet.
**Diff read:** `.superpowers/sdd/final-review.diff`, whole branch against merge base `44e3faf`.
**Panel diff size:** `check-panel-diff-size.sh` measured **400** changed lines against a cap of **2000** — exit 0, under cap, no operator prompt.

**Optional slots.** All four triggers fired against `final-review.diff`: Security (path/file handling in the new guard; a config file changed), Adversarial (400 changed lines, over ~300), Lens B (400 changed lines, over ~200), Lens C (the guard's error-handling paths and the config/env edit). Under `light` these go to the operator as one multi-select prompt. **The prompt was put and no explicit answer was given, so the recommended answer was applied and all four were included** — recorded as *no explicit answer — recommended answer applied*, distinct from a slot the operator declined and from a slot whose trigger never fired.

**Substitution.** Slot 4 (Security) is normally dispatched by `subagent_type: security-review`. No such subagent type is available in this session, so the slot ran as a `general-purpose` reviewer on this repository's own `skills/myflow-do/security-reviewer-prompt.md`, dispatched on `models.reviewPanel` = sonnet. The slot was **not** dropped; the substitution is recorded here because an unavailable harness feature is not a way to weaken review. Its ledger entry names sonnet rather than `unknown (agent-defined)`, because this dispatcher named the model.

| Slot | Model | Result |
|------|-------|--------|
| 0 — Primary | sonnet | clean |
| 2 — Principles (Merged) | sonnet | 2 findings |
| 3 — Code review (low) | sonnet | clean |
| 4 — Security (substituted) | sonnet | 3 findings |
| 5 — Adversarial | sonnet | 10 findings |
| 6 — Lens B (simplicity & state) | sonnet | 1 finding |
| 6 — Lens C (robustness & ops) | sonnet | 5 findings |

### Pass 1 findings, unioned and deduped by file:line + theme

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Security, Principles, Lens C, Adversarial | Critical | `skills/myflow-do/SKILL.md:519` | the pre-dispatch step runs a command taken verbatim from subagent-authored record text, with no constraint, no cwd pin and no bound |
| F2 | Lens C | Critical | `scripts/check-panel-reproducers.sh:28` | `-r` is true for a directory, so a directory-shaped record path reaches exit 0 instead of exit 2 |
| F3 | Adversarial | Critical | `scripts/check-panel-reproducers.sh:45` | the well-formedness regex accepts a bare `none` with no reason |
| F4 | Security | Major | `scripts/check-panel-reproducers.sh:39` | no `--` before the record path, so a worktree path beginning with `-` makes every grep parse it as options and the guard exits 0 |
| F5 | Lens C | Major | `scripts/check-panel-reproducers.sh:22` | `set -uo pipefail` without `-e`, diverging from every sibling guard and enabling F2 |
| F6 | Security | Minor | `scripts/check-panel-reproducers.sh:51` | `reproducers-total` accepts unbounded digits, so a real mismatch is swallowed by an integer-comparison error |
| F7 | Lens C | Major | `.myflow/project.md:36` | the 18-command `## test` list measures 118.63s against a 120000ms default tool timeout |
| F8 | Principles, Adversarial | Major | `skills/myflow-do/SKILL.md:533` | `none — not supplied by <slot>` is accepted from any slot, including a general-purpose slot that was given the instruction |
| F9 | Adversarial | Major | `skills/myflow-do/SKILL.md:523` | bounce accounting keys off `F<n>`, so re-raising the same defect under a new identifier resets the one-bounce budget |
| F10 | Lens B | Major | `skills/myflow-do/SKILL.md:520` | the bounced-once state is persisted nowhere, so a resumed run can bounce the same finding again |
| F11 | Adversarial | Minor | `scripts/test-check-panel-reproducers.sh:1` | four coverage gaps: scattered markers, reversed identifier order, a leading-zero total, and case 9 not asserting the directory check's own message |
| F12 | Adversarial | Major | `skills/myflow-do/SKILL.md:499` | claims replacing `a public contract` with `a guard's behaviour` narrows coverage, and asks for both |



**Reproducers were verified before the fix round was dispatched**, which is the rule this change introduces, applied to its own review. Four of the seven runnable reproducers demonstrated a real defect at exit 0. **Adversarial's reproducers for scattered markers, reversed identifier order and a leading-zero total demonstrated the opposite of what they claimed** — the guard already rejects all three correctly — so those were re-scoped from Critical/Major behaviour defects to a single Minor coverage finding (F11), and the severities in the table are the parent's after verification, not the raising slot's.

**F12 is disputed and goes to the operator rather than to the fix subagent.** KAN-108's own text says of this clause: "Either drop the trigger here or reword it to 'altered a delta spec or a guard's behaviour'." Restoring `a public contract` alongside the new clause reinstates precisely the non-discriminating condition the change exists to remove, so implementing it would undo the ticket. The gap Adversarial names is real — a contract with no guard behind it, such as the state file's documented fields or the panel record's own format, no longer escalates on its own — but the remedy is the operator's to choose.

## Pass 2 — Full escalation

**Why Full, automatically:** fix round 1's diff measured **301** changed lines, over the ~150 threshold. No operator was asked, per the ladder's own rule.
**Conditional slots re-run:** all four. Each trigger still fired against `fix-round-1.diff` — Security (the guard's argument handling and path canonicalisation changed), Adversarial (301 changed lines, and the harness itself was modified), Lens B (301 changed lines), Lens C (the exit-code paths and `set -e` discipline changed). None qualified for *not re-run — subject unchanged*.
**Diff read:** rewritten `.superpowers/sdd/final-review.diff`, 637 changed lines, measured under the 2000 cap.

| Slot | Model | Result |
|------|-------|--------|
| 0 — Primary | sonnet | 1 finding |
| 2 — Principles (Merged) | sonnet | clean |
| 3 — Code review (low) | sonnet | 6 findings |
| 4 — Security (substituted) | sonnet | 2 findings |
| 5 — Adversarial | sonnet | 5 findings |
| 6 — Lens B | sonnet | clean |
| 6 — Lens C | sonnet | 2 findings |

Every pass-1 fix was independently confirmed to hold, by four separate slots. Two slots re-raised the withdrawn F12; it stays withdrawn on the operator's recorded reason and is not reopened.

### Pass 2 findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F13 | Code review (low) | Major | `scripts/check-panel-reproducers.sh:99` | the guard accepts a reproducer line carrying shell metacharacters, a leading `-`, or a URL at exit 0 — the injection barrier is prose with nothing mechanical behind it |
| F14 | Code review (low) | Major | `scripts/test-check-panel-reproducers.sh:10` | the harness runs on `set -uo pipefail`, so a failed `mktemp` reuses the previous case's path and the suite can still report all cases passing |
| F15 | Primary, Adversarial, Code review (low) | Major | `skills/myflow-do/SKILL.md:585` | the bounce is said to be recorded in the pass log, but the pass log entry's own definition was never amended to carry a bounce or a defect identity, so the pointer resolves to nothing |
| F16 | Adversarial, Code review (low) | Major | `skills/myflow-do/SKILL.md:550` | *theme*, half of the bounce identity key, is defined nowhere, so a reworded theme or a line number nudged by one resets the one-bounce budget |
| F17 | Security | Major | `skills/myflow-do/SKILL.md:520` | only the reproducer's path token is containment-checked; its arguments are not, so a contained script can be handed `../../../etc/passwd` |
| F18 | Lens C | Major | `skills/myflow-do/SKILL.md:532` | the bound is named by citation with no concrete seconds, and a reproducer killed at the bound exits non-zero, which the rule reads as *defect demonstrated* |
| F19 | Adversarial | Major | `skills/myflow-do/SKILL.md:539` | the symmetric re-run checks that the reproducer flips fail to pass, never that the production code is what changed to flip it |
| F20 | Adversarial | Minor | `scripts/check-panel-reproducers.sh:169` | swapping the two `comm` directions mislabels every finding and still passes all 20 cases, because the assertions match on the identifier substring rather than the message |
| F21 | Adversarial | Minor | `scripts/check-panel-reproducers.sh:116` | no case exercises a command beginning with the literal `none`, so the word-boundary in the exemption pattern is unlocked |
| F22 | Adversarial, Security | Minor | `skills/myflow-do/SKILL.md:526` | *a leading `-`* does not say whether it binds the path token or every token, and read literally it refuses the ordinary `script.sh --strict` case; the ban list also omits `~ * ? [ ] #`, and the prose never says whether the run is a shell or a direct exec |
| F23 | Lens C | Minor | `scripts/check-panel-reproducers.sh:36` | the worktree is re-resolved after its own existence check, so a worktree that vanishes between the two exits 1 with bash's own message instead of the documented exit 2 |
| F24 | Code review (low) | Minor | `.myflow/project.md:59` | the runtime note hardcodes *eighteen commands*, which this file's own policy forbids — a written count went stale the first time a guard was added |
| F25 | Code review (low) | Major | `scripts/check-panel-reproducers.sh:91` | the marker machinery is a near-verbatim second copy of `check-unfinished-work.sh`'s, and has already drifted: this guard bounds the total's digits, the sibling does not, so the 26-digit crash this change fixed here still exists there |

**F25 is not fixed in this change, and the reason is recorded rather than left to look like an oversight.** Deduplicating the two guards' marker machinery means editing `check-unfinished-work.sh`, which this change is constrained not to touch, and the sibling's unbounded `findings-total` regex at `check-unfinished-work.sh:347` is a pre-existing defect this change neither introduced nor widened. Both belong in a follow-up issue filed at `/myflow-finish` run 1. The drift the finding names is real and is why the follow-up is filed rather than the finding dismissed.

**F13, F14, F24 and the sibling-drift half of F25 were verified by the parent before dispatch**, by running the guard and reading the file. F13's three fixtures — a metacharacter chain, a bare `-rf`, and a URL — each returned `REPRODUCERS-OK` at exit 0.

## Pass 3 — Full escalation

**Why Full, automatically:** fix round 2's diff measured **313** changed lines, over the ~150 threshold. All four conditional triggers still fired against `fix-round-2.diff`.
**Diff read:** rewritten `.superpowers/sdd/final-review.diff`, 870 changed lines, under the 2000 cap.

| Slot | Model | Result |
|------|-------|--------|
| 0 — Primary | sonnet | 1 finding |
| 2 — Principles (Merged) | sonnet | 1 finding |
| 3 — Code review (low) | sonnet | 1 finding |
| 4 — Security (substituted) | sonnet | 5 findings |
| 5 — Adversarial | sonnet | 3 findings |
| 6 — Lens B | sonnet | 4 findings |
| 6 — Lens C | sonnet | 1 finding |

No Critical this pass, and every pass-1 and pass-2 fix was re-confirmed by at least two slots. Three slots converged independently on the same false-parity sentence, which is the strongest signal in the pass.

### Pass 3 findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F26 | Primary, Principles, Lens B | Major | `skills/myflow-do/SKILL.md:551` | claims the 20s bound and 2s grace are "the same two numbers" as `check-cleanup-complete.sh`'s, but that file sets `SURVIVORS_TIMEOUT=60`; only the grace matches |
| F27 | Lens B | Important | `skills/myflow-do/SKILL.md:529` | the constraint and defect-identity paragraphs restate the delta spec's normative text closely, and account for most of the file's 10038-byte growth |
| F28 | Lens B | Minor | `skills/myflow-do/SKILL.md:544` | the "no newline" ban is enforced nowhere and is unreachable anyway — a grep-matched marker line cannot carry an embedded newline |
| F29 | Lens B | Minor | `scripts/check-panel-reproducers.sh:101` | `lines_of` and `full_lines_of` are near-duplicate bodies differing only in grep's `-o` flag |
| F30 | Security | Major | `skills/myflow-do/SKILL.md:530` | "never through a shell" is unearned: no argv-exec runner exists, and this harness's executor is a shell, so the metacharacter list is the only barrier rather than a belt |
| F31 | Security | Major | `scripts/check-panel-reproducers.sh` | per-token containment is prose-only — `/etc/passwd` and `../../../../../../etc/passwd` both reach exit 0 |
| F32 | Security | Minor | `scripts/check-panel-reproducers.sh:196` | a NUL byte in a reproducer line is silently dropped rather than rejected |
| F33 | Security | Minor | `scripts/check-panel-reproducers.sh:196` | a trailing backslash line-continuation is not banned, so a second line sits outside every check |
| F34 | Security | Major | `skills/myflow-do/SKILL.md:550` | a reproducer that double-forks or detaches escapes the bound's process-group kill, and the rule says nothing about detecting it or about the worktree state it leaves |
| F35 | Adversarial | Minor | `scripts/test-check-panel-reproducers.sh` | only 2 of 15 banned metacharacters are individually covered; shrinking the list to `\|;&` leaves all 25 cases passing |
| F36 | Adversarial | Minor | `skills/myflow-do/SKILL.md:511` | the theme definition has no worked derivation for a compound finding note, so two readers key the same defect differently |
| F37 | Adversarial | Minor | `skills/myflow-do/SKILL.md:598` | the diff-must-touch rule checks membership, not materiality — a comment or whitespace edit on a named path satisfies it |
| F38 | Lens C | Important | `skills/myflow-do/SKILL.md:525` | the caller's exit-1 prose describes only the missing-field class, so a rejected — possibly malicious — reproducer line reads as something to rewrite rather than escalate |
| F39 | Code review (low) | Minor | `scripts/check-panel-reproducers.sh:196` | a malformed `reproducers-total: 01` beside a well-formed one passes at exit 0 — the untested gap between cases 13 and 20 |

**F26, F31 and F39 were verified by the parent before dispatch.** `check-cleanup-complete.sh:529` reads `SURVIVORS_TIMEOUT=60`; an absolute path and a `..` traversal each returned `REPRODUCERS-OK` at exit 0; and the mixed-duplicate total fixture likewise returned exit 0.

**F27 is accepted in part and the rejected half is recorded here rather than left silent.** Its fix asks that `skills/myflow-do/SKILL.md` be reduced to a pointer at the delta spec. That is refused: a `/myflow-do` run loads the skill and never the live spec — the same reason **Model policy** (`skills/myflow-contracts/pipeline.md`) states that runtime reads its own section rather than the capability — so the operational rules must remain in the skill or the run has no instructions at all. What is accepted is the bloat: sentences that restate a rule the same file already states are trimmed.

## Pass 4 — Full escalation, curtailed by the operator

**Why Full:** fix round 3 exceeded ~150 changed lines and three or more fix rounds had run — two independent triggers.

**The operator stopped three slots mid-flight: Adversarial, Lens B and Lens C.** They were dispatched and killed before reporting. This is recorded as **stopped by the operator**, which is distinct from *declined*, from *not re-run — subject unchanged*, and from a trigger that never fired. Their most recent results are from **pass 3**, and the branch changed after that in fix round 3 and the trim, so **those three results are stale**.

**The handoff bar is therefore not met.** It requires zero open findings at any severity *and* a non-stale clean result from every slot in the roster. Three slots' results are stale by the definition this section uses, and the operator's instruction is the reason. Nothing about that is presented as clean.

**The operator also directed that panel composition is never the run's decision.** Every pass from here is dispatched only on an explicit answer, including the optional-slot prompt, which is no longer resolved by applying the recommended answer to silence.

| Slot | Model | Result |
|------|-------|--------|
| 0 — Primary | sonnet | clean |
| 2 — Principles (Merged) | sonnet | 1 finding (F41) |
| 3 — Code review (low) | sonnet | 1 finding (F42) |
| 4 — Security (substituted) | sonnet | 1 finding (F40) |
| 5 — Adversarial | — | **stopped by the operator** — pass-3 result stale |
| 6 — Lens B | — | **stopped by the operator** — pass-3 result stale |
| 6 — Lens C | — | **stopped by the operator** — pass-3 result stale |

### Pass 4 findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F40 | Security | Important | `skills/myflow-do/SKILL.md:552` | a detached or double-forked reproducer child escapes the process-group kill and is reported but never terminated, and the one-shot `git status` recheck is scoped to the killed case, so a surviving child can keep writing during the operator's pause |
| F41 | Principles | Important | `skills/myflow-do/SKILL-rationale.md:216` | the trim moved an operational rule into the rationale file, duplicating `SKILL.md` — no run loads that file, so a rule stated there does not execute — and left a pointer from the rationale file into itself |
| F42 | Code review (low) | Minor | `skills/myflow-do/SKILL.md:537` | the prose lists seventeen banned characters where the script bans eighteen; the backslash added for F33 was never documented |

**F42's direction matters and is recorded rather than left to be inferred:** the guard is *stricter* than the prose, so this was a documentation defect, never a hole.

All three were fixed in one round. `skills/myflow-do/SKILL.md` ended at **54843** bytes against its 58623 budget; `skills/myflow-do/SKILL-rationale.md` at **14122** against 15721, the F41 deletion having returned 731 bytes.

**One correction to a fix report, recorded because it nearly became a false lint claim.** The fix agent ran `scripts/check-vocabulary.sh .` and reported exit 1, dismissing the hits as pre-existing. `.myflow/project.md` declares that guard with **no argument**; in the declared form it exits 0. The hits came from a wider scan over archived artifacts the change never touched, so the conclusion was right and the invocation wrong — the declared form is the one a verification claim rests on.

## Pass 5 — required slots only, by the operator's choice

The operator directed that panel composition is never the run's own decision, so this pass ran the three required slots and no conditional ones. Adversarial, Lens B and Lens C keep their **stale pass-3 results**.

| Slot | Model | Result |
|------|-------|--------|
| 0 — Primary | sonnet | 1 finding (F43) |
| 2 — Principles (Merged) | sonnet | clean |
| 3 — Code review (low) | sonnet | 3 findings (F44, F45, F46) |

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F43 | Primary | Minor | `skills/myflow-do/SKILL.md:537` | the single-backtick code span closes early at the escaped backtick inside it, so half the character list renders as plain text — **predates F42**, which only added a second offender |
| F44 | Code review (low) | Major | `skills/myflow-do/SKILL.md:555` | the fix for F34 cited `ps -o sid=`, which does not exist on Darwin — `ps: sid: keyword not found` |
| F45 | Code review (low) | Major | `scripts/check-panel-reproducers.sh:265` | neither quote character was in the banned set, so an unbalanced quote passed at exit 0 and reached a shell |
| F46 | Code review (low) | Minor | `skills/myflow-do/SKILL.md:594` | the materiality clause may disqualify a correct single-file fix to a guard script, where the reproducer's target *is* the finding's named path; raised at low confidence by its own reporter |

**The operator chose to fix the two Majors only.** F43 and F46 are recorded not-fixed on that instruction: F43 is a rendering defect in documentation that predates this round, and F46 is speculative by its reporter's own account, with two earlier script-level fixes having proceeded fine under the same wording.

## Task 7 — the runner, added mid-run

**Eight findings across three passes — F17, F18, F19, F30, F34, F40, F44, F46 — were all defects in prose describing how to *run* a reproducer, with no code behind it.** A rule nothing implements can only be re-read, never tested, which is why each pass found another gap in the same paragraphs while the guard and its 35 cases went clean and stayed clean. F44 is the plainest evidence: a flag that does not exist on this platform survived three passes of review because nobody could run it.

The operator directed that the runner be built inside this change. `scripts/run-reproducer.sh` and `scripts/test-run-reproducer.sh` now enforce what the prose described, and `skills/myflow-do/SKILL.md` fell from 54849 to **52389 bytes** by citing an invocation instead of carrying a procedure. Scope was recorded in `proposal.md`, as task 7 in `tasks.md`, and as a comment on KAN-108 before any code was written.

**Two of task 7's own premises turned out wrong, and only running them showed it:** `ps -o sess=`, the natural replacement for the invalid `ps -o sid=`, always prints `0` here; and bare `cd && pwd` does not resolve symlinks, so with macOS's symlinked `$TMPDIR` every contained fixture was refused as an escape until `pwd -P`.

## Pass 6 — required slots over task 7

**The panel diff measured 2013 changed lines, over the 2000 cap.** `check-panel-diff-size.sh` exited 1, the operator was asked with named options, and chose **proceed with the panel anyway** — the contract's default and recommended answer. The measurement, the cap in force and that answer are recorded here, as the cap requires on every run. That this change trips the cap its own predecessor added is noted rather than smoothed over.

Slots: Primary, Principles (Merged), Code review (low), on sonnet. Conditional slots not run, by the operator's standing instruction; their pass-3 results remain stale.

## Pass 6 results, and rounds 4 and 5

| Slot | Result |
|------|--------|
| 0 — Primary | clean; independently reproduced both platform claims (`ps -o sid=` unknown, `ps -o sess=` always `0`, `$TMPDIR` a real symlink) |
| 2 — Principles (Merged) | 1 Important (F48) + 1 Minor note |
| 3 — Code review (low) | 5 findings (F49–F53) |

A **peer session's reuse review**, not a panel slot, contributed F47 and F48 independently; three of its five candidates restated F25 and were not counted twice.

### Findings F47–F56

| ID | Source | Severity | Location | Note |
|---|---|---|---|---|
| F47 | peer reuse review | Major | `scripts/check-panel-reproducers.sh:47` | canonicalised with logical `pwd` while the runner used `pwd -P` for the same variable, so the two disagreed on a symlinked worktree |
| F48 | Principles, peer | Important | both new scripts | the banned-character set duplicated byte-for-byte; **it had already drifted twice inside this change** — F33 added the backslash, F45 the quotes, each in one file first |
| F49 | Code review (low) | Minor | `scripts/run-reproducer.sh:104` | the path token split on space while the tokenizer split on space and tab, so a tab-separated command was refused as non-existent |
| F50 | Code review (low) | Major | `scripts/run-reproducer.sh` | the reproducer's output was captured and deleted unread, while the bounce rule requires carrying it back to the raising slot |
| F51 | Code review (low) | Major | `scripts/run-reproducer.sh` | the survivor sweep was gated on the timeout branch, so a reproducer that detached a child and exited fast leaked it |
| F52 | Code review (low) | Minor | `scripts/run-reproducer.sh` | a subshell plumbing failure was indistinguishable from the reproducer's own non-zero exit, so it read as *defect demonstrated* |
| F53 | Code review (low) | Important | `skills/myflow-do/SKILL-rationale.md` §7 | **the operator-directed trim tore sentences apart** — a paragraph promising a reason that never arrived, a sentence ending mid-clause at "the trade", and a blockquote whose first line lost its marker |
| F54 | Adversarial | Critical | `scripts/run-reproducer.sh` | `pgrep -P` walks one level, so a real double fork left a live process behind under a clean verdict |
| F55 | Adversarial | Important | `scripts/run-reproducer.sh` | the sentinel was a guessable `$TMPDIR` path the reproducer could delete, turning a demonstrated defect into *cannot answer* |
| F56 | Adversarial, parent | Minor | both scripts | a missing shared metachar file exited 1 in both, where the honest answer is *cannot answer* — 2 for the guard, 4 for the runner |

**F53 is the one finding this run's own process caused.** A trim the operator directed to reclaim budget headroom moved reasoning into the rationale file and shredded it, and the parent's verification missed it: the trimmer supplied a rule-presence list, that list covered `SKILL.md`, and nobody read the file the text landed in. `scripts/check-references.sh` passes on that file because it resolves headings, not sentences. Recorded here because the next trim will be tempted by the same evidence.

**F54 is fixed in part, and is recorded as partial rather than closed.** The walk is now a breadth-first sweep over `pgrep -P` at every depth, collected each poll and killed deepest-first, and a real double fork whose intermediate lives long enough to be observed is caught, named with its pid and reaches exit 3. What remains unclosed, verified by the parent after the fix:

```
scripts/run-reproducer.sh <wt> "df.sh"   →  rc=1, "defect not demonstrated"
grandchild 83971 STILL ALIVE
```

A grandchild whose intermediate exits within milliseconds re-parents to `launchd` before the first poll observes it, and no parentage-based technique can see it afterwards. `ps -o sid=` and `ps -o sess=` are both unusable here, which is what F44 established, so `pgrep -P` is the only tool available. The script documents this limit beside the code. The operator accepted the residual gap rather than adding a process-group wrapper or a snapshot diff to a script already near 400 lines.

**Fix rounds 4 and 5** closed F47–F53 and F54–F56 respectively, each fix mutation-checked in scratch copies. The shared definition is `scripts/reproducer-metachars.sh`, sourced by both scripts, with a harness case asserting neither carries its own literal. The sentinel is now an unlinked FIFO held open on a descriptor — `coproc` was tried first and rejected, because bash closes coproc descriptors in every forked child, so the subshell could never reach them. Harnesses stand at 15 cases for the runner and 37 for the guard, and the suite leaks no processes.

findings-total: 56
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
finding-status: F12 withdrawn operator's decision at the panel handback — restoring the clause reinstates the non-discriminating condition KAN-108 exists to remove, and the ticket's own text names rewording as the remedy; the unguarded-contract gap is accepted as the price
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
finding-status: F25 open
finding-status: F26 fixed
finding-status: F27 fixed
finding-status: F28 fixed
finding-status: F29 fixed
finding-status: F30 fixed
finding-status: F31 fixed
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
finding-status: F43 open
finding-status: F44 fixed
finding-status: F45 fixed
finding-status: F46 open
finding-status: F47 fixed
finding-status: F48 fixed
finding-status: F49 fixed
finding-status: F50 fixed
finding-status: F51 fixed
finding-status: F52 fixed
finding-status: F53 fixed
finding-status: F54 open
finding-status: F55 fixed
finding-status: F56 fixed

reproducers-total: 56
finding-reproducer: F1 none — prose defect; at pass 1 no constraint existed yet to violate mechanically
finding-reproducer: F2 none — verified as an ad-hoc record fixture: a directory-shaped record path returned REPRODUCERS-OK at exit 0
finding-reproducer: F3 none — verified as an ad-hoc record fixture: a bare `none` with no reason returned REPRODUCERS-OK at exit 0
finding-reproducer: F4 none — verified as an ad-hoc record fixture: a worktree path beginning with `-` returned REPRODUCERS-OK at exit 0
finding-reproducer: F5 none — general hardening; F2's fixture is the concrete instance this omission produced
finding-reproducer: F6 none — verified as an ad-hoc record fixture: a 26-digit total returned REPRODUCERS-OK at exit 0
finding-reproducer: F7 none — verified by measurement: the test list ran 118.63s against a 120s default
finding-reproducer: F8 none — prose defect; the guard has no slot-awareness to test against
finding-reproducer: F9 none — process gap; demonstrating it needs a re-identified finding across two rounds
finding-reproducer: F10 none — needs a session interrupted mid-loop, which no scriptable check reproduces
finding-reproducer: F11 none — the inputs behaved correctly; the defect was that no case locked the behaviour in
finding-reproducer: F12 none — withdrawn by the operator; the reason sits on its marker line
finding-reproducer: F13 scripts/check-panel-reproducers.sh
finding-reproducer: F14 scripts/test-check-panel-reproducers.sh
finding-reproducer: F15 none — prose defect; a pointer resolving to nothing has no runnable check
finding-reproducer: F16 none — prose defect; an undefined key cannot be tested until it is defined
finding-reproducer: F17 none — prose defect at the time; scripts/run-reproducer.sh is what makes this rule testable
finding-reproducer: F18 none — prose defect at the time; now enforced by scripts/run-reproducer.sh
finding-reproducer: F19 none — process gap across two rounds; no single command demonstrates it
finding-reproducer: F20 none — established by mutation in a scratch copy: swapping the comm directions left all 20 cases passing
finding-reproducer: F21 none — established by mutation in a scratch copy: stripping the word boundary left all 20 cases passing
finding-reproducer: F22 none — prose ambiguity at the time; now enforced by scripts/run-reproducer.sh
finding-reproducer: F23 none — needs a race window between two shell builtins; no script drives it deterministically
finding-reproducer: F24 .myflow/project.md
finding-reproducer: F25 scripts/check-unfinished-work.sh
finding-reproducer: F26 scripts/check-cleanup-complete.sh
finding-reproducer: F27 none — a prose-volume judgment; no command measures whether a sentence is a restatement
finding-reproducer: F28 none — the banned condition was unreachable, so nothing could demonstrate it
finding-reproducer: F29 none — a duplication judgment on two helper bodies; no command demonstrates it
finding-reproducer: F30 none — prose defect at the time; the runner is what makes the direct-exec claim true
finding-reproducer: F31 scripts/check-panel-reproducers.sh
finding-reproducer: F32 scripts/check-panel-reproducers.sh
finding-reproducer: F33 scripts/check-panel-reproducers.sh
finding-reproducer: F34 none — prose defect at the time; now enforced by scripts/run-reproducer.sh
finding-reproducer: F35 none — established by mutation in a scratch copy: shrinking the metacharacter list left all 25 cases passing
finding-reproducer: F36 none — prose ambiguity; nothing implements the key to test
finding-reproducer: F37 none — process gap; no command distinguishes a material edit from a cosmetic one
finding-reproducer: F38 none — prose contradiction between two paragraphs; no runnable check
finding-reproducer: F39 scripts/check-panel-reproducers.sh
finding-reproducer: F40 none — prose defect at the time; now enforced by scripts/run-reproducer.sh
finding-reproducer: F41 none — a docs coherence defect; no script exercises a pointer's target
finding-reproducer: F42 none — a prose-versus-source comparison; the guard was the stricter of the two
finding-reproducer: F43 none — a Markdown rendering defect; no script in this repository renders Markdown to check it
finding-reproducer: F44 none — verified by running it: `ps -o sid= -p $$` answers `ps: sid: keyword not found` on this platform
finding-reproducer: F45 scripts/check-panel-reproducers.sh
finding-reproducer: F46 none — raised at low confidence as a prose ambiguity; two earlier script-level fixes proceeded fine under the same wording
finding-reproducer: F47 none — a resolution inconsistency between two scripts; established by reading both canonicalisation lines
finding-reproducer: F48 scripts/test-check-panel-reproducers.sh
finding-reproducer: F49 scripts/test-run-reproducer.sh
finding-reproducer: F50 scripts/test-run-reproducer.sh
finding-reproducer: F51 scripts/test-run-reproducer.sh
finding-reproducer: F52 none — a plumbing failure the reporter could not force deterministically without instrumenting the script under test
finding-reproducer: F53 none — torn prose; no script in this repository checks that a sentence is whole
finding-reproducer: F54 scripts/test-run-reproducer.sh
finding-reproducer: F55 scripts/test-run-reproducer.sh
finding-reproducer: F56 scripts/test-run-reproducer.sh
