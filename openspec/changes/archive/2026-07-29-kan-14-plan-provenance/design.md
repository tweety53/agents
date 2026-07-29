## Context

Full design: `docs/superpowers/specs/2026-07-29-kan-14-plan-provenance-design.md`, committed
`d38372a`. This file carries what an implementer needs at hand plus the decision record.

The motivating evidence is the KAN-6 run in `/Users/tweety53/Projects/intellij-review-queue`. Six
plan claims were wrong; the costly one was structural rather than slow. The plan said the diff
marker could be read from `DiffRequest`. It cannot — `ChangeDiffRequestChain` never copies user data
onto the requests it produces, so `request.getUserData(...)` is permanently `null`. Had the
implementer transcribed the snippet, the context-menu guard would have been permanently false and
the feature a no-op **that every unit test in that task still passed.** It was caught only because
that one task carried a hand-written instruction to check.

## Goals / Non-Goals

**Goals.** Make an unchecked claim in a plan visible at the proposal gate and refusable at the
implementation gate. Enforce mechanically the part a script can actually enforce.

**Non-Goals.** Making plans correct — nothing here would have made the KAN-6 plan correct. Verifying
that a `verified:` tag is truthful. Parallelising independent task groups (KAN-15). Changing what
`writing-plans` requires of a plan's structure beyond adding the tags.

## Approach

### The four tags

On a fenced block's info string: `verified:<how>` or `unverified:<what-to-check>`.
On a numeric claim, an HTML comment on the following line: `measured:<command> @ <ref>` or
`predicted:<what-confirms-it>`.

The asymmetry between the two pairs is deliberate. A code block always gets a tag — in-repo code
satisfies it with `verified:compiled in-tree`, so the cost lands on external claims, which is where
every KAN-6 error was. A number, by contrast, **may not appear at all** without provenance. That is
the stronger rule because the weaker one would not have helped: "194 tests" was not mislabelled, it
was invented, and no labelling discipline catches a number nobody tried to source.

### The guard

`scripts/check-plan-provenance.sh`, in this repository's established style: a header comment stating
the rule *and the failure that motivated it*, `file:line` output, and a `CHECK_PLAN_PROVENANCE_ROOT`
environment override so the harness can point it at a fixture tree rather than `cd`-ing into one —
`check-references.sh` sets that pattern exactly.

**Scope: `openspec/changes/*/tasks.md`, excluding `archive/`.** This is the single most important
implementation constraint. `check-references.sh`'s own header records what happened when a guard in
this repository over-fired: 28 false failures on its own tree, after which suppression markers were
the only way to silence them, and those markers then switched off the real checks sharing those
lines. A guard that scans skill docs, READMEs or archived plans earns that outcome. Archived plans
predate the rule and are history.

The guard checks that provenance is **stated**. It cannot check that `javap` was run.

### The harness

`scripts/test-check-plan-provenance.sh`, following `test-check-references.sh`: sandboxed fixture
trees under `TMPDIR`, assertions on exit status *and* output, never touching the real tree. Both
directions covered — tagged passes, untagged fails, archive is skipped, non-plan files are skipped.
A guard whose passing direction is untested will be "fixed" into silence the first time it
misfires.

### The skill edits

`myflow-start` step D gains the tagging duty and a guard run before publishing. `myflow-do`'s
implementer dispatch gains a fourth standing clause beside the existing three (no-commits, TDD,
engineering principles):

> **PLAN PROVENANCE:** a fenced block tagged `unverified:` is a hypothesis, not code to transcribe.
> Establish the real API before writing against it, and report what you found. A block tagged
> `verified:<how>` was checked as stated; if it does not compile, report that — do not contort the
> code to match it.

Plus a contracts-table row in `rules/myflow-manual-review.mdc`.

### Two stale facts, corrected because accuracy forces it

