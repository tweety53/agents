# kan-850-agents-port-the-visual-verify-scripts-to-the-go

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Ports the five scripts `flow.visual-verify` runs into `flow-guard`: the shared foundation (task 1),
the three bash ports (tasks 2–4), the in-process trigger (task 5), the PNG decode (task 6), the two
Python ports (tasks 7–8), the deletion and citation sweep (task 9), and the before/after
measurement (task 10). `design.md` is canonical for every decision; each task cites its entry by
ID.

**The script at `3915fbc0` is each port's specification.** Its header comment (for the Python
scripts, the module docstring) states the contract; its body is the behaviour. A port reproduces
arguments, environment overrides, every output line byte for byte, the stdout/stderr split, every
side effect on the filesystem, and every exit code — except where `design.md` names a departure
(`any-channel-diff-mask`, `png-rgb-decode`'s 16-bit rule, `exact-option-names`,
`python-io-semantics`' tracebacks, `bytewise-screenshot-sort`). Where the source and its header
disagree anywhere else, stop and report — never pick one silently.

**Parity floors** are in `design.md`'s **Measurements** (**Decision:** carry-prior-port-decisions —
KAN-760's `parity-by-case-count`).

**Every port task (2–4, 7–8) follows the same shape:** port the harness's cases to a Go table test
first (red — the guard is unregistered), port the script, run the test green, replace the script's
body with the shim, `git rm` the harness (and the `.py`). Each Go subtest is named after the
harness case's `ok:` label, so parity is a `--- PASS` count. The script body's comments move into
the Go file beside the code they explain, updated where the mechanism changed (KAN-760's
`flow-guard-binary-rationale-in-go`). The Go file is
`stats/internal/guard/<name without - and without check->.go`; the test file is
`stats/internal/guard/<name with - replaced by _>_test.go` — the name `scripts/run-guard-tests.sh`'s
companion rule accepts for a shim. Each port registers its basename in `guard.Registry` from its
own file's `init()`, as `basemoved.go` does. The guard signature is
`func(args []string, env Env, stdout, stderr io.Writer) int`.

**Shim template** — the header comment block kept, corrected only where the port makes a
statement false (`$SCRIPT_DIR/<lib>` sourcing, bash/Python plumbing, Pillow), then:

```bash verified:copied from the tail of scripts/check-references.sh @ 3915fbc0
set -euo pipefail
. "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "<name>: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec <name> 2 "<name>:" "$@"
```

`<name>` is the basename without `.sh`, written literally; the cannot-answer code is 2 for all five
(`design.md` **Context**). The two `FLOW_GUARD_REPO_ROOT` lines are kept only where the Go guard
reads its own checkout; none of the five ports does once the siblings are in Go, so each shim
drops them. For `compose-mockup-frames.sh` and `measure-visual-properties.sh` the header becomes the
`.py` docstring's interface, layout/property and exit-code paragraphs — `project-configuration.md`
cites the `.sh` header as canonical for compose's exit codes — with "Pillow absent" removed from
exit 2; the docstring's algorithm prose moves into the Go file.

**Test isolation:** every Go test calls `t.Parallel()`; fixture trees are built once in a
`sync.Once` helper (`helpers_test.go` has `writeFile`, `mkdir`, `gitRun`, `fixtureRoot`) and each
case copies its own into `t.TempDir()`; stdin is `Env.Stdin` (task 1), never `os.Stdin`; `git` is
the guard package's existing git helper. No test reads or writes the operator's real `$HOME`.

**Per-task verify** is `cd stats && gofmt -l . && go vet ./internal/guard/ && go test
./internal/guard/ -run '^<Test>$' -count=1 -race`, then the shimmed script run once for real,
through a symlink in a temp directory, as its last step names. The full package and
`scripts/run-guard-tests.sh` run in task 10 and `flow.verify`.

Live verification: task 10 runs the real suite on this machine before and after. No running
service or persistent state is touched — the ports are CLI tools over files.

## Review Focus

- **The shim run from an installed skill directory** (`~/.claude/skills/flow/scripts/<name>.sh`, a
  symlink into this checkout) must resolve and build the same `flow-guard` the bash/Python ran
  beside — each port's last step runs the shim through a symlink in a temp directory.
- **Path spelling.** `compose-mockup-frames` prints `os.path.abspath` paths (physical cwd);
  `resolve-visual-screenshots` prints `cd … && pwd` paths (logical — `$PWD`-relative) plus
  `find`'s joins; `check-visual-trigger` prints `$ROOT` as given. Each port carries a subtest run
  from a cwd reached through a symlink (`t.TempDir()` on macOS is under `/var` → `/private/var`)
  pinning the spelling the source printed.
- **Python number types in `measure-visual-properties`' JSON** — `peak_delta` is the int `0` with
  no shadow band but a float otherwise; `a_scaled` stays an int under an unscaled key with an int
  `a`; `round(x)` returns an int, `round(x, n)` a float; `-0.0` keeps its sign. Task 8's goldens
  cover every property in both single-image and paired form so a wrong type fails a byte compare.
- **Attacker-influenced `.flow/project.md`** — BOM, CRLF, `\|` inside a cell, control bytes in a
  cell that reaches an error line: the ported cases carry them, and task 1's parity tests pin the
  twins on the same inputs, a NUL byte included (`ccSanitize` escapes it; the awk loop starts at 1).
- **Python text input quirks in `compose-mockup-frames`** — a map line separated by `\x1c`, a stdin
  list with a lone `\r` between paths, CRLF, and a map file that is not UTF-8 (`python-io-semantics`)
  — each a subtest in task 7.

---

- [x] 1. Env.Stdin, the visual section reader and the Go twins of strip-bom, sanitize-display and visual-table-cells

**Files:** `stats/internal/guard/guard.go`, `stats/cmd/flow-guard/main.go`, `stats/internal/guard/visualsection.go`, `stats/internal/guard/libtwins_test.go`, `stats/internal/guard/visual_section_test.go`
**Tests:** `TestStripBOMParity`, `TestSanitizeDisplayParity`, `TestVisualTableCellsParity`, `TestVisualSection`
**Regression:** each parity test fails if its Go twin's output differs from `scripts/lib/strip-bom.sh`,
`scripts/lib/sanitize-display.sh` or `scripts/lib/visual-table-cells.awk` on the same input;
`TestVisualSection` fails if the heading count, the section's line set or a trimmed glob element
changes.
**Baseline:** before=5 after=9
<!-- measured: cat stats/internal/guard/libtwins_test.go stats/internal/guard/visual_section_test.go 2>/dev/null | grep -cE '^func Test' @ 3915fbc0 -->
**After:** none
**Commit:** `feat(stats): add visual section reader and Go twins of the visual table helpers`
**Build:** green

**Decision:** kan850-helper-twins

**Decision:** env-stdin

  - [x] **Step 1: Env.Stdin.** Add `Stdin io.Reader // os.Stdin in production` to `guard.Env`
    (`stats/internal/guard/guard.go`) and pass `Stdin: os.Stdin` in
    `stats/cmd/flow-guard/main.go`'s `guard.Env{…}` literal. A nil `Stdin` reads as empty input.
  - [x] **Step 2: Failing parity tests** in `libtwins_test.go`, reusing its `bashLib`/`assertParity`
    helpers (the `.awk` library runs as `awk -f scripts/lib/visual-table-cells.awk -f <driver>`
    with a driver that prints each function's result per input line):
    - `TestStripBOMParity` — a file with a leading UTF-8 BOM, one with a BOM not at byte 0, one
      of two bytes `\xef\xbb`, an empty file, a missing file; compare the bytes `strip_bom_cat`
      prints with `stripBOM`'s.
    - `TestSanitizeDisplayParity` — every byte `0x00`–`0x1f`, `0x7f`, `\`, a multi-byte UTF-8
      run, a line ending in `\r`; compare `sanitize_display`'s stdout with the twin's. Use
      `ccSanitize` if this passes; otherwise add `sanitizeDisplay` to `visualsection.go` and leave
      `ccSanitize` as it is.
    - `TestVisualTableCellsParity` — `split_cells`/`trimcell`/`foldcell` over: a two-cell row, a
      row without the trailing `|`, `\|` and `\\` inside a cell, a lone `\` before a letter, CRLF,
      leading/trailing spaces and backticks, interior runs of spaces and tabs, a line not starting
      with `|`, an empty trailing cell, mixed case.
    Run `cd stats && go test ./internal/guard/ -run 'Parity$' -count=1` — expect compile failure
    (twins undefined).
  - [x] **Step 3: Twins.** In `visualsection.go`: `stripBOM(b []byte) []byte`,
    `vtSplitCells(line string) []string`, `vtTrimCell(c string) string`, `vtFoldCell(c string)
    string`, and `trimGlobElement(s string) string` (`scripts/lib/trim-glob-element.sh`'s
    `sub(/^[ \t`]+/,""); sub(/[ \t`]+$/,"")`, LC_ALL=C). Each library header's reasoning moves into
    the Go comment above its twin. `[[:space:]]` in the awk is the C locale's set (space, `\t`,
    `\n`, `\v`, `\f`, `\r`); `tolower` is ASCII-only under the libraries' callers.
    `measured: the callers set no LC_ALL/LANG, so awk/grep inherit the invoker's locale; under en_US.UTF-8 (this machine's LANG) tolower folds non-ASCII (Ä→ä) and [[:space:]] matches NBSP, under C neither — the parity test pins LC_ALL=C (the plan's stated semantics), and a non-ASCII row fails it under en_US.UTF-8: grep -n 'LC_ALL\|LANG' scripts/check-visual-{trigger,verification}.sh scripts/resolve-visual-screenshots.sh; printf 'Ä\xc2\xa0x\n' | LC_ALL=en_US.UTF-8 awk '{s=tolower($0); gsub(/[[:space:]]+/,"_",s); print s}' @ 3915fbc0`
  Correction (2026-09-28): the plan declared `ccSanitize` as the sanitize twin if it passed parity; it
  failed (the one-true-awk ends a record at NUL, so `a\x00b` prints `a`), so `sanitizeDisplay`
  (truncate at NUL, then `ccSanitize`, then `\n`) was added to `visualsection.go` and `ccSanitize` left
  unchanged. Step 3's "`tolower` is ASCII-only under the libraries' callers" holds only under a C
  locale: the bash callers inherited the invoker's locale; the twins pin C-locale semantics
  (**Decision:** c-locale-table-semantics). The heading match lowercases byte-wise (`gsASCIILower`)
  instead of `(?i)`, whose Unicode folding accepts `ſ`/Kelvin.

  - [x] **Step 4: Section reader, test first.** `TestVisualSection` in `visual_section_test.go`,
    then in `visualsection.go`: `vvHeadingCount(text string) int` — lines matching
    `(?i)^##[[:space:]]+visual verification[[:space:]]*$` after `stripBOM`; and
    `vvSectionLines(text string) []vvLine` (`vvLine{No int; Text string}`, `No` 1-based) — the
    lines inside the section under the scripts' shared awk rule: a `#+[[:space:]]` heading line is
    never a section line; the own heading opens the section; any other heading closes it unless it
    is deeper than level 2 while inside. Cases: section at file end, closed by `## next`, `### sub`
    inside it, `# top` closing it, heading case and trailing spaces, CRLF, BOM, two sections (count
    2), no section. `trimGlobElement`: backticks and spaces on both sides, an interior space kept,
    all-backtick input to empty.
  - [x] **Step 5: Green and lint.** `cd stats && gofmt -l . && go vet ./... && go test
    ./internal/guard/ -run 'Parity$|^TestVisualSection$' -count=1 -race` and `go test
    ./cmd/flow-guard/ -count=1`.

- [x] 2. Port check-visual-trigger

**Files:** `stats/internal/guard/visualtrigger.go`, `stats/internal/guard/check_visual_trigger_test.go`, `scripts/check-visual-trigger.sh`, `scripts/test-check-visual-trigger.sh`
**Tests:** `TestCheckVisualTrigger`
**Regression:** fails if any of the harness's 51 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_visual_trigger_test.go 2>/dev/null | grep -cE '^func Test' @ 3915fbc0 -->
**After:** Task 1
**Commit:** `feat(stats): port check-visual-trigger to Go`
**Build:** green

**Decision:** scope-visual-verify-scripts

**Decision:** env-stdin

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-check-visual-trigger.sh`, one
    subtest per `ok:` label, stdin through `Env.Stdin`. Add one subtest run with `Env.Dir` reached
    through a symlink and a relative root, pinning that the verdict line prints `$ROOT` as given.
    Run — expect failure.
  - [x] **Step 2: Port**, registering `check-visual-trigger`, exported for task 5 as
    `visualTrigger(env Env, root string, changed []string, stdout, stderr io.Writer) int` (a relative root resolved against `env.Dir`) beside the
    registered wrapper (which reads `Env.Stdin` line by line, a final line without `\n` included,
    as `read -r … || [ -n … ]` does). Uses task 1's `vvHeadingCount`, `vvSectionLines`,
    `vtSplitCells`, `vtFoldCell`, `vtTrimCell`, `trimGlobElement`, the sanitize twin. The glob
    translation is the bash `glob_to_ere`: `**` → `.*`, `*` → `[^/]*`, `?` → `[^/]`, the listed
    metacharacters escaped, anchored `^…$`; a leading `./` stripped from both sides and a leading
    `/` from the glob; each `MATCH:` line goes through the sanitize twin.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckVisualTrigger$' -count=1
    -race -v | grep -c -- '--- PASS: TestCheckVisualTrigger/'` — at least 51.
  - [x] **Step 4: Shim and delete** — shim template, no `FLOW_GUARD_REPO_ROOT`; `git rm
    scripts/test-check-visual-trigger.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`; `printf
    'stats/web/src/App.tsx\n' | scripts/check-visual-trigger.sh .` prints the line it printed at
    `3915fbc0` and exits 0; the same through a symlink to `scripts/check-visual-trigger.sh` in a
    temp directory.

  Correction (2026-09-28): Step 5's "through a symlink in a temp directory" is a temp directory
  holding a symlink to `scripts/check-visual-trigger.sh` beside a `lib` symlink to `scripts/lib` —
  the shape `skills/flow/scripts/` installs; a lone script symlink exits 2 at base and in the shim
  alike. The ported subtests number 71: the harness's 51 labels plus 20 for behaviour it never
  reached. A read failure of `.flow/project.md` after the access check (a race) prints the "grep
  exited 2" refusal where the bash said "declares no section"; both exit 2.

- [x] 3. Port check-visual-verification

**Files:** `stats/internal/guard/visualverification.go`, `stats/internal/guard/check_visual_verification_test.go`, `scripts/check-visual-verification.sh`, `scripts/test-check-visual-verification.sh`
**Tests:** `TestCheckVisualVerification`
**Regression:** fails if any of the harness's 111 `ok:` behaviours regress.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/check_visual_verification_test.go 2>/dev/null | grep -cE '^func Test' @ 3915fbc0 -->
**After:** Task 1
**Commit:** `feat(stats): port check-visual-verification to Go`
**Build:** green

**Decision:** scope-visual-verify-scripts

**Decision:** kan850-helper-twins

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-check-visual-verification.sh`,
    one subtest per `ok:` label. The harness's stubbed `git` (the `git remote get-url origin`
    failure that is not "no such remote") becomes a stub on `PATH` for that case only, or the
    package's injected git hook if one exists.
    `unverified: check whether the guard package's git helper can be stubbed in-process; a PATH stub is the fallback, as KAN-842's cases did before dispatches-via-env-hook`
    Run — expect failure.
  - [x] **Step 2: Port**, registering `check-visual-verification`: the settings and commands table
    checks, closed vocabularies, the `regression checkout`/`regression repo` identity assertion, the
    `file:line: <message>` violation lines through the sanitize twin, the `VISUAL-OK`/
    `VISUAL-INVALID` verdicts. `lib/git-clean.sh`'s `git_clean` (unset every `GIT_*` variable before
    `git`) is ported here, its header's reasoning with it — through the package's existing git
    helper if it already strips `GIT_*`.
    `unverified: read envGit in the guard package to see whether it strips GIT_* already`
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckVisualVerification$'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckVisualVerification/'` — at least 111.
  - [x] **Step 4: Shim and delete** — shim template, no `FLOW_GUARD_REPO_ROOT`; `git rm
    scripts/test-check-visual-verification.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-visual-verification.sh .` prints the `VISUAL-OK` line it printed at `3915fbc0`
    and exits 0; the same through a symlink in a temp directory.

  Correction (2026-09-28): `envGit` does not strip `GIT_*`; `vvGitClean` builds the command through
  `envGit` and drops every `GIT_*` variable. A PATH stub through `pathEnv` stubs git in-process.
  Fixtures are built per subtest in `t.TempDir()` rather than one `sync.Once` tree; case 22 runs the
  built `flow-guard` as a subprocess with `GIT_DIR` set, since `t.Setenv` cannot combine with
  `t.Parallel()`. Step 5's symlink run uses the installed shape (a `lib` symlink beside the shim).

- [x] 4. Port resolve-visual-screenshots

**Files:** `stats/internal/guard/resolvevisualscreenshots.go`, `stats/internal/guard/resolve_visual_screenshots_test.go`, `scripts/resolve-visual-screenshots.sh`, `scripts/test-resolve-visual-screenshots.sh`
**Tests:** `TestResolveVisualScreenshots`
**Regression:** fails if any of the harness's 38 `ok:` behaviours regress, or if the output order
stops being bytewise.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/resolve_visual_screenshots_test.go 2>/dev/null | grep -cE '^func Test' @ 3915fbc0 -->
**After:** Task 1
**Commit:** `feat(stats): port resolve-visual-screenshots to Go`
**Build:** green

**Decision:** scope-visual-verify-scripts

**Decision:** bytewise-screenshot-sort

  - [x] **Step 1: Failing test.** Port every case of `scripts/test-resolve-visual-screenshots.sh`,
    one subtest per `ok:` label, plus: matches named `B.png`, `a.png`, `_x.png`, `A-1.png` print in
    byte order; a relative root from a cwd reached through a symlink prints the logical
    (`$PWD`-joined) prefix `cd … && pwd` printed. Run — expect failure.
  - [x] **Step 2: Port**, registering `resolve-visual-screenshots`: the `screenshots` and
    `regression checkout` rows (first of each), the regression checkout's `git worktree list
    --porcelain` lookup of this root's branch, then a walk of `<base>/<screenshots>` that prunes
    every directory named `.worktrees`, keeps regular `*.png` files only (symlinks neither
    followed nor matched, as `find -type f` without `-L`), and keeps a path when any `/`-separated
    segment of it starts with the spec basename; sorted bytewise; exit 1 with the "zero PNGs"
    line when none.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestResolveVisualScreenshots$'
    -count=1 -race -v | grep -c -- '--- PASS: TestResolveVisualScreenshots/'` — at least 38.
  - [x] **Step 4: Shim and delete** — shim template, no `FLOW_GUARD_REPO_ROOT`; `git rm
    scripts/test-resolve-visual-screenshots.sh`.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/resolve-visual-screenshots.sh . baseline.spec.ts` prints and exits as it did at
    `3915fbc0`; the same through a symlink in a temp directory.

  Correction (2026-09-28): Step 5's symlink run uses a temp directory holding symlinks to the shim
  and to `scripts/lib`, the shape `skills/flow/scripts/` installs. On an unreadable subdirectory the
  port prints `resolve-visual-screenshots: open <dir>: permission denied` where `find` printed
  `find: <dir>: Permission denied`; exit code and stdout match (**Decision:**
  prefixed-walk-errors). A path containing a newline is matched and printed whole, where the bash
  split it. Case 17 uses a second project repository. The subtests number 41: the harness's 38 plus
  byte order, a symlinked cwd and unfollowed symlinks.

- [x] 5. Evaluate the trigger in-process in check-visual-verify-dispatched

**Files:** `stats/internal/guard/visualverifydispatched.go`, `stats/internal/guard/check_visual_verify_dispatched_test.go`, `scripts/check-visual-verify-dispatched.sh`
**Tests:** `TestCheckVisualVerifyDispatchedInProcessTrigger`
**Regression:** fails if the guard execs a sibling again — a checkout with no
`scripts/check-visual-trigger.sh` must still answer all three trigger outcomes (`not configured`,
`no UI paths touched`, the dispatch check).
**Baseline:** before=1 after=2
<!-- measured: cat stats/internal/guard/check_visual_verify_dispatched_test.go 2>/dev/null | grep -cE '^func Test' @ 3915fbc0 -->
**After:** Task 2
**Commit:** `refactor(stats): evaluate the visual trigger in-process in check-visual-verify-dispatched`
**Build:** green

**Decision:** trigger-in-process

  - [x] **Step 1: Test first.** In `check_visual_verify_dispatched_test.go`, delete the
    `TestCheckVisualVerifyDispatched` cases that exist only for the exec — "required sibling module
    not found or not executable" (both), "exited 3, neither 0, 1 nor 2", and "FLOW_GUARD_SELF unset"
    if `FLOW_GUARD_SELF` has no other reader — and add `TestCheckVisualVerifyDispatchedInProcessTrigger`:
    a checkout with no `scripts/check-visual-trigger.sh` still answers `not configured`, `no UI paths
    touched` and a dispatch verdict, one subtest each. Run — expect failure.
  - [x] **Step 2: Port.** Replace the `exec.Command(trigger, worktree)` block with
    `visualTrigger(env, worktree, strings.Split(changed, "\n"), io.Discard, &trigErr)` (the diff's paths,
    an empty diff one empty line as before); drop `guardSelfDir` here if nothing else in the guard
    needs it. The header paragraph "THE TRIGGER QUESTION IS DELEGATED" and the Go comments are
    corrected to say it is evaluated in-process.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestCheckVisualVerifyDispatched'
    -count=1 -race -v | grep -c -- '--- PASS: TestCheckVisualVerifyDispatched/'` — at least 18
    (its KAN-842 floor); record the count.
  - [x] **Step 4: Shim.** Drop `FLOW_GUARD_SELF` and its `$SCRIPT_DIR` sibling comment from
    `scripts/check-visual-verify-dispatched.sh` when the Go no longer reads it; correct the header.
  - [x] **Step 5: Verify.** `gofmt -l`, `go vet ./internal/guard/`;
    `scripts/check-visual-verify-dispatched.sh` with no arguments exits 2 with the line it printed
    at `3915fbc0`; `scripts/check-guard-symlinks.sh` exits 0.

  Correction (2026-09-28): Step 2's call is `visualTrigger(env, worktree, func() []string { return
  strings.Split(changed, "\n") }, io.Discard, io.Discard)` — `changed` a func since task 2's fix, so the
  shim reads stdin only after every refusal — the "exited N" branch was `trigErr`'s only reader. Step 1
  also deleted "the trigger reads the diff's paths on stdin and the worktree as its argument", a
  case that tested only the exec'd stub. `guardSelfDir` stays (`taskreviewersingledispatch.go`
  reads it); the shim dropped `FLOW_GUARD_SELF` and `SCRIPT_DIR`.

- [x] 6. PNG-to-RGB decode

**Files:** `stats/internal/guard/pngrgb.go`, `stats/internal/guard/png_rgb_test.go`
**Tests:** `TestDecodeRGB`
**Regression:** fails if any colour type or bit depth decodes to different RGB than Pillow's
`convert("RGB")` (16-bit: the high byte).
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/png_rgb_test.go 2>/dev/null | grep -cE '^func Test' @ 3915fbc0 -->
**After:** none
**Commit:** `feat(stats): add a PNG decode with Pillow's convert("RGB") semantics`
**Build:** green

