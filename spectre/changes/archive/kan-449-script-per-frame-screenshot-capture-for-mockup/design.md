# kan-449-script-per-frame-screenshot-capture-for-mockup — design

**Change:** `kan-449-script-per-frame-screenshot-capture-for-mockup` · **Jira:** KAN-449 ·
**Date:** 2026-09-10

## Context

`flow.visual-verify` already starts the stack (`start`, e.g. gymie's `./gradlew devStart …`),
runs the per-change Playwright spec (`capture`) and resolves its PNGs
(`resolve-visual-screenshots.sh`). What it does **not** do is put a captured frame next to the
mockup it was drawn from. That comparison is done by the verifier in its head and written down as
prose in `visual-verification.md` — unreproducible, and invisible to the review panel and the
handoff, which see the prose but never the pair of images it describes. KAN-423's own run is the
motivating case: nineteen mockup frames (`gymie/docs/design/group-sessions/screens/G1.png …
K3.png`) against `gymie-playwright/tests/group-session-mockup-*.spec.ts-snapshots/*.png`, compared
by hand.

## 2. What this change adds

Three pieces, all inside the existing stage rather than beside it:

1. **`scripts/compose-mockup-frames.py`** (+ `compose-mockup-frames.sh` wrapper, matching every
   other Python guard's shape) — reads a **map** of `<screenshot name> <frame id>` pairs, the
   resolved capture PNG paths on stdin, a mockup root and an output directory, and writes one
   side-by-side PNG per pair.
2. **A `mockups` setting** in `## visual verification` — the root directory holding
   `<frame id>.png` files. Optional; absent means the stage composes nothing and reports
   `mockups: not declared`.
3. **A new step in `flow.visual-verify`** between "read every captured PNG" and "write
   `visual-verification.md`": run the compositor, read every composite, and list each pair in the
   report.

Plus the edits those imply: `check-visual-verification.sh` (closed `Setting` vocabulary gains
`mockups`), its test harness, `project-configuration.md` (the contract table), `verify-and-handoff.md`
(the step and the report shape), and gymie's `.flow/project.md` (the `mockups` row).

## 3. The map — a sidecar beside the spec

A per-change capture spec `<spec>` may carry a sibling `<spec>.mockups` (e.g.
`kan-423-group-sessions.spec.ts.mockups`) in the same directory — the regression checkout when one is
declared, else the project root, exactly where the spec itself lands. One pair per line:

```text verified:authored in-tree for this change — names are gymie-playwright/tests/group-session-mockup-dialogs.spec.ts-snapshots/ basenames and gymie/docs/design/group-sessions/screens/ frame ids
# <screenshot name as passed to toHaveScreenshot()>   <frame id>
j5-finish-dialog-mixed-readiness.png   J5
k2-menu-no-reopen.png                  K2
k3-delete-dialog.png                   K3
```

- Whitespace-separated, two fields; blank lines and lines starting with `#` are ignored; a line
  with any other field count is a violation.
- `<screenshot name>` is the name the spec passed to `toHaveScreenshot()`. Playwright writes it to
  disk as `<name-without-.png>-<platform>.png` (`…-darwin.png`); the compositor matches a resolved
  PNG whose basename is either the name verbatim or the name with a `-<platform>` suffix inserted
  before `.png`. That keeps the map platform-independent, which the spec already is.
- `<frame id>` names `<mockups>/<frame id>.png`. Case is significant — the frame files are.
- The map is explicit by operator decision (`explicit-map`): the existing capture names mostly start
  with a lowercased frame id (`j5-…`, `k2-…`) but not always (`entry-no-grant`), and a convention
  that guesses would silently pair the wrong frame.

## 4. The compositor

`compose-mockup-frames.sh <map file> <mockups root> <out dir>`, resolved PNG paths on stdin, one per
line (the exact stdout of `resolve-visual-screenshots.sh`).

Layout per pair: captured frame on the **left**, mockup on the **right**, both at native size,
top-aligned, a 16px gutter, on a plain background; no labels, no scaling — the order is stated once
here and in the contract, and a label would need a font the script would then have to find. Output
file: `<out dir>/<screenshot name without .png>.png`. Output directory created if absent.

Exit codes, in the shape every guard here uses:

- `0` — every map line composed; one `<composite path>` per line on stdout.
- `1` — a map line names a screenshot no stdin path matches, or a frame file absent under the
  mockups root, or a malformed map line. Every finding is printed as `<map>:<line>: <message>` and
  the run still composes every other pair, so one bad line does not hide the rest.
- `2` — cannot answer: usage error, the map unreadable, the mockups root not a directory, the
  output directory not creatable, an unreadable PNG, or **Pillow absent** — reported as
  `compose-mockup-frames: Pillow is required — python3 -m pip install pillow`, naming the
  prerequisite the way `setup` names chromium. Pillow is a machine prerequisite, not a project
  dependency: nothing in this repository declares it, and the script checks for it at start.

A capture with no map line is not a finding — not every captured view is a mockup frame. A frame
file no map line names is not a finding either — a change touches a subset of frames.

The script is Python with a `.sh` wrapper because the other Python guards here
(`check-plan-shape.py`/`.sh`, `plan-dispatch-bundles.py`/`.sh`) are shaped that way and
`run-guard-tests.sh` finds the `.sh`.

## 5. The stage

`flow.visual-verify` (`skills/flow/verify-and-handoff.md`) gains a step between today's 8 and 9,
run by the verifier:

> **Compose captured frames against their mockups, if `mockups` is declared.** With a
> `<spec>.mockups` sidecar beside the capture spec, pipe step 8's printed paths into
> `compose-mockup-frames.sh <spec>.mockups <mockups> <changeRoot>/visual-verification`. Exit 0
> prints one composite path per line — **read every one** and state, per pair, whether the capture
> matches the frame and where it departs. Exit 1 (a broken map — a screenshot no capture matched, a
> frame file absent, a malformed line) or 2 (cannot answer) blocks. No sidecar → report
> `mockups: no map for <spec basename>` and continue. `mockups` not declared → report
> `mockups: not declared` and continue.

Step 9's `visual-verification.md` gains one entry per composed pair: the composite's absolute path,
the frame id, and what was seen. Step 10 commits `<changeRoot>/visual-verification/` with the rest
of the change root. The report block gains:

```text verified:authored in-tree for this change — extends the ## Report block of skills/flow/verify-and-handoff.md
- mockups: <not declared | no map for <spec> | exit <n>>
- <frame id>: <absolute composite path> — <match, or the departure seen>
```

**Blocking** gains: a compositor exit 1 or 2, and a departure from the mockup the verifier reports
in a composite — the same standing the existing "a defect in a captured screenshot" already has.

The composites live in the change root, committed, by operator decision (`composites-in-change-root`):
the panel reads them by path, the archive keeps them, and the cost is a few MB of PNG per UI change.

## 6. The contract and its guard

`## visual verification` (`skills/flow-contracts/project-configuration.md`) settings table gains:

| Setting | Required | Meaning |
|---------|----------|---------|
| `mockups` | no | Directory holding one `<frame id>.png` per drawn frame, relative to the project root. Declared, it enables the compose step above for any capture spec carrying a `<spec>.mockups` sidecar; absent, the stage composes nothing. |

`check-visual-verification.sh` adds `mockups` to the closed `Setting` vocabulary — the whole of its
change; it validates neither that the directory exists (a run-time question the compositor answers)
nor the sidecar. `test-check-visual-verification.sh` pins that `mockups` is accepted and that the
vocabulary is otherwise unchanged.

gymie's `.flow/project.md` declares `| \`mockups\` | \`docs/design/group-sessions/screens\` |` — the
one project with drawn frames today.

## 7. Out of scope

- Pairing by filename convention — rejected (`explicit-map`).
- Any change to `start`/`capture`: the stage already runs devStart and the spec.
- `gymie-frontend`'s missing `## visual verification` section — recorded under open questions.
- Scaling, labels or diff heat-maps on the composite: a reviewer reads two images; the script
  only puts them beside each other.

## Decisions

### The comparison is wired into the stage, not a hand-run script

**ID:** wire-into-stage
**Status:** active
**Chosen:** a compositor script plus a `mockups` setting and a `flow.visual-verify` step — the stage already starts the stack and runs the spec, so only the composite step was missing, and only a stage step is reproducible and visible in the panel and handoff.
**Considered:** a standalone script in `agents/scripts/` run by hand — nothing would make a run produce it, so the manual comparison would remain the default.

### Python + Pillow composes the PNGs

**ID:** pillow-compositor
**Status:** active
**Chosen:** `compose-mockup-frames.py` on Pillow, checked for at start and named as a prerequisite on absence — the shortest working script; Pillow 12.3.0 is present on the operator's machine.
**Considered:** ImageMagick `montage` — not installed; Playwright rendering an HTML pair page to PNG — no new dependency, but the script would have to run inside the regression checkout's `node_modules` and become a Node/Playwright file rather than a guard; HTML output only — the panel's reviewers read PNGs through the Read tool and cannot see an HTML page.

### A capture finds its frame through an explicit sidecar map

**ID:** explicit-map
**Status:** active
**Chosen:** `<spec>.mockups` beside the capture spec, one `<screenshot name> <frame id>` line per pair — no guessing; the file lands where the spec and its PNGs land.
**Considered:** filename convention (first `-` segment uppercased is the frame id) — most existing captures fit it but `entry-*` does not, and a convention pairs wrong silently; a table inside `visual-verification.md` — keeps the map in the change rather than beside the spec that produced the captures.

### Only a broken map blocks

**ID:** broken-map-blocks
**Status:** active
**Chosen:** a map line naming a screenshot no capture matched, a frame file absent, or a malformed line blocks; no sidecar at all is reported and does not block — backend-only diffs trigger this stage with no frame to compare (KAN-423's run was one).
**Considered:** blocking whenever `mockups` is declared and no sidecar exists — would force every triggered run to write a map or an explicit empty one.

### Composites are committed in the change root

**ID:** composites-in-change-root
**Status:** active
**Chosen:** `<changeRoot>/visual-verification/<screenshot name>.png`, committed with the change — the panel reads them by path and the archive keeps the evidence; a few MB of PNG per UI change.
**Considered:** `.superpowers/` scratch referenced by absolute path — nothing committed, evidence gone with the worktree.

## Open questions

### `gymie-frontend` declares no `## visual verification` section

**ID:** frontend-section-missing
**Status:** open
**Why it is open:** out of this change's scope by operator decision; a follow-up candidate.
**What it affects:** a frontend-only change in `gymie-frontend` never triggers `flow.visual-verify`, so the composite step lands there only when the gymie backend diff triggers the stage (as KAN-423's did).
