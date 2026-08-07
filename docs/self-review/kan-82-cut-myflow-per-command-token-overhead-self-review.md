# Self-review — kan-82-cut-myflow-per-command-token-overhead

**Rating: 4/5 — good.** The operator's own words for that level: solid, with real friction — four
plan amendments mid-run, three fixes that introduced new findings, and a target that was never
reachable as specified.

**What landed.** `pipeline.md` 103,326 → 88,253 bytes and `jira-integration.md` 52,783 → 15,591, so
a `/myflow-start` run loads **103,844 bytes of these two files where it loaded 156,109 — a 33%
cut**. Two rationale appendices and one single-command file were created, and
`scripts/check-contract-budget.sh` now ratchets every contract file at achieved size + 25%.

**What did not.** KAN-82 asked for `pipeline.md` alone to fall from roughly 34k tokens to roughly
8k. It fell 15%. The gap is the subject of angle 1.

---

## 1. Problems, and the pipeline change that would avoid each

### The plan's verification method silently constrained the plan's ambition

The change's central safety property is a **line-multiset check**: sort the non-blank lines of the
original file, sort those of core + appendix, and prove nothing disappeared. It is cheap, needs no
parser, and it worked — it caught real losses and it proved the final state clean across ~1,800
changed lines.

It also made the plan's own instruction unexecutable. The corpus is hard-wrapped at ~100 characters,
so a sentence boundary almost never coincides with a line break. "Split at the sentence boundary" —
which the approved design required — means re-wrapping the lines that stay, and a re-wrapped line is
indistinguishable from a deleted one under a line multiset. Task 2.1's implementer discovered this by
trying it, and reported it rather than quietly re-wrapping.

The consequence was structural, not cosmetic: the achievable unit became a whole paragraph, and this
corpus writes most of its rules as *paragraph mixing a rule with its justification*. Those stay in
the core entire. `jira-integration.md` yielded 11% to its appendix; `pipeline.md` yielded 15%.

**The pipeline change:** `/myflow-start` should require a plan whose central verification method is
novel to be **exercised against a real sample before the plan is published** — not merely described.
One `sed`-and-`diff` against a single section would have exposed this in the planning stage, when the
design was still cheap to change, instead of in task 2.1 when two files had already been cut.

*Filed: no — this is the most valuable finding of the run and the least mechanical. It is recorded
here rather than turned into a ticket because the concrete form of the rule is not yet clear enough
to implement, and a vague ticket is worse than a precise paragraph.*

### Three fixes introduced the next finding

Of 27 panel findings, **three were created by fixes to earlier findings**:

| Fix for | Introduced |
|---------|------------|
| F11 — a broken symlink misreported the directory as empty | **F21** — adding `[ -L "$path" ]` made symlinks *readable*, so `wc -c` sized whatever they pointed at; a `.md` symlink to a file outside the tree had its byte count printed. A size check became a path oracle. |
| F6 — a hand-maintained count nothing checks | **F15** — the fix wrote the identical defect into `SKILL.md`, two files away, in the same round |
| F13 — a duplicated heading | **F23** — the title edit changed a file's size, invalidating the budget table |

The escalation rule handled this correctly and without being asked: the fix round exceeded 150
changed lines and touched a delta spec, so pass 2 ran the **full** roster rather than a targeted
subset. It found nine more findings. **A targeted re-run would have missed F21 entirely**, because
the security slot had raised no finding against the symlink code in pass 1.

**No pipeline change proposed.** The mechanism worked. What this run adds is evidence for its
threshold: 3 in 27 is the base rate of fix-introduced defects on a change of this shape, and it is
high enough that the full re-run earns its cost.

### Two guards were structurally blind to this change's central question

`scripts/check-references.sh` verifies a citation resolves to a **heading**. The mirroring rule this
change introduced guarantees every heading exists in *both* the core and its appendix — so every
self-citation resolved regardless of where the content went. The guard was green throughout while
three findings (F1, F2, F26) sat in exactly that gap.

The line-multiset check has its own blind spot, correctly documented and correctly exploited by the
adversarial slot: it sorts, so reordering is invisible; it drops blanks, so paragraph structure is
invisible; and it compares against the **union** of core and appendix, so a verbatim relocation
between them produces no output at all.

**Filed as KAN-84.**

### A checkbox can be ticked against a stale expectation

Task 6.1's `Expected` said "no `<` lines at all"; the position-word reversal made a non-zero count
correct, and the box was ticked anyway. Task 4.1's index check used a substring `grep` that matched
two paths which are not files in the directory, so its "Expected: empty" was unreachable — also
ticked. Neither hid a real defect. Both hid that **nobody re-ran the command as written**.

**Filed as KAN-86.**

### The SDD ledger was written where nothing collects it

`/myflow-do` requires a ledger and never says where it lives.
`scripts/preserve-session-records.sh` reads exactly one path. The mismatch produced
`skipped: … (absent)` at finish run 1 — true, and easy to read past, since a change with no
subagents legitimately has no ledger. The model-policy audit trail would have been destroyed with
the worktree.

