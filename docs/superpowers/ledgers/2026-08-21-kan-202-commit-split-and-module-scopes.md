# SDD ledger — kan-202-commit-split-and-module-scopes

Run: `/myflow-fast`, harness `claude-code`, session token `mf-k202scope1`.
Recorded models: implementation `sonnet`, review panel `sonnet`, panel fixes `sonnet`,
roster `light` — `/myflow-fast`'s recorded defaults, taken without asking per **Recorded defaults
favor speed** (`skills/myflow-fast/SKILL.md`).

Base branch `main` at `355fe28` after the rebase described below. Branch
`openspec/kan-202-commit-split-and-module-scopes`.

## Dispatches

### Bundle 1 — tasks 1, 2, 7, 8

- **Implementer**, model: `sonnet` (named explicitly on the dispatch). Two squash units.
- **Task 1 → 2: complete** (commit `367cfaa`, review clean with one Minor finding, see F1 below;
  model: sonnet; review: combined). Subject
  `feat(check-task-commit-fields): reject a declared Commit scope naming the change or a task id`.
  Harness `scripts/test-check-task-commit-fields.sh` 29 → 43 assertions, matching the plan's
  declared `after=43`. RED confirmed first: 8 failures before `check_commit_scope` existed.
- **Task 7 → 8: complete** (commit `255914e`; model: sonnet; review: combined). Subject
  `feat(rules): carry the module-scope commit convention as an always-on rule`. Harness
  `scripts/test-setup.sh` 468 → 472 assertions, matching the plan's declared `after=472`. RED
  confirmed first.

### Bundle 2 — task 3

- **Implementer**, model: `sonnet` (named explicitly on the dispatch). In flight at the time of
  writing.

### Per-task reviews

- **Tasks 1+2**, model: `sonnet`, combined slot (spec + quality), range `355fe28..367cfaa`.
  One Minor finding, F1.
- **Tasks 7+8**, model: `sonnet`, combined slot (spec + quality), range `367cfaa..255914e`.
  In flight at the time of writing.

## Per-task review findings

| ID | Task | Severity | Location | Note |
|---|---|---|---|---|
| F1 | 1–2 | Minor | `scripts/check-task-commit-fields.py:337` | The change-name branch has no case pinning equality against substring, so mutating `scope == change_name` to `change_name in scope` survives the whole suite — unlike the Jira-key branch, which case 23 pins. |

F1 is a **surviving mutant**, established by the reviewer running the mutation rather than by
reading: with the comparison widened to a substring test the harness still reports `all cases
passed`. The Jira-key branch is already pinned by case 23 (`kan-900-helpers` against key `kan-900`);
the change-name branch needs its equivalent — a declared scope that strictly *contains* the change
name and must therefore pass.

Reproducer for F1:

```
cd /Users/tweety53/Projects/agents-worktrees/kan-202-commit-split-and-module-scopes
sed -i '' 's/    if scope == change_name:/    if change_name in scope:/' scripts/check-task-commit-fields.py
bash scripts/test-check-task-commit-fields.sh 2>&1 | tail -1   # -> "all cases passed"
git checkout -- scripts/check-task-commit-fields.py
```

## Parent-side repairs during the run

### Task 8's `**Files:**` was incomplete — a plan defect, repaired in the plan

Adding an always-on rule file trips `check-references.sh`'s per-file coverage check
(`rules/commit-scope-is-the-module.mdc:0: 0 checked, and not declared expected-zero`) unless the file
is registered in that guard's `EXPECTED_ZERO_RULE_FILES` array, as all ten pre-existing always-on
rule files are. Task 8 did not declare `scripts/check-references.sh` under `**Files:**`, so its own
commit failed the very guard task 2 had just built.

The implementer made the one-line registration in alphabetical order — the established pattern, with
no reason string weakened and no suppression added — and **reported the undeclared path rather than
trimming the change to fit the declaration**, which is the correct disposition: `tasks.md` is a
planning path an implementer must never touch. The parent amended task 8's `**Files:**` and added an
explicit registration step, after which `check-task-commit-fields.sh` exits 0 for task 8.

### The guard wrapper could not resolve, because the branch base was stale

