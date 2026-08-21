# Review panel — kan-77-sdd-ledger-canonical-path

**Roster:** `light` · **Panel model:** Sonnet (`models.reviewPanel`) · **Fix model:** Opus (`models.panelFix`)

## Pass 1

**Mode:** full roster (pass 1 always runs the full roster selected for the change).
**Diff:** `.superpowers/sdd/final-review.diff`, 916 lines, 7 files, from merge base `f09f2b1`.
**Diff size:** `check-panel-diff-size.sh` measured **1307**, cap not exceeded, exit 0. No operator
prompt was reached, and the cap moved nothing about the roster or the bar.
**Bounces:** none this pass.
**Unverifiable reproducers:** F1. `run-reproducer.sh` exited **2** — refusal: "the path token 'grep'
does not exist inside the worktree — never executed". Recorded unverifiable and put to the operator,
per the refusal disposition; **not** rewritten to satisfy the runner. The parent separately confirmed
the duplication by reading lines 384-390 against 405-411.

| Slot | Ran? | Model | Result |
|------|------|-------|--------|
| 0 — Primary (plan alignment + code quality) | yes | sonnet | clean |
| 2 — Principles (lens Merged) | yes | sonnet | 1 Minor |
| 3 — Code review (low) | yes | sonnet | 3 Important |
| 1 — Bugbot | no | — | not in the `light` roster |
| 4 — Security | no | — | trigger fired (path/file handling); **operator declined** |
| 5 — Adversarial | no | — | triggers fired (tests modified; >~300 lines); **operator declined** |
| 6 — Lens B (simplicity & state) | no | — | trigger fired (>~200 lines); **operator declined** |
| 6 — Lens C (robustness & ops) | no | — | trigger fired (error handling); **operator declined** |

The four optional slots were put to the operator as one multi-select prompt, per the `light` preset,
and declined. Recorded as **declined** — distinct from a trigger that never fired. Declining moves
nothing about the handoff bar: zero open findings at any severity still governs.

## Findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Minor | `scripts/check-task-commit-fields.py:357-412` | `resolve_folded_task`'s two branches duplicate the accumulate-then-`replace` widening step, so the widening rule is written twice |
| F2 | Code review (low) | Important | `scripts/check-task-commit-fields.py:393-412` | multi-partner fold: the green-side branch unions the red task's files but never a sibling green partner's, so a 3-task fold falsely reports the sibling's file as undeclared |
| F3 | Code review (low) | Important | `scripts/check-task-commit-fields.py:362-380` | multi-partner fold: the expected subject is taken from the first-listed partner only (`if index == 0`), with no check that every partner declares the same subject |
| F4 | Code review (low) | Important | `scripts/check-task-commit-fields.py:395-402` | the green-side branch never re-validates the red task's full partner set, so an invalid fold the red id rejects passes when checked via a green id |

F2, F3 and F4 were confirmed by the parent against the source before dispatch, and against
**The build-green tag** (`skills/myflow-contracts/build-green.md`), which states a `Squash-with:`
field names "one or more other tasks in the same plan (comma- or whitespace-separated dotted ids,
e.g. `2.1, 3.4`)" — so multi-partner is a supported shape, not a hypothetical one.

F4 contradicts a contract this change itself wrote: `resolve_folded_task`'s own docstring promises
"one verdict from either id", and the delta spec carries the scenario **Either task id gives the
same verdict**.

findings-total: 23
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 withdrawn — operator's decision: message text only, the verdict is correct, and a duplicate partner id is a malformed plan field rather than a guard defect
finding-status: F11 withdrawn — operator's decision: arguably the designed behaviour, since one verdict from either id implies one violation set from either id; changing it risks altering behaviour that is currently right
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 fixed
finding-status: F16 fixed
finding-status: F18 fixed
finding-status: F19 fixed
finding-status: F20 fixed
finding-status: F21 fixed
finding-status: F22 fixed
finding-status: F23 fixed
finding-status: F17 withdrawn — operator's decision: pre-existing, mirrors the documented per-task design of the Tests: field, and the slot flagged its own confidence as lower than its other findings

reproducers-total: 23
finding-reproducer: F1 grep -n allowed_collateral=task.allowed_collateral scripts/check-task-commit-fields.py
finding-reproducer: F2 none — a multi-partner fold needs a three-task git fixture that cannot be expressed as one relative-path argv command; the fix round adds the harness case instead
finding-reproducer: F3 none — same fixture constraint as F2; the fix round adds the harness case instead
finding-reproducer: F4 none — same fixture constraint as F2; the fix round adds the harness case instead
finding-reproducer: F5 none — the runner refused the slot's command (absolute path outside the worktree); confirmed instead by the parent reading check_commit_subject lines 519-521
finding-reproducer: F6 none — the slot reproduced it by direct execution against the module with a throwaway three-task fixture, which is not expressible as a shell-free in-worktree argv
finding-reproducer: F7 none — a structural observation about a future edit, not a runtime defect; cases 40-42 pin the current behaviour
finding-reproducer: F8 none — needs a throwaway repo and a multi-task tasks.md built with mkdir/git init/heredoc, which is not expressible as a bare argv call of a worktree executable; both slots reproduced it live in their own transcripts
finding-reproducer: F9 none — needs a throwaway repo and a three-task tasks.md; the slot reproduced it live and the parent reproduced it independently, confirming an undeclared file passes uncaught
finding-reproducer: F10 none — same fixture constraint; the slot reproduced it live with a red naming the same partner twice
finding-reproducer: F11 none — same fixture constraint; the slot reproduced it live with two mutually-red tasks
finding-reproducer: F12 none — a stale comment, read directly at scripts/preserve-session-records.sh lines 9-38
finding-reproducer: F13 none — a structural SSOT finding, not independently runnable
finding-reproducer: F14 none — needs a throwaway repo with a wrapped Squash-with field; both slots reproduced it live, in opposite directions
finding-reproducer: F15 none — needs a throwaway repo with three tasks and two commits; the slot reproduced it live, failing a valid unrelated commit
finding-reproducer: F16 none — same fixture constraint; the slot reproduced it live by querying the green partner named in malformed free text
finding-reproducer: F18 none — needs the guard pair copied to a directory lacking a lib/ sibling; the slot verified it directly, getting ModuleNotFoundError and exit 1 instead of a guarded exit 2
finding-reproducer: F19 none — needs a git-backed worktree with a two-Squash-with-line plan and a real commit; both the slot and the parent reproduced it directly, and the parent confirmed both halves on one fixture
finding-reproducer: F20 none — needs a task body with two Build: lines; the fix subagent found it and the parent confirmed both readings directly on one fixture
finding-reproducer: F21 none — needs a plan whose task body carries a tilde-fenced example field; the parent reproduced it directly, reading files=['example-only.txt'] where the real declaration was real.txt
finding-reproducer: F22 none — needs a task body carrying a bare tilde line; the slot reproduced it against a real commit and the parent reproduced it independently, reading files=[] and commit=None
finding-reproducer: F23 none — a comment-only finding, no reproducible behaviour
finding-reproducer: F17 none — traced by reading check_regression/check_baseline against their call site; not executed, and the slot flagged its own confidence as lower


