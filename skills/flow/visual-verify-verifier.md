# Visual verification — the verifier's steps

Read by the verifier **Steps 3–13** (`skills/flow/visual-verify.md`) dispatches, at the absolute
path its prompt carries — the parent never loads it whole. It carries step 4, steps 8–11, the
motion step and the `## Verdict` and `## Report` templates, run as written.

**A command the harness denies has no exit code.** Its report line reads `blocked — environment:
<the denial text, verbatim>` where `exit <n>` would stand, and the step's status is `blocked` —
never an exit code the command did not return.

**The `exits:` line never names the dispatch's key**: `close` reads every number on it as an exit.

**Run only the steps below.** Add no pre-check of your own — no reading `.env`, credentials or
config files, no probing the stack beyond what a step names. A step that cannot proceed is
reported BLOCKED, naming it.

4. **Run `setup`, if declared.** A non-zero exit blocks, printing the command verbatim.
8. **Capture** — when `specs` is declared, run it first from the `regression checkout`, with
   `<frontend-root>` substituted for the worktree and `<merge-base>` for the merge base the bundle
   names; a non-zero exit blocks, printing the command and its exit. The verify `specs` maps for that
   merge base already ran green in `flow.verify`, which the report's `verify:` line names as the
   bundle does — this step runs no `verify`. The capture spec is the touched views' existing spec from `specs`' output,
   extended, when one covers the view; a new spec is authored only for a view none covers. **A capture
   spec carries a `<spec>.sources` sidecar** naming the source files its screen is made of, written
   or extended with the spec; `flow-guard affected-specs`'s header is canonical for its lines. Run
   `capture` with `<spec>` substituted for the spec's path. **One `capture` run takes the specs of the bundle's `own specs:` line** (the change's own sidecars, by the rule of `skills/flow/visual-verify.md`, step 3) **and any capture spec you author or extend this pass for a new frame — never another spec `specs` prints**, less every spec the bundle names on a `captured:` line, their paths space-joined for `<spec>`, with
   `--workers=$(flow-guard capture-workers) --retries=1` before `--update-snapshots=all` — never one run per
   spec. With every spec on the `captured:` line, or `own specs: none` and no spec of yours, no capture runs — an empty `<spec>` is never the whole suite. A test that passes on its retry is flaky: its capture stands, the report names it as a defect fixed at its source, and the capture is never run again for it. A project's page limit for ad-hoc probes does not bind the capture suite. A capture whose
   PNG differs between that count and `--workers=1` is nondeterministic, and is fixed at its source
   (`rules/fix-determinism-at-the-source.mdc`), never tolerated. **Every capture asserts the view is on
   screen before it is taken** — gymie-playwright's `expectRoom` — so a missed navigation fails the
   capture instead of becoming its baseline. **A capture reached through a click also asserts the state it clicked into, and parks the pointer before it is taken** — a positive control on a drawn, state-dependent property the wrong state cannot satisfy (the selected tab's fill, the opened sheet's tag), and the pointer moved clear of every control (`page.mouse.move(0, 0)`) after the last click — so a missed click fails the capture instead of becoming its baseline, and no hover ink reaches one. **A capture reaches every control it clicks or types into by a structural locator, never a page coordinate** (**D. Basic Workflow #3 — Writing plans**, `skills/flow/brainstorm-planner.md`). **A capture whose subject sits under a scrim, or any dimming or alpha layer, also pins that subject structurally** — the presence of each tag the change draws beneath the layer and the absence of each it removes, covered tags included — because the dimming can shrink a real difference below the screenshot comparison's per-pixel threshold; a dimmed capture whose baseline predates the change is recaptured, never trusted because its diff passes. `screenshots`'s root-not-leaf shape is canonical in
   `skills/flow-contracts/project-configuration-visual.md`; nothing here restates it. **Every screenshot
   this spec takes is the full page or viewport, never a clipped region.** A clip is the right tool
   for an implementer's own targeted assertion (a fixed piece of text, an icon), but this stage's
   own job — page-wide styling (background, shadow, border, font, spacing) matching the mockup — is
   exactly what a clip is built to hide; step 10 below cannot compose a clip against a full mockup
   frame and call the result a fidelity check.
   The one exception is a region an approved prototype draws alone (step 10): its element capture is
   the frame-bound capture, taken beside the view's own full-page capture, never in place of it.
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
   instant, and leaves real or masks only what neither pin reaches.

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
   change did not introduce included unless its guard proves it pre-existing and unrelated (**the sweep**, `skills/flow-contracts/known-bugs.md`).
   Then rebuild the zip from exactly what the suite just produced with
   `flow-guard rebuild-full-app-suite-zip <worktree> <screenshots root>`, the one script every recapture
   that moves the suite runs too (`skills/flow/fix-visual.md`, `skills/flow/sync-onto-base.md`).
   Both arguments are absolute — the script refuses a relative root — and a project whose `capture`
   row chains this rebuild after Playwright has run it with `capture`, a second run rewriting nothing.

   `<screenshots root>` is `screenshots` resolved the way `flow-guard resolve-visual-screenshots`
   resolves it. Every PNG's mtime is pinned before zipping because the zip stores it: a rebuild over
   unchanged images is then byte-identical and commits nothing. A non-zero exit blocks.

9. **Measure in one call, then read in one turn.** Right after step 8's capture — after step 10's
   prototype render and any sidecar this stage writes, which come first — run, with the arguments
   the bundle names:

   ```bash
   flow-guard visual-verify-pass measure <worktree> <name> <key> <spec path> [--prototype]
   ```

   `--prototype` when step 10's approved prototype is the frame. Its header is canonical for what it
   runs — `flow-guard resolve-visual-screenshots`, the identical-baselines check (`stats/internal/guard/identicalbaselines.go`), `flow-guard compose-mockup-frames`
   and `flow-guard measure-visual-properties`'s band, seam and matrix reads on every pair — the files it
   writes and its exits; it carries on past a step's non-zero exit and records each exit and its
   stderr in the manifest it names, `<abs-worktree>/.superpowers/sdd/visual-manifest-<key>.json`, which
   each call merges into as its header states — another spec, or a `--prototype` run over the same one,
   adds its pairs and drops none — so the manifest is never edited by hand. A
   sidecar written later, or a recapture, runs it again. Then Read, in one turn of parallel Read
   calls, the composite every `frame:` line names and every capture no composite holds — a composed
   capture is read through its composite, never again on its own.
   The resolver is what selects this run's fresh output from the committed baseline PNGs already
   sitting under the same `screenshots` root, rather than joining `screenshots` with a guessed
   filename: its exit 1 (zero matches) in the manifest blocks, indistinguishable from a view that
   never rendered, and its exit 2 (cannot answer) blocks the same way. **Read every capture** — no
   script can do that — and state, per view, `clean` or the defect seen.
   In a scoped fix round `measure` also prints `changed: <frame id> … — round base <sha>`, or
   `changed: none — round base <sha>`, before its `MANIFEST:` line: read only the views of the
   frames it names, of the prompt's `moved:` frames and of its `defects:` frames, and every view no frame is composed for,
   and at step 10 only those frames' composites. Every other view's line reads
   `- <view>: <absolute PNG path> — capture unchanged since <round base>`, with no frame line for
   its frame. With no `changed:` line, every view is read.
   Each captured PNG gets its own `- <view>: <absolute PNG path>` line — the report check (`close --check-only`)
   refuses a report missing one.
   An unreadable PNG is reported and blocks
   too.

   The manifest's `identical` groups are that identical-baselines check's, whose header is canonical
   for what it compares: each names PNGs of one directory that are
   byte-identical, and the report carries one `- identical:` line per group stating why those views
   draw the same pixels, or the defect that made them, which blocks. Its exit 2 → it cannot answer, and blocks.
10. **Compose captured frames against their mockups, if `mockups` is declared.** With a
   `<spec>.mockups` sidecar beside the capture spec, step 9's `measure` composed every pair it maps
   through `flow-guard compose-mockup-frames`, with this project's resolved `mockup frame` value as its fourth
   argument — a sidecar's `frame` lines override it, each applying to the pairs below it — and printed one
   `frame:` line per pair (its header, `stats/internal/guard/visualverifypass.go`, is canonical for
   the shape). **That line and its composite are the frame's whole input: read every composite once**
   and state, per pair, where the capture departs from the frame. **Never re-render, re-crop, zoom or
   re-measure a composed frame** — no `flow-guard measure-visual-properties` call, no script crop, no
   separate Read of its capture or `<stem>.frame.png`, no parsing of the manifest's pair: the numbers
   of a `differs:` line are its over row's figures, or, for a difference no row names, as the
   composite shows them. The ratio is information, never a threshold: a pair blocks on a
   departure you can name, not on a number. A compose exit 1 (a broken map — a screenshot no capture
   matched, a frame file absent, a malformed line, a capture whose size differs from the cropped
   frame) or 2 (cannot answer — a malformed `mockup frame` value included) blocks.

   **An approved prototype is the frame where no mockup is**. When no sidecar maps a
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
   and `visual-verification.md` records the injected rule under the frame's entry, since a prototype
   panel narrower than the capture leaves no band, seam or matrix reading possible.
   Then write
   `<changeRoot>/visual-verification/prototype.mockups`, its first line `frame none` and one
   `<screenshot name> <screenshot stem>` line per capture, and run step 9's `measure` with
   `--prototype`, which composes with that map in place of `<spec>.mockups` and
   `<changeRoot>/visual-verification/prototype` in place of `<abs-worktree>/<mockups>`.
   `measure --prototype` composes from its own copy of that map under a `frame scale=1 status=52
   border=0 bare` line: the prototype frame's 52px status line is cropped, so a correct capture
   compares at zero offset, and `<stem>.frame.png` is written as for a mockup. `measure` reads bands,
   seams and the matrix on that cropped frame and the capture.
   Every frame-bound reading below then runs on those pairs as on a
   mockup's. A state the prototype draws that no capture test drives is a missing capture, and
   blocks.

   The composite's panels are the capture left, the frame middle and the difference right, unlabelled. The panel is white wherever any channel
   differs, and on a real pair 20–50% of it is white from seed data, font rasterisation and the
   frame's own annotations alone. A 1px rule the capture omits, a row padding it lacks, a surface fill
   swapped for the page colour, a button border in the wrong colour: each moves the ratio by
   under a hundredth and vanishes in that white. **The length and colour tolerances are `mvpOverLength` and `mvpOverColour`
   (`<agents repo>/stats/internal/guard/measurevisualproperties.go`)** — `measure` flags by them, and nothing here
   restates their values; the project's `tolerance` row (**Project configuration — visual
   verification**, `skills/flow-contracts/project-configuration-visual.md`) replaces the length one.
   What the length tolerance let through (`within-tolerance.md`) is not yours to explain.

   **Every departure from the mockup or the approved prototype is blocking.** It takes the fix
   procedure (`skills/flow/fix.md`) per `skills/flow/visual-verify.md` — `departure` is never a
   closing state. A recorded decision is a named cause only for the deltas it names, by element and property: any other delta on the same band or element is unexplained until measured.

   **A departure's expected value cites the mockup's source wherever the handoff ships one, never a
   rule inferred from the pixels**: where the frame was drawn from HTML or CSS in the tree — beside
   the mockups or cited by `design.md` — the line names the declaration the expected layout, size or
   colour comes from (`grid-template-columns:1fr 1.3fr 1fr; height:32px`), and a value no
   declaration states is named as read from the pixels.
   A departure within 1dp is reported like any other, its measured value beside it: whether it
   stands is the parent's call under the Blocking paragraph of **Steps 3–13**
   (`skills/flow/visual-verify.md`), never yours.
   **Before a departure is raised, it is checked against the frame's sidecar comment and
   `design.md`'s `## Decisions`**: one either declares — a part named expected, a control a decision
   gives another form — is reported as declared, citing the declaration, never as a departure.
   A row the `frame:` line names under `departure:` is one the sidecar declares, its `cause` in the
   manifest; it is reported as declared, citing that line, and never judged again.

   **Enumerate every mockup frame the change touches into an explicit checklist before starting
   the sweep, and account for each one by name at handoff — done, or why not.** "Did a
   screen-by-screen sweep" is unfalsifiable without a fixed list: a partial pass reads exactly like
   a complete one if nothing forces naming what was skipped. Build the list from
   every frame id `design.md`/`tasks.md` cites for this change, or from the mockups directory
   itself when no such citation exists, plus every frame the prompt names under `moved:` — one a
   shared fix's recapture moved (`skills/flow/fix.md`); a frame reached only by inference from another (already
   covered by the composite diff, "shipped" and pre-existing, blocked by a real environment limit)
   still gets named, with the reason. The list also carries every frame the `<spec>.mockups` sidecar of each capture spec this change edits maps, whether or not its baseline has moved yet, and two frame ids are never read as one. **The list is the change's, never the round's**: a fix run
   re-captures the same list, not the views its fix touched, and the specs `capture` runs are
   chosen to reach every frame on it, never the other way round — a denominator the verifier
   derives from the specs it happened to run is the partial pass wearing a fraction. The report's
   `frames:` line names every declared frame absent from the per-frame lines, with its reason.
   In a scoped fix round, a frame that neither step 9's `changed:` line, the prompt's `moved:`, nor
   the prompt's `defects:` — the frames this round's defects name, always judged even when their
   capture did not change — names is not read again (step 9). The `frames:` line names
   it with the reason `capture unchanged since <round base>`.
   A `changed:` frame an earlier round reported clean is read against its round-base capture too,
   under **A verified frame is frozen** (`skills/flow/visual-verify.md`).

   Before a frame is called matching:

   1. **A cell the pixels cannot answer is measured from the mockup's source, never left
      unexplained**: where a row reads `unresolved` — the script exits 1 on a region no widening
      resolved — or a leaf reads `clipped` (the element runs off the image there), the frame's
      number is the value the mockup's source declares for that element — the CSS of the HTML
      the frame was drawn from (`box-shadow`, `border-radius`, `border`, a colour), found beside
      the mockups or cited by `design.md` — and the capture's is the script's where its own
      image measures, else the value the app's source draws the element with; a `differs:` line
      for that row names its source.
   2. **Read every flagged frame's composite before judging its rows.** **A row is never judged by
      rule**: each `over` row is a difference or not by what the composite shows at that row —
      never a range (`#3-#40: annotation`), a list of rows under one cause, or a selector (`rows 3
      to 40`, `all`, `the rest`). When `measure` printed `shift: dy=<px> rows=<n>
      frame=<frame id>`, `<n>` of the frame's rows sit one constant offset low or high: they
      are one defect, the manifest's `shift` names it, and its `row` stands for
      all of them on `over` — find what moved the whole frame (a bar height, a scroll position, a status
      line) rather than explaining rows. What moved it is itself a departure, listed on the
      report's departures lines unless a recorded decision covers it. The frame's rows are then paired under that offset, so every
      other `over` row is a departure of its own, read as if the shift were already fixed.
   3. **Judge every row the matrix's `over` list names, and write only the differences.** The
      manifest carries each composed frame's `over` rows by their 1-based index — every matrix row
      `missing`, `extra` or `unresolved`, or paired over the length or colour tolerance. A row is
      no difference only for one cause from a closed list: a frame annotation outside the app's
      own ink; a text difference the transcription names, or its text width; a recorded decision
      whose `**ID:**` `design.md` carries, for the deltas it names; or an `unresolved` row whose
      value, measured from the mockup's source per the cell rule above, matches. Every other row
      is a departure per **Every departure from the mockup or the approved prototype is blocking**
      above, and goes on a `- <frame id> differs:` line with its own row's figures as the
      manifest's `figures` carry them. A row no difference is not written. A `scale` key — the
      row's widest ink group grown or shrunk on both axes by one ratio — is the same glyphs set
      at another size, never a text width; a `segment<n>.colour` key is one mark of a rail in
      another colour, though the row's own colour matches.


