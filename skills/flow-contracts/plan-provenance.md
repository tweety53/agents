# Plan provenance

**This file is canonical for plan provenance.** Skills and guards reference it by name; none of
them restate the tag vocabulary. If a rule below and a skill ever disagree, this file wins.

## The four tags

Every fenced code block in a plan carries a provenance tag on its info string, and every numeric
claim carries one in an HTML comment within the two lines after it:

- `verified:<how>` — on a fenced block's info string. The snippet was checked against something
  real; `<how>` names the check, e.g. `verified:javap intellij.platform.diff.jar`.
- `unverified:<what-to-check>` — on a fenced block's info string. The snippet is the implementer's
  best guess, not a checked fact; `<what-to-check>` names exactly what to confirm before trusting
  it, e.g. `unverified:confirm the member is a property, not a function`.
- `measured:<command> @ <ref>` — in an HTML comment within the two lines after a numeric claim
  (the claim's own line, the line after it, or the line after that — so a blank note line between
  claim and comment is tolerated). The number came from actually running `<command>` at `<ref>`
  (a commit, tag, branch, or other named point), e.g. `<!-- measured: ./gradlew test @ c515c42 -->`.

  **Choosing a ref while the work is uncommitted.** A plan under `/myflow-do` sits on a branch whose
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

**Why the asymmetry, and not just "label everything":** in a prior run (KAN-6, an IntelliJ
plugin), a plan claimed "194 tests" as a baseline. That number was invented, not mislabelled —
nobody ran the suite and wrote down a stale count; the figure never came from a run at all.
Labelling alone would not have caught it, because a fabricated number can be labelled exactly as
confidently as a real one. What catches it is refusing the number a home unless the label names a
command and a ref that can be re-run — `measured:<command> @ <ref>` forces the claim to point at
something checkable, and `predicted:<what-confirms-it>` forces it to name the check that has not
happened yet. A plan that cannot produce either for a number should not state the number.

The same run also carried the motivating failure for the block rule: a plan snippet said a marker
key could be read from `DiffRequest`. It cannot — the platform never propagates chain user data to
requests. Had the implementer transcribed that snippet verbatim, the feature's guard would have
been permanently false and the whole feature a no-op **that every unit test still passed** — a
false guard produces no failing assertion, only a feature that silently never fires. It was caught
only because that one task happened to carry a hand-written instruction telling the implementer to
check. The tag vocabulary exists so that instruction is never a special case again: every snippet
either says how it was checked, or says plainly that it was not.

## The implementer's duty

Because the guard cannot verify truth, the obligation falls on whoever writes the tag: write
`verified:<how>` only after actually performing `<how>`, and write `measured:<command> @ <ref>`
only after actually running `<command>` at `<ref>` and reading the result. Writing either tag
without doing the check it names is worse than leaving the block `unverified` or the number
`predicted` — it tells the next reader a check happened when it did not, which is exactly the
failure this contract exists to prevent.

## When a measurement contradicts the plan

A plan is an argument, not a script. An implementer who measures something that contradicts the plan
reports the measurement and stops, rather than following the plan into a wrong result. The plan or
the design is amended before the work continues. Silent deviation and silent compliance are both
failures: the first is indistinguishable from a mistake, the second wastes the measurement.

This is where the tag vocabulary comes due. A guess labelled `unverified:` or `predicted:` is
honest; the same guess after a run has disproved it is a defect in the plan, and the plan is where
it is fixed — not routed around in the implementation, and not obeyed anyway.
