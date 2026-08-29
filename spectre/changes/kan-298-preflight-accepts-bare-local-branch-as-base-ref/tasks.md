# kan-298-preflight-accepts-bare-local-branch-as-base-ref

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

- [x] 1. State the base-ref rule in `check-finish-preflight.sh`'s usage message

**Build:** green
**Files:** `scripts/check-finish-preflight.sh`, `scripts/test-check-finish-preflight.sh`, `scripts/lib/base-ref-usage.sh`
**Tests:** `usage message states the base-ref rule` — a source assertion in `scripts/test-check-finish-preflight.sh`; plus `missing arguments: stdout is empty` and `missing arguments: stderr is exactly the usage message`, two runtime assertions the review panel required (F1/F2)
**Regression:** reverting this commit removes the `<base-ref>` qualification from the guard's stderr usage message, and the new source assertion in `test-check-finish-preflight.sh` fails with `usage message no longer states the base-ref rule`.
**Baseline:** before=44 after=47
<!-- measured: ./scripts/test-check-finish-preflight.sh 2>&1 | grep -c '^ok: ' @ branch main -->
<!-- measured: ./scripts/test-check-finish-preflight.sh 2>&1 | grep -c '^ok: ' @ branch spectre/kan-298-preflight-accepts-bare-local-branch-as-base-ref -->
**Commit:** `docs(scripts): state the base-ref rule in the preflight usage message`

  - [x] **Step 1: Replace the single-line usage `echo` with a quoted heredoc.** The guard's
    missing-argument branch currently runs `echo "usage: check-finish-preflight.sh <worktree>
    <base-ref> <recorded-merge-base|->" >&2`. Replace it with a `cat >&2 <<'EOF'` block carrying
    the synopsis line unchanged plus the four wrapped continuation lines below. Quote the heredoc
    delimiter (`<<'EOF'`, not `<<EOF`) so the `<...>` placeholders and the `origin/` paths are
    written literally and no expansion can occur.

```sh unverified:confirm against the file after editing that the heredoc sits inside the same `if [ -z ... ]` branch and is still followed by `exit 2`
cat >&2 <<'EOF'
usage: check-finish-preflight.sh <worktree> <base-ref> <recorded-merge-base|->
  <base-ref>  the base branch name, bare (main) or remote-tracking
              (origin/main). The guard prefers refs/remotes/origin/<base-ref>
              when it resolves, so a bare name is never tested against a
              stale local branch.
EOF
```

  - [x] **Step 2: Leave the header comment alone.** The header's "What THIS script owns (KAN-88,
    design.md: preflight-resolves-remote-tracking)" paragraph already states the substitution and
    its reasoning. Do not restate the rule in the `# Usage:` synopsis line — `design.md`'s
    `usage-message-is-the-only-site` decision keeps exactly one copy of the phrase per guard so the
    harness assertion in step 3 has a single target.

  - [x] **Step 3: Add the source assertion to `scripts/test-check-finish-preflight.sh`.** Append a
    new numbered case after the last existing one, following case 14's precedent for text whose
    deletion no verdict can reveal.

```sh unverified:confirm the appended case uses this harness's own `pass`/`fail` helpers and `$SCRIPT_DIR`, matching case 14 immediately above it
grep -qF -- 'prefers refs/remotes/origin/<base-ref>' \
  "$SCRIPT_DIR/check-finish-preflight.sh" \
  && pass "usage message states the base-ref rule" \
  || fail "usage message no longer states the base-ref rule"
```

  - [x] **Step 4: Verify.** Run `./scripts/test-check-finish-preflight.sh` — it must print
    `check-finish-preflight: all cases pass` and one more `ok:` line than before. Run
    `./scripts/check-finish-preflight.sh` with no arguments and confirm the wrapped message reaches
    stderr and the exit status is still 2.

  - [x] **Step 5: Mutation-test the new assertion.** Delete the word `prefers` from the guard's
    usage message, re-run the harness, and confirm the new case fails with
    `usage message no longer states the base-ref rule`. Restore the word.

  - [x] **Step 6: Two runtime assertions the review panel required (F1, F2).** The source grep
    above cannot see which stream the guard wrote to, and `run_guard` merges the two, so dropping
    `>&2` from the heredoc survived every test. On the no-arguments path the harness now also
    asserts that stdout is empty and that stderr is **exactly** the usage message — the second
    assertion additionally covering the explicit `exit 2`, whose deletion otherwise falls through
    to the next check and appends a second diagnostic.

  - [x] **Step 7: One shared golden value (F3).** The expected usage text lives once in
    `scripts/lib/base-ref-usage.sh`, whose `base_ref_usage_message <guard-name>` both harnesses
    source and call. **It is an independent golden value and must never be derived from either
    guard** — reading, sourcing or executing the guard to produce it would make both assertions
    tautological. That file is created by this task's commit and is why it appears in `**Files:**`
    above; task 2's harness sources it from there.

