# kan-367-check-task-commit-fields-sh-refuses-to-run-while

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

Two tasks: the guard's new direct-resolution path (with its own test coverage), and the call site
that actually supplies the argument that makes it fire. `design.md` is canonical for the decisions;
`proposal.md` for why the change exists.

**Baseline, measured before any edit:**

- 44 harnesses matched by `scripts/test-*.sh`.
  <!-- measured: ls scripts/test-*.sh | wc -l @ d251721 -->
- `scripts/test-check-task-commit-fields.sh` passes all 76 of its own cases.
  <!-- measured: bash scripts/test-check-task-commit-fields.sh @ d251721 -->
- `check-task-commit-fields.sh` accepts 3–5 positional arguments today; a 6th is a usage error.
  <!-- verified: sed -n '113,122p' scripts/check-task-commit-fields.sh @ d251721 -->
- `skills/flow/implement.md` invokes the guard with exactly 5 arguments, once, at its "Guard the
  commit before dispatching review" step.
  <!-- measured: grep -c 'check-task-commit-fields.sh <worktree>' skills/flow/implement.md @ d251721 -->

- [x] 1. `check-task-commit-fields.sh` — resolve directly when given the change name

The wrapper globs `<worktree>/spectre/changes/*/tasks.md` and refuses (exit 2) whenever more than
one unrelated root turns up — correct when the caller genuinely does not know which change a task
belongs to, wrong every time the caller already does. Add an optional 6th positional argument,
`<change-name>`; when given, resolve straight against `$CHANGES_DIR/<change-name>` instead of
scanning every directory under `spectre/changes/`.

Insert this block immediately after `CHANGES_DIR` is computed (currently line 129) and before the
existing `MATCHES=()` glob loop, so a supplied change name short-circuits the whole ambiguity path
via `exec`:

```bash unverified:confirm this sits after CHANGES_DIR is set and before the MATCHES=() loop, and that the case pattern rejects a name carrying '/', '.', or '..'
if [ -n "$CHANGE_NAME" ]; then
  case "$CHANGE_NAME" in
    *[!A-Za-z0-9._-]* | . | .. | */*)
      echo "check-task-commit-fields.sh: invalid change name: $CHANGE_NAME" >&2
      exit 2
      ;;
  esac

  ROOT_DIR="$CHANGES_DIR/$CHANGE_NAME"
  TASKS_MD=""

  if [ -f "$ROOT_DIR/tasks.md" ]; then
    # Scoped version of the existing highest-numbered-fix-sibling rule (see
    # the unchanged glob path below): look only at $CHANGE_NAME's own
    # -fix-N family, never at any other directory under $CHANGES_DIR.
    CHOSEN="$CHANGE_NAME"
    CHOSEN_N=-1
    for change_dir in "$CHANGES_DIR/$CHANGE_NAME"-fix-*/; do
      [ -d "$change_dir" ] || continue
      cname="${change_dir%/}"
      cname="${cname##*/}"
      suffix="${cname##*-fix-}"
      case "$suffix" in '' | *[!0-9]*) continue ;; esac
      [ -f "$CHANGES_DIR/$cname/tasks.md" ] || continue
      if [ "$((10#$suffix))" -gt "$CHOSEN_N" ]; then
        CHOSEN_N="$((10#$suffix))"
        CHOSEN="$cname"
      fi
    done
    TASKS_MD="$CHANGES_DIR/$CHOSEN/tasks.md"
  elif [ -f "$ROOT_DIR/link.md" ]; then
    TASKS_MD="$(change_plan_path "$WORKTREE" "$CHANGE_NAME" "$CANONICAL_WORKTREE" 2>/dev/null || true)"
  fi

  if [ -z "$TASKS_MD" ] || [ ! -f "$TASKS_MD" ]; then
    echo "check-task-commit-fields.sh: no tasks.md found for change '$CHANGE_NAME' under $CHANGES_DIR" >&2
    exit 2
  fi

  if [ -n "$PARENT_SHA" ]; then
    exec python3 "$PYTHON_GUARD" "$TASKS_MD" "$TASK_ID" "$WORKTREE" "$COMMIT_SHA" "$PARENT_SHA"
  fi
  exec python3 "$PYTHON_GUARD" "$TASKS_MD" "$TASK_ID" "$WORKTREE" "$COMMIT_SHA"
fi
```

Wire the new argument in alongside the existing three:

```bash verified:sed -n '113,122p' scripts/check-task-commit-fields.sh @ d251721 — this is the block being edited, reproduced so the diff is unambiguous
if [ "$#" -lt 3 ] || [ "$#" -gt 6 ]; then
  echo "usage: check-task-commit-fields.sh <worktree> <task-id> <commit-sha> [parent-sha] [canonical-worktree] [change-name]" >&2
  exit 2
fi

WORKTREE="$1"
TASK_ID="$2"
COMMIT_SHA="$3"
PARENT_SHA="${4:-}"
CANONICAL_WORKTREE="${5:-}"
CHANGE_NAME="${6:-}"
```

Update the header comment (the "This wrapper's own job..." paragraph and the calling-convention line
near the top of the file) to state the new argument and that it takes priority over the glob when
given — the existing "MORE THAN ONE ROOT IS STILL A REFUSAL" and "THE HIGHEST-NUMBERED FIX SIBLING
WINS" comments stay exactly as they are, since that logic is unchanged for the omitted-argument path.

**Write the test cases FIRST, RED before GREEN**, appended to
`scripts/test-check-task-commit-fields.sh` after case 76 and before the `if [ "$FAILURES" -gt 0 ]`
trailer:

```bash unverified:confirm run_guard's 6-argument pass-through works unchanged (it forwards "$@" as-is) and that new_repo/write_tasks_md behave as in the neighboring cases
# ===========================================================================
# Case 77 (KAN-367): the exact reported scenario — TWO OR MORE unrelated live
# root changes exist under spectre/changes/, which used to be an unconditional
# refusal. Given the change name explicitly, the guard resolves it directly
# without even looking at the unrelated directory.
# ===========================================================================
new_repo "kan-367-demo"
write_tasks_md "$REPO" '- [ ] 1. Named change task

**Files:** `alpha.txt`
**Tests:** `test_alpha`
**Commit:** test: alpha
**Build:** green
'
mkdir -p "$REPO/spectre/changes/some-other-live-change"
printf '%s' '- [ ] 1. Unrelated change task

**Files:** `beta.txt`
**Commit:** test: beta
**Build:** green
' > "$REPO/spectre/changes/some-other-live-change/tasks.md"
git -C "$REPO" add "spectre/changes"
git -C "$REPO" commit -q -m "plan"
printf 'def test_alpha(): pass\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "test: alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA" "" "" "kan-367-demo"
[ "$RC" -eq 0 ] && pass "case 77: a named change resolves directly despite an unrelated live root" \
  || fail "case 77: rc=$RC out=$OUT"
case "$OUT" in
  *"more than one tasks.md"*) fail "case 77: still fell back to the ambiguity refusal: $OUT" ;;
  *) pass "case 77: no ambiguity message" ;;
esac

# ===========================================================================
# Case 78 (KAN-367): the named change is itself mid fix-round — its own
# -fix-N sibling must still win, scoped to just that family, while an
# unrelated third change sitting in the same directory is never consulted.
# ===========================================================================
new_repo "kan-367-fix-demo"
write_tasks_md "$REPO" '- [ ] 1. Parent task, already done

**Files:** `parent-only.txt`
**Commit:** test: parent
**Build:** green
'
mkdir -p "$REPO/spectre/changes/kan-367-fix-demo-fix-1"
printf '%s' '- [ ] 1. Fix-round task

**Files:** `alpha.txt`
**Tests:** `test_alpha`
**Commit:** fix: add alpha
**Build:** green
' > "$REPO/spectre/changes/kan-367-fix-demo-fix-1/tasks.md"
mkdir -p "$REPO/spectre/changes/unrelated-third-change"
printf '%s' '- [ ] 1. Unrelated

**Files:** `gamma.txt`
**Commit:** test: gamma
**Build:** green
' > "$REPO/spectre/changes/unrelated-third-change/tasks.md"
git -C "$REPO" add "spectre/changes"
git -C "$REPO" commit -q -m "plan"
printf 'def test_alpha(): pass\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "fix: add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA" "" "" "kan-367-fix-demo"
[ "$RC" -eq 0 ] && pass "case 78: the named change's own fix sibling wins, scoped to its family" \
  || fail "case 78: rc=$RC out=$OUT"

# ===========================================================================
# Case 79 (KAN-367): a change name that does not exist under spectre/changes/
# still fails loudly rather than silently falling back to the glob.
# ===========================================================================
new_repo "kan-367-demo-79"
printf 'a\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "test: alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
run_guard "$REPO" 1 "$SHA" "" "" "no-such-change"
[ "$RC" -eq 2 ] && pass "case 79: an unknown change name exits 2" \
  || fail "case 79: expected exit 2, got rc=$RC out=$OUT"
case "$OUT" in
  *"no tasks.md found for change 'no-such-change'"*) pass "case 79: names the unresolved change" ;;
  *) fail "case 79: expected a message naming the change, out=$OUT" ;;
esac

# ===========================================================================
# Case 80 (KAN-367): a satellite change resolves through its link.md when
# named directly, exactly as the unnamed glob path already does (case 72),
# proving the new argument reaches the same change_plan_path resolution.
# ===========================================================================
new_repo "kan-367-sat"
cat > "$REPO/spectre/changes/kan-367-sat/link.md" <<'EOF'
## Part of

`peerx:canon-demo-367`
EOF
git -C "$REPO" add "spectre/changes/kan-367-sat/link.md"
git -C "$REPO" commit -q -m "link"
printf 'def test_alpha(): pass\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "add alpha for real"
SHA="$(git -C "$REPO" rev-parse HEAD)"

CANON_WT_80="$(mktemp -d "${TMPDIR:-/tmp}/task-commit-fields-canon.XXXXXX")"
mkdir -p "$CANON_WT_80/spectre/changes/canon-demo-367"
printf '%s' '- [ ] 1. Satellite task named directly

**Files:** `alpha.txt`
**Tests:** `test_alpha`
**Commit:** add alpha for real
**Build:** green
' > "$CANON_WT_80/spectre/changes/canon-demo-367/tasks.md"

run_guard "$REPO" 1 "$SHA" "" "$CANON_WT_80" "kan-367-sat"
[ "$RC" -eq 0 ] && pass "case 80: a satellite named directly resolves through its link" \
  || fail "case 80: rc=$RC out=$OUT"
```

