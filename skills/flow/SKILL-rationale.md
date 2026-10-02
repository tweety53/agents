# flow — rationale

Reasoning behind `skills/flow/`'s phase files: why a rejected alternative was rejected, which
design.md decision a passage implements, and choices a task made on its own. Moved here verbatim
from the run-loaded files; **no `/flow` run loads this file.** Each heading names the source file
and the section the passage came from.

## SKILL.md — Stage keys

Every stage `/flow` marks uses a `flow.*` key, minted fresh for this command rather than reusing
`start.*`/`do.*`/`finish.*` — the old namespacing was tied to the three commands this one replaces,
and reusing it here would misdescribe a stage that no longer runs under the command its key names.
**This is a design decision this task makes, not one resolved upstream of it**: the alternative —
keeping the old keys, unioned, as `/flow-fast` did — was rejected because several stages here are
not the old stage unchanged (`flow.verify` merges two, `flow.landing-routes` absorbs two more,
`flow.kickoff` is new), so a reused key would sometimes mean something different than its history
records.

## SKILL.md — preamble

> one state file, the same three states, content reorganized by topic rather than by the old
> start/do/finish boundary (design.md's `flow-rename-content-split`). This file is the router: it

## SKILL.md — Model resolution

> An empty list can never reach this table from `/flow-settings`: `<agents repo>/stats/cmd/flow/settings.go`'s
> `settings set` refuses an empty `-reviewers` as a caller mistake before any write reaches the
> store. The empty-list row exists because this resolver must still define a value for a state the
> store's schema permits, not because an operator can produce one.

## SKILL.md — Guardrails

> **Generate this run's session token once, right here, before the first mark any phase file below
> makes — a short, unique literal string — and reuse that exact same value at every `stage begin` this
> run makes, including inside every phase file it dispatches into.** One run, one token, never a fresh
> one per mark or per phase file (design.md's "one token per session, not one per mark").

## brainstorm.md — preamble

> Superpowers Basic Workflow steps **#1** (brainstorming) and **#3** (writing-plans), intertwined
> with spectre artifact creation, run here — without an options-question round or a proposal
> publish (design.md's `ask-options-removed`, `publish-proposal-removed`).

## brainstorm.md — A. Resolve the change and write `STARTED`

> This is design.md's `started-redefined`: `STARTED` is a kickoff marker,
> "the operator started this," not a record that a design was approved and a proposal published.

> `planningEffort` and `models.default` are written `null` and stay `null` for the life of the
> change: `/flow` asks no planning-effort or model question on a creating run
> (`ask-options-removed`), and models are chosen per dispatch by the Decide step
> (`planner-chooses-models`), not recorded per change. `artifactUrl` stays `null` —
> `/flow` publishes no proposal artifact (`publish-proposal-removed`).

## integrate.md — preamble

> Run 1, with two folds design.md decides:
> `move-in-review-fold` (the Jira "move to In Review" step becomes a sub-step of `flow.landing-routes`
> rather than its own mark) and this task's own resolution of open question `write-in-progress-fold`
> (below).

## integrate.md — rebase

> — per design.md's `rebase-is-a-confirmed-choice`: the rebase never runs on its own, only after the
> operator picks this option, and only against the worktree(s) that actually moved

Superseded: the rebase now runs with no prompt on every `MOVED` worktree — **Sync the branch onto
the base** (`skills/flow/sync-onto-base.md`).

> , per design.md's `never-auto-abort`

## integrate.md — Scoped re-verification

> , per design.md's `scoped-reverify-not-full-suite`

## integrate.md — landing routes

> This stage carries three sub-steps under one mark, per this task's own resolution of open question
> `write-in-progress-fold` and design.md's `move-in-review-fold`: the git route, the state write, and
> the Jira transition — three sub-steps that used to be three separate top-level marks
> (`finish.landing-routes`, `finish.write-in-progress`, `finish.move-in-review`) now recorded as one.
> **This task's own choice**: fold `write-in-progress` in alongside `move-in-review`, rather than
> leave it standalone. The write is a genuine no-op (`IN_PROGRESS` → `IN_PROGRESS`, nothing changes
> but `prUrl` and `updatedAt`/`updatedBy`) exactly as design.md's open question describes, and by the
> time this mark reaches it, the route sub-step immediately above has already decided what `prUrl`
> becomes — folding the write in with the route that produces its one real input, and with the Jira
> transition that only ever follows a successful route, keeps one mark's three sub-steps in the causal
> order they already have to run in, rather than three marks whose middle one records nothing a reader
> could not already infer from the other two.

## brainstorm.md — the worktree is created inside `flow.kickoff`

kan-488 moved worktree creation to the end of planning (its `worktree-created-at-end-of-planning`
decision), so sections **B**, **C** and **D** wrote into the main checkout's own
`<project>/spectre/changes/<name>/` and the run moved that directory into the worktree afterwards. That
decision is superseded by an explicit operator instruction: `/flow`, `/flow-fast` and `/flow-plan`
runs kept touching whatever branch the main checkout sat on — planning artifacts left there when a
run stopped early, a `/flow-plan` commit on the checkout's current branch — and the operator asked
that every run "always create a worktree in the very beginning and operate only on worktrees, not
touching main/develop/regular branches." Creating the worktree inside `flow.kickoff`, right after
the `STARTED` write, is what makes that hold by construction: no later phase has a main-checkout
path to write to. The interim `.superpowers-sdd-decision.json` path and the two `mv` lines that
existed only to bridge the late creation went with it. **Considered:** writing planning output to a
scratch directory and moving it in later — rejected the same way kan-488 rejected it, and now with
no gap left for it to bridge.

## git-boundaries.md — Branch backup

Pushing the branch at creation and after every commit was asked for in the same instruction, so a
worktree lost with the machine, or removed by a stray cleanup, is rebuilt from `origin/<branch>`
rather than lost. The cost accepted: one push per commit, and integrate's push after the
reshape becomes `--force-with-lease`, on a branch only the run writes.

## integrate.md — 2. Ask how the branch should land, the rebase's planning-artifact aside

Incident behind `aside-planning-artifacts.sh`, cited there as a parenthetical: KAN-628.

## archive.md — 4. Commit the archive

Incident behind `check-archive-scope.sh`'s cannot-answer exit, cited there as a parenthetical:
KAN-601.

## archive.md — Worktree cleanup

Moved verbatim from the section's opening sentence, where it followed "**Worktree cleanup**
(`skills/flow-contracts/finish-contract-run2.md`), canonical for it": , and is not restated in full
here beyond one override.

## archive.md — Guardrails

Moved verbatim, the bullet addressed to the file's editor rather than to a run:

- **Never** state a cleanup rule here — **Temporary artifacts registry** (`skills/flow-contracts/artifacts-registry.md`)
  and **Worktree cleanup** (`skills/flow-contracts/finish-contract-run2.md`) are canonical.


## Moved by the 2026-09-22 prompt audit — incident records and recorded reasons

Each passage below sat inline in a run-loaded phase file beside the rule it motivated; the rule and its
in-sentence reason stay there. The quoted fragment before each dash is that rule, for attribution; the
text after it is the passage, verbatim.

### brainstorm-planner.md — The checklist

- ***A frame-specified design needs its handoff assets reachable from the tree.*** — gymie kan-29's routes A–F were built with the handoffs outside the tree, which is what left every "different from the mockup" report unanswerable and let an unchecked caption survive review.
- ***The seeded-note path is legitimate, never a bypass to prevent.*** — gymie kan-468's run converged its plan this way; this paragraph preserves the path.
- ***The note is the research, never the plan's form.*** — gymie kan-485's run corrected every `Files:` field by hand before task 1's guard ran clean, and gymie kan-468's fixed the seeded H1 at load-context; this requirement is both corrections, applied where the note is consumed.

### brainstorm-planner.md — Decisions

- ***ID** is assigned once, at creation, and is **immutable** — the match key a later round uses to **supersede** a decision.* — The same key is what let gymie kan-459's review panel name the specific decision a finding found stale instead of re-arguing the design.

### brainstorm-planner.md — D. Writing plans

- ***Write a verification change so its found defects become their own tasks.*** — (the gymie KAN-29/gymie KAN-30 precedent)
- ***A task premise about a guard's behaviour is run at plan time, never assumed.*** — kan-542's task 2 shows the cost of leaving the premise to review: planned on a stale premise, it survived until a reviewer struck it with the guard's own exit — a defect this duty kills at planning.
- ***A task cites the decision it implements, never restates it.** Restated decision prose drifts from its entry the first time either is edited* — which is why gymie kan-468's seeded plan cited instead
- *`**Baseline:**` … record a command whose stdout is one integer* — (the `| grep -c` pipelines in kan-271 and kan-298's plans are the shape)
- *`**Files:**` … is expanded by the writer, never left in the field for the guard to expand* — (KAN-636); kan-579's plan declared gs/core/... behind its own legend and hand-repaired every field back to literal paths at its first task boundary, which is the repair this rule removes.
- *A block of `**Decision:**` lines sits between blank lines … never glued to the `**Commit:**` subject above it* — (KAN-636): a citation with no blank line rode in the subject's continuation in every reader that had not learned the field's name, and kan-579's run repaired each such field by hand at its task boundary.

### brainstorm-planner.md — D. Writing plans (plan guards)

- *The refusal exits are never repaired by editing the plan* — (KAN-601)
- *…record its size as the change's first task-count observation — the planned figure every later fix round's growth is measured against* — (KAN-415)

### brainstorm-planner.md — D. Writing plans (live-verification task)

- ***Write a live-verification task when the change touches a running service or persistent state.** … the justification is required, the task is not.* — This is deliberately not an integration-test stage on every change: a step that usually resolves to "nothing to do" trains everyone to skip it.

### verify-and-handoff.md — Inline verify

- *…the baseline is the project's own statement that the failure exists on an unmodified tree* — (KAN-547)
- *`environment` where the environment itself failed twice* — KAN-510

### visual-verify.md — The verifier dispatch

- ***A final report that is a block closes its dispatch `-outcome blocked -cause <cause>`, never `completed`.*** — The cause is what makes three environment-caused blocks in one run a query the store answers instead of a footnote in one run's ledger (KAN-510).

### verify-and-handoff.md — Resolve the run instructions

- ***The stack behind the URLs is checked, not trusted.** … the project's declared `fingerprint` row … proves what the stack serves is the worktree's own build* — the check KAN-334's handoff lacked when it printed a URL backed by a bundle eleven commits stale.

### visual-verify.md — Visual verification, step 3 (pre-flight)

- *A verifier re-dispatch cannot repair an environment that cannot pass* — gymie KAN-459's visual-verify stage retried three times over roughly five hours on pre-existing workspace-isolation gaps before the operator stopped it.

### visual-verify.md — Visual verification, step 6 (fingerprint)

- *…a stack step 5 found already running may be serving a build older than the worktree* — gymie KAN-29's last fix round captured, and nearly accepted, the bug the fix had removed.
- *Steps 1, 2, 3, 5, 6, 12 and 13 are the parent's* — gymie KAN-845: the verifier's step-6 restart was refused by the auto-mode permission classifier, so the dispatch closed `blocked` on a stale bundle one approved restart fixed, and integrate stopped on `VISUAL-VERIFY-MISSING`; the operator's own `devRestart` then dropped `-PisolateOAuthPorts` and rebound the stack to the shared 3000/8080.

### visual-verify.md — Visual verification, step 10 (sidecar)

- ***No sidecar is never a silent skip.*** — A mockups directory sitting unused is what let gymie kan-30's own screens ship four fix rounds deep with their real, drawn frames never once diffed against the app — `mockups: no map` was reported and accepted every round, because nothing required the sidecar that triggers the compose step to exist.

### visual-verify.md — Visual verification, step 11 (record)

- *No entry cites the worktree-absolute path of a file the worktree holds alone* — gymie KAN-29's record cited two screenshots that way, they died with `worktree remove --force` at archive, and the record kept two dead citations.

### visual-verify.md — Visual verification, step 12 (commit)

- *…never a bare `git commit`: the index of a main checkout may carry a pre-staged foreign tree* — and gymie kan-469's plain commit swept ~130 such files into a baselines commit

### verify-and-handoff.md — Resolve the run instructions (start the stack)

- *…if the applications those instructions name are still serving pre-fix code, the operator reviews one thing and runs another* — measured, not hypothetical, in this repository's own `<project>/stats/internal/web/embed.go`, whose `//go:embed all:dist` makes a running daemon blind to an SPA source change until it is rebuilt.

### visual-verify.md — Visual verification (Blocking)

- ***a defect the verifier reports in a captured screenshot — even when every assertion passed.*** — That last one is the whole point of this stage: three defects have shipped invisible to a diff, a five-pass review panel and both test suites, and obvious the moment the page was opened.

### visual-verify.md — Visual verification, step 8 (capture)

- *A `toHaveScreenshot` passing over a baseline this change wrote is the app agreeing with itself, never with the mockup — step 10 is the only comparison, and `capture: exit 0` is never evidence of a frame's fidelity* — (gymie KAN-437 final verification: Q1's baseline was the implementer's own first capture, green on every later run while drawing three controls of the wrong kind)
- *…d, goal or marker value outside the plotted range; a dataset whose derived numbers (axis ticks, averages, deltas, unit conversions) do not come out round; a list longer than the viewport; and the empty or first-time entry path beside the populated one* — (gymie KAN-437: a goal line drawn from a value outside the axis range rendered over the list below the chart, and axis labels read `82.333333333333 kg`; the spec's fixture kept the goal in range and its ticks round, so 22 frames passed and an operator found both by hand)
- *Every date the fixture writes — a workout's day, a weigh-in, a target's `from` — derives from that same instant in the app's own timezone, never from `new Date()` and never from its UTC day where the app renders local* — (gymie KAN-339: the baselines embedded `THU 27 AUGUST` and expired at the next midnight, and the fixture's UTC date disagreed with the day the app drew, each costing a re-capture round)

### visual-verify.md — Visual verification, step 10 (band and seam pairing)

- *…arture, so the first read of every pair is the script's band pairing, never the panel.** The panel is white wherever any channel differs, and on a real pair 20–50% of it is white from seed data, font rasterisation and the frame's own annotations alone* — (gymie KAN-437 final verification: every composed frame read `diff=0.20`–`0.52`)
- *The pairing resyncs after an unpaired band; a pair count far below the frame's band count is itself the finding that the layout differs wholesale, and the composite is then read to say how* — (gymie KAN-437 fix round 5, Q2 pre-fix: the frame's four "HOW IT GOT THERE" hairlines listed `missing`, the "Adjust first" band paired with `edge #0088b0` against `#d7d3d3` — the outlined-accent variant shipped as the grey one — and the answer rows' `since_pair` short by the padding the row lacked; Q1 pre-fix: the ACTIVITY card and the GOAL segmented control each `missing` as the frame's bordered band and `extra` as the capture's unbordered one)
- *…vertical structure inside every boxed band.** A segmented control with its two cell dividers gone and its wrapped labels left-anchored where the frame centres them is one band in both images, the same height and the same border colour, and pairs clean* — (gymie KAN-437 fix round 5, Q1's GOAL control: both shipped past the band pairing and every sweep, and an operator found them by eye)

### visual-verify.md — A missed defect — the tooling analysis

- *A miss means the sweeps cannot see that defect's class — the property, element kind or state no sweep read — and fixing the one instance leaves every other defect of the class in place; the analyst writes sweeps for the class, and the fix run's verifier runs them on every touched view* — (gymie KAN-437, KAN-550: round 1 missed nine defects; a dispatched tooling analysis produced four new sweeps, and the re-run with them immediately caught two further real defects — the GOAL dashed line struck through its label, and the calories-tab scroll)

### visual-verify.md — Visual verification, step 10 (control-level sweeps)

- *Full-page comparison catches wrong regions and wrong overall layout, and wrong text only where the capture's data is the frame's own; it does not catch a border style, an icon's glyph, or a colour step, all of which are invisible at full-page scale* — (gymie KAN-30 fix round 6: a field's underline-only focus border, drawn against a mockup showing a full outline, read as a match at composite scale and was found only once the two were cropped and zoomed side by side)
- *…a list row carrying a leading radio, or a quiet row whose only indicator is a trailing filled circle on the selected one; a segmented control of N cells each fitting its own label — a differing kind is a departure before any number is taken* — (gymie KAN-437 final verification, Q1: the WEIGHT field rendered in the weight tab's headline-figure style at about three times the frame's field height; every ACTIVITY row carried a leading outlined radio where the frame draws no indicator on unselected rows and a trailing filled circle on the selected one; the GOAL segments' two-line labels ran into the neighbouring cell — none subtle at 1x, none seen, the frame never on the list below)
- *…fore starting the sweep, and account for each one by name at handoff — done, or why not.** "Did a screen-by-screen sweep" is unfalsifiable without a fixed list: a partial pass reads exactly like a complete one if nothing forces naming what was skipped* — (gymie KAN-30 fix round 7: two prior rounds each reported partial coverage as if it were the whole set, and the gap was caught only when the operator asked "so is X fully compared and fixed?" of a specific frame)
- *…gainst that list and names every declared frame absent from the per-frame lines, with its reason; the parent reconciles it against `design.md`'s own frame list before applying **Blocking**, and a declared frame with no line blocks as a departure would* — (gymie KAN-437 final verification: `design.md` declared 22 frames, two `flow.visual-verify` rounds each ran three of five fidelity specs and reported 15/17 composed — Q1–Q3 and M1–M2 never composed, never named as skipped, unnoticed by the parent, and every Q1 defect above surfaced only when the operator opened the screen)
- *…s below and capture each one; a state nobody drove into is a state nobody verified.** Each is a state a resting capture and a full-page composite cannot show, so each is reached deliberately rather than by whatever the walkthrough happens to pass through* — (gymie KAN-30 fix rounds 7 and 8: a blind click landed on a day's "Repeat that workout" suggestion instead of "Create a group session" because an extra card existed on that day only, and the sweep almost recorded the wrong screen as verified; one dialog's "selected" fill differed between two sub-states, two sibling cards diverged in shadow, and "Add user" carried a border on one code path and none on another because the empty and populated roster states routed to two different shared button components; a time wheel scrolled correctly one way only, a state-derivation bug no resting capture shows; an empty picker shipped in its pre-restyle appearance beneath a restyled populated sibling; a picker offered an option submit always rejected, surfaced only as a post-hoc toast)
- *…e shows and state, per run, that it is a display value at the precision the frame shows — a raw float, a `NaN`, `null`, `undefined`, an empty string where the frame draws a value, a placeholder is a defect whatever the frame's own numbers are* — (gymie KAN-437: `82.333333333333 kg` on an axis the frame drew as `82.3 kg`, read as a data difference)
- *…ading, chart, chips, list, link — and compare their order: a control present in both but in a different position relative to its siblings is a departure the per-control crop never sees, since each crop matches its own control wherever it sits* — (gymie KAN-437: preset chips rendered above the chart the frame drew them below)
- *Then confirm no element's ink crosses its container's bounds into a sibling — a plotted line, a marker, a label — driving the out-of-range value step 8 seeded and stating whether the app clamps it or hides it* — (gymie KAN-437: a goal line drawn at a Y past the chart's clip, over the rows beneath)
- *A screen whose content extends past the viewport with nothing to scroll it is a defect, and the fixed-viewport capture cannot show it — it proves the visible viewport and nothing beyond* — (gymie KAN-437: a tab with no scroll container at all, its "All N weigh-ins" link structurally unreachable, every capture of its top green)
- *A value that does not follow its input, or that is absent on one entry path the populated path shows (first-time versus returning), is a defect no single capture can show* — (gymie KAN-437: a dialog's derived field stayed stale as the operator typed, and was missing entirely on the first-entry path the spec never took)
- *A row whose every control measures the right size is the case to suspect, not to pass: a uniform container padding around correctly-sized controls is invisible to every reading but this one* — (gymie KAN-437 final verification: a Weight/Calories segmented control and its "+" button, both exactly the mockup's 44dp, sat in a `padding(space4)` on all four sides — 20dp against a frame drawing 2–5px to the dividers above and below — and passed two `flow.visual-verify` fix rounds, one operator sweep and one manual re-check, each of which had measured the controls and none the gaps)
- *This sweep runs on a frame whose content already matches and on a frame already fixed for something else — a fix run re-verifies the whole frame, never the element it fixed* — (gymie KAN-437 final verification, Q2 and Q3: Q2's "HOW IT GOT THERE" rows had their label text fixed and were re-checked for that text only, and Q3 was passed as matching on its labels and values; the hairline rules both frames draw above, between and below their rows were never in the app, and an operator found both by hand)
- *The report's `sweeps:` line then names the sweeps, not the numbers, so a round that re-checked one defect reads exactly like a round that measured the frame* — (gymie KAN-437 final verification, Q1 and Q2, one operator pass after `12a64fa`: Q1's ACTIVITY rows lacked their grey borders, its GOAL segments' labels sat off-centre in their cells, and its fat/carbs slider drew another track and thumb; Q2's headline figure was the wrong size, its PROTEIN/CARBS/FAT row the wrong colour and inset where the frame runs edge to edge, its "HOW IT GOT THERE" caption the wrong colour and its rows spaced differently from the dividers — seven departures across colour, size, width, alignment, border and control kind, on two frames already fixed and re-verified more than once, every one of them a property the script measures and none of them a cell any sweep asked for)

### visual-verify.md — Visual verification, step 10 (measurement preconditions)

- *…sized or cropped image; measure it with `measure-visual-properties.sh`, then eyeball what it measured.** A crop is for reading text and layout, never edges or centres — interpolation and a small viewport shift an edge by pixels and hide a gap outright* — (gymie KAN-30 fix round 9: a card corner read "square, no gap" from a tight crop that ended before the corner; a row-by-row background-colour scan of the same boundary found a rounded corner and a 9px gap — done by hand-written one-off scripts, three attempts, the first two wrong)
- *…it lists is an edge the region can sit on, and a miscrop is the glance again; and `--edge` lowered below the distance `runs` reports between a control's fill and its background where that distance is under the default 24, since a surface-on-page card* — (gymie: `#eae9e9` on `#f3f2f2`, 15.6 apart)
- *Two readings the eye reliably gets wrong: an icon's tint is `content.colour`, never `fill` — the fill is the box behind the glyph, and a grey glyph on the right fill matches on every other property* — (gymie KAN-437: two icons grey where the frame drew accent blue)
- *…is the box behind the glyph, and a grey glyph on the right fill matches on every other property; and a pill is a `radius` equal to half the box height, while a rounded rectangle is any smaller number — read the number, since both look "rounded" at 1x* — (gymie KAN-437: preset chips shipped as rounded rectangles against a pill frame)
- *…inding, a code-level guarantee, a passing test — as that element's box in each image via `--ref-a`/`--ref-b`, never from the nearest similar-looking thing: an unchecked ruler is itself a claim, and a wrong one makes the comparison wrong twice* — (gymie KAN-30 fix round 9: a day-number circle as ruler put a button at 2–3x oversized; an adjacent "+" button confirmed correct earlier put it at 15–20%, traced to a deliberate 44dp touch-target minimum)
- *The box between two landmarks encloses different content in different states, and its size then compares nothing* — (gymie KAN-30 fix round 9: a date-header row measured 3x taller against a mockup drawn with the calendar collapsed and a capture with it expanded)
- *…ather than a percentage, an order of magnitude — as a methodology error, never as a bigger finding.** Re-derive the calibration and the state check before reporting it; a smaller wrong number from the same mistake reads as a finding and ships* — (gymie KAN-30 fix round 9: the 3x row was caught only because it was absurd)
- *An accessibility minimum applies to the tap target only, while the mockup draws the visible shape, so a control that looks oversized is measured twice — hit box and ink — before either number is called wrong* — (gymie KAN-30 manual re-sweep, after round 10: the "+" button precondition 1 had accepted as a deliberate 44dp minimum had grown past 44dp on a child's layout demand, found only by measuring its rendered pixels; and a tap-target circle whose visible fill shared its hit size rendered about 2.5x the mockup's circle, the two numbers never having been separated)
- *…ing — zoom and contrast-check the capture before it becomes a code change.** A small, low-contrast but pixel-correct element reads as wrong at a glance, and "hard to see" is a different question from "drawn wrong" — one for the operator, not for a fix* — (gymie KAN-30 manual re-sweep: a substring highlight read as a smudge, cost three wrong hypotheses about its geometry, and was proved pixel-correct only once re-rendered in a saturated red for a contrast test; the right first move was a zoomed crop and "low-contrast, reads as a smudge at 1x — accepted?" handed back)
- *An operator's re-raised or repeated spacing complaint is a measurement order, not a second look — the second look is what already failed* — (gymie KAN-30 manual re-sweep: a date-row header's "+" button was screenshotted and eyeballed as "compact, matching the mockup's proportions" several times, disputed twice by the operator, and pixel-sampled only on the third complaint — 26px above the control to the divider, 6px below, a `LazyColumn` content padding stacked on the sticky header's top inset with nothing equivalent at its bottom)
- *…alone — a `padding()` placed before the `background()` or `clip()` it was meant to inset the content of, or on the parent before its children's own backgrounds, shrinks the painted area, not the content — a checkable line of code before any screenshot* — (gymie KAN-30 manual re-sweep: a workout picker's selected-row fill stopped a `space2` gutter short of its box's border on every side, the parent `Column`'s padding sitting above the rows' backgrounds; the complaint "rows are still not filled with color fully till the borders" was answered twice with the wrong measurement — a centring check on an unrelated icon, then a colour-existence check with the gap tooling — and once with the right one, a scanline through the fill and both border strokes)
- *…read the JSON's `delta` block: `abs` and `pct` per numeric property, RGB distance per colour* — (gymie KAN-30 fix round 10)
- *A number no eye confirmed is a methodology error waiting to ship* — ; an eye with no number is round 9 again (gymie KAN-30 fix round 9, above)


## Moved by the 2026-09-22 prompt audit (pass A) — review-panel.md and implement.md

*Incident narratives and archived design.md decision pointers moved verbatim from `skills/flow/review-panel.md` and `skills/flow/implement.md`; each sub-heading names the section the passage came from, and the rule it decorated stays in the run-loaded file.*

### review-panel.md — preamble

> Per design.md's `roster-from-settings`, which supersedes
> `review-panel-fixed-3` (design.md's `supersedes-review-panel-fixed-3`): there is no fixed roster
> table and no diff-size/touched-area trigger table.

### review-panel.md — The roster

> `ValidReviewers` in `<agents repo>/stats/internal/store/settings.go` is the id vocabulary this table exhausts —
> six entries, never a seventh.

> There is no parent-model inheritance and no economy tier.

> and both `model` and `-effort` are recorded,
> per design.md's `agent-definitions-universal-handshake`.

### review-panel.md — Experimental slot

> The rendered panel record's Slot column therefore shows the `exp-` id unchanged, so the
> prefix survives into the archive per design.md's `exp-slot-prefix`.

> `flow record pass -round <round> -note 'experimental: skipped — bundle cap'` (design.md's
> `exp-skipped-over-cap`) rather than displacing a persistent role.

### review-panel.md — the diff-size cap

> and the over-cap report names the sum and each
> worktree's own count (design.md's `cap-sum-across-worktrees`).

### review-panel.md — The docs-only reduction

> Per design.md's `docs-only-reduces-to-primary` (narrowing `roster-from-settings`):

### review-panel.md — final-review.diff

> it is a plain working-tree diff
> against the merge base, so work still unstaged or uncommitted is already in it — gymie KAN-459's F37
> was a false positive from reading it otherwise.

> one pass reads every worktree's section, so a seam between two repositories is in one pass's view
> (design.md's `combined-diff-per-round`).

### review-panel.md — Bundled dispatch

> **At most two review dispatches per round, each carrying one to three roles**, on
> decided and `default` panels alike and in both execution modes (design.md's
> `two-dispatch-cap-everywhere`).

> A one-role dispatch is unchanged from today.

> the
> panel record names every worktree's sha beside the delta path (design.md's
> `diff-base-canonical-sha`).

> **`-agent-id` is never typed, never invented** — the daemon captures the launch identifier
> (KAN-322), pairing each launch with the begin whose command sits nearest it in the transcript;

### review-panel.md — the reproducer rule

> so a command that exits 0 because its own
> diagnostic succeeded — the inverted convention every kan-512 round-0 reproducer was authored with,
> and had to be hand-corrected out of — reads as the defect already gone.

> in the grep-output shape
> the author pastes straight off the defect-present tree (KAN-606).

> and the guard rejects the exemption
> at Important (KAN-503).

> and a finding closes on that target's evidence, never on the test target's clean
> exit alone (kan-551, from gymie kan-437: a calories-tab scroll failure unproducible on the desktop
> target closed only at true scroll end in the browser build).

### review-panel.md — MUTATION ENTRY CONTEXT

> every test the plan's `**Tests:**` fields name and every test file in the touched-files list inlined
> directly beneath it (kan-574's round-0 mutation dispatch read ~16.9M cached tokens — four times
> the primary slot — because the brief's search for covering tests was named nowhere):

### review-panel.md — No forking, and a wall-clock ceiling on every slot

> **The ceiling is a record-and-review bound, not a stop** (KAN-612)

### review-panel.md — The mutation-testing brief

> the declaration
> `run-reproducer.sh` reads as the mutation convention (gymie KAN-568)

### review-panel.md — The throwaway worktree

> Dispatching a mutating slot into the same worktree
> a reading slot concurrently reads is the KAN-366
> collision

> **Independent multi-slot detection is the
> panel's core signal, preserved deliberately rather than treated as incidental (KAN-503)**

> The scaffold lines after
> the loop exist because the slot dispatched into the copy resolves the bundle paths its prompt names
> — `dispatch-context.md`, and the report file it writes — against its dispatched root: a copy
> without `<abs-worktree>/.superpowers/sdd` sent the panel back to re-brief the slot mid-dispatch and let
> the slot's report resolve onto the copy, where its removal destroyed it (KAN-529).

> Dispatch each slot present in this round's roster **once**, its prompt listing every copy made for
> that slot as the repository paths to mutate and test in, in place of `<worktree>` (design.md's
> `bugbot-security-one-dispatch`).

> because a report the slot resolved onto its dispatched root
> dies with the copy, and KAN-529's round-2 mutation report survived only because the parent session
> had read it earlier.

> every
> reproducer still runs against the real `<worktree>` at verification time, never against any slot's
> throwaway copy, exactly as today.

### review-panel.md — Panel re-runs

> pushed
> plain like any other commit, and every downstream commit keeps its sha — folding instead via `git
> commit --fixup=<task-sha>` + `git rebase --autosquash` rewrote tasks 7–10's shas in gymie kan-469's
> `gymie-frontend` run and forced a `git push --force-with-lease` re-sync with the remote.

> where an `origin/$BASE` upstream would
> carry them straight into the round's delta, as kan-535's first fix round found when its delta
> swallowed two upstream commits and had to be regenerated by hand.

> reserved for catching independent issues, which is what caught that change's
> round-6 real bugs.

The antecedent of "that change", recovered from the introducing commit (`9c60d2f`): KAN-500, whose fix rounds re-read a growing fix diff at roughly forty-five minutes a round; the sentence naming it was removed in a later edit.

> `check-panel-docs-only.sh <worktree>
> <merge-base>`, per design.md's `fix-rounds-reclassify`.

> **The exit-code contract is checked mechanically before any dispatch decision reads a reproducer by
> hand** (KAN-554 — gymie kan-468's panel supplied an Important finding's reproducer whose exit-code
> condition was inverted, and the inversion reached the deferred self-review pass before anything ran
> it):

> Before anything runs, the guard audits the
> instrument itself (KAN-606)

> **The parent runs these re-runs itself, in its own
> Bash calls — never a "verify fixes" reader or any other subagent**

> **The rerun cap.** The whole-roster re-reads this mechanism adds are capped at two unasked — the
> first final pass and its one repeat.

Before this cap the `full` policy's final whole-branch pass repeated after every clean fix round it
had opened, unbounded. Gymie's `kan-580-step-1-frontend-copy-session-to-another-day` run paid for
five passes — rounds 0, 2, 4, 6 and 8 — at 5.2M–7.9M cache reads apiece (KAN-662's ledger
figures), four Important findings spread across the four repeats. The cap keeps the reads that
earned their keep automatic: the round-2 pass caught F10, a double-copy race, and the round-4 pass
caught F14, missing disabled-button chrome — both real, both introduced or exposed by the fix
round before them, both invisible to a scoped re-run that reads only the fix diff and the sites of
earlier findings. Rounds 6 and 8 each caught one more Important finding; that catch rate against
that price is the trade the cap makes explicit, and the operator's single extension prompt is the
escape hatch over it.

The shape was settled with the operator rather than prototyped around them: an earlier attempt
(commit cc9167b, reverted in full by c66f806) also allowed one unasked repeat but put a prompt at
every later clean round whose silent default ran the pass — a toll booth rather than a bound,
since an unattended run still paid forever. The settled cap inverts the default: the unasked
budget is the first final pass and its one repeat — three whole-roster reads counting pass 1 —
silence closes the panel, and running beyond the cap is an explicit choice, put once per run.

**Rejected — a hard stop at the cap, no prompt.** A silent close is the over-cap failure
`check-panel-diff-size.sh` refuses for the same reason: a decision the operator did not see. The
prompt is the bound; the operator-prompts contract's 🤖 marker in the handoff is what shows the
silent default fired.

**Superseded in part — the prompt is no longer asked.** The operator later asked that every
prompt with a recommended option in an implementation or fix run be taken rather than asked
(**Auto-resolution**, `skills/flow-contracts/operator-prompts.md`). The cap's close is now taken
automatically; this rejection's reason is still met, because the close is not silent — it is
recorded as a pass note and carried by the `rerun cap:` field's 🤖 marker in the handoff — and a
third whole-branch pass still needs the operator's explicit instruction.

**Rejected — a cheaper pair on the unasked repeats.** The rerun pair runs at `low` effort;
the pass would keep its scope and lose its eyes. F10 is a race; a low-effort whole-branch
read is the kind of read that misses one, and a pass that reads everything badly is a worse bargain
than the prompt. Only the operator-chosen third pass, beyond the cap, runs on the budget rules'
demoted pair.

**Rejected — repeat only when the fix touched paths no full pass has read.** Every full pass reads
the whole diff, so after the first one no path is unread and the trigger never fires; F14 sat in a
file the round-2 pass had already read clean. This is a delete of the policy wearing a condition.

**Superseded (2026-09-27) — Rerun policy `full`, the rerun cap and its demotion rule are
retired.** The operator's instruction, verbatim: "get rid of FULL panel rereview. Do the diff
review there too. Maybe neighboring code at most." `delta` is now every class's policy, and no
pass after pass 1 re-reads the whole branch — each reads its fix diff, the sites of the findings
it re-reviews, and at most the code neighbouring those hunks (the FIX-ROUND SCOPE paragraph of **Panel re-runs**, `skills/flow/review-panel.md`). The `Panel:` line's `demoted:` and `rerun cap:` fields went with
the mechanism they reported. The scoped-round rule's closing whole-branch pass went too, since
that pass was **Rerun policy `full`**'s. **Rejected — keeping the whole-branch pass.** On KAN-778
it re-read about 24k lines of branch diff and found only Minors; the catches recorded above
(F10, F14) are the cost this trade accepts.

*Minors are fixed inline, never deferred (KAN-862).* The Minor-deferral default — a Minor-only
round deferred its Minors to `<project>/KNOWN-BUGS.md`, because a fix round cost more than they were worth
— had left 96 entries by `0b7e1c58`, 30 of them from kan-860 alone. Fixing them later took whole
`/flow` runs: KAN-861 planned 17 tasks at opus/high over 67 of them. What made a fix round
expensive was its subagent, reproducers, mutation proof and slot re-runs, never the edits. The
parent's inline fix drops all four, so a Minor now costs about what writing its
`<project>/KNOWN-BUGS.md` entry did.

### review-panel.md — The fix round mutation-proves what it changed

> The
> `fix-mutation:` line for that behaviour carries the measured pre/post observable as its third
> field, `<pre>→<post> <what the observable counts>` in the shape kan-534's census recorded
> (`2→0 orphaned temp lists`), never a bare test name (KAN-587).

> **The parent checks the reported list against the fix diff before the round can close, reading the
> `fix-mutation:` lines and walking the diff itself — never a "mutation re-verify" subagent or any
> other delegate.**

> rather than the regression being discovered next round (KAN-496).

> Exit 2 — the guard's not-a-verdict close (KAN-601, the same course
> `skills/flow/implement.md`'s task-close step gives it) — stops the run

> The stale-field classes the walk used to judge
> alone are the same guard's verdicts now (KAN-511):

> exit 2 — the same not-a-verdict
> close — stops the run (KAN-601).

> keep the round open until every tag reads `green` or `red` (KAN-538).

### review-panel.md — PLAN FIELDS

> never left for a reviewer to
> catch next round (gymie kan-454, gymie KAN-459).

### review-panel.md — the fix step

> A round may not chunk freely: its chunk count is bounded by
> `ceil(findings raised in earlier rounds / 10)` — a bound, not a target, so a well-formed
> per-finding sequence stays caught.

> A
> re-run that finds a fix incomplete — kan-512's round 1 catching task 10's own correction as F4,
> fixed as task 11 — opens the next fix round under the rules above

### implement.md — Dispatch sites — the parent's closed list

> Not to a "verify" reader, a
> "re-verify" or "mutation re-verify" agent, a helper, a background task, or a subagent under any
> other name — the KAN-449 run's six unrecorded subagents (four rogue panel-fix dispatches, a
> "verify fixes" reader and a "mutation re-verify" agent) are exactly the shape this forbids.

### implement.md — 1. Load context and validate the plan

> a concurrent merge can
> outdate both while this change waits (gymie kan-579: KAN-527 merged a range-read endpoint at the
> exact route the plan meant to add, and the collision surfaced only when task 1's spec-delta step
> ran against the current spec — an operator ask and a mid-run redesign of two tasks).

### implement.md — 3. Documenting a fix, before implementing it

> an append past this budget is how a change outgrows its own proposal
> without anyone deciding it should (gymie KAN-29 appended 24 of its 46 tasks this way).

> A test whose own name
> reads as a description of the reported bug is a signal to pause on, not reassurance (gymie KAN-30
> manual re-sweep: `backClosesTheFloorAndKeepsTheSessionRunning` asserted exactly the navigation
> the operator reported as wrong; the decision it was assumed to encode,
> `floor-bar-exits-only-on-contents`, decided only which bar draws the chevron, not where it leads).

> The glance that
> passed the control is what the operator is contesting, and repeating it answers nothing (gymie KAN-30
> manual re-sweep: a "+" button's row was re-eyeballed as matching through two disputes and
> measured only on the third — 26px above, 6px below).

> so the measurement answers the property
> named, never the screen area the complaint happens to sit in (gymie KAN-30 manual re-sweep: "rows
> are still not filled with color fully till the borders" was read as "look at that picker
> again" and cost a centring check and a colour check before the fill's edge was measured).

### implement.md — 4. Execute (SDD + TDD)

> UI fixes
> routinely touch shared files — icon sets, shared components, menu wiring — neither task named
> (gymie KAN-30 manual re-sweep: two "small fix" agents dispatched together at one worktree, with no
> Gradle overlap, collided in the tree — the second found the first's uncommitted, compile-broken
> WIP in an unrelated file and silently patched over it to unblock its own build, and neither
> agent nor the parent noticed until both reports named the same file).

> the daemon now captures the agent's identifier automatically (KAN-322)

> The start and end instants are the daemon's own (KAN-324) — never
> caller inputs.

> The sixth argument scopes the group's `## tasks.md` section to the plan header and the named
> tasks' blocks (per design.md's `scope-tasks-not-files`);

> **A guard you could not run is hand-substituted only on the record (KAN-417).**

### implement.md — FLOW — COMMIT-PER-TASK

> a plain commit sweeps a pre-staged foreign tree in with the
> task's work (gymie kan-469's sweep took ~130 files).

> because a `:(exclude)` governs what an `add` adds and cannot retract what an earlier
> step already staged (gymie kan-468 lost task commits twice to exactly this ordering):

### implement.md — REPORT, DON'T DECIDE

> **The gymie KAN-29 self review credited its implementers for exactly this behaviour — a backdating
> seam, a declined re-litigation of a recorded rule and an own-card asymmetry each reached the
> operator as a recorded decision because the implementer stopped to report instead of building —
> so every implementer dispatch also carries:**

### implement.md — The next implementer overlaps the guard

> the fourth is the empty placeholder that skips the parent-sha — the guard derives
> the commit's parent itself (KAN-330), so the argument a mistyped merge base once corrupted is
> never typed at all:

> The guard reads git objects and `tasks.md` only, so it is safe while the tree changes, and
> never stashes, reverts or resets (KAN-442).

> **exit 2 — the guard's not-a-verdict close (an
> unreadable plan, a task it cannot resolve, a commit range git cannot resolve, a usage error) —
> stops the run** (KAN-601)

> never re-run the
> guard on top of it. (gymie KAN-423: a re-run over a mid-flight revert cost ~55 minutes of hand
> recovery.)

### implement.md — The review gate

> WHY 40: gymie KAN-29's self-review (the source of this gate, KAN-400) measured its
> per-task reviewer rows on trivial tasks returning clean with sub-1k-token output — a review of a
> commit small enough to hold in one glance added nothing the guard and the whole-branch panel did
> not already cover, and roughly a third of that run's ninety dispatches were of that shape. Forty
> changed lines is the boundary below which a diff still is one glance. It is a recorded constant
> like `check-panel-diff-size.sh`'s cap — re-tune it by editing this sentence with the reason,
> never by arguing it away per run.

### implement.md — The record carries its own corrections

> This is expected practice
> on every task, not one implementer's habit (gymie KAN-29's self-review: corrections recorded in the task
> itself made that panel's bookkeeping findings cheap to adjudicate; KAN-407 makes it the rule).

### implement.md — A deviation from the plan records as a dated `Correction:` paragraph

> the panel verifies the deviation instead
> of discovering it (gymie kan-361's task 2: three tests moved to a sibling file after detekt's
> `LargeClass` refused the planned one, recorded on the task this way).

### implement.md — A pivot reconciles the three artifacts together

> gymie kan-579 pivoted two tasks through `tasks.md`
> alone; its `proposal.md` never caught up, and the panel found the drift.

### implement.md — The gated per-task reviewer

> **Never one reviewer dispatch per gate-fired task, and never one per
> group on `small`/`regular`** (gymie KAN-527: nine one-task dispatches on the first twelve tasks of a
> 21-task change, before the operator stopped the run — the review-dispatch count tracks the
> change's size, never its task count).

> never
> implementation WIP (KAN-628, the improvised WIP-commit dance
> two runs performed to get past exactly this).

### implement.md — 3. Documenting a fix, before implementing it (plan growth)

> gate-time re-planning visible as a trend in the app rather than a
> per-change surprise (KAN-415)

## Moved by KAN-856 — planning-session trims

Verbatim passages KAN-856 moved out of the planning session's run-loaded files; each is the reason behind a rule that stays where it was.

### brainstorm.md — A. Resolve the change and write `STARTED` (KAN-856)

`planningEffort` and `models.default` are written `null` and stay `null` for the life of the
change: `/flow` asks no planning-effort or model question on a creating run, and models are chosen per dispatch by the Decide step, not recorded per change.

The
operator sees `STARTED` recorded the moment they invoke `/flow`, whether or not the run goes on to
finish brainstorming in the same sitting.

### brainstorm.md — the worktree is created inside `flow.kickoff` (KAN-856)

The record written at
**A** carries `"worktrees": {}`, and the next durable write would otherwise be
`flow.isolate-workspace`'s, a whole brainstorming, design-gate and planning run later; a run
killed after the add but before that stage must not leave a real worktree and a pushed branch
behind an empty map — the stale `STARTED` record a resumed run reads as a fresh creating run.

**A command's non-zero exit
ends your turn** naming the command and its output — a worktree that cannot be set up fails
`flow.verify` later anyway, and the operator should see it here.

### brainstorm.md — Run brainstorming and planning directly

The operator approving or answering the question is
approving against the summary they were actually shown.

### brainstorm-planner.md — D. Writing plans (KAN-856)

A premise the run disproves strikes the task during
planning, before the panel reads the plan — the same verify-the-premise move the filed-finding
reachability check makes at checklist open, applied to each task's own premise.

KAN-750's task 11 was blocked for
35 minutes on a plan that assumed Playwright's `reducedMotion` took effect under `use`, where it
takes effect only under `contextOptions`.

`**Baseline:**` — the expected test counts, as `before=<N> after=<N>`, counted statically from
test declarations (`@Test` and its language equivalents), never from a test run, and each
task's delta its own: `after` is `before` plus the tests the task's own `**Tests:**` field
names, so two tasks dispatched in parallel count independently — neither's baseline depends on
the other's edits.

The recorded command names no ref of its own — no tree, sha, branch
or revision argument anywhere inside it: the re-measurement checks each of its two points out
itself and runs the command in that working tree, so a command naming a ref measures that one
tree at both points, reports identical counts, and fails any nonzero declared delta.

This line is required and explicit on every plan — never omitted, per
this repository's "missing rather than dropped" convention.

## Moved by KAN-857 — implementation-session trims

Sentences `skills/flow/implement.md` and the files split out of it (`document-fix.md`,
`cross-repo-worktrees.md`) carried as reasons, moved here verbatim. Each rule they explained stays
in the run-loaded file.

### implement.md — Dispatch sites (KAN-857)

A subagent's prompt cache lives five minutes, so a child woken later re-writes its whole context at the cache-write price.

**The `flow-<effort>` family (`agents/flow-low.md`, `agents/flow-medium.md`, `agents/flow-high.md` — three definitions, one per effort, each carrying `effort:` and no `model:`, since the Agent tool's dispatch-time `model` parameter overrides a definition's `model` while `effort` has no dispatch-time parameter) carries a `tools:` allowlist that omits `Agent`** — the NO DELEGATION paragraph is backed by a capability the dispatched agent structurally does not have, not by prompt text alone.

**A dispatch whose agent dies is closed with `-outcome aborted`, reported, and not retried**: print `/flow <name>` for the operator — a re-run resumes from whatever was left (checkbox state, the state file's worktrees, findings in the store) through this file's own re-entry rules, and the operator should see the death rather than have it hidden by a second dispatch.

### implement.md — 1. Load context and validate the plan (KAN-857)

The plan and design were written against a snapshot of the base branch and the capability specs, and a concurrent merge can outdate both while this change waits.

### implement.md — 2. Isolate the workspace (KAN-857)

A run that stops, is interrupted, or is resumed after a context compaction anywhere between here and `flow.write-in-progress` (`skills/flow/verify-and-handoff.md`) must not leave the state record looking like a creating run with no worktrees, when real worktrees, branches and commits already exist — that mismatch is exactly what **Reading the state** (`skills/flow/SKILL.md`) uses to decide whether this is a creating run at all, so a stale record makes a resumed session re-derive everything from scratch or misclassify the run.

### implement.md — 2. Isolate the workspace, additional worktrees (KAN-857)

The working directory is what resolves the peers file's relative entries — `ResolvePeer` stats a declared peer path against the process working directory, so from inside a worktree `../<peer>` resolves into `<project>/.worktrees/` and the link is always refused — and `--root <abs-worktree>/spectre` is what writes the satellite-side `link.md` into the worktree, where `check-unfinished-work.sh` reads it at integrate.

**A refusal is a hard failure of this stage**: report it and stop the run — a change whose cross-repo link cannot be established lands at integrate with a false OUTSTANDING verdict that forces hand verification.

**A worktree add that fails is a hard failure of this stage**, reported and stopping the run exactly like a refused link: a declared app left unresolved is the commit destination a later stage creates by hand, off the wrong base, unrecorded.

### implement.md — 3. Documenting a fix (KAN-857)

When the count has reached **6**, the re-plan budget, this fix round is offered the planning pass before anything is appended: an append past this budget is how a change outgrows its own proposal without anyone deciding it should.

The glance that passed the control is what the operator is contesting, and repeating it answers nothing.

After the planning pass writes its appends and bumps `**Tasks appended:**`, the plan's new size is recorded as the next observation of the change's plan-growth series — gate-time re-planning visible as a trend in the app rather than a per-change surprise:

```bash
flow tasks count -C <worktree> <name>
```

### implement.md — 4. Execute (SDD + TDD) (KAN-857)

The invariant is the working tree, not the build tool, and it holds however file-disjoint two tasks look on paper: UI fixes routinely touch shared files — icon sets, shared components, menu wiring — neither task named.

`begin` is recorded immediately before the dispatch — or, when several launches go out together in one message, in the next Bash call, one call for all (step 3 of **The next implementer overlaps the guard**, below) — carrying no `-agent-id`: the daemon captures the agent's identifier — Claude Code writes it into the parent transcript's own launch tool result, the harvester pairs that result with the begin nearest it in time, and the row's empty `agent_id` is filled from it, which is why `begin` never waits for the launch's id.

A generic warning ("be careful with git") is not a carry: kan-527's chunk 1 worked under one and its git incident happened anyway; chunk 2's dispatch carried the specific three and finished with ancestry confirmed clean.

The child's own claim that nothing went wrong never closes an incident — the same reason the content markers assert by grep, never by the child's prose (KAN-643's rule).

The write journals on store failure like every record write and never blocks the run — the row is the evidence a later cross-repo fix is justified and shaped by, which silence cannot hold.

### implement.md — The parent runs the full suite (KAN-857)

A subagent's prompt cache lives five minutes, the parent's an hour: a suite run outlasting five minutes inside an implementer re-prices that implementer's whole context on its next turn.

### implement.md — The review gate (KAN-857)

The undeclared-path arm is the gate's risk half: a commit reaching past its own plan declaration is exactly the surprise a second reading exists for, however few lines it runs.

### implement.md — The record carries its own corrections (KAN-857)

The review gate's undeclared-path arm reads the paths the refusal named — the pre-correction declaration lives in the refusal, not in any field the transcription can overwrite — so the disclosure cannot disarm the gate.

### implement.md — A deviation from the plan (KAN-857)

The archived plan then reads as what actually shipped, and the panel verifies the deviation instead of discovering it.

### implement.md — The plan tree survives every reviewer dispatch (KAN-857)

A reviewer that closed clean over a tree it changed has reported about evidence it destroyed — the verify answers whether the tree survived, never the prose.

### implement.md — content markers (KAN-857)

The markers sit beside the plan-tree guard, never in its place: that guard asserts git's view of one directory, while markers pin known content anywhere in the tree, independent of git entirely — content that was never committed has no git answer at all, which is exactly what KAN-579's destroyed-and-self-reported-restored artifacts were.

### implement.md — Turn discipline (KAN-857)

The loop is bounded at 240 s rather than the Bash tool's ten-minute cap: a wait long enough to outlast the prompt cache re-prices the whole context on return, while a bounded wait's `still-running` turn reads it at the cache rate and keeps it warm.

### implement.md — Read discipline (KAN-857)

The parent does a large amount of reading across one long-lived context, so keeping that context small is what keeps a warm call cheap.

### implement.md — The handshake (single-model harness, now in model-policy.md)

The mapping already fixes what the handshake exists to establish, and the dispatch's ledger line records it.

## Moved by KAN-857 — review-panel trims

Sentences `skills/flow/review-panel.md` and the files split out of it (`review-panel-late-fix.md`,
`review-panel-fix-round.md`, `review-panel-optional-slots.md`) carried as reasons or as the
renderer's output format, moved here verbatim. Each rule they explained stays in the run-loaded
file.

### review-panel.md — The roster (KAN-857)

The `flow-<effort>` definitions are this repository's own, whose `tools:` allowlist omits `Agent` — **No forking** below is backed by a capability the slot structurally does not have, not by the NO DELEGATION paragraph alone; `general-purpose` is a harness-provided type whose tool set cannot be restricted.

### review-panel.md — final-review.diff (KAN-857)

The first line is the file's own answer to the reviewer who reads it as anything narrower: it is a plain working-tree diff against the merge base, so work still unstaged or uncommitted is already in it.

### review-panel.md — Bundled dispatch (KAN-857)

The same plan-tree discipline rides every round, re-run rounds included: before the launches the parent makes the reviewer-dispatch planning commit (**Planning commits**, `skills/flow-contracts/git-boundaries.md`) and runs `check-plan-unchanged.sh snapshot <worktree> <name> <snapshot-file>`, and once every report file exists it runs `check-plan-unchanged.sh verify <worktree> <name> <snapshot-file>` before any report is read or finding recorded — the slots read those artifacts, and a flight that changed them has invalidated the reviews that flew.

Exit 1 or 2 stops the round the same way the per-task reviewer's stop works (the plan-tree rule of **4. Execute (SDD + TDD)**, `skills/flow/implement.md`); the slots' own read-only briefs are the first line of defense, this verify is the assertion that a breach cannot slide past as a clean report.

### review-panel.md — the reproducer rule (KAN-857)

`prove-reproducer.sh <worktree> <pre-fix-ref> <reproducer-path>` runs both legs and prints both exits for the record — the detached scratch worktree it materializes at the pre-fix commit, with the script copied to the same worktree-relative path inside it, is what keeps the pre-fix leg off the already-fixed live tree, the failure mode where a pre-fix check silently reads a fixed worktree and proves nothing.

### review-panel.md — the rendered record's format (KAN-857; the renderer writes it, stats/internal/records/render.go)

The record carries a findings table, one row per finding:

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Mutation | Minor | `src/Foo.kt:42` | replaced the silent catch |

and, below it, the marker block — one line per row, plus the count:

```
findings-total: 1
finding-status: F1 fixed
```

The marker format is never quoted inside the record itself. The renderer neutralises any marker
label a finding's note or location happens to carry, on the way out only.

**The reproducer each finding's slot supplied gets a marker block of its own**, separate from the
`finding-status:` block above:

```
reproducers-total: 1
finding-reproducer: F1 scripts/test-check-plan-unchanged.sh
```

A finding recorded with no reproducer renders the `none — <reason>` exemption form.

**The table carries no status column, on purpose.** To read a finding's state, look up its `F<n>`
in the marker block.

### review-panel.md — The late-fix reduction (KAN-857)

An appended fix that a full pass would bury under re-covered ground is reviewed by one targeted dispatch on what the fix actually changed; the full-read coverage stays exactly where it earns its keep, on every round this reduction does not fire for.

A consuming project typically carries none of those paths and the condition holds vacuously there; a fix to anything the panel reads as its own brief is never its own reviewer.

### review-panel.md — Panel re-runs (KAN-857)

The entry check ran once, before pass 1; a base that moves while earlier rounds ran would otherwise reach the final round — and then integrate — unchallenged, its conflict surfacing only after review has closed.

A conflict found here surfaces while the panel is still active and the operator is already engaged.

The scoping exists because a round that re-reads a growing fix diff regress-checks by volume, not by site.

A `fixed` recorded in the same call as its re-run stands verified before that exit is known — the ordering KAN-582's F1 rode in on, luck rather than discipline — which is the defect the close guard's fixed-without-clean-rerun class below exists to catch.

### review-panel.md — The fix round mutation-proves what it changed (KAN-857)

This catches an undeclared file the fixup added and a declared test it removed or renamed, read post-autosquash.

This binds the fix round every run — the obligation is the round's, not a slot's, so a run where Mutation is not in the resolved roster or added this run is exactly where the round's own proof is the only mutation reasoning that happens at all.

### review-panel-optional-slots.md — The throwaway worktree (KAN-857)

`git diff HEAD --binary` — against `HEAD`, not a bare `git diff --binary` — is the same "staged and
unstaged together" semantics `final-review.diff` already uses, and covers the transplant in
one diff rather than the working-tree-only diff a bare `git diff` produces: a bare `git diff` misses
anything staged, and (independently) fails to reconstruct a rename whose move is already reflected
in the index. `--allow-empty` on the `apply` side makes the sequence a no-op, not a failure, when
there is nothing to transplant — the common case, since task and fix-round work is committed and
`worktree add --detach ... HEAD` already carries every committed change on its own. The
untracked-file loop reads `git status --porcelain -z`, NUL-delimited, into `read -r -d ''` — the
plain-text `awk` form cannot survive git's quote-escaping of a filename with a space or another
special character, and silently drops that file from the copy; the `-z`/NUL form carries the literal
byte string through untouched, regardless of what the filename contains. The scaffold lines after
the loop exist because the slot dispatched into the copy resolves the bundle paths its prompt names
— `dispatch-context.md`, and the report file it writes — against its dispatched root. The `-d` test
keeps a canonical worktree with no bundle yet a no-op rather than a failure.

It rescues the slot's reproducers beside its reports — both are recorded as worktree-relative paths the parent later runs, and one left only in the copy dangles.

A copy-side file newer than the canonical one replaces it — the wall-clock re-dispatch's fresh report outranking a timed-out attempt's — and the `[ -s ]` tests keep an empty copy-side file from being copied.

Mutation mutates code in place to run its brief; every other slot only reads the diff.

Dispatching a mutating slot into the same worktree a reading slot concurrently reads is the collision — a mutation applied for one slot's test is visible to whatever a concurrently dispatched reading slot reads from `<worktree>` at that moment.

## Moved by KAN-858 — visual verification

### visual-verify-verifier.md — Visual verification, step 8 (capture)

**Every capture asserts the view is on screen before it is taken** — gymie-playwright's `expectRoom` — so a missed navigation fails the capture instead of becoming its baseline (KAN-747).

### visual-verify-verifier.md — Visual verification, step 10 (the `mockups` declaration gates the frame comparison)

The state sweeps run on the captures alone and are the default on every `flow.visual-verify` run, composed or not: a verification that transcribes no text run (sweep 5), checks no element against its container's bounds (sweep 6), scrolls no region to its end (sweep 7) and re-derives no derived value (sweep 8) is the operator-driven round this stage exists to replace.

### visual-verify-verifier.md — Visual verification, step 10 (state 3, every option of every picker)

`rules/design-mockups-are-specs.mdc` puts the same question to the implementer; this is the verifier's side of it.

### visual-verify-verifier.md — Visual verification, step 10 (sweep 9, bounded rows)

Nothing else here sees it — the composite's ratio barely moves when a row's content is right and only the whitespace around it is wrong, the per-control crop matches its control wherever it sits, and the order-and-containment sweep finds every element present and inside its container.

### visual-verify-verifier.md — Visual verification, step 10 (sweep 10, non-text ink)

Every sweep above starts from something the capture shows or the text says: a control, a text run, a row, a value. A line the frame draws and the capture omits is on none of those lists, and the composite's ratio barely moves for a 1px rule, so nothing above asks for it and a verifier who has matched every label and number walks away satisfied.

### visual-verify-verifier.md — Visual verification, step 10 (edge-to-edge measurement)

Not `content.padding`: it boxes every non-fill pixel in the container, the other rows' text included, and a fill covering half the box flips which colour counts as `fill` (both shown on the incident's own geometry by `TestMeasureVisualProperties` in `<agents repo>/stats/internal/guard/measure_visual_properties_test.go`).

### visual-verify-verifier.md — `## Report`

The template's fence carried the authoring provenance `verified:design.md section 3 of this change` (`kan-171-generic-visual-verification-step`).

### visual-verify.md — Visual verification, step 12 (commit)

**Never push** — see `no-automatic-push` (design.md): a file inside a repository cannot authorise a push to another repository, so no guard here grants one, worktree or not.

## Moved by KAN-858 — reviewer templates

### principles-reviewer-prompt.md — preamble

The principle list itself is **not** restated here — the reviewer reads
[engineering-principles.md](engineering-principles.md) (this file's sibling, inside the installed
skill directory), which is the single source of truth. Never paste the list into a dispatch
prompt. The relative link above resolves only for a reader of *this* file; the dispatched
subagent's working directory is the **project worktree**, which has no `skills/` tree of its own,
so the dispatcher must substitute `[PRINCIPLES_PATH]` with the **absolute** path before spawning.

## Moved by KAN-859 — integrate and archive

### integrate.md — 3. Commit the staged work (reshape)

Moved verbatim from the paragraph after the `reshape-branch.sh` call, where it followed "never the state file's now-stale pre-rebase value for that worktree."; the duplicate `reset --soft` sentence after it was cut, **Run 1 — the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`) being canonical for it:

This keeps every planning commit as its own commit on the merge base and collapses every per-task
and fixup commit back into the working tree, uncommitted; using the
stale value here would also collapse in the upstream commits the rebase just brought in, silently
smuggling them into the implementation commit below.

### archive.md — 4. Commit the archive, the `bash -c` pin (KAN-859)

Moved verbatim from the paragraph after step 4's shell block; the duplicate sentences around it were cut, step 4 of **Run 2 — the branch is merged** (`skills/flow-contracts/finish-contract-run2.md`) being canonical for them:

The loop is invoked through `bash -c` (KAN-816):
its `set -- $pair` relies on word-splitting an unquoted parameter, which zsh — the shell an
executing session may run — does not do, leaving `$2` empty and both records uncopied; the pin
fixes the loop's bash semantics rather than rewriting it in zsh-only splitting syntax.

## brainstorm-planner.md — Model and effort, the effort roll (medium-effort-default)

The operator asked for `medium` effort on about 80% of implementer and reviewer dispatches. A free
per-dispatch choice gave no share at all to hold to, so the split rides on a fourth name-derived
roll, `effort_roll = sha256("<name>effort") mod 100`, the `compact`/`experimental`/`bundle` rolls'
own mechanism: below 80, every planner-chosen effort is `medium`; at or above it, the free choice
that existed before. It is per change, not per dispatch, so the 80% holds across changes rather
than inside one. A hard seam keeps `high` in the rolled 80% too, and those dispatches count toward
the non-`medium` share, which is why the target is "about" 80% rather than exactly. **Rejected — a
new per-dispatch roll:** the planner would have to hash each dispatch's own key, a ceremony the
per-change roll avoids. **Rejected — reusing an existing roll with a different threshold:** it
would tie effort to roster size or grouping. The rerun pair stays at its fixed `low`, since a
re-run reads a delta and is outside what the operator asked to change.

## brainstorm-planner.md — Decide, step 5 visual verification (visual-verify-decided-at-decide)

`check-visual-verify-dispatched.sh` fires `VISUAL-VERIFY-MISSING` whenever a diff touches a declared
`ui paths` glob and no verifier dispatch is recorded. It exists because of kan-30, where a run wrote
`flow.visual-verify`'s marks around no verifier and shipped broken screens. A glob is coarse,
though: gymie declares `src/gateway/src/main/**`, so kan-863's actuator-only scrape-token filter
matched it, the implement run decided mid-run that there was no UI surface and skipped silently, and
integrate then put the three-course unfinished-work question to the operator over a verification
nothing could have needed; kan-866 hit the same false positive. Both failures are one defect — the
"is visual verification needed?" judgement was made nowhere visible: implicitly by the glob, or
silently by an implementing agent.

The judgement moves to Decide, beside the other per-change decisions, where it is recorded in
`decision.json`, rendered as a row of the `## Decision` block and approved at the plan gate. That
keeps kan-30's guarantee — no skip the operator did not see — while letting an obvious no-UI change
skip without a prompt. The integrate guard accepts the recorded skip as `VISUAL-VERIFY-OK: skipped
at Decide — <reason>`; `required` with no verifier dispatch stays `MISSING`, now a real signal (the
plan said required and nobody verified). A fix run passes no plan gate, so it may raise `skipped`
to `required` but never lower it — a lowering there would be exactly the unseen skip kan-30 was.
`visual` is optional in the record, and absent reads as `required`, so a change decided before the
field existed, and `/flow-fast`, which never runs `flow.visual-verify`, behave as before.
**Rejected by the operator — turning `VISUAL-VERIFY-MISSING` into a report-only line:** it removed
the prompt for every change, the kan-30 shape included, where the operator wanted the judgement made
at planning, visible and gated.

### Fewest operator actions — no round caps, and a verifier that finishes (KAN-870, 2026-10-02)

The operator, after gymie KAN-870's visual verification ended partly done twice: "Fix all, it must
not happen in the future", and, the same day: "In general the top priority is the least actions
from my side. Less stops. Number of re-review and re visual verify shouldn't be limited."
**Fewest operator actions** (`skills/flow-contracts/pipeline.md`) is the rule that came of it.

**Retired — every round count on a re-review or a re-verification.** The verify stages' five-round
fix loop, the panel's escalation of a defect still open after two automatic rounds, and the
verifier's single re-dispatch ("the second report is final; never a third verifier on the same
HEAD") each converted a defect the run could still fix into a stop the operator had to clear. On
KAN-870 the last of them did worse: it forbade finishing a verification the verifier had simply
not completed. What a cap guarded against — a loop that repeats the same failed fix — is now met
by changing the approach on a round without progress (a fresh opus/high fixer carrying every
earlier report, another reproduction) rather than by counting rounds. **Rejected — a remainder
that stops the stage when it is itself incomplete.** It re-creates the cap one level down; an
incomplete remainder is dispatched again under the same no-progress rule.

**The verifier runs on opus at `low`, and no browser-driving rule moves it.** Both KAN-870
verifiers ran on Sonnet — the operator's global instructions put browser-driving agents there —
and both stopped after the tests and the capture, listing the reading, composing and
record-writing steps under "Not done". **Rejected — raising the verifier's effort to `medium`.**
The operator kept it at `low`: completeness is enforced by the report guard, not by effort. **A
partial report is refused mechanically** (`check-verify-report.sh`), not by reading its prose: the
parent had accepted one whose prose named the undone steps.

**Motion is recorded, not inferred from stills.** KAN-870's substance was motion — exits, slides,
drag gestures, an entrance removed — and a still capture compared against a mockup has nothing to
say about any of it. **Rejected — detecting motion by grepping the diff for animation APIs.** The
API names are per-framework and per-project; the parent naming the motions from `design.md`,
`proposal.md` and the diff is the one rule that holds across projects.