- [x] 2. State the same rule in `check-base-moved.sh`'s usage message

**Build:** green
**Files:** `scripts/check-base-moved.sh`, `scripts/test-check-base-moved.sh`
**Tests:** `usage message states the base-ref rule` — a source assertion in `scripts/test-check-base-moved.sh`; plus `missing arguments: stdout is empty` and `missing arguments: stderr is exactly the usage message`, two runtime assertions the review panel required (F1/F2)
**Regression:** reverting this commit removes the `<base-ref>` qualification from `check-base-moved.sh`'s stderr usage message, and the new source assertion in `test-check-base-moved.sh` fails with `usage message no longer states the base-ref rule`.
**Baseline:** before=51 after=54
<!-- measured: ./scripts/test-check-base-moved.sh 2>&1 | grep -c '^ok: ' @ branch main -->
<!-- measured: ./scripts/test-check-base-moved.sh 2>&1 | grep -c '^ok: ' @ branch spectre/kan-298-preflight-accepts-bare-local-branch-as-base-ref -->
**Commit:** `docs(scripts): state the base-ref rule in the base-moved usage message`

  - [x] **Step 1: Apply task 1's step 1 to `scripts/check-base-moved.sh`**, substituting the guard's
    own name on the synopsis line. The four continuation lines are byte-identical to task 1's —
    the two guards share `resolve_remote_base`, so they share the rule.

  - [x] **Step 2: Leave `check-base-moved.sh`'s header alone.** Its "Base-ref resolution is shared
    with check-finish-preflight.sh via scripts/lib/resolve-remote-base.sh (design.md:
    base-moved-is-a-guard)" paragraph already carries the rule by reference.

  - [x] **Step 3: Add the source assertion to `scripts/test-check-base-moved.sh`**, appended before
    the harness's closing `if [ "$FAILURES" -ne 0 ]` block.

```sh unverified:confirm this harness names its helpers `pass`/`fail` and resolves the guard through `$SCRIPT_DIR`, as test-check-finish-preflight.sh does
grep -qF -- 'prefers refs/remotes/origin/<base-ref>' \
  "$SCRIPT_DIR/check-base-moved.sh" \
  && pass "usage message states the base-ref rule" \
  || fail "usage message no longer states the base-ref rule"
```

  - [x] **Step 4: Verify.** Run `./scripts/test-check-base-moved.sh` — it must print
    `check-base-moved: all cases pass` and one more `ok:` line than before. Run
    `./scripts/check-base-moved.sh` with no arguments and confirm the wrapped message reaches
    stderr and the exit status is still 2.

  - [x] **Step 5: Mutation-test the new assertion**, exactly as task 1's step 5 does, against
    `scripts/check-base-moved.sh`.

  - [x] **Step 6: Apply task 1's steps 6 and 7 to this harness.** The same two runtime assertions,
    and the same call to `base_ref_usage_message` — passing `check-base-moved.sh` as the guard
    name, so each harness asserts its own guard's message rather than the other's. This task's
    commit does not create `scripts/lib/base-ref-usage.sh`; task 1's does, and task 1 is this
    commit's ancestor, so this commit is independently green.
