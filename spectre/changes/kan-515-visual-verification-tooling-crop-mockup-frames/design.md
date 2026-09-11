# kan-515-visual-verification-tooling-crop-mockup-frames

## Context

Seeded from `docs/superpowers/research/kan-515.md`, which this change adopts and deletes.

## The frames

The 34 PNGs under `gymie/docs/design/mockups/` are 2x exports of the phone frames in
`docs/design/group-sessions/design/Gymie Group Session Flow.html`. Every frame is the same markup:
a `390px` wide, `border-box`, `1px`-bordered container; a **26px status line** (`17:43` · `5G · 82%`)
the web app never renders; a 52px wordmark bar the app does have; content hugging its height; and a
dark (42, 42, 42) caption band of varying height beneath the frame.

Measured on `G1.png` with Pillow: 780 wide; border pixels at columns 0–1 and 778–779 and rows 0–1;
status line rows 2–53; frame bottom border rows 1494–1495; caption from row 1496. So the content
area is `776 x (frame - 4 - 52)` px, i.e. **388 x (frame/2 - 2 - 26) logical** — the app is captured
at 388 wide, not 390, because the 1px border sits inside the 390.
<!-- measured: Pillow pixel scan of gymie/docs/design/mockups/G1.png -->

## Declared geometry, detected caption

`## visual verification` gains one optional setting, `mockup frame`, value
`scale=<int> status=<logical px> border=<logical px>` — gymie's is `scale=2 status=26 border=1`.

Crop, in PNG pixels with `s = scale`: left `border*s`, top `(border + status)*s`, right
`w - border*s`, bottom `frame_bottom - border*s`. `frame_bottom` is **detected, not declared**,
because it differs per frame: the page background is the pixel at (0, 0) — outside the rounded
corner, so always the page — and `frame_bottom` is the first row below the top crop whose every
pixel in `[border*s, w - border*s)` equals it; no such row means the frame runs to the last row.

A uniform row is what the caption band's own top margin guarantees, and no screen row is uniformly
the page colour edge to edge because the border sits inside it.

## Three panels and a ratio

With the row declared: capture | cropped frame | diff, same 16px gutter, same (40, 40, 40)
background, top-aligned, no labels, no scaling. The diff panel is
`ImageChops.difference(capture, frame).convert("L").point(lambda v: 255 if v else 0)` — white on
black, so a departure reads as shape rather than as a colour blend. The ratio is differing pixels
over total pixels, printed to four decimals.

stdout becomes `<composite path> diff=<ratio>` per line, `diff=n/a` on the no-geometry path.
Nothing parses that stdout — `flow.visual-verify` step 9 hands it to the verifier — so the shape
change costs one contract sentence.

## Decisions

### Declared geometry over pixel heuristics for the fixed part

**ID:** declared-geometry-not-heuristics
**Status:** active
**Chosen:** a declared `scale`/`status`/`border` row — the three values are fixed properties of the
export, known to whoever wrote the mockups, and a heuristic that guessed them would fail silently on
a frame whose content happens to start with a uniform band.
**Considered:** detecting all three from the pixels (rejected: silent misdetection, no operator
signal); hardcoding gymie's values in the script (rejected: the script ships to every project).

### The caption bottom is detected, not declared

**ID:** caption-bottom-detected
**Status:** active
**Chosen:** scan for the first row below the top crop uniformly equal to the page colour at (0, 0).
Frames run from 520 to 1330 logical px tall because content hugs its height, so a declared bottom
would need one value per frame — 34 rows of configuration for something one scan derives.
**Considered:** a declared per-frame height in the `mockups` sidecar (rejected: 34 values to keep in
sync with a re-export); a fixed caption height (rejected: the caption's height varies with its text).

### The ratio is information, never a threshold

**ID:** no-pixel-diff-threshold
**Status:** active
**Chosen:** print the ratio beside the panels; the verifier reads it. This upholds kan-486's
`no-pixel-diff` decision — an antialiasing difference and a missing button produce comparable
ratios, so a numeric gate would block correct work and pass real departures.
**Considered:** failing above a ratio (rejected: the number does not carry the meaning a gate needs);
omitting the ratio entirely (rejected: a verifier reading three panels still benefits from knowing
whether the difference is 0.0003 or 0.4).

### Size mismatch blocks rather than warns

**ID:** size-mismatch-blocks
**Status:** active
**Chosen:** with the geometry declared, a capture whose size differs from the cropped frame is a
finding on stderr and exit 1 for that pair; siblings still compose. Equal size is the whole
apples-to-apples guarantee — a spec that set the wrong viewport height would otherwise produce a
composite that reads as a layout departure, which is the exact failure this change exists to end.
**Considered:** a warning line and compose anyway (rejected: a warning beside a plausible-looking
composite is what the verifier ignores); scaling one to match the other (rejected: resampling
manufactures differences and hides real ones).

### No row declared changes nothing

**ID:** absent-row-is-todays-behaviour
**Status:** active
**Chosen:** the geometry argument is optional; without it the two-panel native-size composite, the
absent size check and `diff=n/a` keep every other project's composites and the 19 existing harness
cases meaning exactly what they mean today.
**Considered:** making the crop unconditional with defaults (rejected: every project without HTML
phone-frame exports would be cropped wrongly and silently).

## Open questions

*(none)*
