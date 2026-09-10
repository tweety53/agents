#!/usr/bin/env bash
# Assertion harness for compose-mockup-frames.sh / .py.
#
# Needs Pillow exactly as the script does — fixture PNGs are made by an
# inline `python3 - <<'PY'` using Pillow (Image.new(...).save(path)).  If
# Pillow is absent this harness FAILS, never skips, so a green run of it is
# never vacuous (`.flow/project.md`'s "vacuous pass" rule for guards).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/compose-mockup-frames.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

DIRS=()
cleanup() {
  [ "${#DIRS[@]}" -eq 0 ] && return 0
  local d
  for d in "${DIRS[@]}"; do rm -rf "$d"; done
}
trap cleanup EXIT

if ! python3 -c 'from PIL import Image' >/dev/null 2>&1; then
  fail "Pillow is not importable on this machine — this harness needs Pillow exactly as the script does, so it fails rather than skipping"
  echo "FAILURES: $FAILURES"
  exit 1
fi

# new_root -> ROOT (mockups root), OUT (output dir), MAP (map file path).
# Each dir is re-resolved through `cd && pwd` because mktemp preserves a
# double slash when $TMPDIR itself ends in one (e.g. macOS's default), and
# the script's own os.path.abspath() normalizes that away — comparing a
# non-normalized expected path against its normalized output would fail
# for a reason that has nothing to do with the guard under test.
new_root() {
  ROOT="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/compose-mockup-root.XXXXXX")" && pwd)"
  OUT="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/compose-mockup-out.XXXXXX")" && pwd)"
  CAPS="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/compose-mockup-caps.XXXXXX")" && pwd)"
  DIRS+=("$ROOT" "$OUT" "$CAPS")
  MAP="$ROOT/map.mockups"
}

# make_png <path> <width> <height> [r] [g] [b]
make_png() {
  python3 - "$1" "$2" "$3" "${4:-0}" "${5:-0}" "${6:-0}" <<'PY'
import sys
from PIL import Image
path, w, h, r, g, b = sys.argv[1], int(sys.argv[2]), int(sys.argv[3]), int(sys.argv[4]), int(sys.argv[5]), int(sys.argv[6])
Image.new("RGB", (w, h), (r, g, b)).save(path)
PY
}

run_guard() {
  set +e
  OUT_TEXT="$(printf '%s\n' "$@" | "$GUARD" "$MAP" "$ROOT" "$OUT" 2>/tmp/compose-mockup-stderr.$$)"
  RC=$?
  ERR="$(cat "/tmp/compose-mockup-stderr.$$")"
  rm -f "/tmp/compose-mockup-stderr.$$"
  set -e
}

# ===========================================================================
# Case 1: happy path — capture and frame differ in size, composite is the
# sum of widths plus the gutter and the max of heights.
# ===========================================================================
new_root
make_png "$CAPS/j5-finish-dialog-darwin.png" 1200 1164 255 0 0
make_png "$ROOT/J5.png" 780 2028 0 255 0
printf 'j5-finish-dialog.png J5\n' > "$MAP"
run_guard "$CAPS/j5-finish-dialog-darwin.png"
if [ "$RC" -eq 0 ] && [ "$OUT_TEXT" = "$OUT/j5-finish-dialog.png" ]; then
  pass "case 1: happy path exits 0 with the composite path on stdout"
else
  fail "case 1: rc=$RC out=$OUT_TEXT err=$ERR"
fi
SIZE="$(python3 -c "from PIL import Image; i=Image.open('$OUT/j5-finish-dialog.png'); print(f'{i.size[0]}x{i.size[1]}')" 2>&1)"
[ "$SIZE" = "1996x2028" ] && pass "case 1: composite size is 1996x2028" || fail "case 1: size=$SIZE"

# ===========================================================================
# Case 2: exact-name match, no platform suffix.
# ===========================================================================
new_root
make_png "$CAPS/k2-menu.png" 100 100
make_png "$ROOT/K2.png" 100 100
printf 'k2-menu.png K2\n' > "$MAP"
run_guard "$CAPS/k2-menu.png"
[ "$RC" -eq 0 ] && pass "case 2: exact-name match with no platform suffix exits 0" || fail "case 2: rc=$RC out=$OUT_TEXT err=$ERR"

# ===========================================================================
# Case 3: a map line's screenshot matches no stdin path -> exit 1, and a
# second well-formed line in the same map is still composed.
# ===========================================================================
new_root
make_png "$CAPS/ok-darwin.png" 50 50
make_png "$ROOT/OK.png" 50 50
make_png "$ROOT/MISSING.png" 50 50
printf 'missing-cap.png MISSING\nok.png OK\n' > "$MAP"
run_guard "$CAPS/ok-darwin.png"
case "$ERR" in
  *"no captured PNG matches"*) pass "case 3: stderr carries 'no captured PNG matches'" ;;
  *) fail "case 3: err=$ERR" ;;
