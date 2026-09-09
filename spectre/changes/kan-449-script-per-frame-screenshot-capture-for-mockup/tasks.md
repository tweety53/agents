# kan-449-script-per-frame-screenshot-capture-for-mockup

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Six commits. The compositor and its harness first (tasks 1–2), the guard and contract that admit
the `mockups` setting next (tasks 3–4), the stage step that runs it (task 5), and the one project
declaration that uses it (task 6, in the `gymie` repository). `design.md` beside this file is
canonical for every decision; nothing below restates one.

**Baseline, measured before any edit:**

- `scripts/run-guard-tests.sh`: 57 harnesses, 56 passed, 1 failed — `test-check-task-build-green.sh`
  case 13, which passes from the main checkout and fails only from inside this worktree
  (`spec-root: … carries both spectre/changes/ and openspec/changes/`). Pre-existing and
  environmental; no task here touches that guard or its harness.
  <!-- measured: scripts/run-guard-tests.sh @ branch spectre/kan-449-script-per-frame-screenshot-capture-for-mockup, worktree at merge-base 2f1ac1f -->
- `skills/flow/verify-and-handoff.md` is 33002 bytes against a 33199-byte budget — 197 bytes of
  room, less than task 5's addition, so task 5 raises the `budgets()` row.
  <!-- measured: wc -c skills/flow/verify-and-handoff.md; grep verify-and-handoff scripts/check-contract-budget.sh @ merge-base 2f1ac1f -->
- `skills/flow-contracts/project-configuration.md` is 49042 bytes against 61302 — task 4 fits.
  <!-- measured: wc -c skills/flow-contracts/project-configuration.md; grep project-configuration scripts/check-contract-budget.sh @ merge-base 2f1ac1f -->
- Pillow imports on this machine: `python3 -c 'from PIL import Image'` exits 0 with
  `/opt/homebrew/bin/python3` (Pillow 12.3.0). Pillow is a machine prerequisite, never a
  dependency this repository declares — `design.md`'s `pillow-compositor`.
  <!-- measured: python3 -c 'import PIL; print(PIL.__version__)' @ this machine, 2026-09-10 -->
- `gymie/docs/design/group-sessions/screens/` holds 34 tracked `<frame id>.png` files
  (`G1.png … K4.png`).
  <!-- measured: git -C /Users/tweety53/Projects/gymie ls-files docs/design/group-sessions/screens | wc -l @ gymie main, 2026-09-10 -->

---

- [x] 1. Add `compose-mockup-frames.py`, its `.sh` wrapper and their skill symlinks

`scripts/compose-mockup-frames.py` (Python 3; Pillow is its one import beyond the standard library)
and `scripts/compose-mockup-frames.sh`, a thin wrapper shaped like `scripts/check-plan-shape.sh`:
it checks `python3` is present and runs a trivial program, checks the `.py` sibling exists by
`$SCRIPT_DIR/compose-mockup-frames.py` (so `check-guard-symlinks.sh` rule 2 sees the sibling), and
`exec`s it with the arguments unchanged. Symlink both into `skills/flow/scripts/` with relative
targets, exactly as `check-plan-shape.sh`/`.py` are — task 5 cites the wrapper by basename from
`skills/flow/verify-and-handoff.md`, which is what rule 2 requires the symlink for.

Interface, canonical in `design.md` section 4 and restated only as the usage line the script
prints:

```text verified:authored in-tree for this change
usage: compose-mockup-frames.sh <map file> <mockups root> <out dir>   (resolved PNG paths on stdin, one per line)
```

Behaviour the `.py` implements, each point one branch:

- **Pillow probe first.** `from PIL import Image` inside `try`; on `ImportError` print
  `compose-mockup-frames: Pillow is required — python3 -m pip install pillow` to stderr and exit 2,
  before any argument is read.
- **Arguments.** Exactly three, else usage to stderr and exit 2. `<map file>` unreadable, `<mockups
  root>` not a directory, or `<out dir>` not creatable (`os.makedirs(exist_ok=True)` raising) → a
  one-line reason to stderr, exit 2.
- **Stdin.** One path per line; blank lines ignored. A listed path that is not a readable PNG
  (`Image.open` raising) → reason to stderr, exit 2 — the same "unreadable PNG blocks" standing
  step 8 of the stage already gives it.
- **Map parse.** Per line: strip; skip empty and `#`-leading lines; split on whitespace; a field
  count other than 2 is a finding `<map>:<line>: expected \`<screenshot name> <frame id>\`, got
  <n> fields`. Duplicate screenshot names are a finding on the second line.
