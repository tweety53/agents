#!/usr/bin/env python3
"""measure-visual-properties.py — measure one UI control's visual properties
from pixels, in one image or in two (a mockup crop and the live capture of the
same control), and report the numbers plus the calibrated delta as JSON.

Interface:

  measure-visual-properties.sh <image A> [<image B>] [options]

    --region-a x,y,w,h    crop of A to measure (default: the whole image)
    --region-b x,y,w,h    crop of B to measure (default: the whole image)
    --scale <f>           calibration: B pixels per A pixel (a project's
                          `mockup frame` scale is 1 here when A is the frame
                          the composite already cropped and B its capture)
    --ref-a x,y,w,h       calibration from two boxes already known correct in
    --ref-b x,y,w,h       both images: scale = ref-b width / ref-a width
                          (heights must agree with that factor within 5%)
    --props <list>        comma-separated subset of box,radius,border,fill,
                          shadow,content,gap,ink,runs,bands,seams (default: all)
    --edge <n>            adjacent-pixel colour jump (RGB Euclidean, 0-441)
                          that counts as a hard edge (default 24)
    --noise <n>           colour distance to background below which a pixel
                          is background (default 8)

  With two images one calibration input is mandatory — `--scale` or the
  `--ref-*` pair; a cross-image comparison with neither is refused, because an
  uncalibrated ruler is itself an unchecked claim (KAN-30 fix round 9).

Each region is a crop around ONE control with a background margin on all
four sides and nothing else in it — the background colour is the region's
top-left pixel, and all four corners must be background. Everything is
measured on the centre row and centre column of the region, from the outside
inward, so the region's centre must fall inside the control. Only `shadow`
and `gap` look past the region, out to the image's own edge, so a neighbour
belongs outside the region, not in it. `ink` and `runs` need no control edge
at all and are the two properties to run before the crop is trusted.

Properties (pixels of the region's own image; `null` where not found):

  box      left/top/right/bottom edge coordinates (region-relative) and
           width/height. An edge is the first pixel, walking inward along
           the centre scan line, whose colour jumps from its neighbour by at
           least `--edge`; a soft shadow gradient never jumps that far in
           one pixel, a hard edge (with or without antialiasing) does.
  radius   corner radius per corner: rows walked from the top (bottom) edge
           until the row's own edge reaches the box's straight side; the
           area those rows leave between the side and the curve is
           r²(1 − π/4), inverted for r. 0 for a square corner.
  border   per-side width in px and the border colour: pixels from the edge
           inward that are neither the fill nor a background/fill blend.
  fill     modal colour of the interior (inset a quarter of the box, so the
           border and edge antialiasing are excluded) and its share of those
           pixels.
  shadow   per side: width of the band outside the edge whose pixels differ
           from the background by more than `--noise`, the band's peak
           distance from background, and an approximate opacity assuming a
           black shadow. A flat design is width 0 on every side.
  content  bounding box of non-fill pixels inside the border and outside the
           corner curves (an icon, a glyph run, a label), its size, its size
           as a ratio of the box, the padding from each box edge to it, and
           `colour` — the modal colour of those pixels, the icon's or text's
           tint. `fill` is the box behind the glyph, never the glyph: a grey
           icon where the mockup draws an accent one matches on box, radius,
           fill and content size alike, and only this colour's `delta`
           distance sees it (KAN-437).
  gap      per side: background pixels between the box edge (past any
           shadow band) and the next non-background pixel on the same scan
           line, out to the image's edge — the spacing to the nearest
           neighbour; `null` when the scan line reaches the image edge.
  ink      bounding box of every non-background pixel in the region, and
           `colour` — the modal colour of those pixels. Needs no control
           edge: crop to a run of capital letters and request `--props ink`
           — its height is the cap-height, the font-size stand-in; its
           `left`/`top` in a crop spanning the container are the run's
           position and, against the crop's width, its alignment; its
           `colour` is the text's tint, which no other property reads on a
           run with no box (KAN-437).
  runs     colour runs along the region's centre row and centre column:
           `from`/`to` (region-relative), `length` and `colour` of every
           stretch of pixels within `--noise` of the stretch's first pixel.
           The scanline, needing no clean box. Two readings the other
           properties cannot give: (1) an edge-to-edge claim — a fill flush
           with its container's border is the run immediately after the
           border's run, and an inset fill shows as the container-coloured
           run between them, its `length` the inset in px (`content` cannot
           answer this: it boxes every non-fill pixel in the container,
           other rows' text included, and a fill covering half the box
           flips which colour counts as `fill` — KAN-30 manual re-sweep);
           (2) where to crop — every run boundary is an edge a region can
           sit on, read before any property that needs a box. No `delta`:
           compare the two lists by eye.
  bands    the region's horizontal structure, whole-page: every maximal run
           of rows holding a non-background pixel is a band — a text line,
           a control, a card, a 1px divider — with `top`, `height`,
           `gap_above` (background rows since the previous band), `edge`
           (the modal non-background colour of the band's first row: a
           control's border or a card's fill, whichever the eye meets
           first) and `colour` (the modal non-background colour of the whole
           band). Background is the region's modal colour, not its corner,
           and no corner needs to be background: the region is the whole
           capture and the whole cropped frame. Unlike every other
           property, `bands` has a structural `delta`: the two lists are
           paired in order by the gap since the last pair, and every band
           the frame draws with no counterpart in the capture is listed
           `missing`, every capture band the frame lacks `extra`, and each
           pair carries its `gap_above`, `since_pair` (rows since the last
           paired band — the one number a lost padding shows as when the
           divider beside it is also missing), `height`, `edge` and
           `colour` deltas. Text content differing between the two does not
           unpair a band — a band is geometry and non-background ink, never
           what the text says — so the departures this lists are exactly
           the ones a data difference cannot explain: a divider the frame
           draws and the capture omits, a row padding the capture lacks, a
           surface fill swapped for the page colour, a button border in the
           wrong colour (KAN-437 fix round 5: all four passed a composite
           whose diff ratio was already 0.3 from data and fonts alone).
           Vertical structure — a segmented control's inter-cell dividers,
           a wrapped label's internal alignment — is not a band; `seams`
           reads those.
  seams    the vertical structure inside each BOXED band — a band whose
           first row's ink spans at least half the band's ink width: a
           bordered or filled control, a card, a hairline; never a bare
           text row, whose glyph stems would otherwise read as seams. Per
           boxed band (its `top`, `height`, `gap_above`, `edge`, `colour`
           as in `bands`): `seams`, every maximal run of columns inked over
           at least 90% of the band's height and of one modal colour
           (adjacent columns further than `--edge` apart start a new seam)
           — an outer border's side, a cell divider, a filled cell — with
           `left`, `width`, `gap_left` (columns since the previous seam)
           and `colour`; and `cells`, the spans between consecutive seams,
           each with `left`, `width` and its `lines`: every maximal run of
           rows carrying ink inside the span, rows inked across 90% of the
           span (a border, an underline) excluded, with `top`, `height`,
           `left` and `right` (the ink's inset from each cell edge) and
           `offset` (the ink's centre minus the cell's centre, negative
           when left of it). `delta.seams` pairs the boxed bands as `bands`
           does, then within each paired band pairs the seams in order by
           `since_pair` (columns since the last paired seam) and `width`,
           listing `missing` and `extra` seams and per pair the `gap_left`,
           `since_pair`, `width` and `colour` deltas; cells pair where both
           bounding seams paired consecutively, and their lines pair by
           index with `count` per image and per pair the `offset`, `left`
           and `right` deltas. Text width never moves a seam, so a missing
           seam is a divider the capture omits; a wrapped label whose lines
           are centred in the frame and left-anchored in the capture reads
           as an `offset` delta of half the slack with a `left` delta of the
           same size, while a left-anchored label whose width is data reads
           as an `offset` delta with `left` unchanged (KAN-437 fix round 5:
           the GOAL control's two dividers and its wrapped labels' centring
           both shipped wrong past every band and every sweep). A filled
           cell is a seam, not a cell — its lines are not read; a divider in
           a row with no border or fill sits in an unboxed band and is not
           read either.

Output (stdout): one JSON object — `a`, `b` (when given), `scale`, and
`delta`: for every numeric leaf, `a`, `a_scaled` (`a` × scale for lengths,
`a` unchanged for ratios and opacities), `b`, `abs` (b − a_scaled) and `pct`
(abs / a_scaled × 100, null when a_scaled is 0); for every colour, the RGB
Euclidean distance; for `bands`, `delta.bands` is the paired list above and
`delta.bands_summary` its `paired`/`missing`/`extra` counts; for `seams`,
`delta.seams` is the per-band paired list above and `delta.seams_summary`
its `paired`/`missing`/`extra` seam counts, `bands_unpaired` (boxed bands
with no counterpart) and `lines` (`paired`/`unpaired`). Thresholds are
the caller's: this script flags nothing.

Exit codes:
  0  measured.
  1  a region cannot be resolved cleanly: a corner that is not background, a
     scan line that finds no edge, a corner scan that never reaches the
     straight side, or the region's centre outside the box. The reason is on
     stderr; nothing is guessed.
  2  cannot answer: usage error, two images with no calibration, `--ref-*`
     boxes whose aspect disagrees, an unreadable image, or Pillow absent.
"""
import argparse
import json
import math
import sys
from collections import Counter