esac
[ "$RC" -eq 1 ] && pass "case 3: exit 1" || fail "case 3: rc=$RC"
[ "$OUT_TEXT" = "$OUT/ok.png" ] && pass "case 3: the well-formed second line is still composed" || fail "case 3: out=$OUT_TEXT"

# ===========================================================================
# Case 4: a map line names a frame absent under the mockups root -> exit 1,
# stderr carries "no frame".
# ===========================================================================
new_root
make_png "$CAPS/a-darwin.png" 50 50
printf 'a.png NOPE\n' > "$MAP"
run_guard "$CAPS/a-darwin.png"
[ "$RC" -eq 1 ] && pass "case 4: exit 1" || fail "case 4: rc=$RC"
case "$ERR" in
  *"no frame"*) pass "case 4: stderr carries 'no frame'" ;;
  *) fail "case 4: err=$ERR" ;;
esac

# ===========================================================================
# Case 5: a malformed line (three fields) -> exit 1, stderr carries
# "expected".
# ===========================================================================
new_root
printf 'a.png B extra\n' > "$MAP"
run_guard
[ "$RC" -eq 1 ] && pass "case 5: exit 1" || fail "case 5: rc=$RC"
case "$ERR" in
  *"expected"*) pass "case 5: stderr carries 'expected'" ;;
  *) fail "case 5: err=$ERR" ;;
esac

# ===========================================================================
# Case 6: comment and blank lines are ignored -> exit 0, no output.
# ===========================================================================
new_root
printf '# a comment\n\n   \n# another\n' > "$MAP"
run_guard
[ "$RC" -eq 0 ] && [ -z "$OUT_TEXT" ] && pass "case 6: comment/blank-only map exits 0 with no output" || fail "case 6: rc=$RC out=$OUT_TEXT"

# ===========================================================================
# Case 7: two stdin paths match one name -> exit 1 naming both.
# ===========================================================================
new_root
make_png "$CAPS/a-darwin.png" 50 50
make_png "$CAPS/a-linux.png" 50 50
make_png "$ROOT/A.png" 50 50
printf 'a.png A\n' > "$MAP"
run_guard "$CAPS/a-darwin.png" "$CAPS/a-linux.png"
[ "$RC" -eq 1 ] && pass "case 7: exit 1" || fail "case 7: rc=$RC"
case "$ERR" in
  *"a-darwin.png"*"a-linux.png"*|*"a-linux.png"*"a-darwin.png"*) pass "case 7: stderr names both paths" ;;
  *) fail "case 7: err=$ERR" ;;
esac

# ===========================================================================
# Case 8: an unreadable PNG on stdin (a text file named .png) -> exit 2.
# ===========================================================================
new_root
printf 'not a png' > "$CAPS/bad.png"
make_png "$ROOT/X.png" 50 50
printf 'bad.png X\n' > "$MAP"
run_guard "$CAPS/bad.png"
[ "$RC" -eq 2 ] && pass "case 8: unreadable PNG on stdin exits 2" || fail "case 8: rc=$RC out=$OUT_TEXT err=$ERR"

# ===========================================================================
# Case 9: wrong argument count -> exit 2 with the usage line.
# ===========================================================================
set +e
USAGE_OUT="$("$GUARD" a b 2>&1)"
USAGE_RC=$?
set -e
[ "$USAGE_RC" -eq 2 ] && pass "case 9: wrong argument count exits 2" || fail "case 9: rc=$USAGE_RC"
case "$USAGE_OUT" in
  *"usage: compose-mockup-frames.sh"*) pass "case 9: stderr carries the usage line" ;;
  *) fail "case 9: out=$USAGE_OUT" ;;
esac

# ===========================================================================
# Case 10: mockups root not a directory -> exit 2.
# ===========================================================================
new_root
printf 'a.png A\n' > "$MAP"
NOT_A_DIR="$ROOT/not-a-dir"
: > "$NOT_A_DIR"
set +e
OUT_TEXT="$(printf '' | "$GUARD" "$MAP" "$NOT_A_DIR" "$OUT" 2>/tmp/compose-mockup-stderr.$$)"
RC=$?
ERR="$(cat "/tmp/compose-mockup-stderr.$$")"
rm -f "/tmp/compose-mockup-stderr.$$"
set -e
[ "$RC" -eq 2 ] && pass "case 10: mockups root not a directory exits 2" || fail "case 10: rc=$RC out=$OUT_TEXT err=$ERR"

