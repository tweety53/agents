> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

# kan-515-visual-verification-tooling-crop-mockup-frames

KAN-515 · implementation plan.

**Goal:** `compose-mockup-frames.py` learns a declared `mockup frame` geometry, crops each frame to
its content area with the caption band detected per frame, writes a three-panel composite with the
differing-pixel ratio on stdout, and reports a capture whose size differs from the cropped frame as
an exit-1 finding. The contract, the section guard, the stage step and both harnesses move with it.
`docs/superpowers/research/kan-515.md` (sections 2–5) is canonical for every decision a task
implements until `/flow` folds it into this change's `design.md`.

Five tasks, dependency order, one repository (`agents`). No spec, no migration.

**Baseline, measured before any edit, on `main` at `f35ae51`:** `test-compose-mockup-frames.sh`
carries 19 cases, `test-check-visual-verification.sh` 33.
<!-- measured: grep -c '^# Case' on each harness @ main f35ae51 -->

**Every task's verify step is the lint guards its own files need plus its own harness**, never
`scripts/run-guard-tests.sh` — that run is the last bundle's FULL SUITE paragraph and
`flow.verify`'s.

---

- [x] 1. Geometry crop, diff panel, ratio and size gate in `compose-mockup-frames.py`

**Build:** green
**Files:** `scripts/compose-mockup-frames.py`, `scripts/test-compose-mockup-frames.sh`
**Tests:** `test-compose-mockup-frames.sh` — `Case 20: geometry declared — frame cropped to
border/status/caption, composite is three panels wide`, `Case 21: caption band detected per frame
— two frames of different heights crop to their own bottoms`, `Case 22: capture size differs from
the cropped frame — finding names both sizes, exit 1, sibling line still composed`, `Case 23: stdout
carries diff=<ratio> and the ratio is 0.0000 for an identical pair and 1.0000 for an inverted
pair`, `Case 24: no geometry argument — two panels, native size, diff=n/a, exit 0 (the pre-change
path)`, `Case 25: a malformed geometry argument — exit 2, usage on stderr`
**Regression:** reverting this commit returns the compose step to pasting the status line, border
and caption beside the capture; Case 22 is the only proof a wrong viewport blocks rather than
composes, and Case 24 the only proof other projects are untouched.
**Baseline:** `test-compose-mockup-frames.sh` before=19 after=25
<!-- predicted: grep -c '^# Case' scripts/test-compose-mockup-frames.sh after task 1 -->
**Commit:** `feat(scripts): crop mockup frames, add a diff panel and a size gate to compose-mockup-frames`

  - [x] **Step 1: Parse the optional fourth argument.** `argv[4]`, when present, is
    `scale=<int> status=<int> border=<int>` (three whitespace-separated `key=value` fields, any
    order, every key required once, every value a non-negative integer, `scale` ≥ 1); anything
    else is `usage_error()` (exit 2). `len(argv)` is 4 or 5; the usage line gains
    `[scale=<n> status=<px> border=<px>]`.
  - [x] **Step 2: Crop.** With geometry, after the frame PNG loads: `s = scale`, page background
    `bg = frame.getpixel((0, 0))`, `top = (border + status) * s`, `left = border * s`, `right =
    w - border * s`; `frame_bottom` is the first row `y` in `range(top, h)` where every pixel in
    `[left, right)` equals `bg`, else `h`; crop to `(left, top, right, frame_bottom - border * s)`.
    A crop with zero or negative width or height is a finding on that map line: `frame <id>: declared
    geometry leaves no content area`.
  - [x] **Step 3: Size gate.** With geometry, `cap.size != frame.size` after the crop is the
    finding `capture <cw>×<ch> vs frame <fw>×<fh> for <frame id>` (multiplication sign, exactly the
    note's shape); the pair is skipped, later lines still compose, exit 1 at the end as today.
  - [x] **Step 4: Three panels.** Sizes equal: `diff = ImageChops.difference(cap, frame).convert("L")
    .point(lambda v: 255 if v else 0)` converted to RGB; canvas width `cap.width * 2 + frame.width +
    2 * GUTTER` (three panels, two gutters), height the max; paste capture at 0, frame at
    `cap.width + GUTTER`, diff after the second gutter. `ratio = (count of nonzero diff pixels) /
    (cap.width * cap.height)`, from `diff.getbbox()`-independent `diff.histogram()[255]` or
    `ImageStat` — either, as long as it is exact. Without geometry the existing two-panel layout and
    native sizes stay byte-for-byte.
  - [x] **Step 5: stdout.** Each output line becomes `<abs composite path> diff=<ratio:.4f>`, or
    `<abs composite path> diff=n/a` without geometry. Update the module docstring's Interface,
    Layout and Exit-code paragraphs so the header stays canonical for the script's exit codes.
  - [x] **Step 6: Harness.** Add Cases 20–25 named above to `test-compose-mockup-frames.sh` on its
    `make_png`/`new_root` helpers: build a frame PNG with a 2px border colour, 52 rows of status
    band, a content block, 2px bottom border and a caption band of the page colour of differing
    heights for Case 21; assert composite widths (`3·w + 32` vs `2·w + 16`), the stderr finding
    text, the exact `diff=` values and exit codes. Keep Cases 1–19 unchanged.
  - [x] **Step 7: Verify.**

  ```bash verified:both scripts exist on main f35ae51 and take these argument shapes
  scripts/test-compose-mockup-frames.sh
  scripts/check-vocabulary.sh
  ```

- [x] 2. `mockup frame` in the section guard's closed vocabulary

**Build:** green
**Files:** `scripts/check-visual-verification.sh`, `scripts/test-check-visual-verification.sh`
**Tests:** `test-check-visual-verification.sh` — `Case 34: a mockup frame row with a well-formed
value — exit 0`, `Case 35: a mockup frame row whose value is not scale=/status=/border= — exit 1,
names the row and the expected shape`, `Case 36: mockup frame declared twice — exit 1, duplicate
setting`
**Regression:** reverting this commit makes the guard drop every `mockup frame` row as an
unrecognised setting, so the stage never receives the geometry and composes uncropped again with
no violation printed.
**Baseline:** `test-check-visual-verification.sh` before=33 after=36
<!-- predicted: grep -c '^# Case' scripts/test-check-visual-verification.sh after task 2 -->
**Commit:** `feat(scripts): accept a mockup frame row in the visual verification section`

  - [x] **Step 1: Vocabulary.** In `check_setting_row`, add `key != "mockup frame"` to the closed
    list; the violation message and the header comment above it name six settings. `mockup frame`
    is optional — no `require_nonempty` — but a present value must match
    `^scale=[1-9][0-9]* status=[0-9]+ border=[0-9]+$` after `trimcell` (fields in that fixed order
    in the file; the script accepts any order, the guard pins one so a row is greppable), else
    `violation(lineno, "Setting `mockup frame` must be `scale=<int> status=<px> border=<px>`, got `<disp>`")`.
    A second `mockup frame` row is the existing duplicate-setting violation.
  - [x] **Step 2: Harness.** Cases 34–36 on the harness's own fixture-writing helpers, extending
    the minimal valid section of Case 3.
  - [x] **Step 3: Verify.**

  ```bash verified:both scripts exist on main f35ae51
  scripts/test-check-visual-verification.sh
  scripts/check-vocabulary.sh
  ```

- [x] 3. The contract: `mockup frame`, the crop rule, the three-panel composite

**Build:** green
**Files:** `skills/flow-contracts/project-configuration.md`
**Tests:** **none** — a contract carries prose; the guards in the verify step are its checks.
**Regression:** reverting this commit leaves a project with no documented way to declare the
geometry task 2 now accepts, and the sidecar paragraph describing a two-panel composite the script
no longer writes.
**Baseline:** `test-check-visual-verification.sh` before=36 after=36 (unchanged — this task adds no test)
<!-- predicted: no harness change in task 3 -->
**Commit:** `docs(flow-contracts): declare the mockup frame geometry and the three-panel composite`

  - [x] **Step 1: Settings table.** Add the row `| \`mockup frame\` | no | \`scale=<int> status=<px>
    border=<px>\`: the mockup PNGs' scale factor, the status-line height and the border width in
    logical px. Declared, the compose step crops each frame to its content area — the bottom found
    per frame as the first row below the status line uniformly equal to the page colour at (0, 0) —
    and blocks on a capture whose size differs from the crop. Absent, frames compose at native size,
    uncropped, with no size check. |`; change "the five rows below" to six.
  - [x] **Step 2: Sidecar paragraph.** Replace "The composite … is the captured frame on the left
    and the mockup on the right, both at native size" with: with `mockup frame` declared, capture |
    cropped frame | difference panel (white where any channel differs), one `<composite path>
    diff=<ratio>` line per pair on stdout, the ratio a number the verifier reads and never a
    threshold; without it, the two-panel native-size composite and `diff=n/a`.
  - [x] **Step 3: Verify.**

  ```bash verified:the three guards are .flow/project.md's ## lint rows for markdown
  scripts/check-vocabulary.sh
  scripts/check-references.sh
  scripts/check-contract-budget.sh
  ```

- [x] 4. `flow.visual-verify` step 9 passes the geometry and reads three panels

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** **none** — a stage file carries procedure; the guards in the verify step are its checks.
**Regression:** reverting this commit leaves the verifier dispatched without the geometry, so the
compose command runs with three arguments and every project composes uncropped, as before.
**Baseline:** `test-check-visual-verification.sh` before=36 after=36 (unchanged — this task adds no test)
<!-- predicted: no harness change in task 4 -->
**Commit:** `feat(flow): pass the mockup frame geometry to the compose step and read its diff panel`

  - [x] **Step 1: The verifier dispatch.** In the prompt-contents sentence (the one naming
    "its resolved `mockups` root when declared"), add "and its `mockup frame` value when declared".
  - [x] **Step 2: Step 9.** The compose command gains the fourth argument, quoted:
    `compose-mockup-frames.sh <spec>.mockups <worktree>/<mockups> <changeRoot>/visual-verification
    '<mockup frame value>'` — the argument omitted when the row is absent. The "Exit 0 prints one
    composite path per line" sentence becomes one `<composite path> diff=<ratio>` line per pair;
    the verifier reads every composite's third panel and states, per pair, what the white regions
    are and the ratio. Add to the exit-1 list: "a capture whose size differs from the cropped frame".
  - [x] **Step 3: Step 10.** The `visual-verification.md` entry per composed pair also records the
    ratio.
  - [x] **Step 4: Verify.**

  ```bash verified:the three guards are .flow/project.md's ## lint rows for markdown
  scripts/check-vocabulary.sh
  scripts/check-references.sh
  scripts/check-contract-budget.sh
  ```

- [x] 5. Clear the four pre-existing `flow.verify` blockers on `main`

**Build:** green
**Files:** `skills/flow-fast/SKILL.md`, `skills/flow-fast/scripts/check-plan-shape.sh`,
`skills/flow-fast/scripts/check-plan-shape.py`,
`skills/flow-fast/scripts/check-task-commit-fields.py`,
`skills/flow-fast/scripts/plan-class.sh`, `scripts/check-contract-budget.sh`,
`scripts/check-visual-verify-dispatched.sh`
**Tests:** **none** — every check this task repairs is an existing guard or an existing Go test;
the repair is proven by those exiting 0, not by a new case.
**Regression:** reverting this commit restores all four failures — `check-guard-symlinks.sh` exit 1
on four missing `skills/flow-fast/scripts/` symlinks, `check-contract-budget.sh` exit 1 on
`skills/flow-fast/SKILL.md`, `TestStageKeysMatchFlowFastSkillTable` on the absent `## Stage keys`
heading, and `check-vocabulary.sh` exit 1 on `MERGE_BASE` in `check-visual-verify-dispatched.sh`.
**Baseline:** `run-guard-tests.sh` before=FAIL after=PASS (no case added; the flow-fast tree the
existing `check-guard-symlinks.sh` case scans is what changed)
<!-- measured: scripts/run-guard-tests.sh before and after this task -->
**Commit:** `fix(flow-fast): close the guard-symlink, budget, stage-key and vocabulary gaps`

  - [x] **Step 1: Symlinks.** `skills/flow-fast/scripts/` gains `check-plan-shape.sh`,
    `plan-class.sh` (rule 2, invoked by `SKILL.md`) and the two sibling dependencies
    `check-plan-shape.sh` resolves from its own directory, `check-plan-shape.py` and
    `check-task-commit-fields.py`.
  - [x] **Step 2: `## Stage keys`.** `skills/flow-fast/SKILL.md` gains the table
    `TestStageKeysMatchFlowFastSkillTable` pins the vocabulary to — the 26 keys the skill already
    marks, grouped by its own eight steps. No key changed; only the statement of them was missing.
  - [x] **Step 3: Budget.** That table grows the file, so its `budgets()` row rises to the new
    size plus 25%, per the guard's own ratchet rule.
  - [x] **Step 4: Vocabulary.** `check-visual-verify-dispatched.sh`'s shell local `MERGE_BASE` is
    renamed `BASE_SHA`. The retired-vocabulary list matches the state file's retired `MERGE_BASE`
    field, and this variable merely collided with the name; renaming it is the fix, where a
    `vocab-guard:allow` marker would have taught the guard to lie about a live script.
  - [x] **Step 5: Verify.** The whole `## lint` and `## test` lists, both clean.
