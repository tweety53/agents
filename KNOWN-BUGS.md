# Known bugs

- `scripts/test-check-cleanup-complete.sh` — the fixture survivors commands' timing bounds (the
  5-second inside-bound cases and the 10-second escaper-race case) lose their races when the
  machine is under heavy external load — observed at load average ~24 on 10 cores, where
  `./survivors.sh` spawns slower than the bound and the group kill can beat a fork — failing 3-4
  timing cases that pass on an idle machine; the guard under test and every assertion are
  correct, the bounds are simply not load-proof (KAN-376 raised one bound for the same reason) —
  introduced by 18feb597 (refactor(flow): work only in worktrees, back the branch up remotely,
  drop the myflow legacy), an ancestor of main.

## Deferred review findings

- `stats/internal/guard/panelreproducers.go:140` — F6, Minor, kan-778-agents-port-the-next-five-slowest-guards-to-the — four hand-written refusals
  are untested: a null state record, a finding with `status: null`, and `FLOW_GUARD_REPO_ROOT`
  unset in check-references and check-installed-citations — breaks: a regression goes unseen, and
  losing the null-state check fails open (`REPRODUCERS-OK`, exit 0) — fix: one subtest per
  refusal — deferred: coverage-gap.
- `.flow/project.md:17` — F7, Minor, kan-778-agents-port-the-next-five-slowest-guards-to-the — lists `check-task-commit-fields.py` as a running
  Python guard; the guard is a `flow_guard_exec` shim, the `.py` only a module
  `check-plan-shape.py` imports — breaks: misleads a reader about what runs — fix: reword the
  line — deferred: doc-only.
- `skills/flow/SKILL.md:156` — F8, Minor, kan-778-agents-port-the-next-five-slowest-guards-to-the — names 2 of the 7 `flow-guard` shims in
  `skills/flow/scripts/` as needing `lib/flow-guard.sh` — breaks: an installer following it
  misses five sibling dependencies — fix: list all seven — deferred: doc-only.
- `scripts/reproducer-metachars.sh:25` — F9, Minor, kan-778-agents-port-the-next-five-slowest-guards-to-the — this change removed the last bash
  consumer of `reproducer-metachars.sh` and `scripts/lib/change-plan.sh`, which remain as a
  second copy of logic now in Go — breaks: two copies drift — fix: delete both, or record why they
  stay — deferred: out-of-scope.
- `stats/internal/guard/guard.go:17` — F10, Minor, kan-778-agents-port-the-next-five-slowest-guards-to-the — `Env.LookupEnv` sits beside
  `Getenv` instead of one deriving the other — breaks: an `Env` built with only `Getenv`
  panics in the three guards calling `LookupEnv` — fix: derive `Getenv` from `LookupEnv` —
  deferred: other.
- `stats/internal/guard/check_panel_reproducers_test.go:160` — F11, Minor, kan-778-agents-port-the-next-five-slowest-guards-to-the — case 20 is named
  "cannot-answer, not violations-found" but asserts exit 1 — breaks: misleads a reader of the
  test — fix: rename the case — deferred: cosmetic.
- `stats/internal/guard/installedcitations.go:347` — F12, Minor, kan-778-agents-port-the-next-five-slowest-guards-to-the — TMPDIR comes from the
  injected `Env` but `setup.sh` gets `os.Environ()` — breaks: a test injecting an
  environment does not reach `setup.sh` — fix: build `setup.sh`'s environment from `Env` —
  deferred: other.
