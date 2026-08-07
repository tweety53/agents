# Self-review — kan-87-cut-per-command-load-further

**Rating: 4/5** — operator's, recorded verbatim: good, with real gaps. The cut landed and every panel
finding was fixed, but the skill splits under-delivered, a stale-measurement defect recurred for the
fourth time, and run 2 hit two pipeline bugs the contract never described.

**Jira:** KAN-87 (Done). **PR:** [#2](https://github.com/tweety53/agents/pull/2), merged as
`8298dd0`. **Commits:** `2bcb685` implementation, `2672eb3` plan and session records, `3db3e30`
spec sync and archive.

**Models:** implementation Sonnet, review panel Sonnet, panel fixes Sonnet — all three operator
choices recorded at the creating run, overriding the pipeline defaults of Opus / Sonnet / Opus.

## What shipped, measured at merge

| File | before | after | change |
|------|-------:|------:|-------:|
| `skills/myflow-contracts/pipeline.md` | 88,253 | **64,701** | −27% |
| `skills/myflow-do/SKILL.md` | 40,529 | 37,885 | −6.5% |
| `skills/myflow-start/SKILL.md` | 28,944 | 25,999 | −10% |
| `skills/myflow-finish/SKILL.md` | 28,086 | 25,604 | −9% |

New: `finish-contract.md` (25,630, loaded by `/myflow-finish` alone) and three `SKILL-rationale.md`
siblings no run loads. `scripts/check-contract-budget.sh` covers 22 files and reports
`BUDGET-OK: 22 contract file(s) within budget`.

**Those two `after` figures for `pipeline.md` and `finish-contract.md` are not what the change
claimed.** The guide, the PR body and `2bcb685`'s message all say 64,621 and 26,246. See problem 4.

## Problems and fixes

Seven findings, all filed. Each carries its evidence in the issue rather than being summarised
here.

### 1. The preflight answered `RUN1` for a merged branch — KAN-88

Run 2's preflight was handed `main`, the form the contract's base resolution produces. Local `main`
was two commits behind, so the ancestor test compared against a ref that had never moved:

```
RUN1: HEAD is not an ancestor of main — not merged
```

Against `origin/main`: `RUN2`. The verdict is what decides between integrating and archiving, and a
wrong `RUN1` on a merged branch sends the operator back through the landing question and a
duplicate-PR attempt.

Two steps of the same run already disagree about which ref answers this: the contract strips
`origin/` when resolving the base, while **Worktree cleanup**'s check 3 uses `origin/$BASE`.

**It fails safe and it failed silently** — `RUN1` never archives, but nothing flagged the
contradiction. It was caught because local `main` was visibly behind.

### 2. Run 2's pull aborted on the main checkout's stale index — KAN-89

`/myflow-start` stages planning artifacts in the main checkout and never commits them. `/myflow-do`
then edits the copies in the worktree, and run 1 commits *those*. So the incoming merge adds paths
that are already staged, and `git pull --ff-only` refuses:

```
Please commit your changes or stash them before you merge.
Aborting
```

Four of seven files were byte-identical to what was landing; three were superseded — `tasks.md`
(+40/−104), the economy delta spec (93 lines), `design.md`. The stale copies had to be compared by
hash and discarded by hand.

**This is structural, not incidental.** Every change goes through the same
stage-in-main-checkout / edit-in-worktree shape, and run 2 has no step for it.

### 3. A renamed requirement tagged `## MODIFIED` — KAN-90

The economy delta widened its byte-budget requirement and renamed the heading with it, under
`## MODIFIED`. Heading matching finds nothing; applied literally the block appends and the old,
now-false requirement stays. The spec would have carried two byte-budget requirements, both
SHALL-normative, one wrong.

Caught only because the sync listed each operation against its match before applying and one row
read *no match*. `openspec validate --specs` passes either way. The panel reviewed the delta in
isolation, where it is correct — the defect exists only relative to the target spec.

### 4. Stale measurements, fourth occurrence — KAN-91

| Claim | Actual |
|---|---|
| `pipeline.md` is 64,621 bytes — guide ×2, PR body, commit message | 64,701 |
| `finish-contract.md` is 26,246 — PR body | 25,630 |

Line 32 of the guide asks the operator to verify a number that was never true; line 63 computes the
headline percentage from it.

**The sequence is the finding.** KAN-82 had one. F4 in this change had one. F10 corrected F4's table
and a later fix *in the same round* made the correction stale. F10's remedy was to write the
regenerate-last rule into `tasks.md` and `design.md` — and the rule was then broken, in the same
change, in a different artifact. Three rounds of instruction have not held; the budget table is
correct precisely because a *script* regenerates it.

### 5. Two reviewers wrote to the live worktree — KAN-92

Lens C mutated the guard and restored it. The adversarial slot ran `git checkout --` over three live
unstaged edits and restored them from its own earlier output. Both disclosed; the parent verified
all three edits present and the guard byte-correct.

Nothing was lost **this time**. The adversarial slot's restore depended on it still holding the file
contents in context — a slot that had compacted would have destroyed work with no copy anywhere.

The dispatch never says the tree is live, and mutation testing is exactly how F1 and F2 were proven.
Blaming the reviewers would be wrong.

### 6. The spec's Purpose was not widened with its requirements — KAN-93

`myflow-contract-economy` now governs `SKILL.md` files; its Purpose still says "a contract file under
`skills/myflow-contracts/`". A delta carries requirement operations only, so nothing can update a
Purpose. Descriptive, not normative — but it is the first thing a reader meets.

### 7. Four stale remote refs from earlier changes — KAN-94

`refs/remotes/origin/openspec/` still holds kan-7, kan-10, kan-19 and kan-20. All four archived.
`check-cleanup-complete.sh` is scoped to one change name, so it cannot see its own historical
misses; the count only grows.

## Cost

Two `/myflow-do` panel passes of four slots each, five implementer dispatches, plus two fix waves
applied by the parent. All slots ran on Sonnet, all fixes on Opus.

**The expensive parts were not the reviews.** `pipeline.md` had to be read twice per pass because
30,351 tokens exceeds the read cap — which is the measurement the proposal opened with, and the
reason the change exists. Reading the file to shrink it cost more than reviewing the shrink.

The two run-2 bugs (problems 1 and 2) each cost a diagnostic detour with the branch already merged,
which is the worst place to spend that time.

## What went well

**Hoist-before-cut worked, and it was the right call.** Two passages inside the finish block are
depended on by files `/myflow-do` loads. Moving them would have put citations in front of a file
their readers never open — KAN-82's F1 reintroduced by design. The hoist costs ~7KB of the headline
cut, which is why `pipeline.md` landed at 64,701 rather than the ~55,000 the issue estimated. That
is a correct trade, not a shortfall.

**Five implementer dispatches each stopped and reported a real plan-or-spec defect** rather than
working around it. Task 1.3 exists because 2.1's implementer refused to create a dangling citation.
None of the five was catchable by a guard.

**Two findings were proven by mutation, not asserted.** F1 and F2 were test-coverage gaps: mutating
the code left all 17 harness cases green. F13's *fix* was then caught the same way — the first
attempt put the bold token and the path on separate lines, invisible to a line-scoped guard.

**Pass 2 escalated to the full roster automatically** at ~154 changed lines, past the 150 threshold,
and found F10 through F13 — four more, one Major. The escalation rule earned its place.

**F7 was resolved against the reviewer's recommendation and that is visible.** The adversarial slot
wanted `/myflow-info` to read `finish-contract.md`; the operator's recorded `myflow-info-core-only`
decision had already rejected exactly that. It got a pointer instead, and pass 2's principles slot
independently judged the override legitimate.

**The panel's two deviations are recorded rather than inferable.** `bugbot` and `security-review`
do not exist in this harness, so the defect-hunt slot ran as `general-purpose` with a prompt file
and its ledger entry says so; no security slot ran, and the panel record states why rather than
leaving it to an agent count.

## Automation candidates

Ranked by how much they would have caught here.

1. **A guard over numeric claims in the manual test guide** (KAN-91). Extract `<file> is <N> bytes`
   and check against `wc -c`. Extended to `tasks.md` and `design.md` it also catches F4 and F10 —
   three of this change's thirteen findings, plus the one that escaped the panel entirely.
2. **A delta-operation check** (KAN-90). Every `## MODIFIED` heading resolves to exactly one target
   heading; every `## ADDED` heading to none. Run it in `/myflow-do`, where a delta can still be
   fixed, not at run 2 where the branch has merged.
3. **A base-ref regression case** (KAN-88). A merged branch whose local base ref is deliberately
   behind the remote must not return `RUN1`.
4. **One line in the reviewer prompts** (KAN-92): the worktree is live; copy to a scratch path before
   mutating. Cheapest item on this list and it preserves the technique that found F1 and F2.
5. **An archived-change ref sweep** (KAN-94), reporting rather than deleting.

**`scripts/check-references.sh` deserves its own note.** It was green throughout this change and
proved nothing: it resolves a citation whenever a heading of that name exists, and the mirroring rule
guarantees one exists in both files. Four defects lived in exactly that gap, including a citation
carrying **no path at all**. KAN-84 already tracks this from KAN-82. Nothing here changes that
assessment — it accumulates evidence for it.

## Known incomplete, carried from the guide

- The skill splits gave ~9% against `pipeline.md`'s 27% — the rule-density ceiling
  `jira-integration.md` hit last change, since a paragraph mixing a rule with its justification stays
  in the core whole.
- `project-configuration.md` (45,258) and `workspace-isolation.md` (31,079) deliberately unsplit,
  now ratcheted where they stand.
- The `[ ! -r "$path" ]` readability pre-test has no independent coverage: removing it does not turn
  the harness red, because the `wc -c` guard behind it exits 2 for the same input. The exit-code
  contract holds; the line is not independently proven.
- 25 boxes in the manual test guide were never ticked. The guards and harnesses were run by the
  `/myflow-do` session; the operator did not work the guide, and integrated over the gate
  deliberately.