# ===========================================================================
# Case 11: Pillow absent -> exit 2, stderr carries the pip-install hint.
# Shadow the real package for one invocation via PYTHONPATH pointing at a
# fixture dir holding PIL/__init__.py that raises ImportError.
# ===========================================================================
new_root
STUB_DIR="$(mktemp -d "${TMPDIR:-/tmp}/compose-mockup-stub.XXXXXX")"
DIRS+=("$STUB_DIR")
mkdir -p "$STUB_DIR/PIL"
printf 'raise ImportError("stubbed out by test-compose-mockup-frames.sh")\n' > "$STUB_DIR/PIL/__init__.py"
printf 'a.png A\n' > "$MAP"
set +e
NO_PILLOW_OUT="$(PYTHONPATH="$STUB_DIR" printf '' | PYTHONPATH="$STUB_DIR" "$GUARD" "$MAP" "$ROOT" "$OUT" 2>&1)"
NO_PILLOW_RC=$?
set -e
[ "$NO_PILLOW_RC" -eq 2 ] && pass "case 11: Pillow absent exits 2" || fail "case 11: rc=$NO_PILLOW_RC out=$NO_PILLOW_OUT"
case "$NO_PILLOW_OUT" in
  *"python3 -m pip install pillow"*) pass "case 11: stderr carries the pip-install hint" ;;
  *) fail "case 11: out=$NO_PILLOW_OUT" ;;
esac

# ===========================================================================
# Case 12: the wrapper refuses a missing .py sibling with exit 2 (mirrors
# test-check-plan-shape.sh's case 17).
# ===========================================================================
STRIPPED="$(mktemp -d "${TMPDIR:-/tmp}/compose-mockup-stripped.XXXXXX")"
DIRS+=("$STRIPPED")
cp "$GUARD" "$STRIPPED/"
set +e
STRIPPED_OUT="$("$STRIPPED/compose-mockup-frames.sh" a b c 2>&1)"
STRIPPED_RC=$?
set -e
[ "$STRIPPED_RC" -eq 2 ] && pass "case 12: a guard copy with no compose-mockup-frames.py sibling exits 2" || fail "case 12: rc=$STRIPPED_RC out=$STRIPPED_OUT"
case "$STRIPPED_OUT" in
  *"compose-mockup-frames.py"*) pass "case 12: the message names the missing module" ;;
  *) fail "case 12: out=$STRIPPED_OUT" ;;
esac

# ===========================================================================
# Case 13: an unreadable mockup frame file (a text file named .png) -> exit 2.
# ===========================================================================
new_root
make_png "$CAPS/bad-frame.png" 50 50
printf 'not a png' > "$ROOT/X.png"
printf 'bad-frame.png X\n' > "$MAP"
run_guard "$CAPS/bad-frame.png"
[ "$RC" -eq 2 ] && pass "case 13: unreadable mockup frame exits 2" || fail "case 13: rc=$RC out=$OUT_TEXT err=$ERR"
case "$ERR" in
  *"unreadable PNG"*"X.png"*) pass "case 13: stderr names the unreadable frame" ;;
  *) fail "case 13: err=$ERR" ;;
esac

# ===========================================================================
# Case 14: two map lines share the same screenshot name -> exit 1, stderr
# carries "duplicate screenshot name".
# ===========================================================================
new_root
make_png "$CAPS/a-darwin.png" 50 50
make_png "$ROOT/A.png" 50 50
make_png "$ROOT/B.png" 50 50
printf 'a.png A\na.png B\n' > "$MAP"
run_guard "$CAPS/a-darwin.png"
[ "$RC" -eq 1 ] && pass "case 14: duplicate screenshot name in map exits 1" || fail "case 14: rc=$RC out=$OUT_TEXT err=$ERR"
case "$ERR" in
  *"duplicate screenshot name"*"a.png"*) pass "case 14: stderr names the duplicate" ;;
  *) fail "case 14: err=$ERR" ;;
esac

# ===========================================================================
# Case 15: a map line's screenshot name doesn't end in .png -> exit 1.
# ===========================================================================
new_root
make_png "$CAPS/a-darwin.png" 50 50
make_png "$ROOT/A.png" 50 50
printf 'a-no-suffix A\n' > "$MAP"
run_guard "$CAPS/a-darwin.png"
[ "$RC" -eq 1 ] && pass "case 15: screenshot name without .png suffix exits 1" || fail "case 15: rc=$RC out=$OUT_TEXT err=$ERR"
case "$ERR" in
  *"does not end in .png"*) pass "case 15: stderr names the missing .png suffix" ;;
  *) fail "case 15: err=$ERR" ;;
esac