**The motion step — record every motion the prompt names.** With `motions: none`, the report's `motion:`
line reads n/a and the step is skipped. Otherwise, for each motion: in a Chromium page with
reduced motion off (`page.emulateMedia({ reducedMotion: 'no-preference' })`) and the viewport and
clock step 8 pins, bring the view to the motion's start, then record frames from before its
trigger until the view settles — a CDP screencast (`Page.startScreencast`, each frame acknowledged
with `Page.screencastFrameAck`) — while driving the trigger (a tap, a navigation, a drag through `page.mouse` in small
steps). Never sample with `page.screenshot`: the Chromium page this step mandates always
carries the screencast, and sampling skips the frames it catches. **A motion that changes an element's height is recorded in a viewport tall enough to hold that element's settled height** — step 8's viewport made taller for that recording alone — since an element whose bottom edge leaves the viewport reads as no height change. Write the frames, in order, to `<changeRoot>/visual-verification/motion/<motion
id>/<nnn>-<ms>.png`, `<ms>` the frame's offset in milliseconds from the trigger — the screencast
frame's own timestamp — read every one, and
judge the strip frame by frame: **every frame shows the view
being left or the view being entered, never an empty, default or wrong-but-plausible rendering
between them; chrome (a title, a bar) and content change in the same frame; a removed motion's
strip shows no intermediate frame; a gesture's content follows the pointer in every frame and
settles where the motion's end state says** — plus any frame rule the project's own instructions
state. A frame that breaks one is a departure, named by its index and what it shows.

11. **Write `<changeRoot>/visual-verification.md`**, opening with the line `Pass: <this row's key>` (`close` refuses a record naming another pass), whenever any view was captured — one entry per view: its screenshot
    path, resolved by the same recursive search step 9 used, and what was seen; and, per composed
    pair, the composite's path in `<changeRoot>/visual-verification/`, the frame id, its `diff=`
    ratio, what was seen, its
    band and seam departures with their causes, and
    `<changeRoot>/visual-verification/matrix/<frame id>.json`, the frame's element × property
    matrix, cited in place of a table. **Every evidence file the record cites is in
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
    **A scoped pass and a `-rest` pass rewrite the whole record**, never only the frames they re-check: every
    view and pair of the change's list, each entry carrying its latest result under the key of the pass
    that produced it — a frame an earlier pass found departing and this pass re-checked done reads
    `done` under this pass's key, and a frame this pass did not re-check keeps its entry with the key
    of the pass that last checked it. The record the change lands with then never names a departure
    the closing pass closed.

