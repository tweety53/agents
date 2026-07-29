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

  **The guard checks only that a `measured:`/`predicted:` comment is present.** It does not parse
  the `<command>` or the `@ <ref>` half — so that shape is a *convention this contract states and a
  human enforces*, not something the script can hold you to. It is written down here as the shape
  to follow, and named as advisory so the contract is not defined two ways: a plan whose comment
  omits the ref passes the guard and is still wrong under this contract.

  **Choosing a ref while the work is uncommitted.** A plan under `/myflow-do` sits on a branch whose
  commits do not exist yet, so naming the merge base is worse than useless: the commands being cited
  frequently do not exist there, and `git cat-file -e <merge-base>:<script>` fails outright. Name
  the **branch** (`@ branch openspec/<change-name>`), which resolves both while the work is in
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

## The guard's scope, and why it is narrow

The guard that enforces this contract (task 2) reads only a change's own `tasks.md` — the plan
being implemented right now — and explicitly excludes `openspec/changes/archive/`. It does not
scan the whole repository, and it does not scan other changes' plans.

That narrowness is deliberate, not an oversight. `scripts/check-references.sh`'s own header
records that an earlier guard in this repository, once widened past the shape it was designed for,
produced 28 false failures on this repo's own tree — none of them a genuine defect. The only way
to silence a false failure under time pressure is a suppression marker, and a suppression marker
placed to silence a false hit also switches off whatever real check shares that line. A guard that
over-fires does not just annoy; it manufactures the conditions for its own defeat. Scoping this
guard to exactly the file it exists to police — the live plan being acted on — is what keeps every
hit real.

## What the guard does not do

The guard checks that provenance is **stated**: every code block carries `verified:` or
`unverified:`, every number is followed by `measured:` or `predicted:`. It does not, and cannot,
check that the stated provenance is **true**. `verified:javap intellij.platform.diff.jar` passes
the guard whether or not `javap` was actually run — no script can confirm a verification was
performed, only that a claim of one was written down. The guard converts "silently unverified"
into "loudly unlabelled or falsely labelled"; only a human reviewing the plan's own claims can tell
labelled-and-true from labelled-and-false.

**It does not always scan a file to the end.** A fence-like run of backticks/tildes that the
guard's container-prefix grammar cannot resolve — behind a prefix shape it does not recognise, a
bare line indented 4+ columns with no container syntax of its own, or a line whose own container
prefix starts 4+ columns in (where CommonMark reads the whole line as an indented code block and
the marker on it as literal text) — makes the guard refuse to
guess whether that line opens or closes a fence (exit code 4). That refusal stops scanning **that
one file** at the line it could not classify: no line after it in that file is scanned on that run.
The scan is not abandoned — every other change directory keeps being scanned, and a real violation
found anywhere still outranks this exit code — but a genuinely unattributed claim sitting after the
unresolved line will not be reported until the unresolved line is fixed and the guard re-run. This
is disclosed here because it is a property of what the guard *does*, not merely of its exit codes:
treat exit 4 as "fix this line, then run it again," not as "this file has no other problems."

**It refuses to open a `tasks.md` it cannot trust the path of (exit code 3).** Before reading any
change's `tasks.md`, the guard confirms the file really sits at the plain, expected
`openspec/changes/<name>/tasks.md` path: not a symlink, not reached through a change directory that
resolves outside `openspec/changes/`, and not some other non-regular file (a FIFO, a device node)
sitting where a plan should be. Any of those shapes is a containment refusal, not an ordinary
violation — a PR-controlled `tasks.md` that is actually a symlink to something else could otherwise
make the guard read (or hang reading) content its author never intended to be scanned as this
change's plan.

A containment refusal stops the scan only for **that one candidate** `tasks.md` — every other
change directory is still scanned, and every violation, classification abort, or unreadable file
found anywhere is still reported — but exit code 3 **outranks every other exit code this guard can
produce** (1, 2, and 4 included), even when one of those was also found elsewhere in the same run.
A symlink escape must never be downgraded to a mere violation or environment code just because an
unrelated tag was also missing somewhere else; that is exactly the shape that would let a real
escape hide behind an unrelated, easily-fixed nit and get lost in the noise.

**What an operator should do on exit 3:** treat it as a security finding first, not a provenance
nit. Read the reported path(s) — the message names exactly which `tasks.md` failed containment and
why (symlink, directory escape, or non-regular file) — and fix the containment problem itself
(replace the symlink with a real file, correct the directory structure) before addressing any other
violation or abort also listed in the same run's output. The other findings remain true and still
need fixing, but they are not why the process exited non-zero this time.

## The implementer's duty

Because the guard cannot verify truth, the obligation falls on whoever writes the tag: write
`verified:<how>` only after actually performing `<how>`, and write `measured:<command> @ <ref>`
only after actually running `<command>` at `<ref>` and reading the result. Writing either tag
without doing the check it names is worse than leaving the block `unverified` or the number
`predicted` — it tells the next reader a check happened when it did not, which is exactly the
failure this contract exists to prevent.
