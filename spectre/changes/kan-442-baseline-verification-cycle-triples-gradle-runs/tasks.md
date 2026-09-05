# kan-442-baseline-verification-cycle-triples-gradle-runs

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** make `check-task-commit-fields.py` read-only against the worktree — no revert, no
reset, no stash, no test run — so the guard costs no extra suite per task and the overlap with
the next implementer in `skills/flow/implement.md` is safe by construction; and give the
conductor a stop rule for a guard call that times out.

**Architecture:** pure deletion in `scripts/check-task-commit-fields.py` (the `Regression:`/
`Baseline:` runtime checks and the revert/stash context managers) and in its harness, plus one
new harness case asserting the read-only property; one prose paragraph in
`skills/flow/implement.md`.

**Tech Stack:** Python 3 standard library; bash fixture harness; Markdown skill file.

**Spec:** `spectre/changes/kan-442-baseline-verification-cycle-triples-gradle-runs/design.md`

## Global Constraints

- `Baseline:`/`Regression:` stay in the plan grammar: `TaskFields.baseline`, the `Baseline`/
  `Regression` entries in the field regexes at `scripts/check-task-commit-fields.py` lines 136
  and 254, `parse_task_fields`, and `scripts/lib/plan_grammar.py` are not touched
  (design decision `delete-not-relocate`).
- The frozen `openspec/specs/` tree is not edited and `spectre/specs/` stays empty
  (`no-spec-edit`).
- No lock file, no recovery script (`no-lock-no-recovery`).
- New prose in `skills/flow/implement.md` carries no `SHALL`/`MUST` sentence, so
  `scripts/check-normative-inventory.sh`'s inventory is unchanged; the file stays under its
  32500-byte budget in `scripts/check-contract-budget.sh` (29818 bytes today).
  <!-- measured: wc -c skills/flow/implement.md; grep -n 'skills/flow/implement.md' scripts/check-contract-budget.sh @ main 285cb7d -->
- Every guard listed under `## lint` in `.flow/project.md` exits clean before a task is
  reported done.

---

- [x] 1. Remove the runtime `Regression:`/`Baseline:` verification from the guard and its harness

