#!/usr/bin/env bash
# test-run-capped.sh — run-capped.sh kills a command's process tree past its
# memory cap. Kept light: cap 64 MB, child allocates 128 MB.
set -u
here=$(cd "$(dirname "$0")" && pwd)
rc="$here/run-capped.sh"
fail=0
# Random bytes, re-touched every 0.2s for 10s: zero-filled pages left idle are
# compressed out of RSS within ~5s on macOS, so a sample run-capped.sh takes
# late under load read under the cap and the probe flaked (kan-1048 verify).
alloc='import os, time; b = bytearray(os.urandom(128 << 20)); [(b.__setitem__(slice(None, None, 4096), os.urandom(len(b) // 4096)), time.sleep(0.2)) for _ in range(50)]'
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
unset RUN_CAPPED_PGID

# (a) capped, fast, CAPPED: line
start=$(date +%s)
bash "$rc" -m 64 -- python3 -c "$alloc" 2>"$tmp/err"; s=$?
el=$(( $(date +%s) - start ))
[ "$s" = 137 ] && [ "$el" -le 6 ] && grep -q 'CAPPED:' "$tmp/err" || { echo "FAIL (a): status=$s elapsed=$el"; fail=1; }

# (b) grandchild allocator killed too
bash "$rc" -m 64 -- bash -c "python3 -c '$alloc' & echo \$! >'$tmp/pid'; wait" 2>/dev/null; s=$?
gp=$(cat "$tmp/pid" 2>/dev/null)
[ "$s" = 137 ] && [ -n "$gp" ] && ! kill -0 "$gp" 2>/dev/null || { echo "FAIL (b): status=$s grandchild=$gp"; fail=1; }

# (c) child status passes through
bash "$rc" -m 64 -- bash -c 'exit 3' 2>/dev/null; s=$?
[ "$s" = 3 ] || { echo "FAIL (c): status=$s"; fail=1; }

# (d) cap from .flow/project.md ## resource cap
mkdir -p "$tmp/p/.flow"; printf '## resource cap\n\n64\n' >"$tmp/p/.flow/project.md"
(cd "$tmp/p" && bash "$rc" -- python3 -c "$alloc") 2>"$tmp/err"; s=$?
[ "$s" = 137 ] && grep -q 'CAPPED: 64MB' "$tmp/err" || { echo "FAIL (d): status=$s $(cat "$tmp/err")"; fail=1; }

# (e) a setsid child (own session and group, as a Playwright browser) is in the tree
bash "$rc" -m 64 -- python3 -c "
import os, subprocess, time
p = subprocess.Popen(['python3', '-c', '$alloc'], start_new_session=True)
open('$tmp/sid', 'w').write(str(p.pid)); p.wait()" 2>/dev/null; s=$?
sp=$(cat "$tmp/sid" 2>/dev/null)
[ "$s" = 137 ] && [ -n "$sp" ] && ! kill -0 "$sp" 2>/dev/null || { echo "FAIL (e): status=$s setsid child=$sp"; fail=1; }
[ -n "$sp" ] && kill -9 "$sp" 2>/dev/null

# (f) nested run-capped.sh: the outer cap covers the inner command
bash "$rc" -m 64 -- bash "$rc" -m 4096 -- python3 -c "$alloc" 2>"$tmp/err"; s=$?
[ "$s" = 137 ] && grep -q 'CAPPED: 64MB' "$tmp/err" || { echo "FAIL (f): status=$s $(cat "$tmp/err")"; fail=1; }

# (g) a cap with a unit is refused, not misread
printf '## resource cap\n\n8 GB\n' >"$tmp/p/.flow/project.md"
(cd "$tmp/p" && bash "$rc" -- true) 2>"$tmp/err"; s=$?
[ "$s" = 2 ] && grep -q 'plain integer' "$tmp/err" || { echo "FAIL (g): status=$s $(cat "$tmp/err")"; fail=1; }

# (h) project-get.sh refusing (key declared twice) fails loudly
printf '## resource cap\n\n64\n\n## resource cap\n\n64\n' >"$tmp/p/.flow/project.md"
(cd "$tmp/p" && bash "$rc" -- true) 2>"$tmp/err"; s=$?
[ "$s" = 2 ] && grep -q 'project-get.sh failed' "$tmp/err" || { echo "FAIL (h): status=$s $(cat "$tmp/err")"; fail=1; }

# (i) SIGTERM to the wrapper exits 143 and kills the command
bash "$rc" -m 64 -- bash -c "echo \$\$ >'$tmp/tp'; sleep 20" 2>/dev/null & w=$!
sleep 1; kill -TERM "$w"; wait "$w"; s=$?
tp=$(cat "$tmp/tp" 2>/dev/null)
[ "$s" = 143 ] && [ -n "$tp" ] && ! kill -0 "$tp" 2>/dev/null || { echo "FAIL (i): status=$s child=$tp"; fail=1; }

# (j) the command's stdin is /dev/null
out=$(bash "$rc" -m 64 -- bash -c 'read -r x && echo "got $x"' <<<hi)
[ -z "$out" ] || { echo "FAIL (j): child read stdin: $out"; fail=1; }

[ "$fail" = 0 ] && echo "ok: run-capped.sh"
exit "$fail"