`agents-repo-verification`'s guard requirement names `/myflow-review` (deleted by KAN-8) and
`scripts/test-state-advance.sh` (not on disk). This change has to restate that requirement to add a
guard, and cannot restate it accurately while preserving either error. The replacement defines the
guard set by **what `.myflow/project.md` declares** rather than by a count in the title — a count is
what let "five guards" survive a script's removal unnoticed.

## Decisions

### How much verification the rule demands

**ID:** provenance-tag-not-hard-gate
**Status:** active
**Chosen:** A stated-provenance tag enforced by a guard — verification stays a judgment call, and
the script enforces only that the claim is labelled, which is what a script can check.
**Considered:** A hard gate compiling every external snippet and running every measurement before
publishing — catches all six KAN-6 errors, but needs a build in a stage that deliberately has no
worktree, and slows every proposal round whether or not it touches an external API. Banning
external-API implementation bodies from plans entirely — removes the error class, but overrides
`writing-plans`' explicit code-block mandate and makes plans markedly less concrete.

### What an `unverified:` tag obliges downstream

**ID:** unverified-is-a-hypothesis
**Status:** active
**Chosen:** The implementer establishes the fact before writing and reports what it found — the tag
changes behaviour at implementation time, the only place it can save anything.
**Considered:** Also writing the confirmed fact back into `tasks.md` so the plan self-heals — would
have stopped KAN-6's Tasks 3, 4 and 5 each rediscovering the same class of SDK mismatch, but puts
implementers in the business of editing the plan mid-run, a larger change to plan ownership than
this warrants. Leaving the tag purely informational — cheapest, and least likely to change what
actually happened.

### Scope of this change

**ID:** provenance-without-parallel-lanes
**Status:** active
**Chosen:** Plan accuracy only; serialization filed as KAN-15 and left in To Do.
**Considered:** Both together — they share no code, and the parallel-lane work needs per-lane build
isolation, the riskier half, which would hold the cheap win hostage. Serialization only — saves ~20
minutes of wall clock but does nothing about the errors, which cost more and nearly shipped a
defect.

## Risks / Trade-offs

**The guard over-fires and gets suppressed.** The one failure mode this repository has already
lived through. Mitigated by the narrow file scope and by the harness asserting the passing direction
as hard as the failing one. If the guard cannot stay quiet on a correct plan, the guard is wrong.

**Qualified after the pass-8 fix wave (Important 5):** that sentence was written before the
fail-loud decision (list-context-fail-loud, above) and, taken literally, the change now
deliberately violates it. A BARE fence-like run at 4+ columns of indentation, with no container
syntax of its own, is *valid CommonMark as a list-item continuation* when a list is genuinely in
scope — and this guard refuses it (exit 4) rather than silently deciding either way, because the
operator judged a silent misclassification worse than a loud refusal on an ambiguous shape. The
qualifier that keeps the original sentence honest rather than deleting it: this costs nothing
*today*, not nothing — verified against this repository's own plans: zero BARE fence-like runs
indented 4+ columns, out of 76 total fence lines across both real `tasks.md` files (52 in kan-8,
24 in this change).
<!-- measured: grep -nE '^ {4,}(```|~~~)' openspec/changes/{kan-8-myflow-updates,kan-14-plan-provenance}/tasks.md,
     zero hits, out of 52+24 fence lines total @ branch openspec/kan-14-plan-provenance -->
A future plan that legitimately needs a deeply nested BARE continuation will hit this refusal
and must dedent or add an explicit marker, not silently pass.

**A tag becomes ritual.** `verified:` can be typed without verifying and nothing detects it. Stated
honestly: this raises the cost of an unchecked claim from zero to writing a false evidence string,
and puts the claim in front of the human at the proposal gate and the reviewer at the implementation
gate. It does not make it impossible.

**`.myflow/project.md` is this repository's own config**, so an error in `## lint` changes how every
future run verifies this repo. It is edited alongside the harness that proves the script runs.

## Migration Plan

None. Archived plans are out of scope and are not rewritten.

