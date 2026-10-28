#!/usr/bin/env bash
# test-run-capped.sh — run-capped.sh kills a command group past its memory
# cap. Allocations stay small: cap 64/128 MB, child 300 MB.
set -u
here=$(cd "$(dirname "$0")" && pwd)
rc="$here/run-capped.sh"
fail=0
alloc='b=bytearray(300*1024*1024); import time; time.sleep(10)'
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT

# (a) capped, fast, CAPPED: line
start=$(date +%s)
bash "$rc" -m 64 -- python3 -c "$alloc" 2>"$tmp/err"; s=$?
el=$(( $(date +%s) - start ))
[ "$s" = 137 ] && [ "$el" -le 6 ] && grep -q 'CAPPED:' "$tmp/err" || { echo "FAIL (a): status=$s elapsed=$el"; fail=1; }

# (b) grandchild allocator killed too
bash "$rc" -m 64 -- bash -c "python3 -c '$alloc' & echo \$! >'$tmp/pid'; wait" 2>/dev/null; s=$?
gp=$(cat "$tmp/pid" 2>/dev/null)
[ "$s" = 137 ] && [ -n "$gp" ] && ! kill -0 "$gp" 2>/dev/null || { echo "FAIL (b): status=$s grandchild=$gp $(ps -o stat=,command= -p "$gp")"; fail=1; }

# (c) child status passes through
bash "$rc" -m 64 -- bash -c 'exit 3' 2>/dev/null; s=$?
[ "$s" = 3 ] || { echo "FAIL (c): status=$s"; fail=1; }

# (d) cap from .flow/project.md ## resource cap
mkdir -p "$tmp/p/.flow"; printf '## resource cap\n\n128\n' >"$tmp/p/.flow/project.md"
(cd "$tmp/p" && bash "$rc" -- python3 -c "$alloc") 2>"$tmp/err"; s=$?
[ "$s" = 137 ] && grep -q 'CAPPED: 128MB' "$tmp/err" || { echo "FAIL (d): status=$s $(cat "$tmp/err")"; fail=1; }

[ "$fail" = 0 ] && echo "ok: run-capped.sh"
exit "$fail"
