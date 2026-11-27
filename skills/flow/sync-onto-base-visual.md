# Sync the branch onto the base — visual conflicts

**Loaded by the Conflict bullet of `skills/flow/sync-onto-base.md` only when** the guard prints a
`RECAPTURE:` line, or the conflict is in a UI file or a spec's screenshot baseline.

**A path the guard names on a `RECAPTURE: <path>` line is resolved by a targeted recapture,
never by merging its bytes and never by a stop** — a PNG under a `*-snapshots/` directory, or
`full-app-suite.zip`, is a capture of a build, and only a build carrying both sides captures it
honestly:
1. Serve the change's own app worktrees (already synced onto the change's base) per visual-verify
   steps 5–6 (`skills/flow/visual-verify.md`, **Steps 3–13**).
2. Run only the specs owning the named PNGs — the spec each `*-snapshots/` directory is
   named for, and `full-app-suite.spec.ts` for the zip — through `capture` with `<spec>`
   substituted and `--update-snapshots=all` appended (step 8 of
   `skills/flow/visual-verify-verifier.md` states why), in the checkout the conflict is in.
3. Rebuild `full-app-suite.zip` when it was named, with
   `flow-guard rebuild-full-app-suite-zip <worktree> <screenshots root>` as step 8 of
   `skills/flow/visual-verify-verifier.md` runs it.
4. `git add` every recaptured PNG and the zip, stop the stack if step 1 started it, then continue the
   rebase as that bullet does.

**A UI file whose piece-by-piece resolution has already cost a visual round is restored whole.**
Once a conflict in a screen file, resolved by hand, has left a visual difference that took a
fix round, the next resolution of that file restores this change's own version whole (`git
checkout <pre-rebase sha> -- <path>`) and re-applies on top of it the one behaviour the base
introduced there — a port, never a drop of the base's side, so that bullet's never-drop-a-side rule stands. Before any
such round the file is resolved piece by piece as that bullet does.

A spec outside the named ones failing on a screenshot mismatch in any later run of this sync is a
stale baseline on the base: re-baselined per **the sweep**
(`skills/flow-contracts/known-bugs.md`), never asked.
