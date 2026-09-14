#!/usr/bin/env bash
# Assertion harness for measure-visual-properties.sh / .py — the `runs`
# property, on the case that made it exist (KAN-30 manual re-sweep): a
# selected row's fill inside a bordered container, once inset from the
# border by a gutter and once flush against it. Each fixture is the
# incident's own geometry — sheet grey, a 1px neutral300 border, an accent100
# row fill, text runs in a second, unselected row — so the assertion is on
# the reading the tool gives of the real shape, not of a bare rectangle.
#
# Needs Pillow exactly as the script does; if Pillow is absent this harness
# FAILS, never skips (`.flow/project.md`'s "vacuous pass" rule for guards).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/measure-visual-properties.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

if ! python3 -c 'from PIL import Image' >/dev/null 2>&1; then
  fail "Pillow is not importable on this machine — this harness needs Pillow exactly as the script does, so it fails rather than skipping"
  echo "FAILURES: $FAILURES"
  exit 1
fi

DIR="$(mktemp -d "${TMPDIR:-/tmp}/measure-visual.XXXXXX")"
trap 'rm -rf "$DIR"' EXIT

# make_fixture <path> <inset px>: a 440x140 sheet, a 390x100 bordered
# container at (20,20), a 44px selected row whose fill is inset by <inset>
# on every side, and text runs in the unselected row below it.
make_fixture() {
  python3 - "$1" "$2" <<'PY'
import sys
from PIL import Image, ImageDraw
path, inset = sys.argv[1], int(sys.argv[2])
BG, BORDER, FILL, TEXT = (0xEA, 0xE9, 0xE9), (0xD7, 0xD3, 0xD3), (0xE9, 0xF8, 0xFF), (0x22, 0x22, 0x22)
im = Image.new("RGB", (440, 140), BG)
d = ImageDraw.Draw(im)
d.rounded_rectangle((20, 20, 409, 119), radius=2, outline=BORDER, width=1)
d.rounded_rectangle((21 + inset, 21 + inset, 408 - inset, 64 - inset), radius=2, fill=FILL)
d.rectangle((31, 74, 200, 83), fill=TEXT)
d.rectangle((31, 90, 150, 96), fill=TEXT)
im.save(path)
PY
}

# runs_after_border <image> <row|col>: "<length> <colour>" of the run that
# follows the first border-coloured run on that scan line — the reading
# "flush" (fill colour) or "inset" (sheet colour, length = the gutter).
runs_after_border() {
  "$GUARD" "$1" --region-a 10,10,420,120 --props runs | python3 -c '
import json, sys
runs = json.load(sys.stdin)["a"]["runs"][sys.argv[1]]
i = next(i for i, r in enumerate(runs) if r["colour"] == "#d7d3d3")
print(runs[i + 1]["length"], runs[i + 1]["colour"])
' "$2"
}

# Case 1: the incident — fill inset a 10px gutter from the border. The run
# after the border is 10px of sheet colour, the inset itself, on the column
# through the selected row.
make_fixture "$DIR/inset.png" 10
got="$(runs_after_border "$DIR/inset.png" col)"
if [ "$got" = "10 #eae9e9" ]; then
  pass "case 1: inset fill reads as a 10px sheet-coloured run after the border"
else
  fail "case 1: expected '10 #eae9e9' after the border, got '$got'"
fi

# Case 2: the fix — fill flush with the border. The run after the border is
# the fill itself, all 44px of the row.
make_fixture "$DIR/flush.png" 0
got="$(runs_after_border "$DIR/flush.png" col)"
if [ "$got" = "44 #e9f8ff" ]; then
  pass "case 2: flush fill is the run immediately after the border, the full row height"
else
  fail "case 2: expected '44 #e9f8ff' after the border, got '$got'"
fi

# Case 3: `runs` needs no clean box — a region whose centre line finds no
# hard edge (a flat sheet) still measures, where `box` exits 1.
python3 - "$DIR/flat.png" <<'PY'
import sys
from PIL import Image
Image.new("RGB", (60, 60), (0xEA, 0xE9, 0xE9)).save(sys.argv[1])
PY
if "$GUARD" "$DIR/flat.png" --props runs >/dev/null 2>&1; then
  pass "case 3: runs measures a region with no edge"
