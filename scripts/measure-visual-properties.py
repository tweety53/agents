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
                          shadow,content,gap,ink,runs (default: all)
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
           as a ratio of the box, and the padding from each box edge to it.
  gap      per side: background pixels between the box edge (past any
           shadow band) and the next non-background pixel on the same scan
           line, out to the image's edge — the spacing to the nearest
           neighbour; `null` when the scan line reaches the image edge.
  ink      bounding box of every non-background pixel in the region. Needs
           no control edge: crop to a run of capital letters and request
           `--props ink` — its height is the cap-height, the font-size
           stand-in.
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

Output (stdout): one JSON object — `a`, `b` (when given), `scale`, and
`delta`: for every numeric leaf, `a`, `a_scaled` (`a` × scale for lengths,
`a` unchanged for ratios and opacities), `b`, `abs` (b − a_scaled) and `pct`
(abs / a_scaled × 100, null when a_scaled is 0); for every colour, the RGB
Euclidean distance. Thresholds are the caller's: this script flags nothing.

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

ALL_PROPS = ("box", "radius", "border", "fill", "shadow", "content", "gap", "ink", "runs")
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
    def __init__(self, path, box, edge, noise):
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
            return {"width": None, "height": None, "ratio": None, "padding": None}
        cl, cr, ct, cb = min(xs), max(xs), min(ys), max(ys)
        w, h = cr - cl + 1, cb - ct + 1
        return {
            "width": w,
            "height": h,
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
            return {"width": None, "height": None}
        return {"left": min(xs), "top": min(ys), "width": max(xs) - min(xs) + 1, "height": max(ys) - min(ys) + 1}

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

    def measure(self, props):
        out = {"background": hexcolour(self.bg)}
        needs_box = [p for p in props if p not in ("ink", "runs")]
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
        return out


def leaves(d, prefix=""):
    for k, v in d.items():
        key = f"{prefix}{k}"
        if isinstance(v, dict):
            yield from leaves(v, key + ".")
        else:
            yield key, v


def delta(a, b, scale):
    out = {}
    bl = dict(leaves(b))
    for key, av in leaves(a):
        bv = bl.get(key)
        if isinstance(av, str) and av.startswith("#") and isinstance(bv, str) and bv.startswith("#"):
            ca = tuple(int(av[i:i + 2], 16) for i in (1, 3, 5))
            cb = tuple(int(bv[i:i + 2], 16) for i in (1, 3, 5))
            out[key] = {"a": av, "b": bv, "distance": round(dist(ca, cb), 1)}
        elif isinstance(av, (int, float)) and isinstance(bv, (int, float)):
            scaled = av if any(seg in UNSCALED for seg in key.split(".")) else av * scale
            ab = bv - scaled
            out[key] = {
                "a": av,
                "a_scaled": round(scaled, 2),
                "b": bv,
                "abs": round(ab, 2),
                "pct": None if scaled == 0 else round(ab / scaled * 100, 1),
            }
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

    result = {}
    try:
        result["a"] = Region(args.image_a, args.region_a, args.edge, args.noise).measure(props)
        if args.image_b:
            result["b"] = Region(args.image_b, args.region_b, args.edge, args.noise).measure(props)
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
