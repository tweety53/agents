# Final review panel — kan-183-project-name-not-hashed-key

**Roster:** `light` (recorded). Required slots: Primary, Principles, Code review (low). Panel model:
`sonnet` for every slot; none was dispatched by `subagent_type`, so no entry reads
`unknown (agent-defined)`.

**Diff:** `.superpowers/sdd/final-review.diff`, `git diff c6f6dab` — 707 changed lines across 18
files. `scripts/check-panel-diff-size.sh` measured **707**, under the cap, exit 0.

**Optional slots — three triggers fired, and all three were declined by the operator.**

| Slot | Trigger that fired | Disposition |
|------|--------------------|-------------|
| Security | the diff adds SQL query construction (`ProjectKeysByDisplayName`) reached by a user-controlled value | **declined by the operator** |
| Adversarial | existing tests were modified, and behaviour changed in code that had tests | **declined by the operator** |
| Lens B | 707 changed lines, over the ~200 threshold | **declined by the operator** |

The operator was asked once, on this pass, and answered *required three only*. **That answer stands
for every later pass of this change and is not asked again** — re-asking a settled question is not
diligence. A conditional slot whose trigger fires on a later pass is recorded as declined under this
standing answer, never as a slot whose trigger never fired.

**What the declines cost, and what was done about it.** Security's ground — the change's only new
SQL, and the only place a user-controlled value reaches a query — was carried on Primary's dispatch
instead, together with Adversarial's ground of behaviour changing under existing tests. Primary
covered both explicitly. That is a substitution and not equivalent coverage, and it is recorded here
rather than left to be inferred.

## Pass 1

| Slot | Model | Result |
|------|-------|--------|
| Primary | sonnet | clean — SQL parameterised with the pattern as a literal, all three suffix patterns byte-identical, every changed caller updated, "exact key costs no query" and "empty is no filter" both hold |
| Principles — Merged lens | sonnet | 2 Minor (F1, F2) |
| Code review (low) | sonnet | 1 Major (F3) |

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Minor | `stats/internal/api/stats.go:217` | a third copy of the `-[0-9a-f]{8}$` pattern sits beside the SQL one inside the same Go module, crossing no boundary and carrying no note of why it is duplicated, unlike the documented TypeScript/Go pair |
| F2 | Principles | Minor | `stats/internal/api/stats.go:228` | the resolver's method signature is spelled out verbatim in three interfaces in the same package instead of the two stores embedding `projectResolver` |
| F3 | Code review (low) | Major | `stats/web/src/views/StateBoard.tsx:50` | the Project column's `accessor` returns the display name, so `DataTable`'s exact-match filter and substring search merge two projects whose keys differ only in the suffix — the same ambiguity the server refuses with a 400 |

findings-total: 6
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed

reproducers-total: 6
finding-reproducer: F1 grep -rn projectKeySuffixPattern stats/internal/api/stats.go
finding-reproducer: F2 grep -c ProjectKeysByDisplayName stats/internal/api/stats.go
finding-reproducer: F3 none — the slot supplied none, since demonstrating it needs two live projects whose keys share a basename; the inconsistency was established by reading `DataTable`'s filter against `resolveProjectParam`'s ambiguity refusal
finding-reproducer: F4 none — inducing it needs a database failure inside project resolution; the fake's own error field exists for exactly that and was never wired to a test, which is F6
finding-reproducer: F5 none — same precondition as F4
finding-reproducer: F6 grep -rn projectKeysByDisplayNameErr stats/internal/api

**F3 is the interesting one, and its fix reverses a decision this plan made.** Task 2 set the
column's `accessor` to the display name deliberately, so that sorting, search and the filter dropdown
would agree with what the screen shows. The consequence nobody traced is that `DataTable` filters and
searches on that same value, so two projects the server treats as ambiguous become one filter entry —
the client silently merging what the server refuses. The fix returns the `accessor` to the full key,
leaving the display name in `render` with the key as the cell's `title`: the column reads `agents`,
the filter and search stay keyed to identity, and the hash remains searchable. The cost, stated
plainly, is that the column's exact-match dropdown lists full keys — which is the correct reading for
a control whose job is to tell two same-named projects apart.

