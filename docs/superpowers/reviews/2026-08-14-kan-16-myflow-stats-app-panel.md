# Final review panel — kan-16-myflow-stats-app, fix round 1 (tasks 18–25)

## Diff-size gate

`scripts/check-panel-diff-size.sh . 2f643f4` measured **34,326** lines and exited over-cap. Put to
the operator, with the alternative of scoping the panel to this fix round alone (`c2357c5..HEAD`,
7,264 lines), the commits below `c2357c5` having already passed a clean full-roster panel over 21
passes. **The operator chose the whole branch**, against that recommendation. Recorded rather than
summarised, because the more expensive scope was an explicit choice and it is what produced F1.

## Roster

`reviewPanelRoster: full`. Operator selected the **required three only**.

| # | Slot | Model | Note |
|---|------|-------|------|
| 0 | Primary | sonnet | plan alignment + code quality |
| 1 | Defect hunt | sonnet | **substituting for Bugbot** — no `bugbot` agent type exists in this harness; carries the mutation-testing brief verbatim |
| 2 | Principles | sonnet | `[LENS]` = Merged |
| 4 | Security | — | **declined by operator** (trigger fired: migration, config, path handling) |
| 5 | Adversarial | — | **declined by operator** (trigger fired) |
| 6 | Lens B | — | **declined by operator** (trigger fired) |
| 6 | Lens C | — | **declined by operator** (trigger fired) |

A declined slot is recorded distinctly from one whose trigger never fired, and an escalation does not
silently reinstate it.

## Passes

Five passes, four fix rounds. Pass 2 escalated to Full automatically — the fix touched files outside
the set the findings named. Every fixup was folded with `git rebase --autosquash`, so each pass's
results were also stale by sha before the next began.

| Pass | Primary | Defect hunt | Principles |
|------|---------|-------------|------------|
| 1 | 0 findings — **rejected by the dispatcher**, then F1 on re-read | 5 findings, 6 mutations, 4 survived | clean |
| 2 | clean | F7, F8 — both in the fix round's own new code | clean |
| 3 | F9 | clean — both closed, verified by re-mutation | clean |
| 4 | clean | clean — swept for F9's siblings, found none | F10 |
| 5 | clean | not re-run — subject unchanged | clean — F10 closed |

**Pass 1's Primary result was rejected by the dispatcher, not by a reviewer.** A clean verdict on
35,718 lines citing only the newest 7,264 of them is not a whole-branch review, and recording it
would have made the operator's explicit choice of scope purely nominal. The re-read found F1.

## Findings

| ID | Pass | Slot | Severity | Location | Note |
|---|---|---|---|---|---|
| F1 | 1 | Primary | Major | `stats/internal/store/query.go:171` | `metricsKeySegmentRE` forbade hyphens, so no recorded model name was addressable as a metrics key path, leaving the spec's filterable/searchable/sortable requirement unsatisfiable for the `models.<model>` buckets the schema itself introduced |
| F2 | 1 | Defect hunt | Major | `stats/web/src/metrics.ts:113` | token readers read `tokens.main`/`sidechain` as flat numbers while the harvester writes `Bucket` objects — **six readers, not the two the finding named**; every token column in the run-detail table rendered Unavailable for every real run |
| F3 | 1 | Defect hunt | Minor | `stats/internal/harvest/attribute.go:361` | `bestWindow`'s highest-`Attempt` tie-break survived inversion |
| F4 | 1 | Defect hunt | Minor | `stats/internal/harvest/watcher.go:222` | the `!applied` race-loss guard survived neutralisation |
| F5 | 1 | Defect hunt | Minor | `stats/internal/store/aggregate.go:616` | `CountRunsWithoutModel`'s `models = '{}'::jsonb` clause survived removal |
| F6 | 1 | Defect hunt | Minor | `stats/internal/store/aggregate.go:369` | `CacheEfficiency`'s divide-by-zero guard survived removal, admitting a `NaN` ratio |
| F7 | 2 | Defect hunt | Minor | `stats/internal/harvest/wireshape_test.go:18` | the drift guard was blind to `omitempty` fields left at zero — **and was already blind to three fields task 23 had added hours earlier**, while its comment claimed otherwise |
| F8 | 2 | Defect hunt | Minor | `stats/internal/store/query.go:158` | `metrics.tokens.cache_read` was the documented example and four tests' path, but no real stage run has it |
| F9 | 3 | Primary | Minor | `stats/internal/store/query_test.go:340` | the tests guarding *absence is never zero* could not fail: the seed order gave the tiebreaker the expected answer on a tie |
| F10 | 4 | Principles | Important | `stats/internal/store/query_test.go:345,387` | F9's fix left its own invariant recorded only in a comment, so a future reorder would silently restore the masking |