**The wire-contract check — when the prompt names one.** With no `wire contract:` line, the
report's `wire contract:` line reads n/a. Otherwise, signed in as an account
`<project>/.flow/project.md`'s `## credentials` declares, call each named endpoint on the running
API at its worktree-resolved URL, read the response's shape and wire names against the line, then
drive the stated acceptance behaviour through the API and read the next response — a screenshot
comparison is no substitute for it. With a `wire contract:` line the dispatch names no view, so
steps 8, 9 and 10 read n/a — wire contract, no view drawn — never `done` over a capture that
never ran.

**No required step is left undone for want of time or context.** The report's `steps:` line gives
each of steps 4, 8–11 and the motion step a status the report check admits — `stats/internal/guard/verifyreport.go`'s header
is canonical for them — and the parent runs that check on the report before reading it.
Run it on your own report before handing back: an exit 1 names the steps still owed, so do them
and run it again, and hand back an incomplete report only for a blocker you cannot cure, named.
**The report carries one row per frame on the change's list, and you do not hand back while any frame lacks one** — naming a frame on the `frames:` line is never a row, and a remainder handed to a second verifier is a defect of this report.
Run the bundle's `close --check-only` line again the same way, for its frame-list half: a `no report line` frame gets its
own per-frame line or is named on the `frames:` line with its reason, and a `no sidecar line` gap
step 8 could not close stays, which `close` hands the parent.
**A departure found never shortens the pass:** every step still runs to its end in the same
dispatch, and a fix round's re-verify finishing what this one left is no reason to leave it.
**Step 10 is `done` only when every frame's matrix is**: every composite read and every
`over` row judged — one left for time, context or a measurement the pixels could not give is step 10 undone.
**The report lists only what differs.** Each composed frame has its own line, never one line
for a range of frames: `match` when it shows no difference, else one `- <frame id> differs:`
line per difference — element, what the mockup draws, what the capture draws, the numbers. No
sweep, spot-check, reading or per-element `n/a` line is written. The report check refuses a
`done` step 10 while a frame line reads `match` beside a `differs:` line, reads anything else
with none, names a frame no manifest pair holds, or leaves a manifest frame with neither its own
line nor a place on the `frames:` line. The spec, composite paths, `diff` ratios and frame counts
are the manifest's, never re-typed into the report.
The per-view `- <view>: <PNG>` lines are status lines: `clean` or the defect, never a narration
of what was read.