## Fix round 1

**Fixer model:** Opus (`models.panelFix`), named explicitly.
**Fix base:** `2d3d77e` → folded to `c1b900b`, one commit per task preserved.
**Diff read:** `.superpowers/sdd/fix-round-1.diff`, 427 lines, 2 files.
**Assertions:** 68 → **80**, all passing. The harness diff is **purely additive** — 0 deletions, so
no assertion was removed or weakened, and the third mutation case (a removed assertion, which
mutation cannot cover) does not arise.
**Bounces:** none this round.

**Multi-partner subject semantics, settled by the fixer and stated here** — every partner a red
task names must declare the **same** `Commit:` subject; disagreement is a plan defect the guard
reports, naming the red task and printing every declared subject. Reason: build-green.md calls the
set one unit folding into one commit — singular commit over a plural set — so one subject. First-wins
was rejected because it reports a false mismatch whenever the real subject is a later-listed
partner's.

### Mutation proof — run by the parent, not the fixer

fix-mutation: scripts/check-task-commit-fields.py — dropped `group_files`/`group_collateral` from the green-side union, reverting it to red-task-only — case 36 failed (4 assertions: sibling's file reported as undeclared from both green ids)
fix-mutation: scripts/check-task-commit-fields.py — replaced the subject-agreement check with first-wins selection — case 37 failed (2 assertions: blamed the commit instead of the plan)
fix-mutation: scripts/check-task-commit-fields.py — dropped `violations += group_violations` from the green-side branch — cases 38 and 39 failed (3 assertions: an invalid fold passed via a green id)
fix-mutation: scripts/check-task-commit-fields.py — none — the `_fold_group` + `_widen` extraction is a refactor with no behaviour change; the three mutations above all run through both extracted functions, so the claim is exercised rather than merely asserted
fix-mutation: scripts/test-check-task-commit-fields.sh — none — added cases 36-39; adding an assertion is not a behaviour a mutation can revert
fix-mutations-total: 5

Each mutation altered one mechanism and was reverted before the next, so none could pass by
cross-contamination. The working tree was confirmed green (80/80) after the last revert.

### An index defect the fix round surfaced — parent's own, now repaired

The fixer reported that `git stash` and `git rebase --autosquash` both failed with "Entry … not
uptodate. Cannot merge." because the index held an **empty blob** (`e69de29`) for
`docs/superpowers/ledgers/2026-08-20-kan-102-citations-resolve-to-installed-paths.md` while the
working file was 39,148 bytes. Cause: the parent's own `git add -N` on that path, run so the panel
diff would show the task-5 rename rather than a bare deletion. Had it survived to finish run 1's
`git add -A`, **the rescued ledger would have been committed empty** — destroying the record this
change exists to preserve. Re-added properly; the index now holds the real blob
(`b004275`, 39,148 bytes). The fixer folded its commit with `git commit-tree` + `git reset --soft`
instead, touching neither index nor working tree, and reported the deviation rather than working
around it silently.

### Escalation — automatic, to Full

Two triggers fired, either of which escalates on its own:

- the fix **altered a guard's behaviour** (`scripts/check-task-commit-fields.py`);
- the fix **altered a delta spec** (`specs/myflow-task-commit-fields/spec.md`, recording the
  multi-partner union, subject-agreement and re-validation rules plus a new scenario).

So round 2 is a **Full** re-run: every required slot, against a rewritten `final-review.diff` — not
the targeted subset. No conditional slot re-runs: all four were declined by the operator at pass 1,
and a declined slot has no trigger to re-evaluate.


## Round 2 — Full re-run

**Mode:** Full, escalated automatically (the fix altered a guard's behaviour and a delta spec).
**Diff:** rewritten `.superpowers/sdd/final-review.diff`, 1160 lines, 965 changed, under cap
(`check-panel-diff-size.sh` exit 0).
**Bounces:** none this round.

| Slot | Result |
|------|--------|
| 0 — Primary | 1 Minor — **F5**, new |
| 2 — Principles | clean; pass-1 F1 confirmed resolved |
| 3 — Code review (low) | F1-F3 confirmed fixed; 1 Medium — **F6**, new, a regression from fix round 1 |

Conditional slots did not re-run: all four were declined by the operator at pass 1, so there is no
trigger to re-evaluate.

### F5 and F6 are two faces of one unsettled case

Fix round 1 settled *that* a fold's partners must agree on a subject. It never settled what an
**undeclared** subject means, and `partner.commit is None` now flows into the dedup at line 390-391
untouched:

| ID | Shape | Symptom | Origin |
|----|-------|---------|--------|
| F5 | red task's **sole** partner declares no `Commit:` | `subjects = [None]` → `subject = None` → `check_commit_subject` returns `[]` at line 520; the folded commit's subject is checked against nothing | **pre-existing** — the `commit is None → skip` branch predates this change. *(This record earlier said it "also sits at line 584"; the fix round corrected that — line 584 is `check_commit_scope`'s own `None` early return, a different check. Both are untouched either way.)* |
| F6 | **several** partners, one declaring no `Commit:` | `None` is counted as a distinct subject, so a fold with exactly one real declared subject false-fails as a disagreement | **regression from fix round 1** |

Both were confirmed by the parent: F5 by reading `check_commit_subject:519-521`, F6 from the slot's
own direct execution against the module.

**Both reproducers are unverifiable**, and neither was rewritten to satisfy the runner. F5's was
refused as an absolute path outside the worktree; F6's was never expressible as a shell-free
in-worktree argv. `run-reproducer.sh` refused F5's when tried, exactly as it refused F1's in pass 1.

### Semantics settled for fix round 2

A partner that declares **no** `Commit:` field contributes **no** subject:

- it is skipped when collecting subjects, so it can never manufacture a disagreement (closes F6);
- if **no** partner in the fold declared a subject, that is a violation naming the red task — a fold
  whose surviving commit's subject is checked against nothing is not a fold the guard can vouch for
  (closes F5).

Scoped to the fold path only. The ordinary-task branch at line 584 keeps `Commit:`-is-optional
behaviour unchanged — tightening that reaches every task in every plan and is not this change's
subject. `myflow-task-commit-fields`'s **Every task declares mechanically-checkable fields** already
says every task SHALL carry `Commit:`; that the guard does not enforce it for ordinary tasks is a
separate, pre-existing gap, recorded here and not fixed.


## Fix round 2

**Fixer model:** Opus (`models.panelFix`), named explicitly.
**Base:** `c1b900b` → `399d98d`. Commit carries only the two scripts; no planning path in it.
**Assertions:** 80 → **94**. 11 of the 14 new ones failed against the pre-fix guard; case 43 and two
rc-only assertions pass either way, by design, as the untouched-path pins.
**Bounces:** none this round.

**F5 fixed. F6 fixed.**

### The fixer widened F5, correctly

F5 as filed described only "the red task's **sole** partner declares no `Commit:`". The same hole
exists in a **multi**-partner fold where one partner declares a subject and another does not: the
non-declaring partner's own id still resolved to `None`, so that id passed a subject the red id
rejected — two verdicts from one fold, which is the contract this change wrote. `_fold_group` alone
could not close it; `resolve_folded_task`'s partner branch had to stop using `task.commit`
unconditionally. Case 41 pins that face.

### Mutation proof — run by the parent

fix-mutation: scripts/check-task-commit-fields.py — re-admitted `None` to the subject list (dropped the `is not None` guard) — case 40 failed (11 assertions; the false disagreement returned verbatim)
fix-mutation: scripts/check-task-commit-fields.py — disabled the no-subject-anywhere violation (`if not subjects and not violations` → `if False`) — case 42 failed (5 assertions)
fix-mutation: scripts/check-task-commit-fields.py — removed the partner branch's `if subject is None: subject = group_subject` — case 41 failed (2 assertions; a non-declaring partner's id passed a commit the red id rejects)
fix-mutation: scripts/check-task-commit-fields.py — none — module docstring and exit-1 list, documentation only
fix-mutation: scripts/test-check-task-commit-fields.sh — none — cases 40-43 added; adding an assertion is not a behaviour a mutation can revert
fix-mutation: scripts/check-task-commit-fields.sh — none — thin wrapper, untouched
fix-mutations-total: 6

Each altered one mechanism and was reverted before the next. Tree confirmed green (94/94) after the
last revert.

### A commit-mechanics trap the parent created, and the fixer caught

`git commit --fixup` commits the **whole index**, not only what the fixer staged. The parent had
`git add`-ed the rescued kan-102 ledger while repairing the empty-blob entry, leaving it staged — so
the fixup swept a planning path into a task commit, which the git-boundary rule forbids. The fixer
unwound it (`git reset --soft` + `git restore --staged` + a plain commit) and reported the deviation
rather than leaving it. **Second time this one file has caused a problem**: first as an empty blob
from `git add -N`, then as staged content swept by `--fixup`. Correct resting state, now verified:
deletion unstaged, new file untracked, swept by finish run 1's `git add -A`.

### Escalation — Full again

The fix altered the delta spec (`specs/myflow-task-commit-fields/spec.md` gained two paragraphs and
three scenarios). That trigger fires on its own, so round 3 is a **Full** re-run. Principles' round-2
clean result is stale against `399d98d` and does not carry.


## Round 3 — Full re-run

**Mode:** Full, escalated automatically (fix round 2 altered the delta spec). Principles' round-2
clean result was stale against `399d98d` and did not carry.
**Diff:** rewritten `.superpowers/sdd/final-review.diff`, 1939 lines, 1742 changed, under cap.
**Bounces:** none this round.

| Slot | Result |
|------|--------|
| 0 — Primary | F5 confirmed fixed **in both faces**; 1 Minor — **F8**, new |
| 2 — Principles | 1 Minor — **F7**, new; judged `_fold_group` still one coherent unit, no SRP finding |
| 3 — Code review (low) | F6 confirmed fixed; **independently found F8**, same file:line, same theme |

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F7 | Principles | Minor | `scripts/check-task-commit-fields.py:406-421` | the no-subject violation's correctness is inferred from `violations` being empty, not stated — a future check appending for an unrelated reason would silently suppress it |
| F8 | Primary + Code review (low) | Minor | `scripts/check-task-commit-fields.py:494-502` | a green task named by **two different** red tasks resolves its subject from the first red in document order, so the shared partner's id exits 0 on a commit both red ids reject |

**F8 was raised independently by two slots** and deduped to one identity — same `file:line`, same
theme (a partner shared across two folds resolving to one fold's subject). Both reproduced it live.
Primary's construction: red-1→{X, Y declaring subject A}, red-2→{X, Z declaring subject B}, one real
commit; ids 1 and 2 fail, id 3 (the shared partner X) exits 0.

**F8 is newly reachable through fix round 2's own subject-assignment line**, per slot 3 — the third
consecutive round in which a fix to this function opened something.

**F7 and Primary's reading are compatible, not contradictory.** Primary checked the same guard and
found it sound *today* — only missing-partner and red-partner violations can precede it, verified by
reading every append site. F7 is about what a future edit could break, not current behaviour.

### The `code-review` skill ran off-target, and the slot said so

Slot 3's invocation of the harness `code-review` skill returned three items against
`skills/myflow-start/SKILL.md` and `skills/myflow-contracts/jira-integration.md` — **neither file is
in this diff**, and the skill's own closing line said its target was "merged, superseded scratch".
The slot **refused to fold them into its findings** and reported the discrepancy instead of passing
them through. Recorded because a slot silently relaying another tool's out-of-scope output would put
findings about unrelated files into this change's fix round.

### Convergence — handed back to the operator

Two fix rounds, three review rounds; **each fix round has bred a new finding in this one function**:
round 1 → F6, round 2 → F8. The change's own subject — the ledger path, the rescue, the MISSING
outcome, artifact brevity — has been clean since pass 1 and has not been touched by any of it. Every
remaining finding lives in `check-task-commit-fields.py`'s fold resolution, which entered this change
as the operator-approved guard-gap fix.


## Fix round 3 — F7, F8

**Fixer:** Opus. `399d98d` → `99caa1b`. 94 → **113** assertions; 10 of 19 new ones failed pre-fix.
Commit verified to list only `scripts/` files. **Bounces:** none.

F8 fixed by resolving the transitive fold group (`_fold_reds`, grown to a fixed point) instead of
the first red in document order, with a new violation when two joined folds disagree, anchored on the
shared task. F7 fixed with an explicit `partner_error` flag gating both post-loop checks.

### Mutation proof — parent, round 3

fix-mutation: scripts/check-task-commit-fields.py — transitive fold group reduced to the first red — case 44 failed (10 assertions)
fix-mutation: scripts/check-task-commit-fields.py — dropped the joined-fold disagreement violation — case 44 failed
fix-mutation: scripts/check-task-commit-fields.py — ungated the disagreement check — **SURVIVED**, repaired in round 4 below
fix-mutation: scripts/check-task-commit-fields.py — no-subject check reverted from `partner_error` to `not violations` — survived, and is a genuine **equivalent** mutant: a disagreement leaves `subjects` non-empty, so the no-subject check cannot fire either way. No test can distinguish them; left alone deliberately.
fix-mutations-total: 4

## Fix round 4 — the surviving mutant, and the gap round 3 reported

Not a reviewer finding. Two items: the surviving mutation above, which this project repairs in the
round that finds it rather than raising as an `F<n>`; and the per-red no-subject rule that round 3's
own report flagged as its most likely next defect, which the operator chose to close.

**Fixer:** Opus. `99caa1b` → `0509f63`. 113 → **129** assertions; 7 of 16 new ones failed pre-fix.
Commit verified to list only `scripts/` files. **Bounces:** none.

- **Surviving mutant killed.** New case 47 — a red naming three partners, one absent and two green
  declaring different subjects — asserts the missing-partner message fires and the disagreement
  message does not. Confirmed by the parent: ungating the check now fails case 47, where before it
  broke nothing.
- **No-subject rule moved to the combined fold.** `_fold_group` no longer emits it;
  `resolve_folded_task` asks `len(subjects) != 1` once over the joined group and branches into either
  message, anchored on the shared task. A red whose own partners declare none now takes a joined
  red's subject.

### Mutation proof — parent, round 4

fix-mutation: scripts/check-task-commit-fields.py — dropped the combined-fold subject branch — case 42 failed (12 assertions)
fix-mutation: scripts/check-task-commit-fields.py — anchored the message on `fold_reds[0]` instead of the shared partner — case 48 failed
fix-mutation: scripts/check-task-commit-fields.py — narrowed the combined check to `len(subjects) > 1`, reverting the per-red behaviour — case 42 failed (5 assertions)
fix-mutation: scripts/check-task-commit-fields.py — none — `_fold_reds`, `_shared_partner`, `_widen`, `check_commit_subject`, `check_commit_scope` and the ordinary-task path were untouched this round
fix-mutation: scripts/test-check-task-commit-fields.sh — none — cases 47-48 added; adding an assertion is not a behaviour a mutation can revert
fix-mutations-total: 5

Two of the three mutations surface as an `IndexError` traceback rather than a violation — the
`subject = subjects[0]` the fixer itself ranked as its second risk. Correct as written, since every
`len(subjects) != 1` path returns above it; the proof lives in control flow rather than in a check.
The harness fails loudly either way, so the behaviour is covered.

### The fixer's own convergence judgment, recorded

Round 4 reports the function has reached the point where further change costs more than it buys:
fold resolution is a fixed-point group, one rule per group, one anchor rule shared by both messages,
and cases 30-48 pin the behaviour from every id. Its remaining ranked risks are message-accuracy
only — `_shared_partner` returns the first partner named twice, so a three-fold join anchors on
whichever appears twice first in document order: arbitrary but stable, and it changes no verdict.
Recorded rather than acted on.


## Round 5 — Full re-run, diff over cap

**Diff size:** `check-panel-diff-size.sh` measured **2242** against the cap of **2000**, exit 1.
Put to the operator per the contract; the operator chose **proceed with the panel anyway**.

**583 of the measured lines are an accounting artefact**, and the operator's answer was given knowing
it: the kan-102 ledger rename registers as a pure deletion because its addition half is untracked and
invisible to `git diff`. Genuinely reviewable content is roughly 1659 lines, under cap. This is the
untracked-file gap KAN-64 describes, showing up in the size guard rather than in the review itself.
The cap moved nothing about the roster, the slots, the escalation ladder or the zero-open-findings
bar.

**Mode:** Full. Three triggers stand: fix round 4 altered the delta spec, it altered a guard's
behaviour, and four fix rounds have run (the ladder escalates at three).


## Round 5 results

| Slot | Result |
|------|--------|
| 0 — Primary | **clean**; round-3 F8 verified fixed. Built a three-fold chain joined through two shared nodes: verdict identical from all 7 ids. Proved `subjects[0]` safe by full path-trace. |
| 2 — Principles | **clean**; round-3 F1 verified resolved. Traced a reverse-chain fixture needing 3 outer passes, proving `_fold_reds`' fixed point is load-bearing rather than speculative. |
| 3 — Code review (low) | **4 findings**, one of them high-severity and verdict-changing |

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F9 | Code review (low) | **Critical** | `scripts/check-task-commit-fields.py:154-158, 343-346` | `PARTNER_ID_RE.findall()` runs un-anchored over the raw `Squash-with:` value, so any digit run in free text becomes a partner id — admitting an unrelated task to the fold and letting its files pass `check_files` uncaught |
| F10 | Code review (low) | Minor | `scripts/check-task-commit-fields.py:501-514` | `squash_partners` is never de-duplicated, so `Squash-with: Task 2, 2` makes `_shared_partner` see a false repeat and anchor the message on the innocent partner. Verdict unaffected — message text only |
| F11 | Code review (low) | Minor | `scripts/check-task-commit-fields.py:559-609` | `_fold_group` violations from every red in a connected group are concatenated with no per-query filtering, so invoking for one id returns violations belonging to another red. Reads against the function's own "whichever id the guard was invoked for" framing |
| F12 | Code review (low) | Minor | `scripts/preserve-session-records.sh:34-38` | the header still says a `skipped:` line "is the only outcome that is silent, and it is the only one that means nothing was wrong" — false since this change's own `rescued:`/`MISSING:`, which are also silent, on stdout, exit 0, and mean something may be wrong |

### F9 — verified independently by the parent, not taken on faith

**Structural proof.** `check-task-build-green.py` gates the whole field before extracting ids:
`SQUASH_WITH_RE = ^\*\*Squash-with:\*\*\s+Task\s+(?P<partners>[\d.,\s]+)\s*$` — digits, dots,
commas and whitespace only, so free text fails the gate outright.
`check-task-commit-fields.py` has **no** such gate; it runs `PARTNER_ID_RE.findall()` over `FIELD_RE`'s
raw remainder. Its own comment claims parity — *"Deliberately the same shape
check-task-build-green.py extracts partner ids with"* — which is true of the inner token pattern and
**false of the gate**. That comment is why the gap survived review.

**Empirical proof.** Fixture: red task 1 declaring `a.txt` with `**Squash-with:** Task 3 (see step 2)`;
green task 2 declaring `unrelated.txt`; green task 3 declaring `b.txt`. Parsed partners:
`['3', '2']`. A commit touching all three files reports only the fixture's own `tasks.md` — **it never
flags `unrelated.txt`**, a file no legitimate member of the fold declares. Strip the fixture noise and
the guard exits 0 on a commit smuggling an undeclared file past it.

**Four review rounds missed this**, because every previous round probed the fold *resolution* and none
probed the field *parsing* that feeds it.


### F10 and F11 withdrawn by the operator

Put to the operator one option-set at a time, with F9 stated as non-negotiable. The operator chose
to fix F9 and F12 and withdraw F10 and F11, with the reasons recorded on their marker lines above.
Nothing in the run wrote `withdrawn` for itself: the fix subagents left every finding they disputed
as `open` with their argument in the note, exactly as the contract requires.


## Fix round 5 — F9, F12

**Fixer:** Opus. `0509f63` → `433c52b`. Commit lists only `scripts/` files.
`test-check-task-commit-fields.sh` 129 → **138**; 5 of 9 new assertions failed pre-fix.
`test-preserve-session-records.sh` 169 → 169 (F12 is comment-only). **Bounces:** none.

**F9 reproduced before fixing:** the reviewer's fixture printed `rc=0 out=` — a commit carrying an
undeclared file passing silently. Case 50's pre-fix output read `Squash-with: names Task 3, which
does not exist in this plan`, direct proof that the word "step 3" in prose had been read as a partner.

**A malformed `Squash-with:` is a reported violation, fail-closed** — not a no-partners result. The
fixer's reasons: `check-task-build-green.py` already fails a red task whose field it cannot parse;
no-partners would silently demote the red task to the ordinary path and produce a subject mismatch
blaming the commit instead of the plan; and the defect being fixed *is* a guard passing what it could
not check. Scoped to a red task that **declares** the field and yields no id — a red task with no
field at all is unchanged, which case 46's fixture depends on.

**The parity comment was the root cause, and it was fixed too.** The old comment claimed
`check-task-build-green.py` parity — true of the token pattern, false of the field gate. It now
states the gate *is* the parity point and names F9.

### Mutation proof — parent, round 5

fix-mutation: scripts/check-task-commit-fields.py — bypassed the gate, running `findall` over the raw field value (restoring the F9 defect) — case 49 failed (5 assertions), output `rc=0 out=`, the fail-open signature
fix-mutation: scripts/check-task-commit-fields.py — disabled the malformed-field violation — case 49 failed (2 assertions)
fix-mutation: scripts/preserve-session-records.sh — none — F12 changed a comment block only, no code
fix-mutation: scripts/test-check-task-commit-fields.sh — none — cases 49-50 added; adding an assertion is not a behaviour a mutation can revert
fix-mutations-total: 4

### The two guards now agree on the field's shape

Identical gate, identical id extraction, identical treatment of a red task whose field does not gate.
Two residual divergences, both pre-existing and recorded rather than fixed: build-green matches one
physical line where commit-fields joins continuations, so a `Squash-with:` wrapped across two lines
would mean different things to each (no plan in the corpus wraps it); and a green task's malformed
field is inert in both, deliberately. No other consumer parses the field — only these two guards; the
skills and contracts that mention it are prose.

### A defect in this change's own delta spec, found by the fixer

`myflow-finish-cleanup`'s spec said the ordinary skip "SHALL remain the only outcome that is silent"
— the same loose wording that seeded F12's stale header. It is false: `preserved:`, `rescued:` and
`MISSING:` all print to stdout and exit zero too. Corrected to say the skip is the only outcome
meaning **nothing was wrong and nothing need be done**, and to state explicitly that silence does not
distinguish them. Written by the parent, not a reviewer finding.

## Round 6 — Full re-run

**Diff size:** measured **2440** against the cap of **2000**, exit 1. The operator's round-5 answer —
proceed with the panel — is **carried forward** rather than re-asked: the question is identical, and
the growth since (2242 → 2440) is fix round 5's own harness cases. 583 lines remain the rename
artefact, so reviewable content is roughly 1857 lines. Recorded per the contract's requirement to log
the measured count, the cap and the answer on every run.

**Mode:** Full. Every trigger still stands: the last fix altered a guard's behaviour and a delta spec,
and five fix rounds have run against a ladder that escalates at three.

**Why this round exists:** no slot has reviewed `433c52b`. Rounds 5's clean results from Primary and
Principles predate the F9 fix and are stale; the handoff bar requires a non-stale clean result from
every slot in the roster.


## Round 6 results — five new findings, and a convergence problem

| Slot | Result |
|------|--------|
| 0 — Primary | 1 Important — **F14**, a regression from fix round 5 |
| 2 — Principles | 1 Important — **F13**, the structural root cause |
| 3 — Code review (low) | 1 High (**F15**), 1 Medium (**F16**), 1 Low (**F17**), plus a re-raise of withdrawn F10 |

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F13 | Principles | Important | `check-task-commit-fields.py:181` vs `check-task-build-green.py:146-148` | the `Squash-with:` grammar is now two independently-maintained regexes held in sync by a comment — the exact structure that produced F9 |
| F14 | Primary + Code review (low) | Important | `check-task-commit-fields.py:349` × `:181,373-379` | **regression from fix round 5**: the gate is applied to the value *after* continuation lines are joined, so a well-formed `Task 2` followed by an unblanked prose line now fails as malformed. Pre-fix this passed |
| F15 | Code review (low) | **High** | `check-task-commit-fields.py:506-541` | `_fold_reds` admits a named partner as a full member before validating it, so one red's bad `Squash-with:` reference pulls an unrelated, correctly-folded group into one membership set and **fails that group's valid commit** |
| F16 | Code review (low) | Medium | `check-task-commit-fields.py:527`, `:592-601` | a red with a malformed field is excluded from `reds`, so the malformed-field violation fires only when the red itself is queried; querying the partner named in its free text reports a misleading "file not declared" |
| F17 | Code review (low) | Low | `check-task-commit-fields.py:1032-1033` | `check_regression`/`check_baseline` take the unfolded `task` where `check_files`/`check_commit_subject` take `folded`. Mirrors `Tests:`'s documented per-task design, so likely deliberate — but no case exercises it on a folded pair. Pre-existing |

**F10 was re-raised and is NOT reopened.** The operator withdrew it in round 5 with a stated reason;
a withdrawn finding stays withdrawn. Recorded here so the re-raise is visible rather than silently
dropped.

**F14's two faces are one identity.** Primary found the gate rejecting a valid wrapped field;
Code review found the same wrapped field *accepted* here and *rejected* by build-green. Same root —
build-green matches one physical line, this guard joins continuations first — deduped to one finding.

**F13, F14, F15 and F16 are one defect family**, and F13 names it: two guards parse one field with two
grammars and two scoping rules. F9 was the first symptom; F14 and F16 are the next two.

### Convergence — this is the signal to stop and ask

Six review rounds, five fix rounds. Findings by round: 4, 2, 2, 0 (fix-only), 4, 5. **The count is
not falling.** Every finding since round 1 lives in the fold-resolution machinery that entered this
change as the operator-approved guard-gap fix; KAN-77's own subject — the ledger path, the rescue,
the MISSING outcome, artifact brevity — has been clean since pass 1 and has not been touched by any
round since.

Two of this round's findings (**F14**, **F15**) are defects *this change's own fix rounds introduced*.
Handed back to the operator rather than dispatching a sixth fix round on the parent's own judgment.


## Fix rounds 6 and 7 — the root cause, and a surviving mutant

**Round 6 (F13, F14, F15, F16).** Opus. `433c52b` → `f8c0388`. Attacked the root the Principles slot
named rather than the sixth symptom.

- **One shared grammar module**, `scripts/lib/plan_grammar.py` — the first `.py` in a `lib/` that had
  held only shell. Both guards import `DOTTED_ID`, the field-presence and gate patterns, and
  `partner_ids()`; neither defines any of them. **No parity comment anywhere** — the artifact that
  let F9 hide is gone rather than replaced.
- **Import verified through the skill symlinks**, not just from the repo root: invoked via `scripts/`,
  `skills/myflow-do/scripts/` and `skills/myflow-fast/scripts/`, all exit 0.
- **`check-guard-symlinks.sh` rule 2 handled by making the dependency detectable**, not by extending
  the guard: the wrapper now sets `GRAMMAR_MODULE="$SCRIPT_DIR/lib/plan_grammar.py"`, which matches
  rule 2's own `$SCRIPT_DIR/<name>` grep, so `lib` becomes a required sibling. Removing the symlink
  now fails that guard.
- **Scoping settled: line-scoped, first line wins.** The value grammar is closed and short so it never
  needs to wrap; under the joined reading an adjacent prose line silently changed the field's meaning,
  which *is* F14; and build-green has always been line-scoped, so this is the rule both guards can
  hold without either loosening.
- Harnesses: commit-fields 138 → 149 (8 of 11 new assertions failed pre-fix); build-green 34 → 37.

**Round 7 (surviving mutant).** Opus. 149 → **151**. The parent's round-6 mutation found that removing
the wrapper's module-presence check broke nothing. Case 56 now fails with the check mutated to
`if false` (rc 1 and a traceback instead of rc 2) and passes with it restored. Test only — no guard
behaviour touched.

### Mutation proof — parent, round 6

fix-mutation: scripts/lib/plan_grammar.py — `partner_ids` stopped distinguishing an ungated value (`None`) from a gated one naming nobody (`[]`) — **the build-green harness** failed, not commit-fields. That cross-guard failure is exactly what the shared module buys and what the old parity comment only pretended to provide
fix-mutation: scripts/check-task-commit-fields.py — fold membership grown over raw `squash_partners` again, reverting F15 — commit-fields failed (2 assertions)
fix-mutation: scripts/check-task-commit-fields.sh — dropped the module-presence check — **SURVIVED**, repaired in round 7 by case 56
fix-mutation: scripts/check-task-build-green.py — none — its verdicts and messages are unchanged; the one behaviour that moved (`**Squash-with:**Task 2` with no space is now a field) matches what commit-fields' `FIELD_RE` always accepted, and is pinned by the three cases added to its own harness
fix-mutations-total: 4

### Parent-side repairs this round

**A misplaced file, and the orchestrator's error that caused it.** `preserve-session-records.sh` —
task 2's file — sat in the tasks 7/8 commit, because fix round 5's dispatch bundled F9 (a 7/8 file)
with F12 (a task 2 file) into one round. Moved to `6b91aa2` where it belongs. **The first attempt
reverted F12 at HEAD** rather than relocating it, and was caught only by reading the file's actual
content rather than trusting the rebase; repaired. Every commit now carries only its own task's
files, verified per commit and by running the guard against all seven task ids.

**A stale commit message.** The tasks 7/8 message still claimed 94 assertions (now 151), cited shas
the rebases had replaced, and — worst — still asserted the **parity claim fix round 6 deleted from
the code**. Amended: the message now records that the comment was wrong about the gate, that this is
how F9 hid for four rounds, and that the grammar now lives once.

**F17 withdrawn by the operator**; reason on its marker line above.

### Recorded, not acted on

- `scripts/test-check-task-commit-fields.sh` has **no cleanup trap** and leaks every `new_repo`
  fixture under `TMPDIR`. Pre-existing; `test-preserve-session-records.sh` has the `TREES`-array idiom
  this one lacks. Flagged by the round-7 fixer, which also correctly caught that the parent's brief
  had claimed the trap already existed here.
- **The round-6 fixer's structural judgment.** F13 and F14 are closed *by construction* — no second
  grammar copy remains to drift. F15 and F16 are a different family that a shared parser does not
  reach: `resolve_folded_task`/`_fold_group`/`_fold_reds` is now ~200 lines implementing connected
  components over a partially-invalid edge set, with no explicit fold object — every rule a guard
  clause layered on a set recomputed from every task id. Its recommendation, endorsed by the parent:
  if a later round finds something there, **do not fix it in place** — build the fold graph once per
  plan and make every check a query against it, with the 151 assertions as the contract. That is a
  refactor deserving its own change.

## Round 7 — Full re-run

**Diff size:** measured **3011** against the cap of **2000**, exit 1. The operator's round-5 answer —
proceed with the panel — is carried forward again; the question is unchanged and the growth since
(2440 → 3011) is fix rounds 6 and 7's own shared module and harness cases. 583 lines remain the
rename artefact, so reviewable content is roughly 2428 lines.

**Scoped diff:** `since-last-review.diff` (886 lines), `433c52b..56c8293` — everything the round-6
slots have not seen. Accurate rather than reconstructed: `433c52b` is still reachable despite the
rebases, so this is a real commit range and not a guess.

**Mode:** Full. Seven fix rounds against a ladder that escalates at three; the last altered a guard's
behaviour, a second guard, a shared module and a delta spec.

**Why this round exists:** no slot has reviewed `56c8293`. Every round-6 clean result predates the
shared-parser work, the commit relocation and the message amendment.


## Round 7 results

| Slot | Result |
|------|--------|
| 0 — Primary | clean verdict, 1 Minor — **F18**. Independently verified the parent's two history rewrites |
| 2 — Principles | **clean**; F13 confirmed resolved *by construction*, verified by grepping both guards rather than by inspection |
| 3 — Code review (low) | 1 Medium — **F19**, confirmed. **Declined** Primary's F18 with reasons |

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F18 | Primary | Minor | `scripts/check-task-build-green.sh` | fix round 6 gave `check-task-build-green.py` the same `plan_grammar` dependency, but only the sibling wrapper got the `GRAMMAR_MODULE` existence check; this one crashes with `ModuleNotFoundError` and exit 1 instead of a guarded exit 2 |
| F19 | Code review (low) | Medium | `check-task-commit-fields.py:359-372` vs `check-task-build-green.py:187-193` | the two guards disagree on **which line is the field** when a body holds more than one `Squash-with:`-shaped line and the first does not gate |

### F19 — verified by the parent on one fixture, both halves

`check-task-commit-fields.py` takes the **first** `Squash-with:`-shaped line whether or not it gates
(`if squash_value is None: squash_value = …`, line 371). `check-task-build-green.py` commits to a line
only once it gates (`if ids is not None:`, line 195) and otherwise keeps scanning — exactly as its own
docstring specifies.

On one body carrying `**Squash-with:** Task 2 (see note below)` then `**Squash-with:** Task 2`:
build-green exits **0**, resolving partner 2; commit-fields selects `'Task 2 (see note below)'` with
partners `[]`, and through F16's plan-wide reporting path fails **every task in the plan**.

**F13's fix shared the patterns but not the selection.** Each guard kept its own line-selection loop,
so the divergence the shared module was written to eliminate survived in the half nobody moved.

**This is the third false parity comment in this change.** `check-task-commit-fields.py:362-368`
claims the two guards "read one field the same way (fix round 6, F14)" — and they do not. The first
such comment produced F9; fix round 6 deleted it and wrote a new one that is also wrong. A comment
asserting cross-file parity has now been wrong every time it has appeared here.

### Code review declined F18, with reasons

`check-task-build-green.sh` sits on `check-guard-symlinks.sh`'s `EXEMPT` list and is never symlinked
into any skill's `scripts/`, so no current path reaches the missing check — a less friendly traceback
in a scenario that does not occur. Recorded rather than folded in: a slot relaying another slot's
finding it believes is not live would put an unearned item into a fix round.


## Fix round 8 — F18, F19

**Fixer:** Opus. `56c8293` → `cc1f673`. Commit lists only `scripts/` files.
commit-fields 151 → **154** (all 3 new failed pre-fix); build-green 37 → **41** (2 of 4 failed
pre-fix; case 19 is a regression pin that passes either way by design). **Bounces:** none.

**Field selection now lives in `scripts/lib/plan_grammar.py`** — `select_squash_with(body)`, plus
`FENCE_RE` moved there. Rule: the first non-fenced line whose value **gates** is the field. **Neither
guard retains a selection loop**; neither imports `partner_ids` any more. `check-task-commit-fields.py`
still matches the field shape to terminate a preceding field's continuation, as required.

**The false parity comment was deleted, not rewritten.** No comment anywhere now asserts the two
guards agree — the shared import does.

### Mutation proof — parent, round 8

fix-mutation: scripts/lib/plan_grammar.py — `select_squash_with` reverted to taking the first *shaped* line rather than the first *gating* one — **both** harnesses failed (3 + 2). A change to the shared selector now breaks both guards' tests, which is the enforcement a comment never provided
fix-mutation: scripts/check-task-build-green.sh — dropped the module-presence check — build-green harness failed (2 assertions)
fix-mutation: scripts/check-task-build-green.py — none — its Squash-with selection moved to the shared `FENCE_RE`, the same regex it defined locally; no behaviour change
fix-mutations-total: 3

## F20 — the same root, a fifth time, found by asking rather than by review

The round-8 dispatch asked the fixer directly: *is anything left that these two files must agree about
with nothing enforcing it?* That question has produced three defects in this change (F9, F14, F19), and
every previous time it was answered only after the defect surfaced. This time it was answered first.

| ID | Source | Severity | Location | Note |
|---|---|---|---|---|
| F20 | fix round 8's self-assessment | Medium | `check-task-build-green.py` `BUILD_TAG_RE`/first-wins vs `check-task-commit-fields.py` `FIELD_RE`/joined/last-wins | the two guards read `**Build:**` by different rules and disagree on real bodies |

**Confirmed by the parent on one fixture.** A task body carrying `**Build:** red` then `**Build:**
green`: build-green reads **red**, commit-fields reads **green**. A second shape — `**Build:** red`
followed by an unblanked prose line — has build-green read **red** while commit-fields joins the
continuation, fails `BUILD_KIND_RE` and reads **None**.

**Consequence:** build-green resolves the fold while commit-fields puts the task on the ordinary
single-commit path, checking a folded commit against a declared subject the fold deleted.

**This is F14 and F19 one field over**: same file, same guards, two selection-and-scoping rules,
nothing enforcing agreement.

### A sixth instance, scoped out and recorded

What counts as a **fence** is still read two ways: `check-task-commit-fields.py` keeps
`FENCE_LINE_RE = ^``` ` while the shared module and build-green use `^ {0,3}(`{3,}|~{3,})`. After
round 8 `parse_task_fields` carries both — the shared one inside `select_squash_with`, its own for
`Files:`/`Tests:`/`Commit:`/`Build:`. Deliberately not unified: `FENCE_LINE_RE` also parses
`.myflow/project.md`'s `## test` block, so the change reaches past these two guards. Nothing pins the
two fence rules against each other.


## Fix round 9 — the sweep

**Fixer:** Opus. `cc1f673` → `1e26209`. Commit lists only `scripts/` files.
commit-fields 154 → **162** (7 of 8 new failed pre-fix); build-green 41 → **50** (its 9 are
cross-harness pins; its own readings were already canonical). **Bounces:** none.

**The audit found three more dual-defined items beyond F20** — each of which, on the previous
one-per-round pattern, would have cost its own review round:

| Item | Agreed? | Resolution |
|------|---------|-----------|
| `**Build:**` tag (F20) | **no** — build-green line-gated first-wins, commit-fields joined continuations last-wins | `plan_grammar.select_build_tag`; `BUILD_KIND_RE` deleted |
| Task headings / which exist | **no** — commit-fields ignored fences entirely, so a `### <id>` inside a fenced example opened a **real task** | `plan_grammar.iter_tasks` |
| Task body start/end | **no** — patterns shared, selection was not: the F19 shape again | `plan_grammar.iter_tasks` |
| Which task a duplicated id names | **no** — build-green first, commit-fields last | `plan_grammar.select_task`, first wins |
| Dotted-id grammar, `Squash-with:` shape/gate/selection | yes | confirmed, untouched |

Each new shared reading is pinned by a case in **both** harnesses on the same body (cf 58-61 /
bg 21-24). No comment anywhere asserts the two guards agree.

### Mutation proof — parent, round 9

fix-mutation: scripts/lib/plan_grammar.py — `select_build_tag` reverted to last-wins — **both** harnesses failed (2 + 2)
fix-mutation: scripts/lib/plan_grammar.py — `select_build_tag` stopped skipping fenced lines — build-green failed (4 assertions)
fix-mutation: scripts/lib/plan_grammar.py — `select_task` reverted to resolving a duplicated id to its last heading — commit-fields failed (1 assertion)
fix-mutation: scripts/check-task-build-green.py — none — its refactor onto the shared selectors is behaviour-preserving; 41/41 of its pre-existing assertions unchanged
fix-mutations-total: 4

A first, blunt attempt at the `select_build_tag` mutation broke the function outright (61 + 23
failures) rather than flipping one mechanism. Discarded and redone surgically, per the rule that each
mutation alters one mechanism — a mutation that breaks everything proves nothing about the mechanism
it meant to test.

## F21 — the fence residue, now a finding

The fence carve-out was scoped out of round 9 deliberately. Having verified its effect, the parent is
raising it rather than leaving it recorded as acceptable.

| ID | Source | Severity | Location | Note |
|---|---|---|---|---|
| F21 | parent, verifying round 9's recorded residue | Medium | `check-task-commit-fields.py:265`, used at `:360` | `FENCE_LINE_RE = ^``` ` misses tilde and indented fences, so a fenced **example** field inside a task body is read as a real field |

**Reproduced by the parent.** A task body declaring `**Files:** \`real.txt\`` and then carrying a
`~~~markdown` block containing `**Files:** \`example-only.txt\`` parses as
`files = ['example-only.txt']` — the fenced example **replaces** the real declaration, because the
field loop overwrites. A commit touching `real.txt` then fails as undeclared, and one touching
`example-only.txt` passes. **Both directions wrong.**

The same body's `**Build:**` tag and its heading are read correctly, because those now go through the
shared fence-aware selectors. So this is no longer a two-guard disagreement — it is an inconsistency
**inside** commit-fields, between the shared selectors and its own field loop.

**The entanglement is smaller than round 9 was told.** `FENCE_LINE_RE` has exactly two call sites:
line 360, the `tasks.md` field loop (defective), and line 929, `.myflow/project.md`'s `## test`
section (correct for its own purpose). They are separable — the fix is one call site, not a
project-configuration refactor.

**Exposure today:** this repository's own plans use no tilde fences, so nothing currently trips it.


## Fix round 10 — F21

**Fixer:** Opus. `1e26209` → `08f085e` → `617e267` after a message amend. Commit lists only
`scripts/` files. commit-fields 162 → **166**; **all 4 new assertions failed pre-fix**. build-green
untouched at 50. **Bounces:** none.

One call site: the field loop's fence toggle now uses the shared `FENCE_RE`. Line 942's use of
`FENCE_LINE_RE` for `.myflow/project.md`'s `## test` block is untouched and still has that one user;
its comment now says it is scoped to that file.

### Mutation proof — parent, round 10

fix-mutation: scripts/check-task-commit-fields.py — field loop reverted to the backtick-only fence rule — case 62 failed (4 assertions, both directions: the real declaration reported undeclared, and the fenced example's file passing)
fix-mutation: scripts/check-task-commit-fields.py — none — line 942's project.md fence use is untouched by design
fix-mutations-total: 2

### A stale commit message, amended a second time

The tasks 7+8 message claimed 154 assertions beside 41; rounds 9 and 10 both left it. Now 166 beside
50, naming ten fix rounds and F1-F21. **This is the second time this message has gone stale** — a
commit message is written once and then keeps being fixed under it, and nothing checks it. Recorded
as an observation, not fixed structurally.

### Recorded, not acted on

- **`FENCE_RE`'s toggle is naive about delimiter kind**: a `~~~` line inside a backtick block closes
  the fence. Shared, pre-existing, unreachable by any case here, and unifying on it is what the
  operator asked for. Unchanged and stated rather than left implicit.
- The commit-fields harness still leaks fixture directories for want of a cleanup trap. Known,
  pre-existing, out of scope through three rounds now.

## Final state before the last review round

**Branch**, each commit carrying only its own task's declared files, guard verified against all
seven task ids:

| Commit | Task | Files |
|---|---|---|
| `6b91aa2` | 1+2 | `preserve-session-records.sh` + harness |
| `f958148` | 3 | `pipeline.md` |
| `16a8ca8` | 4 | `myflow-do/SKILL.md` |
| `e745a31` | 6 | `pipeline.md` |
| `617e267` | 7+8 | both guards, both wrappers, `lib/plan_grammar.py`, both harnesses |

**Harnesses:** 169 / 166 / 50. **Findings:** 21 — 18 fixed, 3 withdrawn by the operator.
**Fix rounds:** 10. **Review rounds:** 7.


## Round 8 results — the panel's last full round

| Slot | Result |
|------|--------|
| 0 — Primary | **clean**. Searched every non-archived `tasks.md` for the four shapes rounds 9/10 changed; only archived plans on a legacy heading convention the grammar never parses. Diffed the tasks 7+8 message against both prior amended states — 166/50, F1-F21, "ten fix rounds" all accurate |
| 2 — Principles | **F13's root cause structurally closed**. Verified the cross-harness pins pair up (19↔57, 21↔58, 22↔59, 23↔60, 24↔61) and that F21's fence fix correctly has *no* build-green pin, since that guard never had the buggy field loop. 1 Minor — **F23** |
| 3 — Code review (low) | 1 Medium — **F22**, reproduced against a real commit. The `code-review` skill ran **on-target** and found nothing; the finding is the slot's own |

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F22 | Code review (low) | Medium | `check-task-commit-fields.py:364` | a bare `~~~` line in a task body toggles fence state and swallows every field after it to end of body — the declared file is then reported undeclared |
| F23 | Principles | Minor | `check-task-build-green.py:180-190` | `by_id`'s "first heading wins" duplicates `select_task`'s rule with no note that the duplication was considered and kept |

### F22 — reproduced, but the slot's framing is wrong in a way that changes its disposition

**Reproduced by the parent.** A body with `**Build:** green`, a bare `~~~`, then `**Files:**` and
`**Commit:**` parses as `files = []`, `commit = None`.

**The regression claim is true**: round 10 changed this. The old `FENCE_LINE_RE = ^``` ` ignored
tildes, so the fields were read.

**But the slot calls the `~~~` "ordinary prose, not a real fence — e.g. a divider". That is
incorrect.** In CommonMark a bare `~~~` on its own line **opens a fenced code block**, and a fence
never closed runs to the end of the containing block. The fields after it *are* inside a fence: any
renderer shows them as code, not as fields. So the guard's new reading **matches what a human sees in
the rendered plan**, and the old one diverged from it. A `~~~` used as a divider is a malformed plan —
markdown's divider is `---`.

**What is genuinely wrong is the diagnostic, not the verdict.** The guard reports "file not declared"
when the real defect is an unclosed fence in the task body. Put to the operator on that basis rather
than as a request to revert round 10, which would reintroduce F21.


## Fix round 11 — F22, F23

**Fixer:** Opus. `617e267` → `55e029c`. Commit lists only `scripts/` files, and the fixer updated the
message body itself rather than leaving it stale a third time. commit-fields 166 → **170**;
build-green 50 → **54**. **Bounces:** none.

**Detection went into the shared module** — `plan_grammar.unclosed_fence(body)` — and the reasoning is
right: the same unclosed fence blinds *both* guards through the same selectors (commit-fields reads
an empty `Files:`, build-green reads no `Build:` tag), so fixing it in one would leave the other
misdiagnosing the identical body. Two callers already existed, so this is not speculative
abstraction. **Cross-harness pinned** on a byte-identical body (cf 64 / bg 25).

**Replaces rather than accompanies.** The swallowed checks are skipped for that task. The fixer's
reasoning, which the parent endorses: the downstream lines are not merely noisy, they are *false about
the commit* — it touched exactly the file the task declared, and the guard said it did not. Precedent
cited from the same file: an unresolvable fold already returns alone.

**F23** got one comment explaining why `by_id` keeps its own index — it asserts no parity between
implementations, which is the distinction that matters here.

### Mutation proof — parent, round 11

fix-mutation: scripts/lib/plan_grammar.py — `unclosed_fence` always returns None, disabling detection — **both** harnesses failed (2 + 2)
fix-mutation: scripts/lib/plan_grammar.py — `unclosed_fence` returns an offset for a *closed* fence too, a false positive — **both** harnesses failed (6 + 7)
fix-mutation: scripts/check-task-build-green.py — none — F23 is a comment
fix-mutations-total: 3

### Stale planning figures, corrected by the parent

Tasks 7/8's `**Baseline:**` still read `before=113 after=113`, four rounds stale, and build-green's
count was recorded nowhere. Now 170 beside 54, with both provenance refs repointed at fix round 11.
**This is the third time these figures have gone stale**; `Baseline:` is skipped-not-verified in this
project, so no guard catches it. Worth knowing as a property of the field, not just this change.

### Three residual unclosed-fence misdiagnoses, reported by the fixer and not acted on

None is a fail-open; each is the guard naming the wrong defect.

1. **A partner task's** unclosed fence is undetected by commit-fields, which checks only the id on the
   command line — a swallowed partner `Files:` narrows the fold union and blames the commit.
2. An unclosed fence **swallows every following task heading** (`iter_tasks` tracks fence state
   file-wide), so commit-fields invoked for a later id exits 2 `task N not found`.
3. An unclosed fence **before the first heading** is invisible to both: no task opens, and build-green
   exits 0 having scanned zero tasks — the vacuous-pass shape **KAN-114** already tracks.