try:
    from PIL import Image
except ImportError:
    print("measure-visual-properties: Pillow is required — python3 -m pip install pillow", file=sys.stderr)
    sys.exit(2)

ALL_PROPS = ("box", "radius", "border", "fill", "shadow", "content", "gap", "ink", "runs", "bands", "seams")
# Properties that read the region as a whole page: background is the modal
# colour and the corners need not be background.
PAGE_PROPS = ("bands", "seams")
# Share of a band's height a column must ink to be a seam, and of a cell's
# width a row must ink to be a rule rather than a text line.
FULL = 0.9
# Numeric leaves that are not lengths and therefore are not scaled.
UNSCALED = ("ratio", "approx_opacity", "peak_delta", "share")


class Unresolved(Exception):
    """A region that cannot be resolved to a clean bounding box."""


def dist(a, b):
    return math.sqrt(sum((x - y) ** 2 for x, y in zip(a, b)))


def hexcolour(c):
    return "#%02x%02x%02x" % c


def parse_box(text):
    parts = text.split(",")
    if len(parts) != 4 or not all(p.strip().isdigit() for p in parts):
        raise argparse.ArgumentTypeError(f"expected x,y,w,h of non-negative integers, got {text!r}")
    x, y, w, h = (int(p) for p in parts)
    if w < 1 or h < 1:
        raise argparse.ArgumentTypeError(f"width and height must be positive, got {text!r}")
    return x, y, w, h


