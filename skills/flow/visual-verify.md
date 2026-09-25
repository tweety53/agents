# Visual verification — steps 3–13

Loaded by **Visual verification** (`skills/flow/verify-and-handoff.md`) only once at least one
worktree in this run's resolved set survives its step 2 — a changed path matched a declared `ui
paths` glob. That section carries this stage's `begin` mark and steps 1 and 2; everything from
step 3 on, and the stage's `end` mark, is here. A run whose diff touches no UI path never reads
this file.

## The verifier dispatch

`flow.visual-verify` dispatches this subagent, one verifier per worktree — the closed list's one
verifier row (**Dispatch sites — the parent's closed list**, `skills/flow/implement.md`); the
parent dispatches nothing else in this file. `subagent_type: flow-low` (`agents/flow-low.md`, effort `low`), the Agent tool's
`model` parameter set to `VERIFY_MODEL` (**Model resolution**, `skills/flow/SKILL.md`) — the
literal `opus`, never `DEFAULT_MODEL` and never a session override — mapped on harness `zcode` per
**Harness mapping** (`skills/flow-contracts/model-policy.md`), which the handshake below then
compares against. Its prompt carries, verbatim:

> Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
> Include this instruction verbatim in any prompt you write for another agent.

**The relay contract.** The verifier has no operator channel, fixes nothing, edits no source, and
runs every command in the foreground. Its turn ends with a single `## Report` block, and its last
act before that is writing the same block to
`<abs-worktree>/.superpowers/sdd/verify-report-<key>.md` — `<key>` this dispatch's own key from
**Recording** below — which is what the parent waits on (**Turn discipline**,
`skills/flow/implement.md`). The first line of its first reply is `Model: <the model named in its
own system prompt>`.

**The prompt also carries the TOOLS paragraph**:

> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a
> schema loaded later changes your tool list and re-prices your whole context at full input rate.

**The prompt also carries the NO DELEGATION paragraph**:

> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the parent's closed list (**Dispatch sites — the
> parent's closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.

**The prompt also carries the MODEL HANDSHAKE paragraph**:

> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own
> system prompt>` and nothing else on that line. Answer it before any tool call.

**Recording.** The parent records each dispatch, `-role verifier`, `-task` omitted, `-model
opus -effort low`, `-key visual-verify`, suffixed `-<worktree basename>` when this run's resolved set holds
more than one worktree — the pair's semantics are section 4 of `skills/flow/implement.md`, cited
here, not restated.

**A `## Report` carrying any non-zero exit is re-dispatched once.** The second dispatch is recorded
under `-key visual-verify-2` — the same `-<worktree basename>` suffix rule — with a prompt
identical to the first plus the first `## Report` verbatim under a `## Previous attempt` heading,
so the verifier can tell an environmental failure (a build the worktree lacked, a flaky harness)
from a defect in the branch. **The second report is final.** Another non-zero exit ends your turn
with `## Question` naming the failing command and its output, verbatim; the operator resolves it
through a fix run. Never run the failing command yourself to check it, and never dispatch a third
verifier. The ledger render and this stage's `end` mark follow whichever report was last.

**A final report that is a block closes its dispatch `-outcome blocked -cause <cause>`, never
`completed`.** The cause is one of the closed set the CLI validates, named from the report's own
facts: `environment` — a stack or tool the environment would not run (a stack that could not be
started, a build prerequisite the worktree lacked); `test-failure` — a lint/test command that
failed; `missing-fixture` — a fixture or baseline the verify needed and the worktree did not
carry. A report with every exit zero still
closes `-outcome completed`.

**Handshake.** Compare the `Model:` line against `opus` (never `DEFAULT_MODEL` or a session
override) and apply **The handshake** (`skills/flow/implement.md`, **The parent orchestrates directly**),
unchanged: a first mismatch closes `<key>` `-outcome fallback` and re-dispatches once under
`<key>-retry`; a second mismatch closes `<key>-retry` `-outcome fallback` too and ends the turn
with `## Question` naming `opus` and both models that answered, options **Continue on `<the
model the second handshake named>`** or **Stop the run**.

A verifier that ends without a `## Report`, or whose agent dies, is closed `-outcome aborted`
and blocks this handoff exactly as a failed command would, naming the death.

