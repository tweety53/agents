#!/usr/bin/env python3
"""compose-mockup-frames.py — pair each captured screenshot with the
mockup frame it was drawn from, side by side, one composite PNG per pair.

Interface: `compose-mockup-frames.sh <map file> <mockups root> <out dir>`,
resolved capture PNG paths on stdin, one per line (design.md section 4).

The map file is a `<spec>.mockups` sidecar: one `<screenshot name> <frame
id>` pair per line, whitespace-separated; blank lines and `#`-leading lines
are ignored (design.md section 3 — canonical for the map's shape and why
it is explicit rather than a filename convention).

Layout (design.md section 4): captured frame on the left, mockup on the
right, both native size, top-aligned, a 16px gutter, plain (40, 40, 40)
background, no labels, no scaling. Output file is
`<out dir>/<screenshot name without .png>.png`.

Exit codes:
  0  every map line composed; one `<composite path>` per line on stdout.
  1  a map line names a screenshot no stdin path matches, a frame file
     absent under the mockups root, or a malformed map line. Every finding
     is printed as `<map>:<line>: <message>` to stderr; every other,
     well-formed line is still composed.
  2  cannot answer: usage error, the map unreadable, the mockups root not
     a directory, the output directory not creatable, an unreadable PNG,
     or Pillow absent.
"""
import os
import sys

GUTTER = 16
BACKGROUND = (40, 40, 40)

try:
    from PIL import Image
except ImportError:
    print(
        "compose-mockup-frames: Pillow is required — python3 -m pip install pillow",
        file=sys.stderr,
    )
    sys.exit(2)


def usage_error():
    print(
        "usage: compose-mockup-frames.sh <map file> <mockups root> <out dir>"
        "   (resolved PNG paths on stdin, one per line)",
        file=sys.stderr,
    )
    sys.exit(2)


def png_stem(name):
    """`name` without its `.png` suffix."""
    return name[: -len(".png")]


def match_capture(name, stdin_paths):
    """Return the stdin path(s) whose basename is `name` or
    `<stem>-<anything>.png` where `<stem>` is `name` without `.png`
    (Playwright's `-<platform>` suffix)."""
    stem = png_stem(name)
    matches = []
    for path in stdin_paths:
        base = os.path.basename(path)
        if base == name:
            matches.append(path)
        elif base.startswith(stem + "-") and base.endswith(".png"):
            matches.append(path)
    return matches


def main(argv):
    if len(argv) != 4:
        usage_error()

    map_path, mockups_root, out_dir = argv[1], argv[2], argv[3]

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
            canvas = Image.new(
                "RGB",
                (cap.width + GUTTER + mock.width, max(cap.height, mock.height)),
                BACKGROUND,
            )
            canvas.paste(cap, (0, 0))
            canvas.paste(mock, (cap.width + GUTTER, 0))
            canvas.save(out_path)

        outputs.append(os.path.abspath(out_path))

    for out_path in outputs:
        print(out_path)

    for finding in findings:
        print(finding, file=sys.stderr)

    sys.exit(1 if findings else 0)


if __name__ == "__main__":
    main(sys.argv)
