# The visual fix procedure

Loaded by `skills/flow/fix.md` for a visual fix — one whose defect or fix instructions name a
mockup departure, a capture or a `ui paths` file. Every "you" addresses the parent, as in
`skills/flow/fix.md`, whose steps these sections bind.

## Measuring first

Binds step 1 of `skills/flow/fix.md`.

**Fix instructions that dispute a visual judgement this session already made — a spacing, size
or alignment an earlier round eyeballed as fine — open with the measurement, never with another
look.** Run `flow-guard measure-visual-properties` on the disputed region of the current capture
and the mockup (**Measuring by hand** below) and put the numbers in the
implementer's prompt as a `Measured:` line; a spacing dispute is measured on every side the
complaint names. The complaint's own wording names the property — "too big", "oversized" is a size
(`box` and `ink`); "cramped", "uneven", "too close" is a spacing (`gap`); "not filled to the
border", "flush", "reaches" is an edge alignment (`runs` through the container) — so the
measurement answers the property named, never the screen area the complaint sits in.

**Every fix task for a visual mismatch starts with the measurement, disputed or not.** The
first step of the task is the `Measured:` line above for the reported element on both sides; when
the numbers show no difference, the task's result is `not a departure: <the measured numbers, both
sides>` and no code changes, and in every case the measured numbers are in the task's result, so
a later round that contradicts it must quote them.

## The visual fix implementer's brief

Binds the fix implementer beside **The fix implementer's brief** (`skills/flow/fix.md`).

**A defect whose root cause is a shared source — a style, a token, a component other screens draw
too, the data every spec renders such as a UI-test seed or fixture — is fixed at that source, whatever else the fix moves.** A shared-source fix is never offered as "a separate change"
(**Fewest operator actions**, `skills/flow-contracts/pipeline.md`). A shared source's shipped look departing
from one feature's frames is not such a defect (**A departure from the mockup**,
`skills/flow-contracts/operator-prompts-auto-resolution.md`). The fix never patches the
changed screen alone to spare the others, and its steps carry two more. Every baseline the fix
moves is recaptured: every spec `flow-guard affected-specs <worktree> <round base> <capture root>`
prints for the fix's own diff where the capture root holds a `<spec>.sources` sidecar, else every
spec `specs` prints for it — none when it prints
nothing — plus every capture spec the fix's task names in its `**Tests:**` line — when it fixes a shared
component, every spec whose frames show that component, named by the planner or the finding that
raised the fix, never left for the diff to reveal — runs through one `capture` run with `--update-snapshots=all`
appended (step 8, `skills/flow/visual-verify-verifier.md`), and every PNG whose bytes changed is
read first, per **The sweep** (`skills/flow-contracts/known-bugs.md`), and
committed with the fix, in the checkout step 12 of **Steps 3–13** (`skills/flow/visual-verify.md`)
commits the stage's PNGs to. The fix's run takes the tests a
sidecar maps a changed file to and the tests its fix adds, a bare
spec path its every test, never
every spec a shared file reaches, which step 3's `visual-verify` re-run captures once. And every step-3 `visual-verify` re-run of the fix names under `moved:` in its prompt every frame a `<spec>.mockups` sidecar maps to a
spec whose baseline moved — the verifier compares each against its mockup exactly as it does the
change's own frames, so a screen the shared fix regressed is a departure the next fix fixes.

**A suite that also fails for a reason the fix did not cause is re-baselined per capture, never per
flag run** — `--update-snapshots=all` would write the unrelated failures' captures over their
baselines too. The suite runs without the flag; for each failing capture, every changed-pixel
blob of its diff image is classified against the fix's expected footprint — where the shared
source draws and how large its change is — and only a capture whose every blob fits has its
`-actual.png` copied over its baseline, each copied baseline read before it is staged. A test
stops at its first failing capture, so the run repeats until no failure inside the footprint
remains, and a capture with a blob outside it is a defect per **The sweep**
(`skills/flow-contracts/known-bugs.md`), never copied.

**A fix round from the second on carries the full departure list.** Its prompt carries every
departure of the report that opened the loop plus every earlier round's report, never the latest
report alone, and the fixer compares every frame row and departure of that list before it reports
done.

**A fixer checks every consumer of what it changed before it reports done.** It lists every view
and frame that draws a component, copy or style its diff changed — the listed departures' siblings
included — reads each one's recaptured PNG against its frame, and fixes in the same round any it
no longer matches. A departure whose frame contradicts older spec text is fixed to
the frame and the spec text updated in the same round, never handed back as an open question
.

**Mockup departures are classified before the fix is written.** Where every frame that
draws the departing component agrees, the departure is a defect this round fixes.

**A departure that persists two rounds is re-judged before a third fixer opens.** Before the
round's fix is written, each departure that stood in both previous rounds is checked against the
frame's sidecar comment and the recorded decisions (`design.md`'s `## Decisions`), with its
measurement beside the declaration: one a declaration covers is closed as `expected` with the
declaration cited, one the evidence does not bear out (a mis-measurement) is dropped, and neither
opens a fixer round.

