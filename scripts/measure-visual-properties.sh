#!/usr/bin/env bash
# measure-visual-properties.sh — measure one UI control's visual properties
# from pixels, in one image or in two (a mockup crop and the live capture of the
# same control), and report the numbers plus the calibrated delta as JSON.
#
# Interface:
#
#   measure-visual-properties.sh <image A> [<image B>] [options]
#
#     --region-a x,y,w,h    crop of A to measure (default: the whole image)
#     --region-b x,y,w,h    crop of B to measure (default: the whole image)
#     --scale <f>           calibration: B pixels per A pixel (a project's
#                           `mockup frame` scale is 1 here when A is the frame
#                           the composite already cropped and B its capture)
#     --ref-a x,y,w,h       calibration from two boxes already known correct in
#     --ref-b x,y,w,h       both images: scale = ref-b width / ref-a width
#                           (heights must agree with that factor within 5%)
#     --props <list>        comma-separated subset of box,radius,border,fill,
#                           shadow,content,gap,ink,runs,bands,seams (default: all)
#     --edge <n>            adjacent-pixel colour jump (RGB Euclidean, 0-441)
#                           that counts as a hard edge (default 24)
#     --noise <n>           colour distance to background below which a pixel
#                           is background (default 8)
#
#   With two images one calibration input is mandatory — `--scale` or the
#   `--ref-*` pair; a cross-image comparison with neither is refused, because an
#   uncalibrated ruler is itself an unchecked claim (KAN-30 fix round 9).
#
# Each region is a crop around ONE control with a background margin on all
# four sides and nothing else in it — the background colour is the region's
# top-left pixel, and all four corners must be background. Everything is
# measured on the centre row and centre column of the region, from the outside
# inward, so the region's centre must fall inside the control. Only `shadow`
# and `gap` look past the region, out to the image's own edge, so a neighbour
# belongs outside the region, not in it. `ink` and `runs` need no control edge
# at all and are the two properties to run before the crop is trusted.
#
# Properties (pixels of the region's own image; `null` where not found):
#
#   box      left/top/right/bottom edge coordinates (region-relative) and
#            width/height. An edge is the first pixel, walking inward along
#            the centre scan line, whose colour jumps from its neighbour by at
#            least `--edge`; a soft shadow gradient never jumps that far in
#            one pixel, a hard edge (with or without antialiasing) does.
#   radius   corner radius per corner: rows walked from the top (bottom) edge
#            until the row's own edge reaches the box's straight side; the
#            area those rows leave between the side and the curve is
#            r²(1 − π/4), inverted for r. 0 for a square corner.
#   border   per-side width in px and the border colour: pixels from the edge
#            inward that are neither the fill nor a background/fill blend.
#   fill     modal colour of the interior (inset a quarter of the box, so the
#            border and edge antialiasing are excluded) and its share of those
#            pixels.
#   shadow   per side: width of the band outside the edge whose pixels differ
#            from the background by more than `--noise`, the band's peak
#            distance from background, and an approximate opacity assuming a
#            black shadow. A flat design is width 0 on every side.
#   content  bounding box of non-fill pixels inside the border and outside the
#            corner curves (an icon, a glyph run, a label), its size, its size
#            as a ratio of the box, the padding from each box edge to it, and
#            `colour` — the modal colour of those pixels, the icon's or text's
#            tint. `fill` is the box behind the glyph, never the glyph: a grey
#            icon where the mockup draws an accent one matches on box, radius,
#            fill and content size alike, and only this colour's `delta`
#            distance sees it (KAN-437).
#   gap      per side: background pixels between the box edge (past any
#            shadow band) and the next non-background pixel on the same scan
#            line, out to the image's edge — the spacing to the nearest
#            neighbour; `null` when the scan line reaches the image edge.
#   ink      bounding box of every non-background pixel in the region, and
#            `colour` — the modal colour of those pixels. Needs no control
#            edge: crop to a run of capital letters and request `--props ink`
#            — its height is the cap-height, the font-size stand-in; its
#            `left`/`top` in a crop spanning the container are the run's
#            position and, against the crop's width, its alignment; its
#            `colour` is the text's tint, which no other property reads on a
#            run with no box (KAN-437).
#   runs     colour runs along the region's centre row and centre column:
#            `from`/`to` (region-relative), `length` and `colour` of every
#            stretch of pixels within `--noise` of the stretch's first pixel.
#            The scanline, needing no clean box. Two readings the other
#            properties cannot give: (1) an edge-to-edge claim — a fill flush
#            with its container's border is the run immediately after the
#            border's run, and an inset fill shows as the container-coloured
#            run between them, its `length` the inset in px (`content` cannot
#            answer this: it boxes every non-fill pixel in the container,
#            other rows' text included, and a fill covering half the box
#            flips which colour counts as `fill` — KAN-30 manual re-sweep);
#            (2) where to crop — every run boundary is an edge a region can
#            sit on, read before any property that needs a box. No `delta`:
#            compare the two lists by eye.
#   bands    the region's horizontal structure, whole-page: every maximal run
#            of rows holding a non-background pixel is a band — a text line,
#            a control, a card, a 1px divider — with `top`, `height`,
#            `gap_above` (background rows since the previous band), `edge`
#            (the modal non-background colour of the band's first row: a
#            control's border or a card's fill, whichever the eye meets
#            first) and `colour` (the modal non-background colour of the whole
#            band). Background is the region's modal colour, not its corner,
#            and no corner needs to be background: the region is the whole
#            capture and the whole cropped frame. Unlike every other
#            property, `bands` has a structural `delta`: the two lists are
#            paired in order by the gap since the last pair, and every band
#            the frame draws with no counterpart in the capture is listed
#            `missing`, every capture band the frame lacks `extra`, and each
#            pair carries its `gap_above`, `since_pair` (rows since the last
#            paired band — the one number a lost padding shows as when the
#            divider beside it is also missing), `height`, `edge` and
#            `colour` deltas. Text content differing between the two does not
#            unpair a band — a band is geometry and non-background ink, never
#            what the text says — so the departures this lists are exactly
#            the ones a data difference cannot explain: a divider the frame
#            draws and the capture omits, a row padding the capture lacks, a
#            surface fill swapped for the page colour, a button border in the
#            wrong colour (KAN-437 fix round 5: all four passed a composite
#            whose diff ratio was already 0.3 from data and fonts alone).
#            Vertical structure — a segmented control's inter-cell dividers,
#            a wrapped label's internal alignment — is not a band; `seams`
#            reads those.
#   seams    the vertical structure inside each BOXED band — a band whose
#            first row's ink spans at least half the band's ink width: a
#            bordered or filled control, a card, a hairline; never a bare
#            text row, whose glyph stems would otherwise read as seams. Per
#            boxed band (its `top`, `height`, `gap_above`, `edge`, `colour`
#            as in `bands`): `seams`, every maximal run of columns inked over
#            at least 90% of the band's height and of one modal colour
#            (adjacent columns further than `--edge` apart start a new seam)
#            — an outer border's side, a cell divider, a filled cell — with
#            `left`, `width`, `gap_left` (columns since the previous seam)
#            and `colour`; and `cells`, the spans between consecutive seams,
#            each with `left`, `width` and its `lines`: every maximal run of
#            rows carrying ink inside the span, rows inked across 90% of the
#            span (a border, an underline) excluded, with `top`, `height`,
#            `left` and `right` (the ink's inset from each cell edge) and
#            `offset` (the ink's centre minus the cell's centre, negative
#            when left of it). `delta.seams` pairs the boxed bands as `bands`
#            does, then within each paired band pairs the seams in order by
#            `since_pair` (columns since the last paired seam) and `width`,
#            listing `missing` and `extra` seams and per pair the `gap_left`,
#            `since_pair`, `width` and `colour` deltas; cells pair where both
#            bounding seams paired consecutively, and their lines pair by
#            index with `count` per image and per pair the `offset`, `left`
#            and `right` deltas. Text width never moves a seam, so a missing
#            seam is a divider the capture omits; a wrapped label whose lines
#            are centred in the frame and left-anchored in the capture reads
#            as an `offset` delta of half the slack with a `left` delta of the
#            same size, while a left-anchored label whose width is data reads
#            as an `offset` delta with `left` unchanged (KAN-437 fix round 5:
#            the GOAL control's two dividers and its wrapped labels' centring
#            both shipped wrong past every band and every sweep). A filled
#            cell is a seam, not a cell — its lines are not read; a divider in
#            a row with no border or fill sits in an unboxed band and is not
#            read either. A control whose outer border the capture omits is
#            therefore not boxed there at all: it is one of `bands_unpaired`,
#            never a list of missing seams, and its labels' centring goes
#            unread until the border is back (the real KAN-437 pre-fix
#            capture: GOAL and PROTEIN each read as a plain band of the same
#            height and gap as the frame's boxed one).
#
# Output (stdout): one JSON object — `a`, `b` (when given), `scale`, and
# `delta`: for every numeric leaf, `a`, `a_scaled` (`a` × scale for lengths,
# `a` unchanged for ratios and opacities), `b`, `abs` (b − a_scaled) and `pct`
# (abs / a_scaled × 100, null when a_scaled is 0); for every colour, the RGB
# Euclidean distance; for `bands`, `delta.bands` is the paired list above and
# `delta.bands_summary` its `paired`/`missing`/`extra` counts; for `seams`,
# `delta.seams` is the per-band paired list above and `delta.seams_summary`
# its `paired`/`missing`/`extra` seam counts, `bands_unpaired` (boxed bands
# with no counterpart) and `lines` (`paired`/`unpaired`). Thresholds are
# the caller's: this script flags nothing.
#
# Exit codes:
#   0  measured.
#   1  a region cannot be resolved cleanly: a corner that is not background, a
#      scan line that finds no edge, a corner scan that never reaches the
#      straight side, or the region's centre outside the box. The reason is on
#      stderr; nothing is guessed.
#   2  cannot answer: usage error, two images with no calibration, `--ref-*`
#      boxes whose aspect disagrees, or an unreadable image (anything but a PNG
#      included).
#
# The body is Go: stats/internal/guard/measurevisualproperties.go. Option names
# are exact — argparse's prefix abbreviations are not accepted — and a usage
# error is one `measure-visual-properties: …` line on stderr.
set -euo pipefail
. "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "measure-visual-properties: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec measure-visual-properties 2 "measure-visual-properties:" "$@"