else
  fail "case 3: runs exited non-zero on an edgeless region"
fi
if "$GUARD" "$DIR/flat.png" --props box >/dev/null 2>&1; then
  fail "case 3: box unexpectedly resolved an edgeless region"
else
  pass "case 3: box still exits non-zero on the same region"
fi

# Case 4: two images with a calibration — `runs` carries no delta and the
# comparison still succeeds.
if "$GUARD" "$DIR/inset.png" "$DIR/flush.png" --region-a 10,10,420,120 --region-b 10,10,420,120 --scale 1 --props runs \
  | python3 -c 'import json, sys; d = json.load(sys.stdin); sys.exit(0 if not any(k.startswith("runs") for k in d["delta"]) and "runs" in d["b"] else 1)'; then
  pass "case 4: a two-image runs comparison has both lists and no runs delta"
else
  fail "case 4: two-image runs comparison failed"
fi

# Case 5: `content.colour` is the glyph's tint, not the box's fill (KAN-437:
# two icons shipped grey where the mockup drew them accent blue, matching on
# every other property). Same box, same fill, a 12px disc in the middle —
# accent blue in one fixture, grey in the other.
make_icon() {
  python3 - "$1" "$2" <<'PY'
import sys
from PIL import Image, ImageDraw
path, tint = sys.argv[1], tuple(int(sys.argv[2][i:i + 2], 16) for i in (0, 2, 4))
im = Image.new("RGB", (80, 80), (0xEA, 0xE9, 0xE9))
d = ImageDraw.Draw(im)
d.rectangle((16, 16, 63, 63), fill=(0xFF, 0xFF, 0xFF))
d.ellipse((34, 34, 45, 45), fill=tint)
im.save(path)
PY
}
make_icon "$DIR/icon-blue.png" 1e6fe0
make_icon "$DIR/icon-grey.png" 9e9e9e
got="$("$GUARD" "$DIR/icon-blue.png" "$DIR/icon-grey.png" --scale 1 --props content | python3 -c '
import json, sys
d = json.load(sys.stdin)
print(d["a"]["content"]["colour"], d["b"]["content"]["colour"], round(d["delta"]["content.colour"]["distance"]))
')"
if [ "$got" = "#1e6fe0 #9e9e9e 152" ]; then
  pass "case 5: content.colour reads the glyph tint and the delta carries the distance between two tints"
else
  fail "case 5: expected '#1e6fe0 #9e9e9e 152', got '$got'"
fi

# Case 6: `ink.colour` is a text run's tint (KAN-437 final verification: a
# caption and a summary row shipped in the wrong colour; a run has no box, so
# `content.colour` cannot be asked). Same glyph run, one grey and one blue,
# in a crop spanning the container's width — height, left and colour all read.
make_text() {
  python3 - "$1" "$2" <<'PY'
import sys
from PIL import Image, ImageDraw
path, tint = sys.argv[1], tuple(int(sys.argv[2][i:i + 2], 16) for i in (0, 2, 4))
im = Image.new("RGB", (200, 40), (0xEA, 0xE9, 0xE9))
d = ImageDraw.Draw(im)
for x in (24, 40, 56):  # three 10x14 "capitals"
    d.rectangle((x, 13, x + 9, 26), fill=tint)
im.save(path)
PY
}
make_text "$DIR/text-grey.png" 9e9e9e
make_text "$DIR/text-blue.png" 1e6fe0
got="$("$GUARD" "$DIR/text-grey.png" "$DIR/text-blue.png" --scale 1 --props ink | python3 -c '
import json, sys
d = json.load(sys.stdin)
a = d["a"]["ink"]
print(a["left"], a["height"], a["colour"], d["b"]["ink"]["colour"], round(d["delta"]["ink.colour"]["distance"]))
')"
if [ "$got" = "24 14 #9e9e9e #1e6fe0 152" ]; then
  pass "case 6: ink reads a text run's position, cap-height and tint, and the delta carries the tint distance"