`check-task-commit-fields.sh` exited 2 — `more than one tasks.md found under openspec/changes`,
naming this change's and `kan-239-run-2-asserts-base-branch-and-archives-via-pr`'s. The wrapper was
right: local `main` stood at `69df0af`, two commits behind `origin/main` at `355fe28`, which already
carries kan-239's archive (PR #24, merged). The worktree had been cut from the stale base, so
kan-239's change directory was still un-archived inside it.

Not a defect in the guard, and not worked around: `main` was fast-forwarded to `355fe28` and the
branch rebased onto it with `git rebase --onto 355fe28 69df0af`. Nothing had been pushed and no PR
existed, so the rewrite cost nothing. Commits `432b5a6` → `367cfaa` and `8ffaf0e` → `255914e`. The
wrapper resolves cleanly at the new base and both commits pass.

### Three anti-fix symlinks removed from the main checkout

`skills/myflow-fast/` held untracked symlinks `engineering-principles.md`,
`principles-reviewer-prompt.md` and `adversarial-reviewer-prompt.md` pointing at `../myflow-do/`.
Nothing in `setup.sh` or `scripts/` creates them. `skills/myflow-do/SKILL.md` states that
`[PRINCIPLES_PATH]` resolves **beside itself** — `skills/myflow-do/`, including under `/myflow-fast`
— and names symlinking a copy in as the wrong fix; that sentence landed in `d149107` at 19:05 on
2026-08-20, and the symlinks were created at 23:58 the same evening. Removed; both
`check-guard-symlinks.sh` and `check-references.sh` exit 0 afterwards.

The operator then asked for the recurrence to be closed mechanically, which became tasks 9 and 10
(rule 5) and an added requirement in
`specs/myflow-contract-distribution/spec.md`.

## Fix round 1 — F1 and F2

- **Fix subagent**, model: `sonnet`. Two fixups, folded via `git commit --fixup=` +
  `git rebase --autosquash 355fe28`. Branch still three commits, no `fixup!` left unsquashed.
  `432b5a6`/`367cfaa` → `b1545e0`, `8ffaf0e`/`255914e` → `1d08d0c`, `0e9663e` → `79ed92e`.
- **F1 fixed.** Case 24 added to `scripts/test-check-task-commit-fields.sh`: change name
  `kan-900-some-change`, declared scope `kan-900-some-change-helpers` — strictly contains without
  equalling — must pass. Assertions 43 → 45.
- **F2 fixed.** `rules/commit-scope-is-the-module.mdc:35` now cites
  `<agents repo>/scripts/check-task-commit-fields.py`. Nothing else in that file touched.

fix-mutation: `scripts/check-task-commit-fields.py` — `scope == change_name` widened to
`change_name in scope` — case 24 fails (`rc=1`, "names the change, not a module"); restored, suite
green
fix-mutations-total: 1

**The parent ran that mutation itself**, not the fix subagent, per **The fix round mutation-proves
what it changed** (`skills/myflow-do/SKILL.md`). Before the fix the same mutation reported `all cases
passed` — that is what made F1 a surviving mutant rather than a style note.

The plan's declared baselines for tasks 1 and 2 were `after=43`, which the fix round invalidated.
Both were corrected to `after=45` and their provenance comments changed from `predicted:` to
`measured:`, since the count is now a measurement rather than an expectation.

## Per-task review — task 3

Model `sonnet`, combined slot, range `1d08d0c..79ed92e`. Four findings.

| ID | Severity | Location | Note | Disposition |
|---|---|---|---|---|
| F1 | Major | `skills/myflow-start/SKILL.md:405-410` | Enumerates the three prohibited shapes *and* cites the requirement canonical for them — a copy beside its own citation, which this repository treats as drift. | dispatched to fix round 2 |
| F2 | Major | `skills/myflow-start/SKILL.md:410-411` | Cites `<agents repo>/openspec/specs/myflow-commit-scope/spec.md`, a new capability that does not land until this change's own finish run 2, with no caveat saying so. `check-references.sh` silently skips a path that does not resolve, so the pointer is unverified and does not look it. | dispatched to fix round 2 |
| F3 | Major | `skills/myflow-start/SKILL.md:403-406` vs `skills/myflow-do/SKILL.md:230-232` | `myflow-do` still instructs `<n>` as the scope, contradicting task 3's new text. | **closed — not a defect** |
| F4 | Minor | `skills/myflow-start/SKILL.md:403-411` | The bullet runs to ~6 sentences against 1-2 for its siblings, and "the three prohibited shapes are **Requirement: …** (path)" reads as though the citation *is* the shapes rather than naming where they are canonical. | dispatched to fix round 2 |