One active change other than this one sits in `openspec/changes/`: `kan-8-myflow-updates`, complete
at 83/83 tasks but not yet archived.
<!-- measured: ls openspec/changes/ + openspec instructions apply --change kan-8-myflow-updates @ branch openspec/kan-14-plan-provenance -->
Its `tasks.md` contains **52 fence lines**, which the finished guard resolves to 25 untagged blocks.
<!-- measured: grep -cE '^[[:space:]]*(```|~~~)' openspec/changes/kan-8-myflow-updates/tasks.md @ branch openspec/kan-14-plan-provenance -->
<!-- measured: ./scripts/check-plan-provenance.sh 2>&1 | grep -c 'kan-8-myflow-updates/tasks.md' @ branch openspec/kan-14-plan-provenance -->

**An earlier draft of this paragraph claimed zero**, citing `grep -c '^```'` — anchored to column 0
and therefore blind to the indented fences that make up all 52 of them. The tag was truthful about
the command it ran and false about what it implied. The error is described here rather than quietly
overwritten, because it is precisely the defect this change exists to name, committed by this
change's own design document.

Its plan is left untagged and the guard is not narrowed to spare it; the hits clear when kan-8
archives, since the guard excludes `openspec/changes/archive/`.

## Open Questions

None.

## Post-review reversal — the `openspec/specs` widening

Added during implementation, reviewed, then **reverted** after pass 2 of the review panel.

The idea: a dead capability spec had gone unnoticed because `check-vocabulary.sh` never scanned
`openspec/specs`. Widening it made the drift visible immediately.

The defect, found by building fixtures from a pending change's delta bodies and running the real
guard: **19 hits**,
<!-- measured: ./scripts/check-vocabulary.sh $(find openspec/changes/kan-8-myflow-updates/specs -iname spec.md) @ branch openspec/kan-14-plan-provenance -->
all of them live requirements naming a retired term *in order to
prohibit it*. `SHALL NOT contain \`gates\`, \`tested\`, ...` cannot be written without naming
those fields. Rewording deletes the requirement; a suppression marker is forbidden by `CLAUDE.md`.

**The general lesson, which is the part worth keeping:** a vocabulary ban is safe over prose that
*describes* a system and unsafe over text that *legislates* about it. Skills and rules describe;
specs legislate. That distinction, not the directory name, is the axis a future drift check has to
be built on.

### Whether the guard should scan live specs at all

**ID:** vocab-scan-live-specs
**Status:** superseded by vocab-scan-rejected
**Chosen:** Widen `DEFAULT_TARGETS` to `openspec/specs` only — live specs describe what the system
is, so retired vocabulary there is drift.
**Considered:** Scanning `openspec` wholesale, rejected because archived plans legitimately quote
retired vocabulary as history.

### Whether the guard should scan live specs at all (superseding)

**ID:** vocab-scan-rejected
**Status:** active
**Chosen:** Do not scan `openspec/specs`. A spec must be able to name what it forbids, so the scan
flags correct requirements with no honest repair available.
**Considered:** Teaching the guard to ignore a retired term inside a prohibition — a real parser
with its own false-negative risk, and this run twice demonstrated that hand-rolled Markdown matching
in bash finds a new gap each round. Accepting a permanently red guard — rejected because a
declared-lint command that is always red trains operators to ignore it, which is the failure the
guard exists to prevent.

## Field evidence — what a `verified:` tag cannot promise

Recorded from the KAN-6 human gate, which ran while this change was in review. It is the sharpest
available bound on what provenance tagging buys, and it is evidence rather than speculation.

KAN-6's right-click menu **failed at the gate**: the menu did not appear at all. The plan's snippet
used `EditorEx.setContextMenuGroupId`, and that symbol was checked against the real platform jar —
it exists, the signature matched, `verifyPlugin` passed. A `verified:javap ...` tag on that block
would have been **entirely truthful**.

It was still the wrong mechanism. `setContextMenuGroupId` only feeds the *default* popup handler,
and the platform appends its own handler afterwards, which wins. The fix installs the menu with
`installPopupHandler` from a `DiffViewerListener.onInit()`, which runs after the platform's.