else
  fail "case 6: expected '24 14 #9e9e9e #1e6fe0 152', got '$got'"
fi

# Case 7: `bands` pairs a whole page's horizontal structure and lists what a
# data difference cannot explain (KAN-437 fix round 5: four defects on one
# screen passed a composite already 0.3 white from data and fonts). The
# frame: a header bar (surface grey on page grey — closer than `--edge`, so
# no box, but further than `--noise`, so a band), three text rows separated
# by 2px dividers, an outlined accent button. The capture: the same rows
# with different text widths, the dividers absent, the row padding halved,
# the button's border grey, the header bar in the page colour.
make_page() {
  python3 - "$1" "$2" <<'PY'
import sys
from PIL import Image, ImageDraw
path, broken = sys.argv[1], sys.argv[2] == "broken"
PAGE, SURFACE, RULE, TEXT, ACCENT, GREY = (0xF3, 0xF2, 0xF2), (0xEA, 0xE9, 0xE9), (0xD7, 0xD3, 0xD3), (0x20, 0x1E, 0x1D), (0x00, 0x88, 0xB0), (0xD7, 0xD3, 0xD3)
im = Image.new("RGB", (300, 400), PAGE)
d = ImageDraw.Draw(im)
if not broken:
    d.rectangle((0, 0, 299, 39), fill=SURFACE)          # header bar
y = 80
pad = 6 if broken else 12
for i, w in enumerate((120, 90, 150)):
    y += pad
    d.rectangle((20, y, 20 + w + (30 if broken else 0), y + 15), fill=TEXT)  # text row, width is data
    y += 16 + pad
    if not broken:
        d.rectangle((20, y, 279, y + 1), fill=RULE)      # divider
    y += 2
y += 40
d.rectangle((20, y, 279, y + 47), outline=GREY if broken else ACCENT, width=2)
im.save(path)
PY
}
make_page "$DIR/page-frame.png" ok
make_page "$DIR/page-capture.png" broken
got="$("$GUARD" "$DIR/page-frame.png" "$DIR/page-capture.png" --scale 1 --props bands | python3 -c '
import json, sys
d = json.load(sys.stdin)
s = d["delta"]["bands_summary"]
pairs = d["delta"]["bands"]
missing = [p["a"] for p in pairs if p["status"] == "missing"]
rules = sum(1 for m in missing if m["height"] == 2 and m["colour"] == "#d7d3d3")
header = any(m["height"] == 40 and m["colour"] == "#eae9e9" for m in missing)
rows = [p for p in pairs if p["status"] == "paired" and p["a"]["height"] == 16]
button = next(p for p in pairs if p["status"] == "paired" and p["a"]["height"] == 48)
print(s["paired"], s["missing"], s["extra"], rules, header, [p["gap_above"]["abs"] for p in rows][1:], [p["since_pair"]["abs"] for p in rows][1:], round(button["edge"]["distance"]))
')"
# 4 paired (three rows, the button); 4 missing (header, three rules). The
# second and third rows: `gap_above` reads +2 (14 to the previous row in the
# capture against 12 to the rule in the frame — a different neighbour), and
# `since_pair` reads -12 (14 since the previous row against 12+2+12): the
# lost padding is the latter number, and only it. The button border is 230
# units from accent.
if [ "$got" = "4 4 0 3 True [2.0, 2.0] [-12.0, -12.0] 230" ]; then
  pass "case 7: bands lists the missing rules and header band, the halved row padding as since_pair and the button's grey border, and pairs the rows across a text-width difference"
else
  fail "case 7: expected '4 4 0 3 True [2.0, 2.0] [-12.0, -12.0] 230', got '$got'"
fi

