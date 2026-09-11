# kan-493-flow-gather-self-review-context-sh-rejects-every

> **Execution:** `/flow-fast` implements this plan inline. Mark a task's own checkbox once its
> commit lands.
> **Relocation:** no

- [x] 1. Accept a linked worktree as `<repo-root>` in `gather-self-review-context.sh`

**Build:** green

**Files:**
- Modify: `scripts/gather-self-review-context.sh`
- Modify: `scripts/test-gather-self-review-context.sh`

**Tests:** `repo-root-supplied-worktree` — passes a linked worktree as the fourth argument with
cwd outside the repository; asserts exit 0 with the ledger and tasks.md content in the bundle
**Regression:** reverting the commit fails the new case (a worktree root exits 2 again) and
returns the suite to 134 ok
<!-- measured: scripts/test-gather-self-review-context.sh @ branch main -->
**Baseline:** before=134 after=135
<!-- measured: scripts/test-gather-self-review-context.sh @ branch main -->
<!-- predicted: scripts/test-gather-self-review-context.sh after task 1 -->
**Commit:** `fix(scripts): accept a linked worktree as gather-self-review-context repo-root`

  - [ ] **Step 1: add the failing harness case**

  In `scripts/test-gather-self-review-context.sh`, after the
  `repo-root-supplied-relative-archived-path-symlink` case and before the `repo-root-relative`
  refusal case, add:

  ````bash unverified:run scripts/test-gather-self-review-context.sh — step 2 expects this case failing, step 4 passing
  # repo-root-supplied-worktree: a linked worktree is a valid <repo-root> — the shape
  # skills/flow/archive.md step 9 documents (<landing-worktree> as the trust anchor). The
  # fixture files are created INSIDE the worktree (a fresh worktree carries none of the main
  # checkout's untracked files), and cwd is outside the repository, so the bundle can only
  # gather if the fourth argument itself becomes the trusted root.
  new_repo
  WORKTREE_K493="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-worktree-root.XXXXXX")"
  rmdir "$WORKTREE_K493"
  (cd "$REPO" && git worktree add -q -b k493-worktree-root "$WORKTREE_K493" >/dev/null)
  TREES+=("$WORKTREE_K493")
  # Canonicalize the same way the script resolves its arguments (pwd -P), so an OS-level path
  # alias under $TMPDIR (e.g. /var/folders/... -> /private/var/folders/...) cannot produce a
  # spurious lexical/real mismatch before the git check this case exists to exercise — the
  # same handling the F23 case gives $ARCHIVED. Measured against the script's own refusal
  # order: without this, the case fails on "does not exist or resolves through a symlink"
  # instead of the git check it targets.
  WORKTREE_K493="$(cd -P "$WORKTREE_K493" && pwd -P)"
  mkdir -p "$WORKTREE_K493/docs/superpowers/ledgers" "$WORKTREE_K493/spectre/changes/archive/2026-01-01-demo"
  printf 'LEDGER-WORKTREE-ROOT\n' > "$WORKTREE_K493/docs/superpowers/ledgers/2026-01-01-demo.md"
  printf 'TASKS-WORKTREE-ROOT\n' > "$WORKTREE_K493/spectre/changes/archive/2026-01-01-demo/tasks.md"
  OUTSIDE_CWD_K493="$(mktemp -d "${TMPDIR:-/tmp}/gather-test-outside-worktree-root.XXXXXX")"
  TREES+=("$OUTSIDE_CWD_K493")
  set +e
  OUT="$(cd "$OUTSIDE_CWD_K493" && "$SCRIPT" spectre/changes/archive/2026-01-01-demo demo "$STATE_DIR" "$WORKTREE_K493" 2>&1)"
  RC=$?
  set -e
  if [ "$RC" -eq 0 ] && case "$OUT" in *LEDGER-WORKTREE-ROOT*) true ;; *) false ;; esac \
    && case "$OUT" in *TASKS-WORKTREE-ROOT*) true ;; *) false ;; esac; then
    pass "repo-root-supplied-worktree: a linked worktree is accepted as <repo-root>"
  else
    fail "repo-root-supplied-worktree: rc=$RC out=$OUT"
  fi
  (cd "$REPO" && git worktree remove -f "$WORKTREE_K493" >/dev/null 2>&1) || true
  ````

  - [ ] **Step 2: run the harness, expect the new case to fail**

  Run: `scripts/test-gather-self-review-context.sh`
  Expected: `FAIL: repo-root-supplied-worktree: rc=2 ...` ("repo-root '...' is not the root of a
  git repository"), suite exits 1, every other case still ok.

  - [ ] **Step 3: swap the override's validation to `--show-toplevel` equality**

  In `scripts/gather-self-review-context.sh`'s `REPO_ROOT_OVERRIDE` block, replace the
  `git-common-dir` derivation and its equality check:

  ````bash verified:adapted from the REPO_ROOT_OVERRIDE block in scripts/gather-self-review-context.sh @ main, same callers of lexically_collapse and cd -P
  repo_root_toplevel="$(git -C "$repo_root_real" rev-parse --show-toplevel 2>/dev/null || true)"
  if [ -z "$repo_root_toplevel" ] || [ "$repo_root_toplevel" != "$repo_root_real" ]; then
    echo "gather-self-review-context: repo-root '$REPO_ROOT_ARG' is not a git repository root (main checkout or linked worktree)" >&2
    exit 2
  fi
  ````

  In the same commit, update the two comment sites that state the old refusal so the text
  matches: the header NOTE on `<repo-root>` (its "root of a git repository" sentence gains the
  worktree-aware `--show-toplevel` mechanism, KAN-493) and the `REPO_ROOT_OVERRIDE` block
  comment (its "a worktree root is refused here" sentence now states that a linked worktree is
  accepted while `validate_archived_path()` step 1 keeps `--git-common-dir` for the process-cwd
  derivation, F23). No other behavior changes: relative, symlink-divergent, non-repository and
  subdirectory roots still exit 2 before this check runs or by this check itself.

  - [ ] **Step 4: run the harness, expect every case to pass**

  Run: `scripts/test-gather-self-review-context.sh`
  Expected: `gather-self-review-context: all cases pass`, 135 ok lines (134 before + the new
  case).

  - [ ] **Step 5: commit**

  ```bash verified:paths match the Files field
  git add scripts/gather-self-review-context.sh scripts/test-gather-self-review-context.sh
  git commit -m "fix(scripts): accept a linked worktree as gather-self-review-context repo-root"
  ```