**Hand back the `## Verdict` alone.** Write the report file as a `## Verdict` block with the full
`## Report` under it, and end your turn with the `## Verdict` block alone — the parent reads nothing
else. `exits:` names every exit the `## Report` carries, `not declared` where a command is not;
`cause:` appears only when a step is `blocked`, naming `environment`, `test-failure` or
`missing-fixture` from the report's own facts; `departures:` counts every departure the `## Report`
names, and the lines under it name each cause once, with every frame or view it shows on and its
attribution beside it — one cause drawn on 34 frames is one line, never one line per frame
.

```text
## Verdict
- Model: <the model named in your own system prompt>
- exits: setup <n | not declared>, capture <n>, full app suite <n>
- steps: <the ## Report's steps: line, after `steps: `>
- cause: environment | test-failure | missing-fixture
- departures: <n>
- <every frame id or view it shows on>: <the departure's cause, with its attribution>
```

```text
## Report
- setup: <not declared | exit <n>>
- stack: <already running | started and stopped | could not be started — <output>>
- verify: green — <the inline verify row the bundle names>
- capture: exit <n>
  <output, verbatim or last 40 lines>
- full app suite: <created | updated> — <absolute spec path>, <n> screens, capture exit <n>
- full app suite zip: <absolute zip path>, <n> PNGs | not written — <reason>
- identical: <the PNG names of one byte-identical group> — <why the views draw the same pixels, or the defect> | none
- <view>: <absolute PNG path> — clean | <the defect seen> | capture unchanged since <round base>
- <frame id>: match | differs
- <frame id> differs: <element> — mockup <what it draws>, capture <what it draws>, <numbers>
- frames: <every declared frame id with no line above and why> | none absent
- motion: <named motions recorded>/<named> — <each motion id: its strip directory, frame count, and `clean` or each departing frame's index and what it shows> | n/a — no motions named
  - <a strip recorded beyond the named motions>
- wire contract: <method> <path> — <the shape and wire names read, and the acceptance behaviour seen, or each departure> | n/a — no wire contract named
- steps: 4 <status> | 8 <status> | 9 <status> | 10 <status> | 11 <status> | motion <status>
```