**F3 is closed rather than fixed, and the reason is checkable rather than a judgment call.** The
finding rests on the premise "as merged today (task 3 alone)". A task never merges alone: the branch
merges as a unit, and task 4 — in this same plan, on this same branch — is what removes the `<n>`
instruction from `skills/myflow-do/SKILL.md`. The two files therefore never reach `main` in a
conflicting state. What the finding does correctly establish is that the ordering dependency between
tasks 3 and 4 is real, which is recorded here rather than dismissed.

## Fix round 2 + bundle 3 — one dispatch

- **Implementer**, model: `sonnet`. Job 1: fold F1, F2 and F4 into task 3's commit. Job 2: tasks 4,
  5 and 6. In flight at the time of writing.

## Parent-side repairs, continued

### Tasks 3 and 5 declared guards as tests

`check-task-commit-fields.sh` failed task 3 with `declared test scripts/check-references.sh not
found in the diff`. The cause is the field grammar, not the commit:
`_extract_backtick_tokens` treats **every backticked token** in `**Tests:**` as a declared test name
and then requires each to appear in the diff. Both tasks had written prose naming the covering
guards in backticks.

The archive's convention for a prose-only task is an unbackticked sentence
(`**Tests:** none — prose-only change.`). Tasks 3 and 5 were rewritten to match, naming the guards
without backticks and saying why. Task 5 had not yet been implemented when this was found, so the
repair reached it before its implementer did.

**This is the third plan defect the guards caught during execution, and all three were in fields the
parent wrote** — task 8's incomplete `**Files:**`, tasks 3 and 5's backticked `**Tests:**`, and tasks
1 and 2's baselines invalidated by a fix round. Each was mechanically detected and none reached
review as a finding, which is the field family working; what it also shows is that the grammar's
sharp edges surface only when the guard runs, long after planning. Recorded for the self-review at
finish.

## Fix round 2 + bundle 3 — outcome

- **Implementer**, model: `sonnet`, one dispatch covering both jobs.
- **Task 3's fixes folded** into its commit: `79ed92e` → `a5393ef`. F1's enumeration removed, F2's
  unresolved-reference caveat added matching `state-file.md`'s phrasing, F4's clause given a verb.
- **Tasks 4, 5, 6 committed**: `e0a2525`, `44ff2b7`, `6531929`. Seven-commit branch reduced to six
  after the autosquash; six commits total on `355fe28`.

### Targeted re-review of task 3

Re-run by **the same reviewer that raised the findings** (per the Targeted row of **Panel re-runs**,
`skills/myflow-do/SKILL.md`), resumed with its own context rather than dispatched fresh. Range
`1d08d0c..a5393ef`.

F1, F2 and F4 **all confirmed closed**, each with a reproducer demonstrating the closure. **No new
finding introduced by the fix round.** The reviewer also examined the F3 closure on its merits and
**agreed with it**, stating that its own original framing — "as merged today, task 3 alone" — was the
wrong scope, and confirming `grep -n "<n> is the scope" skills/myflow-do/SKILL.md` now returns
nothing because task 4 (`e0a2525`) sits on the same branch. The closure was not re-raised.

It also declined to accept one word of the fix report at face value: the fix claimed to have
"trimmed" the bullet, and the reviewer noted the bullet in fact grew, because F2's caveat added
prose — judging the growth a necessary cost of F2's correctness fix rather than a fresh defect, and
saying so instead of letting the claim stand.

## Parent-side repair — the contract budget was raised to zero headroom

The bundle-3 implementer raised `scripts/check-contract-budget.sh`'s row for
`skills/myflow-contracts/finish-contract.md` twice, ending at **40901 — exactly the file's new
size**. The guard reported `BUDGET-OK`, so nothing failed and the run would have proceeded.

That value is wrong against the guard's own header, which states a budget is "the size its file
actually had when the change that added its row landed, **plus 25%**". Measured against the real
tree, every other row carries roughly 20-22% headroom; this one carried **0.0%**, meaning the next
byte added to `finish-contract.md` would fail lint. A ratchet had been turned into a tripwire.

Corrected to **51126** (40901 + 25%) by the parent and folded into task 6's commit via
`git commit --fixup=53d71cb` + autosquash, taking task 6 to `6531929`. Tasks 5 and 6 both gained
`scripts/check-contract-budget.sh` in their `**Files:**` fields, with the plus-25% rule stated there
so a later implementer does not repeat it — and the tasks 9-10 dispatch carries the rule explicitly
for the same reason.