- **Match.** For a map line's `<name>` (which ends in `.png`; a name without the suffix is a
  finding), the stdin path whose basename is `<name>` or `<stem>-<anything>.png` where `<stem>` is
  `<name>` without `.png` — Playwright's `-<platform>` suffix. Zero matches → finding
  `<map>:<line>: no captured PNG matches \`<name>\``. More than one → finding naming both paths.
- **Frame.** `<mockups root>/<frame id>.png` must exist and be a regular file, else finding
  `<map>:<line>: no frame \`<frame id>.png\` under <mockups root>`.
- **Compose.** Canvas `RGB`, width `cap.w + 16 + mock.w`, height `max(cap.h, mock.h)`, background
  `(40, 40, 40)`; paste the capture at `(0, 0)` and the mockup at `(cap.w + 16, 0)` — captured
  left, mockup right, native size, top-aligned, no labels (`design.md` section 4). Both images are
  converted to `RGB` before pasting. Save as `<out dir>/<stem>.png`; print that absolute path to
  stdout.
- **Exit.** Every well-formed, matched line is composed even when another line is a finding.
  Findings print to stderr as `<map>:<line>: <message>`. Exit 1 if any finding, else 0. A map with
  zero pair lines is exit 0 with no output — an empty map is a declaration of nothing, not an
  error.

The `.py` module docstring carries the interface, the exit-code contract and the layout rule once,
citing `design.md` for why (explicit map, Pillow, left/right order) rather than restating the
decisions.

```python unverified:confirm Image.new/paste signatures against the installed Pillow 12.3.0 while writing
canvas = Image.new("RGB", (cap.width + GUTTER + mock.width, max(cap.height, mock.height)), (40, 40, 40))
canvas.paste(cap.convert("RGB"), (0, 0))
canvas.paste(mock.convert("RGB"), (cap.width + GUTTER, 0))
canvas.save(out_path)
```

**Files:** `scripts/compose-mockup-frames.py`, `scripts/compose-mockup-frames.sh`,
`skills/flow/scripts/compose-mockup-frames.py`, `skills/flow/scripts/compose-mockup-frames.sh`
**Tests:** none
**Regression:** reverting this commit removes the compositor task 5's stage step invokes and the
symlinks `check-guard-symlinks.sh` rule 2 then reports missing once task 5 lands.
**Baseline:** before=57 after=57 guard harnesses in `scripts/run-guard-tests.sh`
<!-- measured: scripts/run-guard-tests.sh @ merge-base 2f1ac1f -->
<!-- predicted: scripts/run-guard-tests.sh after this task -->
**Commit:** `feat(scripts): compose captured frames beside their mockups`
**Build:** green

Verify step:

  - [x] **Step 1: `bash -n scripts/compose-mockup-frames.sh` and `python3 -m py_compile scripts/compose-mockup-frames.py` exit 0**
  - [x] **Step 2: `scripts/check-guard-symlinks.sh` exits 0 — both symlinks resolve with relative targets**
  - [x] **Step 3: a smoke run — two fixture PNGs made with Pillow in `$TMPDIR`, a one-line map — prints one composite path and exits 0; `python3 -c 'from PIL import Image; i=Image.open("<that path>"); print(i.size)'` prints `(cap.w + 16 + mock.w, max heights)`**

- [x] 2. Add `test-compose-mockup-frames.sh`, the compositor's harness

