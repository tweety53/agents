# kan-923 — live verification

Before figures: each guard's base wrapper or Python, run from a detached worktree of the merge base
`bbfe685e` (`kan-923-…-before`). After figures: the new shim on HEAD `e1caa28b`, through
`skills/flow/scripts/<name>.sh` where the guard ships there, else `scripts/<name>.sh`. Same inputs on
both sides; stdout, stderr and exit code compared with `cmp`.

| Guard | Input | Before exit | After exit | stdout/stderr |
|---|---|--:|--:|---|
| check-plan-shape | no argument (scan) | 0 | 0 | identical |
| check-plan-shape | this change's `tasks.md` | 0 | 0 | identical |
| check-task-build-green | no argument (scan) | 0 | 0 | identical |
| check-task-build-green | this change's `tasks.md` | 0 | 0 | identical |
| plan-dispatch-bundles | no argument (scan) | 0 | 0 | identical |
| plan-dispatch-bundles | this change's `tasks.md` | 0 | 0 | identical |
| plan-dispatch-groups | this change's `tasks.md` | 0 | 0 | identical |
| check-task-records | `.` on this worktree | 0 | 0 | identical |
| generate-relocation-comparison | this change, merge base `bbfe685e` | 0 | 0 | identical |
| check-markdown-integrity | this tree (`.py` vs new `.sh`) | 0 | 0 | identical |

Failure look (a shim printing nothing and exiting 0 where the Python printed a verdict): not seen —
every pair matched byte for byte. `check-task-build-green` and `check-markdown-integrity` ship no
`skills/*/scripts/` symlink, so they ran through `scripts/`.
