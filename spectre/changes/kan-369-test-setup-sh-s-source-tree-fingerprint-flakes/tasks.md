# kan-369-test-setup-sh-s-source-tree-fingerprint-flakes

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

- [x] 1. Prune `__pycache__` from `source_tree_fingerprint()`'s scan

**Files:**
- Modify: `scripts/test-setup.sh` (the `source_tree_fingerprint` function, at `:377-388` before
  this change, and its `find` call)

**Tests:** `a .pyc rewrite under __pycache__ leaves the fingerprint unchanged`, `a real
source-file change still moves the fingerprint`

**Regression:** revert this task's diff and both new assertions fail: the first because the
unpruned `find` re-includes `__pycache__/*.pyc` in the scan, so a `.pyc` mtime rewrite changes
the fingerprint again; the second is a sanity check that stays green either way (it never
touches `__pycache__`), so it is included to prove the fix does not blind the assertion to a
real change, not to catch a regression on revert.

**Baseline:** before=428 after=430
<!-- measured: bash scripts/test-setup.sh @ 3a023e5ae7e274004ab85e19fe6c0a7c606476d5 -->

**Commit:** `fix(test-setup): prune __pycache__ from the source-tree fingerprint`

**Build:** green

  - [ ] **Step 1: Extract the `find` idiom into a shared, parameterized helper**

Replace the body of `source_tree_fingerprint()` (`scripts/test-setup.sh:377-388`) so the `find`
call is shared with the new regression case below, rather than duplicated — the case must call
the exact expression production runs, never a hand-copied approximation of it.

Current code:

```bash verified:read scripts/test-setup.sh:377-388 directly
source_tree_fingerprint() {
  local f
  # scripts/ and README.md are included deliberately: without them the harness cannot
  # detect damage to ITSELF or to the vocabulary guard, which is the one blind spot a
  # write-through incident would exploit twice.
  find "$REPO_ROOT/skills" "$REPO_ROOT/rules" "$REPO_ROOT/commands" "$REPO_ROOT/commands-claude" \
       "$REPO_ROOT/scripts" \
       "$REPO_ROOT/setup.sh" "$REPO_ROOT/CLAUDE.md" "$REPO_ROOT/AGENTS.md" "$REPO_ROOT/README.md" \
       -type f -print0 2>/dev/null | sort -z | while IFS= read -r -d '' f; do
    stat_line "$f"
  done
}
```

Replace it with:

```bash unverified:confirm this diff applies cleanly against the current file and gofmt/shellcheck-equivalent conventions in this file (there is no shell auto-formatter in this repo, per CLAUDE.md's lint list) are unaffected
# tree_fingerprint <path>... — stat-lines every regular file found under the given paths,
# skipping any directory named __pycache__: a generated, .gitignore'd Python bytecode cache is
# not source, the same reason .git itself is never walked. Shared between
# source_tree_fingerprint() below and this harness's own KAN-369 regression case, so the case
# can never drift from what production actually runs.
tree_fingerprint() {
  local f
  find "$@" -name '__pycache__' -prune -o -type f -print0 2>/dev/null | sort -z | while IFS= read -r -d '' f; do
    stat_line "$f"
  done
}

source_tree_fingerprint() {
  # scripts/ and README.md are included deliberately: without them the harness cannot
  # detect damage to ITSELF or to the vocabulary guard, which is the one blind spot a
  # write-through incident would exploit twice.
  tree_fingerprint \
    "$REPO_ROOT/skills" "$REPO_ROOT/rules" "$REPO_ROOT/commands" "$REPO_ROOT/commands-claude" \
    "$REPO_ROOT/scripts" \
    "$REPO_ROOT/setup.sh" "$REPO_ROOT/CLAUDE.md" "$REPO_ROOT/AGENTS.md" "$REPO_ROOT/README.md"
}
```

`-name '__pycache__' -prune -o -type f -print0` is the standard find idiom for "don't descend
into a directory named `__pycache__`, but do print every other regular file found": `-prune`
stops descent for a matched directory before `-o`'s right side ever evaluates on it, and for a
non-directory starting point (`setup.sh`, `CLAUDE.md`, `AGENTS.md`, `README.md` are each passed
to `find` as a bare file, not a directory) `-prune` is a no-op and the file is still printed by
the `-o -type f -print0` branch, exactly as before this change.

  - [ ] **Step 2: Add the two new assertions right after the existing before/after comparison**