**The lesson, stated against this change's own interest:** `verified:` attests that a symbol exists
and its signature matches. It does not attest that the symbol is the right one for the job, that it
is called at the right time, or that the surrounding framework will not override its effect. A tag
that is true at the symbol level can still sit on code that cannot work.

That is not an argument against the tags — the KAN-6 defects they *would* have caught were real, and
one of them was a silent no-op. It is an argument for stating the boundary precisely, which the
proposal's *"What this does not claim"* section already does and which this note now backs with a
worked case rather than an assertion.

A second, smaller lesson from the same gate: the guide predicted that if the menu failed, the marker
key would be the cause. It was wrong, and the wrong prediction sent the first investigation in the
wrong direction. **A confident guess about a failure's cause, written into a test guide, is itself an
unsourced claim** — the same defect class this change exists to name, one level up from the code.

## Post-review reshape — the fence classifier leaves bash

Decided after **five panel passes and seven fix waves**, on the panel's own standing test.

### The evidence

Distinct defect classes found in one ~150-line bash routine: column-0-only detection; non-whitespace
prefixes invisible; task-checkbox lines swallowing later violations; `PROVENANCE_RE` with no left
boundary (a comment *denying* provenance satisfying the rule that grants it); prose over-firing
(twice — first letter-led, then container-prefixed); an abort reachable on content inside an open
fence; a quadratic DoS in blockquote stripping; container scoping ignored on the closing test; and a
second quadratic in line storage. **Pass 5 found three of these on a settled tree, after four of seven
slots had already returned "clean".** The canonical, numbered enumeration lives in
`check-plan-provenance.py`'s own module docstring — this paragraph is prose recounting it, not a
second source of truth for the count; an earlier draft of this paragraph independently omitted the
`PROVENANCE_RE` item and so disagreed with that docstring's list, which is exactly the drift a
second enumeration of the same facts invites.

Every round's fix was structurally better than the last. The inversion in wave 3 was a genuine design
improvement, and two principles reviewers independently judged the result "a grammar, not an
allowlist". It still produced three more Criticals.

### The decision

**ID:** fence-parser-in-bash
**Status:** superseded by fence-parser-reshape
**Chosen:** A hand-rolled fence classifier in Bash ERE, matching the repository's existing guard style.
**Considered:** Nothing else — the house style was Bash, and the first implementation looked small.

**ID:** fence-parser-reshape
**Status:** active
**Chosen:** Rewrite the guard with a **real block-structure parser in Python 3 (standard library only)**,
keeping its CLI contract byte-identical — same exit codes, same `CHECK_PLAN_PROVENANCE_ROOT`
override, same `file:line` output. A container stack that records what each fence was *opened*
inside fixes the closing-scope defect by construction; proper block parsing distinguishes a fence
opener from prose inside a list item, removing the false-abort class; and Python list indexing is
O(1), removing the second quadratic.
**Considered:** A CommonMark library — **unavailable**: no `cmark`, `cmark-gfm`, `pandoc`, or any
Python markdown package is installed, and this repository has no dependency management, so adding one
would trade a known bug class for an unguaranteed runtime. Continuing in Bash with better structure —
rejected on the evidence above: seven waves of exactly that. Dropping fence tagging and keeping only
the numeric rule — a real option (the numeric rule has had one bug in seven waves) but it abandons
half the change's purpose, and the operator chose to reshape instead.

### What this costs, stated plainly

The repository becomes **Bash + Python**, not Bash-only, and `.myflow/project.md`'s own description of
itself must change. `/usr/bin/python3` is present on every supported macOS, so this adds no install
step — but it is a genuine widening of the toolchain and should be recorded as one rather than
slipped in.

### What survives unchanged, and why it matters

