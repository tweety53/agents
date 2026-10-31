# Live verification — kan-1072

Measured 2026-10-31 against stand-ins (scratch clones of `spectre/kan-1072-…`, a store copy on 4175).

| Check | Before (merge base / installed build) | After (worktree build) | Verdict |
|---|---|---|---|
| 1. Landing commit | `spectre archive` + main-checkout `commit-archive.sh`: subject `chore(spectre): archive live-check-1`; `spectre/changes/archive/live-check-1` tree present; live dir absent | worktree `commit-archive.sh`: subject `chore(spectre): drop live-check-1`; both `ls-tree` lines empty | matched |
| 2. Board reads a landed plan | dev flowd 4173 (installed build, reads `archive/`): `kan-1064-flow-self-review-run-the-coordination-around-the` → 3 tasks | store copy 4175 (worktree build, `stats/internal/landed`): same change → 3 tasks | matched — equal, non-zero |
| 3. Cleanup's check | before the drop commit: `LEFTOVER … spectre/changes/live-check-1 was not dropped` | after the drop commit: no `spectre/changes/live-check-1` row (remaining LEFTOVERs name only the scratch clone's own branch/worktree) | matched |
