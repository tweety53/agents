#!/usr/bin/env bash
# Moves kan-970's visual-verification record into its archived change root,
# copying every scratchpad file it cites under visual-verification/evidence/
# and citing the copies. Run from the gymie worktree root.
set -euo pipefail
P=/private/tmp/claude-501/-Users-tweety53-Projects-agents/37a08db9-090a-4d3c-9d2e-b1f86ccfe404/scratchpad/vv
R=spectre/changes/archive/kan-970-fix-the-remaining-known-bugs-md-entries-and
E="$R/visual-verification/evidence"
mkdir -p "$E"
cp "$P/verify.log" "$P/cfg/vv.config.ts" "$E/"
cp "$P/snap/kan-940-cardio-rows.spec.ts-snapshots/"*.png "$E/"
for f in "$P/motion/borrowed-delete-cardio/"*.png; do cp "$f" "$E/borrowed-delete-cardio-${f##*/}"; done
python3 - "$P" .superpowers/sdd/visual-verification.md "$R/visual-verification.md" <<'PY'
import sys
p, src, dst = sys.argv[1], sys.argv[2], sys.argv[3]
s = open(src).read()
e = "visual-verification/evidence/"
for old, new in [
    (p + "/snap/kan-940-cardio-rows.spec.ts-snapshots/*.png", e + "*-darwin.png"),
    (p + "/snap/kan-940-cardio-rows.spec.ts-snapshots/", e),
    (p + "/motion/borrowed-delete-cardio", e + "borrowed-delete-cardio-*.png"),
    (p + "/cfg/vv.config.ts", e + "vv.config.ts"),
    (p + "/verify.log", e + "verify.log"),
    (p + "/snap", e),
]:
    s = s.replace(old, new)
assert p not in s, "a scratchpad citation is left"
title, rest = s.split("\n", 1)
note = ("\n> Moved into this change root from the gymie worktree's `.superpowers/sdd/` by the kan-970 "
        "self-review (finding F3); every scratchpad file it cited is copied under "
        "`visual-verification/evidence/` and cited there.\n")
open(dst, "w").write(title + "\n" + note + rest)
PY
rm .superpowers/sdd/visual-verification.md