`scripts/test-compose-mockup-frames.sh`, in the shape of `scripts/test-check-spec-reach.sh`:
`set -euo pipefail`, `fail`/`pass`, a `DIRS` array with an `EXIT` trap, `run` capturing `RC`,
`OUT` (stdout) and `ERR` (stderr) separately since the script's stdout is data. Fixture PNGs are
made by an inline `python3 - <<'PY'` using Pillow (`Image.new("RGB", (w, h), colour).save(path)`)
— the harness needs Pillow exactly as the script does, and says so on the first line of its
header; it must **fail**, never skip, when Pillow is absent, so a green run is never vacuous
(`.flow/project.md`'s "vacuous pass" rule for guards). Pinned cases, each asserting the exit code
and the relevant line:

1. Happy path: a 1200×1164 capture `j5-finish-dialog-darwin.png`, a 780×2028 frame `J5.png`, map
   `j5-finish-dialog.png J5` → exit 0, stdout is `<out>/j5-finish-dialog.png`, and the composite's
   size read back through Pillow is `1996×2028`.
   <!-- predicted: 1200 + 16 + 780 = 1996 wide, max(1164, 2028) = 2028 high, per task 1's compose rule -->
2. Exact-name match with no platform suffix (`k2-menu.png` on disk, map `k2-menu.png K2`) → exit 0.
3. Map line whose screenshot matches no stdin path → exit 1, stderr carries
   `no captured PNG matches`, and a second well-formed line in the same map is still composed
   (its path on stdout).
4. Map line naming a frame absent under the root → exit 1, stderr carries `no frame`.
5. Malformed line (three fields) → exit 1, stderr carries `expected`.
6. Comment and blank lines are ignored: a map of only `#` lines and blanks → exit 0, no output.
7. Two stdin paths matching one name (`a-darwin.png`, `a-linux.png`) → exit 1 naming both.
8. Unreadable PNG on stdin (a text file named `.png`) → exit 2.
9. Wrong argument count → exit 2 with the usage line.
10. Mockups root not a directory → exit 2.
11. Pillow absent → exit 2 and stderr carries `python3 -m pip install pillow`: run the script with
    `PYTHONPATH` pointing at a fixture dir holding `PIL/__init__.py` that does `raise
    ImportError("stubbed out by test-compose-mockup-frames.sh")`, which shadows the real package
    for that one invocation only.
12. The wrapper refuses a missing `.py` sibling with exit 2 (copy the `.sh` alone into a fixture
    dir and run it) — the same case `test-check-plan-shape.sh` pins for its wrapper.

Added by the round-0 panel-fix (F2/F4, F5, F6, F7, F8 — untested guards/behaviours the panel
found unexercised):

13. Unreadable mockup frame file (a text file named `.png` under the mockups root) → exit 2,
    stderr carries `unreadable PNG` and the frame's path.
14. Two map lines sharing the same `<screenshot name>` → exit 1, stderr carries `duplicate
    screenshot name` and the name.
15. A map line's `<screenshot name>` not ending in `.png` → exit 1, stderr carries `does not end
    in .png`.
16. Composite's gutter/background pixel reads back as `(40, 40, 40)`.
17. Mockup is pasted top-aligned: with a capture taller than its mockup, the pixel at the
    mockup's top edge is the mockup's colour and the pixel below the mockup (still within the
    capture's height) is the background colour — distinguishes top- from bottom-alignment.

`scripts/run-guard-tests.sh` discovers the harness by glob; no edit to `.flow/project.md`'s
`## test` is needed.

**Files:** `scripts/test-compose-mockup-frames.sh`
**Tests:** `test-compose-mockup-frames.sh` (cases 1–17 above)
**Regression:** reverting this commit leaves `compose-mockup-frames.py` with no harness — a
mutation to its match rule, its exit codes or its layout arithmetic would pass
`scripts/run-guard-tests.sh` unnoticed.
**Baseline:** before=57 after=58 guard harnesses in `scripts/run-guard-tests.sh`; 30 `ok:`
assertions in `test-compose-mockup-frames.sh` itself (cases 1–17, up from 12 cases /
22 assertions before the round-0 panel-fix)
<!-- measured: scripts/run-guard-tests.sh @ merge-base 2f1ac1f -->
<!-- predicted: scripts/run-guard-tests.sh after this task — one harness added -->
**Commit:** `test(scripts): pin compose-mockup-frames exit codes, matching and layout`
**Build:** green

Verify step:

  - [x] **Step 1: `bash -n scripts/test-compose-mockup-frames.sh` exits 0**
  - [x] **Step 2: `scripts/test-compose-mockup-frames.sh` prints thirty `ok:` lines and exits 0**
  - [x] **Step 3: `scripts/run-guard-tests.sh` reports 58 harnesses, with `test-compose-mockup-frames.sh` among the passes**

- [x] 3. Admit `mockups` in `check-visual-verification.sh` and pin it in its harness

In `scripts/check-visual-verification.sh`'s `check_setting_row`, add `mockups` to the closed
`Setting` vocabulary: extend the `key !=` chain and the violation message so it reads
`` is not one of `ui paths`, `screenshots`, `regression checkout`, `regression repo` or `mockups` ``.
Update the header comment's vocabulary sentence and the comment above `check_setting_row` to
name five settings. No `require_nonempty` for it — `mockups` is optional (`design.md` section 6),
and the guard validates neither that the directory exists nor any sidecar; that is the
compositor's run-time question.

In `scripts/test-check-visual-verification.sh`, append two cases after case 31, in the same shape as
cases 30/31 (`start` row accepted / optional):

- Case 32: a `| \`mockups\` | \`docs/design/screens\` |` row beside the two required settings →
  `assert_rc 0` and `assert_out_not_contains "vocabulary is closed"`.
- Case 33: the unrecognised-Setting message names all five words — reuse case 12's unknown-setting
  fixture and add `assert_out_contains "case 33" "\`mockups\`"` beside assertions for the other
  four, the way case 29 pins the `Command` list.

**Files:** `scripts/check-visual-verification.sh`, `scripts/test-check-visual-verification.sh`
**Tests:** `test-check-visual-verification.sh` (cases 32–33)
**Regression:** reverting this commit makes task 6's `mockups` row a `vocabulary is closed`
violation in gymie's `.flow/project.md`, and `flow.visual-verify`'s step 1 there reports the
section invalid.
**Baseline:** before=58 after=58 guard harnesses in `scripts/run-guard-tests.sh`
<!-- measured: scripts/run-guard-tests.sh @ merge-base 2f1ac1f, plus task 2's harness -->
<!-- predicted: scripts/run-guard-tests.sh after this task -->
**Commit:** `feat(scripts): admit the mockups setting in check-visual-verification`
**Build:** green

Verify step:

  - [x] **Step 1: `bash -n scripts/check-visual-verification.sh` exits 0**
  - [x] **Step 2: `scripts/test-check-visual-verification.sh` exits 0 with cases 32 and 33 among its `ok:` lines**
  - [x] **Step 3: `scripts/check-visual-verification.sh .` still prints `VISUAL-OK` for this repository**

- [x] 4. Document `mockups` in the `## visual verification` contract

In `skills/flow-contracts/project-configuration.md`, under `## visual verification`:

- Change "a name outside the four rows below" to "five".
- Add the settings-table row after `regression repo`:
  `| \`mockups\` | no | Directory holding one \`<frame id>.png\` per drawn frame, relative to the project root. Declared, it enables the compose step of \`flow.visual-verify\` for any capture spec carrying a \`<spec>.mockups\` sidecar; absent, the stage composes nothing and reports \`mockups: not declared\`. |`
- One paragraph after the commands table, before "**No push is ever automatic.**", stating the
  sidecar once: `<spec>.mockups` beside the capture spec, one `<screenshot name> <frame id>` line
  per pair, `#` comments and blank lines ignored, `<screenshot name>` the `toHaveScreenshot()`
  name with Playwright's `-<platform>` suffix tolerated on match; the composite is captured-left,
  mockup-right at native size. Cite `<agents repo>/scripts/compose-mockup-frames.sh`'s header as
  canonical for the exit codes, the way the section already cites
  `check-visual-verification.sh` for its shape.

The stage's own procedure (what blocks, where composites land) is task 5's text in
`verify-and-handoff.md`, cited from here by name, not restated.

**Files:** `skills/flow-contracts/project-configuration.md`
**Tests:** none
**Regression:** reverting this commit leaves the guard (task 3) accepting a setting the contract
that claims to be canonical for the section does not define.
**Baseline:** before=58 after=58 guard harnesses in `scripts/run-guard-tests.sh`
<!-- measured: scripts/run-guard-tests.sh @ merge-base 2f1ac1f, plus task 2's harness -->
<!-- predicted: scripts/run-guard-tests.sh after this task -->
**Commit:** `docs(flow-contracts): define the mockups setting for visual verification`
**Build:** green

Verify step:

  - [x] **Step 1: `scripts/check-vocabulary.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.py` exit 0**
  - [x] **Step 2: `scripts/check-contract-budget.sh` exits 0 with no `budgets()` edit — the file stays under 61302 bytes**

- [x] 5. Add the compose step, report lines and blocking rule to `flow.visual-verify`

In `skills/flow/verify-and-handoff.md`, `## Visual verification`:

- The verifier-prompt sentence ("Its prompt states: … `screenshots` root") also states the
  resolved `mockups` root when declared.
- Renumber: today's step 9 becomes 10, 10 becomes 11, 11 becomes 12, and the new step 9 is:

> 9. **Compose captured frames against their mockups, if `mockups` is declared.** With a
>    `<spec>.mockups` sidecar beside the capture spec, run
>
>    ```bash verified:authored in-tree for this change — resolve-visual-screenshots.sh's usage line and task 1's interface
>    resolve-visual-screenshots.sh <worktree> <spec's basename> | compose-mockup-frames.sh <spec>.mockups <worktree>/<mockups> <changeRoot>/visual-verification
>    ```
>
>    Exit 0 prints one composite path per line — **read every one** and state, per pair, whether
>    the capture matches the frame and where it departs. Exit 1 (a broken map — a screenshot no
>    capture matched, a frame file absent, a malformed line) or 2 (cannot answer, Pillow absent
>    included) blocks. No sidecar → report `mockups: no map for <spec's basename>` and continue.
>    Not declared → report `mockups: not declared` and continue. The sidecar's shape is canonical
>    in **visual verification** (`skills/flow-contracts/project-configuration.md`).

- New step 10 (was 9) adds: "and, per composed pair, the composite's absolute path, the frame id,
  and what was seen".
- New step 11 (was 10) adds: "`<changeRoot>/visual-verification/` is committed with the change root."
- The `## Report` block gains, after the `- <view>:` line:
  `- mockups: <not declared | no map for <spec> | exit <n>>` and
  `- <frame id>: <absolute composite path> — <match, or the departure seen>`.
- **Blocking** gains "a `compose-mockup-frames.sh` exit 1 or 2, and a departure from the mockup the
  verifier reports in a composite" beside the existing captured-screenshot defect clause.

Raise `skills/flow/verify-and-handoff.md`'s row in `scripts/check-contract-budget.sh`'s
`budgets()` to the file's new size plus 25%, per `.flow/project.md`'s "Raising a budget is the
correct response to a genuine addition" — measure with `wc -c` after the edit and record the
figure in the commit body.

**Files:** `skills/flow/verify-and-handoff.md`, `scripts/check-contract-budget.sh`
**Tests:** none
**Regression:** reverting this commit leaves the compositor (task 1) and the setting (tasks 3–4)
with no stage that runs them — the comparison stays manual, which is the defect KAN-449 names.
**Baseline:** before=58 after=58 guard harnesses in `scripts/run-guard-tests.sh`
<!-- measured: scripts/run-guard-tests.sh @ merge-base 2f1ac1f, plus task 2's harness -->
<!-- predicted: scripts/run-guard-tests.sh after this task -->
**Commit:** `feat(flow): compose captured frames against mockups in flow.visual-verify`
**Build:** green

Verify step:

  - [x] **Step 1: `scripts/check-guard-symlinks.sh` exits 0 — `compose-mockup-frames.sh`, now invoked from the skill's text, resolves through task 1's symlink**
  - [x] **Step 2: `scripts/check-references.sh`, `scripts/check-vocabulary.sh`, `scripts/check-stage-mark-calls.sh`, `scripts/check-dispatch-paragraphs.sh` and `scripts/check-markdown-integrity.py` exit 0**
  - [x] **Step 3: `scripts/check-contract-budget.sh` exits 0 against the raised row**

- [x] 6. Declare `mockups` in gymie's `.flow/project.md`

In `/Users/tweety53/Projects/gymie/.flow/project.md`, `## visual verification`, add the settings
row after `regression repo`: `| \`mockups\` | \`docs/design/group-sessions/screens\` |`, and one
sentence in the section's prose naming the sidecar convention for that checkout: a per-change
capture spec at the `gymie-playwright` root carries `<spec>.mockups` beside it. This commit lands
on a branch in the `gymie` repository, not in this one; the implementer commits it there and
prints the branch for the operator, pushing nothing.

**Files:** `/Users/tweety53/Projects/gymie/.flow/project.md`
**Tests:** none
**Regression:** reverting this commit leaves the one project with drawn frames declaring no
`mockups` root, so its `flow.visual-verify` reports `mockups: not declared` on every run.
**Baseline:** before=58 after=58 guard harnesses in `scripts/run-guard-tests.sh`
<!-- measured: scripts/run-guard-tests.sh @ merge-base 2f1ac1f, plus task 2's harness -->
<!-- predicted: scripts/run-guard-tests.sh after this task — this task adds no harness -->
**Commit:** `chore(flow): declare the group-sessions mockup frames for visual verification`
**Build:** green
**After:** Task 3

Verify step:

  - [x] **Step 1: `scripts/check-visual-verification.sh /Users/tweety53/Projects/gymie` prints `VISUAL-OK` with the guard from this branch**
  - [x] **Step 2: `ls /Users/tweety53/Projects/gymie/docs/design/group-sessions/screens/J5.png` — the declared root resolves relative to the gymie root**