**A size the frames, or the agents reading them, disagree on is picked once, from a value a frame
draws, and never reopened.** Where frames disagree on one element's size, position, spacing or
radius — or two verifiers, fixers or rounds read it off the frames at different values — the
parent picks one value a frame draws: the element's own design source when one exists (the frame
that draws it as its subject, or the design-system frame it is taken from), else the value most
frames draw, else the first of those frames in the handoff's order. The pick is written once, before
the fix, as a `design.md` decision whose `**ID:**` is `drift-<element slug>` and whose
`**Chosen:**` names the element, the property, the kept value and its frame, and every other frame
with the value it draws; and as that drift in the store, a withdrawn finding of the parent's:

```bash
flow record finding -change <name> -ref <the ref `flow record next-ref -change <name>` prints> -slot parent -severity minor -location '<element> <property>' -status 'withdrawn drift: kept <value> from <frame>' -note '<every frame and the value it draws>'
```

From then on every verifier attributes the rows that pick explains `decision drift-<element slug>`
(step 10's closed list, `skills/flow/visual-verify-verifier.md`), so they count as no departure; no
round fixes the element toward another frame's value, picks it again or asks about it, and the
handoff's `**Decisions:**` lines name each drift. Whatever looks wrong in use the operator reports
after landing.

**A shared element that departs on some screens and not on others is fixed where those screens
differ, never by a per-screen offset.** When the frames agree on where an element sits and the
captures put it in different places on different screens, the cause is what those screens draw
differently around it — a title block, an inset, a hero — and the fix fixes that. A constant
added for one screen or one tab to land the element where that frame puts it moves the element off
the frames the other screens matched, and the next round reports them.

**Every mockup departure the fix lands a permanent probe in the fix's one commit** —
a test that drives the real composable or component, never a hand-copied constant, and asserts
the value the verifier measured against the frame — so a later change that moves it again fails
a test instead of waiting for the next visual verification. The fix's tests name one probe per
departure it lists, and the implementer hands back a listed departure with no probe it watched
fail before the fix and pass after it as unfixed, never as fixed.

A departure seen in motion or over time — a frame range, a gesture, a delay — is probed
under the input the verifier recorded, its pace, distance and step timing, and sampled inside
the window the verifier saw it in, never with instantaneous steps or a settle wait that outlasts
that window; where the harness cannot reproduce those conditions, the commit body names the
difference.

**The implementer measures before it hands back.** When the defects include a mockup departure,
its dispatch names the absolute path of `skills/flow/visual-verify-verifier.md`, the capture spec
with its `<spec>.mockups` sidecar, the resolved `mockups` root and `mockup frame` value, and the
project's `capture` command, and tells it: on every view its fix touched, recapture against the
stack rebuilt from its fix, as step 3's confirming re-run rebuilds it, and run step 9's
`flow-guard visual-verify-pass measure` on each under the key `fix-<k>-implementer`, reading its
manifest's flagged bands, seams and `over` rows — fixing until they come back clean or each
leftover is named with its cause, and report every view's measured numbers. It measures each pair
on the 2× capture and the 2× mockup frame themselves, never on a `flow-guard
compose-mockup-frames` composite. A leftover of any size — a 1dp offset included — is closed or
named with its cause, never reported clean because it is small.

**A fixer iterates on the dev loop.** Where `## visual verification` declares a `dev loop`
command (**Project configuration — visual verification**,
`skills/flow-contracts/project-configuration-visual.md`), the fixer starts it once and looks at each
edit on what it serves, then stops it and runs one build and capture against the stack rebuilt from
its fix, as below — that capture is the only evidence the verifier reads. The dev server is never
taken to render the same pixels as that build: the closing capture decides, and a difference between
the two is a departure the round fixes, never a pass.

**A fix for a visual departure runs its capture spec before it closes.** Its verify step runs the
project's `capture` command against the stack rebuilt from its fix, `<spec>` the tests of the
findings it fixes — `<spec>:<line>` of each test whose screenshot a `<spec>.mockups` sidecar maps
to a finding's frame, Playwright's own location filter — plus, on a shared-source fix, the
recapture list above, and the fix closes only on that run's exit 0 — a red capture is fixed in the
same commit. Every other frame of the change's list is captured once, by step 3's `visual-verify`
re-run. `flow-guard recapture` needs a `tasks.md` task a fix has none of, so it is not used.

**A departure observed on a running clock closes on repeated running-clock passes.** When the
departure was observed on a running clock — a live walk, a recorded frame sequence, any read the
capture spec's pinned clock (`skills/flow/visual-verify-verifier.md`, step 8) does not
reproduce — the fix first adds a capture test that runs under the observer's conditions (clock
running, same viewport and data) and fails without the fix, and closes only once that test
passes five times over in one run with one worker (Playwright
`--repeat-each 5 --workers 1`), beside the capture spec's own exit 0.