**Decision:** png-rgb-decode

  - [x] **Step 1: Failing test.** `TestDecodeRGB`, table-driven, each row a PNG encoded in the test
    with `image/png` from a Go image of one type, decoded, compared pixel by pixel with the
    expected RGB: `image.NRGBA` with alpha 0 and 128 (colour kept, never premultiplied),
    `image.RGBA` opaque, `image.Gray`, `image.Gray16` (high byte), `image.NRGBA64` (high bytes),
    `image.Paletted` with an opaque palette and with a translucent `color.NRGBA` entry (entry's
    RGB), and a 1-bit grey PNG (0/255) built from raw bytes. Plus: a truncated file and a non-PNG
    file return an error. Run — expect failure.
  - [x] **Step 2: Decode.** `type rgbImage struct{ W, H int; Pix []byte }` (3 bytes per pixel,
    row-major) with `at(x, y int) [3]byte`; `decodeRGB(path string) (*rgbImage, error)` —
    `png.Decode`, then per concrete type read the stored channels (`*image.NRGBA`/`*image.RGBA`
    Pix directly; `*image.Gray`; `*image.Gray16`/`*image.NRGBA64`/`*image.RGBA64` high byte;
    `*image.Paletted` via `color.NRGBAModel` of each entry), falling back to
    `color.NRGBAModel.Convert` per pixel for any other type; `encodeRGB(path string, im
    *rgbImage) error` writing an `*image.RGBA` with alpha 255.
  - [x] **Step 3: Green.** `go test ./internal/guard/ -run '^TestDecodeRGB$' -count=1 -race`.
  - [x] **Step 4: Verify.** `gofmt -l`, `go vet ./internal/guard/`.

