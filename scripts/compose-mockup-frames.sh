#!/usr/bin/env bash
# compose-mockup-frames.sh — pair each captured screenshot with the mockup
# frame it was drawn from, side by side, one composite PNG per pair.
#
# Interface: `compose-mockup-frames.sh <map file> <mockups root> <out dir>
# [scale=<n> status=<px> border=<px>]`, resolved capture PNG paths on stdin,
# one per line (design.md section 4). The fourth argument is the mockups'
# declared frame geometry — the project's `mockup frame` row — and is
# optional: without it nothing is cropped, no size check runs, and the
# composite keeps its two native-size panels.
#
# The map file is a `<spec>.mockups` sidecar: one `<screenshot name> <frame
# id>` pair per line, whitespace-separated; blank lines and `#`-leading lines
# are ignored (design.md section 3 — canonical for the map's shape and why
# it is explicit rather than a filename convention).
#
# Layout (design.md section 4): captured frame on the left, mockup on the
# right, both native size, top-aligned, a 16px gutter, plain (40, 40, 40)
# background, no labels, no scaling. Output file is
# `<out dir>/<screenshot name without .png>.png`.
#
# With the geometry declared the layout is three panels — capture, cropped
# frame, difference — on the same gutter and background. Each frame is
# cropped to its content area: the status line and the border come from the
# declared values, and the bottom comes from the frame itself — at the
# page-coloured margin above the caption band, or, where the caption band
# starts directly below the frame, at the frame's own bottom border (both
# rules in cmfCropBox). It is detected rather than declared because
# content-hugging frames differ in height from one another. The
# difference panel is white wherever any channel differs and black
# elsewhere, so a departure reads as shape rather than as a colour blend.
# The cropped frame itself is also written as `<out dir>/<screenshot name
# without .png>.frame.png` — the same size as the capture, so
# `measure-visual-properties.sh <that>.frame.png <capture> --scale 1` needs no
# region arithmetic to compare the two.
#
# Exit codes:
#   0  every map line composed; one `<composite path> diff=<ratio>` per line
#      on stdout — the differing-pixel ratio to four decimals, or the literal
#      `n/a` when no geometry was declared and no difference was computed.
#      The ratio is information for whoever reads the composite; nothing in
#      this script or its callers treats it as a threshold.
#   1  a map line names a screenshot no stdin path matches, a frame file
#      absent under the mockups root, a malformed map line, or — with the
#      geometry declared — a capture whose size differs from the cropped
#      frame, or a geometry that leaves the frame no content area. Every
#      finding is printed as `<map>:<line>: <message>` to stderr; every
#      other, well-formed line is still composed.
#   2  cannot answer: usage error (including a malformed geometry argument),
#      the map unreadable or not UTF-8, the mockups root not a directory, the
#      output directory not creatable, stdin not UTF-8, an unreadable PNG, or
#      an output PNG not writable.
#
# The body is Go: stats/internal/guard/composemockupframes.go.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "compose-mockup-frames: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec compose-mockup-frames 2 "compose-mockup-frames:" "$@"
