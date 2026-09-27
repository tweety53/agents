# kan-841-agents-port-the-next-ten-slowest-bash-scripts-to — session narrative

## 2026-09-27 — creating run

- Implementation ran as ten SDD groups plus the citation sweep and the after-timings. Most review time went to the signal behaviour of the two long-running ports, mutate-and-verify and prove-reproducer. Each needed three or four fix rounds to match the bash trap's restore-then-re-raise and exit-128+n semantics under Go's signal delivery.
- Two hangs came from mutate-and-verify's EPIPE-halting writer, first around the trap's own restore report and then around `git apply`'s output while the lock was held. Both left orphaned `flow-guard mutate-and-verify` processes from test temp dirs, which were killed by hand (three needed SIGKILL).
- Darwin never delivers a `kill`-sent SIGSEGV/SIGBUS/SIGILL/SIGEMT to `signal.Notify`, so that gap is accepted and recorded rather than tested. A `kill`-sent SIGPIPE is tested through a real closed pipe instead.
- Plan-lint friction:
  - Fix-only files had been listed in a task's `**Files:**`.
  - A wrapped prose line began with `**Files:**`.
  - The KNOWN-BUGS.md budget row had to be raised twice: in task 13, and again at panel round close.
- Review panel:
  - The round-0 panel (primary+principles, opus/high) ran 31 minutes against its 15-minute ceiling. The operator chose to keep its findings rather than discard them and re-dispatch.
  - Fix round 1 closed three Importants (the check-guard-symlinks sibling regression, the 255 exit on a signal-killed git) and three Minors.
  - F4's reproducer hardcoded Go line numbers instead of reading the KNOWN-BUGS citations. It was re-authored by the parent and proved on both legs.
  - Five Minors were deferred to KNOWN-BUGS.md.
- Verify: the guard-tests harness took 70s wall this run, against 49s at the sdd-tdd close.

## 2026-09-27 — integrate run

- Preflight `RUN1`; main checkout `STAGED-CLEAN`/`DRIFT-CLEAN` (its untracked `skills/flow/scripts/guard-autosquash.sh` belongs to the separate `fix/guard-autosquash-symlink` worktree and trips `check-guard-symlinks.sh` rule 6 there — unrelated to this change).
- Unfinished-work gate `CLEAR`; visual verify `VISUAL-VERIFY-OK` (no UI paths).
- `check-base-moved.sh` `CLEAR` — no rebase needed.
- Route `merge and push`, taken from the project's configured default, not asked.
- First `flow state get` call failed with usage exit 2: flags must precede the change name (`-C` after the name is rejected).