- [x] 7. Port compose-mockup-frames

**Files:** `stats/internal/guard/composemockupframes.go`, `stats/internal/guard/compose_mockup_frames_test.go`, `scripts/compose-mockup-frames.sh`, `scripts/compose-mockup-frames.py`, `scripts/test-compose-mockup-frames.sh`, `skills/flow/scripts/compose-mockup-frames.py`
**Allowed-collateral:** `stats/internal/guard/testdata/compose-mockup-frames/**`
**Tests:** `TestComposeMockupFrames`
**Regression:** fails if any of the harness's 59 `ok:` behaviours regress (case 11's Pillow-absent
check excepted, `design.md` **Context**), if any composite or `.frame.png` departs from the Python's
golden pixels outside the difference panel, or if a one-level red or blue difference stops
counting.
**Baseline:** before=0 after=1
<!-- measured: cat stats/internal/guard/compose_mockup_frames_test.go 2>/dev/null | grep -cE '^func Test' @ 3915fbc0 -->
**After:** Task 1, 6
**Commit:** `feat(stats): port compose-mockup-frames to Go`
**Build:** green

**Decision:** any-channel-diff-mask

**Decision:** python-goldens-as-testdata

**Decision:** python-io-semantics

**Decision:** env-stdin

  - [x] **Step 1: Goldens, before any Go.** With the Python still in place: extract each
    `python3 - <<'PY'` fixture snippet of `scripts/test-compose-mockup-frames.sh` into a scratch
    generator (not committed) writing its PNGs, map files and mockup roots to
    `stats/internal/guard/testdata/compose-mockup-frames/<case>/`; run
    `scripts/compose-mockup-frames.sh` for each harness invocation there and keep its stdout,
    stderr, exit code and every PNG it wrote as `<case>/golden/`. For each composite with geometry,
    compute in the generator whether Pillow's luma mask equals an any-channel mask of the same
    pair, and list the cases where it does not in `design.md` **Measurements** (expected: none;
    a hit means that case's diff panel is compared against the any-channel mask in step 2).
    `measured: no harness fixture carries a sub-luma difference — ImageChops.difference(cap, frame).convert("L") mask equals the any-channel mask on all 8 harness three-panel composites (case-20, 21×2, 22, 23a, 23b, 26, 27), Pillow 12.3.0 @ 3915fbc0`
  - [x] **Step 2: Failing test.** Port every case, one subtest per `ok:` label, reading the
    committed fixtures; each composed case additionally compares every written PNG with its golden
    pixel by pixel (via `decodeRGB`) and stdout lines with the golden's, output paths compared
    after replacing the golden run's directory with the test's. Case 11 (Pillow absent) has no
    counterpart. New subtests: a pair differing by `(1,0,0)` in one pixel reports
    `diff=<1/(w·h) to 4 decimals>` and a white pixel there; a map line split by `\x1c`; stdin with
    a lone `\r` between two paths and with CRLF; a map file that is not UTF-8 exits 2 with one
    `compose-mockup-frames: …` line; a relative output directory from a cwd reached through a
    symlink prints the physical absolute path (`syscall.Getwd`). Run — expect failure.
  - [x] **Step 3: Port**, registering `compose-mockup-frames`: usage (4 or 5 args; geometry
    `scale=… status=… border=…`, each key once, non-negative ASCII-digit values, `scale ≥ 1`), map
    read (UTF-8, universal newlines, Python whitespace for `strip`/`split`), stdin paths
    (`Env.Stdin`, same rules), every stdin PNG decoded up front (exit 2 on failure), per map line
    the findings in the Python's order and wording, `match_capture` with the eight platform
    suffixes, `crop_box`'s two bottom rules, the three-panel and two-panel layouts (16px gutter,
    background `(40,40,40)`), the any-channel difference panel and ratio (`%.4f`), outputs on
    stdout then findings on stderr, exit 1 if any finding. Uses `decodeRGB`/`encodeRGB`.
  - [x] **Step 4: Green.** `go test ./internal/guard/ -run '^TestComposeMockupFrames$' -count=1
    -race -v | grep -c -- '--- PASS: TestComposeMockupFrames/'` — at least 59.
  - [x] **Step 5: Shim and delete** — `scripts/compose-mockup-frames.sh` becomes the shim with the
    docstring-derived header (see the shim template note above); `git rm
    scripts/compose-mockup-frames.py scripts/test-compose-mockup-frames.sh`.
  - [x] **Step 6: Verify.** `gofmt -l`, `go vet ./internal/guard/`; the shim, through a symlink in
    a temp directory, composes one committed fixture pair and prints the golden's line.

  Correction (2026-09-28): `**Files:**` widened by `skills/flow/scripts/compose-mockup-frames.py`,
  the installed symlink the `.py`'s deletion left dangling (`check-guard-symlinks.sh` rule 1). Step 3's
  stdin "same rules" is wrong: CPython's `sys.stdin` is `newline="\n"` on POSIX, so stdin splits on
  `\n` only and a lone `\r` stays inside a path, while the map file splits with universal newlines —
  measured: `printf 'a.png\rb.png\r\nc.png\n' | python3 -c 'import sys; print([l for l in
  sys.stdin])'` gives `['a.png\rb.png\r\n', 'c.png\n']`, Python 3.14.6 @ 3915fbc0; Step 2's lone-`\r`
  subtest pins it as one path (exit 2, `unreadable PNG on stdin`). The physical cwd is
  `filepath.EvalSymlinks(env.Dir)` rather than `syscall.Getwd`, the guard running in-process. Case 12
  runs a lone copy of the shim (no `lib/flow-guard.sh` beside it). The subtests number 64: the
  harness's 59 less case 11's 2, plus 7; task 7's review fix added a one-level blue difference to
  the `any-channel` fixture (`diff=0.0010`) and a `stdin-nul` fixture, for 66.

