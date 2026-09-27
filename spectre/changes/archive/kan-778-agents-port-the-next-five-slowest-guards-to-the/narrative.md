# kan-778-agents-port-the-next-five-slowest-guards-to-the — session narrative

## 2026-09-27 — creating run

- The implementation ran as SDD in seven groups. Findings JSON that `flow record findings` never emits
  (empty stdout, `[null]`, `{}`, several values, a non-string `ref`) split the header and body of
  the bash guards. The operator decided to refuse all of it with exit 2, in both
  check-panel-reproducers and check-unfinished-work.
- check-plan-provenance's parity took the longest. The first reviewer showed that the Unicode
  `\s`/`\w` classes, the `{0,20}` bound and `str.strip` sites were unpinned, and one was fail-open.
  94 table rows now pin them. I wrongly judged one `ppLeadingSpace` mutant equivalent, and the
  re-review disproved it.
- The session context was compacted once, mid-panel. After it, a `project-get.sh` result wrapped in
  a ```` ```bash ```` fence was eval'd and hung the citation pre-check; the process was killed
  and the bare command used.
- Several `flow record pass/mutation/status` calls were first sent with an unsupported
  `-session-token` flag and silently failed. They were re-recorded. Some earlier dispatch rows
  passed the token through a shell variable, not as a literal.
- An sdd-phase habit broke the panel-fix key shape: `check-panel-fix-single-dispatch.sh` exits 1 on
  nine `task-<n>-implementer-fix-<k>` rows recorded with `-role panel-fix`. This was auto-resolved
  **Continue**.
- The panel had 4 rounds plus two late-fix reviews. F1 was the only Important: my own plan edit
  widened task 3's `**Files:**` past its commit, which broke `check-task-records.sh`. F2–F5 were
  fixed. F6–F13 are deferred Minors, recorded in `KNOWN-BUGS.md`.
- The operator changed the flow skill mid-run, and both changes landed on this branch:
  - deferred findings go to `KNOWN-BUGS.md` without asking, instead of prompting to file Jira
    (`8dbcbdd6`);
  - rerun policy `full` is retired, and every re-review reads only its diff (`80de983e`).
  Both are recorded as decisions in `design.md`, and each got a primary-only late-fix review.
- The operator asked whether moving guards from shell to Go is worth it. The suite wall time is now
  bounded by `test-setup.sh` (~34s), which is not a guard. The answer given: stop porting small
  guards, and target `test-setup.sh` instead of the planned next five-guard slice.

## 2026-09-27 — integrate run

- Preflight `RUN1`; unfinished-work `CLEAR`; visual-verify `OK` (no UI paths); main checkout `STAGED-CLEAN`/`DRIFT-CLEAN`.
- `origin/main` had moved 4 commits with no path overlap; clean rebase of 42 commits onto `dfb0f6a0`, no conflicts, so no scoped re-verification ran.
- `.flow/project.md`'s `## default landing route` body is `` `merge and push` `` — backtick-wrapped, so the byte-for-byte match dropped it and the landing question was asked; the operator chose merge and push. The backticks are the fix.