## Pass 2 — targeted re-run

**Mode:** targeted. Slot 0 as the integration check, plus every slot that raised a finding — which is
all three, so the same roster ran. Nothing escalated it: no delta spec, migration or guard behaviour
changed, the fix stayed inside the files the findings named, and no new Critical appeared.

**Slots:** the three required. The operator's pass-1 answer — required three only — stands for every
pass of this change and was **not** asked again. Security, Adversarial and Lens B remain declined
under that standing answer.

| Slot | Model | Result |
|------|-------|--------|
| Primary | sonnet | clean — F1, F2, F3 each verified fixed; regenerated the interpolated SQL and confirmed it byte-identical to the prior inline literal |
| Principles — Merged lens | sonnet | clean — F1 and F2 hold, and the `fmt.Sprintf` judged safe on its merits rather than accepted because it closed the finding |
| Code review (low) | sonnet | F3 closed; **3 new findings** (F4, F5, F6) |

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F4 | Code review (low) | Major | `stats/internal/api/changes.go:320` | `list()` maps every `parseChangeQuery` error to 400, but that function now performs a store call, so a database failure surfaces as a 400 carrying raw store text instead of a logged 500 through `mapStoreError` |
| F5 | Code review (low) | Major | `stats/internal/api/stats.go:323` | the same defect in `view()` and `listModels()`: `parsePeriodAndProject` can now fail from a store call, and both sites classify its error as a client error |
| F6 | Code review (low) | Minor | `stats/internal/api/changes_test.go:72` | `projectKeysByDisplayNameErr` was added to both fakes to simulate a failing resolver, but no test ever sets it, so F4 and F5 ship with no regression coverage |

**What pass 1 missed, and why it is worth naming.** Task 3 changed two pure parsing functions into
functions that perform **I/O**, and neither call site's error classification was revisited. Pass 1's
reviewers checked that every caller was *updated to compile* and that behaviour was preserved for
well-formed input; none asked what the new failure mode meant for the existing "any parse error is a
400" rule. A signature change that adds a failure *kind* is exactly where that question belongs.

**The three findings' own rewritten pre-existing test was judged honest by two slots independently** —
Primary and Code review both read it as mechanically forced by F3's fix rather than loosened to
accommodate it.

## Pass 3 — targeted re-run

**Mode:** targeted, after fix round 2. Slot 0 as the integration check plus the slot that raised
F4-F6; Principles ran too, since fix round 2 touched code it had judged and its pass-2 clean result
would otherwise be stale. Nothing escalated it: no Critical appeared, no guard, migration or delta
spec changed, and the fix stayed inside the files the findings named.

**Slots:** the three required, under the operator's standing pass-1 answer. Not asked again.

| Slot | Model | Result |
|------|-------|--------|
| Primary | sonnet | clean — walked every error path out of both parse functions; independently reproduced the pre-fix defect by applying only the test hunk to `6b2bc23` in a throwaway worktree and observing the 400 with `connection refused` in the body |
| Principles — Merged lens | sonnet | clean — the wrapper follows a convention already present in this package (`stages.ErrUnknownStage`'s `errors.As` shape) rather than adding a second one, and the helper's three call sites each previously carried the same defect |
| Code review (low) | sonnet | clean — F4, F5, F6 each verified closed, with the `mapStoreError` default case confirmed to yield 500 and a generic body |

**Zero open findings.** Six findings, all fixed, none withdrawn.

**Every slot's clean result is current.** All three read the branch as it stands at `22f3fd2`, so
nothing here rests on a stale pass — the qualification KAN-182's handoff had to carry does not apply.
