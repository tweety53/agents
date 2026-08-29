NOTE: this slot was SUBSTITUTED. The `bugbot` agent type is not offered by this harness, so the
slot ran as a general-purpose subagent carrying Bugbot's brief plus the mandatory mutation-testing
requirement, per **An unspawnable id is substituted, not skipped** (skills/flow/review-panel.md).
Model recorded: sonnet (the model actually given), never `unknown (agent-defined)`.

`git status` is clean of my changes — the only entries are the two untracked files present before I started (design doc and spectre change dir), not modified by me. All mutations were made in a scratchpad copy (`/private/tmp/.../scratchpad/kan362`), never in the worktree itself.

## Verdict: 1 finding (Minor)

**Finding 1 — Minor.** `scripts/test-check-installed-citations.sh`, CHECK-phase loop (around line 949-953):
```bash
i=0
n=${#CMDS[@]}
while [ "$i" -lt "$n" ]; do
  "${CHECKERS[$i]}" "$i"
  i=$((i + 1))
done
```
The file's own comment directly above this loop states the contract: *"replays every case's already-captured result through its own checker, in the original build order, so ok:/FAIL: lines print in exactly the order a reader of the old, sequential harness would have seen them."* No test enforces this.

**Mutation:** reversed the loop to iterate `i` from `${#CMDS[@]}-1` down to `0` (last case's checker runs first, first case's runs last).

**Test run that failed to catch it:** `bash test-check-installed-citations.sh` — result unchanged: `check-installed-citations: all cases pass`, exit 0, 0 `FAIL:` lines, same total ok count — only the printed order of the `ok:` lines changed. No test in this file, `test-run-guard-tests.sh`, or `test-lib-parallel.sh` inspects this harness's own line ordering.

**Reproducer:**
```bash
cd <worktree>/scripts
cp test-check-installed-citations.sh /tmp/tci.sh
python3 - /tmp/tci.sh <<'INNER'
import re
p="/tmp/tci.sh"; s=open(p).read()
old='''i=0
n=${#CMDS[@]}
while [ "$i" -lt "$n" ]; do
  "${CHECKERS[$i]}" "$i"
  i=$((i + 1))
done'''
new='''i=$((${#CMDS[@]} - 1))
n=${#CMDS[@]}
while [ "$i" -ge 0 ]; do
  "${CHECKERS[$i]}" "$i"
  i=$((i - 1))
done'''
assert old in s
open(p,"w").write(s.replace(old,new,1))
INNER
bash /tmp/tci.sh   # exits 0, "all cases pass" — order-reversal is invisible to the suite
```

Reverted; confirmed `git status` shows the tree clean of my changes (only the two pre-existing untracked files remain, unmodified by me).

All other high-value targets — replay list-order (parallel.sh), aggregate exit code, `JOBS=` boundaries (0/-1/3.5/empty/unset), `RUN_GUARD_TESTS_ROOT` three-way handling, glob flatness (nested `subdir/test-x.sh` confirmed excluded), trap chaining/cleanup on clean exit and real SIGINT (both in `parallel.sh` directly and in the harness's chained trap), and the private-`TMPDIR` fix (no other global-namespace `find` scan exists in `scripts/`) — were each mutated and caught by an existing test. No surviving mutants found in those areas.