**Every recapture the fix runs rebuilds the full app suite's zip after it** —
`flow-guard rebuild-full-app-suite-zip <worktree> <screenshots root>` in the checkout the recapture
wrote to, the rebuilt `full-app-suite.zip` committed with the fix's PNGs — so the zip never
trails the baselines it archives.

## Measuring by hand

These rules bind every `flow-guard measure-visual-properties` call a fix makes; the verifier makes
none (step 10, `skills/flow/visual-verify-verifier.md`).
**Never judge a size, alignment, spacing, corner radius, border, fill, shadow, icon size or
font size by eye from a resized or cropped image; measure it with
`flow-guard measure-visual-properties`, then eyeball what it measured.** A crop is for reading text
and layout, never edges or centres — interpolation and a small viewport shift an edge by
pixels and hide a gap outright. Every per-control comparison runs the tool on a crop around that one
control and reads the JSON's `delta` block; its options, properties, output and exit codes are the
header of `<agents repo>/stats/internal/guard/measurevisualproperties.go`, canonical. **Both halves are mandatory and neither substitutes for the other.** The
script's numbers are the only admissible measurement — no ad-hoc PIL, no reading a coordinate
off a crop; and its output is then eyeballed against the two crops before any number is
reported: state, per control, that the `box` the script found is that control (its edges land
where the control's edges are seen), that `content` is the icon or label and not a corner
artefact, and that a `null` or an exit 1 was resolved by a wider or better-centred region, not
by dropping the property. A number no eye confirmed is a methodology error waiting to ship.
Exit 2 with no calibration is the script refusing
precondition 1 below; supply it, never work around it. Four preconditions on any number
compared across two images:
1. **Calibrate before comparing images of different provenance** — a mockup export against a
   capture, or captures at two viewport sizes. Where the project declares a `mockup frame`
   geometry, its `scale` is the calibration the composite already crops by, and the frame the
   composite cropped is at the capture's own scale, so `--scale 1`; ad-hoc measurements use
   the same factor. Otherwise derive the factor from an element whose correctness in both
   images is already established by other evidence — a prior finding, a code-level guarantee,
   a passing test — as that element's box in each image via `--ref-a`/`--ref-b`, never from
   the nearest similar-looking thing: an unchecked ruler is itself a claim, and a wrong one
   makes the comparison wrong twice.
2. **Confirm both images are the same state of the view** — the same expand/collapse state,
   scroll position and populated/empty condition, not merely the same screen. The box between
   two landmarks encloses different content in different states, and its size then compares
   nothing.
3. **Treat an implausible result — a multiple rather than a percentage, an order of
   magnitude — as a methodology error, never as a bigger finding.** Re-derive the calibration
   and the state check before reporting it; a smaller wrong number from the same mistake reads
   as a finding and ships.
4. **Measure the rendered box, never the declared one — a declared minimum is a floor, not a
   size, and the hit box and the visible ink are two measurements, not one.** `sizeIn(min =
   44.dp)` bounds a box from below and nothing bounds it from above: a child's own intrinsic
   size inflates the box past the mockup's drawn size and past the floor itself, and reading
   the constant in code sees none of it. An accessibility minimum applies to the tap target
   only, while the mockup draws the visible shape, so a control that looks oversized is
   measured twice — hit box and ink — before either number is called wrong.

**A first impression that a control "looks broken" is a hypothesis, not a finding — zoom and
contrast-check the capture before it becomes a code change.** A small, low-contrast but
pixel-correct element reads as wrong at a glance, and "hard to see" is a different question
from "drawn wrong" — one for the operator, not for a fix.
Symmetric with the measurement rule above: the eye that passes a control owes a number, and so
does the eye that fails one.

**"Looks compact", "looks tight", "matches the proportions" is the same hypothesis said of a
gap — a spacing claim is backed by the script's `gap` property on every side it speaks for,
never by a glance.** The eye cannot compare two absolute gaps in a scaled screenshot — a
4x asymmetry between the space above a control and the space below it reads as "about the
same" at 1x — so a claim about the space between two elements (a control and the dividers
around it, a heading and its rule, a row and its neighbours) names the measured background
pixels on each side, from one capture, and the mockup's on the same sides, before it is called
symmetric, tight or matching. Gaps compound where sizes do not: an outer container's content
padding stacks on a header's own inset on one side only, and no per-element size measurement
sees it. An operator's re-raised or repeated spacing complaint is a measurement order, not a
second look — the second look is what already failed.

