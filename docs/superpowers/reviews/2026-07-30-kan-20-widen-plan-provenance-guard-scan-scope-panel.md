# Review panel — kan-20-widen-plan-provenance-guard-scan-scope

**Pass 1 — full roster.** Diff read by every slot: `.superpowers/sdd/final-review.diff`
(whole branch, merge-base c1f5aa6 -> working tree; 1878 insertions / 201 deletions, 9 files).

## Roster selection (evaluated against the diff BEFORE dispatch)

| Slot | Included | Why |
|------|----------|-----|
| 0 Primary | yes | always required |
| 1 Bug hunter | yes | always required |
| 2 Principles (Merged) | yes | always required |
| 3 Security | yes | trigger: path/file handling — containment gate generalised over symlinks, path escapes, non-regular files |
| 4 Adversarial | yes | triggers: >~300 changed lines; existing tests modified (4 exact-string assertions rewritten) |
| 5 Principles Lens B | yes | trigger: >~200 changed lines; new module-level helpers introduced |
| 5 Principles Lens C | yes | trigger: error handling / operator output reworked per-candidate |

No slot excluded. Every conditional trigger fired.

## Harness substitution (recorded)

This harness exposes no `bugbot` or `security-review` subagent_type. Slots 1 and 3 therefore ran as
general-purpose agents reading myflow's own `bug-hunter-reviewer-prompt.md` and
`security-reviewer-prompt.md`, which exist for this fallback. All slots ran on Sonnet.

[PRINCIPLES_PATH] /Users/tweety53/.claude/skills/myflow-do/engineering-principles.md
[STANDARDS_PATHS] <worktree>/CLAUDE.md, <worktree>/AGENTS.md — both form 2, containment-checked,
  none dropped. (Both slots noted these files declare no project-specific standards beyond the
  global fix-first lint policy.)

## Results

| Slot | Verdict | Critical | Important | Minor |
|------|---------|----------|-----------|-------|
| 2 Principles (Merged) | principles-compliant | 0 | 0 | 2 (both named as accepted tradeoffs) |
| 5B Principles Lens B | principles-compliant | 0 | 0 | 2 |
| 5C Principles Lens C | principles-compliant | 0 | 0 | 2 |
| 1 Bug hunter | NOT ready | **1** | 0 | 1 |
| 3 Security | with fixes | 0 | **1** | 0 |
| 4 Adversarial | ready | 0 | 0 | 2 |
| 0 Primary | pending | | | |

## Open findings for the fix wave (union, deduped)

1. **CRITICAL (slot 1)** `_code_span_regions` returns `[]` for BOTH "no code spans on this line" and
   "code-span state indeterminate". `_is_quoted` passes that `[]` to `_quote_regions` as the skip
   list, so quote characters that are really code-span content become live delimiters and can pair
   into a bogus region enclosing a bare asserted number.
   Reproduced end-to-end by the controller: a change whose tasks.md contains
   `` `a` "b` and 7 tests failed" c `` exits 0 with the claim unreported.
   Same fail-open class as the original Task 2 Critical, re-entering via the veto's own conflation.
2. **IMPORTANT (slot 3)** `_quote_regions` calls `_within(i, skip)` per character; `_within` scans
   the skip list linearly -> O(N^2) in the number of code spans on one line. Controller-measured on
   this tree: 500 pairs 0.05s, 1000 0.21s, 2000 0.84s, 4000 3.38s. A crafted line well inside the
   10 MiB cap stalls a CI gate that runs over untrusted PR content. NOTE: this file's own docstring
   already records two prior quadratics ("the third quadratic this guard has had to fix").
3. **MINOR, fixing anyway (slot 4)** `skills/myflow-contracts/plan-provenance.md` claims the veto's
   cost was measured "across every Markdown file in this repository" and that not one genuine
   quotation loses its exemption. Slot 4 found ~165 repo-wide markdown lines with an odd `"` count
   that the stated evidence never exercised. An overclaim, in the contract that forbids overclaiming.