- [x] 8. Port measure-visual-properties

**Files:** `stats/internal/guard/measurevisualproperties.go`, `stats/internal/guard/pyjson.go`, `stats/internal/guard/measure_visual_properties_test.go`, `scripts/measure-visual-properties.sh`, `scripts/measure-visual-properties.py`, `scripts/test-measure-visual-properties.sh`, `skills/flow/scripts/measure-visual-properties.py`
**Allowed-collateral:** `stats/internal/guard/testdata/measure-visual-properties/**`
**Tests:** `TestMeasureVisualProperties`, `TestPyJSON`
**Regression:** `TestMeasureVisualProperties` fails if any of the harness's 11 `ok:` behaviours
regress or any golden's stdout differs by a byte; `TestPyJSON` fails if a float, an int, `None`,
`-0.0` or key order prints differently from Python's `json.dump(indent=2)`.
**Baseline:** before=0 after=2
<!-- measured: cat stats/internal/guard/measure_visual_properties_test.go 2>/dev/null | grep -cE '^func Test' @ 3915fbc0 -->
**After:** Task 6
**Commit:** `feat(stats): port measure-visual-properties to Go`
**Build:** green

**Decision:** byte-identical-json

**Decision:** exact-option-names

**Decision:** python-goldens-as-testdata

