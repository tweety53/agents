# kan-850-agents-port-the-visual-verify-scripts-to-the-go

## Why

- `flow.visual-verify` still runs five scripts outside `flow-guard`: three in bash
  (`check-visual-trigger`, `check-visual-verification`, `resolve-visual-screenshots`) and two in
  Python 3 + Pillow (`compose-mockup-frames`, `measure-visual-properties`) — the stage's one
  third-party dependency, which its harnesses fail rather than skip without.
- KAN-842 excluded the two Python scripts from its slice as a rewrite rather than a port; its
  measurements ranked their harnesses among the slowest in `scripts/run-guard-tests.sh`.
- `check-visual-verify-dispatched`, already Go, still execs `check-visual-trigger.sh` beside its
  shim.
- `compose-mockup-frames`' diff mask is built through Pillow's luma conversion, so a pixel differing
  by one level in red or blue alone reads as unchanged, against the script's own "white wherever any
  channel differs" contract.
  <!-- measured: python3 ImageChops.difference(...).convert("L") of per-pixel diffs (1,0,0),(0,1,0),(0,0,1),(1,0,1),(2,0,0),(0,0,5) → [0,1,0,0,1,1], Pillow 12.3.0 @ 3915fbc0 -->

## What changes

- The five scripts run as Go ports inside `flow-guard`; each `scripts/<name>.sh` becomes a
  `flow_guard_exec` shim keeping its header, and each harness is replaced by in-process Go tests.
- CLI contracts unchanged — arguments, output lines, stdout/stderr split, side effects, exit codes
  — with three stated departures: the compose diff mask counts any differing channel; 16-bit PNGs
  convert by high byte (Pillow clips 16-bit grey to 255); `measure-visual-properties` accepts exact
  option names only (argparse's prefix abbreviations are gone).
- `measure-visual-properties` prints the Python's JSON byte for byte; `compose-mockup-frames` writes
  the Python's pixels outside the diff panel. Python outputs on the fixtures are kept as Go test
  goldens.
- No Python or Pillow remains in the visual-verify path; `compose-mockup-frames.py`,
  `measure-visual-properties.py`, `scripts/lib/trim-glob-element.sh` and `scripts/lib/git-clean.sh`
  are deleted.
- `lib/strip-bom.sh`, `lib/sanitize-display.sh` and `lib/visual-table-cells.awk` gain Go twins
  with parity tests; the bash libraries stay for their remaining bash callers.
- `check-visual-verify-dispatched` evaluates the trigger in-process.
- Before/after suite timings are recorded in `design.md`.