findings-total: 10
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

reproducers-total: 10
finding-reproducer: F1 none — the defect was that a legitimate key shape was rejected; the test asserting the rejection passed, so no command demonstrated it without a new case
finding-reproducer: F2 none — required adding a metrics.test.ts case feeding the readers the real nested shape; reproduced and reverted in an isolated copy
finding-reproducer: F3 none — surviving mutant; demonstration is the mutation itself
finding-reproducer: F4 none — surviving mutant; demonstration is the mutation itself
finding-reproducer: F5 none — surviving mutant; demonstration is the mutation itself
finding-reproducer: F6 none — surviving mutant; demonstration is the mutation itself
finding-reproducer: F7 none — required adding an omitempty field to internal/harvest.Bucket and leaving it zero in the fixture patch
finding-reproducer: F8 grep -n metrics.tokens.cache_read internal/store/query.go internal/store/query_test.go
finding-reproducer: F9 none — a structural gap in discriminating power, proven instead by injecting COALESCE into metricsNumericExpr and observing both tests fail
finding-reproducer: F10 none — proven by reverting the seed order and observing the new guard fire

Every `none — <reason>` here is inherent rather than an omission: a surviving mutant is *defined* by
the absence of a failing test, so the demonstration is the mutation; and F1, F2, F9 and F10 are all
correct-looking tests asserting the wrong thing, which has the same property. F9 and F10 were each
proven twice — once by the fix agent and once independently by the dispatcher.

## The pattern

Every finding is one species: **something trusted further than its actual coverage.** The production
code was rarely wrong; the confidence attached to it was.

| | What was trusted | What it actually covered |
|---|---|---|
| `Price` (task 23) | tests calling it directly | nothing in production called it |
| `aggregate.go` (task 24) | each side's own fixtures | the two sides disagreed about the shape |
| `metrics.ts` (F2) | a hand-written fixture | six readers, all reading a shape never written |
| the drift guard (F7) | value comparison | silent on `omitempty` — including three fields that already existed |
| the absence tests (F9) | an order assertion | an assertion the tiebreaker satisfied by coincidence |
| F9's own fix (F10) | a doc comment | knowledge nothing enforced |

Each was found by a different method, and none by the method that found the previous one. That is
the argument for the roster, and against reading any single clean slot as sufficient.

## Dispatcher errors, recorded

**"Exactly 8 commits in `2f643f4..HEAD`" was wrong, and was asserted to reviewers across three
dispatches.** `2f643f4` is the merge base with `main`, so that range is the whole branch — 27
commits. This round's 8 are `c2357c5..HEAD`. The dispatcher introduced it by verifying with
`git log --oneline 2f643f4..HEAD | head -8` — a command that cannot print more than 8 lines — and
reading the output as confirmation. **It is this panel's own defect class, committed in the
dispatcher's verification**, then laundered into a confirmed fact by a reviewer being asked to
confirm it. Corrected to both affected slots rather than silently fixed.

**A suspicion raised and disproved:** the half-open `[start, end)` convention was queried and found
applied identically in all nine period-restricted queries. Recorded so it is not re-raised.

## Known limitations, accepted rather than fixed

- **The structural fixture guard covers `harvest.Bucket` only**, not `ModelBucket` or `MetricsPatch`;
  an `omitempty` field added to either escapes both tests, verified by mutation. Accepted because the
  comments scope the guarantee honestly, no SPA reader depends structurally on those types, and a
  non-`omitempty` addition is still caught. **The original defect was never incomplete coverage — it
  was a comment claiming coverage it did not have.**
- **`/api/v1/stage-runs` cannot filter by model server-side**; task 25 states a truncation shortfall
  in the panel rather than silently showing a short page. Worth a follow-up issue.
