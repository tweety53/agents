# Visual verification — the verifier's steps

Read by the verifier **Steps 3–13** (`skills/flow/visual-verify.md`) dispatches, at the absolute
path its prompt carries — the parent never loads it whole. It carries step 4, steps 8–11, the
motion step and the `## Verdict` and `## Report` templates, run as written.

**A command the harness denies has no exit code.** Its report line reads `blocked — environment:
<the denial text, verbatim>` where `exit <n>` would stand, and the step's status is `blocked` —
never an exit code the command did not return.

4. **Run `setup`, if declared.** A non-zero exit blocks, printing the command verbatim.
8. **Capture** — when `specs` is declared, run it first from the `regression checkout`, with
   `<frontend-root>` substituted for the worktree and `<merge-base>` for the merge base the bundle
   names; a non-zero exit blocks, printing the command and its exit. The verify `specs` maps for that
   merge base already ran green in `flow.verify`, which the report's `verify:` line names as the
   bundle does — this step runs no `verify`. The capture spec is the touched views' existing spec from `specs`' output,
   extended, when one covers the view; a new spec is authored only for a view none covers. Run
   `capture` with `<spec>` substituted for the spec's path. **Every capture asserts the view is on
   screen before it is taken** — gymie-playwright's `expectRoom` — so a missed navigation fails the
   capture instead of becoming its baseline. **A capture reached through a click also asserts the state it clicked into, and parks the pointer before it is taken** — a positive control on a drawn, state-dependent property the wrong state cannot satisfy (the selected tab's fill, the opened sheet's tag), and the pointer moved clear of every control (`page.mouse.move(0, 0)`) after the last click — so a missed click fails the capture instead of becoming its baseline, and no hover ink reaches one. **A capture reaches every control it clicks or types into by a structural locator, never a page coordinate** (**D. Basic Workflow #3 — Writing plans**, `skills/flow/brainstorm-planner.md`). **A capture whose subject sits under a scrim, or any dimming or alpha layer, also pins that subject structurally** — the presence of each tag the change draws beneath the layer and the absence of each it removes, covered tags included — because the dimming can shrink a real difference below the screenshot comparison's per-pixel threshold; a dimmed capture whose baseline predates the change is recaptured, never trusted because its diff passes. `screenshots`'s root-not-leaf shape is canonical in
   `skills/flow-contracts/project-configuration-visual.md`; nothing here restates it. **Every screenshot
   this spec takes is the full page or viewport, never a clipped region.** A clip is the right tool
   for an implementer's own targeted assertion (a fixed piece of text, an icon), but this stage's
   own job — page-wide styling (background, shadow, border, font, spacing) matching the mockup — is
   exactly what a clip is built to hide; step 10 below cannot compose a clip against a full mockup
   frame and call the result a fidelity check.
   The one exception is a region an approved prototype draws alone (step 10): its element capture is
   the frame-bound capture, taken beside the view's own full-page capture, never in place of it
   (group-allowed-pairs-roles self-review).
   **`capture` creates this change's baseline**: writing
   a PNG that does not yet exist is its success path, not a failure — `verify` is the regression gate
   over an already-committed baseline, `capture` is not, and only a `capture` failure for some other
   reason blocks (see Blocking in `skills/flow/visual-verify.md`). **A Playwright `capture` runs
   with `--update-snapshots=all` appended, last on its command line**: the bare
   `--update-snapshots` is Playwright's `changed` preset, which keeps every PNG whose diff is under
   the comparison threshold, so a re-capture after a subtle colour fix comes back byte-identical
   and the baseline goes on showing the defect; `=all` rewrites every PNG the spec takes, writes a
   missing one with exit 0, and as the last flag wins over a bare one in the declared command.
   A `toHaveScreenshot` passing over a baseline this
   change wrote is the app agreeing with itself, never with the mockup — step 10 is the only
   comparison, and `capture: exit 0` is never evidence of a frame's fidelity. **Seed the spec with data the frame does not
   draw.** A mockup is drawn on a happy case, and a spec whose fixture reproduces it verifies only that case:
   the fixture holds, for every element the view derives from data, at least one input the frame's
   own numbers would never produce — a threshold, goal or marker value outside the plotted range;
   a dataset whose derived numbers (axis ticks, averages, deltas, unit conversions) do not come out
   round; a list longer than the viewport; and the empty or first-time entry path beside the
   populated one.

   **Pin the clock and the viewport before the first navigation.** The spec installs a fixed
   clock — `page.clock.install({ time })` at one chosen instant — before its first `page.goto`,
   never after: `install` is an init script, and an init script reaches only navigations that
   start after it. Every date the fixture writes — a workout's day, a weigh-in, a target's
   `from` — derives from that same instant in the app's own timezone, never from `new Date()`
   and never from its UTC day where the app renders local. Where the checkout already carries a pinned
   instant and a helper (gymie-playwright's `PINNED_TODAY` and `pinClock`), the spec uses them
   rather than choosing a second instant. The spec states its own viewport — the Playwright
   project it belongs to, or `test.use({ viewport })` in the file — so a coordinate and a capture
   are the same size on every machine. **The one exception is a view whose content the server
   derives from its own clock** — a window, an age, a "today" computed server-side — which no
   browser clock can pin: such a spec leaves the clock real, its dates relative to the real day,
   and its header says so and why (gymie-playwright's body-calories spec).
   **A server-side pin narrows that exception.** Where the checkout can send the server its own
   today (gymie-playwright's `pinClock` sends `X-Gymie-Today`), the spec pins both ends to the one
   instant, and leaves real or masks only what neither pin reaches (KAN-802).

   **Every mockup frame this change adds or updates gets a capture and a sidecar line — never
   optional, with `mockups` declared.** The frames are the change's declared list (step 10) plus
   every `<frame id>.png` that
   `git -C <worktree> diff --name-only <merge-base>..HEAD -- <mockups>` prints. For each one the
   capture spec takes a screenshot of the view that frame draws — added where the spec has none,
   updated where the frame changed what it draws — and the `<spec>.mockups` sidecar carries its
   `<screenshot name> <frame id>` line, written by this stage where the change brought none. A
   frame on that list with no capture or no sidecar line blocks.

   **Then update the full app suite — and create it where the checkout has none.** The suite and
   its file names are canonical in **Visual verification**
   (`skills/flow-contracts/project-configuration-visual.md`). Absent → author `full-app-suite.spec.ts`
   beside the capture spec, one full-page capture per screen the app has — enumerated from the
   app's own routes or navigation and from every spec already in the checkout, never from this
   change's diff. Present → add a capture for every screen this change added and update the
   capture of every screen it changed or removed. Run `capture` with `<spec>` substituted for the
   suite's path and `-g '<titles>'` appended, naming only the tests this step added or updated —
   never the whole suite; a non-zero exit blocks as any `capture` failure does, a failing test this
   change did not introduce included (**the sweep**, `skills/flow-contracts/known-bugs.md`).
   Then rebuild the zip from exactly what the suite just produced with
   `rebuild-full-app-suite-zip.sh <worktree> <screenshots root>`, the one script every recapture
   that moves the suite runs too (`skills/flow/verify-fix-loop.md`, `skills/flow/sync-onto-base.md`).
   Both arguments are absolute — the script refuses a relative root — and a project whose `capture`
   row chains this rebuild after Playwright has run it with `capture`, a second run rewriting nothing
   (group-allowed-pairs-roles self-review).

   `<screenshots root>` is `screenshots` resolved the way `resolve-visual-screenshots.sh`
   resolves it. Every PNG's mtime is pinned before zipping because the zip stores it: a rebuild over
   unchanged images is then byte-identical and commits nothing. A non-zero exit blocks.

   Then run `check-spec-reach.sh <worktree>` — the spec
   `capture` just wrote must be reached by a `package.json` script of the `regression checkout`;
   exit 1 (an orphan, named) or 2 (cannot answer) blocks.
9. **Measure in one call, then read in one turn.** Right after step 8's capture — after step 10's
   prototype render and any sidecar this stage writes, which come first — run, with the arguments
   the bundle names:

   ```bash
   visual-verify-pass.sh measure <worktree> <name> <key> <spec path> [--prototype]
   ```

   `--prototype` when step 10's approved prototype is the frame. Its header is canonical for what it
   runs — `resolve-visual-screenshots.sh`, `check-identical-baselines.sh`, `compose-mockup-frames.sh`
   and `measure-visual-properties.sh`'s band, seam and matrix reads on every pair — the files it
   writes and its exits; it carries on past a step's non-zero exit and records each exit and its
   stderr in the manifest it names, `<abs-worktree>/.superpowers/sdd/visual-manifest-<key>.json`. A
   sidecar written later, or a recapture, runs it again. Then read the manifest and every image it
   names — each capture, composite and cropped frame — in one turn of parallel Read calls.
   The resolver is what selects this run's fresh output from the committed baseline PNGs already
   sitting under the same `screenshots` root, rather than joining `screenshots` with a guessed
   filename: its exit 1 (zero matches) in the manifest blocks, indistinguishable from a view that
   never rendered, and its exit 2 (cannot answer) blocks the same way. **Read every capture** — no
   script can do that — and state, per view, what was seen.
   Each captured PNG gets its own `- <view>: <absolute PNG path>` line, and each view no frame is
   composed for its `per view, no frame composed: <view> sweeps:` line — `check-verify-report.sh`
   refuses a report missing either (group-allowed-pairs-roles self-review: two round-1 verifiers
   named only the prototype-framed captures, and a clipped chart label on an unnamed full-page
   capture surfaced a round later).
   An unreadable PNG is reported and blocks
   too.

   The manifest's `identical` groups are `check-identical-baselines.sh`'s, whose header is canonical
   for what it compares: each names PNGs of one directory that are
   byte-identical, and the report carries one `- identical:` line per group stating why those views
   draw the same pixels, or the defect that made them, which blocks: kan-998's pad-up and pad-down
   baselines were one image, and only a reviewer saw it. Its exit 2 → it cannot answer, and blocks.
10. **Compose captured frames against their mockups, if `mockups` is declared.** With a
   `<spec>.mockups` sidecar beside the capture spec, step 9's `measure` composed every pair it maps
   through `compose-mockup-frames.sh`, with this project's resolved `mockup frame` value as its fourth
   argument — a sidecar's own first `frame` line overrides it — and its manifest lists one
   `<composite path> diff=<ratio>` per pair: **read every one** and state,
   per pair, what the difference panel's white regions are, the ratio, and where the capture
   departs from the frame. The ratio is information, never a threshold: a pair blocks on a
   departure you can name, not on a number. A compose exit 1 (a broken map — a screenshot no capture
   matched, a frame file absent, a malformed line, a capture whose size differs from the cropped
   frame) or 2 (cannot answer — a malformed `mockup frame` value included) blocks.

   **An approved prototype is the frame where no mockup is** (kan-1038). When no sidecar maps a
   capture — `mockups` undeclared included — and `design.md`'s `## Context` cites an approved
   prototype HTML file that exists in the worktree (**The checklist**,
   `skills/flow/brainstorm-planner.md`), render that file as each capture test's frame first. With
   the project's own installed Playwright, a Node script you write under
   `<changeRoot>/visual-verification/prototype/` opens `file://<prototype>` at the capture test's
   own viewport, device scale factor and colour scheme, drives the state that test captures —
   every hover, focus or pressed state the prototype draws included — and writes
   `<changeRoot>/visual-verification/prototype/<screenshot stem>.png`.
   For a region capture (step 8's exception) the script sizes the prototype's region to the
   capture's width, injecting CSS that overrides any standalone width cap such as a `max-width`,
   and `visual-verification.md` records the injected rule under the frame's entry
   (group-allowed-pairs-roles self-review: a 1100px prototype panel beside a 1232px capture left no
   band, seam or matrix reading possible).
   Then write
   `<changeRoot>/visual-verification/prototype.mockups`, its first line `frame none` and one
   `<screenshot name> <screenshot stem>` line per capture, and run step 9's `measure` with
   `--prototype`, which composes with that map in place of `<spec>.mockups` and
   `<changeRoot>/visual-verification/prototype` in place of `<abs-worktree>/<mockups>`.
   Under that `frame none` line the compose step pairs each capture with its frame at native size,
   prints each pair's `diff=` ratio as n/a and writes no `<stem>.frame.png` — by design, never a
   defect — so `measure` reads bands, seams and the matrix on the prototype PNG and the capture
   directly.
   Every frame-bound reading below then runs on those pairs as on a
   mockup's. A state the prototype draws that no capture test drives is a missing capture, and
   blocks.

   > **The `mockups` declaration gates the frame comparison, never the sweeps.** The compose
   > step, the band pairing, the seam pairing, the ink inventory (sweep 10), the bounded-row
   > sweep (sweep 9) and the element × property matrix read a frame against the capture, and a
   > project declaring no `mockups` root, with no approved prototype to render (above), composes
   > nothing and reports each of them
   > n/a — no frame. The state sweeps run on the captures alone and are the default on every
   > `flow.visual-verify` run, composed or not. Where a sweep's wording names the frame for something
   > the capture answers without one — sweep 7's last element reached by name, sweep 5's
   > display precision — the run reads that name against the view's own rendered content;
   > sweep 6's order comparison needs the frame's list and is reported n/a — no frame
   > beside the other frame-bound readings.

   **The difference panel cannot show a structural departure, so the first read of every pair
   is the script's band pairing, never the panel.** The composite's panels are the capture left, the frame middle and the difference right, unlabelled; a radius, a stroke width, a glyph size or a colour is read from `<stem>.frame.png` and the capture PNG themselves, never off the composite (kan-1016: a stepper radius read off the composite was raised in two rounds and withdrawn). The panel is white wherever any channel
   differs, and on a real pair 20–50% of it is white from seed data, font rasterisation and the
   frame's own annotations alone. A 1px rule the capture omits, a row padding it lacks, a surface fill
   swapped for the page colour, a button border in the wrong colour: each moves the ratio by
   under a hundredth and vanishes in that white. Beside every composite the compose step writes
   the cropped frame as `<composite stem>.frame.png`, the capture's own size, and for every pair
   `measure` ran `measure-visual-properties.sh <stem>.frame.png <capture PNG> --scale 1 --props
   bands,seams`: the manifest lists each pair's flagged bands and seams, and its full output sits in
   the file the manifest names. Read `delta.bands` (the script's header is canonical for the property): every band the
   frame draws, paired with the capture's in order, a text or data difference never unpairing
   one. **The length and colour tolerances are `mvpOverLength` and `mvpOverColour`
   (`<agents repo>/stats/internal/guard/measurevisualproperties.go`)** — `measure` flags by them, and nothing here
   restates their values. **Every `missing` band, every `extra` band, and every pair whose `since_pair`,
   `gap_above` or `height` delta exceeds the length tolerance or whose `edge` or `colour` distance exceeds the colour tolerance is
   a departure until a cause is named that a band cannot carry** — a wrapped label (a `height`
   delta on a text band, confirmed by transcription), a frame annotation outside the app's own
   ink; an empty-versus-populated state is precondition 2 below, a spec to fix, never a cause to
   accept. Data never creates or removes a band, moves a divider or recolours a border, so "the
   seed data differs" explains none of them. The pairing resyncs after an unpaired band; a pair
   count far below the frame's band count is itself the finding that the layout differs
   wholesale, and the composite is then read to say how.

   **A band is one row of the page's structure and sees nothing across it, so the same call's
   seam pairing is the second read: `delta.seams`, the vertical structure inside every boxed
   band.** A segmented control with its two cell dividers gone and its wrapped labels
   left-anchored where the frame centres them is one band in both images, the same height and
   the same border colour, and pairs clean. For every boxed
   band — a bordered or filled control, a card, a hairline; never a bare text row — the script
   lists its seams (an outer border's side, a cell divider, a filled cell) paired in order, and
   for every cell between two paired seams its text lines with their `offset` from the cell's
   centre and their `left` and `right` insets. **Every `missing` seam, every `extra` seam, every
   seam pair whose `since_pair` or `width` delta exceeds the length tolerance or whose `colour` distance exceeds
   the colour tolerance, every cell whose line `count` differs, and every line whose `offset` delta exceeds the length tolerance
   is a departure until a cause is named that a seam or a line cannot carry.** Text width never
   moves a seam, so a data difference explains no seam; a line's width is data, so an `offset`
   delta whose `left` delta is within the length tolerance (or whose `right` is) is a left- (right-) anchored
   label whose text differs — the cause is then the transcription, and it is named — where an
   `offset` delta with `left` and `right` both moved by the same amount is a label anchored to a
   different edge than the frame's, a departure whatever the text reads. What the seams cannot
   see — a divider in a row with no border or fill, the label of a filled cell, a glyph, a
   corner radius — the per-control sweeps below still measure.

   **Every departure is attributed at the merge base before the round closes — `departure` is
   never a closing state.** A band, seam, matrix or sweep finding no named cause explains is
   measured twice more before the report is written: the same element on the frame at 2x, and the
   same element on the view's own rendering at the merge base — the committed baseline PNG the
   merge base carries for that spec, read out with `git show <merge-base>:<baseline png path>`
   into a temporary file and measured like any other image. A recorded decision is a named cause only for the deltas it names, by element and property: any other delta on the same band or element is unexplained until measured (kan-1016: a band's `colour` delta was put down to decision `warmup-sublines-w4b` for four rounds, and was the checkbox's size). The comparison decides which of the
   three the departure is, every time: **present at the merge base** — pre-existing, blocking and
   fixed in this run per **the sweep** (`skills/flow-contracts/known-bugs.md`); **absent at the merge base** — this change's own defect, blocking and taking
   **The loop** (`skills/flow/verify-fix-loop.md`) per **Steps 3–13**
   (`skills/flow/visual-verify.md`); **corrected or gone at 2x** — a mis-measurement, the
   corrected number standing in its place. A
   view this change creates has no rendering at the merge base: every departure on it is this
   change's. **A view the merge base already draws but carries no baseline PNG for is captured at
   the merge base before any departure on it is attributed — never counted as this change's for
   want of a baseline** (KAN-942: KAN-758's verifier counted some 39 over-tolerance band pairs as
   its change's, every one geometry the merge base already shipped or a frame artefact). One
   capture per spec serves every departure on its views. In each repository the stack and the
   capture run from, add a detached worktree `<name>-merge-base` beside the change's own, at that
   repository's merge base (`git worktree add --detach ../<name>-merge-base <merge-base>`); copy
   this change's capture spec, its `<spec>.mockups` sidecar and its fixtures into the merge-base
   worktree of the repository that holds specs; start the stack there from `start` when declared, else `## run` — on its own
   derived workspace, never the change's running stack — and run `capture` from there with
   `<spec>` substituted. Measure the same element on the PNG it writes and judge it as above. Whether
   or not the capture ran, stop that stack, remove the workspace it derived with the project's declared removal command
   (**Creation and cleanup**, `skills/flow-contracts/workspace-isolation.md`),
   and remove those worktrees before the report is written: nothing the merge-base capture writes
   is committed or stands as this change's baseline. A merge-base capture that cannot run — the
   spec reaches a route or control the merge base lacks, the stack does not start — leaves the
   departure this change's, its line naming `merge-base capture failed — <reason>` beside it.
   A line attributing a departure as present at the merge base names its evidence beside it —
   `merge-base baseline: <merge-base sha>:<baseline png path>` or `merge-base capture: <absolute
   png path>` — and `check-verify-report.sh` refuses a `pre-existing` or `present at the merge
   base` attribution that names neither (kan-1015: the kan-466 weight-tab button was attributed
   pre-existing from the code path alone, and its baseline re-captured over the defect).
   Every line that reports the departure names its cause or its attribution
   beside it — never a bare `departure`.
   **A departure's expected value cites the mockup's source wherever the handoff ships one, never a
   rule inferred from the pixels**: where the frame was drawn from HTML or CSS in the tree — beside
   the mockups or cited by `design.md` — the line names the declaration the expected layout, size or
   colour comes from (`grid-template-columns:1fr 1.3fr 1fr; height:32px`), and a value no
   declaration states is named as read from the pixels (kan-1003: an inferred `flex: 1 1 auto`
   cost four fix rounds against the grid the HTML declared).
   A departure within 1dp is reported like any other, its measured value beside it: whether it
   stands is the parent's call under the Blocking paragraph of **Steps 3–13**
   (`skills/flow/visual-verify.md`), never yours (KAN-1000).
   **A part the frame's sidecar comment declares expected until a named sibling ticket lands is
   no departure.** The line reporting the frame names that part
   `expected — <sibling ticket>, <decision ID>` and nothing blocks on it, while every part the
   comment does not name is measured and attributed as above — the comment being the one
   `skills/flow/brainstorm-planner.md` plans for a frame a sibling ticket finishes (KAN-821).

   **A full-page match — a clean composite read, a structural match — is necessary, never
   sufficient: verify at the control level, measure rather than eyeball, and exercise the states
   a resting frame does not show.** Full-page comparison catches wrong regions and wrong overall
   layout, and wrong text only where the capture's data is the frame's own; it does not catch a
   border style, an icon's glyph, or a colour step, all of
   which are invisible at full-page scale. Before
   accepting any field, button, icon, or toggle as matching its mockup:
   1. **Crop and zoom (2–3x) both the mockup region and the corresponding capture, side by side.**
      Do this for every interactive control the view carries, not only ones that already look
      suspicious at full scale. **Name the control's kind in each image first, before any
      measurement** — a plain text field or a headline display figure; a list row carrying a
      leading radio, or a quiet row whose only indicator is a trailing filled circle on the
      selected one; a segmented control of N cells each fitting its own label — a differing
      kind is a departure before any number is taken.
   2. **Exercise every interactive state the mockup draws for that control, not only its resting
      one** — a focused field, a field with a value typed (so a clear icon or a validation state
      actually renders), a pressed button, an enabled toggle. A control checked only at rest has
      not been checked at all if the mockup draws it focused or active; capture that state
      separately rather than inferring it from the resting frame.

   **Enumerate every mockup frame the change touches into an explicit checklist before starting
   the sweep, and account for each one by name at handoff — done, or why not.** "Did a
   screen-by-screen sweep" is unfalsifiable without a fixed list: a partial pass reads exactly like
   a complete one if nothing forces naming what was skipped. Build the list from
   every frame id `design.md`/`tasks.md` cites for this change, or from the mockups directory
   itself when no such citation exists, plus every frame the prompt names under `moved:` — one a
   shared fix's recapture moved (**The loop**, `skills/flow/verify-fix-loop.md`); a frame reached only by inference from another (already
   covered by the composite diff, "shipped" and pre-existing, blocked by a real environment limit)
   still gets named, with the reason. The list also carries every frame the `<spec>.mockups` sidecar of each capture spec this change edits maps, whether or not its baseline has moved yet, and two frame ids are never read as one (kan-1016: the warm-up sheet's frame `templates-A4b`, mapped by the edited `templates.spec.ts`'s sidecar, joined the list only in round 7, once its baseline moved). **The list is the change's, never the round's**: a fix run
   re-verifies the same list, not the views its fix touched, and the specs `capture` runs are
   chosen to reach every frame on it, never the other way round — a denominator the verifier
   derives from the specs it happened to run is the partial pass wearing a fraction. The report's
   `frames:` line counts against that list and names every declared frame absent from the
   per-frame lines, with its reason.

   **Exercise the states below and capture each one; a state nobody drove into is a state nobody
   verified.** Each is a state a resting capture and a full-page composite cannot show, so each is
   reached deliberately rather than by whatever the walkthrough happens to pass through.
   1. **Every appearance of a named state, and of a shared component, diffed against its other
      appearances.** Crop and diff each appearance of a named state (selected, active, pressed)
      and of a shared component (a button variant, a sibling card) across every place it shows in
      the same flow: two captures can each match their own frame and still disagree with each
      other, because the mockup draws a state once and only the captures can show the drift. Where
      a screen renders the same logical control from different code branches depending on state —
      empty versus populated, first versus subsequent — capture it in every reachable branch and
      diff the branches against each other, never only the branch the walkthrough reached first.
   2. **The whole range of every ranged control, and the empty state of every dynamic list or
      picker.** A wheel, slider, drag handle or multi-step selector is exercised in every
      direction and past where it wraps or clamps, with a screenshot along the way — the resting
      capture and one direction prove nothing about the other. A list or picker is captured with
      zero items as well as populated: an empty state sits outside any normal walkthrough.
   3. **Every option of every picker or selector, tested against submit.** Enumerate its options
      and ask whether any is predictable to fail on submit; one that is gets reported as a defect
      however good the rejection reads. No mockup draws the invalid-selection case, so no
      composite can see it — submit the options the rules already forbid and name every one the
      list should have excluded instead of offered.
   4. **Every intermediate screen of a scripted navigation.** When exercising a flow by scripted
      navigation (not manual clicks), screenshot after every single action and confirm the
      resulting screen against an expected marker — a heading, a test tag, a distinctive label —
      before issuing the next action. Never chain two or more blind actions and inspect only the
      final screenshot: a coordinate that assumed a fixed layout silently steers the whole
      remaining sequence onto the wrong screen the moment real content shifts it — an extra
      suggestion card, a longer note, a wrapped title — and the resulting screenshot can still
      look plausible enough to accept at a glance.
   5. **Every text run in the capture, transcribed and judged as display text.** The difference
      panel is white wherever the capture's data differs from the frame's, and "the data differs"
      is the explanation that absorbs a wrong number: a white label region is not accounted for
      until its text has been read. Transcribe every number, unit and label the capture shows and
      state, per run, that it is a display value at the precision the frame shows — a raw float,
      a `NaN`, `null`, `undefined`, an empty string where the frame draws a value, a placeholder
      is a defect whatever the frame's own numbers are.
   6. **The elements of capture and frame, listed top to bottom, and every element inside its
      own container.** Write both lists — heading, chart, chips, list, link — and compare their
      order: a control present in both but in a different position relative to its siblings is a
      departure the per-control crop never sees, since each crop matches its own control wherever
      it sits. Then
      confirm no element's ink crosses its container's bounds into a sibling — a plotted line, a
      marker, a label — driving the out-of-range value step 8 seeded and stating whether the app
      clamps it or hides it.
   7. **Every scrollable region scrolled to its end, and the frame's last element reached by
      name.** A frame taller than the capture viewport is itself the assertion that the screen
      scrolls: scroll to the frame's bottom element, capture it, and name it in the report. A
      screen whose content extends past the viewport with nothing to scroll it is a defect, and
      the fixed-viewport capture cannot show it — it proves the visible viewport and nothing
      beyond.
   8. **Every derived value re-derived after its input changes.** For each value the view
      computes from an input the operator can edit — a delta, a total, a conversion, a preview —
      type a new input, capture before and after, recompute the expected value by hand, and
      state both. A value that does not follow its input, or that is absent on one entry path the
      populated path shows (first-time versus returning), is a defect no single capture can
      show.
   9. **Every row the frame bounds with a divider, hairline or container edge — a header bar, a
      sticky bar, a toolbar, a list section, a dialog's action row — gap-measured on the side
      facing each bound, in both images, before the row is called matching.** Which rows are
      bounded is read from the frame, never from the capture: a capture that omits the frame's
      hairline shows no bounded row to measure, and the band pairing above is what finds the
      omission. The top and bottom gaps are the row band's `gap_above` and its successor's in
      the band pairing above, already paired per image; the left and right gaps are read from
      the `gap` block of the
      per-control measurement below for the row's controls: the script scans past the region to
      the next neighbour, so the number is already in
      the JSON of every measurement this step makes and costs nothing beyond reading it. This
      sweep runs on every such row, unprompted: the spacing rule below answers a claim or a
      complaint, and a row nobody claimed anything about was never measured. A row
      whose every control measures the right size is the case to suspect, not to pass: a uniform
      container padding around correctly-sized controls is invisible to every reading but this one.
   10. **Every non-text ink the frame draws — hairline, divider, rule, border, background fill,
      shadow band — inventoried from the frame alone, then found in the capture one by one.**
      The horizontal inks — every hairline, rule, fill band and border edge the frame draws across
      the page — are the frame's bands, and the band pairing above has already listed each one
      `missing` or paired with its colour; the vertical inks inside every bordered or filled
      band — a control's sides, its cell dividers, a card's inner rules — are its seams, and the
      seam pairing above has listed each one the same way; this sweep reads both lists and adds
      the vertical inks neither can see, the dividers of a row with no border or fill: run
      `measure-visual-properties.sh` with `--props runs` on the
      frame alone, the region a full-width row through the centre of each such row, write down
      every run whose colour is neither the
      background's nor a text run's — its position and length — then the same region on the
      capture, and pair the lists: a run on the frame's list with no counterpart on the
      capture's is a departure, whatever the row's text reads. This sweep runs on a frame whose
      content already matches and on a frame
      already fixed for something else — a fix run re-verifies the whole frame, never the
      element it fixed.

   **Every element the frame draws, on every property the script measures — a matrix per
   frame, never a list of sweeps done.** Every sweep above is anchored on one kind of element
   and asks the question its own incident taught: a control's kind and box, a text run's
   content, a bounded row's gaps, a column's ink. An element outside every anchor — a headline
   figure, a caption, a summary row, a slider's track and thumb, a row's own border — is
   transcribed at most, and a property no sweep names is read by nobody: the text sweep judges
   what a run says and never its colour or its size, the row sweep measures the gaps around a
   row and never how far the row stretches, and the per-control crop is asked for the one number
   the last incident made memorable. The report's `sweeps:` line then names the sweeps, not the
   numbers, so a round that re-checked one defect reads exactly like a round that measured the
   frame. Before a frame is called matching:

   1. **Enumerate the elements from the frame, not the capture and not the last report** —
      every visually distinct thing it draws, top to bottom: each text run (a heading, a figure,
      a label, a caption, a unit), each control (a field, a button, a segment, a slider's track
      and its thumb separately, an icon), each row, card or section container, each rule, divider
      and fill band. An element in the capture the frame does not draw is itself a row, marked
      `extra`. A fix run enumerates afresh from the frame — the previous round's departures are
      cells in the new matrix, never its rows. Seed the enumeration with the matrix step 9's
      `measure` wrote to `<changeRoot>/visual-verification/matrix/<frame id>.json` —
      `measure-visual-properties.sh --props matrix`, which finds each image's app surface — a sheet's
      top edge on a scrim frame — lists every ink run inside it paired in order, and fills every
      column below for every paired element, each region clipped to its image; every element it
      merges or misses is added by hand (kan-1015: the band read took each scrim-dimmed frame as one
      band, and every matrix was measured with a hand-written script), and a row it does not pair
      takes the per-element calls. That file is the one copy, overwritten every pass — never a second copy under another name (kan-1016: each frame's matrix was committed three times under three names).
   2. **Fill every column for every row from the script's own output**, one call per element:

      | column | text run (`--props ink`, the crop spanning the run's container's full width) | control, row, card (`--props box,radius,border,fill,shadow,content,gap`) | rule, divider, fill band (`--props runs`, the region crossing it) |
      |---|---|---|---|
      | kind | the run's role (heading, figure, caption…) | the control's kind, per sweep 1 | rule / fill band |
      | content | the transcription, per sweep 5 | `content` size and `content.colour` | — |
      | colour | `ink.colour` | `fill.colour`, `border.colour`, `content.colour` | the run's `colour` |
      | size | `ink.height` — the cap-height, the font-size stand-in — and `ink.width` | `box.width` × `box.height` | the run's `length` — a rule's thickness |
      | position | `ink.left`, `ink.top` in the container crop | `gap` on all four sides; `box.left`/`box.top` | the run's `from` |
      | stretch | `ink.left` and the container width minus `ink.left + ink.width` — the insets from each container edge | `gap.left` and `gap.right` to the image edge, or to the neighbour, and `box.width` against the container's | the run's extent on the crossing scan line |
      | alignment | the two insets above, compared: equal is centred; for a line inside a boxed band's cell, its `offset`, `left` and `right` from the seam pairing | `content.padding`, all four sides | — |
      | border | — | `border` per side and its colour | — |
      | fill | — | `fill.colour` and `share` | the run's `colour` |
      | radius | — | `radius` per corner | — |
      | shadow | — | `shadow` per side | — |

      The matrix table's header opens with `element` and names the columns above, and each row's
      `kind` cell opens with the row's class — `text run`; `control`, `row` or `card`; `rule`,
      `divider` or `fill band` — then its role, so `check-verify-report.sh` can read which columns
      may be n/a for it.
      Every cell holds the frame's number, the capture's and the `delta` — or n/a — `<why>`
      where the column does not apply to the row's class (the dashes above). **A cell the pixels
      cannot answer is measured from the mockup's source, never left n/a**: where the script
      exits 1 on a region no widening resolved, or reads a side `clipped` (the element runs off
      the image there), the frame's number is the value the mockup's source declares for that
      element — the CSS of the HTML the frame was drawn from (`box-shadow`, `border-radius`,
      `border`, a colour), found beside the mockups or cited by `design.md` — and the capture's
      is the script's where its own image measures, else the value the app's source draws the
      element with; the cell names its source. Only a frame with no source in the tree reads
      n/a — `no mockup source`, with the region tried. A cell left empty, or a cell filled by
      eye, is the sweep-name report again. The matrix rides `visual-verification.md` (step 11)
      under the frame's entry, and the report's `matrix:` line counts its rows and its n/a
      cells.
   3. **Explain every row the matrix's `over` list names.** The manifest carries each composed
      frame's `over` rows by their 1-based index — every matrix row `missing`, `extra` or
      `unresolved`, or paired over the length or colour tolerance. The report's `- <frame id>
      over-tolerance: <n> — #<row>: <cause>; …` line names every one, `<n>` the frame's count, each
      with one cause from a closed list: `departure`, attributed per **Every departure is attributed
      at the merge base** above; `annotation` — a frame annotation outside the app's own ink;
      `transcription` or `text width` — a text difference the transcription names; `decision <ID>` —
      a recorded decision whose `**ID:**` `design.md` carries, for the deltas it names;
      `present at the merge base`, with its merge-base evidence beside it; `mockup source` — an
      `unresolved` row's value measured from the mockup's source per the cell rule above.
      `check-verify-report.sh` refuses a missing line, a wrong count, an unnamed row or a cause off
      the list, and reads `over` from this dispatch's manifest, so a pass that skipped `measure`
      never stands on last round's numbers.

   **Never judge a size, alignment, spacing, corner radius, border, fill, shadow, icon size or
   font size by eye from a resized or cropped image; measure it with
   `measure-visual-properties.sh`, then eyeball what it measured.** A crop is for reading text
   and layout, never edges or centres — interpolation and a small viewport shift an edge by
   pixels and hide a gap outright. For every per-control comparison this step makes, run

   ```bash
   measure-visual-properties.sh <mockup frame PNG> <capture PNG> --region-a x,y,w,h --region-b x,y,w,h --scale <n>
   ```

   with each region a crop around that one control and a background margin on every side —
   `--props runs` on that crop first where its edges are in doubt, since every run boundary it
   lists is an edge the region can sit on, and a miscrop is the glance again; and `--edge`
   lowered below the distance `runs` reports between a control's fill and its background where
   that distance is under the default 24, since a surface-on-page card has no hard edge at the default and exits 1 on every box property
   until it is — and
   read the JSON's `delta` block: `abs` and `pct` per numeric property, RGB distance per colour
   (the script's own header is canonical for its options, properties, output
   and exit codes). Two readings the eye reliably gets wrong: an icon's tint is `content.colour`,
   never `fill` — the fill is the box behind the glyph, and a grey glyph on the right fill matches
   on every other property; and a pill
   is a `radius` equal to half the box height, while a rounded rectangle is any smaller number —
   read the number, since both look "rounded" at 1x. **Both halves are mandatory and neither substitutes for the other.** The
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
   as the region, centred on the filled row, and read `--props runs`: the run right after the
   border's run is the fill when flush, and the container-coloured run between them when not,
   its `length` the inset in px — the mockup's frame gives the same list to compare against.
   Not `content.padding`: it boxes every non-fill pixel in the container, the other rows' text
   included, and a fill covering half the box flips which colour counts as `fill`. The same
   reading covers a highlight against the row it highlights and an icon against the circle
   drawn behind it. In Compose the usual cause is modifier order alone — a `padding()` placed
   before the `background()` or `clip()` it was meant to inset the content of, or on the parent
   before its children's own backgrounds, shrinks the painted area, not the content — a
   checkable line of code before any screenshot.

   **No sidecar is never a silent skip.** Before reporting `mockups: no map`, list
   `<abs-worktree>/<mockups>` and check it for a frame plausibly matching any view this change
   touched — by filename, by a frame id `design.md` or `tasks.md` cites for this change, or by the
   screen family the touched `ui paths` name. A plausible match exists → **author the
   `<spec>.mockups` sidecar yourself**, one `<screenshot name> <frame id>` line per captured view
   with a real frame, then run step 9's `measure` again — composing against a sidecar this stage
   just wrote is not a special case, and it is committed at step 12 along with everything else this
   stage writes. No plausible match anywhere in the directory → report `mockups: no map for
   <spec's basename> — searched <mockups dir>, no frame for <views>`, naming what was searched, and
   continue. Not declared → report `mockups: not declared` and continue. The sidecar's shape is
   canonical in **visual verification** (`skills/flow-contracts/project-configuration-visual.md`).
**The motion step — record every motion the prompt names.** With `motions: none`, the report's `motion:`
line reads n/a and the step is skipped. Otherwise, for each motion: in a Chromium page with
reduced motion off (`page.emulateMedia({ reducedMotion: 'no-preference' })`) and the viewport and
clock step 8 pins, bring the view to the motion's start, then record frames from before its
trigger until the view settles — a CDP screencast (`Page.startScreencast`, each frame acknowledged
with `Page.screencastFrameAck`) — while driving the trigger (a tap, a navigation, a drag through `page.mouse` in small
steps). Never sample with `page.screenshot`: the Chromium page this step mandates always
carries the screencast, and sampling skips the frames it catches (kan-995). **A motion that changes an element's height is recorded in a viewport tall enough to hold that element's settled height** — step 8's viewport made taller for that recording alone — since an element whose bottom edge leaves the viewport reads as no height change (kan-1008: an 816px viewport twice reported "no height animation"). Write the frames, in order, to `<changeRoot>/visual-verification/motion/<motion
id>/<nnn>-<ms>.png`, `<ms>` the frame's offset in milliseconds from the trigger — the screencast
frame's own timestamp — read every one, and
judge the strip frame by frame: **every frame shows the view
being left or the view being entered, never an empty, default or wrong-but-plausible rendering
between them; chrome (a title, a bar) and content change in the same frame; a removed motion's
strip shows no intermediate frame; a gesture's content follows the pointer in every frame and
settles where the motion's end state says** — plus any frame rule the project's own instructions
state. A frame that breaks one is a departure, named by its index and what it shows.
**A one-frame departure is confirmed before it is reported.** A departure lasting one frame or
less is reported only once confirmed — by a DOM read at the next animation frame after the trigger
(`page.evaluate` resolving inside a `requestAnimationFrame` callback, reading the departing
element's presence or computed style) or by a second recording showing the same frame; an
unconfirmed one is not a departure (kan-1048 self-review, 8-A: a single screencast frame showed a
tooltip lingering after `mouseleave`, yet React flushes that discrete event before paint and a
`requestAnimationFrame` DOM read showed no lag — the report cost a whole fix round).
**A re-record replaces the strip.** A later round's recording of a motion first deletes `<changeRoot>/visual-verification/motion/<motion
id>/` and writes its frames there, never into a suffixed sibling such as `<changeRoot>/visual-verification/motion/<motion
id>-r2/`, and `visual-verification.md` cites an earlier round's strip by its round number,
never by a path (kan-995: superseded strips were 28.6 MB of a 43 MB archive).
**A difference between two triggers is judged at one sampling.** Before reporting that one
trigger draws a frame its sibling does not — Back against the forward navigation, a gesture
against the tap it replaces — record the sibling in the same session by the same capture method
and compare the two strips by their `<ms>` offsets; a frame one strip lacks only because its
sampling skipped that instant is no departure.

11. **Write `<changeRoot>/visual-verification.md`** — one entry per view: its screenshot
    path, resolved by the same recursive search step 9 used, and what was seen; and, per composed
    pair, the composite's path in `<changeRoot>/visual-verification/`, the frame id, its `diff=`
    ratio, what was seen, its
    band pairing's unpaired and over-tolerance bands with their causes, its seam pairing's
    unpaired seams and off-centre lines with theirs, and the
    frame's element × property matrix from step 10. **Every evidence file the record cites is in
    the repository when the record is written.** A cited file the change has not itself placed in
    the repository — a step-9 capture — is copied into
    `<changeRoot>/visual-verification/evidence/` first, and the entry cites
    the copy
    by its path relative to the record, naming the `evidence` subdirectory and the file's own
    name — a subdirectory
    the compose step never writes into, so a copy cannot take a composite's name; a file the
    change has already placed in the repository — a committed baseline, or the composite the
    compose step just wrote into `<changeRoot>/visual-verification/` — is cited at that
    in-repository path. No entry cites the
    worktree-absolute path of a file the worktree holds alone. The report block below keeps its absolute paths: it is the live
    run's handoff, not the committed record. **One entry per motion the motion step recorded**: its
    id, its strip's directory relative to the record, its frame count, and each departing frame's
    index and what it shows — or `clean`.

**The wire-contract check — when the prompt names one.** With no `wire contract:` line, the
report's `wire contract:` line reads n/a. Otherwise, signed in as an account
`<project>/.flow/project.md`'s `## credentials` declares, call each named endpoint on the running
API at its worktree-resolved URL, read the response's shape and wire names against the line, then
drive the stated acceptance behaviour through the API and read the next response — a screenshot
comparison is no substitute for it. With a `wire contract:` line the dispatch names no view, so
steps 8, 9 and 10 read n/a — wire contract, no view drawn — never `done` over a capture that
never ran (kan-1022's self-review).

**No required step is left undone for want of time or context.** The report's `steps:` line gives
each of steps 4, 8–11 and the motion step a status `check-verify-report.sh` admits — its header
is canonical for them — and the parent runs that guard on the report before reading it.
Run it on your own report before handing back: an exit 1 names the steps still owed, so do them
and run it again, and hand back an incomplete report only for a blocker you cannot cure, named.
**A departure found never shortens the pass:** every step still runs to its end in the same
dispatch, and a fix round's re-verify finishing what this one left is no reason to leave it
(kan-1008: a verifier left step 10 "by choice" once it had found departures, and two remainder
dispatches cost 19 and 10 minutes).
**Step 10 is `done` only when every frame's matrix and sweeps are**: a row for every element the
frame draws, every cell filled as step 10 states, and every sweep on each `sweeps:` and `added
sweep` line `done` or `n/a — <why it does not apply>` — a sweep or a cell left for time, context
or a measurement the pixels could not give is step 10 undone.
Each composed frame carries its own `- <frame id> matrix:` line, never one line for a range of frames: `check-verify-report.sh` refuses a `done` step 10 while a frame with a `- <frame id>: <composite path> diff=` line has none (kan-1016: one line covered nine frames, and the warm-up sheet's frame had no matrix until round 8).
It refuses one the same way for a missing `bands:`, `seams:` or `over-tolerance:` line, an empty
matrix cell, an n/a cell in a column that applies to its row's class and not reading
`no mockup source`, and a sweep item reading neither `done` nor n/a — `<why it does not apply>`.

**Hand back the `## Verdict` alone.** Write the report file as a `## Verdict` block with the full
`## Report` under it, and end your turn with the `## Verdict` block alone — the parent reads nothing
else. `exits:` names every exit the `## Report` carries, `not declared` where a command is not;
`cause:` appears only when a step is `blocked`, naming `environment`, `test-failure` or
`missing-fixture` from the report's own facts; `departures:` counts every departure the `## Report`
names, each then on a line of its own with its cause or attribution beside it.

```text
## Verdict
- Model: <the model named in your own system prompt>
- exits: setup <n | not declared>, capture <n>, full app suite <n>, spec reach <n>
- steps: <the ## Report's steps: line, after `steps: `>
- cause: environment | test-failure | missing-fixture
- departures: <n>
- <frame id or view>: <the departure, with its cause or attribution>
```

```text
## Report
- setup: <not declared | exit <n>>
- stack: <already running | started and stopped | could not be started — <output>>
- fingerprint: <not declared — <step 6's reason, from the prompt> | exit 0 | mismatch → restarted → exit <n>>
- verify: green — <the inline verify row the bundle names>
- capture: exit <n>
  <output, verbatim or last 40 lines>
- full app suite: <created | updated> — <absolute spec path>, <n> screens, capture exit <n>
- full app suite zip: <absolute zip path>, <n> PNGs | not written — <reason>
- spec reach: exit <n>
- identical: <the PNG names of one byte-identical group> — <why the views draw the same pixels, or the defect> | none
- spec: <absolute spec path>
- <view>: <absolute PNG path> — <what was seen, including any defect>
- measure: exit <n> — <absolute manifest path>
- mockups: <not declared | no map for <spec> | exit <n>>
- <frame id>: <absolute composite path> diff=<ratio> — <match, or the departure seen>
- <frame id> bands: <paired>/<frame's band count> paired, <missing> missing, <extra> extra — <every missing and extra band by `top` and `height`, and every pair over the tolerance by its delta, each with its named cause, or `departure` attributed per **Every departure is attributed at the merge base** above>
- <frame id> seams: <paired>/<frame's seam count> paired, <missing> missing, <extra> extra, <lines paired> lines — <every missing and extra seam by its band's `top` and its `left`, every seam pair over the tolerance by its delta, every cell whose line count differs, and every line whose `offset` delta is over the length tolerance with its `left` and `right` deltas, each with its named cause, or `departure` attributed per **Every departure is attributed at the merge base** above>
- <frame id> sweeps: text | order | reach | derived | rows | ink — <each `done`, or `n/a — <why it does not apply>`; `rows` names each bounded row and its four gaps per image; `ink` names every non-text run the frame draws and its counterpart in the capture, or the one absent>
- per view, no frame composed: <view id> sweeps: text | order | reach | derived | rows |
  ink — the same sweep line at capture scope; `order` carries its containment half's
  outcome beside n/a — no frame; `rows` and `ink` read n/a — no frame
- per view, no frame composed: <view id> bands | seams | matrix — n/a — no frame
- added sweep <name> (sweeps-<n>.md): <`done`, or `n/a — <why it does not apply>` — every view and frame it read, and each departure it found>
- <frame id> matrix: <n> elements × 11 columns, <k> n/a — <each n/a cell as `<element>.<column>: <why>`; the matrix itself is in visual-verification.md>
- <frame id> over-tolerance: <n> — #<row>: <cause>; … <every row the manifest's `over` list names for the frame, each with one cause from step 10's closed list>
- frames: <n>/<m> — <m> the change's own declared list, then every declared frame id with no line above and why
- motion: <named motions recorded>/<named> — <each motion id: its strip directory, frame count, and `clean` or each departing frame's index and what it shows> | n/a — no motions named
  - <a strip recorded beyond the named motions>
- wire contract: <method> <path> — <the shape and wire names read, and the acceptance behaviour seen, or each departure> | n/a — no wire contract named
- visual-verification.md: written | not written — <reason>
- steps: 4 <status> | 8 <status> | 9 <status> | 10 <status> | 11 <status> | motion <status>
```