**Decision:** python-io-semantics

  - [x] **Step 1: Goldens, before any Go.** With the Python still in place: extract each fixture
    snippet of `scripts/test-measure-visual-properties.sh` into a scratch generator (not committed)
    writing its PNGs to `stats/internal/guard/testdata/measure-visual-properties/`; record every
    harness invocation's argv, stdout, stderr and exit code as `<case>.golden` (argv on line 1 as a
    JSON array, then exit code, then stdout; stderr in `<case>.stderr`). Add invocations until
    every property in `ALL_PROPS` appears in at least one single-image golden and one two-image
    golden (default `--props` on a box fixture with `--scale 1` and `--scale 2`, `--ref-a/--ref-b`,
    `--region-a/--region-b`), plus the exit-1 (`Unresolved`) and exit-2 refusals in the docstring.
    Record the golden count in `design.md` **Measurements**.
  - [x] **Step 2: Failing tests.** `TestPyJSON`: `pyFloat` renders `3.0`, `0.1`, `1e-05`,
    `0.0001`, `1e+16`, `1.5e+16`, `-0.0`, `123456789012345.0`; `pyRound(x, n)` matches Python's
    `round` on `0.125→0.12`, `2.675→2.67`, `2.5→2` (ndigits omitted, int), `-0.04→-0.0`; an ordered
    object marshals keys in insertion order with `indent=2` layout, `[]`/`{}` for empty, `null` for
    nil. Expected strings copied from `python3 -c 'import json; print(json.dumps(…, indent=2))'`
    output, the command in a comment beside each. `TestMeasureVisualProperties`: every harness case
    as a subtest per `ok:` label, and every golden replayed byte for byte. Run — expect failure.
  - [x] **Step 3: pyjson.go.** `pyFloat float64` and `pyInt int` with `MarshalJSON` (float:
    `strconv.FormatFloat(f, 'f', -1, 64)` plus `.0` when integral for `1e-4 ≤ |f| < 1e16`, else
    `'e', -1` — Python's `repr`); `pyRound(x float64, n int) pyFloat` via `FormatFloat(x, 'f', n)`
    then `ParseFloat`; `pyObj` (a slice of key/value pairs, `set` replacing in place) with
    `MarshalJSON`; one `pyDump(v any) string` emitting Python's `indent=2` layout. Values are
    `pyInt`/`pyFloat` so Python's int/float promotion is explicit at each arithmetic site.
  - [x] **Step 4: Port**, registering `measure-visual-properties`: the argparse surface under
    `exact-option-names` (`--region-a`, `--region-b`, `--scale`, `--ref-a`, `--ref-b`, `--props`,
    `--edge`, `--noise`, each `--opt value` or `--opt=value`, last one wins, positionals anywhere,
    `-h`/`--help`), `parse_box` (ASCII digits only), every calibration refusal in the Python's
    wording, `Region` with page mode, each `measure_*`, `pair_in_order`, `pair_seams`, `delta`,
    `leaves`. `Counter.most_common(1)` is a count map plus first-seen order, ties to the earliest;
    `dist` is `math.Sqrt` of the summed squares in channel order; `round` is `pyRound`. Uses
    `decodeRGB`.
  - [x] **Step 5: Green.** `go test ./internal/guard/ -run '^(TestMeasureVisualProperties|TestPyJSON)$'
    -count=1 -race -v | grep -c -- '--- PASS: TestMeasureVisualProperties/'` — at least 11, and
    every golden passing.
  - [x] **Step 6: Shim and delete** — `scripts/measure-visual-properties.sh` becomes the shim with
    the docstring-derived header; `git rm scripts/measure-visual-properties.py
    scripts/test-measure-visual-properties.sh`.
  - [x] **Step 7: Verify.** `gofmt -l`, `go vet ./internal/guard/`; the shim, through a symlink in
    a temp directory, reproduces one golden's stdout byte for byte (`cmp`).

  Correction (2026-09-28): `**Files:**` widened by `skills/flow/scripts/measure-visual-properties.py`,
  the installed symlink the `.py`'s deletion left dangling. Step 3 shipped as: `pyFloat(f float64)
  string` is Python's `repr` — `FormatFloat(f, 'f', -1, 64)` plus `.0` when integral for `f == 0` or
  `1e-4 ≤ |f| < 1e16` (the plan's rule missed zero), else `'e', -1`; `nan`/`inf`/`-inf` for the
  specials — and `pyRound(x, n) float64` via `FormatFloat(x, 'f', n)` then `ParseFloat`; `pyObj` an
  insertion-ordered pair slice; `pyDump` emits `indent=2` with `NaN`/`Infinity`. No `MarshalJSON`
  (`encoding/json` rejects the NaN/Infinity Python prints for `--scale nan`), no `pyInt` (Go's `int`
  carries the distinction), no `set` (objects are built in final order) — measured: `python3 -c
  'print(repr(0.0), repr(-0.0), repr(1e15), repr(1e16), repr(1e-05), round(2.675, 2), round(-0.04, 1))'`
  @ 3915fbc0. Reuses `pyStrip` (`installedcitations.go`), `ppRepr`/`ppOSError` (`planprovenance.go`).
  Goldens: 52, a `modal-tie` fixture added when a last-seen tie mutant survived. Subtests: 78 (11
  harness, 52 goldens, 14 usage, 1 symlinked cwd). Testdata: 127 files, 155,253 bytes.
  <!-- measured: ls stats/internal/guard/testdata/measure-visual-properties/*.golden | wc -l; go test ./internal/guard/ -run '^TestMeasureVisualProperties$' -count=1 -v | grep -c -- '--- PASS: TestMeasureVisualProperties/'; find stats/internal/guard/testdata/measure-visual-properties -type f | wc -l; cat those files | wc -c @ branch spectre/kan-850-agents-port-the-visual-verify-scripts-to-the-go -->

- [x] 9. Delete the sole-user helpers and repoint citations

**Files:** `scripts/lib/trim-glob-element.sh`, `scripts/lib/git-clean.sh`, `.flow/project.md`
**Allowed-collateral:** `.flow/*.md`, `scripts/*.sh`, `scripts/lib/*.sh`, `scripts/lib/*.awk`, `skills/**/*.md`, `rules/*.mdc`, `README.md`, `CONTRIBUTING.md`, `stats/internal/guard/*.go`
**Tests:** none — deletion and citation sweep; the lint guards are the check
**Regression:** none — prose, comments and two unreferenced libraries
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 2, 3, 4, 5, 7, 8
**Commit:** `docs(scripts): repoint citations of the visual-verify scripts to their Go ports`
**Build:** green

**Decision:** kan850-helper-twins

**Decision:** carry-prior-port-decisions

  - [x] **Step 1: Delete.** `grep -rn 'trim-glob-element\|git-clean' scripts stats skills .flow`
    must name only comments; then `git rm scripts/lib/trim-glob-element.sh scripts/lib/git-clean.sh`.
  - [x] **Step 2: Find.** `grep -rlF -e test-check-visual-trigger.sh -e
    test-check-visual-verification.sh -e test-resolve-visual-screenshots.sh -e
    test-compose-mockup-frames.sh -e test-measure-visual-properties.sh -e compose-mockup-frames.py
    -e measure-visual-properties.py -e trim-glob-element -e git-clean.sh -e Pillow
    --exclude-dir=archive --exclude-dir=.worktrees --exclude-dir=node_modules --exclude-dir=.git
    --exclude-dir=self-review .`; then the same tree for citations of the ported scripts' bash
    plumbing (`sources lib/…`, "thin wrapper", "python3 probe").
    `unverified: the file set is known only after tasks 2–8 land; widen **Files:** by a correction if a hit falls outside the collateral globs`
  - [x] **Step 3: Repoint** each citation to the Go file or test that now holds what it cites; a
    sentence describing plumbing that no longer exists is corrected, not repointed. Known at plan
    time: `skills/flow/visual-verify.md` (line 331's "Pillow absent" exit-2 cause, line 654's
    `test-measure-visual-properties.sh`), `skills/flow-contracts/project-configuration.md`
    (the compose header citation), `.flow/project.md`'s Go-guard list gains the five ports and its
    Python list is checked for them, and the headers of `lib/strip-bom.sh`,
    `lib/sanitize-display.sh`, `lib/visual-table-cells.awk` stop naming the ported scripts as
    sourcing callers and name their Go twins.
  - [x] **Step 4: Verify.** Step 2's grep returns only `spectre/changes/kan-850-*`,
    `docs/self-review/` and the Go ports' own history comments; every guard in `.flow/project.md`'s
    `## lint` exits 0.

- [x] 10. Live verification: before/after timings and parity

**Files:** none
**Tests:** none — measurement task; the figures it records are the check
**Regression:** none — no commit
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 1, 2, 3, 4, 5, 6, 7, 8, 9
**Build:** green

**Decision:** suite-median-below-before

**Decision:** carry-prior-port-decisions

  - [x] **Step 1: Base checkout.** `git worktree add --detach <scratch>/kan850-base 3915fbc0`
    (removed at the end of this task).
  - [x] **Step 2: Suite, interleaved.** Three rounds, each: `sysctl -n vm.loadavg`, then
    `FLOW_GUARD_CACHE_DIR=$(mktemp -d) /usr/bin/time -p scripts/run-guard-tests.sh` in the base
    checkout, then the same on the branch head; record real/user/sys, load, exit, harness count
    and the slowest five harnesses (`grep '(Ns)'`, sorted descending) of each.
  - [x] **Step 3: Go package, interleaved.** Three rounds of `cd stats && /usr/bin/time -p go test
    ./internal/guard/... -count=1`, base then head, load before each.
  - [x] **Step 4: Parity.** `cd stats && go test ./internal/guard/ -count=1 -v | grep -c -- '---
    PASS: Test<Name>/'` per port against its floor in `design.md`, plus `--- FAIL` count 0.
  - [x] **Step 5: Record** **Before** and **After** tables under `design.md`'s **Suite
    before/after**, KAN-842's columns, each figure tagged `measured:` with the command and `@
    3915fbc0` or `@ branch spectre/kan-850-agents-port-the-visual-verify-scripts-to-the-go`; name
    the slowest remaining harness.
  - [x] **Step 6: Judge.** Failure looks like: after suite median not below before; any port's
    `--- PASS` count below its floor; the Go package median above 40s real
    (`guard-package-under-40s`, carried); any harness red; any `python3`/`PIL` left in
    `scripts/check-visual-*.sh`, `scripts/resolve-visual-screenshots.sh`,
    `scripts/compose-mockup-frames.sh`, `scripts/measure-visual-properties.sh`. Any of these is
    reported, not recorded as success.
