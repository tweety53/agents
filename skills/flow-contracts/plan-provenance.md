# Plan provenance

**This file is canonical for plan provenance.** Skills and guards reference it by name; none of
them restate the tag vocabulary. If a rule below and a skill ever disagree, this file wins.

## The four tags

Every fenced code block in a plan carries a provenance tag on its info string, and every numeric
claim carries one in an HTML comment within the two lines after it; an assumption the plan could
not verify carries one in the task that depends on it:

- `verified:<how>` — on a fenced block's info string. The snippet was checked against something
  real; `<how>` names the check, e.g. `verified:javap intellij.platform.diff.jar`.
- `unverified:<what-to-check>` — on a fenced block's info string. The snippet is the implementer's
  best guess, not a checked fact; `<what-to-check>` names exactly what to confirm before trusting
  it, e.g. `unverified:confirm the member is a property, not a function`. The tag also sits on a
  prose assumption: a claim about code or environment a task depends on that the plan could not
  verify at plan time carries `unverified:<what-to-check>` in that task, and the implementing task
  confirms it before building on it.
- `measured:<command> @ <ref>` — in an HTML comment within the two lines after a numeric claim
  (the claim's own line, the line after it, or the line after that — so a blank note line between
  claim and comment is tolerated). The number came from actually running `<command>` at `<ref>`
  (a commit, tag, branch, or other named point), e.g. `<!-- measured: ./gradlew test @ c515c42 -->`.

  **Choosing a ref while the work is uncommitted.** A plan under `/flow`'s implement phase sits on a branch whose
  commits do not exist yet, so naming the merge base is worse than useless: the commands being cited
  frequently do not exist there, and `git cat-file -e <merge-base>:<script>` fails outright. Name
  the **branch** (`@ branch spectre/<change-name>`), which resolves both while the work is in
  flight and after it merges. Name a commit only when the measurement really was taken at that
  commit and the command really does exist there — and say which, as in
  `@ merge-base d38372a (the count BEFORE this change)`. When a measurement genuinely cannot be
  re-run — it read a machine-local file, or observed a live event — say that in the tag instead of
  naming a ref that implies otherwise.
- `predicted:<what-confirms-it>` — in an HTML comment within the two lines after a numeric claim,
  same window as `measured:` above. The number is an expectation, not a measurement;
  `<what-confirms-it>` names what would confirm it, e.g.
  `<!-- predicted: ./gradlew test after task 1 -->`.

## Tag syntax examples

````markdown verified:authored in-tree for this change
```kotlin verified:javap intellij.platform.diff.jar
override val toolWindowIds: Array<String>
```

```kotlin unverified:confirm the member is a property, not a function
override fun getToolWindowIds(): Array<String>
```

Baseline: 197 tests, 0 failures
<!-- measured: ./gradlew test @ c515c42 -->

After the deletion: 186 tests
<!-- predicted: ./gradlew test after task 1 -->
````

## The asymmetry rule

Every fenced code block in a plan needs a tag — `verified:<how>` or `unverified:<what-to-check>` —
but a number with no tag may not appear in a plan **at all**. A block can be labelled `unverified`
and stay in the plan, because it is honest about its own uncertainty. A number has no such escape
hatch: an untagged number is not "unverified", it is unattributed, and an implementer reading it
cannot tell whether it was run or guessed.

## The implementer's duty

Because the guard cannot verify truth, the obligation falls on whoever writes the tag: write
`verified:<how>` only after actually performing `<how>`, and write `measured:<command> @ <ref>`
only after actually running `<command>` at `<ref>` and reading the result. Writing either tag
without doing the check it names is worse than leaving the block `unverified` or the number
`predicted` — it tells the next reader a check happened when it did not, which is exactly the
failure this contract exists to prevent. A prose `unverified:` assumption in a task carries the
same duty on the reading side: the implementing task confirms it before building on it.

## When a measurement contradicts the plan

A plan is an argument, not a script. An implementer who measures something that contradicts the plan
reports the measurement and stops, rather than following the plan into a wrong result. The plan or
the design is amended before the work continues. Silent deviation and silent compliance are both
failures: the first is indistinguishable from a mistake, the second wastes the measurement.

This is where the tag vocabulary comes due. A guess labelled `unverified:` or `predicted:` is
honest; the same guess after a run has disproved it is a defect in the plan, and the plan is where
it is fixed — not routed around in the implementation, and not obeyed anyway. The amendment is
that correction written into the task record where the disproved claim sits; which half of the
exchange discloses it and which transcribes it, and how the review gate stays armed across a
corrected `**Files:**`, is [the record carries its own corrections](../flow/implement.md).
