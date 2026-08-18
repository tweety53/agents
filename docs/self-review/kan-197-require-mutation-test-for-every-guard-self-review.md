# Self-review — kan-197-require-mutation-test-for-every-guard

**Date:** 2026-08-18
**Issue:** [KAN-197](https://tweety53.atlassian.net/browse/KAN-197)
**PRs:** [#12](https://github.com/tweety53/agents/pull/12) (implementation), [#13](https://github.com/tweety53/agents/pull/13) (archive)
**Operator rating:** 4 / 5 — good
**Command:** `/myflow-fast`, three invocations (create → integrate → archive)

One combined pass over four angles. Context gathered by `scripts/gather-self-review-context.sh` —
4 of 4 sources found. **Every angle produced findings and every angle was offered for filing**, per
[KAN-200](https://tweety53.atlassian.net/browse/KAN-200), and each finding was explained in full
before anything was filed.

---

## Problems · `myflow-fix`

### P1 — the change shipped the exact defect it was built to prevent

`check-stage-mark-calls.sh` had a target that enumerated to zero files. It never called
`coverage_record`, and `coverage_verdict` reasons only over members that *were* recorded — so the
target vanished from the verdict and the guard reported clean at exit 0.

That is the vacuous pass KAN-197 exists to close, reproduced inside the fix for it. Raised as
Critical by the panel's low-effort slot with a runnable reproducer, and fixed in the same change.

The general lesson is worth stating beyond this instance: **a component that reasons only over what
it was told, never over what it should have been told, has this defect available to it by
construction.** It appeared twice here — once as an unrecorded member (F7) and once as a declaration
for a member never recorded (F3). Both were one root cause.

*Not filed — fixed within the change, and the lesson is recorded here.*

### P2 — verification proved the tests pass without proving they run

The run verified with `for t in scripts/test-*.sh` and reported "all 25 harnesses pass". The panel
then found `scripts/test-check-stage-mark-calls.sh` had never been in `.myflow/project.md`'s
`## test` list, since KAN-172 — so it passed, and would never have run in the declared suite. This
change had just added ~75 lines of assertions to it.

The verification method had the identical blind spot as the guard it was verifying: a check that
passes tells you nothing unless you know it ran over what it claims to cover.

*Filed as [KAN-204](https://tweety53.atlassian.net/browse/KAN-204).*

## Cost · `myflow-cost`

**7 dispatches, ~1.25M subagent tokens** — against KAN-73's 12 / ~1.98M for comparable output,
roughly 37% less. Two repeatable causes:

- **The shared library was extracted before its adopters existed.** KAN-73 let `resolve_file` reach
  five copies and spent a review round plus a fix round pulling it back. Here `scripts/lib/coverage.sh`
  was task 1 and four guards adopted it, so no duplication was created and none had to be removed.
- **Independent tasks were bundled.** Tasks 3, 4 and 5 shared no file and went to one implementer
  producing three commits, instead of three cold starts.

The panel cost 302K and returned 8 findings including a Critical; the fix round then cost **285K** —
nearly the whole panel again — starting cold on evidence the panel had already located, with every
finding carrying a file, a line and a reproducer.

*Appended to [KAN-201](https://tweety53.atlassian.net/browse/KAN-201) as a second datapoint rather
than filed separately.*

## What went well · `myflow-improvement`

**W1 — the `unverified:` provenance tag did real work, twice.** The design guessed the coverage
numbers would be 14/17/7; those were raw symlink counts, and the correct metric was the
post-delegation required set, 13/16/6. The implementer established real behaviour and reported the
discrepancy instead of transcribing the guess — and separately found `openspec-explore` was a corpus
member the plan had missed and `myflow-contracts` was not a member at all.

**W2 — the two panel slots disagreed, and both were right about different guards.** Principles wanted
the corpus narrowed so non-candidates stop being members; Primary wanted it kept and the boilerplate
reason strings fixed. Narrowing suits `check-stage-mark-calls`, where non-candidacy is structural;
better reasons suit `check-references`, where it is not. Routing each remedy where it applied beat
picking a winner.

*Appended to [KAN-198](https://tweety53.atlassian.net/browse/KAN-198), with the observation that a
per-slot finding count would record both slots as "1 Major" and miss that the disagreement was the
productive part.*

**W3 — the fix agent under-claimed its own work.** It disclosed a residual on F3 and asked for
review; testing showed the residual does not exist on the default scan. Erring toward under-claiming
is the behaviour to reinforce.

## Automation · `myflow-automation`

**A1 — three prior follow-ups were applied and all three worked.** This is the first evidence any of
them earned their filing:

| Follow-up | Applied here | Outcome |
|---|---|---|
| [KAN-192](https://tweety53.atlassian.net/browse/KAN-192) | ledger written to `.superpowers/sdd/tasks/progress.md` | preserved on the first run; KAN-73 needed a second |
| [KAN-195](https://tweety53.atlassian.net/browse/KAN-195) | fetched and fast-forwarded `main` before the preflight | `RUN2` first try; KAN-73 got a false `RUN1` |
| [KAN-202](https://tweety53.atlassian.net/browse/KAN-202) | used `scripts/commit-split.sh` | correct two-commit split, no hand-written chain |

KAN-202 is worth noting specifically: the contract still does not name the script, so this run used
it because the ticket existed, not because the pipeline said to. That is the gap the ticket asks to
close.

*Not filed — the evidence belongs on the three existing tickets.*

---

## Verification

25/25 harnesses · 10/10 lint guards · all four guards exit 0 on the real tree · 5/5 commit-fields
guards on the rewritten shas · `check-unfinished-work` `CLEAR` · `check-cleanup-complete` `COMPLETE`.

Every guard claim was mutation-tested rather than trusted: undeclared zero, empty corpus, declared
zero, duplicate target, ghost declaration, a declared member renamed away, and a `SKILL.md` stripped
of its marks. Tree restored and confirmed clean after each.
