# kan-448

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

## Global constraints

- Bash 3.2 is the floor for everything under `scripts/` — indexed arrays only, no associative
  arrays, no `wait -n` (per `scripts/run-guard-tests.sh`'s own header).
- Every `scripts/check-*.sh` guard must exit clean after every task; the command list lives in
  `.flow/project.md`'s `## lint`.
- Exit-code contracts are load-bearing: `mutate-and-verify.sh` keeps `0`/`2`/`3`/`4` (with `2`
  gaining a second documented meaning), and `prepare-archive-branch.sh` keeps its `0`/`1`/`2`/`3`
  (with `2` gaining a second documented meaning). Never renumber an existing code.
- Out of scope, never touched: `scripts/check-unfinished-work.sh`, `spectre` itself, `stats/`,
  `openspec/`.
- `~/.claude/rules/be-brief.md` — cut, never paraphrase: the task 4 edit deletes the non-gating
  sentence outright, never rewords it.

- [x] 1. Add the post-mutation self-check library

**Build:** green

**Files:**
- Create: `scripts/lib/post-mutation-check.sh`
- Create: `scripts/test-lib-post-mutation-check.sh`

**Tests:** `clean tree matches its snapshot`, `new stash entry is named and fails the check`,
`unexpected status line is named and fails the check`, `snapshot-recorded stash still present is
not drift` — in `scripts/test-lib-post-mutation-check.sh`
**Regression:** reverting this commit removes the only mechanism tasks 2 and 3 source; both of
their suites still pass (they do not reference the lib yet), but the ticket's residue detection
does not exist.
**Baseline:** before=0 after=4
<!-- measured: ls scripts/test-lib-post-mutation-check.sh @ cc704bc — no such suite exists, so the before-count is 0 -->
<!-- predicted: scripts/test-lib-post-mutation-check.sh after task 1 -->
**Commit:** `feat(scripts): add the post-mutation self-check library`

  - [x] **Step 1: Write the failing test suite**

  Create `scripts/test-lib-post-mutation-check.sh`, executable, in this repository's harness
  style (`fail()`/`pass()` printing `FAIL:`/`ok:` lines, indexed `REPOS` array cleaned by an EXIT
  trap, throwaway git repositories under a sandboxed `TMPDIR`):

  ```bash unverified:run it in Step 2 before trusting anything it says; fixture layout copied from scripts/test-mutate-and-verify.sh's header
  #!/usr/bin/env bash
  # Assertion harness for scripts/lib/post-mutation-check.sh. Builds
  # throwaway git repositories under a sandboxed TMPDIR and asserts the
  # library's snapshot/check roundtrip: a clean tree matches its snapshot,
  # a new stash entry is named, an unexpected status line is named, and a
  # stash the snapshot recorded is not drift. Never touches the real
  # repository tree.
  set -euo pipefail

  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  # shellcheck source=lib/post-mutation-check.sh
  . "$SCRIPT_DIR/lib/post-mutation-check.sh"

  FAILURES=0
  fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
  pass() { printf 'ok: %s\n' "$1"; }

  REPOS=()
  cleanup() {
    [ "${#REPOS[@]}" -eq 0 ] && return 0
    for p in "${REPOS[@]}"; do rm -rf "$p"; done
  }
  trap cleanup EXIT

  # new_repo — print a fresh throwaway repository path (one empty baseline
  # commit), registered for cleanup.
  new_repo() {
    local repo
    repo="$(mktemp -d "${TMPDIR:-/tmp}/pmc-lib.XXXXXX")"
    REPOS+=("$repo")
    git -C "$repo" init -q
    git -C "$repo" -c user.name=t -c user.email=t@example.com commit -q --allow-empty -m base
    printf '%s\n' "$repo"
  }

  # 1. A tree that changed not at all matches its snapshot.
  repo="$(new_repo)"
  snap="$(snapshot_tree_state "$repo")"
  if check_tree_restored "$repo" "$snap"; then
    pass "clean tree matches its snapshot"
  else
    fail "clean tree matches its snapshot"
  fi

  # 2. A stash entry the snapshot does not hold is named and fails.
  repo="$(new_repo)"
  snap="$(snapshot_tree_state "$repo")"
  printf 'dirty\n' > "$repo/file.txt"
  git -C "$repo" stash push -q --include-untracked -m kan-448-test-residue
  set +e
  out="$(check_tree_restored "$repo" "$snap")"
  rc=$?
  set -e
  if [ "$rc" -ne 0 ] && printf '%s' "$out" | grep -q 'new stash entry: .*kan-448-test-residue'; then
    pass "new stash entry is named and fails the check"
  else
    fail "new stash entry is named and fails the check (rc=$rc out=$out)"
  fi

  # 3. A status line the snapshot does not hold is named and fails.
  repo="$(new_repo)"
  snap="$(snapshot_tree_state "$repo")"
  printf 'stray\n' > "$repo/stray.txt"
  set +e
  out="$(check_tree_restored "$repo" "$snap")"
  rc=$?
  set -e
  if [ "$rc" -ne 0 ] && printf '%s' "$out" | grep -q 'unexpected status line: .*stray.txt'; then
    pass "unexpected status line is named and fails the check"
  else
    fail "unexpected status line is named and fails the check (rc=$rc out=$out)"
  fi

  # 4. A stash the snapshot recorded is expected, not drift.
  repo="$(new_repo)"
  printf 'kept\n' > "$repo/file.txt"
  git -C "$repo" stash push -q --include-untracked -m kan-448-kept
  snap="$(snapshot_tree_state "$repo")"
  if check_tree_restored "$repo" "$snap"; then
    pass "snapshot-recorded stash still present is not drift"
  else
    fail "snapshot-recorded stash still present is not drift"
  fi

  if [ "$FAILURES" -gt 0 ]; then
    printf 'FAIL: %d case(s) failed\n' "$FAILURES" >&2
    exit 1
  fi
  printf '%s: all cases passed\n' "$(basename "$0")"
  ```

  - [x] **Step 2: Run the suite to verify it fails**

  Run: `scripts/test-lib-post-mutation-check.sh`
  Expected: fails immediately — `post-mutation-check.sh: No such file or directory` when the
  source line runs, so every case is missing rather than passing.

  - [x] **Step 3: Write the library**

  Create `scripts/lib/post-mutation-check.sh`:

  ```bash unverified:the Step 2 suite proves it; watch the empty-array expansions, which are guarded for bash 3.2 + set -u
  # scripts/lib/post-mutation-check.sh — snapshot_tree_state and
  # check_tree_restored, the post-guard sanity check KAN-448 part 1
  # mechanizes: a guard that mutates the working tree snapshots
  # `git status --porcelain=v2` and `git stash list` before its first
  # mutation and checks afterwards that the tree still matches the
  # snapshot — any NEW stash entry or status line is residue the guard
  # names and fails on, instead of leaving a conductor to retry blind
  # (KAN-423's incident).
  #
  # Sourced, never executed. Both functions are read-only. Bash 3.2 floor:
  # indexed arrays only, no associative arrays. Every empty-array expansion
  # is guarded with the ${arr[@]+"${arr[@]}"} idiom — under `set -u` on
  # bash 3.2 a bare "${arr[@]}" over an empty array is an unbound-variable
  # error. Callers prefix their own program name on the output; this
  # library prints bare findings, one per line.
  #
  # Only NEW entries are drift: a stash entry the snapshot recorded and
  # that is still present afterwards is expected, and a deleted entry is
  # not residue the contract names.

  POST_MUTATION_CHECK_SENTINEL='--- post-mutation-check: stash section ---'

  # snapshot_tree_state <worktree> — print the snapshot on stdout: every
  # `git status --porcelain=v2` line, the sentinel, then every
  # `git stash list` line. The caller captures it in a variable.
  snapshot_tree_state() {
    git -C "$1" status --porcelain=v2
    printf '%s\n' "$POST_MUTATION_CHECK_SENTINEL"
    git -C "$1" stash list
  }

  # check_tree_restored <worktree> <snapshot> — recompute status and stash
  # list, diff both against <snapshot>, and print one finding per new
  # stash entry (`new stash entry: ...`) and per unexpected status line
  # (`unexpected status line: ...`). Returns 0 when nothing was printed
  # (the tree matches the snapshot), 1 on any drift, 2 when git itself
  # failed on the recompute.
  check_tree_restored() {
    local worktree="$1" snapshot="$2" line prev in_stash=0 drift=0 found
    local snap_status=() snap_stash=() now_status now_stash
    while IFS= read -r line; do
      [ -n "$line" ] || continue
      if [ "$line" = "$POST_MUTATION_CHECK_SENTINEL" ]; then
        in_stash=1
      elif [ "$in_stash" -eq 0 ]; then
        snap_status+=("$line")
      else
        snap_stash+=("$line")
      fi
    done < <(printf '%s\n' "$snapshot")

    now_status="$(git -C "$worktree" status --porcelain=v2)" || return 2
    now_stash="$(git -C "$worktree" stash list)" || return 2

    while IFS= read -r line; do
      [ -n "$line" ] || continue
      found=1
      for prev in ${snap_stash[@]+"${snap_stash[@]}"}; do
        if [ "$prev" = "$line" ]; then found=0; break; fi
      done
      if [ "$found" -eq 1 ]; then
        printf 'new stash entry: %s\n' "$line"
        drift=1
      fi
    done < <(printf '%s\n' "$now_stash")

    while IFS= read -r line; do
      [ -n "$line" ] || continue
      found=1
      for prev in ${snap_status[@]+"${snap_status[@]}"}; do
        if [ "$prev" = "$line" ]; then found=0; break; fi
      done
      if [ "$found" -eq 1 ]; then
        printf 'unexpected status line: %s\n' "$line"
        drift=1
      fi
    done < <(printf '%s\n' "$now_status")

    return "$drift"
  }
  ```

  - [x] **Step 4: Run the suite to verify it passes**

  Run: `scripts/test-lib-post-mutation-check.sh`
  Expected: `test-lib-post-mutation-check.sh: all cases passed` — all 4 cases print `ok:`
  <!-- predicted: scripts/test-lib-post-mutation-check.sh after this task -->

  - [x] **Step 5: Commit**

  ```bash verified:paths match the task's **Files:** field and the subject matches **Commit:** verbatim
  git add scripts/lib/post-mutation-check.sh scripts/test-lib-post-mutation-check.sh
  git commit -m "feat(scripts): add the post-mutation self-check library"
  ```

- [x] 2. Wire the self-check into mutate-and-verify.sh

**Build:** green

**Files:**
- Modify: `scripts/mutate-and-verify.sh`
- Modify: `scripts/test-mutate-and-verify.sh`

**Tests:** `post-restore stash residue exits 2 naming the stash`,
`touched-file residual keeps exit 3 and the stash is still named` — appended to
`scripts/test-mutate-and-verify.sh`
**Regression:** reverting this commit leaves the library orphaned (task 1's suite still passes)
and restores the gap: a fixture harness that leaves a stash behind reads exit 0 again, so a
conductor retries blind over residue — KAN-423's failure shape.
**Baseline:** before=42 after=44
<!-- measured: scripts/test-mutate-and-verify.sh @ cc704bc — 42 ok, 0 FAIL, the count BEFORE this change's tasks -->
<!-- predicted: scripts/test-mutate-and-verify.sh after task 2 — the 42 existing cases unchanged plus the 2 new ones -->
**Commit:** `feat(scripts): self-check the tree after mutate-and-verify restores`

  - [x] **Step 1: Write the two failing cases**

  Append to `scripts/test-mutate-and-verify.sh`, in that suite's own style (registered `REPOS`
  cleanup, fixture guard built by its existing `build_guard`, patch built as the diff between two
  `build_guard` outputs, harness files written into the fixture repo):

  ```bash unverified:align the fixture/patch helper names with the suite's existing ones before running; the assertions are the contract
  # Case: a fixture harness that, on the mutated pass only (the patch is
  # applied then), stashes the applied mutation away and leaves the entry
  # behind. Today's script restores the touched files and exits 0 — the
  # residue goes unnoticed. After task 2's wiring the EXIT trap must exit 2
  # naming the stash.
  #
  # Case: the same stash plus a staged mutation of the touched file — the
  # trap's `git checkout --` restores the working tree from the index, so
  # the touched file stays dirty, residual is non-empty, and exit 3 keeps
  # precedence over the drift's exit 2 — while the stash is still named in
  # the report.

  stash_harness() {
    local file="$1" touched="$2" staged="$3"
    {
      printf '#!/usr/bin/env bash\n'
      printf 'set -e\n'
      printf 'echo "ok: harness case 1"\n'
      # Only act when the patch is in place (the mutated pass).
      printf 'if grep -q "%s" "%s" 2>/dev/null; then\n' "$MUTATION_MARK" "$touched"
      printf '  git stash push -q --include-untracked -m kan-448-residue\n'
      if [ "$staged" = "stage" ]; then
        printf '  printf "%s" > "%s"\n' "$MUTATION_MARK" "$touched"
        printf '  git add "%s"\n' "$touched"
      fi
      printf 'fi\n'
    } > "$file"
    chmod +x "$file"
  }
  ```

  `MUTATION_MARK` is the mutation string the case's patch applies to the fixture guard (the
  suite's existing `flip:`/`catchall` mutations work — pick one and grep for its observable
  text). Each case asserts, in the suite's `set +e` capture style:

  ```bash unverified:same alignment duty as Step 1's block; rc and output assertions are the contract
  # residue case — expect rc=2 and the stash named:
  if [ "$RC" -eq 2 ] && printf '%s' "$OUT" | grep -q 'new stash entry: .*kan-448-residue'; then
    pass "post-restore stash residue exits 2 naming the stash"
  else
    fail "post-restore stash residue exits 2 naming the stash (rc=$RC)"
  fi

  # precedence case — expect rc=3, the residual report, AND the stash named:
  if [ "$RC" -eq 3 ] \
    && printf '%s' "$OUT" | grep -q 'could not fully restore' \
    && printf '%s' "$OUT" | grep -q 'new stash entry: .*kan-448-residue'; then
    pass "touched-file residual keeps exit 3 and the stash is still named"
  else
    fail "touched-file residual keeps exit 3 and the stash is still named (rc=$RC)"
  fi
  ```

  - [x] **Step 2: Run the suite to verify both cases fail**

  Run: `scripts/test-mutate-and-verify.sh`
  Expected: the 2 new cases FAIL — the residue case reads `rc=0` today and the precedence case's
  report carries no `new stash entry:` line — and the existing cases still pass.
  <!-- predicted: scripts/test-mutate-and-verify.sh at this step: 2 FAIL, 42 ok -->

  - [x] **Step 3: Wire the snapshot and the trap check**

  In `scripts/mutate-and-verify.sh`, source the library beside the other setup (before the first
  `git` call, so `$0`-relative resolution holds regardless of the caller's cwd):

  ```bash unverified:confirm sourcing works when the script is invoked through a relative path from the repo root and by absolute path
  . "$(dirname "$0")/lib/post-mutation-check.sh"
  ```

  Capture the snapshot after all refusal paths (after step 6's dirty check) and immediately
  before step 7's trap install:

  ```bash unverified:placement is the contract — after every refuse/cannot_answer path, before anything mutates
  # Step 6b: snapshot the tree the mutation starts from — status and stash
  # list — so the EXIT trap can tell post-restore residue from the state
  # the run actually left.
  TREE_SNAPSHOT="$(snapshot_tree_state "$REPO_ROOT")"
  ```

  Extend the EXIT trap so the drift check runs after the existing touched-file residual test,
  whose exit 3 keeps precedence:

  ```bash unverified:the Step 2 suite proves the wiring; keep the residual branch's existing output byte-identical
  APPLIED=0
  on_exit() {
    local rc=$?
    if [ "$APPLIED" -eq 1 ]; then
      git checkout -- "${TOUCHED[@]}" 2>/dev/null || true
      local residual drift
      residual="$(git status --porcelain -- "${TOUCHED[@]}" 2>/dev/null || true)"
      if [ -n "$residual" ]; then
        echo "mutate-and-verify: could not fully restore — residual status:" >&2
        printf '%s\n' "$residual" >&2
        rc=3
      elif [ "$rc" -eq 0 ]; then
        # Only claim success once the restore has actually been verified
        # clean — printing this before the restore ran would contradict a
        # residual-status failure reported moments later.
        echo "mutate-and-verify: ran clean — touched files restored"
      fi
      drift="$(check_tree_restored "$REPO_ROOT" "$TREE_SNAPSHOT" || true)"
      if [ -n "$drift" ]; then
        echo "mutate-and-verify: post-restore drift detected:" >&2
        printf '%s\n' "$drift" >&2
        if [ "$rc" -ne 3 ]; then rc=2; fi
      fi
    fi
    exit "$rc"
  }
  trap on_exit EXIT
  ```

  Amend the header's exit-code list so exit 2 carries both meanings and 3's precedence is
  stated:

  ```bash verified:path and code numbers match this file's existing header contract; wording authored in-tree for this change
  #   2  refused before mutating anything — the patch does not apply cleanly,
  #      a file it touches already has uncommitted changes, or a named
  #      harness does not exist or is not executable — OR post-restore
  #      drift: the touched files restored clean but the tree no longer
  #      matches its pre-mutation snapshot (a new stash entry or an
  #      unexpected status line, each named in the report).
  #   3  could not fully restore — after `git checkout --` the touched files
  #      are still not clean; the residual `git status --porcelain` output is
  #      printed and the files may still be mutated. Takes precedence over
  #      exit 2 when both drift kinds fire.
  ```

  - [x] **Step 4: Run the suite to verify it passes**

  Run: `scripts/test-mutate-and-verify.sh`
  Expected: `mutate-and-verify: all cases passed` — 44 cases, 0 FAIL
  <!-- predicted: scripts/test-mutate-and-verify.sh after this task -->

  - [x] **Step 5: Commit**

  ```bash verified:paths match the task's **Files:** field and the subject matches **Commit:** verbatim
  git add scripts/mutate-and-verify.sh scripts/test-mutate-and-verify.sh
  git commit -m "feat(scripts): self-check the tree after mutate-and-verify restores"
  ```

- [x] 3. Wire the self-check into prepare-archive-branch.sh

**Build:** green

**Files:**
- Modify: `scripts/prepare-archive-branch.sh`
- Modify: `scripts/test-prepare-archive-branch.sh`

**Tests:** `stash appearing mid-run exits 2 naming the stash` — appended to
`scripts/test-prepare-archive-branch.sh`
**Regression:** reverting this commit leaves the library with one fewer consumer (task 1's suite
and task 2's suite still pass) and restores the gap: a stash left behind while this script moves
the landing worktree to the archive branch goes unnamed and the script exits 0.
**Baseline:** before=65 after=66
<!-- measured: scripts/test-prepare-archive-branch.sh @ cc704bc — 65 ok, 0 FAIL, the count BEFORE this change's tasks -->
<!-- predicted: scripts/test-prepare-archive-branch.sh after task 3 — the 65 existing cases unchanged plus the 1 new one -->
**Commit:** `feat(scripts): self-check the tree after prepare-archive-branch moves HEAD`

  - [x] **Step 1: Write the failing case**

  Append to `scripts/test-prepare-archive-branch.sh`. The drift must appear between the guard's
  own snapshot and its check, so the case builds a PATH-shim `git` that forwards every invocation
  to the real git and, before forwarding its **second** `stash list` call (the check's recompute;
  the first is the snapshot's), leaves one real stash entry in the landing worktree:

  ```bash unverified:align the repo-building helper names with this suite's existing ones; the shim logic and assertions are the contract
  # Case: a stash entry appearing between the guard's snapshot and its
  # post-run check is named and fails the guard (exit 2).
  SHIM="$(mktemp -d "${TMPDIR:-/tmp}/pab-shim.XXXXXX")"
  REAL_GIT="$(command -v git)"
  LANDING_ABS="$LANDING" SHIM_DIR="$SHIM" REAL="$REAL_GIT" \
  {
    printf '#!/usr/bin/env bash\n'
    printf 'count_file="%s/count"\n' "$SHIM_DIR"
    printf 'if [ "$1" = "stash" ] && [ "$2" = "list" ]; then\n'
    printf '  n="$(cat "$count_file" 2>/dev/null || echo 0)"\n'
    printf '  n=$((n + 1))\n'
    printf '  printf "%%s\\n" "$n" > "$count_file"\n'
    printf '  if [ "$n" = "2" ]; then\n'
    printf '    printf stray > "%s/stray.txt"\n' "$LANDING_ABS"
    printf '    "$REAL" -C "%s" stash push -q --include-untracked -m kan-448-injected\n' "$LANDING_ABS"
    printf '  fi\n'
  } > "$SHIM/git"
  printf 'fi\n' >> "$SHIM/git"
  chmod +x "$SHIM/git"
  ```

  Run the guard with `PATH="$SHIM:$PATH"` through the suite's existing capture helper and
  assert:

  ```bash unverified:same alignment duty as Step 1's shim block; rc and output assertions are the contract
  if [ "$RC" -eq 2 ] && printf '%s' "$ERR" | grep -q 'new stash entry: .*kan-448-injected'; then
    pass "stash appearing mid-run exits 2 naming the stash"
  else
    fail "stash appearing mid-run exits 2 naming the stash (rc=$RC)"
  fi
  ```

  Register `$SHIM` for the suite's EXIT-trap cleanup alongside `$REPOS`.

  - [x] **Step 2: Run the suite to verify the case fails**

  Run: `scripts/test-prepare-archive-branch.sh`
  Expected: the new case FAILS — the guard exits 0 today and its output names no stash — and the
  existing cases still pass.
  <!-- predicted: scripts/test-prepare-archive-branch.sh at this step: 1 FAIL, 65 ok -->

  - [x] **Step 3: Wire the snapshot and the check**

  In `scripts/prepare-archive-branch.sh`, source the library beside the argument validation:

  ```bash unverified:confirm sourcing works when the script is invoked through a relative path from the repo root and by absolute path
  . "$(dirname "$0")/lib/post-mutation-check.sh"
  ```

  Capture the snapshot immediately after the dirty-tree refusal block (the tree is known clean
  and on a known branch there) and before the first tree-mutating `git checkout`:

  ```bash unverified:placement is the contract — after every dirty refusal, before anything moves HEAD
  # Snapshot the tree the branch moves start from — status and stash list —
  # so the post-run check can tell residue from the state the run left.
  TREE_SNAPSHOT="$(snapshot_tree_state "$LANDING")"
  ```

  Run the check immediately before the script's success report, so a drift fails the guard
  rather than reading as a clean positioning:

  ```bash unverified:the Step 2 case proves the wiring; keep the success printf byte-identical
  drift="$(check_tree_restored "$LANDING" "$TREE_SNAPSHOT" || true)"
  if [ -n "$drift" ]; then
    echo "prepare-archive-branch: post-run drift detected:" >&2
    printf '%s\n' "$drift" >&2
    exit 2
  fi

  printf '%s -> %s\n' "$CUR" "$ARCHIVE_BRANCH"
  ```

  Amend the header's exit-code prose so exit 2 carries both meanings (the existing
  "a `git checkout` this script performs fails" meaning and the new post-run drift meaning,
  each naming what the output carries).

  - [x] **Step 4: Run the suite to verify it passes**

  Run: `scripts/test-prepare-archive-branch.sh`
  Expected: `prepare-archive-branch: all cases pass` — 66 cases, 0 FAIL
  <!-- predicted: scripts/test-prepare-archive-branch.sh after this task -->

  - [x] **Step 5: Commit**

  ```bash verified:paths match the task's **Files:** field and the subject matches **Commit:** verbatim
  git add scripts/prepare-archive-branch.sh scripts/test-prepare-archive-branch.sh
  git commit -m "feat(scripts): self-check the tree after prepare-archive-branch moves HEAD"
  ```

- [x] 4. Gate the cross-repo spectre link from the primary checkout

**Build:** green

**Files:**
- Modify: `skills/flow/implement.md` (§2 "Isolate the workspace", the paragraph beginning
  "**After creating each worktree beyond the canonical one**")

**Tests:** **none** — one paragraph of instruction prose in a skill file
**Regression:** reverting this commit restores the non-gating in-worktree link call: every
cross-repo change again logs a refused link ("peer not present") and lands at integrate with a
false OUTSTANDING verdict from `check-unfinished-work.sh` (KAN-439's failure shape).
**Baseline:** n/a — no test declared.
**Commit:** `docs(flow): gate the cross-repo spectre link from the primary checkout`

  - [x] **Step 1: Replace the link paragraph**

  In `skills/flow/implement.md` §2, replace the paragraph beginning "**After creating each
  worktree beyond the canonical one, run `spectre link <canonical-peer>:<name>` in it**" —
  including its closing sentence "A failure is reported and the run continues: the link is not a
  gate, and a change with one worktree runs nothing here." — with:

  ```bash verified:path, function name and peers-file path match the paragraph being replaced; wording authored in-tree for this change and approved in design.md's link-runs-from-primary-cwd and link-placement-after-add decisions
  **After creating each additional worktree, run
  `spectre link --root <worktree>/spectre <canonical-peer>:<name>` with the working directory at
  that repository's primary checkout.** The working directory is what resolves the peers file's
  relative entries — `ResolvePeer` (`spectre/internal/tree/peer.go`) stats a declared peer path
  against the process working directory, so from inside a worktree `../<peer>` resolves into
  `.worktrees/` and the link is always refused — and `--root <worktree>/spectre` is what writes
  the satellite-side `link.md` into the worktree, where `check-unfinished-work.sh` reads it at
  integrate. `<canonical-peer>` is the canonical repository's own name in that worktree's
  `<project>/spectre/peers` file. Record what the command wrote alongside that worktree's merge
  base in this run's working notes. **A refusal is a hard failure of this stage**: report it and
  stop the run — a change whose cross-repo link cannot be established lands at integrate with a
  false OUTSTANDING verdict that forces hand verification. A change with one worktree runs
  nothing here.
  ```

  - [x] **Step 2: Run the markdown guards**

  Run: `scripts/check-vocabulary.sh && scripts/check-references.sh && ~/go/bin/spectre validate "kan-448"`
  Expected: all three exit 0.
  <!-- measured: the same three commands at cc704bc — all three exit 0 before this task -->

  - [x] **Step 3: Commit**

  ```bash verified:paths match the task's **Files:** field and the subject matches **Commit:** verbatim
  git add skills/flow/implement.md
  git commit -m "docs(flow): gate the cross-repo spectre link from the primary checkout"
  ```
