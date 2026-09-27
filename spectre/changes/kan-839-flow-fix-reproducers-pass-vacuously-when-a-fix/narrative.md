# kan-839-flow-fix-reproducers-pass-vacuously-when-a-fix — session narrative

## 2026-09-28 — creating run

Resumed at implementation the same day the plan gated; this session implemented, reviewed and
verified. What the run actually hit, in order:

- **Kickoff (previous session)** — the worktree-location guard flagged a stray worktree
  (`.claude/worktrees/agent-ae1ef826791eba5ee`, branch fully merged, tree clean); removed it after
  verifying it disposable, keeping the branch ref.
- **zsh word-splitting bit the first task-1 commit attempt**: an unquoted `$T1FILES` reached `git
  add` as one pathspec and the commit silently didn't land; the fields guard then correctly
  refused the planning commit at HEAD. Retried with explicit paths. The repo's Bash-is-not-the-shell
  hazard in miniature.
- **Two record-grammar corrections at task close**, both transcribed before the guard re-ran: the
  plan declared the shim at its `skills/flow/scripts/` symlink path (the tracked file is the
  target, `scripts/check-panel-reproducer-exit-contract.sh`), and `**Files:**` paths must be
  backtick-quoted tokens, not comma-separated prose.
- **Task 3's RED was the plan's own snippet being wrong about `rrRun`**: the helper takes the
  runner's argv variadically, so flags passed inside the command-line string never reach
  `--pre-fix-verdict` parsing — the first run demonstrated the runner reading the renamed-premise
  re-run as *demonstrated* instead of refusing. Fixed the call shape, recorded as a dated
  Correction on the task.
- **The base moved mid-panel-entry** (5 commits on origin/main, overlapping
  `skills/flow/review-panel.md`); the operator chose rebase. Clean 7-commit replay onto
  `675a55a9`, no conflicts; the rebased branch needed `--force-with-lease` (the push contract's
  only sanctioned rewrite). The recorded `measured:` comments still name the pre-rebase base as
  the annotation ref — substantively accurate, counts identical at both commits.
- **Panel round 0 (primary+principles, one bundle)**: Minor-only. F1 (deduped
  primary+principles) — a declared-but-unasserted premise escapes both enforcement points, the
  design's accepted residual understating it; F2 — the exit-1 disposition enumeration not
  extended with the new premise class. Both deferred per the Minor-deferral default, entries in
  KNOWN-BUGS.md; no fix round, no re-run.
- **One verify flake**: `TestConcurrentAppendVersusRetirePreservesEveryEntry` (reconcile, a
  package this change never touches) failed once under full-suite load, passed 3× in isolation
  and on the one inline re-run the flake rule allows. No baseline declared; no sweep entry —
  it did not reproduce.
- Main deleted `scripts/test-check-panel-findings-closed.sh` while this change was in flight
  (its coverage moved into Go), which is why the post-rebase guard harness count is 68, not 69.

## 2026-09-28 — integrate run

- Preflight `RUN1`; unfinished-work `CLEAR`; visual-verify `OK` (no UI paths).
- The base-moved check ran against the state map's pre-rebase `4a278320` and reported `MOVED` —
  35 commits, since main gained ~30 more while this change was in panel. The sync rebase onto
  `origin/main` (now `b8faae9a`) hit exactly one conflict: `KNOWN-BUGS.md`, where main's own
  kan-842 deferrals and this change's task-1 deferral appended at the same tail — resolved as a
  union, both sides kept, `--continue` clean.
- Per the resolution rule, the full lint and test lists ran on the rebased tree: the lint list
  shrank by one on main (main deleted `check-contract-budget.sh` and its declaration — the
  KNOWN-BUGS budget ratchet is gone), 58 guard harnesses pass (main's Go-port consolidation),
  `go test ./... -race` 21/21 packages, SPA 170/170.
- Scoped re-verification found no discoverable `scripts/test-<name>.sh` harness for any overlap
  path — main moved that coverage into the Go suite, which the test list above already covers.
- Route: `merge and push`, taken from the project's configured default, not asked.
