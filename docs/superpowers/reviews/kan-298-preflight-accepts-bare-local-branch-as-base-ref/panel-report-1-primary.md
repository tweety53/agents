Scope matches proposal.md and tasks.md exactly — four files, no more, no less. Everything verified.

**Findings:**

None — no defects. All checks pass.

**Judgments requested by the dispatch:**

1. **Plan alignment** — full branch diff (`git diff --stat ca7b003..HEAD`) touches exactly the four files `proposal.md`/`tasks.md` name: both guards' usage heredocs and both harnesses' new cases. No more, no less. Both `all cases pass` (finish-preflight: 45 lines incl. new; base-moved similarly), confirmed by live run.

2. **`EXPECTED_USAGE` literal — acceptable fixture, not a drift hazard.** It duplicates the guard's wrapped text inside the test file, but this is not a second copy of the *rule* `usage-message-is-the-only-site` protects — that decision governs where the qualification is *authored* (one site per guard's production source), not whether a test may assert the literal string it emits. A test asserting exact stdout/stderr output against a golden string is the standard shape for "no more, no less" behavioral coverage (same pattern already used implicitly by other exact-match assertions in this harness family). If the production wording changes, this test breaks loudly and requires a one-line update alongside — that is a normal test-maintenance cost, not silent drift. Verdict: acceptable fixture.

3. **`MISSING_ARGS_ERR` mktemp file appended to `REPOS`, cleaned via `rm -rf`.** Safe and consistent — `rm -rf` on a plain file is a no-op-if-exists correct removal (does not require `-r`, `-f` suppresses missing-file errors), and the existing `cleanup()` loop already treats `REPOS` as a mixed bag of directories and files (e.g., `NOT_A_REPO` dirs alongside potential file entries) via unconditional `rm -rf`. No change to `cleanup()` was needed or made.

4. **New case placement, ordering, `set -e`/`set -u` safety.** Placed immediately after the existing "missing arguments" case (case 9/11 respectively) and before the next numbered case — cannot mask or reorder anything, since each case is independent and pass/fail accumulate into `$FAILURES` rather than short-circuiting. `set +e` / `set -e` bracket only the guard invocation, matching the harness's established pattern elsewhere (e.g. `run_guard`). `${TMPDIR:-/tmp}` and quoted mktemp template match existing conventions (bash 3.2-safe).

5. **Nothing else broken or left behind.** Verified live: `check-vocabulary.sh` and `check-references.sh` both clean; git worktree confirmed byte-identical to HEAD after two live mutation tests (F1: dropped `>&2` — reproducer failed exactly as predicted, restored via `git checkout --`; F2: dropped explicit `exit 2` — reproducer failed exactly as predicted, restored via `git checkout --`). A transient `git status` flap immediately after test runs (mtime-only, empty `git diff`) settled to clean on retry — not a real mutation.

**Reproducers:**
- `scripts/test-check-finish-preflight.sh` — full suite pass, incl. new cases
- `scripts/test-check-base-moved.sh` — full suite pass, incl. new cases
- `scripts/check-finish-preflight.sh` — bare invocation, confirms exit 2, empty stdout, exact wrapped usage on stderr

**VERDICT: PASS**