**Build:** green
**Files:** `scripts/check-task-commit-fields.py`, `scripts/test-check-task-commit-fields.sh`
**Tests:** Case 89: on a worktree carrying a staged edit, an unstaged edit and an untracked file,
the guard exits 0 and leaves `HEAD`, `git status --porcelain` and `git stash list` exactly as it
found them
**Regression:** Case 89 — if the guard still stashes, reverts or resets, `git status --porcelain`
after the run differs from before it (the stash empties the tree and the pop restores it only on
success; a `reset --hard` drops the unstaged edit) or `git stash list` is non-empty; reverting
this task's commit restores `_uncommitted_work_protected`, whose `stash push --include-untracked`
plus `stash pop` round-trip *does* pass the status comparison on a clean run — so Case 89 also
asserts that no `.flow/project.md` `## test` command is needed for exit 0 and that the guard's
output carries no `skipped, not verified` notice, which the reverted code prints whenever a
task declares `Baseline:` counts and the project's `## test` is not a single command.
**Baseline:** before=218 after=197
<!-- measured: bash scripts/test-check-task-commit-fields.sh | grep -c '^ok:' @ main 285cb7d — 218; cases 11 (4), 12 (2), 13 (2), 14 (2), 16 (4), 83 (2), 84 (2), 85 (4), 86 (3) carry 25 assertions between them -->
<!-- predicted: 218 − 25 + 4 (Case 89's four assertions) = 197 after this task's commit; bash scripts/test-check-task-commit-fields.sh | grep -c '^ok:' confirms -->
**Commit:** fix(scripts): make check-task-commit-fields read-only against the worktree

  - [x] **Step 1: Write the failing Case 89 at the end of `scripts/test-check-task-commit-fields.sh`**
    (before the `if [ "$FAILURES" -gt 0 ]` epilogue). Build the fixture with the existing
    helpers — `new_repo`, `write_tasks_md` — then a task commit, then dirty the tree three ways
    and snapshot it:

```bash verified:authored in-tree for this change; helpers and assertion style copied from case 88 directly above it
# ===========================================================================
# Case 89 (KAN-442): the guard is read-only against the worktree. With a
# staged edit, an unstaged edit and an untracked file present, and a task
# declaring Baseline: counts and a Regression: line, it exits 0, moves
# neither HEAD nor the index nor the working tree, pushes no stash, and
# prints no "skipped, not verified" notice — there is no runtime
# Regression:/Baseline: check left to skip.
# ===========================================================================
new_repo
write_tasks_md "$REPO" '- [ ] 1. Read-only guard

**Files:** `alpha.txt`
**Tests:** `test_alpha`
**Regression:** test_alpha fails if alpha.txt is reverted
**Baseline:** before=0 after=1
**Commit:** feat: add alpha
**Build:** green
'
git -C "$REPO" add "spectre/changes/$CHANGE_NAME/tasks.md"
git -C "$REPO" commit -q -m "plan"
printf 'def test_alpha(): pass\n' > "$REPO/alpha.txt"
git -C "$REPO" add alpha.txt
git -C "$REPO" commit -q -m "feat: add alpha"
SHA="$(git -C "$REPO" rev-parse HEAD)"
printf 'staged\n' >> "$REPO/root.txt"
git -C "$REPO" add root.txt
printf 'unstaged\n' >> "$REPO/alpha.txt"
printf 'untracked\n' > "$REPO/scratch.txt"
STATUS_BEFORE="$(git -C "$REPO" status --porcelain)"
run_guard "$REPO" 1 "$SHA"
[ "$RC" -eq 0 ] && pass "case 89: guard exits 0 on a dirty worktree with no ## test command" || fail "case 89: rc=$RC out=$OUT"
[ "$(git -C "$REPO" rev-parse HEAD)" = "$SHA" ] && pass "case 89: HEAD unchanged" || fail "case 89: HEAD moved"
[ "$(git -C "$REPO" status --porcelain)" = "$STATUS_BEFORE" ] && pass "case 89: index and working tree untouched" || fail "case 89: status changed: before=[$STATUS_BEFORE] after=[$(git -C "$REPO" status --porcelain)]"
[ -z "$(git -C "$REPO" stash list)" ] && case "$OUT" in *"skipped, not verified"*) false ;; *) true ;; esac && pass "case 89: no stash entry and no skipped-not-verified notice" || fail "case 89: stash=[$(git -C "$REPO" stash list)] out=$OUT"
```

    Run `bash scripts/test-check-task-commit-fields.sh 2>&1 | grep 'case 89'` and record the
    RED output: the fourth assertion fails today, because `check_baseline` prints
    `task 1: Baseline: skipped, not verified — the project's `## test` command cannot be run as
    a single, targetable command` (the fixture has no `.flow/project.md`, so
    `read_single_test_command` returns `None`). The first three pass today — the stash/pop
    round-trip restores the tree on a clean run — which is why the notice assertion is the one
    that pins the deletion.

  - [x] **Step 2: Delete the runtime checks from `scripts/check-task-commit-fields.py`.** Remove,
    in this order so each removal leaves no dangling reference:
    - `check_regression` and `check_baseline` (lines 1136–1239 today);
    - `_commit_reverted`, `_STASH_MESSAGE` with its comment, and `_uncommitted_work_protected`
      (lines 1029–1134);
    - `read_single_test_command` and `_run_test_command` (lines 976–1027);
    - `CheckOutcome` (lines 356–371);
    - the test-runner-contract comment block and `TARGETED_RESULT_RE`, `TOTAL_COUNT_RE`,
      `PROJECT_TEST_SECTION_RE`, `MARKDOWN_HEADING_RE`, `FENCE_LINE_RE` (lines 288–313);
    - `import contextlib` and `import shlex` (lines 214, 218) — `subprocess` stays for
      `run_git`; `NamedTuple` stays for `TestSpec`.
    In `check_task_commit` (lines 1241–1313): drop the `with _uncommitted_work_protected` block
    and the `notices` list, change the signature's return annotation to `List[str]`, return
    `violations`. In `main`: `violations = check_task_commit(...)`, drop the `for notice in
    notices` loop. Line numbers are those of `main` at `285cb7d`; re-locate by name after each
    deletion.
    <!-- measured: grep -n 'def check_regression\|def check_baseline\|def _commit_reverted\|def _uncommitted_work_protected\|def read_single_test_command\|def _run_test_command\|class CheckOutcome\|^TARGETED_RESULT_RE\|^FENCE_LINE_RE\|^import contextlib\|^import shlex\|def check_task_commit\|def main' scripts/check-task-commit-fields.py @ main 285cb7d -->

  - [x] **Step 3: Reword the docstrings.** Module docstring lines 1–16: the first paragraph
    keeps the frozen-tree citation for "A runtime guard checks each field against the real
    commit" and replaces the sentence citing "Regression and Baseline checks skip, rather than
    fail, when unsupported" with a record that that requirement's runtime check was removed by
    KAN-442 — it reverted, tested and reset the canonical worktree the conductor overlaps with
    the next implementer (KAN-423's incident), cost two full suites per task, and verified
    nothing against any real runner, since it required a `COUNT:`/`RESULT` protocol none speaks;
    `Regression:`/`Baseline:` remain plan declarations the grammar still parses. Exit-code table
    lines 192–207: drop the parenthetical about skipped-not-verified notices under `0` and the
    "a `Regression:` revert that did not make its named test fail, or a `Baseline:` count
    mismatch" clause under `1`. `check_task_commit`'s docstring: "Returns the violations that
    fail the run (exit 1)". Delete the harness's own header sentence, if any, that describes
    the revert cycle — `grep -n 'revert\|stash' scripts/test-check-task-commit-fields.sh` after
    Step 4 must return only Case 89's comment.

  - [x] **Step 4: Delete harness cases 11, 12, 13, 14, 16, 83, 84, 85, 86 and the five fixtures**
    — `write_project_md_test_section`, `write_test_runner`, `write_side_effect_test_runner`,
    `write_double_side_effect_test_runner`, `write_unsupported_test_runner` (lines 50–178) —
    each case from its opening `# ====` rule through the last line before the next case's
    opening rule (cases 11–14: lines 418–519; case 16: 551–584; cases 83–86: 3114–3248, at
    `285cb7d`). Do not renumber the surviving cases: case comments cross-reference each other by
    number (case 30, case 33, case 85's mechanism), and a gap is cheaper than rewriting them.
    `grep -c '^# Case' scripts/test-check-task-commit-fields.sh` reads 83 afterwards.
    <!-- measured: grep -c '^# Case' scripts/test-check-task-commit-fields.sh @ main 285cb7d — 91; 91 − 9 + 1 = 83 -->

  - [x] **Step 5: Verify.** `bash scripts/test-check-task-commit-fields.sh` ends `all cases
    passed` with 197 `ok:` lines; `python3 -m py_compile scripts/check-task-commit-fields.py`;
    `grep -n 'stash\|revert\|reset --hard\|subprocess' scripts/check-task-commit-fields.py` shows
    `subprocess` only in `run_git` and its import; then the full `## lint` list from
    `.flow/project.md` — `scripts/check-guard-symlinks.sh` in particular, since the wrapper's
    `$SCRIPT_DIR/<name>` sibling lines are untouched and must still resolve. Commit with the
    `**Commit:**` subject above.

- [x] 2. Add the post-timeout inspection rule to the conductor dispatch in `skills/flow/implement.md`

**Build:** green
**Files:** `skills/flow/implement.md`
**Tests:** **none** — a prose-only skill edit; the guards under `## lint` (`check-references.sh`,
`check-contract-budget.sh`, `check-normative-inventory.sh`, `check-markdown-integrity.py`,
`check-dispatch-paragraphs.sh`) are its verification
**Regression:** not applicable — no test is added; reverting the commit removes the paragraph and
every `## lint` guard still passes, which is the accepted limit of an unguarded prose rule
(the same limit KAN-441's `FULL SUITE` paragraph accepted)
**Baseline:** not applicable — no test count changes
**Commit:** docs(flow): inspect the worktree before retrying a guard that timed out

  - [x] **Step 1: Insert one paragraph** in `skills/flow/implement.md` under **The next
    implementer overlaps the guard**, immediately after step 2's paragraph (the one ending
    "exit 0 ticks the task in the same call.", line 434 at `285cb7d`) and before step 3:

```markdown verified:authored in-tree for this change; indented three spaces to sit under numbered item 2 like the file's existing list continuation paragraphs
   **A guard call that times out is inspected before it is retried.** Run
   `git status --porcelain=v2 --branch` and `git stash list` in that worktree first. A
   reverting, rebasing or merging state on the `# branch` lines, a change the run did not
   make, or a stash entry the conductor did not push means the tree is not the one the run
   left — end the turn with `## Question` carrying both outputs verbatim; never re-run the
   guard on top of it. (KAN-423: a re-run over a mid-flight revert cost ~55 minutes of hand
   recovery.)
```

    Also amend line 431's sentence "The guard reads git objects and `tasks.md` only, so it is
    safe while the tree changes." to add ", and never stashes, reverts or resets (KAN-442)" — it
    is now true, and the reader should know why it was not always.

  - [x] **Step 2: Verify.** `wc -c skills/flow/implement.md` stays under 32500;
    `scripts/check-contract-budget.sh`, `scripts/check-normative-inventory.sh`,
    `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`,
    `scripts/check-dispatch-paragraphs.sh` and the rest of `## lint` exit 0. Commit with the
    `**Commit:**` subject above.
