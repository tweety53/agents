# kan-850-agents-port-the-visual-verify-scripts-to-the-go — session narrative

## 2026-09-28 — creating run

The run resumed at `STARTED`, with the plan already ready, and implemented all ten tasks in SDD waves.

- **Task 7 guard failure.** Task 7's commit adds a deliberately non-UTF-8 compose fixture (`map-not-utf8/map.mockups`). `check-task-commit-fields.sh` crashed on it with a UnicodeDecodeError, so task 7 was ticked on a hand check. That check covered the subject, the trailer, the paths and the tests, and was recorded as a `flow record substitution`. Panel round 0 raised the crash as F4. The fix marked the fixture `-diff` in a new root `.gitattributes`, and the guard now passes on that commit.
- **Dangling symlinks.** Deleting the two `.py` scripts left dangling installed symlinks in `skills/flow/scripts/`. Both task 7 and task 8 widened `**Files:**` and removed their symlink by amending the unpushed pick.
- **Task 10 timings.** The first run of task 10's live timings failed on three counts:
  - Case 22 of `TestCheckVisualVerification` flaked. A no-LANG shim sharing the guard cache built a second binary, and case 22's glob saw two binaries whenever that shim ran first.
  - The run hung once with SIGPIPE. It did not recur.
  - The suite median missed its target.
- **Case 22 fix and re-measure.** The operator chose to fix case 22 at its source and re-measure. Case 22 now runs the one binary the shared cache built (`guardBinary`, `a990af5c`). The re-measure passed all five checks.
- **Panel round 0** raised two Importants and five Minors.
  - Importants:
    - The NUL line-cut rule lived in two callers rather than the shared `vvLines`.
    - `compose-mockup-frames` copied the package's Python-semantics helpers, and the copies had drifted from measure's.
  - The two slots raised two defects in common, so the store holds seven findings.
  - Two of primary's reproducers were bounced once:
    - F3's memory threshold sat inside macOS RSS noise, because the fixtures were uniform and their pages compressed.
    - F4's `# demonstrates:` citation named a byte that cannot be cited.
- **Fix round 0.** One fix dispatch fixed all seven findings in six new commits on top of the pushed branch. Four reproducers then exited 3 on the fixed tree, because their `# premise:` lines cited code the fix deleted. The raising slots re-authored them in place during their targeted round-1 re-runs. `prove-reproducer.sh` held each one against `0a15948d`.
- **Round 1** came back clean from both slots.
- **Panel close.** `check-panel-fix-single-dispatch.sh` flagged task 10's `full-suite-fix-1` dispatch key, recorded during `flow.sdd-tdd` with role `panel-fix`. The run resolved that prompt on Continue.
- **Stale plan snapshots.** Several times, my own mid-flight edits to the plan and to `KNOWN-BUGS.md` tripped the plan and marker snapshot checks while a reviewer was in flight. Each time the diff was hand-checked and showed only those edits.
- **Suite timings not stored.** `flow suite record` could not store the verify run's timings: the exported worktree `FLOW_ADDR` (port 6683) has no daemon behind it. Every suite still passed.
