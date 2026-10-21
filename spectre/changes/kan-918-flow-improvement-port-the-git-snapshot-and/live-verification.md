# kan-918 — live verification

`design.md` **Live check**: each of the five shims, run through a symlink to the worktree's
`scripts/` directory (the installed layout), against the bash at the merge base `c9b8fde8`
(`git archive c9b8fde8 scripts`), with the same arguments the skills pass, on this repository.
Failure look: a shim that prints nothing and exits 0 where the bash printed a verdict line, or any
difference in stdout, stderr or exit code.

| Call | Before (bash @ c9b8fde8) | After (shim) | Match |
|---|---|---|---|
| `check-plan-unchanged.sh snapshot <wt> <name> <snap>` | rc 0, `PLAN-SNAPSHOT-OK: …` | rc 0, same bytes | yes |
| `check-plan-unchanged.sh verify <wt> <name> <snap>` | rc 0, `PLAN-UNCHANGED-OK: …` | rc 0, same bytes | yes |
| `check-tree-markers.sh snapshot <wt> <markers> <snap>` | rc 0, `TREE-MARKERS-SNAPSHOT-OK: …` | rc 0, same bytes | yes |
| `check-tree-markers.sh verify <wt> <markers> <snap>` | rc 0, `TREE-MARKERS-OK: …` | rc 0, same bytes | yes |
| `guard-autosquash.sh targets <wt> HEAD` | rc 0, `guard-autosquash: targets ok` | rc 0, same bytes | yes |
| `guard-autosquash.sh after <wt> <merge-base> <tasks.md>` | rc 0, `guard-autosquash: after ok` | rc 0, same bytes | yes |
| `check-panel-diff-size.sh <wt> <merge-base>` | rc 0, `4192` / `under cap` | rc 0, same bytes | yes |
| `check-worktree-processes.sh <wt>` | rc 1, `HELD: <wt> — <pid> <cwd>; …` | rc 1, same verdict and shape | yes — only the pids of processes that came and went between the two calls (the suite's test children, each guard's own lsof) differ |

<!-- measured: bash scratchpad/live.sh (base side from git archive c9b8fde8 scripts, new side via a symlinked scripts dir; stdout/stderr/rc compared with cmp) @ branch spectre/kan-918-flow-improvement-port-the-git-snapshot-and -->

Result: matched — no shim printed nothing where the bash printed a verdict.