# Case 8: `seams` reads the vertical structure inside a boxed band — the
# KAN-437 GOAL segmented control: a 1px-bordered 3-cell control whose two
# dividers and whose wrapped labels' centring both shipped wrong past the
# band pairing (one band in both images, same height, same edge) and every
# sweep. The frame: a title text row above (unboxed — its stems must produce
# no seams), the control with both dividers, a one-line centred label in
# cells 1 and 3 and a two-line centred label of unequal widths in cell 3.
# The capture: the first divider absent, the second present, cell 3's two
# lines left-anchored 8px in, and cell 1's label wider (data).
make_control() {
  python3 - "$1" "$2" <<'PY'
import sys
from PIL import Image, ImageDraw
path, broken = sys.argv[1], sys.argv[2] == "broken"
PAGE, RULE, TEXT = (0xF3, 0xF2, 0xF2), (0xD7, 0xD3, 0xD3), (0x20, 0x1E, 0x1D)
im = Image.new("RGB", (340, 160), PAGE)
d = ImageDraw.Draw(im)
d.rectangle((20, 20, 21, 31), fill=TEXT)                             # title: a text row — one ascender stem spanning the band's height
d.rectangle((26, 24, 20 + (110 if broken else 90), 31), fill=TEXT)   # and its x-height body, width is data
L, T, R, B = 20, 60, 319, 107                                        # the control: 300x48, 3 cells of 98 inside a 1px border
d.rectangle((L, T, R, B), outline=RULE, width=1)
for x in (L + 100, L + 200):
    if x == L + 100 and broken:
        continue                                                     # the missing divider
    d.line((x, T + 1, x, B - 1), fill=RULE)
def label(x0, x1, y, w, left_anchor=None):
    l = x0 + 8 if left_anchor else (x0 + x1 + 1 - w) // 2
    d.rectangle((l, y, l + w - 1, y + 9), fill=TEXT)
label(L + 1, L + 99, 79, 70 if broken else 50)                       # cell 1: one line, width is data
label(L + 101, L + 199, 79, 50)                                      # cell 2: one line
label(L + 201, R - 1, 72, 60, broken)                                # cell 3: two lines, 60 and 40 wide
label(L + 201, R - 1, 86, 40, broken)
im.save(path)
PY
}
make_control "$DIR/control-frame.png" ok
make_control "$DIR/control-capture.png" broken
got="$("$GUARD" "$DIR/control-frame.png" "$DIR/control-capture.png" --scale 1 --props seams | python3 -c '
import json, sys
d = json.load(sys.stdin)
s = d["delta"]["seams_summary"]
frame_boxed = [b["top"] for b in d["a"]["seams"]]
band = next(b for b in d["delta"]["seams"] if b["status"] == "paired")
missing = [(p["a"]["left"], p["a"]["width"], p["a"]["colour"]) for p in band["seams"] if p["status"] == "missing"]
cells = {(c["a"], c["b"]): [(l["offset"]["abs"], l["left"]["abs"]) for l in c["lines"]] for c in band["cells"]}
print(frame_boxed, s["paired"], s["missing"], s["extra"], s["bands_unpaired"], s["lines"], missing, cells)
')"
# Only the control is boxed (the title row is not, so its stems make no
# seam). Of the frame'"'"'s four seams the left border, the second divider and
# the right border pair and the first divider is missing at x=120. Cell
# pairs: only frame cell 2 with capture cell 1 (frame cells 0 and 1 have no
# counterpart across the missing divider). Its two lines: the 60-wide one
# reads offset -11 (centred at 8+30 against 49) with `left` -11; the
# 40-wide one offset -21 with `left` -21 — a slack of half the cell width
# minus the text, not a data width, which would leave `left` at 0.
want="[60] 3 1 0 0 {'paired': 2, 'unpaired': 0} [(120, 1, '#d7d3d3')] {(2, 1): [(-11.0, -11.0), (-21.0, -21.0)]}"
if [ "$got" = "$want" ]; then
  pass "case 8: seams lists the missing cell divider and the wrapped label's lines left-anchored where the frame centres them, reads no seam from a text row's stems, and pairs cells across a data width"
else
  fail "case 8: expected \"$want\", got \"$got\""
fi

echo "FAILURES: $FAILURES"
[ "$FAILURES" -eq 0 ]