Insert this block in `scripts/test-setup.sh` immediately after the existing
`assert_eq "the repo's own skills, rules and commands are unchanged" ...` line (currently
`scripts/test-setup.sh:1138`, at the very end of the file, just before the final
`----------------------------------------` summary block):

```bash unverified:confirm the insertion point still reads exactly this way after Step 1's edit lands (Step 1 only touches lines 377-388, well above this point, so the line should be unchanged)
group "KAN-369: __pycache__ churn does not flake the source-tree fingerprint"

PYCACHE_FIXTURE="$SANDBOX/pycache-fixture"
mkdir -p "$PYCACHE_FIXTURE/scripts/__pycache__" "$PYCACHE_FIXTURE/skills"
printf 'x\n' > "$PYCACHE_FIXTURE/scripts/real.py"
printf 'y\n' > "$PYCACHE_FIXTURE/scripts/__pycache__/real.cpython-314.pyc"

FP_BEFORE="$(tree_fingerprint "$PYCACHE_FIXTURE/scripts" "$PYCACHE_FIXTURE/skills" | hash_stdin)"

# Simulate a sibling harness's Python invocation rewriting the bytecode cache mid-run — the
# exact race KAN-369 reports. A fixed future mtime (portable across GNU and BSD touch, per the
# same two-tool split this file already uses for stat) guarantees a changed %Y even when the
# whole test runs inside one filesystem-mtime-resolution tick.
touch -t 203001010000 "$PYCACHE_FIXTURE/scripts/__pycache__/real.cpython-314.pyc"
FP_AFTER_PYCACHE="$(tree_fingerprint "$PYCACHE_FIXTURE/scripts" "$PYCACHE_FIXTURE/skills" | hash_stdin)"
assert_eq "a .pyc rewrite under __pycache__ leaves the fingerprint unchanged" "$FP_BEFORE" "$FP_AFTER_PYCACHE"

# Sanity check: the prune must not blind the assertion to a real source change — a rewrite of
# an ordinary file in the same tree still has to move the fingerprint.
touch -t 203001010000 "$PYCACHE_FIXTURE/scripts/real.py"
FP_AFTER_REAL="$(tree_fingerprint "$PYCACHE_FIXTURE/scripts" "$PYCACHE_FIXTURE/skills" | hash_stdin)"
assert_ne "a real source-file change still moves the fingerprint" "$FP_BEFORE" "$FP_AFTER_REAL"
```

This fixture lives entirely under `$SANDBOX` (removed by the harness's existing sandbox
cleanup) — it never touches the real `$REPO_ROOT`, keeping this file's own "nothing here writes
outside the sandbox" guarantee intact.

  - [ ] **Step 3: Add the missing `assert_ne` helper**

`scripts/test-setup.sh` has `assert_eq` (`scripts/test-setup.sh:144-146`) but no negated form;
Step 2 needs one. Insert it immediately after `assert_eq`'s definition:

```bash unverified:confirm no assert_ne already exists elsewhere in the file under a different name before adding this
assert_ne() { # <desc> <not-expected> <actual>
  if [[ "$2" != "$3" ]]; then pass "$1"; else fail "$1" "expected something other than [$2], got [$3]"; fi
}
```

  - [ ] **Step 4: Run the full harness and confirm the new assertions pass**

Run: `bash scripts/test-setup.sh`
Expected: `✓ PASS — 430 assertions, 0 failures` (428 baseline + the 2 new assertions from Step 2)
<!-- predicted: bash scripts/test-setup.sh after Step 2 -->

  - [ ] **Step 5: Run the concurrent guard suite to confirm the flake is gone under the actual runner**

Run: `bash scripts/run-guard-tests.sh` (repeat 3 times — the flake KAN-369 reports was 1 failure
in 3 worktree runs, so a single clean pass does not confirm the fix; matching the reproduction
table's own sample size does)
Expected: all 3 runs report every harness `ok`, including `test-setup.sh`, with no
`the repo's own skills, rules and commands are unchanged` failure in any of them

  - [ ] **Step 6: Commit**

```bash unverified:confirm the exact staged path list once Steps 1-3's edits are final
git add scripts/test-setup.sh
git commit -m "fix(test-setup): prune __pycache__ from the source-tree fingerprint"
```