- `skills/flow-contracts/jira-followups.md:17` — F13, Minor, kan-778-agents-port-the-next-five-slowest-guards-to-the
  — still frames integrate run 1 as one of several filing sites ("every site that files a
  follow-up", the heading `### The filing site's outstanding items`, "that site's list") —
  breaks: a reader looks for a second filing site that no longer exists — fix: reduce the three
  passages to the integrate run's outstanding items — deferred: doc-only.
- `stats/internal/guard/panelexitcontract.go:98`, `stats/internal/guard/panelreproducers.go:87` — task-9 review F4, Minor, kan-841-agents-port-the-next-ten-slowest-bash-scripts-to — both guards canonicalise the worktree with `filepath.EvalSymlinks`, which succeeds on a directory the caller cannot search, where the bash's `cd … && pwd -P` failed and refused it as vanished — breaks: a `chmod 000` worktree passes the check instead of exiting 2 — fix: also require search permission (`syscall.Access(…, 1)`) as `panelfixsingledispatch.go` now does — deferred: out-of-scope (ports from earlier changes).
- `stats/internal/guard/preparearchivebranch.go:76` — task-7 review, Minor, kan-841-agents-port-the-next-ten-slowest-bash-scripts-to — a relative landing path given from outside the main checkout is resolved by `git -C <main> worktree add` against the main checkout, as the bash did — breaks: it creates `<main>/<path>`, exits 2 ("is not a directory") and leaves that stray worktree registered — fix: absolutise the landing path against the caller's directory before `worktree add` — deferred: out-of-scope (the port keeps the bash's behaviour).
- `stats/internal/guard/preparearchivebranch.go:209` — task-7 review, Minor, kan-841-agents-port-the-next-ten-slowest-bash-scripts-to — porcelain v1 quotes some paths (a rename's `"new dir.txt"`, `"tab\tname"`), which never match the unquoted `diff --name-only` names, as in the bash — breaks: such entries are always classified "does not look like" this change's — fix: read `status --porcelain -z` — deferred: out-of-scope (the port keeps the bash's behaviour).
- `stats/internal/guard/mutateandverify.go:198` — task-6 review, Minor, kan-841 — a signal landing during `git apply` does not stop the main flow before the mutated pass — breaks: a harness can start and be orphaned when the trap exits (the restore still runs) — fix: return from `main` when `r.dying` is set after `apply` — deferred: out-of-scope.
- `stats/internal/guard/mutateandverify.go:288` — task-6 review, Minor, kan-841 — a `git` wrapper whose background child keeps git's stdout open blocks the apply while `mu` is held — breaks: a signal then hangs the trap — fix: bound the wait — deferred: out-of-scope.
- `stats/internal/guard/mutate_and_verify_test.go:742`, `stats/internal/guard/check_guard_symlinks_test.go:187`, `stats/internal/guard/check_finish_preflight_test.go:371` — panel F2/F11, Minor, kan-841 — the parity tests `git show d71a2327:scripts/…` — breaks: `go test ./internal/guard/...` fails in a `--depth 1` clone that lacks that commit — fix: vendor the bash originals as test fixtures, or skip the parity subtests when the commit is absent — deferred: test-environment.
- `stats/internal/guard/resolveremotebase.go:34`, `stats/internal/guard/preparearchivebranch.go:58` — panel F3, Minor, kan-841 — run from a deleted working directory the ports pin every git child's directory to it — breaks: they answer "is not a git worktree" and exit 2 where the bash printed CLEAR or RUN1 — fix: leave `cmd.Dir` unset when the cwd no longer exists — deferred: edge case.
- `stats/internal/guard/basemoved.go:31`, `stats/internal/guard/provereproducer.go:121`, `stats/internal/guard/mutateandverify.go:144` — panel F9, Minor, kan-841 — shared helpers (`smcAbs`, `gdcDirname`, `rrExitCode`, `gsReadable`, `trapExitSignals`) live in unrelated guards' files under those guards' prefixes — breaks: a reader looks for them in the wrong file — fix: move them to one shared file under neutral names — deferred: refactor.
- `stats/internal/guard/mutate_and_verify_test.go:596`, `stats/internal/guard/prove_reproducer_test.go:552` — panel F10, Minor, kan-841 — the ignored-signal tests sleep 500ms and then look — breaks: a handler slower than the sleep passes with the defect present — fix: synchronise on a readiness signal from the child instead of a sleep — deferred: test strength.
- `stats/internal/guard/preparearchivebranch.go:204` — F1, Minor, kan-823-prepare-archive-branch-sh-fails-silently — the new post-run recompute rc == 2 exit path is the only new exit path with no covering test (beyond tasks.md's spec) — breaks: a regression in the recompute-failure exit would land untested — fix: a case corrupting the working-tree read after the branch moves, asserting exit 2 — deferred: coverage-gap.
- `stats/internal/guard/prepare_archive_branch_test.go:625` — F2, Minor, kan-823-prepare-archive-branch-sh-fails-silently — new refusal cases 23-25 drop the file's own c.snap/c.unchanged discipline (all 15 pre-existing refusal cases pin state unchanged; these pin branch only, or nothing) — breaks: a refusal that mutates the landing tree could pass the new cases — fix: add snap/unchanged to the new cases where their fixtures allow a readable state — deferred: coverage-gap.
- `stats/internal/guard/prepare_archive_branch_test.go:625` — F3, Minor, kan-823-prepare-archive-branch-sh-fails-silently — testing principles: refusal cases should pin the full observable behavior including no-mutation (the same defect F2 raises, re-raised under this pass's angle) — breaks: same as F2 — fix: same as F2 — deferred: coverage-gap.
- `stats/internal/guard/prepare_archive_branch_test.go:570` — F8, Minor,
  kan-823-flow-fix-prepare-archive-branch-sh-fails — four test comments (:570,
  :645, :664, :817) still say git's stderr prints "beneath the named line"
  after the prose was corrected to "before" — breaks: a reader of the comments
  learns the wrong output order — fix: four one-word edits ("beneath" →
  "before") — deferred: cosmetic.
