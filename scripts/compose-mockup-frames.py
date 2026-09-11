#!/usr/bin/env python3
"""compose-mockup-frames.py — pair each captured screenshot with the
mockup frame it was drawn from, side by side, one composite PNG per pair.

Interface: `compose-mockup-frames.sh <map file> <mockups root> <out dir>
[scale=<n> status=<px> border=<px>]`, resolved capture PNG paths on stdin,
one per line (design.md section 4). The fourth argument is the mockups'
declared frame geometry — the project's `mockup frame` row — and is
optional: without it nothing is cropped, no size check runs, and the
composite keeps its two native-size panels.

The map file is a `<spec>.mockups` sidecar: one `<screenshot name> <frame
id>` pair per line, whitespace-separated; blank lines and `#`-leading lines
are ignored (design.md section 3 — canonical for the map's shape and why
it is explicit rather than a filename convention).

Layout (design.md section 4): captured frame on the left, mockup on the
right, both native size, top-aligned, a 16px gutter, plain (40, 40, 40)
background, no labels, no scaling. Output file is
`<out dir>/<screenshot name without .png>.png`.

With the geometry declared the layout is three panels — capture, cropped
frame, difference — on the same gutter and background. Each frame is
cropped to its content area: the status line and the border come from the
declared values, and the bottom comes from the frame itself, as the first
row below the status line every one of whose pixels between the side
borders equals the page colour at (0, 0). That row is the page-coloured
margin above the caption band, and it is detected rather than declared
because content-hugging frames differ in height from one another. The
difference panel is white wherever any channel differs and black
elsewhere, so a departure reads as shape rather than as a colour blend.

Exit codes:
  0  every map line composed; one `<composite path> diff=<ratio>` per line
     on stdout — the differing-pixel ratio to four decimals, or the literal
     `n/a` when no geometry was declared and no difference was computed.
     The ratio is information for whoever reads the composite; nothing in
     this script or its callers treats it as a threshold.
  1  a map line names a screenshot no stdin path matches, a frame file
     absent under the mockups root, a malformed map line, or — with the
     geometry declared — a capture whose size differs from the cropped
     frame, or a geometry that leaves the frame no content area. Every
     finding is printed as `<map>:<line>: <message>` to stderr; every
     other, well-formed line is still composed.
  2  cannot answer: usage error (including a malformed geometry argument),
     the map unreadable, the mockups root not a directory, the output
     directory not creatable, an unreadable PNG, or Pillow absent.
"""
import os
import sys

GUTTER = 16
BACKGROUND = (40, 40, 40)

try:
    from PIL import Image, ImageChops
except ImportError:
    print(
        "compose-mockup-frames: Pillow is required — python3 -m pip install pillow",
        file=sys.stderr,
    )
    sys.exit(2)


def usage_error():
    print(
        "usage: compose-mockup-frames.sh <map file> <mockups root> <out dir>"
        " [scale=<n> status=<px> border=<px>]"
        "   (resolved PNG paths on stdin, one per line)",
        file=sys.stderr,
    )
    sys.exit(2)


def png_stem(name):
    """`name` without its `.png` suffix."""
    return name[: -len(".png")]


# Node's process.platform values (https://nodejs.org/api/process.html#processplatform) —
# the exhaustive set Playwright's own snapshot suffix is drawn from. Matching against this
# closed vocabulary, rather than "any continuation", is what keeps one screenshot name from
# being mistaken for another's platform-suffixed capture merely because it is a string prefix
# of it (`add-participant-finished-excluded.png` vs.
# `add-participant-finished-excluded-added-darwin.png` — a real collision, not hypothetical).
PLATFORM_SUFFIXES = ("aix", "android", "darwin", "freebsd", "linux", "openbsd", "sunos", "win32")


def match_capture(name, stdin_paths):
    """Return the stdin path(s) whose basename is `name` or `<stem>-<platform>.png` where
    `<stem>` is `name` without `.png` and `<platform>` is exactly one Node `process.platform`
    value (Playwright's own snapshot suffix). Deliberately not `<stem>-<anything>.png`: that
    loosely matches a longer, unrelated screenshot name's own capture whenever it happens to
    start with this name's stem followed by a dash."""
    stem = png_stem(name)
    matches = []
    for path in stdin_paths:
        base = os.path.basename(path)
        if base == name:
            matches.append(path)
        elif base.endswith(".png") and png_stem(base) in (f"{stem}-{p}" for p in PLATFORM_SUFFIXES):
            matches.append(path)
    return matches


GEOMETRY_KEYS = ("scale", "status", "border")


def parse_geometry(arg):
    """`scale=<int> status=<int> border=<int>` -> {"scale": …, "status": …, "border": …}.

    Every key required exactly once, in any order; every value a non-negative
    integer and `scale` at least 1. Anything else is a usage error rather than
    a silently-ignored argument: a geometry the caller meant to declare and
    mistyped must not compose uncropped as though none had been given."""
    values = {}
    for field in arg.split():
        key, sep, raw = field.partition("=")
        if not sep or key not in GEOMETRY_KEYS or key in values:
            usage_error()
        if not raw.isdigit():
            usage_error()
        values[key] = int(raw)
    if set(values) != set(GEOMETRY_KEYS) or values["scale"] < 1:
        usage_error()
    return values