**Filed as KAN-83.**

### The budget table drifted three times

Each time for the same reason: it is computed from a tree that every subsequent fix changes.

**Filed as KAN-85.**

---

## 2. Cost, and what would reduce it without losing quality

Roughly **20 subagent dispatches**: 3 implementers, 3 task reviewers, 13 panel slots across two
passes, 1 fix agent. Subagent token usage totalled on the order of 2.4M, and the review panel
dominates that by a wide margin.

**Was the panel worth it?** On this change, unambiguously. It found a **core** term left undefined in
a file every command loads, a path-traversal oracle in a new guard, three fix-introduced defects, and
two blind spots in the repository's own tooling. Pass 1 found 14 findings, pass 2 found 9 — declining,
but nowhere near zero, and pass 2's included the most severe security finding of the run.

**What would genuinely reduce cost:**

- **Smaller, more frequent fix rounds.** Pass 2 escalated to the full roster because the fix round
  was 441 lines — and it was that large mostly because a *rule* changed mid-round, forcing a sweep of
  ~26 position words. Had the rule been settled earlier (see angle 1), the fix round would have been
  small enough for a genuinely targeted re-run: slot 0 plus the four slots that raised findings,
  rather than seven.
- **Not what would reduce cost:** collapsing slots. This run collapsed primary and defect-hunt into
  one pass-2 prompt, which is a deviation from the panel contract, recorded as such. It saved one
  dispatch and muddied two lenses. Not a trade worth repeating.

**One cost this run did avoid:** self-review ran as a single combined pass over four angles, as the
contract requires. Four separate dispatches would have been its own finding under this very heading.

---

## 3. What went well, and how to reproduce it

**Implementers reported plan defects instead of working around them — three times, each correct.**
Task 1.1's declined a heading promotion that would have broken its own acceptance test and said so.
Task 2.1's found the sentence-splitting impossibility. Task 3.1's flagged a spec-versus-command
disagreement and resolved it by satisfying the stronger of the two. Every dispatch prompt carried an
explicit instruction that a reported plan defect is a **good outcome**, and by task 3.1 the prompt
could cite the two previous implementers who had done it. That framing is reproducible and cheap.

**Measure-first made disputes trivial to settle.** Every numeric claim in the plan carried a
`measured:` tag naming the command and the ref. When a reviewer disputed two budget rows, the dispute
was arithmetic, resolved in seconds, with no argument about whose number was right.

**Reviewers verified by breaking things.** Lens C reverted the `wc -c` fix and re-ran the harness to
confirm the new fixtures actually fail without it — testing the test. The security slot fed the guard
`$(touch /tmp/pwned)` as a filename rather than reasoning about whether it was safe. Both disproved
hypotheses as well as confirming them, and both are recorded, because a disproved hypothesis is
evidence too.

**The operator was asked at every genuine fork, and every answer changed the work.** Four decisions:
paragraph granularity, position-word deletion, overturning a withdrawal on new evidence, and moving a
disputed paragraph back. None was cosmetic.

---

## 4. What could be automated or moved to a script

| Candidate | Notes |
|-----------|-------|
| Budget-table regeneration | **KAN-85.** The clearest win: a mechanical fact, hand-maintained, wrong three times. |
| Classifying every `<` line in a partition diff | Currently manual inspection, repeated ~6 times across the run by different agents. A script could label each removed line *repointed citation* / *deleted position word* / **UNEXPLAINED** and exit non-zero on the last. This would have made F1 and F2 mechanical instead of relying on a reviewer noticing. |
| Heading-parity between a core and its appendix | Panel finding F14, withdrawn as a disclosed tradeoff. Worth revisiting if a third contract is ever split — one file pair is judgment, three is a pattern. |
| Detecting a stale `Expected` after a mid-run rule change | **KAN-86**, cheapest form: when a global constraint is amended, re-read every task's `Expected` for statements the amendment invalidates. |

**One anti-pattern worth naming rather than automating.** A verification command of mine reported
"no lines lost" from a `diff` that had errored and never run — `set -- $pair` failed to split, the
paths did not exist, and `|| echo "no lines lost"` printed reassurance on failure. The shape
`<command> || echo "<success message>"` is a false-pass generator, and this repository already cares
intensely about guards that pass having looked at nothing. It belongs in
`engineering-principles.md` as a named shape to avoid, not in a script.

---

## Findings filed

| Issue | Finding |
|-------|---------|
| **KAN-83** | The SDD ledger path is undocumented, so it can be written where preservation will not find it |
| **KAN-84** | `check-references.sh` cannot see a citation whose section was hollowed out |
| **KAN-85** | The contract budget table must be hand-regenerated, and drifted three times |
| **KAN-86** | A ticked checkbox can assert an `Expected` output its own command no longer produces |

Declined: none. The verification-method finding in angle 1 was not offered for filing — it is the
run's most valuable observation and its concrete form is not yet clear enough to implement, so it is
recorded here in full instead.