**"Filled to the border", "flush with the edge", "reaches the corner" is a third measurement,
and neither of the two above sees it — a claim that one shape's paint ends where another
shape's boundary is compares two edge coordinates, never one shape's size (hit box or ink)
and never the empty space between two shapes (`gap`).** A fill inset from its own container's
border has no gap the `gap` property can find — the inset is the container's own interior
colour — and a correct box size says nothing about where the box sits. Crop the *container*
as the region, centred on the filled row, and read `--props runs` (the tool header's `runs`
entry), never `content.padding`. In Compose the usual cause is modifier order alone — a `padding()` placed
before the `background()` or `clip()` it was meant to inset the content of, or on the parent
before its children's own backgrounds, shrinks the painted area, not the content — a
checkable line of code before any screenshot.

## Narrowed visual re-runs

**A failing visual `verify` row is re-run narrowed, never whole.** When the `## visual verification`
`verify` row a `test` group ends with fails, its one inline re-run is `flow-guard run-verify-list
-only <visual-failing.txt> <worktree> visual [<canonical worktree>]`, the `visual-failing.txt` file
the failing run prints, run once with one retry, and one fixer gets every test that run fails and
every test it prints as `flaky` in one batch. A flaky test is a failure the fixer fixes at its
source (`rules/fix-determinism-at-the-source.mdc`), and never a reason to run the selection again.
Each fix round then re-runs that file the same way. Once a narrowed run is green and prints no
`flaky` line, one plain run of the `test`
group under the round's `-fix-<k>` key is the gate, and when it fails the next round narrows to
its failures. That re-run selects from the round's diff, never the change's, plus the spec file of
each `visual-failing.txt` entry, and `-only` selects as the row it follows did — `flow-guard
run-verify-list` reads the round base from the previous verify row itself (THE ROUND'S SELECTION,
its header canonical). `-only` records no dispatch row, writes no green cache and never stands as the gate.
A plain run is the `specs` row's selection — the
change's tests plus the smoke tail —
and never the whole suite. A selection naming more than 40 spec files runs the specs the change edits plus that tail, and is never refused (A WIDE SELECTION, `flow-guard run-verify-list`'s header canonical). Each failing run prints its failures in two groups, the
change's specs and the specs outside the `specs` row's output. A non-visual command keeps its own inline re-run.

**Two consecutive verify rounds whose failures all fall outside the change's specs are a round
without progress too** (**Fewest operator actions**, `skills/flow-contracts/pipeline.md`). When two narrowed visual re-runs in a row fail only in specs outside the
`specs` row's output, the next round takes one
batched fixer on the `no-progress-fixer` role, carrying every earlier round's
report (`skills/flow/fix.md`). When the diff touches shared rendering code every spec is
in that output, so only the same-defect rule of **Fewest operator actions** applies.