## Steps 3–13

Steps 1, 2, 3 and 12 are the parent's — those steps, `prepare-workspace.sh` and the ledger render
are the parent's own Bash calls, never a subagent's. Steps 4–11 and 13 are run by one verifier per worktree
surviving steps 1–3, dispatched per **The verifier dispatch** above with `-key visual-verify`; the
parent applies **Blocking** to its report. Its prompt states: the absolute worktree path; the
`KEY=value` lines **Verify** (`skills/flow/verify-and-handoff.md`) exported for it; this section's resolved `setup`, `verify`, `capture`,
`fingerprint` and `start` commands and `screenshots` root, its resolved `mockups` root when
declared, and its `mockup frame` value when declared; the worktree-resolved URL of each app `ui paths`
matched; the project's `## run` commands; the views touched; `<changeRoot>`; and to run steps 4–11
and 13 below as written, committing and pushing nothing.

3. **Pre-flight the workspace, before anything is dispatched.** The verifier's one re-dispatch
   cannot repair an environment that cannot pass. With the parent's own Bash calls, before the dispatch below, check the environment
   this worktree will verify in:

   - **Ports held only by their own app.** For every port this worktree's resolved environment
     carries — the `port` row's resolved value and the port in each matched app's
     worktree-resolved URL — a listener is a pre-flight failure only when nothing answers it. The
     probe is the app's worktree-resolved URL where the port reaches this check through one;
     where it arrives only through a `port` row, the probe is the port itself — a listener that
     accepts a connection is answering, one that accepts nothing is a dead holder. An answering
     listener is the already-running stack steps 5 and 6 probe and fingerprint and step 13 leaves
     alone — the pipeline's own run-instructions rule starts that stack on every run and hands it
     running to the operator, so a fix run re-entering on a held port is the ordinary case, never
     a failure. A held port whose probe answers nothing — a foreign holder, a sibling workspace,
     yesterday's orphan — fails the pre-flight (`lsof -nP -iTCP:<port> -sTCP:LISTEN`, the probe
     **What the id derives** (`skills/flow-contracts/workspace-isolation.md`) defines), naming
     the port, the holder `lsof` prints, and the probe that went unanswered.
   - **The app's base-URL configuration resolves to the workspace's port** — `ApiBaseUrl`, where
     this failure was measured. For every `## workspace isolation` row whose resolved value (the
     `KEY=value` line **Verify** exported) differs from the row's declared default, the
     application configuration under the matched `ui paths`' roots must take that value from the
     exported variable or from the `start`/`## run` command — a literal naming the declared
     default there, which neither the exported variables nor that command overrides, is the
     hardcoded-port failure; name the file and the line.
   - **Origin allowed.** The worktree's allowed-origins configuration (`ALLOWED_ORIGINS`, where
     this failure was measured) must include the worktree-resolved URL of every matched app — an
     origin list pinned to the declared default URL fails; name the file.
   - **One Playwright checkout per workspace.** Resolve the Playwright module from each matched
     app's package root (`node -e "console.log(require.resolve('@playwright/test/package.json',
     {paths: [root]}))"`) — the resolved absolute path must sit inside this worktree. A resolution
     landing in a sibling worktree or a machine-global checkout is the cross-worktree module
     conflict; name the path, and install this worktree's own before re-running.

   **Any failing check ends the stage here**: no verifier is dispatched, the stage's `end` mark
   carries `-outcome stopped`, and the report names every failing check with the evidence above.
   The run proceeds to the `IN_PROGRESS` handoff, which names the environment cause — the operator
   fixes the environment and re-runs. This is a handoff, never a `## Question` and never a
   re-dispatch: the re-dispatch below is for a verifier report, not for an environment this stage
   has proven cannot pass.

4. **Run `setup`, if declared.** A non-zero exit blocks, printing the command verbatim.
5. **Probe before starting anything.** Probe the URL of each app `ui paths` matched, resolved for
   this worktree per **What the id derives** (`skills/flow-contracts/workspace-isolation.md`) —
   never the project's declared default. If nothing answers, start the stack from `start` when
   declared, else `## run`, and record that this stage started it — needed at step 13.