4. **MINOR (slot 4)** No regression pin on the three-message operator output for a chmod-000 change
   directory (case 51 pins only rc=2).

---

# Pass 2 — FULL re-run (escalated automatically)

**Why full, not targeted.** myflow escalates a re-run from targeted to full, without asking, when
the fix diff exceeds ~150 changed lines. Fix round 2 is 275 changed lines (3 files). The escalation
is automatic and is recorded here rather than decided.

**What happened between passes.**
- Pass 1 raised 2 Critical + 2 Important + 3 Minor. One fix wave addressed all seven.
- The scoped re-review of that wave verified all seven ADDRESSED with mutation proofs, found no new
  breakage — and found a FOURTH fail-open: backslash-escaped delimiters treated as live delimiters
  (CommonMark 2.4 makes them literal). Evidence of completeness: an escape-aware oracle differentially
  fuzzed 300k lines; every divergence contained a backslash; zero divergences without backslashes.
- Operator was asked (SDD says no second fix wave; myflow says never hand off with an open Critical)
  and chose to fix. Fix round 2: veto the line when an odd backslash run precedes a delimiter.
  Striking escapes was rejected as circular — a code span gets no escape processing, so escape
  validity depends on the span structure the strike would have to precede, and "two passes that must
  agree" is the shape of all three earlier fail-opens.
- Post-fix differential fuzz: 11,813 fail-open shapes before -> **0** after, over 700k lines / 5 seeds.

Every slot re-reads the rewritten `.superpowers/sdd/final-review.diff`.

## Pass 2 results

| Slot | Verdict | Critical | Important | Minor |
|------|---------|----------|-----------|-------|
| 0 Primary | with fixes | 0 | 1 (artifact/spec drift) | 0 |
| 1 Bug hunter | NOT ready | **1** (raw-HTML fail-open, the FIFTH) | 0 | 1 |
| 2 Principles (Merged) | compliant | 0 | 0 | 2 |
| 3 Security | ready | 0 | 0 | 0 |
| 4 Adversarial | ready | 0 | 0 | 0 |
| 5B Lens B | — | 0 | 1 (_region_at precondition unenforced) | 1 |
| 5C Lens C | — | 0 | 1 (no veto signal in operator message) | 1 |

**Panel-induced flakiness, resolved.** Three slots independently observed transient harness failures
(cases 207a, 219) and one caught the guard mid-mutation with `_has_escaped_delimiter` replaced by
`if False:`. Cause: reviewers mutation-testing the file concurrently, exactly as instructed. Not a
product defect. Recorded so a future reader does not re-investigate it.

**Fix wave 3 — the design change.** The fifth fail-open was the fifth in a row inside CODE-SPAN
handling. Measurement showed every false positive the exemption was ever built for is double-quote
delimited; none needed code spans. Operator ruled: drop code spans from the delimiter set. That
deletes the CommonMark-inline-parsing surface rather than patching it a fifth time.
Post-wave differential fuzz vs an independent escape-aware/HTML-aware oracle: 0 fail-open
divergences over 271,869 claim matches; the same fuzz finds 13,619 fail-opens against the pre-wave
guard, so it is not vacuous.

## Final outcome

Zero open Critical/Important. Six fail-opens found and fixed across the passes, every one in the
quotation exemption, every one traceable to approximating a CommonMark grammar with a regex. The
last two rounds removed grammar rather than adding a mechanism: code spans dropped from the
delimiter set (no measured false positive ever needed them), then the raw-HTML grammar replaced by
"the line contains `<`". What remains is auditable by reading: escape-parity veto, angle-bracket
veto, left-to-right pairing per quote class.

Methodological note worth keeping: fix wave 3 reported "0 fail-open divergences" over 271k lines and
that number was true and nearly meaningless — its corpus contained no HTML attributes, so it could
not express the defect that was sitting in the tree. Fix wave 4's report states its token set
alongside its count for exactly this reason.