Run it once before writing any implementation to confirm cases 77–80 fail (RED) — the 6-argument
form is a usage error today, so all four should currently exit 2 with the usage message, not the
assertions above. Then apply the implementation blocks above and re-run: all 80 cases, including the
original 76, must pass (GREEN).

```text unverified:run after the implementation is in place
bash scripts/test-check-task-commit-fields.sh
```

**Files:** `scripts/check-task-commit-fields.sh`, `scripts/test-check-task-commit-fields.sh`
**Tests:** `scripts/test-check-task-commit-fields.sh` (cases 77, 78, 79, 80)
**Regression:** reverting this commit returns the guard to refusing on `>1` unrelated root even when
the caller supplies the change name, reproducing KAN-173's own reported failure — every task commit
in a repo with two or more live changes falls through to the hand-run fallback.
**Baseline:** before=44 after=44 harnesses matched by `scripts/test-*.sh`
<!-- predicted: ls scripts/test-*.sh | wc -l after task 1 — no harness file is added or removed, only cases within the existing one -->
**Commit:** `fix(scripts): resolve check-task-commit-fields.sh directly when given the change name`
**Build:** green

- [x] 2. `skills/flow/implement.md` — pass the resolved change name to the guard

The guard is useless to every real `/flow` run until its one call site actually supplies the
argument task 1 adds. `<name>` is already in scope at this call site — it is the change name
resolved once per run and threaded through every stage.

Update the "Guard the commit before dispatching review" paragraph and its fenced invocation:

```bash verified:sed -n '241,249p' skills/flow/implement.md @ d251721 — this is the block being edited, reproduced so the diff is unambiguous
**Guard the commit before dispatching review.** As soon as the implementer reports the task's
commit sha back, and **before** the parent dispatches that task for review, pass the canonical
worktree's absolute path — the worktree created or resumed in step **2** above, recorded in this
run's working notes — as the guard's fifth argument, and this run's own resolved `<name>` as its
sixth, so the guard never has to guess which change a task belongs to even when other changes are
live in the same worktree:

check-task-commit-fields.sh <worktree> <task-id> <task-sha> <task-base> <canonical-worktree> <name>
```

Leave the surrounding "A nonzero exit is a guard failure..." and "When the script cannot be
located..." paragraphs exactly as they are — neither depends on the argument count.

Run `scripts/check-references.sh` and `scripts/check-vocabulary.sh` and confirm both pass; this task
touches no code path, so no other guard applies.

**Files:** `skills/flow/implement.md`
**Tests:** none — this task changes skill prose; task 1's own test suite (cases 77–80) already
covers the guard behavior this call site now exercises, and `scripts/check-references.sh` /
`scripts/check-vocabulary.sh` verify the prose itself.
**Regression:** reverting this commit leaves every real `/flow` run calling the guard with 5
arguments, so task 1's new direct-resolution path never fires and the guard keeps refusing on `>1`
live change exactly as it does today.
**Baseline:** before=44 after=44 harnesses matched by `scripts/test-*.sh`
<!-- predicted: ls scripts/test-*.sh | wc -l after task 2 — a prose-only edit adds no harness -->
**Commit:** `docs(flow): pass the resolved change name to check-task-commit-fields.sh`
**Build:** green