**This is the fourth plan defect the run surfaced, and the second where a guard passed on a wrong
value.** Task 8's incomplete `**Files:**`, tasks 3 and 5's backticked `**Tests:**`, tasks 1 and 2's
fix-round-invalidated baselines, and now tasks 5 and 6's undeclared budget script. The budget case is
the one worth carrying into the self-review: `check-task-commit-fields.sh` catches an undeclared
file, but nothing catches a budget row set to the wrong value — `BUDGET-OK` is returned either way,
and only reading the guard's header against the real numbers finds it.

### Checkboxes ticked

Tasks 1, 2, 3, 7 and 8 — 16 step checkboxes — ticked, each after its task passed both spec and
quality review. Tasks 4, 5, 6, 9 and 10 remain open pending their reviews.

## Panel pass 1 — light roster, three required slots

Optional slots: **declined by the operator**. Security, Adversarial, Lens B and Lens C all had
triggers fire against the 478-line diff and were put in one multi-select prompt per the `light`
preset; the operator chose none. Recorded as declined, which is distinct from a trigger never firing.

| Slot | Model | Result |
|------|-------|--------|
| Primary | sonnet | 1 Minor — task 5's own commit left the contract budget at zero headroom (corrected forward by task 6) |
| Principles (Merged) | sonnet | clean; verified `_leading_jira_key` was genuinely new rather than a DRY duplication |
| Code review (low) | sonnet | **1 Major** — `gather-self-review-context.sh` resolves a change's commits by grepping the change name as the scope, which tasks 5 and 6 remove; plus 1 Minor on rule-5 block order |

**The low-effort slot found the only Major, which three per-task reviews and the principles slot had
all missed.** It runs at the cheapest setting in the roster.

## Fix rounds 3, 4 and 5

Round 3 added tasks 11 and 12 (path-based resolution). **The parent's own mutation pass then found
the fix defective**: every fixture it added wrote the planning commit directly at the archived path,
a shape no real run produces, leaving `PLAN_PATH_LIVE` unexercised. Modelling the real
run-1-then-run-2 sequence showed the archive commit — a `git mv` touching both pathspecs — won
`--max-count=1`, so `PLAN_SHA` resolved to the archive commit. Confirmed empirically in a throwaway
repository before being handed back.

