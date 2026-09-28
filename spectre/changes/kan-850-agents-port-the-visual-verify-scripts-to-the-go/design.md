# kan-850-agents-port-the-visual-verify-scripts-to-the-go

## Context

- Fifth slice of KAN-760's guard port, and the first to port Python. Base `3915fbc0`.
- Pattern established by KAN-760/KAN-778/KAN-841/KAN-842: a guard is one Go file registered in
  `guard.Registry` plus a shim calling `flow_guard_exec <name> <code> "<name>:" "$@"`;
  `scripts/check-references.sh` is the reference shim. The cannot-answer code is 2 for all five
  scripts here.
  <!-- verified: each script's header exit-code paragraph and its usage-error `exit`/`sys.exit` @ 3915fbc0 -->
- The five scripts and who calls them:
  - `check-visual-trigger.sh` — `skills/flow/verify-and-handoff.md` (flow.visual-verify step 2),
    exec'd by `stats/internal/guard/visualverifydispatched.go` from beside its shim
    (`guardSelfDir`, `FLOW_GUARD_SELF`), and `.flow/project.md`'s `## lint`.
  - `check-visual-verification.sh` — a project's own `## lint` (`.flow/project.md` here);
    `guardsymlinks.go`'s `gsDeclaredRule6` declares it.
  - `resolve-visual-screenshots.sh` — `skills/flow/visual-verify.md`, piped into `zip` and into
    `compose-mockup-frames.sh`; `.flow/project.md`'s `## lint`.
  - `compose-mockup-frames.sh` + `.py` — `skills/flow/visual-verify.md` step 4.
  - `measure-visual-properties.sh` + `.py` — `skills/flow/visual-verify.md`,
    `skills/flow/implement.md`, `skills/flow/SKILL-rationale.md`.
  <!-- verified: grep -rlE 'check-visual-trigger|check-visual-verification|resolve-visual-screenshots|compose-mockup-frames|measure-visual-properties' scripts stats skills .flow, excluding test-* @ 3915fbc0 -->