def crop_box(frame, geometry):
    """The frame's content area in PNG pixels, or None when the declared
    geometry leaves none.

    The sides and the top come from the declared values. The bottom is found
    in the image: the page background is the pixel at (0, 0) — outside the
    frame's rounded corner, so always the page and never the border — and the
    first row below the top crop whose every pixel between the side borders
    equals it is the margin above the caption band. No such row means the
    frame runs to the last row."""
    s = geometry["scale"]
    left = geometry["border"] * s
    top = (geometry["border"] + geometry["status"]) * s
    right = frame.width - geometry["border"] * s
    if right <= left or top >= frame.height:
        return None

    page = frame.getpixel((0, 0))
    pixels = frame.load()
    frame_bottom = frame.height
    for y in range(top, frame.height):
        if all(pixels[x, y] == page for x in range(left, right)):
            frame_bottom = y
            break

    bottom = frame_bottom - geometry["border"] * s
    if bottom <= top:
        return None
    return (left, top, right, bottom)


def difference_panel(cap, frame):
    """(panel, ratio) — white wherever any channel differs, black elsewhere,
    and the share of differing pixels."""
    mask = ImageChops.difference(cap, frame).convert("L").point(lambda v: 255 if v else 0)
    differing = mask.histogram()[255]
    return mask.convert("RGB"), differing / (cap.width * cap.height)


def main(argv):
    if len(argv) not in (4, 5):
        usage_error()

    map_path, mockups_root, out_dir = argv[1], argv[2], argv[3]
    geometry = parse_geometry(argv[4]) if len(argv) == 5 else None

    try:
        with open(map_path, "r", encoding="utf-8") as f:
            map_lines = f.readlines()
    except OSError as e:
        print(f"compose-mockup-frames: cannot read map file {map_path}: {e}", file=sys.stderr)
        sys.exit(2)

    if not os.path.isdir(mockups_root):
        print(f"compose-mockup-frames: mockups root is not a directory: {mockups_root}", file=sys.stderr)
        sys.exit(2)

    try:
        os.makedirs(out_dir, exist_ok=True)
    except OSError as e:
        print(f"compose-mockup-frames: cannot create output directory {out_dir}: {e}", file=sys.stderr)
        sys.exit(2)

    stdin_paths = [line.strip() for line in sys.stdin if line.strip()]
    for path in stdin_paths:
        try:
            with Image.open(path) as im:
                im.load()
        except Exception as e:
            print(f"compose-mockup-frames: unreadable PNG on stdin: {path}: {e}", file=sys.stderr)
            sys.exit(2)

    findings = []
    seen_names = set()
    outputs = []

    for lineno, raw in enumerate(map_lines, start=1):
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        fields = line.split()
        if len(fields) != 2:
            findings.append(
                f"{map_path}:{lineno}: expected `<screenshot name> <frame id>`, got {len(fields)} fields"
            )
            continue

        name, frame_id = fields

        if name in seen_names:
            findings.append(f"{map_path}:{lineno}: duplicate screenshot name `{name}`")
            continue
        seen_names.add(name)

        if not name.endswith(".png"):
            findings.append(f"{map_path}:{lineno}: expected `<screenshot name> <frame id>`, `{name}` does not end in .png")
            continue

        matches = match_capture(name, stdin_paths)
        if len(matches) == 0:
            findings.append(f"{map_path}:{lineno}: no captured PNG matches `{name}`")
            continue
        if len(matches) > 1:
            findings.append(
                f"{map_path}:{lineno}: multiple captured PNGs match `{name}`: {', '.join(matches)}"
            )
            continue
        capture_path = matches[0]

        frame_path = os.path.join(mockups_root, f"{frame_id}.png")
        if not os.path.isfile(frame_path):
            findings.append(f"{map_path}:{lineno}: no frame `{frame_id}.png` under {mockups_root}")
            continue

        stem = png_stem(name)
        out_path = os.path.join(out_dir, f"{stem}.png")

        try:
            with Image.open(frame_path) as probe:
                probe.load()
        except Exception as e:
            print(f"compose-mockup-frames: unreadable PNG: {frame_path}: {e}", file=sys.stderr)
            sys.exit(2)

        with Image.open(capture_path) as cap_im, Image.open(frame_path) as mock_im:
            cap = cap_im.convert("RGB")
            mock = mock_im.convert("RGB")

            if geometry is None:
                canvas = Image.new(
                    "RGB",
                    (cap.width + GUTTER + mock.width, max(cap.height, mock.height)),
                    BACKGROUND,
                )
                canvas.paste(cap, (0, 0))
                canvas.paste(mock, (cap.width + GUTTER, 0))
                canvas.save(out_path)
                outputs.append((os.path.abspath(out_path), None))
                continue

            box = crop_box(mock, geometry)
            if box is None:
                findings.append(
                    f"{map_path}:{lineno}: frame {frame_id}: declared geometry leaves no content area"
                )
                continue
            mock = mock.crop(box)

            if cap.size != mock.size:
                findings.append(
                    f"{map_path}:{lineno}: capture {cap.width}\u00d7{cap.height} vs "
                    f"frame {mock.width}\u00d7{mock.height} for {frame_id}"
                )
                continue

            diff, ratio = difference_panel(cap, mock)
            canvas = Image.new(
                "RGB",
                (cap.width + GUTTER + mock.width + GUTTER + diff.width, cap.height),
                BACKGROUND,
            )
            canvas.paste(cap, (0, 0))
            canvas.paste(mock, (cap.width + GUTTER, 0))
            canvas.paste(diff, (cap.width + GUTTER + mock.width + GUTTER, 0))
            canvas.save(out_path)
            outputs.append((os.path.abspath(out_path), ratio))

    for out_path, ratio in outputs:
        print(f"{out_path} diff={'n/a' if ratio is None else format(ratio, '.4f')}")

    for finding in findings:
        print(finding, file=sys.stderr)

    sys.exit(1 if findings else 0)


if __name__ == "__main__":
    main(sys.argv)