Panel pass 2 (Full, escalated — a guard's behaviour changed and three fix rounds had run) returned
**nine findings across all three slots**, every one reproduced by the parent before being acted on:

| Finding | Severity | What |
|---------|----------|------|
| A | Major | `feat(<change-name>)!: …` bypassed `check_commit_scope` entirely — the regex did not match Conventional Commits' breaking-change form |
| B | Major | Comparisons were case-sensitive, so `feat(KAN-202)` passed — the shape a human is most likely to type, since Jira keys are uppercase |
| C | Major | `_leading_jira_key('release-2026-kan-450-cleanup')` returned `release-2026`, both over- and under-flagging |
| D | Major | Rule 5 could not see `skills/myflow-contracts/` — it reused a scan set that deliberately excludes it |
| E | Major | Any later commit touching the plan path became `PLAN_SHA`; a `docs(openspec): fix a typo` commit displaced the real one |
| G | Major | With the implementation commit skipped, an unrelated merge commit from another PR was reported as this change's implementation |
| F1/F2 (principles) | Important | The archive-shape regex hand-copied at four sites; the `PLAN_SHA` pipeline duplicated across an if/else |
| F3 (principles) | Important | The refuse-rather-than-guess branch had **zero** coverage — deleting the whole guard failed no test |
| I | Minor | 16 fixture sites hardcoded a change name that only happened to match the default |

Round 4 fixed all nine. `PLAN_SHA` now matches **path AND subject shape together** — path alone is
not commit-specific, subject alone is not change-specific — which also made the archive-exclusion
grep dead and let `PLAN_PATH_ARCHIVED` go entirely, collapsing the duplicated if/else. `IMPL_SHA`
accepts the parent only when it is non-merge and touches a path outside the two planning trees, the
heuristic the operator chose over labelling-as-unverified or resolving nothing.

Round 5 was the parent's own repair of **two surviving mutants in round 4's tests**: the finding-G
fixture merged an *empty* branch, so its parent failed the non-merge and outside-path conditions
simultaneously and neither was isolated. Two fixtures were added that isolate each.

fix-mutation: `scripts/check-task-commit-fields.py` — `SUBJECT_SCOPE_RE` optional `!` removed — 2 failures
fix-mutation: `scripts/check-task-commit-fields.py` — scope comparison returned to case-sensitive — 4 failures
fix-mutation: `scripts/check-guard-symlinks.sh` — rule 5 returned to the command-skills scan set — 2 failures
fix-mutation: `scripts/gather-self-review-context.sh` — `PLAN_SHA` subject filter removed — 10 failures
fix-mutation: `scripts/gather-self-review-context.sh` — outside-path requirement dropped — 1 failure
fix-mutation: `scripts/gather-self-review-context.sh` — both merge gates removed — 1 failure
fix-mutation: `scripts/gather-self-review-context.sh` — non-merge dropped from the acceptance line only — none; equivalent mutant, the flag also gates `PARENT_TOUCHES_OUTSIDE`
fix-mutations-total: 7

**Two mutation results this run were false negatives caused by the mutation never applying** — a
`sed` with a bad delimiter, and a regex that matched nothing. Both reported identically to a genuine
survivor. Re-run through a parser that asserts the target text was found, they were caught. Worth
carrying into the self-review: an unapplied mutation is indistinguishable from a survived one unless
the harness asserts the edit landed.

## Panel pass 3 — Full re-run

| Slot | Result |
|------|--------|
| Primary | **1 Important** — the old-era subject-only fallback's stated rationale is factually wrong and it is unreachable for any genuine historical commit; its only coverage came from `--allow-empty` fixtures, so nothing tested real old-era behaviour |
| Principles (Merged) | 2 Minor — the deliberate redundancy in `IMPL_SHA`'s acceptance needed a comment saying so; the fallback had no stated retirement trigger. Endorsed three design calls explicitly rather than restating a clean verdict |
| Code review (low) | 1 Minor — `git diff A^..A` fails on a root commit, `2>/dev/null` swallows it, and a genuine implementation commit is silently dropped |

**No Major findings in pass 3** — the first pass where that is true. The trend across passes is
1 Major → 5 Major → 0 Major.

**Both of pass 3's substantive findings were found by checking the code against reality rather than
against its tests.** The primary slot queried this repository's own archived history (`kan-201`,
`kan-209`, `kan-102`, `kan-236`) and found every old-era planning commit resolves through the primary
query. The low-effort slot ran the failing git command directly. Every test suite was green
throughout.

## Fix round 6 — a round that deleted code

- **35 lines removed.** The operator chose deletion over keeping the fallback with a stated expiry.
- **~18 test sections rewritten** from `--allow-empty` commits to real commits touching real paths,
  because a path-filtered `PLAN_SHA` cannot be satisfied by an empty commit — the old fixtures were
  structurally incapable of testing the mechanism.
- **Four ancestor-symlink sections reordered**, since `git add` refuses to write through a symlinked
  ancestor once the fixture commits real files.
- **Root-commit case fixed** with `git diff-tree --no-commit-id --name-only -r --root`.

fix-mutation: `scripts/gather-self-review-context.sh` — `check_boundary` made never to refuse — 19 failures, so the reordered ancestor-symlink sections are not inert
fix-mutation: `scripts/gather-self-review-context.sh` — root-commit fix reverted to `git diff A^..A` — 1 failure, naming the cause
fix-mutations-total: 2

## Panel pass 4 — Full re-run

Escalated twice over: fix round 6 altered a guard's behaviour, and more than three fix rounds had
run. Each slot was resumed with its own prior context rather than dispatched cold, so it re-reviewed
its own finding.

| Slot | Result |
|------|--------|
| Primary | *(pending at the time of writing)* |
| Principles (Merged) | both prior findings fixed; **1 new Minor** — `IMPL_SHA`'s acceptance has grown to four independently-motivated conditions and warrants extraction into a named function |
| Code review (low) | F1 fixed, **no new findings**; verified no reordered section passes for a different reason than its comment claims, and the deletion left nothing dangling |

The principles slot answered the accumulation question directly: each of the four gates was added in
response to a demonstrated bug rather than speculative hardening, which it judged the legitimate way
for a heuristic to grow — while flagging that the shape now warrants a named function.
