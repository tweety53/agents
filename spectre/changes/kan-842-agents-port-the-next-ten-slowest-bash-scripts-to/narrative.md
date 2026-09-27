# kan-842-agents-port-the-next-ten-slowest-bash-scripts-to — session narrative

## 2026-09-28 — creating run

- Per-task review ran into repeated parity gaps. Task 5 (`recover-guard-incident`) took four fix rounds to match bash `cd` semantics: lexical canonicalisation, the physical fallback, and the check that each prefix before `..` is a directory. Tasks 4 and 9 moved their sibling lookup from `$FLOW_GUARD_REPO_ROOT/scripts/` to beside the invoked shim, through `FLOW_GUARD_SELF`.
- The orchestrator fixed Minor-only per-task findings inline until the operator stopped it. Minor-only reviews now defer to KNOWN-BUGS.md: `00a0357d` on main changed the flow skill. The panel then deferred Minors beside an Important, which the operator also corrected. On main, `check-panel-findings-closed` is now a Go guard that refuses that deferral (`e2a81e54`, `768075dd`). `.flow/project.md` now says new guards are written in Go (`18cd6093`).
- Timings: the first After run failed both decisions under load 18–28. The Go package's tests held parallel slots idle and paid macOS's git xcrun trampoline. `9a6aa00a` fixed both, and the re-measure met both decisions. Load stayed high (15–57) for the rest of the session, so the ≤40s figure at HEAD is an A/B equivalence with `9a6aa00a`, not a fresh idle measurement.
- At the panel, origin/main had moved: KAN-809 had rewritten `check-visual-verify-dispatched.sh`. The operator chose to rebase and port. The branch merged origin/main instead of rebasing, because it was already pushed. `97c6f38b` ported KAN-809 to Go.
- The operator removed the contract-budget guard mid-run (`e595f199`), together with task 10's port and task 2's `ownedcorpus.go` twin.
- A hook blocks any Bash call whose cwd is the main checkout. After a shell cwd reset, calls had to `cd` into the worktree first.