- Sibling helpers and their other bash callers at base:
  - `lib/strip-bom.sh` — also `check-dev-stack-fresh.sh`, `check-spec-reach.sh`,
    `lib/project-section.sh`; no Go twin.
  - `lib/sanitize-display.sh` — also `check-spec-reach.sh`; `ccSanitize` (`cleanupcomplete.go`)
    claims to be its Go form but has no parity test, and escapes byte 0x00 where the awk loop
    starts at 1.
    <!-- verified: cleanupcomplete.go ccSanitize `c < 0x20`; lib/sanitize-display.sh `for (j = 1; j < 32; j++)` @ 3915fbc0 -->
  - `lib/visual-table-cells.awk` — also `check-spec-reach.sh`, `check-dev-stack-fresh.sh`; no Go
    twin (`workspaceisolation.go`'s `splitCells` is the wider four-column copy and stays).
  - `lib/trim-glob-element.sh` — only `check-visual-trigger.sh` and `check-visual-verification.sh`.
  - `lib/git-clean.sh` — only `check-visual-verification.sh`.
  <!-- verified: grep -l '<lib>' scripts/*.sh scripts/*.py scripts/lib/* stats/internal/guard/*.go, excluding the lib itself and test-* @ 3915fbc0 -->
- Pillow semantics the Go decode must meet (Pillow 12.3.0): `convert("RGB")` drops alpha without
  compositing (RGBA `(10,20,30,0)` → `(10,20,30)`); a 16-bit grey PNG opens as `I;16` and converts
  to `(255,255,255)` for every value tried (`0x00ff`, `0x0100`, `0x1234`, `0xffff`) — a clip, not a
  scale.
  <!-- measured: python3 Image.open(...).convert("RGB").getpixel((0,0)) on a saved RGBA PNG and on hand-built 16-bit grey PNGs, Pillow 12.3.0 @ 3915fbc0 -->
- Fixtures: both Python harnesses draw their PNGs with Pillow's `ImageDraw`, `rounded_rectangle`
  included (`test-measure-visual-properties.sh` line 41); no harness case pins a diff ratio that
  the luma rounding moves (every pinned ratio is `0.0000`, `1.0000` or `n/a`).
  <!-- verified: grep -n 'diff=' scripts/test-compose-mockup-frames.sh; grep -n rounded_rectangle scripts/test-measure-visual-properties.sh @ 3915fbc0 -->
- `test-compose-mockup-frames.sh` case 11 pins "Pillow absent → exit 2 with a pip hint"; the Go
  port has no Pillow, so that case has no Go counterpart.
- Python I/O the image ports inherit: `str.split()` splits on `\x1c`, `str.strip()` strips `\x1f`
  and `\x85`, a text-mode read ends a line at a lone `\r`, and `os.getcwd()` is the physical path
  (`/private/tmp/…` from a `cd /tmp/…`) — where Go's `os.Getwd` returns `$PWD` (`/tmp/…`) and
  `syscall.Getwd` the physical one.
  <!-- measured: python3 -c '"a\x1cb c".split(); " \x1fz\x85 ".strip(); io.TextIOWrapper(BytesIO(b"a\rb\r\nc\n")) lines; os.getcwd()' and a Go program printing os.Getwd/syscall.Getwd from /tmp/gwtest, go1.26.5 darwin/arm64 @ 3915fbc0 -->
- No ported guard reads stdin: `guard.Env` has no stdin field and `cmd/flow-guard/main.go` passes
  none; `check-visual-trigger` and `compose-mockup-frames` read their input from stdin.
  <!-- verified: grep -n 'Stdin\|stdin' stats/internal/guard/guard.go stats/cmd/flow-guard/*.go — no hit @ 3915fbc0 -->
- Guard harness labels: `test-measure-visual-properties.sh` has 10 cases across 11 `pass` lines;
  `test-compose-mockup-frames.sh` 54 `pass` call sites for 59 `ok` lines (some in loops).
  <!-- measured: grep -c 'pass "' scripts/test-<name>.sh @ 3915fbc0 -->


## Measurements

### Parity floors at base

| Harness | floor | real (5 run concurrently) |
|---|--:|--:|
| `test-check-visual-trigger.sh` | 51 | 2.42s |
| `test-check-visual-verification.sh` | 111 | 2.92s |
| `test-resolve-visual-screenshots.sh` | 38 | 1.56s |
| `test-compose-mockup-frames.sh` | 59 | 7.69s |
| `test-measure-visual-properties.sh` | 11 | 3.56s |

- Floor is the harness's `^ok` line count on a green run; all five exited 0, `FAILURES: 0`.
- Load 3.52 7.75 7.24 before the run.

<!-- measured: FLOW_GUARD_CACHE_DIR=$(mktemp -d) /usr/bin/time -p bash scripts/test-<name>.sh, the five concurrently, grep -c '^ok' of each log, all exit 0 @ 3915fbc0 -->

Each Go test carries at least its harness's floor in cases (`parity-by-case-count`).

- measure-visual-properties goldens: 52, every `ALL_PROPS` property single-image and paired; the
  port matched the base Python on 452 + 244 differential argvs (stdout, stderr, exit).
  <!-- measured: task 8's scratch generator and seeded-random sweeps against python3 3915fbc0:scripts/measure-visual-properties.py, Pillow 12.3.0 -->
- Compose luma vs any-channel mask: equal on all 8 harness three-panel composites; only task 7's
  new one-pixel `(255,0,0)`/`(254,0,0)` fixture differs (Python `diff=0.0000`, port `diff=0.0005`).
  <!-- measured: task 7's scratch generator comparing ImageChops.difference(...).convert("L") with an any-channel mask per golden composite, Pillow 12.3.0 @ 3915fbc0 -->

### Suite before/after

The live-verification task records **Before** (a detached checkout at `3915fbc0`) and **After**
(the branch head), interleaved on the same machine, KAN-842's columns.

## Decisions

### Carry the prior slices' port decisions unchanged

**ID:** carry-prior-port-decisions
**Status:** active
**Chosen:** KAN-760's `flow-guard-binary-rationale-in-go`, `go-tests-replace-harnesses`,
`parity-by-case-count`, `inject-deadlines-in-process`, `helper-parity-tests`,
`port-base-moves-into-go`, `guard-binary-built-from-checkout`, KAN-841's
`shared-helper-go-twins`, `sole-user-helpers-move-into-go`, and KAN-842's
`suite-median-below-before` and `guard-package-under-40s` govern this slice as written; the base
for `port-base-moves-into-go` is `3915fbc0`.
**Considered:** re-deciding each per script — only the Python ports raise cases those decisions do
not cover, and the entries below cover them.

### Scope: the five visual-verify scripts in one change

**ID:** scope-visual-verify-scripts
**Status:** active
**Chosen:** port `check-visual-trigger`, `check-visual-verification`,
`resolve-visual-screenshots`, `compose-mockup-frames` and `measure-visual-properties` — every
script `flow.visual-verify` runs that is not yet Go.
**Considered:** the bash three now and the Python two later — leaves Pillow in the stage for
another cycle, the operator's ask was the whole stage; `generate-relocation-comparison` too — it is
the review panel's, not visual-verify's.

### Shared helpers get Go twins; sole-user helpers move into Go

**ID:** kan850-helper-twins
**Status:** active
**Chosen:** per `shared-helper-go-twins`, `lib/strip-bom.sh`, `lib/sanitize-display.sh` and
`lib/visual-table-cells.awk` get Go twins in one file with parity tests running the bash/awk and the
Go over the same inputs; the libraries stay for their remaining bash callers. The
`sanitize_display` twin is whatever the parity test proves — `ccSanitize` if it passes, otherwise
a separate twin, with `ccSanitize` left to its own caller. Per `sole-user-helpers-move-into-go`,
`lib/trim-glob-element.sh` and `lib/git-clean.sh` are ported into Go and `git rm`'d with their last
bash caller.
**Considered:** reusing `workspaceisolation.go`'s `splitCells` — it parses a wider table and its
trim differs; porting the libraries' other bash callers — out of this slice's scope.

### The trigger is evaluated in-process

**ID:** trigger-in-process
**Status:** active
**Chosen:** `check-visual-verify-dispatched` calls the ported trigger's Go function with the diff's
paths instead of exec'ing `check-visual-trigger.sh`; its three exit codes keep their meaning. The
"sibling not found" and "exited N, neither 0, 1 nor 2" refusals become unreachable and their cases
are removed; the shim drops its `FLOW_GUARD_SELF` export where nothing else in the guard reads it.
**Considered:** keeping the exec — a process per call and a sibling that exists only for this.

### Stdin reaches a guard through Env

**ID:** env-stdin
**Status:** active
**Chosen:** `guard.Env` gains `Stdin io.Reader` — `os.Stdin` in `cmd/flow-guard/main.go`, a
`strings.Reader` in tests — read by `check-visual-trigger` and `compose-mockup-frames`.
**Considered:** reading `os.Stdin` inside each guard — tests could not feed it in-process, the cost
`inject-deadlines-in-process` exists to avoid.

### One PNG decode with Pillow's convert("RGB") semantics

**ID:** png-rgb-decode
**Status:** active
**Chosen:** both image ports decode through one Go function returning an 8-bit RGB buffer: every
PNG colour type's stored colour channels as Pillow's `convert("RGB")` yields them — alpha dropped,
never composited or premultiplied; grey and palette expanded; a palette entry's `tRNS` alpha
ignored. A 16-bit channel takes its high byte. Proven by a decode test over every colour type and
bit depth Go's `image/png` decodes.
**Considered:** Go's `color.RGBAModel` — premultiplies, so a fully transparent pixel reads black
where Pillow keeps its colour; mirroring Pillow's 16-bit-grey clip — the operator chose the
documented semantics over Pillow's quirks (`any-channel-diff-mask`), and no capture or frame is
16-bit grey.

### The compose diff mask counts any differing channel

**ID:** any-channel-diff-mask
**Status:** active
**Chosen:** a pixel is white in the difference panel, and counts toward `diff=`, when any of its
three channels differs — the script's own documented contract — replacing Pillow's
`difference().convert("L")` mask, whose luma rounding reads a one-level red or blue difference as
none (**Context**). The operator's choice at the design gate. Every other pixel the port writes —
the capture panel, the cropped frame, `<stem>.frame.png`, the background and gutters — and every
stdout/stderr line equal the Python's.
**Considered:** mirroring Pillow — a pure port, but ships a mask known to hide small departures.

### measure-visual-properties prints the Python's JSON byte for byte

**ID:** byte-identical-json
**Status:** active
**Chosen:** stdout equals `json.dump(result, indent=2)` plus a newline, byte for byte: an ordered
JSON object type keeps Python's insertion order, a float type prints Python's `repr` (`3.0`, `1e-05`,
`1e+16`), `round()` is Python's (round-half-even on the binary value), and every modal colour breaks
ties by first occurrence in scan order, as `Counter.most_common` does.
**Considered:** semantic equality — Go's default encoding sorts keys and prints `3` for `3.0`, so a
reader comparing an old and a new run sees a different document.

### Exact option names only

**ID:** exact-option-names
**Status:** active
**Chosen:** `measure-visual-properties` accepts `--opt value` and `--opt=value` under the full
option names, `-h`/`--help` (exit 0, a usage text on stdout), and exits 2 with a one-line
`measure-visual-properties: …` message on any other usage error; argparse's prefix abbreviations
are not accepted. Every skill citation spells the full names.
**Considered:** full argparse parity — prefix matching and argparse's own wording, for callers that
do not exist.

### Python outputs become committed goldens

**ID:** python-goldens-as-testdata
**Status:** active
**Chosen:** before either `.py` is deleted, each harness case's fixture PNGs are generated once by
the harness's own Python snippet and committed under
`stats/internal/guard/testdata/<script>/`, with the Python's output for that case beside them:
`measure-visual-properties`' stdout as `<case>.json`, `compose-mockup-frames`' composites and
`.frame.png` files as PNGs. The Go tests compare against them — JSON by bytes, PNGs by pixels
(the diff panel against an any-channel mask of the same pair). The same comparison, run once
against the live Python on every fixture, is the differential recorded in **Measurements**.
**Considered:** redrawing fixtures in Go — `rounded_rectangle` needs a rasterizer matching
Pillow's, or every pinned radius re-derived; keeping only the harness's assertions — 11 measure
cases pin a handful of fields, not the whole JSON.

### Python's text and path semantics are kept; its tracebacks are not

**ID:** python-io-semantics
**Status:** active
**Chosen:** the image ports keep what Python's I/O did on a well-formed input: text input read with
universal newlines (a lone `\r` ends a line), `str.split()`/`str.strip()` over Python's whitespace
set (which adds `\x1c`–`\x1f` and `\x85` to ASCII whitespace), and every absolute path printed as
`os.path.abspath` makes it — joined onto the physical working directory (`syscall.Getwd`), not
`$PWD`. An input on which the Python died with an uncaught exception — a traceback and exit 1,
e.g. a map file that is not UTF-8 — exits 2 with one `<script>: …` line instead.
**Considered:** Go's `strings.Fields`/`bufio.Scanner`/`os.Getwd` — each differs on those inputs
(**Context**); keeping the traceback's exit 1 — it reads as a finding, and a crash is not one.
**Correction (2026-09-28):** measured at task 7 — only the map file is read with universal
newlines; stdin splits on `\n` alone (CPython's `sys.stdin` is `newline="\n"` on POSIX). Python's
whitespace set is `str.isspace`: Go's `unicode.IsSpace` plus `\x1c`–`\x1f`, NBSP and the other
Unicode spaces included.

### resolve-visual-screenshots sorts bytewise

**ID:** bytewise-screenshot-sort
**Status:** active
**Chosen:** the port sorts its matches by byte order; the bash `sort` used the caller's locale
collation. Both consumers (`zip -@`, `compose-mockup-frames`' stdin) are order-insensitive.
**Considered:** exec'ing `sort` to keep the collation, as KAN-778 task 5 did for a guard whose
output order was read — nothing reads this order.

### The table and heading twins pin C-locale semantics

**ID:** c-locale-table-semantics
**Status:** active
**Chosen:** the Go twins of `lib/visual-table-cells.awk` and the `VV_HEADING` match fold and trim
ASCII only, as the awk and `grep -i` do under `LC_ALL=C`. The bash callers set no locale, so under
a UTF-8 locale they folded non-ASCII capitals (`Ä`→`ä`) and trimmed NBSP; the ports do neither.
Auto-resolved at task 1 on the recommended option; the parity tests run the awk under `LC_ALL=C`.
**Considered:** following the ambient locale — Go has none, and a test whose verdict depends on who
runs it is not a parity test.

### Walk errors print under the guard's own prefix

**ID:** prefixed-walk-errors
**Status:** active
**Chosen:** `resolve-visual-screenshots` reports an unreadable directory under the screenshots root
as `resolve-visual-screenshots: open <dir>: permission denied` on stderr, where `find` printed
`find: <dir>: Permission denied`; exit code and stdout are unchanged. Auto-resolved at task 4 on
the recommended option.
**Considered:** copying `find`'s line byte for byte — it names a tool the
port no longer runs.

### Exit codes follow the headers' contracts

**ID:** header-exit-codes
**Status:** active
**Chosen:** where a bash script's `set -e`/`pipefail` let a failing command's own status escape
outside the exit codes its header documents, the port answers within the header's contract. Known
case: `resolve-visual-screenshots` with a `regression checkout` that is a directory but not a git
repository exited 128 at `3915fbc0` (its `git worktree list` pipeline failed); the port finds no
worktree on the branch and exits 0 over the declared checkout. Raised by task 4's review,
auto-resolved on the recommended option.
**Considered:** reproducing 128 — an undocumented code a
caller cannot tell from a signal death.

### The trigger's globs match bytes

**ID:** byte-wise-trigger-globs
**Status:** active
**Chosen:** `check-visual-trigger`'s `?` matches one byte and `*`/`**` count bytes, whatever the
caller's locale — the bash's behaviour under `LC_ALL=C`. Under a UTF-8 locale the bash's `?`
matched one character (`c?d` matched `céd`); the port does not. A non-ASCII literal in a glob
matches its own bytes under both. Raised by task 2's review, auto-resolved on the recommended
option; the shim header says so.
**Considered:** character-wise matching — Go has no locale to
follow, the same reason as `c-locale-table-semantics`.

### compose-mockup-frames reads ASCII geometry and PNG only, with Go-worded decode errors

**ID:** compose-ascii-png-only
**Status:** active
**Chosen:** geometry values accept ASCII digits only (the Python's `isdigit()`/`int()` took any
Unicode decimal digit, `scale=٣` as 3); stdin captures and frames decode as PNG only
(`png-rgb-decode`), where Pillow opened any format; a corrupt PNG's message after `<path>: ` is Go's
`image/png` error; non-UTF-8 map/stdin and an unwritable output PNG exit 2 with the port's own line.
OSError text and Pillow's `cannot identify image file` are reproduced. Raised at task 7,
auto-resolved on the recommended option.
**Considered:** Unicode digits and Pillow's wording —
no caller writes either, and matching Pillow's per-decoder messages means embedding them.
**Correction (2026-09-28):** task 7's review added two more, auto-resolved the same way: a map line
naming the screenshot `.png` composes `out/.png` and exits 0 where the Python's `save` raised
(traceback, exit 1); a non-UTF-8 path argument prints raw in stderr lines where the Python's
`backslashreplace` stderr printed `\udcXX` (APFS refuses such names, so only error lines reach it).

### measure-visual-properties reads PNG and ASCII numbers only

**ID:** measure-png-ascii-only
**Status:** active
**Chosen:** the port measures PNG only (`png-rgb-decode`; another format exits 2 `cannot identify image
file`); a corrupt PNG's message is Go's decoder error; `parse_box` and float options take ASCII digits
only, and an integer past Go's `int` range is `invalid parse_box value`; `-h` prints argparse's
80-column uncoloured help; with abbreviations gone `-hx -h` prints help; a non-UTF-8 argument echoes
as U+FFFD where Python printed `\udcXX`. Exit codes match in every case but `-hx -h`. Raised at task
8, auto-resolved on the recommended option.
**Considered:** other decoders, Pillow's messages and
Unicode digits — no caller supplies them.

## Open questions
