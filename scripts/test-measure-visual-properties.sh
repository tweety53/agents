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

echo "FAILURES: $FAILURES"
[ "$FAILURES" -eq 0 ]