6. **Fingerprint the served bundle, if `fingerprint` is declared.** A screenshot is evidence only
   of what the app was serving when it was taken, and a stack step 5 found already running may be
   serving a build older than the worktree. Run `check-dev-stack-fresh.sh <worktree>` — it reads the row and
   runs it, so the stage never parses the table by hand; its header is canonical for its three
   exits. Exit 0 → the served bundle is the worktree's build; continue. Exit 1 → stop the stack,
   start it from `start` when declared, else `## run`, record that this stage started it (step 13
   stops it), and run the guard once more. A second exit 1 blocks, carrying the guard's output.
   Exit 2 → report `fingerprint: not declared — <the guard's stderr reason>` and continue; the
   report makes the gap visible in every handoff, but this stage cannot prove what it was never
   told how to check.
7. **Run `verify`.** A non-zero exit blocks — except a failing test **the sweep**
   (`skills/flow-contracts/known-bugs.md`) classifies pre-existing, which takes that contract's
   known-failure course instead.
8. **Capture** — author a spec covering the views this change touched, then run `capture` with
   `<spec>` substituted for the spec's path. `screenshots`'s root-not-leaf shape is canonical in
   `skills/flow-contracts/project-configuration.md`; nothing here restates it. **Every screenshot
   this spec takes is the full page or viewport, never a clipped region.** A clip is the right tool
   for an implementer's own targeted assertion (a fixed piece of text, an icon), but this stage's
   own job — page-wide styling (background, shadow, border, font, spacing) matching the mockup — is
   exactly what a clip is built to hide; step 10 below cannot compose a clip against a full mockup
   frame and call the result a fidelity check. **`capture` creates this change's baseline**: writing
   a PNG that does not yet exist is its success path, not a failure — `verify` is the regression gate
   over an already-committed baseline, `capture` is not, and only a `capture` failure for some other
   reason blocks (see **Blocking** below). A `toHaveScreenshot` passing over a baseline this
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

   **Every mockup frame this change adds or updates gets a capture and a sidecar line — never
   optional, with `mockups` declared.** The frames are the change's declared list (step 10) plus
   every `<frame id>.png` that
   `git -C <worktree> diff --name-only <merge-base>..HEAD -- <mockups>` prints. For each one the
   capture spec takes a screenshot of the view that frame draws — added where the spec has none,
   updated where the frame changed what it draws — and the `<spec>.mockups` sidecar carries its
   `<screenshot name> <frame id>` line, written by this stage where the change brought none. A
   frame on that list with no capture or no sidecar line blocks.

   **Then update the full app suite — and create it where the checkout has none.** The suite and
   its file names are canonical in **The full app suite**
   (`skills/flow-contracts/project-configuration.md`). Absent → author `full-app-suite.spec.ts`
   beside the capture spec, one full-page capture per screen the app has — enumerated from the
   app's own routes or navigation and from every spec already in the checkout, never from this
   change's diff. Present → add a capture for every screen this change added and update the
   capture of every screen it changed or removed. Run `capture` with `<spec>` substituted for the
   suite's path; a non-zero exit blocks as any `capture` failure does — except a failing test
   **the sweep** (`skills/flow-contracts/known-bugs.md`) classifies pre-existing, which takes
   that contract's known-failure course instead. Then rebuild the zip from
   exactly what the suite just produced:

   ```bash
   rm -f <screenshots root>/full-app-suite.zip
   resolve-visual-screenshots.sh <worktree> full-app-suite.spec.ts | zip -X -j -@ <screenshots root>/full-app-suite.zip
   ```

   `<screenshots root>` is `screenshots` resolved the way `resolve-visual-screenshots.sh`
   resolves it. A resolver exit 1 or 2, or a non-zero `zip`, blocks.

   Then run `check-spec-reach.sh <worktree>` — the spec
   `capture` just wrote must be reached by a `package.json` script of the `regression checkout`;
   exit 1 (an orphan, named) or 2 (cannot answer) blocks.
9. **Read every captured PNG — resolve their paths with the guard, not by eye.** Run

   ```bash
   resolve-visual-screenshots.sh <worktree> <spec's basename>
   ```

   Exit 0 prints one absolute PNG path per line — that is what selects this run's fresh output from
   the committed baseline PNGs already sitting under the same `screenshots` root, rather than joining
   `screenshots` with a guessed filename. Exit 1 (zero matches) blocks, indistinguishable from a view
   that never rendered; exit 2 (cannot answer) blocks the same way. **Read every printed path** — no
   script can do that — and state, per view, what was seen. An unreadable PNG is reported and blocks
   too.