# ===========================================================================
# Case 16: the composite's background (gutter) pixel is exactly (40, 40, 40).
# ===========================================================================
new_root
make_png "$CAPS/bg-check-darwin.png" 100 100 255 0 0
make_png "$ROOT/BG.png" 100 200 0 255 0
printf 'bg-check.png BG\n' > "$MAP"
run_guard "$CAPS/bg-check-darwin.png"
[ "$RC" -eq 0 ] && pass "case 16: happy path for background check exits 0" || fail "case 16: rc=$RC out=$OUT_TEXT err=$ERR"
BG_PIXEL="$(python3 -c "from PIL import Image; print(Image.open('$OUT/bg-check.png').getpixel((100, 5)))")"
[ "$BG_PIXEL" = "(40, 40, 40)" ] && pass "case 16: gutter pixel is (40, 40, 40)" || fail "case 16: pixel=$BG_PIXEL"

# ===========================================================================
# Case 17: the mockup is pasted top-aligned, not bottom-aligned — a capture
# taller than its mockup places distinct-colored mockup content flush with
# the top edge, not the bottom.
# ===========================================================================
new_root
make_png "$CAPS/align-check-darwin.png" 50 200 255 0 0
make_png "$ROOT/ALIGN.png" 50 50 0 255 0
printf 'align-check.png ALIGN\n' > "$MAP"
run_guard "$CAPS/align-check-darwin.png"
[ "$RC" -eq 0 ] && pass "case 17: happy path for alignment check exits 0" || fail "case 17: rc=$RC out=$OUT_TEXT err=$ERR"
TOP_PIXEL="$(python3 -c "from PIL import Image; print(Image.open('$OUT/align-check.png').getpixel((66, 0)))")"
BOTTOM_PIXEL="$(python3 -c "from PIL import Image; print(Image.open('$OUT/align-check.png').getpixel((66, 199)))")"
[ "$TOP_PIXEL" = "(0, 255, 0)" ] && pass "case 17: mockup pixel at top edge is the mockup's color" || fail "case 17: top=$TOP_PIXEL"
[ "$BOTTOM_PIXEL" = "(40, 40, 40)" ] && pass "case 17: pixel below the top-aligned mockup is the background" || fail "case 17: bottom=$BOTTOM_PIXEL"

# ===========================================================================
# Case 18: a screenshot name that is a string-prefix of ANOTHER screenshot's
# own platform-suffixed capture must not match that other capture — the
# kan-486 proof-run collision: `add-participant-finished-excluded.png`'s
# stem is a dash-prefix of `add-participant-finished-excluded-added-darwin.png`,
# a real capture for a *different* screenshot
# (`add-participant-finished-excluded-added.png`). Before the fix this
# resolved as an ambiguous multi-match (or worse, a silent wrong pairing)
# even though every capture and every map line is individually correct.
# ===========================================================================
new_root
make_png "$CAPS/add-participant-finished-excluded-darwin.png" 50 50 255 0 0
make_png "$CAPS/add-participant-finished-excluded-added-darwin.png" 50 50 0 0 255
make_png "$ROOT/H2.png" 50 50
printf 'add-participant-finished-excluded.png H2\n' > "$MAP"
run_guard "$CAPS/add-participant-finished-excluded-darwin.png" "$CAPS/add-participant-finished-excluded-added-darwin.png"
if [ "$RC" -eq 0 ] && [ "$OUT_TEXT" = "$OUT/add-participant-finished-excluded.png" ]; then
  pass "case 18: a shorter name is not confused with a longer name's own platform-suffixed capture"
else
  fail "case 18: rc=$RC out=$OUT_TEXT err=$ERR"
fi
COMP_PIXEL="$(python3 -c "from PIL import Image; print(Image.open('$OUT/add-participant-finished-excluded.png').getpixel((5, 5)))")"
[ "$COMP_PIXEL" = "(255, 0, 0)" ] && pass "case 18: the composed capture is the exact-stem one, not the longer-named one" || fail "case 18: pixel=$COMP_PIXEL"

# ===========================================================================
# Case 19: an actual platform suffix outside the closed vocabulary (a typo,
# or a future Node platform not yet in the list) does not match — this is
# the deliberate flip side of case 18: loosening the check back to "any
# continuation" is exactly the regression case 18 guards against, so this
# case fails loudly if a future edit widens it that way.
# ===========================================================================
new_root
make_png "$CAPS/a-bogusplatform.png" 50 50
make_png "$ROOT/A.png" 50 50
printf 'a.png A\n' > "$MAP"
run_guard "$CAPS/a-bogusplatform.png"
[ "$RC" -eq 1 ] && pass "case 19: a suffix outside the closed platform vocabulary does not match" || fail "case 19: rc=$RC out=$OUT_TEXT err=$ERR"
case "$ERR" in
  *"no captured PNG matches"*) pass "case 19: stderr carries 'no captured PNG matches'" ;;
  *) fail "case 19: err=$ERR" ;;
esac

echo "FAILURES: $FAILURES"
[ "$FAILURES" -eq 0 ]