class Region:
    def __init__(self, path, box, edge, noise, page=False):
        try:
            with Image.open(path) as im:
                self.im = im.convert("RGB")
        except Exception as e:  # unreadable, missing, not an image
            print(f"measure-visual-properties: unreadable image {path}: {e}", file=sys.stderr)
            sys.exit(2)
        if box is None:
            box = (0, 0, self.im.width, self.im.height)
        self.x0, self.y0, self.w, self.h = box
        if self.x0 + self.w > self.im.width or self.y0 + self.h > self.im.height:
            print(f"measure-visual-properties: region {box} exceeds {path} ({self.im.width}x{self.im.height})", file=sys.stderr)
            sys.exit(2)
        self.px = self.im.load()
        self.edge, self.noise = edge, noise
        if page:
            self.bg = Counter(self.at(x, y) for y in range(self.h) for x in range(self.w)).most_common(1)[0][0]
            return
        self.bg = self.at(0, 0)
        for x, y in ((self.w - 1, 0), (0, self.h - 1), (self.w - 1, self.h - 1)):
            if dist(self.at(x, y), self.bg) > noise:
                raise Unresolved(f"region corner ({x},{y}) is not background {hexcolour(self.bg)} — widen the crop")

    def at(self, x, y):
        return self.px[self.x0 + x, self.y0 + y][:3]

    def outward(self, side):
        """Region-relative points from just outside the box's `side` to the IMAGE's
        edge — shadow and gap look past the region, so the region can be a tight
        crop and still measure the spacing to a neighbour outside it."""
        b, cx, cy = self.box, self.w // 2, self.h // 2
        return {
            "left": self.row(cy, b["left"] - 1, -self.x0),
            "right": self.row(cy, b["right"] + 1, self.im.width - 1 - self.x0),
            "top": self.col(cx, b["top"] - 1, -self.y0),
            "bottom": self.col(cx, b["bottom"] + 1, self.im.height - 1 - self.y0),
        }[side]

    def is_bg(self, c):
        return dist(c, self.bg) <= self.noise

    # --- scan lines -------------------------------------------------------

    def first_edge(self, points):
        """Index into `points` (outside → inward) of the first hard jump, or None."""
        prev = self.at(*points[0])
        for i in range(1, len(points)):
            cur = self.at(*points[i])
            if dist(cur, prev) >= self.edge:
                return i
            prev = cur
        return None

    def row(self, y, x_from, x_to):
        step = 1 if x_to >= x_from else -1
        return [(x, y) for x in range(x_from, x_to + step, step)]

    def col(self, x, y_from, y_to):
        step = 1 if y_to >= y_from else -1
        return [(x, y) for y in range(y_from, y_to + step, step)]

    # --- properties -------------------------------------------------------

    def measure_box(self):
        cx, cy = self.w // 2, self.h // 2
        found = {}
        for side, points in (
            ("left", self.row(cy, 0, cx)),
            ("right", self.row(cy, self.w - 1, cx)),
            ("top", self.col(cx, 0, cy)),
            ("bottom", self.col(cx, self.h - 1, cy)),
        ):
            i = self.first_edge(points)
            if i is None:
                raise Unresolved(f"no edge on the centre scan line from the {side} (jump >= {self.edge}); the region's centre may lie outside the control, or lower --edge")
            found[side] = points[i][0] if side in ("left", "right") else points[i][1]
        b = dict(found, width=found["right"] - found["left"] + 1, height=found["bottom"] - found["top"] + 1)
        self.box = b
        return b

    def measure_radius(self):
        b = self.box
        cx, cy = self.w // 2, self.h // 2
        limit = min(b["width"], b["height"]) // 2 + 1
        out = {}
        for name, ys, x_from, target in (
            ("top_left", range(b["top"], b["top"] + limit), 0, b["left"]),
            ("top_right", range(b["top"], b["top"] + limit), self.w - 1, b["right"]),
            ("bottom_left", range(b["bottom"], b["bottom"] - limit, -1), 0, b["left"]),
            ("bottom_right", range(b["bottom"], b["bottom"] - limit, -1), self.w - 1, b["right"]),
        ):
            # The area between the straight side and the curve is r²(1 − π/4);
            # summing each row's inset from the side and inverting that is
            # robust to antialiasing where counting rows until the inset hits
            # zero is not (it converges ~√(2r) rows early).
            area = None
            insets = []
            for y in ys:
                if y < 0 or y >= self.h:
                    break
                points = self.row(y, x_from, cx)
                i = self.first_edge(points)
                if i is None:
                    break
                inset = abs(points[i][0] - target)
                if inset == 0:
                    area = sum(insets)
                    break
                insets.append(inset)
            if area is None:
                raise Unresolved(f"{name} corner never reaches the straight side x={target} within {limit} rows — not a clean rectangle")
            out[name] = round(math.sqrt(area / (1 - math.pi / 4)))
        self.radius = out
        return out

    def interior_inset(self):
        b = self.box
        ix, iy = b["width"] // 4, b["height"] // 4
        return b["left"] + ix, b["top"] + iy, b["right"] - ix, b["bottom"] - iy

    def measure_fill(self):
        l, t, r, btm = self.interior_inset()
        counts = Counter(self.at(x, y) for y in range(t, btm + 1) for x in range(l, r + 1))
        colour, n = counts.most_common(1)[0]
        self.fill = colour
        return {"colour": hexcolour(colour), "share": round(n / sum(counts.values()), 3)}

    def is_blend(self, c, p=None, q=None):
        """`c` lies on the segment between `p` and `q` (default background and fill) — antialiasing."""
        p, q = p or self.bg, q or self.fill
        return dist(c, p) + dist(c, q) <= dist(p, q) + self.noise

    def measure_border(self):
        b = self.box
        cx, cy = self.w // 2, self.h // 2
        out = {}
        colour = None
        for side, points in (
            ("left", self.row(cy, b["left"], cx)),
            ("right", self.row(cy, b["right"], cx)),
            ("top", self.col(cx, b["top"], cy)),
            ("bottom", self.col(cx, b["bottom"], cy)),
        ):
            run = []
            for p in points:
                c = self.at(*p)
                if dist(c, self.fill) <= self.noise:
                    break
                run.append(c)
            solid = [c for c in run if not self.is_blend(c)]
            out[side] = len(solid)
            if solid and colour is None:
                colour = solid[len(solid) // 2]
        self.border = out
        self.border_colour = colour
        out = dict(out)
        out["colour"] = hexcolour(colour) if colour else None
        return out

    def measure_shadow(self):
        black = dist(self.bg, (0, 0, 0)) or 1
        out = {}
        for side in ("left", "right", "top", "bottom"):
            points = self.outward(side)
            deltas = []
            for p in points:
                d = dist(self.at(*p), self.bg)
                if d <= self.noise:
                    break
                deltas.append(d)
            peak = max(deltas) if deltas else 0
            out[side] = {"width": len(deltas), "peak_delta": round(peak, 1), "approx_opacity": round(min(1.0, peak / black), 2)}
        return out

    def measure_content(self):
        b = self.box
        bd, rad, bc = self.border, self.radius, self.border_colour
        l, t = b["left"] + bd["left"] + 2, b["top"] + bd["top"] + 2
        r, btm = b["right"] - bd["right"] - 2, b["bottom"] - bd["bottom"] - 2
        corners = (  # (x0, y0, x1, y1) squares the curve cuts out of the rectangle
            (b["left"], b["top"], b["left"] + rad["top_left"], b["top"] + rad["top_left"]),
            (b["right"] - rad["top_right"], b["top"], b["right"], b["top"] + rad["top_right"]),
            (b["left"], b["bottom"] - rad["bottom_left"], b["left"] + rad["bottom_left"], b["bottom"]),
            (b["right"] - rad["bottom_right"], b["bottom"] - rad["bottom_right"], b["right"], b["bottom"]),
        )
        xs, ys = [], []
        for y in range(t, btm + 1):
            for x in range(l, r + 1):
                if any(cx0 <= x <= cx1 and cy0 <= y <= cy1 for cx0, cy0, cx1, cy1 in corners):
                    continue
                c = self.at(x, y)
                if dist(c, self.fill) <= self.noise or self.is_blend(c) or (bc and self.is_blend(c, bc)):
                    continue
                xs.append(x)
                ys.append(y)
        if not xs:
            return {"width": None, "height": None, "ratio": None, "padding": None, "colour": None}
        cl, cr, ct, cb = min(xs), max(xs), min(ys), max(ys)
        w, h = cr - cl + 1, cb - ct + 1
        # The tint: the modal colour of the content pixels themselves. `fill`
        # is the box behind an icon, never the icon — a grey glyph on the
        # right fill reads as a match on every other property (KAN-437).
        tint = Counter(self.at(x, y) for x, y in zip(xs, ys)).most_common(1)[0][0]
        return {
            "width": w,
            "height": h,
            "colour": hexcolour(tint),
            "ratio": {"width": round(w / b["width"], 3), "height": round(h / b["height"], 3)},
            "padding": {"left": cl - b["left"], "top": ct - b["top"], "right": b["right"] - cr, "bottom": b["bottom"] - cb},
        }

    def measure_gap(self):
        out = {}
        for side in ("left", "right", "top", "bottom"):
            points = self.outward(side)
            i = 0
            while i < len(points) and not self.is_bg(self.at(*points[i])):
                i += 1  # shadow band
            while i < len(points) and self.is_bg(self.at(*points[i])):
                i += 1  # the gap itself
            out[side] = i if i < len(points) else None
        return out

    def measure_ink(self):
        xs, ys = [], []
        for y in range(self.h):
            for x in range(self.w):
                if not self.is_bg(self.at(x, y)):
                    xs.append(x)
                    ys.append(y)
        if not xs:
            return {"width": None, "height": None, "colour": None}
        # The tint of the ink itself — a text run has no box for `content` to
        # look inside, so this is the one reading of a label's colour (KAN-437
        # final verification: a caption and a summary row shipped in the wrong
        # colour on a frame whose every label had been transcribed).
        tint = Counter(self.at(x, y) for x, y in zip(xs, ys)).most_common(1)[0][0]
        return {"left": min(xs), "top": min(ys), "width": max(xs) - min(xs) + 1, "height": max(ys) - min(ys) + 1, "colour": hexcolour(tint)}

    def measure_runs(self):
        cx, cy = self.w // 2, self.h // 2
        out = {}
        for name, points in (("row", self.row(cy, 0, self.w - 1)), ("col", self.col(cx, 0, self.h - 1))):
            runs = []
            for i, p in enumerate(points):
                c = self.at(*p)
                # Compared with the run's FIRST pixel, not the previous one, so a
                # gradient breaks into runs instead of drifting into one.
                if runs and dist(c, runs[-1]["colour"]) <= self.noise:
                    runs[-1]["to"] = i
                else:
                    runs.append({"from": i, "to": i, "colour": c})
            out[name] = [
                {"from": r["from"], "to": r["to"], "length": r["to"] - r["from"] + 1, "colour": hexcolour(r["colour"])}
                for r in runs
            ]
        return out

    def measure_bands(self):
        bands = []
        prev_end = -1
        y = 0
        while y < self.h:
            row = [self.at(x, y) for x in range(self.w)]
            if all(self.is_bg(c) for c in row):
                y += 1
                continue
            top = y
            edge = Counter(c for c in row if not self.is_bg(c)).most_common(1)[0][0]
            ink = Counter()
            while y < self.h:
                row = [c for c in (self.at(x, y) for x in range(self.w)) if not self.is_bg(c)]
                if not row:
                    break
                ink.update(row)
                y += 1
            bands.append({
                "top": top,
                "height": y - top,
                "gap_above": top - prev_end - 1,
                "edge": hexcolour(edge),
                "colour": hexcolour(ink.most_common(1)[0][0]),
            })
            prev_end = y - 1
        return bands

    def measure_seams(self, bands):
        out = []
        for band in bands:
            y0, h = band["top"], band["height"]
            ys = range(y0, y0 + h)
            cols = [[c for c in (self.at(x, y) for y in ys) if not self.is_bg(c)] for x in range(self.w)]
            inked = [x for x, col in enumerate(cols) if col]
            first_row = sum(1 for x in inked if not self.is_bg(self.at(x, y0)))
            if first_row * 2 < inked[-1] - inked[0] + 1:
                continue  # a bare text row: its glyph stems are not seams
            seams = []
            for x, col in enumerate(cols):
                if len(col) < FULL * h:
                    continue
                colour = Counter(col).most_common(1)[0][0]
                if seams and seams[-1]["right"] == x - 1 and dist(colour, seams[-1]["rgb"]) < self.edge:
                    seams[-1]["right"] = x
                else:
                    seams.append({"left": x, "right": x, "rgb": colour})
            cells = []
            for s, t in zip(seams, seams[1:]):
                x0, x1 = s["right"] + 1, t["left"] - 1
                if x1 < x0:
                    continue
                span = x1 - x0 + 1
                lines, run = [], None
                for y in ys:
                    xs = [x for x in range(x0, x1 + 1) if not self.is_bg(self.at(x, y))]
                    if not xs or len(xs) >= FULL * span:  # empty, or a rule across the cell
                        run = None
                        continue
                    if run is None:
                        run = {"top": y, "height": 0, "l": xs[0], "r": xs[-1]}
                        lines.append(run)
                    run["height"] += 1
                    run["l"], run["r"] = min(run["l"], xs[0]), max(run["r"], xs[-1])
                cells.append({
                    "left": x0,
                    "width": span,
                    "lines": [{
                        "top": ln["top"] - y0,
                        "height": ln["height"],
                        "left": ln["l"] - x0,
                        "right": x1 - ln["r"],
                        "offset": round((ln["l"] + ln["r"]) / 2 - (x0 + x1) / 2, 1),
                    } for ln in lines],
                })
            prev = -1
            for s in seams:
                s["width"] = s.pop("right") - s["left"] + 1
                s["gap_left"] = s["left"] - prev - 1
                s["colour"] = hexcolour(s.pop("rgb"))
                prev = s["left"] + s["width"] - 1
            out.append(dict(band, seams=seams, cells=cells))
        return out

    def measure(self, props):
        out = {"background": hexcolour(self.bg)}
        needs_box = [p for p in props if p not in ("ink", "runs", "bands", "seams")]
        if needs_box:
            box = self.measure_box()
            if "box" in props:
                out["box"] = box
            if "radius" in props or "content" in props:
                radius = self.measure_radius()
                if "radius" in props:
                    out["radius"] = radius
            if any(p in props for p in ("fill", "border", "content")):
                fill = self.measure_fill()
                if "fill" in props:
                    out["fill"] = fill
            if "border" in props or "content" in props:
                border = self.measure_border()
                if "border" in props:
                    out["border"] = border
            if "shadow" in props:
                out["shadow"] = self.measure_shadow()
            if "content" in props:
                out["content"] = self.measure_content()
            if "gap" in props:
                out["gap"] = self.measure_gap()
        if "ink" in props:
            out["ink"] = self.measure_ink()
        if "runs" in props:
            out["runs"] = self.measure_runs()
        if "bands" in props or "seams" in props:
            bands = self.measure_bands()
            if "bands" in props:
                out["bands"] = bands
            if "seams" in props:
                out["seams"] = self.measure_seams(bands)
        return out


def leaves(d, prefix=""):
    for k, v in d.items():
        key = f"{prefix}{k}"
        if isinstance(v, dict):
            yield from leaves(v, key + ".")
        else:
            yield key, v


def colour_delta(av, bv):
    ca = tuple(int(av[i:i + 2], 16) for i in (1, 3, 5))
    cb = tuple(int(bv[i:i + 2], 16) for i in (1, 3, 5))
    return {"a": av, "b": bv, "distance": round(dist(ca, cb), 1)}


def number_delta(av, bv, scale):
    scaled = av * scale
    ab = bv - scaled
    return {"a": av, "a_scaled": round(scaled, 2), "b": bv, "abs": round(ab, 2), "pct": None if scaled == 0 else round(ab / scaled * 100, 1)}


def pair_in_order(a, b, scale, pos, size, gap, colours):
    """Pair the frame's bands (or seams) with the capture's in order along
    one axis — `pos`/`size`/`gap` name the record's keys. A record pairs
    when its gap since the last pair and its size both match within a
    tolerance of 8px plus a quarter of the larger record, so a padding
    defect shows once as that pair's gap delta and the records after it
    still pair, instead of every one after reading as shifted — and a 2px
    rule never pairs with a text row that happens to sit where the rule was.

    ponytail: greedy in-order pairing; a shift larger than the tolerance
    unpairs one record and resyncs on the next — enough for a page of rows,
    replace with sequence alignment if frames with many near-identical
    bands mis-pair."""
    out, i, j = [], 0, 0
    a_anchor, b_anchor = -1, -1  # last row/column of the last paired record, per image
    while i < len(a) or j < len(b):
        if i < len(a) and j < len(b):
            ga = (a[i][pos] - a_anchor - 1) * scale
            gb = b[j][pos] - b_anchor - 1
            tol = 8 + max(a[i][size] * scale, b[j][size]) / 4
            if abs(gb - ga) <= tol and abs(b[j][size] - a[i][size] * scale) <= tol:
                pair = {"status": "paired", "a": a[i], "b": b[j], gap: number_delta(a[i][gap], b[j][gap], scale)}
                # Distance from the last PAIRED record, so a padding lost
                # around a missing divider still reads as one number here.
                pair["since_pair"] = number_delta(a[i][pos] - a_anchor - 1, gb, scale)
                pair[size] = number_delta(a[i][size], b[j][size], scale)
                for c in colours:
                    pair[c] = colour_delta(a[i][c], b[j][c])
                out.append(pair)
                a_anchor, b_anchor = a[i][pos] + a[i][size] - 1, b[j][pos] + b[j][size] - 1
                i += 1
                j += 1
                continue
            if ga < gb:
                out.append({"status": "missing", "a": a[i], "b": None})
                i += 1
            else:
                out.append({"status": "extra", "a": None, "b": b[j]})
                j += 1
        elif i < len(a):
            out.append({"status": "missing", "a": a[i], "b": None})
            i += 1
        else:
            out.append({"status": "extra", "a": None, "b": b[j]})
            j += 1
    return out


def pair_bands(a, b, scale):
    return pair_in_order(a, b, scale, "top", "height", "gap_above", ("edge", "colour"))


def count_status(pairs):
    return {s: sum(1 for p in pairs if p["status"] == s) for s in ("paired", "missing", "extra")}


def pair_seams(a, b, scale):
    """Pair the boxed bands, then each band's seams; a cell pairs where its
    two bounding seams paired with consecutive seams in the other image, and
    its lines pair by index."""
    out = []
    summary = {"paired": 0, "missing": 0, "extra": 0, "bands_unpaired": 0, "lines": {"paired": 0, "unpaired": 0}}
    for band in pair_bands(a, b, scale):
        entry = {"status": band["status"], "a": band["a"] and {"top": band["a"]["top"], "height": band["a"]["height"]}, "b": band["b"] and {"top": band["b"]["top"], "height": band["b"]["height"]}}
        if band["status"] != "paired":
            summary["bands_unpaired"] += 1
            out.append(entry)
            continue
        sa, sb = band["a"]["seams"], band["b"]["seams"]
        seams = pair_in_order(sa, sb, scale, "left", "width", "gap_left", ("colour",))
        for k, v in count_status(seams).items():
            summary[k] += v
        # Index of each paired seam in its own list, in order.
        idx = [(sa.index(p["a"]), sb.index(p["b"])) for p in seams if p["status"] == "paired"]
        cells = []
        for (ia, ib), (ja, jb) in zip(idx, idx[1:]):
            if ja != ia + 1 or jb != ib + 1:
                continue
            ca, cb = band["a"]["cells"][ia], band["b"]["cells"][ib]
            lines = [{
                "offset": number_delta(la["offset"], lb["offset"], scale),
                "left": number_delta(la["left"], lb["left"], scale),
                "right": number_delta(la["right"], lb["right"], scale),
            } for la, lb in zip(ca["lines"], cb["lines"])]
            summary["lines"]["paired"] += len(lines)
            summary["lines"]["unpaired"] += abs(len(ca["lines"]) - len(cb["lines"]))
            cells.append({"a": ia, "b": ib, "count": {"a": len(ca["lines"]), "b": len(cb["lines"])}, "lines": lines})
        out.append(dict(entry, seams=seams, cells=cells))
    return out, summary


def delta(a, b, scale):
    out = {}
    if "bands" in a and "bands" in b:
        pairs = pair_bands(a["bands"], b["bands"], scale)
        out["bands"] = pairs
        out["bands_summary"] = count_status(pairs)
    if "seams" in a and "seams" in b:
        out["seams"], out["seams_summary"] = pair_seams(a["seams"], b["seams"], scale)
    bl = dict(leaves(b))
    for key, av in leaves(a):
        bv = bl.get(key)
        if key in ("bands", "seams"):
            continue
        if isinstance(av, str) and av.startswith("#") and isinstance(bv, str) and bv.startswith("#"):
            out[key] = colour_delta(av, bv)
        elif isinstance(av, (int, float)) and isinstance(bv, (int, float)):
            out[key] = number_delta(av, bv, 1 if any(seg in UNSCALED for seg in key.split(".")) else scale)
        elif av is None or bv is None:
            out[key] = {"a": av, "b": bv, "abs": None, "pct": None}
    return out


def main(argv):
    p = argparse.ArgumentParser(prog="measure-visual-properties.sh", add_help=True)
    p.add_argument("image_a")
    p.add_argument("image_b", nargs="?")
    p.add_argument("--region-a", type=parse_box)
    p.add_argument("--region-b", type=parse_box)
    p.add_argument("--scale", type=float)
    p.add_argument("--ref-a", type=parse_box)
    p.add_argument("--ref-b", type=parse_box)
    p.add_argument("--props", default=",".join(ALL_PROPS))
    p.add_argument("--edge", type=float, default=24)
    p.add_argument("--noise", type=float, default=8)
    try:
        args = p.parse_args(argv[1:])
    except SystemExit as e:
        sys.exit(0 if e.code == 0 else 2)

    props = [s.strip() for s in args.props.split(",") if s.strip()]
    unknown = [s for s in props if s not in ALL_PROPS]
    if unknown or not props:
        print(f"measure-visual-properties: unknown --props {unknown}; choose from {','.join(ALL_PROPS)}", file=sys.stderr)
        sys.exit(2)

    scale = None
    if args.image_b:
        if args.scale is not None and (args.ref_a or args.ref_b):
            print("measure-visual-properties: give --scale or the --ref-a/--ref-b pair, not both", file=sys.stderr)
            sys.exit(2)
        if args.scale is not None:
            if args.scale <= 0:
                print("measure-visual-properties: --scale must be positive", file=sys.stderr)
                sys.exit(2)
            scale = args.scale
        elif args.ref_a and args.ref_b:
            scale = args.ref_b[2] / args.ref_a[2]
            by_height = args.ref_b[3] / args.ref_a[3]
            if abs(by_height - scale) / scale > 0.05:
                print(f"measure-visual-properties: --ref boxes disagree: width factor {scale:.4f}, height factor {by_height:.4f}", file=sys.stderr)
                sys.exit(2)
        else:
            print("measure-visual-properties: two images need a calibration — --scale <B px per A px>, or --ref-a and --ref-b boxes already known correct in both images", file=sys.stderr)
            sys.exit(2)
    elif args.scale is not None or args.ref_a or args.ref_b or args.region_b:
        print("measure-visual-properties: --scale/--ref-*/--region-b need a second image", file=sys.stderr)
        sys.exit(2)

    page = all(p in PAGE_PROPS for p in props)
    result = {}
    try:
        result["a"] = Region(args.image_a, args.region_a, args.edge, args.noise, page).measure(props)
        if args.image_b:
            result["b"] = Region(args.image_b, args.region_b, args.edge, args.noise, page).measure(props)
            result["scale"] = scale
            result["delta"] = delta(result["a"], result["b"], scale)
    except Unresolved as e:
        which = "b" if "a" in result else "a"
        print(f"measure-visual-properties: image {which}: {e}", file=sys.stderr)
        sys.exit(1)

    json.dump(result, sys.stdout, indent=2)
    print()
    sys.exit(0)


if __name__ == "__main__":
    main(sys.argv)