10. **Compose captured frames against their mockups, if `mockups` is declared.** With a
   `<spec>.mockups` sidecar beside the capture spec, run

   ```bash
   resolve-visual-screenshots.sh <worktree> <spec's basename> | compose-mockup-frames.sh <spec>.mockups <worktree>/<mockups> <changeRoot>/visual-verification '<mockup frame value>'
   ```

   The fourth argument is this project's resolved `mockup frame` value, quoted; omit it entirely
   when the row is absent.

   Exit 0 prints one `<composite path> diff=<ratio>` line per pair — **read every one** and state,
   per pair, what the difference panel's white regions are, the ratio, and where the capture
   departs from the frame. The ratio is information, never a threshold: a pair blocks on a
   departure you can name, not on a number. Exit 1 (a broken map — a screenshot no capture
   matched, a frame file absent, a malformed line, a capture whose size differs from the cropped
   frame) or 2 (cannot answer — a malformed `mockup frame` value, Pillow absent, included) blocks.

   > **The `mockups` declaration gates the frame comparison, never the sweeps.** The compose
   > step, the band pairing, the seam pairing, the ink inventory (sweep 10), the bounded-row
   > sweep (sweep 9) and the element × property matrix read a frame against the capture, and a
   > project declaring no `mockups` root composes nothing and reports each of them
   > n/a — no frame. The state sweeps run on the captures alone and are the default on every
   > `flow.visual-verify` run, composed or not: a verification that transcribes no text run
   > (sweep 5), checks no element against its container's bounds (sweep 6), scrolls no region
   > to its end (sweep 7) and re-derives no derived value (sweep 8) is the operator-driven
   > round this stage exists to replace. Where a sweep's wording names the frame for something
   > the capture answers without one — sweep 7's last element reached by name, sweep 5's
   > display precision — the run reads that name against the view's own rendered content;
   > sweep 6's order comparison needs the frame's list and is reported n/a — no frame
   > beside the other frame-bound readings.

   **The difference panel cannot show a structural departure, so the first read of every pair
   is the script's band pairing, never the panel.** The panel is white wherever any channel
   differs, and on a real pair 20–50% of it is white from seed data, font rasterisation and the
   frame's own annotations alone. A 1px rule the capture omits, a row padding it lacks, a surface fill
   swapped for the page colour, a button border in the wrong colour: each moves the ratio by
   under a hundredth and vanishes in that white. Beside every composite the compose step writes
   the cropped frame as `<composite stem>.frame.png`, the capture's own size; for every pair, run

   ```bash
   measure-visual-properties.sh <changeRoot>/visual-verification/<stem>.frame.png <capture PNG> --scale 1 --props bands,seams
   ```

   and read `delta.bands` (the script's header is canonical for the property): every band the
   frame draws, paired with the capture's in order, a text or data difference never unpairing
   one. **Every `missing` band, every `extra` band, and every pair whose `since_pair`,
   `gap_above` or `height` delta exceeds 2px or whose `edge` or `colour` distance exceeds 12 is
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
   seam pair whose `since_pair` or `width` delta exceeds 2px or whose `colour` distance exceeds
   12, every cell whose line `count` differs, and every line whose `offset` delta exceeds 2px
   is a departure until a cause is named that a seam or a line cannot carry.** Text width never
   moves a seam, so a data difference explains no seam; a line's width is data, so an `offset`
   delta whose `left` delta is within 2px (or whose `right` is) is a left- (right-) anchored
   label whose text differs — the cause is then the transcription, and it is named — where an
   `offset` delta with `left` and `right` both moved by the same amount is a label anchored to a
   different edge than the frame's, a departure whatever the text reads. What the seams cannot
   see — a divider in a row with no border or fill, the label of a filled cell, a glyph, a
   corner radius — the per-control sweeps below still measure.

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
   itself when no such citation exists; a frame reached only by inference from another (already
   covered by the composite diff, "shipped" and pre-existing, blocked by a real environment limit)
   still gets named, with the reason. **The list is the change's, never the round's**: a fix run
   re-verifies the same list, not the views its fix touched, and the specs `capture` runs are
   chosen to reach every frame on it, never the other way round — a denominator the verifier
   derives from the specs it happened to run is the partial pass wearing a fraction. The report's
   `frames:` line counts against that list and names every declared frame absent from the
   per-frame lines, with its reason; the parent reconciles it against `design.md`'s own frame
   list before applying **Blocking**, and a declared frame with no line blocks as a departure
   would.

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
      list should have excluded instead of offered. `rules/design-mockups-are-specs.mdc` puts the
      same question to the implementer; this is the verifier's side of it.
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
      complaint, and a row nobody claimed anything about was never measured. Nothing else here
      sees it — the composite's ratio barely moves when a row's content is right and only the
      whitespace around it is wrong, the per-control crop matches its control wherever it sits, and
      the order-and-containment sweep finds every element present and inside its container. A row
      whose every control measures the right size is the case to suspect, not to pass: a uniform
      container padding around correctly-sized controls is invisible to every reading but this one.
   10. **Every non-text ink the frame draws — hairline, divider, rule, border, background fill,
      shadow band — inventoried from the frame alone, then found in the capture one by one.**
      Every sweep above starts from something the capture shows or the text says: a control, a
      text run, a row, a value. A line the frame draws and the capture omits is on none of those
      lists, and the composite's ratio barely moves for a 1px rule, so nothing above asks for it
      and a verifier who has matched every label and number walks away satisfied. The
      horizontal inks — every hairline, rule, fill band and border edge the frame draws across
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
      cells in the new matrix, never its rows.
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

      Every cell holds the frame's number, the capture's and the `delta` — or n/a — `<why>`: the
      column does not apply to the row's class (the dashes above), or the script exited 1 on a
      region no widening resolved, with the region tried. A cell left empty, or a cell filled by
      eye, is the sweep-name report again. The matrix rides `visual-verification.md` (step 11)
      under the frame's entry, and the report's `matrix:` line counts its rows and its n/a
      cells; the parent opens the frame beside the matrix and a visible element with no row, or
      an empty cell, blocks as a departure would.

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
   included, and a fill covering half the box flips which colour counts as `fill` (both shown
   on the incident's own geometry by `<agents repo>/scripts/test-measure-visual-properties.sh`). The same
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
   with a real frame, then run the compose command above — composing against a sidecar this stage
   just wrote is not a special case, and it is committed at step 12 along with everything else this
   stage writes. No plausible match anywhere in the directory → report `mockups: no map for
   <spec's basename> — searched <mockups dir>, no frame for <views>`, naming what was searched, and
   continue. Not declared → report `mockups: not declared` and continue. The sidecar's shape is
   canonical in **visual verification** (`skills/flow-contracts/project-configuration.md`).
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
    run's handoff, not the committed record.
12. **Commit the spec and its PNGs, and stop there.** A declared `regression checkout` receives
    them; with none declared, commit to the change's own branch instead. **The commit is
    pathspec-scoped** — `git add --` the spec, its PNGs and `<changeRoot>/visual-verification/`
    first, then `git commit -m "<subject>" --` those same paths — the add first because the
    stage's outputs are new untracked files, which a pathspec commit cannot pick up — carrying
    only what this stage wrote, never a bare `git commit`: the index of a main checkout
    may carry a pre-staged foreign tree (**A commit a run instructs defaults to the pathspec-scoped form**,
    `skills/flow-contracts/git-boundaries.md`). **Resolve the
    `regression checkout` root the same way every other declared app root in this file is
    resolved** — from `git worktree list` in that repository, or the state file's `worktrees`
    map, per **Roots in `## apps` are main checkouts** (`skills/flow-contracts/project-configuration.md`) <!-- refs-guard:allow -->
    — never the main checkout while a worktree for it holds the change's work. A `regression
    checkout` not also declared in `## apps` has no worktree to resolve and this step commits to
    the main checkout directly. **Never push** — see `no-automatic-push`
    (design.md): a file inside a repository cannot authorise a push to another repository, so no
    guard here grants one, worktree or not. `regression repo` records which repository the
    checkout is expected to be, and `check-visual-verification.sh` reports a mismatch against
    its real `origin`, but that is an identity assertion, not an authorisation. When a commit landed
    in a `regression checkout`, print the push command for the operator to run by hand, naming
    whichever branch actually received the commit — the change's own `spectre/<name>` when a
    worktree resolved, the main checkout's current branch otherwise:

    ```bash
    git -C <regression checkout> push
    ```

    `<changeRoot>/visual-verification/` is committed with the change root.
    **The full app suite rides the same commit**: `full-app-suite.spec.ts`, its PNGs and
    `full-app-suite.zip` (step 8) join the pathspec above, in the same checkout the capture spec
    lands in.
13. **Stop the stack only if step 5 or step 6 started it.** A stack the operator already had running is left
    alone.

```text verified:design.md section 3 of this change
## Report
- setup: <not declared | exit <n>>
- stack: <already running | started and stopped | could not be started — <output>>
- fingerprint: <not declared | exit 0 | mismatch → restarted → exit <n>>
- verify: exit <n>
  <output, verbatim or last 40 lines>
- capture: exit <n>
  <output, verbatim or last 40 lines>
- full app suite: <created | updated> — <absolute spec path>, <n> screens, capture exit <n>
- full app suite zip: <absolute zip path>, <n> PNGs | not written — <reason>
- spec reach: exit <n>
- spec: <absolute spec path>
- <view>: <absolute PNG path> — <what was seen, including any defect>
- mockups: <not declared | no map for <spec> | exit <n>>
- <frame id>: <absolute composite path> diff=<ratio> — <match, or the departure seen>
- <frame id> bands: <paired>/<frame's band count> paired, <missing> missing, <extra> extra — <every missing and extra band by `top` and `height`, and every pair over the tolerance by its delta, each with its named cause or `departure`>
- <frame id> seams: <paired>/<frame's seam count> paired, <missing> missing, <extra> extra, <lines paired> lines — <every missing and extra seam by its band's `top` and its `left`, every seam pair over the tolerance by its delta, every cell whose line count differs, and every line whose `offset` delta is over 2px with its `left` and `right` deltas, each with its named cause or `departure`>
- <frame id> sweeps: text | order | reach | derived | rows | ink — <each done, or why not; `rows` names each bounded row and its four gaps per image; `ink` names every non-text run the frame draws and its counterpart in the capture, or the one absent>
- per view, no frame composed: <view id> sweeps: text | order | reach | derived | rows |
  ink — the same sweep line at capture scope; `order` carries its containment half's
  outcome beside n/a — no frame; `rows` and `ink` read n/a — no frame
- <frame id> matrix: <n> elements × 11 columns, <k> n/a — <each n/a cell as `<element>.<column>: <why>`; the matrix itself is in visual-verification.md>
- frames: <n>/<m> — <m> the change's own declared list, then every declared frame id with no line above and why
- visual-verification.md: written | not written — <reason>
```

**Blocking.** This stage blocks the `IN_PROGRESS` handoff on: **a workspace pre-flight that failed
(step 3), which ends the stage before any dispatch**, a failed `setup`, a failed `verify`,
a genuine `capture` failure — the full app suite's included — **a frame this change added or
updated with no capture or no sidecar line, a full app suite or a `full-app-suite.zip` step 8 did
not write**, a stack that could not be started, **a `fingerprint` that still exits
non-zero after step 6's restart**, a `check-spec-reach.sh` exit 1 or 2,
an unreadable PNG, **a `compose-mockup-frames.sh` exit 1 or 2 and a departure from the mockup the
verifier reports in a composite**, **a frame on the change's declared list with no line in the
report, a composed frame with no `bands:` line or with a `missing`, `extra` or over-tolerance
band whose line names no cause, a composed frame with no `seams:` line or with a `missing`,
`extra` or over-tolerance seam or an off-centre line whose line names no cause, and a composed frame with no `matrix:` line, a matrix row missing for an element the
frame visibly draws, or a cell that is neither the script's numbers nor an n/a with its reason
— the parent's own reconciliation, step 10**, and **a defect the
verifier reports in a captured screenshot — even when every assertion passed.**

```bash
flow stage end -command '/flow' -stage flow.visual-verify -outcome completed <name>
```

**A step-3 pre-flight failure closes this mark `-outcome stopped` instead of `completed`**, per
that step — the one outcome variant this stage's end mark carries.