**The assertion harness stays in Bash and is not rewritten.** It tests the guard's *CLI contract* —
exit codes, output shape, scan scope — not its internals. That makes it the regression suite proving
the reshape preserves behaviour: every defect the five passes found already has a fixture, and the
new implementation must satisfy all of them before anything else is believed. Its count has grown as
each review pass added fixtures for what it found — 97 assertions as of the pass-6 fix wave (a
historical snapshot, not a live claim: this count is deliberately NOT tagged `measured:`, because
that tag promises a reader who re-runs the named command gets the same number, and here they will
not — the harness has grown every fix wave since, and Minor 7 of the pass-8 fix wave is the record
of that choice). It is deliberately not restated as a bare number anywhere else in this repository,
for the same reason the defect-class count above is not: a number copied away from the thing it
counts is a number that goes stale the next time that thing changes.

## Post-pass-7 reversal — the ambient list-column tracker

Added during the pass-7 fix wave to resolve a shape `classify_line`'s grammar could not otherwise
answer: a BARE line (no container syntax of its own) reducing to a fence run at 4+ columns of
indentation, where whether an enclosing list puts that column in scope depends on state from
EARLIER lines the classifier does not otherwise keep. The pass-7 fix threaded a scalar,
`_next_list_content_col`, across lines to answer that question.

**The evidence.** That tracker produced three separate Criticals in one review pass, on a settled
tree: a thematic break (`- - -`) opened a phantom list scope it should not have; an uncapped
trailing-whitespace gap after a marker inflated the tracked column past what the marker actually
established; and a scalar (not a stack) lost all state on any nested-list dedent, misreading
correctly-dedented content as still inside the deepest list seen so far. All three are the same
root shape: state inferred for a line that carries no marker of its own, threaded across lines the
classifier is not actually looking at when it makes the decision.

### The decision

**ID:** list-context-ambient-tracker
**Status:** superseded by list-context-fail-loud
**Chosen:** Track the content column of the innermost open list across lines (`_next_list_content_col`),
consulted by `classify_line` whenever a BARE line's indentation is ambiguous, so the guard could
resolve the 4+-column BARE case instead of refusing it.
**Considered:** Nothing else at the time — the fail-loud alternative (refuse rather than guess) was
not yet on the table; resolving the ambiguity looked like the more complete answer.

**ID:** list-context-fail-loud
**Status:** active
**Chosen:** Delete the ambient tracker. A BARE line at 4+ columns of indentation with no container
syntax of its own now aborts (exit 4, content-classification) rather than being resolved by
cross-line state. `classify_line` carries no cross-line state of any kind after this reversal.
**Considered:** A stack-based repair of the same tracker (fixing the thematic-break, trailing-
whitespace, and dedent defects individually) — shown to the operator as a working option. Rejected:
a guard that cannot tell whether a list is in scope should refuse, not guess, and refusing is
available here in a way it is not for other cross-line ambiguities this parser already accepts (see
the "KNOWN LIMITATION" notes in `check-plan-provenance.py`'s module docstring). Measured against
this repository's own plans, the case being refused does not occur — see the "zero BARE fence-like
runs indented 4+ columns, out of 76 total fence lines" measurement earlier in this document (Risks
section, tagged `measured:`) rather than restating that count here untagged: this is "costs nothing
today," not "costs nothing," and is recorded as such in the Risks section below.

**What this is not, and why the distinction matters:** `FenceContext.list_content_col` (Critical 1,
pass-9 fix wave; renamed from `list_col` at pass-11 — see the decision below) is NOT a return of
this deleted tracker. `list_content_col` is per-fence state, captured from a list marker physically
present on that fence's own opening line, consulted only while that one fence is open, and
discarded when it closes — structurally identical to `bq_pre`/`bq_post`, which this design
document's own reshape section already established as correct. The deleted tracker inferred a
list's content column for lines carrying no marker of their own, threaded across blank lines and
dedents with no notion of when it should stop applying. `list_content_col` never does that: a line
with no list marker of its own never contributes to it, and a fence's `list_content_col` is scoped
strictly to that fence's own open/close pair.

## Post-pass-11 reversal (declined) — container-prefix support versus fail-loud

