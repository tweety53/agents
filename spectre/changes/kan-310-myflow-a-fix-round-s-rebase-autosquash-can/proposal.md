# kan-310-myflow-a-fix-round-s-rebase-autosquash-can

## Why

KAN-310: a review-panel fix round commits a repair as `git commit --fixup=<task-sha>` then
immediately `git rebase --autosquash`s it into that commit. Where the fixup and the commit it
folds into touch nearby lines, git's 3-way auto-merge can resolve in favour of the **pre-fix**
side — it exits 0, prints no conflict marker, and leaves no stray `fixup!` commit behind. The
repair is silently gone.

Observed twice on KAN-295 (fix rounds 2 and 3): three repaired findings (F11, F12, F14) were
reverted by the rebase after being applied and verified. Each time it was caught only because the
round's own report was required to carry pasted command output for every claim, so the check
happened to run after the rebase rather than being trusted to have already run. `git log`, the
folded fixup commit, and every structural pipeline check all passed — only reading the file
content proved the fix had not survived.

## What changes

`skills/flow/review-panel.md`'s **Panel re-runs** section states, as an explicit rule right next
to the fixup/rebase instructions, that a clean `git rebase --autosquash` is not evidence a fix
survived it, and that the reproducer rerun and diff check the section already runs after the fix
subagent reports are what must catch a silent auto-merge revert — never the fixup commit's
presence or the rebase's exit code. No new script, no new guard; this makes an already-run check's
purpose explicit rather than leaving it implicit, which is what let the revert through twice
unnoticed until an unrelated reporting requirement happened to catch it.