Not a reversal that happened — a reversal considered and **declined**, for the third time, at the
pass-11 fix wave. Recorded here because the alternative (extending `list-context-fail-loud`'s
refuse-rather-than-guess principle to container-prefixed fences generally, rather than continuing
to patch `FenceContext.list_content_col`/`strip_container_prefix`/`is_closing_line`) has been on the
table at every one of the last three fix waves and has been rejected every time, silently, by simply
fixing the layer in front of the reviewers instead of raising the question this section now raises
in writing.

### The running cost

Three consecutive fix waves — passes 9, 10 and 11 — each found and fixed a Critical in the exact
same code path: a column width computed for a fence opened behind container syntax (a blockquote,
a list marker, or both), carried into `FenceContext`, and consumed later by `is_closing_line` to
decide whether a later line closes that fence.

- Pass 9: the gap between a blockquote marker and a list marker (0-3 tolerated columns) was not
  recognised at all — the blockquote run absorbed at most one such column (`_BQ_RUN_RE`'s `\s?`,
  replaced at pass 13 by `_consume_bq_run`'s one-COLUMN step), so a two-space gap
  fell through to unresolved prose.
- Pass 10: once the gap was recognised, its width was computed and then discarded before reaching
  `list_content_col` — the field existed, but was short by exactly the gap's width.
- Pass 11 (this fix wave): the outermost leading indentation in front of the whole container
  prefix — stripped at the very top of `strip_container_prefix`, before `bq_pre` is even
  consumed — was measured and then discarded the same way, for the same reason: a width computed in
  one place, needed one function later, and nobody carried it across the boundary.

**The counting rule, stated before the number** (Important 3, pass-13 fix wave — three review slots
reached three different totals from the same facts, which is itself the proof that no rule had ever
been written down; a `measured:` tag over an undefined rule is not a measurement):

> A Critical counts toward this path if its **verified root-cause fix touches the shared container
> machinery** — `strip_container_prefix`, `classify_line`'s container branch, `is_closing_line`,
> `FenceContext`'s width fields, or `_advance_col`. A Critical found in the same fix wave whose fix
> lands elsewhere (an environment gate, a containment check, a message's wording) does not count,
> however severe it was.

**Re-derived at pass 14, because the previous itemisation was curated rather than rule-derived**
(Important 3 of the pass-14 fix wave). Applying the rule above to the whole record admits four
Criticals the old list silently omitted — every one of them satisfies it, and none had a stated
reason for exclusion. Two documents agreeing on the same curated list is not reproducibility; it
is the same list copied twice, which is the defect class this change exists to name. The rule is
now applied exhaustively, pass by pass, with the *excluded* Criticals named as well as the
included ones so a reader can check the filter rather than trust it:

| Source | Counts | Excluded, and why |
|---|---|---|
| Original Bash version (module docstring) | items **2** (non-whitespace prefixes invisible), **3** (checkbox lines swallowing later violations — `strip_container_prefix`'s checkbox step), **6** (over-firing on container-prefixed lines), **8** (quadratic in blockquote-prefix stripping — the blockquote run consumer), **9** (container scoping ignored on the closing test) = **5** | 1 (column-0 detection, not container syntax), 4 (`PROVENANCE_RE`), 5 (prose over-firing, bare lines), 7 (abort on content lines inside an open fence) |
| Pass 6 | Critical 1, blockquote depth tracked as a boolean rather than a count — `FenceContext`'s width fields and `is_closing_line` = **1** | C2 BOM, C3 `CLAIM_RE` ReDoS, C4 `OSError` |
| Pass 7 | Critical 2, leading indentation unbounded before container recognition — `classify_line`'s container branch = **1** | C1 unreadable scan entry (`main()`), C3 `CLAIM_RE`'s left boundary |
| Pass 8 | Critical 1, the ambient-tracker reversal — one heading bundling **3** distinct defects (a thematic break opening a phantom list scope; an uncapped trailing gap after a marker inflating the tracked column; a scalar losing state on any dedent) = **3** | C2, the abort exiting mid-scan (`check_file`/`main()`) |
| Pass 9 | Criticals 1 and 2 = **2** | C3, documentation |
| Pass 10 | Criticals 2 and 3 = **2** | C1, the environment gates |
| Pass 11 | Critical 1 = **1** | C2, containment |
| Pass 12 | Criticals 1 and 2 = **2** | — |
| Pass 13 | Critical 1 = **1** | — |
| Pass 14 | Criticals 1, 2 and 3 (the unbounded whitespace runs in the list/checkbox/blockquote markers; `_consume_bq_run` consuming a whole run where an exact count was demanded; the fence-indent budget discarded on the `bq_post` branch of `is_closing_line`) = **3** | C4, the wrapper's interpreter probe; C5, the `openspec/changes` containment anchor |

**5 + 1 + 1 + 3 + 2 + 2 + 1 + 2 + 1 + 3 = 21 through pass 14** (18 through pass 13).
<!-- measured: each pass's Criticals read off its own fix-wave brief's `## Critical` headings under
     .superpowers/sdd/, plus the module docstring's numbered item list, then filtered by the rule
     stated immediately above; both the included and excluded sets are itemised in the table so the
     sum is re-derivable from the table alone @ branch openspec/kan-14-plan-provenance -->

**Two attribution errors are corrected here rather than carried forward.** The old itemisation
called the tracker's three defects "the pass-7 ambient-tracker reversal's 3 Criticals". The ambient
tracker was *added* at pass 7 — pass 7's own three Criticals are unrelated findings — and its three
defects were *found* by the **pass-8** panel and bundled under one heading in `fix-wave-pass8.md`.
And the old totals ("11 through pass 11", "13 through pass 12", "14 through pass 13") were reached
by a filter narrower than the written rule, which is how three review slots previously reached
three different numbers from one set of facts. The rule has not changed; only its application is
now complete.

Ten of the twenty-one (passes 9 through 14: 2 + 2 + 1 + 2 + 1 + 2) turn on one quantity — the column
at which a container's content begins — either computed in one place and not carried to another,
carried in the wrong unit, or measured without a bound where it is used. That is why pass 12's fix
was structural (one running cursor, not another per-site patch), why pass 13's was narrow (that
cursor was verified correct across 9,702 prefix combinations; only the width fed into it was
measured in characters instead of columns), and why pass 14's is again structural: every whitespace
run in the parser is now consumed through one primitive that requires an explicit column budget, so
"a run measured somewhere and unbounded where it is used" — five recurrences across passes 7, 9, 10,
11 and 14 — is no longer expressible without a greppable sentinel saying so.

### The measurement that made the alternative tempting

Zero of the 76 fence lines across this repository's two real `tasks.md` files (52 in kan-8, 24 in
this change) carry any container syntax — a blockquote marker or a list marker — on the fence line
itself.
<!-- measured: grep -nE '^[[:space:]]*(>|[-*+]|[0-9]+[.)])[[:space:]]*(```|~~~)' openspec/changes/{kan-8-myflow-updates,kan-14-plan-provenance}/tasks.md, zero hits, out of 76 fence lines total @ branch openspec/kan-14-plan-provenance -->
Every one of the twenty-one defects above was found and fixed on a code path this repository's own
plans have never yet exercised. That is precisely the shape of evidence that makes "fail loud
instead" attractive: `list-context-fail-loud` (above) already set the precedent of refusing a
shape rather than maintaining cross-line/cross-function machinery to resolve it, on materially
weaker justification (a 4+-column BARE line with no list in scope, a single ambiguity, not a whole
grammar).

### The decision

**ID:** container-prefix-support-kept
**Status:** active, reconsidered three times
**Chosen:** Keep full container-prefix support (blockquote depth, list markers, checkboxes, the
gap between them, and the compound content-column arithmetic all of that produces) rather than
narrowing the guard to fail loud (exit 4) on any fence opened behind container syntax at all. Made
three times: at pass 9 (fix the gap, do not refuse it), at pass 10 (fix the gap's width reaching
`list_content_col`, do not refuse the shape), and at pass 11 (fix the outermost indentation reaching
`list_content_col`, do not refuse the shape) — each time after a new Critical in this exact path,
and each time the operator chose to patch the next layer rather than cut the feature.
**Considered:** Refuse (exit 4) any fence whose opening line carries a blockquote marker, a list
marker, or both, and support only bare, uncontained fences — collapsing `FenceContext.bq_pre`,
`bq_post` and `list_content_col` to dead fields, deleting `strip_container_prefix`'s gap-tolerance
and checkbox handling, and removing the corresponding harness sections outright. This would have
prevented all five of the pass-9/10/11 Criticals from being reachable at all — five, not three
(Important 4, pass-13 fix wave: this sentence restated the disproven one-Critical-per-pass
itemisation two paragraphs from the corrected count, from the same error) — at the cost the
measurement above quantifies: zero real regressions today, for a construct real Markdown plans are
free to use.
**Reasoning for keeping it, stated plainly rather than as a foregone conclusion:** a plan is
free-form prose with fenced code, and container-nested fences (a numbered task with an embedded
snippet, a blockquoted aside quoting a command) are ordinary, valid CommonMark that a plan author
has no reason to expect this guard to reject. Refusing them outright would be `list-context-
fail-loud`'s "refuse rather than guess" principle applied to a case that is not actually
ambiguous — CommonMark defines exactly what these constructs mean — so the refusal would not be
buying safety, only convenience for this guard's own implementation. That reasoning has held for
three passes running; **it is recorded here, rather than left implicit, so a future reader with a
twenty-second defect in hand can weigh it against this exact cost-and-benefit accounting instead of
re-deriving both from scratch** — and can conclude, if the count keeps climbing, that this decision
was wrong.


## The implementer model policy

Added to this change after implementation, at the operator's direction, when auditing this change's
own records showed that the model each implementer subagent ran on had never been recorded — so no
one could say whether any policy had been followed.

**ID:** implementers-at-the-ceiling
**Status:** active
**Chosen:** Implementer subagents dispatched by `/myflow-do` run on Opus — or the harness's
strongest available model — named explicitly at dispatch, never inherited. Review-panel slots are
untouched by this and stay on Sonnet, which `myflow-review-panel-economics` owns and this change
does not restate. Every dispatch records its model in the SDD ledger; where the dispatcher cannot
observe the model (a slot dispatched by `subagent_type`, which carries its own agent definition),
the entry records `unknown (agent-defined)` rather than a plausible value.
**Considered:** superpowers:subagent-driven-development's economy tiering — "the least powerful
model that can handle each role", with the cheapest tier where the plan already contains the code
to write. **Rejected**, and the rejection is the whole point of writing this down: that guidance
optimises the cost of a single dispatch, and this pipeline does not pay for implementation defects
at dispatch time. It pays for them at the review panel, which *finds* them rather than preventing
them, and then in a fix wave that re-runs every slot. A cheaper implementer buys a more expensive
review. Two further instructions in the same upstream skill are overridden for related reasons and
are named in `pipeline.md` §Model policy rather than left to be discovered: dispatch the final
review on the most capable model (myflow fixes the panel at Sonnet and escalates panel *breadth*
instead), and escalate the model in fix rounds 4-5 (unsatisfiable when implementers already sit at
the ceiling from round 1).
**Also considered:** moving this policy into its own change, since it arrived mid-review and
produced four Criticals of its own within an hour. The operator chose to keep it here.
**Known limit, stated rather than left to be found:** the ledger that records the models lives
under `.superpowers/`, which is gitignored, in a worktree `/myflow-finish` removes — so the record
is session-scoped and does not support an after-the-fact audit. The spec says so plainly rather
than implying a durability the artifacts do not have. Making it durable is a change to
`/myflow-finish`'s archive step and belongs to whichever change owns it.
